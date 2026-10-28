"""Whole unchanged GAME.EXE Blink create/update, wake and FISTP roots.

Mode/buff reads, balance return, spatial choices, type lookup/allocation,
CreateAt, FX, sound/audio, teleport and attribution are declared services.
A scratch FLD/RET thunk supplies only the balance callee's double result.
Original loads, spills, conversion and control flow are never replaced.
"""
import argparse,collections,copy,hashlib,json,math,struct
from pathlib import Path
from unicorn import Uc,UC_ARCH_X86,UC_MODE_32,UC_HOOK_CODE,UC_HOOK_MEM_INVALID
from unicorn.x86_const import *
p=argparse.ArgumentParser(description=__doc__)
for key in ('game','fixture','proof'):p.add_argument('--'+key,type=Path,required=True)
p.add_argument('--control',type=lambda x:int(x,0),choices=[0xe3f,0x23f],default=0xe3f)
a=p.parse_args();blob=a.game.read_bytes();sha=lambda b:hashlib.sha256(b).hexdigest();assert sha(blob)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,60)[0];opt=pe+24;base=struct.unpack_from('<I',blob,opt+28)[0];table=opt+struct.unpack_from('<H',blob,pe+20)[0]
u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,(struct.unpack_from('<I',blob,opt+56)[0]+4095)&~4095)
for n in range(struct.unpack_from('<H',blob,pe+6)[0]):
 _,va,size,off=struct.unpack_from('<IIII',blob,table+40*n+8)
 if size:u.mem_write(base+va,blob[off:off+size])
heap,stack,code=0x21000000,0x22000000,0x23000000
for v in (heap,stack,code):u.mem_map(v,0x10000)
record=heap+0x1000;objects=[heap+0x2000+1024*i for i in range(6)];updates=[heap+0x7000,heap+0x7800];collide=[heap+0xa000,heap+0xa100];destination=heap+0xb000
sp,stop,init,thunk,scratch,cw=stack+0xe000,code+0x100,code+0x200,code+0x300,code+0x400,code+0x500
frameaddr,tickaddr,cacheaddr=0x84ea04,0x85b3fc,0x833c00
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw));u.mem_write(thunk,b'\xdd\x05'+struct.pack('<I',scratch)+b'\xc3')
get=lambda p:struct.unpack('<I',u.mem_read(p,4))[0]
put=lambda p,v:u.mem_write(p,struct.pack('<I',v&0xffffffff))
words=lambda p:[get(p),get(p+4)]
point=lambda x,y:[struct.unpack('<I',struct.pack('<f',n))[0] for n in (x,y)]
f64=lambda value:struct.pack('<d',value).hex()
ref=lambda i:0 if i<0 else objects[i]
idx=lambda p:-1 if p==0 else objects.index(p)
regions={'create':(0x530310,0x53037e),'update':(0x530380,0x530556),'wake':(0x530560,0x5305cf),'converter':(0x419a70,0x419a81)}
traces={name:{} for name in regions};entries=collections.Counter();services=collections.Counter();mutations=collections.Counter();faults=[]
current={};events=[];calls=collections.Counter();mutated=False;active=False;allowed=set()

def state():
 return dict(source=idx(get(record+12)),caster=idx(get(record+16)),target=idx(get(record+48)),spell=get(record+4),level=get(record+8),flag=get(record+20),timer=get(record+68),frame=get(frameaddr),tick=get(tickaddr),cache=get(cacheaddr),
  positions=[words(o+56) for o in objects],classes=[get(o+8) for o in objects],flags=[get(o+16) for o in objects],buffs=[get(o+340) for o in objects],
  gates=[idx(get(d+308)) for d in updates],wake_expiry=get(objects[4]+136),wake_data=collide.index(get(objects[4]+700)),wake_destinations=[words(d) for d in collide],destination=words(destination))
def allow(address,count):allowed.update(range(address-heap,address-heap+count))
def write(address,value):allow(address,4);put(address,value)
def writepoint(address,value):
 for i,v in enumerate(value):write(address+4*i,v)
