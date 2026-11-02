"""Execute both complete Chakram original bodies; only world services are replaced."""
from pathlib import Path
import argparse, collections, datetime, gzip, hashlib, json, struct
import capstone
from unicorn import Uc,UcError,UC_ARCH_X86,UC_MODE_32,UC_HOOK_CODE,UC_HOOK_MEM_WRITE,UC_HOOK_MEM_INVALID
from unicorn.x86_const import *

p=Path(__file__).resolve().parent
ap=argparse.ArgumentParser(description=__doc__)
ap.add_argument('--exe',type=Path,required=True)
ap.add_argument('--control',type=lambda s:int(s,0),choices=[0xe3f,0x23f],required=True)
ap.add_argument('--output',type=Path,required=True)
ap.add_argument('--proof',type=Path,required=True)
a=ap.parse_args();sha=lambda b:hashlib.sha256(b).hexdigest()
encoded=(p/'chakram_retarget_independent_inputs_4eb250.json.gz').read_bytes()
assert sha(encoded)=='3c2583c1fd4df60f98a8b9162c3de1409a97cee24390278f5f46ae01f0a8ab92'
inputs=json.loads(gzip.decompress(encoded));blob=a.exe.read_bytes()
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
objects=[heap+0x1000+i*0x1000 for i in range(6)]
updates=[heap+0x8000+i*0x1000 for i in range(3)]
sp,stop,init,cw=stack+0xe000,aux+0x100,aux+0x200,aux+0x300
resume_callback,resume_nested=aux+0x110,aux+0x120
u.mem_write(resume_callback,b'\x90');u.mem_write(resume_nested,b'\x90')
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw));u.mem_write(cw,struct.pack('<H',a.control))
scratch={'source':0x753248,'owner':0x7531f4,'last':0x753250,'selected':0x75324c,'best':0x7531f0}
scratch_original={v:bytes(u.mem_read(v,4))for v in scratch.values()}
md=capstone.Cs(capstone.CS_ARCH_X86,capstone.CS_MODE_32);instructions={};ranges=[]
for pc,n in [(0x4eb250,226),(0x4eb340,150)]:
    raw=bytes(u.mem_read(pc,n));seq=list(md.disasm(raw,pc))
    assert sum(i.size for i in seq)==n and seq[-1].mnemonic=='ret'
    instructions.update({i.address:bytes(i.bytes)for i in seq})
    ranges.append(dict(entry=hex(pc),bytes=n,sha256=sha(raw)))
saved={UC_X86_REG_EBX:0x12345678,UC_X86_REG_ESI:0x23456789,UC_X86_REG_EDI:0x3456789a,UC_X86_REG_EBP:0x456789ab}
rd=lambda ptr:struct.unpack('<I',u.mem_read(ptr,4))[0]
objptr=lambda i:0 if i==-1 else 0x41000000 if i==-2 else objects[i]
updptr=lambda i:0 if i==-1 else 0x41000000 if i==-2 else updates[i]
role=lambda v:objects.index(v)if v in objects else -1 if v==0 else -2 if v==0x41000000 else (_ for _ in ()).throw(AssertionError(('unknown pointer',hex(v))))
uprole=lambda v:updates.index(v)if v in updates else -1 if v==0 else -2 if v==0x41000000 else (_ for _ in ()).throw(AssertionError(('unknown update',hex(v))))
active=False;expected=bytearray();global_expected={};row={};events=[];writes=[];faults=[];frames=[];seen=set();visited=set();calls=collections.Counter()
enum_index=buff_index=map_index=0

def write(ptr,value,n=4):
    raw=int(value&((1<<(8*n))-1)).to_bytes(n,'little');u.mem_write(ptr,raw)
    if heap<=ptr<ptr+n<=heap+0x10000:expected[ptr-heap:ptr-heap+n]=raw
    elif ptr in scratch_original and n==4:global_expected[ptr]=raw
    else:raise AssertionError(('undeclared service write',hex(ptr),n))

