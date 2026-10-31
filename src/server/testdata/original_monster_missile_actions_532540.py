"""Execute unchanged missile roots, prediction, stack, buff helpers and RNG.

Declared world services are allocation, collision trace, creation/deletion,
audio, buff publication, player attack and registered action callbacks.
No trajectory, range, stack admission, RNG or prediction result is supplied.
"""
import argparse, collections, gzip, hashlib, itertools, json, struct
from pathlib import Path
from capstone import Cs, CS_ARCH_X86, CS_MODE_32
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_WRITE
from unicorn.x86_const import *

parser = argparse.ArgumentParser(description=__doc__)
for key in ['game', 'fixture', 'proof', 'input-seal']:
    parser.add_argument('--' + key, type=Path, required=True)
parser.add_argument('--control', type=lambda v: int(v, 0), choices=[0xe3f, 0x23f], required=True)
parser.add_argument('--limit', type=int, default=0)
args = parser.parse_args()
sha = lambda raw: hashlib.sha256(raw).hexdigest()
blob = args.game.read_bytes()
assert sha(blob) == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe = struct.unpack_from('<I', blob, 60)[0]
optional = pe + 24
base = struct.unpack_from('<I', blob, optional + 28)[0]
size = (struct.unpack_from('<I', blob, optional + 56)[0] + 4095) & ~4095
table = optional + struct.unpack_from('<H', blob, pe + 20)[0]
u = Uc(UC_ARCH_X86, UC_MODE_32)
u.mem_map(base, size)
for i in range(struct.unpack_from('<H', blob, pe + 6)[0]):
    _, rva, length, offset = struct.unpack_from('<IIII', blob, table + i * 40 + 8)
    if length:
        u.mem_write(base + rva, blob[offset:offset + length])
intervals = [(0x532540, 208), (0x532610, 496), (0x533080, 64),
    (0x534b10, 73), (0x534b60, 47), (0x534b90, 47),
    (0x4ec580, 38), (0x419130, 22), (0x419150, 43), (0x40a5c0, 18),
    (0x50a0d0, 53), (0x4ff350, 35), (0x5341d0, 26), (0x4ff5b0, 105),
    (0x424920, 12), (0x424800, 22), (0x50a260, 251), (0x50a160, 254),
    (0x50a110, 76), (0x50a010, 13), (0x4e39d0, 23), (0x424300, 31), (0x415fa0, 67)]
cs = Cs(CS_ARCH_X86, CS_MODE_32)
instructions, original_ranges, bodies = {}, [], {}
for address, length in intervals:
    raw = bytes(u.mem_read(address, length))
    decoded = list(cs.disasm(raw, address))
    assert sum(v.size for v in decoded) == length, hex(address)
    instructions.update({v.address: bytes(v.bytes) for v in decoded})
    original_ranges.append(dict(address=address, size=length, sha256=sha(raw)))
    if address in [0x532540, 0x532610, 0x533080]:
        bodies[address] = sum(v.size for v in decoded if v.mnemonic != 'nop')
heap, stack, helpers = 0x21000000, 0x22000000, 0x23000000
u.mem_map(heap, 0x20000)
u.mem_map(stack, 0x10000)
u.mem_map(helpers, 0x1000)
unit, target = heap + 0x1000, heap + 0x3000
updates = [heap + 0x5000, heap + 0x7000]
definitions = [heap + 0x9000, heap + 0xa000]
sounds = [heap + 0xb000, heap + 0xb100]
projectiles = [heap + 0xc000, heap + 0xd000]
types, type_record, type_name = heap + 0x10000, heap + 0x11000, heap + 0x11100
sp, stop, init, cw = stack + 0xe000, helpers + 0x100, helpers + 0x200, helpers + 0x300
cancel_pc, end_pc = helpers + 0x500, helpers + 0x600
wr = lambda a, v: u.mem_write(a, struct.pack('<I', v & 0xffffffff))
rd = lambda a: struct.unpack('<I', u.mem_read(a, 4))[0]
bits = lambda v: struct.unpack('<I', struct.pack('<f', v))[0]
u.mem_write(init, b'\xdb\xe3\xd9\x2d' + struct.pack('<I', cw))
u.mem_write(cw, struct.pack('<H', args.control))
frame_address, fps_address, changed_address, random_index = 0x84ea04, 0x85b3fc, 0x75ae60, 0x62efd0
random_table = bytes(u.mem_read(0x590404, 4096 * 4))
assert sha(random_table) == '7dfccdce34d74e348fae4181630a4c41a20a560432658ba2526969bb7000605f'
action_ids = [0, 1, 4, 24, 25, 31, 41, 50, 60, 0xffffffff]
registry_addresses = [((0x5bfed0 + ((a << 4) & 0xffffffff) + o) & 0xffffffff) for a in action_ids for o in [0, 4]]
saved = {UC_X86_REG_EBX: 0x12345678, UC_X86_REG_ESI: 0x23456789, UC_X86_REG_EDI: 0x3456789a, UC_X86_REG_EBP: 0x456789ab}
reset_fields = [(8, 4), (268, 4), (296, 4), (364, 4), (481, 1), (482, 1), (483, 1), (496, 4), (548, 4)]
allowed = [(unit + o, n) for o, n in [(8, 8), (56, 8), (124, 4), (136, 4), (176, 4), (340, 4), (344, 2), (390, 2), (408, 1), (431, 1), (748, 4)]]
for update in updates:
    allowed += [(update + o, n) for o, n in reset_fields + [(484, 4), (488, 4), (512, 4), (544, 1), (552, 576)]]
