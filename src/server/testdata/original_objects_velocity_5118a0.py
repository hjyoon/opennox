"""Whole unchanged GAME.EXE 5118A0 root. External services are declared.

Trace's final argument is a C char: compare its observable low byte; the
original caller leaves unrelated high DWORD bits in the argument slot.
No original arithmetic/conversion instruction is replaced.
"""
import argparse,collections,copy,hashlib,json,random,struct
from pathlib import Path
from unicorn import Uc,UcError,UC_ARCH_X86,UC_MODE_32,UC_HOOK_CODE,UC_HOOK_MEM_INVALID
from unicorn.x86_const import *
p=argparse.ArgumentParser(description=__doc__)
for key in ('game','actual','nearest','proof'):p.add_argument('--'+key,type=Path,required=True)
p.add_argument('--prefix-inputs',type=Path,default=Path(__file__).with_name('objects_velocity_prefix_inputs_5118a0.json'))
a=p.parse_args();blob=a.game.read_bytes();sha=lambda b:hashlib.sha256(b).hexdigest()
assert sha(blob)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,60)[0];opt=pe+24;base=struct.unpack_from('<I',blob,opt+28)[0];size=(struct.unpack_from('<I',blob,opt+56)[0]+4095)&~4095;table=opt+struct.unpack_from('<H',blob,pe+20)[0]
u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,size)
for i in range(struct.unpack_from('<H',blob,pe+6)[0]):
 _,va,count,off=struct.unpack_from('<IIII',blob,table+40*i+8)
 if count:u.mem_write(base+va,blob[off:off+count])
heap,stack,code=0x21000000,0x22000000,0x23000000
for v in (heap,stack,code):u.mem_map(v,0x10000)
sp=stack+0xe000;stop=code+0x200;init=code+0x100;cwaddr=heap+0xf000
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cwaddr)+b'\xc3')
objects=[heap+128+i*2048 for i in range(3)];health=heap+0xd000;cacheaddr=0x81b028;start,end=0x5118a0,0x511b57
region=bytes(u.mem_read(start,end-start));get=lambda v:struct.unpack('<I',u.mem_read(v,4))[0];put=lambda v,x:u.mem_write(v,struct.pack('<I',x&0xffffffff));words=lambda v,n:[get(v+4*i) for i in range(n)];f32=lambda x:struct.unpack('<I',struct.pack('<f',x))[0]
names=['SmallFlameCleanse','SmallFlameCleanse','MediumFlameCleanse','FlameCleanse','LargeFlameCleanse','SmallBlueFlameCleanse','SmallBlueFlameCleanse','MediumBlueFlameCleanse','BlueFlameCleanse','LargeBlueFlameCleanse']
lookup=[400,400,401,402,403,404,404,405,406,407]
properties={'pos':56,'new':64,'vel':80,'force':88,'acc':96}
# Reject an engine that preserves SNaN on FLD m32fp. This is an emulator
# calibration, before original execution/coverage collection, with no hooks.
probe=code+0x300;nanaddr=heap+0xf100
u.mem_write(probe,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cwaddr)+b'\xd9\x05'+struct.pack('<I',nanaddr))
for control in (0xe3f,0x23f):
 for word in (0x7f812345,0xff812345,0x7f832345,0x7fc12345):
  u.mem_write(cwaddr,struct.pack('<H',control));put(nanaddr,word)
  u.emu_start(probe,probe+14,count=4);sw=u.reg_read(UC_X86_REG_FPSW)
  sig,exp=u.reg_read(UC_X86_REG_FP0+((sw>>11)&7))
  actual=((exp&0x8000)<<16)|0x7f800000|((sig>>40)&0x7fffff)
  assert actual==word|0x400000 and sw&1==int(word&0x400000==0), 'Engine FLD m32fp SNaN defect; apply documented patch'
recording=False;trace={};mode_trace={};events=[];faults=[];calls=collections.Counter();current={};phase=-1;node=-1;mutation_count=0;raw_trace_words=[]
def idx(v):return -1 if not v else objects.index(v)
def snapshot():
 rows=[]
 for ptr in objects:
  row={k:words(ptr+off,2) for k,off in properties.items()};row.update(typ=struct.unpack('<H',u.mem_read(ptr+4,2))[0],cls=get(ptr+8),flags=get(ptr+16),drag=get(ptr+112),health=get(ptr+556)!=0);rows.append(row)
 return dict(cache=words(cacheaddr,10),objects=rows)