def state():
    return dict(objects=[dict(cls=rd(v+8),flags=rd(v+16),pos=[rd(v+56),rd(v+60)],
        speed=rd(v+544),vel=[rd(v+80),rd(v+84)],owner=role(rd(v+508)),update=uprole(rd(v+748)))for v in objects],
        updates=[dict(last=role(rd(v+12)),state=bytes(u.mem_read(v+24,1))[0])for v in updates],
        scratch={key:rd(v)if key=='best'else role(rd(v))for key,v in scratch.items()})

def mutate(m):
    for change in m.get('objects',[]):
        v=objects[change['id']]
        for key,off in [('cls',8),('flags',16),('speed',544)]:
            if key in change:write(v+off,change[key])
        for key,off in [('pos',56),('vel',80)]:
            if key in change:
                for i,value in enumerate(change[key]):write(v+off+4*i,value)
        if 'owner'in change:write(v+508,objptr(change['owner']))
        if 'update'in change:write(v+748,updptr(change['update']))
    for change in m.get('updates',[]):
        v=updates[change['id']]
        if 'last'in change:write(v+12,objptr(change['last']))
        if 'state'in change:write(v+24,change['state'],1)
    for key,value in m.get('scratch',{}).items():write(scratch[key],value if key=='best'else objptr(value))

def registers():return {k:u.reg_read(k)for k in saved}
def check_regs(values):assert all(u.reg_read(k)==v for k,v in values.items()),(row['name'],'nonvolatile service/callback register')
def ret(esp,value):
    u.reg_write(UC_X86_REG_EAX,value&0xffffffff)
    u.reg_write(UC_X86_REG_ESP,esp+4);u.reg_write(UC_X86_REG_EIP,rd(esp))
def schedule_callback(frame):
    candidate=frame['plan']['candidates'][frame['index']]
    frame['candidate']=candidate
    esp=frame['esp']-12;u.mem_write(esp,struct.pack('<III',resume_callback,objptr(candidate),frame['point']))
    u.reg_write(UC_X86_REG_ESP,esp);u.reg_write(UC_X86_REG_EIP,0x4eb340)
def finish_enum(frame):
    mutate(frame['plan'].get('after',{}))
    events.append(dict(kind='enum-return',index=frame['number'],after=state()))
    check_regs(frame['registers']);frames.pop();ret(frame['esp'],0x76543210)
def continue_enum(frame):
    if frame['plan']['candidates']:schedule_callback(frame)
    else:finish_enum(frame)
def finish_service(frame):
    step=frame['step'];mutate(step.get('after_nested',{}))
    value=step.get('value',0 if frame['kind']=='buff'else 1)
    events.append(dict(kind=frame['kind']+'-return',index=frame['number'],value=value,after=state()))
    check_regs(frame['registers']);frames.pop();ret(frame['esp'],value)
def start_nested(frame,nested):
    esp=frame['esp']-8;u.mem_write(esp,struct.pack('<II',resume_nested,objptr(nested['source'])))
    events.append(dict(kind='nested',service=frame['kind'],index=frame['number'],source=nested['source'],before=state()))
    u.reg_write(UC_X86_REG_ESP,esp);u.reg_write(UC_X86_REG_EIP,0x4eb250)

