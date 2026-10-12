"""Unchanged packet/prepare/draw execution with declared lookup, UI and image services.

Not a Windows transport or framebuffer oracle.
"""
import collections
import hashlib
import json
import struct
import sys
from pathlib import Path
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import UC_X86_REG_EAX, UC_X86_REG_EIP, UC_X86_REG_ESP

pe=Path(sys.argv[1]).read_bytes()
assert hashlib.sha256(pe).hexdigest()=="0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a"
h=struct.unpack_from("<I",pe,0x3c)[0]
o=h+24
base=struct.unpack_from("<I",pe,o+28)[0]
size=struct.unpack_from("<I",pe,o+56)[0]
s=o+struct.unpack_from("<H",pe,h+20)[0]
cpu=Uc(UC_ARCH_X86,UC_MODE_32)
cpu.mem_map(base,(size+4095)&~4095)
for i in range(struct.unpack_from("<H",pe,h+6)[0]):
    _,virtual,length,raw=struct.unpack_from("<IIII",pe,s+40*i+8)
    cpu.mem_write(base+virtual,pe[raw:raw+length])
cpu.mem_map(0x20000000,0x20000)
PKT,DR,DATA,IMAGES,VP=0x20000020,0x20001020,0x20003020,0x20004020,0x20005020
STACK,STOP=0x20018000,0x2001f000
images=[0x30000010+16*i for i in range(7)]
ranges={0x48ea70:22613,0x4072e0:47,0x43c700:6,0x578b30:8,0x578b70:11,0x4bc720:33,0x4bc750:425}
sealed={at:bytes(cpu.mem_read(at,n)) for at,n in ranges.items()}
tables={0x4942d4:bytes(cpu.mem_read(0x4942d4,832)),0x4bc8fc:bytes(cpu.mem_read(0x4bc8fc,24))}
entries,services,faults=collections.Counter(),collections.Counter(),[]
events,inp=[],{}

def put(at,value):
    cpu.mem_write(at,struct.pack("<I",value&0xffffffff))

def get(at):
    return struct.unpack("<I",cpu.mem_read(at,4))[0]

def args(n):
    return struct.unpack("<"+"I"*n,cpu.mem_read(cpu.reg_read(UC_X86_REG_ESP)+4,4*n))

def ret(value=0):
    sp=cpu.reg_read(UC_X86_REG_ESP)
    cpu.reg_write(UC_X86_REG_EAX,value)
    cpu.reg_write(UC_X86_REG_EIP,get(sp))
    cpu.reg_write(UC_X86_REG_ESP,sp+4)

def hook(_cpu,address,_size,_):
    if address in ranges:
        entries[hex(address)]+=1
    if any(at<=address<at+n for at,n in ranges.items()):
        return
    if address==0x470a80:
        services["packet-ui-counter"]+=1
        ret()
    elif address in (0x45a6f0,0x45a720):
        assert args(1)==(0x1234,) and (address==0x45a720)==inp["static"]
        services["lookup-static" if inp["static"] else "lookup-dynamic"]+=1
        ret(0 if inp.get("missing",False) else DR)
    elif address==0x4c4770:
        vp,dr,img=args(3)
        assert vp==VP and dr==DR and img in images
        events.append(images.index(img))
        services["draw"]+=1
        ret()
    elif address!=STOP:
        raise AssertionError("undeclared execution "+hex(address))

def invalid(_cpu,access,address,length,value,_):
    faults.append(dict(input=inp,address=address,size=length,instruction=cpu.reg_read(UC_X86_REG_EIP)))
    return False

cpu.hook_add(UC_HOOK_CODE,hook)
cpu.hook_add(UC_HOOK_MEM_INVALID,invalid)

def unchanged_code():
    assert all(bytes(cpu.mem_read(at,n))==sealed[at] for at,n in ranges.items())
    assert all(bytes(cpu.mem_read(at,len(data)))==data for at,data in tables.items())

def setup(p):
    global inp,events
    inp,events=p,[]
    cpu.mem_write(0x20000000,bytes(0x17000))
    for at,n in ((PKT,7),(DR,512),(DATA,56),(IMAGES,1020),(VP,48)):
        cpu.mem_write(at-16,b"\x5a"*16)
        cpu.mem_write(at+n,b"\xa5"*16)
    code=0x1234|(0x8000 if p["static"] else 0)
    cpu.mem_write(PKT,struct.pack("<BHI",101,code,p["new"]))
    cls=(0x20400000 if p["static"] else 0)|0x80000|(0x20000 if p["generator"] else 0)
    put(DR+112,cls)
    put(DR+120,0x21000004)
    put(DR+128,0x1234)
    put(DR+280,p["old"])
    put(DR+304,DATA)
    put(DR+432,777)
    put(DATA,56)
    for state in range(5):
        put(DATA+4+4*state,IMAGES)
        cpu.mem_write(DATA+24+state,bytes((p["count"],)))
        cpu.mem_write(DATA+29+state,bytes((p["delay"],)))
        put(DATA+36+4*state,p.get("kind",0))
    cpu.mem_write(IMAGES,struct.pack("<"+"I"*255,*[images[i%7] for i in range(255)]))
    put(0x69b828,int(p["connected"]))
    put(0x85b7a0,0)
    return {at:bytes(cpu.mem_read(at-16,n+32)) for at,n in ((PKT,7),(DR,512),(DATA,56),(IMAGES,1020),(VP,48))}

