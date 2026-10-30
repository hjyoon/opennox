"""Execute unmodified101-byte RandomWalk and171-byte direction helper.

RNG, tile query and MoveAudio are declared services. Both root raw EAX and
its low public byte are retained, separately from the handled-bool Go API.
"""
import argparse,collections,copy,hashlib,itertools,json,struct
from pathlib import Path
from unicorn import Uc,UC_ARCH_X86,UC_MODE_32,UC_HOOK_CODE
from unicorn.x86_const import *
p=argparse.ArgumentParser(description=__doc__)
for name in ['game','fixture','proof']:p.add_argument('--'+name,type=Path,required=True)
p.add_argument('--control',type=lambda s:int(s,0),choices=[0xe3f,0x23f],required=True)
a=p.parse_args();blob=a.game.read_bytes();sha=lambda b:hashlib.sha256(b).hexdigest()
assert sha(blob)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,60)[0];op=pe+24;tb=op+struct.unpack_from('<H',blob,pe+20)[0];base=struct.unpack_from('<I',blob,op+28)[0]
u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,(struct.unpack_from('<I',blob,op+56)[0]+4095)&~4095)
for i in range(struct.unpack_from('<H',blob,pe+6)[0]):
 _,va,n,off=struct.unpack_from('<IIII',blob,tb+i*40+8)
 if n:u.mem_write(base+va,blob[off:off+n])
heap,stack,code=0x21000000,0x22000000,0x23000000
for addr in [heap,stack,code]:u.mem_map(addr,0x10000)
unit=heap+0x1000;updates=[heap+0x3000,heap+0x4000];defs=[heap+0x6000,heap+0x7000,heap+0x8000]
sp,stop,init,cw=stack+0xe000,code+0x100,code+0x200,code+0x300
rd=lambda v:struct.unpack('<I',u.mem_read(v,4))[0]
wr=lambda v,n:u.mem_write(v,struct.pack('<I',n&0xffffffff))
r16=lambda v:struct.unpack('<H',u.mem_read(v,2))[0]
w16=lambda v,n:u.mem_write(v,struct.pack('<H',n&0xffff))
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw))
ranges={0x545020:101,0x545090:171};original={pc:bytes(u.mem_read(pc,n)) for pc,n in ranges.items()}
seed=bytes((i*31+19)&255 for i in range(0x10000));events=[];traces={};services=collections.Counter();current={};active=False
registers={UC_X86_REG_EBX:0x12345678,UC_X86_REG_ESI:0x23456789,UC_X86_REG_EDI:0x3456789a,UC_X86_REG_EBP:0x456789ab}
def state():
 return dict(update=updates.index(rd(unit+748)),cls=rd(unit+8),subclass=rd(unit+12),flags=rd(unit+16),position=[rd(unit+56),rd(unit+60)],directions=[r16(unit+124),r16(unit+126)],force=[rd(unit+88),rd(unit+92)],speed=rd(unit+544),statuses=[rd(v+1440) for v in updates],definitions=[defs.index(rd(v+484)) for v in updates],multipliers=[rd(v+96) for v in defs])
def mutate(stage):
 m=current['mutation']
 if m['stage']!=stage:return
 action,value=m['action'],m['value']
 if action in ['direction','all']:w16(unit+124,0xffff if value else 129);w16(unit+126,0x1234)
 if action in ['subclass','all']:wr(unit+12,0x480 if value else 0x80)
 if action in ['class','all']:wr(unit+8,0x400000 if value else 2)
 if action in ['update','all']:wr(unit+748,updates[value])
 if action in ['entry-status','all']:wr(updates[0]+1440,0x4000 if value else 0)
 if action in ['alternate-status','all']:wr(updates[1]+1440,0 if value else 0x4000)
 if action in ['entry-definition','all']:wr(updates[0]+484,defs[1])
 if action in ['alternate-definition','all']:wr(updates[1]+484,defs[value])
 if action in ['run-multiplier','all']:wr(defs[0]+96,0x3f000001 if value else 0x3fe00000)
 if action in ['speed','all']:wr(unit+544,0xbf900000 if value else 0x40499999)
 if action in ['position','all']:wr(unit+56,0xc2160000);wr(unit+60,0x419a0000)
