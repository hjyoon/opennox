"""Execute complete unchanged original Readiness4E0960 on independent inputs.

This root invokes no helpers. It compares the 004E0950 callback identity and
reloads the successful modifier slot before reading its DWORD strength.
No native implementation, original oracle manifest, or source is modified.
"""
from pathlib import Path
import argparse, collections, gc, gzip, hashlib, json, struct
import capstone
from unicorn import (Uc, UcError, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE,
                     UC_HOOK_MEM_READ, UC_HOOK_MEM_WRITE, UC_HOOK_MEM_INVALID,
                     UC_HOOK_INTR)
from unicorn.x86_const import *

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--exe', type=Path, required=True)
parser.add_argument('--control', type=lambda v: int(v, 0), choices=[0xe3f, 0x23f], required=True)
parser.add_argument('--inputs', type=Path, required=True)
parser.add_argument('--output', type=Path, required=True)
parser.add_argument('--proof', type=Path, required=True)
options = parser.parse_args()
sha = lambda value: hashlib.sha256(value).hexdigest()
encoded = options.inputs.read_bytes()
assert sha(encoded) == 'cb6330a5622ee63f15cc64b2979d3521fac88df4b3184651517bdc2e88987d4f'
decoded_inputs = gzip.decompress(encoded)
assert sha(decoded_inputs) == 'daca8b8c2e4e0ed897a33394dba93657b15cfbfa78dc0bc3ffb3f09f22e8413f'
cases = json.loads(decoded_inputs)
assert len(cases) == 113
raw = options.exe.read_bytes()
assert sha(raw) == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe = struct.unpack_from('<I', raw, 60)[0]
opt = pe+24
base = struct.unpack_from('<I', raw, opt+28)[0]
image = bytearray((struct.unpack_from('<I', raw, opt+56)[0]+4095) & ~4095)
table_offset = opt+struct.unpack_from('<H', raw, pe+20)[0]
for number in range(struct.unpack_from('<H', raw, pe+6)[0]):
    _, rva, length, disk = struct.unpack_from('<IIII', raw, table_offset+40*number+8)
    image[rva:rva+length] = raw[disk:disk+length]
image = bytes(image)
root, extent = 0x4e0960, 68
body = image[root-base:root-base+extent]
assert sha(body) == 'dd2e32642e9c728d4d6c6efb83957760d70c9cc4d7054fbc042d32e3db9319a8'
assert image[0x4e0950-base:0x4e0951-base] == b'\xc3'
assert image[0x4e09a4-base:0x4e09b0-base] == b'\x90'*12
md = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_32)
code_items = list(md.disasm(body, root))
assert sum(item.size for item in code_items) == extent and code_items[-1].mnemonic == 'ret'
instructions = {item.address: (bytes(item.bytes), item.mnemonic) for item in code_items}
assert len(instructions) == 25

heap, stack, aux, bad = 0x21000000, 0x22000000, 0x23000000, 0x41000000
sp, stop, initialize, cw = stack+0xe000, aux+0x100, aux+0x200, aux+0x300
init_bytes = b'\xdb\xe3\xd9\x2d'+struct.pack('<I', cw)
saved = {UC_X86_REG_EBX: 0x12345678, UC_X86_REG_ESI: 0x23456789,
         UC_X86_REG_EDI: 0x3456789a, UC_X86_REG_EBP: 0x456789ab}
locations = {0x4e0965: ('item', 'init-data', 692), 0x4e096f: ('item', 'class', 8),
             0x4e0980: ('table', 'slot', None), 0x4e0986: ('modifier', 'callback', 40),
             0x4e099c: ('table', 'slot', None), 0x4e09a0: ('modifier', 'strength', 48)}
active = False
u = None
seen = set()
fault_domains, accesses_by_field = collections.Counter(), collections.Counter()
accesses, writes, calls, returns, pending, applied = [], [], [], [], [], []
read_counts = collections.Counter()
fault = None
roles, current = {}, {}


