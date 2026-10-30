"""Execute original schedulers and real push, reset, predicate and distance code.

Only related-spell selection and the diagnostic print sink are declared
services. Push success, stack changes, predicates and FP arithmetic execute
unchanged original instructions; none of their results is substituted.
"""
import argparse, collections, gzip, hashlib, itertools, json, struct
from pathlib import Path
from capstone import Cs, CS_ARCH_X86, CS_MODE_32
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_WRITE
from unicorn.x86_const import *

parser=argparse.ArgumentParser(description=__doc__)
for name in ['game','fixture','proof']:parser.add_argument('--'+name,type=Path,required=True)
parser.add_argument('--control',type=lambda v:int(v,0),choices=[0xe3f,0x23f],required=True)
parser.add_argument('--limit',type=int,default=0,help='Development smoke prefix; zero executes all cases.')
args=parser.parse_args()
blob=args.game.read_bytes();sha=lambda raw:hashlib.sha256(raw).hexdigest()
assert sha(blob)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,60)[0];optional=pe+24
base=struct.unpack_from('<I',blob,optional+28)[0]
size=(struct.unpack_from('<I',blob,optional+56)[0]+4095)&~4095
table=optional+struct.unpack_from('<H',blob,pe+20)[0]
u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,size)
for i in range(struct.unpack_from('<H',blob,pe+6)[0]):
    _,rva,length,offset=struct.unpack_from('<IIII',blob,table+i*40+8)
    if length:u.mem_write(base+rva,blob[offset:offset+length])
intervals=[(0x531b40,208),(0x531c10,80),(0x531c60,240),(0x531d50,208),
           (0x4e6c00,212),(0x4ff350,35),(0x534220,83),(0x534280,58),
           (0x534300,23),(0x534710,52),(0x50a260,251),(0x50a010,13),
           (0x50a110,76),(0x4e39d0,23)]
md=Cs(CS_ARCH_X86,CS_MODE_32);instructions={};code_seals=[]
for address,length in intervals:
    raw=bytes(u.mem_read(address,length));decoded=list(md.disasm(raw,address))
    assert sum(v.size for v in decoded)==length
    instructions.update({v.address:bytes(v.bytes) for v in decoded})
    code_seals.append(dict(address=hex(address),size=length,sha256=sha(raw)))
heap,stack,code=0x21000000,0x22000000,0x23000000
u.mem_map(heap,0x20000);u.mem_map(stack,0x10000);u.mem_map(code,0x1000)
unit,target=heap+0x1000,heap+0x3000
updates=[heap+0x5000,heap+0x7000];definitions=[heap+0x9000,heap+0xa000]
type_table,type_record,type_name=heap+0xc000,heap+0xd000,heap+0xd100
sp,stop,init,cw=stack+0xe000,code+0x100,code+0x200,code+0x300
rd=lambda address:struct.unpack('<I',u.mem_read(address,4))[0]
wr=lambda address,value:u.mem_write(address,struct.pack('<I',value&0xffffffff))
u.mem_write(init,bytes([0xdb,0xe3,0xd9,0x2d])+struct.pack('<I',cw))
u.mem_write(cw,struct.pack('<H',args.control))
frame_address,changed_address=0x84ea04,0x75ae60
cancel_pc=code+0x500
u.mem_write(cancel_pc,b'\xc3')
cancel_actions=[0,1,4,7,31,39,40,60,0xffffffff]
cancel_table_addresses={a:(0x5bfed4+((a<<4)&0xffffffff))&0xffffffff for a in cancel_actions}
original_cancel_table={address:bytes(u.mem_read(address,4)) for address in cancel_table_addresses.values()}
cancels=[]
entry_addresses={address for address,_ in intervals}
saved={UC_X86_REG_EBX:0x12345678,UC_X86_REG_ESI:0x23456789,
       UC_X86_REG_EDI:0x3456789a,UC_X86_REG_EBP:0x456789ab}
