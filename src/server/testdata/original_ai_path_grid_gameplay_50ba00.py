"""Execute the unchanged GAME.EXE 0050BA00 coordinate prefix and 00419A70.
This observes binary32 product spills and FISTP under the actual gameplay CW;
no wall precheck, path search, map state or whole-game behavior is emulated.
"""
import hashlib,json,pathlib,struct,sys
from collections import Counter
from unicorn import Uc,UC_ARCH_X86,UC_MODE_32,UC_HOOK_CODE
from unicorn.x86_const import UC_X86_REG_ESP,UC_X86_REG_EAX,UC_X86_REG_FPCW,UC_X86_REG_FPSW,UC_X86_REG_FPTAG
repo,original=map(pathlib.Path,sys.argv[1:3]);blob=original.read_bytes();manifest=json.loads((repo/'toolchain/oracle/game-exe-functions.json').read_text())
assert hashlib.sha256(blob).hexdigest()==manifest['file_sha256']=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,0x3c)[0];opt=pe+24;base=struct.unpack_from('<I',blob,opt+28)[0];assert base==0x400000
st=opt+struct.unpack_from('<H',blob,pe+20)[0];sections=[struct.unpack_from('<IIII',blob,st+i*40+8) for i in range(struct.unpack_from('<H',blob,pe+6)[0])]
def read(a,n):
 for _,v,z,r in sections:
  if v<=a-base and a-base+n<=v+z:return blob[r+a-base-v:r+a-base-v+n]
 raise AssertionError(hex(a))
addresses={0x50ba00,0x419a70,0x43e2c1,0x4031f2};ranges=[f for f in manifest['functions'] if int(f['address'],16) in addresses];assert len(ranges)==4
for f in ranges:assert hashlib.sha256(read(int(f['address'],16),f['size'])).hexdigest()==f['sha256']
assert read(0x583a00,4)==bytes.fromhex('4316323d')
cpu=Uc(UC_ARCH_X86,UC_MODE_32);cpu.mem_map(base,(struct.unpack_from('<I',blob,opt+56)[0]+4095)&~4095)
for _,v,z,r in sections:
 if z:cpu.mem_write(base+v,blob[r:r+z])
heap,stack=0x10000000,0x20000000;cpu.mem_map(heap,0x1000);cpu.mem_map(stack,0x10000)
entrysp=stack+0x8000;hits=Counter()
def hook(c,a,z,_):
 hits[hex(a)]+=1
 if a==0x50ba7e:c.emu_stop();return
 if 0x50ba00<=a<0x50ba7e or 0x419a70<=a<0x419a81:return
 raise AssertionError('unexpected instruction '+hex(a))
cpu.hook_add(UC_HOOK_CODE,hook)
def bits(x):return struct.unpack('<I',struct.pack('<f',x))[0]
inputs=[]
def add(name,x,y):inputs.append((name,bits(x),bits(y)))
add('zaris-spawn',3917,4892);add('zaris-outbound-waypoint113',3969,4843);add('zaris-outbound-waypoint116',4020,4820);add('zaris-gate-waypoint118',4083,4843)
add('positive-halves',34.5,57.5);add('negative-halves',-34.5,-57.5);add('positive-3.5',80.5,11.5)
for n in range(257):
 b=bits(n*23.0)
 for step in [-1,0,1]:
  if b+step>=0:inputs.append((f'grid-boundary-{n}-ulp{step:+d}',b+step,(b+step)^0x80000000))
for b in [0,0x80000000,1,0x80000001,0x7f800000,0xff800000,0x7fc00000,0x7f7fffff,0xff7fffff,0x51380000,0xd1380000]:inputs.append((f'special-{b:08x}',b,b))
rows=[]
for name,x,y in inputs:
 observed={}
 for mode,cw in [('nearest',0x023f),('gameplay',0x0e3f)]:
  cpu.mem_write(heap,struct.pack('<II',x,y));cpu.mem_write(entrysp,struct.pack('<7I',0,0,heap+0x100,heap,heap,0,0));cpu.mem_write(0x81ae8c,struct.pack('<I',1))
  cpu.reg_write(UC_X86_REG_ESP,entrysp);cpu.reg_write(UC_X86_REG_FPCW,cw);cpu.reg_write(UC_X86_REG_FPSW,0);cpu.reg_write(UC_X86_REG_FPTAG,0xffff)
  cpu.emu_start(0x50ba00,0x50ba7e,count=200)
  sp=cpu.reg_read(UC_X86_REG_ESP);assert sp==entrysp-108
  observed[mode]=list(struct.unpack('<ii',cpu.mem_read(sp+0x54,8)))
  assert cpu.reg_read(UC_X86_REG_FPCW)==cw
 rows.append(dict(name=name,input_bits=[x,y],**observed))
out={'original_sha256':hashlib.sha256(blob).hexdigest(),'ranges':ranges,'entry':'0x0050BA00','stop_before':'0x0050BA7E','inverse_bits':0x3d321643,'gameplay_control_word':0x0e3f,'nearest_control_word':0x023f,'scope':'Unchanged original source-coordinate prefix and FISTP. Main-loop/CRT CW provenance sealed; wall precheck and full path search are outside this bounded observation.','rows':rows,'instruction_entries':dict(hits)}
json.dump(out,sys.stdout,separators=(',',':'));print()
