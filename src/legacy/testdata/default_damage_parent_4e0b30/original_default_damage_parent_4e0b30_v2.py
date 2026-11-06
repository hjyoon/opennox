#!/usr/bin/env python3
"""Prepared PE32 recorder. No original code executes on import or without --execute.

Only parent1783, child112/34, converter17 execute as original bodies. All other
services are explicit frozen input boundaries, with no helper execution credit.
The producer observes original outcomes; it contains no native damage algorithm.
V2 additionally records bounded emulator-observed partial writes on a memory
fault; the instruction stays incomplete and native fault atomicity is unclaimed.
"""
from __future__ import annotations

import argparse
from collections import Counter
import gc
import gzip
import hashlib
import json
from pathlib import Path
import struct
import sys
import traceback

EXE_SHA = '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
MAIN_SHA = 'cbe217010a1bffd6dd75299ac02629d7bda88b396ea7d52450bd8524cf972afd'
PARENT = 0x004e0b30
STACK, STACK_SIZE, CALLER_SP = 0x22000000, 0x10000, 0x2200e000
AUX, AUX_SIZE, STOP = 0x23000000, 0x10000, 0x23000100
INIT, CW_ADDRESS, PAYLOAD, TRAMPOLINE = 0x23000200, 0x23000300, 0x23000400, 0x23000500
CALLBACKS = {0x004e0d05: 0x23001000, 0x004e1049: 0x23002000, 0x004e104f: 0x23002000}
MODES = [('startup-53-chop', 0x0e3f), ('nearest-53', 0x023f)]
MAX_INSTRUCTIONS = 20000
MAX_GRAPH_BYTES = 0x200000


def sha(data):
    return hashlib.sha256(data).hexdigest()


def integer(value):
    assert isinstance(value, (int, str)) and not isinstance(value, bool), value
    if isinstance(value, str):
        if len(value) == 8 and all(c in '0123456789abcdefABCDEF' for c in value):
            return int(value, 16)  # canonical unprefixed PE address/site strings
        return int(value, 0)
    return value


def word(value):
    value = integer(value)
    assert 0 <= value <= 0xffffffff, value
    return value


def hx(value):
    return f'0x{value:08x}'


def canonical(obj):
    return json.dumps(obj, sort_keys=True, separators=(',', ':')).encode()


def record(path):
    path = Path(path).resolve()
    data = path.read_bytes()
    return dict(path=str(path), bytes=len(data), sha256=sha(data))


def bound(row):
    observed = record(row['path'])
    assert observed['bytes'] == row['bytes'] and observed['sha256'] == row['sha256'], observed
    return observed


def write_json(path, obj):
    with Path(path).open('x') as f:
        json.dump(obj, f, indent=2)
        f.write('\n')


class PE:
    def __init__(self, path):
        self.blob = Path(path).read_bytes()
        assert len(self.blob) == 1929216 and sha(self.blob) == EXE_SHA
        pe = struct.unpack_from('<I', self.blob, 0x3c)[0]
        assert self.blob[:2] == b'MZ' and self.blob[pe:pe+4] == b'PE\0\0'
        assert struct.unpack_from('<H', self.blob, pe+4)[0] == 0x14c
        opt = pe+24
        assert struct.unpack_from('<H', self.blob, opt)[0] == 0x10b
        self.base = struct.unpack_from('<I', self.blob, opt+28)[0]
        assert self.base == 0x00400000
        self.size = (struct.unpack_from('<I', self.blob, opt+56)[0]+4095) & ~4095
        header_size = struct.unpack_from('<I', self.blob, opt+60)[0]
        table = opt+struct.unpack_from('<H', self.blob, pe+20)[0]
        self.sections = []
        self.image = bytearray(self.size)
        self.image[:header_size] = self.blob[:header_size]
        for i in range(struct.unpack_from('<H', self.blob, pe+6)[0]):
            at = table+40*i
            vs, va, size, raw = struct.unpack_from('<IIII', self.blob, at+8)
            flags = struct.unpack_from('<I', self.blob, at+36)[0]
            assert va+size <= self.size and raw+size <= len(self.blob)
            self.sections.append(dict(address=self.base+va, virtual_size=vs, raw_size=size,
                                      raw_offset=raw, flags=flags))
            self.image[va:va+size] = self.blob[raw:raw+size]

    def original(self, address, size):
        for section in self.sections:
            start = section['address']
            if start <= address and address+size <= start+section['raw_size']:
                at = section['raw_offset']+address-start
                return self.blob[at:at+size]
        raise AssertionError(('not original raw-backed range', hx(address), size))

    def mutable(self, address, size):
        return any(s['address'] <= address and address+size <= s['address']+max(s['virtual_size'], s['raw_size'])
                   and s['flags'] & 0x80000000 and not s['flags'] & 0x20000000 for s in self.sections)


class BlobStore:
    """Complete snapshots, deduplicated and compressed without raw disk copies."""
    def __init__(self, root):
        self.root = root
        root.mkdir()
        self.items = {}

    def keep(self, data):
        data = bytes(data)
        digest = sha(data)
        if digest not in self.items:
            packed = gzip.compress(data, compresslevel=6, mtime=0)
            assert gzip.decompress(packed) == data
            name = digest+'.bin.gz'
            with (self.root/name).open('xb') as f:
                f.write(packed)
            self.items[digest] = dict(file='memory/'+name, bytes=len(data), sha256=digest,
                                      gzip_bytes=len(packed), gzip_sha256=sha(packed))
        return self.items[digest]


