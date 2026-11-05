"""Execute the complete three original raw death bodies with declared world boundaries."""
from pathlib import Path
import argparse
import collections
import gzip
import hashlib
import json
import struct

import capstone
from unicorn import (Uc, UcError, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE,
                     UC_HOOK_MEM_WRITE, UC_HOOK_MEM_INVALID, UC_HOOK_INTR)
from unicorn.x86_const import *

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--exe', type=Path, required=True)
parser.add_argument('--control', type=lambda value:int(value, 0), choices=[0xe3f, 0x23f], required=True)
parser.add_argument('--output', type=Path, required=True)
parser.add_argument('--proof', type=Path, required=True)
options = parser.parse_args()
stage = Path(__file__).resolve().parent
sha = lambda value:hashlib.sha256(value).hexdigest()
packed = (stage/'marker_boulder_generator_independent_inputs_54e460.json.gz').read_bytes()
assert sha(packed) == '33513f4d4be79a95526c408703bbe11ecb71bd7dfa69283c3d29f84d3dbf2cf8'
inputs = json.loads(gzip.decompress(packed))
assert len(inputs) == 200
raw = options.exe.read_bytes()
assert sha(raw) == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe = struct.unpack_from('<I', raw, 60)[0]
optional = pe + 24
base = struct.unpack_from('<I', raw, optional + 28)[0]
image_size = (struct.unpack_from('<I', raw, optional + 56)[0] + 4095) & ~4095
table = optional + struct.unpack_from('<H', raw, pe + 20)[0]
original_image = bytearray(image_size)
for index in range(struct.unpack_from('<H', raw, pe + 6)[0]):
    _, address, size, disk = struct.unpack_from('<IIII', raw, table + 40*index + 8)
    original_image[address:address+size] = raw[disk:disk+size]
original_image = bytes(original_image)
heap, stack, aux = 0x21000000, 0x22000000, 0x23000000
roles = [heap+offset for offset in [0x1000, 0x2000, 0x3000, 0x4000, 0x5000, 0x6000, 0x9000, 0xa000]]
pointer = lambda role:0 if role == -1 else 0x41000000 if role == -2 else roles[role]
role = lambda value:next((number for number in [-1, -2, *range(8)] if pointer(number) == value), None)
sp, stop, initialize, cw, resume, float_return, float_value = [
    stack+0xe000, aux+0x100, aux+0x200, aux+0x300, aux+0x400, aux+0x500, aux+0x600]
initial_aux = bytearray(0x10000)
initial_aux[initialize-aux:initialize-aux+8] = b'\xdb\xe3\xd9\x2d'+struct.pack('<I', cw)
initial_aux[cw-aux:cw-aux+2] = struct.pack('<H', options.control)
initial_aux[resume-aux] = 0x90
initial_aux[float_return-aux:float_return-aux+7] = b'\xdd\x05'+struct.pack('<I', float_value)+b'\xc3'
index_at, divisor_at, fps_at, frame_at = 0x834acc, 0x5cde24, 0x85b3fc, 0x84ea04
names_at, mass_at = 0x5ce2b8, 0x5cde18
names = ['BigRock', 'MediumRock', 'SmallRock', 'DestroyedGenerator']
def initial_u32(address):
    return struct.unpack_from('<I', original_image, address-base)[0]
def initial_name(address):
    return original_image[address-base:address-base+48].split(b'\0', 1)[0].decode('ascii')
stock_names = [initial_name(initial_u32(names_at+4*number)) for number in range(12)]
stock_mass = list(original_image[mass_at-base:mass_at-base+12])
assert stock_names == ['BigRock', 'MediumRock', 'SmallRock', 'MediumRock', 'SmallRock',
                       'SmallRock', 'MediumRock', 'SmallRock', 'SmallRock', 'SmallRock', 'SmallRock', 'SmallRock']
assert stock_mass == [1, 3, 5, 3, 5, 5, 3, 5, 5, 5, 5, 5]
assert initial_u32(divisor_at) == 12 and initial_name(0x5ce43c) == 'DestroyedGenerator'
name_addresses = [next(initial_u32(names_at+4*i) for i in range(12)
                       if stock_names[i] == name) for name in names[:3]] + [0x5ce43c]
