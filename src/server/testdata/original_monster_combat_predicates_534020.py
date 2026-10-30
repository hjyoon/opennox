"""Execute thirteen unchanged pure predicates and their real internal callees.

Range outputs use independent native x86 execution of all20 unchanged bytes.
Unicorn's two SNaN conversion discrepancies remain explicit in the proof.
No predicate result is modeled or stubbed. Heap and stack guards, callee-saved
registers, original code/data, return width and x87 depth are observed.
"""
import argparse
import gzip
import hashlib
import itertools
import json
import struct
from pathlib import Path

from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_WRITE
from unicorn.x86_const import *

parser = argparse.ArgumentParser(description=__doc__)
for name in ['game','fixture','proof']:
    parser.add_argument('--'+name,type=Path,required=True)
parser.add_argument('--control',type=lambda v:int(v,0),choices=[0xe3f,0x23f],required=True)
parser.add_argument('--native32',type=Path,required=True)
parser.add_argument('--native64',type=Path,required=True)
args = parser.parse_args()
blob = args.game.read_bytes()
sha = lambda data:hashlib.sha256(data).hexdigest()
assert sha(blob)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
stage = Path(__file__).resolve().parent
audit = json.loads((stage/'original-predicates-readonly-v1.json').read_text())
instructions = {int(v['address'],16):bytes.fromhex(v['bytes'])
                for row in audit['functions'] for v in row['instructions']}
assert sum(map(len,instructions.values()))==598
entries = [int(row['address'],16)for row in audit['functions']]
native_sources = [
    (args.native32,32,'full-original-pe32',20,'f2e8bf24caaa0815c4f10c455b427f0b0a371886d56b2bf806e30b679c5914bb'),
    (args.native64,64,'original-fld-only-control',3,'eaea14051d55b57e644db0ac7673eb87c1745c1c6f408bc290e6ab90fc3938c8'),
]
native_values = {}
native_hashes = {}
original_range = b''.join(instructions[address]for address in sorted(instructions)if 0x534470<=address<0x534484)
assert len(original_range)==20 and sha(original_range)=='d300cac309459429773283935c5ec28f306e649f02fd63fc6931be92706e87a3'
for path,bits,mode,executed,expected_sha in native_sources:
    raw=gzip.decompress(path.read_bytes())
    assert sha(raw)==expected_sha
    native_hashes[str(bits)]=sha(raw)
    native=[json.loads(line)for line in raw.splitlines()]
    meta=native[0]
    assert len(native)==239 and meta['pointer_bits']==bits and meta['mode']==mode
    assert meta['executed_instruction_bytes']==executed and bytes.fromhex(meta['original_hex'])==original_range
    values={}
    for row in native[1:]:
        control,value=int(row['control'],16),int(row['input'],16)
        assert control in [0xe3f,0x23f] and int(row['loaded_control'],16)==control|0x40
        assert (int(row['returned_status'],16)>>11)&7==7 and (int(row['stored_status'],16)>>11)&7==0
        assert (control,value)not in values
        values[control,value]=int(row['output'],16)
    assert len(values)==238
    native_values[bits]=values
assert native_values[32]==native_values[64]
native_provenance=dict(ci_run_id=41108651935,ci_source='05ca2ad472591f6567e4e89ecad5417c3dc191c5',
    original_bytes=20,original_sha256=sha(original_range),range_contexts=119,
    native32_result_sha256=native_hashes['32'],native64_result_sha256=native_hashes['64'],
    source='native i386 execution of the unchanged complete original root; x86_64 FLD-only independent control')
native_differences=[]
pe = struct.unpack_from('<I',blob,60)[0]
optional = pe+24
base = struct.unpack_from('<I',blob,optional+28)[0]
table = optional+struct.unpack_from('<H',blob,pe+20)[0]
u = Uc(UC_ARCH_X86,UC_MODE_32)
u.mem_map(base,(struct.unpack_from('<I',blob,optional+56)[0]+4095)&~4095)
for i in range(struct.unpack_from('<H',blob,pe+6)[0]):
    _,va,size,offset = struct.unpack_from('<IIII',blob,table+i*40+8)
    if size:
        u.mem_write(base+va,blob[offset:offset+size])
