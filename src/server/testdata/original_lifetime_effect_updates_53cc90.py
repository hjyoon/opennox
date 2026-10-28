"""Execute six unchanged GAME.EXE lifetime/effect updates with declared world services."""
from pathlib import Path
import argparse,collections,copy,hashlib,json,struct
from unicorn import Uc,UC_ARCH_X86,UC_MODE_32,UC_HOOK_CODE,UC_HOOK_MEM_INVALID
from unicorn.x86_const import *
parser=argparse.ArgumentParser(description=__doc__)
parser.add_argument('--game',type=Path,required=True)
parser.add_argument('--fixture',type=Path,required=True)
parser.add_argument('--proof',type=Path,required=True)
parser.add_argument('--fault-proof',type=Path)
args=parser.parse_args();original=args.game;blob=original.read_bytes()
sha=lambda b:hashlib.sha256(b).hexdigest()
assert sha(blob)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,60)[0];opt=pe+24;base=struct.unpack_from('<I',blob,opt+28)[0]
size=(struct.unpack_from('<I',blob,opt+56)[0]+4095)&~4095;tab=opt+struct.unpack_from('<H',blob,pe+20)[0]
u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,size)
for n in range(struct.unpack_from('<H',blob,pe+6)[0]):
 _,va,count,offset=struct.unpack_from('<IIII',blob,tab+40*n+8)
 if count:u.mem_write(base+va,blob[offset:offset+count])
stack,heap,stop=0x20000000,0x21000000,0x30000000
for a in (stack,heap,stop):u.mem_map(a,0x10000)
sp=stack+0xe000
get=lambda a:struct.unpack('<I',u.mem_read(a,4))[0]
put=lambda a,v:u.mem_write(a,struct.pack('<I',v&0xffffffff))
bits=lambda x:struct.unpack('<I',struct.pack('<f',x))[0]
objects=[heap+128+1024*n for n in range(4)]
updates=[heap+8192,heap+9216];players=[heap+12000,heap+20000]
offsets=(8,16,56,60,72,76,104,128,136,156,160,176)
frame_address,fps_address=0x84ea04,0x85b3fc
specs=[('self',0x53cc90,29),('fragment',0x53d220,74),('moonglow',0x53d270,192),('telekinesis',0x53d330,202),('fist',0x53d400,264),('flame',0x53d510,136)]
regions=[dict(name=n,address=hex(a),size=z,sha256=sha(bytes(u.mem_read(a,z)))) for n,a,z in specs]
ref=lambda n:0 if n<0 else objects[n]
objref=lambda p:-1 if not p else objects.index(p)
udref=lambda p:-1 if not p else updates.index(p)
plref=lambda p:-1 if not p else players.index(p)
def state():
 return dict(objects=[[get(p+o) for o in offsets] for p in objects],owners=[objref(get(p+508)) for p in objects],update_refs=[udref(get(p+748)) for p in objects],player_refs=[plref(get(p+276)) for p in updates],cursors=[[get(p+2284),get(p+2288)] for p in players],frame=get(frame_address),fps=get(fps_address))
trace={};entries=collections.Counter();events=[];faults=[];recording=False;current={};fx_count=0
def ret(value=0):
 esp=u.reg_read(UC_X86_REG_ESP);u.reg_write(UC_X86_REG_EAX,value&0xffffffff);u.reg_write(UC_X86_REG_EIP,get(esp));u.reg_write(UC_X86_REG_ESP,esp+4)
def mutate(service,point=0):
 global fx_count
 if service=='fx':fx_count+=1
 key=current['mutation']
 if not key or not key.startswith(service+':'):return
 action=key.split(':',1)[1]
 if service=='fx' and fx_count!=1:return
 if action=='owner':put(objects[0]+508,objects[2])
 elif action=='owner-nil':put(objects[0]+508,0)
 elif action=='update':put(objects[1]+748,updates[1])
 elif action=='player':put(updates[0]+276,players[1])
 elif action=='cursor':put(players[0]+2284,0x7fffffff);put(players[0]+2288,0x80000001)
 elif action=='position':put(objects[0]+56,bits(500.1));put(objects[0]+60,bits(-600.3))
 elif action=='radius':put(objects[0]+176,bits(53.7))
 elif action=='height':put(objects[0]+104,bits(201))
 elif action=='unmark':put(objects[0]+16,0)
 elif action=='deadline':put(objects[0]+136,0)
 elif action=='created':put(objects[0]+128,get(frame_address))
 elif action=='clock':put(frame_address,0xffffffff);put(fps_address,0)
 elif action=='trace-point':put(point+8,bits(17.3));put(point+12,bits(-91.7))
 elif action=='pos-pointer':put(point,bits(611.3));put(point+4,bits(-817.7))
 elif action=='prev-position':put(objects[0]+72,bits(99));put(objects[0]+76,bits(88))
 else:raise AssertionError(key)
