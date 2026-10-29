"""Execute unchanged Plasma create/update/candidate/destroy and FISTP code.

Type lookup, position/facing/distance, enemy/visibility, binary64 balance,
ray/point effects, virtual damage, state/audio/charge reporting and circle
enumeration are declared services. Enumeration invokes the actual original
candidate function with its original context. Caller packing, arithmetic,
reloads, x87 operations, the converter and candidate code are not replaced.
The complete logical record, native-mapped scratch and object/update/wand state
are outputs. Destroy retains its weapon link; native lifetime cleanup belongs
to the duration owner, outside this original-root comparison. Only Player
cursor links and a zero nonpointer monster +288 slot are valid cursor inputs.
"""
import argparse, collections, copy, hashlib, json, math, struct
from pathlib import Path
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_INVALID, UC_HOOK_MEM_READ
from unicorn.x86_const import *

p = argparse.ArgumentParser(description=__doc__)
for key in ['game','fixture','proof']: p.add_argument('--'+key,type=Path,required=True)
p.add_argument('--control',type=lambda v:int(v,0),choices=[0xe3f,0x23f],default=0xe3f)
a=p.parse_args(); blob=a.game.read_bytes(); sha=lambda b:hashlib.sha256(b).hexdigest()
assert sha(blob)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,60)[0];opt=pe+24;base=struct.unpack_from('<I',blob,opt+28)[0]
table=opt+struct.unpack_from('<H',blob,pe+20)[0]
u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,(struct.unpack_from('<I',blob,opt+56)[0]+4095)&~4095)
for n in range(struct.unpack_from('<H',blob,pe+6)[0]):
    _,va,size,off=struct.unpack_from('<IIII',blob,table+40*n+8)
    if size:u.mem_write(base+va,blob[off:off+size])
heap,stack,code=0x21000000,0x22000000,0x23000000
for address in [heap,stack,code]:u.mem_map(address,0x40000)
record=heap+0x1000;objects=[heap+0x2000+0x400*i for i in range(8)]
updates=[heap+0x10000+0x600*i for i in range(8)]
players=[heap+0x18000+0x1000*i for i in range(8)]
wands=[heap+0x24000+0x200*i for i in range(8)]
sp,stop,init,thunk,scratch,cw,callback_end,damage_entry=stack+0x3e000,code+0x100,code+0x200,code+0x300,code+0x400,code+0x500,code+0x600,code+0x700
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw))
u.mem_write(thunk,b'\xdd\x05'+struct.pack('<I',scratch)+b'\xc3')
fps,frame,selected,hec,orb,nearest,limit=0x85b3fc,0x84ea04,0x833c10,0x833c14,0x833c18,0x5c6934,0x583c18
get=lambda addr:struct.unpack('<I',u.mem_read(addr,4))[0]
get16=lambda addr:struct.unpack('<H',u.mem_read(addr,2))[0]
get8=lambda addr:bytes(u.mem_read(addr,1))[0]
put=lambda addr,value:u.mem_write(addr,struct.pack('<I',value&0xffffffff))
put16=lambda addr,value:u.mem_write(addr,struct.pack('<H',value&0xffff))
put8=lambda addr,value:u.mem_write(addr,bytes([value&255]))
words=lambda addr:[get(addr),get(addr+4)]
signed=lambda v,b:((v&((1<<b)-1))^(1<<(b-1)))-(1<<(b-1))
f32=lambda v:struct.unpack('<I',struct.pack('<f',v))[0]
point=lambda x,y:[f32(x),f32(y)]
f64=lambda v:struct.pack('<d',v).hex()
ref=lambda i:0 if i<0 else objects[i]
index=lambda v:-1 if v==0 else objects.index(v)
up_index=lambda v:updates.index(v)
wand_index=lambda v:wands.index(v)
player_index=lambda v:players.index(v)
cstr=lambda addr:bytes(u.mem_read(addr,128)).split(b'\0',1)[0].decode('ascii')
regions={'create':(0x531580,0x5315f2),'update':(0x531600,0x531911),'candidate':(0x531920,0x5319d5),'destroy':(0x5319e0,0x5319f6),'converter':(0x419a70,0x419a81)}
traces={n:{} for n in regions};entries,services,mutations,constant_reads=collections.Counter(),collections.Counter(),collections.Counter(),collections.Counter()
active,current,events,calls,changed,allowed,faults,circle=False,{},[],collections.Counter(),False,set(),[],None

