"""Execute complete original TrapDoor/UndeadKiller update callers."""
import argparse,collections,copy,hashlib,json,struct
from pathlib import Path
from unicorn import Uc,UC_ARCH_X86,UC_MODE_32,UC_HOOK_CODE,UC_HOOK_MEM_INVALID
from unicorn.x86_const import *
p=argparse.ArgumentParser(description=__doc__)
for n in ('game','fixture','proof','fault-proof'):p.add_argument('--'+n,type=Path,required=True)
args=p.parse_args();blob=args.game.read_bytes();sha=lambda b:hashlib.sha256(b).hexdigest()
assert sha(blob)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,60)[0];opt=pe+24;base=struct.unpack_from('<I',blob,opt+28)[0]
size=(struct.unpack_from('<I',blob,opt+56)[0]+4095)&~4095;tab=opt+struct.unpack_from('<H',blob,pe+20)[0]
u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,size)
for n in range(struct.unpack_from('<H',blob,pe+6)[0]):
    _,va,count,off=struct.unpack_from('<IIII',blob,tab+40*n+8)
    if count:u.mem_write(base+va,blob[off:off+count])
stack,heap,stop=0x20000000,0x21000000,0x30000000
for a in (stack,heap,stop):u.mem_map(a,0x10000)
sp=stack+0xe000;source=heap+128;traps=[heap+0x4000,heap+0x4100];undead=[heap+0x4200,heap+0x4300];spells=[heap+0x5000,heap+0x5200]
get=lambda a:struct.unpack('<I',u.mem_read(a,4))[0]
put=lambda a,v:u.mem_write(a,struct.pack('<I',v&0xffffffff))
ref=lambda rows,n:0 if n<0 else rows[n]
idx=lambda rows,a:-1 if a==0 else rows.index(a)
frame_addr,fps_addr=0x84ea04,0x85b3fc
specs=[('trapdoor',0x53de80,184),('undead',0x53e190,58)]
regions=[dict(name=n,address=hex(a),size=z,sha256=sha(bytes(u.mem_read(a,z)))) for n,a,z in specs]
events=[];faults=[];trace={};entries=collections.Counter();current={};applied=False;recording=False
def state():
    return dict(flags=get(source+16),status=get(source+20),next=get(source+136),collide_ref=idx(traps if current['kind']==0 else undead,get(source+700)),traps=[[get(a+4*n) for n in range(7)] for a in traps],undead_refs=[idx(spells,get(a)) for a in undead],spell_flags=[get(a+88) for a in spells],frame=get(frame_addr),fps=get(fps_addr))
def event(name,values):events.append(dict(service=name,args=values,**state()))
def mutate(service):
    global applied
    key=current['mutation']
    if applied or not key.startswith(service+':'):return
    applied=True;action=key.split(':',1)[1]
    if action=='clock':put(frame_addr,0xfffffffc)
    elif action=='fps':put(fps_addr,0x80000001)
    elif action=='flags':put(source+16,0x80000000)
    elif action=='status':put(source+20,0x10000002)
    elif action=='next':put(source+136,0x12345678)
    elif action=='deadline':put(traps[0]+16,0xfffffffc)
    elif action=='replacement-deadline':put(traps[1]+16,99)
    elif action=='collide-swap':put(source+700,(traps if current['kind']==0 else undead)[1])
    elif action=='collide-nil':put(source+700,0)
    elif action=='spell-swap':put(undead[0],spells[1])
    elif action=='cancel':put(spells[0]+88,1)
    elif action=='activate':put(traps[0]+24,99)
    else:raise AssertionError(key)
def ret(v=0):
    esp=u.reg_read(UC_X86_REG_ESP);u.reg_write(UC_X86_REG_EAX,v);u.reg_write(UC_X86_REG_EIP,get(esp));u.reg_write(UC_X86_REG_ESP,esp+4)
