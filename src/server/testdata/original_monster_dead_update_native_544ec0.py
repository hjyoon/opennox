"""Execute original DeadUpdate163 bytes and Cleanup117 bytes without rewriting code.

IsZombie, RemoveUpdatable, SparkExplosion, BurnDelete, RaiseZombie and
DelayedDelete are declared services. Cleanup executes as real original code,
including its update-pointer return. No544FF0 instruction or padding executes.
"""
import argparse, collections, copy, hashlib, itertools, json, struct
from pathlib import Path
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE
from unicorn.x86_const import *

p = argparse.ArgumentParser(description=__doc__)
for n in ['game', 'fixture', 'proof']: p.add_argument('--'+n, type=Path, required=True)
p.add_argument('--control', type=lambda v:int(v,0), choices=[0xe3f,0x23f], default=0xe3f)
a = p.parse_args(); blob = a.game.read_bytes(); sha = lambda b:hashlib.sha256(b).hexdigest()
assert sha(blob) == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,60)[0]; opt=pe+24; base=struct.unpack_from('<I',blob,opt+28)[0]; table=opt+struct.unpack_from('<H',blob,pe+20)[0]
u=Uc(UC_ARCH_X86,UC_MODE_32); u.mem_map(base,(struct.unpack_from('<I',blob,opt+56)[0]+4095)&~4095)
for i in range(struct.unpack_from('<H',blob,pe+6)[0]):
    _,va,size,off=struct.unpack_from('<IIII',blob,table+40*i+8)
    if size:u.mem_write(base+va,blob[off:off+size])
heap,stack,code=0x21000000,0x22000000,0x23000000
for addr in [heap,stack,code]:u.mem_map(addr,0x10000)
unit=heap+0x1000; updates=[heap+0x3000,heap+0x6000]; definitions=[heap+0x9000,heap+0xa000]
sp,stop,init,cw=stack+0xe000,code+0x100,code+0x200,code+0x300
read32=lambda addr:struct.unpack('<I',u.mem_read(addr,4))[0]
put32=lambda addr,v:u.mem_write(addr,struct.pack('<I',v&0xffffffff))
put8=lambda addr,v:u.mem_write(addr,bytes([v&255]))
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw))
seed=bytes((i*31+19)&255 for i in range(0x10000))
registers={UC_X86_REG_EBX:0x12345678,UC_X86_REG_ESI:0x23456789,UC_X86_REG_EDI:0x3456789a,UC_X86_REG_EBP:0x456789ab}
ranges={0x544ec0:163,0x544f70:117}; original={pc:bytes(u.mem_read(pc,n)) for pc,n in ranges.items()}
active=False; current={}; events=[]; traces={}; services=collections.Counter()

def update_state(ptr):
    return dict(waypoints=read32(ptr+296),waypoint_ids=[read32(ptr+300+i*4) for i in range(16)],
        seen=u.mem_read(ptr+1129,1)[0],enemy_ids=[read32(ptr+1132+i*4) for i in range(16)],
        enemy=read32(ptr+1196)!=0,field91=read32(ptr+364),electric=u.mem_read(ptr+2094,1)[0],
        electric_canary=u.mem_read(ptr+2095,1)[0],stamina_canaries=list(u.mem_read(ptr+1128,1)+u.mem_read(ptr+1130,2)),
        status=read32(ptr+1440),definition=definitions.index(read32(ptr+484)),dead_frame=read32(ptr+548),duration=read32(ptr+492))

def state():
    return dict(update=updates.index(read32(unit+748)),update_nil=read32(unit+744)==0,
        cls=read32(unit+8),subclass=read32(unit+12),flags=read32(unit+16),
        position=list(struct.unpack('<2f',u.mem_read(unit+56,8))),frame=read32(0x84ea04),
        updates=[update_state(v) for v in updates],definitions=[read32(v+92) for v in definitions])

def mutate(stage):
    m=current['mutation']
    if m['stage']!=stage:return
    action=m['action']; value=m['value']
    assert action in ['swap-update','primary-status','alternate-status','definition-alt','primary-delete','alternate-delete','frame','counts','enemy','all']
    if action in ['swap-update','all']:put32(unit+748,updates[1])
    if action in ['primary-status','all']:put32(updates[0]+1440,value)
    if action in ['alternate-status','all']:put32(updates[1]+1440,value^0x80000)
    if action in ['definition-alt','all']:put32(updates[0]+484,definitions[1])
    if action in ['primary-delete','all']:put32(definitions[0]+92,value)
    if action in ['alternate-delete','all']:put32(definitions[1]+92,value)
    if action in ['frame','all']:put32(0x84ea04,value)
    if action in ['counts','all']:
        put32(updates[0]+296,1);put8(updates[0]+1129,1)
        put32(updates[1]+296,16);put8(updates[1]+1129,16)
    if action in ['enemy','all']:put32(updates[0]+1196,0);put32(updates[1]+1196,0x44000001)
    if action=='all':
        put32(unit+8,0x400000);put32(unit+12,0xfedcba98);put32(unit+16,0x87654321)
        u.mem_write(unit+56,struct.pack('<2f',-37.5,19.25))

