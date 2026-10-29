"""Execute unchanged Drain Mana roots and actual small helpers.

Update/transfer/select/candidate/has-mana, 419A70, buff/recent/old/max-mana,
team presence/comparison and sync execute original machine code. Binary64
balance/table, enumeration, enemy/same-team/trace, Quest, ray/audio and mana are
declared services. Full native-mapped state and service arguments are observed.
All original arithmetic, spill, conversion, packing and live reloads execute.
This is not transitive full-game equivalence.
"""
import argparse, collections, copy, hashlib, json, struct
from pathlib import Path
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import *

p=argparse.ArgumentParser(description=__doc__)
for key in ['game','fixture','proof']:p.add_argument('--'+key,type=Path,required=True)
p.add_argument('--control',type=lambda v:int(v,0),choices=[0xe3f,0x23f],default=0xe3f)
a=p.parse_args();blob=a.game.read_bytes();sha=lambda b:hashlib.sha256(b).hexdigest()
assert sha(blob)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,60)[0];opt=pe+24;base=struct.unpack_from('<I',blob,opt+28)[0];table=opt+struct.unpack_from('<H',blob,pe+20)[0]
u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,(struct.unpack_from('<I',blob,opt+56)[0]+4095)&~4095)
for n in range(struct.unpack_from('<H',blob,pe+6)[0]):
 _,va,size,off=struct.unpack_from('<IIII',blob,table+40*n+8)
 if size:u.mem_write(base+va,blob[off:off+size])
heap,stack,code=0x21000000,0x22000000,0x23000000
for addr in [heap,stack,code]:u.mem_map(addr,0x40000)
record=heap+0x1000;objects=[heap+0x2000+0x400*i for i in range(4)];updates=[heap+0x10000+0x800*i for i in range(4)];infos=[heap+0x20000+0x1000*i for i in range(4)];health=[heap+0x30000+0x20*i for i in range(4)]
sp,stop,init,thunk,temp,cw,callback_end=stack+0x3e000,code+0x100,code+0x200,code+0x300,code+0x400,code+0x500,code+0x600
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw));u.mem_write(thunk,b'\xdd\x05'+struct.pack('<I',temp)+b'\xc3')
get=lambda addr:struct.unpack('<I',u.mem_read(addr,4))[0]
get16=lambda addr:struct.unpack('<H',u.mem_read(addr,2))[0]
get8=lambda addr:u.mem_read(addr,1)[0]
put=lambda addr,v:u.mem_write(addr,struct.pack('<I',v&0xffffffff))
put16=lambda addr,v:u.mem_write(addr,struct.pack('<H',v&0xffff))
put8=lambda addr,v:u.mem_write(addr,bytes([v&255]))
f32=lambda v:struct.unpack('<I',struct.pack('<f',v))[0]
f64=lambda v:struct.pack('<d',v).hex()
signed=lambda v:((v&0xffffffff)^0x80000000)-0x80000000
ref=lambda i:0 if i<0 else objects[i]
index=lambda v:-1 if v==0 else objects.index(v)
upindex=lambda v:updates.index(v)
infoindex=lambda v:infos.index(v)
point=lambda addr:[get(addr),get(addr+4)]
bits=lambda x,y:[f32(x),f32(y)]
cstr=lambda addr:bytes(u.mem_read(addr,128)).split(b'\0',1)[0].decode('ascii')
fps,frame,selected,center,nearest=0x85b3fc,0x84ea04,0x833ba8,0x833bb0,0x833bd8
multiplier_addresses=[0x5d35d4,0x5d35f4,0x5d35e4]
regions={'update':(0x52e210,0x52e441),'transfer':(0x52e450,0x52e602),'select':(0x52e610,0x52e660),'candidate':(0x52e660,0x52e7b6),'has-mana':(0x52e7c0,0x52e819),'converter':(0x419a70,0x419a81)}
helpers={'buff':(0x4ff350,0x4ff373),'recent':(0x4e6bd0,0x4e6bfa),'old-mana':(0x4eec80,0x4eeca7),'max-mana':(0x4eecb0,0x4eeccd),'has-team':(0x419130,0x419146),'compare-teams':(0x419150,0x41917b),'sync':(0x4e44f0,0x4e44ff)}
traces={n:{} for n in dict(regions,**helpers)};entries,services,mutations=collections.Counter(),collections.Counter(),collections.Counter()
active,current,events,calls,changed,allowed,faults,circle,quest=False,{},[],collections.Counter(),False,set(),[],None,0

