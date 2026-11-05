"""Execute three unmodified PE32 bodies, including the actual nested Grip call."""
from pathlib import Path
import argparse
import collections
import gzip
import hashlib
import json
import struct

import capstone
from unicorn import (Uc, UcError, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE,
                     UC_HOOK_MEM_WRITE, UC_HOOK_MEM_INVALID)
from unicorn.x86_const import *

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--exe', type=Path, required=True)
parser.add_argument('--control', type=lambda text:int(text, 0), choices=[0xe3f, 0x23f], required=True)
parser.add_argument('--output', type=Path, required=True)
parser.add_argument('--proof', type=Path, required=True)
args = parser.parse_args()
stage = Path(__file__).resolve().parent
sha = lambda data:hashlib.sha256(data).hexdigest()
packed = (stage/'modifier_defend_predicates_independent_inputs_4e03d0.json.gz').read_bytes()
assert sha(packed) == 'adc9b7435ce18b9006cf77e7e84057d18b1ed42272bda5432dd81d99b99ad5c7'
cases = json.loads(gzip.decompress(packed))
assert len(cases) == 298
raw = args.exe.read_bytes()
assert sha(raw) == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe = struct.unpack_from('<I', raw, 60)[0]
opt = pe+24
base = struct.unpack_from('<I', raw, opt+28)[0]
size = (struct.unpack_from('<I', raw, opt+56)[0]+4095)&~4095
image = bytearray(size)
table = opt+struct.unpack_from('<H', raw, pe+20)[0]
for number in range(struct.unpack_from('<H', raw, pe+6)[0]):
    _, rva, length, disk = struct.unpack_from('<IIII', raw, table+number*40+8)
    image[rva:rva+length] = raw[disk:disk+length]
image = bytes(image)
md = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_32)
roots = dict(inversion=0x4e03d0, scan=0x4e03f0, grip=0x4e0480)
instructions, ranges = {}, []
for name, length in [('inversion', 22), ('scan', 139), ('grip', 22)]:
    address = roots[name]
    body = image[address-base:address-base+length]
    decoded = list(md.disasm(body, address))
    assert sum(item.size for item in decoded) == length and decoded[-1].mnemonic == 'ret'
    instructions.update({item.address:bytes(item.bytes) for item in decoded})
    ranges.append(dict(name=name, address=hex(address), size=length, sha256=sha(body)))
assert ranges[0]['sha256'] == 'fe6639dcaebfa8919164437706cf56e6298ca32a42c741a2fc1e770373922d6f'
assert ranges[2]['sha256'] == '0ea4bb7047e57a59eaa5ee0ebf55591a37f990c49ab8cde7aae8bafdc3924a9f'

heap, stack, aux, bad = 0x21000000, 0x22000000, 0x23000000, 0x41000000
sp, stop, initialize, cw = stack+0xe000, aux+0x100, aux+0x200, aux+0x300
effect0, effect1, item, data, output = [heap+offset for offset in [0x1000, 0x2000, 0x3000, 0x4000, 0x5000]]
saved = {UC_X86_REG_EBX:0x12345678, UC_X86_REG_ESI:0x23456789,
         UC_X86_REG_EDI:0x3456789a, UC_X86_REG_EBP:0x456789ab}
seen, faults = set(), collections.Counter()
results = []
u = Uc(UC_ARCH_X86, UC_MODE_32)
for address, length in [(base, len(image)), (heap, 0x10000), (stack, 0x10000), (aux, 0x1000)]:
    u.mem_map(address, length)
u.mem_write(base, image)
init_bytes = b'\xdb\xe3\xd9\x2d'+struct.pack('<I', cw)
u.mem_write(initialize, init_bytes)
u.mem_write(cw, struct.pack('<H', args.control))
active = False
current, writes, fault, calls = None, [], None, []
def code(uc, address, length, unused):
    if not active:
        return
    assert address in instructions and length == len(instructions[address]), hex(address)
    assert bytes(uc.mem_read(address, length)) == instructions[address]
    seen.add(address)
    if address in roots.values():
        calls.append(hex(address))
def write(uc, access, address, length, value, unused):
    if active:
        assert heap <= address and address+length <= heap+0x10000 or stack <= address and address+length <= stack+0x10000
        writes.append(dict(address=hex(address), size=length, value=value&((1<<(8*length))-1)))
def invalid(uc, access, address, length, value, unused):
    global fault
    fault = dict(access=access, address=hex(address), size=length,
                 instruction=hex(uc.reg_read(UC_X86_REG_EIP)))
    return False
u.hook_add(UC_HOOK_CODE, code)
u.hook_add(UC_HOOK_MEM_WRITE, write)
u.hook_add(UC_HOOK_MEM_INVALID, invalid)
def ptr(kind, valid):
    return {'valid':valid, 'null':0, 'bad':bad}[kind]
