"""Run original PE32 team enumeration and ten-byte member packet bodies.

Team/player enumeration, membership, descriptor send, object lookup and packet
transport are declared controlled services. The actual two original function
bodies and field reads are executed without instruction replacement.
"""
import hashlib
import json
import struct
import sys
from pathlib import Path
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import UC_X86_REG_EAX, UC_X86_REG_EBX, UC_X86_REG_ECX, UC_X86_REG_EDX, UC_X86_REG_EDI, UC_X86_REG_ESI, UC_X86_REG_EBP, UC_X86_REG_EIP, UC_X86_REG_ESP

pe = Path(sys.argv[1]).read_bytes()
assert hashlib.sha256(pe).hexdigest() == "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a"
h = struct.unpack_from("<I", pe, 0x3c)[0]
o = h + 24
base = struct.unpack_from("<I", pe, o + 28)[0]
size = struct.unpack_from("<I", pe, o + 56)[0]
sections = o + struct.unpack_from("<H", pe, h + 20)[0]
cpu = Uc(UC_ARCH_X86, UC_MODE_32)
cpu.mem_map(base, (size + 4095) & ~4095)
for i in range(struct.unpack_from("<H", pe, h + 6)[0]):
    _, va, length, raw = struct.unpack_from("<IIII", pe, sections + i * 40 + 8)
    cpu.mem_write(base + va, pe[raw:raw + length])
cpu.mem_map(0x20000000, 0x20000)
STACK, STOP = 0x2001e000, 0x2001f000
TEAM = [0x20000100 + i * 0x100 for i in range(3)]
UNIT = [0x20001000 + i * 0x1000 for i in range(5)]
MEMBER = [u + 48 for u in UNIT]
ranges = {0x4198a0: 96, 0x4df2e0: 128}
sha = {0x4198a0: "8411a72e7bc786a47cb2aa7d8b2b773616493782b1fb54d2d89d663d9f159f4f",
       0x4df2e0: "1f3c9f560970551b4e9bb14f0b4ca0a6a407463825ef7c251549ff59d0cb21a6"}
membership_abi = bytes(cpu.mem_read(0x419180, 80))
assert hashlib.sha256(membership_abi).hexdigest() == "4913018b168d7e91223131ee8f373fe24287c89b724d8c5785967111d5c18f6e"
assert bytes.fromhex("3ac8") in membership_abi and bytes.fromhex("25ff000000") in membership_abi
sealed = {at: bytes(cpu.mem_read(at, n)) for at, n in ranges.items()}
assert all(hashlib.sha256(sealed[at]).hexdigest() == sha[at] for at in ranges)
team_ids, member_ids = [1, 2, 255], [1, 0, 2, 1, 255]
codes = [0x1234, 0xffff8012, 0x1234abcd, 0x80008000, 0xffffffff]
types = [0x3456, 0x1234, 0xfedc, 0xffff, 0]
events, packets, rows, faults = [], [], [], []
team_count, unit_count, lookup = 3, 5, None
calls = {}

def put(at, value):
    cpu.mem_write(at, struct.pack("<I", value & 0xffffffff))

def get(at):
    return struct.unpack("<I", cpu.mem_read(at, 4))[0]

def args(n):
    sp = cpu.reg_read(UC_X86_REG_ESP)
    return list(struct.unpack("<" + "I" * n, cpu.mem_read(sp + 4, n * 4)))

def ret(value=0):
    sp = cpu.reg_read(UC_X86_REG_ESP)
    cpu.reg_write(UC_X86_REG_EAX, value & 0xffffffff)
    cpu.reg_write(UC_X86_REG_ESP, sp + 4)
    cpu.reg_write(UC_X86_REG_EIP, get(sp))

def hook(_cpu, at, _size, _):
    calls[hex(at)] = calls.get(hex(at), 0) + 1 if at in ranges else calls.get(hex(at), 0)
    if at == STOP or any(start <= at < start + n for start, n in ranges.items()):
        return
    if at == 0x418b10:
        events.append(["team-first"])
        ret(TEAM[0] if team_count else 0)
    elif at == 0x418b60:
        pointer = args(1)[0]
        i = TEAM.index(pointer)
        events.append(["team-next", i])
        ret(TEAM[i + 1] if i + 1 < team_count else 0)
    elif at == 0x4197c0:
        pointer, recipient = args(2)
        events.append(["descriptor", TEAM.index(pointer), recipient])
        ret()
    elif at == 0x4da7c0:
        events.append(["unit-first"])
        ret(UNIT[0] if unit_count else 0)
    elif at == 0x4da7f0:
        pointer = args(1)[0]
        i = UNIT.index(pointer)
        events.append(["unit-next", i])
        ret(UNIT[i + 1] if i + 1 < unit_count else 0)
    elif at == 0x419180:
        pointer, team_id = args(2)
        team_id &= 0xff  # 00419180 compares CL with AL and masks EAX to one byte.
        i = MEMBER.index(pointer)
        events.append(["contains", i, team_id])
        # Exact active ID qualification is a declared controlled service.
        ret(int(member_ids[i] != 0 and member_ids[i] == team_id))
    elif at == 0x4eccb0:
        code = args(1)[0]
        events.append(["lookup", code])
        ret(lookup or 0)
    elif at == 0x4e5390:
        recipient, pointer, length, reliable, channel = args(5)
        data = bytes(cpu.mem_read(pointer, length)).hex()
        packets.append([recipient, data, reliable, channel])
        events.append(["packet", recipient, data, reliable, channel])
        ret(77)
    else:
        raise AssertionError("undeclared execution " + hex(at))

