"""Execute original video initializers; only compass formatter/image services are controlled.

The trig entry, FSIN and original truncating x87 helper execute without hooks.
This is not a Windows framebuffer or a physical image-loader verification.
"""
import hashlib
import json
import struct
import sys
from pathlib import Path

from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import (
    UC_X86_REG_EAX, UC_X86_REG_EDX, UC_X86_REG_EIP, UC_X86_REG_ESP, UC_X86_REG_FPCW,
)

pe = Path(sys.argv[1]).read_bytes()
assert hashlib.sha256(pe).hexdigest() == "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a"
h = struct.unpack_from("<I", pe, 0x3c)[0]
optional = h + 24
base = struct.unpack_from("<I", pe, optional + 28)[0]
image_size = struct.unpack_from("<I", pe, optional + 56)[0]
sections_at = optional + struct.unpack_from("<H", pe, h + 20)[0]
cpu = Uc(UC_ARCH_X86, UC_MODE_32)
cpu.mem_map(base, (image_size + 4095) & ~4095)
for i in range(struct.unpack_from("<H", pe, h + 6)[0]):
    _, virtual, length, raw = struct.unpack_from("<IIII", pe, sections_at + 40*i + 8)
    cpu.mem_write(base + virtual, pe[raw:raw+length])
cpu.mem_map(0x20000000, 0x10000)
COMPASS, TRIG, FTOL = 0x4a9c80, 0x4aee30, 0x566dcc
FORMAT, LOAD = 0x565c27, 0x42f970
STACK, STOP = 0x20008000, 0x20009000
codes = {COMPASS: bytes(cpu.mem_read(COMPASS, 126)),
         TRIG: bytes(cpu.mem_read(TRIG, 67)), FTOL: bytes(cpu.mem_read(FTOL, 39))}
assert all(code[-1] == 0xc3 for code in codes.values())
entries = {COMPASS: 0, TRIG: 0, FTOL: 0}
calls, formats, faults = [], [], []

def get(at):
    return struct.unpack("<I", cpu.mem_read(at, 4))[0]

def cstr(at):
    out = bytearray()
    while True:
        value = bytes(cpu.mem_read(at + len(out), 1))[0]
        if value == 0:
            return out.decode("ascii")
        out.append(value)
        assert len(out) < 64

def hook(_cpu, address, _length, _):
    if address in entries:
        entries[address] += 1
    if any(start <= address < start + len(code) for start, code in codes.items()) or address == STOP:
        return
    assert address in (FORMAT, LOAD), "undeclared execution " + hex(address)
    sp = cpu.reg_read(UC_X86_REG_ESP)
    if address == FORMAT:
        dst, fmt, index = struct.unpack("<III", cpu.mem_read(sp + 4, 12))
        text = cstr(fmt) % index
        cpu.mem_write(dst, text.encode("ascii") + b"\0")
        formats.append([cstr(fmt), index, text])
        returned = len(text)
    else:
        name = cstr(get(sp + 4))
        returned = 0x80000000 | (len(calls) + 1)
        calls.append([name, returned])
    cpu.reg_write(UC_X86_REG_EAX, returned)
    cpu.reg_write(UC_X86_REG_EIP, get(sp))
    cpu.reg_write(UC_X86_REG_ESP, sp + 4)

def invalid(_cpu, access, address, length, value, _):
    faults.append(dict(address=address, length=length, pc=cpu.reg_read(UC_X86_REG_EIP)))
    return False

cpu.hook_add(UC_HOOK_CODE, hook)
cpu.hook_add(UC_HOOK_MEM_INVALID, invalid)
rows = []
for pattern in (0, 0x5a, 0xff):
    calls, formats = [], []
    cpu.mem_write(0x7140e0 - 32, bytes([pattern]) * (152 + 64))
    cpu.mem_write(0x714224 - 32, bytes([pattern]) * (1280 + 64))
    returns = []
    for entry in (COMPASS, TRIG):
        cpu.mem_write(STACK - 256, bytes([pattern]) * 512)
        cpu.mem_write(STACK, struct.pack("<I", STOP))
        cpu.reg_write(UC_X86_REG_ESP, STACK)
        cpu.reg_write(UC_X86_REG_FPCW, 0x37f)
        cpu.emu_start(entry, STOP, count=30000)
        assert cpu.reg_read(UC_X86_REG_EIP) == STOP and cpu.reg_read(UC_X86_REG_ESP) == STACK + 4
        returns.append([cpu.reg_read(UC_X86_REG_EAX), cpu.reg_read(UC_X86_REG_EDX)])
    assert len(calls) == len(formats) == 36 and not faults
    assert get(0x714174) == 0
    for address, length in ((0x7140e0, 152), (0x714224, 1280)):
        assert bytes(cpu.mem_read(address - 32, 32)) == bytes([pattern]) * 32
        assert bytes(cpu.mem_read(address + length, 32)) == bytes([pattern]) * 32
    assert all(bytes(cpu.mem_read(start, len(code))) == code for start, code in codes.items())
    rows.append(dict(pattern=pattern, calls=calls, formats=formats,
                     compass_words=list(struct.unpack("<38I", cpu.mem_read(0x7140e0, 152))),
                     trig_words=list(struct.unpack("<320i", cpu.mem_read(0x714224, 1280))),
                     returned=returns))
assert entries == {COMPASS: 3, TRIG: 3, FTOL: 960}
assert all(row["calls"] == rows[0]["calls"] and row["formats"] == rows[0]["formats"]
           and row["trig_words"] == rows[0]["trig_words"] and row["returned"] == rows[0]["returned"]
           for row in rows)
constants = {hex(at): bytes(cpu.mem_read(at, 8)).hex() for at in (0x583a78, 0x583a70, 0x5839a0)}
print(json.dumps(dict(original_sha256=hashlib.sha256(pe).hexdigest(),
    code=[dict(entry=hex(at), bytes=len(code), sha256=hashlib.sha256(code).hexdigest()) for at, code in codes.items()],
    entries={hex(at): n for at, n in entries.items()}, constants=constants,
    faults=faults, code_and_guards_unchanged=True, rows=rows,
    scope="original compass loop/name/store with declared formatter/image token services; unhooked original trig FSIN/truncating helper; not Windows image loader/framebuffer"),
    separators=(",", ":")))
