"""Execute complete original Mana Bomb roots, converter and local getters.

Lookup, type resolution/allocation, creation, buff application/removal, deletion,
radial damage, quake, FX, audio and mana subtraction are declared world-service
bodies. Their ordered arguments and entry states are observed. All root reads,
stores, branches and x87 instructions execute unchanged, including the original
dictionary indirection, player/mana reads and walls flag setter. Input levels
are 1..6 (six is outside the five-value table); FPS and buff durations are within
the stock signed-word range. Nil-caster non-glyph creation is undefined in the
original and excluded. Destructor EAX is retained but the manager discards it.
No transitive service or earned campaign equivalence is claimed.
"""
import argparse, collections, copy, hashlib, itertools, json, struct
from pathlib import Path
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE
from unicorn.x86_const import *

parser = argparse.ArgumentParser(description=__doc__)
for name in ['game', 'fixture', 'proof']:
    parser.add_argument('--'+name, type=Path, required=True)
parser.add_argument('--control', type=lambda s:int(s, 0), choices=[0xe3f, 0x23f], default=0xe3f)
args = parser.parse_args()
blob = args.game.read_bytes()
sha = lambda b:hashlib.sha256(b).hexdigest()
assert sha(blob) == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe = struct.unpack_from('<I', blob, 60)[0]
opt = pe+24
base = struct.unpack_from('<I', blob, opt+28)[0]
image_size = (struct.unpack_from('<I', blob, opt+56)[0]+4095)&~4095
table = opt+struct.unpack_from('<H', blob, pe+20)[0]
u = Uc(UC_ARCH_X86, UC_MODE_32)
u.mem_map(base, image_size)
for i in range(struct.unpack_from('<H', blob, pe+6)[0]):
    _, va, size, off = struct.unpack_from('<IIII', blob, table+40*i+8)
    if size:u.mem_write(base+va, blob[off:off+size])
heap, stack, code = 0x21000000, 0x22000000, 0x23000000
u.mem_map(heap, 0x20000);u.mem_map(stack, 0x10000);u.mem_map(code, 0x10000)
record = heap+0x1000
objects = [heap+0x2000+0x1000*i for i in range(4)]
updates = [heap+0x6000, heap+0x7000]
dictionary = heap+0x8000
keys = ['ManaBombInitPower', 'ManaBombGlyphDuration', 'ManaBombInRadius',
        'ManaBombOutRadius', 'ManaBombShakeMag', 'ManaBombDeltaPower']
entries_by_key = {key:heap+0x9000+0x100*i for i,key in enumerate(keys)}
values_by_key = {key:heap+0xa000+0x100*i for i,key in enumerate(keys)}
sp, stop, init, cw = stack+0xe000, code+0x100, code+0x200, code+0x300
cache_address, frame_address, fps_address, walls_address = 0x833c04, 0x84ea04, 0x85b3fc, 0x5b8198
roots = [('create',0x530f90,303),('update',0x5310c0,454),('destroy',0x531290,116)]
regions = roots+[('converter',0x419a70,17),('scalar-getter',0x419d40,36),
                 ('indexed-getter',0x419d70,50),('current-mana',0x4eec80,39),
                 ('is-player',0x4e7bc0,19),('walls-setter',0x4e25a0,10)]
read32 = lambda a:struct.unpack('<I', u.mem_read(a,4))[0]
read16 = lambda a:struct.unpack('<H', u.mem_read(a,2))[0]
put32 = lambda a,v:u.mem_write(a,struct.pack('<I',v&0xffffffff))
put16 = lambda a,v:u.mem_write(a,struct.pack('<H',v&0xffff))
bits = lambda x:struct.unpack('<I',struct.pack('<f',x))[0]
point = lambda x,y:[bits(x),bits(y)]
readpoint = lambda a:[read32(a),read32(a+4)]
ref = lambda i:0 if i<0 else objects[i]
index = lambda p:-1 if p==0 else objects.index(p)
string = lambda a:bytes(u.mem_read(a,80)).split(b'\0',1)[0].decode('ascii')
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw))
traces, root_entries, service_counts = {}, collections.Counter(), collections.Counter()
active, current, events, calls, changed, allowed = False, {}, [], collections.Counter(), False, set()

