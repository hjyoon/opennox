"""Execute three complete unchanged GAME.EXE update callers."""
import argparse,collections,copy,hashlib,json,struct
from pathlib import Path
from unicorn import Uc,UC_ARCH_X86,UC_MODE_32,UC_HOOK_CODE,UC_HOOK_MEM_INVALID
from unicorn.x86_const import *
p=argparse.ArgumentParser(description=__doc__)
for n in ('game','fixture','proof','fault-proof'):p.add_argument('--'+n,type=Path,required=True)
args=p.parse_args();blob=args.game.read_bytes();sha=lambda b:hashlib.sha256(b).hexdigest()
assert sha(blob)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,60)[0];opt=pe+24;base=struct.unpack_from('<I',blob,opt+28)[0];size=(struct.unpack_from('<I',blob,opt+56)[0]+4095)&~4095;tab=opt+struct.unpack_from('<H',blob,pe+20)[0]
u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,size)
for n in range(struct.unpack_from('<H',blob,pe+6)[0]):
    _,va,count,off=struct.unpack_from('<IIII',blob,tab+40*n+8)
    if count:u.mem_write(base+va,blob[off:off+count])
stack,heap,stop=0x20000000,0x21000000,0x30000000
for a in (stack,heap,stop):u.mem_map(a,0x10000)
sp=stack+0xe000;objects=[heap+128+1024*n for n in range(3)];data=[heap+0x4000,heap+0x4100];callbacks=[stop+0x1000,stop+0x1010]
get=lambda a:struct.unpack('<I',u.mem_read(a,4))[0]
put=lambda a,v:u.mem_write(a,struct.pack('<I',v&0xffffffff))
ref=lambda rows,n:0 if n<0 else rows[n]
idx=lambda rows,a:-1 if a==0 else rows.index(a)
frame_addr,fps_addr=0x84ea04,0x85b3fc;caches=[0x833efc,0x833ef8]
specs=[('toggle',0x53b060,332),('trigger',0x53b1b0,327),('switch',0x53b320,91)]
regions=[dict(name=n,address=hex(a),size=z,sha256=sha(bytes(u.mem_read(a,z)))) for n,a,z in specs]
events=[];faults=[];trace={};entries=collections.Counter();current={};applied=False;recording=False
def state():
    words=[]
    for a in data:
        row=[get(a+4*k) for k in range(15)];row[1]=0;words.append(row)
    return dict(type=struct.unpack('<H',u.mem_read(objects[0]+4,2))[0],flags=get(objects[0]+16),anim=get(objects[0]+132),next=get(objects[0]+136),sync=get(objects[0]+152),update_ref=idx(data,get(objects[0]+748)),collide_ref=idx(callbacks,get(objects[0]+696)),records=words,callers=[idx(objects,get(a+4)) for a in data],frame=get(frame_addr),fps=get(fps_addr),caches=[get(a) for a in caches])
def event(name,values):events.append(dict(service=name,args=values,**state()))
def mutate(service):
    global applied
    key=current['mutation']
    if applied or not key.startswith(service+':'):return
    applied=True;action=key.split(':',1)[1]
    if action=='clock':put(frame_addr,0xfffffffc)
    elif action=='fps':put(fps_addr,0x80000001)
    elif action=='flags':put(objects[0]+16,0x1000041)
    elif action=='anim':put(objects[0]+132,99)
    elif action=='deadline':put(objects[0]+136,0xfffffffe)
    elif action=='update':put(objects[0]+748,data[1])
    elif action=='caller':put(data[0]+4,objects[2])
    elif action=='caller-nil':put(data[0]+4,0)
    elif action=='request':put(data[0],get(data[0])^1)
    elif action=='one-shot':put(data[0],get(data[0])|2)
    elif action=='mode':u.mem_write(data[0]+8,b'\x01')
    elif action=='sound':put(data[0]+36,99);put(data[0]+40,100)
    elif action=='script':put(data[0]+20,0x77777777);put(data[0]+28,0x88888888)
    elif action=='cache':put(caches[0],99);put(caches[1],100)
    elif action=='type':u.mem_write(objects[0]+4,struct.pack('<H',18))
    elif action=='collide':put(objects[0]+696,0)
    else:raise AssertionError(key)
def ret(v=0):
    esp=u.reg_read(UC_X86_REG_ESP);u.reg_write(UC_X86_REG_EAX,v&0xffffffff);u.reg_write(UC_X86_REG_EIP,get(esp));u.reg_write(UC_X86_REG_ESP,esp+4)
