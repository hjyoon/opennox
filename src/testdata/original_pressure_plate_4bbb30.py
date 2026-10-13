"""Execute unchanged PE32 PressurePlateDraw and x87 float conversion.

Only RGB/point/line services are declared. Endpoints and low-byte RGB are
original results; this is not an original Windows framebuffer/world oracle.
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
ranges = {0x4bbb30: 504, 0x419a70: 17}
sealed = {at: bytes(cpu.mem_read(at, length)) for at, length in ranges.items()}
assert hashlib.sha256(sealed[0x4bbb30]).hexdigest() == "ab498ba330038e5d8ac268f6c6d8d94b10b5ee77dfda4be07487f2b5a8f32659"
assert hashlib.sha256(sealed[0x419a70]).hexdigest() == "10e536e2a987ea0fe14c91f007b2d5361cd87d4f0ca4cf5385fc82e53913d186"
cpu.reg_write(UC_X86_REG_FPCW, 0x037f)
entries, services = collections.Counter(), collections.Counter()
points, lines, colors, events, faults = [], [], [], [], []

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
    if address in ranges:
        entries[hex(address)] += 1
    if any(at <= address < at + length for at, length in ranges.items()):
        return
    if address == 0x434430:
        rgb = [v & 0xff for v in args(3)]
        colors.append(rgb)
        events.append(["rgb", *rgb])
        services["rgb"] += 1
        ret()
    elif address == 0x49f500:
        points.append([signed(v) for v in args(2)])
        assert len(points) <= 2
        services["point"] += 1
        ret()
    elif address == 0x49e4b0:
        assert len(points) == 2
        lines.append(points.copy())
        points.clear()
        events.append(["line"])
        services["line"] += 1
        ret(1)
    elif address != STOP:
        raise AssertionError("undeclared execution " + hex(address))

def invalid(_cpu, access, address, length, value, _):
    faults.append(dict(case=name, address=address, size=length, instruction=cpu.reg_read(UC_X86_REG_EIP)))
    return False

def bits(v):
    return struct.unpack("<I", struct.pack("<f", v))[0]

shapes = [
    ("diamond", [bits(v) for v in [-8, 0, 0, 8, 0, -8, 8, 0]]),
    ("ties-even", [bits(v) for v in [-8.5, -0.5, -2.5, 7.5, 2.5, -7.5, 8.5, 0.5]]),
    ("ties-odd", [bits(v) for v in [-9.5, -1.5, -3.5, 8.5, 3.5, -8.5, 9.5, 1.5]]),
    ("zero", [0, 0x80000000, 0, 0x80000000, 0, 0, 0, 0]),
    ("fractions", [bits(v) for v in [-7.25, -0.25, 1.25, 7.75, -1.25, -7.75, 7.25, 0.25]]),
    ("asymmetric", [bits(v) for v in [-12.75, -4.25, -3.75, 10.25, 4.75, -10.25, 12.75, 4.25]]),
    ("nan-inf", [0x7fc12345, 0x7f800000, 0xff800000, 0x80000000, 0x7fa12345, 0xffc12345, 0, 0x3f800000]),
    ("int-high", [0x4effffff, 0x4effffff, 0x4f000000, 0x4f000000, 0x4effffff, 0x4f000001, 0, 0x4effffff]),
    ("int-low", [0xcf000000, 0xcf000000, 0xceffffff, 0xceffffff, 0xcf000001, 0xff800000, 0xcf000000, 0])
]
rgbs = [("zero", [0]*6)]
for i in range(6):
    rgb = [0]*6
    rgb[i] = 255 if i % 2 == 0 else 129
    rgbs.append(("channel-" + str(i), rgb))
rgbs.append(("mixed", [17, 133, 255, 249, 65, 8]))
cpu.hook_add(UC_HOOK_CODE, hook)
cpu.hook_add(UC_HOOK_MEM_INVALID, invalid)
rows = []
for shape, fields in shapes:
    for color, rgb in rgbs:
        name = shape + "/" + color
        cpu.mem_write(0x20000000, bytes(0x8000))
        for at, length in ((VP, 48), (DR, 512)):
            cpu.mem_write(at - 16, b"\x5a"*16)
            cpu.mem_write(at + length, b"\x5a"*16)
        for at, value in ((VP, 7), (VP + 4, 11), (VP + 16, 13), (VP + 20, 23),
                          (DR + 12, 6), (DR + 16, 12)):
            put(at, value)
        for i, word in enumerate(fields):
            put(DR + 64 + 4*i, word)
        cpu.mem_write(DR + 432, bytes(rgb) + b"\xef\xbe")
        immutable = {at: bytes(cpu.mem_read(at - 16, length + 32)) for at, length in ((VP, 48), (DR, 512))}
        points, lines, colors, events = [], [], [], []
        cpu.mem_write(STACK, struct.pack("<III", STOP, VP, DR))
        cpu.reg_write(UC_X86_REG_ESP, STACK)
        cpu.emu_start(0x4bbb30, STOP, count=5000)
        visible = any(rgb)
        assert cpu.reg_read(UC_X86_REG_EIP) == STOP and cpu.reg_read(UC_X86_REG_ESP) == STACK + 4
        assert cpu.reg_read(UC_X86_REG_EAX) == 1 and len(lines) == (8 if visible else 0) and not points
        assert colors == ([rgb[:3], rgb[3:]] if visible else [])
        assert cpu.reg_read(UC_X86_REG_FPCW) == 0x037f
        assert all(bytes(cpu.mem_read(at, length)) == sealed[at] for at, length in ranges.items())
        assert all(bytes(cpu.mem_read(at - 16, len(data))) == data for at, data in immutable.items())
        rows.append(dict(name=name, shape_bits=fields, rgb=rgb, visible=visible, colors=colors, lines=lines, events=events, returned=1, guarded_bytes_unchanged=True))
assert not faults and entries == {"0x4bbb30": 72, "0x419a70": 504}
assert services == {"rgb": 126, "point": 1008, "line": 504}
print(json.dumps(dict(original_sha256=hashlib.sha256(pe).hexdigest(),
                      functions={hex(at): dict(bytes=length, sha256=hashlib.sha256(sealed[at]).hexdigest()) for at, length in ranges.items()},
                      original_base=[0, 0], x87_control_word=0x037f, rows=rows,
                      unhooked_entries=dict(entries), declared_services=dict(services),
                      memory_faults=faults, executable_bytes_unchanged=True,
                      scope="unchanged original pressure plate and x87 nearest-even conversion, low-byte RGB and ordered endpoints; declared line services, not original Windows raster/full world"), separators=(",", ":")))
