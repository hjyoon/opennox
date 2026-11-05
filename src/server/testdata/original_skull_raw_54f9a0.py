"""Execute all five original raw Skull bodies; no native implementation is read."""
import argparse
import collections
import gzip
import hashlib
import json
import struct
from pathlib import Path

import capstone
from unicorn import (Uc, UcError, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE,
                     UC_HOOK_MEM_WRITE, UC_HOOK_MEM_INVALID)
from unicorn.x86_const import *

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--exe', type=Path, required=True)
parser.add_argument('--control', type=lambda s: int(s, 0), choices=[0xe3f, 0x23f], required=True)
parser.add_argument('--output', type=Path, required=True)
parser.add_argument('--proof', type=Path, required=True)
parser.add_argument('--filter', default='')
options = parser.parse_args()
sha = lambda b: hashlib.sha256(b).hexdigest()
packed = Path(__file__).with_name('skull_raw_independent_inputs_54f9a0.json.gz').read_bytes()
assert sha(packed) == '2fa912b92071a20903b0596adce24a80942874efeb6e80fe5b8d8f6def6dcc36'
inputs = json.loads(gzip.decompress(packed))
assert len(inputs) == 753
selected = [i for i in inputs if not options.filter or options.filter in i['name']]
assert selected
raw = options.exe.read_bytes()
assert sha(raw) == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe = struct.unpack_from('<I', raw, 60)[0]
opt = pe + 24
base = struct.unpack_from('<I', raw, opt + 28)[0]
image_size = (struct.unpack_from('<I', raw, opt + 56)[0] + 4095) & ~4095
table = opt + struct.unpack_from('<H', raw, pe + 20)[0]
image = bytearray(image_size)
for n in range(struct.unpack_from('<H', raw, pe + 6)[0]):
    _, va, size, offset = struct.unpack_from('<IIII', raw, table + 40*n + 8)
    if size:
        image[va:va+size] = raw[offset:offset+size]

heap, stack, aux = 0x21000000, 0x22000000, 0x23000000
roles = [heap + 0x1000*(n+1) for n in range(8)]
sp, stop, initialize, cw, resume, float_return, float_value = [
    stack+0xe000, aux+0x100, aux+0x200, aux+0x300, aux+0x400, aux+0x500, aux+0x600]
globals_at, fps_at = [0x834b08 + 4*n for n in range(6)], 0x85b3fc
specs = [
    (0x54f9a0, 215, 'b79f13c5294ff503c6c585b8c967c3e3b463019bd943a1e60a2065db0b47c862'),
    (0x54fa80, 300, 'c580ca147c534ab48e9186e4ad3f126f9a513a31bde673b0e45a29fdc8309bdd'),
    (0x54fbb0, 51, '42356876554beba9c143631ba4fdc958744d9a73fe567ff437c1d75423e0d5e0'),
    (0x54fbf0, 94, '215b664efef388461519e9d1e18a74d13ec447b2a1b7d7574a05fc32e82cc9d8'),
    (0x54fc50, 91, 'e4f6e35e82ae0c21b86ac7fc68dc9c9cb89ee879116cdd01d4bb14a262421585'),
    (0x419a70, 17, '10e536e2a987ea0fe14c91f007b2d5361cd87d4f0ca4cf5385fc82e53913d186'),
]
md = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_32)
md.detail = True
instructions, ranges, guest_stores = {}, [], set()
for address, size, digest in specs:
    body = bytes(image[address-base:address-base+size])
    assert sha(body) == digest
    decoded = list(md.disasm(body, address))
    assert sum(i.size for i in decoded) == size and decoded[-1].mnemonic == 'ret'
    instructions.update({i.address: bytes(i.bytes) for i in decoded})
    guest_stores.update(i.address for i in decoded for op in i.operands
                        if op.type == capstone.x86.X86_OP_MEM and op.access & capstone.CS_AC_WRITE)
    ranges.append(dict(address=hex(address), size=size, sha256=digest))
