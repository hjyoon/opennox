"""Execute complete original Force of Nature roots and original buff helper.

Six world-service bodies are declared: allocation, creation, deletion, ray,
player-state and audio. Original field reads/stores, signed shape branches,
direction-table reads and every x87 operation execute unchanged. Directions
are stock0..255. The duration manager discards destructor EAX and its dead
record's charge slot; both raw observations are retained in the fixture.
No transitive world-service or earned gameplay equivalence is claimed.
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
    if size:
        u.mem_write(base+va, blob[off:off+size])
heap, stack, code = 0x21000000, 0x22000000, 0x23000000
u.mem_map(heap, 0x20000)
u.mem_map(stack, 0x10000)
u.mem_map(code, 0x10000)
record = heap+0x1000
objects = [heap+0x2000+0x1000*i for i in range(8)]
updates = [heap+0xa000, heap+0xb000]
uses = [heap+0xc000, heap+0xd000]
sp, stop, init, cw = stack+0xe000, code+0x100, code+0x200, code+0x300
frame_address = 0x84ea04
roots = [('create', 0x52ef30, 149), ('update', 0x52efd0, 510), ('destroy', 0x52f1d0, 68)]
regions = roots+[('buff', 0x4ff350, 35)]
read32 = lambda a:struct.unpack('<I', u.mem_read(a, 4))[0]
read16 = lambda a:struct.unpack('<H', u.mem_read(a, 2))[0]
put32 = lambda a,v:u.mem_write(a, struct.pack('<I', v&0xffffffff))
put16 = lambda a,v:u.mem_write(a, struct.pack('<H', v&0xffff))
bits = lambda x:struct.unpack('<I', struct.pack('<f', x))[0]
point = lambda x,y:[bits(x), bits(y)]
readpoint = lambda a:[read32(a), read32(a+4)]
ref = lambda i:0 if i < 0 else objects[i]
index = lambda p:-1 if p == 0 else objects.index(p)
u.mem_write(init, b'\xdb\xe3\xd9\x2d'+struct.pack('<I', cw))
traces, entries, service_counts = {}, collections.Counter(), collections.Counter()
active, current, events, calls, changed, allowed = False, {}, [], collections.Counter(), False, set()

def state():
    return dict(caster=index(read32(record+16)), mode=read32(record+20),
                anchor=index(read32(record+24)), pos=readpoint(record+28),
                field72=read32(record+72), charge=index(read32(record+76)),
                flags88=read32(record+88), frame=read32(frame_address), expiry=read32(record+68),
                caster_pos=readpoint(objects[0]+56), caster_direction=read16(objects[0]+124),
                wand_flags=read32(uses[0]+96), alternate_wand_flags=read32(uses[1]+96),
                charge_pos=readpoint(objects[4]+56), ball_pos=readpoint(objects[5]+56),
                ball_velocity=readpoint(objects[5]+80), ball_speed=read32(objects[5]+544),
                ball_direction=[read16(objects[5]+124), read16(objects[5]+126)])

def write32(a, v):
    allowed.update(range(a, a+4))
    put32(a, v)

def mutate(kind):
    global changed
    n = calls[kind]
    calls[kind] += 1
    m = current['mutation']
    if changed or m['stage'] != kind or m['at'] != n:
        return
    changed = True
    action = m['action']
    if action == 'caster-c': write32(record+16, objects[2])
    elif action == 'mode-one': write32(record+20, 1)
    elif action == 'mode-zero': write32(record+20, 0)
    elif action == 'caster-position':
        for j,v in enumerate(point(8193.125, -4097.25)): write32(objects[0]+56+4*j, v)
    elif action == 'caster-direction':
        allowed.update(range(objects[0]+124, objects[0]+126));put16(objects[0]+124, 192)
    elif action == 'record-position':
        for j,v in enumerate(point(-73.25, 97.125)): write32(record+28+4*j, v)
    elif action == 'frame-launch': write32(frame_address, (read32(record+68)-1)&0xffffffff)
    elif action == 'expiry-next': write32(record+68, (read32(frame_address)+1)&0xffffffff)
    elif action == 'ball-speed': write32(objects[5]+544, bits(3.125))
    elif action == 'wand-flags': write32(uses[0]+96, 0xfffffffd)
    else: raise AssertionError(action)

def event(kind, values):
    events.append(dict(kind=kind, args=values, state=state()))
    mutate(kind)

def ret(value=0):
    esp = u.reg_read(UC_X86_REG_ESP)
    u.reg_write(UC_X86_REG_EAX, value&0xffffffff)
    u.reg_write(UC_X86_REG_EIP, read32(esp))
    u.reg_write(UC_X86_REG_ESP, esp+4)

def hook(_u, pc, size, _):
    if not active: return
    if pc == stop: u.emu_stop();return
    for name, lo, length in regions:
        if lo <= pc < lo+length:
            traces[pc] = bytes(u.mem_read(pc, size))
            if pc == lo: entries[name] += 1
            return
    esp = u.reg_read(UC_X86_REG_ESP)
    arg = lambda n:read32(esp+4*n)
    service_counts[hex(pc)] += 1
    if pc == 0x4e3810:
        name = bytes(u.mem_read(arg(1), 40)).split(b'\0', 1)[0].decode('ascii')
        assert name == ('ForceOfNatureCharge' if current['function'] == 'create' else 'DeathBall')
        event('new', [name]);ret(ref(4 if current['function'] == 'create' else 5) if current['allocate'] else 0)
    elif pc == 0x4daa50:
        obj, owner = index(arg(1)), index(arg(2))
        assert obj in [4, 5] and arg(5) == 0
        position = [arg(3), arg(4)]
        event('create', [obj, owner, position, arg(5)])
        for j,v in enumerate(position): write32(objects[obj]+56+4*j, v)
        ret(0x13572468)
    elif pc == 0x4e5cc0:
        assert index(arg(1)) in [3, 4]
        event('delete', [index(arg(1))]);ret(0x13572468)
    elif pc == 0x535250:
        assert arg(2) == arg(3) == 0 and arg(4) == 5
        event('trace', [readpoint(arg(1)), readpoint(arg(1)+8), arg(4)])
        ret(int(current['trace']))
    elif pc == 0x4fa020:
        assert arg(2) == 10
        event('state', [index(arg(1)), arg(2)]);ret(0x13572468)
    elif pc == 0x501960:
        assert arg(1) == 38 and arg(3) == arg(4) == 0
        event('audio', [arg(1), index(arg(2)), arg(3), arg(4)]);ret(0x13572468)
    else: raise AssertionError((current['name'], hex(pc)))

u.hook_add(UC_HOOK_CODE, hook)
registers = {UC_X86_REG_EBX:0x12345678, UC_X86_REG_ESI:0x23456789,
             UC_X86_REG_EDI:0x3456789a, UC_X86_REG_EBP:0x456789ab}
seed = bytes((i*31+19)&255 for i in range(0x20000))

def run(c):
    global active, current, events, calls, changed, allowed
    active, current, events, calls, changed, allowed = False, c, [], collections.Counter(), False, set()
    u.mem_write(heap, seed)
    for i,obj in enumerate(objects):
        put32(obj+8, 2);put32(obj+12, 0);put32(obj+16, 0);put32(obj+340, 0)
        for off in [56, 60, 80, 84]: put32(obj+off, bits(0))
        put16(obj+124, 0);put16(obj+126, 0);put32(obj+544, c['speed'])
        put32(obj+172, 0);put32(obj+176, bits(0));put32(obj+184, bits(0));put32(obj+188, bits(0))
    for off,value in [(16, ref(-1 if c['caster_nil'] else 0)), (20, c['mode']), (24, objects[1]),
                      (68, c['expiry']), (72, c['field72']), (76, ref(3 if c['charge'] else -1)), (88, c['flags88'])]: put32(record+off, value)
    for j,v in enumerate(c['pos']): put32(record+28+4*j, v)
    for j,v in enumerate(c['caster_pos']): put32(objects[0]+56+4*j, v)
    for j,v in enumerate(point(71.125, 93.5)): put32(objects[1]+56+4*j, v)
    for j,v in enumerate(point(211.25, -41.5)): put32(objects[2]+56+4*j, v)
    put16(objects[0]+124, c['direction']);put16(objects[1]+124, 192);put16(objects[2]+124, 33)
    put32(objects[0]+8, c['class']);put32(objects[2]+8, 4);put32(objects[0]+340, c['buffs'])
    put32(objects[0]+172, c['shape']);put32(objects[0]+176, c['radius'])
    put32(objects[0]+184, c['width']);put32(objects[0]+188, c['height'])
    for obj,update,wand,use in [(objects[0], updates[0], objects[6], uses[0]), (objects[2], updates[1], objects[7], uses[1])]:
        put32(obj+748, update);put32(update+104, wand if c['wand'] else 0)
        put32(wand+12, c['wand_subclass']);put32(wand+736, use);put32(use+96, c['wand_flags'])
    put32(frame_address, c['frame'])
    before = bytes(u.mem_read(heap, len(seed)))
    u.mem_write(cw, struct.pack('<H', args.control));u.emu_start(init, init+8, count=2)
    u.mem_write(sp-512, bytes(1024));put32(sp, stop);put32(sp+4, record)
    for reg,v in registers.items():u.reg_write(reg, v)
    u.reg_write(UC_X86_REG_EFLAGS, 0x202);u.reg_write(UC_X86_REG_ESP, sp)
    lo = next(a for name,a,_ in roots if name == c['function'])
    active = True;u.emu_start(lo, stop+1, count=10000);active = False
    assert u.reg_read(UC_X86_REG_EIP) == stop and u.reg_read(UC_X86_REG_ESP) == sp+4
    assert all(u.reg_read(reg) == v for reg,v in registers.items())
    assert u.reg_read(UC_X86_REG_FPCW) == args.control and u.reg_read(UC_X86_REG_FPSW)&0x3800 == 0
    assert bytes(u.mem_read(sp-512, 256)) == bytes(256) and bytes(u.mem_read(sp+8, 256)) == bytes(256)
    for off,n in [(72,2), (76,4), (88,1)]:allowed.update(range(record+off, record+off+n))
    for off,n in [(80,8), (124,4)]:allowed.update(range(objects[5]+off, objects[5]+off+n))
    if c['function'] == 'destroy':
        for use in uses: allowed.update(range(use+96, use+100))
    after = bytes(u.mem_read(heap, len(seed)))
    assert all(x == y or heap+i in allowed for i,(x,y) in enumerate(zip(before, after))), c['name']
    assert read32(frame_address) == c['frame'] or frame_address in allowed
    return dict(input=c, result=dict(return_value=u.reg_read(UC_X86_REG_EAX), events=copy.deepcopy(events), state=state()))

default = dict(function='update', mode=0, caster_nil=False, **{'class':4}, buffs=0,
               frame=100, expiry=101, charge=True, allocate=True, trace=True,
               field72=0x6a5a0080, flags88=0x12340001, pos=point(100,200),
               caster_pos=point(123.125,-71.25), direction=33, shape=2,
               radius=bits(3.25), width=bits(7.125), height=bits(9.25), speed=bits(5.125),
               wand=True, wand_subclass=0x200000, wand_flags=5,
               mutation=dict(stage='', at=0, action=''))
cases = []
def case(name, **kw):
    c = copy.deepcopy(default);c.update(kw, name=name);cases.append(c)
for mode,cls,wand,sub,flags,allocation in itertools.product([0,1,0xffffffff], [2,4], [False,True], [0,0x200000], [0,4,0xffffffff], [False,True]):
    case(f'create-{len(cases)}', function='create', mode=mode, **{'class':cls}, wand=wand, wand_subclass=sub, wand_flags=flags, allocate=allocation)
shapes = [(0,3.25,7.125,9.25),(1,3.25,7.125,9.25),(2,3.25,7.125,9.25),
          (3,3.25,7.125,9.25),(3,3.25,9.25,7.125),(3,3.25,7.125,7.125),(4,3.25,7.125,9.25)]
for shape,r,w,h in shapes:
    for direction,cls,trace in itertools.product([0,1,33,64,127,128,192,255], [2,4], [False,True]):
        case(f'launch-{len(cases)}', shape=shape, radius=bits(r), width=bits(w), height=bits(h), direction=direction, **{'class':cls}, trace=trace)
for mode,direction in itertools.product([1,0xffffffff], [0,1,33,64,127,128,192,255]):
    case(f'anchored-{len(cases)}', mode=mode, direction=direction, field72=0x6a5a0000|direction, caster_nil=True)
for phase,frame,expiry in [('idle',100,120),('delete',113,120),('wrap-launch',0xffffffff,0),('wrap-delete',0xfffffffc,3)]:
    for charge,allocation,cls,mode in itertools.product([False,True], [False,True], [2,4], [0,1]):
        case(f'{phase}-{len(cases)}', frame=frame, expiry=expiry, charge=charge, allocate=allocation, **{'class':cls}, mode=mode)
case('buffed', buffs=1<<8)
case('idle-nil-caster', frame=100, expiry=120, caster_nil=True)
case('allocation-fails', allocate=False)
for value in [0x00000000,0x80000000,0x4b800001,0xcb800001,0x7f7fffff,0x7f800000,0xff800000,0x7fc12345,0xffc12345]:
    case(f'edge-{value:08x}', caster_pos=[value,value], radius=value, speed=value)
    case(f'box-edge-{value:08x}', shape=3, width=value, height=bits(7.125))
named = [('create',dict(function='create')), ('launch',{}), ('delete',dict(frame=113,expiry=120)),
         ('destroy',dict(function='destroy')), ('anchored',dict(mode=1)), ('idle',dict(frame=100,expiry=120))]
for name,kw in named:
    for stage,action in itertools.product(['new','create','delete','trace','state','audio'],
        ['caster-c','mode-one','mode-zero','caster-position','caster-direction','record-position','frame-launch','expiry-next','ball-speed','wand-flags']):
        case(f'{name}-mutate-{stage}-{action}', **kw, mutation=dict(stage=stage,at=0,action=action))
for nil,cls,wand,sub,flags,charge in itertools.product([False,True],[2,4],[False,True],[0,0x200000],[0,4,0xffffffff],[False,True]):
    case(f'destroy-{len(cases)}', function='destroy', caster_nil=nil, **{'class':cls}, wand=wand, wand_subclass=sub, wand_flags=flags, charge=charge)
fixture = dict(control_word=args.control, original_machine_code_executed=True,
               destructor_EAX_and_dead_record_charge_excluded_from_native_semantic_comparison=True,
               cases=[run(c) for c in cases])
covered = {a+i for a,b in traces.items() for i in range(len(b))}
range_proofs = [dict(name=name,address=hex(a),size=n,sha256=sha(bytes(u.mem_read(a,n))),
                     executed_bytes=len(set(range(a,a+n))&covered),missing_bytes=[hex(p) for p in range(a,a+n) if p not in covered]) for name,a,n in regions]
data_proofs = [dict(name=name,address=hex(a),size=n,sha256=sha(bytes(u.mem_read(a,n)))) for name,a,n in
               [('directions',0x5b6658,2048),('zero',0x58307c,4),('four',0x583a1c,4),('twenty-four',0x583d2c,4)]]
proof = dict(original_sha256=sha(blob),control_word=args.control,contexts=len(cases),regions=range_proofs,
             data_regions=data_proofs,entries=dict(entries),services=dict(service_counts),faults=0,
             register_stack_FPCW_TOP_heap_canaries_preserved=True,original_machine_code_executed=True,
             independent_comparator_NOT_YET_BUILT=True,scope=__doc__)
assert args.game.read_bytes() == blob
for path,value in [(args.fixture,fixture),(args.proof,proof)]:
    with path.open('x') as f:json.dump(value,f,indent=2);f.write('\n')
print(json.dumps(proof))
