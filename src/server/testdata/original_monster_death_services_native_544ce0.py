"""Execute the original113-byte BurnDelete and84-byte ReleasedSoul callers.

Shield, minimap, reward, scorch, deletion, name lookup, indexed allocation and
world creation are declared services. Repeated Soul calls execute the real
lazy DWORD cache, including callback mutations. Original code is unmodified.
"""
import argparse,collections,copy,hashlib,itertools,json,struct
from pathlib import Path
from unicorn import Uc,UC_ARCH_X86,UC_MODE_32,UC_HOOK_CODE
from unicorn.x86_const import *
p=argparse.ArgumentParser(description=__doc__)
for n in ['game','fixture','proof']:p.add_argument('--'+n,type=Path,required=True)
p.add_argument('--control',type=lambda v:int(v,0),choices=[0xe3f,0x23f],default=0xe3f)
a=p.parse_args();blob=a.game.read_bytes();sha=lambda b:hashlib.sha256(b).hexdigest()
assert sha(blob)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,60)[0];opt=pe+24;base=struct.unpack_from('<I',blob,opt+28)[0];table=opt+struct.unpack_from('<H',blob,pe+20)[0]
u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,(struct.unpack_from('<I',blob,opt+56)[0]+4095)&~4095)
for i in range(struct.unpack_from('<H',blob,pe+6)[0]):
 _,va,size,off=struct.unpack_from('<IIII',blob,table+40*i+8)
 if size:u.mem_write(base+va,blob[off:off+size])
heap,stack,code=0x21000000,0x22000000,0x23000000
u.mem_map(heap,0x20000)
for addr in [stack,code]:u.mem_map(addr,0x10000)
unit=heap+0x1000;owners=[heap+0x3000,heap+0x4000];owner_updates=[heap+0x5000,heap+0x6000];players=[heap+0x8000,heap+0xa000];souls=[heap+0xc000,heap+0xd000]
cache=0x834204;sp,stop,init,cw=stack+0xe000,code+0x100,code+0x200,code+0x300
read32=lambda addr:struct.unpack('<I',u.mem_read(addr,4))[0]
put32=lambda addr,v:u.mem_write(addr,struct.pack('<I',v&0xffffffff))
read16=lambda addr:struct.unpack('<H',u.mem_read(addr,2))[0]
put16=lambda addr,v:u.mem_write(addr,struct.pack('<H',v&0xffff))
put8=lambda addr,v:u.mem_write(addr,bytes([v&255]))
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw))
seed=bytes((i*31+19)&255 for i in range(0x20000))
registers={UC_X86_REG_EBX:0x12345678,UC_X86_REG_ESI:0x23456789,UC_X86_REG_EDI:0x3456789a,UC_X86_REG_EBP:0x456789ab}
ranges={0x544ce0:113,0x544e60:84};original={pc:bytes(u.mem_read(pc,n)) for pc,n in ranges.items()}
active=False;current={};events=[];traces={};services=collections.Counter();lookup_count=0;allocation_count=0;iteration=0
def object_index(pointer):
 if pointer==0:return -1
 return [souls[0],souls[1],unit].index(pointer)
def state():
 return dict(owner=-1 if read32(unit+508)==0 else owners.index(read32(unit+508)),owner_classes=[read32(v+8) for v in owners],
  owner_updates=[owner_updates.index(read32(v+748)) for v in owners],players=[players.index(read32(v+276)) for v in owner_updates],
  player_indices=[u.mem_read(v+2064,1)[0] for v in players],subclass=read32(unit+12),flags=read32(unit+16),position=[read32(unit+56),read32(unit+60)],
  directions=[[read16(v+124),read16(v+126)] for v in [unit]+souls],cache=read32(cache))
def mutate(stage,obj=-1):
 m=current['mutation']
 if m['stage']!=stage:return
 action,value=m['action'],m['value']
 if current['entry']=='burn':
  assert action in ['owner','owner-update','player','index','subclass','position','class','all']
  if action in ['owner','all']:put32(unit+508,owners[value])
  if action in ['owner-update','all']:put32(owners[0]+748,owner_updates[value])
  if action in ['player','all']:put32(owner_updates[0]+276,players[value])
  if action in ['index','all']:put8(players[0]+2064,255 if value else 1)
  if action in ['subclass','all']:put32(unit+12,0xfedcba98 if value else 0x12340080)
  if action in ['position','all']:put32(unit+56,0xc2160000);put32(unit+60,0x419a0000)
  if action in ['class','all']:put32(owners[0]+8,0x400000 if value else 0)
 else:
  assert action in ['cache','unit-direction','position','soul-direction','all']
  if action in ['cache','all']:put32(cache,18 if value else 0)
  if action in ['unit-direction','all']:put16(unit+124,0xffff if value else 0x1234)
  if action in ['position','all']:put32(unit+56,0xc2160000);put32(unit+60,0x419a0000)
  if action in ['soul-direction','all']:
   target=[souls[0],souls[1],unit][max(0,obj)];put16(target+124,0xa55a);put16(target+126,0x5aa5)
