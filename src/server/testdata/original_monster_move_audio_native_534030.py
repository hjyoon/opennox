"""Execute unchanged 230-byte MoveAudio and both 31-byte getters.

Only AudioEvent is a declared body. Animation count/duration and sound-set
lookup execute unchanged. Primary rows use valid frame divisors and native
record inputs. Separate malformed-divisor diagnostics retain original CPU
faults, and an ancillary nil sound getter covers its early return.
"""
import argparse,collections,gzip,hashlib,itertools,json,struct
from pathlib import Path
from unicorn import Uc,UcError,UC_ARCH_X86,UC_MODE_32,UC_HOOK_CODE
from unicorn.x86_const import *

p=argparse.ArgumentParser(description=__doc__)
for name in ['game','fixture','proof']:p.add_argument('--'+name,type=Path,required=True)
p.add_argument('--control',type=lambda s:int(s,0),choices=[0xe3f,0x23f],required=True)
a=p.parse_args();blob=a.game.read_bytes();sha=lambda data:hashlib.sha256(data).hexdigest()
assert sha(blob)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,60)[0];op=pe+24;tb=op+struct.unpack_from('<H',blob,pe+20)[0];base=struct.unpack_from('<I',blob,op+28)[0]
u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,(struct.unpack_from('<I',blob,op+56)[0]+4095)&~4095)
for i in range(struct.unpack_from('<H',blob,pe+6)[0]):
 _,va,n,off=struct.unpack_from('<IIII',blob,tb+i*40+8)
 if n:u.mem_write(base+va,blob[off:off+n])
heap,stack,code=0x21000000,0x22000000,0x23000000
for address in [heap,stack,code]:u.mem_map(address,0x10000)
unit=heap+0x1000;updates=[heap+0x3000,heap+0x5000];defs=[heap+0x7000,heap+0x8000];sounds=[heap+0xa000,heap+0xb000]
sp,stop,init,cw=stack+0xe000,code+0x100,code+0x200,code+0x300
rd=lambda address:struct.unpack('<I',u.mem_read(address,4))[0]
wr=lambda address,value:u.mem_write(address,struct.pack('<I',value&0xffffffff))
r8=lambda address:bytes(u.mem_read(address,1))[0]
w8=lambda address,value:u.mem_write(address,bytes([value&255]))
ptrref=lambda pointer,values:values.index(pointer) if pointer else -1
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw))
ranges={0x534030:230,0x424300:31,0x4f9f90:31}
original={pc:bytes(u.mem_read(pc,n)) for pc,n in ranges.items()}
seed=bytes((i*31+19)&255 for i in range(0x10000))
registers={UC_X86_REG_EBX:0x12345678,UC_X86_REG_ESI:0x23456789,UC_X86_REG_EDI:0x3456789a,UC_X86_REG_EBP:0x456789ab}
active=False;current={};events=[];traces={};service_count=0

def state():
 return dict(cls=rd(unit+8),subclass=rd(unit+12),net=rd(unit+36),
  update=updates.index(rd(unit+748)),definitions=[defs.index(rd(v+484)) for v in updates],
  sound_sets=[ptrref(rd(v+488),sounds) for v in updates],
  animation=[[r8(v+481),r8(v+482)] for v in updates],
  steps=[[rd(v+100),rd(v+104)] for v in defs],
  sound_ids=[rd(v+72) for v in sounds],frame=rd(0x84ea04),
  frames=rd(0x753450),delay=rd(0x753454))

def hook(_u,pc,size,_):
 global service_count
 if not active:return
 if pc==stop:u.emu_stop();return
 if any(start<=pc<start+n for start,n in ranges.items()):
  traces[pc]=bytes(u.mem_read(pc,size));return
 assert pc==0x501960,(current['name'],hex(pc))
 esp=u.reg_read(UC_X86_REG_ESP)
 assert rd(esp+8)==unit and rd(esp+12)==0 and rd(esp+16)==0
 events.append(dict(kind='audio',args=[rd(esp+4),0,0,0],state=state()))
 service_count+=1
 u.reg_write(UC_X86_REG_EAX,current['audio_return'])
 u.reg_write(UC_X86_REG_EIP,rd(esp));u.reg_write(UC_X86_REG_ESP,esp+4)

