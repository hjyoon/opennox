"""Execute all49 original dying-end bytes; IsZombie/BurnDelete bodies are services.

The original cached entry update, live flags after IsZombie, call order and
unit identity execute unchanged. Void ABI has no return-value comparison;
raw original EAX is retained for audit only. Fifteen NOPs remain unexecuted.
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
for addr in [heap,stack,code]:u.mem_map(addr,0x10000)
unit,primary,alternate=heap+0x1000,heap+0x3000,heap+0x6000
sp,stop,init,cw=stack+0xe000,code+0x100,code+0x200,code+0x300
read32=lambda addr:struct.unpack('<I',u.mem_read(addr,4))[0]
put32=lambda addr,v:u.mem_write(addr,struct.pack('<I',v&0xffffffff))
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw))
seed=bytes((i*31+19)&255 for i in range(0x10000))
registers={UC_X86_REG_EBX:0x12345678,UC_X86_REG_ESI:0x23456789,UC_X86_REG_EDI:0x3456789a,UC_X86_REG_EBP:0x456789ab}
active=False;current={};events=[];traces={};services=collections.Counter()
def state():
 ptr=read32(unit+748);assert ptr in [primary,alternate]
 return dict(update=0 if ptr==primary else 1,cls=read32(unit+8),subclass=read32(unit+12),flags=[read32(primary+1440),read32(alternate+1440)])
def mutate(stage):
 m=current['mutation']
 if m['stage']!=stage:return
 action=m['action']
 if action in ['update','all']:put32(unit+748,alternate)
 if action in ['primary','all']:put32(primary+1440,m['flags'])
 if action in ['alternate','all']:put32(alternate+1440,m['flags'])
 if action in ['class','all']:put32(unit+8,0x400000)
 if action in ['subclass','all']:put32(unit+12,0xfedcba98)
def hook(_u,pc,size,_):
 if not active:return
 if pc==stop:u.emu_stop();return
 if 0x544ca0<=pc<0x544cd1:traces[pc]=bytes(u.mem_read(pc,size));return
 kind={0x534a40:'zombie',0x544ce0:'burn'}.get(pc);assert kind,(current['name'],hex(pc))
 esp=u.reg_read(UC_X86_REG_ESP);assert read32(esp+4)==unit
 events.append(dict(kind=kind,unit=0,state=state()));services[kind]+=1;mutate(kind)
 u.reg_write(UC_X86_REG_EAX,current['zombie'] if kind=='zombie' else current['burn_return'])
 u.reg_write(UC_X86_REG_EIP,read32(esp));u.reg_write(UC_X86_REG_ESP,esp+4)
u.hook_add(UC_HOOK_CODE,hook)
def run(c):
 global active,current,events
 active,current,events=False,c,[];u.mem_write(heap,seed)
 put32(unit+8,2);put32(unit+12,0xa5);put32(unit+748,primary)
 put32(primary+1440,c['flags']);put32(alternate+1440,c['alternate_flags'])
 before=bytes(u.mem_read(heap,len(seed)))
 u.mem_write(cw,struct.pack('<H',a.control));u.emu_start(init,init+8,count=2)
 u.mem_write(sp-512,bytes(1024));put32(sp,stop);put32(sp+4,unit)
 for reg,v in registers.items():u.reg_write(reg,v)
 u.reg_write(UC_X86_REG_EFLAGS,0x202);u.reg_write(UC_X86_REG_ESP,sp)
 active=True
 try:u.emu_start(0x544ca0,stop+1,count=1000)
 except Exception as e:raise AssertionError((c['name'],hex(u.reg_read(UC_X86_REG_EIP)))) from e
 active=False
 assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4,c['name']
 assert all(u.reg_read(reg)==v for reg,v in registers.items()),c['name']
 assert u.reg_read(UC_X86_REG_FPCW)==a.control and u.reg_read(UC_X86_REG_FPSW)&0x3800==0,c['name']
 assert bytes(u.mem_read(sp-512,256))==bytes(256) and bytes(u.mem_read(sp+8,256))==bytes(256),c['name']
 after=bytes(u.mem_read(heap,len(seed)))
 allowed=set().union(*(range(addr,addr+4) for addr in [unit+8,unit+12,unit+748,primary+1440,alternate+1440]))
 assert all(x==y or heap+i in allowed for i,(x,y) in enumerate(zip(before,after))),c['name']
 return dict(input=c,result=dict(events=copy.deepcopy(events),state=state()),raw_original_eax=u.reg_read(UC_X86_REG_EAX))
cases=[]
def case(prefix,**kw):
 c=dict(zombie=1,flags=0x80000,alternate_flags=0,burn_return=0x87654321,mutation=dict(stage='',action='',flags=0))
 c.update(kw,name=f'{prefix}-{len(cases)}');cases.append(c)
flags=[0,1,0x7ffff,0x80000,0x80001,0xffffffff,0x100000,0x7fffffff]
for q,f in itertools.product([0,1,2,127,128,0x80000000,0xffffffff],flags):case('predicate-flags',zombie=q,flags=f)
for stage,action,f,v in itertools.product(['zombie','burn'],['update','primary','alternate','class','subclass','all'],[0,0x80000],[0,0x80000,0xffffffff]):case('mutation',flags=f,mutation=dict(stage=stage,action=action,flags=v))
for ret in [0,1,0x7fffffff,0x80000000,0xffffffff]:case('void-return-audit',burn_return=ret)
assert len(cases)==133
rows=[run(c) for c in cases]
executed={pc+j for pc,b in traces.items() for j in range(len(b))};assert executed==set(range(0x544ca0,0x544cd1))
fixture=dict(original_sha256=sha(blob),control_word=a.control,rows=rows)
proof=dict(original_sha256=sha(blob),control_word=a.control,contexts=len(cases),faults=0,root_bytes=49,executed_bytes=len(executed),existing_manifest_envelope=64,padding_bytes_unexecuted=15,root_sha256=sha(bytes(u.mem_read(0x544ca0,49))),services=dict(services),only_two_service_bodies_declared=['534A40 IsZombie','544CE0 BurnDelete'],raw_original_EAX_retained_void_return_excluded=True,register_stack_FPCW_TOP_heap_canaries_preserved=True,cached_entry_update_and_live_status_load_execute_unmodified=True,instructions=[dict(pc=hex(pc),bytes=b.hex()) for pc,b in sorted(traces.items())])
for path,data in [(a.fixture,fixture),(a.proof,proof)]:
 with path.open('x') as f:json.dump(data,f,indent=2);f.write('\n')
print(json.dumps(dict(contexts=len(cases),executed_bytes=len(executed),control_word=hex(a.control),services=dict(services),fixture_sha256=sha(a.fixture.read_bytes()))))
