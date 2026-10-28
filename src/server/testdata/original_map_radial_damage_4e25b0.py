"""Execute unchanged radial-damage root/target/setter and both conversion helpers.

Rectangle enumeration, map trace, object damage and wall damage are declared
services. Their original callees are not replaced with game-result claims.
"""
import argparse, collections, copy, hashlib, json, struct
from pathlib import Path
from unicorn import Uc, UcError, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import *

p = argparse.ArgumentParser(description=__doc__)
for name in ('game', 'fixture', 'nearest', 'proof'):
    p.add_argument('--'+name, type=Path, required=True)
a = p.parse_args(); blob = a.game.read_bytes()
sha = lambda b: hashlib.sha256(b).hexdigest()
assert sha(blob) == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe = struct.unpack_from('<I', blob, 60)[0]; opt = pe+24
base = struct.unpack_from('<I', blob, opt+28)[0]; size = (struct.unpack_from('<I', blob, opt+56)[0]+4095)&~4095
tab = opt+struct.unpack_from('<H', blob, pe+20)[0]
u = Uc(UC_ARCH_X86, UC_MODE_32); u.mem_map(base, size)
for n in range(struct.unpack_from('<H', blob, pe+6)[0]):
    _, va, count, off = struct.unpack_from('<IIII', blob, tab+40*n+8)
    if count: u.mem_write(base+va, blob[off:off+count])
stack, heap, stop = 0x20000000, 0x21000000, 0x30000000
for v in (stack, heap): u.mem_map(v, 0x10000)
u.mem_map(stop,0x100000)
sp = stack+0xe000; objects = [heap+128+n*1024 for n in range(4)]
points = [heap+0x5000, heap+0x5100]; context = heap+0x6000; damage = stop+0x100; stub = stop+0x1000
flag = 0x5b8198
specs = [('setter',0x4e25a0,10),('root',0x4e25b0,328),('target',0x4e2700,208),('dword',0x419a70,17),('qword',0x566dcc,39)]
regions = [dict(name=n,address=hex(v),size=z,sha256=sha(bytes(u.mem_read(v,z)))) for n,v,z in specs]
get = lambda v: struct.unpack('<I',u.mem_read(v,4))[0]
put = lambda v,x: u.mem_write(v,struct.pack('<I',x&0xffffffff))
f32 = lambda v: struct.unpack('<I',struct.pack('<f',v))[0]
ref = lambda rows,n: 0 if n<0 else rows[n]
idx = lambda rows,v: -1 if v==0 else rows.index(v)
words = lambda v,n: [get(v+4*k) for k in range(n)]
trace={}; entries=collections.Counter(); events=[]; faults=[]; current={}; active=0; calls=collections.Counter(); recording=False
stub_cursor=stub;target_entries=[];prefix_receipts=[]

def state():
    ctx = None
    if active:
        ctx = dict(center=idx(points,get(active)),outer=get(active+4),inner=get(active+8),
                   damage=get(active+12),typ=get(active+16),source=idx(objects,get(active+20)),excluded=idx(objects,get(active+24)))
    return dict(objects=[dict(pos=words(v+56,2),flags=get(v+16),callback=get(v+716)!=0) for v in objects],
                points=[words(v,2) for v in points],flag=get(flag),context=ctx)

def event(name,args):
    events.append(dict(service=name,args=args,state=state()))

