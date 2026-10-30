"""Replay six unchanged state predicates plus the real WORD name-ID helper.

Only dictionary lookup 004E3830 is a declared service. It returns a valid
PE32 type-record pointer or nil; the original 004E3AA0 loads its WORD ID.
No predicate return, frame/FPS arithmetic or cache publication is modeled.
"""
import argparse,gzip,hashlib,itertools,json,struct
from pathlib import Path
from unicorn import Uc,UC_ARCH_X86,UC_MODE_32,UC_HOOK_CODE,UC_HOOK_MEM_WRITE
from unicorn.x86_const import *

parser=argparse.ArgumentParser(description=__doc__)
for name in ['game','fixture','proof']:parser.add_argument('--'+name,type=Path,required=True)
parser.add_argument('--control',type=lambda v:int(v,0),choices=[0xe3f,0x23f],required=True)
args=parser.parse_args()
stage=Path(__file__).resolve().parent
blob=args.game.read_bytes();sha=lambda raw:hashlib.sha256(raw).hexdigest()
assert sha(blob)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
audit=json.loads((stage/'original-state-predicates-readonly-v2.json').read_text())
instructions={int(v['address'],16):bytes.fromhex(v['bytes'])for row in audit['functions']for v in row['instructions']}
assert sum(map(len,instructions.values()))==244
entries=[0x534710,0x5347a0,0x5347c0,0x5347f0,0x534810,0x534840]
pe=struct.unpack_from('<I',blob,60)[0];optional=pe+24
base=struct.unpack_from('<I',blob,optional+28)[0];table=optional+struct.unpack_from('<H',blob,pe+20)[0]
u=Uc(UC_ARCH_X86,UC_MODE_32)
u.mem_map(base,(struct.unpack_from('<I',blob,optional+56)[0]+4095)&~4095)
for i in range(struct.unpack_from('<H',blob,pe+6)[0]):
    _,va,size,offset=struct.unpack_from('<IIII',blob,table+i*40+8)
    if size:u.mem_write(base+va,blob[offset:offset+size])
heap,stack,code=0x21000000,0x22000000,0x23000000
for address in [heap,stack,code]:u.mem_map(address,0x10000)
unit,update,health,type_record=heap+0x1000,heap+0x3000,heap+0x6000,heap+0xa000
sp,stop,init,cw=stack+0xe000,code+0x100,code+0x200,code+0x300
frame_address,fps_address,cache_address,literal=0x84ea04,0x85b3fc,0x833e60,0x5c76d8
assert bytes(u.mem_read(literal,6))==b'Mimic\0'
u.mem_write(stop,b'\x90')
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw));u.mem_write(cw,struct.pack('<H',args.control))
rd=lambda address:struct.unpack('<I',u.mem_read(address,4))[0]
wr=lambda address,value:u.mem_write(address,struct.pack('<I',value&0xffffffff))
seed=bytes((i*31+19)&255 for i in range(0x10000))
saved={UC_X86_REG_EBX:0x12345678,UC_X86_REG_ESI:0x23456789,UC_X86_REG_EDI:0x3456789a,UC_X86_REG_EBP:0x456789ab}
active=False;current={};traces={};lookups=[];writes=[];call_index=0;actual_name_helper_calls=0
def on_code(_u,pc,size,_):
    global actual_name_helper_calls
    if not active:return
    if pc==0x4e3830:
        assert current['entry']==0x534840
        pointer=u.reg_read(UC_X86_REG_ESP)
        assert rd(pointer)==0x4e3aaa and rd(pointer+4)==literal
        assert bytes(u.mem_read(literal,6))==b'Mimic\0'
        found=current['input']['lookup_index']!=0
        lookups.append(dict(service='dictionary_type_by_name_4E3830',name='Mimic',found=found,
            type_index=current['input']['lookup_index']if found else None,cache_before=rd(cache_address),call=call_index))
        u.reg_write(UC_X86_REG_EAX,type_record if found else 0)
        u.reg_write(UC_X86_REG_ECX,0x87654321);u.reg_write(UC_X86_REG_EDX,0x98765432)
        u.reg_write(UC_X86_REG_ESP,pointer+4);u.reg_write(UC_X86_REG_EIP,rd(pointer))
        return
    assert pc in instructions,(current['name'],hex(pc))
    raw=bytes(u.mem_read(pc,size));assert raw==instructions[pc]
    traces[pc]=raw
    if pc==0x4e3aa0:
        assert current['entry']==0x534840
        actual_name_helper_calls+=1
