"""Execute four complete GAME.EXE Skeleton/drop bodies; only named world services are replaced."""
from pathlib import Path
import argparse, collections, gzip, hashlib, json, struct
import capstone
from unicorn import Uc,UcError,UC_ARCH_X86,UC_MODE_32,UC_HOOK_CODE,UC_HOOK_MEM_WRITE,UC_HOOK_MEM_INVALID,UC_HOOK_INTR
from unicorn.x86_const import *
ap=argparse.ArgumentParser(description=__doc__);ap.add_argument('--exe',type=Path,required=True);ap.add_argument('--control',type=lambda s:int(s,0),choices=[0xe3f,0x23f],required=True);ap.add_argument('--output',type=Path,required=True);ap.add_argument('--proof',type=Path,required=True);a=ap.parse_args();p=Path(__file__).resolve().parent
sha=lambda raw:hashlib.sha256(raw).hexdigest()
encoded=(p/'skeleton_death_independent_inputs_54a310.json.gz').read_bytes();assert sha(encoded)=='70a545ed723b29909a80fa5dc084869fad81d26566b316d395c44152f4cc4776'
inputs=json.loads(gzip.decompress(encoded));assert len(inputs)==169
raw=a.exe.read_bytes();assert sha(raw)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',raw,60)[0];op=pe+24;table=op+struct.unpack_from('<H',raw,pe+20)[0];base=struct.unpack_from('<I',raw,op+28)[0];image_size=(struct.unpack_from('<I',raw,op+56)[0]+4095)&~4095
u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,image_size)
for i in range(struct.unpack_from('<H',raw,pe+6)[0]):
 _,va,size,off=struct.unpack_from('<IIII',raw,table+40*i+8)
 if size:u.mem_write(base+va,raw[off:off+size])
original_image=bytes(u.mem_read(base,image_size));heap,stack,aux=0x21000000,0x22000000,0x23000000
for v in [heap,stack,aux]:u.mem_map(v,0x10000)
roles=[heap+n for n in [0x1000,0x2000,0x3000,0x4000,0x5000,0x6000,0x9000,0xa000]]
ptr=lambda role:0 if role==-1 else 0x41000000 if role==-2 else roles[role]
role=lambda v:next((n for n in [-1,-2,*range(8)] if ptr(n)==v),None)
sp,stop,init,cw,resume,double_return,double_value=[stack+0xe000,aux+0x100,aux+0x200,aux+0x300,aux+0x400,aux+0x500,aux+0x600]
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw));u.mem_write(cw,struct.pack('<H',a.control));u.mem_write(resume,b'\x90');u.mem_write(double_return,b'\xdd\x05'+struct.pack('<I',double_value)+b'\xc3')
names=['TestItem','WeaponPower1','Material2','FireEffect','OtherEffect','Skull','ArmBone','LegBone','Sword','WoodenShield','SteelShield']
name_addresses={s:aux+0x1000+i*64 for i,s in enumerate(names)}
for s,at in name_addresses.items():u.mem_write(at,s.encode()+b'\0')
name_ptr=lambda n:0 if n==-1 else 0x41000000 if n==-2 else name_addresses[names[n]]
rd=lambda at:struct.unpack('<I',u.mem_read(at,4))[0]
cache_at,fps_at,divisor_at,bone_at=0x834a58,0x85b3fc,0x5cd274,0x5cd864
services={0x40a5c0:('flags',1),0x415fa0:('int',4),0x416030:('float',4),0x4e3810:('new',1),0x4ed970:('point',3),0x4daa50:('create',5),0x4e46f0:('raise',2),0x52df80:('force',3),0x501960:('audio',4),0x522ff0:('pointfx',2),0x511660:('decay',2),0x413290:('modid',1),0x413330:('moddesc',1),0x4e4990:('apply',2),0x53eae0:('armor',1),0x53aab0:('weapon',1)}
md=capstone.Cs(capstone.CS_ARCH_X86,capstone.CS_MODE_32);instructions={};ranges=[]
for pc,size,digest in [(0x54a310,124,'fc81e1cbc8cf0c7d255a3ce89924495fc208b5af1811684b20e7fdffd549f274'),(0x54a390,297,'a7456c5dea5f19a123f5570f4e705d30fd1e44f4d3d497b43481ef88f2c227e8'),(0x54a4c0,644,'ed201956c5e80ca9a6581eccaf71698954cd9ab249da5c07a6e4c0a25eb061fd'),(0x54a750,124,'3f9bd64fe0debb63b37176b6c1f0bf6aee4d760ebd2ba096b1d4be8c1062a51b')]:
 body=bytes(u.mem_read(pc,size));assert sha(body)==digest;seq=list(md.disasm(body,pc));assert sum(i.size for i in seq)==size and seq[-1].mnemonic=='ret';instructions.update({i.address:bytes(i.bytes) for i in seq});ranges.append(dict(entry=hex(pc),bytes=size,sha256=digest))