for projectile in projectiles:
    allowed += [(projectile + o, n) for o, n in [(56, 8), (80, 8), (124, 4), (544, 4)]]
active, current, events, pushes, trace = False, None, [], [], set()
call_counts, service_counts = collections.Counter(), collections.Counter()

def role(value):
    if value == 0: return {'kind': 'nil'}
    if value == unit: return {'kind': 'unit'}
    if value == target: return {'kind': 'target'}
    for kind, group in [('update', updates), ('definition', definitions), ('sound', sounds), ('projectile', projectiles)]:
        for index, item in enumerate(group):
            if value == item: return {'kind': kind, 'index': index}
    for index, update in enumerate(updates):
        if update + 552 <= value < update + 1128 and (value - update - 552) % 24 == 0:
            return {'kind': 'slot', 'update': index, 'index': (value - update - 552) // 24}
    return {'kind': 'u32', 'value': value}

def return_service(value=0):
    pointer = u.reg_read(UC_X86_REG_ESP)
    u.reg_write(UC_X86_REG_EAX, value & 0xffffffff)
    u.reg_write(UC_X86_REG_EIP, rd(pointer))
    u.reg_write(UC_X86_REG_ESP, pointer + 4)

def mutate(effect, cached):
    if effect == 'none': return
    if effect == 'swap-update': wr(unit + 748, updates[1])
    elif effect == 'clear-class': wr(unit + 8, 0)
    elif effect == 'swap-update-and-clear-class': wr(unit + 748, updates[1]); wr(unit + 8, 0)
    elif effect == 'advance-frame': wr(frame_address, rd(frame_address) + 3)
    elif effect == 'replace-sound': wr(cached + 488, sounds[1])
    elif effect == 'null-sound': wr(cached + 488, 0)
    elif effect == 'replace-definition': wr(cached + 484, definitions[1])
    elif effect == 'complete-cached': u.mem_write(cached + 483, b'\x81')
    elif effect == 'change-head': u.mem_write(cached + 544, b'\x02')
    elif effect == 'cap-cached-head': u.mem_write(cached + 544, b'\x16')
    elif effect == 'change-head-args':
        head = struct.unpack('b', u.mem_read(cached + 544, 1))[0]
        assert 0 <= head < 24
        wr(cached + 556 + head * 24, bits(410)); wr(cached + 560 + head * 24, bits(510)); wr(cached + 564 + head * 24, 0)
    elif effect == 'change-direction': u.mem_write(unit + 124, struct.pack('<h', 96))
    elif effect == 'change-position': wr(unit + 56, bits(325)); wr(unit + 60, bits(305))
    else: raise AssertionError(effect)

def on_code(uc, pc, length, _):
    if not active: return
    if pc == stop: uc.emu_stop(); return
    pointer, data = uc.reg_read(UC_X86_REG_ESP), current['input']
    services = {0x4e48f0: 'publish-buffs', 0x501960: 'audio', 0x5341a0: 'diagnostic',
        0x538960: 'player-attack', 0x4e3810: 'allocate', 0x535250: 'trace',
        0x4daa50: 'create', 0x4e5cc0: 'delete', cancel_pc: 'cancel', end_pc: 'end'}
    if pc in services:
        name = services[pc]; service_counts[name] += 1
        if name == 'diagnostic': return_service(); return
        if name == 'publish-buffs':
            assert rd(pointer + 4) == unit
            flags = rd(pointer + 8); wr(unit + 340, flags)
            events.append(dict(service=name, flags=flags))
            if sum(v['service'] == name for v in events) == 1: mutate(data['buff_effect'], updates[0])
            return_service(); return
        if name == 'audio':
            assert rd(pointer + 8) == unit and rd(pointer + 12) == 0 and rd(pointer + 16) == 0
            events.append(dict(service=name, sound=rd(pointer + 4), live_update=role(rd(unit + 748))))
        elif name == 'allocate':
            address = rd(pointer + 4)
            text = bytes(uc.mem_read(address, 64)).split(b'\0')[0].decode('ascii')
            assert text in ['FixtureMissileA', 'FixtureMissileB']
            allocated = 0 if data['allocation'] == 0 else projectiles[data['allocation'] - 1]
            events.append(dict(service=name, type_name=text, returned=role(allocated)))
            mutate(data['allocate_effect'], updates[0]); return_service(allocated); return
        elif name == 'trace':
            line = rd(pointer + 4)
            assert [rd(pointer + o) for o in [8, 12, 16]] == [0, 0, 5]
            events.append(dict(service=name, line_bits=[rd(line + 4 * i) for i in range(4)], flags=5, result=data['trace']))
            mutate(data['trace_effect'], updates[0]); return_service(data['trace']); return
        elif name == 'create':
            projectile = rd(pointer + 4); assert projectile in projectiles and rd(pointer + 8) == unit
            pos = [rd(pointer + 12), rd(pointer + 16)]
            events.append(dict(service=name, projectile=role(projectile), position_bits=pos))
            wr(projectile + 56, pos[0]); wr(projectile + 60, pos[1])
        elif name == 'delete':
            assert rd(pointer + 4) in projectiles
            events.append(dict(service=name, projectile=role(rd(pointer + 4))))
        else:
            assert rd(pointer + 4) == unit
            events.append(dict(service=name, class_before=rd(unit + 8), live_update_before=role(rd(unit + 748)), effect=data[name + '_effect']))
        cached = uc.reg_read(UC_X86_REG_ESI) if name in ['cancel', 'end'] else updates[0]
        assert cached in updates
        mutate(data[name + '_effect'], cached)
        return_service(data['attack_return'] if name == 'player-attack' else 0); return
    assert pc in instructions and bytes(uc.mem_read(pc, length)) == instructions[pc], (current['name'], hex(pc))
    trace.add(pc)
    if any(pc == start for start, _ in intervals): call_counts[hex(pc)] += 1
    if pc == 0x50a260:
        assert rd(pointer + 4) == unit
        pushes.append(dict(action=rd(pointer + 8), live_update=role(rd(unit + 748))))
    if pc in [0x50a27b, 0x50a28d, 0x50a2aa, 0x50a35a]:
        assert pushes and 'returned' not in pushes[-1]
        pushes[-1]['returned'] = role(uc.reg_read(UC_X86_REG_EAX))

def on_write(uc, access, address, length, value, _):
    if not active: return
    if stack <= address and address + length <= stack + 0x10000: return
    if address in [frame_address, changed_address, random_index] and length == 4: return
    if any(a <= address and address + length <= a + n for a, n in allowed): return
    raise AssertionError((current['name'], 'unexpected write', hex(address), length, value))
u.hook_add(UC_HOOK_CODE, on_code)
u.hook_add(UC_HOOK_MEM_WRITE, on_write)

defaults = dict(cls=2, subclass=0, frame=1400, deadline=1401, fps=60,
    head=0, action=25, root_action=1, started=0, callback_present=0,
    buffs=0, sound=1, range=bits(10), delay_min=20, delay_max=30, rng=19,
    direction=0, unit_pos=[bits(320), bits(300)], radius=bits(10),
    target_pos=[bits(350), bits(340)], target_velocity=[bits(1), bits(-1)],
    head_pos=[bits(350), bits(340)], targeted=1, speed=bits(12),
    animation=2, strike_frame=2, strike_done=0, complete=0, alt_complete=0,
    allocation=1, trace=1, attack_return=1, initial_changed=0, condition_below=0)
for name in ['buff', 'audio', 'player-attack', 'allocate', 'trace', 'create', 'delete', 'cancel', 'end']:
    defaults[name + '_effect'] = 'none'
cases, per_entry = [], collections.Counter()
def add(entry, family, **values):
    index = per_entry[entry]; per_entry[entry] += 1
    cases.append(dict(entry=entry, name=f'{entry:06x}/{index:05d}', family=family, input=dict(defaults, **values)))
for cls, head, action, started, present in itertools.product([0, 2, 4, 6], [-1, 0, 1, 20, 21, 22, 23], [0, 1, 4, 25, 31, 41, 50, 0xffffffff], [0, 1], [0, 1]):
    add(0x532540, 'actual-cooldown-stack-admission', cls=cls, head=head, action=action, started=started, callback_present=present)
float_bits = {0, 0x80000000, 1, 0x007fffff, 0x00800000, 0x7f7fffff, 0xff7fffff, 0x7f800000, 0xff800000, 0x7fc12345, 0x7fa12345, bits(-1)}
for center in [bits(1), bits(10), bits(15)]: float_bits.update(center + v for v in range(-3, 4))
for value, head in itertools.product(sorted(float_bits), [-1, 21, 22, 23]):
    add(0x532540, 'range-F32-constant-original-multiply-and-spill', range=value, head=head)
for cls, buffs, effect, sound in itertools.product([0, 2, 4, 6], [0, 1, 1 << 23, (1 << 23) | 1, 0xffffffff], ['none', 'swap-update', 'clear-class', 'advance-frame', 'replace-sound', 'null-sound', 'replace-definition'], [0, 1]):
    add(0x532540, 'ready-buff-publication-cached-data-live-sound', cls=cls, deadline=1400, buffs=buffs, buff_effect=effect, sound=sound)
for effect, head in itertools.product(['none', 'swap-update', 'clear-class', 'advance-frame', 'cap-cached-head'], [0, 1, 21, 22]):
    add(0x532540, 'cancel-mutation-during-real-Push', action=4, head=head, started=1, callback_present=1, cancel_effect=effect)
for index, bounds in itertools.product([0, 1, 19, 4095], [(0, 0), (1, 1), (20, 30), (-5, 5), (20, 19)]):
    add(0x532540, 'actual-RNG-table-wrap-and-signed-bounds', deadline=1400, rng=index, delay_min=bounds[0], delay_max=bounds[1])
for frame, deadline in itertools.product([0, 1, 0xfffffffe, 0xffffffff], repeat=2):
    add(0x532540, 'unsigned-frame-and-deadline', frame=frame, deadline=deadline)
for animation, strike_frame, done, complete, allocation, sound in itertools.product([0, 1, 2, 255], [0, 2, 255, 256, 0xffffffff], [0, 1, 128], [0, 1, 128, 255], [0, 1, 2], [0, 1]):
    add(0x532610, 'animation-allocation-completion-and-live-audio', animation=animation, strike_frame=strike_frame, strike_done=done, complete=complete, allocation=allocation, sound=sound)
effects = ['none', 'swap-update', 'clear-class', 'swap-update-and-clear-class', 'advance-frame', 'replace-sound', 'null-sound', 'replace-definition', 'complete-cached', 'change-head', 'change-head-args', 'change-direction', 'change-position']
for name in ['allocate', 'trace', 'create', 'delete', 'audio']:
    for effect, complete in itertools.product(effects, [0, 128]):
        add(0x532610, 'world-callback-' + name, complete=complete, alt_complete=255, trace=0 if name == 'delete' else 1, **{name + '_effect':effect})
for value, effect, head in itertools.product([0, 1, 128, 0x80000000, 0xffffffff], ['none', 'swap-update', 'clear-class', 'complete-cached', 'advance-frame'], [0, 1, 23]):
    add(0x532610, 'NPC-player-attack-return-and-real-Pop', subclass=16, attack_return=value, head=head, **{'player-attack_effect':effect})
for head, action, started, present, condition in itertools.product([-1, 0, 1, 21, 23], [0, 1, 4, 25, 41, 50, 0xffffffff], [0, 1], [0, 1], [0, 1]):
    add(0x532610, 'non-firing-original-Pop-end-admission', animation=0, strike_frame=2, complete=1, head=head, action=action, started=started, callback_present=present, condition_below=condition)
for effect in effects:
    add(0x532610, 'end-callback-cached-completion', animation=0, complete=1, head=1, started=1, callback_present=1, end_effect=effect)
for direction, speed, radius, targeted in itertools.product([0, 1, 16, 32, 63, 64, 96, 127, 128, 160, 192, 224, 255], [bits(0), bits(-1), bits(0.1), bits(12), 1, 0x7f7fffff, 0x7f800000, 0x7fc12345], [bits(0), bits(10), bits(-4), 0x7f7fffff], [0, 1]):
    add(0x532610, 'actual-prediction-trajectory-direction-spills', direction=direction, speed=speed, radius=radius, targeted=targeted)
for cls, head in itertools.product([0, 2, 4, 6], [-1, 0, 1, 23]):
    add(0x532610, 'non-firing-class-and-no-head', cls=cls, head=head, animation=0)
input_bytes = (json.dumps(cases, sort_keys=True, separators=(',', ':')) + '\n').encode()
if args.input_seal.exists(): assert args.input_seal.read_bytes() == input_bytes
else:
    with args.input_seal.open('xb') as f: f.write(input_bytes)
selected_cases = cases[:args.limit] if args.limit else cases
seed = bytes((i * 31 + 19) & 255 for i in range(0x20000))
protected, cursor = [], heap
for address, length in sorted(allowed):
    if address > cursor: protected.append((cursor, address - cursor))
    cursor = max(cursor, address + length)
if cursor < heap + len(seed): protected.append((cursor, heap + len(seed) - cursor))
cache_addresses = [0x833e54, 0x833e58, 0x833e5c, 0x833e70, 0x833e74, 0x833e78, 0x833e7c]
global_addresses = [frame_address, fps_address, changed_address, random_index]
rows = []
for current in selected_cases:
    data = current['input']
    u.mem_write(heap, seed); u.mem_write(stack, b'\xcc' * 0x10000)
    for address, length in [(unit, 0x400), (target, 0x400), *[(v, 0x1000) for v in updates], *[(v, 0x400) for v in definitions], *[(v, 0x400) for v in projectiles]]:
        u.mem_write(address, b'\0' * length)
    wr(unit + 4, 117); wr(unit + 8, data['cls']); wr(unit + 12, data['subclass']); wr(unit + 36, 118)
    wr(unit + 56, data['unit_pos'][0]); wr(unit + 60, data['unit_pos'][1])
    u.mem_write(unit + 124, struct.pack('<h', data['direction']))
    wr(unit + 136, 97); wr(unit + 176, data['radius']); wr(unit + 340, data['buffs']); wr(unit + 748, updates[0])
    for enchant in [0, 23]:
        u.mem_write(unit + 344 + 2 * enchant, struct.pack('<H', 0x5511 + enchant))
        u.mem_write(unit + 408 + enchant, bytes([0x41 + enchant]))
    wr(target + 56, data['target_pos'][0]); wr(target + 60, data['target_pos'][1])
    wr(target + 80, data['target_velocity'][0]); wr(target + 84, data['target_velocity'][1])
    for index, update in enumerate(updates):
        wr(update + 484, definitions[index]); wr(update + 488, sounds[index] if data['sound'] else 0)
        wr(update + 512, data['deadline'] if index == 0 else 701); wr(update + 1196, target)
        for offset, length in reset_fields: u.mem_write(update + offset, bytes([0x71 + index]) * length)
        u.mem_write(update + 481, bytes([data['animation']]))
        u.mem_write(update + 482, bytes([data['strike_done']]))
        u.mem_write(update + 483, bytes([data['complete'] if index == 0 else data['alt_complete']]))
        u.mem_write(update + 544, struct.pack('b', data['head']))
        for slot in range(24):
            record = update + 552 + slot * 24; wr(record, 1)
            for argument in range(4): wr(record + 4 + 4 * argument, 0x61000000 + index * 0x10000 + slot * 0x100 + argument * 0x10 + 19)
            wr(record + 20, 0)
        if data['head'] >= 0:
            record = update + 552 + data['head'] * 24
            wr(record, data['action']); wr(record + 20, data['started'])
            wr(record + 4, data['head_pos'][0]); wr(record + 8, data['head_pos'][1]); wr(record + 12, target if data['targeted'] else 0)
            if data['head'] > 0: wr(update + 552, data['root_action'])
            if data['condition_below']:
                for slot in range(1, data['head']): wr(update + 552 + slot * 24, 41)
        u.mem_write(definitions[index] + 148, (b'FixtureMissileA' if index == 0 else b'FixtureMissileB') + b'\0')
        wr(definitions[index] + 212, data['range'] if index == 0 else bits(19))
        wr(definitions[index] + 216, data['strike_frame'])
        wr(definitions[index] + 220, data['delay_min'] if index == 0 else 31)
        wr(definitions[index] + 224, data['delay_max'] if index == 0 else 35)
    for index, projectile in enumerate(projectiles):
        wr(projectile + 56, bits(701 + index)); wr(projectile + 60, bits(801 + index))
        wr(projectile + 80, bits(17)); wr(projectile + 84, bits(-19))
        u.mem_write(projectile + 124, struct.pack('<hh', 77, 78)); wr(projectile + 544, data['speed'])
    for index, sound in enumerate(sounds):
        for slot in range(17): wr(sound + 4 * slot, 600 + 100 * index + slot)
    for action in action_ids:
        for offset, address in [(0, end_pc), (4, cancel_pc)]:
            wr((0x5bfed0 + ((action << 4) & 0xffffffff) + offset) & 0xffffffff, address if data['callback_present'] else 0)
    for address, value in zip(cache_addresses, range(301, 308)): wr(address, value)
    wr(types + 117 * 4, type_record); wr(type_record + 4, type_name)
    u.mem_write(type_name, b'MissileFixture\0'); wr(0x7520d4, types)
    for enchant, spell_id in [(0, 10), (23, 11)]:
        wr(0x596ea4 + 4 * enchant, spell_id); wr(0x663f34 + 80 * spell_id + 8, 900 + enchant)
    wr(frame_address, data['frame']); wr(fps_address, data['fps'])
    wr(changed_address, data['initial_changed']); wr(random_index, data['rng'])
    before = [bytes(u.mem_read(address, length)) for address, length in protected]
    guards = {address: bytes(u.mem_read(address - 16, 36)) for address in global_addresses}
    cache_before = {address: rd(address) for address in cache_addresses}
    registry_before = {address: rd(address) for address in registry_addresses}
    u.reg_write(UC_X86_REG_ESP, sp); u.emu_start(init, init + 8, count=4)
    for register, value in saved.items(): u.reg_write(register, value)
    wr(sp, stop); wr(sp + 4, unit); wr(sp + 8, 0x77665544)
    events, pushes = [], []; active = True
    try: u.emu_start(current['entry'], stop + 1, count=100000)
    except Exception as error: raise AssertionError((current['name'], data, str(error))) from error
    finally: active = False
    assert u.reg_read(UC_X86_REG_EIP) == stop and u.reg_read(UC_X86_REG_ESP) == sp + 4, current['name']
    assert u.reg_read(UC_X86_REG_FPCW) == args.control and (u.reg_read(UC_X86_REG_FPSW) >> 11) & 7 == 0, current['name']
    assert all(u.reg_read(register) == value for register, value in saved.items()), current['name']
    assert before == [bytes(u.mem_read(address, length)) for address, length in protected], current['name']
    assert bytes(u.mem_read(stack, 0xc000)) == b'\xcc' * 0xc000
    assert bytes(u.mem_read(sp + 12, 0x10000 - (sp + 12 - stack))) == b'\xcc' * (0x10000 - (sp + 12 - stack))
    for address, raw in guards.items():
        after = bytes(u.mem_read(address - 16, 36))
        for i, (old, new) in enumerate(zip(raw, after)):
            absolute = address - 16 + i
            assert old == new or any(other <= absolute < other + 4 for other in [frame_address, changed_address, random_index]), (current['name'], hex(absolute))
    assert cache_before == {address: rd(address) for address in cache_addresses}
    assert registry_before == {address: rd(address) for address in registry_addresses}
    assert all('returned' in item for item in pushes)
    states = []
    for update in updates:
        states.append(dict(head=struct.unpack('b', u.mem_read(update + 544, 1))[0],
            reset_fields=[int.from_bytes(u.mem_read(update + offset, length), 'little') for offset, length in reset_fields],
            definition=role(rd(update + 484)), sound=role(rd(update + 488)), deadline=rd(update + 512),
            slots=[dict(action=rd(update + 552 + 24 * i), args=[role(rd(update + 556 + 24 * i + 4 * j)) for j in range(4)], started=rd(update + 572 + 24 * i)) for i in range(24)]))
    projectile_states = [dict(position_bits=[rd(p + 56), rd(p + 60)], velocity_bits=[rd(p + 80), rd(p + 84)], directions=list(struct.unpack('<hh', u.mem_read(p + 124, 4))), speed=rd(p + 544)) for p in projectiles]
    returned = u.reg_read(UC_X86_REG_EAX)
    rows.append(dict(current, output=dict(return_value=role(returned) if current['entry'] == 0x532540 else dict(kind='byte', value=returned & 255),
        cls=rd(unit + 8), subclass=rd(unit + 12), unit_update=role(rd(unit + 748)), field34=rd(unit + 136),
        unit_position_bits=[rd(unit + 56), rd(unit + 60)], direction=struct.unpack('<h', u.mem_read(unit + 124, 2))[0],
        buffs=rd(unit + 340), durations=[int.from_bytes(u.mem_read(unit + 344 + 2 * i, 2), 'little') for i in [0, 23]],
        powers=[u.mem_read(unit + 408 + i, 1)[0] for i in [0, 23]], updates=states, projectiles=projectile_states,
        frame=rd(frame_address), rng=rd(random_index), stack_changed=rd(changed_address), events=events, pushes=pushes)))

for item in original_ranges:
    assert sha(bytes(u.mem_read(item['address'], item['size']))) == item['sha256']
assert bytes(u.mem_read(0x590404, len(random_table))) == random_table
assert sha(args.game.read_bytes()) == sha(blob)
executed_per_range = {hex(a): sum(len(instructions[pc]) for pc in trace if a <= pc < a + n) for a, n in intervals}
if not args.limit:
    assert all(executed_per_range[hex(a)] == n for a, n in bodies.items()), (executed_per_range, bodies)
fixture = dict(schema='original-monster-missile-actions-532540-v1', original_sha256=sha(blob), control_word=args.control,
    input_sha256=sha(input_bytes), declared_engine_services=['allocation', 'map-trace', 'create-at', 'delayed-delete',
        'audio-return-zero', 'buff-flag-publication', 'diagnostic', 'player-attack', 'registered-cancel', 'registered-end'], rows=rows)
raw = (json.dumps(fixture, sort_keys=True, separators=(',', ':')) + '\n').encode()
with args.fixture.open('xb') as f:
    with gzip.GzipFile(filename='', mode='wb', fileobj=f, mtime=0, compresslevel=6) as z:
        z.write(raw)
proof = dict(status='smoke_only' if args.limit else 'passed', contexts=len(rows), full_input_contexts=len(cases),
    control_word=hex(args.control), smoke_limit=args.limit, input_sha256=sha(input_bytes),
    original_GAME_EXE_sha256=sha(blob), original_ranges=original_ranges,
    per_entry=dict(collections.Counter(hex(r['entry']) for r in rows)),
    per_family=dict(collections.Counter(r['family'] for r in rows)),
    original_calls=dict(call_counts), declared_service_calls=dict(service_counts),
    original_root_body_bytes={hex(a):n for a,n in bodies.items()}, executed_bytes_per_range=executed_per_range,
    all_original_code_RNG_table_type_cache_and_callback_registry_bytes_preserved=True,
    no_substituted_trajectory_prediction_range_FP_RNG_push_or_pop_result=True,
    full_heap_readonly_guards_stack_bounds_global_neighbors_NV_FPCW_and_FPU_TOP_checked=True,
    initialized_type_and_sound_resources_are_input=True,
    world_service_scope='World allocation, collision, placement/deletion, publication/audio, attacks and registered action callbacks are declared boundaries. Root math, prediction, stack and RNG execute original instructions. This does not certify full transitive world implementations or the audio service return value.',
    fixture_sha256=sha(args.fixture.read_bytes()), raw_fixture_sha256=sha(raw), overall_porting_100_percent=False)
with args.proof.open('x') as f:
    json.dump(proof, f, indent=2)
    f.write('\n')
print(json.dumps(proof))
