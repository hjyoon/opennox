"""Execute unchanged original Oval Shield create/update/destroy roots.

Offensive cancellation, buff apply/test/off and position-delta are declared
services. FPS/frame reads, dword arithmetic, reloads, argument packing and all
root instructions execute unchanged. Destroy's unused EAX is not compared.
"""
import argparse, collections, copy, hashlib, json, struct
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
record = heap+0x1000; objects = [heap+0x2000+0x400*i for i in range(5)]
sp,stop,init,cw = stack+0x1e000,code+0x100,code+0x200,code+0x500
fps,frame = 0x85b3fc,0x84ea04
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw))
get = lambda address:struct.unpack('<I',u.mem_read(address,4))[0]
put = lambda address,value:u.mem_write(address,struct.pack('<I',value&0xffffffff))
words = lambda address:[get(address),get(address+4)]
signed = lambda value,bits:((value&((1<<bits)-1))^(1<<(bits-1)))-(1<<(bits-1))
ref = lambda index:0 if index<0 else objects[index]
index = lambda value:-1 if value==0 else objects.index(value)
regions = {'create':(0x531490,0x5314eb),'update':(0x5314f0,0x531551),'destroy':(0x531560,0x531577)}
traces = {name:{} for name in regions}
entries,services,mutations = collections.Counter(),collections.Counter(),collections.Counter()
active,current,events,calls,mutated,allowed,faults = False,{},[],collections.Counter(),False,set(),[]

def state():
    return dict(target=index(get(record+48)),level=get(record+8),expiry=get(record+68),
        position=words(record+28),fps=get(fps),frame=get(frame),
        classes=[get(o+8) for o in objects],flags=[get(o+16) for o in objects],positions=[words(o+56) for o in objects])

def allow(address,size):allowed.update(range(address,address+size))
def write(address,value):allow(address,4);put(address,value)
def mutate(kind):
    global mutated
    at = calls[kind];calls[kind]+=1;m=current['mutation']
    if mutated or m['stage']!=kind or m['at']!=at:return
    mutated=True;action=m['action'];mutations[kind+'/'+str(at)+'/'+action]+=1
    if action=='target-b':write(record+48,objects[1])
    elif action=='target-monster':write(record+48,objects[2])
    elif action=='target-nil':write(record+48,0)
    elif action=='level':write(record+8,0xfffffffe)
    elif action=='frame':write(frame,0xfffffff9)
    elif action=='fps':write(fps,0xffffffff)
    elif action=='target-flags':write(objects[0]+16,0x8020)
    elif action=='target-class':write(objects[0]+8,2)
    elif action=='record-position':write(record+28,0x4b800001);write(record+32,0xcb800001)
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
    esp=u.reg_read(UC_X86_REG_ESP);arg=lambda n:get(esp+4*n);services[hex(pc)]+=1
    if pc==0x4ff310:event('cancel',[index(arg(1))]);ret(0x13572468)
    elif pc==0x4ff380:
        assert arg(2)==27
        event('apply',[index(arg(1)),arg(2),signed(arg(3),16),signed(arg(4),8)]);ret(0x24681357)
    elif pc==0x4ff350:
        assert arg(2)==8
        event('test',[index(arg(1)),arg(2)]);ret(current['test_result'])
    elif pc==0x4fea70:
        assert arg(2)==record+28
        event('position',[index(arg(1)),*words(arg(2))]);ret(current['position_result'])
    elif pc==0x4ff5b0:
        assert arg(2)==27
        event('off',[index(arg(1)),arg(2)]);ret(0x87654321)
    else:raise AssertionError((current['name'],'undeclared instruction/service',hex(pc)))
def invalid(_u,access,address,size,value,_):
    faults.append(dict(name=current['name'],pc=hex(u.reg_read(UC_X86_REG_EIP)),address=hex(address)));return False
u.hook_add(UC_HOOK_CODE,hook);u.hook_add(UC_HOOK_MEM_INVALID,invalid)
seed=bytes((i*31+19)&255 for i in range(0x20000))
registers={UC_X86_REG_EBX:0x13572468,UC_X86_REG_ESI:0x24681357,UC_X86_REG_EDI:0x98765432,UC_X86_REG_EBP:0xfedcba98}