name_pointer = lambda number:0 if number == -1 else 0x41000000 if number == -2 else name_addresses[number]
services = {
    0x4ec580:('owner', 1), 0x522ff0:('pointfx', 2), 0x4e5cc0:('delete', 1),
    0x501960:('audio', 4), 0x415fa0:('int', 4), 0x4e3810:('new', 1),
    0x4ed970:('point', 3), 0x4daa50:('create', 5), 0x416030:('float', 4),
    0x4e46f0:('raise', 2), 0x52df80:('force', 3), 0x511660:('decay', 2),
    0x4d71e0:('timer', 1), 0x4d7520:('mode', 1), 0x502490:('script', 3),
    0x523200:('generatorfx', 2), 0x40a5c0:('flags', 1), 0x4d61b0:('credit', 1),
}
md = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_32)
instructions, ranges = {}, []
for address, size, expected in [
    (0x54e460, 77, '20f351cb74ee7120bd3f87114ee15b8d9fb437344ec5e9a0a218fad792af4d57'),
    (0x54e4b0, 366, '269ac493c49a805f472061a9f622cdeef9167b0ebb43cd1e9c58d8545b9be7f3'),
    (0x54e630, 178, 'db2bdc0a2fa4c4d468eb5a270f6d75069355ac46d40882b321e6ea3ee5c8680c'),
]:
    body = original_image[address-base:address-base+size]
    assert sha(body) == expected
    decoded = list(md.disasm(body, address))
    assert sum(item.size for item in decoded) == size and decoded[-1].mnemonic == 'ret'
    instructions.update({item.address:bytes(item.bytes) for item in decoded})
    ranges.append(dict(address=hex(address), size=size, sha256=expected))
saved = {UC_X86_REG_EBX:0x12345678, UC_X86_REG_ESI:0x23456789,
         UC_X86_REG_EDI:0x3456789a, UC_X86_REG_EBP:0x456789ab}
u = None
active = False
seen, calls, observed_faults = set(), collections.Counter(), collections.Counter()
current, expected_heap, expected_image, expected_aux = {}, bytearray(), bytearray(), bytearray()
events, frames, occurrences, writes = [], [], {}, []
fault = fault_detail = None
quest, owner_return, reentries = 0, 1, 0
rd = lambda address:struct.unpack('<I', u.mem_read(address, 4))[0]
registers = lambda:{key:u.reg_read(key) for key in saved}
def write(address, value):
    u.mem_write(address, value)
    for start, expected in [(heap, expected_heap), (base, expected_image), (aux, expected_aux)]:
        if start <= address and address+len(value) <= start+len(expected):
            expected[address-start:address-start+len(value)] = value
            return
    assert stack <= address and address+len(value) <= stack+0x10000, hex(address)
def world_read(address, size, site):
    global fault, fault_detail
    try:
        return bytes(u.mem_read(address, size))
    except UcError:
        fault = dict(kind='memory', site=site)
        fault_detail = dict(kind='declared-world-read', site=site, address=hex(address), size=size)
        raise
def origin(address):
    return list(struct.unpack('<II', world_read(address, 8, 'origin')))
def motion(address):
    return [struct.unpack('<I', world_read(address+offset, 4, 'part'))[0]
            for offset in [16, 108, 116]]
def state():
    return dict(part_index=rd(index_at), divisor=rd(divisor_at), fps=rd(fps_at), frame=rd(frame_at),
                quest=quest, owner_return=owner_return)
def slot_state():
    return [role(rd(roles[6]+116+4*index)) for index in range(4)]
def mutate(plan):
    global quest, owner_return
    for key, address in [('index', index_at), ('divisor', divisor_at), ('fps', fps_at), ('frame', frame_at)]:
        if key in plan:
            write(address, struct.pack('<I', plan[key]&0xffffffff))
    for key, offset in [('position', 56), ('flags', 16), ('class', 8)]:
        if key in plan:
            target, *values = plan[key]
            write(pointer(target)+offset, struct.pack('<'+'I'*len(values), *values))
    for key, offset in [('update', 748), ('owner', 520)]:
        if key in plan:
            target, value = plan[key]
            write(pointer(target)+offset, struct.pack('<I', pointer(value)))
    if 'slot' in plan:
        index, target = plan['slot']
        write(roles[6]+116+4*index, struct.pack('<I', pointer(target)))
    if 'table' in plan:
        index, target = plan['table']
        write(names_at+4*index, struct.pack('<I', name_pointer(target)))
    if 'mass' in plan:
        index, value = plan['mass']
        write(mass_at+index, bytes([value]))
    if 'quest' in plan:
        quest = plan['quest']
    if 'owner_return' in plan:
        owner_return = plan['owner_return']
