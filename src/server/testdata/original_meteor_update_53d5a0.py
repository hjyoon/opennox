"""Execute complete unchanged GAME.EXE Meteor update bodies and integer conversion."""
import argparse
import collections
import copy
import hashlib
import json
from pathlib import Path
import struct
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import *

parser = argparse.ArgumentParser(description=__doc__)
for key in ('game', 'fixture', 'proof'):
    parser.add_argument('--' + key, type=Path, required=True)
parser.add_argument('--fault-proof', type=Path)
args = parser.parse_args()
blob = args.game.read_bytes()
sha = lambda b: hashlib.sha256(b).hexdigest()
assert sha(blob) == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe = struct.unpack_from('<I', blob, 60)[0]
opt = pe + 24
base = struct.unpack_from('<I', blob, opt + 28)[0]
size = (struct.unpack_from('<I', blob, opt + 56)[0] + 4095) & ~4095
tab = opt + struct.unpack_from('<H', blob, pe + 20)[0]
u = Uc(UC_ARCH_X86, UC_MODE_32)
u.mem_map(base, size)
for n in range(struct.unpack_from('<H', blob, pe + 6)[0]):
    _, va, count, offset = struct.unpack_from('<IIII', blob, tab + 40*n + 8)
    if count:
        u.mem_write(base + va, blob[offset:offset + count])
stack, heap, stop = 0x20000000, 0x21000000, 0x30000000
for p in (stack, heap, stop):
    u.mem_map(p, 0x10000)
sp = stack + 0xe000
get = lambda p: struct.unpack('<I', u.mem_read(p, 4))[0]
put = lambda p, v: u.mem_write(p, struct.pack('<I', v & 0xffffffff))
f32 = lambda v: struct.unpack('<I', struct.pack('<f', v))[0]
f64 = lambda v: struct.unpack('<Q', struct.pack('<d', v))[0]
s32 = lambda v: struct.unpack('<i', struct.pack('<I', v & 0xffffffff))[0]
objects = [heap + 128 + 1024*n for n in range(4)]
data = [heap + 8192 + 128*n for n in range(3)]
offsets = (8, 16, 20, 56, 60, 104, 108, 128, 136)
frame_address, fps_address = 0x84ea04, 0x85b3fc
specs = [('shower', 0x53d5a0, 319), ('meteor', 0x53d6e0, 368), ('integer', 0x419a70, 17)]
regions = [dict(name=n, address=hex(p), size=z, sha256=sha(bytes(u.mem_read(p, z)))) for n,p,z in specs]
directions = [get(0x5b6658 + 4*n) for n in range(512)]
table = dict(address='0x5b6658', size=2048, sha256=sha(bytes(u.mem_read(0x5b6658, 2048))), words=directions)
ref = lambda n: 0 if n < 0 else objects[n]
objref = lambda p: -1 if p == 0 else objects.index(p)
dataref = lambda p: -1 if p == 0 else data.index(p)
def state():
    return dict(objects=[[get(p + off) for off in offsets] for p in objects],
                owners=[objref(get(p + 508)) for p in objects],
                update_refs=[dataref(get(p + 748)) for p in objects],
                data=[get(p) for p in data], frame=get(frame_address), fps=get(fps_address))

events, faults, trace = [], [], {}
entries = collections.Counter()
current = {}
recording = False
def event(name, values):
    events.append(dict(service=name, args=values, **state()))
def ret(value=0):
    esp = u.reg_read(UC_X86_REG_ESP)
    u.reg_write(UC_X86_REG_EAX, value & 0xffffffff)
    u.reg_write(UC_X86_REG_EIP, get(esp))
    u.reg_write(UC_X86_REG_ESP, esp + 4)
def mutate(service, pointer=0):
    key = current['mutation']
    if not key.startswith(service + ':'):
        return
    action = key.split(':', 1)[1]
    if action == 'position':
        put(objects[0]+56, f32(500.1)); put(objects[0]+60, f32(-600.3))
    elif action == 'owner': put(objects[0]+508, objects[3])
    elif action == 'owner-nil': put(objects[0]+508, 0)
    elif action == 'update': put(objects[0]+748, data[2])
    elif action == 'spawn-update': put(objects[1]+748, data[2])
    elif action == 'damage': put(data[0], 73)
    elif action == 'negative-damage': put(data[0], 0x80000001)
    elif action == 'clock': put(frame_address, 0xfffffffc); put(fps_address, 0)
    elif action == 'created': put(objects[0]+128, get(frame_address))
    elif action == 'flags': put(objects[1]+20, 0x400)
    elif action == 'velocity': put(objects[1]+108, f32(101))
    elif action == 'point': put(pointer, f32(611.3)); put(pointer+4, f32(-817.7))
    elif action == 'trace-point': put(pointer+8, f32(17.3)); put(pointer+12, f32(-91.7))
    else: raise AssertionError(key)

