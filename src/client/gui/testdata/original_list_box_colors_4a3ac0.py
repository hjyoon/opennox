"""Bounded stock GAME.EXE list-box palette and insertion color oracle.

Execute unchanged 4A94A0 and 4A3AC0 instructions. RGB5551 packing, UTF-16
copy/length, font height and reflow are declared service boundaries. This
checks row color selection, not an original Windows framebuffer.
Usage: python original_list_box_colors_4a3ac0.py REPO GAME.EXE
"""
import hashlib
import json
import pathlib
import struct
import sys
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import UC_X86_REG_EAX, UC_X86_REG_EIP, UC_X86_REG_ESP

repo, original = map(pathlib.Path, sys.argv[1:3])
blob = original.read_bytes()
manifest = json.loads((repo / 'toolchain/oracle/game-exe-functions.json').read_text())
digest = hashlib.sha256(blob).hexdigest()
assert digest == manifest['file_sha256'] == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe = struct.unpack_from('<I', blob, 0x3c)[0]
opt = pe + 24
base = struct.unpack_from('<I', blob, opt + 28)[0]
table = opt + struct.unpack_from('<H', blob, pe + 20)[0]
sections = [struct.unpack_from('<IIII', blob, table + i*40 + 8)
            for i in range(struct.unpack_from('<H', blob, pe + 6)[0])]
size = (struct.unpack_from('<I', blob, opt + 56)[0] + 4095) & ~4095
cpu = Uc(UC_ARCH_X86, UC_MODE_32)
cpu.mem_map(base, size)
for _, virtual, raw_size, raw in sections:
    if raw_size:
        cpu.mem_write(base + virtual, blob[raw:raw + raw_size])
addresses = {0x4a3ac0}
ranges = [r for r in manifest['functions'] if int(r['address'], 16) in addresses]
assert len(ranges) == len(addresses)
sealed = {}
for r in ranges:
    start = int(r['address'], 16)
    code = bytes(cpu.mem_read(start, r['size']))
    assert hashlib.sha256(code).hexdigest() == r['sha256']
    sealed[start] = code
# The whole-file digest seals the initializer, which is not a separate entry
# in the older function manifest. Its first byte through RET is 0x214 bytes.
start, count = 0x4a94a0, 0x214
sealed[start] = bytes(cpu.mem_read(start, count))
ranges.append({'name': 'nox_xxx_loadDefColor_4A94A0', 'address': hex(start),
               'size': count, 'sha256': hashlib.sha256(sealed[start]).hexdigest()})
heap, stack, stop = 0x10000000, 0x20000000, 0x30000000
for address in (heap, stack, stop):
    cpu.mem_map(address, 0x10000)
win, data, items, text = (heap + n for n in (0x100, 0x500, 0x1000, 0x2000))
get = lambda p: struct.unpack('<I', cpu.mem_read(p, 4))[0]
put = lambda p, v: cpu.mem_write(p, struct.pack('<I', v & 0xffffffff))
rgb_calls, faults = [], []

def hook(_cpu, address, _size, _):
    if any(start <= address < start + len(code) for start, code in sealed.items()):
        return
    sp = cpu.reg_read(UC_X86_REG_ESP)
    arg = lambda n: get(sp + 4 + n * 4)
    result = 0
    if address == 0x4344a0:
        rgb = [arg(n) for n in range(3)]
        rgb_calls.append(rgb)
        r, g, b = rgb
        packed = ((r & 248) << 7) | ((g & 248) << 2) | ((b & 248) >> 3)
        result = packed | (packed << 16)
    elif address == 0x565cc8:
        dest, src, count = [arg(n) for n in range(3)]
        for i in range(count):
            value = bytes(cpu.mem_read(src + i*2, 2))
            cpu.mem_write(dest + i*2, value)
            if value == b'\0\0':
                break
        result = dest
    elif address == 0x56688d:
        while bytes(cpu.mem_read(arg(0) + result*2, 2)) != b'\0\0':
            result += 1
    elif address == 0x43f320:
        result = 10
    elif address == 0x4a3a70:
        result = 1
    else:
        raise AssertionError('undeclared service ' + hex(address))
    cpu.reg_write(UC_X86_REG_EAX, result)
    cpu.reg_write(UC_X86_REG_EIP, get(sp))
    cpu.reg_write(UC_X86_REG_ESP, sp + 4)

def invalid(_cpu, access, address, size, value, _):
    faults.append([access, hex(address), size, value])
    return False

cpu.hook_add(UC_HOOK_CODE, hook)
cpu.hook_add(UC_HOOK_MEM_INVALID, invalid)

def call(address, *args):
    sp = stack + 0x8000
    for i, value in enumerate((stop,) + args):
        put(sp + i*4, value)
    cpu.reg_write(UC_X86_REG_ESP, sp)
    cpu.emu_start(address, stop, count=20000)
    assert cpu.reg_read(UC_X86_REG_EIP) == stop
    return cpu.reg_read(UC_X86_REG_EAX)

assert call(0x4a94a0) == 1
assert len(rgb_calls) == 17
palette = [get(get(0x85b480 + i*4)) for i in range(17)]
rows = []
for default in (0x12341234, 0x80000000):
    for index in (-2147483648, -2, -1, *range(18), 2147483647):
        cpu.mem_write(win, bytes(0x200))
        cpu.mem_write(data, bytes(0x100))
        cpu.mem_write(items, bytes(524))
        put(win + 4, 0x4000)
        put(win + 8, 200)
        put(win + 32, data)
        put(win + 104, default)
        put(data + 24, items)
        cpu.mem_write(text, 'Auto Save\0'.encode('utf-16le'))
        call(0x4a3ac0, text, index, win)
        color = get(items + 516)
        assert color == (palette[index] if 0 <= index < 17 else default)
        rows.append({'index': index, 'default': default, 'color': color})
assert not faults
assert all(bytes(cpu.mem_read(start, len(code))) == code for start, code in sealed.items())
assert original.read_bytes() == blob
print(json.dumps({'original_sha256': digest, 'functions': ranges,
                  'rgb_calls': rgb_calls, 'palette': palette, 'rows': rows,
                  'memory_faults': faults, 'executable_and_code_unchanged': True,
                  'scope': __doc__}, indent=2))
