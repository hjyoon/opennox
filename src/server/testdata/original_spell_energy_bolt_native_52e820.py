"""Execute unchanged EnergyBolt roots/candidate/converter and real small helpers.

Cancel, binary64 balance/table, circle enumeration, facing, surface distance,
enemy/LOS, ray/FX, virtual damage, state/audio/cast sound and mana debit are
declared services. Enumeration invokes the actual original candidate with a
fixed native-mapped caster context. Buff test, recent-damage and old-mana
helpers execute original code. Root packing/reloads, carry arithmetic and
x87 conversion are never replaced. Selected/mode/center/nearest scratch and
all mapped record/object/update state are observed. No transitive game claim.
"""
import argparse,collections,copy,hashlib,json,math,struct
from pathlib import Path
from unicorn import Uc,UC_ARCH_X86,UC_MODE_32,UC_HOOK_CODE,UC_HOOK_MEM_INVALID
from unicorn.x86_const import *
p=argparse.ArgumentParser(description=__doc__)
for k in ['game','fixture','proof']:p.add_argument('--'+k,type=Path,required=True)
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
record=heap+0x1000;objects=[heap+0x2000+0x400*i for i in range(8)];updates=[heap+0x10000+0x600*i for i in range(8)]
health=[heap+0x18000+0x20*i for i in range(8)]
sp,stop,init,thunk,temp,cw,callback_end,damage_entry=stack+0x3e000,code+0x100,code+0x200,code+0x300,code+0x400,code+0x500,code+0x600,code+0x700
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw));u.mem_write(thunk,b'\xdd\x05'+struct.pack('<I',temp)+b'\xc3')
fps,frame,mode,center,selected,nearest=0x85b3fc,0x84ea04,0x833bac,0x833bd0,0x833bdc,0x833be0
get=lambda addr:struct.unpack('<I',u.mem_read(addr,4))[0]
get16=lambda addr:struct.unpack('<H',u.mem_read(addr,2))[0]
put=lambda addr,v:u.mem_write(addr,struct.pack('<I',v&0xffffffff))
put16=lambda addr,v:u.mem_write(addr,struct.pack('<H',v&0xffff))
words=lambda addr:[get(addr),get(addr+4)]
signed=lambda v,b:((v&((1<<b)-1))^(1<<(b-1)))-(1<<(b-1))
f32=lambda v:struct.unpack('<I',struct.pack('<f',v))[0]
point=lambda x,y:[f32(x),f32(y)]
f64=lambda v:struct.pack('<d',v).hex()
ref=lambda i:0 if i<0 else objects[i]
index=lambda v:-1 if v==0 else objects.index(v)
upindex=lambda v:updates.index(v)
cstr=lambda addr:bytes(u.mem_read(addr,128)).split(b'\0',1)[0].decode('ascii')
regions={'create':(0x52e820,0x52e84c),'update':(0x52e850,0x52ec58),'candidate':(0x52ec60,0x52ed22),'converter':(0x419a70,0x419a81)}
helpers={'buff':(0x4ff350,0x4ff373),'recent':(0x4e6bd0,0x4e6bfa),'mana':(0x4eec80,0x4eeca7)}
traces={n:{} for n in dict(regions,**helpers)};entries,services,mutations=collections.Counter(),collections.Counter(),collections.Counter()
active,current,events,calls,changed,allowed,faults,circle=False,{},[],collections.Counter(),False,set(),[],None

def state():
 return dict(caster=index(get(record+16)),source=index(get(record+12)),flag20=get(record+20),target=index(get(record+48)),ray=index(get(record+36)),carry=get(record+72),level=get(record+8),expiry=get(record+68),start=get(record+60),position=words(record+28),fps=get(fps),frame=get(frame),mode=get(mode),center=words(center),selected=index(get(selected)),nearest=get(nearest),classes=[get(o+8) for o in objects],subclasses=[get(o+12) for o in objects],flags=[get(o+16) for o in objects],positions=[words(o+56) for o in objects],directions=[get16(o+124) for o in objects],buffs=[get(o+340) for o in objects],health=[int(get(o+556)!=0) for o in objects],damage_frames=[get(o+536) for o in objects],updates=[upindex(get(o+748)) for o in objects],cursors=[index(get(d+288)) for d in updates],mana=[get16(d+4) for d in updates])
