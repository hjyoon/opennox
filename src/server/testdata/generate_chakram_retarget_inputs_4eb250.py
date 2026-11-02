"""Deterministic input plans, independent of native or original outputs."""
from pathlib import Path
import copy, gzip, hashlib, json, math, random, struct

p = Path(__file__).resolve().parent
f32 = lambda v: struct.unpack('<I', struct.pack('<f', v))[0]
rows = []

def case(name, entry='driver', **kw):
    objects = [dict(cls=4, flags=0, pos=[f32(10),f32(20)], speed=f32(11),
                    vel=[0x40c00000,0x40a00000], owner=-1, update=0) for _ in range(6)]
    objects[0]['owner'] = 1
    for i in range(1,6): objects[i]['pos'] = [f32(10+i*3), f32(20+i*4)]
    out = dict(name=name, entry=entry, source=0, candidate=3, point=0, point_offset=56,
               objects=objects, updates=[dict(last=2,state=0xa5),dict(last=1,state=0x5a),dict(last=-1,state=0xff)],
               scratch=dict(source=5,owner=4,last=3,selected=2,best=0x4b189680),
               enums=[dict(candidates=[1,2,3,4,5])], buffs=[], maps=[])
    out.update(kw)
    rows.append(out)
    return out

for route in [[],[0],[1],[2],[3],[5,4,3],[3,3,4],[3,4,5]]:
    case('driver/route/'+','.join(map(str,route)),enums=[dict(candidates=route)])
for bit in range(32):
    for field in ['cls','flags']:
        c=case(f'callback/{field}/bit{bit}',entry='callback')
        c['objects'][3][field]=1<<bit
for value in [0,2,4,0x20000,0x20006,0xffffffff,0x40000000,0x3f800000]:
    c=case(f'callback/class/{value:08x}',entry='callback')
    c['objects'][3]['cls']=value
for value in [0,0x20,0x8000,0x8020,0xffffffff,1,0x40000000]:
    c=case(f'callback/flags/{value:08x}',entry='callback')
    c['objects'][3]['flags']=value
for gate in ['buff','map','owner','last','excluded-after-buff','source-after-buff','point-after-map','live-class-after-buff']:
    c=case('driver/gate/'+gate,enums=[dict(candidates=[3,4])])
    if gate=='buff': c['buffs']=[dict(value=1),dict(value=0)]
    elif gate=='map': c['maps']=[dict(value=0),dict(value=1)]
    elif gate=='owner': c['objects'][0]['owner']=3
    elif gate=='last': c['updates'][0]['last']=3
    elif gate=='excluded-after-buff': c['buffs']=[dict(value=0,mutation=dict(scratch=dict(owner=3)))]
    elif gate=='source-after-buff': c['buffs']=[dict(value=0,mutation=dict(scratch=dict(source=5)))]
    elif gate=='point-after-map': c['maps']=[dict(value=1,mutation=dict(objects=[dict(id=0,pos=[f32(50),f32(-60)])]))]
    else: c['buffs']=[dict(value=0,mutation=dict(objects=[dict(id=3,cls=0,flags=0x8020)]))]
for which in ['update','last','owner','source','selected','point']:
    c=case('driver/live/'+which,enums=[dict(candidates=[3])])
    mutation = dict(objects=[dict(id=0,update=1)]) if which=='update' else \
        dict(updates=[dict(id=0,last=3)]) if which=='last' else \
        dict(objects=[dict(id=0,owner=3)]) if which=='owner' else \
        dict(scratch=dict(source=5)) if which=='source' else \
        dict(scratch=dict(selected=4)) if which=='selected' else \
        dict(objects=[dict(id=0,pos=[f32(-30),f32(70)])])
    c['enums'][0]['after']=mutation
