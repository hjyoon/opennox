"""Run complete original ImpEgg/Polyp/Potion bodies and the real FISTP helper."""
import argparse
import collections
import gzip
import hashlib
import json
import struct
from pathlib import Path

import capstone
from unicorn import (Uc, UcError, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE,
                     UC_HOOK_MEM_WRITE, UC_HOOK_MEM_INVALID)
from unicorn.x86_const import *

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument("--exe", type=Path, required=True)
parser.add_argument("--control", type=lambda v: int(v, 0), choices=[0xe3f, 0x23f], required=True)
parser.add_argument("--output", type=Path, required=True)
parser.add_argument("--proof", type=Path, required=True)
options = parser.parse_args()
sha = lambda data: hashlib.sha256(data).hexdigest()
encoded = Path(__file__).with_name("simple_polyp_death_independent_inputs_54cae0.json.gz").read_bytes()
# Pinned after the independent generator is run, before any native entry changes.
INPUT_SHA256 = "54334f7371c06bc3f046293acc31e87004e286bb7293600b3d1819245d705d26"
assert sha(encoded) == INPUT_SHA256
inputs = json.loads(gzip.decompress(encoded))
assert len(inputs) == 149
raw = options.exe.read_bytes()
assert sha(raw) == "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a"
pe = struct.unpack_from("<I", raw, 60)[0]
optional = pe + 24
table = optional + struct.unpack_from("<H", raw, pe + 20)[0]
base = struct.unpack_from("<I", raw, optional + 28)[0]
image_size = (struct.unpack_from("<I", raw, optional + 56)[0] + 4095) & ~4095
image = bytearray(image_size)
for i in range(struct.unpack_from("<H", raw, pe + 6)[0]):
    _, va, size, offset = struct.unpack_from("<IIII", raw, table + 40 * i + 8)
    if size:
        image[va:va + size] = raw[offset:offset + size]
heap, stack, auxiliary = 0x21000000, 0x22000000, 0x23000000
roles = [heap + offset for offset in [0x1000, 0x2000, 0x3000, 0x4000, 0x5000, 0x6000, 0x7000]]


def pointer(role):
    return {-4: heap - 16, -3: heap - 60, -2: 0x41000000, -1: 0}.get(role, roles[role] if role >= 0 else None)


def role(value):
    found = next((n for n in range(-4, 7) if pointer(n) == value), None)
    assert found is not None, (current["name"], "undeclared pointer", hex(value))
    return found


sp, stop, initialize, cw, resume, float_return, float_value = [
    stack + 0xe000, auxiliary + 0x100, auxiliary + 0x200, auxiliary + 0x300,
    auxiliary + 0x400, auxiliary + 0x500, auxiliary + 0x600]
cache_at, fps_at = 0x834aac, 0x85b3fc
services = {0x4e3aa0: ("name", 1), 0x4e3450: ("new", 1),
            0x4daa50: ("create", 5), 0x419d40: ("lifetime", 1),
            0x501960: ("audio", 4), 0x4e5cc0: ("delete", 1)}
instructions, ranges = {}, []
md = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_32)
for address, size, digest in [
    (0x54cae0, 33, "6f42bc269662b3a3b55c158088b8b2d8e8f9672afc01f675830e826bf1cf5b5b"),
    (0x54cb10, 150, "c6edef6b059d3697a5859492ab544beb3000860b6482d78cdf6a119ca414cd14"),
    (0x54cbb0, 31, "09411c20fc834d61f4bf98d6b418e4a71ab5d0fc16832995c667bf3308a830d0"),
    (0x419a70, 17, "10e536e2a987ea0fe14c91f007b2d5361cd87d4f0ca4cf5385fc82e53913d186"),
]:
    body = bytes(image[address - base:address - base + size])
    assert sha(body) == digest
    sequence = list(md.disasm(body, address))
    assert sum(item.size for item in sequence) == size and sequence[-1].mnemonic == "ret"
    instructions.update({item.address: bytes(item.bytes) for item in sequence})
    ranges.append(dict(entry=hex(address), bytes=size, sha256=digest))
saved = {UC_X86_REG_EBX: 0x12345678, UC_X86_REG_ESI: 0x23456789,
         UC_X86_REG_EDI: 0x3456789a, UC_X86_REG_EBP: 0x456789ab}
seen, calls, faults = set(), collections.Counter(), collections.Counter()
reentries = conversions = 0
active = False


def read_u32(address):
    return struct.unpack("<I", u.mem_read(address, 4))[0]


def registers():
    return {key: u.reg_read(key) for key in saved}


def write(address, data):
    u.mem_write(address, data)
    if heap <= address and address + len(data) <= heap + 0x10000:
        expected_heap[address - heap:address - heap + len(data)] = data
    elif base <= address and address + len(data) <= base + image_size:
        expected_image[address - base:address - base + len(data)] = data
    else:
        raise AssertionError(("undeclared world write", hex(address)))


def state():
    return dict(cache=read_u32(cache_at), fps=read_u32(fps_at))


