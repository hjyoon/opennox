"""Execute complete218-byte DeadStart and actual17-byte float converter.

Soul creation, the selected DeadFunc body, IsZombie, balance lookup and logic
RNG bodies are declared services. Both original rounding profiles are kept
separate; the public char is recorded separately from full original EAX.
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
unit,primary,alternate=heap+0x1000,heap+0x4000,heap+0x8000
defs=[heap+0xc000,heap+0xe000,heap+0x10000];functions=[0,code+0x1000,code+0x1100,code+0x1200]
sp,stop,init,cw,stub,value=stack+0xe000,code+0x100,code+0x200,code+0x300,code+0x400,code+0x500
rd=lambda addr:struct.unpack('<I',u.mem_read(addr,4))[0]
def put(addr,v):u.mem_write(addr,struct.pack('<I',v&0xffffffff))
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw));u.mem_write(stub,b'\xd9\x05'+struct.pack('<I',value)+b'\xc3')
seed=bytes((i*31+19)&255 for i in range(0x20000))
regs={UC_X86_REG_EBX:0x12345678,UC_X86_REG_ESI:0x23456789,UC_X86_REG_EDI:0x3456789a,UC_X86_REG_EBP:0x456789ab}
active=False;current={};events=[];balance=[];queries=0;trace={};converter={};services=collections.Counter()
def state():
 ptr=rd(unit+748);assert ptr in [primary,alternate]
 return dict(update=0 if ptr==primary else 1,cls=rd(unit+8),subclass=rd(unit+12),flags=rd(unit+16),field131=rd(unit+524),motion=[rd(unit+i) for i in range(80,104,4)],definitions=[defs.index(rd(ud+484)) for ud in [primary,alternate]],callbacks=[functions.index(rd(d+232)) for d in defs],durations=[rd(ud+492) for ud in [primary,alternate]],balance=list(balance))
def mutate(stage):
 m=current['mutation']
 if m['stage']!=stage:return
 action=m['action'];assert action in ['update','definition','callback','flags','motion','primary_duration','alternate_duration','class','subclass','field131','balance','all']
 if action in ['update','all']:put(unit+748,alternate)
 if action in ['definition','all']:put(primary+484,defs[1])
 if action in ['callback','all']:put(rd(primary+484)+232,functions[2])
 if action in ['flags','all']:put(unit+16,0x530000a5)
 if action in ['motion','all']:
  for offset,word in zip(range(80,104,4),[0x80000000,0x7fc00001,0x40a00000,0xc0a00000,0x3f000000,0xbf000000]):put(unit+offset,word)
 if action in ['primary_duration','all']:put(primary+492,0x02030405)
 if action in ['alternate_duration','all']:put(alternate+492,0x05060708)
 if action in ['class','all']:put(unit+8,0x400000)
 if action in ['subclass','all']:put(unit+12,0xfedcba98)
 if action in ['field131','all']:put(unit+524,0x3344)
 if action in ['balance','all']:balance[:]=[0x40200000,0x40600000]
def hook(_u,pc,size,_):
 global queries
 if not active:return
 if pc==stop:u.emu_stop();return
 if 0x544d80<=pc<0x544e5a:trace[pc]=bytes(u.mem_read(pc,size));return
 if 0x419a70<=pc<0x419a81:converter[pc]=bytes(u.mem_read(pc,size));return
 if stub<=pc<stub+7:return
 esp=u.reg_read(UC_X86_REG_ESP);arg=0;args=[]
 if pc==0x544e60:kind,stage='soul','soul';assert rd(esp+4)==unit;result=0x98765432
 elif pc in functions[1:]:kind,stage='dead','dead';arg=functions.index(pc);assert rd(esp+4)==unit;result=0x87654321
 elif pc==0x534a40:
  kind='zombie';queries+=1;stage=f'zombie{queries}';arg=queries;assert rd(esp+4)==unit and queries<=len(current['zombie'])
  result=current['zombie'][queries-1]
 elif pc==0x419d70:
  kind='balance';arg=rd(esp+8);stage=f'balance{arg}';assert arg in [0,1] and bytes(u.mem_read(rd(esp+4),32)).split(b'\0',1)[0]==b'ZombieDeadDuration';result=0
 elif pc==0x415fa0:
  kind,stage='random','random';args=[struct.unpack('<i',u.mem_read(esp+i,4))[0] for i in [4,8]];assert rd(esp+16)==326;result=current['random_return']
 else:raise AssertionError((current['name'],hex(pc)))
 e=dict(kind=kind,arg=arg,args=args,state=state())
 if kind=='random':e['original_debug_file']=bytes(u.mem_read(rd(esp+12),96)).split(b'\0',1)[0].decode('ascii');e['original_debug_line']=rd(esp+16)
 events.append(e);services[kind]+=1;mutate(stage)
 if kind=='balance':put(value,balance[arg]);u.reg_write(UC_X86_REG_EIP,stub);return
 u.reg_write(UC_X86_REG_EAX,result&0xffffffff);u.reg_write(UC_X86_REG_EIP,rd(esp));u.reg_write(UC_X86_REG_ESP,esp+4)
u.hook_add(UC_HOOK_CODE,hook)
def run(c):
 global active,current,events,balance,queries
 active,current,events,balance,queries=False,c,[],list(c['balance']),0
 u.mem_write(heap,seed);put(unit+8,2);put(unit+12,c['subclass']);put(unit+16,c['flags']);put(unit+524,c['field131']);put(unit+748,primary)
 for offset,word in zip(range(80,104,4),[0x3f800000,0x40000000,0x40400000,0x40800000,0x40a00000,0x40c00000]):put(unit+offset,word)
 put(primary+484,defs[c['definition']]);put(alternate+484,defs[2]);put(primary+492,0xcccc1111);put(alternate+492,0xdddd2222)
 for i,d in enumerate(defs):put(d+232,functions[i+1])
 put(defs[c['definition']]+232,functions[c['callback']])
 before=bytes(u.mem_read(heap,len(seed)));put(cw,a.control);u.emu_start(init,init+8,count=2)
 u.mem_write(sp-512,bytes(1024));put(sp,stop);put(sp+4,unit)
 for reg,v in regs.items():u.reg_write(reg,v)
 u.reg_write(UC_X86_REG_ESP,sp);u.reg_write(UC_X86_REG_EFLAGS,0x202);active=True
 try:u.emu_start(0x544d80,stop+1,count=3000)
 except Exception as e:raise AssertionError((c['name'],hex(u.reg_read(UC_X86_REG_EIP)))) from e
 active=False
 assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4,c['name']
 assert all(u.reg_read(reg)==v for reg,v in regs.items()),c['name']
 assert u.reg_read(UC_X86_REG_FPCW)==a.control and not u.reg_read(UC_X86_REG_FPSW)&0x3800,c['name']
 assert bytes(u.mem_read(sp-512,256))==bytes(256) and bytes(u.mem_read(sp+8,256))==bytes(256),c['name']
 after=bytes(u.mem_read(heap,len(seed)))
 stores=[unit+8,unit+12,unit+16,unit+524,unit+748,*range(unit+80,unit+104,4),primary+484,primary+492,alternate+492,*[d+232 for d in defs]]
 allowed=set().union(*(range(addr,addr+4) for addr in stores))
 assert all(x==y or heap+i in allowed for i,(x,y) in enumerate(zip(before,after))),c['name']
 raw=u.reg_read(UC_X86_REG_EAX)
 return dict(input=copy.deepcopy(c),result=dict(return_byte=raw&255,events=copy.deepcopy(events),state=state()),raw_original_eax=raw)
cases=[]
def case(prefix,**kw):
 c=dict(field131=14,subclass=0x10000,flags=0x12000004,definition=0,callback=1,zombie=[1],balance=[0x40fccccd,0x414e6666],random_return=0x8899aabb,mutation=dict(stage='',action=''))
 c.update(kw,name=f'{prefix}-{len(cases)}');cases.append(c)
patterns=[[1],[0,0],[0,1]]
for field,subclass,zombie,flags,callback in itertools.product([0,14,15],[0,0x10000,0x10010],patterns,[0,127,128,255,0x12000004,0xffffffff],[0,1]):case('branch-flags',field131=field,subclass=subclass,zombie=zombie,flags=flags,callback=callback)
pairs=[[0,0],[0x3f800000,0x40000000],[0x3f000000,0x3fc00000],[0x40200000,0x40600000],[0x40fccccd,0x414e6666],[0xc0fccccd,0xc14e6666],[0x414e6666,0x40fccccd],[0x4effffff,0x4f000000],[0xcf000000,0xcf000001],[0x7f800000,0xff800000],[0x7fc00001,0x3f800000],[1,0x80000001],[0,0x80000000],[0x4effffff,0x4efffffe],[0x3fffffff,0x40000001],[0x80000000,0]]
for pair,zombie in itertools.product(pairs,[[1],[0x80000000],[0xffffffff]]):case('duration-convert',balance=pair,zombie=zombie)
for fn,soul,zombie,definition in itertools.product([0,1,2,3],[False,True],patterns,[0,1,2]):case('live-definition',callback=fn,field131=14 if soul else 0,zombie=zombie,definition=definition)
for stage,action,fn in itertools.product(['soul','dead','zombie1','zombie2','balance0','balance1','random'],['update','definition','callback','flags','motion','primary_duration','alternate_duration','class','subclass','field131','balance','all'],[0,1]):case('mutation',callback=fn,zombie=[0,0] if stage=='zombie2' else [1],mutation=dict(stage=stage,action=action))
for low,returned in itertools.product([0,128,255],range(7,13)):case('native-duration-alias',field131=0,subclass=0,flags=low,callback=0,zombie=[1],random_return=returned)
assert len(cases)==630
rows=[run(c) for c in cases]
executed={pc+i for pc,b in trace.items() for i in range(len(b))};converted={pc+i for pc,b in converter.items() for i in range(len(b))}
assert executed==set(range(0x544d80,0x544e5a)) and converted==set(range(0x419a70,0x419a81))
fixture=dict(original_sha256=sha(blob),control_word=a.control,rows=rows)
proof=dict(original_sha256=sha(blob),control_word=a.control,contexts=len(cases),faults=0,root_instruction_bytes=218,executed_root_bytes=len(executed),converter_bytes=len(converted),existing_root_envelope=224,padding_bytes_unexecuted=6,root_sha256=sha(bytes(u.mem_read(0x544d80,218))),converter_sha256=sha(bytes(u.mem_read(0x419a70,17))),services=dict(services),service_bodies_declared=['ReleasedSoul544E60','selectedDeadFunction','IsZombie534A40','Gamedata419D70 binary32 return','Random415FA0'],raw_original_EAX_and_char_retained=True,original_random_debug_metadata_retained_separately=True,register_stack_FPCW_TOP_heap_canaries_preserved=True,instructions=[dict(pc=hex(pc),bytes=b.hex()) for pc,b in sorted(trace.items())],converter_instructions=[dict(pc=hex(pc),bytes=b.hex()) for pc,b in sorted(converter.items())])
for path,data in [(a.fixture,fixture),(a.proof,proof)]:
 with path.open('x') as f:json.dump(data,f,indent=2);f.write('\n')
print(json.dumps(dict(contexts=len(cases),root_bytes=len(executed),converter_bytes=len(converted),control_word=hex(a.control),services=dict(services),fixture_sha256=sha(a.fixture.read_bytes()))))
