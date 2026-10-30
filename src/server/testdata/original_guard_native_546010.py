"""Execute unchanged Guard, aggression, recent-sound, facing and Buff bodies.

NoticeThreat, LookAtDamager, InterestingSound, type predicates, heal, RNG and
PushAction are declared services. Pure frame/FPS loads and facing inputs are
read-only observations. All original code, data and raw EAX/AL are retained.
"""
import argparse, collections, copy, gzip, hashlib, itertools, json, struct
from pathlib import Path
from unicorn import Uc, UcError, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_WRITE
from unicorn.x86_const import *

p = argparse.ArgumentParser(description=__doc__)
for name in ['game', 'fixture', 'proof']:
    p.add_argument('--'+name, type=Path, required=True)
p.add_argument('--control', type=lambda s: int(s, 0), choices=[0xe3f, 0x23f], required=True)
a = p.parse_args()
blob = a.game.read_bytes()
sha = lambda data: hashlib.sha256(data).hexdigest()
assert sha(blob) == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe = struct.unpack_from('<I', blob, 60)[0]
op = pe+24
tb = op+struct.unpack_from('<H', blob, pe+20)[0]
base = struct.unpack_from('<I', blob, op+28)[0]
u = Uc(UC_ARCH_X86, UC_MODE_32)
u.mem_map(base, (struct.unpack_from('<I', blob, op+56)[0]+4095)&~4095)
for i in range(struct.unpack_from('<H', blob, pe+6)[0]):
    _, va, n, off = struct.unpack_from('<IIII', blob, tb+i*40+8)
    if n:
        u.mem_write(base+va, blob[off:off+n])
heap, stack, code = 0x21000000, 0x22000000, 0x23000000
for address in [heap, stack, code]:
    u.mem_map(address, 0x10000)
objects = [heap+i*0x1000 for i in range(1, 5)]
unit = objects[0]
updates = [heap+0x6000, heap+0x8000]
items = [heap+0xc000, heap+0xd000, updates[0]+552, updates[1]+552]
sp, stop, init, cw = stack+0xe000, code+0x100, code+0x200, code+0x300
rd = lambda address: struct.unpack('<I', u.mem_read(address, 4))[0]
r16 = lambda address: struct.unpack('<H', u.mem_read(address, 2))[0]
r8 = lambda address: bytes(u.mem_read(address, 1))[0]
active = False
def wr(address, value):
    u.mem_write(address, struct.pack('<I', value&0xffffffff))
    if active:
        write_hook(u, 0, address, 4, value, None)
def w16(address, value):
    u.mem_write(address, struct.pack('<H', value&0xffff))
def w8(address, value):
    u.mem_write(address, bytes([value&255]))
signed = lambda value: struct.unpack('<i', struct.pack('<I', value&0xffffffff))[0]
ref = lambda value: objects.index(value) if value else -1
canonical = lambda value: 0xf0000000+objects.index(value) if value in objects else value
u.mem_write(init, b'\xdb\xe3\xd9\x2d'+struct.pack('<I', cw))
ranges = {0x546010:1019, 0x534120:57, 0x534390:38, 0x5343c0:57,
          0x534440:38, 0x545da0:179, 0x4ff350:35}
original = {pc:bytes(u.mem_read(pc, n)) for pc,n in ranges.items()}
constants = [0x583bcc, 0x583d44, 0x583d48, 0x583d4c, 0x583d84,
             0x5839b8, 0x583c7c, 0x583b28]
data = {pc:bytes(u.mem_read(pc, 4)) for pc in constants}
data[0x5b6650] = bytes(u.mem_read(0x5b6650, 2056))
queries = {0x546096:'frame', 0x546182:'frame', 0x5461e9:'frame',
           0x546395:'frame', 0x5463a0:'tick', 0x545dba:'frame',
           0x545dc6:'tick', 0x545df6:'tick', 0x545e0c:'frame'}
