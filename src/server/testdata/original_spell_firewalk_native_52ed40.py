"""Execute original Firewalk52ED40 under declared name/RNG/world services.

The whole original root executes its own arithmetic, FSQRT, spills, signed
level tests, cache stores, record writes and live reloads. Six call targets
are bounded world services, not transitive engine equivalence claims.
"""
import argparse,collections,copy,hashlib,json,struct
from pathlib import Path
from unicorn import Uc,UC_ARCH_X86,UC_MODE_32,UC_HOOK_CODE
from unicorn.x86_const import *
p=argparse.ArgumentParser(description=__doc__)
for k in ['game','fixture','proof']:p.add_argument('--'+k,type=Path,required=True)
p.add_argument('--control',type=lambda s:int(s,0),choices=[0xe3f,0x23f],default=0xe3f)
a=p.parse_args();blob=a.game.read_bytes();sha=lambda b:hashlib.sha256(b).hexdigest()
assert sha(blob)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,60)[0];opt=pe+24;base=struct.unpack_from('<I',blob,opt+28)[0];table=opt+struct.unpack_from('<H',blob,pe+20)[0]
u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,(struct.unpack_from('<I',blob,opt+56)[0]+4095)&~4095)
for n in range(struct.unpack_from('<H',blob,pe+6)[0]):
 _,va,size,off=struct.unpack_from('<IIII',blob,table+40*n+8)
 if size:u.mem_write(base+va,blob[off:off+size])
heap,stack,code=0x21000000,0x22000000,0x23000000
for addr in [heap,stack,code]:u.mem_map(addr,0x10000)
record=heap+0x1000;objects=[heap+0x2000,heap+0x3000];flames=[heap+0x4000,heap+0x5000]
sp,stop,init,cw=stack+0xe000,code+0x100,code+0x200,code+0x300
get=lambda addr:struct.unpack('<I',u.mem_read(addr,4))[0]
put=lambda addr,v:u.mem_write(addr,struct.pack('<I',v&0xffffffff))
bits=lambda v:struct.unpack('<I',struct.pack('<f',v))[0]
point=lambda x,y:[bits(x),bits(y)]
readpoint=lambda addr:[get(addr),get(addr+4)]
ref=lambda i:0 if i<0 else objects[i]
index=lambda p:-1 if not p else objects.index(p)
signed=lambda v:((v&0xffffffff)^0x80000000)-0x80000000
cache=[0x833be4,0x833be8,0x833bec];fps=0x85b3fc
lo,hi=0x52ed40,0x52ef23
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw))
traces={};entries,services=collections.Counter(),collections.Counter();active=False;current={};events=[];calls=collections.Counter();allowed=set();changed=False
def state():
 return dict(target=index(get(record+48)),level=get(record+8),frame60=get(record+60),frame64=get(record+64),previous=readpoint(record+72),anchor=readpoint(record+80),positions=[readpoint(o+56) for o in objects],flags=[get(o+16) for o in objects],radii=[get(o+176) for o in objects],cache=[get(c) for c in cache],fps=get(fps))
def write(addr,v):allowed.update(range(addr,addr+4));put(addr,v)
def mutate(kind):
 global changed
 at=calls[kind];calls[kind]+=1;m=current['mutation']
 if changed or m['stage']!=kind or m['at']!=at:return
 changed=True;action=m['action']
 if action=='target-b':write(record+48,objects[1])
 elif action=='target-position':
  for j,v in enumerate(point(81.25,-91.5)):write(objects[0]+56+4*j,v)
 elif action=='previous':
  for j,v in enumerate(point(3.25,-4.5)):write(record+72+4*j,v)
 elif action=='anchor':
  for j,v in enumerate(point(7.25,-8.5)):write(record+80+4*j,v)
 elif action=='level':write(record+8,4)
 elif action=='frame60':write(record+60,11)
 elif action=='frame64':write(record+64,999)
 elif action=='fps':write(fps,60)
 elif action=='cache':
  for i,c in enumerate(cache):write(c,91+i)
 else:raise AssertionError(action)