def state():
    actors = [dict(class_bits=read32(o+8),mana=read16(update+4),mass=read32(o+120),
                   pos=readpoint(o+56),velocity=readpoint(o+80),force=readpoint(o+88),
                   pos24=readpoint(o+96)) for o,update in zip(objects,updates)]
    return dict(level=read32(record+8),caster=index(read32(record+16)),mode=read32(record+20),
                pos=readpoint(record+28),expiry=read32(record+68),power=read32(record+72),
                charge=index(read32(record+76)),saved_mass=read32(record+80),exploded=read32(record+84),
                cache=read32(cache_address),frame=read32(frame_address),fps=read32(fps_address),
                walls=read32(walls_address),actors=actors,charge_pos=readpoint(objects[3]+56))

def write32(a,v):
    allowed.update(range(a,a+4));put32(a,v)

def mutate(kind):
    global changed
    n=calls[kind];calls[kind]+=1
    m=current['mutation']
    if changed or m['stage']!=kind or m['at']!=n:return
    changed=True
    action=m['action']
    if action=='caster-b':write32(record+16,objects[1])
    elif action=='caster-nil':write32(record+16,0)
    elif action=='mode-one':write32(record+20,1)
    elif action=='mode-zero':write32(record+20,0)
    elif action=='record-position':
        for j,v in enumerate(point(-73.25,97.125)):write32(record+28+4*j,v)
    elif action=='caster-position':
        for j,v in enumerate(point(8193.125,-4097.25)):write32(objects[0]+56+4*j,v)
    elif action=='power':write32(record+72,0x76543210)
    elif action=='level':write32(record+8,2)
    elif action=='fps':write32(fps_address,120)
    elif action=='frame':write32(frame_address,0xfffffffd)
    elif action=='expiry':write32(record+68,(read32(frame_address)+1)&0xffffffff)
    elif action=='saved-mass':write32(record+80,bits(9.75))
    elif action=='cache':write32(cache_address,137)
    elif action=='walls':write32(walls_address,0)
    else:raise AssertionError(action)

def event(kind,values):
    events.append(dict(kind=kind,args=values,state=state()));mutate(kind)

def ret(value=0):
    esp=u.reg_read(UC_X86_REG_ESP)
    u.reg_write(UC_X86_REG_EAX,value&0xffffffff)
    u.reg_write(UC_X86_REG_EIP,read32(esp));u.reg_write(UC_X86_REG_ESP,esp+4)

