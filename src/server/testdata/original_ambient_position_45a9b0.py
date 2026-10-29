"""Execute stock positional-loop geometry and handle lifecycle.

Whole45A9B0, actual452EB0/452E90 and CRT566DCC conversion execute original
bytes. Definition/viewport lookup, channel create/stop and volume/pan setters
are declared services. No geometry, FSQRT, int32 arithmetic or conversion is
replaced. This proves this root under those services, not the Miles engine.
"""
from pathlib import Path
import argparse, collections, copy, hashlib, json, struct
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE
from unicorn.x86_const import *
p=argparse.ArgumentParser(description=__doc__)
for key in ['game','fixture','proof']:p.add_argument('--'+key,type=Path,required=True)
p.add_argument('--control',type=lambda s:int(s,0),default=0xe3f)
a=p.parse_args();blob=a.game.read_bytes();sha=lambda b:hashlib.sha256(b).hexdigest()
assert sha(blob)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,60)[0];opt=pe+24;base=struct.unpack_from('<I',blob,opt+28)[0];table=opt+struct.unpack_from('<H',blob,pe+20)[0]
u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,(struct.unpack_from('<I',blob,opt+56)[0]+4095)&~4095)
for n in range(struct.unpack_from('<H',blob,pe+6)[0]):
 _,va,size,off=struct.unpack_from('<IIII',blob,table+40*n+8)
 if size:u.mem_write(base+va,blob[off:off+size])
heap,stack,code=0x21000000,0x22000000,0x23000000
for addr in [heap,stack,code]:u.mem_map(addr,0x10000)
source,listener,viewport,definition,channel=heap+0x1000,heap+0x2000,heap+0x3000,heap+0x4000,heap+0x5000
sp,stop,init,cw=stack+0xe000,code+0x100,code+0x200,code+0x300
get=lambda addr:struct.unpack('<I',u.mem_read(addr,4))[0]
put=lambda addr,v:u.mem_write(addr,struct.pack('<I',v&0xffffffff))
signed=lambda v:((v&0xffffffff)^0x80000000)-0x80000000
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw))
regions={'position':(0x45a9b0,0x45ab33),'handle':(0x452eb0,0x452ed8),'store':(0x452e90,0x452eae),'conversion':(0x566dcc,0x566df3)}
traces={k:{} for k in regions};entries=collections.Counter();services=collections.Counter();events=[];current={};active=False
def ret(v=0):
 esp=u.reg_read(UC_X86_REG_ESP);u.reg_write(UC_X86_REG_EAX,v&0xffffffff);u.reg_write(UC_X86_REG_EIP,get(esp));u.reg_write(UC_X86_REG_ESP,esp+4)
def hook(_u,pc,length,_):
 if not active:return
 if pc==stop:u.emu_stop();return
 for name,(lo,hi) in regions.items():
  if lo<=pc<hi:
   traces[name][pc]=bytes(u.mem_read(pc,length))
   if pc==lo:entries[name]+=1
   return
 esp=u.reg_read(UC_X86_REG_ESP);arg=lambda n:get(esp+4*n);services[hex(pc)]+=1
 if pc==0x452270:assert arg(1)==973;ret(definition if current['definition'] else 0)
 elif pc==0x437250:ret(viewport if current['viewport'] else 0)
 elif pc==0x4522a0:assert arg(1)==definition;ret(current['radius'])
 elif pc==0x452300:assert arg(1)==definition;events.append(['create']);ret(channel if current['create'] else 0)
 elif pc in [0x452ee0,0x452f80,0x452fe0,0x452f50]:
  assert arg(1)==channel;events.append([{0x452ee0:'initial-volume',0x452f80:'initial-pan',0x452fe0:'pan',0x452f50:'volume'}[pc],signed(arg(2))]);ret()
 elif pc==0x4523d0:assert arg(1)==channel;events.append(['stop']);ret()
 else:raise AssertionError(hex(pc))
