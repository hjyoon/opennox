"""Execute unchanged Teleport Pop/ToMark roots and their FISTP converter.

Balance return, mode/buff reads, random, wake, FX/sound/audio, teleport, delayed
deletion and attribution are declared services. Canonical Mark IDs122..125
use the original marker array; existing OtherToMark117..120 native behavior
is a separately retained correction, not original-root equivalence.
No caller, argument packing, conversion, memory access or control flow is replaced.
"""
import argparse, collections, copy, hashlib, json, math, struct
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import *
from pathlib import Path

p = argparse.ArgumentParser(description=__doc__)
for key in ['game', 'fixture', 'proof']:
    p.add_argument('--' + key, type=Path, required=True)
p.add_argument('--control', type=lambda value: int(value, 0), choices=[0xe3f, 0x23f], default=0xe3f)
a = p.parse_args()
blob = a.game.read_bytes()
sha = lambda value: hashlib.sha256(value).hexdigest()
assert sha(blob) == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe = struct.unpack_from('<I', blob, 60)[0]
opt = pe + 24
base = struct.unpack_from('<I', blob, opt + 28)[0]
table = opt + struct.unpack_from('<H', blob, pe + 20)[0]
u = Uc(UC_ARCH_X86, UC_MODE_32)
u.mem_map(base, (struct.unpack_from('<I', blob, opt + 56)[0] + 4095) & ~4095)
for n in range(struct.unpack_from('<H', blob, pe + 6)[0]):
    _, va, size, off = struct.unpack_from('<IIII', blob, table + 40*n + 8)
    if size:
        u.mem_write(base + va, blob[off:off+size])
heap, stack, code = 0x21000000, 0x22000000, 0x23000000
for address in [heap, stack, code]: u.mem_map(address, 0x20000)
record = heap+0x1000
objects = [heap+0x2000+0x400*i for i in range(8)]
updates = [heap+0x6000+0x600*i for i in range(4)]
sp, stop, init, thunk, scratch, cw = stack+0x1e000, code+0x100, code+0x200, code+0x300, code+0x400, code+0x500
frameaddr = 0x84ea04
u.mem_write(init, b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw))
u.mem_write(thunk, b'\xdd\x05'+struct.pack('<I',scratch)+b'\xc3')
get=lambda address:struct.unpack('<I',u.mem_read(address,4))[0]
put=lambda address,value:u.mem_write(address,struct.pack('<I',value&0xffffffff))
words=lambda address:[get(address),get(address+4)]
f32=lambda v:struct.unpack('<I',struct.pack('<f',v))[0]
point=lambda x,y:[f32(x),f32(y)]
f64=lambda v:struct.pack('<d',v).hex()
ref=lambda n:0 if n<0 else objects[n]
index=lambda p:-1 if p==0 else objects.index(p)
update_index=lambda p:-1 if p==0 else updates.index(p)
cstr=lambda p:bytes(u.mem_read(p,128)).split(b'\0',1)[0].decode('ascii')
regions={'mark-create':(0x5305d0,0x530648),'mark-update':(0x530650,0x53081c),
         'pop-create':(0x530820,0x530874),'pop-update':(0x530880,0x530a21),'converter':(0x419a70,0x419a81)}
traces={name:{} for name in regions}
entries,services,mutations=collections.Counter(),collections.Counter(),collections.Counter()
active,current,events,calls,mutated,allowed,faults=False,{},[],collections.Counter(),False,set(),[]

def state():
 return dict(source=index(get(record+12)),caster=index(get(record+16)),target=index(get(record+48)),spell=get(record+4),level=get(record+8),flag=get(record+20),flags88=get(record+88),timer=get(record+68),frame=get(frameaddr),
  positions=[words(o+56) for o in objects],classes=[get(o+8) for o in objects],flags=[get(o+16) for o in objects],buffs=[get(o+340) for o in objects],netcodes=[get(o+36) for o in objects],field34=[get(o+136) for o in objects],
  object_updates=[update_index(get(o+748)) for o in objects],markers=[[index(get(d+116+4*j)) for j in range(4)] for d in updates],charges=[get(d+156) for d in updates])
