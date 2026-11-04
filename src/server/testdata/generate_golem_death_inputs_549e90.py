"""Independent raw MechGolem/Golem inputs; no native implementation or output is read."""
from pathlib import Path
import argparse, copy, gzip, hashlib, json, struct
ap=argparse.ArgumentParser(description=__doc__)
ap.add_argument('--output',type=Path,required=True)
a=ap.parse_args()
rows=[]
bits=lambda value:struct.unpack('<Q',struct.pack('<d',value))[0]
base=dict(entry='549fa0',unit=0,coop=1,cache=0,divisor=12,fps=30,parts=2,seconds=13,
          float_bits=bits(1.25),mech_stop=8,new={name:2 for name in
          ['MechGear','MechHead','MechArm','MechLeg','BigRock','MediumRock','SmallRock']},script={})
def add(name,**changes):
    row=copy.deepcopy(base);row.update(changes);row['name']=name;rows.append(row)
for entry in ['549e90','549fa0']:
    for coop in [0,1,2]:add(f'{entry}-coop-{coop}',entry=entry,coop=coop)
    for unit in [-1,-2]:
        add(f'{entry}-unit-{unit}-origin-read',entry=entry,unit=unit)
        add(f'{entry}-unit-{unit}-allocation-fails',entry=entry,unit=unit,new={})
        add(f'{entry}-unit-{unit}-empty',entry=entry,unit=unit,mech_stop=0,parts=0)
    for role in [0,1,3,5,-2]:
        add(f'{entry}-constructor-alias-{role}',entry=entry,new={name:role for name in base['new']})
    for occurrence in [0,1,3]:
        add(f'{entry}-allocation-fails-at-{occurrence}',entry=entry,script={f'new:{occurrence}':{'return_nil':True}})
    for fps in [0,1,0x7fffffff,0x80000000,0xffffffff]:
        add(f'{entry}-fps-{fps:08x}',entry=entry,fps=fps,seconds=0x7fffffff)
    for value in [0,0x8000000000000000,1,0x0010000000000000,0x3ff0000000000001,
                  0xbff0000000000001,0x3fefffffffffffff,0x7fefffffffffffff,
                  0xffefffffffffffff,0x7ff0000000000000,0xfff0000000000000,0x7ff8000000001234]:
        add(f'{entry}-float-{value:016x}',entry=entry,float_bits=value)
    for kind in ['audio','pointfx','new','point','create','raise','float','decay']+(['flags','int','force'] if entry=='549fa0' else ['int']):
        add(f'{entry}-move-source-at-{kind}',entry=entry,script={kind+':0':{'position':[0,0x447a0000,0xc47a0000]}})
        add(f'{entry}-reenter-at-{kind}',entry=entry,script={kind+':0':{'reenter':['549fa0' if entry=='549e90' else '549e90',1]}})
    add(f'{entry}-same-entry-reentry',entry=entry,script={'decay:0':{'reenter':[entry,1]}})
    add(f'{entry}-live-part-flags-before-vertical-store',entry=entry,script={'float:1':{'flags':[2,0x01020304]}})
for stop in range(9):add(f'mech-stop-{stop}',entry='549e90',mech_stop=stop)
for parts in [-2,-1,0,1,2,13,31]:add(f'golem-part-count-{parts}',parts=parts)
for cache,divisor in [(0,1),(1,12),(11,12),(0,0),(0,-1),(0,-2),
                      (12,12),(0xffffffff,12),(0x7fffffff,-1),(0x80000000,12),(0x10000000,12)]:
    add(f'golem-cache-{cache:08x}-divisor-{divisor}',cache=cache,divisor=divisor,parts=1)
for value in [0,128,255]:add(f'golem-unsigned-mass-{value}',script={'new:0':{'mass':[0,value]}})
for kind in ['audio','pointfx','flags','int','new','point','create','raise','float','force','decay']:
    add(f'golem-live-index-divisor-at-{kind}',script={kind+':0':{'cache':11,'divisor':-2,'fps':0xffffffff,'coop':0}})
for index in [1,11,12,0xffffffff,0x10000000]:
    add(f'golem-live-mass-index-{index:08x}',script={'new:0':{'cache':index}})
add('golem-signed-divide-overflow',script={'decay:0':{'cache':0x7fffffff,'divisor':-1}})
add('golem-live-table-next-slot',parts=2,script={'decay:0':{'rock':[1,4]}})
add('golem-live-mass-after-type-is-cached',script={'new:0':{'mass':[0,255],'cache':1}})
add('mech-live-next-name',entry='549e90',script={'decay:0':{'mech':[1,3]}})
add('mech-live-null-next-slot',entry='549e90',script={'decay:0':{'mech':[1,-1]}})
add('mech-live-first-slot-no-reload',entry='549e90',script={'new:0':{'mech':[0,-1]}})
assert len({row['name'] for row in rows})==len(rows)
raw=json.dumps(rows,sort_keys=True,separators=(',',':')).encode()+b'\n'
packed=bytearray(gzip.compress(raw,compresslevel=9,mtime=0));packed[9]=0x13
with a.output.open('xb') as out:out.write(packed)
print(json.dumps(dict(cases=len(rows),sha256=hashlib.sha256(packed).hexdigest(),decoded_sha256=hashlib.sha256(raw).hexdigest(),native_source_used=False)))