registers = {UC_X86_REG_EBX:0x12345678, UC_X86_REG_ESI:0x23456789,
             UC_X86_REG_EDI:0x3456789a, UC_X86_REG_EBP:0x456789ab}
seed = bytes((i*31+19)&255 for i in range(0x10000))
current, events, traces, writes = {}, [], {}, {}
services, counters = collections.Counter(), collections.Counter()

def state():
    return dict(update=updates.index(rd(unit+748)),
        classes=[rd(v+8) for v in objects], flags=[rd(v+16) for v in objects],
        positions=[[rd(v+56),rd(v+60)] for v in objects],
        previous=[rd(unit+72),rd(unit+76)], direction=r16(unit+124),
        net=rd(unit+36), buffs=rd(unit+340), indices=[r8(v+544) for v in updates],
        aggression=[rd(v+1304) for v in updates], sight=[rd(v+1312) for v in updates],
        sound=[[rd(v+o) for o in [388,396,400,404]] for v in updates],
        last_move=[rd(v+548) for v in updates], enemies=[ref(rd(v+1196)) for v in updates],
        seen_count=[r8(v+1129) for v in updates],
        seen=[[ref(rd(v+1132+j*4)) for j in range(16)] for v in updates],
        items=[[canonical(rd(v+j*4)) for j in range(5)] for v in items],
        fps=rd(0x85b3fc), frame=rd(0x84ea04))

def mutate(stage):
    m = current['mutation']
    if m['stage'] != stage:
        return
    action,value = m['action'],m['value']
    is_ = lambda name: action in [name,'all']
    if is_('update'):
        wr(unit+748, updates[value])
    if is_('enemy'):
        wr(updates[0]+1196, objects[1+value])
        wr(updates[1]+1196, objects[2-value])
    if action == 'nil-enemy':
        wr(updates[0]+1196, 0)
    if is_('aggression'):
        wr(updates[0]+1304, 0x3f800000 if value else 0x7fc12345)
        wr(updates[1]+1304, 0x3f000000 if value else 0)
    if is_('sight'):
        wr(updates[0]+1312, 0x42c80001 if value else 0x7fc54321)
        wr(updates[1]+1312, 0x3f000001)
    if is_('positions'):
        for i,v in enumerate(objects):
            wr(v+56, 0x42c80001+i if value else 0x80000000)
            wr(v+60, 0xc1a00001-i if value else 0x7fc54321)
    if is_('previous'):
        wr(unit+72, 0x3f000001 if value else 0x7fc12345)
        wr(unit+76, 0x80000000 if value else 0x41a00000)
    if is_('direction'):
        w16(unit+124, 0xffff if value else 129)
    if is_('net'):
        wr(unit+36, 0xffffffff if value else 0)
    if is_('buffs'):
        wr(unit+340, 0x20000000 if value else 0)
    if is_('classes'):
        for i,v in enumerate(objects):
            wr(v+8, 4 if value and i==2 else 2)
    if is_('flags'):
        for i,v in enumerate(objects):
            wr(v+16, 0x01000000 if i==value else 4)
    if is_('frame'):
        wr(0x84ea04, 0xfffffff0 if value else 17)
    if is_('fps'):
        wr(0x85b3fc, 0x80000001 if value else 0xffffffff)
    if is_('last-move'):
        wr(updates[0]+548, 0xfffffff0 if value else 1)
        wr(updates[1]+548, 17)
    if is_('sound'):
        for i,v in enumerate(updates):
            for off,bits in [(388,7 if value else 0),(396,0x41a00001+i),
                             (400,0xc1200001-i),(404,0xfffffff0 if value else 3)]:
                wr(v+off, bits)
    if is_('seen'):
        for v in updates:
            w8(v+1129, 3 if value else 0)
            for j in range(16):
                wr(v+1132+j*4, objects[3-j%3])
    if is_('head-args'):
        for i,v in enumerate(items[2:]):
            for j,bits in enumerate([0x42c80001+i,0xc1a00001-i,64,0x6b000004]):
                wr(v+4+j*4, bits)
    if is_('item-args'):
        for i,v in enumerate(items[:2]):
            for j in range(1,5):
                wr(v+j*4, 0x6b000000+i*256+j)

