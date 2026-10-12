"""Execute unchanged packet decoder and AnimateDraw with declared lookup/draw/delete services.

The original instruction stream and stack probe execute, not Windows transport or raster.
"""
import collections
import hashlib
import json
import struct
import sys
from pathlib import Path

from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import UC_X86_REG_EAX, UC_X86_REG_EIP, UC_X86_REG_ESP

pe = Path(sys.argv[1]).read_bytes()
assert hashlib.sha256(pe).hexdigest() == "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a"
h = struct.unpack_from("<I", pe, 0x3c)[0]
o = h + 24
base = struct.unpack_from("<I", pe, o + 28)[0]
size = struct.unpack_from("<I", pe, o + 56)[0]
s = o + struct.unpack_from("<H", pe, h + 20)[0]
cpu = Uc(UC_ARCH_X86, UC_MODE_32)
cpu.mem_map(base, (size + 4095) & ~4095)
for i in range(struct.unpack_from("<H", pe, h + 6)[0]):
    _, virtual, length, raw = struct.unpack_from("<IIII", pe, s + 40*i + 8)
    cpu.mem_write(base + virtual, pe[raw:raw+length])
cpu.mem_map(0x20000000, 0x20000)
PKT, DR, DATA, IMAGES, VP = 0x20000020, 0x20001020, 0x20003020, 0x20004020, 0x20005020
STACK, STOP = 0x20018000, 0x2001f000
images = [0x30000010, 0x30000020, 0x30000030]
ranges = {0x48ea70: 0x4942c5-0x48ea70, 0x4072e0:47, 0x43c700:6,
          0x578b30:8, 0x578b70:11, 0x45a9a0:9, 0x4bbd60:400}
sealed = {at: bytes(cpu.mem_read(at,n)) for at,n in ranges.items()}
tables = {0x4942d4:bytes(cpu.mem_read(0x4942d4,832)), 0x4bbef0:bytes(cpu.mem_read(0x4bbef0,24))}
entries, services, faults = collections.Counter(), collections.Counter(), []
events = []
inp = {}

def put(at,value):
    cpu.mem_write(at,struct.pack("<I",value & 0xffffffff))

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
    if address==0x45a9a0:
        assert args(1)==(DR,)
        events.append(["deactivate"])
    if any(at<=address<at+n for at,n in ranges.items()):
        return
    if address==0x470a80:
        services["packet-ui-counter"]+=1
        ret()
    elif address in (0x45a6f0,0x45a720):
        assert args(1)==(0x1234,)
        assert (address==0x45a720)==inp["static"]
        services["lookup-static" if inp["static"] else "lookup-dynamic"]+=1
        ret(0 if inp["shape"]=="missing" else DR)
    elif address==0x45a4e0:
        assert args(1)==(DR,)
        events.append(["delete"])
        services["delete"]+=1
        ret()
    elif address==0x4c4770:
        vp,dr,img=args(3)
        assert vp==VP and dr==DR and img in images
        events.append(["draw",images.index(img)])
        services["draw"]+=1
        ret()
    elif address!=STOP:
        raise AssertionError("undeclared original execution "+hex(address))

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
    for at,n in ((PKT,3),(DR,512),(DATA,16),(IMAGES,12),(VP,48)):
        cpu.mem_write(at-16,b"\x5a"*16)
        cpu.mem_write(at+n,b"\xa5"*16)
    cpu.mem_write(PKT,struct.pack("<BH",p["opcode"],0x1234 | (0x8000 if p["static"] else 0)))
    put(DR+112,0x20400000 if p["static"] else 0)
    put(DR+120,0x01000004)
    put(DR+300,0x4bbd60 if p["shape"]!="foreign" else 0x4bbd50)
    put(DR+304,0 if p["shape"]=="nil-data" else DATA)
    put(DR+316,p.get("start",100))
    put(DATA,16)
    put(DATA+4,IMAGES)
    cpu.mem_write(DATA+8,bytes((3,p.get("delay",0))))
    put(DATA+12,p["kind"])
    cpu.mem_write(IMAGES,struct.pack("<III",*images))
    put(0x69b828,int(p["connected"]))
    put(0x85b7a0,0)
    put(0x59abc4,0)
    put(0x852980,0)
    return {at:bytes(cpu.mem_read(at-16,n+32)) for at,n in ((PKT,3),(DR,512),(DATA,16),(IMAGES,12),(VP,48))}

