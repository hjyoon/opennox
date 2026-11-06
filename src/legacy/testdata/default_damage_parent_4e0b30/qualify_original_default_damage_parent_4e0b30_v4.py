#!/usr/bin/env python3
"""Independent streaming certificate qualifier; never imports/runs the producer.

Observed original faults remain observations, with no native expected-outcome
algorithm. Full compressed memory bytes are replayed independently. No Go,
game, original CPU emulation, network, raw JSON dump or canonical-file mutation.
"""
import argparse
from collections import Counter, OrderedDict
import gzip
import hashlib
import json
from pathlib import Path
import re
import struct
import sys
import traceback

EXE_SHA = '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
MAIN_SHA = 'cbe217010a1bffd6dd75299ac02629d7bda88b396ea7d52450bd8524cf972afd'
PRODUCER_SHA = '394c1aa35958d9ebcc1ed76d790721bbbea0308e6c26cb52daa676c603098aa4'
SERVICE_SHA = 'e481d4bf3fa671436ff3d2b2b9c056b8250a202bffcee7b8de4290474a261887'
ROOT, STACK, STACK_SIZE, SP = 0x4e0b30, 0x22000000, 0x10000, 0x2200e000
AUX, AUX_SIZE, STOP = 0x23000000, 0x10000, 0x23000100
INIT, CW_AT, PAYLOAD, TRAMP = 0x23000200, 0x23000300, 0x23000400, 0x23000500
MODES = {'startup-53-chop': 0x0e3f, 'nearest-53': 0x023f}
SAVED = dict(ebx=0x12345678, esi=0x23456789, edi=0x3456789a, ebp=0x456789ab)


def digest(data):
    return hashlib.sha256(data).hexdigest()


def canon(obj):
    return json.dumps(obj, sort_keys=True, separators=(',', ':')).encode()


def number(value):
    assert isinstance(value, (int, str)) and not isinstance(value, bool)
    if isinstance(value, str):
        return int(value, 16) if re.fullmatch('[0-9a-fA-F]{8}', value) else int(value, 0)
    return value


def hx(value):
    return f'0x{value:08x}'


def file_record(path):
    path = Path(path).resolve()
    size, hasher = 0, hashlib.sha256()
    with path.open('rb') as f:
        while chunk := f.read(1 << 20):
            size += len(chunk)
            hasher.update(chunk)
    return dict(path=str(path), bytes=size, sha256=hasher.hexdigest())


def verify_file(row):
    actual = file_record(row['path'])
    assert (actual['bytes'], actual['sha256']) == (row['bytes'], row['sha256']), actual
    return actual


def fpu(value, cw):
    assert value['control'] == cw
    assert 0 <= value['status'] <= 0xffff and 0 <= value['tag'] <= 0xffff
    assert len(value['raw_physical80']) == 8
    for mantissa, exponent in value['raw_physical80']:
        assert isinstance(mantissa, int) and 0 <= mantissa < 1 << 64
        assert isinstance(exponent, int) and 0 <= exponent < 1 << 16


def registers(value):
    assert set(value) == {'eax', 'ebx', 'ecx', 'edx', 'esi', 'edi', 'ebp', 'esp', 'eip', 'eflags'}
    assert all(isinstance(v, int) and 0 <= v <= 0xffffffff for v in value.values())