def allow(p,size):allowed.update(range(p-heap,p-heap+size))
def write(p,v):allow(p,4);put(p,v)
def writepoint(p,v):
 for i,w in enumerate(v):write(p+4*i,w)
def owner_data():
 role=current['caster'] if current['kind'].startswith('pop') else current['source']
 return updates[current['object_updates'][role]]
def selected():
 if current['kind'].startswith('mark'):return current['spell']-122
 return current['random'][min(calls['random']-1,len(current['random'])-1)]
def mutate(kind):
 global mutated
 at=calls[kind];calls[kind]+=1;m=current['mutation']
 if mutated or m['stage']!=kind or m['at']!=at:return
 mutated=True;action=m['action'];mutations[kind+'/'+str(at)+'/'+action]+=1
 if action=='target-b':write(record+48,objects[3])
 elif action=='source-b':write(record+12,objects[3])
 elif action=='caster-b':write(record+16,objects[3])
 elif action=='caster-nil':write(record+16,0)
 elif action=='caster-destroyed':write(objects[1]+16,get(objects[1]+16)|0x20)
 elif action=='spell':write(record+4,43)
 elif action=='position-target':writepoint(objects[2]+56,point(-17.5,23.25))
 elif action=='position-marker':writepoint(objects[4]+56,point(-37.5,43.25))
 elif action=='invisible-target':write(objects[2]+340,get(objects[2]+340)|1)
 elif action=='visible-target':write(objects[2]+340,get(objects[2]+340)&~1)
 elif action=='netcode-target':write(objects[2]+36,0xdeadbeef)
 elif action=='marker-b':write(owner_data()+116+4*selected(),objects[7])
 elif action=='charge':write(owner_data()+156,0xfedcba01)
 elif action=='update-owner':
  role=current['caster'] if current['kind'].startswith('pop') else current['source'];write(objects[role]+748,updates[3])
 elif action=='frame':put(frameaddr,0xfffffffd)
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
  assert cstr(arg(1))=='TeleportDelay';event('delay',[arg(2)]);u.mem_write(scratch,bytes.fromhex(current['delay_bits64']));u.reg_write(UC_X86_REG_EIP,thunk)
 elif pc==0x4ff350:assert arg(2) in [0,14];ret(int(bool(get(arg(1)+340)&(1<<arg(2)))))
 elif pc==0x415fa0:
  assert arg(1)==0 and arg(2)==3 and cstr(arg(3))=='C:\\NoxPost\\src\\Server\\Magic\\Spell\\ExecDur.c' and arg(4)==2840
  value=current['random'][min(calls['random'],len(current['random'])-1)];event('random',[value]);ret(value)
 elif pc==0x530560:
  assert arg(2)==arg(1)+56 and arg(3)-56 in objects
  event('wake',[index(arg(1)),index(arg(3)-56),*words(arg(2)),*words(arg(3))]);ret(0x13572468)
 elif pc==0x522ff0:
  assert arg(1) in [137,129];event('fx',[arg(1),*words(arg(2))]);ret(0x24681357)
 elif pc==0x424800:
  mode=0 if current['kind'].startswith('pop') else 1;assert arg(2)==mode
  event('sound',[arg(1),mode]);ret(current['sound']+calls['sound']-1)
 elif pc==0x501960:event('audio',[arg(1),index(arg(2)),arg(3),arg(4)]);ret(0x13572468)
 elif pc==0x4e7190:
  assert arg(2)-56 in objects;target,destination=arg(1),words(arg(2));event('teleport',[index(target),index(arg(2)-56),*destination]);writepoint(target+56,destination);ret(1)
 elif pc==0x4e5cc0:event('delete',[index(arg(1))]);ret(0x24681357)
 elif pc==0x4e7540:event('attribution',[index(arg(1)),index(arg(2))]);ret(0x24681357)
 else:raise AssertionError((current['name'],'undeclared instruction/service',hex(pc)))
def invalid(_u,access,address,size,value,_):
 faults.append(dict(name=current['name'],pc=hex(u.reg_read(UC_X86_REG_EIP)),address=hex(address)));return False
