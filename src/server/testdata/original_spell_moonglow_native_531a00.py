"""Execute unchanged Moonglow create/destroy roots and their FISTP converter.

Balance return, object allocation, create-at, buff application/removal and
delayed deletion are declared services. Create-at writes its declared object's
position. Buff ABI observations use signed low16 duration and low8 power.
Destroy's unused EAX is recorded separately, not compared by the void consumer.
No root caller, packing, arithmetic, conversion or memory access is replaced.
"""
import argparse, collections, copy, hashlib, json, math, struct
from pathlib import Path
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import *

p = argparse.ArgumentParser(description=__doc__)
for key in ['game', 'fixture', 'proof']: p.add_argument('--'+key, type=Path, required=True)
p.add_argument('--control', type=lambda v:int(v,0), choices=[0xe3f,0x23f], default=0xe3f)
a = p.parse_args()
blob = a.game.read_bytes(); sha = lambda b:hashlib.sha256(b).hexdigest()
assert sha(blob) == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe = struct.unpack_from('<I',blob,60)[0]; opt = pe+24
base = struct.unpack_from('<I',blob,opt+28)[0]; table = opt+struct.unpack_from('<H',blob,pe+20)[0]
u = Uc(UC_ARCH_X86,UC_MODE_32)
u.mem_map(base,(struct.unpack_from('<I',blob,opt+56)[0]+4095)&~4095)
for n in range(struct.unpack_from('<H',blob,pe+6)[0]):
    _,va,size,off = struct.unpack_from('<IIII',blob,table+40*n+8)
    if size:u.mem_write(base+va,blob[off:off+size])
heap,stack,code = 0x21000000,0x22000000,0x23000000
for address in [heap,stack,code]:u.mem_map(address,0x20000)
record = heap+0x1000
objects = [heap+0x2000+0x400*i for i in range(6)]
updates = [heap+0x6000+0x600*i for i in range(3)]
players = [heap+0x8000+0x1000*i for i in range(3)]
sp,stop,init,thunk,scratch,cw = stack+0x1e000,code+0x100,code+0x200,code+0x300,code+0x400,code+0x500
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw))
u.mem_write(thunk,b'\xdd\x05'+struct.pack('<I',scratch)+b'\xc3')
get = lambda address:struct.unpack('<I',u.mem_read(address,4))[0]
put = lambda address,value:u.mem_write(address,struct.pack('<I',value&0xffffffff))
words = lambda address:[get(address),get(address+4)]
signed = lambda value,bits:((value&((1<<bits)-1))^(1<<(bits-1)))-(1<<(bits-1))
f32 = lambda value:struct.unpack('<I',struct.pack('<f',value))[0]
point = lambda x,y:[f32(x),f32(y)]
f64 = lambda value:struct.pack('<d',value).hex()
ref = lambda index:0 if index<0 else objects[index]
index = lambda value:-1 if value==0 else objects.index(value)
ud_index = lambda value:-1 if value==0 else updates.index(value)
pl_index = lambda value:-1 if value==0 else players.index(value)
cstr = lambda address:bytes(u.mem_read(address,128)).split(b'\0',1)[0].decode('ascii')
regions = {'create':(0x531a00,0x531ae2),'destroy':(0x531af0,0x531b3b),'converter':(0x419a70,0x419a81)}
traces = {name:{} for name in regions}
entries,services,mutations = collections.Counter(),collections.Counter(),collections.Counter()
active,current,events,calls,mutated,allowed,faults = False,{},[],collections.Counter(),False,set(),[]
def state():
    return dict(target=index(get(record+48)),level=get(record+8),visual=index(get(record+72)),
        classes=[get(o+8) for o in objects],flags=[get(o+16) for o in objects],positions=[words(o+56) for o in objects],
        object_updates=[ud_index(get(o+748)) for o in objects],players=[pl_index(get(d+276)) for d in updates],
        player_flags=[get(o+3680) for o in players],cursors=[[signed(get(o+2284+4*i),32) for i in range(2)] for o in players])
