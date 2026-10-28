"""Complete unchanged Blow root/candidate, indexed direction and mass getter.
Only rectangle enumeration and CanInteract are declared world services.
"""
from pathlib import Path
import argparse,collections,copy,hashlib,json,math,struct
from unicorn import Uc,UC_ARCH_X86,UC_MODE_32,UC_HOOK_CODE,UC_HOOK_MEM_INVALID
from unicorn.x86_const import *

parser=argparse.ArgumentParser(description=__doc__)
parser.add_argument('--game',required=True,type=Path)
parser.add_argument('--fixture',required=True,type=Path)
parser.add_argument('--proof',required=True,type=Path)
args=parser.parse_args()
original=args.game
blob=original.read_bytes()
sha=lambda b:hashlib.sha256(b).hexdigest()
assert sha(blob)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,60)[0];opt=pe+24
base=struct.unpack_from('<I',blob,opt+28)[0]
size=(struct.unpack_from('<I',blob,opt+56)[0]+4095)&~4095
tab=opt+struct.unpack_from('<H',blob,pe+20)[0]
u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,size)
for i in range(struct.unpack_from('<H',blob,pe+6)[0]):
 _,va,count,offset=struct.unpack_from('<IIII',blob,tab+40*i+8)
 if count:u.mem_write(base+va,blob[offset:offset+count])
stack,heap,stop=0x20000000,0x21000000,0x30000000
for a in (stack,heap,stop):u.mem_map(a,0x10000)
sp=stack+0xe000
get=lambda a:struct.unpack('<I',u.mem_read(a,4))[0]
put=lambda a,v:u.mem_write(a,struct.pack('<I',v&0xffffffff))
bits=lambda v:struct.unpack('<I',struct.pack('<f',v))[0]
trace={};entries=collections.Counter();events=[];faults=[];current={};recording=False
def ret(value=0):
 esp=u.reg_read(UC_X86_REG_ESP)
 u.reg_write(UC_X86_REG_EAX,value&0xffffffff)
 u.reg_write(UC_X86_REG_EIP,get(esp));u.reg_write(UC_X86_REG_ESP,esp+4)
def invalid(_u,access,address,size,value,_):
 faults.append(dict(access=access,address=address,size=size,instruction=u.reg_read(UC_X86_REG_EIP)))
 return False
specs=[('Blow-root',0x53c160,223),('Blow-candidate',0x53c240,691),('indexed-direction',0x509e20,98),('mass-getter',0x4e4a70,8)]
regions=[dict(name=n,address=hex(a),size=z,sha256=sha(bytes(u.mem_read(a,z)))) for n,a,z in specs]
objects=[heap+128+2048*n for n in range(3)]
offsets=(16,8,56,60,88,92,120)
table_specs=[('indexed-vector-table',0x5b5e58,2048),('float-vector-table',0x5b6658,2048),('indexed-threshold',0x5bf2cc,4),('jump-table',0x53c4f4,36)]+[('constant',a,4) for a in (0x583c18,0x583bb0,0x583d64,0x583d68,0x583d6c,0x583d70,0x583d74,0x58307c)]
tables=[dict(name=n,address=hex(a),size=z,hex=bytes(u.mem_read(a,z)).hex(),sha256=sha(bytes(u.mem_read(a,z)))) for n,a,z in table_specs]
assert get(0x5bf2cc)==6
ref=lambda n:0 if n<0 else objects[n]
ptr=lambda p:-1 if p==0 else objects.index(p)
def state():
 return dict(objects=[[get(p+o) for o in offsets] for p in objects],directions=[struct.unpack('<H',u.mem_read(p+124,2))[0] for p in objects])
def event(name,**kw):events.append(dict(service=name,**kw,**state()))
def mutate(kind):
 if not kind:return
 if kind=='source-position':put(objects[0]+56,bits(-300.1));put(objects[0]+60,bits(987.3))
 elif kind=='candidate-position':put(objects[1]+56,bits(-1e10));put(objects[1]+60,bits(5000.3))
 elif kind=='mass-zero':put(objects[1]+120,0)
 elif kind=='mass-negative':put(objects[1]+120,bits(-2.3))
 elif kind=='force':put(objects[1]+88,bits(123.1));put(objects[1]+92,bits(-1e6))
 elif kind.startswith('direction-'):u.mem_write(objects[0]+124,struct.pack('<H',int(kind[10:])))
 elif kind=='source-disabled':put(objects[0]+16,0)
 elif kind=='candidate-destroyed':put(objects[1]+16,0x20)
 elif kind=='candidate-immobile':put(objects[1]+8,0x400000)
 elif kind=='wake-second':put(objects[2]+16,0);put(objects[2]+8,0);put(objects[2]+56,bits(120.1));put(objects[2]+60,bits(201.3))
 else:raise AssertionError(kind)
