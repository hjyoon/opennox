"""Execute unchanged HurtSound532800 and SoundSet424300 with declared RNG/audio.

Frame/FPS loads are read-only observations. RNG may mutate declared live
inputs; cached update writes and the actual sound lookup remain original.
The compressed reference is not a native implementation or whole MainAI.
"""
import argparse, gzip, hashlib, itertools, json, struct
from pathlib import Path
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_READ, UC_HOOK_MEM_WRITE
from unicorn.x86_const import *

p = argparse.ArgumentParser(description=__doc__)
for name in ['game', 'fixture', 'proof']:
    p.add_argument('--'+name, type=Path, required=True)
p.add_argument('--control', type=lambda s: int(s, 0), choices=[0xe3f, 0x23f], required=True)
a = p.parse_args()
blob = a.game.read_bytes()
sha = lambda raw: hashlib.sha256(raw).hexdigest()
assert sha(blob) == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe = struct.unpack_from('<I', blob, 60)[0]; op = pe+24
tb = op+struct.unpack_from('<H', blob, pe+20)[0]
base = struct.unpack_from('<I', blob, op+28)[0]
u = Uc(UC_ARCH_X86, UC_MODE_32)
u.mem_map(base, (struct.unpack_from('<I', blob, op+56)[0]+4095)&~4095)
for i in range(struct.unpack_from('<H', blob, pe+6)[0]):
    _, va, n, off = struct.unpack_from('<IIII', blob, tb+i*40+8)
    if n:
        u.mem_write(base+va, blob[off:off+n])
heap, stack, code = 0x21000000, 0x22000000, 0x23000000
for address in [heap, stack, code]:
    u.mem_map(address, 0x10000)
unit, updates, sounds = heap+0x1000, [heap+0x3000, heap+0x5000], [heap+0x8000, heap+0xa000]
sp, stop, init, cw = stack+0xe000, code+0x100, code+0x200, code+0x300
frame_address, fps_address = 0x84ea04, 0x85b3fc
rd = lambda address: struct.unpack('<I', u.mem_read(address, 4))[0]
wr = lambda address, value: u.mem_write(address, struct.pack('<I', value&0xffffffff))
ptrref = lambda pointer, values: values.index(pointer) if pointer else -1
u.mem_write(init, b'\xdb\xe3\xd9\x2d'+struct.pack('<I', cw))
ranges = {0x532800: 113, 0x424300: 31}
original = {pc: bytes(u.mem_read(pc, n)) for pc, n in ranges.items()}
assert sha(original[0x532800]) == '444eab1f404602b062927196eb13e58788b50a21ad981ad6affc35a1d9c3b4ef'
seed = bytes((i*31+19)&255 for i in range(0x10000))
registers = {UC_X86_REG_EBX: 0x12345678, UC_X86_REG_ESI: 0x23456789, UC_X86_REG_EDI: 0x3456789a, UC_X86_REG_EBP: 0x456789ab}
active = False
current, events, writes, traces, raw_random_args = {}, [], [], {}, []
service_counts = {'random': 0, 'audio': 0}

def state():
    return dict(cls=rd(unit+8), frame=rd(frame_address), fps=rd(fps_address),
                update=updates.index(rd(unit+748)), deadlines=[rd(v+532) for v in updates],
                sound_sets=[ptrref(rd(v+488), sounds) for v in updates], sound_ids=[rd(v+8) for v in sounds])

def mutate():
    mode = current['mutation']
    cached = current['update']
    live = updates.index(rd(unit+748))
    if mode in ['frame', 'frame-update']:
        wr(frame_address, current['frame'] ^ 0x80000011)
    if mode == 'fps':
        wr(fps_address, current['fps'] ^ 0x80000007)
    if mode in ['update', 'frame-update', 'live-set', 'clear-live']:
        live = 1-cached
        wr(unit+748, updates[live])
    if mode in ['cached-set', 'live-set']:
        owner = cached if mode == 'cached-set' else live
        old = ptrref(rd(updates[owner]+488), sounds)
        wr(updates[owner]+488, sounds[1 if old == 0 else 0])
    if mode in ['clear-cached', 'clear-live']:
        wr(updates[cached if mode == 'clear-cached' else live]+488, 0)
    if mode == 'class-remove':
        wr(unit+8, rd(unit+8)&~2)
    if mode == 'deadlines':
        wr(updates[0]+532, 0x11223344)
        wr(updates[1]+532, 0xaabbccdd)
    if mode == 'ids':
        wr(sounds[0]+8, 0)
        wr(sounds[1]+8, 777)

