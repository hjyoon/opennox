"""Original root/search/callback/flags/poison; spatial/visibility/Push are services."""
from pathlib import Path
import argparse, collections, datetime, gzip, hashlib, json, struct
import capstone
from unicorn import Uc,UcError,UC_ARCH_X86,UC_MODE_32,UC_HOOK_CODE,UC_HOOK_MEM_WRITE,UC_HOOK_MEM_INVALID
from unicorn.x86_const import *

p=Path(__file__).resolve().parent
ap=argparse.ArgumentParser(description=__doc__)
ap.add_argument('--exe',type=Path,required=True)
ap.add_argument('--control',type=lambda x:int(x,0),choices=[0xe3f,0x23f],required=True)
ap.add_argument('--output',type=Path,required=True);ap.add_argument('--proof',type=Path,required=True)
a=ap.parse_args();sha=lambda b:hashlib.sha256(b).hexdigest()
encoded=(p/'monster_edibles_independent_inputs_544a00.json.gz').read_bytes()
assert sha(encoded)=='245b01ed1e35249a6e9adb171c560c6ecaf311479fbc28622ad3fd6f4bb2a483'
inputs=json.loads(gzip.decompress(encoded))
blob=a.exe.read_bytes()
assert sha(blob)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,60)[0];op=pe+24;table=op+struct.unpack_from('<H',blob,pe+20)[0]
base=struct.unpack_from('<I',blob,op+28)[0];image_size=(struct.unpack_from('<I',blob,op+56)[0]+4095)&~4095
u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,image_size)
for i in range(struct.unpack_from('<H',blob,pe+6)[0]):
    _,va,n,off=struct.unpack_from('<IIII',blob,table+40*i+8)
    if n:u.mem_write(base+va,blob[off:off+n])
image_before=bytes(u.mem_read(base,image_size))
heap,stack,aux=0x21000000,0x22000000,0x23000000
for ptr in [heap,stack,aux]:u.mem_map(ptr,0x10000)
objects=[heap+i*0x1000 for i in range(1,6)];items=[heap+i*0x1000 for i in range(8,12)]
sp,stop,init,cw=stack+0xe000,aux+0x100,aux+0x200,aux+0x300
resume_enum,resume_nested=aux+0x110,aux+0x120
u.mem_write(resume_enum,b'\x90');u.mem_write(resume_nested,b'\x90')
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw));u.mem_write(cw,struct.pack('<H',a.control))
flags,best,chosen=0x5d53a4,0x8341f8,0x834200
globals_original={ptr:bytes(u.mem_read(ptr,4))for ptr in [flags,best,chosen]}
regions={0x5455e0:197,0x544a00:53,0x544a40:153,0x40a5c0:18,0x5347f0:18}
md=capstone.Cs(capstone.CS_ARCH_X86,capstone.CS_MODE_32);instructions={};code=[]
for pc,n in regions.items():
    raw=bytes(u.mem_read(pc,n));seq=list(md.disasm(raw,pc));assert sum(i.size for i in seq)==n
    instructions.update({i.address:bytes(i.bytes)for i in seq});code.append(dict(entry=hex(pc),bytes=n,sha256=sha(raw)))
assert sum(regions.values())==439
saved={UC_X86_REG_EBX:0x12345678,UC_X86_REG_ESI:0x23456789,UC_X86_REG_EDI:0x3456789a,UC_X86_REG_EBP:0x456789ab}
rd=lambda ptr:struct.unpack('<I',u.mem_read(ptr,4))[0]
role=lambda value:objects.index(value)if value in objects else-1
objptr=lambda index:0 if index<0 else objects[index]
itemptr=lambda index:0 if index<0 else items[index]
canonical=lambda value:0xf0000000+objects.index(value)if value in objects else value
decode_word=lambda value:objects[value-0xf0000000]if 0xf0000000<=value<0xf0000000+len(objects)else value
active=False;row={};events=[];writes=[];faults=[];frames=[];seen=set();visited=set();calls=collections.Counter()
expected=bytearray();global_expected={};enum_index=interaction_index=push_index=0

def write(ptr,value,n=4):
    raw=int(value&((1<<(n*8))-1)).to_bytes(n,'little');u.mem_write(ptr,raw)
    if heap<=ptr<ptr+n<=heap+0x10000:expected[ptr-heap:ptr-heap+n]=raw
    elif ptr in globals_original and n==4:global_expected[ptr]=raw
    else:raise AssertionError(('undeclared service write',hex(ptr),n))