def event(name,values):events.append(dict(service=name,args=values,**state()))
def point(a):return [get(a),get(a+4)]
def hook(_u,a,n,_):
 if a in (stop,0x4014c4):u.emu_stop();return
 if not recording:return
 for name,lo,z in specs:
  if lo<=a<lo+z:
   trace[a]=bytes(u.mem_read(a,n)).hex()
   if a==lo:entries[name]+=1
   return
 esp=u.reg_read(UC_X86_REG_ESP);arg=lambda i:get(esp+4*i)
 if a==0x4e5cc0:
  assert arg(1)==objects[0];event('delete',[0]);mutate('delete');ret()
 elif a==0x4e25b0:
  assert arg(1)==objects[0]+56 and [arg(i) for i in range(2,8)]==[bits(25),0,20,2,objects[0],0]
  event('damage',[bits(25),0,20,2,0,-1]+point(arg(1)));mutate('damage',arg(1));ret()
 elif a==0x4ff5b0:
  event('buff',[objref(arg(1)),arg(2)]);mutate('buff');ret()
 elif a==0x517590:
  event('valid',[arg(1),arg(2)]);mutate('valid');ret(int(current['valid']))
 elif a==0x4e7010:
  assert arg(1)==objects[0];event('move',[0]+point(arg(2)));mutate('move',arg(2))
  u.mem_write(objects[0]+56,bytes(u.mem_read(arg(2),8)));ret()
 elif a==0x424800:
  assert [arg(1),arg(2)]==[127,2];event('sound',[127,2]);mutate('sound');ret(current['sound'])
 elif a==0x501960:
  assert [arg(3),arg(4)]==[0,0];event('audio',[arg(1),objref(arg(2)),0,0]);mutate('audio');ret()
 elif a==0x535250:
  assert [arg(2),arg(3)]==[0,0];p=arg(1);event('trace',[get(p+4*i) for i in range(4)]+[0,0,arg(4)]);mutate('trace',p);ret(int(current['valid']))
 elif a==0x537af0:
  assert arg(1)==objects[0]+56 and arg(2)==2;event('scorch',[2]+point(arg(1)));mutate('scorch',arg(1));ret()
 elif a==0x522ff0:
  assert arg(1)==138;event('fx',[138]+point(arg(2)));mutate('fx',arg(2));ret()
 elif a==0x4d9110:
  assert arg(1)==objects[0]+56 and arg(2)==30;event('quake',[30]+point(arg(1)));mutate('quake',arg(1));ret()
 else:raise AssertionError((hex(a),current['name']))
def invalid(_u,access,a,n,v,_):
 faults.append(dict(address=a,size=n,instruction=u.reg_read(UC_X86_REG_EIP)));return False
u.hook_add(UC_HOOK_CODE,hook);u.hook_add(UC_HOOK_MEM_INVALID,invalid)
u.reg_write(UC_X86_REG_ESP,sp);u.reg_write(UC_X86_REG_FPCW,0x27f);u.emu_start(0x4014b5,0x4014c4,count=20000)
assert u.reg_read(UC_X86_REG_FPCW)==0xe3f
initial=u.context_save();recording=True;cases=[]
def add(kind,name,**kw):
 c=dict(name=name,kind=kind,objects=[[0,0,bits(12.1),bits(34.3),bits(12.1),bits(34.3),bits(1),100,1000,bits(1.1),bits(2.3),bits(11.1)],[4,0,bits(300.1),bits(400.3),0,0,0,0,0,0,0,0],[4,0,bits(-70),bits(80),0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0,0,0,0]],owners=[1,-1,-1,-1],update_refs=[0,0,1,1],player_refs=[0,1],cursors=[[321,654],[0xffffff85,0xfffffe38]],frame=111,fps=30,valid=True,sound=0x1234abcd,mutation='')
 c.update(kw);cases.append(c);return c
