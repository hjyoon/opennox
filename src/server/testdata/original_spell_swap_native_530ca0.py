"""Execute unchanged original Swap create/update, FISTP and rectangle roots.
Only pure mode/buff reads, balance return, LOS, message, teleport, FX,
sound/audio and attribution are declared services. No original caller,
rectangle comparison, arithmetic, spill or conversion instruction is replaced.
"""
import argparse,collections,copy,hashlib,json,math,struct
from pathlib import Path
from unicorn import Uc,UC_ARCH_X86,UC_MODE_32,UC_HOOK_CODE,UC_HOOK_MEM_INVALID
from unicorn.x86_const import *
p=argparse.ArgumentParser(description=__doc__)
for key in ['game','fixture','proof']:p.add_argument('--'+key,type=Path,required=True)
p.add_argument('--control',type=lambda s:int(s,0),choices=[0xe3f,0x23f],default=0xe3f)
a=p.parse_args();blob=a.game.read_bytes();sha=lambda b:hashlib.sha256(b).hexdigest();assert sha(blob)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,60)[0];opt=pe+24;base=struct.unpack_from('<I',blob,opt+28)[0];table=opt+struct.unpack_from('<H',blob,pe+20)[0]
u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,(struct.unpack_from('<I',blob,opt+56)[0]+4095)&~4095)
for n in range(struct.unpack_from('<H',blob,pe+6)[0]):
 _,va,size,off=struct.unpack_from('<IIII',blob,table+40*n+8)
 if size:u.mem_write(base+va,blob[off:off+size])
heap,stack,code=0x21000000,0x22000000,0x23000000
for addr in [heap,stack,code]:u.mem_map(addr,0x10000)
record=heap+0x1000;objects=[heap+0x2000+0x400*i for i in range(4)];updates=[heap+0x5000+0x600*i for i in range(4)];players=[heap+0x9000+0x100*i for i in range(4)]
sp,stop,init,thunk,scratch,cw=stack+0xe000,code+0x100,code+0x200,code+0x300,code+0x400,code+0x500;frameaddr=0x84ea04
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw));u.mem_write(thunk,b'\xdd\x05'+struct.pack('<I',scratch)+b'\xc3')
get=lambda addr:struct.unpack('<I',u.mem_read(addr,4))[0]
put=lambda addr,value:u.mem_write(addr,struct.pack('<I',value&0xffffffff))
u16=lambda addr:struct.unpack('<H',u.mem_read(addr,2))[0]
put16=lambda addr,value:u.mem_write(addr,struct.pack('<H',value))
words=lambda addr:[get(addr),get(addr+4)]
f32=lambda value:struct.unpack('<I',struct.pack('<f',value))[0]
point=lambda x,y:[f32(x),f32(y)]
f64=lambda value:struct.pack('<d',value).hex()
ref=lambda n:0 if n<0 else objects[n]
idx=lambda address:-1 if address==0 else objects.index(address)
regions={'create':(0x530ca0,0x530d27),'update':(0x530d30,0x530f8d),'converter':(0x419a70,0x419a81),'rectangle':(0x428220,0x428262)}
traces={n:{} for n in regions};entries=collections.Counter();services=collections.Counter();mutations=collections.Counter();faults=[]
active=False;current={};events=[];calls=collections.Counter();mutated=False;allowed=set()
def state():
 return dict(source=idx(get(record+12)),caster=idx(get(record+16)),target=idx(get(record+48)),spell=get(record+4),level=get(record+8),flag=get(record+20),flags88=get(record+88),timer=get(record+68),frame=get(frameaddr),
  positions=[words(o+56) for o in objects],classes=[get(o+8) for o in objects],subclasses=[get(o+12) for o in objects],flags=[get(o+16) for o in objects],buffs=[get(o+340) for o in objects],
  object_updates=[updates.index(get(o+748)) for o in objects],update_players=[players.index(get(d+276)) for d in updates],dimensions=[[u16(d+10),u16(d+12)] for d in players])
def allow(addr,size):allowed.update(range(addr-heap,addr-heap+size))
def write(addr,value):allow(addr,4);put(addr,value)
def writepoint(addr,value):
 for i,v in enumerate(value):write(addr+4*i,v)