def run(case):
    global active,current,events,calls,mutated,allowed
    active,current,events,calls,mutated,allowed=False,case,[],collections.Counter(),False,set();u.mem_write(heap,seed)
    for address,value in [(fps,case['fps']),(frame,case['frame']),(record+8,case['level']),(record+48,ref(case['target'])),(record+68,case['expiry'])]:put(address,value)
    for j,value in enumerate(case['position']):put(record+28+4*j,value)
    for i,obj in enumerate(objects):
        for offset,value in [(8,case['classes'][i]),(16,case['flags'][i])]:put(obj+offset,value)
        for j,value in enumerate(case['positions'][i]):put(obj+56+4*j,value)
    u.mem_write(cw,struct.pack('<H',a.control));u.emu_start(init,init+8,count=2)
    before=bytes(u.mem_read(heap,len(seed)));global_before={v:get(v) for v in [fps,frame]}
    u.mem_write(sp-512,bytes(1024));put(sp,stop);put(sp+4,record)
    for reg,value in registers.items():u.reg_write(reg,value)
    u.reg_write(UC_X86_REG_EFLAGS,0x202);u.reg_write(UC_X86_REG_ESP,sp)
    active=True;u.emu_start(regions[case['kind']][0],stop+1,count=30000);active=False
    assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4 and get(sp)==stop
    assert all(u.reg_read(reg)==value for reg,value in registers.items())
    assert bytes(u.mem_read(sp-512,256))==bytes(256) and bytes(u.mem_read(sp+8,256))==bytes(256)
    assert u.reg_read(UC_X86_REG_FPCW)==a.control and u.reg_read(UC_X86_REG_FPSW)&0x3800==0
    allow(record+68,4)
    after=bytes(u.mem_read(heap,len(seed)))
    assert all(x==y or heap+i in allowed for i,(x,y) in enumerate(zip(before,after))),(case['name'],'nonoutput heap canary')
    assert all(get(v)==old or v in allowed for v,old in global_before.items())
    if case['mutation']['stage']:assert mutated,(case['name'],'mutation not reached')
    return dict(input=copy.deepcopy(case),result=dict(ret=None if case['kind']=='destroy' else u.reg_read(UC_X86_REG_EAX),state=state(),events=copy.deepcopy(events)))

template=dict(name='',kind='create',target=0,level=3,fps=30,frame=0xfffffff0,expiry=0x87654321,
    position=[0x3fdccccc,0xc1b20000],classes=[4,4,2,4,2],flags=[0]*5,
    positions=[[0x42c88000,0xc2c90000],[0xc1380000,0x41b20000],[0x40e00000,0x41300000],[0x41f00000,0x42200000],[0x42480000,0x42700000]],
    test_result=0,position_result=0,mutation=dict(stage='',at=0,action=''))
cases=[]
def add(kind,name,**changes):
    case=copy.deepcopy(template);case.update(kind=kind);case.update(changes);case['name']=kind+'/'+name;cases.append(case)
for kind in regions:
    for target in [-1,0,1,2,3,4]:add(kind,'target-'+str(target),target=target)
for cls in [0,1,2,3,4,5,6,7,0x100,0x80000004]:
    classes=list(template['classes']);classes[0]=cls
    for kind in ['create','update']:add(kind,'class-'+str(cls),classes=classes)
for flag in [0,0x20,0x8000,0x8020,2,0x1000000,0xffffffff]:
    flags=list(template['flags']);flags[0]=flag;add('update','flags-'+hex(flag),flags=flags)
for level in [0,1,2,127,128,255,256,32767,65535,65536,0xfffffffe,0xffffffff]:
    for rate in [0,1,30,60,65535,65536,0x80000000,0xffffffff]:add('create','level-'+str(level)+'-fps-'+str(rate),level=level,fps=rate)
for tick in [0,1,0x7fffffff,0x80000000,0xffffffff]:add('create','frame-'+str(tick),frame=tick)
for value in [0,1,0xffffffff,0x80000000]:
    add('update','test-'+str(value),test_result=value)
    add('update','position-'+str(value),target=2,position_result=value)
actions=['target-b','target-monster','level','frame','fps','target-flags','target-class','record-position']
for stage in ['cancel','apply','test','position','off']:
    kind={'cancel':'create','apply':'create','test':'update','position':'update','off':'destroy'}[stage]
    for action in actions:
        add(kind,stage+'-'+action,target=2 if stage=='position' else 0,mutation=dict(stage=stage,at=0,action=action))
for stage in ['apply','off']:add('create' if stage=='apply' else 'destroy',stage+'-target-nil',mutation=dict(stage=stage,at=0,action='target-nil'))
add('update','test-nil-early-return',test_result=1,mutation=dict(stage='test',at=0,action='target-nil'))
assert len({c['name'] for c in cases})==len(cases)
rows=[run(case) for case in cases]
sealed=[]
for name,(lo,hi) in regions.items():
    ins=traces[name];covered={pc+i for pc,body in ins.items() for i in range(len(body))}
    assert covered==set(range(lo,hi)),(name,sorted(set(range(lo,hi))-covered))
    sealed.append(dict(name=name,address=hex(lo),end_exclusive=hex(hi),bytes=hi-lo,sha256=sha(bytes(u.mem_read(lo,hi-lo))),instructions=[dict(pc=hex(pc),bytes=body.hex()) for pc,body in sorted(ins.items())]))
encoded=(json.dumps(dict(original_sha256=sha(blob),control_word=a.control,rows=rows),indent=2)+'\n').encode()
with a.fixture.open('xb') as f:f.write(encoded)
proof=dict(original_sha256=sha(blob),control_word=a.control,contexts=len(rows),entries=dict(entries),regions=sealed,all_original_instruction_bytes_executed=True,zero_caller_packing_arithmetic_or_conversion_replacements=True,callee_saved_registers_stack_control_word_TOP_and_nonoutput_heap_and_global_canaries_preserved=True,services=dict(services),mutations=dict(mutations),faults=faults,fixture_sha256=sha(encoded),destroy_EAX_ignored_by_void_consumer=True,scope=__doc__)
with a.proof.open('x') as f:json.dump(proof,f,indent=2);f.write('\n')
print(json.dumps({k:v for k,v in proof.items() if k not in ['regions','mutations','scope']}))