u.hook_add(UC_HOOK_CODE,hook)
def run(c,diagnostic=False,nil_getter=False):
 global active,current,events
 active=False;current=c;events=[];u.mem_write(heap,seed)
 wr(unit+8,c['cls']);wr(unit+12,c['subclass']);wr(unit+36,c['net']);wr(unit+748,updates[c['update']])
 for i,v in enumerate(updates):
  wr(v+484,defs[c['definition']] if i==c['update'] else defs[1-c['definition']])
  wr(v+488,0 if c['sound_set']<0 else sounds[c['sound_set'] if i==c['update'] else 1-c['sound_set']])
  w8(v+481,c['animation'][0] if i==c['update'] else 255)
  w8(v+482,c['animation'][1] if i==c['update'] else 1)
 for i,v in enumerate(defs):
  wr(v+100,c['steps'][0] if i==c['definition'] else 2)
  wr(v+104,c['steps'][1] if i==c['definition'] else 255)
 for v,sound_id in zip(sounds,c['sound_ids']):wr(v+72,sound_id)
 wr(0x84ea04,c['frame']);wr(0x753450,c['frames']);wr(0x753454,c['delay'])
 before=bytes(u.mem_read(heap,len(seed)));wr(cw,a.control);u.emu_start(init,init+8,count=2)
 u.mem_write(sp-512,bytes(1024));wr(sp,stop);wr(sp+4,0 if nil_getter else unit)
 for reg,value in registers.items():u.reg_write(reg,value)
 u.reg_write(UC_X86_REG_EAX,0xa5a5a5a5);u.reg_write(UC_X86_REG_ESP,sp);u.reg_write(UC_X86_REG_EFLAGS,0x202)
 active=True;fault=None
 try:u.emu_start(0x424300 if nil_getter else 0x534030,stop+1,count=1000)
 except UcError as error:
  assert diagnostic,(c['name'],hex(u.reg_read(UC_X86_REG_EIP)),str(error))
  fault=dict(pc=hex(u.reg_read(UC_X86_REG_EIP)),error=str(error))
 active=False
 if not fault:
  assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4,c['name']
  assert all(u.reg_read(reg)==value for reg,value in registers.items()),c['name']
 assert u.reg_read(UC_X86_REG_FPCW)==a.control and u.reg_read(UC_X86_REG_FPSW)&0x3800==0,c['name']
 assert bytes(u.mem_read(sp-512,256))==bytes(256) and bytes(u.mem_read(sp+8,256))==bytes(256),c['name']
 assert bytes(u.mem_read(heap,len(seed)))==before,c['name']
 eax=u.reg_read(UC_X86_REG_EAX)
 return dict(input=c,result=dict(events=list(events),state=state()),raw_original_EAX=eax,raw_original_AL=eax&255,
  caller_argument_word_after_return=rd(sp+4),fault=fault)

cases=[]
def case(prefix,**kw):
 c=dict(cls=2,subclass=0,net=0,frame=0,frames=7,delay=0,update=0,definition=0,
        sound_set=0,animation=[0,0],steps=[0,2],sound_ids=[123,987],audio_return=0x87654321)
 c.update(kw,name=prefix+'-'+str(len(cases)));cases.append(c)
for cls,frame,latch,steps,sound_set in itertools.product([0,2,4,6,0x1000002],[0,1,2,255],[0,1,255],
 [[0,2],[1,255],[255,255],[0xffffffff,1]],[-1,0,1]):
 case('monster-steps',cls=cls,animation=[frame,latch],steps=steps,sound_set=sound_set,
      update=frame&1,definition=latch&1)
for cls,frames,delay,net,frame in itertools.product([2,4,6],
 [1,2,3,7,255,0x80000000,0xffffffff],[0,1,2,0x7fffffff,0x80000000,0xfffffffe],
 [0,1,0x7fffffff,0x80000000,0xfffffff0,0xffffffff],[0,1,2,16,0xfffffffe,0xffffffff]):
 case('NPC-DWORD-animation',cls=cls,subclass=0x30,frames=frames,delay=delay,net=net,frame=frame,
      steps=[0,1],update=net&1,definition=frame&1)
