from pathlib import Path
import json,struct,hashlib,datetime,collections
from unicorn import Uc,UC_ARCH_X86,UC_MODE_32,UC_HOOK_CODE
from unicorn.x86_const import *
from capstone import Cs,CS_ARCH_X86,CS_MODE_32
p=Path(__file__).resolve().parent;game=Path('/Users/ghost/codex-sandbox/tmp/opennox-warrior-r318-20261019/save-load-v2-normal-virtual/runtime-data/GAME.EXE');blob=game.read_bytes();sha=lambda b:hashlib.sha256(b).hexdigest();assert sha(blob)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,60)[0];opt=pe+24;base=struct.unpack_from('<I',blob,opt+28)[0];tab=opt+struct.unpack_from('<H',blob,pe+20)[0];u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,(struct.unpack_from('<I',blob,opt+56)[0]+4095)&~4095)
for i in range(struct.unpack_from('<H',blob,pe+6)[0]):
 _,rva,size,off=struct.unpack_from('<IIII',blob,tab+40*i+8)
 if size:u.mem_write(base+rva,blob[off:off+size])
heap,stack,stop=0x21000000,0x22000000,0x23000000
for addr in [heap,stack,stop]:u.mem_map(addr,0x10000)
sp=stack+0xe000;dr=heap+0x1000;functions=[];instructions={};dis=Cs(CS_ARCH_X86,CS_MODE_32)
for start,end in [(0x4756e0,0x475740),(0x475740,0x4757a0)]:
 raw=bytes(u.mem_read(start,end-start));ins=[dict(address=hex(v.address),bytes=v.bytes.hex(),op=v.mnemonic,args=v.op_str)for v in dis.disasm(raw,start)];functions.append(dict(address=hex(start),size=len(raw),sha256=sha(raw),instructions=ins));instructions.update({int(v['address'],16):bytes.fromhex(v['bytes'])for v in ins})
seen=set();active=False
def hook(_u,pc,size,_):
 if not active:return
 assert pc in instructions,hex(pc);assert bytes(u.mem_read(pc,size))==instructions[pc];seen.add(pc)
u.hook_add(UC_HOOK_CODE,hook)
saved={UC_X86_REG_EBX:0x12345678,UC_X86_REG_ESI:0x23456789,UC_X86_REG_EDI:0x3456789a,UC_X86_REG_EBP:0x456789ab}
rows=[]
for mode,callback in [('nil',0),('static',0x4bcc20),('random',0x4bcc60),('other1',1),('other2',0xffffffff)]:
 for f in range(32):
  flags=sum(bit for i,bit in enumerate([1,8,0x40,0x800,0x1000])if f&(1<<i))
  for c in range(8):
   cls=sum(bit for i,bit in enumerate([0x80000000,0x800000,0x400000])if c&(1<<i));returns=[]
   u.mem_write(dr,b'\xa5'*1024)
   for off,val in [(112,cls),(120,flags),(300,callback)]:u.mem_write(dr+off,struct.pack('<I',val))
   before=bytes(u.mem_read(heap,0x10000))
   for entry in [0x4756e0,0x475740]:
    u.mem_write(sp-256,b'\xcd'*512);u.mem_write(sp,struct.pack('<II',stop,dr));u.reg_write(UC_X86_REG_ESP,sp);u.reg_write(UC_X86_REG_EAX,0xdeadface);u.reg_write(UC_X86_REG_EFLAGS,0x202)
    for reg,val in saved.items():u.reg_write(reg,val)
    active=True
    try:u.emu_start(entry,stop,count=250)
    finally:active=False
    assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4
    assert all(u.reg_read(reg)==val for reg,val in saved.items())
    assert bytes(u.mem_read(heap,0x10000))==before
    returns.append(u.reg_read(UC_X86_REG_EAX))
   rows.append(dict(callback=mode,cls=cls,flags=flags,returns=returns))
payload=(json.dumps(rows,separators=(',',':'))+'\n').encode()
with(p/'original-drawable-predicates-v1.json').open('xb')as f:f.write(payload)
proof=dict(status='passed',utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),source_GAME_sha256=sha(blob),contexts=len(rows),actual_original_calls=2*len(rows),functions=functions,executed_instruction_bytes=sum(len(instructions[a])for a in seen),expected_sha256=sha(payload),callback_bodies_called=False,scope='Unchanged original integer predicates. Complete 5 callback identities x32flags x8classes. Native warning candidate already partially edited before this independent reference execution; this is not a pre-edit C-before certificate or new full port closure.')
with(p/'original-drawable-predicates-qualified-v1.json').open('x')as f:json.dump(proof,f,indent=2);f.write('\n')
print(json.dumps({k:v for k,v in proof.items()if k!='functions'}),flush=True)
