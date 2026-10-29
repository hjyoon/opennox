"""Execute whole original GreaterHeal create52F220 under six world services.

Flags/search/adjust/enemy/ray/message bodies are declared; original argument
caching, pointer reloads, stores, branching and return behavior execute.
No transitive world-service effect or visibility equivalence is claimed.
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
record=heap+0x1000;objects=[heap+0x2000,heap+0x3000,heap+0x4000]
sp,stop,init,cw=stack+0xe000,code+0x100,code+0x200,code+0x300
get=lambda addr:struct.unpack('<I',u.mem_read(addr,4))[0]
put=lambda addr,v:u.mem_write(addr,struct.pack('<I',v&0xffffffff))
bits=lambda v:struct.unpack('<I',struct.pack('<f',v))[0]
point=lambda x,y:[bits(x),bits(y)]
readpoint=lambda addr:[get(addr),get(addr+4)]
ref=lambda i:0 if i<0 else objects[i]
index=lambda p:-1 if not p else objects.index(p)
lo,hi=0x52f220,0x52f2de
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw))
traces={};entries=collections.Counter();services=collections.Counter();active=False;current={};events=[];calls=collections.Counter();allowed=set();changed=False;hp=[];rays=messages=0
def state():
 return dict(caster=index(get(record+16)),target=index(get(record+48)),mode=get(record+20),spell=get(record+4),aim=readpoint(record+52),hp=hp.copy(),rays=rays,messages=messages)
def write(addr,v):allowed.update(range(addr,addr+4));put(addr,v)
def mutate(kind):
 global changed
 at=calls[kind];calls[kind]+=1;m=current['mutation']
 if changed or m['stage']!=kind or m['at']!=at:return
 changed=True;action=m['action']
 if action=='caster-b':write(record+16,objects[1])
 elif action=='caster-nil':write(record+16,0)
 elif action=='target-c':write(record+48,objects[2])
 elif action=='target-nil':write(record+48,0)
 elif action=='aim':
  for j,v in enumerate(point(81.25,-91.5)):write(record+52+4*j,v)
 elif action=='mode':write(record+20,1)
 elif action=='spell':write(record+4,0x87654321)
 else:raise AssertionError(action)
def event(kind,args):events.append(dict(kind=kind,args=args,state=state()));mutate(kind)
def ret(v=0):
 esp=u.reg_read(UC_X86_REG_ESP);u.reg_write(UC_X86_REG_EAX,v&0xffffffff);u.reg_write(UC_X86_REG_EIP,get(esp));u.reg_write(UC_X86_REG_ESP,esp+4)
def hook(_u,pc,length,_):
 global rays,messages
 if not active:return
 if pc==stop:u.emu_stop();return
 if lo<=pc<hi:
  traces[pc]=bytes(u.mem_read(pc,length))
  if pc==lo:entries['create']+=1
  return
 esp=u.reg_read(UC_X86_REG_ESP);arg=lambda n:get(esp+4*n);services[hex(pc)]+=1
 if pc==0x424a70:event('flags',[arg(1)]);ret(current['flags_return'])
 elif pc==0x540610:
  assert arg(1)==record+52 and arg(4)==bits(400) and arg(5)==1
  event('search',[readpoint(arg(1)),index(arg(2)),arg(3),arg(4),arg(5),index(arg(6))]);ret(ref(current['search_return']))
 elif pc==0x4ee460:
  obj,amount=index(arg(1)),arg(2);assert obj>=0 and amount==20
  event('adjust',[obj,amount]);hp[obj]+=amount;ret(0x13572468)
 elif pc==0x5330c0:event('enemy',[index(arg(1)),index(arg(2))]);ret(current['enemy_return'])
 elif pc==0x4ff130:
  assert arg(1)==record;event('ray',['record']);rays+=1;ret(0x13572468)
 elif pc==0x4da2c0:
  name=bytes(u.mem_read(arg(2),64)).split(b'\0',1)[0].decode('ascii');assert name=='ExecDur.c:GreaterHealNoTarget' and arg(3)==0
  event('message',[index(arg(1)),name,arg(3)]);messages+=1;ret(0x13572468)
 else:raise AssertionError((current['name'],hex(pc)))
u.hook_add(UC_HOOK_CODE,hook)
seed=bytes((i*31+19)&255 for i in range(0x10000));registers={UC_X86_REG_EBX:0x12345678,UC_X86_REG_ESI:0x23456789,UC_X86_REG_EDI:0x3456789a,UC_X86_REG_EBP:0x456789ab}
def run(c):
 global active,current,events,calls,allowed,changed,hp,rays,messages
 active,current,events,calls,allowed,changed=False,c,[],collections.Counter(),set(),False
 hp=c['hp'].copy();rays,messages=c['rays'],c['messages'];u.mem_write(heap,seed)
 for off,value in [(4,c['spell']),(16,ref(c['caster'])),(20,c['mode']),(48,ref(c['target']))]:put(record+off,value)
 for j,v in enumerate(c['aim']):put(record+52+4*j,v)
 before=bytes(u.mem_read(heap,len(seed)));u.mem_write(cw,struct.pack('<H',a.control));u.emu_start(init,init+8,count=2)
 u.mem_write(sp-512,bytes(1024));put(sp,stop);put(sp+4,record)
 for reg,v in registers.items():u.reg_write(reg,v)
 u.reg_write(UC_X86_REG_EFLAGS,0x202);u.reg_write(UC_X86_REG_ESP,sp);active=True;u.emu_start(lo,stop+1,count=10000);active=False
 assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4 and all(u.reg_read(reg)==v for reg,v in registers.items())
 assert u.reg_read(UC_X86_REG_FPCW)==a.control and u.reg_read(UC_X86_REG_FPSW)&0x3800==0
 assert bytes(u.mem_read(sp-512,256))==bytes(256) and bytes(u.mem_read(sp+8,256))==bytes(256)
 allowed.update(range(record+48,record+52));after=bytes(u.mem_read(heap,len(seed)))
 assert all(x==y or heap+i in allowed for i,(x,y) in enumerate(zip(before,after)))
 return dict(input=c,result=dict(return_value=u.reg_read(UC_X86_REG_EAX),events=copy.deepcopy(events),state=state()))
default=dict(name='default',caster=0,target=2,mode=0,spell=32,aim=point(100,200),hp=[5,6,7],rays=0,messages=0,flags_return=0x12345678,search_return=1,enemy_return=0,mutation=dict(stage='',at=0,action=''))
cases=[]
def case(name,**kw):c=copy.deepcopy(default);c.update(kw,name=name);cases.append(c)
for caster in [-1,0,1]:
 for mode in [0,1,0xffffffff]:
  for target in [-1,0,1,2]:
   for enemy in [0,1]:case(f'caster-{caster}-mode-{mode}-search-{target}-enemy-{enemy}',caster=caster,mode=mode,search_return=target,enemy_return=enemy)
named=[('normal-friendly',{}),('glyph-caster',dict(mode=1)),('glyph-nil',dict(mode=1,caster=-1)),('no-target',dict(search_return=-1)),('enemy',dict(enemy_return=1)),('nil-caster',dict(caster=-1))]
for name,kw in named:
 case(name,**kw)
 for stage in ['flags','search','adjust','enemy','ray','message']:
  for action in ['caster-b','caster-nil','target-c','target-nil','aim','mode','spell']:
   case(f'{name}-mutate-{stage}-{action}',**kw,mutation=dict(stage=stage,at=0,action=action))
for flags in [0,1,0xffffffff]:
 for spell in [0,32,0x7fffffff,0xffffffff]:case(f'flags-{flags}-spell-{spell}',flags_return=flags,spell=spell)
for aim in [[0,0],[0x80000000,0],[0x7f800000,0xff800000],[0x7fc12345,0xff812345]]:case(f'aim-bits-{aim}',aim=aim)
fixture=dict(control_word=a.control,original_machine_code_executed=True,cases=[run(c) for c in cases])
proof=dict(original_sha256=sha(blob),control_word=a.control,contexts=len(cases),root=dict(start=hex(lo),end_exclusive=hex(hi),bytes=hi-lo,sha256=sha(bytes(u.mem_read(lo,hi-lo))),covered_bytes=sum(len(b) for b in traces.values()),missing_bytes=[hex(i) for i in range(lo,hi) if not any(pc<=i<pc+len(b) for pc,b in traces.items())]),entries=dict(entries),services=dict(services),register_stack_FPCW_TOP_heap_canaries_preserved=True,faults=0,original_machine_code_executed=True,independent_comparator_is_external=True,scope=__doc__)
for path,value in [(a.fixture,fixture),(a.proof,proof)]:
 with path.open('x') as f:json.dump(value,f,indent=2);f.write('\n')
print(json.dumps(proof))
