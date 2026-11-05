"""Execute the original raw inventory checker and its actual pure callees."""
from pathlib import Path
import argparse
import collections
import gzip
import hashlib
import json
import struct

import capstone
from unicorn import (Uc,UcError,UC_ARCH_X86,UC_MODE_32,UC_HOOK_CODE,
                     UC_HOOK_MEM_READ,UC_HOOK_MEM_WRITE,UC_HOOK_MEM_INVALID)
from unicorn.x86_const import *

parser=argparse.ArgumentParser(description=__doc__)
parser.add_argument('--exe',type=Path,required=True)
parser.add_argument('--control',type=lambda text:int(text,0),choices=[0xe3f,0x23f],required=True)
parser.add_argument('--output',type=Path,required=True)
parser.add_argument('--proof',type=Path,required=True)
options=parser.parse_args()
stage=Path(__file__).resolve().parent
sha=lambda data:hashlib.sha256(data).hexdigest()
packed=(stage/'inversion_inventory_independent_inputs_4fa4f0.json.gz').read_bytes()
assert sha(packed)=='69c228a2eb2ef67ce89f8d2127766d0b209cb29d80433a3d43fda9ec6e515e1a'
cases=json.loads(gzip.decompress(packed));assert len(cases)==207
raw=options.exe.read_bytes()
assert sha(raw)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',raw,60)[0];opt=pe+24
base=struct.unpack_from('<I',raw,opt+28)[0]
size=(struct.unpack_from('<I',raw,opt+56)[0]+4095)&~4095
image=bytearray(size)
table=opt+struct.unpack_from('<H',raw,pe+20)[0]
for number in range(struct.unpack_from('<H',raw,pe+6)[0]):
    _,rva,length,disk=struct.unpack_from('<IIII',raw,table+number*40+8)
    image[rva:rva+length]=raw[disk:disk+length]
image=bytes(image)
md=capstone.Cs(capstone.CS_ARCH_X86,capstone.CS_MODE_32)
roots={0x4fa4f0:('checker',159,2,'8bf018e2ff189e5e4f3725004765a960730e7b8d0748a27f6c5a148c0e6967c2'),
       0x4ec580:('parent',38,1,'7e7580cf0eb5c57ba46812b911b2fd4fa3224fc26ee646d777e207b350e91859'),
       0x4e03d0:('inversion',22,6,'fe6639dcaebfa8919164437706cf56e6298ca32a42c741a2fc1e770373922d6f')}
instructions,ranges={},[]
for address,(name,length,argc,expected)in roots.items():
    body=image[address-base:address-base+length];assert sha(body)==expected
    decoded=list(md.disasm(body,address));assert sum(item.size for item in decoded)==length and decoded[-1].mnemonic=='ret'
    instructions.update({item.address:(bytes(item.bytes),item.mnemonic,name)for item in decoded})
    ranges.append(dict(name=name,address=hex(address),size=length,sha256=expected))

heap,stack,aux,bad=0x21000000,0x22000000,0x23000000,0x41000000
objects=[heap+n for n in [0x3000,0x5000,0x6000,0x7000,0x8000,0xd000,0xe000]]
arrays=[heap+n for n in [0x4000,0x9000,0xa000]]
mods=[heap+n for n in [0x1000,0x2000,0xb000,0xc000]]
pointer=lambda role,roles:0 if role==-1 else bad if role==-2 else roles[role]
sp,stop,initialize,cw=stack+0xe000,aux+0x100,aux+0x200,aux+0x300
saved={UC_X86_REG_EBX:0x12345678,UC_X86_REG_ESI:0x23456789,
       UC_X86_REG_EDI:0x3456789a,UC_X86_REG_EBP:0x456789ab}