def finish(frame):
    assert registers() == frame['registers'] and u.reg_read(UC_X86_REG_FPSW)&0x3800 == 0
    events.append(dict(kind=frame['kind']+'-return', index=frame['index'], result=frame['semantic']))
    if frame['kind'] == 'float':
        write(float_value, struct.pack('<Q', frame['result']))
        u.reg_write(UC_X86_REG_ESP, frame['esp'])
        u.reg_write(UC_X86_REG_EIP, float_return)
    else:
        u.reg_write(UC_X86_REG_EAX, frame['result']&0xffffffff)
        u.reg_write(UC_X86_REG_ESP, frame['esp']+4)
        u.reg_write(UC_X86_REG_EIP, rd(frame['esp']))
def on_code(_u, address, size, _):
    global reentries
    if not active:
        return
    esp = u.reg_read(UC_X86_REG_ESP)
    if address in (float_return, float_return+6):
        expected = bytes(expected_aux[address-aux:address-aux+size])
        assert bytes(u.mem_read(address, size)) == expected
        return
    if address == resume:
        frame = frames.pop()
        assert esp == frame['esp']-4
        events.append(dict(kind='nested-return', index=frame['index']))
        finish(frame)
        return
    if address in services:
        kind, count = services[address]
        args = [rd(esp+4*number) for number in range(1, count+1)]
        sequence = occurrences.get(kind, 0)
        occurrences[kind] = sequence+1
        calls[kind] += 1
        at = sum(event['kind'] in {value[0] for value in services.values()} for event in events)
        plan = current['script'].get(f'{kind}:{sequence}', {})
        event = dict(kind=kind, index=at, **state())
        result = semantic = 0xa5a50000+at
        if kind == 'owner':
            event.update(unit=role(args[0]))
            semantic = plan.get('return_role', owner_return)
            result = pointer(semantic)
        elif kind == 'pointfx':
            assert args[0] == 0x8a
            event.update(effect=args[0], origin_pointer=role((args[1]-56)&0xffffffff), slots=slot_state())
        elif kind == 'delete':
            event.update(unit=role(args[0]))
        elif kind == 'audio':
            assert args[0] in (757, 1000) and args[2:] == [0, 0]
            event.update(sound=args[0], unit=role(args[1]))
        elif kind == 'int':
            assert args[2] in (0x5ce374, 0x5ce414) and args[3] in (0x6a3, 0x6b4)
            assert initial_name(args[2]) == 'C:\\NoxPost\\src\\Server\\Object\\die\\Die.c'
            event.update(min=args[0], max=args[1])
            result = semantic = current['parts'] if args[3] == 0x6a3 else current['seconds']
        elif kind == 'float':
            assert args[2] in (0x5ce39c, 0x5ce3c4, 0x5ce3ec)
            assert args[3] in (0x6ad, 0x6ae, 0x6b1)
            event.update(min=args[0], max=args[1])
            result = semantic = current['float_bits']
        elif kind == 'new':
            name = world_read(args[0], 48, 'name').split(b'\0', 1)[0].decode('ascii') if args[0] else None
            semantic = plan.get('return_role', current['new'].get(name, -1))
            result = pointer(semantic)
            event.update(name=name)
        elif kind == 'point':
            event.update(radius=args[0], origin=origin(args[1]))
            write(args[2], struct.pack('<II', 0x42c80000+sequence, 0x43480000+sequence))
        elif kind == 'create':
            assert role(args[0]) is not None and role(args[1]) == -1 and args[4] == 0
            event.update(object=role(args[0]), owner=-1, position=args[2:4], extra=0)
        elif kind == 'raise':
            event.update(object=role(args[0]), height=args[1])
        elif kind == 'force':
            event.update(object=role(args[1]), origin=origin(args[0]), force=args[2], motion=motion(args[1]))
        elif kind == 'decay':
            event.update(object=role(args[0]), duration=args[1], motion=motion(args[0]))
        elif kind == 'timer':
            event.update(value=args[0])
        elif kind == 'mode':
            assert args == [0]
            event.update(value=0)
        elif kind == 'script':
            event.update(callback_data_role=role((args[0]-56)&0xffffffff), caller=role(args[1]), trigger=role(args[2]), event=3)
        elif kind == 'generatorfx':
            assert args[1] == 200
            event.update(origin_pointer=role((args[0]-56)&0xffffffff), intensity=args[1])
        elif kind == 'flags':
            assert args == [4096]
            result = semantic = quest
        elif kind == 'credit':
            event.update(unit=role(args[0]))
        else:
            raise AssertionError(kind)
        events.append(event)
        mutate(plan)
        frame = dict(kind=kind, index=at, esp=esp, result=result, semantic=semantic, registers=registers())
        if 'reenter' in plan:
            entry, unit = plan['reenter']
            frames.append(frame)
            reentries += 1
            events.append(dict(kind='nested', index=at, entry=entry, unit=unit))
            u.mem_write(esp-8, struct.pack('<II', resume, pointer(unit)))
            u.reg_write(UC_X86_REG_ESP, esp-8)
            u.reg_write(UC_X86_REG_EIP, int(entry, 16))
        else:
            finish(frame)
        return
    assert address in instructions and bytes(u.mem_read(address, size)) == instructions[address], (current['name'], hex(address))
    seen.add(address)