def on_code(_u,pc,n,_):
    global enum_index,buff_index,map_index
    if not active:return
    esp=u.reg_read(UC_X86_REG_ESP)
    if pc==resume_callback:
        frame=frames[-1];assert frame['kind']=='enum'and esp==frame['esp']-8
        check_regs(frame['registers']);assert u.reg_read(UC_X86_REG_FPSW)&0x3800==0
        u.reg_write(UC_X86_REG_ESP,frame['esp'])
        events.append(dict(kind='callback-return',enum=frame['number'],index=frame['index'],candidate=frame['candidate'],after=state()))
        changes=frame['plan'].get('after_each',[])
        if frame['index']<len(changes):mutate(changes[frame['index']])
        frame['index']+=1
        if frame['index']<len(frame['plan']['candidates']):schedule_callback(frame)
        else:finish_enum(frame)
        return
    if pc==resume_nested:
        frame=frames[-1];assert esp==frame['esp']-4
        check_regs(frame['registers']);assert u.reg_read(UC_X86_REG_FPSW)&0x3800==0
        events.append(dict(kind='nested-return',service=frame['kind'],index=frame['number'],result=role(u.reg_read(UC_X86_REG_EAX)),after=state()))
        u.reg_write(UC_X86_REG_ESP,frame['esp'])
        if frame['kind']=='enum':continue_enum(frame)
        else:finish_service(frame)
        return
    if pc==0x517c10:
        rect,callback,point=[rd(esp+4*i)for i in range(1,4)]
        assert callback==0x4eb340
        number=enum_index;enum_index+=1
        plan=row['enums'][number]if number<len(row['enums'])else dict(candidates=[])
        frame=dict(kind='enum',number=number,esp=esp,point=point,index=0,plan=plan,registers=registers());frames.append(frame)
        point_object=next((i for i,v in enumerate(objects)if point==v+56),None);assert point_object is not None
        events.append(dict(kind='enum',index=number,rectangle=[rd(rect+4*i)for i in range(4)],point_object=point_object,before=state()))
        mutate(plan.get('before',{}))
        if 'reenter'in plan:start_nested(frame,plan['reenter'])
        else:continue_enum(frame)
        return
    if pc in [0x4ff350,0x537110]:
        candidate,other=rd(esp+4),rd(esp+8)
        if pc==0x4ff350:
            assert other==0;kind='buff';number=buff_index;buff_index+=1;plans=row['buffs']
        else:kind='map';number=map_index;map_index+=1;plans=row['maps']
        step=plans[number]if number<len(plans)else {}
        frame=dict(kind=kind,number=number,esp=esp,step=step,registers=registers());frames.append(frame)
        event=dict(kind=kind,index=number,candidate=role(candidate),before=state())
        event['enchant'if kind=='buff'else 'source']=other if kind=='buff'else role(other)
        events.append(event);mutate(step.get('mutation',{}))
        if 'reenter'in step:start_nested(frame,step['reenter'])
        else:finish_service(frame)
        return
    assert pc in instructions and bytes(u.mem_read(pc,n))==instructions[pc],(row['name'],hex(pc),n)
    seen.add(pc)
    if pc in [0x4eb250,0x4eb340]:calls[hex(pc)]+=1

def on_write(_u,_access,ptr,n,value,_):
    if not active:return
    pc=u.reg_read(UC_X86_REG_EIP)
    if stack<=ptr<stack+0x10000:
        if pc==0x4eb2f3 and ptr==sp+4 and n==4:
            caller_expected[ptr-stack:ptr-stack+n]=int(value&0xffffffff).to_bytes(4,'little');return
        assert sp-8192<=ptr<ptr+n<=sp,(row['name'],hex(pc),hex(ptr),n);return
    key=next((key for key,v in scratch.items()if ptr==v),None)
    if key is not None:
        allowed={0x4eb264:'selected',0x4eb275:'last',0x4eb280:'owner',0x4eb286:'source',0x4eb28c:'best',0x4eb3c4:'best',0x4eb3ca:'selected'}
        assert n==4 and allowed.get(pc)==key,(row['name'],hex(pc),key)
        global_expected[ptr]=int(value&0xffffffff).to_bytes(4,'little')
        writes.append(dict(after_event=len(events),pc=hex(pc),target=key,size=n,value=value&0xffffffff if key=='best'else role(value&0xffffffff)))
        return
    if pc==0x4eb2df:
        index=next((i for i,v in enumerate(updates)if ptr==v+24),None)
        assert index is not None and n==1
        target=dict(target='state',record=index)
    else:
        assert pc in [0x4eb317,0x4eb328]and n==4,(row['name'],hex(pc),hex(ptr),n)
        off=80 if pc==0x4eb317 else 84
        index=next((i for i,v in enumerate(objects)if ptr==v+off),None);assert index is not None
        target=dict(target='velocityX'if off==80 else'velocityY',record=index)
    expected[ptr-heap:ptr-heap+n]=int(value&((1<<(8*n))-1)).to_bytes(n,'little')
    writes.append(dict(after_event=len(events),pc=hex(pc),size=n,value=value&((1<<(8*n))-1),**target))
def on_invalid(_u,access,ptr,n,_value,_):
    faults.append(dict(kind='original-memory',access=access,address=hex(ptr),bytes=n,pc=hex(u.reg_read(UC_X86_REG_EIP))))
    return False