def on_write(_u,_access,address,size,value,_):
    if not active:return
    if stack<=address and address+size<=stack+0x10000:return
    if address==cache_address and size==4:
        assert current['entry']==0x534840 and u.reg_read(UC_X86_REG_EIP)==0x534856
        writes.append(dict(field='Mimic_cache',value=value&0xffffffff,call=call_index))
        return
    raise AssertionError((current['name'],'unexpected write',hex(address),size,value))
u.hook_add(UC_HOOK_CODE,on_code);u.hook_add(UC_HOOK_MEM_WRITE,on_write)
defaults=dict(cls=2,status=0,head=0,action=4,prefix528=0,type_index=1,poison=0,
    health_present=True,health_cur=50,health_max=100,frame=1400,fps=31,move_frame=1400,
    cache=0x11223344,lookup_index=1,repeat=1)
cases=[];per_entry={}
def add(entry,values):
    i=per_entry.get(entry,0);per_entry[entry]=i+1
    cases.append(dict(name=f'{entry:06x}/{i:05d}',entry=entry,input=dict(defaults,**values)))
words=[0,0xffffffff]+[1<<i for i in range(32)]
classes=[0,2,6,0xffffffff]
for status,head,action in itertools.product([0,0x40,0x200,0xffffffff,0xffffffbf],
        [-1,0,1,7,23],list(range(81))+[0xffffffff,0x80000000]):
    add(0x534710,dict(status=status,head=head,action=action,prefix528=action if head==-1 else 0xdeadbeef))
for cls,status in itertools.product(classes,words):add(0x5347a0,dict(cls=cls,status=status))
health_values=[0,1,2,3,255,256,32767,32768,65534,65535]
for cls in classes:
    add(0x5347c0,dict(cls=cls,health_present=False))
    for cur,maximum in itertools.product(health_values,health_values):
        add(0x5347c0,dict(cls=cls,health_cur=cur,health_max=maximum))
for cls,poison in itertools.product(classes,range(256)):add(0x5347f0,dict(cls=cls,poison=poison))
for fps in [0,1,10,30,31,60,120,0x55555555,0x55555556,0x80000000,0xffffffff]:
    edge=(3*fps)&0xffffffff
    elapsed=sorted({0,1,2,0x7fffffff,0x80000000,0xffffffff}|{(edge+d)&0xffffffff for d in range(-3,4)})
    for frame,delta in itertools.product([0,1,2,1400,0x7fffffff,0x80000000,0xfffffffe,0xffffffff],elapsed):
        add(0x534810,dict(frame=frame,fps=fps,move_frame=(frame-delta)&0xffffffff))
type_indices=[0,1,2,7,17,0x7fff,0x8000,0x8001,0xfffe,0xffff]
caches=sorted(set(words+type_indices))
for cache,type_index,lookup in itertools.product(caches,type_indices,[0,1,2,7,17]):
    add(0x534840,dict(cache=cache,type_index=type_index,lookup_index=lookup))
for cache,type_index,lookup,repeat in itertools.product([0,1],type_indices,[0,1,2,7,17],[2,3]):
    add(0x534840,dict(cache=cache,type_index=type_index,lookup_index=lookup,repeat=repeat))