# A declared random service returns its exact input double through x87 ST0.
# This FLD/RET return stub is outside the original image and performs no math.
stub, radius_value = stop + 0x100, stop + 0x2000
u.mem_write(stub, b'\xdd\x05' + struct.pack('<I', radius_value) + b'\xc3')
def hook(_u, address, length, _):
    if address in (stop, 0x4014c4):
        u.emu_stop(); return
    if not recording or stub <= address < stub + 7:
        return
    for name, lo, size in specs:
        if lo <= address < lo + size:
            trace[address] = bytes(u.mem_read(address, length)).hex()
            if address == lo: entries[name] += 1
            return
    esp = u.reg_read(UC_X86_REG_ESP)
    arg = lambda n: get(esp + 4*n)
    if address == 0x415fa0:
        bounds = [s32(arg(1)), s32(arg(2))]
        service = 'direction' if bounds == [0, 255] else 'delay'
        assert bounds in ([0,255], [4,8])
        event(service, bounds); mutate(service)
        ret(current['direction'] if service == 'direction' else current['delay'])
    elif address == 0x416030:
        assert [arg(1),arg(2)] == [f32(4), f32(12)]
        event('radius', [arg(1),arg(2)]); mutate('radius')
        u.reg_write(UC_X86_REG_EIP, stub)
    elif address == 0x5374b0:
        p = arg(1); assert arg(2) == objects[0]
        event('trace', [get(p+4*n) for n in range(4)] + [0]); mutate('trace', p)
        ret(current['trace_result'])
    elif address == 0x4e3810:
        raw = bytes(u.mem_read(arg(1), 32)).split(b'\0',1)[0].decode()
        assert raw == ('Meteor' if current['kind'] == 0 else 'MeteorExplode')
        event('new', [current['kind']]); mutate('new')
        ret(objects[1] if current['allocate'] else 0)
    elif address == 0x4daa50:
        assert arg(1) == objects[1] and arg(5) == 0
        owner, x, y = arg(2), arg(3), arg(4)
        event('create', [1,objref(owner),x,y,0]); mutate('create')
        put(objects[1]+508, owner); put(objects[1]+56, x); put(objects[1]+60, y); ret()
    elif address == 0x4e46f0:
        assert arg(1) == objects[1] and arg(2) == f32(255)
        event('raise', [1,arg(2)]); mutate('raise'); put(objects[1]+104,arg(2)); ret()
    elif address == 0x501960:
        assert [arg(n) for n in (1,2,3,4)] == [87,objects[0],0,0]
        event('audio', [87,0,0,0]); mutate('audio'); ret()
    elif address in (0x537af0, 0x4d9110):
        assert arg(1) == objects[0]+56
        name, expected = ('scorch',2) if address == 0x537af0 else ('quake',10)
        assert arg(2) == expected
        p = arg(1); event(name,[expected,get(p),get(p+4)]); mutate(name,p); ret()
    elif address == 0x4ec580:
        assert arg(1) == objects[0]
        owner = get(objects[0]+508)
        event('parent',[0]); mutate('parent'); ret(owner)
    elif address == 0x4e25b0:
        assert arg(1) == objects[0]+56 and [arg(2),arg(3),arg(5),arg(7)] == [f32(80),f32(30),7,0]
        p=arg(1); event('units',[get(p),get(p+4),arg(2),arg(3),s32(arg(4)),7,objref(arg(6)),-1]); mutate('units',p); ret()
    elif address == 0x534fc0:
        assert arg(2) == objects[0]+56 and [arg(3),arg(5),arg(6)] == [f32(80),7,objects[0]]
        p=arg(2); event('walls',[s32(get(arg(1)+4*n)) for n in range(4)] + [get(p),get(p+4),arg(3),s32(arg(4)),7,0]); mutate('walls',p); ret()
    elif address == 0x4e5cc0:
        assert arg(1) == objects[0]
        event('delete',[0]); mutate('delete'); ret()
    else:
        raise AssertionError((hex(address),current['name']))
def invalid(_u, access, address, length, value, _):
    faults.append(dict(address=address,size=length,instruction=u.reg_read(UC_X86_REG_EIP)))
    return False