def hook(_u,pc,size,_):
 global lookup_count,allocation_count
 if not active:return
 if pc==stop:u.emu_stop();return
 if any(start<=pc<start+n for start,n in ranges.items()):traces[pc]=bytes(u.mem_read(pc,size));return
 kind={0x4d9200:'shield',0x417300:'unmark',0x4ee500:'reward',0x537af0:'scorch',0x4e5cc0:'delete',0x4e3aa0:'lookup',0x4e3450:'new',0x4daa50:'create'}.get(pc)
 assert kind,(current['name'],hex(pc));esp=u.reg_read(UC_X86_REG_ESP);args=[];ret=0x87654321;obj=-1
 if kind=='shield':assert read32(esp+8)==unit;args=[read32(esp+4),0]
 elif kind=='unmark':assert read32(esp+8)==unit and read32(esp+12)==1;args=[read32(esp+4),0,1]
 elif kind in ['reward','delete']:assert read32(esp+4)==unit;args=[0]
 elif kind=='scorch':assert read32(esp+4)==unit+56 and read32(esp+8)==1;args=[read32(unit+56),read32(unit+60),1]
 elif kind=='lookup':
  pointer=read32(esp+4);assert bytes(u.mem_read(pointer,13))==b'ReleasedSoul\x00';args=[]
  ret=current['lookup_values'][lookup_count];lookup_count+=1
 elif kind=='new':
  args=[read32(esp+4)];obj=current['allocations'][allocation_count];allocation_count+=1
  ret=0 if obj<0 else [souls[0],souls[1],unit][obj]
 else:
  obj=object_index(read32(esp+4));assert obj>=0 and read32(esp+8)==0 and read32(esp+20)==0
  args=[obj,0,read32(esp+12),read32(esp+16),0]
 events.append(dict(kind=kind,args=args,iteration=iteration,state=state()));services[kind]+=1;mutate(kind,obj)
 u.reg_write(UC_X86_REG_EAX,ret);u.reg_write(UC_X86_REG_EIP,read32(esp));u.reg_write(UC_X86_REG_ESP,esp+4)
u.hook_add(UC_HOOK_CODE,hook)
allowed=set()
for ptr,off,n in [(unit,12,8),(unit,56,8),(unit,124,4),(unit,508,4)]+[(v,8,4) for v in owners]+[(v,748,4) for v in owners]+[(v,276,4) for v in owner_updates]+[(v,2064,1) for v in players]+[(v,124,4) for v in souls]:allowed.update(range(ptr+off,ptr+off+n))
def run(c):
 global active,current,events,lookup_count,allocation_count,iteration
 active,current,events,lookup_count,allocation_count=False,c,[],0,0;u.mem_write(heap,seed)
 put32(unit+8,2);put32(unit+12,c['subclass']);put32(unit+16,0x12000004);put32(unit+508,0 if c['owner']<0 else owners[c['owner']])
 put32(unit+56,c['position'][0]);put32(unit+60,c['position'][1]);put16(unit+124,c['direction']);put16(unit+126,0xbabe)
 for n,v in enumerate(owners):put32(v+8,c['owner_class'] if n==0 else 4);put32(v+748,owner_updates[n])
 for n,v in enumerate(owner_updates):put32(v+276,players[n])
 for n,v in enumerate(players):put8(v+2064,c['player_index'] if n==0 else 9)
 for n,v in enumerate(souls):put16(v+124,0xa500+n);put16(v+126,0x5a00+n)
 put32(cache,c['cache']);before=bytes(u.mem_read(heap,len(seed)));raw_returns=[];snapshots=[]
 for iteration in range(c['repeats']):
  u.mem_write(cw,struct.pack('<H',a.control));u.emu_start(init,init+8,count=2)
  u.mem_write(sp-512,bytes(1024));put32(sp,stop);put32(sp+4,unit)
  for reg,v in registers.items():u.reg_write(reg,v)
  u.reg_write(UC_X86_REG_EFLAGS,0x202);u.reg_write(UC_X86_REG_ESP,sp);active=True
  try:u.emu_start(0x544ce0 if c['entry']=='burn' else 0x544e60,stop+1,count=10000)
  except Exception as e:raise AssertionError((c['name'],iteration,hex(u.reg_read(UC_X86_REG_EIP)))) from e
  active=False
  assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4,(c['name'],iteration)
  assert all(u.reg_read(reg)==v for reg,v in registers.items()),(c['name'],iteration)
  assert u.reg_read(UC_X86_REG_FPCW)==a.control and u.reg_read(UC_X86_REG_FPSW)&0x3800==0,(c['name'],iteration)
  assert bytes(u.mem_read(sp-512,256))==bytes(256) and bytes(u.mem_read(sp+8,256))==bytes(256),(c['name'],iteration)
  raw_returns.append(u.reg_read(UC_X86_REG_EAX));snapshots.append(state())
 after=bytes(u.mem_read(heap,len(seed)));assert all(x==y or heap+i in allowed for i,(x,y) in enumerate(zip(before,after))),c['name']
 return dict(input=c,result=dict(events=copy.deepcopy(events),state=state(),snapshots=snapshots),raw_original_eax=raw_returns)
