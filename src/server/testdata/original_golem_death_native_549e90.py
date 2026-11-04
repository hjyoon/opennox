"""Execute both complete GAME.EXE MechGolem/Golem bodies with declared world services only."""
from pathlib import Path
import argparse, collections, gzip, hashlib, json, struct, os, sys
import capstone
from unicorn import Uc,UcError,UC_ARCH_X86,UC_MODE_32,UC_HOOK_CODE,UC_HOOK_MEM_WRITE,UC_HOOK_MEM_INVALID,UC_HOOK_INTR
from unicorn.x86_const import *
ap=argparse.ArgumentParser(description=__doc__)
ap.add_argument('--exe',type=Path,required=True)
ap.add_argument('--control',type=lambda s:int(s,0),choices=[0xe3f,0x23f],required=True)
ap.add_argument('--output',type=Path,required=True)
ap.add_argument('--proof',type=Path,required=True)
a=ap.parse_args();p=Path(__file__).resolve().parent
sha=lambda value:hashlib.sha256(value).hexdigest()
encoded=(p/'golem_death_independent_inputs_549e90.json.gz').read_bytes()
assert sha(encoded)=='c9c5c2b03c74b9d6bd30c7e18e55b2edee5fde4f0627d8997c4da3fef8f0a791'
inputs=json.loads(gzip.decompress(encoded));assert len(inputs)==164
raw=a.exe.read_bytes();assert sha(raw)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',raw,60)[0];op=pe+24;table=op+struct.unpack_from('<H',raw,pe+20)[0]
base=struct.unpack_from('<I',raw,op+28)[0];image_size=(struct.unpack_from('<I',raw,op+56)[0]+4095)&~4095
u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,image_size)
for i in range(struct.unpack_from('<H',raw,pe+6)[0]):
    _,va,size,off=struct.unpack_from('<IIII',raw,table+40*i+8)
    if size:u.mem_write(base+va,raw[off:off+size])
original_image=bytes(u.mem_read(base,image_size))
heap,stack,aux=0x21000000,0x22000000,0x23000000
for address in [heap,stack,aux]:u.mem_map(address,0x10000)
roles=[heap+off for off in [0x1000,0x2000,0x3000,0x4000,0x5000,0x6000,0x9000,0xa000]]
ptr=lambda role:0 if role==-1 else 0x41000000 if role==-2 else roles[role]
role=lambda value:next((n for n in [-1,-2,*range(8)] if ptr(n)==value),None)
sp,stop,init,cw,resume,double_return,double_value=[stack+0xe000,aux+0x100,aux+0x200,aux+0x300,aux+0x400,aux+0x500,aux+0x600]
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw));u.mem_write(cw,struct.pack('<H',a.control))
u.mem_write(resume,b'\x90');u.mem_write(double_return,b'\xdd\x05'+struct.pack('<I',double_value)+b'\xc3')
rd=lambda at:struct.unpack('<I',u.mem_read(at,4))[0]
cache_at,fps_at,divisor_at,mass_at,mech_at,rock_at=0x834a50,0x85b3fc,0x5cd270,0x5cd264,0x5cd4e8,0x5cd5f0
names=['MechGear','MechHead','MechArm','MechLeg','BigRock','MediumRock','SmallRock']
def plain_name(address):return bytes(u.mem_read(address,48)).split(b'\0',1)[0].decode('ascii') if address else None
stock_mech=[plain_name(rd(mech_at+4*i)) for i in range(9)]
stock_rock=[plain_name(rd(rock_at+4*i)) for i in range(12)]
stock_mass=list(u.mem_read(mass_at,12))
assert stock_mech==['MechGear','MechHead','MechArm','MechGear','MechLeg','MechLeg','MechGear','MechGear',None]
assert stock_rock==['BigRock','MediumRock','SmallRock','MediumRock','SmallRock','SmallRock','MediumRock','SmallRock','SmallRock','SmallRock','SmallRock','SmallRock']
assert stock_mass==[1,3,5,3,5,5,3,5,5,5,5,5] and rd(rock_at-4)==0x632e63
name_addresses={name:next(rd(at+4*i) for at,count in [(mech_at,8),(rock_at,12)] for i in range(count) if plain_name(rd(at+4*i))==name) for name in names}
name_ptr=lambda index:0 if index==-1 else 0x41000000 if index==-2 else name_addresses[names[index]]
services={0x40a5c0:('flags',1),0x415fa0:('int',4),0x416030:('float',4),0x4e3810:('new',1),0x4ed970:('point',3),0x4daa50:('create',5),0x4e46f0:('raise',2),0x52df80:('force',3),0x501960:('audio',4),0x522ff0:('pointfx',2),0x511660:('decay',2)}
service_names={value[0] for value in services.values()}
md=capstone.Cs(capstone.CS_ARCH_X86,capstone.CS_MODE_32);instructions={};ranges=[]
for pc,size,digest in [(0x549e90,265,'089be747744c081c29d4a9f7ff9d20c6d2505f8e806c4cbc7f725842f26e17f6'),(0x549fa0,426,'ab7579907728c6f2f94f480296cceaa0b7b2085ad40e1ba80464fa5c2fd479d7')]:
    body=bytes(u.mem_read(pc,size));assert sha(body)==digest
    seq=list(md.disasm(body,pc));assert sum(ins.size for ins in seq)==size and seq[-1].mnemonic=='ret'
    instructions.update({ins.address:bytes(ins.bytes) for ins in seq});ranges.append(dict(entry=hex(pc),bytes=size,sha256=digest))