def state():
 return dict(caster=index(get(record+16)),source=index(get(record+12)),glyph=get(record+20),target=index(get(record+48)),ray=index(get(record+36)),carry=get(record+72),expiry=get(record+68),level=get(record+8),position=point(record+28),selected=index(get(selected)),center=point(center),nearest=get(nearest),fps=get(fps),frame=get(frame),quest=quest,classes=[get(o+8) for o in objects],subclasses=[get(o+12) for o in objects],flags=[get(o+16) for o in objects],team_ids=[get8(o+52) for o in objects],sync_flags=[get(o+152) for o in objects],positions=[point(o+56) for o in objects],buffs=[get(o+340) for o in objects],health=[int(get(o+556)!=0) for o in objects],damage_frames=[get(o+536) for o in objects],updates=[upindex(get(o+748)) for o in objects],mana=[get16(d+4) for d in updates],maximum=[get16(d+8) for d in updates],obelisk_mana=[signed(get(d)) for d in updates],monster_status=[get(d+1440) for d in updates],players=[infoindex(get(d+276)) for d in updates],player_classes=[get8(i+2251) for i in infos],player_flags=[get(i+3680) for i in infos],multipliers=[get(i) for i in multiplier_addresses])
def allow(addr,size):allowed.update(range(addr,addr+size))
def write(addr,v):allow(addr,4);put(addr,v)
def write16(addr,v):allow(addr,2);put16(addr,v)
def write8(addr,v):allow(addr,1);put8(addr,v)
def writepoint(addr,v):
 for j,b in enumerate(v):write(addr+4*j,b)
def mutate(kind):
 global changed,quest
 at=calls[kind];calls[kind]+=1;m=current['mutation']
 if changed or m['stage']!=kind or m['at']!=at:return
 changed=True;action=m['action'];mutations[kind+'/'+str(at)+'/'+action]+=1
 if action=='caster-b':write(record+16,objects[2])
 elif action=='caster-nil':write(record+16,0)
 elif action=='source-b':write(record+12,objects[2])
 elif action=='target-b':write(record+48,objects[3])
 elif action=='ray-b':write(record+36,objects[3])
 elif action=='ray-nil':write(record+36,0)
 elif action=='carry':write(record+72,f32(-.75))
 elif action=='level':write(record+8,0xffffffff)
 elif action=='glyph':write(record+20,1)
 elif action=='expiry':write(record+68,0x13579bdf)
 elif action=='record-position':writepoint(record+28,bits(99.25,-88.5))
 elif action=='caster-position':writepoint(objects[0]+56,bits(103.25,204.5))
 elif action=='target-position':writepoint(objects[1]+56,bits(111.25,222.5))
 elif action=='candidate-position':writepoint(objects[2]+56,bits(11.25,-22.5))
 elif action=='fps':write(fps,60)
 elif action=='frame':write(frame,0xfffffff9)
 elif action=='selected-b':write(selected,objects[3])
 elif action=='nearest':write(nearest,f32(4))
 elif action=='center':writepoint(center,bits(3,4))
 elif action=='caster-update-b':write(objects[0]+748,updates[2])
 elif action=='target-update-b':write(objects[1]+748,updates[2])
 elif action=='candidate-update-b':write(objects[2]+748,updates[3])
 elif action=='monster-update-b':write(objects[3]+748,updates[2])
 elif action=='caster-player-b':write(updates[0]+276,infos[2])
 elif action=='candidate-player-b':write(updates[2]+276,infos[3])
 elif action=='caster-class-zero':write(objects[0]+8,0)
 elif action=='caster-class-player':write(objects[0]+8,4)
 elif action=='target-class-player':write(objects[1]+8,4)
 elif action=='target-dead':write(objects[1]+16,get(objects[1]+16)|0x8020)
 elif action=='candidate-status-zero':write(updates[3]+1440,0)
 elif action=='candidate-flags-one':write(infos[2]+3680,1)
 elif action=='candidate-class-zero':write(objects[2]+8,0)
 elif action=='caster-mana-full':write16(updates[0]+4,100)
 elif action=='caster-mana-zero':write16(updates[0]+4,0)
 elif action=='target-mana-zero':write16(updates[2]+4,0)
 elif action=='obelisk-zero':write(updates[1],0)
 elif action=='obelisk-negative':write(updates[1],-5)
 elif action=='maximum-zero':write16(updates[0]+8,0)
 elif action=='quest-on':quest=1
 elif action=='quest-off':quest=0
 elif action=='class-wizard':write8(infos[0]+2251,1)
 elif action=='multiplier':write(multiplier_addresses[0],f32(1.9999999))
 elif action=='team-change':write8(objects[0]+52,1);write8(objects[1]+52,2)
 else:raise AssertionError(action)
