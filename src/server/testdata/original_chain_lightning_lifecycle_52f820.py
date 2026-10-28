"""Execute unchanged GAME.EXE Chain Lightning create/destroy instructions.

Cancel, point FX and recursive-free are declared services. This does not
qualify the update routine's arithmetic or transitive service implementations.
"""
import argparse
import collections
import copy
import hashlib
import json
from pathlib import Path
import struct

from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import *

parser = argparse.ArgumentParser(description=__doc__)
for key in ('game', 'fixture', 'proof'):
    parser.add_argument('--' + key, type=Path, required=True)
args = parser.parse_args()
blob = args.game.read_bytes()
sha = lambda data: hashlib.sha256(data).hexdigest()
assert sha(blob) == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe = struct.unpack_from('<I', blob, 60)[0]
opt = pe + 24
base = struct.unpack_from('<I', blob, opt + 28)[0]
size = (struct.unpack_from('<I', blob, opt + 56)[0] + 4095) & ~4095
table = opt + struct.unpack_from('<H', blob, pe + 20)[0]
u = Uc(UC_ARCH_X86, UC_MODE_32)
u.mem_map(base, size)
for n in range(struct.unpack_from('<H', blob, pe + 6)[0]):
    _, va, count, offset = struct.unpack_from('<IIII', blob, table + 40*n + 8)
    if count:
        u.mem_write(base + va, blob[offset:offset + count])
stack, heap, stop = 0x20000000, 0x21000000, 0x30000000
for address in (stack, heap, stop):
    u.mem_map(address, 0x10000)
sp, record = stack + 0xe000, heap + 0x1000
objects = [heap + 0x2000 + 1024*i for i in range(4)]
players = [heap + 0x6000 + 512*i for i in range(2)]
wands = [heap + 0x7000 + 256*i for i in range(2)]
rays = [heap + 0x8000 + 256*i for i in range(4)]
get = lambda p: struct.unpack('<I', u.mem_read(p, 4))[0]
put = lambda p, value: u.mem_write(p, struct.pack('<I', value & 0xffffffff))
ref = lambda values, i: 0 if i < 0 else values[i]
index = lambda values, p: -1 if p == 0 else values.index(p)
specs = {'create': (0x52f820, 0x52f8a0), 'destroy': (0x530100, 0x530160)}
traces = {kind: {} for kind in specs}
entries = collections.Counter()
events, current, mutations = [], {}, collections.Counter()
did_mutate = False


def state(final=False):
    value = dict(caster=index(objects, get(record+16)), flag=get(record+20),
                 pos=[get(record+28), get(record+32)], carry=get(record+76), flags=get(record+88),
                 sub104=index(rays, get(record+104)), sub108=index(rays, get(record+108)),
                 classes=[get(p+8) for p in objects[:2]],
                 equipped=[index(objects, get(p+104)) for p in players],
                 subclasses=[get(p+12) for p in objects[2:]],
                 wand_flags=[get(p+96) for p in wands],
                 next=[index(rays, get(p+116)) for p in rays])
    if not (final and current['kind'] == 'destroy'):
        value['stored'] = index(objects, get(record+72))
    return value


def mutate(service, ray=None):
    global did_mutate
    mutation = current['mutation']
    if did_mutate or not mutation.startswith(service + '-'):
        return
    did_mutate = True
    mutations[mutation] += 1
    if mutation == 'cancel-caster-b': put(record+16, objects[1])
    elif mutation == 'cancel-trap-on': put(record+20, 1)
    elif mutation == 'cancel-trap-off': put(record+20, 0)
    elif mutation == 'cancel-equipped-none': put(players[0]+104, 0)
    elif mutation == 'cancel-equipped-b': put(players[0]+104, objects[3])
    elif mutation == 'cancel-clear-flag': put(wands[0]+96, get(wands[0]+96) & ~4)
    elif mutation == 'cancel-pos':
        put(record+28, 0xc49a5000); put(record+32, 0x447a1000)
    elif mutation == 'cancel-flags': put(record+88, 0xa1b2c301)
    elif mutation == 'free-next-nil': put(ray+116, 0)
    elif mutation == 'free-old-head': put(record+104, rays[3])
    elif mutation == 'free-weapon-b': put(record+72, objects[3])
    elif mutation == 'free-weapon-nil': put(record+72, 0)
    elif mutation == 'free-flags': put(wands[0]+96, 0xf1223344)
    else: raise AssertionError(mutation)


def ret(value=0):
    esp = u.reg_read(UC_X86_REG_ESP)
    u.reg_write(UC_X86_REG_EAX, value & 0xffffffff)
    u.reg_write(UC_X86_REG_EIP, get(esp))
    u.reg_write(UC_X86_REG_ESP, esp+4)