# Capstone 5.0.6 omits write-access metadata on these actual x87 FSTPs.
# Declare only their exact original byte sequences; keep all store guards.
for address, opcode in [(0x54fb03, 'd95e50'), (0x54fb17, 'd95e54')]:
    assert instructions[address].hex() == opcode
    guest_stores.add(address)
services = {0x4e3aa0: ('name', 1), 0x4e3450: ('new', 1),
            0x4daa50: ('create', 5), 0x419d40: ('damage', 1),
            0x501960: ('audio', 4), 0x5238a0: ('fx', 2),
            0x517c10: ('enumerate', 3), 0x5330c0: ('enemy', 2),
            0x537110: ('map', 2), 0x4e6e50: ('front', 3)}
saved = {UC_X86_REG_EBX: 0x12345678, UC_X86_REG_ESI: 0x23456789,
         UC_X86_REG_EDI: 0x3456789a, UC_X86_REG_EBP: 0x456789ab}
seen, calls, faults = set(), collections.Counter(), collections.Counter()
reentries = conversions = callbacks = 0
active = False


def pointer(n):
    if n == -1:
        return 0
    if n == -2:
        return 0x41000000
    assert 0 <= n < len(roles)
    return roles[n]


def role(p):
    found = next((n for n in range(-2, 8) if pointer(n) == p), None)
    assert found is not None, (current['name'], 'undeclared object/data pointer', hex(p))
    return found


def point_role(p):
    return role((p-56) & 0xffffffff)


def get(p):
    return struct.unpack('<I', u.mem_read(p, 4))[0]


def registers():
    return {r: u.reg_read(r) for r in saved}


def write(p, data):
    u.mem_write(p, data)
    if heap <= p and p+len(data) <= heap+0x10000:
        expected_heap[p-heap:p-heap+len(data)] = data
    elif base <= p and p+len(data) <= base+image_size:
        expected_image[p-base:p-base+len(data)] = data
    else:
        raise AssertionError(('undeclared world write', hex(p)))


def put(p, value, width=4):
    write(p, int(value & ((1 << (8*width))-1)).to_bytes(width, 'little'))


def state():
    return dict(caches=[get(p) for p in globals_at], fps=get(fps_at))


def frame_state():
    return dict(**state(), heap_sha256=sha(expected_heap))


def mutate(plan):
    for key, width in [('words', 4), ('words16', 2), ('bytes', 1)]:
        for target, offset, value in plan.get(key, []):
            put(pointer(target)+offset, value, width)
    for target, offset, value in plan.get('pointers', []):
        put(pointer(target)+offset, pointer(value))
    for index, value in plan.get('globals', []):
        put(globals_at[index], value)
    if 'fps' in plan:
        put(fps_at, plan['fps'])


def entry_args(plan):
    entry = plan['entry']
    unit = plan.get('unit', current['unit'])
    target = plan.get('target', current['target'])
    projectile_type = plan.get('projectile_type', current['projectile_type'])
    if entry == '54fc50':
        return [pointer(target), pointer(unit)]
    if entry == '54fa80':
        return [pointer(unit), projectile_type]
    assert entry in ['54f9a0', '54fbb0', '54fbf0']
    return [pointer(unit)]


def advance(frame):
    global callbacks
    if frame['actions']:
        action = frame['actions'].pop(0)
        if action['kind'] == 'candidate':
            callbacks += 1
        frame['action'] = action
        frames.append(frame)
        values = action['args']
        new_sp = frame['esp'] - 4*(len(values)+1)
        u.mem_write(new_sp, struct.pack('<'+'I'*(len(values)+1), resume, *values))
        u.reg_write(UC_X86_REG_ESP, new_sp)
        u.reg_write(UC_X86_REG_EIP, int(action['entry'], 16))
        events.append(dict(kind=action['kind'], index=frame['index'], entry=action['entry'],
                           arguments=action['semantic'], **frame_state()))
        return
    assert registers() == frame['registers']
    assert u.reg_read(UC_X86_REG_FPSW) & 0x3800 == 0
    events.append(dict(kind=frame['kind']+'-return', index=frame['index'],
                       result=frame['semantic'], **frame_state()))
    u.reg_write(UC_X86_REG_ESP, frame['esp'])
    if frame['kind'] == 'damage':
        u.mem_write(float_value, struct.pack('<I', frame['result']))
        u.reg_write(UC_X86_REG_EIP, float_return)
    else:
        u.reg_write(UC_X86_REG_EAX, frame['result'])
        u.reg_write(UC_X86_REG_ESP, frame['esp']+4)
        u.reg_write(UC_X86_REG_EIP, get(frame['esp']))


