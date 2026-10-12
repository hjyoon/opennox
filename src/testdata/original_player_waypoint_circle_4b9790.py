"""Run the unchanged PE32 waypoint and circle bodies with declared line services.

This verifies integer endpoints and call order, not a Windows framebuffer.
The direction table is read unchanged from the original executable.
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
ranges = {0x4b9790: 269, 0x4c32a0: 237}
sealed = {at: bytes(cpu.mem_read(at, length)) for at, length in ranges.items()}
assert hashlib.sha256(sealed[0x4b9790]).hexdigest() == "0996ff27d5007997d3f978a087a68e4cd9260eab69d22e0a207fdffd77163e3b"
assert hashlib.sha256(sealed[0x4c32a0]).hexdigest() == "52173cb750585d45c3c31c27bb9b355ee39f93ba968277948613261baa398aa3"
table = bytes(cpu.mem_read(0x5b5e58, 2048))
assert hashlib.sha256(table).hexdigest() == "69f2eae1505c2e27de6bf104de077085994f00e4eb1d4ab3c68bd13916d7b079"
entries, services = collections.Counter(), collections.Counter()
points, lines, events, faults = [], [], [], []

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
    if address == 0x434460:
        assert args(1) == (0x03e0,)
        events.append(["color", 0x03e0])
        services["color"] += 1
        ret()
    elif address == 0x434560:
        value, = args(1)
        assert value in (0, 1)
        events.append(["alpha", value])
        services["alpha"] += 1
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
    faults.append(dict(frame=frame, address=address, size=length, instruction=cpu.reg_read(UC_X86_REG_EIP)))
    return False

cpu.hook_add(UC_HOOK_CODE, hook)
cpu.hook_add(UC_HOOK_MEM_INVALID, invalid)
rows, circle, control_events = [], None, None
for frame in list(range(128)) + [128, 256, 0x80000000, 0xffffffff]:
    cpu.mem_write(0x20000000, bytes(0x8000))
    for at, length in ((VP, 48), (DR, 512)):
        cpu.mem_write(at - 16, b"\x5a"*16)
        cpu.mem_write(at + length, b"\x5a"*16)
    for at, value in ((VP, 7), (VP + 4, 11), (VP + 16, 13), (VP + 20, 23),
                      (DR + 12, 13), (DR + 16, 23), (0x84ea04, frame), (0x85b7a8, 0x03e0)):
        put(at, value)
    immutable = {at: bytes(cpu.mem_read(at - 16, length + 32)) for at, length in ((VP, 48), (DR, 512))}
    points, lines, events = [], [], []
    cpu.mem_write(STACK, struct.pack("<III", STOP, VP, DR))
    cpu.reg_write(UC_X86_REG_ESP, STACK)
    cpu.emu_start(0x4b9790, STOP, count=5000)
    assert cpu.reg_read(UC_X86_REG_EIP) == STOP and cpu.reg_read(UC_X86_REG_ESP) == STACK + 4
    assert cpu.reg_read(UC_X86_REG_EAX) == 1 and len(lines) == 21 and not points
    assert all(bytes(cpu.mem_read(at, length)) == sealed[at] for at, length in ranges.items())
    assert bytes(cpu.mem_read(0x5b5e58, 2048)) == table and get(0x84ea04) == frame
    assert all(bytes(cpu.mem_read(at - 16, len(data))) == data for at, data in immutable.items())
    if circle is None:
        circle, control_events = lines[:16], events.copy()
    assert lines[:16] == circle and events == control_events
    rows.append(dict(frame=frame, star=lines[16:], returned=1, guarded_bytes_unchanged=True))
assert not faults and all(entries[hex(at)] == len(rows) for at in ranges)
print(json.dumps(dict(original_sha256=hashlib.sha256(pe).hexdigest(),
                      functions={hex(at): dict(bytes=length, sha256=hashlib.sha256(sealed[at]).hexdigest()) for at, length in ranges.items()},
                      direction_table_sha256=hashlib.sha256(table).hexdigest(),
                      circle=circle, events=control_events, rows=rows, unhooked_entries=dict(entries),
                      declared_services=dict(services), memory_faults=faults, executable_bytes_unchanged=True,
                      scope="unchanged original waypoint and circle integer endpoints, ordered line services and static stock direction table; not original Windows raster/full world"), separators=(",", ":")))