def hook(_u, address, length, _):
    if address == stop:
        u.emu_stop()
        return
    lo, hi = specs[current['kind']]
    if lo <= address < hi:
        traces[current['kind']][address] = bytes(u.mem_read(address, length))
        if address == lo: entries[current['kind']] += 1
        return
    esp = u.reg_read(UC_X86_REG_ESP)
    arg = lambda n: get(esp+4*n)
    if address == 0x4feb10:
        assert arg(1) == 24
        events.append(dict(kind='cancel', args=[24, index(objects, arg(2))], state=state()))
        mutate('cancel')
        ret(0x87654321)
    elif address == 0x522ff0:
        events.append(dict(kind='fx', args=[arg(1), get(arg(2)), get(arg(2)+4)], state=state()))
        ret(0x98765432)
    elif address == 0x4fe980:
        events.append(dict(kind='free', args=[index(rays, arg(1))], state=state()))
        mutate('free', arg(1))
        ret(0xaabbccdd)
    else:
        raise AssertionError((current['name'], 'undeclared instruction/service', hex(address)))


faults = []
def invalid(_u, access, address, size, value, _):
    faults.append(dict(name=current['name'], pc=hex(u.reg_read(UC_X86_REG_EIP)), address=hex(address)))
    return False


u.hook_add(UC_HOOK_CODE, hook)
u.hook_add(UC_HOOK_MEM_INVALID, invalid)
seed = bytes((i*31+19) & 255 for i in range(0x10000))
registers = {UC_X86_REG_EBX: 0x13572468, UC_X86_REG_ESI: 0x24681357,
             UC_X86_REG_EDI: 0x98765432, UC_X86_REG_EBP: 0xfedcba98}


def run(case):
    global current, events, did_mutate
    current, events, did_mutate = case, [], False
    u.mem_write(heap, seed)
    for offset, value in [(16, ref(objects, case['caster'])), (20, case['flag']),
                          (28, case['pos'][0]), (32, case['pos'][1]),
                          (72, ref(objects, case['stored'])), (76, case['carry']), (88, case['flags'])]:
        put(record+offset, value)
    for i in range(2):
        put(objects[i]+8, case['classes'][i]); put(objects[i]+748, players[i])
        put(players[i]+104, ref(objects, case['equipped'][i]))
        put(objects[i+2]+12, case['subclasses'][i]); put(objects[i+2]+736, wands[i])
        put(wands[i]+96, case['wand_flags'][i])
    for p in rays: put(p+116, 0)
    for offset, chain in [(104, case['sub104']), (108, case['sub108'])]:
        put(record+offset, ref(rays, chain[0]) if chain else 0)
        for i, node in enumerate(chain):
            put(rays[node]+116, ref(rays, chain[i+1]) if i+1 < len(chain) else 0)
    before = bytes(u.mem_read(heap, len(seed)))
    u.mem_write(sp-128, bytes(256))
    put(sp, stop); put(sp+4, record)
    for register, value in registers.items(): u.reg_write(register, value)
    u.reg_write(UC_X86_REG_EFLAGS, 0x202)
    u.reg_write(UC_X86_REG_ESP, sp)
    u.emu_start(specs[case['kind']][0], stop+1, count=10000)
    assert u.reg_read(UC_X86_REG_EIP) == stop and u.reg_read(UC_X86_REG_ESP) == sp+4
    assert get(sp) == stop and get(sp+4) == record
    for register, value in registers.items(): assert u.reg_read(register) == value
    after = bytes(u.mem_read(heap, len(seed)))
    # Only documented outputs and declared callback mutations may change.
    allowed = set()
    for offset in ([72, 76, 88] if case['kind'] == 'create' else [104, 108]):
        allowed.update(range(record-heap+offset, record-heap+offset+4))
    mutation = case['mutation']
    ranges = {'cancel-caster-b': [(record+16, 4)], 'cancel-trap-on': [(record+20, 4)],
              'cancel-trap-off': [(record+20, 4)], 'cancel-equipped-none': [(players[0]+104, 4)],
              'cancel-equipped-b': [(players[0]+104, 4)], 'cancel-clear-flag': [(wands[0]+96, 4)],
              'cancel-pos': [(record+28, 8)], 'cancel-flags': [(record+88, 4)],
              'free-next-nil': [(rays[case['sub108'][0] if case['sub108'] else 0]+116, 4)],
              'free-old-head': [(record+104, 4)], 'free-weapon-b': [(record+72, 4)],
              'free-weapon-nil': [(record+72, 4)], 'free-flags': [(wands[0]+96, 4)]}
    for address, count in ranges.get(mutation, []): allowed.update(range(address-heap, address-heap+count))
    if case['kind'] == 'destroy':
        for address in wands: allowed.update(range(address-heap+96, address-heap+100))
    assert all(a == b or i in allowed for i, (a, b) in enumerate(zip(before, after)))
    result = dict(ret=u.reg_read(UC_X86_REG_EAX) & (255 if case['kind'] == 'destroy' else 0xffffffff),
                  state=state(final=True), events=copy.deepcopy(events))
    if case['mutation'] != 'none': assert did_mutate
    return dict(input=copy.deepcopy(case), result=result)