cases=[]
def case(prefix,**kw):
 c=dict(entry='burn',owner=0,owner_class=4,player_index=4,subclass=0x12340080,position=[0x43a08000,0x44238000],direction=13,cache=0,lookup_values=[17,29,41],allocations=[0,0,0],repeats=1,mutation=dict(stage='',action='',value=0))
 c.update(kw,name=f'{prefix}-{len(cases)}');cases.append(c)
for cls,sub,index in itertools.product([0,2,4,6,0x104,0xffffffff],[0,0x7f,0x80,0xff,0x12345678,0xdeadbeef,0xffffffff],[0,31,127,255]):case('burn-owner-mask',owner_class=cls,subclass=sub,player_index=index)
for sub in [0,0x7f,0x80,0xff,0x12345678,0xdeadbeef,0xffffffff]:case('burn-no-owner',owner=-1,subclass=sub)
for stage,action,value in itertools.product(['shield','unmark','reward','scorch','delete'],['owner','owner-update','player','index','subclass','position','class','all'],range(2)):case('burn-live',mutation=dict(stage=stage,action=action,value=value))
for cached,lookup,allocation,direction in itertools.product([0,17,0x80000000,0xffffffff],[0,17,0xffffffff],[-1,0,1,2],[0,255,0xffff]):case('soul-cache',entry='soul',cache=cached,lookup_values=[lookup,lookup^0x22,lookup^0x44],allocations=[allocation]*3,direction=direction,repeats=2)
for stage,action,value,allocation in itertools.product(['lookup','new','create'],['cache','unit-direction','position','soul-direction','all'],range(2),[0,2]):case('soul-live',entry='soul',allocations=[allocation]*3,repeats=3,mutation=dict(stage=stage,action=action,value=value))
for bits in [0,0x80000000,0x7f800000,0xff800000,0x7fc12345,0x7fa12345,0x7f7fffff,1]:case('soul-position-bits',entry='soul',position=[bits,bits^0x80000000])
assert len(cases)==467
rows=[run(c) for c in cases]
executed={pc+j for pc,b in traces.items() for j in range(len(b))};assert executed=={pc+j for pc,n in ranges.items() for j in range(n)}
assert all(bytes(u.mem_read(pc,n))==original[pc] for pc,n in ranges.items())
fixture=dict(original_sha256=sha(blob),control_word=a.control,rows=rows)
proof=dict(original_sha256=sha(blob),control_word=a.control,contexts=len(cases),burn_contexts=255,soul_contexts=212,original_invocations=sum(c['repeats'] for c in cases),faults=0,root_bytes=ranges,executed_bytes=len(executed),root_sha256={hex(pc):sha(b) for pc,b in original.items()},services=dict(services),
 register_stack_FPCW_TOP_heap_canaries_preserved=True,original_lazy_DWORD_cache_and_callback_mutations_execute_unmodified=True,Soul_create_fifth_zero_word_retained_in_raw_arguments=True,public_void_raw_EAX_retained_but_excluded_from_contract=True,
 unexecuted_padding={'544CE0':15,'544E60':12},new_exact_root_ranges_required=True,cache_address=hex(cache),cache_width=4,
 instructions=[dict(pc=hex(pc),bytes=b.hex()) for pc,b in sorted(traces.items())])
for path,data in [(a.fixture,fixture),(a.proof,proof)]:
 with path.open('x') as f:json.dump(data,f,indent=2);f.write('\n')
print(json.dumps(dict(contexts=len(cases),invocations=proof['original_invocations'],executed_bytes=len(executed),control_word=hex(a.control),services=dict(services),fixture_sha256=sha(a.fixture.read_bytes()))))
