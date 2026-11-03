"""Execute the complete original Troll death body and float-to-int converter."""
from pathlib import Path
import argparse, collections, gzip, hashlib, json, struct
import capstone
from unicorn import Uc, UcError, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_WRITE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import *

ap = argparse.ArgumentParser(description=__doc__)
ap.add_argument('--exe', type=Path, required=True)
ap.add_argument('--control', type=lambda v: int(v, 0), choices=[0xe3f, 0x23f], required=True)
ap.add_argument('--output', type=Path, required=True)
ap.add_argument('--proof', type=Path, required=True)
a = ap.parse_args(); p = Path(__file__).resolve().parent
sha = lambda raw: hashlib.sha256(raw).hexdigest()
encoded = (p / 'troll_death_independent_inputs_54a270.json.gz').read_bytes()
assert sha(encoded) == '203cc04c74a1a0aa07dbb046c1fcf83ce685657e134fffdc03068b80883c42b2'
inputs = json.loads(gzip.decompress(encoded)); assert len(inputs) == 136
raw = a.exe.read_bytes()
assert sha(raw) == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe = struct.unpack_from('<I', raw, 60)[0]; op = pe + 24
table = op + struct.unpack_from('<H', raw, pe + 20)[0]
base = struct.unpack_from('<I', raw, op + 28)[0]
image_size = (struct.unpack_from('<I', raw, op + 56)[0] + 4095) & ~4095
u = Uc(UC_ARCH_X86, UC_MODE_32); u.mem_map(base, image_size)
for i in range(struct.unpack_from('<H', raw, pe + 6)[0]):
    _, va, size, off = struct.unpack_from('<IIII', raw, table + 40 * i + 8)
    if size: u.mem_write(base + va, raw[off:off + size])
original_image = bytes(u.mem_read(base, image_size))
heap, stack, aux = 0x21000000, 0x22000000, 0x23000000
for ptr in [heap, stack, aux]: u.mem_map(ptr, 0x10000)
roles = [heap + n for n in [0x1000, 0x2000, 0x3000, 0x4000, 0x6000, 0x7000]]
ptr = lambda role: 0 if role == -1 else 0x41000000 if role == -2 else roles[role]
role = lambda address: next((n for n in [-1, -2, 0, 1, 2, 3, 4, 5] if ptr(n) == address), None)
sp, stop, init, cw, resume, double_return, double_value = [stack + 0xe000, aux + 0x100, aux + 0x200, aux + 0x300, aux + 0x400, aux + 0x500, aux + 0x600]
u.mem_write(init, b'\xdb\xe3\xd9\x2d' + struct.pack('<I', cw))
u.mem_write(cw, struct.pack('<H', a.control)); u.mem_write(resume, b'\x90')
u.mem_write(double_return, b'\xdd\x05' + struct.pack('<I', double_value) + b'\xc3')
cache_at, fps_at = 0x834a54, 0x85b3fc
services = {0x4e3aa0: ('resolve', 1), 0x4e3450: ('new', 1), 0x4daa50: ('create', 5), 0x501960: ('audio', 4), 0x419d40: ('balance', 1)}
md = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_32)
instructions, ranges = {}, []
for pc, size, expected_sha in [(0x54a270, 154, 'c21a25775cdac2ec64520866b21071063b02b89aaa5818ebfb511a685cc3f79e'), (0x419a70, 17, '10e536e2a987ea0fe14c91f007b2d5361cd87d4f0ca4cf5385fc82e53913d186')]:
    body = bytes(u.mem_read(pc, size)); assert sha(body) == expected_sha
    seq = list(md.disasm(body, pc)); assert sum(i.size for i in seq) == size and seq[-1].mnemonic == 'ret'
    instructions.update({i.address: bytes(i.bytes) for i in seq})
    ranges.append(dict(entry=hex(pc), bytes=size, sha256=sha(body)))
saved = {UC_X86_REG_EBX: 0x12345678, UC_X86_REG_ESI: 0x23456789, UC_X86_REG_EDI: 0x3456789a, UC_X86_REG_EBP: 0x456789ab}
rd = lambda address: struct.unpack('<I', u.mem_read(address, 4))[0]
active = False; seen = set(); calls = collections.Counter(); current = {}; expected_heap = bytearray(); expected_image = bytearray(); events = []; frames = []; fault = None; index = 0; writes = []
def registers(): return {key: u.reg_read(key) for key in saved}
def wr(address, raw):
    u.mem_write(address, raw)
    if heap <= address < heap + 0x10000: expected_heap[address - heap:address - heap + len(raw)] = raw
    elif base <= address < base + image_size: expected_image[address - base:address - base + len(raw)] = raw
    else: raise AssertionError(hex(address))
