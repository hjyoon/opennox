"""Bounded unchanged PE32 DefaultDamage plus UnitDamageClear's nil/zero-Max gates.
Flags/buffs/enemy/audio/Shield are declared services. Damage attribution, root
return and the health gates execute original instructions; no full-world claim.
"""
import hashlib
import json
import pathlib
import struct
import sys
from collections import Counter
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import (UC_X86_REG_EAX, UC_X86_REG_EIP, UC_X86_REG_ESP,
    UC_X86_REG_FPCW, UC_X86_REG_FPSW, UC_X86_REG_FPTAG)

repo, original = map(pathlib.Path, sys.argv[1:3])
blob = original.read_bytes()
manifest = json.loads((repo / 'toolchain/oracle/game-exe-functions.json').read_text())
assert hashlib.sha256(blob).hexdigest() == manifest['file_sha256'] == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe = struct.unpack_from('<I', blob, 0x3c)[0]
opt = pe + 24
base = struct.unpack_from('<I', blob, opt + 28)[0]
assert base == 0x400000 and struct.unpack_from('<H', blob, pe + 4)[0] == 0x14c
table = opt + struct.unpack_from('<H', blob, pe + 20)[0]
sections = [struct.unpack_from('<IIII', blob, table + i*40 + 8)
    for i in range(struct.unpack_from('<H', blob, pe + 6)[0])]
size = (struct.unpack_from('<I', blob, opt + 56)[0] + 4095) & ~4095
def read(address, count):
    for _, virtual, raw_size, raw in sections:
        if virtual <= address-base and address-base+count <= virtual+raw_size:
            off = raw+address-base-virtual
            return blob[off:off+count]
    raise AssertionError(hex(address))
addresses = {0x4e0b30, 0x4e1212, 0x4e1217, 0x4e1230, 0x4e1320,
    0x4e13b0, 0x4e1400, 0x4e1470, 0x4ec580, 0x532880, 0x4ee5e0}
ranges = [r for r in manifest['functions'] if int(r['address'], 16) in addresses]
assert len(ranges) == len(addresses)
for r in ranges:
    assert hashlib.sha256(read(int(r['address'], 16), r['size'])).hexdigest() == r['sha256']
cpu = Uc(UC_ARCH_X86, UC_MODE_32)
cpu.mem_map(base, size)
for _, virtual, raw_size, raw in sections:
    if raw_size:
        cpu.mem_write(base+virtual, blob[raw:raw+raw_size])
heap, stack, stop = 0x10000000, 0x20000000, 0x30000000
cpu.mem_map(heap, 0x10000)
cpu.mem_map(stack, 0x10000)
cpu.mem_map(stop, 0x1000)
v, a, h = heap+0x100, heap+0x1100, heap+0x2100
get = lambda p: struct.unpack('<I', cpu.mem_read(p, 4))[0]
put = lambda p, x: cpu.mem_write(p, struct.pack('<I', x & 0xffffffff))
sealed = {int(r['address'], 16): bytes(cpu.mem_read(int(r['address'], 16), r['size'])) for r in ranges}
entries, faults, rows = Counter(), [], []
current, events = {}, []

def hook(_cpu, address, _size, _):
    if address in sealed:
        entries[hex(address)] += 1
    if address == 0x4ee5e0:
        sp = cpu.reg_read(UC_X86_REG_ESP)
        assert [get(sp+4), get(sp+8)] == [v, current['damage'] & 0xffffffff]
        events.append('clear') # Observe only; original nil/Max tests are NOT hooked.
        return
    if any(start <= address < start+len(code) for start, code in sealed.items()):
        return
    sp = cpu.reg_read(UC_X86_REG_ESP)
    arg = lambda i: get(sp+4+4*i)
    result = 0
    if address == 0x4ff350:
        assert arg(0) in (v, a)
        result = bool(get(arg(0)+340) & (1 << arg(1)))
    elif address == 0x417da0:
        assert arg(0) == 1
        result = 1
    elif address == 0x40a5c0:
        assert arg(0) == 6144
    elif address == 0x5330c0:
        assert [arg(0), arg(1)] == [v, a]
        events.append('enemy')
    elif address == 0x534a40:
        assert arg(0) == v
    elif address == 0x501960:
        assert [arg(0), arg(1), arg(2), arg(3)] == [71, v, 0, 0]
        events.append('audio')
    elif address == 0x4ff5b0:
        assert [arg(0), arg(1)] == [v, 0]
        put(v+340, get(v+340) & ~1)
        events.append('buff')
    elif address == 0x532e20:
        assert [arg(0), arg(1)] == [v, a]
        events.append('sound')
    elif address == 0x52f710:
        assert [arg(0), arg(2), arg(3)] == [v, 2, a]
        put(arg(1), 0)
        events.append('shield')
    else:
        raise AssertionError('undeclared instruction ' + hex(address) + ' ' + current['name'])
    cpu.reg_write(UC_X86_REG_EAX, int(result))
    cpu.reg_write(UC_X86_REG_ESP, sp+4)
    cpu.reg_write(UC_X86_REG_EIP, get(sp))