def invoke(at,arguments):
    cpu.mem_write(STACK,struct.pack("<"+"I"*(len(arguments)+1),STOP,*arguments))
    cpu.reg_write(UC_X86_REG_ESP,STACK)
    cpu.emu_start(at,STOP,count=50000)
    assert cpu.reg_read(UC_X86_REG_EIP)==STOP and cpu.reg_read(UC_X86_REG_ESP)==STACK+4
    unchanged_code()
    return cpu.reg_read(UC_X86_REG_EAX)

def packet(p):
    before=setup(p)
    result=invoke(0x48ea70,(31,PKT,3))
    assert result==1
    expected=[]
    if p["connected"] and p["shape"]!="missing" and not (p["shape"] not in ("nil-data","foreign") and p["kind"]==1):
        expected=[["deactivate" if p["static"] else "delete"]]
    assert events==expected
    counters=[get(DR+x) for x in (480,484,488)]
    expect_counter=1 if p["connected"] and p["shape"]!="missing" and p["opcode"]==51 else 0
    assert counters==[expect_counter]*3
    flags=get(DR+120)
    assert flags==(0x01000000 if expected==[["deactivate"]] else 0x01000004)
    mutable={16+x for x in range(120,124)}|{16+x for x in range(480,492)}
    after=bytes(cpu.mem_read(DR-16,544))
    assert all(a==b for i,(a,b) in enumerate(zip(before[DR],after)) if i not in mutable)
    for at in (PKT,DATA,IMAGES,VP):
        assert bytes(cpu.mem_read(at-16,len(before[at])))==before[at]
    return dict(input=p,returned=result,actions=list(events),flags=flags,counters=counters,guards_and_other_input_unchanged=True)

rows=[]
for opcode in (49,50,51):
    for static in (False,True):
        for connected in (False,True):
            for shape,kind in [("kind-"+str(i),i) for i in range(6)]+[("nil-data",1),("foreign",1),("missing",1)]:
                rows.append(packet(dict(opcode=opcode,static=static,connected=connected,shape=shape,kind=kind)))

timelines=[]
for opcode in (49,50,51):
    for static in (False,True):
        for delay in (0,2):
            for start in (100,0xfffffffc):
                p=dict(opcode=opcode,static=static,connected=True,shape="kind-1",kind=1,delay=delay,start=start)
                packet_row=packet(p)
                ticks=[]
                data_before=bytes(cpu.mem_read(DATA-16,48))
                for elapsed in range(3*(delay+1)+1):
                    events=[]
                    frame=(start+elapsed)&0xffffffff
                    put(0x84ea04,frame)
                    result=invoke(0x4bbd60,(VP,DR))
                    expected=[["draw",elapsed//(delay+1)]] if elapsed<3*(delay+1) else [["delete"]]
                    assert events==expected and result==int(expected[0][0]=="draw")
                    assert bytes(cpu.mem_read(DATA-16,48))==data_before
                    ticks.append(dict(elapsed=elapsed,frame=frame,returned=result,actions=list(events)))
                timelines.append(dict(input=p,packet=packet_row,ticks=ticks))
assert not faults
print(json.dumps(dict(original_sha256=hashlib.sha256(pe).hexdigest(),
    functions={hex(at):dict(bytes=n,sha256=hashlib.sha256(sealed[at]).hexdigest()) for at,n in ranges.items()},
    table_hashes={hex(at):hashlib.sha256(data).hexdigest() for at,data in tables.items()},
    packet_rows=rows,timelines=timelines,unhooked_entries=dict(entries),declared_services=dict(services),
    memory_faults=faults,executable_and_table_bytes_unchanged=True,
    scope="unchanged whole packet decoder, stack probe, status, code namespace, static deactivation and AnimateDraw; declared lookup/UI counter/draw/delete services; not Windows transport/raster"),separators=(",",":")))