def mutate(kind):
 global mutated
 at=calls[kind];calls[kind]+=1;m=current['mutation']
 if mutated or m['stage']!=kind or m['at']!=at:return
 mutated=True;action=m['action'];mutations[kind+'/'+str(at)+'/'+action]+=1
 if action=='target-b':write(record+48,objects[3])
 elif action=='caster-b':write(record+16,objects[0])
 elif action=='spell':write(record+4,43)
 elif action=='position-target':writepoint(objects[2]+56,point(-17.5,23.25))
 elif action=='position-caster':writepoint(objects[1]+56,point(-77.5,83.25))
 elif action=='invisible-target':write(objects[2]+340,get(objects[2]+340)|1)
 elif action=='invisible-caster':write(objects[1]+340,get(objects[1]+340)|1)
 elif action=='visible-target':write(objects[2]+340,get(objects[2]+340)&~1)
 elif action=='frame':put(frameaddr,0xfffffffd)
 elif action=='update-caster':write(objects[1]+748,updates[0])
 elif action=='player-caster':write(updates[1]+276,players[0])
 elif action=='dimensions':
  allow(players[1]+10,4);put16(players[1]+10,65535);put16(players[1]+12,1)
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
 if pc==0x40a5c0:assert arg(1)==2048;ret(int(current['coop']))
 elif pc==0x419d70:
  assert bytes(u.mem_read(arg(1),32)).split(b'\0',1)[0]==b'TeleportDelay';event('delay',[arg(2)]);u.mem_write(scratch,bytes.fromhex(current['delay_bits64']));u.reg_write(UC_X86_REG_EIP,thunk)
 elif pc==0x4ff350:assert arg(2) in [0,14];ret(int(bool(get(arg(1)+340)&(1<<arg(2)))))
 elif pc==0x5370e0:assert arg(3)==0;event('interact',[idx(arg(1)),idx(arg(2)),0]);ret(int(current['interacts']))
 elif pc==0x4da2c0:
  assert arg(3)==0 and bytes(u.mem_read(arg(2),32)).split(b'\0',1)[0]==b'ExecDur.c:NeedClearLOSForSwap';event('inform',[idx(arg(1))]);ret(0x13572468)
 elif pc==0x4e7190:
  target,dest=arg(1),words(arg(2));event('teleport',[idx(target),*dest]);writepoint(target+56,dest);ret(1)
 elif pc==0x522ff0:assert arg(1)==137;event('fx',[arg(1),*words(arg(2))]);ret(0x24681357)
 elif pc==0x424800:assert arg(2)==0;event('sound',[arg(1),0]);ret(current['sound']+calls['sound']-1)
 elif pc==0x501960:event('audio',[arg(1),idx(arg(2)),arg(3),arg(4)]);ret(0x13572468)
 elif pc==0x4e7540:event('attribution',[idx(arg(1)),idx(arg(2))]);ret(0x24681357)
 else:raise AssertionError((current['name'],'undeclared instruction/service',hex(pc)))
def invalid(_u,access,address,size,value,_):faults.append(dict(name=current['name'],pc=hex(u.reg_read(UC_X86_REG_EIP)),address=hex(address)));return False
u.hook_add(UC_HOOK_CODE,hook);u.hook_add(UC_HOOK_MEM_INVALID,invalid)
seed=bytes((i*31+19)&255 for i in range(0x10000));registers={UC_X86_REG_EBX:0x13572468,UC_X86_REG_ESI:0x24681357,UC_X86_REG_EDI:0x98765432,UC_X86_REG_EBP:0xfedcba98}
def run(case):
 global current,events,calls,mutated,active,allowed
 active=False;current=case;events=[];calls=collections.Counter();mutated=False;allowed=set();u.mem_write(heap,seed)
 for off,value in [(4,115),(8,case['level']),(12,ref(case['source'])),(16,ref(case['caster'])),(20,case['flag']),(48,ref(case['target'])),(68,case['timer']),(88,case['flags88'])]:put(record+off,value)
 for i,obj in enumerate(objects):
  for off,value in [(8,case['classes'][i]),(12,case['subclasses'][i]),(16,case['flags'][i]),(340,case['buffs'][i]),(748,updates[i])]:put(obj+off,value)
  for j,value in enumerate(case['positions'][i]):put(obj+56+j*4,value)
  put(updates[i]+276,players[i]);put16(players[i]+10,case['dimensions'][i][0]);put16(players[i]+12,case['dimensions'][i][1])
 u.mem_write(frameaddr-16,bytes(range(36)));put(frameaddr,case['frame']);global_before=bytes(u.mem_read(frameaddr-16,36))
 u.mem_write(cw,struct.pack('<H',a.control));u.emu_start(init,init+8,count=2);before=bytes(u.mem_read(heap,len(seed)));u.mem_write(sp-512,bytes(1024));put(sp,stop);put(sp+4,record)
 for reg,value in registers.items():u.reg_write(reg,value)
 u.reg_write(UC_X86_REG_EFLAGS,0x202);u.reg_write(UC_X86_REG_ESP,sp);active=True;u.emu_start(regions[case['kind']][0],stop+1,count=20000);active=False
 assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4 and get(sp)==stop
 assert all(u.reg_read(reg)==value for reg,value in registers.items())
 assert bytes(u.mem_read(sp-512,256))==bytes(256) and bytes(u.mem_read(sp+8,256))==bytes(256)
 assert u.reg_read(UC_X86_REG_FPCW)==a.control and u.reg_read(UC_X86_REG_FPSW)&0x3800==0
 if case['kind']=='create':allow(record+68,4)
 after=bytes(u.mem_read(heap,len(seed)));assert all(x==y or i in allowed for i,(x,y) in enumerate(zip(before,after)))
 after=bytes(u.mem_read(frameaddr-16,36));assert all(x==y or 16<=i<20 for i,(x,y) in enumerate(zip(global_before,after)))
 if case['mutation']['stage']:assert mutated,(case['name'],'mutation not reached')
 return dict(input=copy.deepcopy(case),result=dict(ret=u.reg_read(UC_X86_REG_EAX),state=state(),events=copy.deepcopy(events)))
