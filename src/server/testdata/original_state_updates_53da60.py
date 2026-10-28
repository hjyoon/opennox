"""Run four complete unchanged GAME.EXE updates with declared world services."""
import argparse, collections, copy, hashlib, json, struct
from pathlib import Path
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import *

p=argparse.ArgumentParser(description=__doc__)
for name in ('game','fixture','proof','fault-proof'):p.add_argument('--'+name,type=Path,required=True)
args=p.parse_args();blob=args.game.read_bytes();sha=lambda b:hashlib.sha256(b).hexdigest()
assert sha(blob)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,60)[0];opt=pe+24;base=struct.unpack_from('<I',blob,opt+28)[0]
size=(struct.unpack_from('<I',blob,opt+56)[0]+4095)&~4095;tab=opt+struct.unpack_from('<H',blob,pe+20)[0]
u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,size)
for n in range(struct.unpack_from('<H',blob,pe+6)[0]):
    _,va,count,off=struct.unpack_from('<IIII',blob,tab+40*n+8)
    if count:u.mem_write(base+va,blob[off:off+count])
stack,heap,stop=0x20000000,0x21000000,0x30000000
for addr in (stack,heap,stop):u.mem_map(addr,0x10000)
sp=stack+0xe000;objects=[heap+128+1024*n for n in range(4)]
get=lambda a:struct.unpack('<I',u.mem_read(a,4))[0]
put=lambda a,v:u.mem_write(a,struct.pack('<I',v&0xffffffff))
f32=lambda v:struct.unpack('<I',struct.pack('<f',v))[0]
s32=lambda v:struct.unpack('<i',struct.pack('<I',v&0xffffffff))[0]
ref=lambda n:0 if n<0 else objects[n]
objref=lambda a:-1 if a==0 else objects.index(a)
frame_addr,fps_addr,cache_addr=0x84ea04,0x85b3fc,0x833f18
offsets=(16,20,56,60,128,136,484)
specs=[('arachnaphobia',0x53da60,157),('break',0x53db30,124),('open',0x53dbb0,116),('break-remove',0x53dc30,134)]
regions=[dict(name=n,address=hex(a),size=z,sha256=sha(bytes(u.mem_read(a,z)))) for n,a,z in specs]
def state():
    return dict(objects=[[get(a+off) for off in offsets] for a in objects],owners=[objref(get(a+508)) for a in objects],frame=get(frame_addr),fps=get(fps_addr),cache=get(cache_addr))
events=[];faults=[];trace={};entries=collections.Counter();current={};applied=False;recording=False
def event(name,values):events.append(dict(service=name,args=values,**state()))
def mutate(service):
    global applied
    key=current['mutation']
    if applied or not key.startswith(service+':'):return
    applied=True;action=key.split(':',1)[1];a=objects[0]
    if action=='clock':put(frame_addr,0xfffffffc)
    elif action=='fps':put(fps_addr,0x80000001)
    elif action=='owner':put(a+508,objects[3])
    elif action=='owner-nil':put(a+508,0)
    elif action=='position':put(a+56,0x80000000);put(a+60,0x7fc12345)
    elif action=='created':put(a+128,0xfffffffa)
    elif action=='deadline':put(a+136,0x12345678)
    elif action=='cache':put(cache_addr,99)
    elif action=='flags':put(a+16,0x12348001)
    elif action=='status':put(a+20,0x10000002)
    else:raise AssertionError(key)
def ret(value=0):
    esp=u.reg_read(UC_X86_REG_ESP);u.reg_write(UC_X86_REG_EAX,value&0xffffffff);u.reg_write(UC_X86_REG_EIP,get(esp));u.reg_write(UC_X86_REG_ESP,esp+4)
def hook(_u,address,length,_):
    if address in (stop,0x4014c4):u.emu_stop();return
    if not recording:return
    for name,lo,z in specs:
        if lo<=address<lo+z:
            trace[address]=bytes(u.mem_read(address,length)).hex()
            if address==lo:entries[name]+=1
            return
    esp=u.reg_read(UC_X86_REG_ESP);arg=lambda n:get(esp+4*n)
    if address==0x4e3aa0:
        assert bytes(u.mem_read(arg(1),12))==b'SmallSpider\0';event('lookup',[]);mutate('lookup');ret(current['lookup'])
    elif address==0x4e3450:event('new',[s32(arg(1))]);mutate('new');ret(ref(current['spawn_ref']))
    elif address==0x4daa50:
        assert arg(5)==0;event('create',[objref(arg(1)),objref(arg(2)),arg(3),arg(4),0]);mutate('create');ret(1)
    elif address==0x415fa0:
        assert [arg(1),arg(2)]==[1,5];event('random',[1,5]);mutate('random');ret(current['delay'])
    elif address in (0x4e4780,0x4e4800):
        service='unset' if address==0x4e4780 else 'set';a=arg(1);mask=arg(2)
        event(service,[objref(a),mask]);put(a+20,get(a+20)&~mask if service=='unset' else get(a+20)|mask);mutate(service);ret()
    elif address==0x4da920:
        a=arg(1);event('remove',[objref(a)]);put(a+484,0);mutate('remove');ret()
    elif address==0x4e5cc0:event('delete',[objref(arg(1))]);mutate('delete');ret()
    else:raise AssertionError((hex(address),current['name']))
def invalid(_u,access,address,length,value,_):
    faults.append(dict(address=address,size=length,instruction=u.reg_read(UC_X86_REG_EIP)));return False