def hook(_u,pc,size,_):
 if not active:return
 if pc==stop:u.emu_stop();return
 if any(start<=pc<start+n for start,n in ranges.items()):traces[pc]=bytes(u.mem_read(pc,size));return
 esp=u.reg_read(UC_X86_REG_ESP)
 if pc==0x415fa0:
  assert rd(esp+4)==0xffffffec and rd(esp+8)==20 and rd(esp+12)==0x5cc470 and rd(esp+16)==435
  kind,args,ret='random',[-20,20,435],current['random']&0xffffffff
 elif pc==0x411160:
  ptr=rd(esp+4);assert stack<=ptr<stack+0x10000
  kind,args,ret='tile',[rd(ptr),rd(ptr+4)],current['tile']&0xffffffff
 elif pc==0x534030:
  assert rd(esp+4)==unit;kind,args,ret='move_audio',[0],current['audio_return']
 else:raise AssertionError((current['name'],hex(pc)))
 events.append(dict(kind=kind,args=args,state=state()));services[kind]+=1;mutate(kind)
 u.reg_write(UC_X86_REG_EAX,ret);u.reg_write(UC_X86_REG_EIP,rd(esp));u.reg_write(UC_X86_REG_ESP,esp+4)
u.hook_add(UC_HOOK_CODE,hook)
allowed=set()
for ptr,off,n in [(unit,8,12),(unit,56,8),(unit,88,8),(unit,124,4),(unit,544,4),(unit,748,4)]+[(v,1440,4) for v in updates]+[(v,484,4) for v in updates]+[(v,96,4) for v in defs]:allowed.update(range(ptr+off,ptr+off+n))
def run(c):
 global current,events,active
 active=False;current=c;events=[];u.mem_write(heap,seed)
 wr(unit+8,2);wr(unit+12,c['subclass']);wr(unit+16,0x12000004);wr(unit+56,c['position'][0]);wr(unit+60,c['position'][1]);wr(unit+88,0x3f400001);wr(unit+92,0xbf000001);wr(unit+544,c['speed']);wr(unit+748,updates[0]);w16(unit+124,c['direction']);w16(unit+126,0xbabe)
 for i,v in enumerate(updates):wr(v+1440,c['status'] if i==0 else 0x4000);wr(v+484,defs[0 if i==0 else 2])
 for v,bits in zip(defs,[c['run_multiplier'],0x40500000,0x3f400000]):wr(v+96,bits)
 before=bytes(u.mem_read(heap,len(seed)));wr(cw,a.control);u.emu_start(init,init+8,count=2);u.mem_write(sp-512,bytes(1024));wr(sp,stop);wr(sp+4,unit)
 for reg,value in registers.items():u.reg_write(reg,value)
 u.reg_write(UC_X86_REG_ESP,sp);u.reg_write(UC_X86_REG_EFLAGS,0x202);active=True
 try:u.emu_start(0x545020 if c['entry']=='root' else 0x545090,stop+1,count=10000)
 except Exception as e:raise AssertionError((c['name'],hex(u.reg_read(UC_X86_REG_EIP)))) from e
 active=False
 assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4,c['name']
 assert all(u.reg_read(reg)==value for reg,value in registers.items()),c['name']
 assert u.reg_read(UC_X86_REG_FPCW)==a.control and u.reg_read(UC_X86_REG_FPSW)&0x3800==0,c['name']
 assert bytes(u.mem_read(sp-512,256))==bytes(256) and bytes(u.mem_read(sp+8,256))==bytes(256),c['name']
 after=bytes(u.mem_read(heap,len(seed)));assert all(x==y or heap+i in allowed for i,(x,y) in enumerate(zip(before,after))),c['name']
 eax=u.reg_read(UC_X86_REG_EAX)
 if c['entry']=='direction':assert eax<=255
 return dict(input=c,result=dict(events=copy.deepcopy(events),state=state()),raw_original_eax=eax,public_byte=eax&255)