def target_field(role, field, index=None):
    if role == 'item':
        return roles[role]+{'init-data': 692, 'class': 8}[field]
    if role in ('table', 'alternate_table'):
        assert field == 'slot' and index in range(4)
        return roles[role]+4*index
    assert role in ('first', 'second', 'replacement')
    return roles[role]+{'callback': 40, 'strength': 48}[field]


def write_stimulus(index):
    assert index not in applied
    mutation = current['mutations'][index]
    role, field = mutation['target_role'], mutation['target_field']
    address = target_field(role, field, mutation.get('target_index'))
    value = mutation['value']
    if isinstance(value, str):
        value = 0 if value == 'null' else bad if value == 'bad' else roles[value]
    value &= 0xffffffff
    assert heap <= address and address+4 <= heap+0x10000
    u.mem_write(address, struct.pack('<I', value))
    record = dict(kind='stimulus', role=role, field=field, address=hex(address),
                  size=4, value=value, complete=True, source_input_mutation=index,
                  instruction=hex(u.reg_read(UC_X86_REG_EIP)))
    if field == 'slot':
        record['index'] = mutation['target_index']
    accesses.append(record)
    writes.append(record.copy())
    applied.append(index)


def code(uc, address, length, unused):
    if not active:
        return
    while pending:
        write_stimulus(pending.pop(0))
    assert address in instructions and length == len(instructions[address][0]), hex(address)
    assert bytes(uc.mem_read(address, length)) == instructions[address][0]
    seen.add(address)
    if address == root:
        esp = uc.reg_read(UC_X86_REG_ESP)
        args = list(struct.unpack('<I', uc.mem_read(esp+4, 4)))
        assert args == [roles['item']]
        calls.append(dict(function='readiness', address=hex(root), args=args, depth=0, external=False))
    if instructions[address][1] == 'ret':
        returns.append(dict(function='readiness', result=uc.reg_read(UC_X86_REG_EAX), depth=0, external=False))


def describe(uc, address, length):
    instruction = uc.reg_read(UC_X86_REG_EIP)
    assert instruction in locations and length == 4, (hex(instruction), address, length)
    role, field, offset = locations[instruction]
    record = dict(kind='read', role=role, field=field, address=hex(address), size=length,
                  instruction=hex(instruction))
    if field == 'slot':
        index = uc.reg_read(UC_X86_REG_EDX)
        assert index in (2, 3)
        record['index'] = index
        record['base'] = hex((address-index*4) & 0xffffffff)
    else:
        record['base'] = hex((address-offset) & 0xffffffff)
    return record


def read(uc, access, address, length, unused_value, unused):
    if not active or stack <= address < stack+0x10000:
        return
    record = describe(uc, address, length)
    try:
        value = int.from_bytes(uc.mem_read(address, length), 'little')
    except UcError:
        value = None
    record.update(value=value, complete=value is not None)
    accesses.append(record)
    key = record['role'], record['field'], record.get('index')
    read_counts[key] += 1
    if value is not None:
        for index, mutation in enumerate(current['mutations']):
            if (mutation['when'] == 'after-read' and mutation['role'] == key[0]
                    and mutation['field'] == key[1]
                    and mutation.get('index') == key[2]
                    and mutation['occurrence'] == read_counts[key]):
                pending.append(index)


def write(uc, access, address, length, value, unused):
    if not active:
        return
    assert stack <= address and address+length <= stack+0x10000 and length == 4
    writes.append(dict(kind='write', address=hex(address), size=length, value=value & 0xffffffff,
                       instruction=hex(uc.reg_read(UC_X86_REG_EIP))))


def invalid(uc, access, address, length, value, unused):
    global fault
    record = describe(uc, address, length)
    fault = dict(kind='memory', access=access, address=hex(address), size=length,
                 instruction=record['instruction'])
    assert access in (19, 23), fault
    if accesses and accesses[-1]['kind'] == 'read' and accesses[-1]['address'] == hex(address):
        accesses[-1].update(complete=False, value=None)
    else:
        record.update(complete=False, value=None)
        accesses.append(record)
    return False