def event(kind,args):events.append(dict(kind=kind,args=args,state=state()));mutate(kind)
def ret(v=0):
 esp=u.reg_read(UC_X86_REG_ESP);u.reg_write(UC_X86_REG_EAX,v&0xffffffff);u.reg_write(UC_X86_REG_EIP,get(esp));u.reg_write(UC_X86_REG_ESP,esp+4)
def hook(_u,pc,length,_):
 if not active:return
 if pc==stop:u.emu_stop();return
 if lo<=pc<hi:
  traces[pc]=bytes(u.mem_read(pc,length))
  if pc==lo:entries['update']+=1
  return
 esp=u.reg_read(UC_X86_REG_ESP);arg=lambda n:get(esp+4*n);services[hex(pc)]+=1
 if pc==0x4e3aa0:
  name=bytes(u.mem_read(arg(1),64)).split(b'\0',1)[0].decode('ascii');i=['SmallFlame','MediumFlame','Flame'].index(name);event('name',[name]);ret(current['name_ids'][i])
 elif pc==0x415fa0:
  assert arg(1)==0 and 0<=signed(arg(2))<=2 and arg(4)==898
  at=calls['random'];choice=min(current['random'][at],arg(2));event('random',[0,signed(arg(2)),choice]);ret(choice)
 elif pc==0x4e3450:
  at=calls['allocate'];value=flames[at] if current['allocate'][at] else 0;event('allocate',[arg(1),at if value else -1]);ret(value)
 elif pc==0x4daa50:
  assert arg(2)==arg(5)==0;event('create',[flames.index(arg(1)),arg(3),arg(4)]);ret(0x13572468)
 elif pc==0x501a30:
  assert arg(1)==46 and arg(3)==arg(4)==0;event('audio',[arg(1),*readpoint(arg(2))]);ret(0x13572468)
 elif pc==0x511660:event('decay',[flames.index(arg(1)),arg(2)]);ret(0x13572468)
 else:raise AssertionError((current['name'],hex(pc)))
u.hook_add(UC_HOOK_CODE,hook)
seed=bytes((i*31+19)&255 for i in range(0x10000));registers={UC_X86_REG_EBX:0x12345678,UC_X86_REG_ESI:0x23456789,UC_X86_REG_EDI:0x3456789a,UC_X86_REG_EBP:0x456789ab}
def run(c):
 global active,current,events,calls,allowed,changed
 active,current,events,calls,allowed,changed=False,c,[],collections.Counter(),set(),False;u.mem_write(heap,seed)
 for off,value in [(8,c['level']),(48,ref(c['target'])),(60,c['frame60']),(64,c['frame64'])]:put(record+off,value)
 for off,values in [(72,c['previous']),(80,c['anchor'])]:
  for j,v in enumerate(values):put(record+off+4*j,v)
 for i,o in enumerate(objects):
  put(o+16,c['flags'][i]);put(o+176,c['radii'][i])
  for j,v in enumerate(c['positions'][i]):put(o+56+4*j,v)
 u.mem_write(cache[0]-16,bytes(range(64)))
 for addr,v in zip(cache,c['cache']):put(addr,v)
 put(fps,c['fps']);globals_before={cache[0]-16:bytes(u.mem_read(cache[0]-16,64)),fps:bytes(u.mem_read(fps,4))}
 before=bytes(u.mem_read(heap,len(seed)));u.mem_write(cw,struct.pack('<H',a.control));u.emu_start(init,init+8,count=2)
 u.mem_write(sp-512,bytes(1024));put(sp,stop);put(sp+4,record)
 for reg,v in registers.items():u.reg_write(reg,v)
 u.reg_write(UC_X86_REG_EFLAGS,0x202);u.reg_write(UC_X86_REG_ESP,sp);active=True;u.emu_start(lo,stop+1,count=10000);active=False
 assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4 and all(u.reg_read(reg)==v for reg,v in registers.items())
 assert u.reg_read(UC_X86_REG_FPCW)==a.control and u.reg_read(UC_X86_REG_FPSW)&0x3800==0
 assert bytes(u.mem_read(sp-512,256))==bytes(256) and bytes(u.mem_read(sp+8,256))==bytes(256)
 for off in [64,72,76,80,84]:allowed.update(range(record+off,record+off+4))
 after=bytes(u.mem_read(heap,len(seed)));assert all(x==y or heap+i in allowed for i,(x,y) in enumerate(zip(before,after)))
 for addr,b in globals_before.items():
  observed=bytes(u.mem_read(addr,len(b)))
  assert all(x==y or addr+i in allowed or cache[0]<=addr+i<cache[0]+12 for i,(x,y) in enumerate(zip(b,observed)))
 return dict(input=c,result=dict(return_value=signed(u.reg_read(UC_X86_REG_EAX)),events=copy.deepcopy(events),state=state()))
