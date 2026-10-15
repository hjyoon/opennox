"""Execute unchanged GAME.EXE 004D2160, not the decompiled C.

Enumeration, cached count, RNG, membership and pickup are declared controlled
services. Team+76 reads execute in the original body; native tests use the
already established side table instead of changing the fixed Team ABI.
"""
import hashlib
import json
import struct
import sys
from pathlib import Path
from unicorn import Uc, UcError, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_READ, UC_HOOK_MEM_INVALID
from unicorn.x86_const import UC_X86_REG_EAX, UC_X86_REG_EBX, UC_X86_REG_ESI, UC_X86_REG_EDI, UC_X86_REG_EBP, UC_X86_REG_EIP, UC_X86_REG_ESP

PE_SHA = "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a"
BODY_SHA = "4e636081e9b36e3bb9532c6fefec50c6941e30eb2e35fb936afb87c181415e63"
pe = Path(sys.argv[1]).read_bytes()
assert hashlib.sha256(pe).hexdigest() == PE_SHA
h = struct.unpack_from('<I', pe, 0x3c)[0]
o = h + 24
base, size = struct.unpack_from('<I', pe, o + 28)[0], struct.unpack_from('<I', pe, o + 56)[0]
sections = o + struct.unpack_from('<H', pe, h + 20)[0]
cpu = Uc(UC_ARCH_X86, UC_MODE_32)
cpu.mem_map(base, (size + 4095) & ~4095)
for i in range(struct.unpack_from('<H', pe, h + 6)[0]):
    _, va, length, raw = struct.unpack_from('<IIII', pe, sections + i * 40 + 8)
    cpu.mem_write(base + va, pe[raw:raw + length])
body = bytes(cpu.mem_read(0x4d2160, 208))
assert hashlib.sha256(body).hexdigest() == BODY_SHA
member_body = bytes(cpu.mem_read(0x419180, 80))
assert hashlib.sha256(member_body).hexdigest() == '4913018b168d7e91223131ee8f373fe24287c89b724d8c5785967111d5c18f6e'
cpu.mem_map(0x20000000, 0x20000)
TEAM = [0x20000100 + i * 0x100 for i in range(3)]
PLAYER = [0x20001000 + i * 0x1000 for i in range(4)]
UNIT = [0x20006000 + i * 0x1000 for i in range(4)]
CROWN = [0x2000b000 + i * 0x1000 for i in range(4)]
STACK, STOP = 0x2001e000, 0x2001f000
names = ['empty', 'empty-cached-32', 'held', 'no-members', 'first', 'last',
         'retry-nil-and-wrong-team', 'three-teams', 'pickup-zero-not-retried',
         'shared-crown-now-held', 'reload-after-membership', 'cached-count-int-min',
         'held-no-members', 'nil-crown-fault', 'zero-roster-fault', 'short-roster-fault']
events, faults, guards, rows = [], [], {}, []
case, nt, np, cached, rolls, ri = 0, 1, 1, 1, [0], 0
members, ids, flags = [1, 1, 1], [1, 2, 255], CROWN[:3]

def put(at, value):
    cpu.mem_write(at, struct.pack('<I', value & 0xffffffff))

def get(at):
    return struct.unpack('<I', cpu.mem_read(at, 4))[0]

def mutate(at, value):
    put(at, value)
    for start, old in list(guards.items()):
        if start <= at and at + 4 <= start + len(old):
            copy = bytearray(old)
            copy[at-start:at-start+4] = struct.pack('<I', value & 0xffffffff)
            guards[start] = bytes(copy)
            return
    raise AssertionError('mutation outside guarded allocation')

def args(n):
    return struct.unpack('<' + 'I' * n, cpu.mem_read(cpu.reg_read(UC_X86_REG_ESP) + 4, n * 4))

def signed(n):
    return n if n < 0x80000000 else n - 0x100000000

def ret(value=0):
    sp = cpu.reg_read(UC_X86_REG_ESP)
    cpu.reg_write(UC_X86_REG_EAX, value & 0xffffffff)
    cpu.reg_write(UC_X86_REG_ESP, sp + 4)
    cpu.reg_write(UC_X86_REG_EIP, get(sp))

def hook(_cpu, at, _size, _):
    global ri
    if at == STOP or 0x4d2160 <= at < 0x4d2230:
        return
    if at == 0x416f40:
        events.append('count'); ret(cached)
    elif at == 0x418b10:
        events.append('team-first'); ret(TEAM[0] if nt else 0)
    elif at == 0x418b60:
        i = TEAM.index(args(1)[0]); events.append('team-next %d' % i)
        ret(TEAM[i + 1] if i + 1 < nt else 0)
    elif at == 0x418bc0:
        i = TEAM.index(args(1)[0]); events.append('members %d' % i); ret(members[i])
    elif at == 0x415fa0:
        lo, hi, source, line = args(4)
        assert lo == 0 and source == 0x5b70e0 and line == 151
        assert ri < len(rolls)
        events.append('random %d %d' % (signed(lo), signed(hi)))
        value = rolls[ri]; ri += 1; ret(value)
    elif at == 0x416ea0:
        events.append('player-first'); ret(PLAYER[0] if np else 0)
    elif at == 0x416ee0:
        i = PLAYER.index(args(1)[0]); events.append('player-next %d' % i)
        ret(PLAYER[i + 1] if i + 1 < np else 0)
    elif at == 0x419180:
        member, tid = args(2); tid &= 255
        i = UNIT.index(member - 48); events.append('contains %d %d' % (i, tid))
        result = int(get(UNIT[i] + 52) & 255 == tid and tid != 0)
        if case == 10:
            flags[0] = CROWN[3]
            mutate(TEAM[0] + 76, CROWN[3]); mutate(PLAYER[0] + 2056, UNIT[3])
        ret(result)
    elif at == 0x4f3400:
        who, crown, a, b = args(4)
        i, j = UNIT.index(who), CROWN.index(crown)
        events.append('pickup %d %d %d %d' % (i, j, a, b))
        if case == 9:
            mutate(crown + 492, who)
        ret(0 if case == 8 else 0x80000001)
    else:
        raise AssertionError('undeclared original execution ' + hex(at))