def mutate(plan):
    if 'position' in plan:
        obj, x, y = plan['position']; wr(ptr(obj) + 56, struct.pack('<II', x, y))
    if 'update' in plan:
        obj, data = plan['update']; wr(ptr(obj) + 748, struct.pack('<I', ptr(data)))
    if 'fps' in plan: wr(fps_at, struct.pack('<I', plan['fps']))
    if 'cache' in plan: wr(cache_at, struct.pack('<I', plan['cache']))
def finish(frame):
    assert registers() == frame['registers'] and u.reg_read(UC_X86_REG_FPSW) & 0x3800 == 0
    events.append(dict(kind=frame['kind'] + '-return', index=frame['index'], result=frame['result']))
    if frame['kind'] == 'balance':
        u.mem_write(double_value, struct.pack('<Q', frame['result'])); u.reg_write(UC_X86_REG_ESP, frame['esp']); u.reg_write(UC_X86_REG_EIP, double_return)
    else:
        u.reg_write(UC_X86_REG_EAX, frame['result'] & 0xffffffff); u.reg_write(UC_X86_REG_ESP, frame['esp'] + 4); u.reg_write(UC_X86_REG_EIP, rd(frame['esp']))
def on_code(_u, pc, size, _):
    global index
    if not active: return
    esp = u.reg_read(UC_X86_REG_ESP)
    if pc in [double_return, double_return + 6]: return
    if pc == resume:
        frame = frames.pop(); assert esp == frame['esp'] - 4 and u.reg_read(UC_X86_REG_EAX) == 1
        events.append(dict(kind='nested-return', index=frame['index'], result=1)); finish(frame); return
    if pc in services:
        kind, n = services[pc]; args = [rd(esp + i * 4) for i in range(1, n + 1)]
        at = index; index += 1; calls[kind] += 1; plan = current['script'].get(str(at), {})
        event = dict(kind=kind, index=at, cache=rd(cache_at), fps=rd(fps_at))
        if kind == 'resolve':
            assert bytes(u.mem_read(args[0], 16)).split(b'\0')[0] == b'SmallToxicCloud'; event['type'] = 'SmallToxicCloud'; result = current['resolved']
        elif kind == 'new': event['type_index'] = args[0]; result = ptr(current['new'])
        elif kind == 'create':
            assert role(args[0]) is not None and role(args[1]) is not None and args[4] == 0
            event.update(object=role(args[0]), owner=role(args[1]), position=args[2:4], trailing_zero=args[4]); result = 0xa5a50000 + at
        elif kind == 'audio':
            assert args[0] == 644 and args[2:] == [0, 0] and role(args[1]) is not None
            event.update(sound=644, unit=role(args[1]), kind_arg=0, arg=0); result = 0xa5a50000 + at
        else:
            assert bytes(u.mem_read(args[0], 24)).split(b'\0')[0] == b'SmallToxicCloudLifetime'; event['key'] = 'SmallToxicCloudLifetime'; result = current['lifetime']
        events.append(event); mutate(plan)
        frame = dict(kind=kind, index=at, esp=esp, result=result, registers=registers())
        if 'reenter' in plan:
            frames.append(frame); events.append(dict(kind='nested', index=at, unit=plan['reenter']))
            u.mem_write(esp - 8, struct.pack('<II', resume, ptr(plan['reenter'])))
            u.reg_write(UC_X86_REG_ESP, esp - 8); u.reg_write(UC_X86_REG_EIP, 0x54a270)
        else: finish(frame)
        return
    assert pc in instructions and bytes(u.mem_read(pc, size)) == instructions[pc], (current['name'], hex(pc), size)
    seen.add(pc)
def on_write(_u, access, address, size, value, _):
    if not active or stack <= address and address + size <= stack + 0x10000: return
    pc = u.reg_read(UC_X86_REG_EIP)
    if pc == 0x54a289: assert address == cache_at and size == 4
    elif pc == 0x54a2f4:
        assert size == 4 and address in roles[4:] + [0, 0x41000000]
        if address in [0, 0x41000000]: return  # The original invalid store is recorded by on_fault.
    else: raise AssertionError((current['name'], 'undeclared original write', hex(pc), hex(address), size))
    raw = value.to_bytes(size, 'little')
    if heap <= address < heap + 0x10000: expected_heap[address - heap:address - heap + size] = raw
    else: expected_image[address - base:address - base + size] = raw
    writes.append(dict(pc=hex(pc), role=role(address), cache=address == cache_at, value=value))