def state():
    return dict(caster=index(get(record+16)),target=index(get(record+48)),weapon=index(get(record+72)),ray=index(get(record+36)),
        grace=get(record+76),expiry=get(record+68),flags=get(record+88),position=words(record+28),
        fps=get(fps),frame=get(frame),caches=[get(hec),get(orb)],selected=index(get(selected)),nearest=get(nearest),
        types=[get16(o+4) for o in objects],classes=[get(o+8) for o in objects],subclasses=[get(o+12) for o in objects],
        object_flags=[get(o+16) for o in objects],positions=[words(o+56) for o in objects],directions=[get16(o+124) for o in objects],
        updates=[up_index(get(o+748)) for o in objects],use_data=[wand_index(get(o+736)) for o in objects],
        equipped=[index(get(d+104)) for d in updates],cursors=[index(get(d+288)) for d in updates],
        player_links=[player_index(get(d+276)) for d in updates],player_indices=[get8(p+2064) for p in players],
        wands=[[get(w+96),get8(w+108),get8(w+109),get(w+112)] for w in wands])

def allow(addr,size):allowed.update(range(addr,addr+size))
def write(addr,value):allow(addr,4);put(addr,value)
def write8(addr,value):allow(addr,1);put8(addr,value)
def writepoint(addr,value):
    for i,w in enumerate(value):write(addr+4*i,w)
def mutate(stage):
    global changed
    at=calls[stage];calls[stage]+=1;m=current['mutation']
    if changed or m['stage']!=stage or m['at']!=at:return
    changed=True;action=m['action'];mutations[stage+'/'+str(at)+'/'+action]+=1
    if action=='caster-b':write(record+16,objects[1])
    elif action=='caster-nil':write(record+16,0)
    elif action=='target-b':write(record+48,objects[3])
    elif action=='target-nil':write(record+48,0)
    elif action=='weapon-b':write(record+72,objects[7])
    elif action=='weapon-nil':write(record+72,0)
    elif action=='ray-b':write(record+36,objects[4])
    elif action=='ray-nil':write(record+36,0)
    elif action=='grace':write(record+76,2)
    elif action=='grace-zero':write(record+76,0)
    elif action=='expiry':write(record+68,0x13579bdf)
    elif action=='record-flags':write(record+88,get(record+88)^0x20)
    elif action=='record-position':writepoint(record+28,point(99.25,-88.5))
    elif action=='fps':write(fps,60)
    elif action=='frame':write(frame,0xfffffff9)
    elif action=='cache-hec':write(hec,20)
    elif action=='cache-orb':write(orb,20)
    elif action=='selected-b':write(selected,objects[5])
    elif action=='nearest':write(nearest,f32(4))
    elif action=='caster-update-b':write(objects[0]+748,updates[1])
    elif action=='equipped-b':write(updates[0]+104,objects[7])
    elif action=='cursor-b':write(updates[0]+288,objects[3])
    elif action=='player-b':write(updates[0]+276,players[1])
    elif action=='caster-position':writepoint(objects[0]+56,point(3,4))
    elif action=='target-position':writepoint(objects[2]+56,point(1.25,-2.5))
    elif action=='target-dead':write(objects[2]+16,get(objects[2]+16)|0x8020)
    elif action=='wand-flags':write(wands[6]+96,0x12345678)
    elif action=='wand-charge':write8(wands[6]+108,1)
    elif action=='wand-max':write8(wands[6]+109,9)
    elif action=='wand-data-b':write(objects[6]+736,wands[7])
    else:raise AssertionError(action)

def event(kind,args):events.append(dict(kind=kind,args=args,state=state()));mutate(kind)
def ret(value=0):
    esp=u.reg_read(UC_X86_REG_ESP);u.reg_write(UC_X86_REG_EAX,value&0xffffffff);u.reg_write(UC_X86_REG_EIP,get(esp));u.reg_write(UC_X86_REG_ESP,esp+4)
