"""Execute the unchanged 58-byte generator-credit body, including live pointer reloads."""
from pathlib import Path
import argparse
import collections
import gzip
import hashlib
import json
import struct
import capstone
from unicorn import Uc, UcError, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_WRITE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import *

parser = argparse.ArgumentParser(description=__doc__)
for name in ['exe', 'output', 'proof']:
    parser.add_argument('--'+name, type=Path, required=True)
parser.add_argument('--control', type=lambda value:int(value, 0), choices=[0xe3f, 0x23f], required=True)
options = parser.parse_args()
stage = Path(__file__).resolve().parent
sha = lambda value:hashlib.sha256(value).hexdigest()
encoded = (stage/'credit_independent_inputs_4d61b0.json.gz').read_bytes()
assert sha(encoded) == 'a086a617a8b90e951ac06183f77992dfb07f9c6b811757a36c668be3dff1e542'
inputs = json.loads(gzip.decompress(encoded))
assert len(inputs) == 286
raw = options.exe.read_bytes()
assert sha(raw) == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe = struct.unpack_from('<I', raw, 60)[0]
optional = pe+24
base = struct.unpack_from('<I', raw, optional+28)[0]
size = (struct.unpack_from('<I', raw, optional+56)[0]+4095)&~4095
table = optional+struct.unpack_from('<H', raw, pe+20)[0]
original = bytearray(size)
for index in range(struct.unpack_from('<H', raw, pe+6)[0]):
    _, address, count, offset = struct.unpack_from('<IIII', raw, table+40*index+8)
    original[address:address+count] = raw[offset:offset+count]
original = bytes(original)
body = original[0x4d61b0-base:0x4d61b0-base+58]
assert sha(body) == 'd2cd4243d76370abbb0e7498bb59b9e4aee8617dc183598f2b31ede0847ca717'
md = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_32)
instructions = {item.address:bytes(item.bytes) for item in md.disasm(body, 0x4d61b0)}
assert sum(map(len, instructions.values())) == 58
heap, stack, aux = 0x21000000, 0x22000000, 0x23000000
unit, update, player = heap+0x4000, heap+0x9000, heap+0x1000
sp, stop, initialize, cw = stack+0xe000, aux+0x100, aux+0x200, aux+0x300
saved = {UC_X86_REG_EBX:0x12345678, UC_X86_REG_ESI:0x23456789,
         UC_X86_REG_EDI:0x3456789a, UC_X86_REG_EBP:0x456789ab}
u, active, expected = None, False, bytearray()
seen, faults = set(), collections.Counter()
current, writes, fault, detail = {}, [], None, None
def write(address, value):
    data = struct.pack('<I', value&0xffffffff)
    u.mem_write(address, data)
    expected[address-heap:address-heap+4] = data
def on_code(_u, address, count, _):
    if active:
        assert address in instructions and bytes(u.mem_read(address, count)) == instructions[address]
        seen.add(address)
def on_write(_u, access, address, count, value, _):
    if not active:
        return
    pc = u.reg_read(UC_X86_REG_EIP)
    assert pc in (0x4d61d6, 0x4d61e2) and count == 4
    writes.append(dict(pc=hex(pc), address=hex(address), value=value&0xffffffff))
    if heap <= address and address+count <= heap+0x10000:
        expected[address-heap:address-heap+count] = (value&0xffffffff).to_bytes(count, 'little')
def on_fault(_u, access, address, count, value, _):
    global fault, detail
    pc = u.reg_read(UC_X86_REG_EIP)
    site = {0x4d61b8:'unit-flags', 0x4d61c3:'unit-update', 0x4d61c9:'update-player',
            0x4d61cf:'player-counter', 0x4d61d6:'player-counter',
            0x4d61dc:'update-player-reload', 0x4d61e2:'player-flags'}[pc]
    fault = dict(kind='memory', site=site)
    detail = dict(pc=hex(pc), address=hex(address), size=count, access=access)
    return False