def write_hook(_u, _access, address, size, value, _):
    if not active:
        return
    for i,v in enumerate(items):
        if v<=address and address+size<=v+20:
            assert size==4 and (address-v)%4==0, (hex(address),size)
            writes.setdefault(str(i)+':'+str((address-v)//4),[]).append(
                dict(size=size,value=canonical(value&0xffffffff)))

def event(kind,args=()):
    events.append(dict(kind=kind,args=list(args),state=state()))

def hook(_u, pc, size, _):
    if not active:
        return
    if pc==stop:
        u.emu_stop()
        return
    if any(start<=pc<start+n for start,n in ranges.items()):
        traces[pc]=bytes(u.mem_read(pc,size))
        if pc in queries:
            event(queries[pc])
        if pc==0x534120:
            esp=u.reg_read(UC_X86_REG_ESP)
            assert rd(esp+4)==unit
            vec=rd(esp+8)
            event('facing',[rd(vec),rd(vec+4)])
        return
    esp=u.reg_read(UC_X86_REG_ESP)
    if pc==0x50a260:
        assert rd(esp+4)==unit
        counters['push']+=1
        kind,args='push',[rd(esp+8)]
        n=counters['push']
        mode=current['alias']
        index=(n-1)%2 if mode==0 else (mode+1 if mode in [1,2] else (n-1)%4)
        ret=items[index] if current['accept']&(1<<(n-1)) else 0
        stage='push'+str(n)
    elif pc==0x415fa0:
        assert rd(esp+12)==0x5cc6b8 and rd(esp+16)==42
        kind,args,ret='random',[signed(rd(esp+4)),signed(rd(esp+8)),42],current['random']
        stage=kind
    elif pc in [0x545e60,0x5466b0,0x5466f0,0x534840,0x534a10,0x5411a0]:
        assert rd(esp+4)==unit
        kind={0x545e60:'notice',0x5466b0:'look',0x5466f0:'interesting',
              0x534840:'mimic',0x534a10:'plant',0x5411a0:'heal'}[pc]
        counters[kind]+=1
        args=[]
        ret=current['mimics'][counters[kind]-1] if kind=='mimic' else current[kind]
        stage=kind+str(counters[kind]) if kind=='mimic' else kind
    else:
        raise AssertionError((current['name'],hex(pc)))
    event(kind,args)
    services[kind]+=1
    mutate(stage)
    if kind=='push' and ret:
        wr(ret,args[0])
    u.reg_write(UC_X86_REG_EAX,ret&0xffffffff)
    u.reg_write(UC_X86_REG_EIP,rd(esp))
    u.reg_write(UC_X86_REG_ESP,esp+4)

u.hook_add(UC_HOOK_CODE,hook)
u.hook_add(UC_HOOK_MEM_WRITE,write_hook)
allowed=set()
for ptr,off,n in ([(v,8,12) for v in objects]+[(v,56,8) for v in objects]+
    [(unit,36,4),(unit,72,8),(unit,124,2),(unit,340,4),(unit,748,4)]+
    [(v,388,4) for v in updates]+[(v,396,12) for v in updates]+
    [(v,544,1) for v in updates]+[(v,548,4) for v in updates]+
    [(v,1129,1) for v in updates]+[(v,1132,68) for v in updates]+
    [(v,1304,4) for v in updates]+[(v,1312,4) for v in updates]+
    [(v,0,20) for v in items]):
    allowed.update(range(ptr+off,ptr+off+n))
protected=[]
start=heap
for address in sorted(allowed):
    if address>start:
        protected.append((start-heap,address-heap))
    start=address+1
protected.append((start-heap,len(seed)))

def run(c):
    global active,current,events,writes,counters
    active=False
    current=c
    events=[]
    writes={}
    counters=collections.Counter()
    u.mem_write(heap,seed)
    for i,v in enumerate(objects):
        wr(v+8,c['classes'][i])
        wr(v+16,c['flags'][i])
        wr(v+56,c['positions'][i][0])
        wr(v+60,c['positions'][i][1])
    wr(unit+36,c['net'])
    wr(unit+72,c['previous'][0])
    wr(unit+76,c['previous'][1])
    w16(unit+124,c['direction'])
    wr(unit+340,c['buffs'])
    wr(unit+748,updates[0])
    for i,v in enumerate(updates):
        w8(v+544,0)
        wr(v+548,c['last_move'] if i==0 else 17)
        wr(v+1304,c['aggression'] if i==0 else 0x3f800000)
        wr(v+1312,c['sight'] if i==0 else 0x42c80000)
        for off,bits in zip([388,396,400,404],c['sound'] if i==0 else [7,0x41a00000,0xc1200000,0]):
            wr(v+off,bits)
        wr(v+1196,objects[c['enemy']] if i==0 and c['enemy']>=0 else (0 if i==0 else objects[2]))
        w8(v+1129,len(c['seen']) if i==0 else 2)
        for j in range(16):
            wr(v+1132+j*4,objects[c['seen'][j]] if i==0 and j<len(c['seen']) else (0 if i==0 else objects[1+j%3]))
    for i,v in enumerate(items):
        wr(v,4 if i>=2 else 0xa5a50000+i)
        for j in range(1,5):
            bits=0x5a000000+i*256+j
            if i>=2 and j<=3:
                bits=c['guard'][j-1] if i==2 else [0x41a00000,0xc1200000,64][j-1]
            wr(v+j*4,bits)
    wr(0x85b3fc,c['fps'])
    wr(0x84ea04,c['frame'])
    before=bytes(u.mem_read(heap,len(seed)))
    wr(cw,a.control)
    u.emu_start(init,init+8,count=2)
    u.mem_write(sp-512,bytes(1024))
    wr(sp,stop)
    wr(sp+4,0 if c['entry']=='nil-buff' else unit)
    wr(sp+8,29 if c['entry']=='nil-buff' else 0)
    for reg,value in registers.items():
        u.reg_write(reg,value)
    u.reg_write(UC_X86_REG_ESP,sp)
    u.reg_write(UC_X86_REG_EFLAGS,0x202)
    active=True
    fault=False
    try:
        u.emu_start(0x4ff350 if c['entry']=='nil-buff' else 0x546010,stop+1,count=10000)
    except UcError as error:
        pc=u.reg_read(UC_X86_REG_EIP)
        assert c['mutation']['action']=='nil-enemy' and pc in [0x546089,0x546176],(c['name'],hex(pc),str(error))
        fault=True
    active=False
    if not fault:
        assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4,c['name']
        assert all(u.reg_read(reg)==value for reg,value in registers.items()),c['name']
    assert u.reg_read(UC_X86_REG_FPCW)==a.control and u.reg_read(UC_X86_REG_FPSW)&0x3800==0,c['name']
    assert bytes(u.mem_read(sp-512,256))==bytes(256) and bytes(u.mem_read(sp+12,256))==bytes(256),c['name']
    after=bytes(u.mem_read(heap,len(seed)))
    assert all(before[x:y]==after[x:y] for x,y in protected),c['name']
    eax=u.reg_read(UC_X86_REG_EAX)
    return dict(input=c,result=dict(events=copy.deepcopy(events),state=state(),fault=fault,item_writes=copy.deepcopy(writes)),raw_original_EAX=eax,raw_original_AL=eax&255)

cases=[]
def case(prefix,**kw):
    c=dict(entry='root',aggression=0x3f000000,sight=0x41200000,
        positions=[[0,0],[0x3f800000,0x3f800000],[0xc1200000,0],[0,0x41200000]],
        previous=[0,0],guard=[0,0,0],classes=[2,2,4,2],flags=[0x1000004,0x1000000,4,4],
        direction=0,net=0,buffs=0,enemy=-1,seen=[],sound=[0,0x41a00001,0xc1200001,0],
        last_move=0,fps=30,frame=16,random=31,notice=0,look=0,interesting=0,
        mimics=[0,0,0,0],plant=0,heal=0x87654321,accept=255,alias=0,
        mutation=dict(stage='',action='',value=0))
    c.update(kw,name=prefix+'-'+str(len(cases)))
    cases.append(c)

agg=[0,0x3da3d709,0x3da3d70a,0x3da3d70b,0x3ea8f5c2,0x3ea8f5c3,
     0x3ea8f5c4,0x3f28f5c2,0x3f28f5c3,0x3f28f5c4,0x3f800000,
     0x7f800000,0xff800000,0x7fc12345,0x7fa12345]
for aggression,notice,enemy,mimic,look in itertools.product(agg,[0,0x87654321],[-1,1],[0,1],[0,1]):
    case('root-branches',aggression=aggression,notice=notice,enemy=enemy,
         mimics=[mimic]*4,look=look,interesting=look)
for plant,accept in itertools.product([0,1],range(128)):
    case('moderate-capacity',enemy=1,plant=plant,accept=accept,net=1)
for aggression,accept in itertools.product([0,0x3f000000,0x3f800000],range(16)):
    case('return-to-guard',aggression=aggression,guard=[0x41200000,0,0],accept=accept)
for mask,direction,guard_direction,accept in itertools.product([0,5,10,15],[0,1,96,255,65535],
    [0,1,64,255,0xffffffff,0x20000000,0x20000001],range(4)):
    case('direction-and-live-mimic',mimics=[(mask>>i)&1 for i in range(4)],
         direction=direction,guard=[0,0,guard_direction],accept=accept)
for count,player,enabled,direction,accept in itertools.product([1,2,3,16],[-1,0,1,2],
    [1,2],[0,64,128,192],range(4)):
    seen=[1+i%3 for i in range(count)]
    classes=[2,2,2,2]
    if player>=0:
        classes[1+player]=4
    flags=[4,4,4,4]
    flags[enabled]=0x1000000
    case('seen-class-selection',seen=seen,classes=classes,flags=flags,direction=direction,accept=accept)
bits=[0,0x80000000,1,0x00800000,0x3f666665,0x3f666666,0x3f666667,0x42c80001,
      0x4b7fffff,0x7f7fffff,0x7f800000,0xff800000,0x7fc12345,0x7fa12345]
for x,y,direction in itertools.product(bits,[0,0x3f000001,0x7fc54321],[0,1,64,255]):
    case('target-normalization-bits',seen=[1],positions=[[0,0],[x,y],[0,0],[0,0]],direction=direction)
for x,y in itertools.product(bits,bits):
    case('return-and-position-comparison-bits',positions=[[x,y],[0,0],[0,0],[0,0]],previous=[0,0],net=1)
for fps,edge,accept,random in itertools.product([0,1,30,0x55555555,0x80000000,0x80000001,0xffffffff],
    [-1,0,1,2],range(16),[-7,31]):
    delta=((3*fps)&0xffffffff)+edge
    case('recent-sound-DWORD',fps=fps,frame=0xfffffffd,
         sound=[7,0x41a00001,0xc1200001,(0xfffffffd-delta)&0xffffffff],accept=accept,random=random,net=1)
for alias,accept,kind in itertools.product([1,2,3],range(32),['moderate','move','recent']):
    kw=dict(alias=alias,accept=accept,net=1)
    if kind=='moderate':
        kw.update(enemy=1)
    if kind=='move':
        kw.update(guard=[0x41200000,0,0],net=0)
    if kind=='recent':
        kw.update(sound=[7,0x41a00000,0xc1200000,16])
    case('slot-alias-'+kind,**kw)
base_contexts=[dict(enemy=1,aggression=0x3f800000),dict(enemy=1,plant=0,net=1),
    dict(enemy=1,plant=1,net=1),dict(sound=[7,0x41a00000,0xc1200000,16]),
    dict(aggression=0x3f800000,guard=[0x41200000,0,0]),dict(seen=[1,2,3]),
    dict(mimics=[1,1,1,1],guard=[0,0,128]),dict(previous=[0x3f800000,0],net=1),
    dict(aggression=0x3f800000)]
actions=['update','enemy','aggression','sight','positions','previous','direction','net',
    'buffs','classes','flags','frame','fps','last-move','sound','seen','head-args','item-args','all']
for base_context in base_contexts:
    case('mutation-route',**base_context)
    route=cases.pop()
    result=run(route)
    counts=collections.Counter()
    stages=[]
    for e in result['result']['events']:
        kind=e['kind']
        if kind not in ['push','mimic','notice','look','interesting','plant','heal','random']:
            continue
        counts[kind]+=1
        stage=kind+str(counts[kind]) if kind in ['push','mimic'] else kind
        stages.append(stage)
    for stage,action,value in itertools.product(stages,actions,range(2)):
        case('live-callback',**base_context,mutation=dict(stage=stage,action=action,value=value))
for aggression,stage in [(0x3f800000,'push1'),(0x3f000000,'push4')]:
    case('nil-live-enemy-fault',enemy=1,aggression=aggression,net=1,
         mutation=dict(stage=stage,action='nil-enemy',value=0))
case('nil-Buff-ancillary',entry='nil-buff')
traces.clear()
services.clear()
rows=[run(c) for c in cases]
visited={pc+i for pc,raw in traces.items() for i in range(len(raw))}
expected={pc+i for pc,n in ranges.items() for i in range(n)}
assert visited==expected,[(hex(pc),hex(pc+1)) for pc in sorted(expected-visited)]
assert all(bytes(u.mem_read(pc,len(raw)))==raw for pc,raw in original.items())
assert all(bytes(u.mem_read(pc,len(raw)))==raw for pc,raw in data.items())
fixture=dict(original_sha256=sha(blob),control_word=a.control,rows=rows)
proof=dict(original_sha256=sha(blob),control_word=a.control,contexts=len(rows),
    fault_contexts=sum(r['result']['fault'] for r in rows),
    root_bytes={hex(pc):n for pc,n in ranges.items()},executed_bytes=len(visited),
    root_sha256={hex(pc):sha(raw) for pc,raw in original.items()},
    original_code_and_data_unmodified=True,register_stack_FPCW_TOP_heap_canaries_preserved=True,
    raw_original_EAX_AL_retained_and_void_update_caller_proven_separately=True,
    declared_services=dict(services),
    read_only_frame_FPS_and_facing_input_observations=True,
    original_item_write_history_for_native_pointer_and_DWORD_expansion=True,
    separate_padding_Nops_not_executed=[5,7,10,7,10,13,13],
    original_data=[dict(address=hex(pc),bytes=raw.hex(),sha256=sha(raw)) for pc,raw in data.items()],
    instructions=[dict(pc=hex(pc),bytes=raw.hex()) for pc,raw in sorted(traces.items())])
for path,value in [(a.fixture,fixture),(a.proof,proof)]:
    if path.suffix == '.gz':
        with path.open('xb') as f:
            with gzip.GzipFile(filename='',mode='wb',fileobj=f,compresslevel=9,mtime=0) as stream:
                stream.write(json.dumps(value,separators=(',',':')).encode()+b'\n')
    else:
        with path.open('x') as f:
            json.dump(value,f,indent=2)
            f.write('\n')
print(json.dumps({k:v for k,v in proof.items() if k not in ['instructions','original_data']}))
