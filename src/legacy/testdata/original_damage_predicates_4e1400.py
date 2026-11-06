"""Execute both complete original damage predicates, without helper substitution.

Both bodies are leaves. All146 sealed bytes execute; no external body or
ordinary native alias is used. The only optional writes are independently
declared after-read input stimuli, applied at the next instruction.
"""
from pathlib import Path
import argparse, collections, gc, gzip, hashlib, json, struct
import capstone
from unicorn import Uc, UcError, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_READ, UC_HOOK_MEM_WRITE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import *

p = argparse.ArgumentParser(description=__doc__)
for name in ['exe', 'manifest', 'inputs', 'output', 'proof']:
    p.add_argument('--'+name, type=Path, required=True)
p.add_argument('--control', type=lambda v:int(v,0), choices=[0xe3f,0x23f], required=True)
a = p.parse_args()
sha = lambda data:hashlib.sha256(data).hexdigest()
raw = a.exe.read_bytes()
assert sha(raw) == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
manifest_bytes = a.manifest.read_bytes()
manifest = json.loads(manifest_bytes)
encoded = a.inputs.read_bytes()
cases = json.loads(gzip.decompress(encoded))
assert len(cases) == 487 and len({r['name'] for r in cases}) == len(cases)
assert sha(encoded) == '1f96b7f8b86e8fd9f37cd9ee8c71993b8f1c9a2093ca0e5ba9afb0586f6fdac6'
pe = struct.unpack_from('<I',raw,0x3c)[0]
opt = pe+24
base = struct.unpack_from('<I',raw,opt+28)[0]
image = bytearray((struct.unpack_from('<I',raw,opt+56)[0]+4095)&~4095)
table = opt+struct.unpack_from('<H',raw,pe+20)[0]
for i in range(struct.unpack_from('<H',raw,pe+6)[0]):
    _,rva,length,disk = struct.unpack_from('<IIII',raw,table+40*i+8)
    image[rva:rva+length] = raw[disk:disk+length]
image = bytes(image)
roots = {
    0x4e1400:('attack',112,2,'0ac2ef45c5c13d020b99663c9fcac15321cdf3b825cabb3b44c746aae5a4fb73'),
    0x4e1470:('friendly',34,1,'fa632520ae8b970b679dd6c5b8bd947f81f4b41d65295f6d3a26edbcb975b92e'),
}
entry_address = {r[0]:address for address,r in roots.items()}
md = capstone.Cs(capstone.CS_ARCH_X86,capstone.CS_MODE_32)
instructions, ranges = {}, []
for address,(name,length,argc,digest) in roots.items():
    body = image[address-base:address-base+length]
    assert sha(body)==digest
    seal = next(r for r in manifest['functions'] if int(r['address'],16)==address)
    assert seal['size']==length and seal['sha256']==digest
    decoded = list(md.disasm(body,address))
    assert sum(i.size for i in decoded)==length and decoded[-1].mnemonic=='ret'
    assert not any(i.mnemonic.startswith('call') for i in decoded), 'report noncompact dependency before execution'
    instructions.update({i.address:(bytes(i.bytes),i.mnemonic,name) for i in decoded})
    ranges.append(dict(name=name,address=hex(address),size=length,sha256=digest))

heap,stack,aux,bad = 0x21000000,0x22000000,0x23000000,0x41000000
sp,stop,initialize,cw = stack+0xe000,aux+0x100,aux+0x200,aux+0x300
init_bytes = b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw)
saved = {UC_X86_REG_EBX:0x12345678, UC_X86_REG_ESI:0x23456789, UC_X86_REG_EDI:0x3456789a, UC_X86_REG_EBP:0x456789ab}
locations = {
    0x4e140c:('source','class',8,4), 0x4e141d:('source','subclass',12,1),
    0x4e142c:('weapon','class',8,4), 0x4e1434:('weapon','subclass',12,4),
    0x4e143d:('weapon','use',736,4), 0x4e1443:('use','flags',96,1),
    0x4e1456:('weapon','subclass',12,4),
    0x4e1478:('weapon','class',8,4), 0x4e1481:('weapon','subclass',12,4),
}
field_offsets = {('source','class'):(8,4), ('source','subclass'):(12,1),
    ('weapon','class'):(8,4), ('weapon','subclass'):(12,4),
    ('weapon','use'):(736,4), ('use','flags'):(96,1)}
