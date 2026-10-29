"""Execute all84 original bytes of three identical terminal action roots.

PopAction50A160 is the only declared body: its exact unit/entry state and
controlled raw return are retained. It can change the ready flag or replace
the update link, exercising the root's complete call/return boundary. This
does not independently revalidate PopAction or started End callbacks. Raw EAX
is retained separately; the public char ABI exposes only its low byte.
"""
import argparse,collections,copy,hashlib,itertools,json,struct
from pathlib import Path
from unicorn import Uc,UC_ARCH_X86,UC_MODE_32,UC_HOOK_CODE
from unicorn.x86_const import *

p=argparse.ArgumentParser(description=__doc__)
for n in ['game','fixture','proof']:p.add_argument('--'+n,type=Path,required=True)
p.add_argument('--control',type=lambda v:int(v,0),choices=[0xe3f,0x23f],default=0xe3f)
args=p.parse_args();blob=args.game.read_bytes();sha=lambda b:hashlib.sha256(b).hexdigest()
assert sha(blob)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,60)[0];opt=pe+24;base=struct.unpack_from('<I',blob,opt+28)[0]
table=opt+struct.unpack_from('<H',blob,pe+20)[0]
u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,(struct.unpack_from('<I',blob,opt+56)[0]+4095)&~4095)
for i in range(struct.unpack_from('<H',blob,pe+6)[0]):
    _,va,size,off=struct.unpack_from('<IIII',blob,table+40*i+8)
    if size:u.mem_write(base+va,blob[off:off+size])
heap,stack,code=0x21000000,0x22000000,0x23000000
for addr in [heap,stack,code]:u.mem_map(addr,0x10000)
update,alternate=heap+0x4000,heap+0x6000
sp,stop,init,cw=stack+0xe000,code+0x100,code+0x200,code+0x300
regions=[('shield-finish',0x5320e0,28),('weapon-block',0x532110,28),('dying',0x544d60,28)]
read32=lambda a:struct.unpack('<I',u.mem_read(a,4))[0]
put32=lambda a,v:u.mem_write(a,struct.pack('<I',v&0xffffffff))
put8=lambda a,v:u.mem_write(a,bytes([v&255]))
read8=lambda a:bytes(u.mem_read(a,1))[0]
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw))
seed=bytes((i*31+19)&255 for i in range(0x10000))
registers={UC_X86_REG_EBX:0x12345678,UC_X86_REG_ESI:0x23456789,UC_X86_REG_EDI:0x3456789a,UC_X86_REG_EBP:0x456789ab}
traces,entries,services={},collections.Counter(),collections.Counter()
active,current,events,unit=False,{},[],0

def state():
    ptr=read32(unit+748);assert ptr in [update,alternate]
    return dict(update=0 if ptr==update else 1,ready=read8(update+483),alternate_ready=read8(alternate+483))

def hook(_u,pc,size,_):
    if not active:return
    if pc==stop:u.emu_stop();return
    for name,start,n in regions:
        if start<=pc<start+n:
            traces[pc]=bytes(u.mem_read(pc,size))
            if pc==start:entries[name]+=1
            return
    assert pc==0x50a160,(current['name'],hex(pc))
    esp=u.reg_read(UC_X86_REG_ESP);assert read32(esp+4)==unit
    services[hex(pc)]+=1;events.append(dict(kind='pop',unit=0,state=state()))
    if current['mutation']=='ready':put8(update+483,current['mutation_ready'])
    elif current['mutation']=='replace-update':
        put32(unit+748,alternate);put8(alternate+483,current['mutation_ready'])
    else:assert current['mutation']=='none'
    u.reg_write(UC_X86_REG_EAX,current['pop_return'])
    u.reg_write(UC_X86_REG_EIP,read32(esp));u.reg_write(UC_X86_REG_ESP,esp+4)

u.hook_add(UC_HOOK_CODE,hook)