def on_fault(_u, access, address, size, value, _):
    global fault
    fault = dict(kind='original-memory-access', pc=hex(u.reg_read(UC_X86_REG_EIP)), access=access, address=address, size=size)
    return False
u.hook_add(UC_HOOK_CODE, on_code); u.hook_add(UC_HOOK_MEM_WRITE, on_write); u.hook_add(UC_HOOK_MEM_INVALID, on_fault)
outputs = []
for current in inputs:
    active = False; u.mem_write(base, original_image); expected_image = bytearray(original_image)
    expected_heap = bytearray((i * 17 + (i >> 7) * 29) & 255 for i in range(0x10000)); u.mem_write(heap, bytes(expected_heap))
    for n in range(4):
        wr(roles[n] + 56, struct.pack('<II', *current['positions'][n])); wr(roles[n] + 748, struct.pack('<I', ptr(current['updates'][n])))
    wr(roles[4], struct.pack('<i', -123)); wr(roles[5], struct.pack('<i', -456))
    wr(cache_at, struct.pack('<I', current['cache'])); wr(fps_at, struct.pack('<I', current['fps']))
    u.mem_write(stack, b'\xcc' * 0x10000); caller = struct.pack('<II', stop, ptr(current['unit'])) + bytes(range(48)); u.mem_write(sp, caller)
    for key, value in saved.items(): u.reg_write(key, value)
    u.reg_write(UC_X86_REG_ESP, sp); u.emu_start(init, init + 8, count=2)
    events = []; frames = []; fault = None; index = 0; writes = []; active = True
    try: u.emu_start(0x54a270, stop, count=10000)
    except UcError:
        assert fault is not None, current['name']
    active = False
    assert bytes(u.mem_read(heap, 0x10000)) == expected_heap, (current['name'], 'complete heap')
    assert bytes(u.mem_read(base, image_size)) == expected_image, (current['name'], 'complete image')
    assert bytes(u.mem_read(sp, len(caller))) == caller, (current['name'], 'caller guard')
    assert u.reg_read(UC_X86_REG_FPCW) == a.control and u.reg_read(UC_X86_REG_FPSW) & 0x3800 == 0
    result = None
    if not fault:
        assert not frames and u.reg_read(UC_X86_REG_EIP) == stop and u.reg_read(UC_X86_REG_ESP) == sp + 4 and registers() == saved
        result = u.reg_read(UC_X86_REG_EAX); assert result == 1
    outputs.append(dict(name=current['name'], result=result, fault=fault, events=events, writes=writes, cache=rd(cache_at), fps=rd(fps_at), durations=[rd(roles[4]), rd(roles[5])], positions=[list(struct.unpack('<II', u.mem_read(v + 56, 8))) for v in roles[:4]], updates=[role(rd(v + 748)) for v in roles[:4]], heap_sha256=sha(expected_heap)))
assert seen == instructions.keys(), ('missing original instructions', [hex(n) for n in instructions.keys() - seen])
decoded = json.dumps(outputs, sort_keys=True, separators=(',', ':')).encode() + b'\n'; packed = bytearray(gzip.compress(decoded, mtime=0)); packed[9] = 0xff; packed = bytes(packed)
with a.output.open('xb') as out: out.write(packed)
proof = dict(original_exe_sha256=sha(raw), input_sha256=sha(encoded), cases=len(inputs), control=hex(a.control), ranges=ranges, all171_instruction_bytes_executed=True, complete_heap_and_image_guards=True, caller_and_valid_return_stack_registers_preserved=True, FPU_CW_TOP_preserved=True, original_faults=sum(row['fault'] is not None for row in outputs), only_declared_world_services_replaced={hex(k): v[0] for k, v in services.items()}, original_reentrant_calls_executed=True, world_calls=dict(calls), output_sha256=sha(packed), decoded_output_sha256=sha(decoded), native_source_used=False)
with a.proof.open('x') as out: json.dump(proof, out, indent=2); out.write('\n')
print(json.dumps(proof), flush=True)