saved={UC_X86_REG_EBX:0x12345678,UC_X86_REG_ESI:0x23456789,UC_X86_REG_EDI:0x3456789a,UC_X86_REG_EBP:0x456789ab}
registers=lambda:{k:u.reg_read(k) for k in saved}
active=False;seen=set();calls=collections.Counter();observed_faults=collections.Counter();scratch_words=collections.Counter();current={};expected_heap=bytearray();expected_image=bytearray();events=[];frames=[];fault=None;occ={};coop=1;fault_detail=None

def wr(at,data):
 u.mem_write(at,data)
 if heap<=at and at+len(data)<=heap+0x10000:expected_heap[at-heap:at-heap+len(data)]=data
 elif base<=at and at+len(data)<=base+image_size:expected_image[at-base:at-base+len(data)]=data
 elif stack<=at and at+len(data)<=stack+0x10000:pass
 else:raise AssertionError(('undeclared world write',hex(at)))
def world_read(at,size,site):
 global fault,fault_detail
 try:return bytes(u.mem_read(at,size))
 except UcError:
  fault=dict(kind='memory',site=site);fault_detail=dict(kind='declared-world-read',site=site,address=hex(at),size=size);raise

def read_name(v):
 if v==0:return None
 raw=world_read(v,48,'name');return raw.split(b'\0',1)[0].decode('ascii')
def origin(v):
 # Preserve the original pointer argument until the declared world reads it.
 return list(struct.unpack('<II',world_read(v,8,'origin')))
def state():return dict(cache=rd(cache_at),divisor=rd(divisor_at),fps=rd(fps_at),coop=coop)
def mutate(plan):
 global coop
 for key,off in [('position',56),('class',8),('subclass',12),('flags',16)]:
  if key in plan:
   r,*vals=plan[key];wr(ptr(r)+off,struct.pack('<'+'I'*len(vals),*vals))
 if 'use' in plan:r,data=plan['use'];wr(ptr(r)+736,struct.pack('<I',ptr(data)))
 for key,at in [('cache',cache_at),('divisor',divisor_at),('fps',fps_at)]:
  if key in plan:wr(at,struct.pack('<I',plan[key]&0xffffffff))
 if 'coop' in plan:coop=plan['coop']
def finish(frame):
 assert registers()==frame['registers'] and u.reg_read(UC_X86_REG_FPSW)&0x3800==0
 events.append(dict(kind=frame['kind']+'-return',index=frame['index'],result=frame['semantic']))
 if frame['kind']=='float':u.mem_write(double_value,struct.pack('<Q',frame['result']));u.reg_write(UC_X86_REG_ESP,frame['esp']);u.reg_write(UC_X86_REG_EIP,double_return)
 else:u.reg_write(UC_X86_REG_EAX,frame['result']&0xffffffff);u.reg_write(UC_X86_REG_ESP,frame['esp']+4);u.reg_write(UC_X86_REG_EIP,rd(frame['esp']))