def prepare(seal_path):
    seal_path = Path(seal_path).resolve()
    seal = json.loads(seal_path.read_text())
    assert seal['schema'] == 'opennox-default-damage-parent-input-freeze-v1'
    assert seal['root_reviewed_before_execution'] is True
    assert seal['native_engine_exists'] is False
    source = bound(seal['producer'])
    assert source == record(__file__), 'seal must bind THIS producer source'
    services_record = bound(seal['services'])
    services = json.loads(Path(services_record['path']).read_text())
    assert services['schema'] == 'opennox-default-damage-parent-service-boundaries-v1'
    original = bound(seal['original_EXE'])
    assert original['sha256'] == EXE_SHA
    main = bound(seal['unchanged_main_manifest'])
    assert main['sha256'] == MAIN_SHA
    main_obj = json.loads(Path(main['path']).read_text())
    assert len(main_obj['functions']) == 3373 and len(main_obj['data']) == 713
    dedicated = bound(seal['dedicated_manifest'])
    dedicated_obj = json.loads(Path(dedicated['path']).read_text())
    assert dedicated_obj['file_sha256'] == EXE_SHA
    data_record = bound(seal['inputs'])
    packed = Path(data_record['path']).read_bytes()
    data = gzip.decompress(packed) if packed[:2] == b'\x1f\x8b' else packed
    inputs = json.loads(data)
    assert inputs['schema'] == 'opennox-default-damage-parent-independent-inputs-v1'
    cases = inputs['cases']
    assert len(cases) == seal['inputs']['count'] and len(cases) > 0
    assert len({row['name'] for row in cases}) == len(cases)
    pe = PE(original['path'])
    for row in services['actual_original_ranges']+[services['padding9']]:
        assert sha(pe.original(integer(row['address']), row['size'])) == row['sha256']
        # Explicit tuple membership, never number-of-name/alias assumptions.
        section = dedicated_obj['data'] if row is services['padding9'] else dedicated_obj['functions']
        assert any(integer(r['address']) == integer(row['address']) and r['size'] == row['size']
                   and r['sha256'] == row['sha256'] for r in section)
    assert sha(pe.original(PARENT, 1792)) == services['parent_union1792_sha256']
    for row in cases:
        validate_case(row, pe, services)
    proof = dict(seal=record(seal_path), producer=source, services=services_record,
                 original_EXE=original, unchanged_main_manifest=main, dedicated_manifest=dedicated,
                 inputs=data_record, inputs_uncompressed_sha256=sha(data), cases=len(cases),
                 independent_domain_description=inputs.get('domain_description'),
                 original_execution_performed=False, native_engine_used=False)
    return seal, services, cases, pe, proof


def validate_case(row, pe, services):
    assert isinstance(row['name'], str) and row['name']
    assert len(row['args']) == 5
    for value in row['args']:
        word(value)
    assert set(row) <= {'name', 'args', 'roles', 'regions', 'image_initializers', 'services',
                        'after_reads', 'after_services', 'domain', 'notes', 'family', 'seed'}
    for value in row.get('roles', {}).values():
        word(value)
    ranges = []
    for region in row['regions']:
        address, size = word(region['address']), integer(region['size'])
        assert 0x21000000 <= address < address+size <= STACK
        assert address % 4096 == size % 4096 == 0 and size > 0
        assert region['permissions'] in ('r', 'rw', 'none')
        assert 0 <= region.get('fill', 0xa7) <= 255
        assert not any(max(a, address) < min(b, address+size) for a, b in ranges)
        ranges.append((address, address+size))
        for init in region.get('initializers', []):
            offset, data = integer(init['offset']), bytes.fromhex(init['hex'])
            assert 0 <= offset <= offset+len(data) <= size
    assert sum(b-a for a, b in ranges) <= MAX_GRAPH_BYTES
    for init in row.get('image_initializers', []):
        assert pe.mutable(word(init['address']), len(bytes.fromhex(init['hex'])))
    external = {integer(site): service for service in services['external_direct_services']
                for site in service['sites']}
    external.update({integer(service['site']): service for service in services['indirect_calls']})
    for site, schedule in row['services'].items():
        site = word(site)
        assert site in external
        responses = schedule if isinstance(schedule, list) else schedule['responses']
        assert responses
        if isinstance(schedule, dict):
            assert set(schedule) <= {'responses', 'repeat_last'}
            assert isinstance(schedule.get('repeat_last', False), bool)
        for response in responses:
            assert set(response) <= {'eax', 'ecx', 'edx', 'x87', 'effects'}
            for key in ('eax', 'ecx', 'edx'):
                if key in response:
                    word(response[key])
            assert 'eax' in response
            assert ('x87' in response) == (external[site]['return_kind'] == 'x87')
            if 'x87' in response:
                x87 = response['x87']
                assert len(bytes.fromhex(x87['hex'])) == {'f32': 4, 'f64': 8, 'ext80': 10}[x87['format']]
    for stimulus in row.get('after_reads', []):
        word(stimulus['pc']); word(stimulus['address'])
        assert integer(stimulus['width']) > 0 and integer(stimulus['occurrence']) >= 1
        assert stimulus['effects']
    for stimulus in row.get('after_services', []):
        assert word(stimulus['site']) in external and integer(stimulus['occurrence']) >= 1
        assert stimulus['effects']