def hook(_u,pc,size,_):
    if not active:return
    if pc==stop:u.emu_stop();return
    for name,lo,length in regions:
        if lo<=pc<lo+length:
            traces[pc]=bytes(u.mem_read(pc,size))
            if pc==lo:
                root_entries[name]+=1
                if name=='walls-setter':
                    esp=u.reg_read(UC_X86_REG_ESP);event('walls',[read32(esp+4)])
            return
    esp=u.reg_read(UC_X86_REG_ESP)
    arg=lambda n:read32(esp+4*n)
    service_counts[hex(pc)]+=1
    if pc==0x426890:
        assert arg(1)==dictionary
        key=string(arg(2));assert key in keys
        event('lookup',[key]);ret(0 if key in current['missing'] else entries_by_key[key])
    elif pc==0x4e3aa0:
        assert string(arg(1))=='ManaBombCharge'
        event('name',['ManaBombCharge']);ret(current['resolved_type'])
    elif pc==0x4e3450:
        event('allocate',[arg(1)]);ret(ref(3) if current['allocate'] else 0)
    elif pc==0x4daa50:
        obj,owner=index(arg(1)),index(arg(2));assert obj==3 and owner==-1 and arg(5)==0
        position=[arg(3),arg(4)];event('create',[obj,owner,position,arg(5)])
        for j,v in enumerate(position):write32(objects[obj]+56+4*j,v)
        ret(0x13572468)
    elif pc==0x4ff380:
        assert arg(2) in [5,14,29] and arg(4)==5 and arg(3)<=32767
        event('buff',[index(arg(1)),arg(2),arg(3),arg(4)]);ret(0x13572468)
    elif pc==0x4ff5b0:
        assert arg(2) in [5,14,29]
        event('off',[index(arg(1)),arg(2)]);ret(0x13572468)
    elif pc==0x4e5cc0:
        assert index(arg(1)) in [2,3]
        event('delete',[index(arg(1))]);ret(0x13572468)
    elif pc==0x4e25b0:
        assert arg(5)==15 and arg(7)==0
        event('damage',[readpoint(arg(1)),arg(2),arg(3),arg(4),arg(5),index(arg(6)),arg(7)])
        ret(0x13572468)
    elif pc==0x4d9110:
        event('quake',[readpoint(arg(1)),arg(2)]);ret(0x13572468)
    elif pc==0x522ff0:
        assert arg(1) in [129,154,163]
        event('fx',[arg(1),readpoint(arg(2))]);ret(0x13572468)
    elif pc==0x501960:
        assert arg(1)==81 and arg(3)==arg(4)==0
        event('audio',[81,index(arg(2)),0,0]);ret(0x13572468)
    elif pc==0x4eebf0:
        event('mana',[index(arg(1)),arg(2)]);ret(0x13572468)
    else:raise AssertionError((current['name'],hex(pc)))

u.hook_add(UC_HOOK_CODE,hook)
registers={UC_X86_REG_EBX:0x12345678,UC_X86_REG_ESI:0x23456789,
           UC_X86_REG_EDI:0x3456789a,UC_X86_REG_EBP:0x456789ab}
seed=bytes((i*31+19)&255 for i in range(0x20000))