def allow(address,size):allowed.update(range(address-heap,address-heap+size))
def write(address,value):allow(address,4);put(address,value)
def writepoint(address,value):
    for i,word in enumerate(value):write(address+4*i,word)
def mutate(kind):
    global mutated
    at = calls[kind];calls[kind]+=1;m=current['mutation']
    if mutated or m['stage']!=kind or m['at']!=at:return
    mutated=True;action=m['action'];mutations[kind+'/'+str(at)+'/'+action]+=1
    if action=='target-b':write(record+48,objects[1])
    elif action=='target-monster':write(record+48,objects[2])
    elif action=='target-nil':write(record+48,0)
    elif action=='level':write(record+8,0xfffffffe)
    elif action=='visual-b':write(record+72,objects[5])
    elif action=='visual-nil':write(record+72,0)
    elif action=='target-destroyed':write(objects[0]+16,get(objects[0]+16)|0x20)
    elif action=='target-dead':write(objects[0]+16,get(objects[0]+16)|0x8000)
    elif action=='update-b':write(objects[0]+748,updates[1])
    elif action=='player-b':write(updates[0]+276,players[1])
    elif action=='cursor-mode':write(players[0]+3680,get(players[0]+3680)^0x10)
    elif action=='cursor':writepoint(players[0]+2284,[16777219,(-16777219)&0xffffffff])
    else:raise AssertionError(action)
def event(kind,args):events.append(dict(kind=kind,args=args,state=state()));mutate(kind)
def ret(value=0):
    esp=u.reg_read(UC_X86_REG_ESP);u.reg_write(UC_X86_REG_EAX,value&0xffffffff)
    u.reg_write(UC_X86_REG_EIP,get(esp));u.reg_write(UC_X86_REG_ESP,esp+4)
def hook(_u,pc,length,_):
    if not active:return
    if pc==stop:u.emu_stop();return
    for name,(lo,hi) in regions.items():
        if lo<=pc<hi:
            traces[name][pc]=bytes(u.mem_read(pc,length))
            if pc==lo:entries[name]+=1
            return
    if thunk<=pc<thunk+7:return
    esp=u.reg_read(UC_X86_REG_ESP);arg=lambda n:get(esp+4*n);services[hex(pc)]+=1
    if pc==0x419d40:
        assert cstr(arg(1))=='MoonglowEnchantmentDuration';event('balance',[])
        u.mem_write(scratch,bytes.fromhex(current['balance_bits64']));u.reg_write(UC_X86_REG_EIP,thunk)
    elif pc==0x4e3810:
        assert cstr(arg(1))=='Moonglow';event('new',['Moonglow']);ret(ref(current['allocated']))
    elif pc==0x4daa50:
        assert arg(5)==0
        visual,owner,position=arg(1),arg(2),[arg(3),arg(4)]
        event('create',[index(visual),index(owner),*position]);writepoint(visual+56,position);ret(0x13572468)
    elif pc==0x4ff380:
        assert arg(2) in [1,15]
        event('apply',[index(arg(1)),arg(2),signed(arg(3),16),signed(arg(4),8)]);ret(0x24681357)
    elif pc==0x4e5cc0:event('delete',[index(arg(1))]);ret(0x13572468)
    elif pc==0x4ff5b0:
        assert arg(2) in [1,15]
        event('off',[index(arg(1)),arg(2)]);ret(0x24681357)
    else:raise AssertionError((current['name'],'undeclared instruction/service',hex(pc)))
def invalid(_u,access,address,size,value,_):
    faults.append(dict(name=current['name'],pc=hex(u.reg_read(UC_X86_REG_EIP)),address=hex(address)));return False
