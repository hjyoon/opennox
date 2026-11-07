#!/usr/bin/env python3
"""Bind retained ARM64 fault observations to sealed code; never execute code.

No original/reference row or expected result is read. The 41-word record lacks
GPRs: this proves code identity/opcode width and the recorded operand fault
range, not a register-derived effective address or whole-memory equivalence.
"""
import argparse
import bisect
import hashlib
import io
import json
import os
from pathlib import Path
import re
import struct
import sys
import uuid
import zipfile

SCHEMA = 'opennox-default-damage-native-PC-contract-v1'
GO_PREFIX = 'github.com/opennox/opennox/v1/legacy.'
ROOT_C = '_nox_xxx_damageDefaultProc_4E0B30'
ROOT_GO = GO_PREFIX + 'rawDefaultDamageNative4E0B30'
READ_GO = GO_PREFIX + 'rawDefaultDamageReadNativeOperand4E0B30'
WRITE_GO = GO_PREFIX + 'rawDefaultDamageWriteNativeOperand4E0B30'
PREDICATE_GO = GO_PREFIX + 'damagePredicateRawRead4E1400'
BEFORE_CALLBACK_GO = GO_PREFIX + 'rawDefaultDamageBeforeNativeCallback4E0B30'
CALLBACK_C = ('_rdd_native_damage_callback_4e0b30', '_rdd_native_sound_callback_4e0b30')
CONTROL_MODES = ('healthy-WORD', 'guarded-WORD', 'nil-C-callback',
                 'unrelated-user-signal', 'other-thread-hardware-fault',
                 'actual-caller-mutate-then-fault')
HEADER_NAMES = ('magic version status signal_number signal_code address pc '
 'expected_operand expected_width site operation loader callback progress '
 'expected_pc expected_sp actual_sp branch_site actual_lr expected_lr '
 'owner_thread actual_thread armed signal_matches_operation unqualified_reason '
 'caller_address caller_entry_bits caller_sample_bits caller_sample_valid '
 'caller_scope_active caller_scope_thread caller_samples caller_observed_bits '
 'caller_observed_progress caller_observed_site caller_observed_count '
 'validation_failed validation_position root_dispatched root_entry root_backend').split()
assert len(HEADER_NAMES) == 41  # Source invariant only; evidence uses require().
MAX_MAIN = 512 * 1024 * 1024
MAX_JSON = 4 * 1024 * 1024

class EvidenceError(Exception):
    pass

def require(condition, message):
    if not condition:
        raise EvidenceError(message)

def digest(data):
    return hashlib.sha256(data).hexdigest()

def file_digest(path):
    with path.open('rb') as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()

def integer(value, label, low=0, high=(1 << 64)-1):
    require(type(value) is int and low <= value <= high, label + ': bounded integer required')
    return value

def hex_sha(value, label):
    require(type(value) is str and re.fullmatch('[0-9a-f]{64}', value), label + ': lowercase SHA256 required')
    return value

def pairs_unique(pairs):
    out = {}
    for key, value in pairs:
        require(key not in out, 'duplicate JSON key: ' + key)
        out[key] = value
    return out

def decode_json(data):
    def bad_constant(value):
        raise EvidenceError('non-finite JSON constant: ' + value)
    return json.loads(data, object_pairs_hook=pairs_unique, parse_constant=bad_constant)

class Inputs:
    def __init__(self):
        self.records = {}

    def read(self, path, limit=MAX_JSON):
        path = Path(path).resolve(strict=True)
        require(path.is_file(), 'regular input file required: ' + str(path))
        require(path.stat().st_size <= limit, 'input exceeds bounded extent: ' + str(path))
        with path.open('rb') as stream:
            data = stream.read(limit+1)
        require(len(data) <= limit, 'input grew beyond bounded extent')
        row = {'path': str(path), 'bytes': len(data), 'sha256': digest(data)}
        previous = self.records.setdefault(str(path), row)
        require(previous == row, 'input changed between reads: ' + str(path))
        return data

    def json(self, path):
        return decode_json(self.read(path))

    def recheck(self):
        for name, row in self.records.items():
            path = Path(name)
            require(path.stat().st_size == row['bytes'] and file_digest(path) == row['sha256'],
                    'source/evidence input changed after qualification: ' + name)

