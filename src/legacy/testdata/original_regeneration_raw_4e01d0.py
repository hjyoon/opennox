"""Execute the complete unchanged raw Regeneration body and both HP getters.

Armor lookup and HP adjustment are explicit observed external boundaries.
Neither callback's body, final HP mutation nor natural gameplay is credited.
Clock/field perturbations are declared independent input stimuli, not results.
"""
from pathlib import Path
import argparse,collections,gzip,hashlib,json,struct,gc
import capstone
from unicorn import (Uc,UcError,UC_ARCH_X86,UC_MODE_32,UC_HOOK_CODE,
                     UC_HOOK_MEM_READ,UC_HOOK_MEM_WRITE,UC_HOOK_MEM_INVALID,
                     UC_HOOK_INTR)
from unicorn.x86_const import *

parser=argparse.ArgumentParser(description=__doc__)
parser.add_argument('--exe',type=Path,required=True)
parser.add_argument('--control',type=lambda value:int(value,0),choices=[0xe3f,0x23f],required=True)
parser.add_argument('--output',type=Path,required=True)
parser.add_argument('--proof',type=Path,required=True)
options=parser.parse_args()
sha=lambda data:hashlib.sha256(data).hexdigest()
stage=Path(__file__).resolve().parent
encoded=(stage/'regeneration_raw_independent_inputs_4e01d0.json.gz').read_bytes()
assert sha(encoded)=='e7a9dc92b938fd023aef24b062e14b0e7127259350ec67cd2bdf8b14c030b713'
cases=json.loads(gzip.decompress(encoded));assert len(cases)==198
raw=options.exe.read_bytes()
assert sha(raw)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',raw,60)[0];opt=pe+24
base=struct.unpack_from('<I',raw,opt+28)[0]
image=bytearray((struct.unpack_from('<I',raw,opt+56)[0]+4095)&~4095)
table=opt+struct.unpack_from('<H',raw,pe+20)[0]
for number in range(struct.unpack_from('<H',raw,pe+6)[0]):
 _,rva,length,disk=struct.unpack_from('<IIII',raw,table+40*number+8)
 image[rva:rva+length]=raw[disk:disk+length]
image=bytes(image)
roots={0x4e01d0:('regeneration',204,2,'7b5720413e3ae9a93e233ed1ec5d6e07fd973d1ae6f79bc1a9fd8055d4d2daea'),
       0x4ee780:('current-hp',26,1,'c842830466581834b537f438907e121a61547ce43c6d07ef033207e13970b674'),
       0x4ee7a0:('maximum-hp',27,1,'f45238ed5aa92034eef46ed574b2967bc0595a25e2a1d3b149fe6bc20f6475d6')}
external={0x415c70:('armor-flags',1),0x4ee460:('adjust-hp',2)}
md=capstone.Cs(capstone.CS_ARCH_X86,capstone.CS_MODE_32)
instructions,ranges={},[]
for address,(name,length,argc,expected) in roots.items():
 body=image[address-base:address-base+length];assert sha(body)==expected
 decoded=list(md.disasm(body,address))
 assert sum(item.size for item in decoded)==length and decoded[-1].mnemonic=='ret'
 instructions.update({item.address:(bytes(item.bytes),item.mnemonic,name) for item in decoded})
 ranges.append(dict(name=name,address=hex(address),size=length,sha256=expected))

heap,stack,aux,bad=0x21000000,0x22000000,0x23000000,0x41000000
sp,stop,initialize,cw=stack+0xe000,aux+0x100,aux+0x200,aux+0x300
frame_address,fps_address=0x84ea04,0x85b3fc
u=None
init_bytes=b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw)
saved={UC_X86_REG_EBX:0x12345678,UC_X86_REG_ESI:0x23456789,
       UC_X86_REG_EDI:0x3456789a,UC_X86_REG_EBP:0x456789ab}
active=False
seen=set();fault_domains=collections.Counter();all_calls=collections.Counter()
accesses,writes,calls,returns,frames,pending=[],[],[],[],[],[]
call_counts,read_counts=collections.Counter(),collections.Counter()
fault=None
applied=set()
current={};roles={}
locations={
 0x4e01df:('item','holder'),0x4e01ed:('owner','health'),
 0x4e01fb:('world','frame'),0x4e0200:('owner','last-damage'),
 0x4e0206:('world','fps'),0x4e0216:('owner','flags'),
 0x4e023a:('item','class'),0x4e0242:('effect','rate'),
 0x4e0269:('world','fps'),0x4e0282:('world','frame'),
 0x4ee788:('owner','health'),0x4ee792:('health','current'),
 0x4ee7a8:('owner','health'),0x4ee7b2:('health','maximum')}