u.hook_add(UC_HOOK_CODE,on_code);u.hook_add(UC_HOOK_MEM_WRITE,on_write);u.hook_add(UC_HOOK_MEM_INVALID,on_invalid)
seed=bytes((i*37+0xa5)&255 for i in range(0x10000));outputs=[];fault_count=0
for row in inputs['rows']:
    active=False;events=[];writes=[];faults=[];frames=[];seen=set();calls=collections.Counter()
    enum_index=buff_index=map_index=0
    u.mem_write(heap,seed);expected=bytearray(seed);u.mem_write(stack,b'\x5a'*0x10000)
    global_expected={ptr:raw for ptr,raw in scratch_original.items()}
    for ptr,raw in global_expected.items():u.mem_write(ptr,raw)
    mutate(dict(objects=[dict(id=i,**v)for i,v in enumerate(row['objects'])],updates=[dict(id=i,**v)for i,v in enumerate(row['updates'])],scratch=row['scratch']))
    args=[objptr(row['source'])]if row['entry']=='driver'else[objptr(row['candidate']),objptr(row['point'])+row['point_offset']]
    u.mem_write(sp,struct.pack('<'+'I'*(len(args)+1),stop,*args));caller=bytes(u.mem_read(stack,0x10000));caller_expected=bytearray(caller)
    u.emu_start(init,init+8,count=2)
    for reg,value in saved.items():u.reg_write(reg,value)
    u.reg_write(UC_X86_REG_ESP,sp);u.reg_write(UC_X86_REG_EFLAGS,0x202)
    active=True;error=None
    try:u.emu_start(0x4eb250 if row['entry']=='driver'else 0x4eb340,stop,count=100000)
    except UcError as err:
        assert faults,(row['name'],str(err));error=str(err)
    finally:active=False
    assert bytes(u.mem_read(heap,0x10000))==bytes(expected),(row['name'],'whole 64KiB heap')
    for ptr,raw in global_expected.items():assert bytes(u.mem_read(ptr,4))==raw,(row['name'],hex(ptr),'scratch')
    after=bytes(u.mem_read(stack,0x10000));offset=sp-stack
    assert caller[:offset-8192]==after[:offset-8192]and bytes(caller_expected[offset:])==after[offset:],(row['name'],'caller guard')
    assert u.reg_read(UC_X86_REG_FPCW)==a.control
    returned=None
    if not faults:
        assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4 and not frames,row['name']
        check_regs(saved);assert u.reg_read(UC_X86_REG_FPSW)&0x3800==0,row['name']
        if row['entry']=='driver':returned=role(u.reg_read(UC_X86_REG_EAX))
        visited.update(seen)
    else:fault_count+=1
    outputs.append(dict(input=row,output=dict(result=returned,events=events,writes=writes,state=state(),fault=faults,error=error,
        calls=dict(calls),whole_heap_scratch_guarded_caller_valid_nonvolatile_and_FPU_preserved=True)))
assert visited==set(instructions),[hex(pc)for pc in set(instructions)-visited]
for ptr,raw in scratch_original.items():u.mem_write(ptr,raw)
assert bytes(u.mem_read(base,image_size))==image_before
raw=json.dumps(dict(schema=1,rows=outputs),separators=(',',':')).encode();packed=gzip.compress(raw,mtime=0)
assert not a.output.exists()and not a.proof.exists();a.output.write_bytes(packed)
proof=dict(utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),control=hex(a.control),original_GAME_EXE_SHA256=sha(blob),
    rows=len(outputs),faults=fault_count,instruction_bytes=376,original_ranges=ranges,independent_input_SHA256=sha(encoded),
    declared_services=['spatial517C10','buff4FF350','visibility537110'],
    both_original_bodies_and_reentrant_original_calls_executed=True,
    no_original_instruction_arithmetic_FPU_or_result_substitution=True,
    all_original_instructions_covered_in_valid_returns=True,
    whole_original_image_heap_scratch_guarded_caller_valid_nonvolatile_and_FPU_guards_passed=True,
    gzip_SHA256=sha(packed),decoded_SHA256=sha(raw))
a.proof.write_text(json.dumps(proof,indent=2)+'\n');print(json.dumps(proof))