class MachO:
    """Bounded adaptation of the reviewed local Q's Mach-O reader.

    Function ranges for Go come from Go1.27 pclntab, independently cross-bound
    to LC_SYMTAB. C uses the next unstripped text symbol as a conservative end;
    primitive assembly additionally requires its exact reviewed instruction
    certificate. No disassembler binary or ART module is imported.
    """
    def __init__(self, data):
        self.data = data
        self.sections = []
        self.symbols = {}
        self.header_segments = []
        self.non_function_header_symbols = []
        self.go_ranges = None
        require(len(data) >= 32, 'truncated Mach-O header')
        magic, cpu, subtype, filetype, count, command_bytes, flags, reserved = struct.unpack_from('<8I', data)
        require(magic == 0xfeedfacf and cpu == 0x0100000c and filetype == 2,
                'thin little-endian ARM64 MH_EXECUTE required')
        require(count <= 65536 and 32+command_bytes <= len(data), 'Mach-O load-command bounds')
        self.cpu_subtype = subtype
        at, symtab, section_index = 32, None, 0
        for unused in range(count):
            require(at+8 <= 32+command_bytes, 'truncated load command')
            cmd, size = struct.unpack_from('<II', data, at)
            require(size >= 8 and size % 8 == 0 and at+size <= 32+command_bytes, 'load-command extent')
            if cmd == 0x19:
                require(size >= 72, 'LC_SEGMENT_64 extent')
                name = data[at+8:at+24].split(b'\0',1)[0].decode('ascii')
                va, vsize, offset, length, maximum, permissions, nsects, segflags = struct.unpack_from('<4Q4I', data, at+24)
                require(offset+length <= len(data) and 72+80*nsects <= size, 'segment/section bounds')
                if name == '__TEXT':
                    self.header_segments.append({'address':va,'size':vsize,'file_offset':offset,
                                                 'file_size':length,'permissions':permissions})
                # Go1.27 macho_combine_dwarf.go:machoUpdateDwarfHeader makes
                # __DWARF file-backed but non-loadable (Addr/Memsz/Prot=0).
                # Section addresses can remain nonzero; flags can be zero.
                file_only_dwarf = name == '__DWARF' and va == vsize == 0 and permissions == 0
                for i in range(nsects):
                    section_index += 1
                    pos = at+72+80*i
                    secname = data[pos:pos+16].split(b'\0',1)[0].decode('ascii')
                    segname = data[pos+16:pos+32].split(b'\0',1)[0].decode('ascii')
                    address, extent = struct.unpack_from('<QQ', data, pos+32)
                    fileoff, align, relocoff, nreloc, secflags = struct.unpack_from('<5I', data, pos+48)
                    require(segname == name, 'section segment identity')
                    if file_only_dwarf:
                        require(secname.startswith(('__debug_', '__zdebug_')) and secflags & 0xff == 0
                                and secflags & 0x80000400 == 0, 'non-instruction file-only DWARF section')
                        require(offset <= fileoff <= offset+length and extent <= offset+length-fileoff,
                                'file-only DWARF section inside segment file extent')
                    else:
                        require(va <= address <= va+vsize and extent <= va+vsize-address,
                                'section virtual extent')
                    zerofill = secflags & 0xff in (1, 0xc, 0x12)
                    if not zerofill:
                        require(fileoff+extent <= len(data), 'file-backed section bounds')
                    self.sections.append({'index':section_index, 'name':secname, 'segment':segname,
                        'start':address,'end':address+extent,'offset':fileoff,'size':extent,
                        'executable':not file_only_dwarf and bool(permissions & 4) and secname == '__text' and not zerofill})
            elif cmd == 2:
                require(size >= 24 and symtab is None, 'unique LC_SYMTAB required')
                symtab = struct.unpack_from('<4I', data, at+8)
            at += size
        require(at == 32+command_bytes and symtab is not None, 'complete commands and unstripped symbol table required')
        symoff, nsyms, stroff, strsize = symtab
        require(symoff+16*nsyms <= len(data) and stroff+strsize <= len(data), 'symbol/string table bounds')
        strings = data[stroff:stroff+strsize]
        text_indices = {s['index'] for s in self.sections if s['executable']}
        for i in range(nsyms):
            string, typ, section, desc, value = struct.unpack_from('<IBBHQ', data, symoff+16*i)
            if typ & 0xe0 or typ & 0x0e != 0x0e or section not in text_indices:
                continue
            require(0 < string < len(strings), 'symbol string offset')
            end = strings.find(b'\0', string)
            require(end != -1, 'unterminated symbol name')
            name = strings[string:end].decode('utf-8')
            if name == '__mh_execute_header':
                # SDK mach-o/ldsyms.h defines this fixed symbol as the address
                # of the MH_EXECUTE header, not a function in the __text body.
                require(typ == 0x0f and section == 1 and len(self.header_segments) == 1,
                        'fixed MH_EXECUTE header symbol shape')
                header = self.header_segments[0]
                require(header['file_offset'] == 0 and header['file_size'] >= 32+command_bytes
                        and header['size'] >= 32+command_bytes and header['permissions'] & 4
                        and value == header['address'], 'header symbol must bind actual file-zero Mach-O header')
                require(any(s['index'] == section and s['segment'] == '__TEXT' and s['name'] == '__text'
                            and value < s['start'] for s in self.sections), 'header symbol is outside text instructions')
                self.non_function_header_symbols.append({'name':name,'address':value,'section':section,
                    'header_file_offset':0,'header_sha256':digest(data[:32]),'function_symbol':False})
                continue
            require(any(s['index'] == section and s['start'] <= value < s['end'] for s in self.sections),
                    'text symbol outside declared section')
            self.symbols.setdefault(name, set()).add(value)
        self.text_starts = sorted({x for values in self.symbols.values() for x in values})
        require(self.text_starts, 'text symbols required')

    def one(self, name):
        # Go Mach-O may retain a platform underscore; both spellings refer to
        # the same exact Go package/function namespace, never a suffix match.
        spellings = (name, '_'+name) if name.startswith(('github.', 'runtime.')) else (name,)
        values = set().union(*(self.symbols.get(x, set()) for x in spellings))
        require(len(values) == 1, 'missing or ambiguous declared text symbol: '+name)
        return next(iter(values))

    def at(self, address, count):
        require(type(address) is int and address % 4 == 0 and count > 0, 'aligned instruction extent')
        for section in self.sections:
            if section['executable'] and section['start'] <= address and address+count <= section['end']:
                offset = section['offset']+address-section['start']
                return self.data[offset:offset+count]
        raise EvidenceError('PC/code certificate outside file-backed __text')

    def word(self, address):
        return struct.unpack('<I', self.at(address,4))[0]

    def _go_functions(self):
        if self.go_ranges is not None:
            return self.go_ranges
        sections = [s for s in self.sections if s['name'] == '__gopclntab']
        require(len(sections) == 1, 'unique Go pclntab section required for Go function bounds')
        section = sections[0]
        raw = self.data[section['offset']:section['offset']+section['size']]
        require(len(raw) >= 72 and struct.unpack_from('<I',raw)[0] == 0xfffffff1 and raw[4:8] == b'\0\0\x04\x08',
                'Go1.27 ARM64 pclntab header required')
        nfunc,nfiles,unused,names,cu,files,pctab,funcdata = struct.unpack_from('<8Q', raw,8)
        require(0 < nfunc <= 2_000_000 and funcdata+(nfunc*2+1)*4 <= len(raw) and names < len(raw),
                'bounded Go function table')
        text = self.one('runtime.text')
        ranges = {}
        for i in range(nfunc):
            entry, info = struct.unpack_from('<II',raw,funcdata+8*i)
            following = struct.unpack_from('<I',raw,funcdata+8*(i+1))[0]
            require(entry < following and funcdata+info+8 <= len(raw), 'ordered Go function extent')
            entry_copy, nameoff = struct.unpack_from('<Ii',raw,funcdata+info)
            require(entry_copy == entry and 0 <= nameoff < len(raw)-names, 'Go function name/entry binding')
            end = raw.find(b'\0',names+nameoff)
            require(end != -1 and end-(names+nameoff) < 4096, 'bounded Go function name')
            name = raw[names+nameoff:end].decode('utf-8')
            ranges.setdefault(name, []).append((text+entry,text+following))
        self.go_ranges = ranges
        return ranges

    def function(self, name):
        start = self.one(name)
        if name.startswith(GO_PREFIX):
            candidates = self._go_functions().get(name, [])
            require(len(candidates) == 1 and candidates[0][0] == start,
                    'exact Go pclntab and symbol entry agreement: '+name)
            begin,end = candidates[0]
            return {'name':name,'start':begin,'end':end,'boundary':'Go1.27-pclntab-and-LC_SYMTAB'}
        section = next(s for s in self.sections if s['executable'] and s['start'] <= start < s['end'])
        index = bisect.bisect_right(self.text_starts,start)
        end = min(self.text_starts[index] if index < len(self.text_starts) else section['end'],section['end'])
        require(start < end, 'nonempty C text symbol interval')
        return {'name':name,'start':start,'end':end,'boundary':'next-unstripped-text-symbol-conservative-upper-bound'}

    def version(self):
        marker=b'\xff Go buildinf:'
        candidates=[]
        offset=0
        while True:
            offset=self.data.find(marker,offset)
            if offset == -1:
                break
            if offset%16 == 0 and self.data[offset+14:offset+16] == bytes([8,2]):
                at=offset+32;value=0;shift=0
                for unused in range(10):
                    require(at < len(self.data),'build info varint bounds')
                    b=self.data[at];at+=1;value |= (b&127)<<shift
                    if b < 128:
                        break
                    shift+=7
                else:
                    raise EvidenceError('oversized buildinfo varint')
                require(value < 128 and at+value <= len(self.data),'Go version extent')
                candidates.append(self.data[at:at+value].decode('ascii'))
            offset+=1
        require(len(candidates) == 1 and candidates[0] in ('go1.27.3','go1.27.3-X:cgocheck2'),
                'unique pinned Go1.27.3 build information required')
        return candidates[0]