lookup_addresses=set()
def hook(_u,address,n,_):
 if address in (stop,0x4014c4):u.emu_stop();return
 if not recording:return
 for name,begin,z in specs:
  if begin<=address<begin+z:
   trace[address]=bytes(u.mem_read(address,n)).hex()
   if address==begin:entries[name]+=1
   if address==0x509e20:
    esp=u.reg_read(UC_X86_REG_ESP);direction=get(esp+4);direction=direction-(1<<32) if direction>>31 else direction
    event('direction',index=direction)
    for a in (0x5b5e58+8*direction,0x5b5e5c+8*direction):lookup_addresses.add(a)
   if address in (0x53c4c7,0x53c4d9):
    reg=UC_X86_REG_EAX if address==0x53c4c7 else UC_X86_REG_ECX
    index=u.reg_read(reg);index=index-(1<<32) if index>>31 else index
    word=2*index+(1 if address==0x53c4d9 else 0);lookup_addresses.add(0x5b6658+4*word)
    event('float-word',index=word)
   return
 esp=u.reg_read(UC_X86_REG_ESP)
 if address==0x517ed0:
  rect,callback,source=[get(esp+4*i) for i in (1,2,3)]
  assert callback==0x53c240 and source==objects[0]
  event('scan',rect_bits32=list(struct.unpack('<4I',u.mem_read(rect,16))))
  mutate(current['scan_mutation'])
  u.mem_write(stop+0x200,b'\x90')
  for candidate in current['candidates']:
   inner=esp-256;put(inner,stop+0x200);put(inner+4,ref(candidate));put(inner+8,source)
   saved=u.context_save();u.reg_write(UC_X86_REG_ESP,inner)
   # Nested emulation returns at a distinct sentinel so outer scan resumes.
   old_stop=current.get('_inner',False);current['_inner']=True
   u.emu_start(callback,stop+0x200,count=10000)
   assert u.reg_read(UC_X86_REG_EIP)==stop+0x200 and u.reg_read(UC_X86_REG_ESP)==inner+4
   u.context_restore(saved);current['_inner']=old_stop
  ret()
 elif address==0x5370e0:
  args=[get(esp+4*i) for i in (1,2,3)]
  assert args[0]==objects[0] and args[2]==0
  event('interact',candidate=ptr(args[1]))
  mutate(current['interact_mutation']);ret(int(current['can_interact']))
 else:raise AssertionError((hex(address),current['name']))
u.hook_add(UC_HOOK_CODE,hook);u.hook_add(UC_HOOK_MEM_INVALID,invalid)
u.reg_write(UC_X86_REG_ESP,sp);u.reg_write(UC_X86_REG_FPCW,0x27f);u.emu_start(0x4014b5,0x4014c4,count=20000)
assert u.reg_read(UC_X86_REG_FPCW)==0xe3f
initial=u.context_save();recording=True;cases=[]
def add(name,**kw):
 c=dict(name=name,entry='root',objects=[[0x1000000,0x400000,bits(100.1),bits(200.3),bits(.3),bits(-.2),bits(1)],[0,0,bits(110.1),bits(202.3),bits(3),bits(4),bits(2)],[0x20,0,bits(120.1),bits(201.3),bits(-1),bits(2),bits(5)]],directions=[0,13,29],candidates=[1,2],can_interact=True,scan_mutation='',interact_mutation='')
 c.update(kw);cases.append(c);return c
for direction in range(256):
 for label,dx,dy in [('east',10,2),('north',2,10),('west',-10,-2)]:
  c=add(f'root-direction/{direction}/{label}',directions=[direction,13,29]);c['objects'][1][2:4]=[bits(100.1+dx),bits(200.3+dy)]
 c=add(f'candidate-direction/{direction}',entry='candidate',directions=[direction,13,29])
word=0x53c16001
def rand(a,b):
 global word
 word=(1664525*word+1013904223)&0xffffffff
 return a+(b-a)*word/4294967296
for n in range(96):
 c=add(f'geometry/{n}',entry='candidate' if n&1 else 'root',directions=[int(rand(0,256)),13,29])
 c['objects'][0][2:4]=[bits(rand(-1e7,1e7)),bits(rand(-1e7,1e7))]
 for k in (1,2):
  c['objects'][k][2:4]=[bits(struct.unpack('<f',struct.pack('<I',c['objects'][0][2]))[0]+rand(-450,450)),bits(struct.unpack('<f',struct.pack('<I',c['objects'][0][3]))[0]+rand(-450,450))]
  c['objects'][k][4:7]=[bits(rand(-100,100)),bits(rand(-100,100)),bits(rand(.1,100))]
for f in (0,0x20,0x8000,0x1000000,0xffffffff):
 for cls in (0,0x400000,0x420000):
  c=add(f'candidate-gate/{f:x}/{cls:x}',entry='candidate');c['objects'][1][:2]=[f,cls]
for flags in (0,0x20,0x8000,0xffffffff):
 c=add(f'root-gate/{flags:x}');c['objects'][0][0]=flags
