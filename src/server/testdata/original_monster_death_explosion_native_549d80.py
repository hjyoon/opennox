"""Execute four complete original PE32 death bodies with declared world services."""
from pathlib import Path
import argparse, collections, datetime, gzip, hashlib, json, struct
import capstone
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_WRITE
from unicorn.x86_const import *

ap = argparse.ArgumentParser(description=__doc__)
ap.add_argument('--exe', type=Path, required=True)
ap.add_argument('--control', type=lambda s: int(s, 0), choices=[0xe3f, 0x23f], required=True)
ap.add_argument('--output', type=Path, required=True)
ap.add_argument('--proof', type=Path, required=True)
a = ap.parse_args(); p = Path(__file__).resolve().parent
sha = lambda b: hashlib.sha256(b).hexdigest()
encoded = (p / 'monster_death_explosion_independent_inputs_549d80.json.gz').read_bytes()
assert sha(encoded) == 'acec7841c6a433cfaa36972977c9b471049c1f5780e4605e5cf53521bb39caeb'
inputs = json.loads(gzip.decompress(encoded)); blob = a.exe.read_bytes()
assert sha(blob) == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe = struct.unpack_from('<I', blob, 60)[0]; op = pe + 24
table = op + struct.unpack_from('<H', blob, pe + 20)[0]
base = struct.unpack_from('<I', blob, op + 28)[0]
image_size = (struct.unpack_from('<I', blob, op + 56)[0] + 4095) & ~4095
u = Uc(UC_ARCH_X86, UC_MODE_32); u.mem_map(base, image_size)
for i in range(struct.unpack_from('<H', blob, pe + 6)[0]):
    _, va, size, offset = struct.unpack_from('<IIII', blob, table + 40*i + 8)
    if size: u.mem_write(base + va, blob[offset:offset + size])
image_before = bytes(u.mem_read(base, image_size))
heap, stack, aux = 0x21000000, 0x22000000, 0x23000000
for ptr in [heap, stack, aux]: u.mem_map(ptr, 0x10000)
objects = [heap + 0x1000, heap + 0x3000]
sp, stop, init, cw, resume = stack + 0xe000, aux + 0x100, aux + 0x200, aux + 0x300, aux + 0x400
u.mem_write(init, b'\xdb\xe3\xd9\x2d' + struct.pack('<I', cw))
u.mem_write(cw, struct.pack('<H', a.control)); u.mem_write(resume, b'\x90')
md = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_32)
instructions, ranges = {}, []
for pc, size in [(0x549d80, 124), (0x549e00, 104), (0x549e70, 27), (0x54a250, 27)]:
    raw = bytes(u.mem_read(pc, size)); seq = list(md.disasm(raw, pc))
    assert sum(i.size for i in seq) == size and seq[-1].mnemonic == 'ret'
    instructions.update({i.address: bytes(i.bytes) for i in seq})
    ranges.append(dict(entry=hex(pc), bytes=size, sha256=sha(raw)))
services = {0x40a5c0: ('flag', 1), 0x52e040: ('push', 7), 0x4e25b0: ('damage', 7), 0x5231b0: ('spark', 2), 0x501960: ('audio', 4), 0x4e5cc0: ('delete', 1), 0x522ff0: ('point', 2)}
saved = {UC_X86_REG_EBX: 0x12345678, UC_X86_REG_ESI: 0x23456789, UC_X86_REG_EDI: 0x3456789a, UC_X86_REG_EBP: 0x456789ab}
rd = lambda ptr: struct.unpack('<I', u.mem_read(ptr, 4))[0]
objptr = lambda i: 0 if i == -1 else 0x41000000 if i == -2 else objects[i]
role = lambda ptr: next((i for i in [-1, -2, 0, 1] if objptr(i) == ptr), None)
point_role = lambda ptr: next((i for i in [-1, -2, 0, 1] if objptr(i) + 56 == ptr), None)
active = False; seen = set(); calls = collections.Counter()
row = {}; expected = bytearray(); events = []; frames = []; fault = None; service_index = 0; flag_result = 0