def interrupt(uc, number, unused):
    raise AssertionError(('Unexpected Readiness interrupt', number, hex(uc.reg_read(UC_X86_REG_EIP))))


def address_for(role, shape):
    if shape == 'null':
        return 0
    if shape == 'bad':
        return bad
    offsets = {'item': 0x3000, 'table': 0x5000, 'first': 0x7000, 'second': 0x9000,
               'replacement': 0xb000, 'alternate_table': 0xd000}
    edges = {'last-init': 696, 'last-class': 12, 'last-slot2': 12, 'last-slot3': 16,
             'last-callback': 44, 'last-strength': 52}
    if shape in edges:
        return heap+65536-edges[shape]
    assert shape == 'valid'
    return heap+offsets[role]


results = []
for current in cases:
    roles = {role: address_for(role, current.get(role, 'valid'))
             for role in ('item', 'table', 'first', 'second', 'replacement', 'alternate_table')}
    initial = bytearray(((i*37+current['seed']*19) ^ (i >> 3)) & 255 for i in range(65536))
    initializers = [('item', 'init-data', None, roles['table']), ('item', 'class', None, current['obj_class'])]
    for index, value in enumerate(current['slots']):
        initializers.append(('table', 'slot', index, 0 if value == 'null' else bad if value == 'bad' else roles[value]))
    for index, value in enumerate([bad, bad, roles['replacement'], 0]):
        initializers.append(('alternate_table', 'slot', index, value))
    for role in ('first', 'second', 'replacement'):
        initializers.extend([(role, 'callback', None, current['callbacks'][role]),
                             (role, 'strength', None, current['strengths'][role])])
    for role, field, index, value in initializers:
        address = target_field(role, field, index)
        if heap <= address and address+4 <= heap+65536:
            initial[address-heap:address-heap+4] = struct.pack('<I', value)

    u = Uc(UC_ARCH_X86, UC_MODE_32)
    for address, length in [(base, len(image)), (heap, 65536), (stack, 65536), (aux, 4096)]:
        u.mem_map(address, length)
    u.mem_write(base, image)
    u.mem_write(heap, bytes(initial))
    initial_stack = bytearray(((i*11+current['seed']) ^ (i >> 4)) & 255 for i in range(65536))
    caller = struct.pack('<2I', stop, roles['item'])
    initial_stack[sp-stack:sp-stack+len(caller)] = caller
    u.mem_write(stack, bytes(initial_stack))
    u.mem_write(initialize, init_bytes)
    u.mem_write(cw, struct.pack('<H', options.control))
    u.hook_add(UC_HOOK_CODE, code)
    u.hook_add(UC_HOOK_MEM_READ, read)
    u.hook_add(UC_HOOK_MEM_WRITE, write)
    u.hook_add(UC_HOOK_MEM_INVALID, invalid)
    u.hook_add(UC_HOOK_INTR, interrupt)
    u.reg_write(UC_X86_REG_ESP, sp)
    u.emu_start(initialize, initialize+len(init_bytes), count=8)
    for register, value in saved.items():
        u.reg_write(register, value)
    u.reg_write(UC_X86_REG_EAX, 0x98765432)
    u.reg_write(UC_X86_REG_ECX, 0x87654321)
    u.reg_write(UC_X86_REG_EDX, 0x76543210)
    u.reg_write(UC_X86_REG_EFLAGS, 0x202)
    accesses, writes, calls, returns, pending, applied = [], [], [], [], [], []
    read_counts = collections.Counter()
    fault = None
    active = True
    try:
        u.emu_start(root, stop, count=256)
    except UcError:
        assert fault is not None
    finally:
        active = False
    assert not pending and len(calls) == 1
    final_heap = bytes(u.mem_read(heap, 65536))
    final_stack = bytes(u.mem_read(stack, 65536))
    expected_heap, expected_stack = bytearray(initial), bytearray(initial_stack)
    for record in writes:
        address = int(record['address'], 16)
        value = struct.pack('<I', record['value'])
        if record['kind'] == 'stimulus':
            expected_heap[address-heap:address-heap+4] = value
        else:
            expected_stack[address-stack:address-stack+4] = value
    assert final_heap == bytes(expected_heap) and final_stack == bytes(expected_stack)
    assert bytes(u.mem_read(base, len(image))) == image
    assert bytes(u.mem_read(sp, len(caller))) == caller
    assert u.reg_read(UC_X86_REG_FPCW) == options.control and u.reg_read(UC_X86_REG_FPSW) == 0
    if fault is None:
        assert len(returns) == 1 and u.reg_read(UC_X86_REG_EIP) == stop and u.reg_read(UC_X86_REG_ESP) == sp+4
        assert all(u.reg_read(register) == value for register, value in saved.items())
    else:
        assert not returns and u.reg_read(UC_X86_REG_EIP) == int(fault['instruction'], 16)
        fault_domains[fault['instruction']] += 1
    for access in accesses:
        if access['kind'] == 'read':
            accesses_by_field[access['role']+'/'+access['field']] += 1
    results.append(dict(name=current['name'], roles=roles, initial_heap=bytes(initial).hex(),
                        final_heap=final_heap.hex(), initial_stack=bytes(initial_stack).hex(),
                        final_stack=final_stack.hex(), caller_before=caller.hex(), caller_after=caller.hex(),
                        initial_image_sha256=sha(image), final_image_sha256=sha(bytes(u.mem_read(base, len(image)))),
                        registers=dict(eax=u.reg_read(UC_X86_REG_EAX), ecx=u.reg_read(UC_X86_REG_ECX),
                                       edx=u.reg_read(UC_X86_REG_EDX), ebx=u.reg_read(UC_X86_REG_EBX),
                                       esi=u.reg_read(UC_X86_REG_ESI), edi=u.reg_read(UC_X86_REG_EDI),
                                       ebp=u.reg_read(UC_X86_REG_EBP), esp=u.reg_read(UC_X86_REG_ESP),
                                       eip=u.reg_read(UC_X86_REG_EIP), eflags=u.reg_read(UC_X86_REG_EFLAGS)),
                        accesses=accesses, writes=writes, calls=calls, returns=returns,
                        applied_input_mutations=applied, fault=fault))
    u = None
    gc.collect()