def allow(addr,size):allowed.update(range(addr,addr+size))
def write(addr,v):allow(addr,4);put(addr,v)
def write16(addr,v):allow(addr,2);put16(addr,v)
def writepoint(addr,p):
 for i,v in enumerate(p):write(addr+4*i,v)
def mutate(kind):
 global changed
 at=calls[kind];calls[kind]+=1;m=current['mutation']
 if changed or m['stage']!=kind or m['at']!=at:return
 changed=True;action=m['action'];mutations[kind+'/'+str(at)+'/'+action]+=1
 if action=='caster-b':write(record+16,objects[1])
 elif action=='caster-nil':write(record+16,0)
 elif action=='source-b':write(record+12,objects[1])
 elif action=='target-b':write(record+48,objects[3])
 elif action=='target-nil':write(record+48,0)
 elif action=='ray-b':write(record+36,objects[4])
 elif action=='ray-nil':write(record+36,0)
 elif action=='carry':write(record+72,f32(-0.75))
 elif action=='level':write(record+8,0xffffffff)
 elif action=='flag20':write(record+20,1)
 elif action=='expiry':write(record+68,0x13579bdf)
 elif action=='start':write(record+60,0xfffffff0)
 elif action=='record-position':writepoint(record+28,point(99.25,-88.5))
 elif action=='fps':write(fps,60)
 elif action=='frame':write(frame,0xfffffff9)
 elif action=='selected-b':write(selected,objects[5])
 elif action=='nearest':write(nearest,f32(4))
 elif action=='mode':write(mode,1)
 elif action=='center':writepoint(center,point(3,4))
 elif action=='caster-update-b':write(objects[0]+748,updates[1])
 elif action=='cursor-b':write(updates[0]+288,objects[3])
 elif action=='caster-position':writepoint(objects[0]+56,point(3,4))
 elif action=='target-position':writepoint(objects[2]+56,point(1.25,-2.5))
 elif action=='target-dead':write(objects[2]+16,get(objects[2]+16)|0x8020)
 elif action=='caster-class-zero':write(objects[0]+8,0)
 elif action=='caster-class-monster':write(objects[0]+8,2)
 elif action=='mana-zero':write16(updates[0]+4,0)
 elif action=='mana-one':write16(updates[0]+4,1)
 else:raise AssertionError(action)
def event(kind,args):events.append(dict(kind=kind,args=args,state=state()));mutate(kind)
def ret(v=0):
 esp=u.reg_read(UC_X86_REG_ESP);u.reg_write(UC_X86_REG_EAX,v&0xffffffff);u.reg_write(UC_X86_REG_EIP,get(esp));u.reg_write(UC_X86_REG_ESP,esp+4)