for kind in range(6):
 for fps in (0,1,30,0x40000000,0x7fffffff,0x80000000,0xffffffff):
  ages=[0,1,2,3,4,9,10,11,60,61,0x7fffffff,0x80000000,0xfffffff6]
  factor=[0,2,300,20,3,0][kind]
  if factor:ages+=list({(factor*fps+x)&0xffffffff for x in (-1,0,1)})
  for age in sorted(set(ages)):
   c=add(kind,f'{specs[kind][0]}/age/{fps}/{age}',fps=fps,frame=(100+age)&0xffffffff)
   if kind==5:c['objects'][0][8]=(c['frame']+100)&0xffffffff
for kind in (2,3):
 for x,y in [(0,0),(16777217,-16777217),(16777219,-16777219),(0x7fffffff,-0x80000000),(-0x7fffffff,0x7ffffffe),(-321,654)]:
  for valid in (False,True):add(kind,f'{specs[kind][0]}/cursor/{x}/{y}/{valid}',cursors=[[x&0xffffffff,y&0xffffffff],[3,4]],valid=valid)
 for flags in (0,0x20,0x8000,0x8020):
  c=add(kind,f'{specs[kind][0]}/owner-flags/{flags}');c['objects'][1][1]=flags
 for service in ('delete','sound','valid','trace'):
  for action in ('owner','owner-nil','update','player','cursor','position','created','clock'):
   c=add(kind,f'{specs[kind][0]}/mutation/{service}/{action}',mutation=service+':'+action)
   if service in ('delete','sound'):c['objects'][1][1]=0x20
 for owner in (-1,0,2):
  if kind==3 and owner==-1:continue
  c=add(kind,f'{specs[kind][0]}/owner-ref/{owner}',owners=[owner,-1,-1,-1])
  if owner==0:c['objects'][0][0]=4
for cls in (0,2,4,0x200004):
 for flags in (0,0x20):
  c=add(2,f'moonglow/class/{cls}/{flags}');c['objects'][1][:2]=[cls,flags]
add(1,'fragment/old-future-creation-exact',frame=10);cases[-1]['objects'][0][7]=20
for mutation in ('damage:clock','damage:created','damage:position','damage:pos-pointer'):add(1,'fragment/'+mutation,frame=161,mutation=mutation)
for z in (0,0x80000000,bits(-1),bits(1),bits(199.9),bits(200),bits(201),0x7f800000,0xff800000,0x7fc12345,0xffc12345):
 for flags in (0,0x80000000,0x20,0x80000020):
  c=add(4,f'fist/height/{z:x}/{flags:x}');c['objects'][0][6]=z;c['objects'][0][1]=flags
for n in range(32):
 c=add(4,f'fist/point-rounding/{n}');c['objects'][0][6]=0;c['objects'][0][2]=bits(100.1+n*.031);c['objects'][0][3]=bits(-200.3-n*.073);c['objects'][0][11]=bits(11.1+n*.113)
for service in ('audio','scorch','fx','quake','delete'):
 for action in ('position','radius','height','unmark','created','clock','pos-pointer'):
  if action=='pos-pointer' and service not in ('scorch','quake'):continue
  c=add(4,'fist/'+service+':'+action,mutation=service+':'+action);c['objects'][0][6]=0
  if service=='delete':c['objects'][0][6]=bits(201);c['objects'][0][1]=0x80000000
for valid in (False,True):
 for mutation in ('','trace:position','trace:prev-position','trace:created','trace:deadline','trace:clock','trace:trace-point'):
  add(5,f'flame/trace/{valid}/{mutation}',valid=valid,mutation=mutation)
for raw in (0,0x80000000,1,bits(12.1),0x7f800000,0xff800000,0x7fc12345,0xffc12345):
 for axis in (0,1):
  c=add(5,f'flame/equality/{axis}/{raw:x}');c['objects'][0][4+axis]=raw