template=dict(name='',kind='update',coop=False,flag=0,flags88=0,level=3,timer=101,frame=100,source=0,caster=1,target=2,classes=[4,4,4,2],subclasses=[0]*4,flags=[0]*4,buffs=[0]*4,positions=[point(7,11),point(100.25,-100.5),point(110.25,-80.5),point(104,-90)],dimensions=[[7,11],[10,20],[21,23],[31,37]],interacts=True,delay_bits64=f64(2.5),sound=351,mutation=dict(stage='',at=0,action=''))
cases=[]
def add(name,**changes):case=copy.deepcopy(template);case.update(changes);case['name']=case['kind']+'/'+name;cases.append(case)
def create(name,**changes):changes.update(kind='create',timer=0x87654321);add(name,**changes)
create('normal');create('coop',coop=True);create('null-caster',caster=-1);create('busy',flags88=0x77556620);create('flags-noise',flags88=0x77556601);create('null-target',target=-1);create('same',caster=2)
for flag in [0x20,0x8000]:flags=[0]*4;flags[2]=flag;create('flags-'+hex(flag),flags=flags)
for cls,sub in [(2,0),(2,0x4000),(4,0x4000)]:
 classes=list(template['classes']);classes[2]=cls;subs=[0]*4;subs[2]=sub;create('class-sub-'+str(cls)+'-'+str(sub),classes=classes,subclasses=subs)
for value in [.49,.51,1.5,2.5,3.5,-1.5,-2.5,-3.5,1.999999999999,2.999999999999,2147483647.,-2147483647.,math.inf,math.nan]:create('delay-'+f64(value),coop=True,delay_bits64=f64(value))
for level in [0,19,0xffffffff]:create('level-'+str(level),coop=True,level=level)
create('frame-wrap',frame=0xffffffff)
for action in ['frame','target-b','caster-b','spell']:create('delay-'+action,coop=True,mutation=dict(stage='delay',at=0,action=action))
for caster,target in [(-1,2),(1,-1)]:add('nil-'+str(caster)+'-'+str(target),caster=caster,target=target)
for unit in [1,2]:
 for flag in [0x20,0x8000]:flags=[0]*4;flags[unit]=flag;add('dead-'+str(unit)+'-'+str(flag),flags=flags)
add('wrong-frame',timer=103);add('zero-timer',timer=0);add('frame-wrap',timer=0,frame=0xffffffff)
for unit in [1,2]:buffs=[0]*4;buffs[unit]=1<<14;add('anchor-'+str(unit),buffs=buffs)
add('visible');add('no-interaction',interacts=False);add('glyph',flag=1,interacts=False)
classes=list(template['classes']);classes[1]=2;add('monster-caster',classes=classes)
for unit in [1,2]:buffs=[0]*4;buffs[unit]=1;add('invisible-'+str(unit),buffs=buffs)
for axis in [0,1]:
 for upper in [False,True]:
  pos=copy.deepcopy(template['positions']);pointf=[110.25,-80.5];pointf[axis]=([100.25,-100.5][axis]+([-10,-20][axis] if not upper else [10,20][axis]))+(-1 if not upper else 1);pos[2]=point(*pointf);add('outside-'+str(axis)+'-'+str(upper),positions=pos)
