"""Execute unchanged Retreat, health/predicate/cast, buff and Edibles bodies.

Related-spell selection, Push/Pop, RNG and food search are declared services.
GameFlags and Buff getters execute unchanged. Tick/frame observations are
read-only instruction hooks. Raw root EAX is retained; the stock update
dispatcher discards it. Ancillary nil-Buff execution covers its early RET.
"""
import argparse, collections, copy, hashlib, itertools, json, struct
from pathlib import Path
from unicorn import Uc, UcError, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_WRITE
from unicorn.x86_const import *

p=argparse.ArgumentParser(description=__doc__)
for name in ['game','fixture','proof']:p.add_argument('--'+name,type=Path,required=True)
p.add_argument('--control',type=lambda s:int(s,0),choices=[0xe3f,0x23f],required=True)
a=p.parse_args();blob=a.game.read_bytes();sha=lambda data:hashlib.sha256(data).hexdigest()
assert sha(blob)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,60)[0];op=pe+24;tb=op+struct.unpack_from('<H',blob,pe+20)[0];base=struct.unpack_from('<I',blob,op+28)[0]
u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,(struct.unpack_from('<I',blob,op+56)[0]+4095)&~4095)
for i in range(struct.unpack_from('<H',blob,pe+6)[0]):
 _,va,n,off=struct.unpack_from('<IIII',blob,tb+i*40+8)
 if n:u.mem_write(base+va,blob[off:off+n])
heap,stack,code=0x21000000,0x22000000,0x23000000
for address in [heap,stack,code]:u.mem_map(address,0x10000)
objects=[heap+0x1000,heap+0x9000,heap+0xa000,heap+0xb000]
unit=objects[0];health=[heap+0x3000,heap+0x4000];updates=[heap+0x5000,heap+0x7000];items=[heap+0xd000,heap+0xe000]
sp,stop,init,cw=stack+0xe000,code+0x100,code+0x200,code+0x300
rd=lambda address:struct.unpack('<I',u.mem_read(address,4))[0]
def wr(address,value):
 u.mem_write(address,struct.pack('<I',value&0xffffffff))
 if active:write_hook(u,0,address,4,value,None)
r16=lambda address:struct.unpack('<H',u.mem_read(address,2))[0]
w16=lambda address,value:u.mem_write(address,struct.pack('<H',value&0xffff))
signed=lambda value:struct.unpack('<i',struct.pack('<I',value&0xffffffff))[0]
ref=lambda value:objects.index(value) if value else -1
canonical=lambda value:0xf0000000+objects.index(value) if value in objects else value
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw))
ranges={0x545440:218,0x545520:83,0x545580:40,0x5455b0:43,0x4ff350:35,0x5455e0:197,0x40a5c0:18}
original={pc:bytes(u.mem_read(pc,n)) for pc,n in ranges.items()}
seed=bytes((i*31+19)&255 for i in range(0x10000));active=False;current={};events=[];traces={};services=collections.Counter();push_count=0;writes={}
registers={UC_X86_REG_EBX:0x12345678,UC_X86_REG_ESI:0x23456789,UC_X86_REG_EDI:0x3456789a,UC_X86_REG_EBP:0x456789ab}

def state():
 return dict(update=updates.index(rd(unit+748)),health=health.index(rd(unit+556)),cls=rd(unit+8),buffs=rd(unit+340),
  positions=[[rd(v+56),rd(v+60)] for v in objects],healths=[[r16(v),r16(v+4)] for v in health],
  statuses=[rd(v+1440) for v in updates],resume=[rd(v+1344) for v in updates],enemies=[ref(rd(v+1196)) for v in updates],
  items=[[canonical(rd(v+i*4)) for i in range(5)] for v in items],fps=rd(0x85b3fc),frame=rd(0x84ea04),flags=rd(0x5d53a4))

def mutate(stage):
 m=current['mutation']
 if m['stage']!=stage:return
 action,value=m['action'],m['value']
 if action in ['enemy','all']:
  wr(updates[0]+1196,objects[2 if value else 1]);wr(updates[1]+1196,objects[1 if value else 2])
 if action=='nil-enemy':wr(updates[0]+1196,0)
 if action in ['update','all']:wr(unit+748,updates[value])
 if action in ['health-pointer','all']:wr(unit+556,health[value])
 if action in ['health','all']:
  for v in health:w16(v,100 if value else 1);w16(v+4,100 if value else 0)
 if action in ['status','all']:wr(updates[0]+1440,0x20 if value else 0);wr(updates[1]+1440,0 if value else 0x20)
 if action in ['resume','all']:wr(updates[0]+1344,0x3f400000 if value else 0x7fc12345);wr(updates[1]+1344,0x3e800000)
 if action in ['buffs','all']:wr(unit+340,0x20000000 if value else 0)
 if action in ['class','all']:wr(unit+8,0x400000 if value else 2)
 if action in ['positions','all']:
  for i,v in enumerate(objects):wr(v+56,0x80000000 if i&1 else 0x7fc54321);wr(v+60,0x42c80001+i)
 if action in ['fps','all']:wr(0x85b3fc,0x40000001 if value else 0xffffffff)
 if action in ['frame','all']:wr(0x84ea04,0xfffffffd if value else 17)
 if action in ['flags','all']:wr(0x5d53a4,0 if value else 4096)
 if action in ['item-args','all']:
  for i,v in enumerate(items):
   for j in range(1,5):wr(v+j*4,0x6b000000+i*256+j)

