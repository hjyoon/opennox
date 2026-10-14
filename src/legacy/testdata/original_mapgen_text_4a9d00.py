"""Execute original map-generation text entry with declared UI services.

Presentation begin/end and copy are recorded, not Windows/GPU emulated.
The decoded C short return remains a separate compatibility contract: both
original callers ignore it, and PE32 EAX comes from the last platform service.
"""
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
optional = h + 24
base = struct.unpack_from("<I", pe, optional + 28)[0]
size = struct.unpack_from("<I", pe, optional + 56)[0]
sections = optional + struct.unpack_from("<H", pe, h + 20)[0]
cpu = Uc(UC_ARCH_X86, UC_MODE_32)
cpu.mem_map(base, (size + 4095) & ~4095)
for i in range(struct.unpack_from("<H", pe, h + 6)[0]):
    _, virtual, length, raw = struct.unpack_from("<IIII", pe, sections + i*40 + 8)
    cpu.mem_write(base + virtual, pe[raw:raw+length])
cpu.mem_map(0x20000000, 0x20000)
ENTRY, STACK, STOP, PACKET = 0x4a9d00, 0x20008000, 0x20009000, 0x20010001
TEXT, SEQUENCE, FRAME = 0x714170, 0x714178, 0x71417c
code = bytes(cpu.mem_read(ENTRY, 395))
assert code[-1] == 0xc3
strings = {i: 0x20011000 + i*256 for i in range(5)}
for i, address in strings.items():
    cpu.mem_write(address, ("text-" + str(i)).encode("utf-16le") + b"\0\0")
services = {
    0x452d80: ("sound", 2), 0x430b50: ("clip", 4), 0x440900: ("clear", 0),
    0x47d2c0: ("image", 3), 0x40f1d0: ("load", 4), 0x43f840: ("measure", 5),
    0x43f320: ("fontheight", 1), 0x434390: ("textcolor", 1),
    0x4343e0: ("color", 1), 0x43f730: ("text", 4),
    0x48a220: ("begin", 0), 0x4ad170: ("copy", 0), 0x48a290: ("end", 0),
}
events, platform, faults, current = [], [], [], {}
entries = 0

def word(at):
    return struct.unpack("<I", cpu.mem_read(at, 4))[0]

def put(at, value):
    cpu.mem_write(at, struct.pack("<I", value & 0xffffffff))

def signed(value):
    return value if value < 0x80000000 else value - 0x100000000

def cstr(at):
    data = bytearray()
    while bytes(cpu.mem_read(at + len(data), 1)) != b"\0":
        data += cpu.mem_read(at + len(data), 1)
        assert len(data) < 160
    return data.decode("ascii")

def kind(address):
    assert address in strings.values(), "incorrect cached text " + hex(address)
    return next(i for i, at in strings.items() if at == address)

def hook(_cpu, address, length, _):
    global entries
    if address == ENTRY:
        entries += 1
    if ENTRY <= address < ENTRY + len(code) or address == STOP:
        return
    name, count = services[address]
    sp = cpu.reg_read(UC_X86_REG_ESP)
    args = list(struct.unpack("<" + "I"*count, cpu.mem_read(sp + 4, count*4))) if count else []
    value = 0
    if name == "load":
        assert args[1] == 0
        key, source, line = cstr(args[0]), cstr(args[2]), args[3]
        k = {"Generating": 1, "Assembling": 2, "Populating": 3}[key]
        events.append([name, k, source, line])
        value = strings[k]
    elif name == "measure":
        assert args[0] == args[3] == args[4] == 0
        events.append([name, kind(args[1])])
        put(args[2], current["measured"])
        if current["replace_text"]:
            put(TEXT, strings[4])
        value = 1
    elif name == "fontheight":
        events.append([name, args[0]])
        value = current["font"]
    elif name == "text":
        events.append([name, args[0], kind(args[1]), signed(args[2]), signed(args[3])])
        value = 777
    elif name in ("begin", "end"):
        platform.append([name])
        value = 0xabcdef01
    else:
        if name in ("clip", "image"):
            args = [args[0]] + [signed(v) for v in args[1:]] if name == "image" else [signed(v) for v in args]
        events.append([name] + args)
    cpu.reg_write(UC_X86_REG_EAX, value & 0xffffffff)
    cpu.reg_write(UC_X86_REG_EIP, word(sp))
    cpu.reg_write(UC_X86_REG_ESP, sp + 4)