def field(role,name):
 if role=='world':return (frame_address if name=='frame' else fps_address,4)
 offsets={('item','holder'):492,('item','class'):8,('owner','health'):556,
          ('owner','last-damage'):536,('owner','flags'):16,
          ('effect','rate'):108,('health','current'):0,('health','maximum'):4}
 return roles[role]+offsets[role,name],2 if role=='health' else 4
def write_stimulus(index):
 mutation=current['mutations'][index]
 role=mutation.get('target_role',mutation['role'])
 name=mutation.get('target_field',mutation['field'])
 address,length=field(role,name);value=mutation['value']
 if isinstance(value,str):value=0 if value=='null' else bad if value=='bad' else roles[value]
 value&=(1<<(8*length))-1
 assert heap<=address and address+length<=heap+0x10000 or address in (frame_address,fps_address)
 u.mem_write(address,value.to_bytes(length,'little'))
 record=dict(kind='stimulus',role=role,field=name,address=hex(address),size=length,value=value,
             complete=True,source_input_mutation=index,instruction=hex(u.reg_read(UC_X86_REG_EIP)))
 accesses.append(record);writes.append(record.copy());applied.add(index)
def code(uc,address,length,unused):
 if not active:return
 while pending:write_stimulus(pending.pop(0))
 if address in roots or address in external:
  name,argc=(roots[address][0],roots[address][2]) if address in roots else external[address]
  esp=uc.reg_read(UC_X86_REG_ESP)
  args=list(struct.unpack('<'+'I'*argc,uc.mem_read(esp+4,4*argc)))
  call_counts[name]+=1;all_calls[name]+=1
  for index,mutation in enumerate(current['mutations']):
   if mutation['when']=='before-call' and mutation['function']==name and mutation['occurrence']==call_counts[name]:write_stimulus(index)
  calls.append(dict(function=name,address=hex(address),args=args,depth=len(frames),external=address in external))
  if address in external:
   if name=='armor-flags':
    assert args==[roles['item']]
    value=current['armor_flags']
   else:
    assert args==[roles['owner'],1]
    value=current['service_return']
   returns.append(dict(function=name,result=value,depth=len(frames),external=True))
   destination=struct.unpack('<I',uc.mem_read(esp,4))[0]
   uc.reg_write(UC_X86_REG_EAX,value);uc.reg_write(UC_X86_REG_ESP,esp+4);uc.reg_write(UC_X86_REG_EIP,destination)
   return
  frames.append(name)
 assert address in instructions and length==len(instructions[address][0]),hex(address)
 assert bytes(uc.mem_read(address,length))==instructions[address][0]
 seen.add(address)
 if instructions[address][1]=='ret':
  name=frames.pop();assert name==instructions[address][2]
  returns.append(dict(function=name,result=uc.reg_read(UC_X86_REG_EAX),depth=len(frames),external=False))
def read(uc,access,address,length,unused_value,unused):
 if not active or stack<=address<stack+0x10000:return
 instruction=uc.reg_read(UC_X86_REG_EIP);assert instruction in locations,hex(instruction)
 role,name=locations[instruction]
 try:value=int.from_bytes(uc.mem_read(address,length),'little')
 except UcError:value=None
 accesses.append(dict(kind='read',role=role,field=name,address=hex(address),size=length,
                      value=value,complete=value is not None,instruction=hex(instruction)))
 read_counts[role,name]+=1
 if value is not None:
  for index,mutation in enumerate(current['mutations']):
   if mutation['when']=='after-read' and mutation['role']==role and mutation['field']==name and mutation['occurrence']==read_counts[role,name]:pending.append(index)
def write(uc,access,address,length,value,unused):
 if not active:return
 assert stack<=address and address+length<=stack+0x10000
 writes.append(dict(kind='write',address=hex(address),size=length,value=value&((1<<(8*length))-1),instruction=hex(uc.reg_read(UC_X86_REG_EIP))))
