"""Execute the original sender and actual scalar/selector/RNG helpers.

Phase A admits warm predicate caches. Only network publication services are
declared. Cold dictionary lookup/mutation and native comparators remain a
separate required gate; this producer never certifies overall completion.
"""
import argparse
import collections
import gzip
import hashlib
import itertools
import json
from pathlib import Path
import struct

from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_WRITE, UC_HOOK_MEM_READ
from unicorn.x86_const import *

parser = argparse.ArgumentParser(description=__doc__)
for name in ['game', 'fixture', 'proof']:
    parser.add_argument('--' + name, type=Path, required=True)
parser.add_argument('--control', type=lambda s: int(s, 0), choices=[0xe3f, 0x23f], required=True)
parser.add_argument('--limit', type=int, default=0)
args = parser.parse_args()
stage = Path(__file__).resolve().parent
audit = json.loads((stage / 'original-functions-and-tables-static-v1.json').read_text())
blob = args.game.read_bytes()
sha = lambda b: hashlib.sha256(b).hexdigest()
assert sha(blob) == audit['original_sha256']
pe = struct.unpack_from('<I', blob, 60)[0]
opt = pe + 24
base = struct.unpack_from('<I', blob, opt + 28)[0]
tab = opt + struct.unpack_from('<H', blob, pe + 20)[0]
u = Uc(UC_ARCH_X86, UC_MODE_32)
u.mem_map(base, (struct.unpack_from('<I', blob, opt + 56)[0] + 4095) & ~4095)
for i in range(struct.unpack_from('<H', blob, pe + 6)[0]):
    _, rva, size, off = struct.unpack_from('<IIII', blob, tab + 40 * i + 8)
    if size:
        u.mem_write(base + rva, blob[off:off + size])
instructions = {int(r['address'], 16): bytes.fromhex(r['bytes']) for f in audit['functions'] for r in f['instructions']}
assert sum(map(len, instructions.values())) == 1192
for f in audit['functions']:
    assert sha(bytes(u.mem_read(int(f['entry'], 16), f['size']))) == f['sha256']
for d in audit['data']:
    assert sha(bytes(u.mem_read(int(d['address'], 16), d['size']))) == d['sha256']
heap, stack, code = 0x21000000, 0x22000000, 0x23000000
for address in [heap, stack, code]:
    u.mem_map(address, 0x10000)
unit, update, target, health = heap + 0x1000, heap + 0x3000, heap + 0x7000, heap + 0x9000
sp, stop, init, cw = stack + 0xe000, code + 0x100, code + 0x200, code + 0x300
u.mem_write(init, b'\xdb\xe3\xd9\x2d' + struct.pack('<I', cw))
u.mem_write(cw, struct.pack('<H', args.control))
rd = lambda a: struct.unpack('<I', u.mem_read(a, 4))[0]
wr = lambda a, v: u.mem_write(a, struct.pack('<I', v & 0xffffffff))
bits = lambda f: struct.unpack('<I', struct.pack('<f', f))[0]
cache_base, rng_index, game_frame, frame_table = 0x833e60, 0x62efd0, 0x84ea04, 0x753430
saved = {UC_X86_REG_EBX: 0x12345678, UC_X86_REG_ESI: 0x23456789, UC_X86_REG_EDI: 0x3456789a, UC_X86_REG_EBP: 0x456789ab}
active = False
visited = set()
calls = collections.Counter()
table_reads = {d['name']: set() for d in audit['data']}
current = None
events, writes, stack_writes = [], [], []