for d in (-2,-1,0,1,2):
 c=add(f'distance-boundary/{d}',entry='candidate');c['objects'][0][2:4]=[0,0];c['objects'][1][2:4]=[bits(399.9)+d,0]
for slope,label in [(0.2679,'east'),(0.3732,'north'),(.5773,'diag-min'),(.1732,'diag-max')]:
 for sign in (-1,1):
  for ulp in (-1,0,1):
   c=add(f'cone/{label}/{sign}/{ulp}',entry='candidate',directions=[0 if label=='east' else 64 if label=='north' else 32,13,29]);c['objects'][0][2:4]=[0,0];c['objects'][1][2:4]=[bits(10),bits(sign*slope*10)+ulp]
for mutation in ['source-position','candidate-position','mass-zero','mass-negative','force','direction-16','direction-64','direction-128','direction-256','direction-65535','source-disabled','candidate-destroyed','candidate-immobile','wake-second']:
 for accepted in (False,True):add(f'interact/{mutation}/{accepted}',interact_mutation=mutation,can_interact=accepted)
for mutation in ['source-position','direction-64','source-disabled','wake-second']:add(f'scan/{mutation}',scan_mutation=mutation)
for k in range(7):
 for raw in (0,0x80000000,1,0x80000001,0x7f7fffff,0xff7fffff,0x7f800000,0xff800000,0x7fc12345,0xffc12345):
  c=add(f'nonfinite/field{k}/{raw:x}',entry='candidate');c['objects'][1][k]=raw if k>=2 else c['objects'][1][k]
for raw in (0,0x80000000,1,0x80000001,0x7f800000,0xff800000,0x7fc12345):
 for field in (2,3):
  c=add(f'origin/{field}/{raw:x}');c['objects'][0][field]=raw
for direction in (256,300,512,0x7fff,0x8000,0xffff):
 c=add(f'raw-classification/{direction}',directions=[direction,13,29])
for entry in ('root','candidate'):
 c=add(f'old-unit-same-input/{entry}',entry=entry);c['objects'][0][2:4]=[bits(100),bits(200)];c['objects'][1][2:4]=[bits(110),bits(202)]
rows=[]
for mode,cw in [('actual-startup',0xe3f),('nearest-counterfactual',0x23f)]:
 for current in cases:
  u.context_restore(initial);u.reg_write(UC_X86_REG_FPCW,cw);u.mem_write(heap,b'\xa5'*0x10000);u.mem_write(stack,bytes(0x10000));events.clear()
  for p,values,direction in zip(objects,current['objects'],current['directions']):
   for off,val in zip(offsets,values):put(p+off,val)
   u.mem_write(p+124,struct.pack('<H',direction))
  before=bytes(u.mem_read(heap,0x10000))
  put(sp,stop);put(sp+4,objects[0] if current['entry']=='root' else objects[1]);put(sp+8,objects[0]);u.reg_write(UC_X86_REG_ESP,sp)
  u.emu_start(0x53c160 if current['entry']=='root' else 0x53c240,stop,count=30000)
  assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4 and u.reg_read(UC_X86_REG_FPCW)==cw
  # All tracked field writes are compared; every other object/canary byte stays identical.
  expected=bytearray(before)
  for p in objects:
   for off in offsets:expected[p-heap+off:p-heap+off+4]=u.mem_read(p+off,4)
   expected[p-heap+124:p-heap+126]=u.mem_read(p+124,2)
  assert bytes(u.mem_read(heap,0x10000))==bytes(expected)
  inp={k:v for k,v in current.items() if not k.startswith('_')}
  rows.append(dict(mode=mode,input=copy.deepcopy(inp),output=state(),events=copy.deepcopy(events),nonoutput_bytes_preserved=True))
assert not faults and original.read_bytes()==blob
raw=[]
for a in sorted(lookup_addresses):
 if not any(begin<=a<begin+z for _,begin,z in table_specs):raw.append(dict(address=hex(a),bits=get(a)))
proof=dict(original_sha256=sha(blob),original_file_unchanged=True,original_startup_FPCW=0xe3f,regions=regions,tables=tables,raw_table_words=raw,contexts=len(cases),executions=len(rows),entries=dict(entries),rows=rows,executed_instructions=[dict(address=hex(a),bytes=b) for a,b in sorted(trace.items())],faults=faults,arithmetic_conversion_lookup_or_mass_substitutions=0,scope='Complete original root/candidate and unchanged indexed-direction/mass getter. Rectangle enumeration and CanInteract are declared services. Prepared reentrant mutations only at those boundaries. Raw adjacent PE words are unit inputs; native blob initialization and full world scan/interaction composition are not inferred.')
with args.proof.open('x') as f:json.dump(proof,f,separators=(',',':'))
fixture={k:v for k,v in proof.items() if k not in ('executed_instructions','rows')}
fixture['rows']=[r for r in rows if r['mode']=='actual-startup']
with args.fixture.open('x') as f:json.dump(fixture,f,separators=(',',':'))
print(json.dumps({k:proof[k] for k in ['contexts','executions','entries','faults']}))