class OriginalPE:
    def __init__(self, path):
        raw = Path(path).read_bytes()
        assert len(raw) == 1929216 and digest(raw) == EXE_SHA
        pe, = struct.unpack_from('<I', raw, 60)
        assert raw[:2] == b'MZ' and raw[pe:pe+4] == b'PE\0\0'
        assert struct.unpack_from('<H', raw, pe+4)[0] == 0x14c
        optional = pe+24
        assert struct.unpack_from('<H', raw, optional)[0] == 0x10b
        self.base, = struct.unpack_from('<I', raw, optional+28)
        assert self.base == 0x400000
        size, headers = struct.unpack_from('<II', raw, optional+56)
        self.image = bytearray((size+4095)//4096*4096)
        self.image[:headers] = raw[:headers]
        self.sections = []
        table = optional+struct.unpack_from('<H', raw, pe+20)[0]
        for at in range(table, table+40*struct.unpack_from('<H', raw, pe+6)[0], 40):
            virtual_size, rva, count, offset = struct.unpack_from('<IIII', raw, at+8)
            flags, = struct.unpack_from('<I', raw, at+36)
            assert rva+count <= len(self.image) and offset+count <= len(raw)
            self.image[rva:rva+count] = raw[offset:offset+count]
            self.sections.append((self.base+rva, max(virtual_size, count), flags))

    def read(self, address, count):
        offset = address-self.base
        assert 0 <= offset <= offset+count <= len(self.image)
        return bytes(self.image[offset:offset+count])

    def writable_data(self, address, count):
        return any(a <= address and address+count <= a+n and f & 0x80000000 and not f & 0x20000000
                   for a, n, f in self.sections)


class Memory:
    def __init__(self, pe, case):
        self.areas = [('image', pe.base, bytearray(pe.image), 7)]
        for r in case['regions']:
            data = bytearray([r.get('fill', 0xa7)])*r['size']
            for init in r.get('initializers', []):
                at, payload = number(init['offset']), bytes.fromhex(init['hex'])
                assert 0 <= at <= at+len(payload) <= len(data)
                data[at:at+len(payload)] = payload
            self.areas.append((r['name'], number(r['address']), data, {'none': 0, 'r': 1, 'rw': 3}[r['permissions']]))
        self.areas.extend([('stack', STACK, bytearray(b'\x6d'*STACK_SIZE), 3),
                           ('framework-auxiliary', AUX, bytearray(b'\x91'*AUX_SIZE), 7)])
        assert len({a[0] for a in self.areas}) == len(self.areas)
        ordered = sorted(self.areas, key=lambda a: a[1])
        assert all(a[1]+len(a[2]) <= b[1] for a, b in zip(ordered, ordered[1:]))

    def spans(self, address, count):
        out = []
        for name, base, data, permissions in self.areas:
            start, end = max(address, base), min(address+count, base+len(data))
            if start < end:
                out.append(dict(area=name, address=hx(start), width=end-start, write_offset=start-address,
                                permissions=permissions, before_hex=bytes(data[start-base:end-base]).hex()))
        return sorted(out, key=lambda s: number(s['address']))

    def read(self, address, count):
        fragments = self.spans(address, count)
        if sum(f['width'] for f in fragments) != count:
            return None
        assert all(number(f['address']) == address+sum(g['width'] for g in fragments[:i]) for i, f in enumerate(fragments))
        return b''.join(bytes.fromhex(f['before_hex']) for f in fragments)

    def word(self, address):
        data = self.read(address, 4)
        assert data is not None
        return int.from_bytes(data, 'little')

    def write(self, address, payload):
        assert self.read(address, len(payload)) is not None, (hx(address), len(payload))
        for _, base, data, _ in self.areas:
            start, end = max(address, base), min(address+len(payload), base+len(data))
            if start < end:
                data[start-base:end-base] = payload[start-address:end-address]

    def invalid_bytes(self, address, count, permission):
        """Mapping/permission facts only, not an expected fault algorithm."""
        result = []
        for at in range(address, address+count):
            spans = self.spans(at, 1)
            if not spans or not spans[0]['permissions'] & permission:
                result.append(at)
        return result


class Snapshots:
    def __init__(self, runout, rows, cache_bytes):
        self.runout, self.maximum = runout, cache_bytes
        self.members, self.cache, self.cache_size, self.verified = {}, OrderedDict(), 0, {}
        for row in rows:
            name = row['file']
            assert re.fullmatch('memory/[0-9a-f]{64}\\.bin\\.gz', name)
            assert name == 'memory/'+row['sha256']+'.bin.gz' and name not in self.members
            assert 0 <= row['bytes'] <= 32 << 20
            self.members[name] = row
        physical_members = list((runout/'memory').iterdir())
        assert all(p.is_file() and not p.is_symlink() for p in physical_members)
        actual_names = {str(p.relative_to(runout)) for p in physical_members}
        assert actual_names == set(self.members), 'unlisted/missing canonical memory member'

    def get(self, ref):
        name = ref['file']
        assert self.members[name] == ref
        if name in self.cache:
            data = self.cache.pop(name)
            self.cache[name] = data
            return data
        path = self.runout/name
        assert not path.is_symlink()
        raw = file_record(path)
        assert (raw['bytes'], raw['sha256']) == (ref['gzip_bytes'], ref['gzip_sha256'])
        with gzip.open(path, 'rb') as f:
            data = f.read(ref['bytes']+1)
            assert not f.read(1)
        assert len(data) == ref['bytes'] and digest(data) == ref['sha256']
        self.verified[name] = raw
        while self.cache and self.cache_size+len(data) > self.maximum:
            _, old = self.cache.popitem(last=False)
            self.cache_size -= len(old)
        if len(data) <= self.maximum:
            self.cache[name] = data
            self.cache_size += len(data)
        return data

    def compare(self, memory, snapshots):
        assert len(memory.areas) == len(snapshots)
        for (name, address, data, permissions), snapshot in zip(memory.areas, snapshots):
            assert (snapshot['name'], number(snapshot['address']), snapshot['size'], snapshot['permissions']) == (name, address, len(data), permissions)
            assert self.get(snapshot['snapshot']) == bytes(data), ('whole-memory mismatch', name)


def resolve(effect, case, args):
    selectors = set(effect) & {'address', 'arg_index', 'role'}
    assert len(selectors) == 1
    if 'address' in effect:
        address = number(effect['address'])
    elif 'role' in effect:
        address = number(case['roles'][effect['role']])
    else:
        address = args[number(effect['arg_index'])]
    return address+number(effect.get('offset', 0)), bytes.fromhex(effect['hex'])


def instruction_table(pe, services):
    # Static disassembly only; never construct a Unicorn engine.
    import capstone
    from capstone import Cs, CS_ARCH_X86, CS_MODE_32, CS_OP_MEM
    decoder = Cs(CS_ARCH_X86, CS_MODE_32)
    decoder.detail = True
    table, groups = {}, {}
    for region in services['actual_original_ranges']:
        start, size = number(region['address']), region['size']
        raw = pe.read(start, size)
        assert digest(raw) == region['sha256']
        groups[region['name']] = []
        for ins in decoder.disasm(raw, start):
            operands = [dict(width=o.size, base=ins.reg_name(o.mem.base) if o.mem.base else None,
                             index=ins.reg_name(o.mem.index) if o.mem.index else None, scale=o.mem.scale,
                             displacement=o.mem.disp, segment=ins.reg_name(o.mem.segment) if o.mem.segment else None,
                             decoder_access=o.access) for o in ins.operands if o.type == CS_OP_MEM]
            table[ins.address] = dict(size=ins.size, encoding=ins.bytes.hex(), mnemonic=ins.mnemonic,
                                      name=region['name'], memory_operands=operands)
            groups[region['name']].append(dict(pc=hx(ins.address), bytes=ins.size, encoding=ins.bytes.hex()))
        assert sum(i['bytes'] for i in groups[region['name']]) == size
    return table, groups, capstone.__version__


class RowCheck:
    def __init__(self, pe, services, instructions, snapshots, case, row):
        self.pe, self.spec, self.instructions, self.snapshots, self.case, self.row = pe, services, instructions, snapshots, case, row
        self.cw = MODES[row['mode']]
        self.memory = Memory(pe, case)
        self.external = {number(site): service for service in services['external_direct_services'] for site in service['sites']}
        self.external.update({number(service['site']): service for service in services['indirect_calls']})
        self.direct = {number(call['site']): number(call['target']) for call in services['direct_calls']}
        self.call_sizes = {number(call['site']): len(bytes.fromhex(call['encoding'])) for call in services['direct_calls']+services['indirect_calls']}
        assert all(size == instructions[site]['size'] for site, size in self.call_sizes.items())
        self.pending_call, self.pending_return, self.current = None, None, None
        self.frames, self.returned, self.faults = [], [], []
        self.counts, self.entered, self.completed, self.entries = Counter(), Counter(), Counter(), Counter()
        self.reads, self.services = Counter(), Counter()
        self.read_used, self.service_used, self.expected_host, self.writes = set(), set(), [], {}
        self.partial_events, self.fault_write_observations = {}, []

    def host_expected(self, reason, effects, args=(), **extra):
        for effect in effects:
            address, payload = resolve(effect, self.case, args)
            self.expected_host.append(dict(reason=reason, address=address, payload=payload, **extra))

    def host(self, event, initial=False):
        assert self.expected_host, ('undeclared host write', event)
        expected = self.expected_host.pop(0)
        assert event['reason'] == expected['reason'] and number(event['address']) == expected['address']
        data = bytes.fromhex(event['after'])
        assert data == expected['payload'] and len(data) == event['width']
        assert self.memory.read(expected['address'], len(data)) == bytes.fromhex(event['before'])
        for key in ('service_site', 'stimulus_index', 'read_sequence'):
            if key in expected:
                assert event[key] == expected[key]
        if self.pe.base <= expected['address'] < self.pe.base+len(self.pe.image):
            assert self.pe.writable_data(expected['address'], len(data))
        self.memory.write(expected['address'], data)

    def initial(self):
        row, case = self.row, self.case
        assert row['input'] == case and row['input_sha256'] == digest(canon(case))
        assert row['control_word'] == self.cw and row['name'] == case['name']
        assert row['status'] in ('returned', 'original-fault') and row['harness_error'] is None
        for init in case.get('image_initializers', []):
            self.host_expected('input-image-initializer', [init])
        caller = struct.pack('<6I', STOP, *map(number, case['args']))
        self.expected_host.extend([
            dict(reason='input-caller-words', address=SP, payload=caller),
            dict(reason='framework-fninit-fldcw', address=INIT, payload=b'\xdb\xe3\xd9\x2d'+struct.pack('<I', CW_AT)),
            dict(reason='framework-control-word', address=CW_AT, payload=struct.pack('<H', self.cw))])
        prefix = row['input_initialization_events_before_snapshot']
        assert prefix == len(self.expected_host)
        events = row['complete_events']
        assert all(e['sequence'] == i for i, e in enumerate(events))
        for event in events[:prefix]:
            assert event['kind'] == 'host-write'
            self.host(event, initial=True)
        assert not self.expected_host
        self.snapshots.compare(self.memory, row['complete_memory_before'])
        assert row['caller_before_hex'] == caller.hex() == self.memory.read(SP, 24).hex()
        registers(row['registers_before'])
        assert row['registers_before'] == dict(eax=0, ecx=0, edx=0, esp=SP, eip=ROOT, eflags=0x202, **SAVED)
        candidates = row['completed_call_frames']+row['open_call_frames']
        if row['pending_return_at_terminal'] is not None:
            candidates.append(row['pending_return_at_terminal'])
        roots = [frame for frame in candidates if frame['call_site'] is None]
        assert len(roots) == 1
        root = roots[0]
        assert (root['kind'], root['name'], number(root['entry']), number(root['stack_entry']), number(root['return_pc'])) == (
            'actual-original', 'parent', ROOT, SP, STOP)
        assert root['args'] == list(map(number, case['args'])) and root['registers_before'] == row['registers_before']
        fpu(root['FPU_before'], self.cw)
        assert root['FPU_before']['status'] == 0 and root['FPU_before']['tag'] == 0xffff
        self.frames.append(json.loads(canon({k: root[k] for k in (
            'kind', 'name', 'entry', 'call_site', 'stack_entry', 'return_pc', 'args', 'registers_before', 'FPU_before')})))
        return prefix

    def check(self):
        prefix = self.initial()
        events = self.row['complete_events']
        for event in events[prefix:]:
            kind = event['kind']
            self.counts[kind] += 1
            if kind in ('original-instruction', 'external-x87-trampoline-instruction', 'call-entered', 'return-completed', 'fault'):
                assert not self.expected_host, ('effect not applied before next execution boundary', self.expected_host, event)
            if kind == 'host-write':
                self.host(event)
            elif kind in ('original-instruction', 'external-x87-trampoline-instruction'):
                self.instruction(event)
            elif kind == 'read':
                self.read(event)
            elif kind == 'write':
                self.write(event)
            elif kind == 'call-issued':
                self.issue(event)
            elif kind == 'call-entered':
                self.enter(event['frame'])
            elif kind == 'return-completed':
                self.returned_frame(event['frame'])
            elif kind == 'external-service-input':
                self.service(event)
            elif kind == 'fault':
                self.fault(event)
            elif kind == 'observed-partial-write-fragment':
                self.fragment(event)
            elif kind == 'faulting-write-observed':
                self.observation(event)
            else:
                raise AssertionError(('unknown event kind', kind))
        assert not self.expected_host
        self.final()
        return dict(status=self.row['status'], entered=self.entered, completed=self.completed, entries=self.entries,
                    event_counts=self.counts, service_counts=self.services, reads=self.reads,
                    faults=self.faults, partial_fragments=len(self.partial_events), faulting_write_observations=len(self.fault_write_observations),
                    used_reads=len(self.read_used), used_services=len(self.service_used),
                    unused_reads=len(self.case.get('after_reads', []))-len(self.read_used),
                    unused_services=len(self.case.get('after_services', []))-len(self.service_used),
                    changed_caller_bytes=[i for i, (a, b) in enumerate(zip(bytes.fromhex(self.row['caller_before_hex']),
                                                                         bytes.fromhex(self.row['caller_after_hex']))) if a != b])

    def instruction(self, event):
        pc = number(event['pc'])
        assert isinstance(event['completed'], bool)
        self.current = event
        if event['kind'] == 'original-instruction':
            decoded = self.instructions[pc]
            assert (event['size'], event['encoding'], event['range_name']) == (decoded['size'], decoded['encoding'], decoded['name'])
            assert event['frame_depth'] == len(self.frames)
            self.entered[pc] += 1
            if event['completed']:
                self.completed[pc] += 1
            if pc in (ROOT, 0x4e1400, 0x4e1470, 0x419a70):
                self.entries[pc] += 1
            is_return = decoded['mnemonic'] == 'ret'
        else:
            assert self.frames and self.frames[-1]['kind'] == 'external-input-boundary'
            response = self.active_response
            assert 'x87' in response and event['original_execution_credit'] is False
            opcode = {'f32': b'\xd9\x05', 'f64': b'\xdd\x05', 'ext80': b'\xdb\x2d'}[response['x87']['format']]
            expected = opcode+struct.pack('<I', PAYLOAD) if pc == TRAMP else b'\xc3'
            assert pc in (TRAMP, TRAMP+6) and bytes.fromhex(event['encoding']) == expected and event['size'] == len(expected)
            is_return = pc == TRAMP+6
        if is_return:
            assert self.pending_return is None and self.frames
            self.pending_return = self.frames.pop()

    def read(self, event):
        pc, address, width = number(event['pc']), number(event['address']), event['width']
        assert self.current is not None and number(self.current['pc']) == pc
        source = 'actual-original' if pc in self.instructions else 'external-x87-trampoline'
        assert event['source'] == source
        self.reads[(pc, address, width)] += 1
        assert event['occurrence'] == self.reads[(pc, address, width)]
        actual = self.memory.read(address, width)
        assert event['value_hex'] == (None if actual is None else actual.hex())
        if event['completed']:
            assert actual is not None and not event.get('fault')
        if source == 'actual-original' and event['completed']:
            for i, stimulus in enumerate(self.case.get('after_reads', [])):
                if tuple(map(number, (stimulus['pc'], stimulus['address'], stimulus['width'], stimulus['occurrence']))) == (
                        pc, address, width, event['occurrence']):
                    assert i not in self.read_used
                    self.read_used.add(i)
                    self.host_expected('after-completed-original-read', stimulus['effects'], stimulus_index=i, read_sequence=event['sequence'])

    def write(self, event):
        pc, address, width = number(event['pc']), number(event['address']), event['width']
        assert pc in self.instructions and self.current['pc'] == event['pc'] and event['source'] == 'actual-original'
        assert 1 <= width <= 8
        current = self.memory.read(address, width)
        assert event['before'] == (None if current is None else current.hex())
        assert event['mapped_spans_before'] == self.memory.spans(address, width)
        registers(event['registers_at_write'])
        assert event['registers_at_write']['eip'] == pc
        expected = (event['hook_value_low64'] & ((1 << (8*width))-1)).to_bytes(width, 'little')
        self.writes[event['sequence']] = event
        if event['completed']:
            assert current is not None and not event.get('fault')
            assert bytes.fromhex(event['after']) == expected
            self.memory.write(address, expected)
        else:
            assert 'after' not in event

    def issue(self, event):
        assert self.pending_call is None
        site = number(event['site'])
        assert self.current['pc'] == event['site'] and self.instructions[site]['mnemonic'] == 'call'
        registers(event['registers']); fpu(event['FPU'], self.cw)
        sp = number(event['stack_before'])
        assert event['registers']['esp'] == sp and event['registers']['eip'] == site
        target = None if event['target'] is None else number(event['target'])
        if site in self.direct:
            assert target == self.direct[site]
        elif site == 0x4e0d05:
            bits = self.memory.read((event['registers']['edi']+716) & 0xffffffff, 4)
            assert target == (None if bits is None else int.from_bytes(bits, 'little'))
        else:
            assert site in (0x4e1049, 0x4e104f) and target == event['registers']['eax']
        argc = self.external[site]['argument_words'] if site in self.external else {0x4e1400: 2, 0x4e1470: 1, 0x419a70: 1}[target]
        assert event['argument_words'] == argc and event['args'] == [self.memory.word(sp+4*i) for i in range(argc)]
        self.pending_call = dict(site=site, target=target, return_pc=site+self.call_sizes[site], stack_before=sp)

    def enter(self, frame):
        call = self.pending_call
        assert call is not None
        self.pending_call = None
        site, address, sp = number(frame['call_site']), number(frame['entry']), number(frame['stack_entry'])
        assert site == call['site'] and address == call['target'] and sp == call['stack_before']-4
        assert self.memory.word(sp) == number(frame['return_pc']) == call['return_pc']
        spec = self.external.get(site)
        argc = spec['argument_words'] if spec else {0x4e1400: 2, 0x4e1470: 1, 0x419a70: 1}[address]
        assert frame['args'] == [self.memory.word(sp+4+4*i) for i in range(argc)]
        assert frame['kind'] == ('external-input-boundary' if spec else 'actual-original')
        assert frame['name'] == (spec['name'] if spec else self.instructions[address]['name'])
        registers(frame['registers_before']); fpu(frame['FPU_before'], self.cw)
        assert frame['registers_before']['esp'] == sp and frame['registers_before']['eip'] == address
        assert frame['completed'] is False
        self.frames.append(json.loads(canon(frame)))

    def service(self, event):
        site = number(event['site'])
        spec, frame = self.external[site], self.frames[-1]
        assert number(frame['call_site']) == site and frame['kind'] == 'external-input-boundary'
        assert event['name'] == spec['name'] and event['args'] == frame['args'] and event['original_helper_executed'] is False
        self.services[site] += 1
        occurrence = self.services[site]
        assert event['occurrence'] == occurrence
        schedules = {number(k): v for k, v in self.case['services'].items()}
        schedule = schedules[site]
        responses = schedule if isinstance(schedule, list) else schedule['responses']
        repeated = occurrence > len(responses)
        assert not repeated or (isinstance(schedule, dict) and schedule.get('repeat_last') is True)
        selected = min(occurrence-1, len(responses)-1)
        response = responses[selected]
        assert event['response_index'] == selected and event['repeated'] == repeated and event['input_response'] == response
        self.active_response = response
        frame['service_occurrence'] = occurrence
        frame['_response'] = response
        self.host_expected('external-service-input-effect', response.get('effects', []), frame['args'], service_site=hx(site))
        if spec['return_kind'] == 'x87':
            payload = response['x87']
            code = {'f32': b'\xd9\x05', 'f64': b'\xdd\x05', 'ext80': b'\xdb\x2d'}[payload['format']]+struct.pack('<I', PAYLOAD)+b'\xc3'
            self.expected_host.extend([dict(reason='framework-exact-x87-service-payload', address=PAYLOAD,
                                            payload=bytes.fromhex(payload['hex']), service_site=hx(site)),
                                       dict(reason='framework-exact-x87-fld-ret', address=TRAMP, payload=code, service_site=hx(site))])

    def returned_frame(self, frame):
        if self.pending_return is not None:
            opened, self.pending_return = self.pending_return, None
        else:
            assert self.frames and self.frames[-1]['kind'] == 'external-input-boundary'
            opened = self.frames.pop()
        for key in ('kind', 'name', 'call_site', 'entry', 'stack_entry', 'return_pc', 'args', 'registers_before', 'FPU_before'):
            assert frame[key] == opened[key], key
        before, after = frame['registers_before'], frame['registers_after']
        registers(after); fpu(frame['FPU_after'], self.cw)
        assert frame['completed'] is True and frame['callee_saved_preserved'] is True
        assert all(after[n] == before[n] for n in SAVED)
        assert after['esp'] == number(frame['stack_entry'])+4 == number(frame['stack_exit'])
        assert after['eip'] == number(frame['return_pc']) == number(frame['returned_to'])
        self.returned.append(frame)
        if frame['kind'] == 'external-input-boundary':
            site, occurrence = number(frame['call_site']), opened['service_occurrence']
            assert frame['service_occurrence'] == occurrence
            response = opened['_response']
            for name in ('eax', 'ecx', 'edx'):
                assert after[name] == (number(response[name]) if name in response else before[name])
            assert after['eflags'] == before['eflags']
            if 'x87' in response:
                assert frame['x87_return_input'] == response['x87']
            for i, stimulus in enumerate(self.case.get('after_services', [])):
                if number(stimulus['site']) == site and number(stimulus['occurrence']) == occurrence:
                    assert i not in self.service_used
                    self.service_used.add(i)
                    self.host_expected('after-completed-external-service', stimulus['effects'], frame['args'], stimulus_index=i, service_site=hx(site))

    def fault(self, event):
        assert event['access'] in (19, 20, 21, 22, 23, 24) and 1 <= event['width'] <= 16
        registers(event['registers']); fpu(event['FPU'], self.cw)
        pc, address, width, access = number(event['pc']), number(event['address']), event['width'], event['access']
        assert event['registers']['eip'] == pc
        permission = {19: 1, 20: 2, 21: 4, 22: 2, 23: 1, 24: 4}[access]
        invalid = self.memory.invalid_bytes(address, width, permission)
        assert invalid, 'reported invalid range is fully mapped with required guest permission'
        if access in (19, 20, 21):
            assert any(self.memory.read(at, 1) is None for at in invalid), 'UNMAPPED fault has no unmapped byte'
        else:
            assert any(self.memory.read(at, 1) is not None for at in invalid), 'PROT fault has no mapped forbidden byte'
        if access in (21, 24):
            # In this frozen domain caller RET slots are preserved and STOP is
            # mapped. The admitted fetch-fault source is an actual pending CALL.
            assert self.pending_call is not None and self.pending_call['target'] is not None
            assert pc == self.pending_call['target'] and address <= pc < address+width
            assert number(self.current['pc']) == self.pending_call['site'] and self.current['completed'] is True
            assert self.memory.word(self.pending_call['stack_before']-4) == self.pending_call['return_pc']
        else:
            assert self.current is not None and number(self.current['pc']) == pc
            assert self.current['completed'] is False and self.current.get('fault') is True
            assert pc in self.instructions, 'data fault outside declared original instruction scope'
            decoded, regs, candidates = self.instructions[pc], event['registers'], []
            # LEA/NOP may have a syntactic MEM operand but do not load/store it.
            operands = [] if decoded['mnemonic'] in ('lea', 'nop') else decoded['memory_operands']
            for operand in operands:
                if operand['segment'] is not None:
                    continue
                # Capstone reports access0 for the two original TEST frame reads;
                # a missing access flag must not erase a real decoded read.
                declared = operand['decoder_access']
                if declared and not declared & permission:
                    continue
                base = regs[operand['base']] if operand['base'] else 0
                index = regs[operand['index']] if operand['index'] else 0
                effective = (base+index*operand['scale']+operand['displacement']) & 0xffffffff
                candidates.append((effective, operand['width'], 'decoded-explicit-memory-operand'))
            if permission == 1 and decoded['mnemonic'] in ('pop', 'ret'):
                start = number(self.pending_return['stack_entry']) if decoded['mnemonic'] == 'ret' and self.pending_return else regs['esp']
                candidates.append((start, 4, 'implicit-cdecl-stack-read'))
            if permission == 2 and decoded['mnemonic'] == 'call' and self.pending_call is not None:
                candidates.append((self.pending_call['stack_before']-4, 4, 'actual-CALL-return-PC-stack-write'))
            if permission == 2 and decoded['mnemonic'] == 'push':
                # Stack is an owned64KiB RW allocation in this frozen domain.
                # Do not infer an unobserved PUSH destination from post-fault ESP.
                for write in self.writes.values():
                    if number(write['pc']) == pc and not write['completed'] and write['width'] == 4:
                        start = number(write['address'])
                        if STACK <= start and start+4 <= STACK+STACK_SIZE:
                            candidates.append((start, 4, 'observed-owned-stack-PUSH-write'))
            assert any(any(start <= at < start+count for at in invalid) for start, count, _ in candidates), (
                'fault range is not justified by a decoded operand/actual stack extent', event, candidates)
        self.faults.append(event)

    def fragment(self, event):
        write = self.writes[event['original_write_sequence']]
        assert write['completed'] is False and write.get('fault') is True
        assert event['original_instruction_completed'] is False and event['emulator_physical_observation_only'] is True
        address, width, offset = number(event['address']), event['width'], event['write_offset']
        assert event['pc'] == write['pc'] and 0 <= offset < offset+width <= write['width'] <= 8
        assert address == number(write['address'])+offset
        before, after = bytes.fromhex(event['before']), bytes.fromhex(event['after'])
        desired = (write['hook_value_low64'] & ((1 << (8*write['width']))-1)).to_bytes(write['width'], 'little')[offset:offset+width]
        assert len(before) == len(after) == width and after == desired == bytes.fromhex(event['original_hook_value_slice'])
        assert all(a != b for a, b in zip(before, after))
        assert self.memory.read(address, width) == before
        assert event['sequence'] in write['faulting_write_observation']['changed_fragment_sequences']
        self.memory.write(address, after)
        self.partial_events[event['sequence']] = event

    def observation(self, event):
        write = self.writes[event['original_write_sequence']]
        width, start, pc = write['width'], number(write['address']), number(write['pc'])
        assert write['completed'] is False and write.get('fault') is True and width <= 8
        assert event['pc'] == write['pc'] and number(event['address']) == start and event['width'] == width
        assert event['original_instruction_completed'] is False and event['emulator_physical_observation_only'] is True
        assert any(f['access'] in (20, 22) and number(f['pc']) == pc and
                   max(number(f['address']), start) < min(number(f['address'])+f['width'], start+width) for f in self.faults)
        operands, regs = [], write['registers_at_write']
        for operand in self.instructions[pc]['memory_operands']:
            if operand['width'] == width and operand['segment'] is None:
                base = regs[operand['base']] if operand['base'] else 0
                index = regs[operand['index']] if operand['index'] else 0
                if (base+index*operand['scale']+operand['displacement']) & 0xffffffff == start:
                    operands.append(operand)
        assert len(operands) == 1
        observed = write['faulting_write_observation']
        assert observed['decoded_memory_operand'] == operands[0]
        assert observed['original_instruction_completed'] is False and observed['emulator_physical_observation_only'] is True
        assert observed['changed_fragment_sequences'] == event['changed_fragment_sequences']
        assert all(i in self.partial_events for i in event['changed_fragment_sequences'])
        changed = {}
        for before, after in zip(write['mapped_spans_before'], observed['mapped_spans_after']):
            assert all(after[k] == v for k, v in before.items())
            old, new = bytes.fromhex(before['before_hex']), bytes.fromhex(after['after_hex'])
            assert self.memory.read(number(before['address']), before['width']) == new
            differences = [i for i, (a, b) in enumerate(zip(old, new)) if a != b]
            assert differences == after['changed_byte_offsets']
            changed.update({before['write_offset']+i: new[i] for i in differences})
        assert len(write['mapped_spans_before']) == len(observed['mapped_spans_after'])
        fragments = {}
        for index in event['changed_fragment_sequences']:
            fragment = self.partial_events[index]
            assert fragment['original_write_sequence'] == write['sequence']
            for i, byte in enumerate(bytes.fromhex(fragment['after'])):
                at = fragment['write_offset']+i
                assert at not in fragments
                fragments[at] = byte
        assert fragments == changed
        self.fault_write_observations.append(event)

    def final(self):
        row = self.row
        assert row['faults'] == self.faults and row['completed_call_frames'] == self.returned
        # Runtime-only response annotations are removed before comparing captured open frames.
        opened = [{k: v for k, v in frame.items() if not k.startswith('_')} for frame in self.frames]
        pending = None if self.pending_return is None else {k: v for k, v in self.pending_return.items() if not k.startswith('_')}
        assert row['open_call_frames'] == opened and row['pending_return_at_terminal'] == pending
        assert row['pending_call_at_terminal'] == self.pending_call
        assert row['instruction_coverage'] == {hx(k): v for k, v in sorted(self.entered.items())}
        assert row['completed_original_instruction_coverage'] == {hx(k): v for k, v in sorted(self.completed.items())}
        assert row['actual_original_entry_counts'] == {hx(k): v for k, v in sorted(self.entries.items())}
        assert self.entries[ROOT] == 1
        external_entries = sum(frame['kind'] == 'external-input-boundary' for frame in self.row['completed_call_frames'])
        external_entries += sum(frame['kind'] == 'external-input-boundary' for frame in self.row['open_call_frames'])
        if self.row['pending_return_at_terminal'] is not None:
            external_entries += self.row['pending_return_at_terminal']['kind'] == 'external-input-boundary'
        assert self.row['instructions'] == sum(self.entered.values()) + self.counts['external-x87-trampoline-instruction'] + external_entries
        assert row['service_call_counts'] == {hx(k): v for k, v in sorted(self.services.items())}
        for field, actual, total in [('read', self.read_used, len(self.case.get('after_reads', []))),
                                     ('service', self.service_used, len(self.case.get('after_services', [])))]:
            assert row['used_after_'+field+'_stimuli'] == sorted(actual)
            assert row['unused_after_'+field+'_stimuli'] == sorted(set(range(total))-actual)
        assert row['observed_partial_write_fragments'] == len(self.partial_events)
        assert row['faulting_write_observations'] == len(self.fault_write_observations)
        observed_ids = [event['original_write_sequence'] for event in self.fault_write_observations]
        assert len(observed_ids) == len(set(observed_ids))
        assert set(observed_ids) == {key for key, write in self.writes.items() if 'faulting_write_observation' in write}
        assert set(observed_ids) == {key for key, write in self.writes.items() if write.get('fault') and not write['completed']}
        assert sum(len(w.get('faulting_write_observation', {}).get('changed_fragment_sequences', [])) for w in self.writes.values()) == len(self.partial_events)
        self.snapshots.compare(self.memory, row['complete_memory_after'])
        assert row['caller_after_hex'] == self.memory.read(SP, 24).hex()
        assert len(row['all_memory_write_reconstruction']) == len(self.memory.areas)
        for (name, _, data, _), claim in zip(self.memory.areas, row['all_memory_write_reconstruction']):
            assert claim == dict(name=name, bytes=len(data), reconstructed_from_completed_writes_and_observed_fault_fragments=True,
                                 actual_sha256=digest(data), reconstructed_sha256=digest(data))
        assert row['original_executable_bytes_unchanged'] is True
        for region in self.spec['actual_original_ranges']:
            assert self.memory.read(number(region['address']), region['size']) == self.pe.read(number(region['address']), region['size'])
        registers(row['registers_after']); fpu(row['FPU_after'], self.cw)
        if row['status'] == 'returned':
            assert not self.faults and row['emulation_error'] is None and not opened and pending is None and self.pending_call is None
            assert row['registers_after']['eip'] == STOP and row['registers_after']['esp'] == SP+4
            assert all(row['registers_after'][n] == v for n, v in SAVED.items())
        else:
            assert self.faults and row['emulation_error']['type'] == 'UcError'
            assert row['registers_after'] == self.faults[-1]['registers'] and row['FPU_after'] == self.faults[-1]['FPU']
        cleanup = row['recorder_cleanup']
        assert cleanup['registered_hooks'] == cleanup['removed_hooks'] == 4 and cleanup['errors'] == []
        assert cleanup['native_engine_reference_released'] is True


def qualify(seal_path, runout, cache_bytes):
    qualifier_before = file_record(__file__)
    seal_row = file_record(seal_path)
    seal = json.loads(Path(seal_path).read_text())
    assert seal['schema'] == 'opennox-default-damage-parent-input-freeze-v1'
    assert seal['root_reviewed_before_execution'] is True and seal['native_engine_exists'] is False
    files = {key: verify_file(seal[key]) for key in ('producer', 'services', 'original_EXE', 'unchanged_main_manifest', 'dedicated_manifest', 'inputs')}
    assert files['producer']['sha256'] == PRODUCER_SHA and files['services']['sha256'] == SERVICE_SHA
    assert files['original_EXE']['sha256'] == EXE_SHA and files['unchanged_main_manifest']['sha256'] == MAIN_SHA
    services = json.loads(Path(files['services']['path']).read_text())
    inputs_raw = Path(files['inputs']['path']).read_bytes()
    unpacked = gzip.decompress(inputs_raw) if inputs_raw[:2] == b'\x1f\x8b' else inputs_raw
    inputs = json.loads(unpacked)
    assert inputs['schema'] == 'opennox-default-damage-parent-independent-inputs-v1'
    cases = inputs['cases']
    assert len(cases) == seal['inputs']['count'] and len({c['name'] for c in cases}) == len(cases)
    by_name = {case['name']: case for case in cases}
    pe = OriginalPE(files['original_EXE']['path'])
    main = json.loads(Path(files['unchanged_main_manifest']['path']).read_text())
    assert len(main['functions']) == 3373 and len(main['data']) == 713
    dedicated = json.loads(Path(files['dedicated_manifest']['path']).read_text())
    assert dedicated['file_sha256'] == EXE_SHA and len(dedicated['functions']) == 14 and len(dedicated['data']) == 3
    for region in dedicated['functions']+dedicated['data']:
        assert digest(pe.read(number(region['address']), region['size'])) == region['sha256']
    assert digest(pe.read(number(services['padding9']['address']), services['padding9']['size'])) == services['padding9']['sha256']
    assert digest(pe.read(ROOT, 1792)) == services['parent_union1792_sha256']
    instructions, groups, capstone_version = instruction_table(pe, services)
    proof_path = runout/'actual-original-parent-recorder-proof-v2.json'
    proof_before = file_record(proof_path)
    proof = json.loads(proof_path.read_text())
    assert proof['schema'] == 'opennox-default-damage-parent-actual-original-recorder-proof-v2' and proof['fatal'] is None
    assert proof['preparation']['seal'] == seal_row and proof['preparation']['cases'] == len(cases)
    assert proof['preparation']['inputs_uncompressed_sha256'] == digest(unpacked)
    assert all(proof['preparation'][key] == value for key, value in files.items())
    assert proof['postexecution_source_input_bindings'] == {**files, 'input_seal': seal_row}
    binding = json.loads((runout/'input-and-source-binding-v2.json').read_text())
    assert binding == proof['preparation']
    snapshots = Snapshots(runout, proof['memory_snapshots'], cache_bytes)
    rows_path = runout/'actual-original-parent-rows-v2.jsonl.gz'
    rows_before = file_record(rows_path)
    assert proof['rows'] == rows_before
    seen, totals, entered, completed, entries, events = Counter(), Counter(), Counter(), Counter(), Counter(), Counter()
    modes = {name: Counter() for name in MODES}
    fault_domains, changed_callers = Counter(), Counter()
    partial_count = observation_count = used_reads = unused_reads = used_services = unused_services = 0
    with gzip.open(rows_path, 'rb') as stream:
        for line_number, line in enumerate(stream, 1):
            assert line.endswith(b'\n') and len(line) <= 64 << 20
            row = json.loads(line)
            identity = (row['name'], row['mode'])
            assert row['name'] in by_name and row['mode'] in MODES and seen[identity] == 0, identity
            seen[identity] += 1
            try:
                result = RowCheck(pe, services, instructions, snapshots, by_name[row['name']], row).check()
            except Exception as exc:
                raise AssertionError(dict(row=line_number, name=row['name'], mode=row['mode'], reason=str(exc))) from exc
            totals[result['status']] += 1
            modes[row['mode']][result['status']] += 1
            entered.update(result['entered']); completed.update(result['completed']); entries.update(result['entries']); events.update(result['event_counts'])
            partial_count += result['partial_fragments']; observation_count += result['faulting_write_observations']
            used_reads += result['used_reads']; unused_reads += result['unused_reads']
            used_services += result['used_services']; unused_services += result['unused_services']
            for fault in result['faults']:
                fault_domains[(row['mode'], fault['pc'], fault['access'], fault['address'], fault['width'])] += 1
            for offset in result['changed_caller_bytes']:
                changed_callers[offset] += 1
    assert set(seen) == {(case['name'], mode) for case in cases for mode in MODES}
    assert all(value == 1 for value in seen.values())
    assert proof['requested_case_count'] == len(cases) and proof['requested_original_root_calls'] == proof['actual_rows'] == len(seen) == 2*len(cases)
    assert proof['statuses'] == dict(totals)
    assert proof['actual_entered_originals'] == {hx(k): v for k, v in sorted(entries.items())}
    assert proof['original_instruction_coverage_union'] == {hx(k): v for k, v in sorted(entered.items())}
    assert proof['original_completed_instruction_coverage_union'] == {hx(k): v for k, v in sorted(completed.items())}
    assert proof['original_static_instruction_boundaries'] == groups
    missing = {name: [i['pc'] for i in group if not entered[number(i['pc'])]] for name, group in groups.items()}
    never_completed = {name: [i['pc'] for i in group if not completed[number(i['pc'])]] for name, group in groups.items()}
    assert proof['unentered_original_instruction_boundaries'] == missing
    assert proof['never_completed_original_instruction_boundaries'] == never_completed
    assert proof['emulator_observed_partial_write_fragments'] == partial_count and proof['faulting_write_observations'] == observation_count
    assert proof['native_fault_atomicity_qualified'] is False
    assert len(proof['modes']) == 2 and {row['label']: number(row['cw']) for row in proof['modes']} == MODES
    assert proof['immutable_original_EXE'] is True and proof['no_native_engine_used'] is True
    for ref in proof['memory_snapshots']:
        snapshots.get(ref)  # verify even unused/failed-row snapshot members, without a raw copy
    assert len(snapshots.verified) == len(proof['memory_snapshots'])
    post = {key: verify_file(seal[key]) for key in files}
    assert post == files and file_record(seal_path) == seal_row
    assert file_record(proof_path) == proof_before and file_record(rows_path) == rows_before
    assert file_record(__file__) == qualifier_before
    for name, observed in snapshots.verified.items():
        assert file_record(runout/name) == observed
    return dict(schema='opennox-default-damage-parent-independent-original-postrun-qualified-v4', qualified=True,
                qualifier=file_record(__file__), input_seal=seal_row, original_proof=file_record(proof_path),
                actual_rows=file_record(rows_path), source_and_input_bindings=files,
                case_count=len(cases), exact_case_two_mode_multiplicity=True,
                actual_entered_originals={hx(k): v for k, v in sorted(entries.items())},
                observed_statuses=dict(totals), observed_statuses_per_mode={name: dict(value) for name, value in modes.items()},
                observed_fault_domains=[dict(mode=k[0], pc=k[1], access=k[2], address=k[3], width=k[4], count=v)
                                        for k, v in sorted(fault_domains.items())],
                observed_emulator_partial_write_fragments=partial_count, faulting_write_observations=observation_count,
                native_fault_atomicity_qualified=False,
                entered_instruction_counts={hx(k): v for k, v in sorted(entered.items())},
                completed_instruction_counts={hx(k): v for k, v in sorted(completed.items())},
                unentered_original_boundaries=missing, never_completed_original_boundaries=never_completed,
                all_static_original_boundaries_entered=not any(missing.values()),
                all_static_original_boundaries_completed=not any(never_completed.values()),
                event_counts=dict(events), stimulus_counts=dict(used_after_reads=used_reads, unused_after_reads=unused_reads,
                                                              used_after_services=used_services, unused_after_services=unused_services),
                changed_caller_byte_offsets_observed={str(k): v for k, v in sorted(changed_callers.items())},
                canonical_memory_member_count=len(snapshots.verified),
                canonical_memory_member_manifest_sha256=digest(canon(sorted(snapshots.verified.items()))),
                maximum_snapshot_cache_bytes=cache_bytes, all_full_memory_bytes_independently_reconstructed=True,
                every_canonical_memory_gzip_SHA_and_uncompressed_SHA_verified=True,
                exact_call_issued_and_entered_arguments_from_reconstructed_stack=True,
                every_fault_PC_range_and_mapping_permission_closed=True,
                caller24bytes_preserved_including_real_scratch_writes=True,
                current_source_seal_reverified_after_analysis=True,
                capstone_static_disassembly_version=capstone_version,
                producer_imported_or_executed=False, original_or_Go_or_game_execution=False,
                expected_fault_or_return_outcomes_computed=False, native_engine_or_product_credit=0,
                scope='Original recorder certificates only. Fault/return counts are observations, not expected native outcomes. Original external service body credit remains zero.')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--input-seal', type=Path, required=True)
    parser.add_argument('--runout', type=Path, required=True)
    parser.add_argument('--output', type=Path, required=True)
    parser.add_argument('--snapshot-cache-MiB', type=int, default=24)
    args = parser.parse_args()
    assert 1 <= args.snapshot_cache_MiB <= 64
    assert not args.output.exists()
    assert not args.output.resolve().is_relative_to(args.runout.resolve()), 'output must be outside canonical runout'
    try:
        result = qualify(args.input_seal.resolve(), args.runout.resolve(), args.snapshot_cache_MiB << 20)
    except Exception as exc:
        result = dict(schema='opennox-default-damage-parent-independent-original-postrun-failure-v4', qualified=False,
                      error_type=type(exc).__name__, error=str(exc), traceback=traceback.format_exc(),
                      qualifier=file_record(__file__), input_seal_path=str(args.input_seal.resolve()),
                      runout=str(args.runout.resolve()), canonical_inputs_or_runout_written_by_qualifier=False,
                      original_or_Go_or_game_execution=False)
    with args.output.open('x') as f:
        json.dump(result, f, indent=2)
        f.write('\n')
    print(json.dumps(dict(qualified=result['qualified'], output=str(args.output),
                          actual_case_count=result.get('case_count'), observed_statuses=result.get('observed_statuses'))))
    return 0 if result['qualified'] else 1


if __name__ == '__main__':
    sys.exit(main())