def event(kind,args):events.append(dict(kind=kind,args=args,state=snapshot()))
def mutate(kind,point=0,ray=0):
 global mutation_count
 count=calls[kind];calls[kind]+=1;m=current['mutation']
 if kind!=m['stage'] or count!=m['at']:return
 mutation_count+=1;ptr=objects[m['object']];action=m['action']
 if action=='clear-acc':put(ptr+96,0);put(ptr+100,0)
 elif action=='motion':
  for off,val in [(88,-1.1),(92,.29),(96,.4),(100,-.6),(112,.125)]:put(ptr+off,f32(val))
 elif action=='frozen':put(ptr+16,get(ptr+16)|2)
 elif action=='fly':put(ptr+16,get(ptr+16)|0x4000);put(ptr+556,0)
 elif action=='health-on':put(ptr+556,health)
 elif action=='new':put(ptr+64,f32(7.25));put(ptr+68,f32(9.5))
 elif action=='reset-new':put(ptr+64,get(ptr+56));put(ptr+68,get(ptr+60))
 elif action=='point':
  assert point==ptr+64
  put(point,f32(7.25));put(point+4,f32(9.5))
 elif action=='trace-max':
  assert ray
  put(ray+8,f32(7.25));put(ray+12,f32(9.5))
 elif action=='cache':
  for i in range(10):put(cacheaddr+4*i,700+i)
  u.mem_write(ptr+4,struct.pack('<H',706))
 else:raise AssertionError(action)
def ret(value=0):
 esp=u.reg_read(UC_X86_REG_ESP);u.reg_write(UC_X86_REG_EAX,value&0xffffffff);u.reg_write(UC_X86_REG_EIP,get(esp));u.reg_write(UC_X86_REG_ESP,esp+4)
def hook(_u,pc,length,_):
 global phase,node
 if not recording:return
 if pc==stop:u.emu_stop();return
 if start<=pc<end:
  encoded=bytes(u.mem_read(pc,length)).hex();trace[pc]=encoded;mode_trace[pc]=encoded;return
 esp=u.reg_read(UC_X86_REG_ESP);arg=lambda n:get(esp+4*n)
 if pc==0x4e3aa0:
  k=calls['lookup'];value=bytes(u.mem_read(arg(1),64)).split(b'\0',1)[0].decode();assert value==names[k];event('lookup',[k]);mutate('lookup');ret(current['lookup'][k])
 elif pc==0x537740:
  phase+=1;assert phase<2;event('first',[phase]);mutate('first');order=current['order'+str(phase)];node=order[0] if order else -1;ret(0 if node<0 else objects[node])
 elif pc==0x537750:
  old=idx(arg(1));event('next',[phase,old]);mutate('next');order=current['order'+str(phase)];i=order.index(old)+1;node=order[i] if i<len(order) else -1;ret(0 if node<0 else objects[node])
 elif pc==0x5481c0:
  event('prepare',[idx(arg(1))]);mutate('prepare');ret()
 elif pc==0x5113a0:event('springs',[]);mutate('springs');ret()
 elif pc==0x50a0d0:
  n=idx(arg(1));assert arg(2)==67;event('action',[n,67]);mutate('action');ret(current['actions'][n])
 elif pc==0x4ff350:
  n=idx(arg(1));b=arg(2);assert b in (5,25,28);event('buff',[n,b]);mutate('buff');ret(current['buffs'][n][(5,25,28).index(b)])
 elif pc==0x535250:
  ray=arg(1);assert arg(2)==0 and arg(3)==0;raw_trace_words.append(arg(4));event('trace',words(ray,4)+[arg(4)&255,0,0]);mutate('trace',ray=ray);ret(current['traces'][node])
 elif pc==0x411160:
  point=arg(1);assert point==objects[node]+64;event('tile',[node]+words(point,2));mutate('tile',point=point);ret(current['tiles'][node])
 elif pc==0x548630:
  n=idx(arg(1));assert arg(2)==6 and words(arg(3),2)==[0,0];event('collision',[n,6,0,0]);mutate('collision');ret()
 elif pc in (0x4e44f0,0x4e7290,0x517970):
  kind={0x4e44f0:'sync',0x4e7290:'coords',0x517970:'move'}[pc];event(kind,[idx(arg(1))]);mutate(kind);ret()
 else:raise AssertionError('undeclared instruction '+hex(pc))
def invalid(_u,access,address,length,value,_):
 if recording:faults.append(dict(access=access,address=hex(address),length=length,pc=hex(u.reg_read(UC_X86_REG_EIP))))
 return False
