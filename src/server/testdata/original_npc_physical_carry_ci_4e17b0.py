"""Execute original NPC physical arithmetic for the historical CI inputs.

Usage: python original_npc_physical_carry_ci_4e17b0.py GAME.EXE output.json
Only nonnumeric buff/mode/type/facing/armor/HP services are observed stubs.
The numeric root, carry helper, conversion and startup execute original bytes.
"""
import hashlib
import json
from pathlib import Path
import struct
import sys

from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE
from unicorn.x86_const import UC_X86_REG_EAX, UC_X86_REG_EIP, UC_X86_REG_ESP, UC_X86_REG_FPCW

blob = Path(sys.argv[1]).read_bytes()
sha = lambda b: hashlib.sha256(b).hexdigest()
assert sha(blob) == "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a"
pe = struct.unpack_from("<I", blob, 60)[0]
opt = pe + 24
base = struct.unpack_from("<I", blob, opt + 28)[0]
size = (struct.unpack_from("<I", blob, opt + 56)[0] + 4095) & ~4095
tab = opt + struct.unpack_from("<H", blob, pe + 20)[0]
u = Uc(UC_ARCH_X86, UC_MODE_32)
u.mem_map(base, size)
for i in range(struct.unpack_from("<H", blob, pe + 6)[0]):
    _, address, n, offset = struct.unpack_from("<IIII", blob, tab + 40 * i + 8)
    if n:
        u.mem_write(base + address, blob[offset:offset + n])
regions = [
    ("PlayerDamage", 0x4e17b0, 2368),
    ("carry", 0x4e20f0, 131),
    ("signed-DWORD", 0x419a70, 17),
    ("NPC-action", 0x50a020, 27),
]
region_seals = [dict(name=n, address=hex(a), size=z, sha256=sha(bytes(u.mem_read(a, z)))) for n, a, z in regions]
stack, heap, stop = 0x20000000, 0x21000000, 0x30000000
for address in [stack, heap, stop]:
    u.mem_map(address, 0x10000)
sp, target, update, weapon = stack + 0xe000, heap + 128, heap + 2048, heap + 8192
get = lambda address: struct.unpack("<I", u.mem_read(address, 4))[0]
bits = lambda value: struct.unpack("<I", struct.pack("<f", value))[0]
trace, excluded, events = {}, {}, []
recording = False


def hook(_u, address, n, _):
    if address in [stop, 0x4014c4]:
        u.emu_stop()
        return
    if not recording:
        return
    if any(a <= address < a + z for _, a, z in regions):
        trace[address] = bytes(u.mem_read(address, n)).hex()
        return
    excluded[address] = bytes(u.mem_read(address, n)).hex()
    esp = u.reg_read(UC_X86_REG_ESP)
    args = [get(esp + i * 4) for i in range(1, 6)]
    result = 0
    if address == 0x4ff350:
        events.append(dict(service="buff", enchant=args[1], result=0))
    elif address == 0x40a5c0:
        events.append(dict(service="mode", flag=args[0], result=int(args[0] == 0x40000)))
        result = int(args[0] == 0x40000)
    elif address == 0x4e3aa0:
        events.append(dict(service="excluded-type", result=2000))
        result = 2000
    elif address == 0x4e6e50:
        events.append(dict(service="facing", result=1))
        result = 1
    elif address in [0x4e2180, 0x4e0b30]:
        assert args[:3] == [target, weapon, weapon] and args[4] == 3
        events.append(dict(service="armor" if address == 0x4e2180 else "HP", amount_bits32=args[3], carry_bits32=get(update + 4)))
        result = int(address == 0x4e0b30)
    else:
        raise AssertionError(("unexpected original service", hex(address)))
    u.reg_write(UC_X86_REG_EAX, result)
    u.reg_write(UC_X86_REG_EIP, get(esp))
    u.reg_write(UC_X86_REG_ESP, esp + 4)


