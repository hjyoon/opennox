"""Execute unchanged GAME.EXE 004E02C0, with charge reporting controlled.

Only the frame/input records and packet-report service are controlled. The
original modulus, class gate, byte increment/clamp, holder/player traversal,
call arguments and fault ordering execute without instruction replacement.
"""
import hashlib
import json
import struct
import sys
from pathlib import Path
from unicorn import Uc, UcError, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import UC_X86_REG_EAX, UC_X86_REG_EBX, UC_X86_REG_EDI, UC_X86_REG_ESI, UC_X86_REG_EBP, UC_X86_REG_EIP, UC_X86_REG_ESP

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
EFFECT, ITEM, USE, OWNER, UPDATE, PLAYER = (0x20000100, 0x20001000, 0x20002000, 0x20003000, 0x20004000, 0x20005000)
STACK, STOP, FRAME, START, LENGTH = (0x2001e000, 0x2001f000, 0x84ea04, 0x4e02c0, 137)
code = bytes(cpu.mem_read(START, LENGTH))
assert code[-1] == 0xc3
reports, faults = [], []

def put(at, value):
    cpu.mem_write(at, struct.pack("<I", value & 0xffffffff))

def get(at):
    return struct.unpack("<I", cpu.mem_read(at, 4))[0]

def hook(_cpu, at, _length, _):
    if at == STOP or START <= at < START + LENGTH:
        return
    assert at == 0x4d82b0, "undeclared original service " + hex(at)
    sp = cpu.reg_read(UC_X86_REG_ESP)
    index, item, charge, maximum = struct.unpack("<IIII", cpu.mem_read(sp + 4, 16))
    assert item == ITEM
    reports.append([index, charge, maximum])
    cpu.reg_write(UC_X86_REG_EAX, 77)
    cpu.reg_write(UC_X86_REG_ESP, sp + 4)
    cpu.reg_write(UC_X86_REG_EIP, get(sp))

def invalid(_cpu, access, at, length, value, _):
    faults.append([access, at, length, cpu.reg_read(UC_X86_REG_EIP)])
    return False

cpu.hook_add(UC_HOOK_CODE, hook)
cpu.hook_add(UC_HOOK_MEM_INVALID, invalid)
fresh_context = cpu.context_save()
cases = []
rates = [240, 210, 180, 150, 1, 0x7fffffff, 0x80000000, 0xffffffff]
pairs = [(0, 0), (0, 1), (4, 5), (5, 5), (6, 5), (254, 255), (255, 255), (255, 1), (1, 0), (255, 0)]
for rate in rates:
    for frame in sorted(set([0, 1, rate - 1, rate, (rate + 1) & 0xffffffff, (2 * rate) & 0xffffffff, 0xffffffff])):
        for charge, maximum in pairs:
            for owner in [0, 2, 4]:
                cases.append(dict(name="matrix", frame=frame, rate=rate, charge=charge, maximum=maximum,
                                  cls=0x1000, owner=owner, index=255 if charge == 255 else 31,
                                  nil_item=False, nil_effect=False, nil_use=False, nil_update=False, nil_player=False))

def case(name, **values):
    row=dict(name=name, frame=240, rate=240, charge=4, maximum=5, cls=0x1000, owner=4,
             index=0, nil_item=False, nil_effect=False, nil_use=False, nil_update=False, nil_player=False)
    row.update(values)
    cases.append(row)