u.hook_add(UC_HOOK_CODE,hook);u.hook_add(UC_HOOK_MEM_INVALID,invalid)
seed=bytes((i*31+19)&255 for i in range(0x20000))
registers={UC_X86_REG_EBX:0x13572468,UC_X86_REG_ESI:0x24681357,UC_X86_REG_EDI:0x98765432,UC_X86_REG_EBP:0xfedcba98}
def run(case):
 global active,current,events,calls,mutated,allowed
 active,current,events,calls,mutated,allowed=False,case,[],collections.Counter(),False,set();u.mem_write(heap,seed)
 for offset,value in [(4,case['spell']),(8,case['level']),(12,ref(case['source'])),(16,ref(case['caster'])),(20,case['flag']),(48,ref(case['target'])),(68,case['timer']),(88,case['flags88'])]:put(record+offset,value)
 for i,obj in enumerate(objects):
  for offset,value in [(8,case['classes'][i]),(16,case['flags'][i]),(36,case['netcodes'][i]),(136,case['field34'][i]),(340,case['buffs'][i]),(748,0 if case['object_updates'][i]<0 else updates[case['object_updates'][i]])]:put(obj+offset,value)
  for j,v in enumerate(case['positions'][i]):put(obj+56+4*j,v)
 for i,data in enumerate(updates):
  for j,marker in enumerate(case['markers'][i]):put(data+116+4*j,ref(marker))
  put(data+156,case['charges'][i])
 u.mem_write(frameaddr-16,bytes(range(36)));put(frameaddr,case['frame']);global_before=bytes(u.mem_read(frameaddr-16,36))
 u.mem_write(cw,struct.pack('<H',a.control));u.emu_start(init,init+8,count=2);before=bytes(u.mem_read(heap,len(seed)))
 u.mem_write(sp-512,bytes(1024));put(sp,stop);put(sp+4,record)
 for reg,value in registers.items():u.reg_write(reg,value)
 u.reg_write(UC_X86_REG_EFLAGS,0x202);u.reg_write(UC_X86_REG_ESP,sp);active=True;u.emu_start(regions[case['kind']][0],stop+1,count=30000);active=False
 assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4 and get(sp)==stop
 assert all(u.reg_read(reg)==value for reg,value in registers.items())
 assert bytes(u.mem_read(sp-512,256))==bytes(256) and bytes(u.mem_read(sp+8,256))==bytes(256)
 assert u.reg_read(UC_X86_REG_FPCW)==a.control and u.reg_read(UC_X86_REG_FPSW)&0x3800==0
 if case['kind'].endswith('create'):allow(record+68,4)
 if case['kind'].endswith('update'):
  for data in updates:allow(data+116,16);allow(data+156,4)
  if case['kind']=='mark-update':
   for obj in objects:allow(obj+136,4)
 after=bytes(u.mem_read(heap,len(seed)));assert all(x==y or i in allowed for i,(x,y) in enumerate(zip(before,after))),(case['name'],'nonoutput heap canary')
 after=bytes(u.mem_read(frameaddr-16,36));assert all(x==y or 16<=i<20 for i,(x,y) in enumerate(zip(global_before,after)))
 if case['mutation']['stage']:assert mutated,(case['name'],'mutation not reached')
 return dict(input=copy.deepcopy(case),result=dict(ret=u.reg_read(UC_X86_REG_EAX),state=state(),events=copy.deepcopy(events)))

template=dict(name='',kind='pop-update',coop=False,spell=121,level=3,timer=101,frame=100,source=0,caster=1,target=2,flag=1,flags88=0x76543201,
 classes=[4,4,4,2,2,2,2,2],flags=[0]*8,buffs=[0]*8,netcodes=[0x80000000+i for i in range(8)],field34=[0x12345670+i for i in range(8)],
 positions=[point(7,11),point(100.25,-100.5),point(110.25,-80.5),point(104,-90),point(300.75,-400.5),point(400.75,500.5),point(-600.25,700.5),point(800.25,-900.5)],
 object_updates=[0,1,2,3,-1,-1,-1,-1],markers=[[4,5,6,7] for _ in range(4)],charges=[0x02020202]*4,random=[0],delay_bits64=f64(2.5),sound=351,mutation=dict(stage='',at=0,action=''))