def contains(function, address):
    return function['start'] <= address < function['end']

def scalar_access(word):
    """Conservative scalar decoder; register-offset support has no GPR/EA claim."""
    require(word & (1<<26) == 0, 'SIMD load/store is unsupported')
    opc=(word>>22)&3;size=(word>>30)&3
    index=None
    if word & 0x3b000000 == 0x39000000:
        addressing='unsigned-immediate';immediate=((word>>10)&4095)*(1<<size)
    elif word & 0x3b200c00 == 0x38000000:
        addressing='unscaled-immediate';immediate=(word>>12)&511
        if immediate&256:
            immediate-=512
    elif word & 0x3b200c00 == 0x38200800:
        # Actual Go1.27.3 predicate load b8626804 is LDR W4,[X0,X2].
        # Admit only the unsigned, unshifted Xm/LSL form: no sign/zero
        # extension, scaled index, writeback or effective-address inference.
        require(opc in (0,1) and (word>>13)&7 == 3 and (word>>12)&1 == 0,
                'unsupported register-offset scalar extension/shift/sign')
        addressing='register-offset-LSL-unshifted';immediate=None
        index={'register':(word>>16)&31,'extend':'LSL','shift':0,
               'register_value_captured':False,'effective_address_reconstructed':False}
    else:
        raise EvidenceError('unsupported scalar opcode/addressing class (no pair/atomic/literal/register-extension guess)')
    require(opc in (0,1) or (size == 0 and opc in (2,3)), 'unsupported signed scalar opcode')
    width=1<<size
    load=opc != 0
    return {'word':word,'hex':f'{word:08x}','width':width,'operation':'read' if load else 'write',
            'mnemonic':('LDRSB' if opc in (2,3) else 'LDRB' if width == 1 and load else
                        'STRB' if width == 1 else 'LDRH' if width == 2 and load else
                        'STRH' if width == 2 else 'LDR' if load else 'STR'),
            'addressing':addressing,'immediate':immediate,'base_register':(word>>5)&31,
            'data_register':word&31,'signed_byte':opc in (2,3),
            'index':index}