u.hook_add(UC_HOOK_CODE,hook);u.hook_add(UC_HOOK_MEM_INVALID,invalid)
seed=bytes((i*31+19)&255 for i in range(0x20000))
registers={UC_X86_REG_EBX:0x13572468,UC_X86_REG_ESI:0x24681357,UC_X86_REG_EDI:0x98765432,UC_X86_REG_EBP:0xfedcba98}
def run(case):
    global active,current,events,calls,mutated,allowed
    active,current,events,calls,mutated,allowed=False,case,[],collections.Counter(),False,set();u.mem_write(heap,seed)
    for offset,value in [(8,case['level']),(48,ref(case['target'])),(72,ref(case['visual']))]:put(record+offset,value)
    for i,obj in enumerate(objects):
        for offset,value in [(8,case['classes'][i]),(16,case['flags'][i]),(748,0 if case['object_updates'][i]<0 else updates[case['object_updates'][i]])]:put(obj+offset,value)
        for j,value in enumerate(case['positions'][i]):put(obj+56+4*j,value)
    for i,data in enumerate(updates):put(data+276,players[case['players'][i]])
    for i,player in enumerate(players):
        put(player+3680,case['player_flags'][i])
        for j,value in enumerate(case['cursors'][i]):put(player+2284+4*j,value)
    u.mem_write(cw,struct.pack('<H',a.control));u.emu_start(init,init+8,count=2)
    before=bytes(u.mem_read(heap,len(seed)))
    u.mem_write(sp-512,bytes(1024));put(sp,stop);put(sp+4,record)
    for reg,value in registers.items():u.reg_write(reg,value)
    u.reg_write(UC_X86_REG_EFLAGS,0x202);u.reg_write(UC_X86_REG_ESP,sp)
    active=True;u.emu_start(regions[case['kind']][0],stop+1,count=30000);active=False
    assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4 and get(sp)==stop
    assert all(u.reg_read(reg)==value for reg,value in registers.items())
    assert bytes(u.mem_read(sp-512,256))==bytes(256) and bytes(u.mem_read(sp+8,256))==bytes(256)
    assert u.reg_read(UC_X86_REG_FPCW)==a.control and u.reg_read(UC_X86_REG_FPSW)&0x3800==0
    allow(record+72,4)
    after=bytes(u.mem_read(heap,len(seed)))
    assert all(x==y or i in allowed for i,(x,y) in enumerate(zip(before,after))),(case['name'],'nonoutput heap canary')
    if case['mutation']['stage']:assert mutated,(case['name'],'mutation not reached')
    return dict(input=copy.deepcopy(case),result=dict(ret=None if case['kind']=='destroy' else u.reg_read(UC_X86_REG_EAX),state=state(),events=copy.deepcopy(events)))
template=dict(name='',kind='create',target=0,level=3,visual=3,allocated=4,
    classes=[4,4,2,2,2,2],flags=[0]*6,positions=[point(100.25,-100.5),point(-11.5,22.25),point(7,11),point(30,40),point(50,60),point(70,80)],
    object_updates=[0,1,2,-1,-1,-1],players=[0,1,2],player_flags=[0,0x10,0x10],cursors=[[-123,456],[16777219,-16777219],[2147483647,-2147483648]],
    balance_bits64=f64(70.9),mutation=dict(stage='',at=0,action=''))
cases=[]
def add(kind,name,**changes):
    case=copy.deepcopy(template);case.update(kind=kind);case.update(changes);case['name']=kind+'/'+name;cases.append(case)
for target in [-1,0,1,2]:
    add('create','target-'+str(target),target=target)
    add('destroy','target-'+str(target),target=target)
for flag in [0x20,0x8000,0x8020,0x2,0x1000000]:
    flags=[0]*6;flags[0]=flag;add('create','target-flags-'+hex(flag),flags=flags)
    add('destroy','target-flags-'+hex(flag),flags=flags)
for cls in [0,1,2,4,5,6,0x80000004]:
    classes=list(template['classes']);classes[0]=cls
    add('create','class-'+str(cls),classes=classes)
    add('destroy','class-'+str(cls),classes=classes)
