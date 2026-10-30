"""Execute 13 unchanged fight/run entries, sound getter and frame-copy body.

Audio, script callback and sight bodies are declared services. The actual
frame-copy body executes with a read-only entry observation. Every root's
raw DWORD return and all observable effects stay in the reference.
"""
import argparse, gzip, hashlib, itertools, json, struct
from pathlib import Path
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_WRITE
from unicorn.x86_const import *

p = argparse.ArgumentParser(description=__doc__)
for name in ['game', 'fixture', 'proof']:
    p.add_argument('--'+name, type=Path, required=True)
p.add_argument('--control', type=lambda s:int(s, 0), choices=[0xe3f, 0x23f], required=True)
a = p.parse_args(); blob = a.game.read_bytes()
sha = lambda raw: hashlib.sha256(raw).hexdigest()
assert sha(blob) == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe = struct.unpack_from('<I', blob, 60)[0]; op = pe+24
tb = op+struct.unpack_from('<H', blob, pe+20)[0]
base = struct.unpack_from('<I', blob, op+28)[0]
u = Uc(UC_ARCH_X86, UC_MODE_32)
u.mem_map(base, (struct.unpack_from('<I', blob, op+56)[0]+4095)&~4095)
for i in range(struct.unpack_from('<H', blob, pe+6)[0]):
    _, va, n, off = struct.unpack_from('<IIII', blob, tb+i*40+8)
    if n: u.mem_write(base+va, blob[off:off+n])
heap, stack, code = 0x21000000, 0x22000000, 0x23000000
for address in [heap, stack, code]: u.mem_map(address, 0x10000)
unit = heap+0x1000; updates = [heap+0x3000, heap+0x5000]
enemies = [heap+0x7000, heap+0x8000]; sounds = [heap+0xa000, heap+0xb000]
sp, stop, init, cw = stack+0xe000, code+0x100, code+0x200, code+0x300
frame_address, copy_address = 0x84ea04, 0x833b18
rd = lambda pc: struct.unpack('<I', u.mem_read(pc, 4))[0]
wr = lambda pc, value: u.mem_write(pc, struct.pack('<I', value&0xffffffff))
ptrref = lambda pointer, values: values.index(pointer) if pointer else -1
u.mem_write(init, b'\xdb\xe3\xd9\x2d'+struct.pack('<I', cw))
ranges = {0x531e20:100, 0x531e90:33, 0x534750:33, 0x534780:31,
          0x544740:12, 0x544750:12, 0x544920:12, 0x544930:12, 0x544940:12,
          0x5456b0:12, 0x5456c0:12, 0x546410:12, 0x546420:12,
          0x424300:31, 0x5281e0:11}
entries = list(ranges)[:13]
original = {pc: bytes(u.mem_read(pc, n)) for pc,n in ranges.items()}
assert sum(ranges.values()) == 347
seed = bytes((i*31+19)&255 for i in range(0x10000))
regs = {UC_X86_REG_EBX:0x12345678, UC_X86_REG_ESI:0x23456789, UC_X86_REG_EDI:0x3456789a, UC_X86_REG_EBP:0x456789ab}
active = False
current, events, writes, traces = {}, [], [], {}
service_counts = {'audio':0, 'script':0, 'sight':0, 'copy-frame':0}

def state():
    return dict(cls=rd(unit+8), update=updates.index(rd(unit+748)),
                flags=[rd(v+1440) for v in updates],
                sound_sets=[ptrref(rd(v+488), sounds) for v in updates],
                enemies=[ptrref(rd(v+1196), enemies) for v in updates],
                blocks=[[rd(v+1240), rd(v+1244)] for v in updates],
                sound_ids=[rd(v+20) for v in sounds],
                frame=rd(frame_address), copied_frame=rd(copy_address))

def mutate(kind):
    mode = current['mutations'][['audio', 'script', 'sight'].index(kind)]
    cached = current['update']; live = updates.index(rd(unit+748))
    if mode == 'swap-update': wr(unit+748, updates[1-live])
    if mode == 'class-clear': wr(unit+8, rd(unit+8)&~2)
    if mode == 'set-flags':
        wr(updates[cached]+1440, 0x10000)
        wr(updates[1-cached]+1440, 0x80008000)
    if mode == 'set-enemy':
        old = ptrref(rd(updates[cached]+1196), enemies)
        wr(updates[cached]+1196, 0 if old == 1 else enemies[1])
    if mode == 'clear-sound': wr(updates[cached]+488, 0)
    if mode == 'swap-sound':
        old = ptrref(rd(updates[cached]+488), sounds)
        wr(updates[cached]+488, sounds[1 if old == 0 else 0])
    if mode == 'frame': wr(frame_address, rd(frame_address)^0x80000011)
    if mode == 'set-block':
        wr(updates[cached]+1240, rd(updates[cached]+1240)^0x100)
        wr(updates[cached]+1244, rd(updates[cached]+1244)^0x7fffffff)