def mutate(kind):
 global mutated
 at=calls[kind];calls[kind]+=1;m=current['mutation']
 if mutated or m['stage']!=kind or m['at']!=at:return
 mutated=True;mutations[kind+'/'+str(at)+'/'+m['action']]+=1;action=m['action']
 if action=='target-b':write(record+48,objects[3])
 elif action=='caster-b':write(record+16,objects[0])
 elif action=='source-b':write(record+12,objects[0])
 elif action=='spell':write(record+4,43)
 elif action=='position-a':writepoint(objects[2]+56,point(-17.5,23.25))
 elif action=='position-b':writepoint(objects[3]+56,point(-77.5,83.25))
 elif action=='invisible-a':write(objects[2]+340,get(objects[2]+340)|1)
 elif action=='visible-a':write(objects[2]+340,get(objects[2]+340)&~1)
 elif action=='collision-b':write(objects[4]+700,collide[1])
 elif action=='frame':put(frameaddr,0xfffffffd)
 elif action=='tick':put(tickaddr,73)
 elif action=='gate-b':write(updates[0]+308,objects[3])
 elif action=='destination':writepoint(destination,point(-27.5,33.25))
 else:raise AssertionError(action)
def event(kind,args):events.append(dict(kind=kind,args=args,state=state()));mutate(kind)
def ret(value=0):
 esp=u.reg_read(UC_X86_REG_ESP);u.reg_write(UC_X86_REG_EAX,value&0xffffffff);u.reg_write(UC_X86_REG_EIP,get(esp));u.reg_write(UC_X86_REG_ESP,esp+4)
def hook(_u,pc,length,_):
 if not active:return
 if pc==stop:u.emu_stop();return
 for name,(lo,hi) in regions.items():
  if lo<=pc<hi:
   traces[name][pc]=bytes(u.mem_read(pc,length))
   if pc==lo:entries[name]+=1
   return
 if thunk<=pc<thunk+7:return
 esp=u.reg_read(UC_X86_REG_ESP);arg=lambda n:get(esp+4*n);services[hex(pc)]+=1
 if pc==0x40a5c0:
  assert arg(1) in [2048,4096];ret(int(current['coop'] if arg(1)==2048 else current['quest']))
 elif pc==0x419d70:
  assert bytes(u.mem_read(arg(1),32)).split(b'\0',1)[0]==b'TeleportDelay'
  event('delay',[arg(2)]);u.mem_write(scratch,bytes.fromhex(current['delay_bits64']));u.reg_write(UC_X86_REG_EIP,thunk)
 elif pc==0x4ff350:
  assert arg(2) in [0,14];ret(int(bool(get(arg(1)+340)&(1<<arg(2)))))
 elif pc in [0x579f00,0x4f7ab0]:
  kind='waypoint' if pc==0x579f00 else 'start';event(kind,[idx(arg(2))]);writepoint(arg(1),current['destination']);ret(current['waypoint'] if kind=='waypoint' else 0x13572468)
 elif pc==0x4ed970:
  assert arg(1)==point(60,0)[0];event('random',[arg(1),*words(arg(2))]);writepoint(arg(3),current['destination']);ret(0x13572468)
 elif pc==0x4e3aa0:
  assert bytes(u.mem_read(arg(1),32)).split(b'\0',1)[0]==b'TeleportWake';event('lookup',[]);ret(current['lookup'])
 elif pc==0x4e3450:event('new',[arg(1)]);ret(objects[4] if current['wake'] else 0)
 elif pc==0x4daa50:
  assert idx(arg(1))==4 and arg(5)==0;event('create',[4,idx(arg(2)),arg(3),arg(4),arg(5)]);writepoint(arg(1)+56,[arg(3),arg(4)]);ret(0x24681357)
 elif pc==0x522ff0:
  assert arg(1)==137;event('fx',[arg(1),*words(arg(2))]);ret(0x87654321)
 elif pc==0x424800:
  assert arg(2)==0;event('sound',[arg(1),arg(2)]);ret(current['sound']+calls['sound']-1)
 elif pc==0x501960:event('audio',[arg(1),idx(arg(2)),arg(3),arg(4)]);ret(0x87654321)
 elif pc==0x4e7190:
  target=arg(1);dest=words(arg(2));event('teleport',[idx(target),*dest]);writepoint(target+56,dest);ret(1)
 elif pc==0x4e7540:event('attribution',[idx(arg(1)),idx(arg(2))]);ret(0x3456789a)
 else:raise AssertionError((current['name'],'undeclared instruction/service',hex(pc)))
def invalid(_u,access,address,size,value,_):
 faults.append(dict(name=current['name'],pc=hex(u.reg_read(UC_X86_REG_EIP)),address=hex(address)));return False
u.hook_add(UC_HOOK_CODE,hook);u.hook_add(UC_HOOK_MEM_INVALID,invalid)
seed=bytes((i*31+19)&255 for i in range(0x10000));registers={UC_X86_REG_EBX:0x13572468,UC_X86_REG_ESI:0x24681357,UC_X86_REG_EDI:0x98765432,UC_X86_REG_EBP:0xfedcba98}

