"""Run original Vortex, projection and float-to-int code with declared draw services.

Ordered color/geometry/state oracle; not an original Windows framebuffer.
"""
import collections
import hashlib
import json
import struct
import sys
from pathlib import Path

from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import UC_X86_REG_EAX, UC_X86_REG_EIP, UC_X86_REG_ESP, UC_X86_REG_FPCW

pe = Path(sys.argv[1]).read_bytes()
assert hashlib.sha256(pe).hexdigest() == "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a"
h = struct.unpack_from("<I", pe, 0x3c)[0]
o = h + 24
base = struct.unpack_from("<I", pe, o + 28)[0]
size = struct.unpack_from("<I", pe, o + 56)[0]
s = o + struct.unpack_from("<H", pe, h + 20)[0]
sections = [struct.unpack_from("<IIII", pe, s + 40*i + 8) for i in range(struct.unpack_from("<H", pe, h + 6)[0])]
cpu = Uc(UC_ARCH_X86, UC_MODE_32)
cpu.mem_map(base, (size + 4095) & ~4095)
for _, virtual, length, raw in sections:
    cpu.mem_write(base + virtual, pe[raw:raw + length])
cpu.mem_map(0x20000000, 0x10000)
VP, DR, STACK, STOP = 0x20000020, 0x20001020, 0x20008000, 0x20009000
ranges = {0x4b9f50: 723, 0x4739e0: 46, 0x419a70: 17}
sealed = {at: bytes(cpu.mem_read(at, n)) for at, n in ranges.items()}
table = bytes(cpu.mem_read(0x5b5e58, 2048))
assert hashlib.sha256(table).hexdigest() == "69f2eae1505c2e27de6bf104de077085994f00e4eb1d4ab3c68bd13916d7b079"
float_constants = bytes(cpu.mem_read(0x583a88, 8))
entries, services = collections.Counter(), collections.Counter()
events, points, faults = [], [], []
color = 0

def put(at, value):
    cpu.mem_write(at, struct.pack("<I", value & 0xffffffff))

def get(at):
    return struct.unpack("<I", cpu.mem_read(at, 4))[0]

def args(n):
    return struct.unpack("<" + "I"*n, cpu.mem_read(cpu.reg_read(UC_X86_REG_ESP) + 4, 4*n))

def signed(v):
    return v if v < 0x80000000 else v - 0x100000000

def ret(value=0):
    sp = cpu.reg_read(UC_X86_REG_ESP)
    cpu.reg_write(UC_X86_REG_EAX, value)
    cpu.reg_write(UC_X86_REG_EIP, get(sp))
    cpu.reg_write(UC_X86_REG_ESP, sp + 4)

def hook(_cpu, address, _size, _):
    global color
    if address in ranges:
        entries[hex(address)] += 1
    if any(at <= address < at + n for at, n in ranges.items()):
        return
    if address == 0x4344a0:
        assert args(3) == (170, 170, 170)
        events.append(["rgb", 170, 170, 170, 0x56b556b5])
        services["rgb"] += 1
        ret(0x56b556b5)
    elif address == 0x4b6720:
        point, cl, inner, outer = args(4)
        x, y = struct.unpack("<ii", cpu.mem_read(point, 8))
        events.append(["glow", x, y, cl, inner, outer])
        services["glow"] += 1
        ret()
    elif address == 0x434460:
        color, = args(1)
        events.append(["color", color])
        services["color"] += 1
        ret()
    elif address == 0x499b70:
        x, y, radius = args(3)
        events.append(["dot", signed(x), signed(y), radius, color])
        services["dot"] += 1
        ret()
    elif address == 0x49f500:
        x, y = [signed(v) for v in args(2)]
        points.append([x, y])
        events.append(["add", x, y])
        services["add"] += 1
        ret()
    elif address == 0x49e4b0:
        assert len(points) == 2
        events.append(["line", *points[0], *points[1], color])
        points.clear()
        services["line"] += 1
        ret(1)
    elif address == 0x45a4e0:
        assert args(1) == (DR,)
        events.append(["delete"])
        services["delete"] += 1
        ret()
    elif address != STOP:
        raise AssertionError("undeclared execution " + hex(address))

def invalid(_cpu, access, address, length, value, _):
    faults.append(dict(case=case, address=address, size=length, instruction=cpu.reg_read(UC_X86_REG_EIP)))
    return False

