"""Execute unchanged GAME.EXE 004183C0 with declared enumeration/team services."""
import hashlib
import json
import struct
import sys
from pathlib import Path
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import (UC_X86_REG_EAX, UC_X86_REG_EBX, UC_X86_REG_ECX,
    UC_X86_REG_EDX, UC_X86_REG_ESI, UC_X86_REG_EDI, UC_X86_REG_EBP,
    UC_X86_REG_ESP, UC_X86_REG_EIP, UC_X86_REG_FPCW, UC_X86_REG_FPSW, UC_X86_REG_FPTAG)

pe = Path(sys.argv[1]).read_bytes()
assert hashlib.sha256(pe).hexdigest() == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
h = struct.unpack_from('<I', pe, 60)[0]
o = h + 24
base = struct.unpack_from('<I', pe, o + 28)[0]
size = struct.unpack_from('<I', pe, o + 56)[0]
sections = o + struct.unpack_from('<H', pe, h + 20)[0]
cpu = Uc(UC_ARCH_X86, UC_MODE_32)
cpu.mem_map(base, (size + 4095) & ~4095)
for i in range(struct.unpack_from('<H', pe, h + 6)[0]):
    _, va, length, raw = struct.unpack_from('<IIII', pe, sections + i * 40 + 8)
    cpu.mem_write(base + va, pe[raw:raw + length])
START, LENGTH = 0x4183c0, 272
sealed = bytes(cpu.mem_read(START, LENGTH))
assert hashlib.sha256(sealed).hexdigest() == '3201fec59d0de584d3bbf363ef55e9a03d1549629400e26666f50468eea3ed47'
cpu.mem_map(0x20000000, 0x40000)
UNIT = [0x20001000 + i * 0x1000 for i in range(3)]
UPDATE = [0x20008000 + i * 0x1000 for i in range(3)]
PLAYER = [0x20010000 + i * 0x2000 for i in range(3)]
TEAM = [0x20020000 + i * 0x100 for i in range(3)]
FLAG = [0x20021000 + i * 0x1000 for i in range(3)]
STACK, STOP = 0x2003e000, 0x2003f000
NET_CODE, ENGINE = 0x85319c, 0x85b7a0
events, faults, rows = [], [], []
case = None

def put(at, value):
    cpu.mem_write(at, struct.pack('<I', value & 0xffffffff))

def get(at):
    return struct.unpack('<I', cpu.mem_read(at, 4))[0]

def args(n):
    return list(struct.unpack('<' + 'I' * n, cpu.mem_read(cpu.reg_read(UC_X86_REG_ESP) + 4, 4 * n)))

def ret(value=0):
    sp = cpu.reg_read(UC_X86_REG_ESP)
    cpu.reg_write(UC_X86_REG_EAX, value & 0xffffffff)
    cpu.reg_write(UC_X86_REG_ESP, sp + 4)
    cpu.reg_write(UC_X86_REG_EIP, get(sp))

def member_index(pointer):
    return [p + 48 for p in UNIT].index(pointer)

def hook(_cpu, at, _size, _):
    if at == STOP or START <= at < START + LENGTH:
        return
    if at == 0x4da7c0:
        events.append(['unit-first'])
        ret(UNIT[0] if case['units'] else 0)
    elif at == 0x4da7f0:
        i = UNIT.index(args(1)[0]); events.append(['unit-next', i])
        ret(UNIT[i + 1] if i + 1 < len(case['units']) else 0)
    elif at == 0x418ab0:
        team_id = args(1)[0] & 255
        events.append(['by-id', team_id])
        ret(next((TEAM[i] for i, t in enumerate(case['teams']) if t['id'] == team_id), 0))
    elif at == 0x418b10:
        events.append(['team-first'])
        ret(TEAM[0] if case['teams'] else 0)
    elif at == 0x418b60:
        i = TEAM.index(args(1)[0]); events.append(['team-next', i])
        ret(TEAM[i + 1] if i + 1 < len(case['teams']) else 0)
    elif at == 0x4196d0:
        member, team, code, flags = args(4)
        events.append(['change', member_index(member), TEAM.index(team), code, flags])
        ret(1)
    elif at == 0x4191d0:
        team_id, member, active, code, flags = args(5)
        # The original callee accepts unsigned char: only low stack byte is live.
        events.append(['attach', team_id & 255, member_index(member), active, code, flags])
        ret()
    else:
        raise AssertionError('undeclared original execution ' + hex(at))

def invalid(_cpu, access, at, length, value, _):
    faults.append([access, at, length, cpu.reg_read(UC_X86_REG_EIP)])
    return False

cpu.hook_add(UC_HOOK_CODE, hook)
cpu.hook_add(UC_HOOK_MEM_INVALID, invalid)

def bits(value):
    return struct.unpack('<I', struct.pack('<f', value))[0]

def unit(x=0.0, y=0.0, team=0, status=0, code=0x1234abcd, update=1):
    return dict(x=bits(x), y=bits(y), team=team, status=status, code=code, update=update)

def team(team_id, x=0.0, y=0.0, flag=1):
    return dict(id=team_id, x=bits(x), y=bits(y), flag=flag)

cases = []
def add(name, teams, units, client=0xffffffff, engine=0):
    cases.append(dict(name=name, teams=teams, units=units, client=client, engine=engine))

add('empty-units', [team(1)], [])
add('empty-teams', [], [unit(), unit(team=255)])
add('missing-all-flags', [team(1, flag=0), team(255, flag=0)], [unit()])
add('attach-nearest-last', [team(1, 40, 50), team(2, 20, 10), team(255, 1, 1)], [unit()])
add('retain-same-team', [team(1, 1, 1), team(255, 10, 10)], [unit(team=1)])
add('change-team', [team(1, 40, 50), team(255, 1, 1)], [unit(team=1)])
add('ties-keep-first', [team(255, -1, 1), team(2, 1, -1)], [unit()])
add('mixed-three-units', [team(1, -2, 3), team(2, 10, 20), team(255, 100, 80)],
    [unit(-3, 4, team=2), unit(10, 19, team=2, code=0xffffffff), unit(95, 77, code=0x80008000)])