seen = set()
fault_domains, counts = collections.Counter(), collections.Counter()
active,u,current,roles = False,None,{},{}
accesses,stack_reads,writes,calls,returns,pending = [],[],[],[],[],[]
read_counts,applied,fault = collections.Counter(),set(),None

def stimulus(index):
    change = current['mutations'][index]
    role,field = change['role'],change['field']
    offset,width = field_offsets[role,field]
    address = roles[role]+offset
    value = change['value']
    if isinstance(value,str): value = bad if value=='bad' else 0 if value=='null' else roles[value]
    value &= (1<<(8*width))-1
    assert heap<=address and address+width<=heap+0x10000
    u.mem_write(address,value.to_bytes(width,'little'))
    record = dict(kind='stimulus',role=role,field=field,address=hex(address),size=width,value=value,
        complete=True,source_input_mutation=index,instruction=hex(u.reg_read(UC_X86_REG_EIP)))
    accesses.append(record);writes.append(record.copy());applied.add(index)

def code(uc,address,length,unused):
    if not active:return
    while pending:stimulus(pending.pop(0))
    assert address in instructions and bytes(uc.mem_read(address,length))==instructions[address][0],hex(address)
    seen.add(address)
    if address==entry_address[current['entry']]:
        argc=roots[address][2]
        args=list(struct.unpack('<'+'I'*argc,uc.mem_read(uc.reg_read(UC_X86_REG_ESP)+4,4*argc)))
        assert not calls
        calls.append(dict(function=current['entry'],address=hex(address),args=args,depth=0,external=False))
    if instructions[address][1]=='ret':
        returns.append(dict(function=current['entry'],result=uc.reg_read(UC_X86_REG_EAX),depth=0,external=False))

def read(uc,access,address,width,unused_value,unused):
    if not active:return
    instruction=uc.reg_read(UC_X86_REG_EIP)
    if stack<=address and address+width<=stack+0x10000:
        stack_reads.append(dict(kind='read',address=hex(address),size=width,
            value=int.from_bytes(uc.mem_read(address,width),'little'),complete=True,instruction=hex(instruction)))
        return
    assert instruction in locations,hex(instruction)
    role,field,offset,expected_width=locations[instruction]
    assert width==expected_width and address==roles[role]+offset,(current['name'],hex(instruction),hex(address),width)
    try:value=int.from_bytes(uc.mem_read(address,width),'little')
    except UcError:value=None
    accesses.append(dict(kind='read',role=role,field=field,address=hex(address),size=width,
        value=value,complete=value is not None,instruction=hex(instruction)))
    read_counts[role,field]+=1
    if value is not None:
        for index,change in enumerate(current['mutations']):
            if change['role']==role and change['field']==field and change['occurrence']==read_counts[role,field]:pending.append(index)

def write(uc,access,address,width,value,unused):
    if active:raise AssertionError('sealed leaf instructions must not write any heap/stack/image byte')

def invalid(uc,access,address,width,value,unused):
    global fault
    instruction=uc.reg_read(UC_X86_REG_EIP)
    assert instruction in locations and access in (19,23),(hex(instruction),access)
    role,field,offset,expected_width=locations[instruction]
    assert width==expected_width and address==roles[role]+offset
    fault=dict(kind='memory',access=access,address=hex(address),size=width,instruction=hex(instruction))
    if accesses and accesses[-1]['kind']=='read' and accesses[-1]['address']==hex(address):accesses[-1].update(value=None,complete=False)
    else:accesses.append(dict(kind='read',role=role,field=field,address=hex(address),size=width,value=None,complete=False,instruction=hex(instruction)))
    return False