def event(kind,args):events.append(dict(kind=kind,args=args,state=state()));mutate(kind)
def ret(v=0):
 esp=u.reg_read(UC_X86_REG_ESP);u.reg_write(UC_X86_REG_EAX,v&0xffffffff);u.reg_write(UC_X86_REG_EIP,get(esp));u.reg_write(UC_X86_REG_ESP,esp+4)
def floatret(v):u.mem_write(temp,bytes.fromhex(v));u.reg_write(UC_X86_REG_EIP,thunk)
def next_candidate():
 global circle
 if circle['next']>=len(current['candidates']):u.reg_write(UC_X86_REG_ESP,circle['esp']);circle=None;ret(0x13572468);return
 target=current['candidates'][circle['next']];circle['next']+=1;event('visit',[target,index(circle['caster'])]);esp=circle['esp']-64;put(esp,callback_end);put(esp+4,ref(target));put(esp+8,circle['caster']);u.reg_write(UC_X86_REG_ESP,esp);u.reg_write(UC_X86_REG_EIP,circle['callback'])
def hook(_u,pc,length,_):
 global circle
 if not active:return
 if pc==stop:u.emu_stop();return
 for name,(lo,hi) in dict(regions,**helpers).items():
  if lo<=pc<hi:
   traces[name][pc]=bytes(u.mem_read(pc,length))
   if pc==lo:entries[name]+=1
   return
 if thunk<=pc<thunk+7:return
 if pc==callback_end:next_candidate();return
 esp=u.reg_read(UC_X86_REG_ESP);arg=lambda n:get(esp+4*n);services[hex(pc)]+=1
 if pc==0x419d40:
  assert cstr(arg(1))=='ManaDrainRange';event('balance',['ManaDrainRange']);floatret(current['range'])
 elif pc==0x419d70:
  assert cstr(arg(1))=='ManaDrainCoeff';event('balance-level',['ManaDrainCoeff',arg(2)]);floatret(current['coefficient'])
 elif pc==0x517f90:
  assert arg(3)==regions['candidate'][0];event('circle',[index(arg(4)),*point(arg(1)),arg(2)]);assert circle is None;circle=dict(esp=esp,callback=arg(3),caster=arg(4),next=0);next_candidate()
 elif pc==0x5330c0:
  i,j=index(arg(1)),index(arg(2));value=current['enemy'][i][j];event('enemy',[i,j]);ret(value)
 elif pc==0x4ec520:
  i,j=index(arg(1)),index(arg(2));value=current['same_team'][i][j+1];event('same-team',[i,j]);ret(value)
 elif pc==0x535250:
  assert arg(2)==arg(3)==0 and arg(4)==5;value=current['trace'];event('trace',[*[get(arg(1)+4*i) for i in range(4)],arg(4)]);ret(value)
 elif pc==0x40a5c0:
  assert arg(1)==4096;value=quest;event('quest',[arg(1)]);ret(value)
 elif pc==0x4fea70:
  value=current['moved'];event('position',[index(arg(1)),*point(arg(2))]);ret(value)
 elif pc in [0x4eebf0,0x4eeb80]:
  owner,amount=arg(1),signed(arg(2));kind='mana-sub' if pc==0x4eebf0 else 'mana-add';event(kind,[index(owner),amount])
  if get(owner+8)&4:
   data=get(owner+748);value=get16(data+4);delta=-amount if kind=='mana-sub' else amount;write16(data+4,min(get16(data+8),max(0,value+delta)))
  ret(0x13572468)
 elif pc==0x4fef90:assert arg(1)==record;event('stop',[index(arg(2))]);ret(0x24681357)
 elif pc==0x4ff130:assert arg(1)==record;event('start',[]);ret(0x24681357)
 elif pc==0x501960:assert arg(3)==arg(4)==0;event('audio',[arg(1),index(arg(2))]);ret(0x24681357)
 else:raise AssertionError((current['name'],'undeclared instruction/service',hex(pc)))