def run(c):
    global active,current,events,unit
    active,current,events=False,c,[];unit=heap+0x1000+c['unit_low_byte']
    u.mem_write(heap,seed);put32(unit+8,2);put32(unit+16,0);put32(unit+748,update)
    put8(update+483,c['ready']);put8(alternate+483,0x53)
    before=bytes(u.mem_read(heap,len(seed)))
    u.mem_write(cw,struct.pack('<H',args.control));u.emu_start(init,init+8,count=2)
    u.mem_write(sp-512,bytes(1024));put32(sp,stop);put32(sp+4,unit)
    for reg,v in registers.items():u.reg_write(reg,v)
    u.reg_write(UC_X86_REG_EFLAGS,0x202);u.reg_write(UC_X86_REG_ESP,sp)
    start=next(a for name,a,_ in regions if name==c['function'])
    active=True
    try:u.emu_start(start,stop+1,count=1000)
    except Exception as e:raise AssertionError((c['name'],hex(u.reg_read(UC_X86_REG_EIP)))) from e
    active=False
    assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4,c['name']
    assert all(u.reg_read(reg)==v for reg,v in registers.items()),c['name']
    assert u.reg_read(UC_X86_REG_FPCW)==args.control and u.reg_read(UC_X86_REG_FPSW)&0x3800==0,c['name']
    assert bytes(u.mem_read(sp-512,256))==bytes(256) and bytes(u.mem_read(sp+8,256))==bytes(256),c['name']
    after=bytes(u.mem_read(heap,len(seed)))
    allowed={update+483,alternate+483,*range(unit+748,unit+752)}
    assert all(a==b or heap+i in allowed for i,(a,b) in enumerate(zip(before,after))),c['name']
    raw=u.reg_read(UC_X86_REG_EAX)
    return dict(input=c,result=dict(return_byte=raw&255,events=copy.deepcopy(events),state=state()),raw_original_eax=raw)

cases=[]
for fn,ready,low,ret,mutation in itertools.product([x[0] for x in regions],[0,1,0x7f,0x80,0xff],[0,8,16,40,128,248],[0,1,0x12345678,0x80000000,0xffffffff],['none','ready','replace-update']):
    cases.append(dict(name=f'{fn}-{len(cases)}',function=fn,ready=ready,unit_low_byte=low,pop_return=ret,mutation=mutation,mutation_ready=ready^0xff))
assert len(cases)==1350
rows=[run(c) for c in cases]
coverage=[]
for name,start,size in regions:
    executed={pc+j for pc,b in traces.items() if start<=pc<start+size for j in range(len(b))}
    assert executed==set(range(start,start+size)),name
    raw=bytes(u.mem_read(start,size));coverage.append(dict(name=name,va=hex(start),size=size,executed_bytes=len(executed),sha256=sha(raw),entries=entries[name],instructions=[dict(pc=hex(pc),bytes=b.hex()) for pc,b in sorted(traces.items()) if start<=pc<start+size]))
fixture=dict(original_sha256=sha(blob),control_word=args.control,rows=rows)
proof=dict(original_sha256=sha(blob),control_word=args.control,contexts=len(cases),faults=0,regions=coverage,root_bytes=84,all_root_instruction_bytes_executed=True,services=dict(services),one_PopAction_service_body_declared=True,declared_service_scope='Entry unit/state and arbitrary raw return; optional ready/update-link mutations. No PopAction or End-body equivalence claim.',raw_EAX_retained_separately_char_low_byte_compared=True,register_stack_FPCW_TOP_heap_canaries_preserved=True,nil_or_nonmonster_inputs_excluded=True)
for path,data in [(args.fixture,fixture),(args.proof,proof)]:
    with path.open('x') as f:json.dump(data,f,indent=2);f.write('\n')
print(json.dumps(dict(contexts=len(cases),root_bytes=84,control_word=hex(args.control),services=dict(services),fixture_sha256=sha(args.fixture.read_bytes()))))