def state():
    return [list(struct.unpack('<II', u.mem_read(v + 56, 8))) for v in objects]

def mutate(positions):
    for ptr, values in zip(objects, positions):
        raw = struct.pack('<II', *values); u.mem_write(ptr + 56, raw)
        expected[ptr + 56 - heap:ptr + 64 - heap] = raw

def registers(): return {key: u.reg_read(key) for key in saved}
def ret(esp, value):
    u.reg_write(UC_X86_REG_EAX, value & 0xffffffff)
    u.reg_write(UC_X86_REG_ESP, esp + 4); u.reg_write(UC_X86_REG_EIP, rd(esp))
def finish(frame):
    if 'after_positions' in frame['plan']: mutate(frame['plan']['after_positions'])
    assert registers() == frame['registers'], (row['name'], 'world/nested nonvolatile registers')
    assert u.reg_read(UC_X86_REG_FPSW) & 0x3800 == 0
    events.append(dict(kind=frame['kind'] + '-return', index=frame['index'], result=frame['result'], positions=state()))
    ret(frame['esp'], frame['result'])

def on_code(_u, pc, size, _):
    global service_index, flag_result, fault
    if not active: return
    esp = u.reg_read(UC_X86_REG_ESP)
    if pc == resume:
        frame = frames.pop(); assert esp == frame['esp'] - 4
        assert u.reg_read(UC_X86_REG_EAX) == 1
        events.append(dict(kind='nested-return', index=frame['index'], result=1, positions=state()))
        u.reg_write(UC_X86_REG_ESP, frame['esp']); finish(frame); return
    if pc in services:
        kind, count = services[pc]; args = [rd(esp + 4*i) for i in range(1, count + 1)]
        index = service_index; service_index += 1; calls[kind] += 1
        plan = row['script'].get(str(index), {})
        event = dict(kind=kind, index=index, positions=state())
        point = None
        if kind == 'flag':
            assert args == [2048]; event['mask'] = args[0]
        elif kind in ['push', 'damage']:
            point = args[0]; unit = role(args[4] if kind == 'push' else args[5]); assert unit is not None
            assert point_role(point) == unit
            event.update(unit=unit, point_unit=unit, outer=args[1], inner=args[2])
            if kind == 'push':
                assert args[5:] == [0, 0]; event.update(force=args[3], callback=0, callback_arg=0)
            else:
                assert args[4] == 7 and args[6] == 0; event.update(damage=args[3], damage_type=7, exclude=-1)
        elif kind in ['spark', 'point']:
            point, code = args if kind == 'spark' else [args[1], args[0]]
            unit = point_role(point); assert unit is not None
            event.update(point_unit=unit, code=code)
        elif kind == 'audio':
            assert args[0] == 42 and args[2:] == [0, 0]
            unit = role(args[1]); assert unit is not None
            event.update(sound=42, unit=unit, kind_arg=0, arg=0)
        else:
            unit = role(args[0]); assert unit is not None; event['unit'] = unit
        if point is not None and row['read_positions']:
            if point_role(point) < 0:
                events.append(event)
                fault = dict(kind='declared-world-point-read', service=kind, index=index, point_unit=point_role(point))
                u.emu_stop(); return
            event['point'] = list(struct.unpack('<II', u.mem_read(point, 8)))
        events.append(event)
        result = flag_result if kind == 'flag' else 0xa5a50000 + index
        if 'positions' in plan: mutate(plan['positions'])
        if 'flag_result' in plan: flag_result = plan['flag_result']
        frame = dict(kind=kind, index=index, result=result, esp=esp, plan=plan, registers=registers())
        if 'reenter' in plan:
            nested = plan['reenter']; assert nested['entry'] in [0x549d80, 0x549e00, 0x549e70, 0x54a250]
            events.append(dict(kind='nested', index=index, entry=nested['entry'], unit=nested['unit'], positions=state()))
            frames.append(frame); u.mem_write(esp - 8, struct.pack('<II', resume, objptr(nested['unit'])))
            u.reg_write(UC_X86_REG_ESP, esp - 8); u.reg_write(UC_X86_REG_EIP, nested['entry'])
        else: finish(frame)
        return
    assert pc in instructions and bytes(u.mem_read(pc, size)) == instructions[pc], (row['name'], hex(pc), size)
    seen.add(pc)