def pointer(kind,role):
    valid=heap+{'source':0x1000,'weapon':0x3000,'use':0x5000}[role]
    if kind=='null':return 0
    if kind=='bad':return bad
    if kind=='last-class':return heap+0x10000-12
    if kind=='last-subclass-byte':return heap+0x10000-13
    if kind=='last-subclass':return heap+0x10000-16
    if kind=='last-use-pointer':return heap+0x10000-740
    if kind=='last-flags-byte':return heap+0x10000-97
    if kind.startswith('unaligned'):return valid+int(kind.removeprefix('unaligned'))
    assert kind=='valid',kind
    return valid

results=[]
for current in cases:
    u=Uc(UC_ARCH_X86,UC_MODE_32)
    for address,length in [(base,len(image)),(heap,0x10000),(stack,0x10000),(aux,0x1000)]:u.mem_map(address,length)
    u.mem_write(base,image);u.mem_write(initialize,init_bytes);u.mem_write(cw,struct.pack('<H',a.control))
    u.hook_add(UC_HOOK_CODE,code);u.hook_add(UC_HOOK_MEM_READ,read)
    u.hook_add(UC_HOOK_MEM_WRITE,write);u.hook_add(UC_HOOK_MEM_INVALID,invalid)
    roles={'weapon':pointer(current['weapon'],'weapon'),'use':pointer(current['use'],'use')}
    roles['source']=roles['weapon'] if current['source']=='weapon' else pointer(current['source'],'source')
    initial=bytearray(((i*37+current['seed']*19)^(i>>3))&255 for i in range(0x10000))
    initialized=[]
    for role,field,width,value in [('source','class',4,current['source_class']),('source','subclass',4,current['source_subclass']),
            ('weapon','class',4,current['weapon_class']),('weapon','subclass',4,current['weapon_subclass']),
            ('weapon','use',4,roles['use']),('use','flags',4,current['use_flags'])]:
        address=roles[role]+field_offsets[role,field][0]
        if not(heap<=address and address+width<=heap+0x10000):
            if role in ['source','use'] and field in ['subclass','flags'] and heap<=address<heap+0x10000:width=1
            else:continue
        value &= (1<<(8*width))-1
        initial[address-heap:address-heap+width]=value.to_bytes(width,'little')
        initialized.append(dict(role=role,field=field,address=hex(address),size=width,value=value))
    u.mem_write(heap,bytes(initial))
    initial_stack=bytearray(((i*11+current['seed'])^(i>>4))&255 for i in range(0x10000))
    args=[roles['source'],roles['weapon']] if current['entry']=='attack' else [roles['weapon']]
    caller=struct.pack('<'+'I'*(len(args)+1),stop,*args)
    initial_stack[sp-stack:sp-stack+len(caller)]=caller;u.mem_write(stack,bytes(initial_stack))
    u.reg_write(UC_X86_REG_ESP,sp);u.emu_start(initialize,initialize+len(init_bytes),count=8)
    for reg,value in saved.items():u.reg_write(reg,value)
    for reg,value in [(UC_X86_REG_EAX,0x98765432),(UC_X86_REG_ECX,0x87654321),(UC_X86_REG_EDX,0x76543210),(UC_X86_REG_EFLAGS,0x202)]:u.reg_write(reg,value)
    accesses,stack_reads,writes,calls,returns,pending=[],[],[],[],[],[]
    read_counts,applied,fault=collections.Counter(),set(),None
    active=True
    try:u.emu_start(entry_address[current['entry']],stop,count=256)
    except UcError:assert fault is not None
    finally:active=False
    assert not pending and applied==set(range(len(current['mutations'])))
    final_heap=bytes(u.mem_read(heap,0x10000));final_stack=bytes(u.mem_read(stack,0x10000))
    expected=bytearray(initial)
    for record in writes:
        address=int(record['address'],16)-heap
        expected[address:address+record['size']]=record['value'].to_bytes(record['size'],'little')
    assert final_heap==bytes(expected) and final_stack==bytes(initial_stack) and bytes(u.mem_read(base,len(image)))==image
    assert bytes(u.mem_read(sp,len(caller)))==caller and len(calls)==1 and calls[0]['args']==args
    assert u.reg_read(UC_X86_REG_FPCW)==a.control and u.reg_read(UC_X86_REG_FPSW)==0
    registers={name:u.reg_read(reg) for name,reg in [('eax',UC_X86_REG_EAX),('ecx',UC_X86_REG_ECX),('edx',UC_X86_REG_EDX),('ebx',UC_X86_REG_EBX),('esi',UC_X86_REG_ESI),('edi',UC_X86_REG_EDI),('ebp',UC_X86_REG_EBP),('esp',UC_X86_REG_ESP),('eip',UC_X86_REG_EIP),('eflags',UC_X86_REG_EFLAGS)]}
    if fault is None:
        assert len(returns)==1 and returns[0]['result']==registers['eax'] and registers['eax']in(0,1)
        assert registers['esp']==sp+4 and registers['eip']==stop and all(u.reg_read(reg)==value for reg,value in saved.items())
        counts[current['entry'],'return-'+str(registers['eax'])]+=1
    else:
        assert not returns and accesses[-1]['complete']is False and accesses[-1]['instruction']==fault['instruction']
        fault_domains[current['entry'],fault['instruction']]+=1
    results.append(dict(name=current['name'],entry=current['entry'],roles=roles,initializers=initialized,
        initial_heap=bytes(initial).hex(),final_heap=final_heap.hex(),initial_stack=bytes(initial_stack).hex(),final_stack=final_stack.hex(),
        initial_image_sha256=sha(image),final_image_sha256=sha(bytes(u.mem_read(base,len(image)))),
        caller_before=caller.hex(),caller_after=bytes(u.mem_read(sp,len(caller))).hex(),registers=registers,
        accesses=accesses,stack_reads=stack_reads,writes=writes,calls=calls,returns=returns,
        applied_input_mutations=sorted(applied),fault=fault))
    u=None;gc.collect()