u.hook_add(UC_HOOK_CODE,hook);u.hook_add(UC_HOOK_MEM_INVALID,invalid)
def object_state():return dict(typ=77,cls=2,flags=0x10000004,pos=[f32(100.3),f32(50.7)],new=[f32(100.3),f32(50.7)],vel=[f32(1.3),f32(-2.7)],force=[f32(.229),f32(-.341)],acc=[f32(.123),f32(2.718)],drag=f32(.081),health=True)
def base_case(name):
 o=object_state();p=copy.deepcopy(o);q=copy.deepcopy(o);p.update(pos=[f32(300.25),f32(400.75)],new=[f32(300.25),f32(400.75)],typ=88);q.update(typ=99)
 return dict(name=name,step=f32(.2),cache=[0]*10,lookup=lookup.copy(),objects=[o,p,q],order0=[0],order1=[0],actions=[0]*3,buffs=[[0]*3 for _ in range(3)],traces=[1]*3,tiles=[0]*3,mutation=dict(stage='',at=0,action='',object=0))
cases=[]
def add(name,edit=None):
 c=base_case(name)
 if edit:edit(c)
 cases.append(c);return c
add('ordinary')
add('empty/init',lambda c:c.update(order0=[],order1=[]))
add('empty/prefilled',lambda c:c.update(order0=[],order1=[],cache=lookup.copy()))
add('cache/partial',lambda c:c.update(cache=[400]+[0]*9))
add('cache/lookup-zero',lambda c:c.update(lookup=[0]+lookup[1:]))
add('cache/lookup-negative',lambda c:c.update(lookup=[0xffffffff]*10))
add('iteration/two',lambda c:c.update(order0=[0,1],order1=[0,1]))
add('iteration/reordered',lambda c:c.update(order0=[0,1,2],order1=[2,0,1]))
add('iteration/prepare-only',lambda c:c.update(order0=[0,1],order1=[]))
add('iteration/integrate-only',lambda c:c.update(order0=[],order1=[2]))
for flags in (0,2,4,6,0x4004,0xffffffff):
 c=add('flags/'+hex(flags));c['objects'][0]['flags']=flags
for cls in (0,1,2,0xffffffff):
 for result in (0,1,-1,-2147483648):
  c=add('action/'+str(cls)+'/'+str(result));c['objects'][0]['cls']=cls;c['actions'][0]=result
for b in range(3):
 for result in (1,-1,-2147483648):
  c=add('buff/'+str(b)+'/'+str(result));c['buffs'][0][b]=result
for result in (0,1,-1,-2147483648):
 c=add('trace/'+str(result));c['traces'][0]=result
for slot in range(10):
 c=add('type-slot/'+str(slot));c['cache']=list(range(500,510));c['objects'][0]['typ']=500+slot
c=add('type-high-word-must-not-match');c['cache']=[0x1004d]*10
for health_flag in (False,True):
 for flying in (False,True):
  for tile in (0,6,-1):
   c=add('tile/'+str(health_flag)+'/'+str(flying)+'/'+str(tile));c['objects'][0]['health']=health_flag;c['objects'][0]['flags']|=0x4000 if flying else 0;c['tiles'][0]=tile
for axis in (0,1):
 for delta in (-7e-10,-1e-10,0,1e-10,7e-10):
  c=add('sync/'+str(axis)+'/'+str(delta));o=c['objects'][0];o['pos']=[0,0];o['new']=[0,0];o['pos'][axis]=f32(delta);o['new'][axis]=f32(.01);c['traces'][0]=0
prefix=json.loads(a.prefix_inputs.read_text())
for row in prefix['rows']:
 c=add('prefix/'+row['name']+'/'+str(row['buff_branch']));o=c['objects'][0];v=row['input_words'];o['vel']=v[:2];o['force']=v[2:4];o['acc']=v[4:6];o['drag']=v[6];c['step']=row['step_word'];c['buffs'][0][0]=int(row['buff_branch'])
special=[0,0x80000000,1,0x80000001,0x007fffff,0x00800000,0x7f7fffff,0xff7fffff,0x7f800000,0xff800000,0x7fc12345,0x7f812345,0xffc23456,0xff812345]
for field,index in [('step',0),('drag',0),('vel',0),('vel',1),('force',0),('force',1),('acc',0),('acc',1),('pos',0),('pos',1),('new',0),('new',1)]:
 for word in special:
  c=add('word/'+field+'/'+str(index)+'/'+format(word,'08x'))
  if field=='step':c['step']=word
  elif field=='drag':c['objects'][0]['drag']=word
  else:c['objects'][0][field][index]=word
