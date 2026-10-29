"""Execute complete unchanged Turn Undead roots, converter and table loads.

Balance return, name-to-type resolution, typed allocation, create-at, FX,
first/next-object traversal and delayed deletion are declared services. Name
resolution and allocation counts are observed. Creation writes declared type,
owner and position outputs; traversal returns the current links after deletion.
Original caller packing, arithmetic, conversions and table loads are untouched.
Destroy EAX is unused by its void consumer. Nil-source/nil-collision defensive
native policies are outside these valid original-root inputs.
"""
import argparse, collections, copy, hashlib, json, math, struct
from pathlib import Path
from unicorn import Uc,UC_ARCH_X86,UC_MODE_32,UC_HOOK_CODE,UC_HOOK_MEM_INVALID,UC_HOOK_MEM_READ
from unicorn.x86_const import *
p=argparse.ArgumentParser(description=__doc__)
for key in ['game','fixture','proof']:p.add_argument('--'+key,type=Path,required=True)
p.add_argument('--control',type=lambda v:int(v,0),choices=[0xe3f,0x23f],default=0xe3f)
a=p.parse_args();blob=a.game.read_bytes();sha=lambda b:hashlib.sha256(b).hexdigest()
assert sha(blob)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,60)[0];opt=pe+24;base=struct.unpack_from('<I',blob,opt+28)[0];table=opt+struct.unpack_from('<H',blob,pe+20)[0]
u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,(struct.unpack_from('<I',blob,opt+56)[0]+4095)&~4095)
for n in range(struct.unpack_from('<H',blob,pe+6)[0]):
    _,va,size,off=struct.unpack_from('<IIII',blob,table+40*n+8)
    if size:u.mem_write(base+va,blob[off:off+size])
heap,stack,code=0x21000000,0x22000000,0x23000000
for address in [heap,stack,code]:u.mem_map(address,0x20000)
record,other=heap+0x1000,heap+0x1800
objects=[heap+0x2000+0x400*i for i in range(46)]
collisions=[heap+0xe000+0x20*i for i in range(46)]
sp,stop,init,thunk,scratch,cw=stack+0x1e000,code+0x100,code+0x200,code+0x300,code+0x400,code+0x500
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw));u.mem_write(thunk,b'\xdd\x05'+struct.pack('<I',scratch)+b'\xc3')
get=lambda address:struct.unpack('<I',u.mem_read(address,4))[0]
put=lambda address,value:u.mem_write(address,struct.pack('<I',value&0xffffffff))
get16=lambda address:struct.unpack('<H',u.mem_read(address,2))[0]
put16=lambda address,value:u.mem_write(address,struct.pack('<H',value&0xffff))
words=lambda address:[get(address),get(address+4)]
f32=lambda v:struct.unpack('<I',struct.pack('<f',v))[0]
point=lambda x,y:[f32(x),f32(y)]
f64=lambda v:struct.pack('<d',v).hex()
ref=lambda n:0 if n<0 else objects[n]
index=lambda p:-1 if p==0 else objects.index(p)
spellref=lambda n:0 if n<0 else record if n==0 else other
spellindex=lambda p:-1 if p==0 else 0 if p==record else 1 if p==other else -99
cstr=lambda p:bytes(u.mem_read(p,128)).split(b'\0',1)[0].decode('ascii')
createcache,destroycache,angles,multiplier=0x833c08,0x833c0c,0x5b6658,0x583a1c
assert get(multiplier)==f32(4)
regions={'create':(0x531310,0x531404),'destroy':(0x531420,0x531481),'converter':(0x419a70,0x419a81)}
traces={name:{} for name in regions};entries,services,mutations=collections.Counter(),collections.Counter(),collections.Counter();table_reads=collections.Counter()
active,current,events,calls,mutated,allowed,faults=False,{},[],collections.Counter(),False,set(),[]
links,owners,head=[],[],0
def corestate():
    return dict(level=get(record+8),caster=index(get(record+16)),anchor=index(get(record+24)),flag=get(record+20),budget=get(record+72),caches=[get(createcache),get(destroycache)],source_positions=[words(o+56) for o in objects[:3]],lookups=calls['lookup'],attempts=calls['new'])
