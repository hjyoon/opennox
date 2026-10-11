"""Execute the unchanged PE32 NPCDraw body against explicit renderer services.

The material/image/equipment/buff/HP boundaries are supplied, not a Windows
renderer or a complete original world. No code bytes or return paths are patched.
"""
import collections
import hashlib
import json
import struct
import sys
from pathlib import Path

from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import UC_X86_REG_EAX, UC_X86_REG_EIP, UC_X86_REG_ESP

pe = Path(sys.argv[1]).read_bytes()
pe_sha = hashlib.sha256(pe).hexdigest()
assert pe_sha == "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a"
h = struct.unpack_from("<I", pe, 0x3c)[0]
o = h + 24
base = struct.unpack_from("<I", pe, o + 28)[0]
image_size = struct.unpack_from("<I", pe, o + 56)[0]
s = o + struct.unpack_from("<H", pe, h + 20)[0]
sections = [struct.unpack_from("<IIII", pe, s + 40*i + 8) for i in range(struct.unpack_from("<H", pe, h + 6)[0])]
cpu = Uc(UC_ARCH_X86, UC_MODE_32)
cpu.mem_map(base, (image_size + 4095) & ~4095)
for _, virtual, size, raw in sections:
    cpu.mem_write(base + virtual, pe[raw:raw + size])
cpu.mem_map(0x20000000, 0x20000)
VP, DR, LOCAL, NPC, DD, NAKED, FRAMES, SKULL, CROWN, CALLBACK, STOP, STACK = (
    0x20000020, 0x20001020, 0x20002020, 0x20003020, 0x20004020,
    0x20008020, 0x20009020, 0x2000a020, 0x2000b020, 0x2000c020,
    0x2000d020, 0x20018000,
)
START, LENGTH, WHITE, BLUE = 0x4b9050, 1047, 0x7fff, 0x001f
sealed = bytes(cpu.mem_read(START, LENGTH))
entries, services = collections.Counter(), collections.Counter()
faults, events, effect_palettes, effect_drawables = [], [], [], []

def put(at, value):
    cpu.mem_write(at, struct.pack("<I", value & 0xffffffff))

def get(at):
    return struct.unpack("<I", cpu.mem_read(at, 4))[0]

def args(count):
    return struct.unpack("<" + "I"*count, cpu.mem_read(cpu.reg_read(UC_X86_REG_ESP) + 4, 4*count))

def ret(value=0):
    sp = cpu.reg_read(UC_X86_REG_ESP)
    cpu.reg_write(UC_X86_REG_EAX, value)
    cpu.reg_write(UC_X86_REG_EIP, get(sp))
    cpu.reg_write(UC_X86_REG_ESP, sp + 4)

def split(word):
    return ((word >> 10) & 31)*8, ((word >> 5) & 31)*8, (word & 31)*8

