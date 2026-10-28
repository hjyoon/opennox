"""Run complete unchanged Spark/ProjectileTrail/Push callers and FISTP helper."""
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
sp=stack+0xe000;objects=[heap+128+1024*n for n in range(3)];data=[heap+0x4000,heap+0x4100]
get=lambda a:struct.unpack('<I',u.mem_read(a,4))[0]
put=lambda a,v:u.mem_write(a,struct.pack('<I',v&0xffffffff))
f32=lambda v:struct.unpack('<I',struct.pack('<f',v))[0]
f64=lambda v:struct.unpack('<Q',struct.pack('<d',v))[0]
s16=lambda v:struct.unpack('<h',struct.pack('<H',v&0xffff))[0]
ref=lambda rows,n:0 if n<0 else rows[n]
idx=lambda rows,a:-1 if a==0 else rows.index(a)
offsets=(56,60,72,76,88,92,112,544)
specs=[('spark',0x53adc0,53),('trail',0x53aec0,355),('push',0x53b030,38),('convert',0x419a70,17)]
regions=[dict(name=n,address=hex(a),size=z,sha256=sha(bytes(u.mem_read(a,z)))) for n,a,z in specs]
table=dict(address='0x5b6658',size=2048,sha256=sha(bytes(u.mem_read(0x5b6658,2048))),words=[get(0x5b6658+4*n) for n in range(512)])
stub,value_addr=stop+0x100,stop+0x2000;u.mem_write(stub,b'\xdd\x05'+struct.pack('<I',value_addr)+b'\xc3')
events=[];faults=[];trace={};entries=collections.Counter();current={};applied=False;recording=False;random_index=0;created=0
def state():
    return dict(objects=[[get(a+off) for off in offsets] for a in objects],direction=struct.unpack('<H',u.mem_read(objects[0]+124,2))[0],update_ref=idx(data,get(objects[0]+748)),data=[[get(a+4*n) for n in range(4)] for a in data],random_index=random_index,created=created)
def event(name,values):events.append(dict(service=name,args=values,**state()))
def mutate(service,number,pointer=0):
    global applied
    key=current['mutation']
    if applied or not key.startswith(service+':') or current['at']!=number:return
    applied=True;action=key.split(':',1)[1];a=objects[0]
    if action=='position':put(a+56,f32(100.1));put(a+60,f32(-200.3))
    elif action=='previous':put(a+72,f32(31.7));put(a+76,f32(-19.9))
    elif action=='force':put(a+88,f32(3.1));put(a+92,f32(-7.3))
    elif action=='speed':put(a+544,f32(99.1))
    elif action=='direction':u.mem_write(a+124,struct.pack('<H',127))
    elif action=='update':put(a+748,data[1])
    elif action=='life':put(data[0]+4,0xffffffff)
    elif action=='kind':put(data[0]+12,4)
    elif action=='radius':put(data[0],f32(99.1))
    elif action=='power':put(data[0]+8,f32(-33.3))
    elif action=='point':put(pointer,f32(17.3));put(pointer+4,f32(-29.7))
    else:raise AssertionError(key)
def ret(v=0):
    esp=u.reg_read(UC_X86_REG_ESP);u.reg_write(UC_X86_REG_EAX,v&0xffffffff);u.reg_write(UC_X86_REG_EIP,get(esp));u.reg_write(UC_X86_REG_ESP,esp+4)
def hook(_u,address,length,_):
    global random_index,created
    if address in (stop,0x4014c4):u.emu_stop();return
    if not recording or stub<=address<stub+7:return
    for n,a,z in specs:
        if a<=address<a+z:
            trace[address]=bytes(u.mem_read(address,length)).hex()
            if address==a:entries[n]+=1
            return
    esp=u.reg_read(UC_X86_REG_ESP);arg=lambda n:get(esp+4*n)
    if address==0x4e5cc0:
        assert arg(1)==objects[0];event('delete',[0]);mutate('delete',0);ret()
    elif address==0x416030:
        number=random_index;assert [arg(1),arg(2)]==[f32(-2 if number%4<2 else -4),f32(2 if number%4<2 else 4)]
        event('random',[arg(1),arg(2),number]);mutate('random',number)
        u.mem_write(value_addr,struct.pack('<Q',current['random'][number]));random_index+=1;u.reg_write(UC_X86_REG_EIP,stub)
    elif address==0x54fd80:
        assert arg(3)==1 and arg(4)==6 and arg(7)==0 and arg(8)==objects[0]
        number=created;event('create',[arg(1),arg(2),1,6,arg(5),arg(6),0,0]);mutate('create',number);created+=1;ret(ref(objects,current['returns'][number]))
    elif address==0x52e040:
        assert arg(1)==objects[0]+56 and [arg(3),arg(5),arg(6),arg(7)]==[0,0,0,0]
        ptr=arg(1);event('push',[0,get(ptr),get(ptr+4),arg(2),0,arg(4),0,0,0]);mutate('push',0,ptr);ret()
    else:raise AssertionError((hex(address),current['name']))