def arm_primitive(image,slide,w,callback):
    """Exact reviewed ADR + shared-header-store + STRH/BLR primitive sequence."""
    symbol='_rdd_shared_invoke_4e0b30' if callback else '_rdd_shared_strh_4e0b30'
    loader=w[11]-slide
    require(loader == image.one(symbol),'primitive loader symbol identity')
    pc=w[17 if callback else 14]-slide
    count=7 if callback else 5
    start=pc-4*(count-1)
    require(loader <= start <= pc < loader+128,'bounded primitive instruction interval')
    function=image.function(symbol)
    require(contains(function,start) and contains(function,pc),'primitive in declared symbol interval')
    code=image.at(start,4*count);op=struct.unpack('<'+'I'*count,code)
    adr=op[0]
    require(adr&0x9f00001f == 0x10000009,'ADR X9 certificate')
    immediate=((adr>>29)&3)|(((adr>>5)&0x7ffff)<<2)
    if immediate&(1<<20):
        immediate-=1<<21
    require(start+immediate == pc,'ADR names actual primitive opcode')
    def store(word,rt,offset,rn=None):
        require(word&0xffc0001f == 0xf9000000|rt and ((word>>10)&4095)*8 == offset,
                'exact shared-header STR certificate')
        base=(word>>5)&31
        require(base != 31 and (rn is None or base == rn),'single actual header base register')
        return base
    if callback:
        base=store(op[1],9,136)
        require(op[2] == 0x9100112a,'ADD X10,X9,#4')
        store(op[3],10,152,base)
        require(op[4] == 0x910003ea,'MOV X10,SP')
        store(op[5],10,120,base)
        require(op[6]&0xfffffc1f == 0xd63f0000,'BLR actual callback register')
    else:
        base=store(op[1],9,112)
        require(op[2] == 0x910003ea,'MOV X10,SP')
        store(op[3],10,120,base)
        require(op[4]&0xffc00000 == 0x79000000 and (op[4]>>10)&4095 == 0,
                'STRH Wt,[Xn,#0] actual instruction')
    return {'function':function,'loader_VA':loader,'instruction_VA':pc,'code_hex':code.hex(),
            'code_sha256':digest(code),'mnemonic':'BLR' if callback else 'STRH',
            'PC_SP_LR_header_stores_decoded':True}