def state():
    return dict(**corestate(),head=head,objects=[[get16(o+4),spellindex(get(collisions[i])),get16(o+124),get16(o+126),get(o+80),get(o+84),get(o+112),*words(o+56),links[i],owners[i]] for i,o in enumerate(objects)])
def allow(address,size):allowed.update(range(address-heap,address-heap+size))
def write(address,value):allow(address,4);put(address,value)
def writepoint(address,value):
    for i,w in enumerate(value):write(address+4*i,w)
def mutate(stage):
    global mutated,head
    at=calls[stage];calls[stage]+=1;m=current['mutation']
    if mutated or m['stage']!=stage or m['at']!=at:return
    mutated=True;action=m['action'];mutations[stage+'/'+str(at)+'/'+action]+=1
    if action=='caster-b':write(record+16,objects[1])
    elif action=='anchor-b':write(record+24,objects[2])
    elif action=='flag':write(record+20,get(record+20)^1)
    elif action=='source-pos':writepoint(objects[0]+56,point(777.25,-888.5))
    elif action=='budget':write(record+72,0x13579bdf)
    elif action=='level':write(record+8,0xfffffffe)
    elif action=='create-cache':put(createcache,99)
    elif action=='destroy-cache':put(destroycache,20)
    elif action=='head-b':head=6
    elif action=='next-nil':links[current_object]=-1
    elif action=='next-skip':links[current_object]=8
    elif action=='next-type':allow(objects[6]+4,2);put16(objects[6]+4,20)
    elif action=='next-spell':write(collisions[6],other)
    elif action=='projectile-fields':
        o=objects[current_object];allow(o+124,4);put16(o+124,0x1234);put16(o+126,0x5678);writepoint(o+80,point(1.25,2.5));write(o+112,f32(9))
    elif action=='projectile-spell':write(collisions[current_object],other)
    else:raise AssertionError(action)
def event(kind,args):events.append(dict(kind=kind,args=args,state=corestate()));mutate(kind)
def ret(value=0):
    esp=u.reg_read(UC_X86_REG_ESP);u.reg_write(UC_X86_REG_EAX,value&0xffffffff);u.reg_write(UC_X86_REG_EIP,get(esp));u.reg_write(UC_X86_REG_ESP,esp+4)
def hook(_u,pc,length,_):
    global current_object
    if not active:return
    if pc==stop:u.emu_stop();return
    for name,(lo,hi) in regions.items():
        if lo<=pc<hi:
            traces[name][pc]=bytes(u.mem_read(pc,length))
            if pc==lo:entries[name]+=1
            return
    if thunk<=pc<thunk+7:return
    esp=u.reg_read(UC_X86_REG_ESP);arg=lambda n:get(esp+4*n);services[hex(pc)]+=1
    if pc==0x419d70:
        assert cstr(arg(1))=='TurnUndeadKillPoints';event('balance',[arg(2)]);u.mem_write(scratch,bytes.fromhex(current['balance_bits64']));u.reg_write(UC_X86_REG_EIP,thunk)
    elif pc==0x4e3aa0:
        assert cstr(arg(1))=='UndeadKiller';mutate('lookup');ret(current['lookup'])
    elif pc==0x4e3450:
        attempt=calls['new'];current_object=3+attempt
        success=attempt not in current['failures'] and arg(1)<=0xffff and arg(1)!=0
        value=objects[current_object] if success else 0
        event('new',[arg(1),index(value)])
        if value:allow(value+4,2);put16(value+4,arg(1))
        ret(value)
    elif pc==0x4daa50:
        assert arg(5)==0;obj,owner,pos=arg(1),arg(2),[arg(3),arg(4)];current_object=index(obj)
        event('create',[current_object,index(owner),*pos,spellindex(get(collisions[current_object]))]);writepoint(obj+56,pos);owners[current_object]=index(owner);ret(0x24681357)
    elif pc==0x522ff0:
        assert arg(1)==160;event('fx',[arg(1),*words(arg(2))]);ret(0x13572468)
    elif pc==0x4da790:ret(ref(head))
    elif pc==0x4da7a0:ret(ref(links[index(arg(1))]))
    elif pc==0x4e5cc0:current_object=index(arg(1));event('delete',[current_object]);ret(0x24681357)
    else:raise AssertionError((current['name'],'undeclared service/instruction',hex(pc)))