def on_code(_u, pc, size, _):
    if not active:
        return
    pointer = u.reg_read(UC_X86_REG_ESP)
    if pc in [0x4e5420, 0x40ef40]:
        length = rd(pointer + 12)
        assert length == (5 if pc == 0x4e5420 else 11)
        assert rd(pointer + 4) == current['player']
        if pc == 0x4e5420:
            assert rd(pointer + 16) == 0 and rd(pointer + 20) == 1
        events.append({'service': hex(pc), 'player': rd(pointer + 4), 'packet': bytes(u.mem_read(rd(pointer + 8), length)).hex()})
        u.reg_write(UC_X86_REG_EAX, current['queue_return'] if pc == 0x40ef40 else 1)
        u.reg_write(UC_X86_REG_ECX, 0x87654321)
        u.reg_write(UC_X86_REG_EDX, 0x98765432)
        u.reg_write(UC_X86_REG_ESP, pointer + 4)
        u.reg_write(UC_X86_REG_EIP, rd(pointer))
        return
    assert pc in instructions, (current['name'], 'unadmitted PC (cold lookup is outside Phase A)', hex(pc))
    assert bytes(u.mem_read(pc, size)) == instructions[pc]
    visited.add(pc)
    for f in audit['functions']:
        if pc == int(f['entry'], 16):
            calls[f['name']] += 1
            event = {'entry': hex(pc)}
            if pc in [0x419a70, 0x4f9f90, 0x509e20]:
                event['argument'] = rd(pointer + 4)
            if pc == 0x415fa0:
                event.update(minimum=rd(pointer + 4), maximum=rd(pointer + 8), index_before=rd(rng_index))
            events.append(event)

def on_write(_u, _access, address, size, value, _):
    if not active:
        return
    if stack <= address and address + size <= stack + 0x10000:
        assert sp - 256 <= address and address + size <= sp + 16, (hex(address), size, hex(u.reg_read(UC_X86_REG_EIP)))
        if address >= sp:
            assert (sp + 4 <= address and address + size <= sp + 8) or (sp + 12 <= address and address + size <= sp + 16), (hex(address), size)
        stack_writes.append((address, size, value))
        return
    assert (address == rng_index and size == 4) or (address == update + 412 + 2 * current['player'] and size == 2), (current['name'], hex(address), size)
    writes.append({'address': hex(address), 'size': size, 'value': value & ((1 << (8 * size)) - 1)})

def on_read(_u, _access, address, size, _value, _):
    if not active:
        return
    for d in audit['data']:
        start = int(d['address'], 16)
        table_reads[d['name']].update(range(max(address, start), min(address + size, start + d['size'])))

u.hook_add(UC_HOOK_CODE, on_code)
u.hook_add(UC_HOOK_MEM_WRITE, on_write)
u.hook_add(UC_HOOK_MEM_READ, on_read)
defaults = dict(player=0, report=0, hp_present=True, hp=50, previous_hp=60, frame=1000, last_frame=900,
                type_index=117, object_class=2, subclass=0, primary_code=118, secondary_code=119,
                x=bits(300.25), y=bits(401.75), direction=37, head=0, action=0, status=0, enemy=False,
                sprite_frame=5, NPC_random=0, rng=0, player_frames=7, player_duration=13, queue_return=1)
cases = []

def add(group, **changes):
    c = dict(defaults, **changes)
    c['name'] = group
    cases.append(c)

for direction in range(256):
    add('all-directions', direction=direction)
for head, action, status, enemy, typ in itertools.product([-1, 0, 7, 23], list(range(39)) + [61, 0x7fffffff, 0x80000000, 0xffffffff], [0, 0x4000, 0x40000, 0x80000, 0xc4000], [False, True], [117, 11, 22, 33, 44]):
    add('warm-selector-matrix', head=head, action=action, status=status, enemy=enemy, type_index=typ)
for index in range(4096):
    add('actual-RNG-table-and-wrap', subclass=16, NPC_random=1, rng=index)
for frames, duration, index in itertools.product([0, 1, 2, 7, 255, 256, 511, -1, -2147483648, 2147483647], [0, 1, 13, 255], [0, 10, 4095]):
    add('NPC-frame-output-boundaries', subclass=16, NPC_random=1, rng=index, player_frames=frames, player_duration=duration)
for player, hp, previous, elapsed in itertools.product([0, 1, 15, 31], [0, 1, 32767, 32768, 65535], [0, 1, 32767, 32768, 65535], [0, 1, 2, 3, 100, 0xffffffff]):
    add('health-delta-and-frame-wrap', player=player, report=1, hp=hp, previous_hp=previous, frame=1000, last_frame=(1000 - elapsed) & 0xffffffff)
for report, present in itertools.product([0, 1, 0xffffffff], [False, True]):
    add('health-admission', report=report, hp_present=present)