heap,stack,code = 0x21000000,0x22000000,0x23000000
for address in [heap,stack,code]:
    u.mem_map(address,0x10000)
unit,update,definition,vector = heap+0x1000,heap+0x3000,heap+0x6000,heap+0x8000
sp,stop,init,cw,out = stack+0xe000,code+0x100,code+0x200,code+0x300,code+0x400
u.mem_write(init,bytes([0xdb,0xe3,0xd9,0x2d])+struct.pack('<I',cw))
u.mem_write(cw,struct.pack('<H',args.control))
u.mem_write(stop,bytes([0xdd,0x1d])+struct.pack('<I',out)+bytes([0x90]))
rd = lambda address:struct.unpack('<I',u.mem_read(address,4))[0]
wr = lambda address,value:u.mem_write(address,struct.pack('<I',value&0xffffffff))
constants = {address:bytes(u.mem_read(address,4))for address in
    [0x583bcc,0x58307c,0x583a9c,0x583d44,0x583d48,0x583d4c]}
directions = bytes(u.mem_read(0x5b6658,2048))
seed = bytes((i*31+19)&255 for i in range(0x10000))
callee_saved = {UC_X86_REG_EBX:0x12345678,UC_X86_REG_ESI:0x23456789,
                UC_X86_REG_EDI:0x3456789a,UC_X86_REG_EBP:0x456789ab}
active = False
current = {}
traces,call_counts = {},{}
return_top = None
unexpected_writes = []

def hook_code(_u,pc,size,_):
    global return_top
    if not active:
        return
    if pc==stop:
        return_top = (u.reg_read(UC_X86_REG_FPSW)>>11)&7
        if current['entry']!=0x534470:
            u.emu_stop()
        return
    if pc==stop+6:
        assert current['entry']==0x534470
        u.emu_stop()
        return
    assert pc in instructions,(current['name'],hex(pc))
    raw = bytes(u.mem_read(pc,size))
    assert raw==instructions[pc]
    traces[pc]=raw
    if pc in [0x534280,0x534300] and current['entry']==0x534220:
        call_counts[hex(pc)]=call_counts.get(hex(pc),0)+1

def hook_write(_u,_access,address,size,value,_):
    if not active:
        return
    if stack<=address and address+size<=stack+0x10000:
        return
    if current['entry']==0x534470 and out<=address and address+size<=out+8 and u.reg_read(UC_X86_REG_EIP)==stop:
        return
    unexpected_writes.append([hex(address),size,value])
    raise AssertionError((current['name'],'unexpected pure-predicate write',hex(address),size))

u.hook_add(UC_HOOK_CODE,hook_code)
u.hook_add(UC_HOOK_MEM_WRITE,hook_write)
defaults = dict(cls=2,subclass=0,status=0,weapons=0,armor=0,range_bits=0x41200000,
                missile_first=0,speed_bits=0x3dcccccd,aggression_bits=0x3f000000,
                head=0,action=0,prefix528=0,direction=0,vector=[0x3f800000,0])
cases = []
case_counts = {}

def add(entry,values):
    value=dict(defaults,**values)
    value['vector']=list(value['vector'])
    index=case_counts.get(entry,0)
    case_counts[entry]=index+1
    cases.append(dict(name=f'{entry:06x}/{index:05d}',entry=entry,input=value))

def float_corpus(centers):
    values={0,0x80000000,1,0x80000001,0x007fffff,0x807fffff,0x00800000,0x80800000,
            0x3f800000,0xbf800000,0x7f7fffff,0xff7fffff,0x7f800000,0xff800000,
            0x7fc00000,0xffc00000,0x7fc12345,0xffc12345,0x7f800001,0xff800001}
    for center in centers:
        values.update(center+delta for delta in range(-16,17))
    return sorted(values)

word_bits=[0,0xffffffff]+[1<<i for i in range(32)]
for cls,subclass in itertools.product([0,2,4,6,0xffffffff],word_bits+[0xfffffbff,0x400]):
    add(0x534020,dict(cls=cls,subclass=subclass))