def invalid(_u,access,address,length,value,_):
    faults.append(dict(address=address,size=length,instruction=u.reg_read(UC_X86_REG_EIP)));return False
u.hook_add(UC_HOOK_CODE,hook);u.hook_add(UC_HOOK_MEM_INVALID,invalid)
u.reg_write(UC_X86_REG_ESP,sp);u.reg_write(UC_X86_REG_FPCW,0x27f);u.emu_start(0x4014b5,0x4014c4,count=20000)
assert u.reg_read(UC_X86_REG_FPCW)==0xe3f
initial=u.context_save();recording=True;cases=[]
def add(kind,name,**kw):
    c=dict(name=specs[kind][0]+'/'+name,kind=kind,objects=[[f32(123.7),f32(-77.1),f32(119.3),f32(-74.9),f32(1.1),f32(-2.3),f32(0.5),f32(6.1)]]+[[0]*8 for _ in range(2)],direction=13,update_ref=0,data=[[10,2,0x33445566,4],[20,30,0x778899aa,7]],random=[f64(x) for _ in range(8) for x in (-1.9,1.1,-3.7,3.9)],returns=[1,2,1,2,1,2,1,2],mutation='',at=0)
    c.update(kw)
    if kind==2:c['data']=[[f32(81.25),f32(912.5),f32(-17.75),0x12345678],[f32(33.3),0,f32(5.7),0x56789abc]]
    direction=s16(c['direction']);c['direction_words']=[get(0x5b6658+direction*8+4*n) for n in range(2)]
    cases.append(c);return c
for life in (0,1,2,0x7fffffff,0x80000000,0xffffffff):
    for kind in (0,1,4,0x104,0xffffffff):
        c=add(0,f'life/{life}/{kind}');c['data'][0][1]=life;c['data'][0][3]=kind
for direction in list(range(256))+[-1,256,-32768,32767]:add(1,f'direction/{direction}',direction=direction&0xffff)
for bits in (0,0x80000000,1,0x80000001,0x3f800001,0xbf800001,0x4b000001,0x4b800001,0x4effffff,0x4f000000,0xcf000000,0x7f7fffff,0xff7fffff,0x7f800000,0xff800000,0x7fc12345):
    for field in range(8):
        c=add(1,f'word/{field}/{bits:x}');c['objects'][0][field]=bits
for sequence in [(0.0,0.0,0.0,0.0),(-0.0,0.0,-0.0,0.0),(1.999999999, -1.999999999,3.999999999,-3.999999999),(1.0+2**-24,-1.0-2**-24,1.0+2**-24,-1.0-2**-24)]:
    add(1,'random/'+','.join(x.hex() for x in sequence),random=[f64(x) for _ in range(8) for x in sequence])
for final in (-1,1,2):add(1,'last-result/'+str(final),returns=[1]*7+[final])
for service,indices in [('random',(0,8,31)),('create',(0,3,7))]:
    for at in indices:
        for action in ('position','previous','force','speed','direction','update'):add(1,f'mutate/{service}/{at}/{action}',mutation=service+':'+action,at=at)
for bits in (0,0x80000000,1,0x3f800001,0x4b800001,0x7f7fffff,0x7f800000,0x7fc12345):
    for field in (0,2):
        c=add(2,f'data/{field}/{bits:x}');c['data'][0][field]=bits
for action in ('position','previous','update','radius','power','point'):add(2,'mutate/'+action,mutation='push:'+action)
for action in ('position','update','life','kind'):
    c=add(0,'delete/'+action,mutation='delete:'+action);c['data'][0][1]=0