cases=[]
def add(family,stage,name,**changes):
 case=copy.deepcopy(template);case.update(kind=family+'-'+stage,spell=121 if family=='pop' else 122);case.update(changes)
 case['name']=case['kind']+'/'+name;cases.append(case)
def create(family,name,**changes):
 changes.update(timer=0x87654321);add(family,'create',name,**changes)
def marker_values(owner,index,value):
 charges=list(template['charges']);charges[owner]=(0xfedcba98&~(255<<(8*index)))|(value<<(8*index));return charges
for family in ['pop','mark']:
 create(family,'arena');create(family,'frame-wrap',frame=0xffffffff)
 for value in [.49,.51,1.5,2.5,3.5,-1.5,-2.5,-3.5,1.999999999999,2.999999999999,2147483647.,-2147483647.,math.inf,math.nan]:create(family,'delay-'+f64(value),coop=True,delay_bits64=f64(value))
 for level in [0,19,0xffffffff]:create(family,'level-'+str(level),coop=True,level=level)
 for action in ['frame','target-b','caster-b','source-b','spell']:create(family,'delay-'+action,coop=True,mutation=dict(stage='delay',at=0,action=action))
 owner=1 if family=='pop' else 0
 for index0 in range(4):
  for charge in [0,1,2,255]:
   options=dict(charges=marker_values(owner,index0,charge))
   if family=='pop':options['random']=[index0]
   else:options['spell']=122+index0
   add(family,'update','slot-'+str(index0)+'-charge-'+str(charge),**options)
 add(family,'update','wrong-frame',timer=103);add(family,'update','zero-timer',timer=0);add(family,'update','frame-wrap',timer=0,frame=0xffffffff)
 buffs=[0]*8;buffs[2]=1<<14;add(family,'update','anchor',buffs=buffs)
 add(family,'update','anchor-audio-target',buffs=buffs,mutation=dict(stage='audio',at=0,action='target-b'))
 for cls in [2,4]:
  classes=list(template['classes']);classes[2]=cls;buffs=[0]*8;buffs[2]=1
  add(family,'update','invisible-'+str(cls),classes=classes,buffs=buffs)
 classes=list(template['classes']);classes[owner]=2
 add(family,'update','nonplayer-owner',classes=classes)
 for objn in ([1,2] if family=='pop' else [0,2]):
  for flags0 in [0x20,0x8000]:
   flags=[0]*8;flags[objn]=flags0
   add(family,'update','flags-'+str(objn)+'-'+str(flags0),flags=flags)
 for role in (['caster','target'] if family=='pop' else ['source','target','caster']):add(family,'update','nil-'+role,**{role:-1})
 for stage,at,actions in [
  ('wake',0,['target-b','caster-b','caster-nil','caster-destroyed','source-b','spell','position-target','position-marker','marker-b','charge','update-owner','frame','invisible-target']),
  ('teleport',0,['target-b','caster-b','caster-nil','caster-destroyed','source-b','spell','position-target','position-marker','marker-b','charge','update-owner','frame','invisible-target']),
  ('fx',0,['target-b','caster-b','caster-nil','source-b','spell','position-target','position-marker','marker-b','charge','update-owner','frame']),
  ('fx',1,['target-b','caster-b','caster-nil','source-b','spell','position-target','position-marker','marker-b','charge','frame']),
  ('sound',0,['target-b','caster-b','caster-nil','caster-destroyed','source-b','spell','position-marker','marker-b','charge','frame']),
  ('audio',0,['target-b','caster-b','caster-nil','caster-destroyed','source-b','position-target','position-marker','marker-b','charge','update-owner','frame']),
  ('fx',2,['marker-b','caster-nil','target-b','charge','frame']),
  ('delete',0,['marker-b','target-b','caster-b','charge','frame'])]:
  for action in actions:
   options=dict(mutation=dict(stage=stage,at=at,action=action))
   if stage=='delete' or stage=='fx' and at==2:options['charges']=marker_values(owner,0,1)
   add(family,'update',stage+'-'+str(at)+'-'+action,**options)