def read(_cpu, _access, at, length, _value, _):
    if at in [p + 76 for p in TEAM] and cpu.reg_read(UC_X86_REG_EIP) in (0x4d217d, 0x4d21f2):
        assert length == 4
        events.append('flag-load %d' % TEAM.index(at - 76))

def invalid(_cpu, access, at, length, _value, _):
    faults.append(dict(address=at, length=length, pc=cpu.reg_read(UC_X86_REG_EIP)))
    return False

cpu.hook_add(UC_HOOK_CODE, hook)
cpu.hook_add(UC_HOOK_MEM_READ, read)
cpu.hook_add(UC_HOOK_MEM_INVALID, invalid)
saved = {UC_X86_REG_EBX: 0x11223344, UC_X86_REG_ESI: 0x55667788,
         UC_X86_REG_EDI: 0x99aabbcc, UC_X86_REG_EBP: 0xddeeff11}
for pattern in (0, 0x5a, 0xff):
    for case, name in enumerate(names):
        events.clear(); faults.clear(); guards.clear()
        nt, np, cached, rolls, ri = 1, 1, 1, [0], 0
        members, flags = [1, 1, 1], CROWN[:3]
        player_units, unit_ids = [UNIT[i] for i in range(4)], [1, 2, 255, 255]
        held = case in (2, 12)
        if case in (0, 1): nt, np, cached = 0, 0, 0 if case == 0 else 32
        if case in (3, 12): members[0] = 0
        if case == 5: np, cached, rolls, unit_ids = 3, 3, [2], [2, 255, 1, 255]
        if case == 6:
            np, cached, rolls, player_units[0], unit_ids[2] = 4, 4, [0, 1, 2], 0, 1
        if case == 7: nt, np, cached, rolls = 3, 3, 3, [0, 1, 2]
        if case == 9: nt, np, cached, flags[1] = 2, 2, 2, CROWN[0]
        if case == 11: cached = -2147483648
        if case == 13: flags[0] = 0
        if case == 14: np, cached = 0, 0
        if case == 15: cached, rolls = 3, [2]
        for i, at in enumerate(TEAM):
            cpu.mem_write(at, bytes([pattern]) * 80)
            cpu.mem_write(at + 57, bytes([ids[i]])); put(at + 76, flags[i])
            guards[at] = bytes(cpu.mem_read(at, 80))
        for i, at in enumerate(PLAYER):
            cpu.mem_write(at, bytes([pattern]) * 3072); put(at + 2056, player_units[i])
            guards[at] = bytes(cpu.mem_read(at, 3072))
        for i, at in enumerate(UNIT):
            cpu.mem_write(at, bytes([pattern]) * 1024); put(at + 52, unit_ids[i])
            guards[at] = bytes(cpu.mem_read(at, 1024))
        for i, at in enumerate(CROWN):
            cpu.mem_write(at, bytes([pattern]) * 1024); put(at + 492, UNIT[0] if i == 0 and held else 0)
            guards[at] = bytes(cpu.mem_read(at, 1024))
        cpu.mem_write(STACK - 1024, bytes([0xcc]) * 1024); put(STACK, STOP)
        for reg, value in saved.items(): cpu.reg_write(reg, value)
        cpu.reg_write(UC_X86_REG_ESP, STACK)
        try:
            cpu.emu_start(0x4d2160, STOP, count=2000)
        except UcError:
            assert case >= 13 and len(faults) == 1
        assert bool(faults) == (case >= 13)
        if not faults:
            assert cpu.reg_read(UC_X86_REG_EIP) == STOP and cpu.reg_read(UC_X86_REG_ESP) == STACK + 4
            assert all(cpu.reg_read(reg) == value for reg, value in saved.items())
        assert all(bytes(cpu.mem_read(at, len(value))) == value for at, value in guards.items())
        assert bytes(cpu.mem_read(0x4d2160, 208)) == body
        rows.append(dict(index=case, name=name, pattern=pattern, events=list(events), fault=bool(faults)))
assert len(rows) == 48
print(json.dumps(dict(original_sha256=PE_SHA, body_sha256=BODY_SHA, body_bytes=208,
    code_and_input_guards_unchanged=True, successful_callee_saved_registers_unchanged=True,
    rng_original_source="0x5b70e0", rng_original_line=151, cases=rows,
    scope="unchanged original whole function; controlled enumeration/count/RNG/membership/pickup; not Windows or multiplayer end-to-end"), separators=(',', ':')))