add('create','allocation-fail',allocated=-1)
add('destroy','missing-visual',visual=-1)
add('destroy','nil-target-and-visual',target=-1,visual=-1)
for value in [.49,.51,1.5,2.5,-.49,-.51,-1.5,-2.5,1.999999999999,-1.999999999999,32767.9,32768.9,-32768.9,65535.9,65536.9,2147483647.,-2147483647.,2147483648.,-2147483649.,math.inf,-math.inf,math.nan]:
    add('create','duration-'+f64(value),target=2,balance_bits64=f64(value))
for level in [0,127,128,255,256,0xffffffff]:add('create','power-'+str(level),target=2,level=level)
for x,y in [(16777217,-16777217),(16777219,-16777219),(2147483647,-2147483647),(-2147483648,2147483646)]:
    cursors=copy.deepcopy(template['cursors']);cursors[0]=[x,y]
    add('create','cursor-'+str(x)+'-'+str(y),player_flags=[0x10,0x10,0x10],cursors=cursors)
for stage,actions in [
    ('balance',['target-b','target-monster','target-nil','level','visual-b','visual-nil','target-destroyed','target-dead','update-b','player-b','cursor-mode','cursor']),
    ('new',['target-b','level','visual-b','visual-nil','target-destroyed','target-dead','update-b','player-b','cursor-mode','cursor']),
    ('create',['target-b','target-monster','level','visual-b','visual-nil','target-destroyed','target-dead','update-b','player-b','cursor-mode','cursor']),
    ('apply',['target-b','target-monster','level','visual-b','visual-nil','target-destroyed','update-b','player-b','cursor-mode','cursor'])]:
    for action in actions:add('create',stage+'-'+action,mutation=dict(stage=stage,at=0,action=action))
for action in ['target-b','target-nil','level','visual-b','visual-nil','cursor']:
    add('create','light-apply-'+action,target=2,mutation=dict(stage='apply',at=0,action=action))
for stage,actions in [
    ('delete',['target-b','target-monster','level','visual-b','visual-nil','target-destroyed','update-b','player-b','cursor-mode','cursor']),
    ('off',['target-b','target-monster','target-nil','level','visual-b','visual-nil','target-destroyed','update-b','player-b','cursor-mode','cursor'])]:
    for action in actions:add('destroy',stage+'-'+action,mutation=dict(stage=stage,at=0,action=action))
for action in ['target-b','target-nil','level','visual-b','visual-nil']:add('destroy','light-off-'+action,target=2,mutation=dict(stage='off',at=0,action=action))
assert len({c['name'] for c in cases})==len(cases)
rows=[run(case) for case in cases]
sealed=[]
for name,(lo,hi) in regions.items():
    ins=traces[name];covered={pc+i for pc,body in ins.items() for i in range(len(body))}
    assert covered==set(range(lo,hi)),(name,sorted(set(range(lo,hi))-covered))
    sealed.append(dict(name=name,address=hex(lo),end_exclusive=hex(hi),bytes=hi-lo,sha256=sha(bytes(u.mem_read(lo,hi-lo))),instructions=[dict(pc=hex(pc),bytes=body.hex()) for pc,body in sorted(ins.items())]))
encoded=(json.dumps(dict(original_sha256=sha(blob),control_word=a.control,rows=rows),indent=2)+'\n').encode()
with a.fixture.open('xb') as f:f.write(encoded)
proof=dict(original_sha256=sha(blob),control_word=a.control,contexts=len(rows),entries=dict(entries),regions=sealed,all_original_instruction_bytes_executed=True,zero_caller_packing_arithmetic_or_conversion_replacements=True,callee_saved_registers_stack_control_word_TOP_and_nonoutput_heap_canaries_preserved=True,services=dict(services),mutations=dict(mutations),faults=faults,fixture_sha256=sha(encoded),destroy_EAX_ignored_by_void_consumer=True,scope=__doc__)
with a.proof.open('x') as f:json.dump(proof,f,indent=2);f.write('\n')
print(json.dumps({k:v for k,v in proof.items() if k not in ['regions','mutations','scope']}))
