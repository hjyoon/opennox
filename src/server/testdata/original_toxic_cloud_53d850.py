"""Execute four complete unchanged GAME.EXE cloud roots and target callbacks."""
from pathlib import Path
import argparse, collections, copy, hashlib, json, struct
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import *

parser=argparse.ArgumentParser(description=__doc__)
for key in ('game','fixture','proof'):parser.add_argument('--'+key,type=Path,required=True)
parser.add_argument('--fault-proof',type=Path)
args=parser.parse_args();blob=args.game.read_bytes()
sha=lambda b:hashlib.sha256(b).hexdigest()
assert sha(blob)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,60)[0];opt=pe+24;base=struct.unpack_from('<I',blob,opt+28)[0]
size=(struct.unpack_from('<I',blob,opt+56)[0]+4095)&~4095;tab=opt+struct.unpack_from('<H',blob,pe+20)[0]
u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,size)
for n in range(struct.unpack_from('<H',blob,pe+6)[0]):
    _,va,count,off=struct.unpack_from('<IIII',blob,tab+40*n+8)
    if count:u.mem_write(base+va,blob[off:off+count])
stack,heap,stop=0x20000000,0x21000000,0x30000000
for p in (stack,heap,stop):u.mem_map(p,0x10000)
sp=stack+0xe000;continuation=stop+0x400
get=lambda p:struct.unpack('<I',u.mem_read(p,4))[0]
put=lambda p,v:u.mem_write(p,struct.pack('<I',v&0xffffffff))
f32=lambda v:struct.unpack('<I',struct.pack('<f',v))[0]
s32=lambda v:struct.unpack('<i',struct.pack('<I',v&0xffffffff))[0]
objects=[heap+128+1024*n for n in range(7)]
data=[heap+0x4000,heap+0x4100];callbacks=[stop+0x5000,stop+0x5010]
offsets=(8,16,56,60,136);frame_address=0x84ea04
specs=[('root',0x53d850,110),('target',0x53d8c0,156),('small-root',0x53d960,110),('small-target',0x53d9d0,131)]
regions=[dict(name=n,address=hex(p),size=z,sha256=sha(bytes(u.mem_read(p,z)))) for n,p,z in specs]
ref=lambda n:0 if n<0 else objects[n]
objref=lambda p:-1 if p==0 else objects.index(p)
dataref=lambda p:-1 if p==0 else data.index(p)
fnref=lambda p:-1 if p==0 else callbacks.index(p)
def state():
    return dict(objects=[[get(p+off) for off in offsets] for p in objects],owners=[objref(get(p+508)) for p in objects],update_refs=[dataref(get(p+748)) for p in objects],callbacks=[fnref(get(p+716)) for p in objects],data=[get(p) for p in data],frame=get(frame_address))
events=[];faults=[];trace={};entries=collections.Counter();recording=False;current={};applied=False;scan=None
def event(name,values):events.append(dict(service=name,args=values,**state()))
def ret(value=0):
    esp=u.reg_read(UC_X86_REG_ESP);u.reg_write(UC_X86_REG_EAX,value&0xffffffff);u.reg_write(UC_X86_REG_EIP,get(esp));u.reg_write(UC_X86_REG_ESP,esp+4)
def mutate(service,pointer=0):
    global applied
    key=current['mutation']
    if applied or not key.startswith(service+':'):return
    applied=True;action=key.split(':',1)[1]
    if action=='position':put(objects[0]+56,f32(500.1));put(objects[0]+60,f32(-600.3))
    elif action=='target-position':put(objects[1]+56,f32(611.3));put(objects[1]+60,f32(-817.7))
    elif action=='owner':put(objects[0]+508,objects[6])
    elif action=='owner-nil':put(objects[0]+508,0)
    elif action=='update':put(objects[0]+748,data[1])
    elif action=='duration':put(data[0],9)
    elif action=='expire':put(data[0],0)
    elif action=='clock':put(frame_address,0xfffffffc)
    elif action=='deadline':put(objects[0]+136,0)
    elif action=='callback':put(objects[1]+716,callbacks[1])
    elif action=='target-class':put(objects[1]+8,0)
    elif action=='point':put(pointer,f32(711.3));put(pointer+4,f32(-917.7))
    elif action=='trace-point':put(pointer+8,f32(17.3));put(pointer+12,f32(-91.7))
    else:raise AssertionError(key)
def next_visit():
    global scan
    u.reg_write(UC_X86_REG_ESP,scan['esp'])
    if scan['remaining']:
        target=scan['remaining'].pop(0);esp=scan['esp']-12
        put(esp,continuation);put(esp+4,ref(target));put(esp+8,scan['cloud'])
        u.reg_write(UC_X86_REG_ESP,esp);u.reg_write(UC_X86_REG_EIP,scan['callback'])
    else:
        scan=None;ret()