def validate_contract(contract):
    require(contract.get('schema') == SCHEMA,'exact source-contract schema')
    require(contract.get('root_symbols') == {'C':ROOT_C,'Go':ROOT_GO},'fixed planned root symbol identities')
    require(contract.get('operand_functions') == {'read':READ_GO,'write':WRITE_GO,'predicate_read':PREDICATE_GO},
            'fixed getter/write/predicate identities')
    require(contract.get('production_callback_wrappers') == list(CALLBACK_C),'fixed production callback wrapper identities')
    require(contract.get('shared_header_words') == 41 and contract.get('shared_header_version') == 2,
            'fixed 41-word shared schema')
    require(contract.get('original_or_expected_data_used') is False,'contract must exclude original/expected results')
    require(contract.get('supported_scalar_addressing') == [
        'integer unsigned-immediate', 'integer unscaled-immediate without writeback',
        'unsigned integer register-offset LSL Xm unshifted (option3,S0)'],
        'exact declared scalar addressing scope')

def shared_record(inputs,directory,audit):
    layout=inputs.json(directory/'actual-native-shared-layout-before-entry.json')
    require(layout.get('schema') == 'opennox-raw-default-damage-shared-native-layout-v2','shared layout schema')
    page=integer(layout.get('page'),'page',4096,65536)
    require(page&(page-1) == 0,'power-of-two actual page')
    n=integer(layout.get('allocation_count'),'allocation_count',1,256)
    allocations=layout.get('allocations')
    require(type(allocations) is list and len(allocations) == n and layout.get('header_words') == 41 and
            layout.get('file_pages') == 1+2*n,'full shared extent')
    raw=inputs.read(directory/'actual-native-shared-AFTER.bin',(1+2*n)*page)
    require(len(raw) == (1+2*n)*page,'exact complete shared AFTER bytes')
    w=list(struct.unpack_from('<41Q',raw))
    audit['raw_header']=dict(zip(HEADER_NAMES,w))
    audit['shared_AFTER_sha256']=digest(raw)
    require(w[0] == 0x3146444452574152 and w[1] == 2,'shared magic/version')
    require(raw[41*8:page] == bytes(page-41*8),'fixed unused header bytes')
    native_extents=[]
    for i,a in enumerate(allocations):
        require(a.get('actual_offset') == (i+1)*page and a.get('expected_offset') == (1+n+i)*page,
                'exact actual/shadow bank offsets')
        base=integer(a.get('native_actual_base'),'native actual page',1<<32)
        shadow=integer(a.get('native_expected_base'),'native shadow page',1<<32)
        require(base%page == shadow%page == 0,'native page alignment')
        for begin in (base,shadow):
            require(all(begin+2*page <= old or old+2*page <= begin for old in native_extents),
                    'actual/shadow guarded reservation overlap')
            native_extents.append(begin)
    audit['memory_scope']='Complete AFTER bytes hashed only; no original-row or shadow correctness comparison.'
    return layout,w

def last_observation(inputs,directory,w,operation):
    raw=inputs.read(directory/'actual-native-observable.jsonl',64*1024*1024)
    require(raw.endswith(b'\n'),'complete observable JSONL terminal line')
    last=None
    for i,line in enumerate(raw.splitlines()):
        require(line and len(line) <= MAX_JSON,'bounded nonempty observable event')
        event=decode_json(line)
        require(event.get('Sequence') == i,'actual observable monotonic sequence')
        last=event
    require(last is not None and last.get('Position') == w[13] and last.get('Site') == w[9],
            'last actual observation/progress/site binding')
    if operation == 2:
        require(last.get('Kind') == 'callback' and last.get('Callback') == w[12] and last.get('Complete') is False,
                'actual incomplete callback observation')
    else:
        require(last.get('Kind') == 'operand' and last.get('Address') == w[7] and last.get('Width') == w[8]
                and last.get('Write') is (operation == 1) and last.get('Complete') is False,
                'actual incomplete operand observation')
    return last