def floatret(bits):u.mem_write(scratch,bytes.fromhex(bits));u.reg_write(UC_X86_REG_EIP,thunk)
def next_candidate():
    global circle
    if circle['next']>=len(current['candidates']):
        u.reg_write(UC_X86_REG_ESP,circle['esp']);circle=None;ret(0x13572468);return
    target=current['candidates'][circle['next']];circle['next']+=1
    event('visit',[target,index(circle['caster'])])
    esp=circle['esp']-48;put(esp,callback_end);put(esp+4,ref(target));put(esp+8,circle['caster'])
    u.reg_write(UC_X86_REG_ESP,esp);u.reg_write(UC_X86_REG_EIP,circle['callback'])

def hook(_u,pc,length,_):
    global circle
    if not active:return
    if pc==stop:u.emu_stop();return
    for name,(lo,hi) in regions.items():
        if lo<=pc<hi:
            traces[name][pc]=bytes(u.mem_read(pc,length))
            if pc==lo:entries[name]+=1
            return
    if thunk<=pc<thunk+7:return
    if pc==callback_end:next_candidate();return
    esp=u.reg_read(UC_X86_REG_ESP);arg=lambda n:get(esp+4*n);services[hex(pc)]+=1
    if pc==0x4e3aa0:
        key=cstr(arg(1));assert key in ['Hecubah','HecubahWithOrb'];event('lookup',[key]);ret(current['lookup'][key=='HecubahWithOrb'])
    elif pc==0x4fea70:event('position',[index(arg(1)),*words(arg(2))]);ret(current['position_result'])
    elif pc==0x4e6e50:
        first=index(arg(1)-56);second=index(arg(3)-56)
        event('facing',[first,signed(arg(2),16),second,*words(arg(1)),*words(arg(3))]);ret(current['facing'][first][second])
    elif pc==0x4e6c00:
        first,second=index(arg(1)),index(arg(2));event('distance',[first,second]);floatret(current['distance_bits64'][first][second])
    elif pc==0x5330c0:
        first,second=index(arg(1)),index(arg(2));event('enemy',[first,second]);ret(current['enemy'][first][second])
    elif pc==0x5370e0:
        assert arg(3)==0;first,second=index(arg(1)),index(arg(2));event('los',[first,second]);ret(current['los'][first][second])
    elif pc==0x517f90:
        assert arg(2)==f32(400) and arg(3)==regions['candidate'][0] and arg(1)==arg(4)+56
        event('circle',[index(arg(4)),*words(arg(1)),arg(2)])
        assert circle is None;circle=dict(esp=esp,callback=arg(3),caster=arg(4),next=0);next_candidate()
    elif pc==0x4fef90:
        assert arg(1)==record;event('stop',[index(arg(2))]);ret(0x24681357)
    elif pc==0x4ff130:
        assert arg(1)==record;event('start',[]);ret(0x24681357)
    elif pc==0x419d40:
        key=cstr(arg(1));assert key in current['balances'];event('balance',[key]);floatret(current['balances'][key])
    elif pc==damage_entry:
        assert arg(3)==0 and arg(5)==14;event('damage',[index(arg(1)),index(arg(2)),signed(arg(4),32)]);ret(0x13572468)
    elif pc==0x522ff0:
        assert arg(1)==131;owner=-1 if arg(2)==record+28 else index(arg(2)-56)
        event('fx',[arg(1),owner,*words(arg(2))]);ret(0x13572468)
    elif pc==0x4fa020:
        assert arg(2)==22;event('state',[index(arg(1)),arg(2)]);ret(1)
    elif pc==0x501960:
        assert arg(1)==98 and arg(3)==arg(4)==0;event('audio',[arg(1),index(arg(2))]);ret(0x24681357)
    elif pc==0x4d82b0:event('report',[arg(1),index(arg(2)),arg(3),arg(4)]);ret(0x13572468)
    else:raise AssertionError((current['name'],'undeclared instruction/service',hex(pc)))
def invalid(_u,access,address,size,value,_):faults.append(dict(name=current['name'],pc=hex(u.reg_read(UC_X86_REG_EIP)),address=hex(address)));return False
def memread(_u,access,address,size,value,_):
    if active and address==limit:constant_reads[(address,size)]+=1