outputs, details = [], []
for current in inputs:
    active = False
    u = Uc(UC_ARCH_X86, UC_MODE_32)
    u.mem_map(base, size)
    u.mem_write(base, original)
    for address in (heap, stack, aux):
        u.mem_map(address, 0x10000)
    expected = bytearray((number*17+(number>>7)*29)&255 for number in range(0x10000))
    u.mem_write(heap, bytes(expected))
    chosen_player = player
    alias = current['alias']
    if alias == 'counter-is-update-player':
        chosen_player = update+276-4668
    elif alias == 'counter-is-unit-update':
        chosen_player = unit+748-4668
    elif alias == 'counter-is-unit-flags':
        chosen_player = unit+16-4668
    elif alias == 'player-flags-is-update-player':
        chosen_player = update+276-4692
    elif current['player'] == 'unit':
        chosen_player = unit
    write(chosen_player+4668, current['counter'])
    write(chosen_player+4692, current['player_flags'])
    selected_player = {'normal':chosen_player, 'nil':0, 'unmapped':0x41000000, 'unit':unit}[current['player']]
    write(update+276, selected_player)
    selected_update = {'normal':update, 'nil':0, 'unmapped':0x41000000, 'unit':unit}[current['update']]
    write(unit+748, selected_update)
    write(unit+16, current['flags'])
    argument = {'normal':unit, 'nil':0, 'unmapped':0x41000000}[current['unit']]
    caller = struct.pack('<II', stop, argument)+bytes(range(48))
    u.mem_write(stack, b'\xcc'*0x10000)
    u.mem_write(sp, caller)
    u.mem_write(initialize, b'\xdb\xe3\xd9\x2d'+struct.pack('<I', cw))
    u.mem_write(cw, struct.pack('<H', options.control))
    aux_before = bytes(u.mem_read(aux, 0x10000))
    for register, value in saved.items():
        u.reg_write(register, value)
    u.reg_write(UC_X86_REG_ESP, sp)
    u.emu_start(initialize, initialize+8, count=2)
    for hook, callback in [(UC_HOOK_CODE, on_code), (UC_HOOK_MEM_WRITE, on_write), (UC_HOOK_MEM_INVALID, on_fault)]:
        u.hook_add(hook, callback)
    writes, fault, detail = [], None, None
    active = True
    try:
        u.emu_start(0x4d61b0, stop, count=1000)
    except UcError:
        assert fault is not None, current['name']
    active = False
    assert bytes(u.mem_read(heap, 0x10000)) == expected, (current['name'], 'whole heap')
    assert bytes(u.mem_read(base, size)) == original, (current['name'], 'whole image')
    assert bytes(u.mem_read(aux, 0x10000)) == aux_before
    assert bytes(u.mem_read(sp, len(caller))) == caller
    assert u.reg_read(UC_X86_REG_FPCW) == options.control
    if fault:
        faults[fault['site']] += 1
        details.append(dict(name=current['name'], **detail))
    else:
        assert u.reg_read(UC_X86_REG_EIP) == stop and u.reg_read(UC_X86_REG_ESP) == sp+4
        assert all(u.reg_read(register) == value for register, value in saved.items())
        assert u.reg_read(UC_X86_REG_FPSW)&0x3800 == 0
    outputs.append(dict(name=current['name'], fault=fault, original_writes=writes, heap_sha256=sha(expected)))
assert seen == set(instructions)
decoded = json.dumps(outputs, sort_keys=True, separators=(',', ':')).encode()+b'\n'
encoded_output = bytearray(gzip.compress(decoded, compresslevel=9, mtime=0))
encoded_output[9] = 0xff
with options.output.open('xb') as stream:
    stream.write(encoded_output)
proof = dict(original_executable_sha256=sha(raw), input_sha256=sha(encoded), cases=len(inputs),
             control=hex(options.control), body_address='0x004D61B0', body_bytes=58, body_sha256=sha(body),
             all58_original_instruction_bytes_executed=True, no_world_services_replaced=True,
             whole_heap_image_and_helper_guards=True, caller_bytes_unchanged=True,
             valid_return_stack_registers_and_FPU_TOP_preserved=True, FPU_control_preserved=True,
             faults=dict(faults), fault_details=details, declared_PE32_aliases=[row['name'] for row in inputs if row['alias'] != 'none'],
             general_native_named_field_alias_credit=False, native_source_used=False,
             output_sha256=sha(encoded_output), decoded_output_sha256=sha(decoded),
             actual_unchanged_C_fault_credit=False, native_comparison_credit=False)
with options.proof.open('x') as stream:
    json.dump(proof, stream, indent=2)
    stream.write('\n')
print(json.dumps(proof), flush=True)