def floatret(bits):u.mem_write(temp,bytes.fromhex(bits));u.reg_write(UC_X86_REG_EIP,thunk)
def next_candidate():
 global circle
 if circle['next']>=len(current['candidates']):
  u.reg_write(UC_X86_REG_ESP,circle['esp']);circle=None;ret(0x13572468);return
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
 if pc==0x4feb10:assert arg(1)==43;event('cancel',[arg(1),index(arg(2))]);ret(0x13572468)
 elif pc==0x4fea70:assert arg(2)==record+28;event('position',[index(arg(1)),*words(arg(2))]);ret(current['position_result'])
 elif pc==0x4e6e50:
  i,j=index(arg(1)-56),index(arg(3)-56);event('facing',[i,signed(arg(2),16),j,*words(arg(1)),*words(arg(3))]);ret(current['facing'][i][j])
 elif pc==0x4e6c00:
  i,j=index(arg(1)),index(arg(2));event('distance',[i,j]);floatret(current['distance_bits64'][i][j])
 elif pc in [0x5330c0,0x5370e0]:
  i,j=index(arg(1)),index(arg(2));kind='enemy' if pc==0x5330c0 else 'los'
  if kind=='los':assert arg(3)==0
  event(kind,[i,j]);ret(current[kind][i][j])
 elif pc==0x517f90:
  assert arg(3)==regions['candidate'][0]
  event('circle',[index(arg(4)),*words(arg(1)),arg(2)])
  assert circle is None;circle=dict(esp=esp,callback=arg(3),caster=arg(4),next=0);next_candidate()
 elif pc==0x4fef90:assert arg(1)==record;event('stop',[index(arg(2))]);ret(0x24681357)
 elif pc==0x4ff130:assert arg(1)==record;event('start',[]);ret(0x24681357)
 elif pc==0x419d40:
  key=cstr(arg(1));assert key in current['balances'];event('balance',[key]);floatret(current['balances'][key])
 elif pc==0x419d70:
  key=cstr(arg(1));assert key=='EnergyBoltDamage';event('balance-level',[key,arg(2)]);floatret(current['damage_bits64'])
 elif pc==damage_entry:assert arg(3)==0 and arg(5)==17;event('damage',[index(arg(1)),index(arg(2)),signed(arg(4),32)]);ret(0x13572468)
 elif pc==0x522ff0:
  assert arg(1)==130;owner=-1 if arg(2)==record+28 else index(arg(2)-56);event('fx',[arg(1),owner,*words(arg(2))]);ret(0x13572468)
 elif pc==0x4fa020:assert arg(2)==10;event('state',[index(arg(1)),arg(2)]);ret(1)
 elif pc==0x501960:assert arg(3)==arg(4)==0;event('audio',[arg(1),index(arg(2))]);ret(0x24681357)
 elif pc==0x424800:assert arg(1)==24 and arg(2)==0;event('cast-sound',[]);ret(current['cast_sound'])
 elif pc==0x4eebf0:
  owner=arg(1);assert arg(2)==1;event('mana-sub',[index(owner),1])
  if get(owner+8)&4:addr=get(owner+748)+4;write16(addr,max(0,get16(addr)-1))
  ret(0x13572468)
 else:raise AssertionError((current['name'],'undeclared instruction/service',hex(pc)))
def invalid(_u,access,addr,size,v,_):faults.append(dict(name=current['name'],pc=hex(u.reg_read(UC_X86_REG_EIP)),address=hex(addr)));return False
u.hook_add(UC_HOOK_CODE,hook);u.hook_add(UC_HOOK_MEM_INVALID,invalid)
seed=bytes((i*31+19)&255 for i in range(0x40000));registers={UC_X86_REG_EBX:0x13572468,UC_X86_REG_ESI:0x24681357,UC_X86_REG_EDI:0x98765432,UC_X86_REG_EBP:0xfedcba98}
def run(case):
 global active,current,events,calls,changed,allowed,circle
 active,current,events,calls,changed,allowed,circle=False,case,[],collections.Counter(),False,set(),None;u.mem_write(heap,seed)
 for off,v in [(16,ref(case['caster'])),(12,ref(case['source'])),(20,case['flag20']),(48,ref(case['target'])),(36,ref(case['ray'])),(72,case['carry']),(8,case['level']),(68,case['expiry']),(60,case['start'])]:put(record+off,v)
 for j,v in enumerate(case['position']):put(record+28+4*j,v)
 for i,o in enumerate(objects):
  for off,v in [(8,case['classes'][i]),(12,case['subclasses'][i]),(16,case['flags'][i]),(340,case['buffs'][i]),(556,health[i] if case['health'][i] else 0),(536,case['damage_frames'][i]),(748,updates[case['updates'][i]]),(716,damage_entry)]:put(o+off,v)
  put16(o+124,case['directions'][i])
  for j,v in enumerate(case['positions'][i]):put(o+56+4*j,v)
 for i,d in enumerate(updates):put(d+288,ref(case['cursors'][i]));put16(d+4,case['mana'][i])
 u.mem_write(mode-16,bytes(range(80)))
 for addr,v in [(fps,case['fps']),(frame,case['frame']),(mode,case['mode']),(selected,ref(case['selected'])),(nearest,case['nearest'])]:put(addr,v)
 for j,v in enumerate(case['center']):put(center+4*j,v)
 globals_before={addr:bytes(u.mem_read(addr,size)) for addr,size in [(mode-16,80),(fps,4),(frame,4)]}
 u.mem_write(cw,struct.pack('<H',a.control));u.emu_start(init,init+8,count=2);before=bytes(u.mem_read(heap,len(seed)))
 u.mem_write(sp-512,bytes(1024));put(sp,stop);put(sp+4,record)
 for reg,v in registers.items():u.reg_write(reg,v)
 u.reg_write(UC_X86_REG_EFLAGS,0x202);u.reg_write(UC_X86_REG_ESP,sp);active=True
 try:u.emu_start(regions[case['kind']][0],stop+1,count=100000)
 except Exception:
  print(json.dumps(dict(failed_case=case['name'],faults=faults)),flush=True);raise
 active=False
 assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4 and get(sp)==stop and circle is None
 assert all(u.reg_read(reg)==v for reg,v in registers.items())
 assert bytes(u.mem_read(sp-512,256))==bytes(256) and bytes(u.mem_read(sp+8,256))==bytes(256)
 assert u.reg_read(UC_X86_REG_FPCW)==a.control and u.reg_read(UC_X86_REG_FPSW)&0x3800==0
 if case['kind']=='update':
  for off in [48,36,72,68]:allow(record+off,4)
  for addr in [mode,selected,nearest,center,center+4]:allow(addr,4)
 after=bytes(u.mem_read(heap,len(seed)));assert all(x==y or heap+i in allowed for i,(x,y) in enumerate(zip(before,after))),(case['name'],'nonoutput heap canary')
 for addr,body in globals_before.items():assert all(x==y or addr+i in allowed for i,(x,y) in enumerate(zip(body,bytes(u.mem_read(addr,len(body)))))),(case['name'],'tracked global canary')
 if case['mutation']['stage']:assert changed,(case['name'],'mutation not reached')
 return dict(input=copy.deepcopy(case),result=dict(ret=u.reg_read(UC_X86_REG_EAX),state=state(),events=copy.deepcopy(events)))