def invalid(uc,access,address,length,value,unused):
 global fault
 instruction=uc.reg_read(UC_X86_REG_EIP)
 fault=dict(kind='memory',access=access,address=hex(address),size=length,instruction=hex(instruction))
 assert access in (19,23) and instruction in locations,fault
 role,name=locations[instruction]
 if accesses and accesses[-1]['kind']=='read' and accesses[-1]['address']==hex(address):
  accesses[-1].update(complete=False,value=None)
 else:accesses.append(dict(kind='read',role=role,field=name,address=hex(address),size=length,value=None,complete=False,instruction=hex(instruction)))
 return False
def interrupt(uc,number,unused):
 global fault
 assert number==0,(current['name'],number,hex(uc.reg_read(UC_X86_REG_EIP)),active)
 instruction=uc.reg_read(UC_X86_REG_EIP);assert instruction in (0x4e0279,0x4e0287)
 fault=dict(kind='integer-divide',interrupt=number,instruction=hex(instruction));uc.emu_stop()
def fresh_emulator():
 emulator=Uc(UC_ARCH_X86,UC_MODE_32)
 for address,length in [(base,len(image)),(heap,0x10000),(stack,0x10000),(aux,0x1000)]:emulator.mem_map(address,length)
 emulator.mem_write(base,image)
 emulator.mem_write(initialize,init_bytes);emulator.mem_write(cw,struct.pack('<H',options.control))
 emulator.hook_add(UC_HOOK_CODE,code);emulator.hook_add(UC_HOOK_MEM_READ,read)
 emulator.hook_add(UC_HOOK_MEM_WRITE,write);emulator.hook_add(UC_HOOK_MEM_INVALID,invalid)
 emulator.hook_add(UC_HOOK_INTR,interrupt)
 return emulator
results=[]
for current in cases:
 # A trapped integer DIV leaves emulator exception state behind. A fresh
 # machine for every input also makes all architectural state independent.
 u=fresh_emulator()
 def pointer(kind,valid,edge):return 0 if kind=='null' else bad if kind=='bad' else edge if kind.startswith('last-') else valid
 roles={'item':pointer(current['item'],heap+0x3000,heap+0x10000-496),
        'owner':pointer(current['owner'],heap+0x5000,heap+0x10000-560),
        'health':pointer(current['health'],heap+0x7000,heap+0x10000-(2 if current['health']=='last-current' else 6)),
        'effect':pointer(current['effect'],heap+0x1000,heap+0x10000-112)}
 if current['owner']=='item':roles['owner']=roles['item']
 initial=bytearray(((number*37+current['seed']*19)^(number>>3))&255 for number in range(0x10000))
 for role,name,value in [('item','holder',roles['owner']),('item','class',current['obj_class']),
                         ('owner','health',roles['health']),('owner','last-damage',current['last_damage']),
                         ('owner','flags',current['owner_flags']),('health','current',current['current_hp']),
                         ('health','maximum',current['maximum_hp']),('effect','rate',current['rate'])]:
  address,length=field(role,name)
  if heap<=address and address+length<=heap+0x10000:initial[address-heap:address-heap+length]=value.to_bytes(length,'little')
 u.mem_write(heap,bytes(initial))
 u.mem_write(frame_address,struct.pack('<I',current['frame']));u.mem_write(fps_address,struct.pack('<I',current['fps']))
 initial_image=bytes(u.mem_read(base,len(image)))
 initial_stack=bytearray(((number*11+current['seed'])^(number>>4))&255 for number in range(0x10000))
 caller=struct.pack('<3I',stop,roles['effect'],roles['item'])
 initial_stack[sp-stack:sp-stack+len(caller)]=caller
 u.mem_write(stack,bytes(initial_stack))
 u.reg_write(UC_X86_REG_ESP,sp)
 u.emu_start(initialize,initialize+len(init_bytes),count=8)
 for register,value in saved.items():u.reg_write(register,value)
 u.reg_write(UC_X86_REG_EAX,0x98765432);u.reg_write(UC_X86_REG_ECX,0x87654321)
 u.reg_write(UC_X86_REG_EDX,0x76543210);u.reg_write(UC_X86_REG_EFLAGS,0x202)
 accesses,writes,calls,returns,frames,pending=[],[],[],[],[],[]
 call_counts,read_counts=collections.Counter(),collections.Counter()
 fault=None;applied=set();active=True
 try:u.emu_start(0x4e01d0,stop,count=4096)
 except UcError:assert fault is not None
 finally:active=False
 assert not pending
 final_heap=bytes(u.mem_read(heap,0x10000));final_stack=bytes(u.mem_read(stack,0x10000))
 expected_heap,expected_stack,expected_image=bytearray(initial),bytearray(initial_stack),bytearray(initial_image)
 for record in writes:
  address=int(record['address'],16);length=record['size'];value=record['value'].to_bytes(length,'little')
  if heap<=address<heap+0x10000:expected_heap[address-heap:address-heap+length]=value
  elif stack<=address<stack+0x10000:expected_stack[address-stack:address-stack+length]=value
  else:expected_image[address-base:address-base+length]=value
 assert final_heap==bytes(expected_heap) and final_stack==bytes(expected_stack)
 assert bytes(u.mem_read(base,len(image)))==bytes(expected_image)
 assert bytes(u.mem_read(sp,len(caller)))==caller
 assert u.reg_read(UC_X86_REG_FPCW)==options.control and u.reg_read(UC_X86_REG_FPSW)==0
 if fault is None:
  assert not frames and u.reg_read(UC_X86_REG_ESP)==sp+4
  assert all(u.reg_read(register)==value for register,value in saved.items())
 else:fault_domains[fault['kind'],fault['instruction']]+=1
 results.append(dict(name=current['name'],roles=roles,initial_heap=bytes(initial).hex(),final_heap=final_heap.hex(),
                     initial_stack=bytes(initial_stack).hex(),final_stack=final_stack.hex(),
                     initial_image_sha256=sha(initial_image),final_image_sha256=sha(bytes(expected_image)),
                     caller_before=caller.hex(),caller_after=bytes(u.mem_read(sp,len(caller))).hex(),
                     registers=dict(eax=u.reg_read(UC_X86_REG_EAX),ecx=u.reg_read(UC_X86_REG_ECX),
                                    edx=u.reg_read(UC_X86_REG_EDX),ebx=u.reg_read(UC_X86_REG_EBX),
                                    esi=u.reg_read(UC_X86_REG_ESI),edi=u.reg_read(UC_X86_REG_EDI),
                                    ebp=u.reg_read(UC_X86_REG_EBP),esp=u.reg_read(UC_X86_REG_ESP)),
                     accesses=accesses,writes=writes,calls=calls,returns=returns,
                     applied_input_mutations=sorted(applied),fault=fault))
 u=None
 gc.collect()