def hook(_u,address,length,_):
    global scan
    if address in (stop,0x4014c4):u.emu_stop();return
    if not recording:return
    if address==continuation:next_visit();return
    for name,lo,z in specs:
        if lo<=address<lo+z:
            trace[address]=bytes(u.mem_read(address,length)).hex()
            if address==lo:entries[name]+=1
            return
    esp=u.reg_read(UC_X86_REG_ESP);arg=lambda n:get(esp+4*n)
    if address==0x517f90:
        assert scan is None and arg(1)==objects[0]+56 and arg(4)==objects[0]
        small=current['kind']==2;callback=specs[3 if small else 1][1]
        assert arg(2)==f32(35 if small else 75) and arg(3)==callback
        p=arg(1);event('circle',[get(p),get(p+4),arg(2),3 if small else 1,0]);mutate('circle',p)
        scan=dict(esp=esp,cloud=arg(4),callback=callback,remaining=list(current['visits']));next_visit()
    elif address==0x415fa0:
        bounds=[s32(arg(1)),s32(arg(2))];name='damage-random' if bounds==[3,10] else 'delay'
        assert bounds in ([3,10],[5,10]);event(name,bounds);mutate(name);ret(current['damage'] if name=='damage-random' else current['delay'])
    elif address==0x535250:
        p=arg(1);assert [arg(2),arg(3),arg(4)]==[0,0,5]
        event('trace',[get(p+4*n) for n in range(4)]+[0,0,5]);mutate('trace',p);ret(current['trace_result'])
    elif address==0x4ec580:
        assert arg(1)==objects[0];owner=get(objects[0]+508)
        event('owner',[0]);mutate('owner');ret(owner)
    elif address in callbacks:
        assert arg(3)==objects[0] and arg(5)==5
        event('damage',[callbacks.index(address),objref(arg(1)),objref(arg(2)),0,s32(arg(4)),5]);mutate('damage');ret(0x1234abcd)
    elif address==0x5330c0:
        event('enemy',[objref(arg(1)),objref(arg(2))]);mutate('enemy');ret(current['enemy_result'])
    elif address==0x4ee7e0:
        assert arg(2)==1 and arg(3)==1;event('poison',[objref(arg(1)),1,1]);mutate('poison');ret(1)
    elif address==0x4e5cc0:
        assert arg(1)==objects[0];event('delete',[0]);mutate('delete');ret()
    else:raise AssertionError((hex(address),current['name']))
def invalid(_u,access,address,length,value,_):
    faults.append(dict(address=address,size=length,instruction=u.reg_read(UC_X86_REG_EIP)));return False
u.hook_add(UC_HOOK_CODE,hook);u.hook_add(UC_HOOK_MEM_INVALID,invalid)
u.reg_write(UC_X86_REG_ESP,sp);u.reg_write(UC_X86_REG_FPCW,0x27f);u.emu_start(0x4014b5,0x4014c4,count=20000)
assert u.reg_read(UC_X86_REG_FPCW)==0xe3f
initial=u.context_save();recording=True;cases=[]
def add(kind,name,**kw):
    value=dict(name=name,kind=kind,objects=[[0,0,f32(10.1),f32(20.3),9],[4,0,f32(30.1),f32(40.3),0],[2,0,f32(50.1),f32(60.3),0],[0,0,f32(70.1),f32(80.3),0],[6,0,f32(90.1),f32(100.3),0],[4,0,0,0,0],[4,0,0,0,0]],owners=[5,-1,-1,-1,-1,-1,-1],update_refs=[0,-1,-1,-1,-1,-1,-1],callbacks=[0,0,0,0,0,0,0],data=[2,99],frame=10,visits=[1,2,3,4],trace_result=1,enemy_result=1,damage=7,delay=6,mutation='')
    value.update(kw);cases.append(value);return value
for kind in (0,2):
    for frame in (0,1,10,0x7fffffff,0x80000000,0xfffffffc,0xffffffff):
        for delta in (-1,0,1):
            for duration in (0,1,2,0x7fffffff,0x80000000,0xffffffff):
                c=add(kind,f'{specs[kind][0]}/clock/{frame}/{delta}/{duration}',frame=frame,data=[duration,99]);c['objects'][0][4]=(frame+delta)&0xffffffff
for kind in (1,3):
    for cls in (0,1,2,4,6,0x100,0x200004,0xffffffff):
        for visible in (0,1,0x100,-1):
            for enemy in (0,1,0x100,-1):
                c=add(kind,f'{specs[kind][0]}/gate/{cls}/{visible}/{enemy}',trace_result=visible,enemy_result=enemy);c['objects'][1][0]=cls
for kind in range(4):
    for service in ('circle','trace','damage-random','owner','damage','enemy','poison','delay','delete'):
        for action in ('position','target-position','owner','owner-nil','update','duration','expire','clock','deadline','callback','target-class'):
            add(kind,f'{specs[kind][0]}/mutation/{service}/{action}',mutation=service+':'+action,visits=[1,1,2,1])
    if kind in (0,2):add(kind,f'{specs[kind][0]}/circle-point-alias',mutation='circle:point')
    add(kind,f'{specs[kind][0]}/trace-point-write',mutation='trace:trace-point')
for kind in range(4):
    for word in (0,0x80000000,1,0x80000001,0x7f800000,0xff800000,0x7fc12345,0xffc12345):
        c=add(kind,f'{specs[kind][0]}/point-bits/{word:x}');c['objects'][0][2:4]=[word,word];c['objects'][1][2:4]=[word,word]
