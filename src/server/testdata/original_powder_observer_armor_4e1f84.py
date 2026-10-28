"""Execute the original armor/carry prefix with declared armor-wear service.

This seeds the prefix registers/locals; it does not execute the full damage
entry or manufacture a gameplay outcome. Nearest control remains separate.
"""
import argparse,collections,hashlib,itertools,json,struct
from pathlib import Path
from unicorn import Uc,UC_ARCH_X86,UC_MODE_32,UC_HOOK_CODE
from unicorn.x86_const import *

p=argparse.ArgumentParser(description=__doc__)
for n in ('game','fixture','proof'):p.add_argument('--'+n,type=Path,required=True)
a=p.parse_args();blob=a.game.read_bytes();sha=lambda b:hashlib.sha256(b).hexdigest()
assert sha(blob)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,60)[0];opt=pe+24;base=struct.unpack_from('<I',blob,opt+28)[0];size=(struct.unpack_from('<I',blob,opt+56)[0]+4095)&~4095;tab=opt+struct.unpack_from('<H',blob,pe+20)[0]
u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,size)
for n in range(struct.unpack_from('<H',blob,pe+6)[0]):
 _,va,count,off=struct.unpack_from('<IIII',blob,tab+40*n+8)
 if count:u.mem_write(base+va,blob[off:off+count])
stack,heap=0x20000000,0x21000000
for v in (stack,heap):u.mem_map(v,0x10000)
sp=stack+0x8000;unit=heap+128;data=heap+0x2000
get=lambda v:struct.unpack('<I',u.mem_read(v,4))[0]
put=lambda v,x:u.mem_write(v,struct.pack('<I',x&0xffffffff))
f32=lambda x:struct.unpack('<I',struct.pack('<f',x))[0]
specs=[('armor-prefix',0x4e1f84,133),('minimum',0x4e2011,20),('carry',0x4e20f0,131),('dword',0x419a70,17)]
regions=[dict(name=n,address=hex(v),size=z,sha256=sha(bytes(u.mem_read(v,z)))) for n,v,z in specs]
trace={};calls=collections.Counter();recording=False
def hook(_u,address,length,_):
 if not recording:return
 for n,v,z in specs:
  if v<=address<v+z:
   trace[address]=bytes(u.mem_read(address,length)).hex()
   if address==v:calls[n]+=1
   return
 if address==0x4e2180:
  esp=u.reg_read(UC_X86_REG_ESP);assert get(esp+4)==unit and get(esp+8)==get(esp+12)==0 and get(esp+20)==7
  u.reg_write(UC_X86_REG_EAX,0x12345678);u.reg_write(UC_X86_REG_EIP,get(esp));u.reg_write(UC_X86_REG_ESP,esp+4)
 else:raise AssertionError(hex(address))
u.hook_add(UC_HOOK_CODE,hook);u.reg_write(UC_X86_REG_ESP,sp);u.reg_write(UC_X86_REG_FPCW,0x27f);u.emu_start(0x4014b5,0x4014c4,count=20000)
assert u.reg_read(UC_X86_REG_FPCW)==0xe3f
initial=u.context_save();recording=True;rows=[]
for cls,raw,armor,carry in itertools.product((4,2),(-9,-1,0,1,9,15),(f32(0),f32(.22999997),f32(.5),f32(1)),(f32(0),f32(.25),f32(-.25),f32(.4),0x3f7fffff)):
 for cw in (0xe3f,0x23f):
  u.context_restore(initial);u.reg_write(UC_X86_REG_FPCW,cw);u.mem_write(heap,b'\xa5'*0x10000);u.mem_write(stack,bytes(0x10000))
  put(unit+8,cls);put(unit+748,data);co=84 if cls==4 else 4;mo=304 if cls==4 else 2188;put(data+co,carry);put(data+mo,0);put(data+mo-4,77)
  put(sp+0x10,armor);put(sp+0x34,raw);put(sp+0x38,7);put(sp+0x30,0);put(sp+0x2c,0);put(sp+0x28,0x12345678)
  before=bytearray(u.mem_read(heap,0x10000))
  u.reg_write(UC_X86_REG_ESP,sp);u.reg_write(UC_X86_REG_ESI,unit);u.reg_write(UC_X86_REG_EDI,data);u.emu_start(0x4e1f84,0x4e2025,count=10000)
  assert u.reg_read(UC_X86_REG_EIP)==0x4e2025
  after=bytearray(u.mem_read(heap,0x10000))
  for off in (co,mo,mo-4):before[data-heap+off:data-heap+off+4]=after[data-heap+off:data-heap+off+4]
  assert before==after
  damage=struct.unpack('<i',u.mem_read(sp+0x28,4))[0]
  rows.append(dict(cls=cls,raw=raw,armor=armor,carry=carry,control_word=cw,effective=damage,remainder=get(data+co),marker=get(data+mo),typ=get(data+mo-4)))
assert len(rows)==480
assert sum(len(b)//2 for b in trace.values())==sum(z for _,_,z in specs)==301
for region,(_,v,z) in zip(regions,specs):assert sha(bytes(u.mem_read(v,z)))==region['sha256']
fixture=dict(game_sha256=sha(blob),regions=regions,scope=__doc__,rows=rows)
proof=dict(game_sha256=sha(blob),regions=regions,contexts=240,executions=480,original_code_bytes=301,arithmetic_or_conversion_substitutions=0,seeded_original_prefix_not_full_damage_entry=True,nonoutput_heap_and_canaries_preserved=True,entries=dict(calls),executed_instructions=[dict(address=hex(v),bytes=b) for v,b in sorted(trace.items())])
for path,value in ((a.fixture,fixture),(a.proof,proof)):
 with path.open('x') as out:json.dump(value,out,indent=2);out.write('\n')
print(json.dumps({k:v for k,v in proof.items() if k!='executed_instructions'}))