class Recorder:
    def __init__(self, pe, services, case, label, cw, blobs):
        # Importing this module and preparing inputs do not import or run Unicorn.
        import unicorn as uc
        from unicorn import x86_const as x
        from capstone import Cs, CS_ARCH_X86, CS_MODE_32, CS_OP_MEM
        self.ucmod, self.x = uc, x
        self.cpu = uc.Uc(uc.UC_ARCH_X86, uc.UC_MODE_32)
        self.hooks = []
        self.pe, self.spec, self.case, self.label, self.cw, self.blobs = pe, services, case, label, cw, blobs
        self.areas, self.shadow, self.before, self.events, self.frames, self.completed_frames = [], {}, {}, [], [], []
        self.read_counts, self.service_counts, self.coverage, self.entries = Counter(), Counter(), Counter(), Counter()
        self.used_read_stimuli, self.used_service_stimuli = set(), set()
        self.pending_accesses, self.pending_call, self.pending_return, self.current_pc = [], None, None, None
        self.pending_instruction = None
        self.phase, self.instructions, self.faults = 'initialization', 0, []
        self.current_frame, self.harness_error = None, None
        self.originals, self.decoded, self.direct, self.external = {}, {}, {}, {}
        dis = Cs(CS_ARCH_X86, CS_MODE_32)
        dis.detail = True
        for original in services['actual_original_ranges']:
            start, size = integer(original['address']), original['size']
            self.originals[start] = original
            for insn in dis.disasm(pe.original(start, size), start):
                memory_operands = [dict(width=op.size, base=insn.reg_name(op.mem.base) if op.mem.base else None,
                                        index=insn.reg_name(op.mem.index) if op.mem.index else None,
                                        scale=op.mem.scale, displacement=op.mem.disp,
                                        segment=insn.reg_name(op.mem.segment) if op.mem.segment else None,
                                        decoder_access=op.access)
                                   for op in insn.operands if op.type == CS_OP_MEM]
                self.decoded[insn.address] = dict(size=insn.size, encoding=insn.bytes.hex(),
                                                   mnemonic=insn.mnemonic, operands=insn.op_str,
                                                   memory_operands=memory_operands,
                                                   range_name=original['name'])
            assert sum(i['size'] for a, i in self.decoded.items() if start <= a < start+size) == size
        for call in services['direct_calls']:
            site = integer(call['site'])
            assert self.decoded[site]['encoding'] == call['encoding']
            self.direct[site] = integer(call['target'])
        for service in services['external_direct_services']:
            for site in service['sites']:
                self.external[integer(site)] = service
        for service in services['indirect_calls']:
            site = integer(service['site'])
            assert self.decoded[site]['encoding'] == service['encoding']
            self.external[site] = service
        self.regids = {n: getattr(x, 'UC_X86_REG_'+n.upper()) for n in
                       ('eax', 'ebx', 'ecx', 'edx', 'esi', 'edi', 'ebp', 'esp', 'eip', 'eflags')}
        self.saved = dict(ebx=0x12345678, esi=0x23456789, edi=0x3456789a, ebp=0x456789ab)
        self.map_area('image', pe.base, bytes(pe.image), uc.UC_PROT_ALL)
        for region in case['regions']:
            size, address = integer(region['size']), word(region['address'])
            data = bytearray([region.get('fill', 0xa7)])*size
            for init in region.get('initializers', []):
                offset, payload = integer(init['offset']), bytes.fromhex(init['hex'])
                data[offset:offset+len(payload)] = payload
            perms = {'r': uc.UC_PROT_READ, 'rw': uc.UC_PROT_READ|uc.UC_PROT_WRITE, 'none': uc.UC_PROT_NONE}[region['permissions']]
            self.map_area(region['name'], address, data, perms)
        self.map_area('stack', STACK, b'\x6d'*STACK_SIZE, uc.UC_PROT_READ|uc.UC_PROT_WRITE)
        self.map_area('framework-auxiliary', AUX, b'\x91'*AUX_SIZE, uc.UC_PROT_ALL)
        for init in case.get('image_initializers', []):
            self.host_write(word(init['address']), bytes.fromhex(init['hex']), 'input-image-initializer')
        self.host_write(CALLER_SP, struct.pack('<6I', STOP, *map(word, case['args'])), 'input-caller-words')
        self.caller_before = bytes(self.cpu.mem_read(CALLER_SP, 24))
        self.host_write(INIT, b'\xdb\xe3\xd9\x2d'+struct.pack('<I', CW_ADDRESS), 'framework-fninit-fldcw')
        self.host_write(CW_ADDRESS, struct.pack('<H', cw), 'framework-control-word')
        self.cpu.reg_write(x.UC_X86_REG_ESP, CALLER_SP)
        self.cpu.emu_start(INIT, INIT+8, count=3)
        assert self.cpu.reg_read(x.UC_X86_REG_FPCW) == cw
        for name, value in self.saved.items():
            self.cpu.reg_write(self.regids[name], value)
        for name in ('eax', 'ecx', 'edx'):
            self.cpu.reg_write(self.regids[name], 0)
        self.cpu.reg_write(x.UC_X86_REG_EFLAGS, 0x202)
        self.cpu.reg_write(x.UC_X86_REG_EIP, PARENT)
        self.before = self.snapshots()
        self.initial_registers = self.registers()
        self.frames = [dict(kind='actual-original', name='parent', entry=hx(PARENT), call_site=None,
                            stack_entry=hx(CALLER_SP), return_pc=hx(STOP), args=list(map(word, case['args'])),
                            registers_before=self.registers(), FPU_before=self.fpu())]
        self.phase = 'original'
        for kind, callback in ((uc.UC_HOOK_CODE, self.code), (uc.UC_HOOK_MEM_READ, self.read),
                               (uc.UC_HOOK_MEM_WRITE, self.write), (uc.UC_HOOK_MEM_INVALID, self.invalid)):
            self.hooks.append(self.cpu.hook_add(kind, callback))

    def close(self):
        # Installed Unicorn exposes hook_del and weakref.finalize -> uc_close,
        # not a public instance close(). Never double-close its private handle.
        errors, removed = [], 0
        for handle in self.hooks:
            try:
                self.cpu.hook_del(handle)
                removed += 1
            except Exception as exc:
                errors.append(dict(type=type(exc).__name__, message=str(exc)))
        self.hooks.clear()
        self.cpu = None
        self.shadow.clear()
        collected = gc.collect()
        return dict(registered_hooks=4, removed_hooks=removed, errors=errors,
                    native_engine_reference_released=True, gc_collect_return=collected,
                    release_contract='public hook_del; drop Uc reference; installed weakref.finalize uc_close')

    def map_area(self, name, address, data, perms):
        assert name not in {a['name'] for a in self.areas}
        assert not any(max(a['address'], address) < min(a['address']+a['size'], address+len(data)) for a in self.areas)
        self.cpu.mem_map(address, len(data), perms)
        self.cpu.mem_write(address, bytes(data))
        self.areas.append(dict(name=name, address=address, size=len(data), permissions=perms))
        self.shadow[address] = bytearray(data)

    def area(self, address, size):
        for area in self.areas:
            if area['address'] <= address and address+size <= area['address']+area['size']:
                return area
        return None

    def event(self, kind, **fields):
        event = dict(sequence=len(self.events), kind=kind, **fields)
        self.events.append(event)
        return event

    def host_write(self, address, data, reason, **extra):
        area = self.area(address, len(data))
        assert area, ('input/service write outside declared mapped areas', hx(address), len(data))
        if area['name'] == 'image':
            assert self.pe.mutable(address, len(data)), 'service/input may not alter original executable/readonly image'
        if self.phase == 'original':
            assert area['permissions'] & self.ucmod.UC_PROT_WRITE, 'declared external effect needs writable destination'
        before = bytes(self.cpu.mem_read(address, len(data)))
        self.cpu.mem_write(address, bytes(data))
        offset = address-area['address']
        self.shadow[area['address']][offset:offset+len(data)] = data
        self.event('host-write', reason=reason, address=hx(address), width=len(data), before=before.hex(),
                   after=bytes(data).hex(), **extra)

    def registers(self):
        return {name: self.cpu.reg_read(reg) for name, reg in self.regids.items()}

    def fpu(self):
        # Unicorn FP0..FP7 expose integer mantissa/exponent, no host-float conversion.
        x = self.x
        return dict(control=self.cpu.reg_read(x.UC_X86_REG_FPCW), status=self.cpu.reg_read(x.UC_X86_REG_FPSW),
                    tag=self.cpu.reg_read(x.UC_X86_REG_FPTAG),
                    raw_physical80=[list(self.cpu.reg_read(x.UC_X86_REG_FP0+i)) for i in range(8)])

    def get32(self, address):
        return struct.unpack('<I', self.cpu.mem_read(address, 4))[0]

    def snapshots(self):
        return [{**area, 'address': hx(area['address']), 'snapshot': self.blobs.keep(self.cpu.mem_read(area['address'], area['size']))}
                for area in self.areas]

    def effects(self, effects, reason, args=(), **extra):
        for effect in effects:
            selectors = [key for key in ('address', 'arg_index', 'role') if key in effect]
            assert len(selectors) == 1
            selector = selectors[0]
            if selector == 'address':
                address = word(effect['address'])
            elif selector == 'arg_index':
                address = args[integer(effect['arg_index'])]
            else:
                address = word(self.case['roles'][effect['role']])
            address += integer(effect.get('offset', 0))
            assert 0 <= address <= 0xffffffff
            self.host_write(address, bytes.fromhex(effect['hex']), reason, **extra)

    def settle(self, completed=True):
        if self.pending_instruction is not None:
            self.pending_instruction['completed'] = completed
            self.pending_instruction = None
        pending, self.pending_accesses = self.pending_accesses, []
        triggered = []
        for event in pending:
            if event.get('fault'):
                continue
            event['completed'] = completed
            if not completed:
                continue
            if event['kind'] == 'write':
                address, width = integer(event['address']), event['width']
                after = bytes(self.cpu.mem_read(address, width))
                event['after'] = after.hex()
                area = self.area(address, width)
                assert area, event
                self.shadow[area['address']][address-area['address']:address-area['address']+width] = after
            elif event['kind'] == 'read' and event['source'] == 'actual-original':
                for index, stimulus in enumerate(self.case.get('after_reads', [])):
                    if (word(stimulus['pc']), word(stimulus['address']), integer(stimulus['width']), integer(stimulus['occurrence'])) == (
                            integer(event['pc']), integer(event['address']), event['width'], event['occurrence']):
                        assert index not in self.used_read_stimuli
                        self.used_read_stimuli.add(index)
                        triggered.append((index, stimulus, event['sequence']))
        # At the next instruction, original read result is already cached in CPU.
        for index, stimulus, read_sequence in triggered:
            self.effects(stimulus['effects'], 'after-completed-original-read', stimulus_index=index, read_sequence=read_sequence)

    def read(self, cpu, access, address, width, value, user):
        source = 'actual-original' if self.current_pc in self.decoded else 'external-x87-trampoline'
        key = (self.current_pc, address, width)
        self.read_counts[key] += 1
        try:
            data = bytes(cpu.mem_read(address, width)).hex()
        except self.ucmod.UcError:
            data = None
        event = self.event('read', pc=hx(self.current_pc), address=hx(address), width=width, value_hex=data,
                           source=source, occurrence=self.read_counts[key], completed=False)
        self.pending_accesses.append(event)

    def write(self, cpu, access, address, width, value, user):
        source = 'actual-original' if self.current_pc in self.decoded else 'external-x87-trampoline'
        assert source == 'actual-original', 'service trampoline may only FLD/RET'
        area = self.area(address, width)
        if area and area['name'] == 'image':
            assert self.pe.mutable(address, width), 'original unexpectedly writes executable/readonly image'
        try:
            before = bytes(cpu.mem_read(address, width)).hex()
        except self.ucmod.UcError:
            before = None
        event = self.event('write', pc=hx(self.current_pc), address=hx(address), width=width,
                           hook_value_low64=value & 0xffffffffffffffff, before=before, source=source, completed=False)
        # A cross-page write may physically commit a mapped prefix in Unicorn
        # before faulting. Capture every mapped intersection BEFORE it executes;
        # an unmapped suffix is not read, initialized, padded or accepted as data.
        spans = []
        for mapped in self.areas:
            start, end = max(address, mapped['address']), min(address+width, mapped['address']+mapped['size'])
            if start < end:
                spans.append(dict(area=mapped['name'], address=hx(start), width=end-start,
                                  write_offset=start-address, permissions=mapped['permissions'],
                                  before_hex=bytes(cpu.mem_read(start, end-start)).hex()))
        event['mapped_spans_before'] = sorted(spans, key=lambda row: integer(row['address']))
        event['registers_at_write'] = self.registers()
        self.pending_accesses.append(event)

    def observe_fault_write_fragments(self):
        """Observe bounded physical write fragments; do not complete the instruction.

        This is an emulator memory certificate, not a native fault-atomicity rule.
        Every changed byte must lie in a statically decoded operand extent and
        equal its corresponding original pre-write hook value byte.
        """
        for write in self.pending_accesses:
            if write['kind'] != 'write' or write['completed']:
                continue
            width, address, pc = write['width'], integer(write['address']), integer(write['pc'])
            assert 1 <= width <= 8, ('unsupported incomplete-write width', write)
            assert any(f['access'] in (self.ucmod.UC_MEM_WRITE_UNMAPPED, self.ucmod.UC_MEM_WRITE_PROT)
                       and integer(f['pc']) == pc
                       and max(integer(f['address']), address) < min(integer(f['address'])+f['width'], address+width)
                       for f in self.faults), ('no matching original store fault', write)
            insn = self.decoded[pc]
            registers = write['registers_at_write']
            matches = []
            for operand in insn['memory_operands']:
                if operand['width'] != width or operand['segment'] is not None:
                    continue
                base = 0 if operand['base'] is None else registers[operand['base']]
                index = 0 if operand['index'] is None else registers[operand['index']]
                effective = (base+index*operand['scale']+operand['displacement']) & 0xffffffff
                if effective == address:
                    matches.append(operand)
            assert len(matches) == 1, ('incomplete-write extent lacks unique decoded memory operand', write, insn)
            desired = (write['hook_value_low64'] & ((1 << (8*width))-1)).to_bytes(width, 'little')
            observed, fragments = [], []
            for span in write['mapped_spans_before']:
                start, count, offset = integer(span['address']), span['width'], span['write_offset']
                before = bytes.fromhex(span['before_hex'])
                after = bytes(self.cpu.mem_read(start, count))
                mapped = self.area(start, count)
                assert mapped and len(before) == count and 0 <= offset <= offset+count <= width
                shadow = self.shadow[mapped['address']]
                at = start-mapped['address']
                assert bytes(shadow[at:at+count]) == before, 'untracked predecessor write before faulting store'
                changed = []
                for i, (old, actual) in enumerate(zip(before, after)):
                    if old != actual:
                        assert actual == desired[offset+i], ('fault fragment differs from original hook byte', write, span, i)
                        changed.append(i)
                observed.append(dict(**span, after_hex=after.hex(), changed_byte_offsets=changed))
                # Only observable changed byte runs are applied. Equal old/new
                # bytes are not credited as evidence of a committed store.
                cursor = 0
                while cursor < len(changed):
                    first = last = changed[cursor]
                    cursor += 1
                    while cursor < len(changed) and changed[cursor] == last+1:
                        last = changed[cursor]
                        cursor += 1
                    data = after[first:last+1]
                    shadow[at+first:at+last+1] = data
                    fragment = self.event('observed-partial-write-fragment', original_write_sequence=write['sequence'],
                                          pc=write['pc'], address=hx(start+first), width=len(data),
                                          write_offset=offset+first, before=before[first:last+1].hex(), after=data.hex(),
                                          original_hook_value_slice=desired[offset+first:offset+last+1].hex(),
                                          original_instruction_completed=False,
                                          emulator_physical_observation_only=True)
                    fragments.append(fragment['sequence'])
            write['faulting_write_observation'] = dict(decoded_memory_operand=matches[0],
                                                      mapped_spans_after=observed,
                                                      changed_fragment_sequences=fragments,
                                                      original_instruction_completed=False,
                                                      emulator_physical_observation_only=True)
            self.event('faulting-write-observed', original_write_sequence=write['sequence'], pc=write['pc'],
                       width=width, address=write['address'], changed_fragment_sequences=fragments,
                       original_instruction_completed=False, emulator_physical_observation_only=True)

    def invalid(self, cpu, access, address, width, value, user):
        # A failed instruction fetch happens AFTER the preceding CALL completed.
        # Its return-PC stack write and indirect operand read must remain complete.
        if access in (self.ucmod.UC_MEM_FETCH_UNMAPPED, self.ucmod.UC_MEM_FETCH_PROT):
            self.settle()
        event = self.event('fault', pc=hx(cpu.reg_read(self.x.UC_X86_REG_EIP)), access=access,
                           address=hx(address), width=width, value_low64=value & 0xffffffffffffffff,
                           registers=self.registers(), FPU=self.fpu())
        self.faults.append(event)
        if self.pending_instruction is not None:
            self.pending_instruction['fault'] = True
        for pending in self.pending_accesses:
            if max(integer(pending['address']), address) < min(integer(pending['address'])+pending['width'], address+width):
                pending['fault'] = True
        return False

    def finish_frame(self, frame, pc):
        after = self.registers()
        frame.update(returned_to=hx(pc), registers_after=after, FPU_after=self.fpu(),
                     stack_exit=hx(after['esp']), completed=True,
                     callee_saved_preserved=all(after[n] == frame['registers_before'][n] for n in self.saved))
        assert pc == integer(frame['return_pc']) and after['esp'] == integer(frame['stack_entry'])+4
        assert frame['callee_saved_preserved']
        self.completed_frames.append(frame)
        self.event('return-completed', frame=json.loads(json.dumps(frame)))
        if frame['kind'] == 'external-input-boundary':
            site = integer(frame['call_site'])
            occurrence = frame['service_occurrence']
            for index, stimulus in enumerate(self.case.get('after_services', [])):
                if word(stimulus['site']) == site and integer(stimulus['occurrence']) == occurrence:
                    assert index not in self.used_service_stimuli
                    self.used_service_stimuli.add(index)
                    self.effects(stimulus['effects'], 'after-completed-external-service', frame['args'],
                                 stimulus_index=index, service_site=hx(site))

    def enter_call(self, address):
        call, self.pending_call = self.pending_call, None
        assert address == call['target'], (call, hx(address))
        sp = self.cpu.reg_read(self.x.UC_X86_REG_ESP)
        assert sp == call['stack_before']-4 and self.get32(sp) == call['return_pc']
        spec = self.external.get(call['site'])
        if spec:
            if call['site'] in CALLBACKS:
                assert address == CALLBACKS[call['site']], ('undeclared executable callback', call)
            argc, kind, name = spec['argument_words'], 'external-input-boundary', spec['name']
        else:
            original = self.originals[address]
            argc, kind, name = {0x4e1400: 2, 0x4e1470: 1, 0x419a70: 1}[address], 'actual-original', original['name']
        args = [self.get32(sp+4+4*i) for i in range(argc)]
        frame = dict(kind=kind, name=name, call_site=hx(call['site']), entry=hx(address),
                     stack_entry=hx(sp), return_pc=hx(call['return_pc']), args=args,
                     registers_before=self.registers(), FPU_before=self.fpu(), completed=False)
        self.frames.append(frame)
        self.event('call-entered', frame=json.loads(json.dumps(frame)))
        return spec, frame

    def service(self, spec, frame):
        site = integer(frame['call_site'])
        self.service_counts[site] += 1
        occurrence = self.service_counts[site]
        frame['service_occurrence'] = occurrence
        schedules = {word(k): v for k, v in self.case['services'].items()}
        assert site in schedules, ('undeclared reached service input', hx(site), self.case['name'])
        schedule = schedules[site]
        responses = schedule if isinstance(schedule, list) else schedule['responses']
        repeat = False if isinstance(schedule, list) else schedule.get('repeat_last', False)
        selected = occurrence-1
        if selected >= len(responses):
            assert repeat, ('service input schedule exhausted', hx(site), occurrence)
            selected = len(responses)-1
        response = responses[selected]
        self.event('external-service-input', site=hx(site), occurrence=occurrence, response_index=selected,
                   repeated=occurrence > len(responses), name=spec['name'], args=frame['args'],
                   input_response=response, original_helper_executed=False)
        self.effects(response.get('effects', []), 'external-service-input-effect', frame['args'], service_site=hx(site))
        for name in ('eax', 'ecx', 'edx'):
            if name in response:
                self.cpu.reg_write(self.regids[name], word(response[name]))
        if spec['return_kind'] == 'x87':
            x87 = response['x87']
            payload = bytes.fromhex(x87['hex'])
            self.host_write(PAYLOAD, payload, 'framework-exact-x87-service-payload', service_site=hx(site))
            opcode = {'f32': b'\xd9\x05', 'f64': b'\xdd\x05', 'ext80': b'\xdb\x2d'}[x87['format']]
            code = opcode+struct.pack('<I', PAYLOAD)+b'\xc3'
            self.host_write(TRAMPOLINE, code, 'framework-exact-x87-fld-ret', service_site=hx(site))
            frame['x87_return_input'] = x87
            self.cpu.reg_write(self.regids['eip'], TRAMPOLINE)
        else:
            # No guest helper instructions execute; model only declared cdecl RET.
            self.cpu.reg_write(self.regids['esp'], integer(frame['stack_entry'])+4)
            self.cpu.reg_write(self.regids['eip'], integer(frame['return_pc']))
            self.pending_return = self.frames.pop()

    def code(self, cpu, address, size, user):
        self.settle()
        if self.pending_return is not None:
            frame, self.pending_return = self.pending_return, None
            self.finish_frame(frame, address)
        if address == STOP:
            assert not self.frames and self.pending_call is None
            cpu.emu_stop()
            return
        self.current_pc = address
        self.instructions += 1
        if self.pending_call is not None:
            spec, frame = self.enter_call(address)
            if spec:
                self.service(spec, frame)
                return
        if address in (TRAMPOLINE, TRAMPOLINE+6):
            assert self.frames[-1]['kind'] == 'external-input-boundary' and 'x87_return_input' in self.frames[-1]
            self.pending_instruction = self.event('external-x87-trampoline-instruction', pc=hx(address), size=size,
                                                  encoding=bytes(cpu.mem_read(address, size)).hex(),
                                                  completed=False, original_execution_credit=False)
            if address == TRAMPOLINE+6:
                self.pending_return = self.frames.pop()
            return
        assert address in self.decoded, ('undeclared instruction', hx(address), self.case['name'])
        insn = self.decoded[address]
        assert size == insn['size'] and bytes(cpu.mem_read(address, size)).hex() == insn['encoding']
        if address in self.originals:
            self.entries[address] += 1
        self.coverage[address] += 1
        self.pending_instruction = self.event('original-instruction', pc=hx(address), size=size,
                                               encoding=insn['encoding'], completed=False,
                                               range_name=insn['range_name'], frame_depth=len(self.frames))
        if insn['mnemonic'] == 'call':
            if address in self.direct:
                target = self.direct[address]
            else:
                assert address in CALLBACKS
                if address == 0x004e0d05:
                    try:
                        target = self.get32(cpu.reg_read(self.regids['edi'])+716)
                    except self.ucmod.UcError:
                        # The actual CALL must perform its own failing operand read.
                        target = None
                else:
                    target = cpu.reg_read(self.regids['eax'])
            self.pending_call = dict(site=address, target=target, return_pc=address+size,
                                     stack_before=cpu.reg_read(self.regids['esp']))
            argc = self.external[address]['argument_words'] if address in self.external else {0x4e1400: 2, 0x4e1470: 1, 0x419a70: 1}[target]
            self.event('call-issued', site=hx(address), target=None if target is None else hx(target), argument_words=argc,
                       args=[self.get32(self.pending_call['stack_before']+4*i) for i in range(argc)],
                       stack_before=hx(self.pending_call['stack_before']), registers=self.registers(), FPU=self.fpu())
        elif insn['mnemonic'] == 'ret':
            self.pending_return = self.frames.pop()

    def run(self):
        emulation_error = None
        try:
            # STOP hook settles final RET stack read; until is deliberately after STOP.
            self.cpu.emu_start(PARENT, STOP+1, count=MAX_INSTRUCTIONS)
            if not self.faults:
                assert self.cpu.reg_read(self.regids['eip']) == STOP, 'instruction bound / unfinished original'
                assert not self.frames and self.pending_return is None and self.pending_call is None
        except Exception as exc:
            emulation_error = dict(type=type(exc).__name__, message=str(exc), traceback=traceback.format_exc())
            if not (isinstance(exc, self.ucmod.UcError) and self.faults):
                self.harness_error = emulation_error
            if self.faults:
                try:
                    self.observe_fault_write_fragments()
                except Exception as fragment_error:
                    self.harness_error = dict(type=type(fragment_error).__name__, message=str(fragment_error),
                                              traceback=traceback.format_exc(), phase='fault-write-fragment-certificate')
            self.settle(completed=False)
        after = self.snapshots()
        memory_exact = []
        for area in self.areas:
            actual = bytes(self.cpu.mem_read(area['address'], area['size']))
            shadow = bytes(self.shadow[area['address']])
            memory_exact.append(dict(name=area['name'], bytes=area['size'],
                                     reconstructed_from_completed_writes_and_observed_fault_fragments=actual == shadow,
                                     actual_sha256=sha(actual), reconstructed_sha256=sha(shadow)))
            if actual != shadow:
                self.harness_error = self.harness_error or dict(type='MemoryGuardMismatch', message=area['name'])
        original_exact = all(bytes(self.cpu.mem_read(start, spec['size'])) == self.pe.original(start, spec['size'])
                             for start, spec in self.originals.items())
        if not original_exact:
            self.harness_error = self.harness_error or dict(type='OriginalBytesChanged')
        regs = self.registers()
        successful = not self.faults and not self.harness_error
        if successful:
            assert regs['esp'] == CALLER_SP+4 and all(regs[n] == value for n, value in self.saved.items())
            assert self.cpu.reg_read(self.x.UC_X86_REG_FPCW) == self.cw
        initial_events = [e for e in self.events if e['kind'] == 'host-write' and e['reason'].startswith(('input-', 'framework-fninit', 'framework-control'))]
        completed_coverage = Counter(e['pc'] for e in self.events if e['kind'] == 'original-instruction' and e['completed'])
        return dict(name=self.case['name'], input_sha256=sha(canonical(self.case)), input=self.case,
                    mode=self.label, control_word=self.cw,
                    status='harness-error' if self.harness_error else ('original-fault' if self.faults else 'returned'),
                    instructions=self.instructions, instruction_coverage={hx(k): v for k, v in sorted(self.coverage.items())},
                    instruction_coverage_semantics='code-hook entry, including a faulting instruction; completed map separately',
                    completed_original_instruction_coverage=dict(sorted(completed_coverage.items())),
                    actual_original_entry_counts={hx(k): v for k, v in sorted(self.entries.items())},
                    registers_before=self.initial_registers, registers_after=regs, FPU_after=self.fpu(),
                    caller_before_hex=self.caller_before.hex(), caller_after_hex=bytes(self.cpu.mem_read(CALLER_SP, 24)).hex(),
                    # Caller source/damage slots are retained verbatim, never blanked/masked.
                    complete_memory_before=self.before, complete_memory_after=after,
                    all_memory_write_reconstruction=memory_exact, original_executable_bytes_unchanged=original_exact,
                    complete_events=self.events, completed_call_frames=self.completed_frames, open_call_frames=self.frames,
                    pending_call_at_terminal=self.pending_call, pending_return_at_terminal=self.pending_return,
                    service_call_counts={hx(k): v for k, v in sorted(self.service_counts.items())},
                    used_after_read_stimuli=sorted(self.used_read_stimuli),
                    unused_after_read_stimuli=sorted(set(range(len(self.case.get('after_reads', []))))-self.used_read_stimuli),
                    used_after_service_stimuli=sorted(self.used_service_stimuli),
                    unused_after_service_stimuli=sorted(set(range(len(self.case.get('after_services', []))))-self.used_service_stimuli),
                    faults=self.faults, emulation_error=emulation_error, harness_error=self.harness_error,
                    observed_partial_write_fragments=sum(e['kind'] == 'observed-partial-write-fragment' for e in self.events),
                    faulting_write_observations=sum(e['kind'] == 'faulting-write-observed' for e in self.events),
                    fault_write_policy='observed mapped fragments <=8B; decoded operand extent; hook-value byte exact; instruction incomplete',
                    original_execution_scope='parent1783 + child112/34 + converter17; all other helpers external',
                    initialized_image_baseline_after_input_writes=True,
                    no_host_float_computation=True, no_native_algorithm=True,
                    original_world_or_HP_helper_execution_credit=False,
                    input_initialization_events_before_snapshot=len(initial_events))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--input-seal', required=True, type=Path)
    parser.add_argument('--out', required=True, type=Path)
    mode = parser.add_mutually_exclusive_group(required=True)
    mode.add_argument('--execute-reviewed-original', action='store_true')
    mode.add_argument('--validate-inputs-only', action='store_true')
    args = parser.parse_args()
    out = args.out.resolve()
    out.mkdir(exist_ok=False)
    try:
        seal, services, cases, pe, preparation = prepare(args.input_seal)
    except Exception as exc:
        write_json(out/'static-preexecution-failure-v2.json',
                   dict(original_calls=0, exception_type=type(exc).__name__, message=str(exc),
                        traceback=traceback.format_exc(), producer=record(__file__),
                        requested_input_seal=str(args.input_seal.resolve())))
        raise
    write_json(out/'input-and-source-binding-v2.json', preparation)
    if args.validate_inputs_only:
        write_json(out/'static-input-validation-only-v2.json', dict(**preparation, original_calls=0))
        print(json.dumps(dict(status='inputs-validated-only', cases=len(cases), original_calls=0)))
        return 0
    blobs = BlobStore(out/'memory')
    counts, coverage, completed_coverage, entered_originals = Counter(), Counter(), Counter(), Counter()
    physical_partial_fragments = faulting_write_observations = 0
    rows_path = out/'actual-original-parent-rows-v2.jsonl.gz'
    fatal = None
    with rows_path.open('xb') as raw:
        with gzip.GzipFile(filename='', fileobj=raw, mode='wb', compresslevel=6, mtime=0) as stream:
            for case in cases:
                for label, cw in MODES:
                    recorder = None
                    try:
                        recorder = Recorder(pe, services, case, label, cw, blobs)
                        row = recorder.run()
                    except Exception as exc:
                        row = dict(name=case['name'], mode=label, input=case,
                                   status='harness-initialization-error' if recorder is None else 'harness-finalization-error',
                                   error=dict(type=type(exc).__name__, message=str(exc), traceback=traceback.format_exc()))
                        if recorder is not None:
                            row.update(complete_events=recorder.events, completed_call_frames=recorder.completed_frames,
                                       open_call_frames=recorder.frames, faults=recorder.faults,
                                       complete_memory_before=recorder.before, complete_memory_after=recorder.snapshots(),
                                       registers_after=recorder.registers(), FPU_after=recorder.fpu(),
                                       instruction_coverage={hx(k): v for k, v in recorder.coverage.items()},
                                       actual_original_entry_counts={hx(k): v for k, v in recorder.entries.items()})
                    finally:
                        if recorder is not None:
                            cleanup = recorder.close()
                        else:
                            cleanup = dict(initialization_failed=True, gc_collect_return=gc.collect())
                    row['recorder_cleanup'] = cleanup
                    if cleanup.get('errors'):
                        row['status'] = 'harness-cleanup-error'
                    stream.write(canonical(row)+b'\n')
                    stream.flush()
                    counts[row['status']] += 1
                    coverage.update({integer(k): v for k, v in row.get('instruction_coverage', {}).items()})
                    completed_coverage.update({integer(k): v for k, v in row.get('completed_original_instruction_coverage', {}).items()})
                    entered_originals.update({integer(k): v for k, v in row.get('actual_original_entry_counts', {}).items()})
                    physical_partial_fragments += row.get('observed_partial_write_fragments', 0)
                    faulting_write_observations += row.get('faulting_write_observations', 0)
                    if row['status'].startswith('harness-'):
                        fatal = dict(name=case['name'], mode=label, error=row.get('harness_error', row.get('error')))
                        break
                if fatal:
                    break
    post_bindings = {}
    try:
        for key in ('producer', 'services', 'original_EXE', 'unchanged_main_manifest', 'dedicated_manifest', 'inputs'):
            post_bindings[key] = bound(seal[key])
        post_bindings['input_seal'] = record(args.input_seal)
        assert post_bindings['input_seal'] == preparation['seal']
    except Exception as exc:
        fatal = fatal or dict(phase='postexecution-source-input-binding', type=type(exc).__name__,
                              message=str(exc), traceback=traceback.format_exc())
    from capstone import Cs, CS_ARCH_X86, CS_MODE_32
    static_boundaries = {}
    dis = Cs(CS_ARCH_X86, CS_MODE_32)
    for original in services['actual_original_ranges']:
        address = integer(original['address'])
        static_boundaries[original['name']] = [dict(pc=hx(i.address), bytes=i.size, encoding=i.bytes.hex())
                                              for i in dis.disasm(pe.original(address, original['size']), address)]
    missing = {name: [i['pc'] for i in instructions if not coverage[integer(i['pc'])]]
               for name, instructions in static_boundaries.items()}
    not_completed = {name: [i['pc'] for i in instructions if not completed_coverage[integer(i['pc'])]]
                     for name, instructions in static_boundaries.items()}
    proof = dict(schema='opennox-default-damage-parent-actual-original-recorder-proof-v2', preparation=preparation,
                 modes=[dict(label=label, cw=hx(cw)) for label, cw in MODES],
                 requested_case_count=len(cases), requested_original_root_calls=2*len(cases),
                 actual_rows=sum(counts.values()), statuses=dict(counts), fatal=fatal,
                 actual_entered_originals={hx(k): v for k, v in sorted(entered_originals.items())},
                 emulator_observed_partial_write_fragments=physical_partial_fragments,
                 faulting_write_observations=faulting_write_observations,
                 native_fault_atomicity_qualified=False,
                 original_instruction_coverage_union={hx(k): v for k, v in sorted(coverage.items())},
                 original_completed_instruction_coverage_union={hx(k): v for k, v in sorted(completed_coverage.items())},
                 original_static_instruction_boundaries=static_boundaries,
                 unentered_original_instruction_boundaries=missing,
                 never_completed_original_instruction_boundaries=not_completed,
                 rows=record(rows_path), memory_snapshots=list(blobs.items.values()),
                 actual_original_ranges=services['actual_original_ranges'],
                 external_services=services['counts'], postexecution_source_input_bindings=post_bindings,
                 immutable_original_EXE=post_bindings.get('original_EXE') == preparation['original_EXE'],
                 original_converter_actual=True, actual_child_predicates_required=True,
                 no_native_engine_used=True, original_service_bodies_excluded=True,
                 original_faults_are_observations_not_automatically_expected=True,
                 no_input_or_expected_output_computation_from_native_code=True)
    write_json(out/'actual-original-parent-recorder-proof-v2.json', proof)
    print(json.dumps(dict(status='recorded' if not fatal else 'harness-failure',
                          actual_rows=sum(counts.values()), statuses=dict(counts), proof=str(out/'actual-original-parent-recorder-proof-v2.json'))))
    return 1 if fatal else 0


if __name__ == '__main__':
    sys.exit(main())