def hook(_u,pc,size,_):
    if not active:return
    if pc==stop:u.emu_stop();return
    if any(start<=pc<start+n for start,n in ranges.items()):traces[pc]=bytes(u.mem_read(pc,size));return
    kind={0x534a40:'zombie',0x4da920:'remove',0x5231b0:'spark',0x544ce0:'burn',0x534ab0:'raise',0x4e5cc0:'delete'}.get(pc)
    assert kind,(current['name'],hex(pc))
    esp=u.reg_read(UC_X86_REG_ESP)
    if kind=='spark':assert read32(esp+4)==unit+56 and read32(esp+8)==100
    else:assert read32(esp+4)==unit
    events.append(dict(kind=kind,unit=0,state=state()));services[kind]+=1;mutate(kind)
    u.reg_write(UC_X86_REG_EAX,current['zombie'] if kind=='zombie' else current['service_return'])
    u.reg_write(UC_X86_REG_EIP,read32(esp));u.reg_write(UC_X86_REG_ESP,esp+4)
u.hook_add(UC_HOOK_CODE,hook)

allowed=set()
for ptr,off,n in [(unit,8,12),(unit,56,8),(unit,744,8)]+[(v,o,n) for v in updates for o,n in [(296,72),(484,4),(492,4),(548,4),(1129,1),(1132,68),(1440,4),(2094,1)]]+[(v,92,4) for v in definitions]:
    allowed.update(range(ptr+off,ptr+off+n))

def run(c):
    global active,current,events
    active,current,events=False,c,[];u.mem_write(heap,seed)
    put32(unit+8,2);put32(unit+12,0xa5);put32(unit+16,0x12008014)
    put32(unit+744,code+0x800);put32(unit+748,updates[c['update']])
    u.mem_write(unit+56,struct.pack('<2f',321.,654.));put32(0x84ea04,c['frame'])
    for index,ptr in enumerate(updates):
        put32(ptr+296,c['waypoints'] if index==0 else c['alternate_waypoints']);put32(ptr+364,0x12345678+index)
        put8(ptr+1129,c['seen'] if index==0 else c['alternate_seen']);put32(ptr+1196,0x44000001 if c['enemy'] else 0)
        put8(ptr+1128,0xa5+index);u.mem_write(ptr+1130,struct.pack('<H',0xa5a5+index*0x101))
        put8(ptr+2094,0x34+index);put8(ptr+2095,0xd5+index)
        put32(ptr+1440,c['status'] if index==0 else c['alternate_status']);put32(ptr+484,definitions[index])
        put32(ptr+548,c['dead_frame']);put32(ptr+492,c['duration'])
        for i in range(16):
            put32(ptr+300+i*4,0x41000000+index*0x100+i+1);put32(ptr+1132+i*4,0x42000000+index*0x100+i+1)
    put32(definitions[0]+92,c['delete']);put32(definitions[1]+92,c['alternate_delete'])
    before=bytes(u.mem_read(heap,len(seed)))
    u.mem_write(cw,struct.pack('<H',a.control));u.emu_start(init,init+8,count=2)
    u.mem_write(sp-512,bytes(1024));put32(sp,stop);put32(sp+4,unit)
    for reg,v in registers.items():u.reg_write(reg,v)
    u.reg_write(UC_X86_REG_EFLAGS,0x202);u.reg_write(UC_X86_REG_ESP,sp);active=True
    try:u.emu_start(0x544ec0 if c['entry']=='update' else 0x544f70,stop+1,count=10000)
    except Exception as e:raise AssertionError((c['name'],hex(u.reg_read(UC_X86_REG_EIP)))) from e
    active=False
    assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4,c['name']
    assert all(u.reg_read(reg)==v for reg,v in registers.items()),c['name']
    assert u.reg_read(UC_X86_REG_FPCW)==a.control and u.reg_read(UC_X86_REG_FPSW)&0x3800==0,c['name']
    assert bytes(u.mem_read(sp-512,256))==bytes(256) and bytes(u.mem_read(sp+8,256))==bytes(256),c['name']
    after=bytes(u.mem_read(heap,len(seed)))
    assert all(x==y or heap+i in allowed for i,(x,y) in enumerate(zip(before,after))),c['name']
    eax=u.reg_read(UC_X86_REG_EAX)
    if c['entry']=='cleanup':assert eax==updates[c['update']],c['name']
    return dict(input=c,result=dict(events=copy.deepcopy(events),state=state()),raw_original_eax=eax,
        cleanup_return_update=updates.index(eax) if c['entry']=='cleanup' else -1)