def hook(_u, pc, size, _):
    if not active:
        return
    if pc == stop:
        u.emu_stop(); return
    if any(start <= pc < start+n for start, n in ranges.items()):
        traces[pc] = bytes(u.mem_read(pc, size)); return
    assert pc in [0x415fa0, 0x501960], (current['name'], hex(pc))
    esp = u.reg_read(UC_X86_REG_ESP)
    if pc == 0x415fa0:
        argv = [rd(esp+4+i*4) for i in range(4)]
        assert argv[2:] == [0x5c6e18, 743]
        raw_random_args.append(argv)
        events.append(dict(kind='random', args=argv[:2], state=state()))
        service_counts['random'] += 1
        mutate()
        value = current['random_return']
    else:
        assert rd(esp+8) == unit and rd(esp+12) == rd(esp+16) == 0
        events.append(dict(kind='audio', args=[rd(esp+4), 0, 0, 0], state=state()))
        service_counts['audio'] += 1
        value = 0x87654321
    u.reg_write(UC_X86_REG_EAX, value)
    u.reg_write(UC_X86_REG_ECX, 0xc1c2c3c4)
    u.reg_write(UC_X86_REG_EDX, 0xd1d2d3d4)
    u.reg_write(UC_X86_REG_EIP, rd(esp))
    u.reg_write(UC_X86_REG_ESP, esp+4)

def read_hook(_u, access, address, size, value, _):
    if active and address in [frame_address, fps_address]:
        assert size == 4
        events.append(dict(kind='frame' if address == frame_address else 'fps', args=[rd(address)], state=state()))

def write_hook(_u, access, address, size, value, _):
    if active and heap <= address < heap+len(seed):
        assert address == updates[current['update']]+532 and size == 4
        writes.append(dict(update=current['update'], field='deadline', value=value&0xffffffff))

u.hook_add(UC_HOOK_CODE, hook)
u.hook_add(UC_HOOK_MEM_READ, read_hook)
u.hook_add(UC_HOOK_MEM_WRITE, write_hook)
allowed = set()
for address in [unit+8, unit+748]+[v+532 for v in updates]+[v+488 for v in updates]+[v+8 for v in sounds]:
    allowed.update(range(address-heap, address-heap+4))

def run(c, nil_getter=False):
    global active, current, events, writes, raw_random_args
    active = False; current = c; events = []; writes = []; raw_random_args = []
    u.mem_write(heap, seed)
    wr(unit+8, c['cls']); wr(unit+748, updates[c['update']])
    for i, v in enumerate(updates):
        wr(v+532, c['deadlines'][i])
        ref = c['sound_sets'][i]
        wr(v+488, sounds[ref] if ref >= 0 else 0)
    for v, value in zip(sounds, c['sound_ids']):
        wr(v+8, value)
    wr(frame_address, c['frame']); wr(fps_address, c['fps'])
    before = bytes(u.mem_read(heap, len(seed)))
    wr(cw, a.control); u.emu_start(init, init+8, count=2)
    u.mem_write(sp-512, bytes(1024)); wr(sp, stop); wr(sp+4, 0 if nil_getter else unit)
    for reg, value in registers.items():
        u.reg_write(reg, value)
    u.reg_write(UC_X86_REG_EAX, 0xa5a5a5a5)
    u.reg_write(UC_X86_REG_ESP, sp); u.reg_write(UC_X86_REG_EFLAGS, 0x202)
    active = True
    u.emu_start(0x424300 if nil_getter else 0x532800, stop+1, count=1000)
    active = False
    assert u.reg_read(UC_X86_REG_EIP) == stop and u.reg_read(UC_X86_REG_ESP) == sp+4, c['name']
    assert all(u.reg_read(reg) == value for reg, value in registers.items()), c['name']
    assert u.reg_read(UC_X86_REG_FPCW) == a.control and u.reg_read(UC_X86_REG_FPSW)&0x3800 == 0
    assert bytes(u.mem_read(sp-512, 256)) == bytes(256) and bytes(u.mem_read(sp+8, 256)) == bytes(256)
    after = bytes(u.mem_read(heap, len(seed)))
    assert all(i in allowed or before[i] == after[i] for i in range(len(seed))), c['name']
    eax = u.reg_read(UC_X86_REG_EAX)
    return dict(input=c, result=dict(events=list(events), state=state()), original_writes=list(writes), raw_declared_RNG_arguments=list(raw_random_args), raw_original_EAX=eax, raw_original_AL=eax&255)