def invoke(at,arguments):
    cpu.mem_write(STACK,struct.pack("<"+"I"*(len(arguments)+1),STOP,*arguments))
    cpu.reg_write(UC_X86_REG_ESP,STACK)
    cpu.emu_start(at,STOP,count=50000)
    assert cpu.reg_read(UC_X86_REG_EIP)==STOP and cpu.reg_read(UC_X86_REG_ESP)==STACK+4
    unchanged_code()
    return cpu.reg_read(UC_X86_REG_EAX)

def state():
    return dict(flags=get(DR+280),timer=get(DR+432),cls=get(DR+112),obj_flags=get(DR+120))

def packet(p):
    before=setup(p)
    returned=invoke(0x48ea70,(31,PKT,7))
    assert returned==1 and not events
    expected_flags=p["new"] if p["connected"] and not p.get("missing",False) else p["old"]
    prepared=p["connected"] and not p.get("missing",False) and p["generator"] and not(p["old"]&0x400) and bool(p["new"]&0x400)
    expected_timer=p["count"]*(p["delay"]+1) if prepared else 777
    actual=state()
    assert actual["flags"]==expected_flags and actual["timer"]==expected_timer
    mutable={16+x for x in range(280,284)}|{16+x for x in range(432,436)}|{16+x for x in range(112,124)}
    after=bytes(cpu.mem_read(DR-16,544))
    assert all(a==b for i,(a,b) in enumerate(zip(before[DR],after)) if i not in mutable)
    for at in (PKT,DATA,IMAGES,VP):
        assert bytes(cpu.mem_read(at-16,len(before[at])))==before[at]
    return dict(input=p,returned=returned,prepared=prepared,**actual,guards_and_other_input_unchanged=True)

rows=[]
shapes=[("start",0,0x400,True),("repeat",0x400,0x400,True),("no-start",0,0,True),
        ("start-and-terminal",0,0xc00,True),("non-generator",0,0x400,False)]
for count in (0,1,3,7,255):
    for delay in (0,2,255):
        for static in (False,True):
            for name,old,new,generator in shapes:
                rows.append(packet(dict(name=name,count=count,delay=delay,static=static,connected=True,generator=generator,old=old,new=new)))
for name,connected,missing in (("disconnected",False,False),("missing",True,True)):
    rows.append(packet(dict(name=name,count=7,delay=2,static=False,connected=connected,missing=missing,generator=True,old=0,new=0x400)))

timelines=[]
for count in (3,7):
    for delay in (0,2):
        for static in (False,True):
            for kind in (0,2):
                p=dict(name="lifetime",count=count,delay=delay,static=static,connected=True,generator=True,old=0,new=0x400,kind=kind)
                result=packet(p)
                ticks=[]
                original_data=bytes(cpu.mem_read(DATA-16,88))
                for elapsed in range(count*(delay+1)+2):
                    events=[]
                    put(0x84ea04,100+elapsed)
                    returned=invoke(0x4bc750,(VP,DR))
                    assert returned==1 and len(events)==1 and bytes(cpu.mem_read(DATA-16,88))==original_data
                    ticks.append(dict(elapsed=elapsed,frame=100+elapsed,returned=returned,images=list(events),**state()))
                assert state()["timer"]==0 and state()["flags"]==0x800 and state()["cls"]&0x80000==0 and state()["obj_flags"]==0x01000005
                timelines.append(dict(input=p,packet=result,ticks=ticks))
assert not faults
print(json.dumps(dict(original_sha256=hashlib.sha256(pe).hexdigest(),
    functions={hex(at):dict(bytes=n,sha256=hashlib.sha256(sealed[at]).hexdigest()) for at,n in ranges.items()},
    table_hashes={hex(at):hashlib.sha256(data).hexdigest() for at,data in tables.items()},packet_rows=rows,timelines=timelines,
    unhooked_entries=dict(entries),declared_services=dict(services),memory_faults=faults,executable_and_table_bytes_unchanged=True,
    scope="whole original packet decoder/stack probe/status/namespace, original startup preparation and generator draw; declared lookup/UI/image services, not Windows transport/raster"),separators=(",",":")))