case("nil-item-and-effect", nil_item=True, nil_effect=True, rate=0)
case("nil-item-zero-period", nil_item=True, rate=0)
case("non-wand", cls=0x01000000)
case("non-wand-nil-use", cls=0x01000000, nil_use=True)
case("non-wand-zero-period-still-faults", cls=0x01000000, rate=0)
case("nil-effect", nil_effect=True)
case("nil-use", nil_use=True)
case("nil-update-after-charge", nil_update=True)
case("nil-player-after-charge", nil_player=True)
case("full-charge-no-update-read", charge=5, maximum=5, nil_update=True)
case("off-tick-no-effect-use-read", frame=241, nil_use=True)
case("player-plus-other-class", owner=6, index=128)
case("negative-period-bits", frame=0xffffffff, rate=0xffffffff)
case("nil-holder-after-charge", owner=0)
rows=[]
executions=0
for row in cases:
    expected=None
    for pattern in [0, 0x5a, 0xff]:
        cpu.context_restore(fresh_context)
        reports.clear()
        faults.clear()
        for at, length in [(EFFECT, 144), (ITEM, 1024), (USE, 116), (OWNER, 1024), (UPDATE, 512), (PLAYER, 5000)]:
            cpu.mem_write(at, bytes([pattern]) * length)
        put(EFFECT + 108, row["rate"])
        put(ITEM + 8, row["cls"])
        put(ITEM + 736, 0 if row["nil_use"] else USE)
        put(ITEM + 492, OWNER if row["owner"] else 0)
        put(OWNER + 8, row["owner"])
        put(OWNER + 748, 0 if row["nil_update"] else UPDATE)
        put(UPDATE + 276, 0 if row["nil_player"] else PLAYER)
        cpu.mem_write(PLAYER + 2064, bytes([row["index"]]))
        cpu.mem_write(USE + 108, bytes([row["charge"], row["maximum"]]))
        put(FRAME, row["frame"])
        guard={at:bytes(cpu.mem_read(at, length)) for at, length in [(EFFECT, 144), (ITEM, 1024), (OWNER, 1024), (UPDATE, 512), (PLAYER, 5000)]}
        use_before=bytes(cpu.mem_read(USE, 116))
        cpu.mem_write(STACK - 1024, bytes([0xcc]) * 1024)
        cpu.mem_write(STACK, struct.pack("<IIII", STOP, 0 if row["nil_effect"] else EFFECT, 0 if row["nil_item"] else ITEM, 0))
        saved={UC_X86_REG_EBX:0x11223344, UC_X86_REG_ESI:0x55667788, UC_X86_REG_EDI:0x99aabbcc, UC_X86_REG_EBP:0xddeeff11}
        for reg, value in saved.items():
            cpu.reg_write(reg, value)
        cpu.reg_write(UC_X86_REG_ESP, STACK)
        error=None
        try:
            cpu.emu_start(START, STOP, count=1000)
        except UcError:
            error="memory" if faults else "divide"
        if error is None:
            assert cpu.reg_read(UC_X86_REG_EIP)==STOP, (row, pattern, hex(cpu.reg_read(UC_X86_REG_EIP)), hex(cpu.reg_read(UC_X86_REG_ESP)))
            assert cpu.reg_read(UC_X86_REG_ESP)==STACK+4
            assert all(cpu.reg_read(reg)==value for reg,value in saved.items())
        result=dict(charge=bytes(cpu.mem_read(USE+108,1))[0], reports=list(reports), fault=error,
                    fault_pc=cpu.reg_read(UC_X86_REG_EIP) if error else None)
        assert expected is None or expected==result, ("guard dependence",row,expected,result)
        expected=result
        assert bytes(cpu.mem_read(START,LENGTH))==code
        assert all(bytes(cpu.mem_read(at,len(value)))==value for at,value in guard.items())
        use_after=bytes(cpu.mem_read(USE,116))
        assert use_after[:108]==use_before[:108] and use_after[109:]==use_before[109:]
        executions+=1
    rows.append(dict(input=row,expected=expected))
print(json.dumps(dict(original_sha256=hashlib.sha256(pe).hexdigest(),
                      function=dict(address=hex(START),bytes=LENGTH,sha256=hashlib.sha256(code).hexdigest()),
                      cases=rows,executions=executions,
                      code_input_guards_and_callee_saved_registers_verified=True,
                      controlled_services=["frame/input records","004D82B0 charge report"],
                      whole_windows_execution=False)))
