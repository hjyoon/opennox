"""Execute both complete original secret-wall packet bodies; no result substitution."""
from pathlib import Path
import argparse, datetime, gzip, hashlib, json, struct
import capstone
from unicorn import Uc, UcError, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_WRITE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import *

ap = argparse.ArgumentParser(description=__doc__)
ap.add_argument('--original', type=Path, required=True)
ap.add_argument('--control', type=lambda x: int(x, 0), choices=[0xe3f, 0x23f], required=True)
ap.add_argument('--output', type=Path, required=True)
ap.add_argument('--proof', type=Path, required=True)
args = ap.parse_args()
sha = lambda b: hashlib.sha256(b).hexdigest()
blob = args.original.read_bytes()
assert sha(blob) == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe = struct.unpack_from('<I', blob, 60)[0]
op = pe + 24
tab = op + struct.unpack_from('<H', blob, pe + 20)[0]
base = struct.unpack_from('<I', blob, op + 28)[0]
image_size = (struct.unpack_from('<I', blob, op + 56)[0] + 4095) & ~4095
u = Uc(UC_ARCH_X86, UC_MODE_32)
u.mem_map(base, image_size)
for i in range(struct.unpack_from('<H', blob, pe + 6)[0]):
    _, va, n, off = struct.unpack_from('<IIII', blob, tab + 40 * i + 8)
    if n:
        u.mem_write(base + va, blob[off:off+n])
image_before = bytes(u.mem_read(base, image_size))
heap, stack, aux = 0x21000000, 0x22000000, 0x23000000
for ptr in [heap, stack, aux]:
    u.mem_map(ptr, 0x10000)
wall = heap + 0x200
units = [heap + 0x1000 + i * 0x800 for i in range(4)]
updates = [heap + 0x4000 + i * 0x800 for i in range(4)]
players = [heap + 0x8000 + i * 0x2000 for i in range(4)]
sp, stop, init, cw = stack + 0xe000, aux + 0x100, aux + 0x200, aux + 0x300
u.mem_write(init, b'\xdb\xe3\xd9\x2d' + struct.pack('<I', cw))
u.mem_write(cw, struct.pack('<H', args.control))
md = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_32)
instructions, ranges = {}, []
for pc in [0x4df120, 0x4df180]:
    raw = bytes(u.mem_read(pc, 86))
    seq = list(md.disasm(raw, pc))
    assert seq[-1].mnemonic == 'ret' and sum(i.size for i in seq) == 86
    instructions.update({i.address: bytes(i.bytes) for i in seq})
    ranges.append(dict(entry=hex(pc), bytes=86, sha256=sha(raw)))
services = {0x4da7c0: 'first', 0x4da7f0: 'next', 0x4e5420: 'send'}
saved = {UC_X86_REG_EBX: 0x12345678, UC_X86_REG_ESI: 0x23456789,
         UC_X86_REG_EDI: 0x3456789a, UC_X86_REG_EBP: 0x456789ab}

def case(entry, name, route, **kw):
    out = dict(name=f'{entry}/{name}', entry=entry, wall_id=0x1234, seed=0xa5a5a5a5,
               wall_kind=0, route=route, updates=[0, 1, 2, 3], players=[0, 1, 2, 3],
               indices=[0, 31, 128, 255], classes=[4, 0, 1, 0xffffffff], edits=[])
    out.update(kw)
    return out

patterns = [([], 'empty', {}), ([0], 'one', {}), ([0, 1], 'two', {}),
            ([0, 1, 2, 3], 'all', {}), ([3, 1, 0, 2], 'reordered', {}),
            ([0, 0, 2, 0], 'repeated-unit', {}),
            ([0, 1, 2, 3], 'shared-update', dict(updates=[0, 0, 0, 0])),
            ([0, 1, 2, 3], 'shared-player', dict(players=[2, 2, 2, 2])),
            ([0, 1, 2], 'first-changes-wall', dict(edits=[dict(after='first', field='wall', slot=0, value=0xbeef)])),
            ([0, 1, 2], 'live-next-recipient', dict(edits=[dict(after='send0', field='index', slot=1, value=203),
                 dict(after='next0', field='player', slot=2, value=1)])),
            ([0, 1, 2], 'live-update-alias', dict(edits=[dict(after='send0', field='update', slot=1, value=3),
                 dict(after='next0', field='index', slot=3, value=17)])),
            ([0, 1, 2], 'packet-survives-send', dict(edits=[dict(after='send0', field='packet', slot=0, value=0xccaabb),
                 dict(after='send1', field='wall', slot=0, value=0x5566)]))]