def hook(_u, pc, size, _):
    if not active: return
    if pc == stop: u.emu_stop(); return
    if any(first <= pc < first+n for first,n in ranges.items()):
        traces[pc] = bytes(u.mem_read(pc, size))
        if pc == 0x5281e0:
            events.append(dict(kind='copy-frame', args=[], state=state()))
            service_counts['copy-frame'] += 1
        return
    assert pc in [0x501960, 0x502490, 0x5281f0], (current['name'], hex(pc))
    esp = u.reg_read(UC_X86_REG_ESP)
    if pc == 0x501960:
        assert rd(esp+8) == unit and rd(esp+12) == rd(esp+16) == 0
        kind, args = 'audio', [rd(esp+4), 0, 0, 0]
    elif pc == 0x502490:
        block = rd(esp+4)
        assert block in [v+1240 for v in updates]
        assert rd(esp+12) == unit
        kind, args = 'script', [[v+1240 for v in updates].index(block), ptrref(rd(esp+8), enemies), 0]
    else:
        assert rd(esp+4) == unit
        kind, args = 'sight', [0]
    events.append(dict(kind=kind, args=args, state=state()))
    service_counts[kind] += 1
    mutate(kind)
    u.reg_write(UC_X86_REG_EAX, 0x87654321)
    u.reg_write(UC_X86_REG_ECX, 0xc1c2c3c4); u.reg_write(UC_X86_REG_EDX, 0xd1d2d3d4)
    u.reg_write(UC_X86_REG_EIP, rd(esp)); u.reg_write(UC_X86_REG_ESP, esp+4)

def write_hook(_u, access, address, size, value, _):
    if not active: return
    if address == copy_address:
        assert size == 4
        writes.append(dict(field='copied_frame', owner=-1, value=value&0xffffffff))
    elif heap <= address < heap+len(seed):
        assert address in [v+1440 for v in updates] and size == 4
        writes.append(dict(field='flags', owner=[v+1440 for v in updates].index(address), value=value&0xffffffff))

u.hook_add(UC_HOOK_CODE, hook); u.hook_add(UC_HOOK_MEM_WRITE, write_hook)
allowed = set()
for address in [unit+8, unit+748]+[v+o for v in updates for o in [1440, 488, 1196, 1240, 1244]]:
    allowed.update(range(address-heap, address-heap+4))

def run(c, nil_getter=False):
    global active, current, events, writes
    active=False;current=c;events=[];writes=[];u.mem_write(heap,seed)
    wr(unit+8,c['cls']);wr(unit+748,updates[c['update']])
    for i,v in enumerate(updates):
        wr(v+1440,c['flags'][i])
        ref=c['sound_sets'][i];wr(v+488,0 if ref<0 else sounds[ref])
        ref=c['enemies'][i];wr(v+1196,0 if ref<0 else enemies[ref])
        wr(v+1240,c['blocks'][i][0]);wr(v+1244,c['blocks'][i][1])
    for v,value in zip(sounds,c['sound_ids']):wr(v+20,value)
    wr(frame_address,c['frame']);wr(copy_address,c['copied_frame'])
    before=bytes(u.mem_read(heap,len(seed)))
    wr(cw,a.control);u.emu_start(init,init+8,count=2)
    u.mem_write(sp-512,bytes(1024));wr(sp,stop);wr(sp+4,0 if nil_getter else unit)
    for reg,value in regs.items():u.reg_write(reg,value)
    u.reg_write(UC_X86_REG_EAX,0xa5a5a5a5);u.reg_write(UC_X86_REG_ESP,sp);u.reg_write(UC_X86_REG_EFLAGS,0x202)
    active=True;u.emu_start(0x424300 if nil_getter else c['entry'],stop+1,count=1000);active=False
    assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4,c['name']
    assert all(u.reg_read(reg)==value for reg,value in regs.items()),c['name']
    assert u.reg_read(UC_X86_REG_FPCW)==a.control and u.reg_read(UC_X86_REG_FPSW)&0x3800==0
    assert bytes(u.mem_read(sp-512,256))==bytes(256) and bytes(u.mem_read(sp+8,256))==bytes(256)
    after=bytearray(u.mem_read(heap,len(seed)))
    for offset in allowed:after[offset]=before[offset]
    assert after==before,c['name']
    eax=u.reg_read(UC_X86_REG_EAX)
    if not nil_getter:assert eax==rd(updates[state()['update']]+1440),c['name']
    return dict(input=c,result=dict(events=list(events),state=state(),returned_status=eax),original_writes=list(writes),raw_original_EAX=eax,raw_original_AL=eax&255)