def on_code(_u, pc, size, _):
    global reentries, conversions, callbacks
    if not active:
        return
    esp = u.reg_read(UC_X86_REG_ESP)
    if pc in [float_return, float_return+6]:
        return
    if pc == resume:
        frame = frames.pop()
        action = frame['action']
        assert esp == frame['esp']-4*len(action['args'])
        assert registers() == frame['registers']
        result = u.reg_read(UC_X86_REG_EAX) if action['entry'] in ['54f9a0', '54fbf0'] else None
        events.append(dict(kind=action['kind']+'-return', index=frame['index'],
                           result=result, **frame_state()))
        advance(frame)
        return
    if pc not in services:
        assert pc in instructions and bytes(u.mem_read(pc, size)) == instructions[pc], (current['name'], hex(pc), size)
        seen.add(pc)
        if pc == 0x419a70:
            conversions += 1
        return
    kind, count = services[pc]
    args = [get(esp+4*n) for n in range(1, count+1)]
    occurrence = occurrences.get(kind, 0)
    occurrences[kind] = occurrence+1
    calls[kind] += 1
    index = sum(1 for e in events if e['kind'] in {s[0] for s in services.values()})
    plan = current['script'].get(f'{kind}:{occurrence}', {})
    event = dict(kind=kind, index=index, **frame_state())
    result, semantic = 0xa5a50000+index, None
    actions = []
    if kind in ['name', 'damage']:
        name = bytes(u.mem_read(args[0], 64)).split(b'\0', 1)[0].decode('ascii')
        assert name in (['MercArcherArrow', 'ArrowTrap1', 'ArrowTrap2'] if kind == 'name' else ['ArrowTrapDamage'])
        event['name'] = name
        result = semantic = plan.get('result', current['name_results'][name] if kind == 'name' else current['damage_bits'])
    elif kind == 'new':
        event['type_id'] = args[0]
        semantic = plan.get('result', current['projectile'])
        result = pointer(semantic)
    elif kind == 'create':
        assert args[4] == 0
        event.update(object=role(args[0]), owner=role(args[1]), position=args[2:4], extra=0)
    elif kind == 'audio':
        assert args[0] == 889 and args[2:] == [0, 0]
        event.update(sound=args[0], unit=role(args[1]))
    elif kind == 'fx':
        assert args[1] in [1, 2]
        event.update(position_owner=point_role(args[0]), position=[get(args[0]), get(args[0]+4)], variant=args[1])
    elif kind == 'enumerate':
        assert args[1] == 0x54fc50
        unit = role(args[2])
        event.update(rectangle=[get(args[0]+4*n) for n in range(4)], source=unit, callback='54fc50')
        for target in plan.get('targets', current['targets']):
            actions.append(dict(kind='candidate', entry='54fc50', args=[pointer(target), args[2]], semantic=[target, unit]))
    elif kind in ['enemy', 'map']:
        event['objects'] = list(map(role, args))
        result = semantic = plan.get('result', current['enemy' if kind == 'enemy' else 'map_check'])
    elif kind == 'front':
        event.update(source=point_role(args[0]), direction=args[1], target=point_role(args[2]),
                     source_position=[get(args[0]), get(args[0]+4)], target_position=[get(args[2]), get(args[2]+4)])
        result = semantic = plan.get('result', current['in_front'])
    else:
        raise AssertionError(kind)
    events.append(event)
    mutate(plan)
    if 'reenter' in plan:
        nested = plan['reenter']
        actions.insert(0, dict(kind='nested', entry=nested['entry'], args=entry_args(nested), semantic=nested))
        reentries += 1
    frame = dict(kind=kind, index=index, esp=esp, result=result, semantic=semantic,
                 actions=actions, registers=registers())
    advance(frame)