template = dict(name='', kind='create', mutation='none', caster=0, flag=0,
                classes=[0x280004, 4], equipped=[2, 3], subclasses=[0x40000, 0x40000],
                wand_flags=[0xaabbccdd, 0x123456ff], stored=-1, carry=0x7fc12345,
                flags=0x77556601, pos=[0x3f923456, 0xc2334567], sub104=[], sub108=[])
cases = []
def add(name, **changes):
    case = copy.deepcopy(template); case.update(name=name, **changes); cases.append(case)

for cls in [4, 0x280004, 2, 0x280002, 0x20]:
    for flag in [0, 1]:
        add(f'create/class-{cls:x}/trap-{flag}', classes=[cls, 4], flag=flag)
add('create/no-caster-trap', caster=-1, flag=1)
add('create/no-equipped', equipped=[-1, 3])
add('create/wrong-subclass', subclasses=[0, 0x40000])
add('create/inactive-wand', wand_flags=[0xfffffffb, 0x123456ff])
for low in [0, 1, 2, 0xfc, 0xff]: add(f'create/flags-{low:x}', flags=0xfedcba00 | low)
for mutation in ['cancel-caster-b', 'cancel-trap-on', 'cancel-trap-off', 'cancel-equipped-none',
                 'cancel-equipped-b', 'cancel-clear-flag', 'cancel-pos', 'cancel-flags']:
    add('create/'+mutation, mutation=mutation, flag=1 if mutation == 'cancel-trap-off' else 0)
for stored in [2, 3]:
    add(f'create/reused-{stored}/no-wand', stored=stored, equipped=[-1, 3])
    add(f'create/reused-{stored}/wand', stored=stored)
    add(f'create/reused-{stored}/trap', stored=stored, flag=1)
for generation in ['empty', 'old', 'new', 'both']:
    for stored in [-1, 2, 3]:
        add(f'destroy/{generation}/wand-{stored}', kind='destroy', stored=stored,
            sub104=[2, 3] if generation in ('old', 'both') else [],
            sub108=[0, 1] if generation in ('new', 'both') else [])
for flags in [0, 4, 0x80, 0xff, 0x12345678, 0xffffffff]:
    add(f'destroy/flags-{flags:x}', kind='destroy', stored=2, wand_flags=[flags, 0x123456ff])
for mutation in ['free-next-nil', 'free-old-head', 'free-weapon-b', 'free-weapon-nil', 'free-flags']:
    add('destroy/'+mutation, kind='destroy', stored=2, sub108=[0, 1], sub104=[2, 3], mutation=mutation)
rows = [run(case) for case in cases]
regions = []
for kind, instructions in traces.items():
    lo = specs[kind][0]; hi = max(pc+len(code) for pc, code in instructions.items())
    covered = {pc+i for pc, code in instructions.items() for i in range(len(code))}
    assert covered == set(range(lo, hi)), (kind, sorted(set(range(lo, hi))-covered))
    regions.append(dict(name=kind, address=hex(lo), end_exclusive=hex(hi), bytes=hi-lo,
                        sha256=sha(bytes(u.mem_read(lo, hi-lo))),
                        instructions=[dict(pc=hex(pc), bytes=code.hex()) for pc, code in sorted(instructions.items())]))
fixture = dict(original_sha256=sha(blob), rows=rows)
encoded = (json.dumps(fixture, indent=2)+'\n').encode()
args.fixture.write_bytes(encoded)
proof = dict(original_sha256=sha(blob), contexts=len(rows), original_executions=len(rows), regions=regions,
             all_original_instruction_bytes_executed=True, zero_arithmetic_or_conversion_replacements=True,
             callee_saved_registers_stack_and_nonoutput_heap_preserved=True,
             entries=dict(entries), declared_services=['cancel24', 'pointFX', 'recursive-free'],
             service_mutations=dict(mutations), faults=faults, fixture_sha256=sha(encoded),
             pointer_projection='PE32 record72 wand identity maps to a native sidecar; native record72 stays0 and the sidecar is released after destroy. Original record72 is excluded from the final destroy projection only.',
             scope='Complete create/destroy roots, declared services. Update arithmetic, recursive free internals and full gameplay are separate evidence.')
assert not faults
args.proof.write_text(json.dumps(proof, indent=2)+'\n')
print(json.dumps({k:v for k,v in proof.items() if k != 'regions'}, indent=2))