def mutate(stage, rect=0):
    n=calls[stage];calls[stage]+=1
    if stage!=current['stage'] or n!=current['at']: return
    action=current['action']
    if action=='center-values': put(points[0],f32(0.3));put(points[0]+4,f32(0.7))
    elif action=='target-position': put(objects[1]+56,f32(-99.1));put(objects[1]+60,f32(17.3))
    elif action=='switch-center': put(active,points[1])
    elif action=='nil-center': put(active,0)
    elif action=='outer': put(active+4,f32(20.3))
    elif action=='inner': put(active+8,f32(200.7))
    elif action=='damage': put(active+12,-117)
    elif action=='typ': put(active+16,9)
    elif action=='source': put(active+20,objects[3])
    elif action=='excluded': put(active+24,objects[1])
    elif action=='enable': put(flag,1)
    elif action=='disable': put(flag,0)
    elif action=='nil-callback': put(objects[1]+716,0)
    elif action=='rect':
        assert rect
        for k,v in enumerate([f32(47.9),f32(-47.9),f32(-23.9),f32(23.9)]):put(rect+4*k,v)
    else: raise AssertionError(action)

def ret(value=0):
    esp=u.reg_read(UC_X86_REG_ESP);u.reg_write(UC_X86_REG_EAX,value&0xffffffff)
    u.reg_write(UC_X86_REG_EIP,get(esp));u.reg_write(UC_X86_REG_ESP,esp+4)

def hook(_u,address,length,_):
    global active,stub_cursor
    if address==stop: u.emu_stop();return
    if not recording:return
    for name,start,count in specs:
        if start<=address<start+count:
            trace[address]=bytes(u.mem_read(address,length)).hex()
            if start==address:
                entries[name]+=1
                if name=='target':target_entries.append(idx(objects,get(u.reg_read(UC_X86_REG_ESP)+4)))
            return
    if stub<=address<stop+0x100000:return
    esp=u.reg_read(UC_X86_REG_ESP);arg=lambda n:get(esp+4*n)
    if address==0x517c10:
        rp,cb,active=arg(1),arg(2),arg(3);assert cb==0x4e2700
        event('each',words(rp,4));mutate('each',rp)
        # Unicorn may retain a translated block after overwriting service
        # stub bytes. Give every declared enumeration its own code address.
        location=stub_cursor;stub_cursor+=128;code=bytearray()
        for n in current['targets']:
            code+=b'\x68'+struct.pack('<I',active)+b'\x68'+struct.pack('<I',ref(objects,n))
            site=location+len(code);code+=b'\xe8'+struct.pack('<i',cb-site-5)+b'\x83\xc4\x08'
        code+=b'\xc3';assert len(code)<=128 and stub_cursor<stop+0x100000
        u.mem_write(location,bytes(code));u.reg_write(UC_X86_REG_EIP,location)
    elif address==0x535250:
        assert [arg(2),arg(3),arg(4)]==[0,0,1]
        event('trace',words(arg(1),4));mutate('trace');ret(current['trace_result'])
    elif address==damage:
        event('damage',[idx(objects,arg(1)),idx(objects,arg(2)),idx(objects,arg(3)),arg(4),arg(5)])
        mutate('damage');ret(0x87654321)
    elif address==0x534fc0:
        active=0
        event('walls',words(arg(1),4)+words(arg(2),2)+[arg(3),arg(4),arg(5),idx(objects,arg(6))])
        mutate('walls');ret(1)
    else:raise AssertionError((hex(address),current['name']))

def invalid(_u,access,address,length,value,_):
    faults.append(dict(address=address,size=length,instruction=u.reg_read(UC_X86_REG_EIP)));return False
u.hook_add(UC_HOOK_CODE,hook);u.hook_add(UC_HOOK_MEM_INVALID,invalid)
u.reg_write(UC_X86_REG_ESP,sp);u.reg_write(UC_X86_REG_FPCW,0x27f);u.emu_start(0x4014b5,0x4014c4,count=20000)
assert u.reg_read(UC_X86_REG_FPCW)==0xe3f
initial=u.context_save(); recording=True; cases=[]
def add(kind,name,**kw):
    c=dict(name=('root/' if kind==0 else 'target/')+name,kind=kind,
           points=[[f32(100.1),f32(50.7)],[f32(-30.3),f32(10.7)]],
           objects=[dict(pos=[f32(125.3),f32(52.7)],flags=0x12345678,callback=True) for _ in range(4)],
           center=0,outer=f32(96),inner=f32(10),damage=96,typ=7,source=0,excluded=-1,flag=1,
           targets=[1],target=1,trace_result=1,stage='',action='',at=0,nil_context=False)
    c.update(kw);cases.append(c);return c