saved={UC_X86_REG_EBX:0x12345678,UC_X86_REG_ESI:0x23456789,UC_X86_REG_EDI:0x3456789a,UC_X86_REG_EBP:0x456789ab}
registers=lambda:{key:u.reg_read(key) for key in saved}
active=False;seen=set();calls=collections.Counter();observed_faults=collections.Counter();reentries=0
current={};expected_heap=bytearray();expected_image=bytearray();events=[];frames=[];fault=None;fault_detail=None;occ={};coop=1

def wr(address,data):
    u.mem_write(address,data)
    if heap<=address and address+len(data)<=heap+0x10000:expected_heap[address-heap:address-heap+len(data)]=data
    elif base<=address and address+len(data)<=base+image_size:expected_image[address-base:address-base+len(data)]=data
    elif stack<=address and address+len(data)<=stack+0x10000:pass
    else:raise AssertionError(('undeclared world write',hex(address)))
def world_read(address,size,site):
    global fault,fault_detail
    try:return bytes(u.mem_read(address,size))
    except UcError:
        fault=dict(kind='memory',site=site)
        fault_detail=dict(kind='declared-world-read',site=site,address=hex(address),size=size)
        raise

def read_name(address):
    return world_read(address,48,'name').split(b'\0',1)[0].decode('ascii') if address else None

def origin(address):return list(struct.unpack('<II',world_read(address,8,'origin')))
def state():return dict(cache=rd(cache_at),divisor=rd(divisor_at),fps=rd(fps_at),coop=coop)
def motion(address):return [struct.unpack('<I',world_read(address+off,4,'object'))[0] for off in [16,108,116]]
def mutate(plan):
    global coop
    for key,offset in [('position',56),('flags',16)]:
        if key in plan:
            r,*values=plan[key];wr(ptr(r)+offset,struct.pack('<'+'I'*len(values),*values))
    for key,address in [('cache',cache_at),('divisor',divisor_at),('fps',fps_at)]:
        if key in plan:wr(address,struct.pack('<I',plan[key]&0xffffffff))
    if 'coop' in plan:coop=plan['coop']
    for key,address in [('mech',mech_at),('rock',rock_at)]:
        if key in plan:
            index,name=plan[key];wr(address+4*index,struct.pack('<I',name_ptr(name)))
    if 'mass' in plan:
        index,value=plan['mass'];wr(mass_at+index,bytes([value]))
def finish(frame):
    assert registers()==frame['registers'] and u.reg_read(UC_X86_REG_FPSW)&0x3800==0
    events.append(dict(kind=frame['kind']+'-return',index=frame['index'],result=frame['semantic']))
    if frame['kind']=='float':
        u.mem_write(double_value,struct.pack('<Q',frame['result']));u.reg_write(UC_X86_REG_ESP,frame['esp']);u.reg_write(UC_X86_REG_EIP,double_return)
    else:
        u.reg_write(UC_X86_REG_EAX,frame['result']&0xffffffff);u.reg_write(UC_X86_REG_ESP,frame['esp']+4);u.reg_write(UC_X86_REG_EIP,rd(frame['esp']))