def state():
    return dict(objects=[dict(cls=rd(v+8),sub=rd(v+12),poison=bytes(u.mem_read(v+540,1))[0],pos=[rd(v+56),rd(v+60)])for v in objects],
        flags=rd(flags),best=rd(best),chosen=role(rd(chosen)),items=[[canonical(rd(v+i*4))for i in range(6)]for v in items])
def mutate(m):
    for change in m.get('objects',[]):
        v=objects[change['id']]
        for key,off,n in [('cls',8,4),('sub',12,4),('poison',540,1)]:
            if key in change:write(v+off,change[key],n)
        if 'pos'in change:write(v+56,change['pos'][0]);write(v+60,change['pos'][1])
    for key,ptr in [('flags',flags),('best',best)]:
        if key in m:write(ptr,m[key])
    if 'chosen'in m:write(chosen,objptr(m['chosen']))
    for change in m.get('items',[]):
        for index,value in enumerate(change['words']):write(items[change['id']]+4*index,decode_word(value))
def regs():return {reg:u.reg_read(reg)for reg in saved}
def check_regs(values):
    assert all(u.reg_read(reg)==value for reg,value in values.items()),(row['name'],'service/callback nonvolatile registers')
def ret(esp,value):
    u.reg_write(UC_X86_REG_EAX,value&0xffffffff);u.reg_write(UC_X86_REG_ESP,esp+4);u.reg_write(UC_X86_REG_EIP,rd(esp))
def schedule_callback(frame):
    index=frame['plan']['candidates'][frame['index']]
    frame['callback_candidate']=index
    esp=frame['esp']-12;u.mem_write(esp,struct.pack('<III',resume_enum,objptr(index),frame['unit']))
    u.reg_write(UC_X86_REG_ESP,esp);u.reg_write(UC_X86_REG_EIP,0x544a40)
def finish_enum(frame):
    mutate(frame['plan'].get('after',{}));events.append(dict(kind='enum-return',index=frame['number'],after=state()))
    check_regs(frame['registers']);frames.pop();ret(frame['esp'],0x76543210)
def finish_service(frame):
    kind,number,step=frame['kind'],frame['number'],frame['step']
    mutate(step.get('after_nested',{}))
    if kind=='push':
        value=itemptr(step['item'])
        if value:write(value,frame['action'])
        events.append(dict(kind='push-return',index=number,item=step['item'],after=state()))
    else:
        value=step.get('accept',1)
        events.append(dict(kind='interact-return',index=number,value=value,after=state()))
    check_regs(frame['registers']);frames.pop();ret(frame['esp'],value)
def begin_service(frame):
    mutate(frame['step'].get('mutation',{}))
    nested=frame['step'].get('reenter')
    if not nested:finish_service(frame);return
    esp=frame['esp']-12
    u.mem_write(esp,struct.pack('<III',resume_nested,objptr(nested['unit']),nested['radius']))
    events.append(dict(kind='nested-search',service=frame['kind'],index=frame['number'],unit=nested['unit'],radius=nested['radius'],before=state()))
    u.reg_write(UC_X86_REG_ESP,esp);u.reg_write(UC_X86_REG_EIP,0x544a00)

def on_write(_u,_access,ptr,n,value,_):
    if not active:return
    pc=u.reg_read(UC_X86_REG_EIP)
    if stack<=ptr<stack+0x10000:
        assert sp-8192<=ptr<ptr+n<=sp,(row['name'],hex(pc),hex(ptr),n);return
    if ptr in globals_original:
        assert n==4 and (pc,ptr)in [(0x544a13,chosen),(0x544a1d,best),(0x544ac5,best),(0x544acb,chosen)],(row['name'],hex(pc),hex(ptr))
        global_expected[ptr]=(value&0xffffffff).to_bytes(4,'little')
        writes.append(dict(after_event=len(events),pc=hex(pc),target='chosen'if ptr==chosen else'best',size=n,value=canonical(value&0xffffffff)if ptr==chosen else value&0xffffffff))
        return
    index=next((i for i,base in enumerate(items)if base<=ptr<ptr+n<=base+24),None)
    allowed={0x545634:(1,4),0x54563a:(2,4),0x54563d:(3,4),0x54564f:(1,4),0x545664:(1,4),0x54566a:(3,4),0x54566e:(2,4),0x545697:(1,4),0x54569e:(3,1)}
    assert index is not None and pc in allowed,(row['name'],hex(pc),hex(ptr),n)
    field=(ptr-items[index])//4;assert (field,n)==allowed[pc]
    expected[ptr-heap:ptr-heap+n]=int(value&((1<<(8*n))-1)).to_bytes(n,'little')
    pointer=pc in [0x54563d,0x54564f,0x54566a]
    writes.append(dict(after_event=len(events),pc=hex(pc),target='item',item=index,field=field,size=n,value=canonical(value&0xffffffff)if pointer else value&((1<<(8*n))-1)))
