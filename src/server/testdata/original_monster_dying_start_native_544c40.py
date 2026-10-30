"""Execute85 original DyingStart bytes and the real31-byte sound getter.

Audio, ScriptCallback and the five selected DIE_FUNCTION bodies are declared
services. The sound getter runs unchanged, including standalone nil/class
controls. Original int EAX and native sound-pointer identity stay recorded.
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
unit=heap+0x1000;updates=[heap+0x3000,heap+0x6000];definitions=[heap+0x9000,heap+0xa000,heap+0xb000];sounds=[heap+0xc000,heap+0xd000]
callbacks=[0,0x54a7d0,0x54a850,0x54a890,0x54a900,0x54a950]
sp,stop,init,cw=stack+0xe000,code+0x100,code+0x200,code+0x300
read32=lambda addr:struct.unpack('<I',u.mem_read(addr,4))[0]
put32=lambda addr,v:u.mem_write(addr,struct.pack('<I',v&0xffffffff))
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw))
seed=bytes((i*31+19)&255 for i in range(0x10000))
registers={UC_X86_REG_EBX:0x12345678,UC_X86_REG_ESI:0x23456789,UC_X86_REG_EDI:0x3456789a,UC_X86_REG_EBP:0x456789ab}
ranges={0x544c40:85,0x424300:31};original={pc:bytes(u.mem_read(pc,n)) for pc,n in ranges.items()}
active=False;current={};events=[];traces={};services=collections.Counter()
def sound_index(pointer):return -1 if pointer==0 else sounds.index(pointer)
def state():
 return dict(update=updates.index(read32(unit+748)),cls=read32(unit+8),subclass=read32(unit+12),flags=read32(unit+16),position=list(struct.unpack('<2f',u.mem_read(unit+56,8))),
  updates=[dict(definition=definitions.index(read32(v+484)),sound=sound_index(read32(v+488)),script_flags=read32(v+1264),script_func=read32(v+1268)) for v in updates],
  callbacks=[callbacks.index(read32(v+228)) for v in definitions],sounds=[[read32(v+i*4) for i in range(16)] for v in sounds])
def mutate(stage):
 m=current['mutation']
 if m['stage']!=stage:return
 action,value=m['action'],m['value'];entry=updates[current['update']];alternate=updates[1-current['update']]
 assert action in ['update','entry-definition','alternate-definition','entry-function','function-A','function-B','function-C','entry-sound','alternate-sound','sound-id','script','class','flags','position','all']
 if action in ['update','all']:put32(unit+748,alternate)
 if action in ['entry-definition','all']:put32(entry+484,definitions[1])
 if action in ['alternate-definition','all']:put32(alternate+484,definitions[1])
 if action=='entry-function':put32(read32(entry+484)+228,callbacks[value])
 if action=='function-A':put32(definitions[0]+228,callbacks[value])
 if action in ['function-B','all']:put32(definitions[1]+228,callbacks[value])
 if action=='function-C':put32(definitions[2]+228,callbacks[value])
 if action in ['entry-sound','all']:put32(entry+488,sounds[1])
 if action in ['alternate-sound','all']:put32(alternate+488,sounds[0])
 if action in ['sound-id','all']:put32(sounds[0]+60,0xfedc0000|value)
 if action in ['script','all']:put32(entry+1264,0x5a);put32(entry+1268,27)
 if action in ['class','all']:put32(unit+8,0x400000);put32(unit+12,0xfedcba98)
 if action in ['flags','all']:put32(unit+16,0x87654321)
 if action in ['position','all']:u.mem_write(unit+56,struct.pack('<2f',-37.5,19.25))
def hook(_u,pc,size,_):
 if not active:return
 if pc==stop:u.emu_stop();return
 if any(start<=pc<start+n for start,n in ranges.items()):traces[pc]=bytes(u.mem_read(pc,size));return
 esp=u.reg_read(UC_X86_REG_ESP);kind={0x501960:'audio',0x502490:'script'}.get(pc)
 callback=callbacks.index(pc) if pc in callbacks[1:] else 0
 if callback:kind='callback'
 assert kind,(current['name'],hex(pc))
 args=[]
 if kind=='audio':
  assert read32(esp+8)==unit and read32(esp+12)==read32(esp+16)==0
  args=[read32(esp+4),0,0,0]
 elif kind=='script':
  assert read32(esp+4)==updates[current['update']]+1264 and read32(esp+8)==0 and read32(esp+12)==unit
  args=[current['update'],0,0]
 else:assert read32(esp+4)==unit;args=[callback,0]
 events.append(dict(kind=kind,args=args,state=state()));services[kind]+=1;mutate(kind)
 u.reg_write(UC_X86_REG_EAX,current['callback_return'] if callback else 0x87654321)
 u.reg_write(UC_X86_REG_EIP,read32(esp));u.reg_write(UC_X86_REG_ESP,esp+4)
u.hook_add(UC_HOOK_CODE,hook)
allowed=set()
for ptr,off,n in [(unit,8,12),(unit,56,8),(unit,748,4)]+[(v,o,n) for v in updates for o,n in [(484,8),(1264,8)]]+[(v,228,4) for v in definitions]+[(sounds[0],60,4)]:allowed.update(range(ptr+off,ptr+off+n))
def run(c):
 global active,current,events
 active,current,events=False,c,[];u.mem_write(heap,seed)
 put32(unit+8,c['cls']);put32(unit+12,0xa5);put32(unit+16,0x12000004);put32(unit+748,updates[c['update']])
 u.mem_write(unit+56,struct.pack('<2f',321.,654.))
 for n,v in enumerate(updates):
  entry=n==c['update'];put32(v+484,definitions[0 if entry else 2]);index=c['sound'] if entry else 1
  put32(v+488,0 if index<0 else sounds[index]);put32(v+1264,0xa5 if entry else 0x5a);put32(v+1268,17 if entry else 19)
 for n,v in enumerate(definitions):put32(v+228,callbacks[c['callback'] if n==0 else n+1])
 for n,v in enumerate(sounds):
  for i in range(16):put32(v+i*4,715+n if i==15 else 0xa5000000+n*0x100+i)
 before=bytes(u.mem_read(heap,len(seed)))
 u.mem_write(cw,struct.pack('<H',a.control));u.emu_start(init,init+8,count=2)
 u.mem_write(sp-512,bytes(1024));put32(sp,stop);put32(sp+4,0 if c['nil_unit'] else unit)
 for reg,v in registers.items():u.reg_write(reg,v)
 u.reg_write(UC_X86_REG_EFLAGS,0x202);u.reg_write(UC_X86_REG_ESP,sp);active=True
 try:u.emu_start(0x544c40 if c['entry']=='dying' else 0x424300,stop+1,count=10000)
 except Exception as e:raise AssertionError((c['name'],hex(u.reg_read(UC_X86_REG_EIP)))) from e
 active=False
 assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4,c['name']
 assert all(u.reg_read(reg)==v for reg,v in registers.items()),c['name']
 assert u.reg_read(UC_X86_REG_FPCW)==a.control and u.reg_read(UC_X86_REG_FPSW)&0x3800==0,c['name']
 assert bytes(u.mem_read(sp-512,256))==bytes(256) and bytes(u.mem_read(sp+8,256))==bytes(256),c['name']
 after=bytes(u.mem_read(heap,len(seed)));assert all(x==y or heap+i in allowed for i,(x,y) in enumerate(zip(before,after))),c['name']
 eax=u.reg_read(UC_X86_REG_EAX)
 if c['entry']=='dying':assert eax==(c['callback_return'] if any(e['kind']=='callback' for e in events) else 0),c['name']
 else:assert not events and before==after,c['name']
 return dict(input=c,result=dict(events=copy.deepcopy(events),state=state()),raw_original_eax=eax,getter_return_sound=sound_index(eax) if c['entry']=='getter' else -2)
cases=[]
def case(prefix,**kw):
 c=dict(entry='dying',update=0,cls=2,nil_unit=False,sound=0,callback=1,callback_return=1,mutation=dict(stage='',action='',value=0))
 c.update(kw,name=f'{prefix}-{len(cases)}');cases.append(c)
for sound,callback,update,ret in itertools.product([-1,0,1],range(6),range(2),[0,1,0x7fffffff,0x80000000,0xffffffff]):case('selected-callback',sound=sound,callback=callback,update=update,callback_return=ret)
actions=['update','entry-definition','alternate-definition','entry-function','function-A','function-B','function-C','entry-sound','alternate-sound','sound-id','script','class','flags','position','all']
for stage,action,update,value in itertools.product(['audio','script','callback'],actions,range(2),[0,2,5]):case('live-mutation',update=update,mutation=dict(stage=stage,action=action,value=value))
case('getter-nil',entry='getter',nil_unit=True)
for cls,sound,update in itertools.product([0,1,2,3,4,6,0x400000,0xffffffff],[-1,0,1],range(2)):case('getter-class-pointer',entry='getter',cls=cls,sound=sound,update=update)
assert len(cases)==499
rows=[run(c) for c in cases]
executed={pc+j for pc,b in traces.items() for j in range(len(b))};assert executed=={pc+j for pc,n in ranges.items() for j in range(n)}
assert all(bytes(u.mem_read(pc,n))==original[pc] for pc,n in ranges.items())
fixture=dict(original_sha256=sha(blob),control_word=a.control,rows=rows)
proof=dict(original_sha256=sha(blob),control_word=a.control,contexts=len(cases),dying_contexts=450,getter_contexts=49,faults=0,root_bytes=ranges,executed_bytes=len(executed),
 original_root_sha256={hex(pc):sha(b) for pc,b in original.items()},services=dict(services),only_Audio_Script_and_five_DIE_FUNCTION_bodies_declared=True,
 real_original_sound_getter_runs_unmodified=True,raw_original_EAX_and_sound_pointer_identity_retained=True,register_stack_FPCW_TOP_heap_canaries_preserved=True,
 eleven_dying_root_NOPs_and_adjacent_getter_setter_entries_unexecuted=True,existing_manifest_envelopes={'544C40':96,'424300':31},
 instructions=[dict(pc=hex(pc),bytes=b.hex()) for pc,b in sorted(traces.items())])
for path,data in [(a.fixture,fixture),(a.proof,proof)]:
 with path.open('x') as f:json.dump(data,f,indent=2);f.write('\n')
print(json.dumps(dict(contexts=len(cases),dying=450,getter=49,executed_bytes=len(executed),control_word=hex(a.control),services=dict(services),fixture_sha256=sha(a.fixture.read_bytes()))))