def write_hook(_u,_access,address,size,_value,_):
 if not active:return
 for i,v in enumerate(items):
  if v<=address and address+size<=v+20:
   assert (size==4 and (address-v)%4==0) or (size==1 and address==v+12),(hex(address),size)
   writes.setdefault(str(i)+':'+str((address-v)//4),[]).append(dict(size=size,value=canonical(_value&((1<<(size*8))-1))))

def hook(_u,pc,size,_):
 global push_count
 if not active:return
 if pc==stop:u.emu_stop();return
 if any(start<=pc<start+n for start,n in ranges.items()):
  traces[pc]=bytes(u.mem_read(pc,size))
  if pc in [0x5454a5,0x5454c7]:events.append(dict(kind='tick' if pc==0x5454a5 else 'frame',args=[],state=state()))
  return
 esp=u.reg_read(UC_X86_REG_ESP)
 if pc==0x541050:
  assert rd(esp+4)==unit;kind,args,ret='related',[],current['related']
 elif pc==0x50a160:
  assert rd(esp+4)==unit;kind,args,ret='pop',[],current['pop_return']
 elif pc==0x50a260:
  assert rd(esp+4)==unit;push_count+=1;action=rd(esp+8)
  kind,args='push',[action]
  ret=items[(push_count-1)%2] if current['accept']&(1<<(push_count-1)) else 0
 elif pc==0x415fa0:
  assert rd(esp+12)==0x5cc500 and rd(esp+16)==149
  kind,args,ret='random',[signed(rd(esp+4)),signed(rd(esp+8)),149],current['random']
 elif pc==0x544a00:
  assert rd(esp+4)==unit;kind,args,ret='search',[rd(esp+8)],objects[current['food']] if current['food']>=0 else 0
 else:raise AssertionError((current['name'],hex(pc)))
 events.append(dict(kind=kind,args=args,state=state()));services[kind]+=1
 mutate('push'+str(push_count) if kind=='push' else kind)
 if kind=='push' and ret:wr(ret,args[0])
 u.reg_write(UC_X86_REG_EAX,ret&0xffffffff);u.reg_write(UC_X86_REG_EIP,rd(esp));u.reg_write(UC_X86_REG_ESP,esp+4)

u.hook_add(UC_HOOK_CODE,hook);u.hook_add(UC_HOOK_MEM_WRITE,write_hook)
allowed=set()
for ptr,off,n in [(unit,8,4),(unit,340,4),(unit,556,4),(unit,748,4)]+[(v,56,8) for v in objects]+[(v,0,2) for v in health]+[(v,4,2) for v in health]+[(v,1196,4) for v in updates]+[(v,1344,4) for v in updates]+[(v,1440,4) for v in updates]+[(v,0,20) for v in items]:allowed.update(range(ptr+off,ptr+off+n))

def run(c):
 global active,current,events,push_count,writes
 active=False;current=c;events=[];push_count=0;writes={};u.mem_write(heap,seed)
 wr(unit+8,2);wr(unit+340,c['buffs']);wr(unit+556,health[0]);wr(unit+748,updates[0])
 for i,v in enumerate(objects):wr(v+56,c['positions'][i][0]);wr(v+60,c['positions'][i][1])
 for i,v in enumerate(health):w16(v,c['cur'] if i==0 else 75);w16(v+4,c['max'] if i==0 else 100)
 for i,v in enumerate(updates):wr(v+1440,c['status'] if i==0 else 0x20);wr(v+1344,c['resume'] if i==0 else 0x3f400000);wr(v+1196,objects[c['enemy']] if i==0 and c['enemy']>=0 else (0 if i==0 else objects[2]))
 for i,v in enumerate(items):
  wr(v,0xa5a50000+i)
  for j in range(1,5):wr(v+j*4,0x5a000000+i*256+j)
 wr(0x85b3fc,c['fps']);wr(0x84ea04,c['frame']);wr(0x5d53a4,4096 if c['quest'] else 0)
 before=bytes(u.mem_read(heap,len(seed)));wr(cw,a.control);u.emu_start(init,init+8,count=2);u.mem_write(sp-512,bytes(1024));wr(sp,stop);wr(sp+4,0 if c['entry']=='nil-buff' else unit);wr(sp+8,29 if c['entry']=='nil-buff' else 0)
 for reg,value in registers.items():u.reg_write(reg,value)
 u.reg_write(UC_X86_REG_ESP,sp);u.reg_write(UC_X86_REG_EFLAGS,0x202);active=True;fault=False
 try:u.emu_start(0x4ff350 if c['entry']=='nil-buff' else 0x545440,stop+1,count=10000)
 except UcError:
  assert c['mutation']['action']=='nil-enemy' and u.reg_read(UC_X86_REG_EIP)==0x5454ed,c['name'];fault=True
 active=False
 if not fault:
  assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4,c['name']
  assert all(u.reg_read(reg)==value for reg,value in registers.items()),c['name']
 assert u.reg_read(UC_X86_REG_FPCW)==a.control and u.reg_read(UC_X86_REG_FPSW)&0x3800==0,c['name']
 assert bytes(u.mem_read(sp-512,256))==bytes(256) and bytes(u.mem_read(sp+12,256))==bytes(256),c['name']
 after=bytes(u.mem_read(heap,len(seed)));assert all(x==y or heap+i in allowed for i,(x,y) in enumerate(zip(before,after))),c['name']
 eax=u.reg_read(UC_X86_REG_EAX)
 return dict(input=c,result=dict(events=copy.deepcopy(events),state=state(),fault=fault,item_writes=dict(writes)),raw_original_EAX=eax,raw_original_AL=eax&255)

cases=[]
def case(prefix,**kw):
 c=dict(entry='root',cur=10,max=100,resume=0x3f000000,status=0,buffs=0,enemy=1,related=0,food=-1,quest=False,accept=31,fps=30,frame=1000,random=151,pop_return=0x87654321,positions=[[0x42c80000,0x43480000],[0x43960000,0x43c80000],[0x43fa0000,0x44160000],[0x442f0000,0x44480000]],mutation=dict(stage='',action='',value=0))
 c.update(kw,name=prefix+'-'+str(len(cases)));cases.append(c)
for cur,status,buffs,enemy,related in itertools.product([0,49,50,51,100],[0,0x20,0x10000,0x10020,0xffffffff,0xffffffdf],[0,0x20000000,0xffffffff,0xdfffffff],[-1,1],[0,0x87654321]):
 case('root-branch',cur=cur,status=status,buffs=buffs,enemy=enemy,related=related)
for cur,maximum,resume in itertools.product([0,1,2,32767,65534,65535],[0,1,3,100,65535],[0,0x80000000,1,0x3dcccccd,0x3eaaaaab,0x3f000000,0x3f800000,0x477fff00,0x7f800000,0xff800000,0x7fc12345,0x7fa12345]):
 case('health-bits',cur=cur,max=maximum,resume=resume,enemy=-1)
for food,quest,accept in itertools.product([-1,0,1,3],[False,True],range(32)):
 case('edibles-push',enemy=-1,food=food,quest=quest,accept=accept)
for accept,fps,frame,random in itertools.product(range(4),[0,30,0x2aaaaaab,0x40000000,0x40000001,0x80000000,0xffffffff],[0,0xfffffffd],[151,-7]):
 case('fallback-DWORD',accept=accept,fps=fps,frame=frame,random=random)
for stage,action,value in itertools.product(['related','push1','push2','random','search','push3','push4','push5','pop'],['enemy','update','health-pointer','health','status','resume','buffs','class','positions','fps','frame','flags','item-args','all'],range(2)):
 case('live-callback',enemy=-1 if stage in ['search','push3','push4','push5'] else 1,cur=100 if stage=='pop' else 10,food=3,mutation=dict(stage=stage,action=action,value=value))
for stage in ['related','push1','push2','random']:
 case('nil-enemy-fault',mutation=dict(stage=stage,action='nil-enemy',value=0))
case('nil-buff-ancillary',entry='nil-buff')
rows=[run(c) for c in cases]
assert all(bytes(u.mem_read(pc,len(raw)))==raw for pc,raw in original.items())
visited=set()
for pc,raw in traces.items():visited.update(range(pc,pc+len(raw)))
expected=set()
for pc,n in ranges.items():expected.update(range(pc,pc+n))
assert visited==expected,[(hex(pc),hex(pc+1)) for pc in sorted(expected-visited)]
fixture=dict(original_sha256=sha(blob),control_word=a.control,rows=rows)
proof=dict(original_sha256=sha(blob),control_word=a.control,contexts=len(rows),fault_contexts=sum(r['result']['fault'] for r in rows),root_bytes={hex(pc):n for pc,n in ranges.items()},root_sha256={hex(pc):sha(raw) for pc,raw in original.items()},executed_bytes=len(visited),original_code_unmodified=True,register_stack_FPCW_TOP_heap_canaries_preserved=True,raw_root_EAX_AL_retained_and_not_substituted_with_handled_bool=True,declared_services=dict(services),native_argument_store_widths_come_from_original_memory_write_trace=True,pure_tick_frame_queries_are_observed_without_substitution=True,one_F32_constant=dict(address='0x583068',bytes=bytes(u.mem_read(0x583068,4)).hex()),instructions=[dict(pc=hex(pc),bytes=raw.hex()) for pc,raw in sorted(traces.items())])
for path,value in [(a.fixture,fixture),(a.proof,proof)]:
 with path.open('x') as f:json.dump(value,f,indent=2);f.write('\n')
print(json.dumps({k:v for k,v in proof.items() if k!='instructions'}))