def invalid(_u,access,addr,size,v,_):faults.append(dict(name=current['name'],pc=hex(u.reg_read(UC_X86_REG_EIP)),address=hex(addr)));return False
u.hook_add(UC_HOOK_CODE,hook);u.hook_add(UC_HOOK_MEM_INVALID,invalid)
seed=bytes((i*31+19)&255 for i in range(0x40000));registers={UC_X86_REG_EBX:0x13572468,UC_X86_REG_ESI:0x24681357,UC_X86_REG_EDI:0x98765432,UC_X86_REG_EBP:0xfedcba98}

def run(case):
 global active,current,events,calls,changed,allowed,circle,quest
 active,current,events,calls,changed,allowed,circle,quest=False,case,[],collections.Counter(),False,set(),None,case['quest'];u.mem_write(heap,seed)
 for off,v in [(8,case['level']),(12,ref(case['source'])),(16,ref(case['caster'])),(20,case['glyph']),(36,ref(case['ray'])),(48,ref(case['target'])),(68,case['expiry']),(72,case['carry'])]:put(record+off,v)
 for j,v in enumerate(case['position']):put(record+28+4*j,v)
 for i,o in enumerate(objects):
  for off,v in [(8,case['classes'][i]),(12,case['subclasses'][i]),(16,case['flags'][i]),(152,case['sync_flags'][i]),(340,case['buffs'][i]),(536,case['damage_frames'][i]),(556,health[i] if case['health'][i] else 0),(748,updates[case['updates'][i]])]:put(o+off,v)
  put8(o+52,case['team_ids'][i])
  for j,v in enumerate(case['positions'][i]):put(o+56+4*j,v)
 for i,data in enumerate(updates):
  put(data,case['obelisk_mana'][i]);put16(data+4,case['mana'][i]);put16(data+8,case['maximum'][i]);put(data+276,infos[case['players'][i]]);put(data+1440,case['monster_status'][i]);put8(infos[i]+2251,case['player_classes'][i]);put(infos[i]+3680,case['player_flags'][i])
 u.mem_write(selected-24,bytes(range(96)))
 for addr,v in [(selected,ref(case['selected'])),(nearest,case['nearest']),(fps,case['fps']),(frame,case['frame'])]:put(addr,v)
 for j,v in enumerate(case['center']):put(center+4*j,v)
 for addr,v in zip(multiplier_addresses,case['multipliers']):put(addr,v)
 globals_before={addr:bytes(u.mem_read(addr,size)) for addr,size in [(selected-24,96),(fps,4),(frame,4),*[(addr,4) for addr in multiplier_addresses]]}
 u.mem_write(cw,struct.pack('<H',a.control));u.emu_start(init,init+8,count=2);before=bytes(u.mem_read(heap,len(seed)));u.mem_write(sp-512,bytes(1024));put(sp,stop)
 args=[record] if case['kind']=='update' else [ref(case['caster']),ref(case['target']),case['amount']]
 for i,v in enumerate(args,1):put(sp+4*i,v)
 for reg,v in registers.items():u.reg_write(reg,v)
 u.reg_write(UC_X86_REG_EFLAGS,0x202);u.reg_write(UC_X86_REG_ESP,sp);active=True
 try:u.emu_start(regions[case['kind']][0],stop+1,count=100000)
 except Exception:print(json.dumps(dict(failed_case=case['name'],faults=faults)),flush=True);raise
 active=False
 assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4 and circle is None
 assert all(u.reg_read(reg)==v for reg,v in registers.items())
 assert bytes(u.mem_read(sp-512,256))==bytes(256) and bytes(u.mem_read(sp+4*(len(args)+1),256))==bytes(256)
 assert u.reg_read(UC_X86_REG_FPCW)==a.control and u.reg_read(UC_X86_REG_FPSW)&0x3800==0
 for off in [36,48,72]:allow(record+off,4)
 for data in updates:allow(data,4)
 for obj in objects:allow(obj+152,4)
 for addr in [selected,center,center+4,nearest]:allow(addr,4)
 after=bytes(u.mem_read(heap,len(seed)));assert all(x==y or heap+i in allowed for i,(x,y) in enumerate(zip(before,after))),(case['name'],'nonoutput heap canary')
 for addr,body in globals_before.items():assert all(x==y or addr+i in allowed for i,(x,y) in enumerate(zip(body,bytes(u.mem_read(addr,len(body)))))),(case['name'],'tracked global canary')
 if case['mutation']['stage']:assert changed,(case['name'],'mutation not reached')
 return dict(input=copy.deepcopy(case),result=dict(ret=u.reg_read(UC_X86_REG_EAX),state=state(),events=copy.deepcopy(events)))