assert seen == set(instructions), [hex(address) for address in sorted(set(instructions)-seen)]
assert sha(options.exe.read_bytes()) == sha(raw)
decoded_output = json.dumps(results, sort_keys=True, separators=(',', ':')).encode()+b'\n'
encoded_output = gzip.compress(decoded_output, mtime=0)
with options.output.open('xb') as stream:
    stream.write(encoded_output)
proof = dict(exe_sha256=sha(raw), complete_original_root_address=hex(root), complete_original_root_bytes=extent,
             complete_original_root_sha256=sha(body), instructions=len(instructions), all25original_instructions_reached=True,
             cases=len(cases), FPU_control=hex(options.control), FPU_status=0, all_full_heaps_stacks_callers_images_checked=True,
             helper_calls_executed=0, callback_identity_compared_only='0x004e0950', fault_domains=dict(fault_domains),
             faults=sum(fault_domains.values()), read_counts=dict(accesses_by_field),
             read_events=sum(accesses_by_field.values()), stimuli=sum(len(row['applied_input_mutations']) for row in results),
             input_encoded_sha256=sha(encoded), input_decoded_sha256=sha(decoded_inputs),
             output_encoded_sha256=sha(encoded_output), output_decoded_sha256=sha(decoded_output),
             producer_sha256=sha(Path(__file__).read_bytes()), native_implementation_credit=False,
             repository_or_original_manifest_modified=False)
with options.proof.open('x') as stream:
    json.dump(proof, stream, indent=2)
    stream.write('\n')
print(json.dumps(proof), flush=True)