def on_write(_u, access, address, size, value, _):
    if not active:
        return
    if stack <= address and address+size <= stack+0x10000:
        assert address+size <= sp, (current['name'], 'caller write', hex(address))
        return
    pc = u.reg_read(UC_X86_REG_EIP)
    assert pc in guest_stores and size in [1, 2, 4], (current['name'], 'undeclared store', hex(pc), size)
    data = int(value & ((1 << (8*size))-1)).to_bytes(size, 'little')
    if heap <= address and address+size <= heap+0x10000:
        expected_heap[address-heap:address-heap+size] = data
    elif address in globals_at and size == 4:
        expected_image[address-base:address-base+size] = data
    else:
        assert address < 0x10000 or 0x41000000 <= address < 0x41010000, (current['name'], 'write outside declared domain', hex(address))


fault_sites = {0x54f9a6: 'update-source-flags', 0x54f9a9: 'update-source-data-pointer',
               0x54f9b6: 'update-disabled-byte', 0x54f9bd: 'update-enabled-byte',
               0x54f9d1: 'update-scan-timer', 0x54f9eb: 'update-state-byte',
               0x54fa8e: 'projectile-source-radius', 0x54fa94: 'projectile-source-direction',
               0x54faea: 'projectile-direction-store', 0x54fb86: 'projectile-damage-store',
               0x54fbb6: 'wrapper-source-data-pointer', 0x54fbc8: 'wrapper-state-byte',
               0x54fbdd: 'wrapper-state-clear', 0x54fbfd: 'scan-source-x',
               0x54fc56: 'candidate-class', 0x54fc5c: 'candidate-flags', 0x54fc85: 'candidate-source-direction'}


def on_fault(_u, access, address, size, value, _):
    global fault, fault_detail
    pc = u.reg_read(UC_X86_REG_EIP)
    assert pc in fault_sites, (current['name'], 'undeclared fault site', hex(pc), hex(address))
    fault = dict(kind='memory', site=fault_sites[pc])
    fault_detail = dict(pc=hex(pc), address=hex(address), access=access, size=size)
    return False