template=dict(name='',kind='update',caster=0,source=2,glyph=0,target=1,ray=2,carry=f32(.25),expiry=0x87654321,level=3,position=bits(100,200),selected=3,center=bits(5,7),nearest=f32(12345),fps=30,frame=150,quest=0,classes=[4,0x400000,4,2],subclasses=[0,8,0,0],flags=[0]*4,team_ids=[0]*4,sync_flags=[0x11223344+i for i in range(4)],positions=[bits(100+20*i,200) for i in range(4)],buffs=[0]*4,health=[0]*4,damage_frames=[0]*4,updates=list(range(4)),mana=[10,10,30,10],maximum=[100]*4,obelisk_mana=[100,100,100,100],monster_status=[0x20]*4,players=list(range(4)),player_classes=[0,1,2,3],player_flags=[0]*4,multipliers=[f32(1.5),f32(1.75),f32(2.25)],range=f64(240),coefficient=f64(1.5),candidates=[1,2,3],enemy=[[1]*4 for _ in range(4)],same_team=[[0]*5 for _ in range(4)],trace=1,moved=0,amount=3,mutation=dict(stage='',at=0,action=''))
cases=[]
def add(kind,name,**changes):
 c=copy.deepcopy(template);c.update(kind=kind);c.update(changes);c['name']=kind+'/'+name;cases.append(c);return c
def vec(key,i,v):out=copy.deepcopy(template[key]);out[i]=v;return out
def mat(key,i,j,v):out=copy.deepcopy(template[key]);out[i][j]=v;return out
for kind in ['update','transfer']:add(kind,'default')
add('update','nil-caster',caster=-1)
for glyph in [1,2,0xffffffff]:
 for caster in [-1,0,2]:add('update','glyph-'+str([glyph,caster]),glyph=glyph,caster=caster)
for cls in [0,1,2,3,4,5,6,7,0x100,0x400000,0x400002,0x400004,0x80000004]:
 add('update','caster-class-'+hex(cls),classes=vec('classes',0,cls))
 add('transfer','caster-class-'+hex(cls),classes=vec('classes',0,cls))
for buff in [1<<8,0xffffffff,1<<7]:add('update','buff-'+str(buff),buffs=vec('buffs',0,buff))
for frame_value in [0,1,2,3,150,0x80000000,0xffffffff]:
 for ago in [0,1,2,0xffffffff]:add('update','recent-'+str([frame_value,ago]),frame=frame_value,health=vec('health',0,1),damage_frames=vec('damage_frames',0,(frame_value-ago)&0xffffffff))
for axis in [0,1]:
 for delta in [0.,-0.,4.999999523162842,5.,5.000000476837158,-4.999999523162842,-5.,-5.000000476837158,float('nan'),float('inf'),float('-inf')]:
  xy=[100.,200.];xy[axis]+=delta;add('update','motion-boundary-'+str([axis,delta]),position=bits(*xy))