u.hook_add(UC_HOOK_CODE,hook);u.hook_add(UC_HOOK_MEM_INVALID,invalid);u.hook_add(UC_HOOK_MEM_READ,memread)
seed=bytes((i*31+19)&255 for i in range(0x40000));registers={UC_X86_REG_EBX:0x13572468,UC_X86_REG_ESI:0x24681357,UC_X86_REG_EDI:0x98765432,UC_X86_REG_EBP:0xfedcba98}

def run(case):
    global active,current,events,calls,changed,allowed,circle
    active,current,events,calls,changed,allowed,circle=False,case,[],collections.Counter(),False,set(),None;u.mem_write(heap,seed)
    for off,value in [(16,ref(case['caster'])),(48,ref(case['target'])),(72,ref(case['weapon'])),(36,ref(case['ray'])),(76,case['grace']),(68,case['expiry']),(88,case['flags'])]:put(record+off,value)
    for j,value in enumerate(case['position']):put(record+28+4*j,value)
    for i,o in enumerate(objects):
        put16(o+4,case['types'][i]);put(o+8,case['classes'][i]);put(o+12,case['subclasses'][i]);put(o+16,case['object_flags'][i]);put16(o+124,case['directions'][i])
        put(o+748,updates[case['updates'][i]]);put(o+736,wands[case['use_data'][i]]);put(o+716,damage_entry)
        for j,value in enumerate(case['positions'][i]):put(o+56+4*j,value)
    for i,d in enumerate(updates):put(d+104,ref(case['equipped'][i]));put(d+288,ref(case['cursors'][i]));put(d+276,players[case['player_links'][i]])
    for i,pl in enumerate(players):put8(pl+2064,case['player_indices'][i])
    for w,row in zip(wands,case['wands']):put(w+96,row[0]);put8(w+108,row[1]);put8(w+109,row[2]);put(w+112,row[3])
    globals_before={}
    for start,size in [(selected-16,44),(nearest-12,28)]:u.mem_write(start,bytes(range(size)))
    for addr,value in [(fps,case['fps']),(frame,case['frame']),(selected,ref(case['selected'])),(hec,case['caches'][0]),(orb,case['caches'][1]),(nearest,case['nearest'])]:put(addr,value)
    for start,size in [(selected-16,44),(nearest-12,28),(fps,4),(frame,4)]:globals_before[start]=bytes(u.mem_read(start,size))
    u.mem_write(cw,struct.pack('<H',a.control));u.emu_start(init,init+8,count=2);before=bytes(u.mem_read(heap,len(seed)))
    u.mem_write(sp-512,bytes(1024));put(sp,stop);put(sp+4,record)
    for reg,value in registers.items():u.reg_write(reg,value)
    u.reg_write(UC_X86_REG_EFLAGS,0x202);u.reg_write(UC_X86_REG_ESP,sp);active=True
    u.emu_start(regions[case['kind']][0],stop+1,count=100000);active=False
    assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4 and get(sp)==stop and circle is None
    assert all(u.reg_read(reg)==v for reg,v in registers.items())
    assert bytes(u.mem_read(sp-512,256))==bytes(256) and bytes(u.mem_read(sp+8,256))==bytes(256)
    assert u.reg_read(UC_X86_REG_FPCW)==a.control and u.reg_read(UC_X86_REG_FPSW)&0x3800==0
    if case['kind']=='create':
        for off in [48,72,76]:allow(record+off,4)
        allow(record+88,1)
    if case['kind']=='update':
        for off in [48,36,76,68]:allow(record+off,4)
        for w in wands:allow(w+108,1);allow(w+112,4)
        for addr in [selected,hec,orb,nearest]:allow(addr,4)
    if case['kind']=='destroy':
        for w in wands:allow(w+96,4)
    after=bytes(u.mem_read(heap,len(seed)));assert all(x==y or heap+i in allowed for i,(x,y) in enumerate(zip(before,after))),(case['name'],'nonoutput heap canary')
    for start,body in globals_before.items():
        assert all(x==y or start+i in allowed for i,(x,y) in enumerate(zip(body,bytes(u.mem_read(start,len(body)))))),(case['name'],'tracked global canary')
    if case['mutation']['stage']:assert changed,(case['name'],'mutation not reached')
    return dict(input=copy.deepcopy(case),result=dict(ret=None if case['kind']=='destroy' else u.reg_read(UC_X86_REG_EAX),state=state(),events=copy.deepcopy(events)))