rows = []
for entry in ['4DF120', '4DF180']:
    for identity in [0, 1, 255, 256, 0x1234, 0x7fff, 0x8000, 0xff00, 0xffff]:
        for route, name, kw in patterns:
            for seed in [0, 0xa5a5a5a5, 0xffffffff]:
                rows.append(case(entry, f'{name}/id{identity:04x}/seed{seed:08x}', route,
                                 wall_id=identity, seed=seed, **kw))
    for index in range(256):
        rows.append(case(entry, f'unsigned-recipient/{index:03d}', [0], indices=[index, 0, 0, 0]))
    for kind in [-1, -2]:
        rows.append(case(entry, f'fault-wall/{kind}', [], wall_kind=kind))
        rows.append(case(entry, f'fault-first-unit/{kind}', [kind]))
        for at in [0, 1]:
            for field in ['updates', 'players']:
                values = [0, 1, 2, 3]
                values[at] = kind
                rows.append(case(entry, f'fault-{field}/{at}/{kind}', [0, 1], **{field: values}))
    for field in ['update', 'player']:
        rows.append(case(entry, f'fault-live-{field}', [0, 1],
                         edits=[dict(after='send0', field=field, slot=1, value=-1)]))

active = False
seen_all = set()
expected = bytearray()
events, invalid, packet_ptr, position, send_count = [], [], None, 0, 0

def pointer(values, index):
    return 0 if index == -1 else 0x41000000 if index == -2 else values[index]

def write(ptr, value, n=4):
    raw = int(value & ((1 << (8*n))-1)).to_bytes(n, 'little')
    u.mem_write(ptr, raw)
    if heap <= ptr < heap + 0x10000:
        expected[ptr-heap:ptr-heap+n] = raw

def edits(after):
    for edit in row['edits']:
        if edit['after'] != after:
            continue
        field, slot, value = edit['field'], edit['slot'], edit['value']
        if field == 'wall': write(wall+10, value, 2)
        elif field == 'update': write(units[slot]+748, pointer(updates, value))
        elif field == 'player': write(updates[slot]+276, pointer(players, value))
        elif field == 'index': write(players[slot]+2064, value, 1)
        elif field == 'packet':
            assert packet_ptr is not None
            write(packet_ptr, value, 3)
        else: raise AssertionError(field)

def read32(ptr):
    return struct.unpack('<I', u.mem_read(ptr, 4))[0]

def on_code(_u, pc, n, _):
    global position, send_count, packet_ptr
    if not active:
        return
    if pc not in services:
        assert pc in instructions and bytes(u.mem_read(pc, n)) == instructions[pc], (row['name'], hex(pc))
        seen_all.add(pc)
        return
    service = services[pc]
    esp = u.reg_read(UC_X86_REG_ESP)
    if service == 'first':
        result = row['route'][0] if row['route'] else -1
        events.append(dict(kind='first', unit=result))
        edits('first')
        result = pointer(units, result)
    elif service == 'next':
        arg = read32(esp+4)
        assert arg == pointer(units, row['route'][position])
        current = row['route'][position]
        position += 1
        result = row['route'][position] if position < len(row['route']) else -1
        events.append(dict(kind='next', unit=current, next=result))
        edits('next'+str(position-1))
        result = pointer(units, result)
    else:
        recipient, payload, size, related, remove = struct.unpack('<IIIII', u.mem_read(esp+4, 20))
        assert size == 3 and related == 0 and remove == 1
        if packet_ptr is None: packet_ptr = payload
        assert packet_ptr == payload == sp-4
        events.append(dict(kind='send', recipient=recipient, packet=bytes(u.mem_read(payload, size)).hex(),
                           size=size, related=related, remove=remove))
        edits('send'+str(send_count))
        send_count += 1
        result = (row['seed'] ^ (send_count * 0x12345678)) & 0xffffffff
    u.reg_write(UC_X86_REG_EAX, result)
    u.reg_write(UC_X86_REG_ECX, 0xcdef1234)
    u.reg_write(UC_X86_REG_EDX, 0x9876abcd)
    u.reg_write(UC_X86_REG_ESP, esp+4)
    u.reg_write(UC_X86_REG_EIP, read32(esp))