def invalid(_cpu, access, address, length, value, _):
    faults.append(dict(address=address, pc=cpu.reg_read(UC_X86_REG_EIP)))
    return False

cpu.hook_add(UC_HOOK_CODE, hook)
cpu.hook_add(UC_HOOK_MEM_INVALID, invalid)
inputs = []
for op in (0x9b, 0x9c, 0x9d):
    for frame in range(32):
        inputs.append(dict(op=op, sequence=frame+1, cached=0xffffffff, frame=frame,
                           width=640, height=480, measured=17, font=13, initial=0, replace_text=False))
    for width, height in ((1, 1), (641, 481), (1024, 768), (-25, -31)):
        for sequence in (0, 1, 0x7fff, 0x8000, 0xffff):
            for duplicate in (False, True):
                inputs.append(dict(op=op, sequence=sequence, cached=sequence if duplicate else 0xffffffff,
                    frame=31, width=width, height=height, measured=23, font=7, initial=2, replace_text=False))
for op in (0, 0xff):
    for initial in range(4):
        inputs.append(dict(op=op, sequence=71, cached=70, frame=11, width=1024, height=768,
                           measured=19, font=20, initial=initial, replace_text=False))
for op in (0x9b, 0x9c, 0x9d):
    inputs.append(dict(op=op, sequence=87, cached=86, frame=31, width=641, height=481,
                       measured=0, font=0, initial=0, replace_text=True))
rows = []
for current in inputs:
    observed = []
    for pattern in (0, 0x5a, 0xff):
        events, platform = [], []
        cpu.mem_write(0x7140e0 - 32, bytes([pattern]) * (160 + 64))
        for i in range(32):
            put(0x7140e0 + 4*i, 0x11000000 + i)
        for i in range(4):
            put(0x714160 + 4*i, 0x22000000 + i)
        put(TEXT, strings[current["initial"]])
        put(SEQUENCE, current["cached"])
        put(FRAME, current["frame"])
        put(0x9756cc, current["width"])
        put(0x975d24, current["height"])
        put(0x83c8c0, 0x7fff)
        put(0x85297c, 0x03e0)
        packet = struct.pack("<BH", current["op"], current["sequence"])
        cpu.mem_write(PACKET - 16, bytes([pattern]) * 35)
        cpu.mem_write(PACKET, packet)
        cpu.mem_write(STACK - 256, bytes([pattern]) * 512)
        cpu.mem_write(STACK, struct.pack("<II", STOP, PACKET))
        cpu.reg_write(UC_X86_REG_ESP, STACK)
        cpu.emu_start(ENTRY, STOP, count=5000)
        assert cpu.reg_read(UC_X86_REG_EIP) == STOP and cpu.reg_read(UC_X86_REG_ESP) == STACK + 4
        assert bytes(cpu.mem_read(ENTRY, len(code))) == code and not faults
        assert bytes(cpu.mem_read(PACKET, 3)) == packet
        assert bytes(cpu.mem_read(PACKET - 16, 16)) == bytes([pattern]) * 16
        assert bytes(cpu.mem_read(PACKET + 3, 16)) == bytes([pattern]) * 16
        assert bytes(cpu.mem_read(0x7140e0 - 32, 32)) == bytes([pattern]) * 32
        assert bytes(cpu.mem_read(0x714180, 32)) == bytes([pattern]) * 32
        observed.append(dict(events=events, platform=platform, sequence=word(SEQUENCE),
                             frame=word(FRAME), text_kind=kind(word(TEXT)),
                             EAX=cpu.reg_read(UC_X86_REG_EAX)))
    assert observed[0] == observed[1] == observed[2]
    rows.append(dict(input=current, expected=observed[0]))
print(json.dumps(dict(original_sha256=hashlib.sha256(pe).hexdigest(), entry=hex(ENTRY),
    code_bytes=len(code), code_sha256=hashlib.sha256(code).hexdigest(), entries=entries,
    cases=len(rows), patterns=[0, 0x5a, 0xff], faults=faults, code_packet_guards_unchanged=True,
    original_presentation_end_EAX_is_declared_service_not_decoded_short_API=True,
    scope="unhooked original packet/dedup/frame/localization/geometry/order with declared UI services; not Windows renderer/driver/framebuffer",
    rows=rows), separators=(",", ":")))