template=dict(name='',kind='update',caster=0,target=2,weapon=6,ray=2,
    grace=0,expiry=0x87654321,flags=0x77556601,position=point(1.725,-22.25),
    fps=30,frame=0xfffffff0,caches=[10,11],selected=5,nearest=f32(12345),
    types=[1,2,12,10,11,13,14,15],classes=[4,4,2,2,2,2,0x1000,0x1000],
    subclasses=[0,0,0,0,0,0,0x4000000,0x4000000],object_flags=[0]*8,
    positions=[point(0,0),point(5,-7),point(100,0),point(20,30),point(-50,5),point(3,4),point(0,0),point(0,0)],
    directions=[0x8001,3,5,7,9,11,13,15],updates=list(range(8)),use_data=list(range(8)),
    equipped=[6]*8,cursors=[2,3,-1,-1,-1,-1,-1,-1],player_links=list(range(8)),player_indices=[7,201,3,4,5,6,8,9],
    wands=[[0x87654324,3,4,75] for _ in range(8)],lookup=[10,11],
    position_result=0,facing=[[0]*8 for _ in range(8)],enemy=[[1]*8 for _ in range(8)],los=[[1]*8 for _ in range(8)],
    distance_bits64=[[f64(100)]*8 for _ in range(8)],
    balances={k:f64(v) for k,v in [('PlasmaDamage',2.75),('PlasmaDamageHecubah',7.75),('PlasmaSearchTime',12.75)]},
    candidates=[2,3,4],mutation=dict(stage='',at=0,action=''))
cases=[]
def add(kind,name,**changes):
    case=copy.deepcopy(template);case.update(kind=kind);case.update(changes);case['name']=kind+'/'+name;cases.append(case);return case
def vec(key,i,value):
    v=copy.deepcopy(template[key]);v[i]=value;return v
def mat(key,i,j,value):
    v=copy.deepcopy(template[key]);v[i][j]=value;return v
for kind in ['create','update','destroy']:add(kind,'default')
for kind in ['create','update']:
    add(kind,'nil-caster',caster=-1)
    for cls in [0,1,2,3,4,5,6,7,0x100,0x20004,0x80000004]:
        add(kind,'class-'+hex(cls),classes=vec('classes',0,cls),cursors=[-1]*8 if cls&4==0 else template['cursors'])
for value in [-1,7]:add('create','equipped-'+str(value),equipped=vec('equipped',0,value))
for sub in [0,1,0x4000000,0xffffffff]:add('create','subclass-'+hex(sub),subclasses=vec('subclasses',6,sub))
for flags in [0,1,4,0x100,0xffffffff]:
    w=copy.deepcopy(template['wands']);w[6][0]=flags;add('create','wand-flags-'+hex(flags),wands=w)
for flags in [0,2,0x20,0x100,0xfffffffd]:add('create','record-flags-'+hex(flags),flags=flags)
for stage in ['fx']:
    for action in ['caster-b','caster-nil','equipped-b','caster-update-b','record-flags','record-position','weapon-b','target-b','grace','wand-flags','wand-data-b']:
        add('create',stage+'-'+action,mutation=dict(stage=stage,at=0,action=action))
for caches in [[0,0],[0,77],[1,11],[0xffffffff,0]]:add('update','caches-'+str(caches),caches=caches)
for lookup in [[0,0],[0,11],[0x1000a,11],[10,0x1000b]]:add('update','lookup-'+str(lookup),caches=[0,0],lookup=lookup)
for flags in [0x20,0xffffffff,0x2000,0x10020]:add('update','flags-'+hex(flags),flags=flags)
for value in [0,1,0xffffffff]:add('update','position-'+str(value),classes=vec('classes',0,2),cursors=[-1]*8,position_result=value)
for target in [-1,3,4]:add('update','target-'+str(target),target=target)
for flags in [0x20,0x8000,0x8020,0x10000]:add('update','target-flags-'+hex(flags),object_flags=vec('object_flags',2,flags))
for facing in [0,1,2,3,0x100,0xffffffff]:
    for grace in [0,1,2,0xffffffff]:add('update','facing-'+str(facing)+'-grace-'+str(grace),facing=mat('facing',0,2,facing),grace=grace)
