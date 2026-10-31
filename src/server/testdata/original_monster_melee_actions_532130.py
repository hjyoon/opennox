"""Execute unchanged melee roots, blocker, stack, RNG, morph and predicates.

Declared engine services: spatial enumeration, buff-flag publication, audio,
diagnostics, player attack and registered action/strike callbacks. Enumeration
invokes the original blocker; predicates, stack admission and FP results are
never replaced. Initialized type/resource caches are explicit input state.
"""
import argparse, collections, gzip, hashlib, itertools, json, struct
from pathlib import Path
from capstone import Cs, CS_ARCH_X86, CS_MODE_32
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_WRITE
from unicorn.x86_const import *

parser = argparse.ArgumentParser(description=__doc__)
for key in ['game', 'fixture', 'proof']:
    parser.add_argument('--' + key, type=Path, required=True)
parser.add_argument('--control', type=lambda value: int(value, 0), choices=[0xe3f, 0x23f], required=True)
parser.add_argument('--limit', type=int, default=0)
args = parser.parse_args()
blob = args.game.read_bytes()
sha = lambda raw: hashlib.sha256(raw).hexdigest()
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
intervals = [(0x532130, 608), (0x532390, 176), (0x532440, 256),
             (0x4f7e80, 99), (0x5330c0, 672), (0x534b10, 73), (0x534b60, 47), (0x534b90, 47),
             (0x4ec580, 38), (0x419130, 22), (0x419150, 43), (0x40a5c0, 18), (0x50a0d0, 53),
             (0x4ff350, 35), (0x5341d0, 26), (0x4ff5b0, 105), (0x424920, 12), (0x424800, 22),
             (0x50a260, 251), (0x50a160, 254), (0x50a110, 76), (0x50a010, 13),
             (0x4e39d0, 23), (0x424300, 31), (0x4faac0, 44), (0x4faaf0, 44), (0x415fa0, 67)]
cs = Cs(CS_ARCH_X86, CS_MODE_32)
instructions, original_ranges = {}, []
for address, length in intervals:
    raw = bytes(u.mem_read(address, length))
    decoded = list(cs.disasm(raw, address))
    assert sum(v.size for v in decoded) == length, hex(address)
    instructions.update({v.address: bytes(v.bytes) for v in decoded})
    original_ranges.append({'address': address, 'size': length, 'sha256': sha(raw)})
heap, stack, helpers = 0x21000000, 0x22000000, 0x23000000
u.mem_map(heap, 0x30000)
u.mem_map(stack, 0x10000)
u.mem_map(helpers, 0x2000)
unit = heap + 0x1000
candidates = [heap + value for value in [0x3000, 0x5000, 0x7000]]
updates = [heap + 0x9000, heap + 0xb000]
definitions = [heap + 0xd000, heap + 0xe000]
sounds = [heap + 0x10000, heap + 0x10100]
player = heap + 0x12000
player_info = heap + 0x16000
candidate_updates = [heap + value for value in [0x18000, 0x1a000, 0x1c000]]
candidate_infos = [heap + value for value in [0x20000, 0x22000, 0x24000]]
types, type_record, type_name = heap + 0x27000, heap + 0x28000, heap + 0x28100
sp, stop, init, cw = stack + 0xe000, helpers + 0x100, helpers + 0x200, helpers + 0x300
cancel_pc, end_pc, strike_pc, enum_pc = [helpers + value for value in [0x500, 0x600, 0x700, 0x1000]]
wr = lambda address, value: u.mem_write(address, struct.pack('<I', value & 0xffffffff))
rd = lambda address: struct.unpack('<I', u.mem_read(address, 4))[0]
bits = lambda value: struct.unpack('<I', struct.pack('<f', value))[0]
u.mem_write(init, b'\xdb\xe3\xd9\x2d' + struct.pack('<I', cw))
u.mem_write(cw, struct.pack('<H', args.control))
frame_address, fps_address, changed_address = 0x84ea04, 0x85b3fc, 0x75ae60
selection, distance, random_index = 0x833c20, 0x833c24, 0x62efd0
random_table = bytes(u.mem_read(0x590404, 4096 * 4))
assert sha(random_table) == '7dfccdce34d74e348fae4181630a4c41a20a560432658ba2526969bb7000605f'
action_ids = [0, 1, 4, 24, 27, 31, 41, 50, 60, 0xffffffff]
registry = {((0x5bfed0 + ((action << 4) & 0xffffffff) + offset) & 0xffffffff): bytes(u.mem_read((0x5bfed0 + ((action << 4) & 0xffffffff) + offset) & 0xffffffff, 4)) for action in action_ids for offset in [0, 4]}
saved = {UC_X86_REG_EBX: 0x12345678, UC_X86_REG_ESI: 0x23456789,
         UC_X86_REG_EDI: 0x3456789a, UC_X86_REG_EBP: 0x456789ab}