template=dict(name='',kind='update',caster=0,source=1,flag20=0,target=2,ray=2,carry=f32(0.25),level=3,expiry=0x87654321,start=0xfffffff0,position=point(1.725,-22.25),fps=30,frame=0xfffffff0,mode=7,center=point(3.25,-4.5),selected=5,nearest=f32(12345),classes=[4,4,2,2,2,2,2,2],subclasses=[0]*8,flags=[0]*8,positions=[point(0,0),point(5,-7),point(100,0),point(20,30),point(-50,5),point(3,4),point(0,0),point(0,0)],directions=[0x8001,3,5,7,9,11,13,15],buffs=[0]*8,health=[0]*8,damage_frames=[100]*8,updates=list(range(8)),cursors=[2,3,-1,-1,-1,-1,-1,-1],mana=[8,9,10,11,12,13,14,15],position_result=0,facing=[[1]*8 for _ in range(8)],enemy=[[1]*8 for _ in range(8)],los=[[1]*8 for _ in range(8)],distance_bits64=[[f64(100)]*8 for _ in range(8)],balances={k:f64(v) for k,v in [('LightningRange',400),('EnergyBoltGlyphDamage',3.5),('LightningSearchTime',5.75)]},damage_bits64=f64(1.5),cast_sound=77,candidates=[2,3,4],mutation=dict(stage='',at=0,action=''))
cases=[]
def add(kind,name,**changes):c=copy.deepcopy(template);c.update(kind=kind);c.update(changes);c['name']=kind+'/'+name;cases.append(c);return c
def vec(key,i,v):out=copy.deepcopy(template[key]);out[i]=v;return out
def mat(key,i,j,v):out=copy.deepcopy(template[key]);out[i][j]=v;return out
for kind in ['create','update']:add(kind,'default');add(kind,'nil-caster',caster=-1)
for caster in [-1,0,1]:
 for flag in [1,2,0xffffffff]:add('update','glyph-'+str(caster)+'-'+str(flag),caster=caster,flag20=flag)
for cls in [0,1,2,3,4,5,6,7,0x100,0x80000004]:add('update','class-'+hex(cls),classes=vec('classes',0,cls))
for buff in [1<<8,0xffffffff,1<<7]:add('update','buff-'+str(buff),buffs=vec('buffs',0,buff))
for flag in [0,1]:
 for caster in [0,1,-1]:
  if caster==-1 and flag==0:continue
  add('update','glyph-no-target-'+str(caster)+'-'+str(flag),caster=caster,flag20=flag,candidates=[],target=-1,cursors=[-1]*8)