def on_invalid(_u,access,ptr,n,_value,_):
    faults.append(dict(kind='original-memory',access=access,address=hex(ptr),bytes=n,pc=hex(u.reg_read(UC_X86_REG_EIP))))
    return False
def on_code(_u,pc,n,_):
    global enum_index,interaction_index,push_index
    if not active:return
    esp=u.reg_read(UC_X86_REG_ESP)
    if pc==resume_enum:
        frame=frames[-1];assert frame['kind']=='enum'and esp==frame['esp']-8
        check_regs(frame['registers']);assert u.reg_read(UC_X86_REG_FPSW)&0x3800==0
        u.reg_write(UC_X86_REG_ESP,frame['esp'])
        events.append(dict(kind='callback-return',enum=frame['number'],index=frame['index'],candidate=frame['callback_candidate'],after=state()))
        mutations=frame['plan'].get('after_each',[])
        if frame['index']<len(mutations):mutate(mutations[frame['index']])
        frame['index']+=1
        if frame['index']<len(frame['plan']['candidates']):schedule_callback(frame)
        else:finish_enum(frame)
        return
    if pc==resume_nested:
        frame=frames[-1];assert frame['kind']in ['push','interact']and esp==frame['esp']-8
        events.append(dict(kind='nested-return',service=frame['kind'],index=frame['number'],result=role(u.reg_read(UC_X86_REG_EAX)),after=state()))
        u.reg_write(UC_X86_REG_ESP,frame['esp']);finish_service(frame);return
    if pc in [0x517f90,0x5370e0,0x50a260]:
        if pc==0x517f90:
            point,radius,callback,unit=[rd(esp+i*4)for i in range(1,5)]
            assert callback==0x544a40 and point==unit+56
            try:position=[rd(point),rd(point+4)]
            except UcError:
                faults.append(dict(kind='declared-spatial-point-admission',address=hex(point),bytes=8,pc=hex(pc)));u.emu_stop();return
            number=enum_index;enum_index+=1
            plan=row['enums'][number]if number<len(row['enums'])else dict(candidates=[])
            frame=dict(kind='enum',esp=esp,unit=unit,number=number,index=0,plan=plan,registers=regs());frames.append(frame)
            events.append(dict(kind='enum',index=number,unit=role(unit),radius=radius,position=position,before=state()))
            mutate(plan.get('before',{}))
            if plan['candidates']:schedule_callback(frame)
            else:finish_enum(frame)
        elif pc==0x5370e0:
            unit,candidate,options=[rd(esp+i*4)for i in range(1,4)];assert options==0
            number=interaction_index;interaction_index+=1
            step=row['interactions'][number]if number<len(row['interactions'])else dict(accept=1)
            frame=dict(kind='interact',esp=esp,number=number,step=step,registers=regs());frames.append(frame)
            events.append(dict(kind='interact',index=number,unit=role(unit),candidate=role(candidate),options=options,before=state()));begin_service(frame)
        else:
            unit,action=rd(esp+4),rd(esp+8);assert unit==objptr(row['unit'])
            number=push_index;push_index+=1
            step=row['pushes'][number]if number<len(row['pushes'])else dict(item=number%4)
            frame=dict(kind='push',esp=esp,number=number,step=step,action=action,registers=regs());frames.append(frame)
            events.append(dict(kind='push',index=number,unit=role(unit),action=action,before=state()));begin_service(frame)
        return
    assert pc in instructions and bytes(u.mem_read(pc,n))==instructions[pc],(row['name'],hex(pc),n)
    seen.add(pc)
    if pc in regions:calls[hex(pc)]+=1
    if pc==0x40a5c0:events.append(dict(kind='flags',mask=rd(esp+4),before=state()))