cpu.hook_add(UC_HOOK_CODE, hook)
cpu.hook_add(UC_HOOK_MEM_INVALID, invalid)
cases = []
for angle in range(256):
    for spin in (-2, 2):
        for glow, center in ((0x03e003e0, 0x001f001f), (0x56b556b5, 0x7fff7fff)):
            cases.append(dict(angle=angle, spin=spin, radius=12, rise=1, z=8, x=164, y=264, glow=glow, center=center))
for name, values in (
    ("outside", dict(x=300)),
    ("left-boundary", dict(x=88)),
    ("top-boundary", dict(y=208)),
    ("expiry", dict(z=399, y=663)),
    ("negative-z-wrap", dict(z=65535)),
    ("signed-z-wrap", dict(z=32767, y=33031)),
):
    cases.append(dict(name=name, **(dict(angle=0, spin=2, radius=12, rise=1, z=8, x=164, y=264, glow=0x03e003e0, center=0x001f001f) | values)))

rows = []
for case, inp in enumerate(cases):
    cpu.mem_write(0x20000000, bytes(0x8000))
    for at, n in ((VP, 48), (DR, 512)):
        cpu.mem_write(at - 16, b"\x5a"*16)
        cpu.mem_write(at + n, b"\x5a"*16)
    for off, value in enumerate((7, 11, 135, 139, 100, 200)):
        put(VP + 4*off, value)
    put(DR + 12, inp["x"])
    put(DR + 16, inp["y"])
    cpu.mem_write(DR + 104, struct.pack("<H", inp["z"]))
    for off, value in ((432, inp["glow"]), (436, inp["center"]), (440, inp["x"]), (444, inp["y"])):
        put(DR + off, value)
    cpu.mem_write(DR + 448, bytes((inp["angle"], inp["spin"] & 255, inp["radius"], inp["rise"])))
    put(0x7151b0, 0)
    put(0x7151ac, 0)
    before_dr = bytes(cpu.mem_read(DR - 16, 544))
    before_vp = bytes(cpu.mem_read(VP - 16, 80))
    events, points, color = [], [], 0x7fff7fff
    cpu.reg_write(UC_X86_REG_FPCW, 0x037f)
    cpu.mem_write(STACK, struct.pack("<III", STOP, VP, DR))
    cpu.reg_write(UC_X86_REG_ESP, STACK)
    cpu.emu_start(0x4b9f50, STOP, count=5000)
    returned = cpu.reg_read(UC_X86_REG_EAX)
    assert cpu.reg_read(UC_X86_REG_EIP) == STOP and cpu.reg_read(UC_X86_REG_ESP) == STACK + 4
    assert returned in (0, 1) and not points and cpu.reg_read(UC_X86_REG_FPCW) == 0x037f
    after_dr = bytes(cpu.mem_read(DR - 16, 544))
    mutable = {16 + off for off in (104, 105, 448, 450)}
    assert all(a == b for i, (a, b) in enumerate(zip(before_dr, after_dr)) if i not in mutable)
    assert bytes(cpu.mem_read(VP - 16, 80)) == before_vp
    assert all(bytes(cpu.mem_read(at, n)) == sealed[at] for at, n in ranges.items())
    assert bytes(cpu.mem_read(0x5b5e58, 2048)) == table and bytes(cpu.mem_read(0x583a88, 8)) == float_constants
    rows.append(dict(input=inp, events=events, returned=returned, final_color=color,
                     final_z=struct.unpack("<H", cpu.mem_read(DR + 104, 2))[0],
                     final_motion=get(DR + 448), guards_and_other_input_unchanged=True))
assert not faults and entries["0x4b9f50"] == len(rows)
print(json.dumps(dict(original_sha256=hashlib.sha256(pe).hexdigest(),
    functions={hex(at): dict(bytes=n, sha256=hashlib.sha256(sealed[at]).hexdigest()) for at, n in ranges.items()},
    direction_table_sha256=hashlib.sha256(table).hexdigest(), float_constants_hex=float_constants.hex(),
    viewport=[7, 11, 135, 139, 100, 200], fpu_control=0x037f, rows=rows,
    unhooked_entries=dict(entries), declared_services=dict(services), memory_faults=faults,
    executable_bytes_unchanged=True, scope="unchanged Vortex/projection/nearest-even conversion with declared RGB5551 and draw/delete services; not original Windows raster/full world"), separators=(",", ":")))