for target in [-1,3,4]:add('update','target-'+str(target),target=target)
for flags in [0x20,0x8000,0x8020,0x10000]:add('update','target-flags-'+hex(flags),flags=vec('flags',2,flags))
for facing in [0,1,2,3,0x100,0xffffffff]:add('update','retained-facing-'+str(facing),facing=mat('facing',0,2,facing))
for v in [399.999999999,400,400.000000001,float('nan'),float('inf'),float('-inf')]:
 for cursor in [False,True]:add('update',('cursor' if cursor else 'retained')+'-distance-'+str(v),target=-1 if cursor else 2,distance_bits64=mat('distance_bits64',0 if cursor else 2,2 if cursor else 0,f64(v)))
add('update','retained-no-los',los=mat('los',0,2,0));add('update','cursor-not-enemy',target=-1,enemy=mat('enemy',0,2,0));add('update','cursor-no-los',target=-1,los=mat('los',0,2,0))
for value in [0,1,0xffffffff]:add('update','position-'+str(value),classes=vec('classes',0,2),position_result=value)
for now in [0,1,2,3,0x7fffffff,0x80000000,0xffffffff]:
 for since in [0,1,2,3,0xffffffff]:
  add('update','recent-'+str(now)+'-'+str(since),frame=now,start=(now-3)&0xffffffff,health=vec('health',0,1),damage_frames=vec('damage_frames',0,(now-since)&0xffffffff))
add('update','recent-without-health',frame=100,start=90,damage_frames=vec('damage_frames',0,100))
for candidates in [[],[0],[2],[3,2],[2,3],[5,4,3,2]]:add('update','candidates-'+str(candidates),target=-1,cursors=[-1]*8,candidates=candidates)
for cls in [0,1,2,4,6,0x20000,0x20002,0x80000000]:add('update','candidate-class-'+hex(cls),target=-1,cursors=[-1]*8,candidates=[2],classes=vec('classes',2,cls))
for flags in [0x20,0x8000,0x8020]:add('update','candidate-flags-'+hex(flags),target=-1,cursors=[-1]*8,candidates=[2],flags=vec('flags',2,flags))
add('update','candidate-excluded',target=-1,cursors=[-1]*8,candidates=[2],subclasses=vec('subclasses',2,0x8000))
for glyph in [0,1]:
 for enemy,los,facing in [(0,1,1),(1,0,1),(1,1,0),(1,1,2)]:add('update','candidate-gates-'+str(glyph)+str([enemy,los,facing]),flag20=glyph,target=-1,cursors=[-1]*8,candidates=[2],enemy=mat('enemy',0,2,enemy),los=mat('los',0,2,los),facing=mat('facing',0,2,facing))
for name,x,y in [('boundary',400,0),('inside',399.9999694824219,0),('outside',400.0000305175781,0),('tie',20,30),('nan',float('nan'),0),('inf',float('inf'),0),('tiny',2**-149,2**-149),('round',16.000001907348633,0.0001)]:add('update','candidate-'+name,target=-1,cursors=[-1]*8,candidates=[2,3],positions=vec('positions',2,point(x,y)))
values=[0,-0.0,0.99,-0.99,1.5,-1.5,2.75,-2.75,1.99999999,-1.99999999,2147483647.0,2147483648.0,-2147483649.0,1e300,-1e300,float('inf'),float('-inf'),float('nan'),2**-150,-2**-150]
for key in template['balances']:
 for value in values:
  b=dict(template['balances']);b[key]=f64(value);add('update','balance-'+key+'-'+str(value),flag20=1 if key=='EnergyBoltGlyphDamage' else 0,balances=b)
for value in values:add('update','damage-table-'+str(value),damage_bits64=f64(value))
for carry in [0,0.25,-0.25,0.75,-0.75,1.5,-1.5,float('nan'),float('inf')]:add('update','carry-'+str(carry),carry=f32(carry))
for mana in [0,1,2,65535]:add('update','mana-'+str(mana),mana=vec('mana',0,mana))
for ray in [-1,2,3]:
 add('update','ray-'+str(ray),ray=ray);add('update','no-target-ray-'+str(ray),target=-1,cursors=[-1]*8,candidates=[],ray=ray)