add('strict-threshold-equal', [team(1, 10000, 30000)], [unit()])
add('strict-threshold-below', [team(1, 9999, 30000)], [unit()])
add('strict-threshold-above', [team(1, 10001, 30000)], [unit()])
for status in (0, 1, 0x20, 0x21, 0xffffffff, 0x80000000):
    for local in (False, True):
        for disabled in (0, 0x40000):
            code = 0x80008000
            add('status-%08x-local-%d-disabled-%d' % (status, local, disabled != 0),
                [team(1, 1, 1)], [unit(status=status, code=code)], code if local else 0xffffffff, disabled)
add('null-update-still-eligible', [team(255, 1, 1)], [unit(update=0)], 0x1234abcd, 0x40000)
for label, word in (('positive-zero', 0), ('negative-zero', 0x80000000),
                    ('positive-infinity', 0x7f800000), ('negative-infinity', 0xff800000),
                    ('quiet-nan', 0x7fc12345), ('signaling-nan', 0x7f812345),
                    ('max-finite', 0x7f7fffff), ('min-subnormal', 1)):
    t = team(1); t['x'] = word
    add(label + '-first', [t, team(255, 2, 3)], [unit()])
    t = team(255); t['y'] = word
    add(label + '-last', [team(1, 2, 3), t], [unit()])
# Spill rounding can choose a later equal-distance team when the first exact
# distance lies below its nearest-even float32 spill. Preserve chop instead.
for x, y in ((0.1, 0.2), (1234.567, 0.00001), (31622.775, 0), (1e-20, 1e-20)):
    add('spill-tie-' + str(x), [team(1, x, y), team(255, x, y)], [unit()])

for guard in (0, 0x5a, 0xff):
    for case in cases:
        events.clear(); faults.clear()
        guards = {}
        for i, at in enumerate(TEAM):
            cpu.mem_write(at, bytes([guard]) * 80)
            t = case['teams'][i] if i < len(case['teams']) else team(0, flag=0)
            cpu.mem_write(at + 57, bytes([t['id']]))
            put(at + 72, FLAG[i] if t['flag'] else 0)
            cpu.mem_write(FLAG[i], bytes([guard]) * 1024)
            put(FLAG[i] + 56, t['x']); put(FLAG[i] + 60, t['y'])
            guards[at] = bytes(cpu.mem_read(at, 80)); guards[FLAG[i]] = bytes(cpu.mem_read(FLAG[i], 1024))
        for i, at in enumerate(UNIT):
            u = case['units'][i] if i < len(case['units']) else unit()
            cpu.mem_write(at, bytes([guard]) * 1024)
            put(at + 36, u['code']); cpu.mem_write(at + 52, bytes([u['team']]))
            put(at + 56, u['x']); put(at + 60, u['y']); put(at + 748, UPDATE[i] if u['update'] else 0)
            cpu.mem_write(UPDATE[i], bytes([guard]) * 512); put(UPDATE[i] + 276, PLAYER[i])
            cpu.mem_write(PLAYER[i], bytes([guard]) * 5000)
            put(PLAYER[i] + 2060, u['code']); put(PLAYER[i] + 3680, u['status'])
            for ptr, length in ((at, 1024), (UPDATE[i], 512), (PLAYER[i], 5000)):
                guards[ptr] = bytes(cpu.mem_read(ptr, length))
        put(NET_CODE, case['client']); put(ENGINE, case['engine'])
        cpu.mem_write(STACK - 1024, bytes([guard]) * 2048); put(STACK, STOP)
        saved = {UC_X86_REG_EBX: 0x11223344, UC_X86_REG_ESI: 0x55667788,
                 UC_X86_REG_EDI: 0x99aabbcc, UC_X86_REG_EBP: 0xddeeff11}
        for reg, value in saved.items(): cpu.reg_write(reg, value)
        cpu.reg_write(UC_X86_REG_FPCW, 0x0e7f)
        cpu.reg_write(UC_X86_REG_FPSW, 0); cpu.reg_write(UC_X86_REG_FPTAG, 0xffff)
        cpu.reg_write(UC_X86_REG_ESP, STACK)
        cpu.emu_start(START, STOP, count=100000)
        assert cpu.reg_read(UC_X86_REG_EIP) == STOP and cpu.reg_read(UC_X86_REG_ESP) == STACK + 4
        assert cpu.reg_read(UC_X86_REG_EAX) == 0 and not faults
        assert all(cpu.reg_read(reg) == value for reg, value in saved.items())
        assert all(bytes(cpu.mem_read(ptr, len(value))) == value for ptr, value in guards.items())
        assert bytes(cpu.mem_read(START, LENGTH)) == sealed
        if guard == 0:
            rows.append(dict(**case, events=list(events)))
        else:
            assert rows[cases.index(case)]['events'] == events
print(json.dumps(dict(original_sha256=hashlib.sha256(pe).hexdigest(),
    function=dict(address='0x4183c0', bytes=LENGTH, sha256=hashlib.sha256(sealed).hexdigest()),
    control_word='0x0e7f', cases=rows, guard_patterns=[0, 90, 255], executions=len(cases)*3,
    faults=faults, code_and_input_guards_unchanged=True,
    scope='unchanged original nearest-team body and x87 arithmetic; controlled First/Next/ByID/attach/change services; no Windows framebuffer or full network game'), separators=(',', ':')))