for cls, primary, secondary in itertools.product([2, 0x20400002, 0x20000002, 0x400002], [0, 1, 0x7fff, 0x8000, 0xffff, 0xffffffff], [0, 1, 0x7fff, 0x8000, 0xffff, 0xffffffff]):
    add('actual-network-code-boundaries', object_class=cls, primary_code=primary, secondary_code=secondary)
coordinates = [0, 0x80000000, 1, 0x80000001, bits(0.5), bits(-0.5), bits(0.75), bits(-0.75), bits(65535.75), bits(-65535.75), 0x4effffff, 0x4f000000, 0xcf000000, 0xcf000001, 0x7f800000, 0xff800000, 0x7fc00001]
for x, y in itertools.product(coordinates, coordinates):
    add('actual-FISTP-coordinate-boundaries', x=x, y=y)
for result in [0, 1, 0x7fffffff, 0x80000000, 0xffffffff]:
    add('full-network-return-word', queue_return=result)
unique, seen = [], set()
for c in cases:
    key = json.dumps({k: v for k, v in c.items() if k != 'name'}, sort_keys=True, separators=(',', ':'))
    if key not in seen:
        seen.add(key)
        c['name'] = f'{len(unique):05d}/' + c['name']
        unique.append(c)
cases = unique[:args.limit] if args.limit else unique
rows = []
for current in cases:
    active = False
    u.mem_write(heap, b'\xa5' * 0x10000)
    u.mem_write(unit, b'\0' * 1024)
    u.mem_write(update, b'\0' * 4096)
    u.mem_write(target, b'\0' * 1024)
    u.mem_write(health, struct.pack('<HH', current['hp'], 65535))
    u.mem_write(unit + 4, struct.pack('<H', current['type_index']))
    for off, value in [(8, current['object_class']), (12, current['subclass']), (16, 0x1000000), (36, current['primary_code']), (40, current['secondary_code']), (56, current['x']), (60, current['y']), (536, current['last_frame']), (556, health if current['hp_present'] else 0), (748, update)]:
        wr(unit + off, value)
    u.mem_write(unit + 124, struct.pack('<h', current['direction']))
    u.mem_write(update + 544, struct.pack('<b', current['head']))
    wr(update + 552 + 24 * current['head'], current['action'])
    wr(update + 1440, current['status'])
    wr(update + 1196, target if current['enemy'] else 0)
    u.mem_write(update + 481, bytes([current['sprite_frame']]))
    u.mem_write(update + 2094, bytes([current['NPC_random']]))
    u.mem_write(update + 412 + 2 * current['player'], struct.pack('<H', current['previous_hp']))
    u.mem_write(cache_base - 64, b'\x6b' * 144)
    for i, value in enumerate([11, 22, 33, 44]):
        wr(cache_base + 4 * i, value)
    wr(rng_index, current['rng'])
    wr(game_frame, current['frame'])
    for i in range(53):
        wr(frame_table + 8 * i, current['player_frames'] if i == 50 else 0x12345678)
        wr(frame_table + 8 * i + 4, current['player_duration'] if i == 50 else 0x23456789)
    heap_before = bytearray(u.mem_read(heap, 0x10000))
    globals_before = bytes(u.mem_read(cache_base - 64, 144))
    frames_before = bytes(u.mem_read(frame_table, 424))
    rng_neighbors = bytearray(u.mem_read(rng_index - 32, 68))
    u.mem_write(stack, b'\xcc' * 0x10000)
    for off, value in [(0, stop), (4, current['player']), (8, unit), (12, current['report'])]:
        wr(sp + off, value)
    stack_before = bytearray(u.mem_read(stack, 0x10000))
    for register, value in saved.items():
        u.reg_write(register, value)
    u.reg_write(UC_X86_REG_ESP, sp)
    u.emu_start(init, init + 8, count=4)
    before_fp = [u.reg_read(r) for r in [UC_X86_REG_FPCW, UC_X86_REG_FPSW, UC_X86_REG_FPTAG]]
    events, writes, stack_writes = [], [], []
    active = True
    u.emu_start(0x518960, stop, count=2000)
    active = False
    assert u.reg_read(UC_X86_REG_EIP) == stop and u.reg_read(UC_X86_REG_ESP) == sp + 4
    assert all(u.reg_read(r) == value for r, value in saved.items())
    after_fp = [u.reg_read(r) for r in [UC_X86_REG_FPCW, UC_X86_REG_FPSW, UC_X86_REG_FPTAG]]
    assert before_fp[0] == after_fp[0] == args.control and before_fp[2] == after_fp[2] and before_fp[1] & 0x3800 == after_fp[1] & 0x3800
    for address, size, value in stack_writes:
        stack_before[address - stack:address - stack + size] = (value & ((1 << (8 * size)) - 1)).to_bytes(size, 'little')
    assert bytes(u.mem_read(stack, 0x10000)) == stack_before
    for w in writes:
        address, size, value = int(w['address'], 16), w['size'], w['value']
        if heap <= address < heap + 0x10000:
            heap_before[address - heap:address - heap + size] = value.to_bytes(size, 'little')
    assert bytes(u.mem_read(heap, 0x10000)) == heap_before
    assert bytes(u.mem_read(cache_base - 64, 144)) == globals_before
    struct.pack_into('<I', rng_neighbors, 32, rd(rng_index))
    assert bytes(u.mem_read(rng_index - 32, 68)) == rng_neighbors
    assert rd(game_frame) == current['frame'] and bytes(u.mem_read(frame_table, 424)) == frames_before
    assert len([e for e in events if e.get('service') == '0x40ef40']) == 1
    rows.append({'name': current['name'], 'input': {k: v for k, v in current.items() if k != 'name'}, 'output': {'return_word': u.reg_read(UC_X86_REG_EAX), 'events': events, 'writes': writes, 'previous_hp_after': struct.unpack('<H', u.mem_read(update + 412 + 2 * current['player'], 2))[0], 'RNG_index_after': rd(rng_index), 'FPU_before': before_fp, 'FPU_after': after_fp, 'full_heap_stack_global_neighbors_nonvolatile_and_FP_control_guards_preserved': True}})