def run(case):
 global current,events,calls,mutated,active,allowed
 active=False;current=case;events=[];calls=collections.Counter();mutated=False;allowed=set();u.mem_write(heap,seed)
 for off,value in [(4,4),(8,case['level']),(12,ref(case['source'])),(16,ref(case['caster'])),(20,case['flag']),(48,ref(case['target'])),(68,case['timer'])]:put(record+off,value)
 positions=[point(7,11),point(13,17),point(123.5,-456.25),point(88.25,99.5),point(0,0),point(22,33)]
 for i,obj in enumerate(objects):
  put(obj+8,case['classes'][i]);put(obj+16,case['flags'][i]);put(obj+340,case['buffs'][i]);put(obj+36,0x12340000+i)
  for n,v in enumerate(positions[i]):put(obj+56+4*n,v)
 for i in range(2):put(objects[i+2]+748,updates[i]);put(updates[i]+308,ref(case['gates'][i]))
 put(objects[4]+700,collide[0]);put(objects[4]+136,0x13579bdf)
 for address in collide:
  put(address,0xbf800000);put(address+4,0xc0000000)
 for n,v in enumerate(case['destination']):put(destination+4*n,v)
 for address,value in [(frameaddr,case['frame']),(tickaddr,case['tick']),(cacheaddr,case['cache'])]:u.mem_write(address-16,bytes(range(36)));put(address,value)
 globals_before={address:bytes(u.mem_read(address-16,36)) for address in [frameaddr,tickaddr,cacheaddr]}
 u.mem_write(cw,struct.pack('<H',a.control));u.emu_start(init,init+8,count=2)
 before=bytes(u.mem_read(heap,len(seed)));u.mem_write(sp-512,bytes(1024));put(sp,stop)
 if case['kind']=='wake':
  put(sp+4,objects[2]);put(sp+8,objects[2]+56);put(sp+12,objects[2]+56 if case['alias'] else destination)
 else:put(sp+4,record)
 for reg,value in registers.items():u.reg_write(reg,value)
 u.reg_write(UC_X86_REG_EFLAGS,0x202);u.reg_write(UC_X86_REG_ESP,sp);active=True;u.emu_start(regions[case['kind']][0],stop+1,count=20000);active=False
 assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4
 assert all(u.reg_read(reg)==value for reg,value in registers.items());assert get(sp)==stop
 assert bytes(u.mem_read(sp-512,256))==bytes(256) and bytes(u.mem_read(sp+(16 if case['kind']=='wake' else 8),256))==bytes(256)
 assert u.reg_read(UC_X86_REG_FPCW)==a.control and u.reg_read(UC_X86_REG_FPSW)&0x3800==0
 if case['kind']=='create':allow(record+68,4)
 if calls['new'] and case['wake']:allow(objects[4]+136,4);allow(collide[0],8)
 after=bytes(u.mem_read(heap,len(seed)));assert all(x==y or i in allowed for i,(x,y) in enumerate(zip(before,after)))
 for address,prior in globals_before.items():
  after=bytes(u.mem_read(address-16,36));assert all(x==y or 16<=i<20 for i,(x,y) in enumerate(zip(prior,after)))
 if case['mutation']['stage']:assert mutated,(case['name'],'mutation not reached')
 return dict(input=copy.deepcopy(case),result=dict(ret=None if case['kind']=='wake' else u.reg_read(UC_X86_REG_EAX),state=state(),events=copy.deepcopy(events)))

template=dict(name='',kind='update',quest=False,coop=False,flag=0,level=3,timer=101,frame=100,tick=30,source=1,caster=1,target=2,classes=[4,4,4,2,0x20,0x20],flags=[0]*6,buffs=[0]*6,gates=[-1,-1],cache=77,lookup=77,wake=True,alias=False,waypoint=1,destination=point(789.5,234.25),delay_bits64=f64(2.5),sound=351,mutation=dict(stage='',at=0,action=''))
cases=[]
def add(name,**changes):
 case=copy.deepcopy(template);case.update(name=name,**changes);cases.append(case)