for subclass,cls in itertools.product(range(256),[0,2]):
 case('subclass-byte-mask',cls=cls,subclass=subclass,frame=7,net=0xfffffffd,frames=7,delay=1)
for frames,delay,net,frame in itertools.product([1,3,7],[0,1],[0,1,0xfffffffd],[0,1,0xfffffffd]):
 for which in [0,1]:
  value=((net+frame)&0xffffffff)//(delay+1)%frames
  pair=[value,0xfffffffe] if which==0 else [0xfffffffe,value]
  case('A-B-frame-selection',subclass=0x10,frames=frames,delay=delay,net=net,frame=frame,steps=pair,
       sound_set=which,sound_ids=[0,987])
rows=[run(c) for c in cases]
ancillary_input=dict(cases[0],name='nil-sound-getter-ancillary')
ancillary=run(ancillary_input,nil_getter=True)
assert ancillary['raw_original_EAX']==0 and not ancillary['result']['events']
# Unicorn CPU exception delivery state survives a failed emu_start. Each
# deliberate DIV diagnostic gets its own CPU, while retaining every input.
template_size=(struct.unpack_from('<I',blob,op+56)[0]+4095)&~4095
template_image=bytes(u.mem_read(base,template_size))
assert all(bytes(u.mem_read(pc,len(raw)))==raw for pc,raw in original.items())
diagnostics=[]
for frames,delay in [(0,0),(7,0xffffffff),(0,0xffffffff)]:
 u=Uc(UC_ARCH_X86,UC_MODE_32)
 u.mem_map(base,template_size);u.mem_write(base,template_image)
 for region in [heap,stack,code]:u.mem_map(region,0x10000)
 u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw))
 u.hook_add(UC_HOOK_CODE,hook)
 c=dict(cases[0],name=f'invalid-divisor-{frames}-{delay}',cls=2,subclass=0x10,frames=frames,delay=delay)
 diagnostics.append(run(c,diagnostic=True))
assert all(x['fault'] for x in diagnostics)
visited={pc+i for pc,raw in traces.items() for i in range(len(raw))}
expected={pc+i for pc,n in ranges.items() for i in range(n)}
assert visited==expected,[hex(pc) for pc in sorted(expected-visited)]
assert all(bytes(u.mem_read(pc,len(raw)))==raw for pc,raw in original.items())
fixture=dict(original_sha256=sha(blob),control_word=a.control,rows=rows)
proof=dict(original_sha256=sha(blob),control_word=a.control,primary_contexts=len(rows),
 ancillary_contexts=1,malformed_divisor_diagnostic_contexts=3,root_bytes={hex(pc):n for pc,n in ranges.items()},
 executed_instruction_bytes=len(visited),root_sha256={hex(pc):sha(raw) for pc,raw in original.items()},
 only_declared_service='AudioEvent501960',declared_audio_calls=service_count,
 actual_animation_and_sound_getter_bodies_execute=True,original_code_unmodified=True,
 register_stack_FPCW_TOP_and_whole_heap_canaries_preserved=True,
 caller_unit_argument_stack_word_overwritten_by_actual_animation_getter_and_then_discarded=True,
 raw_original_EAX_AL_retained_and_return_consumers_proven_separately=True,
 runtime_inputs=dict(frame_global='0x84ea04',animation_count_and_delay=['0x753450','0x753454']),
 invalid_divisor_diagnostics_are_not_primary_comparator_inputs=True,
 ancillary_nil_getter=ancillary,malformed_divisor_diagnostics=diagnostics,
 unexecuted_Nop_tails=[10,1,1],
 instructions=[dict(pc=hex(pc),bytes=raw.hex()) for pc,raw in sorted(traces.items())])
for path,value in [(a.fixture,fixture),(a.proof,proof)]:
 with path.open('xb') as f:
  if path.suffix=='.gz':
   with gzip.GzipFile(filename='',mode='wb',fileobj=f,compresslevel=9,mtime=0) as stream:
    stream.write(json.dumps(value,separators=(',',':')).encode()+b'\n')
  else:f.write((json.dumps(value,indent=2)+'\n').encode())
print(json.dumps({k:v for k,v in proof.items() if k not in ['instructions','ancillary_nil_getter','malformed_divisor_diagnostics']}))