assert not set(instructions)-seen,[hex(address) for address in sorted(set(instructions)-seen)]
assert sha(options.exe.read_bytes())==sha(raw)
data=json.dumps(results,sort_keys=True,separators=(',',':')).encode()+b'\n'
with options.output.open('xb') as stream:stream.write(gzip.compress(data,mtime=0))
proof=dict(original_exe_sha256=sha(raw),original_image_sha256=sha(image),original_file_unchanged=True,
           independent_inputs_sha256=sha(encoded),cases=len(cases),original_function_ranges=ranges,
           all257_original_bytes_reached=True,instruction_addresses=len(instructions),
           actual_original_HP_getter_bodies=True,clock_values_read_from_original_global_addresses=True,
           declared_external_services=['armor-flags lookup: input flags returned, no body credit','HP adjustment: exact cached owner/delta observed; body and effects excluded'],
           external_service_effects_credited=False,root_C_ABI_void=True,hardware_EAX_retained_but_not_a_defined_void_return=True,
           full64KiB_heap_and_stack_before_after_retained=True,all_writes_reconstruct_whole_memory=True,
           fresh_independent_emulator_state_for_each_input=True,
           caller_words_unchanged=True,nonvolatile_registers_and_stack_preserved_on_success=True,
           all_original_image_bytes_unchanged_except_declared_clock_input_stimuli=True,
           FPU_control=options.control,FPU_status_unchanged=True,original_calls=dict(all_calls),
           external_accesses=sum(len(row['accesses']) for row in results),
           memory_and_arithmetic_faults=[dict(kind=key[0],instruction=key[1],count=value) for key,value in sorted(fault_domains.items())],
           native_initial_memory_credit=False,native_implementation_credit=False,actual_product_credit=False,overall100=False)
with options.proof.open('x') as stream:json.dump(proof,stream,indent=2);stream.write('\n')
print(json.dumps(proof),flush=True)