def on_code(_u,pc,size,_):
 global fault
 if not active:return
 esp=u.reg_read(UC_X86_REG_ESP)
 if pc in [double_return,double_return+6]:return
 if pc==resume:
  frame=frames.pop();assert esp==frame['esp']-4*frame['nested_argc']
  result=None if frame['entry']=='54a390' else u.reg_read(UC_X86_REG_EAX);assert result in [None,1]
  events.append(dict(kind='nested-return',index=frame['index'],result=result));finish(frame);return
 if pc in services:
  kind,n=services[pc];args=[rd(esp+4*i) for i in range(1,n+1)];at=len([v for v in events if v['kind'] in services_names]);seq=occ.get(kind,0);occ[kind]=seq+1;calls[kind]+=1;plan=current['script'].get(f'{kind}:{seq}',{});event=dict(kind=kind,index=at,**state());result=0xa5a50000+at;semantic=result
  if kind=='flags':assert args==[2048];result=semantic=coop
  elif kind=='int':
   lo,hi=(struct.unpack('<i',struct.pack('<I',v))[0] for v in args[:2]);is_parts=args[3] in [0x425,0x42c]
   event.update(min=lo,max=hi);result=semantic=current['roll'] if (lo,hi)==(0,100) else current['parts'] if is_parts else current['seconds']
  elif kind=='float':event.update(min=args[0],max=args[1]);result=semantic=current['float_bits']
  elif kind=='new':
   name=read_name(args[0]);r=current['new'].get(name,-1);event.update(name=name);result=ptr(r);semantic=r
  elif kind=='point':
   event.update(radius=args[0],origin=origin(args[1]));wr(args[2],struct.pack('<II',0x42c80000+seq,0x43480000+seq))
  elif kind=='create':
   assert role(args[0]) is not None and role(args[1])==-1 and args[4]==0;event.update(object=role(args[0]),owner=-1,position=args[2:4],extra=0)
  elif kind=='raise':assert role(args[0]) is not None;event.update(object=role(args[0]),height=args[1])
  elif kind=='force':
   assert role(args[1]) is not None;event.update(object=role(args[1]),origin=origin(args[0]),force=args[2])
   # Capture original stores at the next meaningful world boundary.
   event['motion']=list(struct.unpack('<I',world_read(args[1]+16,4,'object')))+list(struct.unpack('<I',world_read(args[1]+108,4,'object')))+list(struct.unpack('<I',world_read(args[1]+116,4,'object')))
  elif kind=='audio':assert args[0]==356 and args[2:]==[0,0];event.update(sound=args[0],unit=role(args[1]))
  elif kind=='pointfx':assert args[0]==0x8a;event.update(effect=args[0],origin_pointer=role((args[1]-56)&0xffffffff))
  elif kind=='decay':assert role(args[0]) is not None;event.update(object=role(args[0]),duration=args[1])
  elif kind=='modid':
   name=read_name(args[0]);event.update(name=name);result=semantic=31 if name is None else names.index(name)+40
  elif kind=='moddesc':
   event.update(id=args[0]);semantic=-1 if plan.get('return_nil') else args[0];result=0 if semantic==-1 else aux+0x4000+args[0]*16
  elif kind=='apply':
   event.update(object=role(args[0]),modifiers=[-1 if rd(args[1]+4*i)==0 else (rd(args[1]+4*i)-aux-0x4000)//16 for i in range(4)])
   # The fifth word is uninitialized by the original caller, not a semantic modifier.
   scratch_words[hex(rd(args[1]+16))]+=1
  else:event.update(object=role(args[0]))
  events.append(event);mutate(plan);frame=dict(kind=kind,index=at,esp=esp,result=result,semantic=semantic,registers=registers())
  if 'reenter' in plan:
   entry,unit=plan['reenter'];frame['entry']=entry;frames.append(frame);events.append(dict(kind='nested',index=at,entry=entry,unit=unit));nested_args=[ptr(unit)]
   if entry=='54a390':nested_args += [name_ptr(current['type']),*[name_ptr(n) for n in current['mods']],current['charge']&0xffffffff]
   u.mem_write(esp-4*(len(nested_args)+1),struct.pack('<'+'I'*(len(nested_args)+1),resume,*nested_args));u.reg_write(UC_X86_REG_ESP,esp-4*(len(nested_args)+1));frame['nested_argc']=len(nested_args);u.reg_write(UC_X86_REG_EIP,int(entry,16))
  else:finish(frame)
  return
 assert pc in instructions and bytes(u.mem_read(pc,size))==instructions[pc],(current['name'],hex(pc),size)
 seen.add(pc)
services_names={v[0] for v in services.values()}
def on_write(_u,access,at,size,value,_):
 if not active:return
 if stack<=at and at+size<=stack+0x10000:
  if at>=sp and at<sp+len(caller):
   assert u.reg_read(UC_X86_REG_EIP)==0x54a5f8 and at==sp+4 and size==4 and current['entry']=='54a4c0'
   caller_expected[4:8]=value.to_bytes(4,'little')
  return
 pc=u.reg_read(UC_X86_REG_EIP)
 allowed={0x54a54a:108,0x54a562:116,0x54a569:16,0x54a694:108,0x54a6ac:116,0x54a6b3:16}
 if pc in allowed:assert size==4 and (role(at-allowed[pc]) is not None)
 elif pc==0x54a72d:assert at==cache_at and size==4
 elif pc==0x54a489:assert size==1 and (role(at-1) in [6,7,-1,-2])
 else:raise AssertionError((current['name'],'undeclared original write',hex(pc),hex(at),size))
 if not(heap<=at and at+size<=heap+0x10000 or base<=at and at+size<=base+image_size):return
 dest=expected_heap if heap<=at<heap+0x10000 else expected_image;off=at-(heap if dest is expected_heap else base);dest[off:off+size]=(value&((1<<(8*size))-1)).to_bytes(size,'little')
def on_fault(_u,access,at,size,value,_):
 global fault,fault_detail
 pc=u.reg_read(UC_X86_REG_EIP);site='bone-table' if pc==0x54a612 else 'object' if pc in [0x54a542,0x54a68c,0x54a3ee,0x54a474,0x54a47d,0x54a483,0x54a48c] else 'charge' if pc==0x54a489 else 'object'
 fault=dict(kind='memory',site=site);fault_detail=dict(kind='original-memory-access',pc=hex(pc),access=access,address=hex(at),size=size);return False
def on_intr(_u,number,_):
 global fault,fault_detail
 assert number==0 and u.reg_read(UC_X86_REG_EIP)==0x54a71c,(current['name'],number,hex(u.reg_read(UC_X86_REG_EIP)))
 fault=dict(kind='arithmetic',site='bone-modulo');fault_detail=dict(kind='original-cpu-divide-error',pc='0x54a71c',interrupt=number);u.emu_stop()
u.hook_add(UC_HOOK_CODE,on_code);u.hook_add(UC_HOOK_MEM_WRITE,on_write);u.hook_add(UC_HOOK_MEM_INVALID,on_fault);u.hook_add(UC_HOOK_INTR,on_intr)
outputs=[];details=[];original_aux=bytes(u.mem_read(aux,0x10000))
for current in inputs:
 active=False;u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,image_size)
 for address in [heap,stack,aux]:u.mem_map(address,0x10000)
 u.mem_write(aux,original_aux);u.hook_add(UC_HOOK_CODE,on_code);u.hook_add(UC_HOOK_MEM_WRITE,on_write);u.hook_add(UC_HOOK_MEM_INVALID,on_fault);u.hook_add(UC_HOOK_INTR,on_intr)
 u.mem_write(base,original_image);expected_image=bytearray(original_image);expected_heap=bytearray((i*17+(i>>7)*29)&255 for i in range(0x10000));u.mem_write(heap,bytes(expected_heap))
 for n in range(6):
  for off,values in [(8,[current['classes'][n]]),(12,[current['subclasses'][n]]),(16,[0x11223344+n]),(56,[0x40e00000+n,0x41000000+n]),(736,[ptr(current['uses'][n])])]:wr(roles[n]+off,struct.pack('<'+'I'*len(values),*values))
 for key,at in [('cache',cache_at),('divisor',divisor_at),('fps',fps_at)]:wr(at,struct.pack('<I',current[key]&0xffffffff))
 coop=current['coop'];u.mem_write(stack,b'\xcc'*0x10000);args=[ptr(current['unit'])]
 if current['entry']=='54a390':args += [name_ptr(current['type']),*[name_ptr(n) for n in current['mods']],current['charge']&0xffffffff]
 caller=struct.pack('<'+'I'*(len(args)+1),stop,*args)+bytes(range(48));u.mem_write(sp,caller);caller_expected=bytearray(caller)
 for k,v in saved.items():u.reg_write(k,v)
 u.reg_write(UC_X86_REG_ESP,sp);u.emu_start(init,init+8,count=2);events=[];frames=[];fault=None;fault_detail=None;occ={};active=True
 try:u.emu_start(int(current['entry'],16),stop,count=100000)
 except UcError:assert fault is not None,current['name']
 active=False
 assert bytes(u.mem_read(heap,0x10000))==expected_heap,(current['name'],'complete heap')
 assert bytes(u.mem_read(base,image_size))==expected_image,(current['name'],'complete image')
 assert bytes(u.mem_read(sp,len(caller)))==caller_expected,(current['name'],'caller guard')
 assert u.reg_read(UC_X86_REG_FPCW)==a.control
 if not fault:
  assert not frames and u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4 and registers()==saved,(current['name'],'return guard')
  assert u.reg_read(UC_X86_REG_FPSW)&0x3800==0
  if current['entry']!='54a390':assert u.reg_read(UC_X86_REG_EAX)==1
 if fault:observed_faults[fault['kind']+':'+fault['site']]+=1;details.append(dict(name=current['name'],**fault_detail))
 outputs.append(dict(name=current['name'],result=None if fault or current['entry']=='54a390' else 1,fault=fault,events=events,**state(),heap_sha256=sha(expected_heap)))
missing=[hex(v) for v in instructions.keys()-seen];assert not missing,('missing original instructions',missing)
decoded=json.dumps(outputs,sort_keys=True,separators=(',',':')).encode()+b'\n';packed=bytearray(gzip.compress(decoded,mtime=0));packed[9]=0xff
with a.output.open('xb')as out:out.write(packed)
proof=dict(original_exe_sha256=sha(raw),input_sha256=sha(encoded),cases=len(inputs),control=hex(a.control),ranges=ranges,all1189_instruction_bytes_executed=True,complete_heap_and_image_guards=True,caller_guard_except_declared_reused_debris_argument=True,valid_return_stack_registers_preserved=True,FPU_CW_and_valid_TOP_preserved=True,original_faults=dict(observed_faults),fault_details=details,only_declared_world_services_replaced={hex(k):v[0] for k,v in services.items()},original_reentrant_calls_executed=True,world_calls=dict(calls),uninitialized_fifth_modifier_word_observed=dict(scratch_words),uninitialized_word_not_compared_as_semantic_modifier=True,output_sha256=sha(packed),decoded_output_sha256=sha(decoded),native_source_used=False)
with a.proof.open('x')as out:json.dump(proof,out,indent=2);out.write('\n')
print(json.dumps(proof),flush=True)