u=Uc(UC_ARCH_X86,UC_MODE_32)
for address,length in [(base,len(image)),(heap,0x10000),(stack,0x10000),(aux,0x1000)]:u.mem_map(address,length)
u.mem_write(base,image)
initialize_bytes=b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw)
u.mem_write(initialize,initialize_bytes);u.mem_write(cw,struct.pack('<H',options.control))
active=False
accesses,writes,calls,returns,frames=[],[],[],[],[]
fault=None
seen=set();faults=collections.Counter();function_calls=collections.Counter()
def code(uc,address,length,unused):
    if not active:return
    assert address in instructions and length==len(instructions[address][0]),hex(address)
    assert bytes(uc.mem_read(address,length))==instructions[address][0]
    seen.add(address)
    if address in roots:
        name,_,argc,_=roots[address]
        current_sp=uc.reg_read(UC_X86_REG_ESP)
        args=list(struct.unpack('<'+'I'*argc,uc.mem_read(current_sp+4,argc*4)))
        calls.append(dict(function=name,address=hex(address),args=args,depth=len(frames)))
        frames.append(name);function_calls[name]+=1
    if instructions[address][1]=='ret':
        name=frames.pop();assert name==instructions[address][2]
        returns.append(dict(function=name,result=uc.reg_read(UC_X86_REG_EAX),depth=len(frames)))
def read(uc,access,address,length,unused_value,unused):
    if not active or stack<=address<stack+0x10000:return
    try:value=int.from_bytes(uc.mem_read(address,length),'little')
    except UcError:value=None
    accesses.append(dict(kind='read',address=hex(address),size=length,value=value,
                         complete=value is not None,instruction=hex(uc.reg_read(UC_X86_REG_EIP))))
def write(uc,access,address,length,value,unused):
    if not active:return
    assert stack<=address and address+length<=stack+0x10000 or heap<=address and address+length<=heap+0x10000
    record=dict(address=hex(address),size=length,value=value&((1<<(8*length))-1))
    writes.append(record)
    if not stack<=address<stack+0x10000:
        accesses.append(dict(record,kind='write',complete=True,instruction=hex(uc.reg_read(UC_X86_REG_EIP))))
def invalid(uc,access,address,length,value,unused):
    global fault
    fault=dict(access=access,address=hex(address),size=length,instruction=hex(uc.reg_read(UC_X86_REG_EIP)))
    assert access in (19,20,23,24),fault
    kind='read'if access in (19,23)else'write'
    if accesses and accesses[-1]['address']==hex(address)and accesses[-1]['size']==length and accesses[-1]['kind']==kind:
        accesses[-1]['complete']=False
        if kind=='read':accesses[-1]['value']=None
    else:
        accesses.append(dict(kind=kind,address=hex(address),size=length,value=None if kind=='read'else value&((1<<(8*length))-1),complete=False,instruction=hex(uc.reg_read(UC_X86_REG_EIP))))
    return False
