"""Execute the complete102-byte original block-start root, not its10NOP pad.

The shield query, PopAction and PushAction bodies are declared services.
Frame/FPS DWORDs are read by the unchanged original instructions. Service
mutations exercise the cached head and live subclass/clock reads. Raw EAX is
retained separately from the public char byte. This is not a whole-world
equivalence claim for missile geometry or action stack/end bodies.
"""
import argparse,collections,copy,hashlib,itertools,json,struct
from pathlib import Path
from unicorn import Uc,UC_ARCH_X86,UC_MODE_32,UC_HOOK_CODE
from unicorn.x86_const import *

p=argparse.ArgumentParser(description=__doc__)
for n in ['game','fixture','proof']:p.add_argument('--'+n,type=Path,required=True)
p.add_argument('--control',type=lambda v:int(v,0),choices=[0xe3f,0x23f],default=0xe3f)
args=p.parse_args();blob=args.game.read_bytes();sha=lambda b:hashlib.sha256(b).hexdigest()
assert sha(blob)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,60)[0];opt=pe+24;base=struct.unpack_from('<I',blob,opt+28)[0];table=opt+struct.unpack_from('<H',blob,pe+20)[0]
u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,(struct.unpack_from('<I',blob,opt+56)[0]+4095)&~4095)
for i in range(struct.unpack_from('<H',blob,pe+6)[0]):
    _,va,size,off=struct.unpack_from('<IIII',blob,table+40*i+8)
    if size:u.mem_write(base+va,blob[off:off+size])
heap,stack,code=0x21000000,0x22000000,0x23000000
for a in [heap,stack,code]:u.mem_map(a,0x10000)
unit,primary,alternate,push_base=heap+0x1000,heap+0x3000,heap+0x6000,heap+0x9000
sp,stop,init,cw=stack+0xe000,code+0x100,code+0x200,code+0x300
frame_addr,fps_addr=0x84ea04,0x85b3fc
read32=lambda a:struct.unpack('<I',u.mem_read(a,4))[0]
put32=lambda a,v:u.mem_write(a,struct.pack('<I',v&0xffffffff))
put8=lambda a,v:u.mem_write(a,bytes([v&255]))
read_index=lambda a:struct.unpack('<b',u.mem_read(a+544,1))[0]
head=lambda update,index:update+552+24*index
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw))
seed=bytes((i*31+19)&255 for i in range(0x10000))
registers={UC_X86_REG_EBX:0x12345678,UC_X86_REG_ESI:0x23456789,UC_X86_REG_EDI:0x3456789a,UC_X86_REG_EBP:0x456789ab}
active,current,events=False,{},[]
traces,services={},collections.Counter()

def state():
    ptr=read32(unit+748);assert ptr in [primary,alternate]
    return dict(update=0 if ptr==primary else 1,subclass=read32(unit+12),frame=read32(frame_addr),fps=read32(fps_addr),indices=[read_index(primary),read_index(alternate)],deadlines=[[read32(head(a,i)+4) for i in range(24)] for a in [primary,alternate]])

def mutation(stage):
    m=current['mutation']
    if m['stage']!=stage:return
    action=m['action'];assert action in ['unit_update','index','class','deadline','clock','all']
    if action in ['unit_update','all']:put32(unit+748,alternate)
    if action in ['index','all']:put8(primary+544,2)
    if action in ['class','all']:put32(unit+12,m['subclass'])
    if action in ['deadline','all']:put32(head(primary,current['head_index'])+4,m['deadline'])
    if action in ['clock','all']:put32(frame_addr,m['frame']);put32(fps_addr,m['fps'])

def hook(_u,pc,size,_):
    if not active:return
    if pc==stop:u.emu_stop();return
    if 0x532070<=pc<0x5320d6:traces[pc]=bytes(u.mem_read(pc,size));return
    kind={0x533e70:'shield',0x50a160:'pop',0x50a260:'push'}.get(pc);assert kind,(current['name'],hex(pc))
    esp=u.reg_read(UC_X86_REG_ESP);assert read32(esp+4)==unit
    event=dict(kind=kind,unit=0,state=state())
    if kind=='push':assert read32(esp+8)==22;event['action']=22
    events.append(event);services[kind]+=1;mutation(kind)
    result=current['shield'] if kind=='shield' else current['pop_return'] if kind=='pop' else 0 if current['push_nil'] else push_base+current['push_low']
    u.reg_write(UC_X86_REG_EAX,result&0xffffffff);u.reg_write(UC_X86_REG_EIP,read32(esp));u.reg_write(UC_X86_REG_ESP,esp+4)

u.hook_add(UC_HOOK_CODE,hook)

