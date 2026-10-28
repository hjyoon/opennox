"""Execute complete unchanged Tag create/update/destroy and FISTP helper.

Balance, minimap mark/unmark, unit-code lookup and packet send are declared
services. A scratch FLD/RET thunk supplies the balance callee's double result;
all original spill/conversion instructions execute unchanged. Destroy's
incidental return register is not a duration-callback result contract.
"""
import argparse, collections, copy, hashlib, json, math, struct
from pathlib import Path
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import *
p=argparse.ArgumentParser(description=__doc__)
for key in ('game','fixture','proof'):p.add_argument('--'+key,type=Path,required=True)
p.add_argument('--control',type=lambda s:int(s,0),default=0xe3f,choices=[0xe3f,0x23f])
a=p.parse_args();blob=a.game.read_bytes();sha=lambda b:hashlib.sha256(b).hexdigest()
assert sha(blob)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,60)[0];opt=pe+24;base=struct.unpack_from('<I',blob,opt+28)[0];table=opt+struct.unpack_from('<H',blob,pe+20)[0]
u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,(struct.unpack_from('<I',blob,opt+56)[0]+4095)&~4095)
for n in range(struct.unpack_from('<H',blob,pe+6)[0]):
 _,va,size,offset=struct.unpack_from('<IIII',blob,table+40*n+8)
 if size:u.mem_write(base+va,blob[offset:offset+size])
heap,stack,code=0x21000000,0x22000000,0x23000000
for addr in (heap,stack,code):u.mem_map(addr,0x10000)
sp,record,frameaddr=stack+0xe000,heap+0x1000,0x84ea04
objects=[heap+0x2000+i*1024 for i in range(4)];data=[heap+0x4000,heap+0x4800];infos=[heap+0x6000,heap+0x7000]
stop,init,thunk,scratch,cw=code+0x100,code+0x200,code+0x300,code+0x400,code+0x500
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw))
u.mem_write(thunk,b'\xdd\x05'+struct.pack('<I',scratch)+b'\xc3')
get=lambda p:struct.unpack('<I',u.mem_read(p,4))[0]
put=lambda p,v:u.mem_write(p,struct.pack('<I',v&0xffffffff))
ref=lambda values,i:0 if i<0 else values[i]
idx=lambda values,p:-1 if not p else values.index(p)
u16=lambda p:struct.unpack('<H',u.mem_read(p,2))[0]
u8=lambda p:u.mem_read(p,1)[0]
f64=lambda value:struct.pack('<d',value).hex()
regions={'create':(0x530160,0x530241),'update':(0x530250,0x530270),'destroy':(0x530270,0x53030a),'converter':(0x419a70,0x419a81)}
traces={name:{} for name in regions};entries=collections.Counter();mutations=collections.Counter();faults=[]
current={};events=[];changed=False;active=False

def state():
 return dict(caster=idx(objects,get(record+16)),target=idx(objects,get(record+48)),level=get(record+8),timer=get(record+68),frame=get(frameaddr),
  data=[idx(data,get(obj+748)) for obj in objects[:2]],players=[idx(infos,get(d+276)) for d in data],indexes=[u8(i+2064) for i in infos],
  classes=[get(obj+8) for obj in objects],flags=[get(obj+16) for obj in objects],types=[u16(obj+4) for obj in objects[2:]])

def event(kind,args):events.append(dict(kind=kind,args=args,state=state()))
def mutate(service):
 global changed
 m=current['mutation']
 if changed or not m.startswith(service+'-'):return
 changed=True;mutations[m]+=1
 if m=='balance-caster-b' or m=='mark-caster-b':put(record+16,objects[1])
 elif m in ('balance-data-b','mark-data-b','code-data-b','unmark-data-b'):put(objects[0]+748,data[1])
 elif m in ('balance-info-b','mark-info-b','unmark-info-b'):put(data[0]+276,infos[1])
 elif m in ('balance-target-b','mark-target-b','code-target-b','unmark-target-b'):put(record+48,objects[3])
 elif m=='balance-level':put(record+8,0xffffffff)
 elif m=='balance-frame':put(frameaddr,0xfffffff9)
 elif m in ('mark-index','unmark-index'):u.mem_write(infos[0]+2064,b'\xff')
 elif m=='code-type-b':u.mem_write(objects[2]+4,struct.pack('<H',0xfe12))
 elif m=='send-timer':put(record+68,0x12345678)
 else:raise AssertionError(m)

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
 esp=u.reg_read(UC_X86_REG_ESP);arg=lambda n:get(esp+4*n)
 if pc==0x419d40:
  assert bytes(u.mem_read(arg(1),32)).split(b'\0',1)[0]==b'TagDurationPerLevel'
  event('balance',[]);mutate('balance');u.mem_write(scratch,bytes.fromhex(current['balance_bits64']));u.reg_write(UC_X86_REG_EIP,thunk)
 elif pc in (0x417190,0x417300):
  kind='mark' if pc==0x417190 else 'unmark';assert arg(3)==1
  event(kind,[arg(1),idx(objects,arg(2)),arg(3)]);mutate(kind);ret(0x12344321)
 elif pc==0x578ac0:
  target=idx(objects,arg(1));event('code',[target]);mutate('code');ret(current['codes'][target-2])
 elif pc==0x4e5420:
  assert [arg(3),arg(4),arg(5)]==[7,0,1]
  event('send',[arg(1),bytes(u.mem_read(arg(2),7)).hex(),arg(4),arg(5)]);mutate('send');ret(0x3456789a)
 else:raise AssertionError((current['name'],'undeclared instruction/service',hex(pc)))