for kind in (0,1):
    for x in (0.,0.3,9.999999,10.,10.000001,23.,47.9,95.99999,96.,96.00001,150.):
        for y in (0.,0.7,10.3):
            c=add(kind,f'distance/{x}/{y}');c['points'][0]=[0,0];c['objects'][1]['pos']=[f32(x),f32(y)]
    for d in (-2147483648,-96,-1,0,1,7,96,2147483647):
        for x in (0.,0.3,11.7,90.3):
            c=add(kind,f'damage/{d}/{x}',damage=d);c['points'][0]=[0,0];c['objects'][1]['pos']=[f32(x),f32(0.7)]
    for outer,inner in ((10,96),(0,0),(-10,10),(10,-10),(-96,-10),(96,96),(96,95.99999),(1e-30,0)):
        add(kind,f'radii/{outer}/{inner}',outer=f32(outer),inner=f32(inner))
    for field in range(6):
        for bits in (0,0x80000000,1,0x3e99999a,0x3f333333,0x3f800001,0x4affffff,0x4b000000,0x4f000000,0x5f000000,0x7f7fffff,0x7f800000,0xff800000,0x7fc12345,0xffc23456,0x7f812345):
            c=add(kind,f'word/{field}/{bits:08x}')
            if field<2:c['points'][0][field]=bits
            elif field<4:c['objects'][1]['pos'][field-2]=bits
            else:c['outer' if field==4 else 'inner']=bits
    for source in (-1,0,1):
        for excluded in (-1,0,1):
            for flagval in (0,1,0x100,-1):
                add(kind,f'identities/{source}/{excluded}/{flagval}',source=source,excluded=excluded,flag=flagval)
    for tr in (0,1,0x100,-1):add(kind,f'trace-result/{tr}',trace_result=tr)
    for stage in (('each','trace','damage') if kind==0 else ('trace','damage')):
        for action in ('center-values','target-position','switch-center','outer','inner','damage','typ','source','excluded','enable','disable'):
            add(kind,f'mutate/{stage}/{action}',stage=stage,action=action,targets=[1,0,2])
    for action in ('nil-center','nil-callback'):
        add(kind,'fault/trace/'+action,stage='trace',action=action,targets=[1,2])
    add(kind,'nil-center/safe-exclusion',center=-1,excluded=1)
    add(kind,'nil-center/fault',center=-1)
    c=add(kind,'nil-callback/fault');c['objects'][1]['callback']=False
add(0,'each-rect',stage='each',action='rect')
add(0,'empty',targets=[])
for bits in (f32(22.999998),f32(23),f32(23.000002),f32(-23.999998),f32(-24),0x4effffff,0x4f000000,0xcf000000):
    c=add(0,f'grid/{bits:08x}',outer=0,inner=0,targets=[]);c['points'][0]=[bits,bits]
add(1,'nil-context/fault',nil_context=True)
add(1,'nil-target/safe',target=-1,source=-1,flag=0,excluded=0)
add(1,'nil-target/excluded-safe',target=-1)
add(1,'nil-target/fault',target=-1,excluded=0)