def on_code(_u,pc,size,_):
    global reentries
    if not active:return
    esp=u.reg_read(UC_X86_REG_ESP)
    if pc in [double_return,double_return+6]:return
    if pc==resume:
        frame=frames.pop();assert esp==frame['esp']-4 and u.reg_read(UC_X86_REG_EAX)==1
        events.append(dict(kind='nested-return',index=frame['index'],result=1));finish(frame);return
    if pc in services:
        kind,count=services[pc];args=[rd(esp+4*i) for i in range(1,count+1)]
        at=sum(event['kind'] in service_names for event in events);seq=occ.get(kind,0);occ[kind]=seq+1;calls[kind]+=1
        plan=current['script'].get(f'{kind}:{seq}',{});event=dict(kind=kind,index=at,**state());result=0xa5a50000+at;semantic=result
        if kind=='flags':assert args==[2048];result=semantic=coop
        elif kind=='int':
            lo,hi=(struct.unpack('<i',struct.pack('<I',value))[0] for value in args[:2]);event.update(min=lo,max=hi)
            result=semantic=current['parts'] if args[3]==0x322 else current['seconds']
        elif kind=='float':event.update(min=args[0],max=args[1]);result=semantic=current['float_bits']
        elif kind=='new':
            name=read_name(args[0]);r=-1 if plan.get('return_nil') else current['new'].get(name,-1)
            event.update(name=name);result=ptr(r);semantic=r
        elif kind=='point':
            event.update(radius=args[0],origin=origin(args[1]));wr(args[2],struct.pack('<II',0x42c80000+seq,0x43480000+seq))
        elif kind=='create':
            assert role(args[0]) is not None and role(args[1])==-1 and args[4]==0
            event.update(object=role(args[0]),owner=-1,position=args[2:4],extra=0)
        elif kind=='raise':assert role(args[0]) is not None;event.update(object=role(args[0]),height=args[1])
        elif kind=='force':
            assert role(args[1]) is not None;event.update(object=role(args[1]),origin=origin(args[0]),force=args[2],motion=motion(args[1]))
        elif kind=='audio':assert args[0] in [494,487] and args[2:]==[0,0];event.update(sound=args[0],unit=role(args[1]))
        elif kind=='pointfx':assert args[0]==0x8a;event.update(effect=args[0],origin_pointer=role((args[1]-56)&0xffffffff))
        elif kind=='decay':assert role(args[0]) is not None;event.update(object=role(args[0]),duration=args[1],motion=motion(args[0]))
        events.append(event);mutate(plan)
        frame=dict(kind=kind,index=at,esp=esp,result=result,semantic=semantic,registers=registers())
        if 'reenter' in plan:
            entry,unit=plan['reenter'];frames.append(frame);reentries+=1
            events.append(dict(kind='nested',index=at,entry=entry,unit=unit))
            u.mem_write(esp-8,struct.pack('<II',resume,ptr(unit)));u.reg_write(UC_X86_REG_ESP,esp-8);u.reg_write(UC_X86_REG_EIP,int(entry,16))
        else:finish(frame)
        return
    assert pc in instructions and bytes(u.mem_read(pc,size))==instructions[pc],(current['name'],hex(pc),size)
    seen.add(pc)
def on_write(_u,access,address,size,value,_):
    if not active:return
    if stack<=address and address+size<=stack+0x10000:
        if address>=sp and address<sp+len(caller):
            assert u.reg_read(UC_X86_REG_EIP) in [0x549fee,0x549ff4] and address==sp+4 and size==4 and current['entry']=='549fa0'
            caller_expected[4:8]=value.to_bytes(4,'little')
        return
    pc=u.reg_read(UC_X86_REG_EIP)
    stores={0x549f4b:108,0x549f5d:116,0x549f64:16,0x54a08a:108,0x54a0c0:16,0x54a0c3:116}
    if pc in stores:assert size==4 and role(address-stores[pc]) is not None
    else:assert pc==0x54a132 and address==cache_at and size==4,(current['name'],'undeclared original write',hex(pc),hex(address),size)
    if not (heap<=address and address+size<=heap+0x10000 or base<=address and address+size<=base+image_size):return
    dest=expected_heap if heap<=address<heap+0x10000 else expected_image
    offset=address-(heap if dest is expected_heap else base);dest[offset:offset+size]=(value&((1<<(8*size))-1)).to_bytes(size,'little')
def on_fault(_u,access,address,size,value,_):
    global fault,fault_detail
    pc=u.reg_read(UC_X86_REG_EIP)
    site='rock-table' if pc==0x54a010 else 'mass-table' if pc==0x54a09f else 'object'
    fault=dict(kind='memory',site=site);fault_detail=dict(kind='original-memory-access',pc=hex(pc),access=access,address=hex(address),size=size)
    return False
def on_intr(_u,number,_):
    global fault,fault_detail
    assert number==0 and u.reg_read(UC_X86_REG_EIP)==0x54a125,(current['name'],number,hex(u.reg_read(UC_X86_REG_EIP)))
    fault=dict(kind='arithmetic',site='rock-modulo');fault_detail=dict(kind='original-cpu-divide-error',pc='0x54a125',interrupt=number);u.emu_stop()