reset_fields=[(8,4),(268,4),(296,4),(364,4),(481,1),(482,1),(483,1),(496,4),(548,4)]
allowed=[]
for update in updates:
    allowed.extend((update+offset,length)for offset,length in reset_fields)
    allowed.extend([(update+544,1),(update+552,24*24)])
active=False;current=None;trace=set();calls=collections.Counter();pushes=[];related=[];logs=0

def role(value):
    if value==0:return dict(kind='nil')
    if value==1:return dict(kind='sentinel',value=1)
    if value==target:return dict(kind='target')
    for index,update in enumerate(updates):
        if update+552<=value<update+552+24*24 and (value-update-552)%24==0:
            return dict(kind='slot',update=index,index=(value-update-552)//24)
    return dict(kind='u32',value=value)

def service_return(value):
    pointer=u.reg_read(UC_X86_REG_ESP)
    u.reg_write(UC_X86_REG_EAX,value)
    u.reg_write(UC_X86_REG_EIP,rd(pointer))
    u.reg_write(UC_X86_REG_ESP,pointer+4)

def hook_code(_u,pc,length,_):
    global logs
    if not active:return
    if pc==stop:u.emu_stop();return
    pointer=u.reg_read(UC_X86_REG_ESP)
    if pc==cancel_pc:
        assert rd(pointer+4)==unit
        cached=u.reg_read(UC_X86_REG_ESI)
        assert cached in updates
        cached_index=updates.index(cached)
        effect=current['input']['cancel_effect']
        if effect in ['swap-update','swap-update-and-old-range']:
            wr(unit+748,updates[1-cached_index])
        if effect=='swap-update-and-old-range':
            wr(definitions[0]+212,current['input']['cancel_range'])
        if effect=='cached-range':wr(definitions[cached_index]+212,current['input']['cancel_range'])
        if effect=='cap-cached-head':u.mem_write(cached+544,b'\x16')
        if effect=='move-target':
            for index,bits in enumerate(current['input']['cancel_target_pos']):wr(target+56+index*4,bits)
        assert effect in ['none','swap-update','swap-update-and-old-range','cached-range','cap-cached-head','move-target']
        action=rd(cached+552+struct.unpack('b',u.mem_read(cached+544,1))[0]*24)
        cancels.append(dict(cached_update=cached_index,live_update_after=updates.index(rd(unit+748)),effect=effect))
        service_return(0);return
    if pc==0x540d90:
        assert rd(pointer+4)==unit and rd(pointer+8)==target
        mutation=current['input']['related_mutation']
        if mutation in ['replace-update','replace-update-and-old-range']:
            wr(unit+748,updates[1])
        if mutation=='replace-update-and-old-range':
            wr(definitions[0]+212,current['input']['old_range_after_related'])
        value=current['input']['related_return']
        related.append(dict(unit_update_after=updates.index(rd(unit+748)),return_u32=value,mutation=mutation))
        service_return(value);return
    if pc==0x5341a0:
        logs+=1
        service_return(0);return
    assert pc in instructions,(current['name'],'unexpected original PC',hex(pc))
    assert bytes(u.mem_read(pc,length))==instructions[pc]
    trace.add(pc)
    if pc in entry_addresses:calls[hex(pc)]+=1
    if pc==0x50a260:
        assert rd(pointer+4)==unit
        pushes.append(dict(action=rd(pointer+8),unit_update=updates.index(rd(unit+748))))
    if pc in [0x50a27b,0x50a28d,0x50a2aa,0x50a35a]:
        assert pushes and 'return_value'not in pushes[-1]
        pushes[-1]['return_value']=role(u.reg_read(UC_X86_REG_EAX))

def hook_write(_u,_access,address,length,value,_):
    if not active:return
    if stack<=address and address+length<=stack+0x10000:return
    if address==changed_address and length==4:return
    if any(start<=address and address+length<=start+size for start,size in allowed):return
    raise AssertionError((current['name'],'unexpected original write',hex(address),length,value))

u.hook_add(UC_HOOK_CODE,hook_code);u.hook_add(UC_HOOK_MEM_WRITE,hook_write)
defaults=dict(cls=2,subclass=0,buffs=1<<29,status=0,weapons=0,head=-1,action=1,
              root_action=1,melee_bits=0x41200000,missile_bits=0x42700000,missile_first=65,
              alt_status=0x400,alt_weapons=0,alt_missile_first=0,alt_melee_bits=0x41200000,
              alt_missile_bits=0x41a00000,unit_pos=[0,0],target_pos=[0x42200000,0],
              unit_shape=0,target_shape=0,unit_radius=0,target_radius=0,
              unit_width=0x40000000,unit_height=0x40800000,target_width=0x40000000,
              target_height=0x40800000,related_return=0,related_mutation='none',
              old_range_after_related=0x3f800000,initial_changed=0,seed=19)
cases=[];counts=collections.Counter()
def add(entry,**values):
    data=dict(defaults,**values);index=counts[entry];counts[entry]+=1
    cases.append(dict(entry=entry,name=f'{entry:06x}/{index:05d}',input=data))
roots=[0x531b40,0x531c10,0x531c60,0x531d50]
for entry in roots:add(entry)
for entry,cls,head,action,root_action in itertools.product(roots,[0,2,4,6,0xffffffff],[-1,0,1,20,21,22,23],[0,1,4,23,26,31,39,60,0xffffffff],[1,31]):
    add(entry,cls=cls,head=head,action=action,root_action=root_action)
word_bits=[0,0xffffffff]+[1<<i for i in range(32)]
for entry,subclass,status in itertools.product(roots,[0,16,0xffffffff],word_bits):
    add(entry,subclass=subclass,status=status)
for entry,weapons in itertools.product(roots,word_bits):add(entry,subclass=16,weapons=weapons)
for entry,head in itertools.product(roots,range(-1,24)):add(entry,head=head,seed=83,initial_changed=1)
float_bits={0,0x80000000,1,2,0x80000001,0x007fffff,0x00800000,0x7f7fffff,
            0xff7fffff,0x7f800000,0xff800000,0x7fc00000,0xffc00000,
            0x7fc12345,0xffc12345,0x7f800001,0xff800001,0x3f800000,0xbf800000}
for center in [0x41200000,0x41f00000,0x42700000]:float_bits.update(center+d for d in range(-4,5))
for entry,bits in itertools.product([0x531b40,0x531c60,0x531d50],sorted(float_bits)):
    add(entry,missile_bits=bits)
    add(entry,melee_bits=bits,missile_first=0)
for value,mutation,buffs in itertools.product([0,1],['none','replace-update','replace-update-and-old-range'],[0,1<<29]):
    add(0x531b40,buffs=buffs,related_return=value,related_mutation=mutation)
for unit_shape,target_shape,width,height in itertools.product([0,2,3],[0,2,3],[0x3f800000,0x40a00000,0x7fc12345],[0x40000000,0x7fc12345]):
    add(0x531b40,unit_shape=unit_shape,target_shape=target_shape,unit_width=width,
        unit_height=height,target_width=height,target_height=width,unit_radius=0x41200000,
        target_radius=0x41a00000,target_pos=[0x41f00000,0])
base_contexts=len(cases)
for entry,head,action,present in itertools.product(roots,[-1,0,1,21,22,23],cancel_actions,[0,1]):
    add(entry,head=head,action=action,started=1,alt_started=0,cancel_present=present,cancel_effect='none',cancel_range=0x3f800000,cancel_target_pos=[0x42c80000,0x43480000])
for entry,effect,alt_started,cls in itertools.product(roots,['swap-update','swap-update-and-old-range','cached-range','cap-cached-head','move-target'],[0,1],[0,2,4,6]):
    add(entry,head=0,action=4,started=1,alt_started=alt_started,cls=cls,cancel_present=1,cancel_effect=effect,cancel_range=0x3f800000,cancel_target_pos=[0x42c80000,0x43480000],alt_status=0,alt_missile_first=65)
for related_mutation,effect,alt_started,related_return in itertools.product(['none','replace-update','replace-update-and-old-range'],['none','swap-update','cached-range'],[0,1],[0,1]):
    add(0x531b40,head=0,action=4,started=1,alt_started=alt_started,buffs=0,related_mutation=related_mutation,related_return=related_return,cancel_present=1,cancel_effect=effect,cancel_range=0x40000000,cancel_target_pos=[0x42c80000,0x43480000],alt_status=0,alt_missile_first=65)
if args.limit:cases=cases[:args.limit]

seed_bytes=bytes((i*31+19)&255 for i in range(0x20000))
protected=[];cursor=heap
for start,length in sorted(allowed+[(unit+748,4),(definitions[0]+212,4),(definitions[1]+212,4),(target+56,8)]):
    if start>cursor:protected.append((cursor,start-cursor))
    cursor=max(cursor,start+length)
if cursor<heap+len(seed_bytes):protected.append((cursor,heap+len(seed_bytes)-cursor))
rows=[]
for current in cases:
    data=current['input'];u.mem_write(heap,seed_bytes)
    for address,raw_table in original_cancel_table.items():u.mem_write(address,raw_table)
    if 'cancel_present'in data:
        for action,address in cancel_table_addresses.items():wr(address,cancel_pc if data['cancel_present']else 0)
    table_before={address:bytes(u.mem_read(address,4))for address in original_cancel_table}
    for address,length in [(unit,0x400),(target,0x400),(updates[0],0x1000),(updates[1],0x1000),(definitions[0],0x400),(definitions[1],0x400)]:u.mem_write(address,b'\0'*length)
    for address,which in [(unit,'unit'),(target,'target')]:
        u.mem_write(address+4,struct.pack('<H',1));wr(address+8,data['cls'])
        wr(address+12,data['subclass']);wr(address+36,117)
        wr(address+56,data[which+'_pos'][0]);wr(address+60,data[which+'_pos'][1])
        wr(address+172,data[which+'_shape']);wr(address+176,data[which+'_radius'])
        wr(address+184,data[which+'_width']);wr(address+188,data[which+'_height'])
    wr(unit+340,data['buffs']);wr(unit+748,updates[0])
    for index,update in enumerate(updates):
        wr(update+484,definitions[index]);wr(update+1440,data['status'if index==0 else'alt_status'])
        wr(update+2056,data['weapons'if index==0 else'alt_weapons'])
        u.mem_write(update+544,struct.pack('b',data['head']))
        for offset,length in reset_fields:u.mem_write(update+offset,bytes([0x71+index])*length)
        for slot in range(24):
            record=update+552+slot*24;wr(record,1)
            for argument in range(4):wr(record+4+argument*4,0x61000000+index*0x10000+slot*0x100+argument*0x10+data['seed'])
            wr(record+20,0)
        if data['head']>=0:
            wr(update+552+data['head']*24,data['action'])
            if data['head']>0:wr(update+552,data['root_action'])
            wr(update+552+data['head']*24+20,data.get('started'if index==0 else'alt_started',0))
        wr(definitions[index]+112,data['melee_bits'if index==0 else'alt_melee_bits'])
        wr(definitions[index]+212,data['missile_bits'if index==0 else'alt_missile_bits'])
        u.mem_write(definitions[index]+148,bytes([data['missile_first'if index==0 else'alt_missile_first']]))
    wr(type_table+4,type_record);wr(type_record+4,type_name);u.mem_write(type_name,b'SchedulerFixture\0')
    wr(0x7520d4,type_table);wr(frame_address,1400);wr(changed_address,data['initial_changed'])
    before=[bytes(u.mem_read(start,length))for start,length in protected]
    saved_changed_neighbors=bytes(u.mem_read(changed_address-16,36))
    u.reg_write(UC_X86_REG_ESP,sp);u.emu_start(init,init+8)
    for register,value in saved.items():u.reg_write(register,value)
    u.reg_write(UC_X86_REG_ESP,sp)
    wr(sp,stop);wr(sp+4,unit);wr(sp+8,target)
    pushes=[];related=[];cancels=[];logs=0;active=True
    try:u.emu_start(current['entry'],stop+1,count=100000)
    finally:active=False
    assert u.reg_read(UC_X86_REG_EIP)==stop,(current['name'],'did not return')
    assert u.reg_read(UC_X86_REG_ESP)==sp+4
    assert u.reg_read(UC_X86_REG_FPCW)==args.control
    assert (u.reg_read(UC_X86_REG_FPSW)>>11)&7==0
    assert all(u.reg_read(register)==value for register,value in saved.items())
    assert before==[bytes(u.mem_read(start,length))for start,length in protected]
    after_neighbors=bytes(u.mem_read(changed_address-16,36))
    assert saved_changed_neighbors[:16]==after_neighbors[:16] and saved_changed_neighbors[20:]==after_neighbors[20:]
    assert all('return_value'in call for call in pushes)
    assert table_before=={address:bytes(u.mem_read(address,4))for address in original_cancel_table}
    if not cancels or data.get('cancel_effect')!='move-target':assert [rd(target+56),rd(target+60)]==data['target_pos']
    states=[]
    for update in updates:
        states.append(dict(head=struct.unpack('b',u.mem_read(update+544,1))[0],
            reset_fields=[int.from_bytes(u.mem_read(update+offset,length),'little')for offset,length in reset_fields],
            slots=[dict(action=rd(update+552+i*24),args=[role(rd(update+552+i*24+4+j*4))for j in range(4)],started=rd(update+552+i*24+20))for i in range(24)]))
    rows.append(dict(**current,output=dict(return_value=role(u.reg_read(UC_X86_REG_EAX)),
        unit_update=updates.index(rd(unit+748)),updates=states,stack_changed=rd(changed_address),
        old_range_after=rd(definitions[0]+212),pushes=pushes,related=related,diagnostic_calls=logs)))
    if 'cancel_present'in data:
        rows[-1]['output'].update(cancels=cancels,target_pos_after=[rd(target+56),rd(target+60)],definition_ranges_after=[rd(d+212)for d in definitions])

fixture=dict(schema='original-monster-fight-schedulers-531b40-v2',original_sha256=sha(blob),
             control_word=args.control,frame=1400,rows=rows)
raw=json.dumps(fixture,separators=(',',':'),sort_keys=True).encode()
with args.fixture.open('xb') as target_file:
    with gzip.GzipFile(filename='',mode='wb',fileobj=target_file,mtime=0) as stream:stream.write(raw)
proof=dict(status='passed',contexts=len(rows),control_word=hex(args.control),development_prefix=bool(args.limit),
    code_intervals=code_seals,loaded_original_bytes=sum(length for _,length in intervals),
    executed_unique_instruction_bytes=sum(len(instructions[pc])for pc in trace),
    instruction_addresses_not_executed=[hex(pc)for pc in sorted(set(instructions)-trace)],
    original_entry_calls=dict(calls),per_root=dict(collections.Counter(hex(row['entry'])for row in rows)),
    declared_services={'23000500':'declared cancel callback, presence in initialized original action-table input; original push dispatch/cached state/reset still execute','540D90':'related spell selection returns declared canonical bool; optional declared live UpdateData/old-range mutation',
                       '5341A0':'diagnostic print sink; no gameplay writes'},
    push_reset_buff_distance_predicate_name_getter_originals_execute=True,
    no_FP_predicate_push_success_or_stack_change_result_substitution=True,
    heap_stack_code_nonvolatile_and_FPU_guards_passed=True,
    fixture_sha256=sha(args.fixture.read_bytes()),raw_fixture_sha256=sha(raw),raw_fixture_bytes=len(raw),
    original_unchanged=sha(args.game.read_bytes())==sha(blob),base_contexts=base_contexts,cancel_callbacks=sum(len(r['output'].get('cancels',[]))for r in rows),overall_porting_100_percent=False)
with args.proof.open('x') as output:json.dump(proof,output,indent=2);output.write('\n')
print(json.dumps({k:v for k,v in proof.items()if k not in ['code_intervals','instruction_addresses_not_executed']}))