for gate in ['buff','map','enum']:
    c=case('driver/reentrant/'+gate,enums=[dict(candidates=[3,4]),dict(candidates=[5])])
    nested=dict(source=5)
    if gate=='buff': c['buffs']=[dict(value=0,reenter=nested)]
    elif gate=='map': c['maps']=[dict(value=1,reenter=nested)]
    else: c['enums'][0]['reenter']=nested

positions=[(0,0),(3,4),(400,0),(-400,0),(0,400),(400,0.00001),(399.99997,0),
           (400.00003,0),(float('inf'),0),(-float('inf'),0),(float('nan'),0),
           (0,float('nan')),(1e-40,-1e-40),(1e20,-1e20),(2**24+2,-2**24),
           (123.4567,-789.125),(-0.0,0.0)]
for index,(x,y) in enumerate(positions):
    for entry in ['driver','callback']:
        c=case(f'{entry}/numeric/{index}',entry=entry,enums=[dict(candidates=[3])])
        c['objects'][0]['pos']=[f32(0),f32(0)]
        c['objects'][3]['pos']=[f32(x),f32(y)]
        if entry=='callback': c['scratch'].update(source=0,owner=1,last=2,selected=-1)
for bits in [0,0x80000000,0x3f800000,0x481c4000,0x481c3fff,0x481c4001,0x7f800000,0xff800000,0x7fc12345,0x7f812345]:
    c=case(f'callback/best/{bits:08x}',entry='callback')
    c['scratch'].update(source=0,owner=1,last=2,selected=-1,best=bits)
for bits in [0,0x80000000,0x3dcccccd,0x3f800001,0x7f7fffff,0x00800000,1,0x7f800000,0xff800000,0x7fc01234]:
    c=case(f'driver/speed/{bits:08x}',enums=[dict(candidates=[3])])
    c['objects'][0]['speed']=bits
for point,offset in [(3,56),(0,80),(3,80),(5,56)]:
    c=case(f'callback/point/{point}/{offset}',entry='callback',point=point,point_offset=offset)
    c['scratch'].update(source=0,owner=1,last=2,selected=-1)

# Fault plans distinguish original dereferences from world-service admission.
for bad in [-1,-2]:
    case(f'fault/driver/source/{bad}',source=bad)
    c=case(f'fault/driver/update/{bad}');c['objects'][0]['update']=bad
    case(f'fault/callback/candidate/{bad}',entry='callback',candidate=bad)
    c=case(f'fault/callback/point/{bad}',entry='callback',point=bad)
    c['scratch'].update(source=0,owner=1,last=2,selected=-1)
    c=case(f'fault/driver/selected/{bad}',enums=[dict(candidates=[],after=dict(scratch=dict(selected=bad)))])
    c=case(f'fault/driver/late-update/{bad}',enums=[dict(candidates=[3],after=dict(objects=[dict(id=0,update=bad)]))])
    c=case(f'callback/ignored-point/{bad}',entry='callback',point=bad);c['objects'][3]['cls']=0

rng=random.Random(0x4eb250)
for i in range(96):
    c=case(f'driver/finite-rounding/{i}',enums=[dict(candidates=[3,4,5])])
    x,y=[rng.uniform(-2000,2000)for _ in range(2)]
    c['objects'][0]['pos']=[f32(x),f32(y)]
    c['objects'][0]['speed']=f32(rng.uniform(-50,50))
    for j in [3,4,5]:c['objects'][j]['pos']=[f32(x+rng.uniform(-400,400)),f32(y+rng.uniform(-400,400))]
names=[x['name']for x in rows];assert len(names)==len(set(names))
raw=json.dumps(dict(schema=1,rows=rows),separators=(',',':')).encode()
data=gzip.compress(raw,mtime=0)
out=p/'chakram_retarget_independent_inputs_4eb250.json.gz';assert not out.exists()
out.write_bytes(data)
print(json.dumps(dict(rows=len(rows),gzip_SHA256=hashlib.sha256(data).hexdigest(),decoded_SHA256=hashlib.sha256(raw).hexdigest(),outputs_consulted=False)))
