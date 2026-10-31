"""Execute original518960 with frozen lookup mutations and repeated calls.
Only dictionary inputs and network publication are declared services.
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
other = heap + 0x5000
names = ["Mimic", "CarnivorousPlant", "Zombie", "VileZombie"]
records = {name: heap + 0xa000 + i * 0x100 for i, name in enumerate(names)}
active = False
visited = set()
calls = collections.Counter()
table_reads = {d["name"]: set() for d in audit["data"]}
current = None
call_index = 0
events, writes, stack_writes, mutations = [], [], [], []
lookup_counts = collections.Counter()
caches = lambda: [rd(cache_base + 4 * i) for i in range(4)]

def mutate(effect):
    field, value = effect['field'], effect['value']
    if field == 'unit_type':
        address, raw = unit + 4, struct.pack('<H', value & 0xffff)
    elif field in ['unit_class', 'unit_subclass', 'unit_x', 'unit_y', 'primary_code', 'secondary_code']:
        address = unit + {'unit_class': 8, 'unit_subclass': 12, 'unit_x': 56, 'unit_y': 60, 'primary_code': 36, 'secondary_code': 40}[field]
        raw = struct.pack('<I', value & 0xffffffff)
    elif field == 'unit_direction':
        address, raw = unit + 124, struct.pack('<H', value & 0xffff)
    elif field == 'live_update':
        assert value in ['entry', 'other']
        address, raw = unit + 748, struct.pack('<I', other if value == 'other' else update)
    elif field in ['entry_status', 'other_status', 'entry_enemy', 'other_enemy']:
        record = update if field.startswith('entry') else other
        address = record + (1440 if field.endswith('status') else 1196)
        raw = struct.pack('<I', value if field.endswith('status') else target if value else 0)
    elif field == 'entry_action':
        address, raw = update + 552 + 24 * current['head'], struct.pack('<I', value)
    elif field in ['entry_sprite_frame', 'other_sprite_frame', 'entry_NPC_random', 'other_NPC_random']:
        record = update if field.startswith('entry') else other
        address, raw = record + (481 if field.endswith('sprite_frame') else 2094), bytes([value])
    elif field.startswith('cache_'):
        address, raw = cache_base + 4 * int(field.removeprefix('cache_')), struct.pack('<I', value)
    else:
        raise AssertionError(field)
    u.mem_write(address, raw)
    mutations.append({'field': field, 'value': value, 'call': call_index, 'address': hex(address), 'bytes': raw.hex()})

def service_return(value):
    pointer = u.reg_read(UC_X86_REG_ESP)
    u.reg_write(UC_X86_REG_EAX, value & 0xffffffff)
    u.reg_write(UC_X86_REG_ECX, 0x87654321)
    u.reg_write(UC_X86_REG_EDX, 0x98765432)
    u.reg_write(UC_X86_REG_ESP, pointer + 4)
    u.reg_write(UC_X86_REG_EIP, rd(pointer))

def on_code(_u, pc, size, _):
    if not active:
        return
    pointer = u.reg_read(UC_X86_REG_ESP)
    if pc == 0x4e3830:
        assert rd(pointer) == 0x4e3aaa
        name = bytes(u.mem_read(rd(pointer + 4), 64)).split(b'\0', 1)[0].decode('ascii')
        assert name in names
        lookup_counts[name] += 1
        event = {'service': 'dictionary4E3830', 'name': name, 'call': call_index, 'occurrence': lookup_counts[name], 'caches_before': caches()}
        effect = current.get('mutation')
        if effect and effect['name'] == name and effect.get('occurrence', 1) == lookup_counts[name]:
            for item in effect['effects']:
                mutate(item)
        found = current['found'][names.index(name)]
        event.update(found=found, index=current['lookup'][names.index(name)] if found else None, caches_after_declared_mutations=caches())
        events.append(event)
        service_return(records[name] if found else 0)
        return
    if pc in [0x4e5420, 0x40ef40]:
        length = rd(pointer + 12)
        assert length == (5 if pc == 0x4e5420 else 11)
        assert rd(pointer + 4) == current['player']
        if pc == 0x4e5420:
            assert rd(pointer + 16) == 0 and rd(pointer + 20) == 1
        events.append({'service': hex(pc), 'player': rd(pointer + 4), 'packet': bytes(u.mem_read(rd(pointer + 8), length)).hex(), 'call': call_index})
        service_return(current['queue_return'] if pc == 0x40ef40 else 1)
        return
    assert pc in instructions, (current['name'], hex(pc))
    assert bytes(u.mem_read(pc, size)) == instructions[pc]
    visited.add(pc)
    for f in audit['functions']:
        if pc == int(f['entry'], 16):
            calls[f['name']] += 1
            event = {'entry': hex(pc), 'call': call_index}
            if pc in [0x419a70, 0x4f9f90, 0x509e20]:
                event['argument'] = rd(pointer + 4)
            if pc == 0x415fa0:
                event.update(minimum=rd(pointer + 4), maximum=rd(pointer + 8), index_before=rd(rng_index))
            if pc in [0x534840, 0x534a10, 0x534a40]:
                event.update(caches_before=caches(), unit_type=struct.unpack('<H', u.mem_read(unit + 4, 2))[0], unit_class=rd(unit + 8), entry_status=rd(update + 1440), live_update='other' if rd(unit + 748) == other else 'entry')
            events.append(event)

def on_write(_u, _access, address, size, value, _):
    if not active:
        return
    if stack <= address and address + size <= stack + 0x10000:
        assert sp - 256 <= address and address + size <= sp + 16
        if address >= sp:
            assert (sp + 4 <= address and address + size <= sp + 8) or (sp + 12 <= address and address + size <= sp + 16)
        stack_writes.append((address, size, value))
        return
    assert size == 4 and address in [rng_index, *[cache_base + 4 * i for i in range(4)]], (current['name'], hex(address), size)
    writes.append({'address': hex(address), 'size': size, 'value': value & 0xffffffff, 'call': call_index})

def on_read(_u, _access, address, size, _value, _):
    if not active:
        return
    for d in audit['data']:
        start = int(d['address'], 16)
        table_reads[d['name']].update(range(max(address, start), min(address + size, start + d['size'])))

u.hook_add(UC_HOOK_CODE, on_code)
u.hook_add(UC_HOOK_MEM_WRITE, on_write)
u.hook_add(UC_HOOK_MEM_READ, on_read)

input_path = stage / 'phaseB-inputs-frozen-v1.json.gz'
assert sha(input_path.read_bytes()) == 'b7af66c39ffd09de1f23b194ce9ec9aee69e6be307b14ea8700b67bf5c5e4d48'
plan = json.loads(gzip.decompress(input_path.read_bytes()))
assert len(plan['rows']) == 15004
cases = plan['rows']
if args.limit:
    assert args.limit >= 2 and args.limit % 2 == 0
    cases = cases[:args.limit // 2] + cases[-args.limit // 2:]
rows = []
for frozen in cases:
    current = dict(frozen['network_input'], **frozen['selector_input'])
    current.update(name=frozen['name'], object_class=current['class'], object_flags=current['flags'])
    assert current['report'] == 0 and current['repeat'] in [1, 2]
    active = False
    u.mem_write(heap, b'\xa5' * 0x10000)
    for address, length in [(unit, 1024), (update, 4096), (other, 4096), (target, 1024)]:
        u.mem_write(address, bytes(length))
    u.mem_write(health, struct.pack('<HH', current['hp'], 65535))
    u.mem_write(unit + 4, struct.pack('<H', current['type_index']))
    for off, value in [(8, current['object_class']), (12, current['subclass']), (16, current['object_flags']), (36, current['primary_code']), (40, current['secondary_code']), (56, current['x']), (60, current['y']), (536, current['last_frame']), (556, health if current['hp_present'] else 0), (748, update)]:
        wr(unit + off, value)
    u.mem_write(unit + 124, struct.pack('<h', current['direction']))
    for address, status, enemy, frame, npc in [
        (update, current['status'], current['enemy'], current['sprite_frame'], current['NPC_random']),
        (other, current['other_status'], current['other_enemy'], current['other_sprite_frame'], current['other_NPC_random']),
    ]:
        u.mem_write(address + 544, struct.pack('<b', current['head']))
        wr(address + 552 + 24 * current['head'], current['action'])
        wr(address + 1440, status)
        wr(address + 1196, target if enemy else 0)
        u.mem_write(address + 481, bytes([frame]))
        u.mem_write(address + 2094, bytes([npc]))
        u.mem_write(address + 412 + 2 * current['player'], struct.pack('<H', current['previous_hp']))
    u.mem_write(cache_base - 64, b'\x6b' * 144)
    for i, name in enumerate(names):
        wr(cache_base + 4 * i, current['caches'][i])
        u.mem_write(records[name], struct.pack('<H', current['lookup'][i]))
    wr(rng_index, current['rng'])
    wr(game_frame, current['frame'])
    for i in range(53):
        wr(frame_table + 8 * i, current['player_frames'] if i == 50 else 0x12345678)
        wr(frame_table + 8 * i + 4, current['player_duration'] if i == 50 else 0x23456789)
    heap_before = bytearray(u.mem_read(heap, 0x10000))
    globals_before = bytearray(u.mem_read(cache_base - 64, 144))
    frames_before = bytes(u.mem_read(frame_table, 424))
    rng_neighbors = bytearray(u.mem_read(rng_index - 32, 68))
    events, writes, mutations, returns, fp_observations = [], [], [], [], []
    lookup_counts = collections.Counter()
    for call_index in range(current['repeat']):
        u.mem_write(stack, b'\xcc' * 0x10000)
        for off, value in [(0, stop), (4, current['player']), (8, unit), (12, current['report'])]:
            wr(sp + off, value)
        stack_before = bytearray(u.mem_read(stack, 0x10000))
        stack_writes = []
        for register, value in saved.items():
            u.reg_write(register, value)
        u.reg_write(UC_X86_REG_ESP, sp)
        u.emu_start(init, init + 8, count=4)
        before_fp = [u.reg_read(r) for r in [UC_X86_REG_FPCW, UC_X86_REG_FPSW, UC_X86_REG_FPTAG]]
        active = True
        u.emu_start(0x518960, stop, count=4000)
        active = False
        assert u.reg_read(UC_X86_REG_EIP) == stop and u.reg_read(UC_X86_REG_ESP) == sp + 4
        assert all(u.reg_read(r) == value for r, value in saved.items())
        after_fp = [u.reg_read(r) for r in [UC_X86_REG_FPCW, UC_X86_REG_FPSW, UC_X86_REG_FPTAG]]
        assert before_fp[0] == after_fp[0] == args.control and before_fp[2] == after_fp[2] and before_fp[1] & 0x3800 == after_fp[1] & 0x3800
        for address, size, value in stack_writes:
            stack_before[address - stack:address - stack + size] = (value & ((1 << (8 * size)) - 1)).to_bytes(size, 'little')
        assert bytes(u.mem_read(stack, 0x10000)) == stack_before
        returns.append(u.reg_read(UC_X86_REG_EAX))
        fp_observations.append({'before': before_fp, 'after': after_fp})
    for effect in mutations:
        address = int(effect['address'], 16)
        raw = bytes.fromhex(effect['bytes'])
        if heap <= address < heap + 0x10000:
            heap_before[address - heap:address - heap + len(raw)] = raw
    assert bytes(u.mem_read(heap, 0x10000)) == heap_before
    for i, value in enumerate(caches()):
        struct.pack_into('<I', globals_before, 64 + 4 * i, value)
    assert bytes(u.mem_read(cache_base - 64, 144)) == globals_before
    struct.pack_into('<I', rng_neighbors, 32, rd(rng_index))
    assert bytes(u.mem_read(rng_index - 32, 68)) == rng_neighbors
    assert rd(game_frame) == current['frame'] and bytes(u.mem_read(frame_table, 424)) == frames_before
    assert len([e for e in events if e.get('service') == '0x40ef40']) == current['repeat']
    final_unit = {'type': struct.unpack('<H', u.mem_read(unit + 4, 2))[0], 'class': rd(unit + 8), 'subclass': rd(unit + 12), 'flags': rd(unit + 16), 'primary_code': rd(unit + 36), 'secondary_code': rd(unit + 40), 'x_bits': rd(unit + 56), 'y_bits': rd(unit + 60), 'direction': struct.unpack('<h', u.mem_read(unit + 124, 2))[0], 'live_update': 'other' if rd(unit + 748) == other else 'entry'}
    final_updates = []
    for address in [update, other]:
        final_updates.append({'status': rd(address + 1440), 'enemy': bool(rd(address + 1196)), 'action': rd(address + 552 + 24 * current['head']), 'sprite_frame': u.mem_read(address + 481, 1)[0], 'NPC_random': u.mem_read(address + 2094, 1)[0]})
    rows.append({'name': frozen['name'], 'input': {'selector_input': frozen['selector_input'], 'network_input': frozen['network_input']}, 'output': {'returns': returns, 'events': events, 'writes': writes, 'mutations': mutations, 'unit_after': final_unit, 'updates_after': final_updates, 'caches_after': caches(), 'RNG_index_after': rd(rng_index), 'FPU': fp_observations, 'full_heap_stack_global_neighbors_nonvolatile_and_FP_control_guards_preserved': True}})

coverage = {f['name']: sum(len(instructions[a]) for a in visited if int(f['entry'], 16) <= a < int(f['entry'], 16) + f['size']) for f in audit['functions']}
for f in audit['functions']:
    assert sha(bytes(u.mem_read(int(f['entry'], 16), f['size']))) == f['sha256']
for d in audit['data']:
    assert sha(bytes(u.mem_read(int(d['address'], 16), d['size']))) == d['sha256']
assert sha(args.game.read_bytes()) == sha(blob)
assert all(row['input'] == {'selector_input': source['selector_input'], 'network_input': source['network_input']} for row, source in zip(rows, cases))
fixture = {'schema': 'original-complex-object-send-phaseB-v1', 'original_sha256': sha(blob), 'control_word': args.control, 'input_fixture_sha256': sha(input_path.read_bytes()), 'declared_services': ['dictionary4E3830', 'packet-send4E5420', 'server-message-list40EF40'], 'rows': rows}
raw = json.dumps(fixture, sort_keys=True, separators=(',', ':')).encode()
with args.fixture.open('xb') as output:
    with gzip.GzipFile(filename='', fileobj=output, mode='wb', mtime=0) as compressed:
        compressed.write(raw)
proof = {'status': 'executed', 'control_word': hex(args.control), 'contexts': len(rows), 'root_calls': sum(len(row['output']['returns']) for row in rows), 'pilot_limit': args.limit, 'producer_sha256': sha(Path(__file__).read_bytes()), 'input_fixture_sha256': sha(input_path.read_bytes()), 'static_dependency_sha256': sha((stage / 'original-functions-and-tables-static-v1.json').read_bytes()), 'coverage_bytes': coverage, 'executed_bytes': sum(coverage.values()), 'static_declared_bytes_not_execution': 1192, 'actual_entries': dict(calls), 'table_bytes_read': {k: len(v) for k, v in table_reads.items()}, 'fixture_sha256': sha(args.fixture.read_bytes()), 'raw_fixture_sha256': sha(raw), 'original_unchanged': True, 'all_frozen_inputs_unchanged': True, 'actual_WORD_helper_dictionary_predicates_execute': True, 'native_before_and_comparators_still_pending': True, 'overall_porting_100_percent': False}
with args.proof.open('x') as output:
    json.dump(proof, output, indent=2)
    output.write('\n')
print(json.dumps(proof), flush=True)