for moved in [0,1,0xffffffff]:add('update','monster-motion-'+str(moved),classes=vec('classes',0,2),moved=moved)
for candidates in [[],[0],[1],[2],[3],[3,2,1],[2,1,3],[1,1]]:add('update','candidates-'+str(candidates),candidates=candidates)
for cls in [0,1,2,3,4,5,6,7,0x400000,0x400002,0x400004,0x80000004]:
 add('update','candidate-class-'+hex(cls),candidates=[2],classes=vec('classes',2,cls),subclasses=vec('subclasses',2,8))
for flag in [0x20,0x8000,0x8020,0x10000]:add('update','candidate-flags-'+hex(flag),candidates=[2],flags=vec('flags',2,flag))
for flags in [0,1,2,3,0xffffffff]:add('update','player-flags-'+str(flags),candidates=[2],player_flags=vec('player_flags',2,flags))
for result in [0,1,2,0xffffffff]:
 add('update','enemy-'+str(result),candidates=[2],enemy=mat('enemy',0,2,result))
 add('update','same-team-'+str(result),candidates=[2],same_team=mat('same_team',2,1,result))
 add('update','trace-'+str(result),trace=result)
for status in [0,0x20,0x40,0xffffffff]:add('update','monster-status-'+str(status),candidates=[3],monster_status=vec('monster_status',3,status))
for mana in [0,1,99,100,65535]:
 add('update','caster-mana-'+str(mana),mana=vec('mana',0,mana));add('update','target-mana-'+str(mana),candidates=[2],mana=vec('mana',2,mana))
for value in [-2147483648,-1,0,1,100,2147483647]:add('update','obelisk-mana-'+str(value),candidates=[1],obelisk_mana=vec('obelisk_mana',1,value))
for x,y in [(10100,200),(10099.9990234375,200),(10100.0009765625,200),(100,float('nan')),(float('inf'),200),(100+2**-20,200+2**-20),(120.00000762939453,200.00001525878906)]:add('update','candidate-point-'+str([x,y]),candidates=[1,2],positions=vec('positions',1,bits(x,y)))
values=[0.,-0.,.99,-.99,1.5,-1.5,2.75,-2.75,1.99999999,-1.99999999,2147483647.,2147483648.,-2147483649.,1e300,-1e300,float('inf'),float('-inf'),float('nan'),2**-150,-2**-150]
for value in values:
 add('update','range-'+str(value),range=f64(value));add('update','coefficient-'+str(value),coefficient=f64(value))
for value in [0.,.25,-.25,.75,-.75,1.5,-1.5,float('nan'),float('inf'),float('-inf')]:add('update','carry-'+str(value),carry=f32(value))
for ray in [-1,1,2,3]:add('update','ray-'+str(ray),ray=ray);add('update','no-target-ray-'+str(ray),candidates=[],ray=ray)
for rate in [2,3,4,30,60,0x80000000,0xffffffff]:add('update','fps-'+str(rate),fps=rate)
for level in [0,1,0x80000000,0xffffffff]:add('update','level-'+str(level),level=level)
for target in [1,2,3]:
 for quest_value in [0,1]:
  for amount in [-2147483648,-65537,-1,0,1,2,3,30,65535,65536,2147483647]:add('transfer','amount-'+str([target,quest_value,amount]),target=target,quest=quest_value,amount=amount)
for cls in [0,1,2,3,4,5,6,7,0x400000,0x400002,0x400004]:add('transfer','target-class-'+hex(cls),target=3,classes=vec('classes',3,cls),subclasses=vec('subclasses',3,8))
for status in [0,0x20]:
 for amount in [-1,0,1,3]:add('transfer','monster-'+str([status,amount]),target=3,monster_status=vec('monster_status',3,status),amount=amount)
for mana in [0,2,3,4,65535]:add('transfer','player-mana-'+str(mana),target=2,mana=vec('mana',2,mana))
for mana in [0,99,100,65535]:add('transfer','caster-mana-'+str(mana),mana=vec('mana',0,mana))
for source_team,target_team in [(0,0),(0,1),(1,1),(1,2),(2,1),(2,2)]:add('transfer','team-'+str([source_team,target_team]),team_ids=[source_team,target_team,0,0])
for player_class in [0,1,2,3,255]:
 for multiplier in [1.5,1.9999999,-.99,2**-149,float('inf'),float('nan')]:
  add('transfer','class-scale-'+str([player_class,multiplier]),quest=1,player_classes=vec('player_classes',0,player_class),multipliers=[f32(multiplier)]*3)