for rate in [3,4,30,60,0x55555556,0x80000000,0xffffffff]:
    add('update','fps-'+str(rate),fps=rate,facing=mat('facing',0,2,2))
for value in [399.999999999,400,400.000000001,float('nan'),float('inf'),float('-inf')]:
    for cursor in [False,True]:add('update',('cursor' if cursor else 'retained')+'-distance-'+str(value),target=-1 if cursor else 2,distance_bits64=mat('distance_bits64',0 if cursor else 2,2 if cursor else 0,f64(value)))
add('update','retained-no-los',los=mat('los',0,2,0))
add('update','cursor-not-enemy',target=-1,enemy=mat('enemy',0,2,0))
add('update','cursor-no-los-still-selected',target=-1,los=mat('los',0,2,0))
for candidates in [[],[0],[2],[3,2],[2,3],[2,2,3],[5,4,3,2]]:
    add('update','candidates-'+str(candidates),target=-1,cursors=[-1]*8,candidates=candidates)
for cls in [0,1,2,4,6,0x20000,0x20002,0x80000000]:
    add('update','candidate-class-'+hex(cls),target=-1,cursors=[-1]*8,candidates=[2],classes=vec('classes',2,cls))
for flags in [0x20,0x8000,0x8020]:add('update','candidate-flags-'+hex(flags),target=-1,cursors=[-1]*8,candidates=[2],object_flags=vec('object_flags',2,flags))
add('update','candidate-excluded-monster',target=-1,cursors=[-1]*8,candidates=[2],subclasses=vec('subclasses',2,0x8000))
add('update','candidate-not-enemy',target=-1,cursors=[-1]*8,candidates=[2],enemy=mat('enemy',0,2,0))
add('update','candidate-no-los',target=-1,cursors=[-1]*8,candidates=[2],los=mat('los',0,2,0))
for facing in [0,1,2,3,4,8,12,0xffffffff]:add('update','candidate-facing-'+str(facing),target=-1,cursors=[-1]*8,candidates=[2],facing=mat('facing',0,2,facing))
for label,x,y in [('boundary',400,0),('inside',399.9999694824219,0),('outside',400.0000305175781,0),('tie',20,30),('nan',float('nan'),0),('inf',float('inf'),0),('tiny',2**-149,2**-149),('round',16.000001907348633,0.0001)]:
    add('update','candidate-'+label,target=-1,cursors=[-1]*8,candidates=[2,3],positions=vec('positions',2,point(x,y)))
for ray in [-1,2,3]:
    add('update','ray-'+str(ray),ray=ray)
    add('update','no-target-ray-'+str(ray),target=-1,cursors=[-1]*8,candidates=[],ray=ray)
for value in [0,10,11,65535]:add('update','target-type-'+str(value),types=vec('types',2,value))
values=[0,-0.0,0.99,-0.99,2.75,-2.75,1.99999999,-1.99999999,2147483647.0,2147483648.0,-2147483649.0,2147483520.0,1e300,-1e300,float('inf'),float('-inf'),float('nan'),2**-150,-2**-150]
for key in template['balances']:
    for value in values:
        balances=dict(template['balances']);balances[key]=f64(value)
        add('update','balance-'+key+'-'+str(value),balances=balances,types=vec('types',2,10) if key=='PlasmaDamageHecubah' else template['types'])
for charge,maxcharge in [(0,4),(1,4),(2,4),(255,255),(255,1)]:
    w=copy.deepcopy(template['wands']);w[6][1:3]=[charge,maxcharge];add('update','charge-'+str(charge)+'-'+str(maxcharge),wands=w)