def invalid(_u,access,address,size,value,_):faults.append(dict(name=current['name'],pc=hex(u.reg_read(UC_X86_REG_EIP)),address=hex(address)));return False
def read(_u,access,address,size,value,_):
    if active and (angles<=address<angles+2048 or address==multiplier):table_reads[(address,size)]+=1
u.hook_add(UC_HOOK_CODE,hook);u.hook_add(UC_HOOK_MEM_INVALID,invalid);u.hook_add(UC_HOOK_MEM_READ,read)
seed=bytes((i*31+19)&255 for i in range(0x20000));registers={UC_X86_REG_EBX:0x13572468,UC_X86_REG_ESI:0x24681357,UC_X86_REG_EDI:0x98765432,UC_X86_REG_EBP:0xfedcba98}
def run(case):
    global active,current,events,calls,mutated,allowed,links,owners,head,current_object
    active,current,events,calls,mutated,allowed=False,case,[],collections.Counter(),False,set();links=list(case['links']);owners=[-1]*46;head=case['head'];current_object=3;u.mem_write(heap,seed)
    for off,val in [(8,case['level']),(16,ref(case['caster'])),(20,case['flag']),(24,ref(case['anchor'])),(72,case['budget'])]:put(record+off,val)
    for i,o in enumerate(objects):
        put16(o+4,case['types'][i]);put(o+700,collisions[i]);put(collisions[i],spellref(case['spells'][i]))
        for j,v in enumerate(case['positions'][i]):put(o+56+4*j,v)
        put16(o+124,0xa100+i);put16(o+126,0xb200+i);put(o+80,f32(1.25));put(o+84,f32(-2.5));put(o+112,f32(7))
    u.mem_write(createcache-16,bytes(range(40)));put(createcache,case['caches'][0]);put(destroycache,case['caches'][1]);global_before=bytes(u.mem_read(createcache-16,40))
    u.mem_write(cw,struct.pack('<H',a.control));u.emu_start(init,init+8,count=2);before=bytes(u.mem_read(heap,len(seed)))
    u.mem_write(sp-512,bytes(1024));put(sp,stop);put(sp+4,record)
    for reg,val in registers.items():u.reg_write(reg,val)
    u.reg_write(UC_X86_REG_EFLAGS,0x202);u.reg_write(UC_X86_REG_ESP,sp);active=True;u.emu_start(regions[case['kind']][0],stop+1,count=100000);active=False
    assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4 and get(sp)==stop and all(u.reg_read(reg)==val for reg,val in registers.items())
    assert bytes(u.mem_read(sp-512,256))==bytes(256) and bytes(u.mem_read(sp+8,256))==bytes(256)
    assert u.reg_read(UC_X86_REG_FPCW)==a.control and u.reg_read(UC_X86_REG_FPSW)&0x3800==0
    if case['kind']=='create':
        allow(record+72,4)
        for i,o in enumerate(objects[3:],3):allow(collisions[i],4);allow(o+124,4);allow(o+80,8);allow(o+112,4)
    after=bytes(u.mem_read(heap,len(seed)));assert all(x==y or i in allowed for i,(x,y) in enumerate(zip(before,after))),(case['name'],'nonoutput heap canary')
    after=bytes(u.mem_read(createcache-16,40));assert all(x==y or 16<=i<24 for i,(x,y) in enumerate(zip(global_before,after)))
    if case['mutation']['stage']:assert mutated,(case['name'],'mutation not reached')
    return dict(input=copy.deepcopy(case),result=dict(ret=None if case['kind']=='destroy' else u.reg_read(UC_X86_REG_EAX),state=state(),events=copy.deepcopy(events)))
template=dict(name='',kind='create',level=3,caster=0,anchor=1,flag=0,budget=0x87654321,caches=[0,0],lookup=19,failures=[],head=3,
    positions=[point(100.25,-100.5),point(300.75,400.5),point(-500.25,600.5)]+[point(i+.25,-i-.5) for i in range(43)],
    types=[11,12,13]+[19,20,19,19,20,19]+[21]*37,spells=[1]*3+[0,0,1,0,-1,0]+[1]*37,links=[-1]*3+[4,5,6,7,8,-1]+[-1]*37,
    balance_bits64=f64(2.5),mutation=dict(stage='',at=0,action=''))
cases=[]
def add(kind,name,**changes):
    case=copy.deepcopy(template);case.update(kind=kind);case.update(changes);case['name']=kind+'/'+name;cases.append(case)