def prepare(c,cw=0xe3f,source_ref=0):
    global applied,random_index,created
    u.context_restore(initial);u.reg_write(UC_X86_REG_FPCW,cw);u.mem_write(heap,b'\xa5'*0x10000);u.mem_write(stack,bytes(0x10000))
    for a,words in zip(objects,c['objects']):
        for off,v in zip(offsets,words):put(a+off,v)
    u.mem_write(objects[0]+124,struct.pack('<H',c['direction']));put(objects[0]+748,ref(data,c['update_ref']))
    for a,words in zip(data,c['data']):
        for n,v in enumerate(words):put(a+4*n,v)
    put(sp,stop);put(sp+4,ref(objects,source_ref));u.reg_write(UC_X86_REG_ESP,sp);events.clear();faults.clear();applied=False;random_index=0;created=0
def preserved(before):
    expected=bytearray(before)
    for a in objects:
        for off in offsets:expected[a-heap+off:a-heap+off+4]=u.mem_read(a+off,4)
    for off,z in ((124,2),(748,4)):expected[objects[0]-heap+off:objects[0]-heap+off+z]=u.mem_read(objects[0]+off,z)
    for a in data:expected[a-heap:a-heap+16]=u.mem_read(a,16)
    return bytes(expected)==bytes(u.mem_read(heap,0x10000))
rows=[];counterfactual=[]
for current in cases:
    result_rows=[]
    for cw in (0xe3f,0x23f):
        prepare(current,cw);before=bytes(u.mem_read(heap,0x10000));u.emu_start(specs[current['kind']][1],stop,count=20000)
        assert not faults and u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4 and u.reg_read(UC_X86_REG_FPCW)==cw and preserved(before)
        result_rows.append(dict(input=copy.deepcopy(current),output=state(),events=copy.deepcopy(events),result_ref=idx(objects,u.reg_read(UC_X86_REG_EAX)) if current['kind']==1 else -1,nonoutput_bytes_preserved=True))
    rows.append(result_rows[0]);counterfactual.append(result_rows[1])
assert args.game.read_bytes()==blob
proof=dict(original_sha256=sha(blob),original_file_unchanged=True,original_startup_FPCW=0xe3f,counterfactual_FPCW=0x23f,contexts=len(cases),executions=2*len(rows),regions=regions,direction_table=table,entries=dict(entries),rows=rows,counterfactual_rows=counterfactual,actual_vs_nearest_state_differences=sum(a['output']!=b['output'] for a,b in zip(rows,counterfactual)),actual_vs_nearest_trace_differences=sum(a['events']!=b['events'] for a,b in zip(rows,counterfactual)),faults=[],executed_instructions=[dict(address=hex(a),bytes=b) for a,b in sorted(trace.items())],arithmetic_or_conversion_substitutions=0,scope='Complete original callers and original419A70 conversion. Declared wide random returns through FLD/RET stub, Spark factory, push and delete services. Original direction words, signed indexing and actual PC53/ToZero are retained; nearest is a distinct counterfactual. No full world/native original-game or general signaling-NaN inference.')
with args.proof.open('x') as f:json.dump(proof,f,separators=(',',':'))
with args.fixture.open('x') as f:json.dump({k:v for k,v in proof.items() if k not in ('executed_instructions','counterfactual_rows')},f,separators=(',',':'))
fault_rows=[]
for kind in range(3):
    for name,source_ref,update_ref in [('nil-source',-1,0),('nil-update',0,-1)]:
        current=copy.deepcopy(cases[0]);current.update(name=specs[kind][0]+'/'+name,kind=kind,update_ref=update_ref,mutation='')
        if kind==2:current['data']=[[f32(5),0,f32(6),0],[0,0,0,0]]
        prepare(current,source_ref=source_ref);before=bytes(u.mem_read(heap,0x10000));error=None
        try:u.emu_start(specs[kind][1],stop,count=20000)
        except Exception as exc:error=str(exc)
        fault=source_ref<0 or kind!=1;assert bool(faults)==fault and preserved(before)
        fault_rows.append(dict(input=current,source_ref=source_ref,fault=fault,faults=list(faults),error=error,events=copy.deepcopy(events),output=state(),result_ref=-1 if fault or kind!=1 else idx(objects,u.reg_read(UC_X86_REG_EAX)),nonoutput_bytes_preserved=True))
with args.fault_proof.open('x') as f:json.dump(dict(original_sha256=sha(blob),rows=fault_rows),f,separators=(',',':'))
print(json.dumps({k:proof[k] for k in ('contexts','executions','entries','actual_vs_nearest_state_differences','actual_vs_nearest_trace_differences')}))