u.hook_add(UC_HOOK_CODE, hook)
u.reg_write(UC_X86_REG_ESP, sp)
u.reg_write(UC_X86_REG_FPCW, 0x27f)
u.emu_start(0x4014b5, 0x4014c4, count=20000)
assert u.reg_read(UC_X86_REG_FPCW) == 0xe3f
initial = u.context_save()
recording = True
inputs = [("fractional-hit-0", 3, .25, 0), ("fractional-hit-1", 3, .25, .25), ("fractional-hit-2", 3, .25, .5), ("ordered-armor", 3, .5, .25)]
inputs += [("binary32-" + str(armor), raw, armor, 0) for armor, raw in [(.1, 3), (.2, 9), (.4, 3)]]
inputs += [("stone-armor-" + str(armor) + "-raw-" + str(raw), raw, armor, .125) for armor in [0, .25, 1] for raw in [1, 3, 19]]
rows = []
for mode, cw in [("actual-startup", 0xe3f), ("nearest-control", 0x27f)]:
    for name, raw, armor, carry in inputs:
        u.context_restore(initial)
        u.reg_write(UC_X86_REG_FPCW, cw)
        frame = bytearray(16384)
        for offset, value in [(128 + 8, 2), (128 + 12, 0x11012), (128 + 748, update), (2048 + 2072, bits(armor)), (2048 + 4, bits(carry)), (2048 + 2188, 99), (2048 + 2184, 77), (8192 + 8, 0x410000)]:
            struct.pack_into("<I", frame, offset, value)
        struct.pack_into("<H", frame, 8192 + 4, 199)
        struct.pack_into("<2f", frame, 8192 + 72, 17, 19)
        u.mem_write(heap, bytes(frame))
        u.mem_write(0x75204c, bytes(24))
        u.mem_write(sp, struct.pack("<6I", stop, target, weapon, weapon, raw, 3))
        u.reg_write(UC_X86_REG_ESP, sp)
        events = []
        u.emu_start(0x4e17b0, stop, count=10000)
        assert u.reg_read(UC_X86_REG_EIP) == stop and u.reg_read(UC_X86_REG_ESP) == sp + 4
        assert u.reg_read(UC_X86_REG_FPCW) == cw and u.reg_read(UC_X86_REG_EAX) == 1
        armor_event, hp_event = [e for e in events if e["service"] in ["armor", "HP"]]
        assert armor_event["service"] == "armor" and hp_event["service"] == "HP"
        after = bytes(u.mem_read(heap, len(frame)))
        allowed = {i for offset in [2048 + 4, 2048 + 2188, 2048 + 2184] for i in range(offset, offset + 4)}
        assert all(a == b or i in allowed for i, (a, b) in enumerate(zip(frame, after)))
        rows.append(dict(name=name, mode=mode, raw=raw, armor_bits32=bits(armor), initial_carry_bits32=bits(carry), damage_bits32=hp_event["amount_bits32"], wear_bits32=armor_event["amount_bits32"], carry_bits32=get(update + 4), nonoutput_bytes_unchanged=True))
for region in region_seals:
    assert sha(bytes(u.mem_read(int(region["address"], 16), region["size"]))) == region["sha256"]
assert Path(sys.argv[1]).read_bytes() == blob
report = dict(original_sha256=sha(blob), original_startup_FPCW=0xe3f, regions=region_seals, rows=rows, executed_original_instructions=[dict(address=hex(a), bytes=b) for a, b in sorted(trace.items())], excluded_observed_services=[dict(address=hex(a), bytes=b) for a, b in sorted(excluded.items())], FP_arithmetic_store_conversion_substitutions=0, scope="Exact historical CI numeric inputs execute the original NPC IMPALE environment root. This proves shared physical arithmetic only; missile admission, HP, audio and armor service implementations are not substituted into this proof.")
with Path(sys.argv[2]).open("x") as output:
    json.dump(report, output, indent=2)
    output.write("\n")