cases=[]
def case(prefix,**kw):
    c=dict(entry=0x531e20,cls=2,update=0,flags=[0,0x8001c100],sound_sets=[0,1],enemies=[0,-1],
           blocks=[[0x30000000,0x70000000],[0x30000001,0x70000001]],sound_ids=[123,987],
           frame=0xfffffff0,copied_frame=0x11223344,mutations=['none','none','none'])
    c.update(kw,name=prefix+'-'+str(len(cases)));cases.append(c)
flag_values=[sum(bit for bit,on in zip([0x100,0x4000,0x8000,0x10000,0x80000000],values)if on)for values in itertools.product([0,1],repeat=5)]
for entry,flags,cls,update in itertools.product(entries,flag_values,[0,2,4,6,0x1000002],[0,1]):
    pair=[flags,flags^0x8001c100];ref=[-1,0,1][len(cases)%3];sets=[ref,1 if ref<0 else 1-ref]
    if update:pair.reverse();sets.reverse()
    case('flags-and-shadow-class',entry=entry,flags=pair,cls=cls,update=update,sound_sets=sets)
modes=['none','swap-update','class-clear','set-flags','set-enemy','clear-sound','swap-sound','frame','set-block']
for update,mutations in itertools.product([0,1],itertools.product(modes,repeat=3)):
    case('audio-script-sight-live-mutations',update=update,mutations=list(mutations),
         flags=[0x100,0x10000]if update==0 else[0x10000,0x100],sound_sets=[update,1-update])
assert len(cases)==5618
rows=[run(c)for c in cases]
ancillary=run(dict(cases[0],name='nil-sound-getter-ancillary'),nil_getter=True)
assert ancillary['raw_original_EAX']==0 and not ancillary['result']['events']
visited={pc+i for pc,raw in traces.items()for i in range(len(raw))}
expected={pc+i for pc,n in ranges.items()for i in range(n)}
assert visited==expected,[hex(pc)for pc in sorted(expected-visited)]
assert all(bytes(u.mem_read(pc,len(raw)))==raw for pc,raw in original.items())
fixture=dict(original_sha256=sha(blob),control_word=a.control,rows=rows)
proof=dict(original_sha256=sha(blob),control_word=a.control,primary_contexts=len(rows),root_entries=13,helper_entries=2,
 ancillary_nil_getter=ancillary,executed_instruction_bytes=len(visited),root_bytes={hex(pc):n for pc,n in ranges.items()},
 root_sha256={hex(pc):sha(raw)for pc,raw in original.items()},
 declared_services=['AudioEvent501960','ScriptCallback502490','UpdateSight5281F0'],
 actual_sound_getter_and_frame_copy_bodies_execute=True,frame_copy_entry_is_read_only_observation=True,
 service_counts=service_counts,original_script_callback_arguments=3,extra_Go_event13_not_an_original_stack_argument=True,
 every_raw_DWORD_return_matches_live_update_status=True,original_code_unmodified=True,
 register_stack_FPCW_TOP_and_heap_canaries_preserved=True,
 whole_heap_outside_declared_mutations_and_status_writes_unchanged=True,
 original_status_framecopy_write_history_and_raw_EAX_AL_retained=True,
 unexecuted_NOP_tails=[12,15,15,1]+[4]*9+[1,5],full_MainAI_movement_or_script_VM_execution=False,
 instructions=[dict(pc=hex(pc),bytes=raw.hex())for pc,raw in sorted(traces.items())])
for path,value in [(a.fixture,fixture),(a.proof,proof)]:
    with path.open('xb')as f:
        if path.suffix=='.gz':
            with gzip.GzipFile(filename='',mode='wb',fileobj=f,compresslevel=9,mtime=0)as stream:
                stream.write(json.dumps(value,separators=(',',':')).encode()+b'\n')
        else:f.write((json.dumps(value,indent=2)+'\n').encode())
print(json.dumps({k:v for k,v in proof.items()if k not in ['instructions','ancillary_nil_getter']}))