vectors=[(0x3f800000,0),(0,0x3f800000),(0xbf800000,0),(0,0xbf800000),(0,0),
         (0x3f666666,0),(0x3f666665,0),(0x3f666667,0),(0x3f000000,0x3f000000),
         (0x3f800000,0x3f800000),(0xbf800000,0xbf800000),
         (0x7f800000,0),(0,0xff800000),(0x7fc00000,0x3f800000),
         (0x3f800000,0x7f800001),(0xffc12345,0x7fc12345),
         (1,0x80000001),(0x7f7fffff,0xff7fffff),(0x80000000,0)]
for direction,vec in itertools.product(range(256),vectors):
    add(0x534120,dict(direction=direction,vector=vec))
for subclass,status,weapons,value in itertools.product([0,0x10],[0,0x20,4,0xffffffff,0xffffffdf],
        [0,1,2,0x47f00fe,0xfb80ff01],float_corpus([0x00000010,0x3f800000])):
    add(0x534220,dict(subclass=subclass,status=status,weapons=weapons,range_bits=value))
for subclass,weapons,first in itertools.product([0,0x10,0xffffffff,0xffffffef],word_bits+[0x47f00fe,0xfb80ff01],[0,1,127,128,255]):
    add(0x534280,dict(subclass=subclass,weapons=weapons,missile_first=first))
for subclass,armor,status in itertools.product([0,0x10,0xffffffff,0xffffffef],word_bits+[0x3000000,0xfcffffff],[0,1,4,0xffffffff,0xfffffffb]):
    add(0x5342c0,dict(subclass=subclass,armor=armor,status=status))
for subclass,status in itertools.product([0,0x10,0xffffffff],word_bits+[0xffffffdf]):
    add(0x534300,dict(subclass=subclass,status=status))
for cls,value in itertools.product([2,6,0xffffffff],float_corpus([0x3c23d70a])):
    add(0x534320,dict(cls=cls,speed_bits=value))
for value in [0,1,4,8,23,25,26,27,0xffffffff,0x80000000]:
    add(0x534340,dict(head=-1,prefix528=value))
for head,action in itertools.product([0,1,7,23],list(range(81))+[0xffffffff,0x80000000]):
    add(0x534340,dict(head=head,action=action,prefix528=0x12345678))
aggression = float_corpus([0x3da3d70a,0x3ea8f5c3,0x3f28f5c3])
for entry,value in itertools.product([0x534390,0x5343c0,0x534400,0x534440],aggression):
    add(entry,dict(aggression_bits=value))
for value in aggression:
    add(0x534470,dict(range_bits=value))