for flag in [0,1,0x80000000]:add('create','source-flag-'+str(flag),flag=flag)
add('create','glyph-without-caster',flag=1,caster=-1)
for caches in [[19,19],[77,88],[0xffffffff,0x10013]]:add('create','warm-'+str(caches[0]),caches=caches)
for failures in [list(range(43)),list(range(0,43,2)),[0,21,42]]:add('create','failures-'+str(len(failures)),failures=failures)
for value in [.49,.51,1.5,2.5,3.5,-1.5,-2.5,-3.5,1.999999999999,-1.999999999999,2147483647.,-2147483647.,2147483648.,math.inf,math.nan]:add('create','budget-'+f64(value),balance_bits64=f64(value),caches=[19,19])
for level in [0,1,19,0xffffffff]:add('create','level-'+str(level),level=level)
for stage,ats,actions in [
    ('balance',[0],['caster-b','anchor-b','flag','source-pos','budget','level','create-cache']),
    ('lookup',[0],['caster-b','anchor-b','flag','source-pos','budget','level','create-cache']),
    ('new',[0,21,42],['caster-b','anchor-b','flag','source-pos','budget','level','create-cache','projectile-fields']),
    ('create',[0,21,42],['caster-b','anchor-b','flag','source-pos','budget','level','create-cache','projectile-fields','projectile-spell']),
    ('fx',[0],['caster-b','anchor-b','flag','budget','create-cache'])]:
    for at in ats:
        for action in actions:add('create',stage+'-'+str(at)+'-'+action,mutation=dict(stage=stage,at=at,action=action))
for caches in [[0,0],[19,19],[0,20],[19,0x10013],[19,0xffffffff]]:add('destroy','cache-'+str(caches[1]),caches=caches)
add('destroy','empty',head=-1)
add('destroy','high-lookup',lookup=0x10013)
add('destroy','zero-lookup',lookup=0)
for action in ['destroy-cache','head-b','budget','caster-b']:add('destroy','lookup-'+action,mutation=dict(stage='lookup',at=0,action=action))
for at in [0,1]:
    for action in ['next-nil','next-skip','next-type','next-spell','destroy-cache','budget','caster-b']:
        add('destroy','delete-'+str(at)+'-'+action,caches=[19,19],mutation=dict(stage='delete',at=at,action=action))
assert len({c['name'] for c in cases})==len(cases)
rows=[run(c) for c in cases];sealed=[]
for name,(lo,hi) in regions.items():
    ins=traces[name];covered={pc+i for pc,body in ins.items() for i in range(len(body))};assert covered==set(range(lo,hi)),(name,sorted(set(range(lo,hi))-covered))
    sealed.append(dict(name=name,address=hex(lo),end_exclusive=hex(hi),bytes=hi-lo,sha256=sha(bytes(u.mem_read(lo,hi-lo))),instructions=[dict(pc=hex(pc),bytes=body.hex()) for pc,body in sorted(ins.items())]))
expected_reads={(angles+8*d+i,4) for d in range(0,256,6) for i in [0,4]}|{(multiplier,4)}
assert set(table_reads)==expected_reads and not faults
encoded=(json.dumps(dict(original_sha256=sha(blob),control_word=a.control,rows=rows),indent=2)+'\n').encode()
with a.fixture.open('xb') as f:f.write(encoded)
proof=dict(original_sha256=sha(blob),control_word=a.control,contexts=len(rows),entries=dict(entries),regions=sealed,all_original_instruction_bytes_executed=True,zero_caller_packing_arithmetic_or_conversion_replacements=True,callee_saved_registers_stack_control_word_TOP_and_nonoutput_heap_global_canaries_preserved=True,services=dict(services),mutations=dict(mutations),faults=faults,fixture_sha256=sha(encoded),table_sha256=sha(bytes(u.mem_read(angles,2048))),multiplier_sha256=sha(bytes(u.mem_read(multiplier,4))),actual_table_read_ranges=[dict(address=hex(k[0]),bytes=k[1],reads=v) for k,v in sorted(table_reads.items())],scope=__doc__)
with a.proof.open('x') as f:json.dump(proof,f,indent=2);f.write('\n')
print(json.dumps({k:v for k,v in proof.items() if k not in ['regions','mutations','scope','actual_table_read_ranges']}))