def on_write(_u, _access, ptr, n, value, _):
    if active:
        assert sp-64 <= ptr < ptr+n <= sp, (row['name'], hex(ptr), n, hex(u.reg_read(UC_X86_REG_EIP)))

def on_invalid(_u, access, ptr, n, _value, _):
    invalid.append(dict(access=access, address=hex(ptr), bytes=n, pc=hex(u.reg_read(UC_X86_REG_EIP))))
    return False

u.hook_add(UC_HOOK_CODE, on_code)
u.hook_add(UC_HOOK_MEM_WRITE, on_write)
u.hook_add(UC_HOOK_MEM_INVALID, on_invalid)
outputs, faults = [], 0
for row in rows:
    active = False
    expected = bytearray(struct.pack('<I', row['seed']) * (0x10000//4))
    u.mem_write(heap, bytes(expected))
    u.mem_write(stack, b'\x5a'*0x10000)
    write(wall+10, row['wall_id'], 2)
    for i in range(4):
        write(units[i]+8, row['classes'][i])
        write(units[i]+748, pointer(updates, row['updates'][i]))
        write(updates[i]+276, pointer(players, row['players'][i]))
        write(players[i]+2064, row['indices'][i], 1)
    u.mem_write(sp, struct.pack('<III', stop, pointer([wall], row['wall_kind']), 0x12345678))
    caller = bytes(u.mem_read(stack, 0x10000))
    u.emu_start(init, init+8, count=2)
    for reg, value in saved.items(): u.reg_write(reg, value)
    u.reg_write(UC_X86_REG_ESP, sp)
    u.reg_write(UC_X86_REG_EFLAGS, 0x202)
    events, invalid, packet_ptr, position, send_count = [], [], None, 0, 0
    error = None
    active = True
    try:
        u.emu_start(int(row['entry'], 16), stop, count=4096)
    except UcError as exc:
        assert invalid, (row['name'], str(exc))
        error = str(exc)
    finally:
        active = False
    assert bytes(u.mem_read(heap, 0x10000)) == bytes(expected), row['name']
    after = bytes(u.mem_read(stack, 0x10000))
    off = sp-stack
    assert caller[:off-64] == after[:off-64] and caller[off:] == after[off:]
    assert u.reg_read(UC_X86_REG_FPCW) == args.control and u.reg_read(UC_X86_REG_FPSW) & 0x3800 == 0
    if error is None:
        assert not invalid and u.reg_read(UC_X86_REG_ESP) == sp+4 and u.reg_read(UC_X86_REG_EIP) == stop
        assert u.reg_read(UC_X86_REG_EAX) == 0
        assert all(u.reg_read(reg) == value for reg, value in saved.items())
    else:
        faults += 1
    outputs.append(dict(input=row, output=dict(events=events, fault=invalid, error=error,
                        return_value=0 if error is None else None, whole_heap_and_guarded_caller_preserved=True)))
assert seen_all == set(instructions)
assert bytes(u.mem_read(base, image_size)) == image_before
raw = json.dumps(dict(schema=1, entries=['4DF120', '4DF180'], rows=outputs), separators=(',', ':')).encode()
packed = gzip.compress(raw, mtime=0)
proof = dict(utc=datetime.datetime.now(datetime.timezone.utc).isoformat(), original_GAME_EXE_sha256=sha(blob),
             control=hex(args.control), rows=len(rows), raw_faults=faults, instruction_bytes=172,
             alignment_bytes_excluded=20, original_ranges=ranges, all_original_instructions_executed=True,
             original_instruction_or_result_substitutions=0, service_boundaries=services,
             whole_original_image_heap_guarded_caller_nonvolatile_and_FPU_guards_passed=True,
             gzip_sha256=sha(packed), decoded_sha256=sha(raw))
assert not args.output.exists() and not args.proof.exists()
args.output.write_bytes(packed)
args.proof.write_text(json.dumps(proof, indent=2)+'\n')
print(json.dumps(proof), flush=True)