def scalar_fault(image,slide,w,observed):
    pc=w[6]-slide
    names=(WRITE_GO,) if w[10] == 1 else (READ_GO,PREDICATE_GO)
    # A compiler nil check can precede a store in the write helper; a read is
    # not admitted in a write helper except for the explicit nil-check case.
    found=[]
    for name in names:
        function=image.function(name)
        if contains(function,pc):
            found.append(function)
    require(len(found) == 1,'fault PC must be in exactly one declared raw operand function')
    opcode=scalar_access(image.word(pc))
    nil_probe=(w[7] == w[5] == 0 and observed.get('Base') == 0 and opcode['operation'] == 'read'
               and opcode['width'] == 1 and opcode['immediate'] == 0 and opcode['data_register'] == 27)
    if nil_probe:
        require(opcode['mnemonic'] in ('LDRB','LDRSB'),'explicit Go ARM64 nil-check byte probe')
        kind='nil-check-before-access'
    else:
        require(not opcode['signed_byte'],'signed byte load is admitted only as explicit nil-check')
        require(opcode['operation'] == ('write' if w[10] == 1 else 'read') and opcode['width'] == w[8],
                'actual load/store direction and width must match retained operand')
        kind='incomplete-scalar-access'
    return {'kind':kind,'function':found[0],'instruction_VA':pc,'instruction_runtime_PC':w[6],
            'instruction_bytes_hex':image.at(pc,4).hex(),'decoded':opcode,
            'semantic_operand_width':w[8],'physical_instruction_width':opcode['width'],
            'access_completed':False,'nil_check_is_not_completed_semantic_access':nil_probe,
            'effective_address_register_reconstructed':False}

def callback_fault(image,slide,w):
    require(w[8] == 0 and w[6] == w[5] == w[12], 'actual fetch PC/address/callback identity')
    require(w[18] >= slide+4 and w[16] != 0 and w[16]%16 == 0,'captured callback LR and aligned SP')
    branch=w[18]-4-slide
    functions=[image.function(name) for name in CALLBACK_C]
    matches=[f for f in functions if contains(f,branch)]
    require(len(matches) == 1,'actual LR-4 must belong to a declared production callback wrapper')
    word=image.word(branch)
    require(word&0xfffffc1f == 0xd63f0000,'actual callback callsite must be BLR')
    require(w[17] == 0 or w[17] == branch+slide,'retained optional branch-site agreement')
    require(w[19] == 0 or w[19] == w[18],'retained optional expected LR agreement')
    require(w[15] == 0 or w[15] == w[16],'retained optional expected SP agreement')
    return {'kind':'incomplete-callback-target-fetch','function':matches[0],
            'instruction_VA':branch,'instruction_runtime_PC':branch+slide,'instruction_bytes_hex':image.at(branch,4).hex(),
            'decoded':{'mnemonic':'BLR','register':(word>>5)&31},'faulting_fetch_PC':w[6],
            'callback_target':w[12],'actual_LR':w[18],'actual_SP':w[16],'access_completed':False,
            'GPR_callback_register_value_reconstructed':False}