for case in cases:
    initial = bytearray(((number*37+case['seed']*19)^(number>>3))&255 for number in range(0x10000))
    def put(address, value):
        struct.pack_into('<I', initial, address-heap, value&0xffffffff)
    opaque = case['opaque_args']
    if case['entry'] != 'scan':
        eff = heap+0x10000-100 if case['effect']=='last-field' else ptr(case['effect'], effect0)
        if eff not in (0, bad):
            put(eff+96, case['strength'])
        dest = {'separate':output, 'field':effect0+96, 'header':effect0+4,
                'unaligned':output+1, 'null':0, 'bad':bad}[case['output']]
        call_args = [eff, *opaque, dest]
    else:
        put(item+8, case['obj_class'])
        put(item+692, ptr(case['data'], data))
        # Slot zero/one are deliberate invalid decoys and must never be read.
        put(data, bad); put(data+4, bad)
        for number, identity in enumerate(case['slots']):
            effect = effect0 if case['alias'] or number == 0 else effect1
            fn = {'grip':roots['grip'], 'inversion':roots['inversion'], 'null-fn':0, 'alien-fn':bad}
            if identity in fn:
                put(effect+88, fn[identity]); put(effect+96, case['strengths'][number])
            put(data+8+4*number, 0 if identity=='none' else bad if identity=='bad' else effect)
        call_args = [opaque[0], opaque[1], ptr(case['item'], item), opaque[2]]
    u.mem_write(heap, bytes(initial))
    initial_stack = bytes(((number*11+case['seed'])^(number>>4))&255 for number in range(0x10000))
    u.mem_write(stack, initial_stack)
    caller = struct.pack('<'+'I'*(1+len(call_args)), stop, *call_args)
    u.mem_write(sp, caller)
    u.reg_write(UC_X86_REG_ESP, sp)
    u.emu_start(initialize, initialize+len(init_bytes), count=8)
    for register, value in saved.items():
        u.reg_write(register, value)
    u.reg_write(UC_X86_REG_EAX, 0x98765432)
    u.reg_write(UC_X86_REG_EFLAGS, 0x202)
    writes, fault, calls = [], None, []
    active = True
    try:
        u.emu_start(roots[case['entry']], stop, count=256)
    except UcError:
        assert fault is not None
    finally:
        active = False
    final_heap = bytes(u.mem_read(heap, 0x10000))
    final_caller = bytes(u.mem_read(sp, len(caller)))
    assert bytes(u.mem_read(base, len(image))) == image
    assert u.reg_read(UC_X86_REG_FPCW) == args.control
    assert u.reg_read(UC_X86_REG_FPSW) == 0
    if fault is None:
        assert all(u.reg_read(register)==value for register, value in saved.items())
        assert u.reg_read(UC_X86_REG_ESP)==sp+4
        if case['entry']=='scan':
            # The original reuses argument three's caller-stack word for its
            # loop counter and argument four for Grip output. Retain both.
            assert final_caller[:12]==caller[:12]
        else:
            assert final_caller==caller
    else:
        faults[case['entry'], fault['instruction'], fault['access']]+=1
    results.append(dict(name=case['name'], initial_heap=bytes(initial).hex(),
                        final_heap=final_heap.hex(), caller_before=caller.hex(),
                        caller_after=final_caller.hex(), return_eax=None if fault else u.reg_read(UC_X86_REG_EAX),
                        fault=fault, writes=writes, actual_entry_calls=calls))
missing = sorted(set(instructions)-seen)
assert not missing, [hex(address) for address in missing]
assert len(results)==len(cases)
with args.output.open('xb') as stream:
    stream.write(gzip.compress(json.dumps(results, sort_keys=True, separators=(',', ':')).encode()+b'\n', mtime=0))
proof = dict(original_exe_sha256=sha(raw), original_image_sha256=sha(image),
             independent_inputs_sha256=sha(packed), original_image_unchanged=True,
             cases=len(cases), original_function_ranges=ranges,
             all183_original_bytes_reached=True, instruction_addresses=len(instructions),
             original_Grip_body_executed_on_real_indirect_calls=True,
             declared_external_services=[], FPU_control=args.control,
             FPU_status_unchanged=True, full64KiB_heap_and_caller_and_write_trace_retained=True,
             fault_domains=[dict(entry=k[0], instruction=k[1], access=k[2], count=v) for k,v in faults.items()],
             native_memory_equivalence=False, native_implementation_credit=False, overall100=False)
with args.proof.open('x') as stream:
    json.dump(proof, stream, indent=2); stream.write('\n')
print(json.dumps(proof), flush=True)