u.hook_add(UC_HOOK_CODE,hook)
registers={UC_X86_REG_EBX:0x12345678,UC_X86_REG_ESI:0x23456789,UC_X86_REG_EDI:0x3456789a,UC_X86_REG_EBP:0x456789ab}
seed=bytes((i*31+19)&255 for i in range(0x10000))
def run(c):
 global current,events,active
 current,events,active=c,[],False;u.mem_write(heap,seed)
 for obj,pos in [(source,c['source']),(listener,c['listener'])]:put(obj+12,pos[0]);put(obj+16,pos[1])
 put(source+120,c['flags']);put(source+280,c['flags70']);put(source+492,973)
 put(source+496,channel if c['handle'] else 0);put(source+500,123 if c['handle']!=2 else 124);put(source+504,definition)
 put(channel+36,definition);put(channel+280,123)
 put(viewport,c['screen_x']);put(viewport+24,c['world_x']);put(0x9756cc,c['width'])
 before=bytes(u.mem_read(heap,len(seed)));width_before=bytes(u.mem_read(0x9756cc,4))
 u.mem_write(cw,struct.pack('<H',a.control));u.emu_start(init,init+8,count=2)
 u.mem_write(sp-512,bytes(1024));put(sp,stop);put(sp+4,source);put(sp+8,listener)
 for reg,v in registers.items():u.reg_write(reg,v)
 u.reg_write(UC_X86_REG_ESP,sp);active=True;u.emu_start(0x45a9b0,stop+1,count=10000);active=False
 assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4
 assert all(u.reg_read(reg)==v for reg,v in registers.items())
 assert u.reg_read(UC_X86_REG_FPCW)==a.control and u.reg_read(UC_X86_REG_FPSW)&0x3800==0
 after=bytes(u.mem_read(heap,len(seed)))
 assert all(x==y or source+496<=heap+i<source+508 for i,(x,y) in enumerate(zip(before,after)))
 assert bytes(u.mem_read(0x9756cc,4))==width_before and bytes(u.mem_read(sp-512,256))==bytes(256) and bytes(u.mem_read(sp+12,256))==bytes(256)
 result={'events':copy.deepcopy(events),'handle':bool(get(source+496))}
 return dict(input=c,result=result)
default=dict(name='default',source=[320,200],listener=[320,200],radius=600,width=640,screen_x=0,world_x=0,flags=0x1000000,flags70=0,handle=0,definition=True,viewport=True,create=True)
cases=[]
def case(name,**kw):c=copy.deepcopy(default);c.update(kw,name=name);cases.append(c)
case('default')
for handle in [0,1,2]:
 for x in [-2147483648,-65536,-46341,-600,-599,-320,-1,0,1,320,599,600,601,46341,65536,2147483647]:
  case(f'x-{handle}-{x}',source=[x,200],listener=[0,200],handle=handle)
 for y in [-601,-600,-599,-1,0,1,599,600,601]:case(f'y-{handle}-{y}',source=[0,y],listener=[0,0],handle=handle)
for radius in [-1,0,1,2,5,599,600,601,2147483647]:
 for d in [(0,0),(3,4),(5,5),(598,0)]:case(f'radius-{radius}-{d}',source=[d[0],d[1]],listener=[0,0],radius=radius)
for flags in [0,1,0xffffff,0x1000000,0x11000000,0xffffffff]:case(f'flags-{flags}',flags=flags,handle=1)
for flags in [0,1,2,4,8,12,16,0xffffffff]:case(f'flags70-{flags}',flags70=flags,handle=1)
for width in [2,3,5,640,641,1024,1920,-2,-3]:
 for x in [-320,0,320,639]:case(f'pan-{width}-{x}',source=[x,0],listener=[0,0],radius=1000,width=width,world_x=37,screen_x=-11)
for handle in [0,1,2]:
 case(f'no-definition-{handle}',definition=False,handle=handle)
 case(f'no-viewport-{handle}',viewport=False,handle=handle)
case('create-failed',create=False)
fixture=dict(control_word=a.control,cases=[run(c) for c in cases])
for path,value in [(a.fixture,fixture),(a.proof,dict(original_sha256=sha(blob),control_word=a.control,contexts=len(cases),entries=dict(entries),services=dict(services),regions=[dict(name=n,start=hex(lo),end_exclusive=hex(hi),bytes=hi-lo,sha256=sha(bytes(u.mem_read(lo,hi-lo))),covered_bytes=sum(len(b) for b in traces[n].values()),missing_bytes=[hex(i) for i in range(lo,hi) if not any(pc<=i<pc+len(b) for pc,b in traces[n].items())]) for n,(lo,hi) in regions.items()],register_stack_control_word_TOP_nonoutput_heap_global_canaries_preserved=True,faults=0,scope=__doc__))]:
 with path.open('x') as f:json.dump(value,f,indent=2);f.write('\n')
print(json.dumps(dict(contexts=len(cases),entries=dict(entries),coverage={n:sum(len(b) for b in t.values()) for n,t in traces.items()})))