def invalid(_cpu, access, address, size, value, _):
    faults.append(dict(case=current['name'], address=address, size=size,
        instruction=cpu.reg_read(UC_X86_REG_EIP)))
    return False

cpu.hook_add(UC_HOOK_CODE, hook)
cpu.hook_add(UC_HOOK_MEM_INVALID, invalid)
targets = [('PlayerStart', 714, 0, 16781893),
    ('InvisibleTeleportPentagram', 717, 4195328, 150995469)]
for name, type_id, target_class, base_flags in targets:
    for health in ['nil', 'zero-max']:
        for weapon in [False, True]:
            for gate in ['normal', 'no-update', 'dead', 'invulnerable', 'shield-zero']:
                for damage in [-3, 0, 1, 50]:
                    current = dict(name=f'{name}/{health}/weapon-{weapon}/{gate}/raw-{damage}',
                        target=name, type=type_id, class_bits=target_class,
                        base_flags=base_flags, health=health, weapon=weapon, gate=gate, damage=damage)
                    events = []
                    cpu.mem_write(heap, bytes([0xa7])*0x10000)
                    cpu.mem_write(v, bytes(0x300))
                    cpu.mem_write(a, bytes(0x300))
                    cpu.mem_write(h, struct.pack('<HHH', 120, 119, 0))
                    flags = base_flags | (2 if gate == 'no-update' else 0) | (0x8000 if gate == 'dead' else 0)
                    buffs = 1 | (1 << 22) | (1 << 23 if gate == 'invulnerable' else 0) | (1 << 26 if gate == 'shield-zero' else 0)
                    cpu.mem_write(v+4, struct.pack('<H', type_id))
                    put(v+8, target_class); put(v+16, flags); put(v+340, buffs)
                    put(v+524, 19); put(v+536, 77)
                    cpu.mem_write(v+528, struct.pack('<ff', 31, 47))
                    put(v+556, h if health == 'zero-max' else 0)
                    cpu.mem_write(a+4, struct.pack('<H', 706))
                    put(a+8, 1572865); put(a+12, 1); put(a+16, 553665028)
                    cpu.mem_write(a+72, struct.pack('<ff', 17, 19))
                    initial_target = bytes(cpu.mem_read(v, 0x300))
                    initial_source = bytes(cpu.mem_read(a, 0x300))
                    initial_health = bytes(cpu.mem_read(h, 6))
                    initial_heap = bytes(cpu.mem_read(heap, 0x10000))
                    put(0x84ea04, 1400)
                    initial_stack = stack+0x8000
                    cpu.mem_write(initial_stack, struct.pack('<IIIIII',
                        stop, v, a, a if weapon else 0, damage & 0xffffffff, 2))
                    cpu.reg_write(UC_X86_REG_ESP, initial_stack)
                    cpu.reg_write(UC_X86_REG_FPCW, 0x027f)
                    cpu.reg_write(UC_X86_REG_FPSW, 0)
                    cpu.reg_write(UC_X86_REG_FPTAG, 0xffff)
                    cpu.emu_start(0x4e0b30, stop, count=12000)
                    assert cpu.reg_read(UC_X86_REG_EIP) == stop and cpu.reg_read(UC_X86_REG_ESP) == initial_stack+4
                    after = bytes(cpu.mem_read(v, 0x300))
                    changed = [i for i, (x,y) in enumerate(zip(initial_target, after)) if x != y]
                    allowed = set(range(340,344)) | set(range(520,540))
                    assert set(changed) <= allowed
                    assert bytes(cpu.mem_read(a, 0x300)) == initial_source
                    assert bytes(cpu.mem_read(h, 6)) == initial_health
                    guarded = bytearray(cpu.mem_read(heap, 0x10000))
                    guarded[0x100:0x400] = initial_target
                    assert bytes(guarded) == initial_heap
                    rows.append(dict(**current, flags=flags, initial_buffs=buffs,
                        result=cpu.reg_read(UC_X86_REG_EAX), buffs=get(v+340),
                        attributed=get(v+520)==a, damage_type=get(v+524), frame=get(v+536),
                        position_bits=[get(v+528),get(v+532)], events=events.copy(),
                        health_unchanged=True, source_and_guards_unchanged=True))
assert not faults
assert all(bytes(cpu.mem_read(start,len(code))) == code for start,code in sealed.items())
print(json.dumps(dict(original_sha256=manifest['file_sha256'],
    function_sha256=hashlib.sha256(read(0x4e0b30,1792)).hexdigest(),
    health_gate_sha256=hashlib.sha256(read(0x4ee5e0,264)).hexdigest(),
    original_ranges=ranges, rows=rows, unhooked_entries=dict(entries),
    executable_bytes_unchanged=True, memory_faults=faults,
    scope=__doc__.strip()), separators=(',', ':')))