mutations=[('lookup',0,'cache'),('prepare',0,'motion'),('springs',0,'motion'),('springs',0,'cache'),('action',0,'frozen'),('buff',0,'motion'),('buff',1,'motion'),('buff',2,'motion'),('trace',0,'trace-max'),('trace',0,'new'),('trace',0,'fly'),('trace',0,'health-on'),('tile',0,'point'),('collision',0,'new'),('sync',0,'reset-new'),('coords',0,'reset-new'),('move',0,'motion'),('next',0,'motion')]
for stage,at,action in mutations:
 c=add('mutation/'+stage+'/'+str(at)+'/'+action);c['mutation']=dict(stage=stage,at=at,action=action,object=0)
 if stage=='collision' or action=='fly' or action=='health-on':c['tiles'][0]=6
 if action=='health-on':c['objects'][0]['health']=False
 if action in ('trace-max','new') and stage=='trace':
  for result in (0,1):
   d=copy.deepcopy(c);d['name']+='/ret'+str(result);d['traces'][0]=result;cases.append(d)
  cases.remove(c)
for stage,action in [('prepare','frozen'),('trace','motion')]:
 c=add('mutation/other-object/'+stage+'/'+action);c.update(order0=[0,1],order1=[0,1]);c['mutation']=dict(stage=stage,at=0,action=action,object=1)
pairs=[(('acc',0),('force',0)),(('acc',1),('force',1)),(('vel',0),('drag',0)),(('vel',1),('drag',0)),(('step',0),('new',0)),(('step',0),('new',1))]
for first,second in pairs:
 for x,y in [(0x7fc12345,0x7f832345),(0x7f812345,0x7fc32345),(0xffc12345,0x7fc32345),(0xff812345,0x7f832345)]:
  c=add('nan-pair/'+str(first)+'/'+str(second)+'/'+format(x,'08x')+'/'+format(y,'08x'))
  for (field,index),value in [(first,x),(second,y)]:
   if field=='step':c['step']=value
   elif field=='drag':c['objects'][0]['drag']=value
   else:c['objects'][0][field][index]=value
for frozen in [False,True]:
 c=add('native-real/'+('frozen' if frozen else 'moving'));o=c['objects'][0]
 o.update(cls=0x100000,flags=0x44 | (2 if frozen else 0),health=False)
 c['cache']=lookup.copy();c['mutation']=dict(stage='prepare',at=0,action='clear-acc',object=0)
assert len({c['name'] for c in cases})==len(cases)
def execute(c,cw):
 global recording,current,phase,node,mutation_count
 current=copy.deepcopy(c);phase=-1;node=-1;mutation_count=0;events.clear();faults.clear();calls.clear();u.mem_write(heap,b'\xa5'*0x10000);u.mem_write(stack,bytes(0x10000));put(cwaddr,cw);put(sp,stop);u.reg_write(UC_X86_REG_ESP,sp);u.emu_start(init,stop,count=100);assert u.reg_read(UC_X86_REG_FPCW)==cw
 for i,row in enumerate(c['objects']):
  ptr=objects[i];u.mem_write(ptr+4,struct.pack('<H',row['typ']&0xffff));put(ptr+8,row['cls']);put(ptr+16,row['flags']);put(ptr+112,row['drag']);put(ptr+556,health if row['health'] else 0)
  for key,off in properties.items():
   for j,w in enumerate(row[key]):put(ptr+off+4*j,w)
 for i,w in enumerate(c['cache']):put(cacheaddr+4*i,w)
 before=bytes(u.mem_read(heap,0x10000));globals_before=bytes(u.mem_read(cacheaddr-16,72));put(sp,stop);put(sp+4,c['step']);u.reg_write(UC_X86_REG_ESP,sp)
 saved_regs=[UC_X86_REG_EBX,UC_X86_REG_ESI,UC_X86_REG_EDI,UC_X86_REG_EBP];saved_values=[0x12345678,0x23456789,0x3456789a,0x456789ab]
 for reg,value in zip(saved_regs,saved_values):u.reg_write(reg,value)
 recording=True
 try:u.emu_start(start,stop+1,count=30000)
 except UcError:
  if not faults:raise
 finally:recording=False
 assert not faults,(c['name'],faults)
 assert u.reg_read(UC_X86_REG_EIP)==stop,(c['name'],hex(u.reg_read(UC_X86_REG_EIP)))
 assert phase==1 and u.reg_read(UC_X86_REG_EAX)==0 and u.reg_read(UC_X86_REG_ESP)==sp+4
 assert [u.reg_read(reg) for reg in saved_regs]==saved_values
 globals_after=bytes(u.mem_read(cacheaddr-16,72));assert globals_before[:16]==globals_after[:16] and globals_before[56:]==globals_after[56:]
 assert mutation_count==int(bool(c['mutation']['stage'])),(c['name'],mutation_count)
 after=bytearray(u.mem_read(heap,0x10000));expected=bytearray(before)
 # Services may alter only the explicitly modeled fields, while the native
 # arithmetic root alters only NewPos, velocity and acceleration. Canaries
 # and every other byte of every original frame remain unchanged.
 allowed={i:[(64,8),(80,8),(96,8)] for i in c['order1']}
 m=c['mutation'];extra={'clear-acc':[(96,8)],'motion':[(88,16),(112,4)],'frozen':[(16,4)],'fly':[(16,4),(556,4)],'health-on':[(556,4)],'new':[(64,8)],'reset-new':[(64,8)],'point':[(64,8)],'trace-max':[],'cache':[(4,2)],'':[]}
 if m['stage']:allowed.setdefault(m['object'],[]).extend(extra[m['action']])
 for i,ranges in allowed.items():
  ptr=objects[i]
  for off,count in ranges:expected[ptr-heap+off:ptr-heap+off+count]=after[ptr-heap+off:ptr-heap+off+count]
 assert after==expected
 return dict(input=c,result=dict(state=snapshot(),events=copy.deepcopy(events),return_value=0),all_nonoutput_bytes_preserved=True)