original_aux=bytes(u.mem_read(aux,0x10000));outputs=[];details=[]
for current in inputs:
    if os.environ.get('GOLEM_ORACLE_TRACE') == '1':print(current['name'],file=sys.stderr,flush=True)
    active=False;u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,image_size)
    for address in [heap,stack,aux]:u.mem_map(address,0x10000)
    u.mem_write(aux,original_aux);u.mem_write(base,original_image);expected_image=bytearray(original_image)
    for hook,callback in [(UC_HOOK_CODE,on_code),(UC_HOOK_MEM_WRITE,on_write),(UC_HOOK_MEM_INVALID,on_fault),(UC_HOOK_INTR,on_intr)]:u.hook_add(hook,callback)
    expected_heap=bytearray((i*17+(i>>7)*29)&255 for i in range(0x10000));u.mem_write(heap,bytes(expected_heap))
    for n in range(6):
        for offset,values in [(16,[0x11223344+n]),(56,[0x40e00000+n,0x41000000+n])]:wr(roles[n]+offset,struct.pack('<'+'I'*len(values),*values))
    for key,address in [('cache',cache_at),('divisor',divisor_at),('fps',fps_at)]:wr(address,struct.pack('<I',current[key]&0xffffffff))
    assert 0<=current['mech_stop']<=8
    wr(mech_at+4*current['mech_stop'],struct.pack('<I',0))
    coop=current['coop'];u.mem_write(stack,b'\xcc'*0x10000)
    caller=struct.pack('<II',stop,ptr(current['unit']))+bytes(range(48));u.mem_write(sp,caller);caller_expected=bytearray(caller)
    for key,value in saved.items():u.reg_write(key,value)
    u.reg_write(UC_X86_REG_ESP,sp);u.emu_start(init,init+8,count=2)
    events=[];frames=[];fault=None;fault_detail=None;occ={};active=True
    try:u.emu_start(int(current['entry'],16),stop,count=100000)
    except UcError:assert fault is not None,current['name']
    active=False
    assert bytes(u.mem_read(heap,0x10000))==expected_heap,(current['name'],'complete heap')
    assert bytes(u.mem_read(base,image_size))==expected_image,(current['name'],'complete image')
    assert bytes(u.mem_read(sp,len(caller)))==caller_expected,(current['name'],'caller guard')
    assert u.reg_read(UC_X86_REG_FPCW)==a.control
    if not fault:
        assert not frames and u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4 and registers()==saved,(current['name'],'return guard')
        assert u.reg_read(UC_X86_REG_FPSW)&0x3800==0 and u.reg_read(UC_X86_REG_EAX)==1
    if fault:observed_faults[fault['kind']+':'+fault['site']]+=1;details.append(dict(name=current['name'],**fault_detail))
    outputs.append(dict(name=current['name'],result=None if fault else 1,fault=fault,events=events,**state(),heap_sha256=sha(expected_heap),
                        tables=dict(mech=[plain_name(rd(mech_at+4*i)) for i in range(9)],rock=[plain_name(rd(rock_at+4*i)) for i in range(12)],mass=list(u.mem_read(mass_at,12)))))
missing=[hex(pc) for pc in instructions.keys()-seen];assert not missing,('missing original instructions',missing)
decoded=json.dumps(outputs,sort_keys=True,separators=(',',':')).encode()+b'\n';packed=bytearray(gzip.compress(decoded,compresslevel=9,mtime=0));packed[9]=0xff
with a.output.open('xb') as out:out.write(packed)
proof=dict(original_exe_sha256=sha(raw),input_sha256=sha(encoded),cases=len(inputs),control=hex(a.control),ranges=ranges,
           all691_instruction_bytes_executed=True,complete_heap_and_image_guards=True,caller_guard_except_declared_reused_Golem_argument=True,
           valid_return_stack_registers_preserved=True,FPU_CW_and_valid_TOP_preserved=True,original_faults=dict(observed_faults),fault_details=details,
           only_declared_world_services_replaced={hex(key):value[0] for key,value in services.items()},original_reentrant_calls_executed=reentries,
           stock_mech=stock_mech,stock_rock=stock_rock,stock_mass=stock_mass,stock_rock_preceding_word='0x632e63',
           fixture_only_declared_live_tables_and_globals_mutated=True,world_calls=dict(calls),output_sha256=sha(packed),decoded_output_sha256=sha(decoded),native_source_used=False)
with a.proof.open('x') as out:json.dump(proof,out,indent=2);out.write('\n')
print(json.dumps(proof),flush=True)