u.hook_add(UC_HOOK_CODE,code);u.hook_add(UC_HOOK_MEM_READ,read)
u.hook_add(UC_HOOK_MEM_WRITE,write);u.hook_add(UC_HOOK_MEM_INVALID,invalid)
results=[]
for case in cases:
    initial=bytearray(((number*23+case['seed']*31)^(number>>5))&255 for number in range(0x10000))
    def put(address,value):struct.pack_into('<I',initial,address-heap,value&0xffffffff)
    for number,address in enumerate(objects):
        put(address+8,case['owner_classes'][number]);put(address+16,0)
        put(address+496,0);put(address+504,0);put(address+508,pointer(case['owners'][number],objects));put(address+692,0)
    put(objects[0]+504,pointer(case['first'],objects))
    for number,address in enumerate(objects[2:5]):
        put(address+8,case['classes'][number]);put(address+16,case['flags'][number])
        put(address+496,pointer(case['next'][number],objects));put(address+692,pointer(case['data'][number],arrays))
    for number,address in enumerate(arrays):
        put(address,bad);put(address+4,mods[3])
        for index,role in enumerate(case['slots'][number]):put(address+8+4*index,pointer(role,mods))
    for number,address in enumerate(mods):
        put(address+88,{'none':0,'inversion':0x4e03d0,'grip':0x4e0480,'alien':bad}[case['functions'][number]])
        put(address+96,case['strengths'][number])
    u.mem_write(heap,bytes(initial))
    initial_stack=bytearray(((number*7+case['seed']*3)^(number>>6))&255 for number in range(0x10000))
    caller=struct.pack('<III',stop,pointer(case['root'],objects),pointer(case['target'],objects))
    initial_stack[sp-stack:sp-stack+len(caller)]=caller
    u.mem_write(stack,bytes(initial_stack));u.reg_write(UC_X86_REG_ESP,sp)
    u.emu_start(initialize,initialize+len(initialize_bytes),count=8)
    for register,value in saved.items():u.reg_write(register,value)
    u.reg_write(UC_X86_REG_EAX,0x98765432);u.reg_write(UC_X86_REG_ECX,0x879abcde)
    u.reg_write(UC_X86_REG_EDX,0xcdef0123);u.reg_write(UC_X86_REG_EFLAGS,0x202)
    accesses,writes,calls,returns,frames=[],[],[],[],[];fault=None;active=True
    try:u.emu_start(0x4fa4f0,stop,count=1200)
    except UcError:assert fault is not None
    finally:active=False
    assert bytes(u.mem_read(heap,0x10000))==bytes(initial),'original checker/callees must not mutate this declared external heap'
    final_stack=bytes(u.mem_read(stack,0x10000));expected_stack=bytearray(initial_stack)
    for event in writes:
        address=int(event['address'],16)
        assert stack<=address and address+event['size']<=stack+0x10000
        expected_stack[address-stack:address-stack+event['size']]=event['value'].to_bytes(event['size'],'little')
    assert bytes(expected_stack)==final_stack
    assert bytes(u.mem_read(sp,len(caller)))==caller
    assert bytes(u.mem_read(base,len(image)))==image
    assert u.reg_read(UC_X86_REG_FPCW)==options.control and u.reg_read(UC_X86_REG_FPSW)==0
    if fault is None:
        assert not frames and u.reg_read(UC_X86_REG_ESP)==sp+4
        assert all(u.reg_read(register)==value for register,value in saved.items())
        assert u.reg_read(UC_X86_REG_EAX)in (0,1)
    else:faults[fault['instruction'],fault['access']]+=1
    results.append(dict(name=case['name'],initial_heap=bytes(initial).hex(),final_heap=bytes(initial).hex(),
                        initial_stack=bytes(initial_stack).hex(),final_stack=final_stack.hex(),caller=caller.hex(),
                        return_eax=None if fault else u.reg_read(UC_X86_REG_EAX),fault=fault,
                        accesses=accesses,writes=writes,calls=calls,returns=returns,
                        final_registers={name:u.reg_read(reg)for name,reg in [
                            ('eax',UC_X86_REG_EAX),('ebx',UC_X86_REG_EBX),('ecx',UC_X86_REG_ECX),
                            ('edx',UC_X86_REG_EDX),('esi',UC_X86_REG_ESI),('edi',UC_X86_REG_EDI),
                            ('ebp',UC_X86_REG_EBP),('esp',UC_X86_REG_ESP),('eip',UC_X86_REG_EIP),('eflags',UC_X86_REG_EFLAGS)]}))
missing=sorted(set(instructions)-seen);assert not missing,[hex(address)for address in missing]
encoded=json.dumps(results,sort_keys=True,separators=(',',':')).encode()+b'\n'
packed_output=gzip.compress(encoded,mtime=0)
with options.output.open('xb')as out:out.write(packed_output)
proof=dict(original_sha256=sha(raw),declared_section_image_sha256=sha(image),original_function_ranges=ranges,
           all219_original_instruction_bytes_executed=True,decoded_instructions=len(instructions),cases=len(cases),
           actual_original_parent_and_inversion_executed=True,declared_external_services=[],
           function_calls=dict(function_calls),observed_external_accesses=sum(len(row['accesses'])for row in results),
           intentional_faults=sum(row['fault']is not None for row in results),
           fault_sites=[dict(instruction=key[0],access=key[1],count=value)for key,value in faults.items()],
           whole64KiB_external_heaps_unchanged=True,whole64KiB_stack_write_certificates=True,callers_unchanged=True,
           complete_declared_section_image_unchanged=True,FPU_control=options.control,FPU_status_unchanged=True,
           encoded_sha256=sha(packed_output),decoded_sha256=sha(encoded),
           native_initial_memory_or_implementation_credit=False,overall100=False)
with options.proof.open('x')as out:json.dump(proof,out,indent=2);out.write('\n')
print(json.dumps(proof),flush=True)
