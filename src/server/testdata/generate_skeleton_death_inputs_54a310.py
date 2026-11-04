"""Independent raw Skeleton/debris/drop inputs; no port or expected output is read."""
from pathlib import Path
import argparse, copy, gzip, hashlib, json, struct
ap=argparse.ArgumentParser(description=__doc__);ap.add_argument('--output',type=Path,required=True);a=ap.parse_args()
rows=[]
bits=lambda v:struct.unpack('<Q',struct.pack('<d',v))[0]
base=dict(entry='54a4c0',unit=0,type=0,mods=[1,2,-1,-1],charge=0,coop=1,cache=0,divisor=2,fps=30,roll=35,parts=2,seconds=13,float_bits=bits(1.25),new={'Skull':2,'ArmBone':3,'LegBone':4,'Sword':5,'WoodenShield':5,'SteelShield':5,'TestItem':5},classes=[0,0,0,0,0,0x1000000],subclasses=[0,0,0,0,0,0x82],uses=[6,6,6,6,6,6],script={})
def add(name,**changes):
 row=copy.deepcopy(base);row.update(changes);row['name']=name;rows.append(row)
for entry in ['54a310','54a750']:
 for roll in [-1,0,20,21,50,51,100,101]:
  for coop in [0,1]:add(f'{entry}-roll-{roll}-coop-{coop}',entry=entry,roll=roll,coop=coop)
for coop in [0,1]:
 for parts in [-1,0,1,2,3]:add(f'debris-coop-{coop}-parts-{parts}',coop=coop,parts=parts)
for name in ['Skull','ArmBone','LegBone']:
 new=base['new'].copy();new[name]=-1;add('allocation-failure-'+name,new=new)
for role in [0,1,3,5,-2]:
 new=base['new'].copy();new['Skull']=role;add(f'skull-alias-{role}',new=new)
for unit in [-1,-2]:
 new=base['new'].copy();new['Skull']=-1
 add(f'bad-unit-{unit}-constructor-fails',unit=unit,new=new)
 add(f'bad-unit-{unit}-world-origin-read',unit=unit)
for cache,divisor in [(1,2),(0,1),(0,0),(1,0),(1,-2),(0,-1),(0x7fffffff,-1),(0xffffffff,2)]:
 add(f'cache-{cache:08x}-divisor-{divisor}',cache=cache,divisor=divisor,parts=1)
for fps in [0,1,0x7fffffff,0x80000000,0xffffffff]:add(f'fps-{fps:08x}',fps=fps,seconds=0x7fffffff)
for value in [0,0x8000000000000000,1,0x0010000000000000,0x3ff0000000000001,0xbff0000000000001,0x3fefffffffffffff,0x7fefffffffffffff,0xffefffffffffffff,0x7ff0000000000000,0xfff0000000000000,0x7ff8000000001234]:
 add(f'float-{value:016x}',float_bits=value,parts=1)
for kind in ['audio','pointfx','new','point','create','raise','float','force','flags','int','decay']:
 add('move-source-at-'+kind,script={kind+':0':{'position':[0,0x447a0000,0xc47a0000]}})
 add('reenter-at-'+kind,script={kind+':0':{'reenter':['54a4c0',1]}})
for kind in ['create','raise','float','force','flags','int','decay']:
 add('live-cache-and-divisor-at-'+kind,script={kind+':1':{'cache':1,'divisor':-2,'fps':0xffffffff,'coop':0}})
# Raw drop admission, live class, all four nil-name lookups, charge and audio branches.
for cls in [0,0x1000,0x1000000,0x2000000,0x3000000,0x10000000,0x4000000]:
 for mods in [[-1,-1,-1,-1],[1,-1,-1,-1],[-1,2,3,4]]:
  classes=base['classes'].copy();classes[5]=cls
  add(f'drop-class-{cls:08x}-mods-'+','.join(map(str,mods)),entry='54a390',classes=classes,mods=mods,charge=0x1234)
for coop in [0,1]:
 for unit in [0,-1,-2]:
  for allocation in [-1,5]:
   new=base['new'].copy();new['TestItem']=allocation
   add(f'drop-admission-coop-{coop}-unit-{unit}-new-{allocation}',entry='54a390',type=0,coop=coop,unit=unit,new=new)
for charge in [0,1,-1,256,0x7fffffff]:add(f'drop-charge-{charge}',entry='54a390',charge=charge)
for use in [-1,-2,7]:
 uses=base['uses'].copy();uses[5]=use
 add(f'drop-use-{use}',entry='54a390',charge=0x1234,uses=uses)
for sub in [0,2,0x80,0x100]:
 subclasses=base['subclasses'].copy();subclasses[5]=sub
 add(f'drop-subclass-{sub}',entry='54a390',charge=4,subclasses=subclasses)
for kind in ['new','point','create','modid','moddesc','apply']:
 add('drop-live-class-at-'+kind,entry='54a390',charge=0x1234,script={kind+':0':{'class':[5,0x3000000],'use':[5,7]}})
 add('drop-reenter-at-'+kind,entry='54a390',charge=0x1234,script={kind+':0':{'reenter':['54a390',1]}})
add('drop-all-descriptors-nil',entry='54a390',script={f'moddesc:{n}':{'return_nil':True} for n in range(4)})
add('drop-invalid-type-not-read-with-flags-off',entry='54a390',type=-2,coop=0,unit=-2)
add('drop-invalid-mods-not-read-before-create-fails',entry='54a390',mods=[-2]*4,new=dict(base['new'],TestItem=-1),unit=-2)
add('live-cache-signed-divide-overflow',script={'decay:1':{'cache':0x7fffffff,'divisor':-1}},parts=1)
assert len({r['name'] for r in rows})==len(rows)
raw=json.dumps(rows,sort_keys=True,separators=(',',':')).encode()+b'\n';packed=bytearray(gzip.compress(raw,mtime=0));packed[9]=0x13
with a.output.open('xb')as out:out.write(packed)
print(json.dumps(dict(cases=len(rows),sha256=hashlib.sha256(packed).hexdigest(),decoded_sha256=hashlib.sha256(raw).hexdigest(),native_source_used=False)))
