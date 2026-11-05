"""Inventory/owner graphs chosen before the raw native checker implementation."""
from pathlib import Path
import copy
import gzip
import hashlib
import itertools
import json

stage=Path(__file__).resolve().parent
cases=[]
strengths=[0,1,2,255,65536,0x7fffffff,0x80000000,0x80000001,0xffff0000,0xfffffffe,0xffffffff]
def add(name,**changes):
    row=dict(name=name,root=0,target=1,first=2,next=[-1,-1,-1],
             flags=[0x100,0x100,0x100],classes=[0x1000,0x1000,0x1000],
             data=[0,1,2],slots=[[0,-1],[-1,-1],[-1,-1]],
             functions=['inversion','inversion','grip','alien'],strengths=[0,1,0,1],
             owners=[-1,-1,-1,-1,-1,-1,-1],owner_classes=[2,2,2,2,2,2,2])
    row.update(copy.deepcopy(changes));cases.append(row)

add('nil-root',root=-1,target=-2)
add('bad-root',root=-2)
add('empty',first=-1,target=-2)
add('bad-first',first=-2)
for flag in [0,1,0x80,0x200,0x80000000]:
    add(f'unequipped-{flag:08x}',flags=[flag,0,0],data=[-2,-2,-2],target=-2)
for cls in [0,1,4,0x800,0x80000000]:
    add(f'class-reject-{cls:08x}',classes=[cls,0,0],data=[-2,-2,-2],target=-2)
for cls in [0x1000,0x1000000,0x2000000,0x10000000,0x13001000,0xffffffff]:
    for flag in [0x100,0xffffffff]:
        add(f'class-admit-{cls:08x}-flags-{flag:08x}',classes=[cls,0,0],flags=[flag,0,0],strengths=[1,0,0,0],target=-1)
for left,right in itertools.product(strengths,repeat=2):
    add(f'signed-{left:08x}-{right:08x}',slots=[[0,1],[-1,-1],[-1,-1]],strengths=[left,right,0,0])
for slot in [0,1]:
    for value in strengths:
        pair=[-1,-1];pair[slot]=0
        add(f'only-slot-{slot+2}-{value:08x}',slots=[pair,[-1,-1],[-1,-1]],strengths=[value,0,0,0])
for fn in ['none','grip','alien']:
    add('wrong-callback-'+fn,functions=[fn,'inversion','grip','alien'],strengths=[1,1,0,0],target=-2)
add('null-modifiers',slots=[[-1,-1],[-1,-1],[-1,-1]],target=-2)
for which in [0,1]:
    pair=[-1,-1];pair[which]=-2
    add('bad-modifier-'+str(which+2),slots=[pair,[-1,-1],[-1,-1]])
for pointer in [-1,-2]:
    add('bad-data-'+str(pointer),data=[pointer,1,2])
add('early-success-skips-bad-slot',slots=[[0,-2],[-1,-1],[-1,-1]],strengths=[1,0,0,0])
add('negative-continues-bad-slot',slots=[[0,-2],[-1,-1],[-1,-1]],strengths=[0xffffffff,0,0,0])
add('weak-continues-bad-next',next=[-2,-1,-1])
add('success-skips-bad-next',next=[-2,-1,-1],strengths=[1,0,0,0])
for value in strengths:
    add(f'shared-modifier-{value:08x}',slots=[[0,0],[-1,-1],[-1,-1]],strengths=[value,0,0,0])
add('skip-flags-follow-next',next=[3,4,-1],flags=[0,0x100,0x100],classes=[0x1000,4,0x1000],slots=[[-2,-2],[-2,-2],[1,-1]])
add('late-item-success',next=[3,4,-1],slots=[[0,-1],[0,-1],[1,-1]])
add('all-items-weak',next=[3,4,-1],slots=[[0,0],[0,0],[0,0]])
add('shared-data-live-item-order',next=[3,-1,-1],data=[0,0,2],slots=[[0,0],[-2,-2],[-1,-1]])
add('bad-target-with-match',target=-2,strengths=[1,0,0,0])
add('bad-target-before-weak-strength',target=-2)
add('nil-target-with-match',target=-1,strengths=[1,0,0,0])
add('nil-target-with-negative',target=-1,strengths=[0xffffffff,0,0,0])
add('terminal-target-skips-class',owner_classes=[2,0xffffffff,2,2,2,2,2],strengths=[1,0,0,0])
owners=[-1]*7;owners[1]=-2
add('player-returns-self-with-bad-owner',owners=owners,owner_classes=[2,4,2,2,2,2,2],strengths=[1,0,0,0])
add('nonplayer-follows-bad-owner',owners=owners,strengths=[1,0,0,0])
owners=[-1]*7;owners[1]=5;owners[5]=6
add('two-owner-links-to-terminal',owners=owners,strengths=[1,0,0,0])
classes=[2]*7;classes[5]=0x10000004
add('intermediate-player-with-owner',owners=owners,owner_classes=classes,strengths=[1,0,0,0])
classes=[2]*7;classes[1]=4
add('target-player-with-owner',owners=owners,owner_classes=classes,strengths=[1,0,0,0])
owners=[-1]*7;owners[1]=5;owners[5]=-2
add('owner-link-read-before-owner-class',owners=owners,strengths=[1,0,0,0])
for number,row in enumerate(cases):
    row['seed']=number+1
assert len({row['name']for row in cases})==len(cases)
encoded=json.dumps(cases,sort_keys=True,separators=(',',':')).encode()+b'\n'
import argparse
parser=argparse.ArgumentParser(description=__doc__)
parser.add_argument('--output',type=Path,required=True)
target=parser.parse_args().output
with target.open('xb')as out:out.write(gzip.compress(encoded,mtime=0))
print(json.dumps(dict(cases=len(cases),sha256=hashlib.sha256(target.read_bytes()).hexdigest(),decoded_sha256=hashlib.sha256(encoded).hexdigest())))