def mutate(plan):
    for key, offset in [("position", 56), ("flags", 16), ("update", 748)]:
        if key in plan:
            target, *values = plan[key]
            if key == "update":
                values = [pointer(values[0])]
            write(pointer(target) + offset, struct.pack("<" + "I" * len(values), *values))
    for key, address in [("cache", cache_at), ("fps", fps_at)]:
        if key in plan:
            write(address, struct.pack("<I", plan[key]))


def finish(frame):
    assert registers() == frame["registers"] and u.reg_read(UC_X86_REG_FPSW) & 0x3800 == 0
    events.append(dict(kind=frame["kind"] + "-return", index=frame["index"], result=frame["semantic"]))
    u.reg_write(UC_X86_REG_ESP, frame["esp"])
    if frame["kind"] == "lifetime":
        u.mem_write(float_value, struct.pack("<I", frame["result"]))
        u.reg_write(UC_X86_REG_EIP, float_return)
    else:
        u.reg_write(UC_X86_REG_EAX, frame["result"])
        u.reg_write(UC_X86_REG_ESP, frame["esp"] + 4)
        u.reg_write(UC_X86_REG_EIP, read_u32(frame["esp"]))


def on_code(_u, pc, size, _):
    global reentries, conversions
    if not active:
        return
    esp = u.reg_read(UC_X86_REG_ESP)
    if pc in [float_return, float_return + 6]:
        return
    if pc == resume:
        frame = frames.pop()
        assert esp == frame["esp"] - 4
        result = u.reg_read(UC_X86_REG_EAX) if frame["nested_entry"] == "54cae0" else None
        events.append(dict(kind="nested-return", index=frame["index"], result=result))
        finish(frame)
        return
    if pc in services:
        kind, count = services[pc]
        args = [read_u32(esp + i * 4) for i in range(1, count + 1)]
        index = sum(event["kind"] in {v[0] for v in services.values()} for event in events)
        occurrence = occurrences.get(kind, 0)
        occurrences[kind] = occurrence + 1
        calls[kind] += 1
        plan = current["script"].get(f"{kind}:{occurrence}", {})
        event = dict(kind=kind, index=index, **state())
        result, semantic = 0xa5a50000 + index, None
        if kind in ["name", "lifetime"]:
            name = bytes(u.mem_read(args[0], 48)).split(b"\0", 1)[0].decode("ascii")
            assert name == ("ToxicCloud" if kind == "name" else "ToxicCloudLifetime")
            event["name"] = name
            result = semantic = plan.get("result", current["name_result" if kind == "name" else "lifetime_bits"])
        elif kind == "new":
            event["type_id"] = args[0]
            semantic = plan.get("result", current["cloud"])
            result = pointer(semantic)
        elif kind == "create":
            assert args[1] == 0 and args[4] == 0
            event.update(object=role(args[0]), owner=-1, position=args[2:4], extra=0)
        elif kind == "audio":
            assert args[0] in [764, 284, 753] and args[2:] == [0, 0]
            event.update(sound=args[0], unit=role(args[1]))
        elif kind == "delete":
            event["unit"] = role(args[0])
        events.append(event)
        mutate(plan)
        frame = dict(kind=kind, index=index, esp=esp, result=result,
                     semantic=semantic, registers=registers())
        if "reenter" in plan:
            entry, unit = plan["reenter"]
            frame["nested_entry"] = entry
            frames.append(frame)
            reentries += 1
            events.append(dict(kind="nested", index=index, entry=entry, unit=unit))
            u.mem_write(esp - 8, struct.pack("<II", resume, pointer(unit)))
            u.reg_write(UC_X86_REG_ESP, esp - 8)
            u.reg_write(UC_X86_REG_EIP, int(entry, 16))
        else:
            finish(frame)
        return
    assert pc in instructions and bytes(u.mem_read(pc, size)) == instructions[pc], (current["name"], hex(pc), size)
    seen.add(pc)
    if pc == 0x419a70:
        conversions += 1


def on_write(_u, access, address, size, value, _):
    if not active:
        return
    if stack <= address and address + size <= stack + 0x10000:
        assert address + size <= sp, (current["name"], "caller stack write", hex(address))
        return
    pc = u.reg_read(UC_X86_REG_EIP)
    assert size == 4
    if pc == 0x54cafc:
        role(address - 16)
    elif pc == 0x54cb29:
        assert address == cache_at
    else:
        assert pc == 0x54cb86
        role(address)
    if heap <= address and address + size <= heap + 0x10000:
        expected_heap[address - heap:address - heap + size] = struct.pack("<I", value)
    elif base <= address and address + size <= base + image_size:
        expected_image[address - base:address - base + size] = struct.pack("<I", value)