for kind in (0,2):
    for visits in ([],[3],[1],[1,1,1],[0,1,4]):add(kind,f'{specs[kind][0]}/visits/'+','.join(map(str,visits)),visits=visits)
def prepare(c,source_ref=0,target_ref=1):
    global applied,scan
    u.context_restore(initial);u.mem_write(heap,b'\xa5'*0x10000);u.mem_write(stack,bytes(0x10000))
    for n,p in enumerate(objects):
        for off,value in zip(offsets,c['objects'][n]):put(p+off,value)
        put(p+508,ref(c['owners'][n]));put(p+748,0 if c['update_refs'][n]<0 else data[c['update_refs'][n]])
        put(p+716,0 if c['callbacks'][n]<0 else callbacks[c['callbacks'][n]])
    for p,value in zip(data,c['data']):put(p,value)
    put(frame_address,c['frame']);put(sp,stop)
    if c['kind'] in (0,2):put(sp+4,ref(source_ref))
    else:put(sp+4,ref(target_ref));put(sp+8,ref(source_ref))
    u.reg_write(UC_X86_REG_ESP,sp);events.clear();faults.clear();applied=False;scan=None
def preserved(before):
    expected=bytearray(before)
    for p in objects:
        for off in offsets+(508,716,748):expected[p-heap+off:p-heap+off+4]=u.mem_read(p+off,4)
    for p in data:expected[p-heap:p-heap+4]=u.mem_read(p,4)
    return bytes(expected)==bytes(u.mem_read(heap,0x10000))
rows=[]
for current in cases:
    prepare(current);before=bytes(u.mem_read(heap,0x10000))
    u.emu_start(specs[current['kind']][1],stop,count=20000)
    assert not faults and u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4 and u.reg_read(UC_X86_REG_FPCW)==0xe3f
    assert preserved(before)
    rows.append(dict(input=copy.deepcopy(current),output=state(),events=copy.deepcopy(events),nonoutput_bytes_preserved=True))
assert args.game.read_bytes()==blob
proof=dict(original_sha256=sha(blob),original_file_unchanged=True,original_startup_FPCW=0xe3f,contexts=len(cases),executions=len(rows),regions=regions,entries=dict(entries),rows=rows,faults=[],executed_instructions=[dict(address=hex(p),bytes=b) for p,b in sorted(trace.items())],arithmetic_or_conversion_substitutions=0,scope='Four complete original routines, including original target callbacks during declared circle enumeration. Declared trace/random/owner/damage/enemy/poison/delete bodies and prepared mutations. Native original-game/full-world bodies and general signaling-NaN arithmetic are not inferred. No arithmetic/conversion instructions in these four bodies require a rounding counterfactual.')
with args.proof.open('x') as f:json.dump(proof,f,separators=(',',':'))
with args.fixture.open('x') as f:json.dump({k:v for k,v in proof.items() if k!='executed_instructions'},f,separators=(',',':'))
if args.fault_proof:
    contexts=[]
    for kind in (0,2):
        for name,source_ref,next_frame,missing in [('nil-cloud',-1,11,False),('waiting-nil-update',0,11,True),('active-nil-update',0,9,True)]:
            c=copy.deepcopy(cases[0]);c.update(kind=kind,name=name,frame=10,visits=[],mutation='');c['objects'][0][4]=next_frame
            if missing:c['update_refs'][0]=-1
            contexts.append((c,source_ref,1,True))
    for kind in (1,3):
        for name,source_ref,target_ref,cls,trace_result,callback,fault in [
            ('nil-target',0,-1,4,1,0,True),('filtered-nil-cloud',-1,1,0,1,0,False),('unit-nil-cloud',-1,1,4,1,0,True),('blocked-nil-callback',0,1,4,0,-1,False),('live-nil-callback',0,1,4,1,-1,True)]:
            c=copy.deepcopy(cases[0]);c.update(kind=kind,name=name,frame=10,trace_result=trace_result,mutation='');c['objects'][1][0]=cls;c['callbacks'][1]=callback
            contexts.append((c,source_ref,target_ref,fault))
    results=[]
    for current,source_ref,target_ref,fault in contexts:
        prepare(current,source_ref,target_ref);before=bytes(u.mem_read(heap,0x10000));error=None
        try:u.emu_start(specs[current['kind']][1],stop,count=20000)
        except Exception as exc:error=str(exc)
        assert bool(faults)==fault,(current['name'],faults)
        assert preserved(before)
        results.append(dict(input=copy.deepcopy(current),source_ref=source_ref,target_ref=target_ref,fault=fault,error=error,faults=list(faults),events=copy.deepcopy(events),output=state(),nonoutput_bytes_preserved=True))
    assert args.game.read_bytes()==blob
    with args.fault_proof.open('x') as f:json.dump(dict(original_sha256=sha(blob),rows=results,faults=sum(v['fault'] for v in results),safe_controls=sum(not v['fault'] for v in results)),f,separators=(',',':'))
print(json.dumps({k:proof[k] for k in ['contexts','executions','entries','faults']}))