for stage,at,action in [('interact',0,'target-b'),('interact',0,'caster-b'),('interact',0,'position-caster'),('interact',0,'update-caster'),('interact',0,'player-caster'),('interact',0,'dimensions'),('teleport',0,'target-b'),('teleport',0,'caster-b'),('teleport',0,'position-target'),('teleport',0,'position-caster'),('teleport',0,'invisible-target'),('teleport',0,'invisible-caster'),('teleport',1,'target-b'),('teleport',1,'caster-b'),('teleport',1,'position-target'),('teleport',1,'invisible-target'),('fx',0,'target-b'),('fx',0,'caster-b'),('fx',0,'position-target'),('fx',0,'position-caster'),('fx',1,'target-b'),('fx',1,'caster-b'),('sound',0,'target-b'),('sound',0,'caster-b'),('sound',0,'spell'),('sound',1,'target-b'),('sound',1,'caster-b'),('sound',1,'spell'),('audio',0,'target-b'),('audio',0,'caster-b'),('audio',0,'spell'),('audio',1,'target-b'),('audio',1,'caster-b')]:add(stage+'-'+str(at)+'-'+action,mutation=dict(stage=stage,at=at,action=action))
add('interaction-false-new-caster',interacts=False,mutation=dict(stage='interact',at=0,action='caster-b'))
buffs=[0]*4;buffs[2]=1<<14;add('anchor-audio-new-caster',buffs=buffs,mutation=dict(stage='audio',at=0,action='caster-b'))
# Generate inputs near representable bounds; the original instructions alone
# determine their outcomes. Neither Python bound arithmetic nor Go is an oracle.
for axis in [0,1]:
 for value,half in [(0.1,1),(0.2,1),(0.51,31),(33554432.,3),(1e30,1),(-1e30,1)]:
  center=struct.unpack('<f',struct.pack('<I',f32(value)))[0]
  for sign in [-1,1]:
   edge=f32(center+sign*half)
   for delta in [-1,0,1]:
    pos=copy.deepcopy(template['positions']);pos[1]=point(0,0);pos[1][axis]=f32(value);pos[2]=point(0,0);pos[2][axis]=(edge+delta)&0xffffffff
    dims=copy.deepcopy(template['dimensions']);dims[1]=[1,1];dims[1][axis]=half
    add('bound-'+str(axis)+'-'+hex(f32(value))+'-'+str(half)+'-'+str(sign)+'-'+str(delta),positions=pos,dimensions=dims)
for raw in [0x00000000,0x80000000,1,0x80000001,0x7f800000,0xff800000,0x7fc00000]:
 pos=copy.deepcopy(template['positions']);pos[1]=[raw,raw];pos[2]=[raw,raw];add('special-'+hex(raw),positions=pos)
rows=[run(case) for case in cases];sealed=[]
for name,(lo,hi) in regions.items():
 insns=traces[name];covered={pc+i for pc,b in insns.items() for i in range(len(b))};assert covered==set(range(lo,hi)),(name,sorted(set(range(lo,hi))-covered))
 sealed.append(dict(name=name,address=hex(lo),end_exclusive=hex(hi),bytes=hi-lo,sha256=sha(bytes(u.mem_read(lo,hi-lo))),instructions=[dict(pc=hex(pc),bytes=b.hex()) for pc,b in sorted(insns.items())]))
encoded=(json.dumps(dict(original_sha256=sha(blob),control_word=a.control,rows=rows),indent=2)+'\n').encode();a.fixture.write_bytes(encoded)
proof=dict(original_sha256=sha(blob),control_word=a.control,contexts=len(rows),entries=dict(entries),regions=sealed,all_original_instruction_bytes_executed=True,zero_arithmetic_comparison_or_conversion_replacements=True,callee_saved_registers_stack_control_word_TOP_and_nonoutput_heap_global_canaries_preserved=True,services=dict(services),mutations=dict(mutations),faults=faults,fixture_sha256=sha(encoded),scope='Complete original Swap create/update and reached converter/rectangle roots under declared services. Pure mode/buff reads; balance uses only scratch FLD/RET return. No transitive-callee or full-campaign equivalence claim.')
assert not faults;a.proof.write_text(json.dumps(proof,indent=2)+'\n');print(json.dumps({k:v for k,v in proof.items() if k not in ['regions','services','mutations']},indent=2));print('mutations',len(mutations),sum(mutations.values()))