def invalid(_u,access,address,size,value,_):
 faults.append(dict(name=current['name'],pc=hex(u.reg_read(UC_X86_REG_EIP)),address=hex(address)));return False
u.hook_add(UC_HOOK_CODE,hook);u.hook_add(UC_HOOK_MEM_INVALID,invalid)
seed=bytes((i*31+19)&255 for i in range(0x10000));registers={UC_X86_REG_EBX:0x13572468,UC_X86_REG_ESI:0x24681357,UC_X86_REG_EDI:0x98765432,UC_X86_REG_EBP:0xfedcba98}

def run(case):
 global current,events,changed,active
 active=False;current=case;events=[];changed=False;u.mem_write(heap,seed)
 for off,value in [(8,case['level']),(16,ref(objects,case['caster'])),(48,ref(objects,case['target'])),(68,case['timer'])]:put(record+off,value)
 for i,obj in enumerate(objects):put(obj+8,case['classes'][i]);put(obj+16,case['flags'][i])
 for i in range(2):
  put(objects[i]+748,data[i]);put(data[i]+276,infos[i]);u.mem_write(infos[i]+2064,bytes([case['indexes'][i]]));u.mem_write(objects[i+2]+4,struct.pack('<H',case['types'][i]))
 u.mem_write(frameaddr-32,bytes(range(68)));put(frameaddr,case['frame']);globals_before=bytes(u.mem_read(frameaddr-32,68))
 u.mem_write(cw,struct.pack('<H',a.control));u.emu_start(init,init+8,count=2)
 before=bytes(u.mem_read(heap,len(seed)));u.mem_write(sp-512,bytes(1024));put(sp,stop);put(sp+4,record)
 for reg,val in registers.items():u.reg_write(reg,val)
 u.reg_write(UC_X86_REG_EFLAGS,0x202);u.reg_write(UC_X86_REG_ESP,sp);active=True
 u.emu_start(regions[case['kind']][0],stop+1,count=10000);active=False
 assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4
 assert get(sp)==stop and get(sp+4)==record
 assert bytes(u.mem_read(sp-512,256))==bytes(256) and bytes(u.mem_read(sp+8,256))==bytes(256)
 assert all(u.reg_read(reg)==val for reg,val in registers.items())
 assert u.reg_read(UC_X86_REG_FPCW)==a.control and u.reg_read(UC_X86_REG_FPSW)&0x3800==0
 allowed=set(range(record-heap+68,record-heap+72)) if case['kind']=='create' else set()
 changes={'balance-caster-b':[(record+16,4)],'mark-caster-b':[(record+16,4)],
  **{m:[(objects[0]+748,4)] for m in ['balance-data-b','mark-data-b','code-data-b','unmark-data-b']},
  **{m:[(data[0]+276,4)] for m in ['balance-info-b','mark-info-b','unmark-info-b']},
  **{m:[(record+48,4)] for m in ['balance-target-b','mark-target-b','code-target-b','unmark-target-b']},
  'balance-level':[(record+8,4)],'mark-index':[(infos[0]+2064,1)],'unmark-index':[(infos[0]+2064,1)],'code-type-b':[(objects[2]+4,2)],'send-timer':[(record+68,4)]}
 for address,count in changes.get(case['mutation'],[]):allowed.update(range(address-heap,address-heap+count))
 after=bytes(u.mem_read(heap,len(seed)));assert all(x==y or i in allowed for i,(x,y) in enumerate(zip(before,after)))
 globals_after=bytes(u.mem_read(frameaddr-32,68));assert all(x==y or (case['mutation']=='balance-frame' and 32<=i<36) for i,(x,y) in enumerate(zip(globals_before,globals_after)))
 if case['mutation']!='none':assert changed
 return dict(input=copy.deepcopy(case),result=dict(ret=None if case['kind']=='destroy' else u.reg_read(UC_X86_REG_EAX),state=state(),events=copy.deepcopy(events)))