def qualify(args,inputs,audit):
    contract_raw=inputs.read(args.source_contract)
    require(digest(contract_raw) == hex_sha(args.source_contract_sha256,'source-contract SHA'),'source-contract SHA mismatch')
    contract=decode_json(contract_raw);validate_contract(contract)
    archive=inputs.read(args.main_archive,MAX_MAIN)
    require(digest(archive) == hex_sha(args.archive_sha256,'archive SHA'),'sealed main ZIP SHA mismatch')
    with zipfile.ZipFile(io.BytesIO(archive)) as z:
        require(z.namelist() == ['actual-test-main'],'exact one canonical PREEXEC executable member')
        info=z.infolist()[0]
        require(0 < info.file_size <= MAX_MAIN and not info.flag_bits&1,'bounded unencrypted main member')
        with z.open(info) as stream:
            main=stream.read(MAX_MAIN+1)
        require(len(main) == info.file_size and digest(main) == hex_sha(args.main_sha256,'main SHA'),
                'whole actual main bytes/SHA mismatch')
    image=MachO(main)
    audit['main']={'bytes':len(main),'sha256':digest(main),'GoVersion':image.version(),'cpu_subtype':image.cpu_subtype}
    directory=Path(args.child_directory).resolve(strict=True)
    require(directory.is_dir(),'retained child directory')
    pre=inputs.json(directory/'actual-child-PREEXEC.json')
    terminal=inputs.json(directory/'actual-child-terminal.json')
    intent=inputs.json(directory/'parent-intent-before-child-exec.json')
    audit['terminal']=terminal
    pid=integer(terminal.get('PID'),'actual terminal PID',1)
    integer(terminal.get('Exit'),'actual terminal exit',-1,255)
    require(pre.get('before_exec') is True and integer(pre.get('pid'),'actual PREEXEC PID',1) == pid,
            'actual same-PID PREEXEC')
    binary=pre.get('binary',{})
    require(binary.get('bytes') == len(main) and binary.get('sha256') == digest(main),'PREEXEC main identity')
    require(type(pre.get('argv')) is list and pre['argv'] and pre['argv'][0] == binary.get('path'),
            'PREEXEC actual executable argv')
    require(intent.get('main_SHA256') == digest(main) and intent.get('timeout_ns') == 30_000_000_000,
            'bounded actual child intent/main identity')
    require(terminal.get('TimedOut') is False and terminal.get('Forced') is False,
            'timeouts or forced children are unqualified')
    layout,w=shared_record(inputs,directory,audit)
    primitive=args.primitive_control
    entry=inputs.json(directory/('child-before-C-entry.json' if primitive else 'child-before-root-entry.json'))
    require(entry.get('PID') == pid and entry.get('actual_main_SHA256') == digest(main),'before-entry same child/main')
    root_symbol=ROOT_C if primitive or w[40] == 2 else ROOT_GO
    root_runtime=entry.get('raw_C_root_entry') if primitive else w[39]
    integer(root_runtime,'runtime root entry',1<<32)
    slide=root_runtime-image.one(root_symbol)
    require(slide >= 0 and slide%4096 == 0,'unique root-derived page-aligned image slide')
    audit['slide']=slide;audit['root_symbol']=root_symbol;audit['root_entry']=root_runtime
    if primitive:
        require(entry.get('mode') == primitive and entry.get('native_root_algorithm_present') is False and
                entry.get('registered_parent_invoked') is False,'primitive is not a native parent execution')
        require(w[38:41] == [0,0,0] and w[13] == 0,'primitive has no parent dispatch/progress')
    else:
        require(w[38] == 1 and w[40] in (1,2) and entry.get('root_entry') == w[39] and
                entry.get('backend') == ('Go' if w[40] == 1 else 'C'),'actual native root/backend dispatch')
        require(entry.get('raw_C_root_entry') == image.one(ROOT_C)+slide,'C root confirms same image slide')
        for key,name in [('generic_read_entry',READ_GO),('generic_write_entry',WRITE_GO),
                         ('callback_before_observer_entry',BEFORE_CALLBACK_GO)]:
            require(entry.get(key) == image.one(name)+slide,'before-entry exact helper address: '+key)
        require(pre['argv'][1:] == ['-test.run=^TestRawDefaultDamageNativeRowWorker4E0B30$','-test.v'],
                'dedicated native row child argv')
    require(w[36] == 0,'validation_failed evidence cannot qualify')
    if primitive == 'healthy-WORD':
        require(terminal.get('Exit') == 0 and w[2] == 2 and w[3] == 0 and w[10] == 1 and w[8] == 2,
                'healthy completed primitive header')
        require(w[20] != 0 and w[21] == 0,'healthy owner exists without a signal-thread observation')
        instruction=arm_primitive(image,slide,w,False)
        return {'Qualified':False,'classification':'completed-primitive-control-no-fault-PC',
                'primitive_instruction_certificate':instruction,'native_parent_engine_qualified':False}
    # Darwin raise(SIGBUS) can reuse a numerically hardware-compatible si_code.
    # Code values never alone prove origin: source PC/instruction/operand and
    # same-owner record are mandatory below, including for rejection controls.
    require((w[3] == 11 and w[4] in (1,2)) or (w[3] == 10 and w[4] in (1,2,3)),
            'signal number/code outside kernel-compatible captured domain')
    require(w[20] != 0 and w[21] == w[20],'actual signal thread must equal captured owner')
    require(w[2] in (3,5) and w[24] == 0 and terminal.get('Exit') == (170 if w[2] == 3 else 172),
            'unqualified status/reason/terminal exit')
    require(w[10] in (1,2,3),'declared read/write/callback operation')
    if w[10] != 2:
        require(w[8] in (1,2,4,8) and w[7] <= w[5] < w[7]+w[8],
                'actual signal address outside retained attempted operand extent')
    else:
        require(w[8] == 0 and w[5] == w[6] == w[12],'callback fetch address/PC/target')
    if primitive:
        require(pre['argv'][1:] == ['-test.run=^TestRawDefaultDamageSharedFaultControls4E0B30$','-test.v'],
                'dedicated primitive child argv')
        require(primitive not in ('unrelated-user-signal','other-thread-hardware-fault'),
                'negative signal/thread control must remain rejected')
        require(terminal.get('Exit') == 170 and w[2] == 3 and w[22] == w[23] == 1,
                'exact armed primitive fault record')
        callback=primitive in ('nil-C-callback','actual-caller-mutate-then-fault')
        require(w[10] == (2 if callback else 1),'primitive operation identity')
        certificate=arm_primitive(image,slide,w,callback)
        require(w[16] == w[15] and w[16] != 0,'actual primitive SP certificate')
        if callback:
            require(w[18] == w[19] == w[17]+4,'actual primitive BLR/LR certificate')
        else:
            require(w[6] == w[14] and w[8] == 2,'actual primitive STRH PC/width')
        instruction={'kind':'primitive-callback-fetch' if callback else 'primitive-incomplete-STRH',**certificate}
    else:
        require(w[2] == 5 and w[22] == w[23] == 0,'generic native fault must not borrow primitive armed certificate')
        observed=last_observation(inputs,directory,w,w[10])
        instruction=callback_fault(image,slide,w) if w[10] == 2 else scalar_fault(image,slide,w,observed)
    return {'Qualified':True,'ExecutableSHA':digest(main),'PC':w[6],'Address':w[5],'Site':w[9],
            'RootEntry':root_runtime,'Backend':None if primitive else entry['backend'],
            'classification':instruction['kind'],'instruction':instruction,
            'native_parent_engine_qualified':False,'original_result_or_memory_equivalence_qualified':False,
            'signal_code_alone_proves_hardware_origin':False,'effective_address_register_reconstructed':False,
            'source_identity_scope':'Exact sealed executable symbol namespace and opcode; caller build-source provenance remains separately bound upstream.'}