def on_write(_u, access, address, size, value, _):
    if not active:
        return
    if stack <= address and address+size <= stack+0x10000:
        assert address+size <= sp, (current['name'], 'original caller write', hex(address), size)
        return
    pc = u.reg_read(UC_X86_REG_EIP)
    assert pc in (0x54e48c, 0x54e57a, 0x54e5b0, 0x54e5b3, 0x54e5fd) and size == 4, (hex(pc), hex(address), size)
    if pc == 0x54e5fd:
        assert address == index_at
    writes.append(dict(pc=hex(pc), address=hex(address), value=value&0xffffffff))
    for start, expected in [(heap, expected_heap), (base, expected_image)]:
        if start <= address and address+size <= start+len(expected):
            expected[address-start:address-start+size] = (value&0xffffffff).to_bytes(size, 'little')
            return
def on_fault(_u, access, address, size, value, _):
    global fault, fault_detail
    pc = u.reg_read(UC_X86_REG_EIP)
    site = {0x54e472:'parent-update', 0x54e47d:'marker-slot', 0x54e48c:'marker-slot',
            0x54e500:'name-table', 0x54e58f:'mass-table', 0x54e63c:'source-update',
            0x54e64e:'source-owner', 0x54e68d:'source-owner', 0x54e69d:'parent-class',
            0x54e6c2:'source-y', 0x54e6c5:'source-x'}.get(pc, 'part')
    fault = dict(kind='memory', site=site)
    fault_detail = dict(kind='original-memory-access', pc=hex(pc), access=access, address=hex(address), size=size)
    return False
def on_interrupt(_u, number, _):
    global fault, fault_detail
    assert number == 0 and u.reg_read(UC_X86_REG_EIP) == 0x54e5f4
    fault = dict(kind='arithmetic', site='boulder-modulo')
    fault_detail = dict(kind='original-divide-error', pc='0x54e5f4', interrupt=number)
    u.emu_stop()