def hook(_cpu, address, _size, _):
    if address == START:
        entries[hex(address)] += 1
    if START <= address < START + LENGTH:
        return
    if address == 0x40a5c0:
        assert args(1) == (0x200000,)
        services["preview-flag"] += 1
        ret(0)
    elif address == 0x49a340:
        assert args(1) == (0x1234,)
        services["npc-by-id"] += 1
        ret(0 if current["missing_npc"] else NPC)
    elif address == 0x4356c0:
        dr, enchant = args(2)
        assert dr == DR and enchant in (23, 25, 16, 30)
        services["has-enchant"] += 1
        ret(bool(current["buffs"] & (1 << enchant)))
    elif address == 0x4341d0:
        index, word = args(2)
        assert 0 <= index < 16 and word <= 0xffff
        palette[index] = word
        events.append(["material", index, word])
        services["material"] += 1
        ret()
    elif address == 0x434480:
        word, r, g, b = args(4)
        for at, value in zip((r, g, b), split(word)):
            put(at, value)
        services["declared-5551-split"] += 1
        ret()
    elif address == 0x4344a0:
        r, g, b = args(3)
        services["declared-5551-pack"] += 1
        ret(((r & 0xf8) << 7) | ((g & 0xf8) << 2) | ((b & 0xf8) >> 3))
    elif address == 0x49a4b0:
        assert args(3) == (DR, 0, 0)
        services["declared-animation-selector"] += 1
        ret(0)
    elif address == 0x4bc5d0:
        assert args(2) == (DR, DD + 4)
        services["declared-frame-selector"] += 1
        ret(0)
    elif address == 0x4c4770:
        assert args(3) == (VP, DR, 0)
        events.append(["body-image", 0])
        services["declared-null-body-image"] += 1
        ret()
    elif address in (0x4b8d40, 0x4b8960):
        vp, dr, equip, data, anim, fi = args(6)
        assert (vp, dr, equip, anim, fi) == (VP, DR, 0, DD + 4, 0)
        name = "weapon" if address == 0x4b8d40 else "armor"
        assert data == NPC + (0x20 if name == "weapon" else 0x2a8)
        events.append([name, 0])
        services["declared-empty-" + name] += 1
        ret()
    elif address == 0x437250:
        services["declared-effect-viewport"] += 1
        ret(VP)
    elif address == CALLBACK:
        vp, drawable = args(2)
        assert vp == VP and drawable in (SKULL, CROWN)
        name = "skull" if drawable == SKULL else "crown"
        assert (get(drawable + 12), get(drawable + 16)) == (57, 61)
        effect_palettes.append(palette.copy())
        effect_drawables.append(name)
        events.append(["effect", name])
        services["declared-effect-draw"] += 1
        ret(1)
    elif address == 0x4951f0:
        assert args(1) == (0x1234,)
        services["declared-not-ally"] += 1
        ret(0)
    elif address != STOP:
        raise AssertionError("undeclared execution " + hex(address) + " in " + current["name"])

def invalid(_cpu, access, address, size, value, _):
    faults.append(dict(case=current["name"], address=address, size=size, instruction=cpu.reg_read(UC_X86_REG_EIP)))
    return False

cpu.hook_add(UC_HOOK_CODE, hook)
cpu.hook_add(UC_HOOK_MEM_INVALID, invalid)
initial = [((i+1) << 10) | ((31-i) << 5) | (2*i+1) for i in range(16)]
colors = [((3+3*i) << 10) | ((4+4*i) << 5) | (5+2*i) for i in range(6)]
cases = []
for color_name, buffs, poison, frame in (
    ("ordinary", 0, 0, 10), ("poison", 0, 1, 10),
    ("frozen", 1 << 25, 0, 10), ("invulnerable-even", 1 << 23, 0, 10),
    ("invulnerable-odd", 1 << 23, 0, 11),
    ("invulnerable-freeze-poison", (1 << 23) | (1 << 25), 1, 10),
):
    for effect_name, effect_buffs in (("none", 0), ("skull", 1 << 16), ("crown", 1 << 30), ("both", (1 << 16) | (1 << 30))):
        for direction in (0, 4, 6):
            cases.append(dict(name=color_name + "/" + effect_name + "/dir-" + str(direction),
                              buffs=buffs | effect_buffs, poison=poison, frame=frame,
                              direction=direction, count=1, missing_npc=False))
for name, count, missing in (("missing-NPC", 1, True), ("zero-frame-count", 0, False)):
    cases.append(dict(name=name, buffs=(1 << 16) | (1 << 30), poison=0, frame=10,
                      direction=0, count=count, missing_npc=missing))