def run(c):
    global active,current,events,calls,changed,allowed
    active,current,events,calls,changed,allowed=False,c,[],collections.Counter(),False,set()
    u.mem_write(heap,seed)
    for i,o in enumerate(objects):
        put32(o+8,c['class'] if i==0 else 4)
        for off,value in [(56,0),(60,0),(80,bits(1+i)),(84,bits(2+i)),(88,bits(3+i)),
                          (92,bits(4+i)),(96,bits(5+i)),(100,bits(6+i)),(120,bits(17.5+i))]:put32(o+off,value)
    for i,update in enumerate(updates):
        put32(objects[i]+748,update);put16(update+4,c['mana'] if i==0 else 37)
    for j,v in enumerate(c['caster_pos']):put32(objects[0]+56+4*j,v)
    for j,v in enumerate(point(211.25,-41.5)):put32(objects[1]+56+4*j,v)
    for off,value in [(8,c['level']),(16,ref(-1 if c['caster_nil'] else 0)),(20,c['mode']),
                      (68,c['expiry']),(72,c['power']),(76,ref(2 if c['charge'] else -1)),
                      (80,c['saved_mass']),(84,c['exploded'])]:put32(record+off,value)
    for j,v in enumerate(c['pos']):put32(record+28+4*j,v)
    put32(0x6552d8,dictionary)
    for key in keys:
        entry,value=entries_by_key[key],values_by_key[key]
        put32(entry,value);u.mem_write(entry+4,bytes([5]))
        for i in range(5):put32(value+4*i,c['balance'][key][i])
    for address,key in [(cache_address,'cache'),(frame_address,'frame'),(fps_address,'fps'),(walls_address,'walls')]:put32(address,c[key])
    before=bytes(u.mem_read(heap,len(seed)))
    globals_before={a:read32(a) for a in [0x6552d8,cache_address,frame_address,fps_address,walls_address]}
    u.mem_write(cw,struct.pack('<H',args.control));u.emu_start(init,init+8,count=2)
    u.mem_write(sp-512,bytes(1024));put32(sp,stop);put32(sp+4,record)
    for reg,v in registers.items():u.reg_write(reg,v)
    u.reg_write(UC_X86_REG_EFLAGS,0x202);u.reg_write(UC_X86_REG_ESP,sp)
    lo=next(a for name,a,_ in roots if name==c['function'])
    active=True
    try:u.emu_start(lo,stop+1,count=10000)
    except Exception as error:raise AssertionError((c['name'],hex(u.reg_read(UC_X86_REG_EIP)))) from error
    active=False
    assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4,c['name']
    assert all(u.reg_read(reg)==v for reg,v in registers.items()),c['name']
    assert u.reg_read(UC_X86_REG_FPCW)==args.control and u.reg_read(UC_X86_REG_FPSW)&0x3800==0,c['name']
    assert bytes(u.mem_read(sp-512,256))==bytes(256) and bytes(u.mem_read(sp+8,256))==bytes(256),c['name']
    if c['function']=='create':
        for off in [68,72,76,80,84]:allowed.update(range(record+off,record+off+4))
        for o in objects[:2]:
            for off in [80,84,88,92,96,100,120]:allowed.update(range(o+off,o+off+4))
        allowed.update(range(cache_address,cache_address+4))
    elif c['function']=='update':
        for off in [72,76,84]:allowed.update(range(record+off,record+off+4))
        allowed.update(range(walls_address,walls_address+4))
    else:
        allowed.update(range(record+76,record+80))
        for o in objects[:2]:allowed.update(range(o+120,o+124))
    after=bytes(u.mem_read(heap,len(seed)))
    assert all(x==y or heap+i in allowed for i,(x,y) in enumerate(zip(before,after))),c['name']
    assert all(read32(a)==v or a in allowed for a,v in globals_before.items()),c['name']
    return dict(input=c,result=dict(return_value=u.reg_read(UC_X86_REG_EAX),events=copy.deepcopy(events),state=state()))

default=dict(function='update',level=3,mode=0,caster_nil=False,**{'class':4},mana=20,
             frame=100,expiry=120,fps=60,cache=0,resolved_type=71,walls=0,
             charge=True,allocate=True,power=0x12340055,saved_mass=bits(8.25),exploded=0,
             pos=point(100,200),caster_pos=point(123.125,-71.25),missing=[],
             balance={k:[bits(v)]*5 for k,v in zip(keys,[10.75,12.75,20.25,60.5,4.75,3.75])},
             mutation=dict(stage='',at=0,action=''))
cases=[]
def case(name,**kw):
    c=copy.deepcopy(default);c.update(kw,name=name);cases.append(c)
for mode,cls,cache,allocation,level in itertools.product([0,1,0xffffffff],[0,2,4,6],[0,97],[False,True],[1,3,5,6]):
    case(f'create-{len(cases)}',function='create',mode=mode,**{'class':cls},cache=cache,allocate=allocation,level=level)
for mode,nil in itertools.product([1,0xffffffff],[False,True]):
    case(f'create-glyph-{len(cases)}',function='create',mode=mode,caster_nil=nil)
for fps in [0,1,30]:case(f'create-fps-{fps}',function='create',fps=fps)
for mode,cls,mana,charge,phase in itertools.product([0,1],[0,2,4,6],[0,14,15,20],[False,True],['idle','last-ten','explode','expired','wrap']):
    frame,expiry={'idle':(100,120),'last-ten':(111,120),'explode':(119,120),'expired':(121,120),'wrap':(0xffffffff,0)}[phase]
    case(f'update-{len(cases)}',mode=mode,**{'class':cls},mana=mana,charge=charge,frame=frame,expiry=expiry)