def write_exclusive(path,value):
    path=Path(path)
    require(path.parent.is_dir(),'existing output directory required')
    flags=os.O_WRONLY|os.O_CREAT|os.O_EXCL|getattr(os,'O_NOFOLLOW',0)
    fd=os.open(path,flags,0o600)
    with os.fdopen(fd,'w',encoding='utf-8') as stream:
        json.dump(value,stream,indent=2,sort_keys=True,allow_nan=False)
        stream.write('\n');stream.flush();os.fsync(stream.fileno())

def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--main-archive',required=True,type=Path)
    parser.add_argument('--archive-sha256',required=True)
    parser.add_argument('--main-sha256',required=True)
    parser.add_argument('--child-directory',required=True,type=Path)
    parser.add_argument('--source-contract',required=True,type=Path)
    parser.add_argument('--source-contract-sha256',required=True)
    parser.add_argument('--primitive-control',choices=CONTROL_MODES)
    parser.add_argument('--output',type=Path)
    args=parser.parse_args()
    inputs=Inputs()
    audit={'schema':'opennox-default-damage-retained-native-PC-audit-v1',
           'primitive_control':args.primitive_control,'original_or_expected_data_used':False,
           'external_processes_executed':0,'limitations':['41-word record has no GPR bank.',
             'No register-derived effective address or complete native stack/heap equality.',
             'No original result, row, service reply, or desired fault drives this decoder.']}
    directory=args.child_directory.resolve()
    output=args.output or directory/'actual-native-PC-qualified.json'
    try:
        inputs.read(Path(__file__),MAX_JSON)
        require(not output.exists(),'qualified output already exists; preserve it without overwrite')
        result=qualify(args,inputs,audit)
        inputs.recheck()
        result.update(audit);result['input_bindings']=list(inputs.records.values())
        if not result['Qualified']:
            path=output.parent/'actual-native-PC-completed-control-observation.json'
        else:
            path=output
        write_exclusive(path,result)
        print(json.dumps({'output':str(path),'sha256':file_digest(path),'Qualified':result['Qualified'],
                          'classification':result['classification']}))
        return 0
    except Exception as exc:
        # An unexpected parser/type failure is also an unqualified observation;
        # keep its evidence and exception class rather than losing the receipt.
        audit.update({'Qualified':False,'rejection':str(exc),'rejection_exception':type(exc).__name__,
                      'input_bindings':list(inputs.records.values())})
        path=output.parent/('actual-native-PC-rejected-'+uuid.uuid4().hex+'.json')
        try:
            write_exclusive(path,audit)
        except (OSError,EvidenceError) as write_error:
            print(json.dumps({'Qualified':False,'rejection':str(exc),'rejection_receipt_error':str(write_error)}),file=sys.stderr)
            return 2
        print(json.dumps({'Qualified':False,'rejection':str(exc),'rejection_receipt':str(path),
                          'sha256':file_digest(path)}),file=sys.stderr)
        return 2

if __name__ == '__main__':
    sys.exit(main())