for action in ['target-b','caster-b','source-b','spell','position-target','position-marker','marker-b','charge','update-owner','frame','invisible-target']:
 add('pop','update','random-'+action,mutation=dict(stage='random',at=0,action=action))
markers=copy.deepcopy(template['markers']);markers[1]=[-1,5,-1,7]
add('pop','update','random-retry',markers=markers,random=[0,2,0,3])
markers=copy.deepcopy(template['markers']);markers[1]=[-1]*4;add('pop','update','no-markers',markers=markers)
for index0 in range(4):
 markers=copy.deepcopy(template['markers']);markers[0][index0]=-1
 create('mark','missing-'+str(index0),spell=122+index0,markers=markers)
 for marker_flags in [0x20,0x8000]:
  flags=[0]*8;flags[4+index0]=marker_flags
  add('mark','update','marker-flags-'+str(index0)+'-'+str(marker_flags),spell=122+index0,flags=flags)
 updates0=list(template['object_updates']);updates0[0]=-1
 create('mark','nil-update-'+str(index0),spell=122+index0,object_updates=updates0)
create('mark','nonplayer-source',classes=[2]+template['classes'][1:])
add('mark','update','self-no-flag',target=0,flag=0)
add('mark','update','different-no-flag',flag=0)
for flag0 in [0x20,0x8000]:
 flags=[0]*8;flags[1]=flag0;add('mark','update','caster-flags-'+str(flag0),flags=flags)
for stage in ['sound','audio']:
 for action in ['target-b','caster-b','caster-nil','caster-destroyed','source-b','spell','marker-b','position-marker','charge','frame']:
  add('mark','update',stage+'-1-'+action,mutation=dict(stage=stage,at=1,action=action))
for action in ['target-b','netcode-target','spell','marker-b']:
 buffs=[0]*8;buffs[2]=1
 add('mark','update','invisible-sound-1-'+action,buffs=buffs,mutation=dict(stage='sound',at=1,action=action))
buffs=[0]*8;buffs[2]=1
add('mark','update','invisible-teleport-visible',buffs=buffs,mutation=dict(stage='teleport',at=0,action='visible-target'))
assert len({c['name'] for c in cases})==len(cases)
rows=[run(c) for c in cases]
sealed=[]
for name,(lo,hi) in regions.items():
 ins=traces[name];covered={pc+i for pc,body0 in ins.items() for i in range(len(body0))}
 assert covered==set(range(lo,hi)),(name,sorted(set(range(lo,hi))-covered))
 sealed.append(dict(name=name,address=hex(lo),end_exclusive=hex(hi),bytes=hi-lo,sha256=sha(bytes(u.mem_read(lo,hi-lo))),instructions=[dict(pc=hex(pc),bytes=body0.hex()) for pc,body0 in sorted(ins.items())]))
encoded=(json.dumps(dict(original_sha256=sha(blob),control_word=a.control,rows=rows),indent=2)+'\n').encode();a.fixture.write_bytes(encoded)
proof=dict(original_sha256=sha(blob),control_word=a.control,contexts=len(rows),entries=dict(entries),regions=sealed,all_original_instruction_bytes_executed=True,zero_caller_packing_arithmetic_or_conversion_replacements=True,callee_saved_registers_stack_control_word_TOP_and_nonoutput_heap_global_canaries_preserved=True,services=dict(services),mutations=dict(mutations),faults=faults,fixture_sha256=sha(encoded),scope='Whole original canonical Mark122..125 and Pop roots under declared services; Wake is declared. Existing OtherToMark117..120 is separately retained native correction, not original-root equivalence. No transitive or full-campaign equivalence claim.')
assert not faults;a.proof.write_text(json.dumps(proof,indent=2)+'\n')
print(json.dumps({k:v for k,v in proof.items() if k not in ['regions','services','mutations']},indent=2));print('mutations',len(mutations),sum(mutations.values()))