def hook(_u,address,length,_):
    if address in (stop,0x4014c4):u.emu_stop();return
    if not recording:return
    for n,a,z in specs:
        if a<=address<a+z:
            trace[address]=bytes(u.mem_read(address,length)).hex()
            if address==a:entries[n]+=1
            return
    esp=u.reg_read(UC_X86_REG_ESP);arg=lambda n:get(esp+4*n)
    if address==0x4e3aa0:
        name=bytes(u.mem_read(arg(1),14)).split(b'\0')[0];which={b'Trigger':0,b'PressurePlate':1}[name];event('lookup',[which]);mutate('lookup');ret(current['lookup'][which])
    elif address==0x4e4880:
        assert arg(1)==objects[0];event('anim',[0,arg(2)]);put(objects[0]+132,arg(2));put(objects[0]+152,0xffffffff);mutate('anim');ret()
    elif address==0x501960:
        assert [arg(2),arg(3),arg(4)]==[objects[0],0,0];event('audio',[arg(1),0,0,0]);mutate('audio');ret()
    elif address==0x502490:
        # The original caller pushes exactly block/caller/trigger, not the
        # later port's event-code argument. Compare these three real inputs.
        block=arg(1);n=next(i for i,a in enumerate(data) if a<=block<a+60);off=block-data[n]
        assert off in (20,28) and arg(3)==objects[0];event('script',[n,off,idx(objects,arg(2)),0]);mutate('script');ret(0x12345678)
    elif address==0x4e44f0:
        assert arg(1)==objects[0];event('sync',[0]);put(objects[0]+152,0xffffffff);mutate('sync');ret()
    elif address==0x537610:
        assert arg(1)==objects[0];event('queue',[0]);mutate('queue');ret(current['queue'])
    else:raise AssertionError((hex(address),current['name']))
def invalid(_u,access,address,length,value,_):
    faults.append(dict(address=address,size=length,instruction=u.reg_read(UC_X86_REG_EIP)));return False
u.hook_add(UC_HOOK_CODE,hook);u.hook_add(UC_HOOK_MEM_INVALID,invalid)
u.reg_write(UC_X86_REG_ESP,sp);u.reg_write(UC_X86_REG_FPCW,0x27f);u.emu_start(0x4014b5,0x4014c4,count=20000)
assert u.reg_read(UC_X86_REG_FPCW)==0xe3f
initial=u.context_save();recording=True;cases=[]
def add(kind,name,**kw):
    c=dict(name=specs[kind][0]+'/'+name,kind=kind,type=19,flags=0x1000000,anim=0,next=9,sync=7,update_ref=0,collide_ref=0,records=[[9,0,0xaabbcc00,11,12,21,22,31,32,101,202,44,48,52,56],[10,0,0x11223305,13,14,23,24,33,34,103,204,45,49,53,57]],callers=[1,2],frame=10,fps=30,caches=[17,18],lookup=[17,18],queue=0x81,mutation='')
    c.update(kw);cases.append(c);return c
for kind in (0,1):
    for enabled in (0,0x1000000):
        for flags in (0,1,2,3,8,9,10,11,0x80000009,0xffffffff):
            for mode in (0,1,3,5,7):
                for delta in (-1,0,1):
                    c=add(kind,f'state/{enabled}/{flags}/{mode}/{delta}',flags=enabled,next=10+delta);c['records'][0][0]=flags;c['records'][0][2]=0xaabbcc00|mode
    for frame in (0,0x80000000,0xffffffff):
        for fps in (0,1,0x80000001,0xffffffff):add(kind,f'wrap/{frame}/{fps}',frame=frame,fps=fps)
for cache in ([0,0],[17,18],[0,99],[0x10011,18],[0xffffffff,0xffffffff]):
    for lookup in ([0,0],[17,18],[0x10011,0x10012]):
        for typ in (0,17,18,19):add(1,f'cache/{cache}/{lookup}/{typ}',caches=cache.copy(),lookup=lookup.copy(),type=typ)
for flags in (0,0x40,0x1000000,0x1000040,0xffffffff):
    for collide in (-1,0,1):
        for anim in (0,1,0xffffffff):
            for queue in (0,1,0x80,0xff):add(2,f'state/{flags}/{collide}/{anim}/{queue}',flags=flags,collide_ref=collide,anim=anim,queue=queue)