for quest,coop,flag in [(True,True,1),(True,False,0),(False,False,0),(False,True,0)]:add(f'create/mode-{int(quest)}-{int(coop)}-{flag}',kind='create',quest=quest,coop=coop,flag=flag)
for value in [.49,.51,1.5,2.5,3.5,-1.5,-2.5,-3.5,1.999999999999,2.999999999999,2147483647.,-2147483647.,math.inf,math.nan]:add('create/delay-'+f64(value),kind='create',coop=True,delay_bits64=f64(value))
add('create/frame-wrap',kind='create',frame=0xffffffff)
for target,flag,timer in [(-1,0,101),(2,0x20,101),(2,0x8000,101),(2,0,102),(2,0,0)]:
 flags=[0]*6;flags[2]=flag;add(f'update/gate-{target}-{flag:x}-{timer}',target=target,flags=flags,timer=timer)
buffs=[0]*6;buffs[2]=1<<14;add('update/anchor',buffs=buffs)
add('update/visible')
for player in [False,True]:
 cls=[4,4,4 if player else 2,2,0x20,0x20];buffs=[0]*6;buffs[2]=1
 add('update/invisible-'+str(player),classes=cls,buffs=buffs)
add('update/start-fallback',waypoint=0)
add('update/quest-start',quest=True)
add('update/quest-gate',quest=True,gates=[5,-1])
add('update/no-wake',wake=False)
add('update/different-source',source=0)
add('update/nil-caster',source=0,caster=-1)
for stage,at,action in [('waypoint',0,'target-b'),('waypoint',0,'position-a'),('start',0,'target-b'),('random',0,'target-b'),('new',0,'position-a'),('new',0,'target-b'),('new',0,'caster-b'),('create',0,'target-b'),('create',0,'collision-b'),('create',0,'frame'),('create',0,'tick'),('fx',0,'target-b'),('fx',0,'position-a'),('sound',0,'target-b'),('sound',0,'spell'),('audio',0,'target-b'),('audio',0,'spell'),('audio',0,'invisible-a'),('fx',1,'target-b'),('teleport',0,'target-b'),('teleport',0,'invisible-a'),('fx',3,'caster-b'),('sound',1,'target-b'),('sound',1,'spell')]:
 changes=dict(mutation=dict(stage=stage,at=at,action=action))
 if stage=='start':changes['waypoint']=0
 if stage=='random':changes.update(quest=True,gates=[5,-1])
 add(f'update/{stage}-{at}-{action}',**changes)
for cache,wake,alias in [(0,True,False),(77,True,False),(0,False,False),(77,False,False),(77,True,True)]:add(f'wake/cache-{cache}-new-{int(wake)}-alias-{int(alias)}',kind='wake',cache=cache,wake=wake,alias=alias)
for stage,action in [('lookup','position-a'),('new','position-a'),('new','destination'),('create','frame'),('create','tick'),('create','collision-b')]:add(f'wake/{stage}-{action}',kind='wake',cache=0,mutation=dict(stage=stage,at=0,action=action))
rows=[run(case) for case in cases];sealed=[]
for name,(lo,hi) in regions.items():
 insns=traces[name];covered={pc+i for pc,b in insns.items() for i in range(len(b))};assert covered==set(range(lo,hi)),(name,sorted(set(range(lo,hi))-covered))
 sealed.append(dict(name=name,address=hex(lo),end_exclusive=hex(hi),bytes=hi-lo,sha256=sha(bytes(u.mem_read(lo,hi-lo))),instructions=[dict(pc=hex(pc),bytes=b.hex()) for pc,b in sorted(insns.items())]))
encoded=(json.dumps(dict(original_sha256=sha(blob),control_word=a.control,rows=rows),indent=2)+'\n').encode();a.fixture.write_bytes(encoded)
proof=dict(original_sha256=sha(blob),control_word=a.control,contexts=len(rows),root_contexts=sum(r['input']['kind']!='wake' for r in rows),helper_contexts=sum(r['input']['kind']=='wake' for r in rows),entries=dict(entries),regions=sealed,all_original_instruction_bytes_executed=True,zero_arithmetic_or_conversion_replacements=True,callee_saved_registers_stack_control_word_and_nonoutput_heap_global_canaries_preserved=True,services=dict(services),mutations=dict(mutations),faults=faults,fixture_sha256=sha(encoded),scope='Complete Blink create/update, reached wake helper and FISTP converter under declared services; no transitive callee/full campaign equivalence. Mode/buff reads pure; scratch FLD/RET supplies only balance return. Helper return register ignored by callers.')
assert not faults;a.proof.write_text(json.dumps(proof,indent=2)+'\n');print(json.dumps({k:v for k,v in proof.items() if k!='regions'},indent=2))