actual=[execute(c,0xe3f) for c in cases];actual_trace=mode_trace.copy();mode_trace.clear();nearest=[execute(c,0x23f) for c in cases];nearest_trace=mode_trace.copy();covered=set()
for pc,h in trace.items():covered.update(range(pc,pc+len(bytes.fromhex(h))))
assert covered==set(range(start,end)),(len(covered),[hex(x) for x in sorted(set(range(start,end))-covered)])
for label,observed in [('actual',actual_trace),('nearest',nearest_trace)]:
 span=set()
 for pc,h in observed.items():span.update(range(pc,pc+len(bytes.fromhex(h))))
 assert span==set(range(start,end)),(label,len(span))
for path,data in [(a.actual,dict(original_sha256=sha(blob),control_word=0xe3f,emulator_fld_patch_sha256='183d9b99af3f1eec4f6a951a8a7e6c314ea7615a4ebdc66f02697e3764d07e12',rows=actual)),(a.nearest,dict(original_sha256=sha(blob),control_word=0x23f,emulator_fld_patch_sha256='183d9b99af3f1eec4f6a951a8a7e6c314ea7615a4ebdc66f02697e3764d07e12',rows=nearest))]:path.write_text(json.dumps(data,separators=(',',':'))+'\n')
proof=dict(original_sha256=sha(blob),regions=[dict(entry=hex(start),end_exclusive=hex(end),bytes=len(region),sha256=sha(region))],all_instruction_bytes_executed=len(covered),each_control_mode_covers_all_695_bytes=True,callee_saved_registers_and_return_stack_preserved=True,adjacent_global_and_nonoutput_heap_canaries_preserved=True,zero_arithmetic_or_conversion_replacements=True,contexts=len(cases),original_executions=2*len(cases),actual_control_word='0xe3f',nearest_counterfactual='0x23f',meaningful_mode_differences=sum(x['result']!=y['result'] for x,y in zip(actual,nearest)),trace_last_argument_declared_C_char=True,original_raw_trace_word_high_bits_seen=any(x>255 for x in raw_trace_words),services=['type lookup','two-pass enumeration','collision preparation','springs','action','buff','trace','tile','collision enqueue','sync','collider update','map-index update'],service_mutations_verified=sum(bool(c['mutation']['stage']) for c in cases),required_faults=0,transitive_callee_composition_claim=False,instructions=[dict(pc=hex(pc),bytes=h) for pc,h in sorted(trace.items())],fixture_sha256=sha(a.actual.read_bytes()),nearest_sha256=sha(a.nearest.read_bytes()))
proof['emulator_fld_correction'] = dict(upstream='unicorn-engine/unicorn',version='2.1.4',patch='unicorn-flds-masked-snan-v1.patch',scope='FLD m32fp masked signaling NaN quieting only; FT0 arithmetic memory conversion unchanged',independent_load_probe='diagnostic-x87-load-v1.stdout.log (Rosetta x86_64; native x86 CI separately required)',original_instruction_bytes_unmodified=True,service_hooks_unchanged=True)
proof['masked_snan_load_probe_verified']=True
a.proof.write_text(json.dumps(proof,indent=2));print(json.dumps({k:v for k,v in proof.items() if k!='instructions'}))