cases = []
def case(prefix, **kw):
    c = dict(cls=2, frame=100, fps=30, update=0, deadlines=[0, 0x81234567], sound_sets=[0, 1], sound_ids=[123, 987], random_return=73, mutation='none')
    c.update(kw, name=prefix+'-'+str(len(cases)))
    cases.append(c)
edge = [0, 1, 31, 99, 0x7fffffff, 0x80000000, 0xfffffffe, 0xffffffff]
for cls, frame, deadline, fps in itertools.product([0, 2, 4, 6, 0x1000002], edge, edge, [0, 1, 30, 0x1fffffff, 0x40000000, 0x7fffffff, 0x80000000, 0xffffffff]):
    update = frame&1
    deadlines = [deadline, deadline^0x80000000]
    if update:
        deadlines.reverse()
    sets = [-1, 1] if deadline&3 == 0 else [deadline&1, 1-(deadline&1)]
    if update:
        sets.reverse()
    case('unsigned-cooldown-and-RNG-bounds', cls=cls, frame=frame, fps=fps, update=update, deadlines=deadlines, sound_sets=sets)
modes = ['none', 'frame', 'fps', 'update', 'cached-set', 'live-set', 'clear-cached', 'clear-live', 'class-remove', 'frame-update', 'deadlines', 'ids']
for update, ref, mode, fps, result in itertools.product([0, 1], [-1, 0, 1], modes, [0, 30, 0x40000001, 0xffffffff], [0, 1, 0x7fffffff, 0x80000000, 0xffffffff]):
    sets = [ref, 1 if ref < 0 else 1-ref]
    if update:
        sets.reverse()
    case('declared-RNG-live-input-mutation', update=update, frame=0xfffffffe, fps=fps, sound_sets=sets, deadlines=[0, 0], mutation=mode, random_return=result)
assert len(cases) == 4000
rows = [run(c) for c in cases]
ancillary = run(dict(cases[0], name='nil-sound-getter-ancillary'), nil_getter=True)
assert ancillary['raw_original_EAX'] == 0 and not ancillary['result']['events']
visited = {pc+i for pc, raw in traces.items() for i in range(len(raw))}
expected = {pc+i for pc, n in ranges.items() for i in range(n)}
assert visited == expected, [hex(pc) for pc in sorted(expected-visited)]
assert all(bytes(u.mem_read(pc, len(raw))) == raw for pc, raw in original.items())
fixture = dict(original_sha256=sha(blob), control_word=a.control, rows=rows)
proof = dict(original_sha256=sha(blob), control_word=a.control, primary_contexts=len(rows), ancillary_contexts=1,
    root_bytes={hex(pc): n for pc, n in ranges.items()}, root_sha256={hex(pc): sha(raw) for pc, raw in original.items()},
    executed_instruction_bytes=len(visited), declared_services=['RandomInt415FA0', 'AudioEvent501960'],
    actual_sound_getter_body_executes=True, read_only_frame_FPS_observations=True, service_counts=service_counts,
    original_code_unmodified=True, register_stack_FPCW_TOP_and_heap_canaries_preserved=True,
    whole_heap_bytes_outside_declared_mutations_and_cached_deadline_unchanged=True,
    original_deadline_write_history_raw_EAX_AL_and_RNG_debug_arguments_retained=True,
    unexecuted_Nop_tails=[15, 1], ancillary_nil_getter=ancillary,
    scope='Declared RNG outcomes and twelve input-mutation compositions; actual unchanged root and sound getter. Native comparator, consumer ABI and complete reference closure remain pending.',
    instructions=[dict(pc=hex(pc), bytes=raw.hex()) for pc, raw in sorted(traces.items())])
for path, value in [(a.fixture, fixture), (a.proof, proof)]:
    with path.open('xb') as f:
        if path.suffix == '.gz':
            with gzip.GzipFile(filename='', mode='wb', fileobj=f, compresslevel=9, mtime=0) as stream:
                stream.write(json.dumps(value, separators=(',', ':')).encode()+b'\n')
        else:
            f.write((json.dumps(value, indent=2)+'\n').encode())
print(json.dumps({k: v for k, v in proof.items() if k not in ['instructions', 'ancillary_nil_getter']}))