for rate in [3,4,30,60,0x80000000,0xffffffff]:add('update','fps-'+str(rate),fps=rate)
for level in [0,1,0x80000000,0xffffffff]:add('update','level-'+str(level),level=level)
for stage in ['cancel','fx']:
 for action in ['caster-b','caster-nil','source-b','target-b','carry','level','flag20','record-position','fps','frame']:add('create','mutate-'+stage+'-'+action,mutation=dict(stage=stage,at=0,action=action))
setups={'balance':{},'position':dict(classes=vec('classes',0,2)),'facing':{},'distance':{},'los':{},'enemy':dict(target=-1),'circle':dict(target=-1,cursors=[-1]*8),'visit':dict(target=-1,cursors=[-1]*8),'balance-level':{},'stop':dict(ray=3),'start':dict(ray=-1),'damage':{},'state':{},'audio':{},'mana-sub':{}}
actions=['caster-b','source-b','target-b','ray-b','carry','level','flag20','expiry','start','record-position','fps','frame','caster-update-b','cursor-b','caster-position','target-position','target-dead','mana-zero','mana-one']
for stage,setup in setups.items():
 for action in actions:add('update','mutate-'+stage+'-'+action,**setup,mutation=dict(stage=stage,at=0,action=action))
for stage in ['circle','visit','enemy','facing','los']:
 for action in ['selected-b','nearest','mode','center']:add('update','scratch-'+stage+'-'+action,target=-1,cursors=[-1]*8,mutation=dict(stage=stage,at=0,action=action))
for stage in ['balance','circle','visit','damage','cast-sound','audio','fx']:
 for action in ['caster-b','source-b','selected-b','nearest','mode','center','target-position']:
  add('update','glyph-mutate-'+stage+'-'+action,flag20=1,mutation=dict(stage=stage,at=0,action=action))
for stage,at,action in [('state',1,'caster-b'),('state',1,'caster-update-b'),('mana-sub',0,'caster-nil'),('mana-sub',0,'caster-class-zero'),('mana-sub',0,'caster-class-monster'),('audio',0,'target-b'),('balance',1,'frame')]:add('update','later-'+stage+'-'+str(at)+'-'+action,mutation=dict(stage=stage,at=at,action=action))
assert len({c['name'] for c in cases})==len(cases)
rows=[run(case) for case in cases]
sealed=[]
for name,(lo,hi) in dict(regions,**helpers).items():
 ins=traces[name];covered={pc+i for pc,body in ins.items() for i in range(len(body))};missing=sorted(set(range(lo,hi))-covered)
 if name in regions:assert not missing,(name,[hex(v) for v in missing])
 sealed.append(dict(name=name,address=hex(lo),end_exclusive=hex(hi),bytes=hi-lo,sha256=sha(bytes(u.mem_read(lo,hi-lo))),covered_bytes=len(covered),missing_bytes=[hex(v) for v in missing],instructions=[dict(pc=hex(pc),bytes=body.hex()) for pc,body in sorted(ins.items())]))
encoded=(json.dumps(dict(original_sha256=sha(blob),control_word=a.control,rows=rows),separators=(',',':'))+'\n').encode()
with a.fixture.open('xb') as f:f.write(encoded)
proof=dict(original_sha256=sha(blob),control_word=a.control,contexts=len(rows),entries=dict(entries),regions=sealed,all_root_candidate_converter_instruction_bytes_executed=True,helpers_executed_with_measured_byte_coverage=True,zero_root_packing_arithmetic_conversion_replacements=True,callee_saved_registers_stack_control_word_TOP_and_nonoutput_heap_and_global_canaries_preserved=True,services=dict(services),mutations=dict(mutations),faults=faults,fixture_sha256=sha(encoded),scope=__doc__)
with a.proof.open('x') as f:json.dump(proof,f,indent=2);f.write('\n')
print(json.dumps({k:v for k,v in proof.items() if k not in ['regions','mutations','scope']}))