cases=[]
def case(prefix,**kw):
 c=dict(entry='root',direction=13,subclass=0x12340080,position=[0x43a08000,0x44238000],speed=0x40200000,run_multiplier=0x3fe00000,status=0,random=0,tile=0,audio_return=0x87654321,mutation=dict(stage='',action='',value=0))
 c.update(kw,name=f'{prefix}-{len(cases)}');cases.append(c)
for direction,random,route in itertools.product(range(256),[-20,0,20],range(3)):
 case('root-direction-grid',direction=direction,random=random,tile=6 if route==1 else 0,subclass=0x12340480 if route==2 else 0x12340080,status=0x4000 if direction&1 else 0)
for direction,random,route in itertools.product([256,32767,32768,65535],[-20,0,20],range(3)):
 case('root-word-wrap',direction=direction,random=random,tile=6 if route==1 else 0,subclass=0x480 if route==2 else 0x80)
for stage,action,value in itertools.product(['random','tile','move_audio'],['direction','subclass','class','update','entry-status','alternate-status','entry-definition','alternate-definition','run-multiplier','speed','position','all'],range(2)):
 case('root-live',status=0x4000,mutation=dict(stage=stage,action=action,value=value))
for speed,multiplier,direction in itertools.product([0,0x80000000,1,0x00800000,0x3f000001,0x7f7fffff,0x7f800000,0xff800000,0x7fc12345,0x7fa12345],[0x3f000001,0x3fe00000,0x7f7fffff,0x7f800000],[1,64]):
 case('root-force-bits',speed=speed,run_multiplier=multiplier,direction=direction,status=0x4000,subclass=0x480)
for pos,direction in itertools.product([0,0x80000000,1,0x00800000,0x42c80000,0x4b7fffff,0x7f7fffff,0x7f800000,0xff800000,0x7fc12345,0x7fa12345], [1,14,64,127]):
 case('root-probe-bits',position=[pos,pos^0x80000000],direction=direction)
for ret in [0,1,0x7f,0x80,0xff,0x80000000,0xffffffff]:case('root-char-forward',audio_return=ret,subclass=0x480)
for direction,random,route in itertools.product([0,1,63,64,127,128,192,255,256,32767,32768,65535],[-20,0,20],range(3)):
 case('helper-int',entry='direction',direction=direction,random=random,tile=6 if route==1 else 0,subclass=0x480 if route==2 else 0x80)
rows=[run(c) for c in cases]
executed={pc+i for pc,b in traces.items() for i in range(len(b))};assert executed=={pc+i for pc,n in ranges.items() for i in range(n)}
assert all(bytes(u.mem_read(pc,n))==original[pc] for pc,n in ranges.items())
table=bytes(u.mem_read(0x5b6658,2048));constant=bytes(u.mem_read(0x583c08,4))
fixture=dict(original_sha256=sha(blob),control_word=a.control,rows=rows)
proof=dict(original_sha256=sha(blob),control_word=a.control,contexts=len(rows),root_contexts=sum(c['entry']=='root' for c in cases),helper_contexts=sum(c['entry']=='direction' for c in cases),faults=0,root_bytes=ranges,executed_bytes=len(executed),root_sha256={hex(pc):sha(b) for pc,b in original.items()},services=dict(services),register_stack_FPCW_TOP_heap_canaries_preserved=True,original_code_unmodified=True,RNG_source_line435_and_original_path_pointer_preserved=True,raw_root_EAX_and_public_byte_retained=True,Go_handled_bool_not_substituted_for_public_byte=True,helper_public_int0_255=True,unexecuted_Nops=[11,5],direction_table=dict(address='0x5b6658',size=2048,sha256=sha(table)),probe_constant=dict(address='0x583c08',size=4,bytes=constant.hex(),sha256=sha(constant)),instructions=[dict(pc=hex(pc),bytes=b.hex()) for pc,b in sorted(traces.items())])
for path,value in [(a.fixture,fixture),(a.proof,proof)]:
 with path.open('x') as f:json.dump(value,f,indent=2);f.write('\n')
print(json.dumps({k:proof[k] for k in ['contexts','root_contexts','helper_contexts','executed_bytes','control_word','services']}))