rows=[]
for case in cases:
    current=case;c=case['input'];active=False
    u.mem_write(heap,seed)
    u.mem_write(unit+4,struct.pack('<H',c['type_index']));wr(unit+8,c['cls'])
    u.mem_write(unit+540,bytes([c['poison']]))
    wr(unit+556,health if c['health_present']else 0);wr(unit+748,update)
    u.mem_write(health,struct.pack('<H',c['health_cur']));u.mem_write(health+4,struct.pack('<H',c['health_max']))
    wr(update+508,c['move_frame']);wr(update+528,c['prefix528'])
    u.mem_write(update+544,struct.pack('<b',c['head']))
    if c['head']>=0:wr(update+552+24*c['head'],c['action'])
    wr(update+1440,c['status'])
    u.mem_write(type_record,struct.pack('<H',c['lookup_index']))
    wr(frame_address,c['frame']);wr(fps_address,c['fps']);wr(cache_address,c['cache'])
    heap_before=bytes(u.mem_read(heap,0x10000))
    cache_guard=bytes(u.mem_read(cache_address-64,132))
    time_before=[bytes(u.mem_read(address-16,36))for address in [frame_address,fps_address]]
    lookups=[];writes=[];returns=[]
    for call_index in range(c['repeat']):
        u.mem_write(stack,seed);u.mem_write(sp,struct.pack('<II',stop,unit))
        low=bytes(u.mem_read(sp-256,128));high=bytes(u.mem_read(sp+16,128))
        for register,value in saved.items():u.reg_write(register,value)
        u.reg_write(UC_X86_REG_EAX,0x76543210);u.reg_write(UC_X86_REG_ECX,0x87654321);u.reg_write(UC_X86_REG_EDX,0x98765432)
        u.reg_write(UC_X86_REG_ESP,sp);u.reg_write(UC_X86_REG_EFLAGS,2)
        u.emu_start(init,init+8,count=3)
        active=True;u.emu_start(case['entry'],stop,count=2000);active=False
        assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4,case['name']
        assert u.reg_read(UC_X86_REG_FPCW)==args.control and u.reg_read(UC_X86_REG_FPSW)==0,case['name']
        assert all(u.reg_read(register)==value for register,value in saved.items()),case['name']
        assert bytes(u.mem_read(sp-256,128))==low and bytes(u.mem_read(sp+16,128))==high,case['name']
        result=u.reg_read(UC_X86_REG_EAX);assert result in [0,1]
        returns.append(result)
    assert bytes(u.mem_read(heap,0x10000))==heap_before,case['name']
    guard=bytes(u.mem_read(cache_address-64,132))
    assert guard[:64]==cache_guard[:64]and guard[68:]==cache_guard[68:],case['name']
    assert [bytes(u.mem_read(address-16,36))for address in [frame_address,fps_address]]==time_before,case['name']
    if case['entry']!=0x534840:assert not lookups and not writes and rd(cache_address)==c['cache']
    rows.append(dict(case,returns=returns,cache_after=rd(cache_address),lookups=lookups,cache_writes=writes))
assert traces==instructions,('uncovered original instructions',sorted(set(instructions)-set(traces)))
assert all(bytes(u.mem_read(address,len(raw)))==raw for address,raw in instructions.items())
assert bytes(u.mem_read(literal,6))==b'Mimic\0'
assert actual_name_helper_calls==sum(len(row['lookups'])for row in rows)
fixture=dict(schema='original-monster-state-predicates-534710-v1',original_sha256=sha(blob),
    control_word=args.control,declared_dictionary_boundary='004E3830 only; complete 004E3AA0 WORD-ID body actually executes',rows=rows)
raw=(json.dumps(fixture,separators=(',',':'))+'\n').encode()
assert not args.fixture.exists()and not args.proof.exists()
with args.fixture.open('xb')as f:
    with gzip.GzipFile(filename='',mode='wb',fileobj=f,mtime=0)as z:z.write(raw)
proof=dict(status='passed',original_sha256=sha(blob),control_word=args.control,primary_roots=6,
    original_helper='004E3AA0',instruction_bytes_executed=244,primary_instruction_bytes=216,
    contexts=len(rows),actual_primary_calls=sum(len(row['returns'])for row in rows),
    per_entry={hex(entry):per_entry[entry]for entry in entries},
    actual_name_ID_helper_calls=actual_name_helper_calls,
    declared_service_boundary=dict(address='0x4e3830',name='dictionary type by name',result='PE32 type-record pointer or nil; original helper reads WORD',calls=actual_name_helper_calls),
    cache_publication_writes=sum(len(row['cache_writes'])for row in rows),
    all_heap_stack_register_control_word_TOP_and_global_guards=True,
    code_readonly_Mimic_literal_and_nonoutput_memory_unchanged=True,
    fixture=dict(path=str(args.fixture),bytes=args.fixture.stat().st_size,sha256=sha(args.fixture.read_bytes()),decoded_bytes=len(raw),decoded_sha256=sha(raw)))
with args.proof.open('x')as f:json.dump(proof,f,indent=2);f.write('\n')
print(json.dumps(dict(status='passed',contexts=len(rows),actual_primary_calls=proof['actual_primary_calls'],
    original_bytes=244,actual_name_ID_helper_calls=actual_name_helper_calls,cache_writes=proof['cache_publication_writes'])))