default=dict(name='default',target=0,flags=[0,0],positions=[point(100,200),point(-100,-200)],radii=[bits(2),bits(3)],frame60=10,frame64=11,level=3,previous=point(30,40),anchor=point(0,0),cache=[11,22,33],name_ids=[11,22,33],fps=30,random=[0,2],allocate=[True,True],mutation=dict(stage='',at=0,action=''))
cases=[]
def case(name,**kw):c=copy.deepcopy(default);c.update(kw,name=name);cases.append(c)
case('default');case('nil',target=-1)
for flags in [0x20,0x8000,0x8020,0x10000]:case(f'flags-{flags}',flags=[flags,0])
case('first',frame60=10,frame64=10);case('first-wrap',frame60=0xffffffff,frame64=0xffffffff);case('cold-first',frame60=10,frame64=10,cache=[0,0,0]);case('cold-later',cache=[0,0,0])
for level in [0,1,2,4,0x7fffffff,0x80000000,0xffffffff]:case(f'level-{level}',level=level)
for v in [14.999999,15,15.000001,17,17.000002,1e20,3.4028234663852886e38,float('inf'),float('nan')]:case(f'distance-{v}',positions=[point(v,0),point(7,8)],previous=point(0,0),radii=[bits(0),bits(3)])
for alloc in [[False,False],[True,False],[False,True]]:case(f'alloc-{alloc}',allocate=alloc)
for anchor in [point(30,0),point(0,40),point(30,40),point(-3.4028234663852886e38,-3.4028234663852886e38)]:case(f'anchor-{anchor}',anchor=anchor)
for stage in ['name','random','allocate','create','audio','decay']:
 for action in ['target-b','target-position','previous','anchor','level','frame64','fps','cache']:
  case(f'mutate-{stage}-{action}',cache=[0,0,0],mutation=dict(stage=stage,at=0,action=action))
# Expand boundary coverage without changing any engine or existing fixture.
for stage,count in [('name',3),('random',2),('allocate',2),('create',2),('audio',2),('decay',2)]:
 for at in range(count):
  for action in ['target-b','target-position','previous','anchor','level','frame60','frame64','fps','cache']:
   case(f'boundary-{stage}-{at}-{action}',cache=[0,0,0],mutation=dict(stage=stage,at=at,action=action))
for stage in ['name','random','allocate','create','audio','decay']:
 for at in range(3 if stage=='name' else 2):
  for action in ['target-b','target-position','previous','anchor','level','frame60','frame64','fps','cache']:
   case(f'first-boundary-{stage}-{at}-{action}',frame60=10,frame64=10,cache=[0,0,0],mutation=dict(stage=stage,at=at,action=action))
for cache_values in [[0,22,33],[11,0,0],[0,0xffffffff,0x80000000],[0xffffffff,0,0]]:
 for first in [False,True]:case(f'partial-cache-{cache_values}-{first}',cache=cache_values,frame64=10 if first else 11)
for values in [[0,0,0],[0xffffffff,0x80000000,0x12345678]]:
 for first in [False,True]:case(f'name-ids-{values}-{first}',cache=[0,0,0],name_ids=values,frame64=10 if first else 11)