for stage,setup in {
 'balance':{},'balance-level':{},'circle':{},'visit':{},'enemy':dict(candidates=[2,3]),'same-team':dict(candidates=[2]),'trace':{},'quest':{},'mana-sub':dict(candidates=[2]),'mana-add':{},'stop':{},'start':{},'audio':{},'position':dict(classes=vec('classes',0,2))
}.items():
 for action in ['caster-b','source-b','target-b','ray-b','carry','level','glyph','expiry','record-position','caster-position','target-position','fps','frame','selected-b','nearest','center','caster-update-b','target-update-b','caster-player-b','caster-class-zero','target-class-player','target-dead','caster-mana-full','caster-mana-zero','obelisk-zero','maximum-zero','quest-on','class-wizard','multiplier']:
  add('update','mutate-'+stage+'-'+action,**setup,mutation=dict(stage=stage,at=0,action=action))
for stage in ['enemy','same-team','trace']:
 for action in ['candidate-update-b','candidate-player-b','candidate-position','candidate-flags-one','candidate-class-zero']:
  add('update','candidate-mutate-'+stage+'-'+action,candidates=[2],player_flags=vec('player_flags',3,1),mutation=dict(stage=stage,at=0,action=action))
for action in ['candidate-update-b','candidate-status-zero']:
 add('update','monster-mutate-'+action,candidates=[3],mutation=dict(stage='enemy',at=0,action=action))
for stage in ['balance','circle','visit','enemy','same-team','trace','mana-sub']:
 for action in ['caster-b','caster-nil','selected-b','center','nearest','target-update-b']:
  add('update','glyph-mutate-'+stage+'-'+action,glyph=1,candidates=[2],mutation=dict(stage=stage,at=0,action=action))
for stage,at,action in [('quest',1,'quest-on'),('quest',2,'quest-off'),('quest',2,'caster-update-b'),('audio',1,'target-b'),('audio',0,'caster-b'),('mana-add',0,'caster-nil')]:
 add('update','later-'+str([stage,at,action]),mutation=dict(stage=stage,at=at,action=action))
add('update','monster-cached-update',candidates=[3],monster_status=vec('monster_status',2,0),mutation=dict(stage='enemy',at=0,action='monster-update-b'))
add('update','position-live-player-class',classes=vec('classes',0,2),mutation=dict(stage='position',at=0,action='caster-class-player'))
add('update','enemy-team-change',candidates=[1,2],mutation=dict(stage='enemy',at=0,action='team-change'))
assert len({c['name'] for c in cases})==len(cases)
rows=[run(case) for case in cases]
sealed=[]
for name,(lo,hi) in dict(regions,**helpers).items():
 ins=traces[name];covered={pc+i for pc,body in ins.items() for i in range(len(body))};missing=sorted(set(range(lo,hi))-covered)
 if name in regions:assert not missing,(name,[hex(v) for v in missing])
 sealed.append(dict(name=name,address=hex(lo),end_exclusive=hex(hi),bytes=hi-lo,sha256=sha(bytes(u.mem_read(lo,hi-lo))),covered_bytes=len(covered),missing_bytes=[hex(v) for v in missing],instructions=[dict(pc=hex(pc),bytes=body.hex()) for pc,body in sorted(ins.items())]))
encoded=(json.dumps(dict(original_sha256=sha(blob),control_word=a.control,rows=rows),separators=(',',':'))+'\n').encode()
with a.fixture.open('xb') as f:f.write(encoded)
proof=dict(original_sha256=sha(blob),control_word=a.control,contexts=len(rows),entries=dict(entries),regions=sealed,all_root_candidate_converter_instruction_bytes_executed=True,zero_root_packing_arithmetic_conversion_replacements=True,register_stack_control_word_TOP_nonoutput_heap_and_global_canaries_preserved=True,services=dict(services),mutations=dict(mutations),faults=faults,fixture_sha256=sha(encoded),scope=__doc__)
with a.proof.open('x') as f:json.dump(proof,f,indent=2);f.write('\n')
print(json.dumps({k:proof[k] for k in ['original_sha256','control_word','contexts','entries','faults','fixture_sha256']}))