for mode,nil,exploded,charge,cls in itertools.product([0,1,0xffffffff],[False,True],[0,1,0xffffffff],[False,True],[0,2,4,6]):
    case(f'destroy-{len(cases)}',function='destroy',mode=mode,caster_nil=nil,exploded=exploded,charge=charge,**{'class':cls})
case('orphan',caster_nil=True,mode=0)
case('nil-glyph-idle',caster_nil=True,mode=1)
case('nil-glyph-explode',caster_nil=True,mode=1,frame=119)
for key in keys:
    function='create' if key in keys[:2] else 'update'
    case('missing-'+key,function=function,mode=1,frame=119 if key in keys[2:5] else 100,missing=[key])
for level in [1,2,3,4,5,6]:case(f'update-level-{level}',level=level)
edges=[0x00000000,0x80000000,bits(0.5),bits(-0.5),bits(1.5),bits(-1.5),bits(3.75),bits(-3.75),
       0x4effffff,0x4f000000,0xcf000000,0xcf000001,0x7f800000,0xff800000,0x7fc12345,0xffc12345]
for key,value in itertools.product(['ManaBombInitPower','ManaBombGlyphDuration','ManaBombShakeMag','ManaBombDeltaPower'],edges):
    balance=copy.deepcopy(default['balance']);balance[key]=[value]*5
    case(f'edge-{key}-{value:08x}',function='create' if key in keys[:2] else 'update',
         mode=1,frame=119 if key=='ManaBombShakeMag' else 100,balance=balance)
named=[('create-player',dict(function='create')),('create-glyph',dict(function='create',mode=1)),
       ('explode',dict(frame=119)),('delete-explode',dict(mana=0)),
       ('delete-idle',dict(mana=14)),('idle',{}),('destroy',dict(function='destroy'))]
for name,kw in named:
    for stage,action in itertools.product(['name','lookup','buff','allocate','create','delete','walls','damage','quake','fx','audio','mana','off'],
        ['caster-b','mode-one','mode-zero','record-position','caster-position','power','level','fps','frame','expiry','saved-mass','cache','walls']):
        case(f'{name}-mutate-{stage}-{action}',**kw,mutation=dict(stage=stage,at=0,action=action))
for stage,at in [('buff',1),('buff',2),('off',1),('off',2),('fx',1),('lookup',1),('lookup',2)]:
    for action in ['caster-b','fps','saved-mass','power','level','walls']:
        function='create' if stage=='buff' else 'destroy' if stage=='off' else 'update'
        case(f'{function}-{stage}-{at}-{action}',function=function,frame=119 if stage in ['fx','lookup'] else 100,
             mutation=dict(stage=stage,at=at,action=action))
case('delete-explode-orphan',mana=0,mutation=dict(stage='delete',at=0,action='caster-nil'))
case('delta-orphan',mutation=dict(stage='lookup',at=0,action='caster-nil'))
fixture=dict(control_word=args.control,original_machine_code_executed=True,
             destructor_EAX_excluded_from_native_semantic_comparison=True,cases=[run(c) for c in cases])
covered={a+i for a,b in traces.items() for i in range(len(b))}
range_proofs=[dict(name=name,address=hex(a),size=n,sha256=sha(bytes(u.mem_read(a,n))),
                   executed_bytes=len(set(range(a,a+n))&covered),missing_bytes=[hex(p) for p in range(a,a+n) if p not in covered]) for name,a,n in regions]
proof=dict(original_sha256=sha(blob),control_word=args.control,contexts=len(cases),regions=range_proofs,
           entries=dict(root_entries),services=dict(service_counts),faults=0,
           register_stack_FPCW_TOP_heap_global_canaries_preserved=True,original_machine_code_executed=True,
           independent_comparator_NOT_YET_BUILT=True,scope=__doc__)
assert args.game.read_bytes()==blob
for path,value in [(args.fixture,fixture),(args.proof,proof)]:
    with path.open('x') as f:json.dump(value,f,indent=2);f.write('\n')
print(json.dumps(proof))