for value in [0,1,60,0x0a3d70a3,0x7fffffff,0x80000000,0xffffffff]:case(f'fps-{value}',fps=value)
for level in [0xfffffffe,0x80000001,0x7ffffffe,3,5]:case(f'extra-signed-level-{level}',level=level,random=[2,2])
for values in [[0,0],[1,1],[2,2],[2,0],[1,2]]:
 for level in [0,2,3,4,0x80000000]:case(f'random-{values}-level-{level}',random=values,level=level)
# Every nonfinite bit pattern is carried as bits; no JSON NaN normalization.
special=[0,0x80000000,1,0x80000001,0x007fffff,0x00800000,0x7f7fffff,0xff7fffff,0x7f800000,0xff800000,0x7fc12345,0xffc12345,0x7f812345,0xff812345]
for value in special:
 for field in ['position-x','position-y','previous-x','previous-y','anchor-x','anchor-y','radius']:
  kw={}
  if field.startswith('position'):
   v=point(100,200);v[0 if field.endswith('x') else 1]=value;kw['positions']=[v,point(7,8)]
  elif field.startswith('previous'):
   v=point(30,40);v[0 if field.endswith('x') else 1]=value;kw['previous']=v
  elif field.startswith('anchor'):
   v=point(0,0);v[0 if field.endswith('x') else 1]=value;kw['anchor']=v
  else:kw['radii']=[value,bits(3)]
  case(f'bits-{field}-{value:08x}',**kw)
# Binary32 neighbors at both diagonal and one-axis thresholds; original PC53
# operations and spills determine outcomes, rather than Python expectations.
for radius in [0,2,5,-2,0.0000001,1e10]:
 center=bits(15+radius)
 for delta in [-2,-1,0,1,2]:
  v=center+delta
  if 0<=v<=0xffffffff:case(f'threshold-{radius}-{delta}',positions=[[v,0],point(7,8)],previous=point(0,0),radii=[bits(radius),bits(3)])
for x,y in [(3,4),(6,8),(9,12),(12,9),(1e-20,1e-20),(1e20,-1e20)]:
 for r in [0,1e-20,1e20]:case(f'diagonal-{x}-{y}-{r}',positions=[point(x,y),point(7,8)],previous=point(0,0),radii=[bits(r),bits(3)])
# Wide exponents reveal the explicit delta spill and PC53 chop behavior.
for previous,anchor in [(point(0.1,0.3),point(0.2,0.4)),(point(16777216,3),point(-1,-3)),(point(1e-30,1e30),point(-1e-30,-1e30)),(point(-0.0,-0.0),point(0.0,0.0))]:
 case(f'spill-{previous}-{anchor}',previous=previous,anchor=anchor,positions=[point(1e31,1e31),point(7,8)])
for stage in ['name','random','allocate','create','audio','decay']:
 case(f'dead-replacement-{stage}',flags=[0,0x8020],cache=[0,0,0],mutation=dict(stage=stage,at=0,action='target-b'))

fixture=dict(control_word=a.control,original_machine_code_executed=True,cases=[run(c) for c in cases])
proof=dict(original_sha256=sha(blob),control_word=a.control,contexts=len(cases),root=dict(start=hex(lo),end_exclusive=hex(hi),bytes=hi-lo,sha256=sha(bytes(u.mem_read(lo,hi-lo))),covered_bytes=sum(len(b) for b in traces.values()),missing_bytes=[hex(i) for i in range(lo,hi) if not any(pc<=i<pc+len(b) for pc,b in traces.items())]),entries=dict(entries),services=dict(services),register_stack_FPCW_TOP_heap_global_canaries_preserved=True,faults=0,original_machine_code_executed=True,independent_comparator_is_external=True,scope=__doc__)
for path,value in [(a.fixture,fixture),(a.proof,proof)]:
 with path.open('x') as f:json.dump(value,f,indent=2);f.write('\n')
print(json.dumps(proof))