template=dict(name='',kind='create',caster=0,target=2,level=3,timer=0xfedcba98,frame=0xfffffff0,
 classes=[0x280004,4,2,4],flags=[0,0,0,0],types=[0xb456,0xc789],codes=[0xa123,0xb234],indexes=[7,9],balance_bits64=f64(2.5),mutation='none')
cases=[]
def add(name,**changes):
 case=copy.deepcopy(template);case.update(name=name,**changes);cases.append(case)
for value in [0,.49,.51,1.5,2.5,3.5,-1.5,-2.5,-3.5,1.999999999999,2.999999999999,-1.999999999999,17.1,-17.1,16777217.9,-16777217.9,2147483647.,-2147483647.,2147483648.,-2147483649.,1e40,-1e40,math.inf,-math.inf,math.nan]:add('create/balance-'+f64(value),balance_bits64=f64(value))
for level in [0,1,16,0xffffffff]:add(f'create/level-{level:x}',level=level)
for cls in [2,0,6,0x280004]:add(f'create/class-{cls:x}',classes=[cls,4,2,4])
for flag in [0x20,0x8000,0x8020]:
 add(f'create/caster-flags-{flag:x}',flags=[flag,0,0,0]);add(f'create/target-flags-{flag:x}',flags=[0,0,flag,0])
add('create/no-target',target=-1)
for mutation in ['balance-caster-b','balance-data-b','balance-info-b','balance-target-b','balance-level','balance-frame','mark-data-b','mark-info-b','mark-target-b','mark-caster-b','mark-index','code-data-b','code-target-b','code-type-b','send-timer']:add('create/'+mutation,mutation=mutation)
for flags in [0,1,0x20,0x40,0x8000,0x8020,0xffffffff]:add(f'update/flags-{flags:x}',kind='update',flags=[0,0,flags,0])
add('update/no-target',kind='update',target=-1)
for target in [-1,2,3]:add(f'destroy/target-{target}',kind='destroy',target=target)
add('destroy/no-caster',kind='destroy',caster=-1)
add('destroy/nonplayer',kind='destroy',classes=[2,4,2,4])
for mutation in ['unmark-data-b','unmark-info-b','unmark-target-b','unmark-index','code-data-b','code-target-b','code-type-b','send-timer']:add('destroy/'+mutation,kind='destroy',mutation=mutation)
rows=[run(case) for case in cases]
sealed=[]
for name,(lo,hi) in regions.items():
 insns=traces[name];covered={pc+i for pc,b in insns.items() for i in range(len(b))};assert covered==set(range(lo,hi)),(name,sorted(set(range(lo,hi))-covered))
 sealed.append(dict(name=name,address=hex(lo),end_exclusive=hex(hi),bytes=hi-lo,sha256=sha(bytes(u.mem_read(lo,hi-lo))),instructions=[dict(pc=hex(pc),bytes=b.hex()) for pc,b in sorted(insns.items())]))
encoded=(json.dumps(dict(original_sha256=sha(blob),control_word=a.control,rows=rows),indent=2)+'\n').encode();a.fixture.write_bytes(encoded)
proof=dict(original_sha256=sha(blob),control_word=a.control,contexts=len(rows),entries=dict(entries),regions=sealed,all_original_instruction_bytes_executed=True,zero_arithmetic_or_conversion_replacements=True,callee_saved_registers_stack_nonoutput_heap_and_global_canaries_preserved=True,original_conversion_helper_executed=True,declared_services=['balance double via scratch FLD/RET thunk','minimap mark','minimap unmark','unit code','packet send'],mutations=dict(mutations),faults=faults,fixture_sha256=sha(encoded),scope='Complete three Tag roots and converter at configured CW. Declared services and callback mutations; no transitive service or full campaign equivalence. Destroy return register intentionally excluded from callback contract.')
assert not faults;a.proof.write_text(json.dumps(proof,indent=2)+'\n');print(json.dumps({k:v for k,v in proof.items() if k!='regions'},indent=2))