def execute(c,cw):
    global current,active
    current=c;calls.clear();events.clear();faults.clear();target_entries.clear();u.context_restore(initial);u.reg_write(UC_X86_REG_FPCW,cw)
    u.mem_write(heap,b'\xa5'*0x10000);u.mem_write(stack,bytes(0x10000))
    for v,o in zip(objects,c['objects']):
        put(v+16,o['flags']);put(v+716,damage if o['callback'] else 0)
        for n,x in enumerate(o['pos']):put(v+56+4*n,x)
    for v,pp in zip(points,c['points']):
        for n,x in enumerate(pp):put(v+4*n,x)
    for n,v in enumerate([ref(points,c['center']),c['outer'],c['inner'],c['damage'],c['typ'],ref(objects,c['source']),ref(objects,c['excluded'])]):put(context+4*n,v)
    def invoke(address,args):
        put(sp,stop)
        for n,v in enumerate(args):put(sp+4+4*n,v)
        u.reg_write(UC_X86_REG_ESP,sp);u.emu_start(address,stop+1,count=20000)
    active=0;invoke(0x4e25a0,[c['flag']])
    before=bytes(u.mem_read(heap,0x10000));active=context if c['kind'] else 0
    if c['nil_context']:active=0
    try:
        if c['kind']==0:invoke(0x4e25b0,[ref(points,c['center']),c['outer'],c['inner'],c['damage'],c['typ'],ref(objects,c['source']),ref(objects,c['excluded'])])
        else:invoke(0x4e2700,[ref(objects,c['target']),0 if c['nil_context'] else context])
    except UcError:
        if not faults:raise
    expected_targets=c['targets'] if c['kind']==0 else [c['target']]
    assert target_entries==expected_targets[:len(target_entries)]
    if not faults:assert target_entries==expected_targets,(c['name'],target_entries,expected_targets)
    prefix_receipts.append(dict(name=c['name'],control_word=cw,targets=list(target_entries),fault=bool(faults)))
    if c['kind']==0:active=0
    result=dict(fault=bool(faults),state=state(),events=copy.deepcopy(events))
    after=bytearray(u.mem_read(heap,0x10000));expected=bytearray(before)
    for v in objects:
        for off,z in ((16,4),(56,8),(716,4)):expected[v-heap+off:v-heap+off+z]=after[v-heap+off:v-heap+off+z]
    for v in points:expected[v-heap:v-heap+8]=after[v-heap:v-heap+8]
    expected[context-heap:context-heap+28]=after[context-heap:context-heap+28]
    assert after==expected
    return result

actual=[];nearest=[];different=[];faultcount=0
for c in cases:
    x=execute(c,0xe3f);y=execute(c,0x23f)
    actual.append(dict(input=c,**x));nearest.append(dict(input=c,**y));faultcount+=x['fault']
    if x!=y:different.append(c['name'])
covered=sum(len(b)//2 for b in trace.values())
assert covered==sum(z for _,_,z in specs)==602,(covered,sum(z for _,_,z in specs))
for region,(_,v,z) in zip(regions,specs):assert sha(bytes(u.mem_read(v,z)))==region['sha256']
scope='Complete original root/target/setter/conversions; enumeration, trace, damage and walls are declared services. No transitive-world or campaign completion claim.'
for dest,rows,cw in ((a.fixture,actual,0xe3f),(a.nearest,nearest,0x23f)):
    with dest.open('x') as out:json.dump(dict(game_sha256=sha(blob),control_word=cw,regions=regions,scope=scope,rows=rows),out,indent=2);out.write('\n')
proof=dict(game_sha256=sha(blob),regions=regions,contexts=len(cases),executions=len(cases)*2,original_code_bytes=602,arithmetic_or_conversion_substitutions=0,original_code_unchanged=True,nonoutput_heap_and_canaries_preserved=True,actual_fault_contexts=faultcount,actual_vs_nearest_different=len(different),different_contexts=different,entries=dict(entries),unique_enumerator_code_addresses=True,callback_prefix_receipts=prefix_receipts,executed_instructions=[dict(address=hex(v),bytes=b) for v,b in sorted(trace.items())],scope=scope)
with a.proof.open('x') as out:json.dump(proof,out,indent=2);out.write('\n')
print(json.dumps({k:v for k,v in proof.items() if k not in ('executed_instructions','different_contexts','callback_prefix_receipts')}))