add('update','no-weapon',weapon=-1)
add('update','nonplayer-charge',classes=vec('classes',0,2),cursors=[-1]*8)
for frame_value in [0,1,9,10,0x7fffffff,0x80000000,0xffffffff]:add('update','frame-'+str(frame_value),frame=frame_value)
for weapon in [-1,6,7]:
    for flags in [0,4,0xffffffff]:
        w=copy.deepcopy(template['wands']);w[6][0]=flags;w[7][0]=flags
        add('destroy','weapon-'+str(weapon)+'-flags-'+hex(flags),weapon=weapon,wands=w)

# Every mutation is applied after the observed service arguments/snapshot.
# Inputs remain within defined original pointer/divisor contracts. A removed
# live target is tested only at boundaries which subsequently handle nil.
stage_setups={
    'lookup':dict(caches=[0,0]),'position':dict(classes=vec('classes',0,2),cursors=[-1]*8),
    'facing':{},'distance':{},'los':{},'enemy':dict(target=-1),
    'circle':dict(target=-1,cursors=[-1]*8),'visit':dict(target=-1,cursors=[-1]*8),
    'stop':dict(ray=3),'start':dict(ray=-1),'balance':{},'damage':{},
    'state':{},'audio':{},'report':{}}
common=['caster-b','target-b','weapon-b','ray-b','grace','expiry','record-flags','record-position','fps','frame','cache-hec','cache-orb','caster-update-b','cursor-b','player-b','caster-position','target-position','wand-flags','wand-charge','wand-max','wand-data-b']
for stage,setup in stage_setups.items():
    for action in common:
        add('update','mutate-'+stage+'-'+action,**setup,mutation=dict(stage=stage,at=0,action=action))
for stage in ['circle','visit','enemy','facing','los']:
    for action in ['selected-b','nearest']:
        add('update','scratch-'+stage+'-'+action,target=-1,cursors=[-1]*8,mutation=dict(stage=stage,at=0,action=action))
for stage in ['lookup','position','facing','distance','los','stop','damage','state','report']:
    setup=stage_setups[stage]
    for action in (['target-nil'] if stage in ['facing','distance','los'] else []):
        # Original would pass nil to declared distance/LOS or dereference it.
        # These controls are intentionally not admitted as valid references.
        pass
for stage,at,action in [('damage',0,'target-dead'),('state',1,'weapon-b'),('state',1,'caster-update-b'),('state',1,'player-b'),('audio',0,'target-b'),('balance',1,'frame'),('stop',0,'ray-nil')]:
    setup=dict(stage_setups[stage]);add('update','later-'+stage+'-'+str(at)+'-'+action,**setup,mutation=dict(stage=stage,at=at,action=action))

assert len({c['name'] for c in cases})==len(cases)
rows=[run(case) for case in cases]
sealed=[]
for name,(lo,hi) in regions.items():
    ins=traces[name];covered={pc+i for pc,body in ins.items() for i in range(len(body))}
    assert covered==set(range(lo,hi)),(name,sorted(set(range(lo,hi))-covered))
    sealed.append(dict(name=name,address=hex(lo),end_exclusive=hex(hi),bytes=hi-lo,sha256=sha(bytes(u.mem_read(lo,hi-lo))),instructions=[dict(pc=hex(pc),bytes=body.hex()) for pc,body in sorted(ins.items())]))
encoded=(json.dumps(dict(original_sha256=sha(blob),control_word=a.control,rows=rows),separators=(',',':'))+'\n').encode()
with a.fixture.open('xb') as f:f.write(encoded)
proof=dict(original_sha256=sha(blob),control_word=a.control,contexts=len(rows),entries=dict(entries),regions=sealed,
    all_original_instruction_bytes_executed=True,zero_caller_packing_arithmetic_or_conversion_replacements=True,
    callee_saved_registers_stack_control_word_TOP_and_nonoutput_heap_and_global_canaries_preserved=True,
    services=dict(services),mutations=dict(mutations),constant_reads={str(k):v for k,v in constant_reads.items()},
    faults=faults,fixture_sha256=sha(encoded),destroy_EAX_ignored_by_void_consumer=True,scope=__doc__)
with a.proof.open('x') as f:json.dump(proof,f,indent=2);f.write('\n')
print(json.dumps({k:v for k,v in proof.items() if k not in ['regions','mutations','scope']}))