u.hook_add(UC_HOOK_CODE,hook);u.hook_add(UC_HOOK_MEM_INVALID,invalid)
u.reg_write(UC_X86_REG_ESP,sp);u.reg_write(UC_X86_REG_FPCW,0x27f);u.emu_start(0x4014b5,0x4014c4,count=20000)
assert u.reg_read(UC_X86_REG_FPCW)==0xe3f
initial=u.context_save();recording=True;cases=[]
def add(kind,name,**kw):
    c=dict(name=specs[kind][0]+'/'+name,kind=kind,objects=[[0x8000,2,f32(10.1),f32(-20.3),1,9,1]]+[[0,0,0,0,0,0,1] for _ in range(3)],owners=[2,-1,-1,-1],frame=10,fps=30,cache=0,lookup=17,spawn_ref=1,delay=3,mutation='')
    c.update(kw);cases.append(c);return c
for frame in (0,1,10,0x7fffffff,0x80000000,0xfffffffc,0xffffffff):
    for delta in (-1,0,1):
        for fps in (0,30,0x80000001):
            for age in (0,3*fps,3*fps+1):
                c=add(0,f'clock/{frame}/{delta}/{fps}/{age}',frame=frame,fps=fps);c['objects'][0][4]=(frame-age)&0xffffffff;c['objects'][0][5]=(frame+delta)&0xffffffff
for cache in (0,17,0xffffffff):
    for lookup in (0,17,0x80000000):
        for spawn in (-1,1,0):add(0,f'cache/{cache}/{lookup}/{spawn}',cache=cache,lookup=lookup,spawn_ref=spawn)
for kind in (1,2,3):
    for status in (0,1,2,4,8,6,10,12,14,0x102,0x104,0x108,0x80000002,0xffffffff):
        for flags in (0,0x8000,0xffffffff):
            for delta in (-1,0,1):
                c=add(kind,f'state/{status}/{flags}/{delta}');c['objects'][0][0]=flags;c['objects'][0][1]=status;c['objects'][0][5]=10+delta
    for frame,fps in ((0,0),(0xffffffff,1),(0xfffffffc,30),(0x80000000,0x80000001),(10,0xffffffff)):
        add(kind,f'wrap/{frame}/{fps}',frame=frame,fps=fps)
for kind,services in [(0,('lookup','new','create','random','delete')),(1,('unset','set','remove')),(2,('unset','set','remove')),(3,('unset','set','remove','delete'))]:
    for service in services:
        for action in ('clock','fps','owner','owner-nil','position','created','deadline','cache','flags','status'):
            c=add(kind,f'mutate/{service}/{action}',mutation=service+':'+action)
            if service in ('remove','delete') and kind!=0:c['objects'][0][1]=8
            if service=='delete' and kind==0:c['objects'][0][4]=0xfffffff0;c['fps']=0
def prepare(c,source_ref=0):
    global applied
    u.context_restore(initial);u.mem_write(heap,b'\xa5'*0x10000);u.mem_write(stack,bytes(0x10000))
    for n,a in enumerate(objects):
        for off,value in zip(offsets,c['objects'][n]):put(a+off,value)
        put(a+508,ref(c['owners'][n]))
    put(frame_addr,c['frame']);put(fps_addr,c['fps']);put(cache_addr,c['cache']);put(sp,stop);put(sp+4,ref(source_ref))
    u.reg_write(UC_X86_REG_ESP,sp);events.clear();faults.clear();applied=False
def preserved(before):
    expected=bytearray(before)
    for a in objects:
        for off in offsets+(508,):expected[a-heap+off:a-heap+off+4]=u.mem_read(a+off,4)
    return bytes(expected)==bytes(u.mem_read(heap,0x10000))
rows=[]
for current in cases:
    prepare(current);before=bytes(u.mem_read(heap,0x10000));u.emu_start(specs[current['kind']][1],stop,count=20000)
    assert not faults and u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4 and u.reg_read(UC_X86_REG_FPCW)==0xe3f
    assert preserved(before);rows.append(dict(input=copy.deepcopy(current),output=state(),events=copy.deepcopy(events),nonoutput_bytes_preserved=True))
assert args.game.read_bytes()==blob
proof=dict(original_sha256=sha(blob),original_file_unchanged=True,original_startup_FPCW=0xe3f,contexts=len(cases),executions=len(rows),regions=regions,entries=dict(entries),rows=rows,faults=[],executed_instructions=[dict(address=hex(a),bytes=b) for a,b in sorted(trace.items())],arithmetic_or_conversion_substitutions=0,scope='Four complete original callers with declared lookup/allocation/create/random/status/remove/delete bodies. No floating arithmetic occurs; point words are copied. EAX residue of void update callbacks is not an output pointer. Full native original-game/world composition is separate.')
with args.proof.open('x') as f:json.dump(proof,f,separators=(',',':'))
with args.fixture.open('x') as f:json.dump({k:v for k,v in proof.items() if k!='executed_instructions'},f,separators=(',',':'))
fault_rows=[]
for kind in range(4):
    current=copy.deepcopy(cases[0]);current.update(kind=kind,name=specs[kind][0]+'/nil-source',mutation='');prepare(current,-1);before=bytes(u.mem_read(heap,0x10000));error=None
    try:u.emu_start(specs[kind][1],stop,count=20000)
    except Exception as exc:error=str(exc)
    assert len(faults)==1 and not events and preserved(before)
    fault_rows.append(dict(input=current,source_ref=-1,fault=True,faults=list(faults),error=error,events=list(events),output=state(),nonoutput_bytes_preserved=True))
with args.fault_proof.open('x') as f:json.dump(dict(original_sha256=sha(blob),rows=fault_rows),f,separators=(',',':'))
print(json.dumps({k:proof[k] for k in ('contexts','executions','entries','faults')}))
