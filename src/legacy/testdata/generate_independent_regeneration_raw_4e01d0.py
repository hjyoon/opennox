"""Choose raw Regeneration inputs before implementing its native C entry.

No expected return, branch, interval, effect or native result is calculated.
Armor lookup and HP adjustment are explicit external service boundaries.
"""
from pathlib import Path
import argparse,gzip,hashlib,json

parser=argparse.ArgumentParser(description=__doc__)
parser.add_argument('--output',type=Path,required=True)
options=parser.parse_args()
rows=[]
def add(name,**changes):
 row=dict(name=name,item='valid',owner='valid',health='valid',effect='valid',
          obj_class=0x01000000,owner_flags=0,last_damage=0,frame=36,fps=30,
          current_hp=99,maximum_hp=100,rate=60,armor_flags=0,
          service_return=0xa1b2c3d4,mutations=[])
 row.update(changes)
 row['seed']=len(rows)+1
 rows.append(row)

add('baseline')
for role in ['item','owner','health','effect']:
 for pointer in ['null','bad']:
  add('pointer/'+role+'/'+pointer,**{role:pointer})
add('pointer/item/last-holder',item='last-holder')
add('pointer/owner/last-health',owner='last-health')
add('pointer/health/last-maximum',health='last-maximum')
add('pointer/health/last-current-fault',health='last-current')
add('pointer/effect/last-rate',effect='last-rate')
add('pointer/owner/self-item',owner='item')
for flag in [0,1,0x20,0x8000,0x8020,0x100,0x80000000,0xffffffff]:
 add('flags/'+format(flag,'08x'),owner_flags=flag)
for frame,last,fps in [(0,0,0),(0,0,1),(29,0,30),(30,0,30),(31,0,30),
                      (36,0,30),(0,0xffffffff,1),(0,0xffffffe2,30),
                      (0xffffffff,0,30),(0x80000000,0,0x80000000),
                      (0x7fffffff,0,0x80000000),(0xffffffff,0,0xffffffff),
                      (0,1,30),(0x80000001,1,0x80000000)]:
 add('clock/'+format(frame,'08x')+'/'+format(last,'08x')+'/'+format(fps,'08x'),frame=frame,last_damage=last,fps=fps)
for current,maximum in [(0,0),(0,1),(1,1),(2,1),(0,100),(99,100),(100,100),(101,100),
                        (0x7fff,0x8000),(0x8000,0x8001),(0xfffe,0xffff),
                        (0xffff,0xffff),(0xffff,1)]:
 add('health/'+str(current)+'/'+str(maximum),current_hp=current,maximum_hp=maximum)
for rate in [0,1,2,3,4,29,30,31,59,60,61,450,1188,2700,7200,
             0x55555555,0x7fffffff,0x80000000,0x80000001,0xfffffffe,0xffffffff]:
 for armor in [False,True]:
  add('rate/'+format(rate,'08x')+('/robe' if armor else '/weapon'),rate=rate,
      obj_class=0x02000000 if armor else 0x01000000,armor_flags=0x4000 if armor else 0)
for obj_class in [0,1,0x01000000,0x02000000,0x03000000,0x04000000,0x80000000,0xffffffff]:
 for flags in [0,0x4000,0x8000,0x4001,0x7fffffff,0x80000000,0xffffffff]:
  add('armor/'+format(obj_class,'08x')+'/'+format(flags,'08x'),obj_class=obj_class,armor_flags=flags)
for frame in [0,1,2,3,17,18,19,35,36,37,0x7fffffff,0x80000000,0xffffffff]:
 add('modulo/'+format(frame,'08x'),frame=frame,last_damage=(frame-30)&0xffffffff)
for rate,fps,maximum in [(0x80000000,2,100),(0xffffffff,0xffffffff,1),
                         (0x7fffffff,2,0xffff),(0x55555555,3,100),
                         (0x10000,0x10000,100),(0xffffffff,30,0xffff)]:
 add('product/'+format(rate,'08x')+'/'+format(fps,'08x')+'/'+str(maximum),
     rate=rate,fps=fps,frame=0xffffffff,last_damage=0,current_hp=0,maximum_hp=maximum)

def before(function,occurrence,role,field,value):
 return dict(when='before-call',function=function,occurrence=occurrence,
             role=role,field=field,value=value)
def after(role,field,occurrence,value,target_role=None,target_field=None):
 return dict(when='after-read',role=role,field=field,occurrence=occurrence,
             target_role=target_role or role,target_field=target_field or field,value=value)
for name,mutation in [
 ('cached-holder',before('maximum-hp',1,'item','holder','bad')),
 ('cached-holder-at-adjust',before('adjust-hp',1,'item','holder','null')),
 ('nil-health-at-first-maximum',before('maximum-hp',1,'owner','health','null')),
 ('bad-health-at-first-maximum',before('maximum-hp',1,'owner','health','bad')),
 ('nil-health-at-current',before('current-hp',1,'owner','health','null')),
 ('bad-health-at-current',before('current-hp',1,'owner','health','bad')),
 ('nil-health-at-second-maximum',before('maximum-hp',2,'owner','health','null')),
 ('bad-health-at-second-maximum',before('maximum-hp',2,'owner','health','bad')),
 ('zero-reloaded-maximum',before('maximum-hp',2,'health','maximum',0)),
 ('changed-reloaded-maximum',before('maximum-hp',2,'health','maximum',50)),
 ('cached-first-frame',after('world','frame',1,37)),
 ('cached-first-fps',after('world','fps',1,1)),
 ('fresh-second-frame',after('world','fps',2,37,'world','frame')),
 ('cached-second-fps',after('world','fps',2,0)),
 ('cached-last-damage',after('owner','last-damage',1,36)),
 ('cached-flags',after('owner','flags',1,0x8020)),
 ('cached-class',after('item','class',1,0)),
 ('cached-rate',after('effect','rate',1,0)),
 ('cached-first-maximum',after('health','maximum',1,50)),
 ('cached-current',after('health','current',1,100)),
]:
 add('mutation/'+name,mutations=[mutation],obj_class=0x02000000,armor_flags=0x4000)
add('mutation/armor-cached-rate',obj_class=0x02000000,armor_flags=0x4000,
    mutations=[before('armor-flags',1,'effect','rate',0)])
add('mutation/armor-live-maximum',obj_class=0x02000000,armor_flags=0x4000,
    mutations=[before('armor-flags',1,'health','maximum',50)])

# Early gates must protect invalid downstream effect/health fields.
for name,changes in [('nil-item',dict(item='null')),('nil-owner',dict(owner='null')),
                      ('nil-health',dict(health='null')),('recent',dict(frame=1)),
                      ('dead',dict(owner_flags=0x20)),('full',dict(current_hp=100))]:
 add('protected-effect/'+name,effect='bad',**changes)
for flag in [0x20,0x8000]:
 add('protected-health/'+format(flag,'08x'),health='bad',owner_flags=flag)
add('protected-health/recent',health='bad',frame=1)

assert len({row['name'] for row in rows})==len(rows)
data=json.dumps(rows,sort_keys=True,separators=(',',':')).encode()+b'\n'
encoded=gzip.compress(data,mtime=0)
with options.output.open('xb') as stream:stream.write(encoded)
print(json.dumps(dict(cases=len(rows),encoded_sha256=hashlib.sha256(encoded).hexdigest(),decoded_sha256=hashlib.sha256(data).hexdigest(),expected_results_calculated=False)),flush=True)