cases=[]
def case(prefix,**kw):
    c=dict(entry='update',update=0,zombie=0,status=0,alternate_status=0x80000,frame=13,dead_frame=10,duration=2,
        enemy=True,waypoints=2,seen=3,alternate_waypoints=1,alternate_seen=2,delete=0,alternate_delete=1,
        service_return=[0,1,0x7fffffff,0x80000000,0xffffffff][len(cases)%5],mutation=dict(stage='',action='',value=0))
    c.update(kw,name=f'{prefix}-{len(cases)}');cases.append(c)
for wp,seen in itertools.product(list(range(17))+[0x80000000,0xfffffffe,0xffffffff],range(17)):
    case('cleanup-prefix',entry='cleanup',update=len(cases)%2,waypoints=wp,seen=seen,alternate_waypoints=wp,alternate_seen=seen)
for count,delete in itertools.product(range(17),[0,1]):case('ordinary-prefix',waypoints=count,seen=16-count,delete=delete)
times=[(12,10,2),(13,10,2),(1,0xfffffffe,2),(0,0xffffffff,0xffffffff),(0x80000000,0,0x7fffffff)]
for query,status,time,enemy in itertools.product([1,2,0x80000000,0xffffffff],[0,1,0x7ffff,0x80000,0x100000,0xffffffff],times,[False,True]):
    case('zombie-branches',zombie=query,status=status,frame=time[0],dead_frame=time[1],duration=time[2],enemy=enemy)
actions=['swap-update','primary-status','alternate-status','definition-alt','primary-delete','alternate-delete','frame','counts','enemy','all']
for stage,mode,action in itertools.product(['zombie','remove','spark','burn','raise','delete'],['ordinary','fire','expired'],actions):
    if stage!='zombie' and mode!={'remove':'ordinary','spark':'fire','burn':'fire','raise':'expired','delete':'ordinary'}[stage]:continue
    case('live-mutation',zombie=0 if mode=='ordinary' else 1,status=0x80000 if mode=='fire' else 0,delete=1,
        mutation=dict(stage=stage,action=action,value=0 if mode=='fire' else 0x80000))
assert len(cases)==694
rows=[run(c) for c in cases]
executed={pc+j for pc,b in traces.items() for j in range(len(b))}
assert executed=={pc+j for pc,n in ranges.items() for j in range(n)}
assert all(bytes(u.mem_read(pc,n))==original[pc] for pc,n in ranges.items())
fixture=dict(original_sha256=sha(blob),control_word=a.control,rows=rows)
proof=dict(original_sha256=sha(blob),control_word=a.control,contexts=len(cases),faults=0,root_bytes=ranges,executed_bytes=len(executed),
    original_root_sha256={hex(pc):sha(b) for pc,b in original.items()},services=dict(services),
    only_six_service_bodies_declared=['534A40 IsZombie','4DA920 RemoveUpdatable','5231B0 SparkExplosion','544CE0 BurnDelete','534AB0 RaiseZombie','4E5CC0 DelayedDelete'],
    original_cleanup_executed_with_native_pointer_identity_contract=True,raw_original_EAX_retained_update_void_return_excluded=True,
    register_stack_FPCW_TOP_heap_canaries_preserved=True,separate_544FF0_and_all_padding_unexecuted=True,
    manifest_544EC0_envelope=176,manifest_544F70_envelope=176,cleanup_envelope_separate_544FF0_bytes=39,cleanup_envelope_NOPs=20,
    instructions=[dict(pc=hex(pc),bytes=b.hex()) for pc,b in sorted(traces.items())])
for path,data in [(a.fixture,fixture),(a.proof,proof)]:
    with path.open('x') as f:json.dump(data,f,indent=2);f.write('\n')
print(json.dumps(dict(contexts=len(cases),executed_bytes=len(executed),control_word=hex(a.control),services=dict(services),fixture_sha256=sha(a.fixture.read_bytes()))))