coverage = {f['name']: sum(len(instructions[a]) for a in visited if int(f['entry'], 16) <= a < int(f['entry'], 16) + f['size']) for f in audit['functions']}
for f in audit['functions']:
    assert sha(bytes(u.mem_read(int(f['entry'], 16), f['size']))) == f['sha256']
for d in audit['data']:
    assert sha(bytes(u.mem_read(int(d['address'], 16), d['size']))) == d['sha256']
assert sha(args.game.read_bytes()) == sha(blob)
fixture = {'schema': 'original-complex-object-send-phaseA-v1', 'original_sha256': sha(blob), 'control_word': args.control, 'warm_cache_only': True, 'declared_services': ['packet-send4E5420', 'server-message-list40EF40'], 'rows': rows}
raw = json.dumps(fixture, sort_keys=True, separators=(',', ':')).encode()
with args.fixture.open('xb') as output:
    with gzip.GzipFile(filename='', fileobj=output, mode='wb', mtime=0) as compressed:
        compressed.write(raw)
proof = {'status': 'executed', 'control_word': hex(args.control), 'contexts': len(rows), 'pilot_limit': args.limit, 'producer_sha256': sha(Path(__file__).read_bytes()), 'static_dependency_sha256': sha((stage / 'original-functions-and-tables-static-v1.json').read_bytes()), 'caller_argument_slots4_and12_observed_as_original_scratch': True, 'coverage_bytes': coverage, 'executed_bytes': sum(coverage.values()), 'static_declared_bytes_not_execution': 1192, 'actual_entries': dict(calls), 'table_bytes_read': {k: len(v) for k, v in table_reads.items()}, 'fixture_sha256': sha(args.fixture.read_bytes()), 'raw_fixture_sha256': sha(raw), 'original_unchanged': True, 'cold_lookup_mutations_and_native_comparators_still_pending': True, 'overall_porting_100_percent': False}
with args.proof.open('x') as output:
    json.dump(proof, output, indent=2)
    output.write('\n')
print(json.dumps(proof), flush=True)
