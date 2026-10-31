"""Execute unchanged 518AE0 and 578AC0; three declared publication services.

The fallback special-reporting body is an explicit boundary, not execution
coverage. Inputs are frozen independently. Linked frame bytes are value XOR
0xa5, distinct from the parent's bytes, and cover all 256 byte values.
"""
import argparse, collections, datetime, gzip, hashlib, json, struct
from pathlib import Path
from unicorn import Uc, UcError, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_WRITE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import *

ap=argparse.ArgumentParser(description=__doc__)
for n in ['game','fixture','proof']:ap.add_argument('--'+n,type=Path,required=True)
ap.add_argument('--control',type=lambda v:int(v,0),choices=[0xe3f,0x23f],required=True)
ap.add_argument('--limit',type=int,default=0)
args=ap.parse_args();stage=Path(__file__).resolve().parent
audit=json.loads((stage/'original-functions-and-source-before-v1.json').read_text())
frozen=json.loads(gzip.decompress((stage/'original-inputs-frozen-v1.json.gz').read_bytes()))
sha=lambda b:hashlib.sha256(b).hexdigest()
blob=args.game.read_bytes();assert sha(blob)==audit['original_sha256']
pe=struct.unpack_from('<I',blob,60)[0];opt=pe+24;base=struct.unpack_from('<I',blob,opt+28)[0]
tab=opt+struct.unpack_from('<H',blob,pe+20)[0]
u=Uc(UC_ARCH_X86,UC_MODE_32)
u.mem_map(base,(struct.unpack_from('<I',blob,opt+56)[0]+4095)&~4095)
for i in range(struct.unpack_from('<H',blob,pe+6)[0]):
    _,rva,size,off=struct.unpack_from('<IIII',blob,tab+40*i+8)
    if size:u.mem_write(base+rva,blob[off:off+size])
instructions={int(r['address'],16):bytes.fromhex(r['bytes'])for f in audit['functions']for r in f['instructions']}
assert sum(map(len,instructions.values()))==388
for f in audit['functions']:assert sha(bytes(u.mem_read(int(f['entry'],16),f['size'])))==f['sha256']
heap,stack,code=0x21000000,0x22000000,0x23000000
for a in [heap,stack,code]:u.mem_map(a,0x10000)
unit,update,linked,linked_update,recipient=heap+0x1000,heap+0x3000,heap+0x5000,heap+0x7000,heap+0x9000
sp,stop,init,cw=stack+0xe000,code+0x100,code+0x200,code+0x300
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw));u.mem_write(cw,struct.pack('<H',args.control))
rd=lambda a:struct.unpack('<I',u.mem_read(a,4))[0]
wr=lambda a,v:u.mem_write(a,struct.pack('<I',v&0xffffffff))
saved={UC_X86_REG_EBX:0x12345678,UC_X86_REG_ESI:0x23456789,UC_X86_REG_EDI:0x3456789a,UC_X86_REG_EBP:0x456789ab}
active=False;current=None;events=[];invalid=[];stack_writes=[]
visited=set();valid_visited=set();calls=collections.Counter();call_index=0

def event(value):
    value['call']=call_index;events.append(value)

def on_code(_u,pc,size,_):
    if not active:return
    ptr=u.reg_read(UC_X86_REG_ESP)
    if pc in [0x4e5390,0x40ebc0,0x527e50]:
        if pc==0x4e5390:
            assert rd(ptr+4)==current['recipient'] and rd(ptr+12)==4 and rd(ptr+16)==0 and rd(ptr+20)==1
            event(dict(service=hex(pc),recipient_word=rd(ptr+4),length=4,zero=0,reliable=1,packet=bytes(u.mem_read(rd(ptr+8),4)).hex()))
            result=current['direct_return']
        elif pc==0x40ebc0:
            assert rd(ptr+4)==current['recipient'] and rd(ptr+8)==1 and rd(ptr+16)==4
            event(dict(service=hex(pc),recipient_word=rd(ptr+4),kind=1,length=4,packet=bytes(u.mem_read(rd(ptr+12),4)).hex()))
            result=current['queue_return']
        else:
            assert rd(ptr+4)==recipient and rd(ptr+8)==unit
            event(dict(service=hex(pc),recipient='recipient',unit='unit'))
            result=current['special_return']
        u.reg_write(UC_X86_REG_EAX,result)
        u.reg_write(UC_X86_REG_ECX,0x87654321);u.reg_write(UC_X86_REG_EDX,0x98765432)
        u.reg_write(UC_X86_REG_ESP,ptr+4);u.reg_write(UC_X86_REG_EIP,rd(ptr));return
    assert pc in instructions,(current['name'],'unadmitted instruction',hex(pc))
    assert bytes(u.mem_read(pc,size))==instructions[pc]
    visited.add(pc)
    if pc in [0x518ae0,0x578ac0]:
        calls[hex(pc)]+=1
        if pc==0x578ac0:
            assert rd(ptr+4)==(unit if current['unit_present']else 0)
            event(dict(entry=hex(pc),unit='unit'if current['unit_present']else None))
    if pc==0x518aeb:event(dict(child_return='0x578ac0',return_word=u.reg_read(UC_X86_REG_EAX)))

def on_write(_u,_access,address,size,value,_):
    if not active:return
    assert sp-256<=address and address+size<=sp+16,(current['name'],'unexpected non-stack write',hex(address),size)
    if address>=sp:assert sp+12<=address and address+size<=sp+16
    stack_writes.append([address-sp,size,value&((1<<(size*8))-1)])