rows=[]
for index,case in enumerate(cases):
    current=case
    c=case['input']
    active=False
    u.mem_write(heap,seed)
    wr(unit+8,c['cls']);wr(unit+12,c['subclass']);wr(unit+748,update)
    wr(unit+548,c['speed_bits'])
    u.mem_write(unit+124,struct.pack('<h',c['direction']))
    wr(update+484,definition);wr(update+528,c['prefix528'])
    u.mem_write(update+544,struct.pack('<b',c['head']))
    if c['head']>=0:
        wr(update+552+24*c['head'],c['action'])
    wr(update+1304,c['aggression_bits']);wr(update+1440,c['status'])
    wr(update+2056,c['weapons']);wr(update+2060,c['armor'])
    wr(definition+112,c['range_bits'])
    u.mem_write(definition+148,bytes([c['missile_first']]))
    wr(vector,c['vector'][0]);wr(vector+4,c['vector'][1])
    before=bytes(u.mem_read(heap,0x10000))
    u.mem_write(stack,seed)
    u.mem_write(sp,struct.pack('<III',stop,unit,vector))
    low_guard=bytes(u.mem_read(sp-256,128));high_guard=bytes(u.mem_read(sp+32,128))
    for register,value in callee_saved.items():
        u.reg_write(register,value)
    u.reg_write(UC_X86_REG_EAX,0x76543210)
    u.reg_write(UC_X86_REG_ECX,0x87654321)
    u.reg_write(UC_X86_REG_EDX,0x98765432)
    u.reg_write(UC_X86_REG_EFLAGS,2)
    u.reg_write(UC_X86_REG_ESP,sp)
    u.emu_start(init,init+8,count=3)
    assert u.reg_read(UC_X86_REG_FPCW)==args.control
    return_top=None
    active=True
    u.emu_start(case['entry'],stop+16,count=1000)
    active=False
    assert u.reg_read(UC_X86_REG_EIP) in [stop,stop+6],case['name']
    assert u.reg_read(UC_X86_REG_ESP)==sp+4,case['name']
    assert u.reg_read(UC_X86_REG_FPCW)==args.control,case['name']
    assert return_top==(7 if case['entry']==0x534470 else 0),case['name']
    assert (u.reg_read(UC_X86_REG_FPSW)>>11)&7==0,case['name']
    assert all(u.reg_read(register)==value for register,value in callee_saved.items()),case['name']
    assert bytes(u.mem_read(heap,0x10000))==before,case['name']
    assert bytes(u.mem_read(sp-256,128))==low_guard and bytes(u.mem_read(sp+32,128))==high_guard,case['name']
    row=dict(case)
    if case['entry']==0x534470:
        emulated=struct.unpack('<Q',u.mem_read(out,8))[0]
        native=native_values[32][args.control,c['range_bits']]
        row['return_f64_bits']=native
        if emulated!=native:
            native_differences.append(dict(name=case['name'],input_bits=c['range_bits'],
                Unicorn_f64_bits=emulated,native_original_f64_bits=native))
    else:
        row['return_u32']=u.reg_read(UC_X86_REG_EAX)
        assert row['return_u32'] in [0,1],case['name']
    rows.append(row)
assert traces==instructions,('uncovered original instructions',sorted(set(instructions)-set(traces)))
assert not unexpected_writes and call_counts.keys()=={'0x534280','0x534300'}
assert all(bytes(u.mem_read(address,len(raw)))==raw for address,raw in instructions.items())
assert all(bytes(u.mem_read(address,4))==raw for address,raw in constants.items())
assert bytes(u.mem_read(0x5b6658,2048))==directions
assert len(native_differences)==2 and {row['input_bits']for row in native_differences}=={0x7f800001,0xff800001}
assert {row['input']['range_bits']for row in rows if row['entry']==0x534470}=={
    value for control,value in native_values[32]if control==args.control}
fixture=dict(schema='original-monster-combat-predicates-534020-v2',original_sha256=sha(blob),
             control_word=args.control,native_range_reference=native_provenance,rows=rows)
raw=(json.dumps(fixture,separators=(',',':'))+'\n').encode()
assert not args.fixture.exists() and not args.proof.exists()
with args.fixture.open('xb')as stream:
    with gzip.GzipFile(filename='',mode='wb',fileobj=stream,mtime=0)as compressed:
        compressed.write(raw)
proof=dict(status='passed',original_sha256=sha(blob),control_word=args.control,roots=len(entries),
           cases=len(rows),per_entry={hex(entry):sum(c['entry']==entry for c in cases)for entry in entries},
           instruction_bytes_executed=sum(map(len,traces.values())),actual_internal_callee_entries=call_counts,
           modeled_services=[],original_code_and_data_unchanged=True,
           native_range_reference=native_provenance,Unicorn_range_discrepancies=native_differences,
           heap_stack_register_control_word_and_TOP_guards=True,
           original_data=dict(direction_table=dict(address=hex(0x5b6658),bytes=2048,sha256=sha(directions)),
               constants={hex(address):dict(bytes=4,hex=raw.hex(),sha256=sha(raw))for address,raw in constants.items()}),
           fixture=dict(path=str(args.fixture),bytes=args.fixture.stat().st_size,
               sha256=sha(args.fixture.read_bytes()),decoded_bytes=len(raw),decoded_sha256=sha(raw)))
with args.proof.open('x')as stream:
    json.dump(proof,stream,indent=2)
    stream.write('\n')
print(json.dumps(dict(status='passed',control=hex(args.control),cases=len(rows),
    original_instruction_bytes=proof['instruction_bytes_executed'],actual_internal_calls=call_counts)))