reset_fields = [(8, 4), (268, 4), (296, 4), (364, 4), (481, 1), (482, 1), (483, 1), (496, 4), (548, 4)]
allowed = [(unit + offset, length) for offset, length in [(8, 8), (136, 4), (340, 4), (344, 2), (390, 2), (408, 1), (431, 1), (748, 4)]]
for update in updates:
    allowed += [(update + offset, length) for offset, length in reset_fields]
    allowed += [(update + offset, length) for offset, length in [(484, 4), (488, 4), (512, 4), (544, 1), (552, 576), (1128, 1), (1440, 4)]]
allowed += [(player + 292, 4)]
for definition in definitions:
    allowed += [(definition + offset, 4) for offset in [112, 128, 132]]
for candidate in candidates:
    allowed += [(candidate + 56, 8)]
active, current, events, pushes, trace = False, None, [], [], set()
call_counts = collections.Counter()
service_counts = collections.Counter()
enemy_returns = {v.address for v in cs.disasm(bytes(u.mem_read(0x5330c0, 672)), 0x5330c0) if v.mnemonic == 'ret'}

def role(value):
    if value == 0: return {'kind': 'nil'}
    if value == unit: return {'kind': 'unit'}
    if value == player: return {'kind': 'player'}
    for index, candidate in enumerate(candidates):
        if value == candidate: return {'kind': 'candidate', 'index': index}
    for index, update in enumerate(updates):
        if value == update: return {'kind': 'update', 'index': index}
        if update + 552 <= value < update + 1128 and (value - update - 552) % 24 == 0:
            return {'kind': 'slot', 'update': index, 'index': (value - update - 552) // 24}
    for index, sound in enumerate(sounds):
        if value == sound: return {'kind': 'sound', 'index': index}
    return {'kind': 'u32', 'value': value}

def service_return(value):
    pointer = u.reg_read(UC_X86_REG_ESP)
    u.reg_write(UC_X86_REG_EAX, value & 0xffffffff)
    u.reg_write(UC_X86_REG_EIP, rd(pointer))
    u.reg_write(UC_X86_REG_ESP, pointer + 4)

def mutate(effect, cached):
    if effect == 'swap-update': wr(unit + 748, updates[1])
    elif effect == 'swap-update-and-clear-class':
        wr(unit + 748, updates[1]); wr(unit + 8, 0)
    elif effect == 'clear-class': wr(unit + 8, 0)
    elif effect == 'class-player': wr(unit + 8, 4)
    elif effect == 'replace-sound': wr(cached + 488, sounds[1])
    elif effect == 'null-sound': wr(cached + 488, 0)
    elif effect == 'clear-bot': wr(cached + 1440, rd(cached + 1440) & ~0x20000)
    elif effect == 'set-bot': wr(cached + 1440, rd(cached + 1440) | 0x20000)
    elif effect == 'advance-frame': wr(frame_address, rd(frame_address) + 3)
    elif effect == 'change-player-link': wr(player + 292, updates[1])
    elif effect == 'complete-cached': u.mem_write(cached + 483, b'\x81')
    elif effect == 'cap-cached-head': u.mem_write(cached + 544, b'\x16')
    else: assert effect == 'none', effect

def on_code(uc, pc, length, _):
    if not active: return
    if pc == stop:
        uc.emu_stop(); return
    pointer = uc.reg_read(UC_X86_REG_ESP)
    data = current['input']
    services = {0x517c10: 'world', 0x4e48f0: 'publish-buffs', 0x501960: 'audio',
                0x5341a0: 'diagnostic', 0x538960: 'player-attack', cancel_pc: 'cancel', end_pc: 'end', strike_pc: 'strike'}
    if pc in services:
        name = services[pc]; service_counts[name] += 1
        if name == 'world':
            assert rd(pointer + 8) == 0x532390 and rd(pointer + 12) == unit
            rect = rd(pointer + 4)
            events.append({'service': name, 'rect': [rd(rect + 4 * i) for i in range(4)]})
            mutate(data['world_effect'], updates[0])
            uc.reg_write(UC_X86_REG_EIP, enum_pc); return
        if name == 'publish-buffs':
            assert rd(pointer + 4) == unit
            flags = rd(pointer + 8); wr(unit + 340, flags)
            events.append({'service': name, 'flags': flags})
            if sum(v['service'] == name for v in events) == 1: mutate(data['buff_effect'], updates[0])
            service_return(0); return
        if name == 'audio':
            assert rd(pointer + 8) == unit and rd(pointer + 12) == 0 and rd(pointer + 16) == 0
            events.append({'service': name, 'sound': rd(pointer + 4), 'live_update': role(rd(unit + 748))})
            service_return(0); return
        if name == 'diagnostic': service_return(0); return
        assert rd(pointer + 4) == unit
        effect = data[name + '_effect']
        cached = uc.reg_read(UC_X86_REG_ESI) if name in ['cancel', 'end'] else updates[0]
        assert cached in updates
        events.append({'service': name, 'class_before': rd(unit + 8), 'live_update_before': role(rd(unit + 748)), 'effect': effect})
        mutate(effect, cached)
        service_return(data['attack_return'] if name == 'player-attack' else data['strike_return'] if name == 'strike' else 0)
        return
    if enum_pc <= pc < enum_pc + enum_length:
        assert bytes(uc.mem_read(pc, length)) == enum_bytes[pc - enum_pc:pc - enum_pc + length]
        return
    assert pc in instructions, (current['name'], 'unexpected PC', hex(pc))
    assert bytes(uc.mem_read(pc, length)) == instructions[pc]
    trace.add(pc)
    if any(pc == start for start, _ in intervals): call_counts[hex(pc)] += 1
    if pc == 0x5330c0:
        events.append({'service': 'original-enemy', 'unit': role(rd(pointer + 4)), 'candidate': role(rd(pointer + 8))})
    if pc in enemy_returns:
        assert events[-1]['service'] == 'original-enemy'
        events[-1]['result'] = uc.reg_read(UC_X86_REG_EAX)
    if pc == 0x50a260:
        assert rd(pointer + 4) == unit
        pushes.append({'action': rd(pointer + 8), 'live_update': role(rd(unit + 748))})
    if pc in [0x50a27b, 0x50a28d, 0x50a2aa, 0x50a35a]:
        assert pushes and 'returned' not in pushes[-1]
        pushes[-1]['returned'] = role(uc.reg_read(UC_X86_REG_EAX))

def on_write(uc, access, address, length, value, _):
    if not active: return
    if stack <= address and address + length <= stack + 0x10000: return
    if address in [selection, distance, changed_address, frame_address, random_index] and length == 4: return
    if any(start <= address and address + length <= start + size for start, size in allowed): return
    raise AssertionError((current['name'], 'unexpected write', hex(address), length, value))
u.hook_add(UC_HOOK_CODE, on_code)
u.hook_add(UC_HOOK_MEM_WRITE, on_write)

defaults = {'cls': 2, 'subclass': 0, 'frame': 1400, 'deadline': 1401, 'fps': 60,
            'head': 0, 'action': 24, 'root_action': 1, 'started': 0, 'callback_present': 0,
            'weapons': 1, 'stamina': 20, 'hammer': 0, 'buffs': 0, 'sound': 1,
            'range': bits(10), 'delay_min': 20, 'delay_max': 30, 'rng': 19,
            'direction': 0, 'unit_pos': [bits(320), bits(300)], 'world': 'empty',
            'world_effect': 'none', 'buff_effect': 'none', 'cancel_effect': 'none', 'end_effect': 'none',
            'player-attack_effect': 'none', 'strike_effect': 'none', 'attack_return': 1, 'strike_return': 1,
            'status': 0, 'alt_status': 0, 'strike_present': 1, 'animation': 2,
            'strike_frame': 2, 'strike_done': 0, 'complete': 0, 'alt_complete': 0, 'initial_changed': 0}
cases, per_entry = [], collections.Counter()
def add(entry, family, **values):
    index = per_entry[entry]; per_entry[entry] += 1
    cases.append({'entry': entry, 'name': f'{entry:06x}/{index:05d}', 'family': family,
                  'input': dict(defaults, **values)})
for cls, head, action, started, present in itertools.product([0, 2, 4, 6], [-1, 0, 1, 20, 21, 22, 23], [0, 1, 4, 24, 31, 41, 50, 0xffffffff], [0, 1], [0, 1]):
    add(0x532130, 'cooldown-stack-admission', cls=cls, head=head, action=action, started=started, callback_present=present)
for frame, deadline in itertools.product([0, 1, 1400, 0x7fffffff, 0x80000000, 0xffffffff], [0, 1, 1399, 1400, 1401, 0x7fffffff, 0x80000000, 0xffffffff]):
    add(0x532130, 'unsigned-frame-deadline', frame=frame, deadline=deadline)
for subclass, weapons, stamina in itertools.product([0, 16, 0xffffffff], [0, 0xffffffff] + [1 << i for i in range(32)], [0, 1, 5, 20, 255]):
    add(0x532130, 'ready-stamina-priority', deadline=1400, subclass=subclass, weapons=weapons, stamina=stamina)
for world, subclass, hammer, head, started in itertools.product(['empty', 'friend', 'enemy', 'friend-then-enemy', 'enemy-then-friend', 'dead-front', 'behind', 'zero-vector', 'angle-boundary'], [0, 16, 0xffffffff], [0, 1], [-1, 0, 21, 23], [0, 1]):
    add(0x532130, 'actual-blocker-and-enemy', deadline=1400, world=world, subclass=subclass, hammer=hammer, head=head, started=started, callback_present=1)
for head, action, root_action in itertools.product([-1, 0, 1, 21, 22, 23], [0, 1, 24, 31, 41, 50], [1, 31]):
    add(0x532130, 'flee-stack-capacity', deadline=1400, world='friend', subclass=16, hammer=1, head=head, action=action, root_action=root_action)
float_bits = {0, 0x80000000, 1, 0x007fffff, 0x00800000, 0x7f7fffff, 0xff7fffff,
              0x7f800000, 0xff800000, 0x7fc12345, 0x7f800001, bits(-1)}
for center in [bits(1), bits(10), bits(15), bits(51)]: float_bits.update(center + value for value in range(-3, 4))
for value, head in itertools.product(sorted(float_bits), [-1, 21, 22, 23]):
    add(0x532130, 'original-range-multiply-and-spill', range=value, head=head)
for buffs, effect, sound in itertools.product([0, 1, 1 << 23, (1 << 23) | 1, 0xffffffff], ['none', 'swap-update', 'advance-frame', 'replace-sound', 'null-sound'], [0, 1]):
    add(0x532130, 'buff-publication-and-cached-data', deadline=1400, buffs=buffs, buff_effect=effect, sound=sound)
for effect in ['none', 'swap-update', 'clear-class', 'advance-frame', 'cap-cached-head']:
    for world in ['empty', 'friend', 'enemy']:
        add(0x532130, 'world-service-mutation', deadline=1400, subclass=16, hammer=1, world=world, world_effect=effect)
for effect, head in itertools.product(['none', 'swap-update', 'cap-cached-head', 'advance-frame'], [0, 1, 21, 22]):
    add(0x532130, 'cancel-cached-versus-live-data', head=head, action=4, started=1, callback_present=1, cancel_effect=effect)
    add(0x532130, 'end-cached-versus-live-data', deadline=1400, world='friend', subclass=16, hammer=1, head=head, action=24, started=1, callback_present=1, end_effect=effect)
for index, bounds, flee in itertools.product([0, 1, 19, 4095], [(0, 0), (1, 1), (20, 30), (-5, 5), (20, 19)], [0, 1]):
    add(0x532130, 'original-RNG-table-and-wrap', deadline=1400, delay_min=bounds[0], delay_max=bounds[1], rng=index, world='friend' if flee else 'empty', subclass=16 if flee else 0, hammer=flee)
for animation, strike_frame, done, complete, present, sound in itertools.product([0, 1, 2, 255], [0, 2, 255, 256, 0xffffffff], [0, 1, 128], [0, 1, 128, 255], [0, 1], [0, 1]):
    add(0x532440, 'strike-frame-completion-and-sound', animation=animation, strike_frame=strike_frame, strike_done=done, complete=complete, strike_present=present, sound=sound)
for value, effect, complete in itertools.product([0, 1, 0x80000000, 0xffffffff], ['none', 'swap-update', 'swap-update-and-clear-class', 'clear-class', 'class-player', 'replace-sound', 'null-sound', 'complete-cached'], [0, 1, 128]):
    add(0x532440, 'strike-return-and-live-sound-cached-completion', strike_return=value, strike_effect=effect, complete=complete, alt_complete=255)
for head, action, started, present, complete in itertools.product([-1, 0, 1, 21, 23], [0, 1, 4, 24, 41, 50, 0xffffffff], [0, 1], [0, 1], [0, 1]):
    add(0x532440, 'actual-Pop-and-end-admission', head=head, action=action, started=started, callback_present=present, complete=complete)
for effect in ['none', 'swap-update', 'clear-class', 'cap-cached-head', 'advance-frame']:
    add(0x532440, 'end-service-mutation', complete=1, head=1, started=1, callback_present=1, end_effect=effect)
for status, attack_return, effect in itertools.product([0, 0x20000, 0xffffffff], [0, 1, 128, 0xffffffff], ['none', 'change-player-link', 'advance-frame', 'set-bot']):
    values = dict(subclass=16, status=status, attack_return=attack_return)
    values['player-attack_effect'] = effect
    add(0x532440, 'NPC-bot-morph-and-player-attack', **values)
for status in [0x20000, 0xffffffff]:
    add(0x532440, 'NPC-cached-BOT-reload', subclass=16, status=status, attack_return=1, **{'player-attack_effect': 'clear-bot'})
for head, action, npc in itertools.product([2, 5, 23], [24, 41, 50], [0, 16]):
    add(0x532440, 'Pop-removes-below-head-conditions', head=head, action=action, complete=1, subclass=npc, attack_return=0, condition_below=1)
if args.limit: cases = cases[:args.limit]

seed = bytes((i * 31 + 19) & 255 for i in range(0x30000))
protected, cursor = [], heap
for start, length in sorted(allowed):
    if start > cursor: protected.append((cursor, start - cursor))
    cursor = max(cursor, start + length)
if cursor < heap + len(seed): protected.append((cursor, heap + len(seed) - cursor))
global_addresses = [frame_address, fps_address, changed_address, selection, distance, random_index]
cache_addresses = [0x833e54, 0x833e58, 0x833e5c, 0x833e70, 0x833e74, 0x833e78, 0x833e7c]
rows = []
for current in cases:
    data = current['input']; u.mem_write(heap, seed); u.mem_write(stack, b'\xcc' * 0x10000)
    for address, length in [(unit, 0x400), *[(v, 0x400) for v in candidates], *[(v, 0x1000) for v in updates], *[(v, 0x400) for v in definitions], (player, 0x3000), (player_info, 0x2000), *[(v, 0x2000) for v in candidate_updates + candidate_infos]]:
        u.mem_write(address, b'\0' * length)
    wr(unit + 4, 117); wr(unit + 8, data['cls']); wr(unit + 12, data['subclass']); wr(unit + 36, 118)
    wr(unit + 56, data['unit_pos'][0]); wr(unit + 60, data['unit_pos'][1])
    u.mem_write(unit + 124, struct.pack('<h', data['direction']))
    wr(unit + 136, 97); wr(unit + 340, data['buffs']); wr(unit + 748, updates[0])
    for enchant in [0, 23]:
        u.mem_write(unit + 344 + 2 * enchant, struct.pack('<H', 0x5511 + enchant))
        u.mem_write(unit + 408 + enchant, bytes([0x41 + enchant]))
    for index, update in enumerate(updates):
        wr(update + 484, definitions[index]); wr(update + 488, sounds[index] if data['sound'] else 0)
        wr(update + 512, data['deadline'] if index == 0 else 701)
        wr(update + 1440, data['status'] if index == 0 else data['alt_status'])
        wr(update + 2056, data['weapons']); wr(update + 2064, candidates[2] if data['hammer'] else 0)
        wr(update + 2180, player); wr(update + 1196, candidates[0])
        u.mem_write(update + 1128, bytes([data['stamina'] if index == 0 else 73]))
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
            wr(update + 552 + data['head'] * 24, data['action'])
            wr(update + 552 + data['head'] * 24 + 20, data['started'])
            if data['head'] > 0: wr(update + 552, data['root_action'])
            if data.get('condition_below', 0):
                for slot in range(1, data['head']): wr(update + 552 + slot * 24, 41)
        wr(definitions[index] + 108, data['strike_frame']); wr(definitions[index] + 112, data['range'])
        wr(definitions[index] + 128, data['delay_min']); wr(definitions[index] + 132, data['delay_max'])
        wr(definitions[index] + 236, strike_pc if data['strike_present'] else 0)
    wr(player + 292, updates[0]); wr(player + 276, player_info)
    for index, candidate in enumerate(candidates):
        wr(candidate + 4, 117); wr(candidate + 8, 2 if index != 1 else 4); wr(candidate + 36, 119 + index)
        wr(candidate + 56, bits(330 + 10 * index)); wr(candidate + 60, bits(300))
        wr(candidate + 748, candidate_updates[index]); wr(candidate_updates[index] + 276, candidate_infos[index])
    wr(candidates[2] + 12, 0x4000)
    world = data['world']
    order = {'empty': [], 'friend': [0], 'enemy': [1], 'friend-then-enemy': [0, 1], 'enemy-then-friend': [1, 0], 'dead-front': [0, 1], 'behind': [0], 'zero-vector': [0], 'angle-boundary': [0]}[world]
    if world == 'dead-front': wr(candidates[0] + 16, 0x8000)
    if world == 'behind': wr(candidates[0] + 56, bits(310))
    if world == 'zero-vector': wr(candidates[0] + 56, bits(320))
    if world == 'angle-boundary': wr(candidates[0] + 56, bits(321)); wr(candidates[0] + 60, bits(301.7320508))
    enum_bytes = b''
    for index in order:
        enum_bytes += b'\x68' + struct.pack('<I', unit) + b'\x68' + struct.pack('<I', candidates[index]) + b'\xb8' + struct.pack('<I', 0x532390) + b'\xff\xd0\x83\xc4\x08'
    enum_bytes += b'\xc3'; enum_length = len(enum_bytes); u.mem_write(enum_pc, enum_bytes)
    u.ctl_remove_cache(enum_pc, enum_pc + 0x100)
    for index, sound in enumerate(sounds):
        for slot in range(17): wr(sound + 4 * slot, 600 + 100 * index + slot)
    for action in action_ids:
        for offset, address in [(0, end_pc), (4, cancel_pc)]:
            wr((0x5bfed0 + ((action << 4) & 0xffffffff) + offset) & 0xffffffff, address if data['callback_present'] else 0)
    for address, index in [(0x833e54, 301), (0x833e58, 302), (0x833e5c, 303), (0x833e70, 304), (0x833e74, 305), (0x833e78, 306), (0x833e7c, 307)]: wr(address, index)
    wr(types + 117 * 4, type_record); wr(type_record + 4, type_name); u.mem_write(type_name, b'MeleeFixture\0'); wr(0x7520d4, types)
    for enchant, spell_id in [(0, 10), (23, 11)]:
        wr(0x596ea4 + 4 * enchant, spell_id); wr(0x663f34 + 80 * spell_id + 8, 900 + enchant)
    wr(frame_address, data['frame']); wr(fps_address, data['fps']); wr(changed_address, data['initial_changed'])
    wr(selection, candidates[2]); wr(distance, bits(17)); wr(random_index, data['rng'])
    before = [bytes(u.mem_read(start, length)) for start, length in protected]
    guards = {address: bytes(u.mem_read(address - 16, 36)) for address in global_addresses}
    caches = {address: rd(address) for address in cache_addresses}
    registry_before = {address: bytes(u.mem_read(address, 4)) for address in registry}
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
    assert before == [bytes(u.mem_read(start, length)) for start, length in protected], current['name']
    assert bytes(u.mem_read(stack, 0xc000)) == b'\xcc' * 0xc000 and bytes(u.mem_read(sp + 12, 0x10000 - (sp + 12 - stack))) == b'\xcc' * (0x10000 - (sp + 12 - stack)), current['name']
    for address, raw in guards.items():
        after = bytes(u.mem_read(address - 16, 36))
        for index, (old, new) in enumerate(zip(raw, after)):
            absolute = address - 16 + index
            assert old == new or any(other <= absolute < other + 4 for other in [frame_address, changed_address, selection, distance, random_index]), (current['name'], hex(absolute))
    assert caches == {address: rd(address) for address in cache_addresses}
    assert registry_before == {address: bytes(u.mem_read(address, 4)) for address in registry}
    assert all('returned' in item for item in pushes), current['name']
    states = []
    for update in updates:
        states.append({'head': struct.unpack('b', u.mem_read(update + 544, 1))[0],
                       'reset_fields': [int.from_bytes(u.mem_read(update + offset, length), 'little') for offset, length in reset_fields],
                       'deadline': rd(update + 512), 'stamina': u.mem_read(update + 1128, 1)[0], 'status': rd(update + 1440),
                       'sound': role(rd(update + 488)),
                       'slots': [{'action': rd(update + 552 + 24 * i), 'args': [role(rd(update + 556 + 24 * i + 4 * j)) for j in range(4)], 'started': rd(update + 572 + 24 * i)} for i in range(24)]})
    returned = u.reg_read(UC_X86_REG_EAX)
    rows.append(dict(current, output={'return_value': role(returned) if current['entry'] == 0x532130 else {'kind': 'byte', 'value': returned & 255},
                'class': rd(unit + 8), 'subclass': rd(unit + 12), 'unit_update': role(rd(unit + 748)), 'field34': rd(unit + 136),
                'buffs': rd(unit + 340), 'durations': [int.from_bytes(u.mem_read(unit + 344 + 2 * i, 2), 'little') for i in [0, 23]],
                'powers': [u.mem_read(unit + 408 + i, 1)[0] for i in [0, 23]], 'updates': states,
                'selection': role(rd(selection)), 'distance': rd(distance), 'frame': rd(frame_address), 'rng': rd(random_index),
                'stack_changed': rd(changed_address), 'player_link': role(rd(player + 292)), 'events': events, 'pushes': pushes}))

for item in original_ranges: assert sha(bytes(u.mem_read(item['address'], item['size']))) == item['sha256']
assert bytes(u.mem_read(0x590404, len(random_table))) == random_table and sha(args.game.read_bytes()) == sha(blob)
assert {hex(start): sum(len(instructions[pc]) for pc in trace if start <= pc < start + length) for start, length in intervals if start in [0x532130, 0x532390, 0x532440]} == {'0x532130': 600, '0x532390': 161, '0x532440': 248}
fixture = {'schema': 'original-monster-melee-actions-532130-v2', 'original_sha256': sha(blob), 'control_word': args.control,
           'declared_engine_services': ['world-enumeration-calls-original-blocker', 'buff-flag-publication', 'audio-return-zero', 'diagnostic', 'player-attack', 'registered-cancel', 'registered-end', 'registered-strike'],
           'rows': rows}
raw = (json.dumps(fixture, sort_keys=True, separators=(',', ':')) + '\n').encode()
with args.fixture.open('xb') as output:
    with gzip.GzipFile(filename='', mode='wb', fileobj=output, compresslevel=9, mtime=0) as stream: stream.write(raw)
proof = {'status': 'passed', 'contexts': len(rows), 'control_word': hex(args.control), 'smoke_limit': args.limit,
         'per_entry': dict(collections.Counter(hex(v['entry']) for v in rows)), 'per_family': dict(collections.Counter(v['family'] for v in rows)),
         'original_sha256': sha(blob), 'ranges': original_ranges, 'original_calls': dict(call_counts), 'declared_service_calls': dict(service_counts),
         'executed_original_bytes': sum(len(instructions[address]) for address in trace),
         'executed_bytes_per_range': {hex(start): sum(len(instructions[pc]) for pc in trace if start <= pc < start + length) for start, length in intervals},
         'no_substituted_FP_predicate_RNG_push_or_pop_result': True, 'full_heap_readonly_guards_stack_bounds_global_neighbors_code_RNG_caches_registry_NV_and_FPU_control_checked': True,
         'FPU_TOP_zero_after_every_call': True, 'original_unchanged': True,
         'fixture_sha256': sha(args.fixture.read_bytes()), 'raw_fixture_sha256': sha(raw)}
with args.proof.open('x') as output: json.dump(proof, output, indent=2); output.write('\n')
print(json.dumps(proof))