u.hook_add(UC_HOOK_CODE,on_code);u.hook_add(UC_HOOK_MEM_WRITE,on_write);u.hook_add(UC_HOOK_MEM_INVALID,on_invalid)
seed=bytes((i*37+0xa5)&255 for i in range(0x10000));outputs=[];fault_count=0
for row in inputs['rows']:
    active=False;events=[];writes=[];faults=[];frames=[];seen=set();calls=collections.Counter()
    enum_index=interaction_index=push_index=0
    u.mem_write(heap,seed);expected=bytearray(seed);u.mem_write(stack,b'\x5a'*0x10000)
    global_expected={ptr:bytes(u.mem_read(ptr,4))for ptr in globals_original}
    for ptr,values in zip(objects,row['objects']):
        write(ptr+8,values['cls']);write(ptr+12,values['sub']);write(ptr+540,values['poison'],1)
        write(ptr+56,values['pos'][0]);write(ptr+60,values['pos'][1])
    for i,ptr in enumerate(items):
        for field in range(6):write(ptr+4*field,0x5a000000+i*256+field+1)
    write(flags,row['flags']);write(best,row['best']);write(chosen,objptr(row['chosen']))
    args=[objptr(row['candidate']),objptr(row['unit'])]if row['entry']=='callback'else[objptr(row['unit']),row['radius']]
    u.mem_write(sp,struct.pack('<III',stop,*args));caller=bytes(u.mem_read(stack,0x10000))
    u.emu_start(init,init+8,count=2)
    for reg,value in saved.items():u.reg_write(reg,value)
    u.reg_write(UC_X86_REG_ESP,sp);u.reg_write(UC_X86_REG_EFLAGS,0x202)
    entry={'root':0x5455e0,'search':0x544a00,'callback':0x544a40}[row['entry']]
    active=True;error=None
    try:u.emu_start(entry,stop,count=50000)
    except UcError as err:
        assert faults,(row['name'],str(err));error=str(err)
    finally:active=False
    assert bytes(u.mem_read(heap,0x10000))==bytes(expected),(row['name'],'whole heap')
    for ptr,value in global_expected.items():assert bytes(u.mem_read(ptr,4))==value,(row['name'],hex(ptr),'whole scratch')
    after=bytes(u.mem_read(stack,0x10000));offset=sp-stack
    assert caller[:offset-8192]==after[:offset-8192]and caller[offset:]==after[offset:],(row['name'],'caller guard')
    assert u.reg_read(UC_X86_REG_FPCW)==a.control
    returned=None
    if not faults:
        assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4 and not frames,row['name']
        assert all(u.reg_read(reg)==value for reg,value in saved.items()),row['name']
        assert u.reg_read(UC_X86_REG_FPSW)&0x3800==0,row['name']
        value=u.reg_read(UC_X86_REG_EAX)
        if row['entry']=='root':assert value==0 or value in items;returned=items.index(value)if value else-1
        elif row['entry']=='search':assert value==0 or value in objects;returned=role(value)
        visited.update(seen)
    else:fault_count+=1
    outputs.append(dict(input=row,output=dict(result=returned,events=events,writes=writes,state=state(),fault=faults,error=error,
        calls=dict(calls),whole_heap_scratch_guarded_caller_valid_nonvolatile_and_FPU_preserved=True)))
assert visited==set(instructions),[hex(pc)for pc in set(instructions)-visited]
for ptr,raw in globals_original.items():u.mem_write(ptr,raw)
assert bytes(u.mem_read(base,image_size))==image_before
raw=json.dumps(dict(schema=1,rows=outputs),separators=(',',':')).encode();packed=gzip.compress(raw,mtime=0)
assert not a.output.exists()and not a.proof.exists();a.output.write_bytes(packed)
proof=dict(utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),control=hex(a.control),original_GAME_EXE_sha256=sha(blob),
    rows=len(outputs),faults=fault_count,instruction_bytes=439,original_ranges=code,
    declared_services=['spatial517F90','visibility5370E0','Push50A260'],
    actual_original_search_callback_flags_and_poison_and_nested_search_executed=True,
    no_original_instruction_arithmetic_FPU_or_result_substitution=True,
    all_original_instructions_covered_in_valid_returns=True,
    whole_original_image_heap_scratch_guarded_caller_valid_nonvolatile_and_FPU_guards_passed=True,
    nil_search_unit_faults_are_declared_spatial_point_read_admission_not_original_search_dereference=True,
    gzip_sha256=sha(packed),decoded_sha256=sha(raw))
a.proof.write_text(json.dumps(proof,indent=2)+'\n');print(json.dumps(proof))