rows = []
for current in cases:
    cpu.mem_write(0x20000000, bytes(0x18000))
    guard_regions = ((VP, 48), (DR, 0x200), (LOCAL, 0x200), (NPC, 0x524),
                     (DD, 264*51 + 4), (NAKED, 40), (FRAMES, 4), (SKULL, 0x200), (CROWN, 0x200))
    for at, size in guard_regions:
        cpu.mem_write(at - 16, b"\x5a"*16)
        cpu.mem_write(at + size, b"\x5a"*16)
    for at, value in ((0x852980, LOCAL), (LOCAL + 0x130, DD), (DR + 0x80, 0x1234),
                      (DR + 12, 63), (DR + 16, 123), (DR + 0x114, 0), (NPC + 0x520, current["poison"]),
                      (0x84ea04, current["frame"]), (0x83c8c0, WHITE), (0x85b7d0, BLUE),
                      (0x715198, SKULL), (0x71519c, CROWN),
                      (SKULL + 0x12c, CALLBACK), (CROWN + 0x12c, CALLBACK),
                      (VP, 7), (VP + 4, 11), (VP + 16, 13), (VP + 20, 23), (DD + 4 + 0x30, NAKED)):
        put(at, value)
    cpu.mem_write(DR + 0x129, bytes((current["direction"],)))
    cpu.mem_write(DD + 4 + 40, struct.pack("<H", current["count"]))
    for i, word in enumerate(colors):
        put(NPC + 8 + 4*i, word)
    for i in range(9):
        put(NAKED + 4 + 4*i, FRAMES)
    immutable = {at: bytes(cpu.mem_read(at, size)) for at, size in guard_regions if at not in (SKULL, CROWN)}
    events, effect_palettes, effect_drawables = [], [], []
    palette = initial.copy()
    cpu.mem_write(STACK, struct.pack("<III", STOP, VP, DR))
    cpu.reg_write(UC_X86_REG_ESP, STACK)
    cpu.emu_start(START, STOP, count=5000)
    assert cpu.reg_read(UC_X86_REG_EIP) == STOP and cpu.reg_read(UC_X86_REG_ESP) == STACK + 4
    assert bytes(cpu.mem_read(START, LENGTH)) == sealed
    assert all(bytes(cpu.mem_read(at, len(data))) == data for at, data in immutable.items())
    assert all(bytes(cpu.mem_read(at-16, 16)) == b"\x5a"*16 and bytes(cpu.mem_read(at+size, 16)) == b"\x5a"*16 for at, size in guard_regions)
    material_events = [e for e in events if e[0] == "material"]
    if current["count"] and not current["missing_npc"]:
        assert [e[1] for e in material_events] == list(range(1, 7)) + list(range(6))
        assert palette[:6] == [WHITE]*6 and palette[7:] == initial[7:]
        expected_effects = [name for bit, name in ((16, "skull"), (30, "crown")) if current["buffs"] & (1 << bit)]
        assert effect_drawables == expected_effects
        assert all(p[:1] == initial[:1] and p[1:7] == [e[2] for e in material_events[:6]] and p[7:] == initial[7:] for p in effect_palettes)
        last_effect = max((i for i, e in enumerate(events) if e[0] == "effect"), default=-1)
        first_reset = next(i for i, e in enumerate(events) if e[:2] == ["material", 0])
        assert last_effect < first_reset
    rows.append(dict(**current, initial=initial, npc_colors=colors, returned=cpu.reg_read(UC_X86_REG_EAX),
                     effect_names=effect_drawables.copy(), effect_palettes=effect_palettes.copy(),
                     final_palette=palette.copy(), events=events.copy(), guarded_bytes_unchanged=True))
assert not faults and entries[hex(START)] == len(cases)
print(json.dumps(dict(original_sha256=pe_sha, function_sha256=hashlib.sha256(sealed).hexdigest(),
                      function_start=START, function_bytes=LENGTH, rows=rows,
                      unhooked_entries=dict(entries), declared_services=dict(services), memory_faults=faults,
                      executable_bytes_unchanged=True,
                      scope="unchanged NPCDraw with declared renderer/material, empty equipment, animation, 5551 split/pack, cached top-effect draw and non-ally services; not stock NPC pixels, HP bar, preview, negative-frame removal or full original world"), separators=(",", ":")))