def hook(_u,address,length,_):
    if address in (stop,0x4014c4):u.emu_stop();return
    if not recording:return
    for n,a,z in specs:
        if a<=address<a+z:
            trace[address]=bytes(u.mem_read(address,length)).hex()
            if address==a:entries[n]+=1
            return
    esp=u.reg_read(UC_X86_REG_ESP);arg=lambda n:get(esp+4*n)
    if address in (0x4e4670,0x4e4780,0x4e4800):
        assert arg(1)==source
        service={0x4e4670:'on',0x4e4780:'unset',0x4e4800:'set'}[address];v=arg(2);event(service,[0,v])
        if service=='on':assert v==1;put(source+16,get(source+16)|0x1000000)
        elif service=='unset':put(source+20,get(source+20)&~v)
        else:put(source+20,get(source+20)|v)
        mutate(service);ret()
    elif address==0x501960:
        assert [arg(1),arg(2),arg(3),arg(4)]==[874,source,0,0];event('audio',[874,0,0,0]);mutate('audio');ret()
    elif address==0x4e5cc0:
        assert arg(1)==source;event('delete',[0]);mutate('delete');ret()
    else:raise AssertionError((hex(address),current['name']))
def invalid(_u,access,address,length,value,_):
    faults.append(dict(address=address,size=length,instruction=u.reg_read(UC_X86_REG_EIP)));return False
u.hook_add(UC_HOOK_CODE,hook);u.hook_add(UC_HOOK_MEM_INVALID,invalid)
u.reg_write(UC_X86_REG_ESP,sp);u.reg_write(UC_X86_REG_FPCW,0x27f);u.emu_start(0x4014b5,0x4014c4,count=20000)
assert u.reg_read(UC_X86_REG_FPCW)==0xe3f
initial=u.context_save();recording=True;cases=[]
def add(kind,name,**kw):
    c=dict(name=specs[kind][0]+'/'+name,kind=kind,flags=0,status=2,next=1,collide_ref=0,traps=[[11,12,13,14,9,0x11220003,7],[21,22,23,24,88,0x33440005,8]],undead_refs=[0,1],spell_flags=[0,1],frame=10,fps=30,mutation='')
    c.update(kw);cases.append(c);return c
for flags in (0,0x1000000,0xffffffff):
    for status in (0,1,2,4,6,8,0x102,0x104,0x10000008):
        for frame in (0,1,10,0x80000000,0xffffffff):
            for delta in (-1,0,1):
                c=add(0,f'state/{flags}/{status}/{frame}/{delta}',flags=flags,status=status,frame=frame);c['traps'][0][4]=(frame+delta)&0xffffffff
for frame in (1,0x80000000,0xffffffff):
    for fps in (0,1,30,0x80000001,0xffffffff):add(0,f'wrap/{frame}/{fps}',frame=frame,fps=fps)
for flags in (0,1,2,0x100,0x101,0xffffffff):
    for spell_ref in (-1,0,1):
        for frame in (0,10,0x80000000,0xffffffff):
            for age in (69,70,71):add(1,f'expiry/{flags}/{spell_ref}/{frame}/{age}',spell_flags=[flags,2],undead_refs=[spell_ref,1],frame=frame,next=(frame-age)&0xffffffff)
for kind,services in [(0,('on','unset','set','audio')),(1,('delete',))]:
    for service in services:
        for action in ('clock','fps','flags','status','next','deadline','replacement-deadline','collide-swap','collide-nil','spell-swap','cancel','activate'):
            add(kind,f'mutate/{service}/{action}',mutation=service+':'+action,spell_flags=[1,2])