u.hook_add(UC_HOOK_CODE,hook)
u.hook_add(UC_HOOK_MEM_INVALID,invalid)
u.reg_write(UC_X86_REG_ESP,sp)
u.reg_write(UC_X86_REG_FPCW,0x27f)
u.emu_start(0x4014b5,0x4014c4,count=20000)
assert u.reg_read(UC_X86_REG_FPCW)==0xe3f
initial=u.context_save(); recording=True

cases=[]
def add(kind,name,**kw):
    value=dict(name=name,kind=kind,objects=[[0,0,0,f32(100.1),f32(200.3),0,0,10,20],[0,0,0,f32(-7),f32(9),0,0,0,0],[4,0,0,0,0,0,0,0,0],[4,0,0,0,0,0,0,0,0]],owners=[2,-1,-1,-1],update_refs=[0,1,-1,-1],data=[41,13,99],frame=20,fps=30,direction=13,radius_bits64=f64(5.1),delay=6,trace_result=1,allocate=True,mutation='')
    value.update(kw);cases.append(value);return value
for fps in (0,1,30,0x40000000,0x7fffffff,0x80000000,0xffffffff):
    for age in sorted({0,1,149,150,151,0x7fffffff,0x80000000,0xffffffff,*[(5*fps+d)&0xffffffff for d in (-1,0,1)]}):
        add(0,f'shower/age/{fps}/{age}',fps=fps,frame=(10+age)&0xffffffff)
for frame in (0,20,0x7fffffff,0x80000000,0xffffffff):
    for difference in (-1,0,1):
        c=add(0,f'shower/deadline/{frame}/{difference}',frame=frame)
        c['objects'][0][7]=frame;c['objects'][0][8]=(frame+difference)&0xffffffff
for direction in range(256):
    for radius in (5.1,11.7):add(0,f'shower/vector/{direction}/{radius}',direction=direction,radius_bits64=f64(radius))
for radius in (4,4.000000000000001,5,7.999999999999999,11.999999999999998,12):
    for direction in (0,31,64,129,255):add(0,f'shower/radius/{radius}/{direction}',direction=direction,radius_bits64=f64(radius))
for value in (0,1,0x100,0x101,-256,-1):add(0,f'shower/trace-low-byte/{value}',trace_result=value)
add(0,'shower/allocation-failure',allocate=False)
c=add(0,'shower/old-spawn-exact',direction=64,radius_bits64=f64(5));c['objects'][0][3:5]=[f32(100),f32(200)]
positions=[0,-0.0,0.1,-0.1,100.1,-100.1,102.9,-102.9,103.5,-103.5,1e9,-1e9,2147483648.,-2147483904.]
for n,x in enumerate(positions):
    for y in (0,200.3,-200.3):
        c=add(1,f'meteor/position/{n}/{y}');c['objects'][0][3:5]=[f32(x),f32(y)]
for raw in (0,0x80000000,1,f32(-1),f32(1),0x7f800000,0xff800000,0x7fc12345,0xffc12345):
    c=add(1,f'meteor/height/{raw:x}');c['objects'][0][5]=raw
for raw in (0,0x80000000,1,0x80000001,0x7f800000,0xff800000,0x7fc12345,0xffc12345):
    for kind in (0,1):
        c=add(kind,f'{specs[kind][0]}/position-bits/{raw:x}');c['objects'][0][3]=raw;c['objects'][0][4]=raw
for kind,services in [(0,('direction','radius','trace','new','create','raise','delay')),(1,('audio','scorch','quake','new','create','parent','units','walls','delete'))]:
    for service in services:
        for action in ('position','owner','owner-nil','update','spawn-update','damage','negative-damage','clock','created','flags','velocity'):
            add(kind,f'{specs[kind][0]}/mutation/{service}/{action}',mutation=service+':'+action)
    for service in (('trace',) if kind==0 else ('scorch','quake','units','walls')):
        add(kind,f'{specs[kind][0]}/mutation/{service}/point',mutation=service+':'+('trace-point' if kind==0 else 'point'))
c=add(1,'meteor/old-impact-exact');c['objects'][0][3:5]=[f32(100),f32(200)];c['data'][0]=37
add(1,'meteor/allocation-failure',allocate=False)

