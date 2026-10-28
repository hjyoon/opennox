"""Execute the five unchanged monster DIE_FUNCTION callers, stubbing the drop helper.

The complete 0054A390 helper is outside this fixture's scope. In particular,
these records do not prove item allocation, cooperative gates or world drops.
"""
import argparse
import hashlib
import json
import struct
from pathlib import Path

from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE
from unicorn.x86_const import UC_X86_REG_EAX, UC_X86_REG_EIP, UC_X86_REG_ESP

p = argparse.ArgumentParser(description=__doc__)
for name in ("game", "fixture", "proof"):
    p.add_argument("--" + name, type=Path, required=True)
args = p.parse_args()
blob = args.game.read_bytes()
sha = lambda value: hashlib.sha256(value).hexdigest()
assert sha(blob) == "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a"
pe = struct.unpack_from("<I", blob, 60)[0]
opt = pe + 24
base = struct.unpack_from("<I", blob, opt + 28)[0]
size = (struct.unpack_from("<I", blob, opt + 56)[0] + 4095) & ~4095
table = opt + struct.unpack_from("<H", blob, pe + 20)[0]
u = Uc(UC_ARCH_X86, UC_MODE_32)
u.mem_map(base, size)
for n in range(struct.unpack_from("<H", blob, pe + 6)[0]):
    _, va, count, off = struct.unpack_from("<IIII", blob, table + 40*n + 8)
    if count:
        u.mem_write(base + va, blob[off:off + count])
stack, heap, stop = 0x20000000, 0x21000000, 0x30000000
for address in (stack, heap, stop):
    u.mem_map(address, 0x10000)
sp, unit = stack + 0xe000, heap + 32
specs = [("swordsman", 0x54a7d0, 114), ("shaman", 0x54a850, 64),
         ("archer", 0x54a890, 105), ("ogre", 0x54a900, 67),
         ("warlord", 0x54a950, 61)]
regions = [dict(name=n, address=hex(a), size=z,
                sha256=sha(bytes(u.mem_read(a, z)))) for n, a, z in specs]
trace, entries, events, current = {}, {}, [], {}

def get(address):
    return struct.unpack("<I", u.mem_read(address, 4))[0]

def put(address, value):
    u.mem_write(address, struct.pack("<I", value & 0xffffffff))

def string(address):
    if not address:
        return ""
    result = bytearray()
    while (v := bytes(u.mem_read(address + len(result), 1))) != b"\0":
        result += v
    return result.decode("ascii")

def ret(value):
    esp = u.reg_read(UC_X86_REG_ESP)
    u.reg_write(UC_X86_REG_EAX, value & 0xffffffff)
    u.reg_write(UC_X86_REG_EIP, get(esp))
    u.reg_write(UC_X86_REG_ESP, esp + 4)

def hook(_u, address, length, _):
    if address == stop:
        u.emu_stop()
        return
    for name, start, size in specs:
        if start <= address < start + size:
            trace[address] = bytes(u.mem_read(address, length)).hex()
            if address == start:
                entries[name] = entries.get(name, 0) + 1
            return
    esp = u.reg_read(UC_X86_REG_ESP)
    arg = lambda n: get(esp + 4*n)
    if address == 0x415fa0:
        assert [arg(1), arg(2)] == [0, 100]
        assert string(arg(3)).endswith("aifunc.c")
        assert arg(4) == [1139, 1154, 1167, 1183, 1196][current["kind"]-1]
        events.append(dict(service="random", flags=get(unit + 16), minimum=0, maximum=100))
        if current["random_mutation"]:
            put(unit + 16, 0x89abcdef)
        ret(current["roll"])
    elif address == 0x54a390:
        assert arg(1) == (unit if current["unit"] else 0)
        events.append(dict(service="drop", flags=get(unit + 16),
                           unit=current["unit"], type=string(arg(2)),
                           modifiers=[string(arg(n)) for n in range(3, 7)], charge=arg(7)))
        if current["drop_mutation"]:
            put(unit + 16, 0x76543210)
        ret(0xbadcadde)
    else:
        raise AssertionError(hex(address))

u.hook_add(UC_HOOK_CODE, hook)
cases = []
def add(kind, name, **kw):
    row = dict(name=specs[kind-1][0] + "/" + name, kind=kind,
               roll=100, unit=True, random_mutation=False, drop_mutation=False)
    row.update(kw)
    cases.append(row)

for kind in range(1, 6):
    for roll in [-2147483648, -1] + list(range(101)) + [101, 2147483647]:
        add(kind, "roll/" + str(roll), roll=roll)
    add(kind, "nil/no-drop", unit=False, roll=0)
    add(kind, "nil/selected", unit=False)
    add(kind, "random-write", random_mutation=True)
    add(kind, "drop-write", drop_mutation=True)
    add(kind, "ordered-writes", random_mutation=True, drop_mutation=True)

rows = []
before = bytes((n*37 + 11) & 255 for n in range(256))
for current in cases:
    events.clear()
    u.mem_write(heap, before)
    put(unit + 16, 0x12345678)
    expected_frame = bytearray(u.mem_read(heap, len(before)))
    put(sp, stop)
    put(sp + 4, unit if current["unit"] else 0)
    u.reg_write(UC_X86_REG_ESP, sp)
    u.reg_write(UC_X86_REG_EAX, 0xdeadbeef)
    u.emu_start(specs[current["kind"]-1][1], stop + 1, count=1000)
    assert u.reg_read(UC_X86_REG_EIP) == stop
    result = u.reg_read(UC_X86_REG_EAX)
    assert result == 1
    flags = get(unit + 16)
    struct.pack_into("<I", expected_frame, 48, flags)
    assert bytes(u.mem_read(heap, len(before))) == bytes(expected_frame)
    rows.append(dict(input=current, result=result, flags=flags, events=list(events)))

assert len(cases) == 550
assert sum(len(code)//2 for code in trace.values()) == sum(z for _, _, z in specs) == 411
assert len(entries) == 5 and set(entries.values()) == {110}
for region, (_, address, count) in zip(regions, specs):
    assert sha(bytes(u.mem_read(address, count))) == region["sha256"]
scope = ("Complete five original callers; random and whole 0054A390 drop helper are "
         "declared services. No helper/world composition or native runtime coverage inferred.")
fixture = dict(game_sha256=sha(blob), regions=regions, scope=scope, rows=rows)
proof = dict(game_sha256=sha(blob), regions=regions, contexts=len(rows),
             original_code_bytes=411, entries=entries, arithmetic_or_conversion_substitutions=0,
             original_code_unchanged=True, complete_initialized_frame_and_canaries_preserved=True,
             executed_instructions=[dict(address=hex(a), bytes=b) for a, b in sorted(trace.items())],
             scope=scope)
for dest, value in ((args.fixture, fixture), (args.proof, proof)):
    with dest.open("x") as out:
        json.dump(value, out, indent=2)
        out.write("\n")
print(json.dumps({k: v for k, v in proof.items() if k != "executed_instructions"}))