def prepare(c,source_ref=0):
    global applied
    u.context_restore(initial);u.mem_write(heap,b'\xa5'*0x10000);u.mem_write(stack,bytes(0x10000))
    for off,key in ((16,'flags'),(20,'status'),(136,'next')):put(source+off,c[key])
    put(source+700,ref(traps if c['kind']==0 else undead,c['collide_ref']))
    for a,words in zip(traps,c['traps']):
        for n,v in enumerate(words):put(a+4*n,v)
    for a,n in zip(undead,c['undead_refs']):put(a,ref(spells,n))
    for a,v in zip(spells,c['spell_flags']):put(a+88,v)
    put(frame_addr,c['frame']);put(fps_addr,c['fps']);put(sp,stop);put(sp+4,source if source_ref==0 else 0)
    u.reg_write(UC_X86_REG_ESP,sp);events.clear();faults.clear();applied=False
def preserved(before):
    expected=bytearray(before)
    for off in (16,20,136,700):expected[source-heap+off:source-heap+off+4]=u.mem_read(source+off,4)
    for a in traps:expected[a-heap:a-heap+28]=u.mem_read(a,28)
    for a in undead:expected[a-heap:a-heap+4]=u.mem_read(a,4)
    for a in spells:expected[a-heap+88:a-heap+92]=u.mem_read(a+88,4)
    return bytes(expected)==bytes(u.mem_read(heap,0x10000))
rows=[]
for current in cases:
    prepare(current);before=bytes(u.mem_read(heap,0x10000));u.emu_start(specs[current['kind']][1],stop,count=20000)
    assert not faults and u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4 and u.reg_read(UC_X86_REG_FPCW)==0xe3f and preserved(before)
    rows.append(dict(input=copy.deepcopy(current),output=state(),events=copy.deepcopy(events),nonoutput_bytes_preserved=True))
assert args.game.read_bytes()==blob
proof=dict(original_sha256=sha(blob),original_file_unchanged=True,original_startup_FPCW=0xe3f,contexts=len(cases),executions=len(rows),regions=regions,entries=dict(entries),rows=rows,faults=[],executed_instructions=[dict(address=hex(a),bytes=b) for a,b in sorted(trace.items())],arithmetic_or_conversion_substitutions=0,scope='Complete unchanged TrapDoor and UndeadKiller callers with declared status/on-off/audio/delete bodies. Integer operations only. EAX residue of void update callbacks is not a pointer output. Native original-game/full-world composition is separate.')
with args.proof.open('x') as f:json.dump(proof,f,separators=(',',':'))
with args.fixture.open('x') as f:json.dump({k:v for k,v in proof.items() if k!='executed_instructions'},f,separators=(',',':'))
fault_rows=[]
for kind,name,source_ref,record,status,next_frame,fault in [(0,'nil-source',-1,0,2,1,True),(0,'enabled-opening-nil-record',0,-1,2,1,False),(0,'enabled-priority-nil-record',0,-1,6,1,False),(0,'enabled-closing-nil-record',0,-1,4,1,True),(0,'enabled-idle-nil-record',0,-1,8,1,True),(0,'disabled-nil-record',0,-1,0,1,True),(1,'nil-source',-1,0,0,1,True),(1,'nil-record',0,-1,0,1,True),(1,'nil-spell-live',0,0,0,10,False),(1,'nil-spell-expired',0,0,0,0xffffff00,False)]:
    current=copy.deepcopy(cases[0]);current.update(name=specs[kind][0]+'/'+name,kind=kind,collide_ref=record,status=status,next=next_frame,frame=10,flags=0x1000000 if name.startswith('enabled') else 0,undead_refs=[-1,1],mutation='')
    prepare(current,source_ref);before=bytes(u.mem_read(heap,0x10000));error=None
    try:u.emu_start(specs[kind][1],stop,count=20000)
    except Exception as exc:error=str(exc)
    assert bool(faults)==fault and preserved(before)
    fault_rows.append(dict(input=current,source_ref=source_ref,fault=fault,faults=list(faults),error=error,events=copy.deepcopy(events),output=state(),nonoutput_bytes_preserved=True))
with args.fault_proof.open('x') as f:json.dump(dict(original_sha256=sha(blob),rows=fault_rows),f,separators=(',',':'))
print(json.dumps({k:proof[k] for k in ('contexts','executions','entries','faults')}))