def on_write(_u, access, address, size, value, _):
    if active: assert stack <= address and address + size <= stack + 0x10000, (row['name'], 'undeclared original write', hex(address), size)

u.hook_add(UC_HOOK_CODE, on_code); u.hook_add(UC_HOOK_MEM_WRITE, on_write)
outputs = []; original_faults = 0; world_faults = 0
for row in inputs:
    active = False
    expected = bytearray((i*17 + (i >> 7)*29) & 255 for i in range(0x10000))
    u.mem_write(heap, bytes(expected)); mutate(row['positions'])
    u.mem_write(stack, b'\xcc'*0x10000)
    caller = struct.pack('<II', stop, objptr(row['unit'])) + bytes(range(48))
    u.mem_write(sp, caller)
    for key, value in saved.items(): u.reg_write(key, value)
    u.reg_write(UC_X86_REG_ESP, sp); u.emu_start(init, init + 8, count=2)
    events = []; frames = []; fault = None; service_index = 0; flag_result = row['flag_result']
    active = True
    u.emu_start(row['entry'], stop, count=10000)
    active = False
    assert bytes(u.mem_read(heap, 0x10000)) == expected, (row['name'], 'complete heap including canaries')
    assert bytes(u.mem_read(base, image_size)) == image_before, (row['name'], 'complete image')
    assert bytes(u.mem_read(sp, len(caller))) == caller, (row['name'], 'guarded caller')
    assert u.reg_read(UC_X86_REG_FPCW) == a.control and u.reg_read(UC_X86_REG_FPSW) & 0x3800 == 0
    result = None
    if fault:
        world_faults += 1
    else:
        assert not frames and u.reg_read(UC_X86_REG_EIP) == stop and u.reg_read(UC_X86_REG_ESP) == sp + 4
        assert registers() == saved
        result = u.reg_read(UC_X86_REG_EAX); assert result == 1
    outputs.append(dict(name=row['name'], result=result, fault=fault, events=events, positions=state(), original_heap_sha256=sha(expected)))
assert len(inputs) == 305 and world_faults == 8 and original_faults == 0
assert seen == instructions.keys(), ('incomplete instruction coverage', set(instructions) - seen)
decoded = json.dumps(outputs, sort_keys=True, separators=(',', ':')).encode() + b'\n'
encoded_output = gzip.compress(decoded, mtime=0)
with a.output.open('xb') as stream: stream.write(encoded_output)
proof = dict(utc=datetime.datetime.now(datetime.timezone.utc).isoformat(), original_exe_sha256=sha(blob), input_sha256=sha(encoded), FPU_control=hex(a.control), cases=len(inputs), original_instruction_faults=original_faults, declared_world_point_read_faults=world_faults, all282_instruction_bytes_executed=True, ranges=ranges, complete_original_image_preserved=True, complete64KiB_heap_matches_only_declared_world_mutations=True, guarded_caller_preserved=True, valid_return_nonvolatile_registers_and_stack_preserved=True, FPU_CW_TOP_preserved=True, nested_calls_execute_original_entries=True, only_declared_world_services_replaced=True, world_services={hex(k): v[0] for k, v in services.items()}, calls=dict(calls), output_sha256=sha(encoded_output), decoded_output_sha256=sha(decoded), native_source_used=False)
with a.proof.open('x') as stream: json.dump(proof, stream, indent=2); stream.write('\n')
print(json.dumps(proof), flush=True)