def on_invalid(_u,access,address,size,_value,_):
    if active:invalid.append(dict(access=access,address=hex(address),size=size,pc=hex(u.reg_read(UC_X86_REG_EIP))))
    return False

u.hook_add(UC_HOOK_CODE,on_code);u.hook_add(UC_HOOK_MEM_WRITE,on_write);u.hook_add(UC_HOOK_MEM_INVALID,on_invalid)

def fields(c):
    u.mem_write(unit+4,struct.pack('<H',c['type_index']))
    for off,value in [(8,c['object_class']),(12,c['subclass']),(16,c['flags']),(36,c['primary_code']),(40,c['secondary_code']),(748,update if c['update_present']else 0)]:wr(unit+off,value)
    for off in [0,12,16,20]:wr(update+off,c['value'])
    wr(update+4,linked if c['link_present']else 0)
    wr(linked+748,linked_update if c['linked_update_present']else 0)
    wr(linked_update+16,c['value']^0xa5)
    for i,value in enumerate(c['caches']):wr(0x81b1b0+4*i,value)

def execute(c,expect_fault=False):
    global active,current,events,invalid,stack_writes,call_index
    current=dict(c);events=[];returns=[];states=[];fault=None
    u.mem_write(heap,b'\xcd'*0x10000)
    for a,n in [(unit,1024),(update,64),(linked,1024),(linked_update,64),(recipient,1024)]:u.mem_write(a,b'\0'*n)
    fields(current)
    for call_index,changes in enumerate([{}]+current['sequence']):
        current.update(changes);fields(current)
        before_heap=bytes(u.mem_read(heap,0x10000));before_globals=bytes(u.mem_read(0x81b000,0x1000))
        u.mem_write(sp-1024,b'\xa7'*2048)
        for off,value in [(0,stop),(4,recipient),(8,current['recipient']),(12,unit if current['unit_present']else 0)]:wr(sp+off,value)
        for reg,value in saved.items():u.reg_write(reg,value)
        u.reg_write(UC_X86_REG_ESP,sp);u.reg_write(UC_X86_REG_EFLAGS,0x202)
        u.emu_start(init,init+8,count=3);assert u.reg_read(UC_X86_REG_FPCW)==args.control
        invalid=[];stack_writes=[];active=True
        try:u.emu_start(0x518ae0,stop,count=2000)
        except UcError:
            if not expect_fault or not invalid:raise
            fault=list(invalid)
        finally:active=False
        assert bytes(u.mem_read(heap,0x10000))==before_heap
        assert bytes(u.mem_read(0x81b000,0x1000))==before_globals
        assert u.reg_read(UC_X86_REG_FPCW)==args.control
        if fault:break
        assert not expect_fault and u.reg_read(UC_X86_REG_EIP)==stop
        assert u.reg_read(UC_X86_REG_ESP)==sp+4
        assert all(u.reg_read(reg)==v for reg,v in saved.items())
        returns.append(u.reg_read(UC_X86_REG_EAX))
        states.append(dict(argument_packet_after=bytes(u.mem_read(sp+12,4)).hex(),stack_write_count=len(stack_writes)))
    if expect_fault:assert fault
    return dict(name=c['name'],input=c,output=dict(returns=returns,events=list(events),fault=fault,local_stack_observations=states,
                full_heap_global_neighbors_nonvolatile_and_FP_guards_preserved=not expect_fault,
                heap_global_neighbors_and_FP_guards_preserved=True))

contexts=frozen['contexts'][:args.limit or None]
rows=[execute(c)for c in contexts];valid_visited.update(visited)
fault_rows=[execute(c,True)for c in frozen['expected_fault_contexts']]
payload=json.dumps(dict(contexts=rows,expected_fault_contexts=fault_rows),separators=(',',':')).encode()
with args.fixture.open('xb') as f:
    with gzip.GzipFile(filename='',mode='wb',fileobj=f,mtime=0)as z:z.write(payload)
coverage=[]
for f in audit['functions']:
    entry=int(f['entry'],16);end=entry+f['size']
    coverage.append(dict(name=f['name'],entry=f['entry'],static_bytes=f['size'],
        valid_executed_bytes=sum(len(b)for pc,b in instructions.items()if entry<=pc<end and pc in valid_visited),
        including_separate_fault_executed_bytes=sum(len(b)for pc,b in instructions.items()if entry<=pc<end and pc in visited),
        missing_instruction_addresses=[hex(pc)for pc in instructions if entry<=pc<end and pc not in visited]))
proof=dict(status='original_root_and_scalar_helper_executed_with_declared_services',utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),
    control=hex(args.control),original_sha256=sha(blob),frozen_input_sha256=sha((stage/'original-inputs-frozen-v1.json.gz').read_bytes()),
    producer_sha256=sha(Path(__file__).read_bytes()),contexts=len(rows),calls=sum(len(x['output']['returns'])for x in rows),
    separate_fault_contexts=len(fault_rows),entries=dict(calls),coverage=coverage,fixture_sha256=sha(args.fixture.read_bytes()),uncompressed_sha256=sha(payload),
    declaration='Only root330 and real unit-code58. Special-reporting527E50 and two publication services are explicit boundaries, not executed child bodies.',
    fixture_linked_frame='value XOR0xa5, distinct from parent update; all byte values exercised',engine_modified=False,all_before_engine_complete=False,overall_porting_100_percent=False)
with args.proof.open('x')as f:json.dump(proof,f,indent=2);f.write('\n')
print(json.dumps(proof))