outputs, fault_details = [], []
for current in selected:
    active = False
    u = Uc(UC_ARCH_X86, UC_MODE_32)
    for address, size in [(base, image_size), (heap, 0x10000), (stack, 0x10000), (aux, 0x10000)]:
        u.mem_map(address, size)
    expected_image = bytearray(image)
    u.mem_write(base, bytes(image))
    expected_heap = bytearray((n*17+(n >> 7)*29) & 255 for n in range(0x10000))
    u.mem_write(heap, bytes(expected_heap))
    for n in range(4):
        target_flags = current['target_flags'] if n == 1 else current['flags']
        for offset, value, width in [(4, current['source_type'], 2), (8, current['target_class'], 4),
                                     (16, target_flags, 4), (56, current['position'][0]+n, 4),
                                     (60, current['position'][1]+n, 4), (124, current['direction'], 2),
                                     (176, current['radius_bits'], 4), (544, current['speed_bits'], 4),
                                     (700, pointer(current['collide']), 4),
                                     (748, pointer(6 if n == 3 else current['update']), 4)]:
            put(roles[n]+offset, value, width)
    for n, offset, value, width in [(4, 0, current['timers'][0], 4), (4, 4, current['timers'][1], 4),
                                     (4, 8, current['state'], 1), (4, 12, current['projectile_type'], 4),
                                     (4, 48, current['enabled'], 1), (6, 0, 17, 4),
                                     (6, 4, 19, 4), (6, 8, 1, 1), (6, 12, 73, 4), (6, 48, 1, 1)]:
        put(roles[n]+offset, value, width)
    for p, value in zip(globals_at, current['caches']):
        put(p, value)
    put(fps_at, current['fps'])
    u.mem_write(stack, b'\xcc'*0x10000)
    arguments = entry_args(current)
    caller = struct.pack('<'+'I'*(len(arguments)+1), stop, *arguments)+bytes(range(48))
    u.mem_write(sp, caller)
    u.mem_write(initialize, b'\xdb\xe3\xd9\x2d'+struct.pack('<I', cw))
    u.mem_write(cw, struct.pack('<H', options.control))
    u.mem_write(resume, b'\x90')
    u.mem_write(float_return, b'\xd9\x05'+struct.pack('<I', float_value)+b'\xc3')
    for key, value in saved.items():
        u.reg_write(key, value)
    u.reg_write(UC_X86_REG_ESP, sp)
    u.emu_start(initialize, initialize+8, count=2)
    for hook, callback in [(UC_HOOK_CODE, on_code), (UC_HOOK_MEM_WRITE, on_write), (UC_HOOK_MEM_INVALID, on_fault)]:
        u.hook_add(hook, callback)
    events, frames, occurrences = [], [], {}
    fault, fault_detail = None, None
    active = True
    try:
        u.emu_start(int(current['entry'], 16), stop, count=100000)
    except UcError:
        assert fault is not None, current['name']
    active = False
    assert bytes(u.mem_read(heap, 0x10000)) == expected_heap, (current['name'], 'complete heap')
    assert bytes(u.mem_read(base, image_size)) == expected_image, (current['name'], 'complete image')
    assert bytes(u.mem_read(sp, len(caller))) == caller, (current['name'], 'caller guard')
    assert u.reg_read(UC_X86_REG_FPCW) == options.control
    if fault:
        faults[fault['site']] += 1
        fault_details.append(dict(name=current['name'], **fault_detail))
    else:
        assert not frames and u.reg_read(UC_X86_REG_EIP) == stop, current['name']
        assert u.reg_read(UC_X86_REG_ESP) == sp+4 and registers() == saved
        assert u.reg_read(UC_X86_REG_FPSW) & 0x3800 == 0
    result = u.reg_read(UC_X86_REG_EAX) if not fault and current['entry'] in ['54f9a0', '54fbf0'] else None
    outputs.append(dict(name=current['name'], result=result, fault=fault, events=events, **frame_state()))

complete = instructions.keys() <= seen
if not options.filter:
    assert complete, [hex(pc) for pc in instructions.keys()-seen]
decoded = json.dumps(outputs, sort_keys=True, separators=(',', ':')).encode()+b'\n'
packed_output = bytearray(gzip.compress(decoded, compresslevel=9, mtime=0))
packed_output[9] = 0xff
with options.output.open('xb') as stream:
    stream.write(packed_output)
proof = dict(original_exe_sha256=sha(raw), input_sha256=sha(packed), cases=len(selected),
             control=hex(options.control), ranges=ranges, full_753_input_set=not options.filter,
             all768_instruction_bytes_executed=complete, new_body_bytes=751, preexisting_converter_bytes=17,
             full_heap_image_caller_guards=True, valid_return_stack_nonvolatile_and_FPU_preserved=True,
             declared_world_services={hex(pc): s[0] for pc, s in services.items()},
             reentrant_original_calls=reentries, actual_original_candidate_calls=callbacks,
             real_FISTP_calls=conversions, FP_arithmetic_conversion_substitutions=0,
             memory_faults=dict(faults), fault_details=fault_details, world_calls=dict(calls),
             output_sha256=sha(packed_output), decoded_output_sha256=sha(decoded), native_source_used=False)
with options.proof.open('x') as stream:
    json.dump(proof, stream, indent=2)
    stream.write('\n')
print(json.dumps(proof), flush=True)