outputs, details = [], []
for current in inputs:
    active = False
    u = Uc(UC_ARCH_X86, UC_MODE_32)
    u.mem_map(base, image_size)
    for address in (heap, stack, aux):
        u.mem_map(address, 0x10000)
    expected_image, expected_aux = bytearray(original_image), bytearray(initial_aux)
    u.mem_write(base, original_image)
    u.mem_write(aux, bytes(initial_aux))
    expected_heap = bytearray((number*17+(number>>7)*29)&255 for number in range(0x10000))
    u.mem_write(heap, bytes(expected_heap))
    for number in range(6):
        for offset, values in [(8, [2]), (16, [0x11223344+number]),
                               (56, [0x40e00000+number, 0x41000000+number]),
                               (520, [pointer(1)]), (748, [pointer(7)])]:
            write(roles[number]+offset, struct.pack('<'+'I'*len(values), *values))
    write(roles[0]+748, struct.pack('<I', pointer(current['source_update'])))
    write(roles[0]+520, struct.pack('<I', pointer(current['source_owner'])))
    write(roles[1]+748, struct.pack('<I', pointer(current['parent_update'])))
    write(roles[1]+8, struct.pack('<I', current['parent_class']))
    for number, target in enumerate(current['slots']):
        write(roles[6]+116+4*number, struct.pack('<I', pointer(target)))
    for key, address in [('index', index_at), ('divisor', divisor_at), ('fps', fps_at), ('frame', frame_at)]:
        write(address, struct.pack('<I', current[key]&0xffffffff))
    quest, owner_return = current['quest'], current['owner_return']
    u.mem_write(stack, b'\xcc'*0x10000)
    caller = struct.pack('<II', stop, pointer(current['unit']))+bytes(range(48))
    u.mem_write(sp, caller)
    for key, value in saved.items():
        u.reg_write(key, value)
    u.reg_write(UC_X86_REG_ESP, sp)
    u.emu_start(initialize, initialize+8, count=2)
    for hook, callback in [(UC_HOOK_CODE, on_code), (UC_HOOK_MEM_WRITE, on_write),
                           (UC_HOOK_MEM_INVALID, on_fault), (UC_HOOK_INTR, on_interrupt)]:
        u.hook_add(hook, callback)
    events, frames, occurrences, writes = [], [], {}, []
    fault = fault_detail = None
    active = True
    try:
        u.emu_start(int(current['entry'], 16), stop, count=100000)
    except UcError:
        assert fault is not None, current['name']
    active = False
    assert bytes(u.mem_read(heap, 0x10000)) == expected_heap, (current['name'], 'whole heap')
    assert bytes(u.mem_read(base, image_size)) == expected_image, (current['name'], 'whole image')
    assert bytes(u.mem_read(aux, 0x10000)) == expected_aux, (current['name'], 'whole helper region')
    assert bytes(u.mem_read(sp, len(caller))) == caller, (current['name'], 'caller guard')
    assert u.reg_read(UC_X86_REG_FPCW) == options.control, current['name']
    if not fault:
        assert not frames and u.reg_read(UC_X86_REG_EIP) == stop and u.reg_read(UC_X86_REG_ESP) == sp+4
        assert registers() == saved and u.reg_read(UC_X86_REG_FPSW)&0x3800 == 0
    else:
        observed_faults[fault['kind']+':'+fault['site']] += 1
        details.append(dict(name=current['name'], **fault_detail))
    outputs.append(dict(name=current['name'], fault=fault, events=events, **state(), slots=slot_state(),
                        heap_sha256=sha(expected_heap), image_sha256=sha(expected_image),
                        original_writes=writes))
missing = sorted(instructions.keys()-seen)
assert not missing, ('unexecuted original instructions', [hex(address) for address in missing])
decoded_output = json.dumps(outputs, sort_keys=True, separators=(',', ':')).encode()+b'\n'
encoded_output = bytearray(gzip.compress(decoded_output, compresslevel=9, mtime=0))
encoded_output[9] = 0xff
with options.output.open('xb') as stream:
    stream.write(encoded_output)
proof = dict(original_executable_sha256=sha(raw), input_sha256=sha(packed), cases=len(inputs),
             control=hex(options.control), ranges=ranges, all621_instruction_bytes_executed=True,
             complete_heap_and_image_guards=True, complete_helper_region_guard=True,
             caller_bytes_unchanged=True, valid_return_stack_and_nonvolatile_registers_preserved=True,
             FPU_control_and_valid_TOP_preserved=True, original_faults=dict(observed_faults),
             fault_details=details, only_declared_world_services_replaced={hex(key):value[0] for key, value in services.items()},
             original_reentrant_calls_executed=reentries, world_calls=dict(calls),
             stock_names=stock_names, stock_mass=stock_mass,
             pointFX_boundary='Pointer observation only; boundary spies do not execute pointFX callees or read pointed coordinates',
             script_boundary='Original body passes three words; semantic event3 is the declared native world-service projection, not a fourth original stack word',
             output_sha256=sha(encoded_output), decoded_output_sha256=sha(decoded_output), native_source_used=False,
             actual_unchanged_C_fault_credit=False, native_comparison_credit=False)
with options.proof.open('x') as stream:
    json.dump(proof, stream, indent=2)
    stream.write('\n')
print(json.dumps(proof), flush=True)