def prepare(c,cw,source=0):
    u.context_restore(initial);u.reg_write(UC_X86_REG_FPCW,cw)
    u.mem_write(heap,b'\xa5'*0x10000);u.mem_write(stack,bytes(0x10000))
    for n,p in enumerate(objects):
        for off,value in zip(offsets,c['objects'][n]):put(p+off,value)
        put(p+508,ref(c['owners'][n]));put(p+748,0 if c['update_refs'][n]<0 else data[c['update_refs'][n]])
    for p,value in zip(data,c['data']):put(p,value)
    put(frame_address,c['frame']);put(fps_address,c['fps'])
    u.mem_write(radius_value,struct.pack('<Q',c['radius_bits64']))
    put(sp,stop);put(sp+4,ref(source));u.reg_write(UC_X86_REG_ESP,sp)
    events.clear();faults.clear()
rows=[]
for mode,cw in [('actual-startup',0xe3f),('nearest-counterfactual',0x23f)]:
    for current in cases:
        prepare(current,cw);before=bytes(u.mem_read(heap,0x10000))
        u.emu_start(specs[current['kind']][1],stop,count=5000)
        assert not faults and u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4 and u.reg_read(UC_X86_REG_FPCW)==cw
        expected=bytearray(before)
        for p in objects:
            for off in offsets+(508,748):expected[p-heap+off:p-heap+off+4]=u.mem_read(p+off,4)
        for p in data:expected[p-heap:p-heap+4]=u.mem_read(p,4)
        assert bytes(expected)==bytes(u.mem_read(heap,0x10000))
        rows.append(dict(mode=mode,input=copy.deepcopy(current),output=state(),events=copy.deepcopy(events),nonoutput_bytes_preserved=True))
assert args.game.read_bytes()==blob
proof=dict(original_sha256=sha(blob),original_file_unchanged=True,original_startup_FPCW=0xe3f,contexts=len(cases),executions=len(rows),regions=regions,direction_table=table,entries=dict(entries),rows=rows,faults=[],executed_instructions=[dict(address=hex(p),bytes=b) for p,b in sorted(trace.items())],arithmetic_or_conversion_substitutions=0,scope='Complete unchanged MeteorShower, Meteor and actual419A70 conversion. Declared random/trace/allocation/world-effect services. Actual PC53/ToZero and nearest counterfactual; prepared callback mutations and complete heap canaries. No full-world/original-native execution or signaling-NaN claim.')
with args.proof.open('x') as f:json.dump(proof,f,separators=(',',':'))
fixture={k:v for k,v in proof.items() if k not in ('rows','executed_instructions')}
fixture['rows']=[r for r in rows if r['mode']=='actual-startup']
with args.fixture.open('x') as f:json.dump(fixture,f,separators=(',',':'))
if args.fault_proof:
    contexts=[]
    for kind in (0,1):
        c=copy.deepcopy(cases[0]);c.update(kind=kind,name='nil-source',mutation='');contexts.append((c,-1,True))
    for kind,name,z,next_frame,allocate,src_data,spawn_data,trace_result,fault in [
        (0,'waiting-nil-data',0,21,True,-1,1,1,False),
        (0,'blocked-nil-data',0,20,True,-1,1,0,False),
        (0,'allocation-failure-nil-data',0,20,False,-1,1,1,False),
        (0,'spawn-with-nil-source-data',0,20,True,-1,1,1,True),
        (0,'spawn-with-nil-spawn-data',0,20,True,0,-1,1,True),
        (1,'airborne-nil-data',f32(1),20,True,-1,1,1,False),
        (1,'impact-nil-data',0,20,True,-1,1,1,True),
        (1,'nan-nil-data',0x7fc00000,20,False,-1,1,1,True)]:
        c=copy.deepcopy(cases[0]);c.update(kind=kind,name=name,frame=20,fps=30,allocate=allocate,mutation='',trace_result=trace_result)
        c['objects'][0][5]=z;c['objects'][0][8]=next_frame;c['update_refs'][:2]=[src_data,spawn_data]
        contexts.append((c,0,fault))
    results=[]
    for current,source,fault in contexts:
        prepare(current,0xe3f,source);error=None
        try:u.emu_start(specs[current['kind']][1],stop,count=5000)
        except Exception as exc:error=str(exc)
        assert bool(faults)==fault,(current['name'],faults)
        results.append(dict(input=copy.deepcopy(current),source_ref=source,fault=fault,error=error,faults=list(faults),events=copy.deepcopy(events),output=state()))
    assert args.game.read_bytes()==blob
    with args.fault_proof.open('x') as f:json.dump(dict(original_sha256=sha(blob),rows=results,faults=sum(r['fault'] for r in results),safe_controls=sum(not r['fault'] for r in results)),f,indent=2)
print(json.dumps(dict(contexts=len(cases),executions=len(rows),entries=proof['entries'],faults=proof['faults'])))