def run(c):
    global active,current,events
    active,current,events=False,c,[];u.mem_write(heap,seed)
    put32(unit+8,2);put32(unit+12,c['subclass']);put32(unit+16,0);put32(unit+748,primary)
    put8(primary+544,c['head_index']);put8(alternate+544,2)
    for a in [primary,alternate]:
        for i in range(24):put32(head(a,i),21);put32(head(a,i)+4,0x24680000+i)
    put32(head(primary,c['head_index'])+4,c['deadline']);put32(head(alternate,2)+4,0x1234)
    put32(frame_addr,c['frame']);put32(fps_addr,c['fps'])
    before=bytes(u.mem_read(heap,len(seed)))
    globals_before={a:bytes(u.mem_read(a&~4095,4096)) for a in [frame_addr,fps_addr]}
    u.mem_write(cw,struct.pack('<H',args.control));u.emu_start(init,init+8,count=2)
    u.mem_write(sp-512,bytes(1024));put32(sp,stop);put32(sp+4,unit)
    for reg,v in registers.items():u.reg_write(reg,v)
    u.reg_write(UC_X86_REG_EFLAGS,0x202);u.reg_write(UC_X86_REG_ESP,sp)
    active=True
    try:u.emu_start(0x532070,stop+1,count=2000)
    except Exception as e:raise AssertionError((c['name'],hex(u.reg_read(UC_X86_REG_EIP)))) from e
    active=False
    assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4,c['name']
    assert all(u.reg_read(reg)==v for reg,v in registers.items()),c['name']
    assert u.reg_read(UC_X86_REG_FPCW)==args.control and u.reg_read(UC_X86_REG_FPSW)&0x3800==0,c['name']
    assert bytes(u.mem_read(sp-512,256))==bytes(256) and bytes(u.mem_read(sp+8,256))==bytes(256),c['name']
    after=bytes(u.mem_read(heap,len(seed)));allowed={primary+544,*range(unit+12,unit+16),*range(unit+748,unit+752),*range(head(primary,c['head_index'])+4,head(primary,c['head_index'])+8)}
    assert all(a==b or heap+i in allowed for i,(a,b) in enumerate(zip(before,after))),c['name']
    for addr,old in globals_before.items():
        new=bytes(u.mem_read(addr&~4095,4096));offset=addr&4095
        assert old[:offset]==new[:offset] and old[offset+4:]==new[offset+4:],c['name']
    raw=u.reg_read(UC_X86_REG_EAX)
    return dict(input=c,result=dict(return_byte=raw&255,events=copy.deepcopy(events),state=state()),raw_original_eax=raw)

default=dict(shield=0,frame=100,fps=30,deadline=99,subclass=0,head_index=1,pop_return=0x12345678,push_nil=False,push_low=128,mutation=dict(stage='',action='',subclass=16,deadline=7,frame=0xffffffff,fps=0xffffffff))
cases=[]
def case(prefix,**kw):
    c=copy.deepcopy(default);c.update(kw,name=f'{prefix}-{len(cases)}');cases.append(c)
frames=[0,1,100,0x7fffffff,0x80000000,0xfffffffe,0xffffffff]
for frame,deadline,subclass in itertools.product(frames,[0,1,99,100,101,0x7fffffff,0x80000000,0xffffffff],[0,16,0xa5,0xff]):case('wait-expire',frame=frame,deadline=deadline,subclass=subclass)
for frame,fps,subclass in itertools.product(frames,[0,1,2,29,30,31,60,0x80000000,0xffffffff],[0,16]):case('refresh',shield=1,frame=frame,fps=fps,subclass=subclass)
for low in range(0,256,8):case('push-byte',push_low=low)
case('push-nil',push_nil=True)
for index,shield,subclass in itertools.product([0,1,2,5,22,23],[0,1],[0,16]):case('head',head_index=index,shield=shield,subclass=subclass)
for stage,action,shield in itertools.product(['shield','pop','push'],['unit_update','index','class','deadline','clock','all'],[0,1]):
    m=copy.deepcopy(default['mutation']);m.update(stage=stage,action=action)
    case('mutation',frame=0xffffffff,deadline=99,shield=shield,mutation=m)
for ret,subclass in itertools.product([0xffffffff,0,1,0x12345678,0x80000000],[0x80,0x90]):case('pop-return',pop_return=ret,subclass=subclass)
for low in range(0,256,8):case('native-wrap-push-byte',shield=1,frame=0xffffffff,fps=30,subclass=0,push_low=low)
assert len(cases)==485
rows=[run(c) for c in cases]
executed={pc+j for pc,b in traces.items() for j in range(len(b))};assert executed==set(range(0x532070,0x5320d6))
raw=bytes(u.mem_read(0x532070,102))
fixture=dict(original_sha256=sha(blob),control_word=args.control,rows=rows)
proof=dict(original_sha256=sha(blob),control_word=args.control,contexts=len(cases),faults=0,root_bytes=102,executed_bytes=len(executed),existing_manifest_envelope=112,padding_bytes_unexecuted=10,root_sha256=sha(raw),services=dict(services),only_three_service_bodies_declared=['533E70 shield query','50A160 pop','50A260 push'],raw_original_EAX_retained_public_char_low_byte_recorded=True,register_stack_FPCW_TOP_heap_global_canaries_preserved=True,frame_FPS_DWORD_loads_execute_unmodified=True,valid_monster_and_head_indices_0_to23_only=True,instructions=[dict(pc=hex(pc),bytes=b.hex()) for pc,b in sorted(traces.items())])
for path,data in [(args.fixture,fixture),(args.proof,proof)]:
    with path.open('x') as f:json.dump(data,f,indent=2);f.write('\n')
print(json.dumps(dict(contexts=len(cases),executed_bytes=len(executed),control_word=hex(args.control),services=dict(services),fixture_sha256=sha(args.fixture.read_bytes()))))