def on_fault(_u, access, address, size, value, _):
    global fault, fault_detail
    pc = u.reg_read(UC_X86_REG_EIP)
    sites = {0x54caf4: "flags", 0x54cafc: "flags", 0x54cb40: "source-y",
             0x54cb43: "source-x", 0x54cb47: "update", 0x54cb86: "duration"}
    assert pc in sites, (current["name"], "undeclared memory fault", hex(pc))
    fault = dict(kind="memory", site=sites[pc])
    fault_detail = dict(pc=hex(pc), access=access, address=hex(address), size=size)
    return False


outputs, fault_details = [], []
for current in inputs:
    active = False
    u = Uc(UC_ARCH_X86, UC_MODE_32)
    for address, size in [(base, image_size), (heap, 0x10000), (stack, 0x10000), (auxiliary, 0x10000)]:
        u.mem_map(address, size)
    expected_image = bytearray(image)
    u.mem_write(base, bytes(image))
    expected_heap = bytearray((i * 17 + (i >> 7) * 29) & 255 for i in range(0x10000))
    u.mem_write(heap, bytes(expected_heap))
    for n in range(4):
        for offset, values in [(16, [(current["flags"] + n) & 0xffffffff]),
                               (56, [0x40e00000 + n, 0x41000000 + n]),
                               (748, [pointer(current["update"])])]:
            write(roles[n] + offset, struct.pack("<" + "I" * len(values), *values))
    for n in range(4, 7):
        write(roles[n], struct.pack("<I", 0xdead0000 + n))
    write(cache_at, struct.pack("<I", current["cache"]))
    write(fps_at, struct.pack("<I", current["fps"]))
    u.mem_write(stack, b"\xcc" * 0x10000)
    caller = struct.pack("<II", stop, pointer(current["unit"])) + bytes(range(48))
    u.mem_write(sp, caller)
    u.mem_write(initialize, b"\xdb\xe3\xd9\x2d" + struct.pack("<I", cw))
    u.mem_write(cw, struct.pack("<H", options.control))
    u.mem_write(resume, b"\x90")
    u.mem_write(float_return, b"\xd9\x05" + struct.pack("<I", float_value) + b"\xc3")
    for key, value in saved.items():
        u.reg_write(key, value)
    u.reg_write(UC_X86_REG_ESP, sp)
    u.emu_start(initialize, initialize + 8, count=2)
    for hook, callback in [(UC_HOOK_CODE, on_code), (UC_HOOK_MEM_WRITE, on_write), (UC_HOOK_MEM_INVALID, on_fault)]:
        u.hook_add(hook, callback)
    events, frames, occurrences = [], [], {}
    fault, fault_detail = None, None
    active = True
    try:
        u.emu_start(int(current["entry"], 16), stop, count=100000)
    except UcError:
        assert fault is not None, current["name"]
    active = False
    assert bytes(u.mem_read(heap, 0x10000)) == expected_heap, (current["name"], "complete heap")
    assert bytes(u.mem_read(base, image_size)) == expected_image, (current["name"], "complete image")
    assert bytes(u.mem_read(sp, len(caller))) == caller, (current["name"], "caller guard")
    assert u.reg_read(UC_X86_REG_FPCW) == options.control
    if not fault:
        assert not frames and u.reg_read(UC_X86_REG_EIP) == stop
        assert u.reg_read(UC_X86_REG_ESP) == sp + 4 and registers() == saved
        assert u.reg_read(UC_X86_REG_FPSW) & 0x3800 == 0
    else:
        faults[fault["site"]] += 1
        fault_details.append(dict(name=current["name"], **fault_detail))
    result = u.reg_read(UC_X86_REG_EAX) if not fault and current["entry"] == "54cae0" else None
    outputs.append(dict(name=current["name"], result=result, fault=fault, events=events,
                        **state(), heap_sha256=sha(expected_heap)))
assert instructions.keys() <= seen, [hex(pc) for pc in instructions.keys() - seen]
decoded = json.dumps(outputs, sort_keys=True, separators=(",", ":")).encode() + b"\n"
packed = bytearray(gzip.compress(decoded, compresslevel=9, mtime=0))
packed[9] = 0xff
with options.output.open("xb") as stream:
    stream.write(packed)
proof = dict(original_exe_sha256=sha(raw), input_sha256=sha(encoded), cases=len(inputs),
             control=hex(options.control), ranges=ranges, all231_instruction_bytes_executed=True,
             newly_inventoried_body_bytes=64, preexisting_Polyp_body_bytes=150, preexisting_converter_bytes=17,
             full_heap_image_and_caller_guards=True, valid_return_stack_registers_and_FPU_preserved=True,
             only_declared_world_services_replaced={hex(pc): value[0] for pc, value in services.items()},
             real_FISTP_calls=conversions, FP_arithmetic_conversion_substitutions=0,
             reentrant_original_calls=reentries, memory_faults=dict(faults), fault_details=fault_details,
             world_calls=dict(calls), output_sha256=sha(packed), decoded_output_sha256=sha(decoded),
             native_source_used=False)
with options.proof.open("x") as stream:
    json.dump(proof, stream, indent=2)
    stream.write("\n")
print(json.dumps(proof), flush=True)