actions=('clock','fps','flags','anim','deadline','update','caller','caller-nil','request','one-shot','mode','sound','script','cache','type','collide')
for kind in (0,1):
    for service in ('lookup','anim','audio','script'):
        for action in actions:
            for mode in (0,1):
                c=add(kind,f'mutate/{service}/{action}/{mode}',mutation=service+':'+action,caches=[0,0] if service=='lookup' else [17,18]);c['records'][0][2]=0xaabbcc00|mode;c['records'][0][0]=8 if kind==1 and mode==1 else 9
for service in ('sync','queue'):
    for action in actions:add(2,f'mutate/{service}/{action}',mutation=service+':'+action,flags=0 if service=='sync' else 0x1000040,anim=0)
def prepare(c,source_ref=0):
    global applied
    u.context_restore(initial);u.mem_write(heap,b'\xa5'*0x10000);u.mem_write(stack,bytes(0x10000))
    u.mem_write(objects[0]+4,struct.pack('<H',c['type']))
    for off,key in ((16,'flags'),(132,'anim'),(136,'next'),(152,'sync')):put(objects[0]+off,c[key])
    put(objects[0]+748,ref(data,c['update_ref']));put(objects[0]+696,ref(callbacks,c['collide_ref']))
    for n,a in enumerate(data):
        for k,v in enumerate(c['records'][n]):put(a+4*k,v)
        put(a+4,ref(objects,c['callers'][n]))
    for a,v in zip(caches,c['caches']):put(a,v)
    put(frame_addr,c['frame']);put(fps_addr,c['fps']);put(sp,stop);put(sp+4,ref(objects,source_ref))
    u.reg_write(UC_X86_REG_ESP,sp);events.clear();faults.clear();applied=False
def preserved(before):
    expected=bytearray(before)
    for off,z in ((4,2),(16,4),(132,4),(136,4),(152,4),(748,4),(696,4)):expected[objects[0]-heap+off:objects[0]-heap+off+z]=u.mem_read(objects[0]+off,z)
    for a in data:expected[a-heap:a-heap+60]=u.mem_read(a,60)
    return bytes(expected)==bytes(u.mem_read(heap,0x10000))
rows=[]
for current in cases:
    prepare(current);before=bytes(u.mem_read(heap,0x10000));u.emu_start(specs[current['kind']][1],stop,count=20000)
    assert not faults and u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4 and preserved(before)
    rows.append(dict(input=copy.deepcopy(current),output=state(),events=copy.deepcopy(events),result=u.reg_read(UC_X86_REG_EAX)&255,nonoutput_bytes_preserved=True))
assert args.game.read_bytes()==blob
proof=dict(original_sha256=sha(blob),original_file_unchanged=True,original_startup_FPCW=0xe3f,contexts=len(cases),executions=len(rows),regions=regions,entries=dict(entries),rows=rows,faults=[],executed_instructions=[dict(address=hex(a),bytes=b) for a,b in sorted(trace.items())],arithmetic_or_conversion_substitutions=0,scope='Three complete original callers with declared lookup/animation/audio/script/sync/queue services. Script compares its three actual stack arguments; Go event codes are separate port metadata. Transient Field4 pointers are normalized into caller identities. Native sidecar/world composition is separate.')
with args.proof.open('x') as f:json.dump(proof,f,separators=(',',':'))
with args.fixture.open('x') as f:json.dump({k:v for k,v in proof.items() if k!='executed_instructions'},f,separators=(',',':'))
fault_rows=[]
for kind in range(3):
    for name,source_ref,update_ref,cold in [('nil-source',-1,0,False),('nil-source-cold',-1,0,True),('nil-update',0,-1,False),('nil-update-cold',0,-1,True)]:
        current=copy.deepcopy(cases[0]);current.update(name=specs[kind][0]+'/'+name,kind=kind,update_ref=update_ref,caches=[0,0] if cold else [17,18],mutation='')
        prepare(current,source_ref);before=bytes(u.mem_read(heap,0x10000));error=None
        try:u.emu_start(specs[kind][1],stop,count=20000)
        except Exception as exc:error=str(exc)
        fault=source_ref<0 or kind!=2
        assert bool(faults)==fault and preserved(before)
        fault_rows.append(dict(input=current,source_ref=source_ref,fault=fault,faults=list(faults),error=error,events=copy.deepcopy(events),output=state(),result=None if fault else u.reg_read(UC_X86_REG_EAX)&255,nonoutput_bytes_preserved=True))
with args.fault_proof.open('x') as f:json.dump(dict(original_sha256=sha(blob),rows=fault_rows),f,separators=(',',':'))
print(json.dumps({k:proof[k] for k in ('contexts','executions','entries','faults')}))