add(3,'telekinesis/trace-target-write',mutation='trace:trace-point')
rows=[]
for mode,cw in [('actual-startup',0xe3f),('nearest-counterfactual',0x23f)]:
 for current in cases:
  u.context_restore(initial);u.reg_write(UC_X86_REG_FPCW,cw);u.mem_write(heap,b'\xa5'*0x10000);u.mem_write(stack,bytes(0x10000));events.clear();fx_count=0
  for n,p in enumerate(objects):
   for off,value in zip(offsets,current['objects'][n]):put(p+off,value)
   put(p+508,ref(current['owners'][n]));put(p+748,0 if current['update_refs'][n]<0 else updates[current['update_refs'][n]])
  for n,p in enumerate(updates):put(p+276,0 if current['player_refs'][n]<0 else players[current['player_refs'][n]])
  for n,p in enumerate(players):put(p+2284,current['cursors'][n][0]);put(p+2288,current['cursors'][n][1])
  put(frame_address,current['frame']);put(fps_address,current['fps']);before=bytes(u.mem_read(heap,0x10000))
  put(sp,stop);put(sp+4,objects[0]);u.reg_write(UC_X86_REG_ESP,sp)
  u.emu_start(specs[current['kind']][1],stop,count=5000)
  assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4 and u.reg_read(UC_X86_REG_FPCW)==cw
  expected=bytearray(before)
  for p in objects:
   for off in offsets+(508,748):expected[p-heap+off:p-heap+off+4]=u.mem_read(p+off,4)
  for p in updates:expected[p-heap+276:p-heap+280]=u.mem_read(p+276,4)
  for p in players:expected[p-heap+2284:p-heap+2292]=u.mem_read(p+2284,8)
  assert bytes(u.mem_read(heap,0x10000))==bytes(expected)
  rows.append(dict(mode=mode,input=copy.deepcopy(current),output=state(),events=copy.deepcopy(events),nonoutput_bytes_preserved=True))
assert not faults and original.read_bytes()==blob
proof=dict(original_sha256=sha(blob),original_file_unchanged=True,original_startup_FPCW=0xe3f,contexts=len(cases),executions=len(rows),regions=regions,entries=dict(entries),rows=rows,faults=faults,executed_instructions=[dict(address=hex(a),bytes=b) for a,b in sorted(trace.items())],arithmetic_or_conversion_substitutions=0,scope='Six complete unchanged original updates, actual PC53/ToZero and nearest counterfactual. Declared damage/delete/buff/valid/move/sound/audio/trace/scorch/FX/quake services with prepared reentrant mutations. Full world bodies, native original-game execution and signaling-NaN parity not inferred.')
with args.proof.open('x') as f:json.dump(proof,f,separators=(',',':'))
fixture={k:v for k,v in proof.items() if k not in ('rows','executed_instructions')};fixture['rows']=[r for r in rows if r['mode']=='actual-startup']
with args.fixture.open('x') as f:json.dump(fixture,f,separators=(',',':'))
print(json.dumps({k:proof[k] for k in ['contexts','executions','entries','faults']}))

if args.fault_proof is not None:
 fault_rows=[]
 required=[(k,'nil-source',-1,1,0,0,-1,-1) for k in range(6)]
 required += [
  (2,'nil-owner-safe',0,-1,0,0,-1,-1),
  (2,'nonplayer-owner-safe',0,1,0,0,-1,-1),
  (2,'destroyed-nil-update-safe',0,1,4,0x20,-1,-1),
  (2,'active-nil-update',0,1,4,0,-1,-1),
  (2,'active-nil-player',0,1,4,0,0,-1),
  (3,'nil-owner',0,-1,0,0,-1,-1),
  (3,'destroyed-nil-update-safe',0,1,0,0x20,-1,-1),
  (3,'active-nil-update',0,1,0,0,-1,-1),
  (3,'active-nil-player',0,1,0,0,0,-1)]
 for k,name,src,own,cls,flags,upd,pl in required:
  current=dict(cases[0],kind=k,name=name,mutation='')
  u.context_restore(initial);u.mem_write(heap,bytes(0x10000))
  put(frame_address,0);put(fps_address,30)
  put(objects[0]+508,ref(own));put(objects[0]+136,100)
  put(objects[1]+8,cls);put(objects[1]+16,flags)
  put(objects[1]+748,0 if upd<0 else updates[upd])
  put(updates[0]+276,0 if pl<0 else players[pl])
  put(sp,stop);put(sp+4,ref(src));u.reg_write(UC_X86_REG_ESP,sp)
  faults.clear();events.clear();error=None
  try:u.emu_start(specs[k][1],stop,count=2000)
  except Exception as exc:error=str(exc)
  expected=not name.endswith('safe')
  assert bool(faults)==expected,(k,name,faults)
  if not expected:assert u.reg_read(UC_X86_REG_EIP)==stop
  fault_rows.append(dict(kind=k,name=name,fault=expected,events=list(events),faults=list(faults),error=error))
 assert original.read_bytes()==blob
 with args.fault_proof.open('x') as f:
  json.dump({'original_sha256':sha(blob),'rows':fault_rows,
             'faults':sum(r['fault'] for r in fault_rows),
             'safe_controls':sum(not r['fault'] for r in fault_rows)},f,indent=2)