missing=set(instructions)-seen
assert not missing,[hex(v)for v in sorted(missing)]
assert sha(a.exe.read_bytes())==sha(raw) and a.manifest.read_bytes()==manifest_bytes and a.inputs.read_bytes()==encoded
data=json.dumps(results,sort_keys=True,separators=(',',':')).encode()+b'\n'
with a.output.open('xb')as out:out.write(gzip.compress(data,mtime=0))
proof=dict(original_exe_sha256=sha(raw),original_image_sha256=sha(image),original_file_unchanged=True,
    original_manifest_sha256=sha(manifest_bytes),original_manifest_unchanged=True,independent_inputs_sha256=sha(encoded),
    cases=len(cases),cases_per_entry={e:sum(r['entry']==e for r in cases)for e in ['attack','friendly']},
    original_function_ranges=ranges,all146_original_bytes_reached=True,instruction_addresses=len(instructions),
    external_service_boundaries=[],actual_callee_bodies_required=0,no_helpers_substituted=True,
    full64KiB_heap_stack_before_after_retained=True,complete_ordered_heap_and_stack_read_values_widths_retained=True,
    complete_caller_call_return_fault_register_certificates=True,actual_original_memory_writes=0,
    declared_input_stimulus_writes=sum(len(r['writes'])for r in results),all_writes_reconstruct_whole_heap=True,
    complete_stack_and_caller_words_unchanged=True,whole_original_PE_image_unchanged=True,
    nonvolatile_registers_stack_and_canonical_int_return_verified=True,fresh_emulator_per_input=True,
    FPU_control=a.control,FPU_status_unchanged=True,
    external_heap_access_records=sum(len(r['accesses'])for r in results),stack_read_records=sum(len(r['stack_reads'])for r in results),
    returns=[dict(entry=e,kind=k,count=v)for(e,k),v in sorted(counts.items())],
    memory_faults=[dict(entry=e,instruction=i,count=v)for(e,i),v in sorted(fault_domains.items())],
    compressed_output_sha256=sha(a.output.read_bytes()),decoded_output_sha256=sha(data),
    native_initial_memory_credit=False,native_implementation_credit=False,actual_product_credit=False,overall100=False)
with a.proof.open('x')as out:json.dump(proof,out,indent=2);out.write('\n')
print(json.dumps(proof),flush=True)