def invalid(_cpu, access, at, length, value, _):
    faults.append([access, at, length, cpu.reg_read(UC_X86_REG_EIP)])
    return False

cpu.hook_add(UC_HOOK_CODE, hook)
cpu.hook_add(UC_HOOK_MEM_INVALID, invalid)
for i, at in enumerate(TEAM):
    cpu.mem_write(at, bytes([0xa5]) * 80)
    cpu.mem_write(at + 57, bytes([team_ids[i]]))
for i, at in enumerate(UNIT):
    cpu.mem_write(at, bytes([0x5a]) * 1024)
    cpu.mem_write(at + 4, struct.pack("<H", types[i]))
    put(at + 36, codes[i])
    cpu.mem_write(at + 52, bytes([member_ids[i]]))
guard = {at: bytes(cpu.mem_read(at, 80)) for at in TEAM}
guard.update({at: bytes(cpu.mem_read(at, 1024)) for at in UNIT})

def call(at, *values):
    cpu.mem_write(STACK - 1024, bytes([0xcc]) * 1024)
    cpu.mem_write(STACK, struct.pack("<" + "I" * (len(values) + 1), STOP, *(v & 0xffffffff for v in values)))
    saved = {UC_X86_REG_EBX: 0x11223344, UC_X86_REG_ESI: 0x55667788, UC_X86_REG_EDI: 0x99aabbcc, UC_X86_REG_EBP: 0xddeeff11}
    for reg, value in saved.items():
        cpu.reg_write(reg, value)
    cpu.reg_write(UC_X86_REG_ESP, STACK)
    cpu.emu_start(at, STOP, count=10000)
    assert cpu.reg_read(UC_X86_REG_EIP) == STOP and cpu.reg_read(UC_X86_REG_ESP) == STACK + 4
    assert all(cpu.reg_read(reg) == value for reg, value in saved.items())
    assert not faults
    assert all(bytes(cpu.mem_read(at, len(value))) == value for at, value in guard.items())

for label, recipient, nt, nu in (("host-index-31", 31, 3, 5), ("empty-teams", 0, 0, 5),
                                ("empty-players", 30, 3, 0), ("all-members", 0, 3, 5),
                                ("last-ordinary-index", 30, 3, 5)):
    events.clear(); packets.clear()
    team_count, unit_count, lookup = nt, nu, UNIT[4]
    call(0x4df2e0, recipient)
    rows.append(dict(name=label, events=list(events), packets=list(packets)))
for label, member, recipient, code, obj in (
        ("nil-member", 0, 17, 0xffffffff, UNIT[2]),
        ("missing-object", MEMBER[0], 17, 0x1234abcd, 0),
        ("zero-team-id", MEMBER[1], 17, 0, UNIT[0]),
        ("high-net-code", MEMBER[2], 30, 0x1234abcd, UNIT[2]),
        ("signed-net-code", MEMBER[4], 0, 0xffffffff, UNIT[3])):
    events.clear(); packets.clear(); lookup = obj
    call(0x4198a0, member, recipient, code)
    rows.append(dict(name=label, events=list(events), packets=list(packets)))
assert all(bytes(cpu.mem_read(at, n)) == sealed[at] for at, n in ranges.items())
print(json.dumps(dict(original_sha256=hashlib.sha256(pe).hexdigest(),
    functions={hex(at): dict(bytes=n, sha256=sha[at]) for at, n in ranges.items()},
    membership_byte_abi=dict(address="0x419180", bytes=80, sha256=hashlib.sha256(membership_abi).hexdigest(), stack_argument_low_byte=True),
    cases=rows, memory_faults=faults, code_and_input_guards_unchanged=True,
    scope="unchanged original 004DF2E0 enumeration and 004198A0 packet body; declared controlled enumeration/membership (original byte ABI)/descriptor/lookup/transport, not full Windows Quest"),
    separators=(",", ":")))
