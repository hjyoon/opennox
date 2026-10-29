"""Execute original facing actions, point/dot helpers and cancel thunk.

All602 instruction bytes and the original stock direction table execute.
Only PopAction50A160 is a declared service: its entry state, unit identity
and full raw return are retained. This does not revalidate that service's
stack/end-callback implementation. Internal normalized-vector observations
are retained separately from observable results; no FP return, operation or
comparison is replaced. Directions1 and angle indices are stock0..255;
Direction2 and SetAngle cover arbitrary word wrapping. The dispatcher ignores
the char return; raw EAX remains recorded. No general AI-idle or earned
campaign equivalence is claimed.
"""
import argparse,collections,copy,hashlib,itertools,json,math,struct
from pathlib import Path
from unicorn import Uc,UC_ARCH_X86,UC_MODE_32,UC_HOOK_CODE
from unicorn.x86_const import *

parser=argparse.ArgumentParser(description=__doc__)
for name in ['game','fixture','proof']:parser.add_argument('--'+name,type=Path,required=True)
parser.add_argument('--control',type=lambda s:int(s,0),choices=[0xe3f,0x23f],default=0xe3f)
args=parser.parse_args();blob=args.game.read_bytes();sha=lambda b:hashlib.sha256(b).hexdigest()
assert sha(blob)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,60)[0];opt=pe+24;base=struct.unpack_from('<I',blob,opt+28)[0]
table=opt+struct.unpack_from('<H',blob,pe+20)[0]
u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,(struct.unpack_from('<I',blob,opt+56)[0]+4095)&~4095)
for i in range(struct.unpack_from('<H',blob,pe+6)[0]):
    _,va,size,off=struct.unpack_from('<IIII',blob,table+40*i+8)
    if size:u.mem_write(base+va,blob[off:off+size])
heap,stack,code=0x21000000,0x22000000,0x23000000
u.mem_map(heap,0x10000);u.mem_map(stack,0x10000);u.mem_map(code,0x10000)
unit,target,update,point_arg=heap+0x1000,heap+0x2000,heap+0x3000,heap+0x6000
sp,stop,init,cw=stack+0xe000,code+0x100,code+0x200,code+0x300
regions=[('location',0x545210,38),('point-helper',0x545240,192),('object',0x545300,57),
         ('angle',0x545340,156),('set-angle',0x5453e0,90),('pop',0x532100,12),('facing-helper',0x534120,57)]
read32=lambda a:struct.unpack('<I',u.mem_read(a,4))[0]
read16=lambda a:struct.unpack('<H',u.mem_read(a,2))[0]
put32=lambda a,v:u.mem_write(a,struct.pack('<I',v&0xffffffff))
put16=lambda a,v:u.mem_write(a,struct.pack('<H',v&0xffff))
bits=lambda v:struct.unpack('<I',struct.pack('<f',v))[0]
value=lambda v:struct.unpack('<f',struct.pack('<I',v))[0]
point=lambda x,y:[bits(x),bits(y)]
readpoint=lambda a:[read32(a),read32(a+4)]
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw))
traces,entries,services={},collections.Counter(),collections.Counter()
active,current,events,internal=False,{},[],[]
action={'location':25,'point-helper':25,'object':26,'angle':27,'set-angle':28,'pop':27,'facing-helper':27}

def state():
    index=struct.unpack('<b',u.mem_read(update+544,1))[0]
    head=update+552+index*24
    arg0=read32(head+4)
    if current['function']=='object':arg0=-1 if arg0==0 else 1 if arg0==target else -2
    return dict(pos=readpoint(unit+56),target_pos=readpoint(target+56),
                direction1=read16(unit+124),direction2=read16(unit+126),
                head_index=index,head_action=read32(head),head_arg0=arg0,head_arg1=read32(head+8))

def ret(v):
    esp=u.reg_read(UC_X86_REG_ESP);u.reg_write(UC_X86_REG_EAX,v&0xffffffff)
    u.reg_write(UC_X86_REG_EIP,read32(esp));u.reg_write(UC_X86_REG_ESP,esp+4)

def hook(_u,pc,size,_):
    if not active:return
    if pc==stop:u.emu_stop();return
    for name,lo,n in regions:
        if lo<=pc<lo+n:
            traces[pc]=bytes(u.mem_read(pc,size))
            if pc==lo:
                entries[name]+=1
                if name=='facing-helper':
                    esp=u.reg_read(UC_X86_REG_ESP);assert read32(esp+4)==unit
                    internal.append(dict(direction=readpoint(read32(esp+8)),state=state()))
            return
    assert pc==0x50a160,(current['name'],hex(pc))
    esp=u.reg_read(UC_X86_REG_ESP);assert read32(esp+4)==unit
    services[hex(pc)]+=1;events.append(dict(kind='pop',args=[0],state=state()))
    ret(current['pop_return'])

u.hook_add(UC_HOOK_CODE,hook)
registers={UC_X86_REG_EBX:0x12345678,UC_X86_REG_ESI:0x23456789,UC_X86_REG_EDI:0x3456789a,UC_X86_REG_EBP:0x456789ab}
seed=bytes((i*31+19)&255 for i in range(0x10000))

def run(c):
    global active,current,events,internal
    active,current,events,internal=False,c,[],[]
    u.mem_write(heap,seed)
    for o,p in [(unit,c['pos']),(target,c['target_pos'])]:
        put32(o+8,2)
        for j,v in enumerate(p):put32(o+56+4*j,v)
    put16(unit+124,c['direction1']);put16(unit+126,c['direction2']);put32(unit+748,update)
    u.mem_write(update+544,struct.pack('<b',c['head_index']))
    head=update+552+24*c['head_index'];put32(head,action[c['function']])
    first,second=c['target_pos']
    if c['function']=='object':first,second=(0 if c['target_nil'] else target),0x12345678
    elif c['function'] in ['angle','set-angle','pop','facing-helper']:first,second=c['angle'],0x12345678
    put32(head+4,first);put32(head+8,second)
    for j,v in enumerate(c['vector'] if c['function']=='facing-helper' else c['target_pos']):put32(point_arg+4*j,v)
    before=bytes(u.mem_read(heap,len(seed)))
    u.mem_write(cw,struct.pack('<H',args.control));u.emu_start(init,init+8,count=2)
    u.mem_write(sp-512,bytes(1024));put32(sp,stop);put32(sp+4,unit);put32(sp+8,point_arg)
    argc=2 if c['function'] in ['point-helper','facing-helper'] else 1
    if argc==1:put32(sp+8,0)
    for reg,v in registers.items():u.reg_write(reg,v)
    u.reg_write(UC_X86_REG_EFLAGS,0x202);u.reg_write(UC_X86_REG_ESP,sp)
    lo=next(a for name,a,_ in regions if name==c['function'])
    active=True
    try:u.emu_start(lo,stop+1,count=20000)
    except Exception as e:raise AssertionError((c['name'],hex(u.reg_read(UC_X86_REG_EIP)))) from e
    active=False
    assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4,c['name']
    assert all(u.reg_read(reg)==v for reg,v in registers.items()),c['name']
    assert u.reg_read(UC_X86_REG_FPCW)==args.control and u.reg_read(UC_X86_REG_FPSW)&0x3800==0,c['name']
    assert bytes(u.mem_read(sp-512,256))==bytes(256),c['name']
    assert bytes(u.mem_read(sp+4*(argc+1),256))==bytes(256),c['name']
    after=bytes(u.mem_read(heap,len(seed)))
    allowed=set(range(unit+124,unit+128))
    assert all(a==b or heap+i in allowed for i,(a,b) in enumerate(zip(before,after))),c['name']
    return dict(input=c,result=dict(return_value=u.reg_read(UC_X86_REG_EAX),events=copy.deepcopy(events),state=state()),
                internal_helper_observations=copy.deepcopy(internal))

default=dict(function='location',pos=point(10,20),target_pos=point(110,20),direction1=0,direction2=0,
             angle=64,head_index=1,target_nil=False,vector=point(1,0),pop_return=0x13572468)
cases=[]
def case(name,**kw):
    c=copy.deepcopy(default);c.update(kw,name=name);cases.append(c)
directions=[0,1,33,64,127,128,192,255]
targets=[(110,20),(10,120),(-90,20),(10,-80),(10,20),(110,120),(-90,-80),(10.000001,20)]
for function,d,p in itertools.product(['location','point-helper','object'],directions,targets):
    case(f'{function}-basic-{len(cases)}',function=function,direction1=d,target_pos=point(*p))
for function,d,a,d2 in itertools.product(['angle'],directions,directions,[0,252,0x7fff,0x8000,0xffff]):
    case(f'angle-{len(cases)}',function=function,direction1=d,angle=a,direction2=d2)
for a in [0,1,7,64,127,255,256,32767,32768,65535,0xffffffff,0x12345678,0x80000001]:
    case(f'set-angle-{a:08x}',function='set-angle',angle=a)
for index,function in itertools.product([0,1,5],['location','object','angle','set-angle']):
    case(f'head-{function}-{index}',function=function,head_index=index)
case('object-nil',function='object',target_nil=True)
for function,d,sign,offset in itertools.product(['location','point-helper','object'],directions,[-1,1],range(-2,3)):
    theta=d*math.tau/256+sign*math.acos(value(read32(0x583bcc)))
    p=point(10+100*math.cos(theta),20+100*math.sin(theta));p[1]=(p[1]+offset)&0xffffffff
    case(f'{function}-cone-{len(cases)}',function=function,direction1=d,target_pos=p)
for function,p,t in itertools.product(['location','point-helper','object'],
    [(10000000,-10000000),(0.03125,-0.0625),(-10000000,0.03125)],
    [(0.03125,-0.0625),(10000000,-10000000),(110,20)]):
    case(f'{function}-subtraction-{len(cases)}',function=function,pos=point(*p),target_pos=point(*t))
threshold=value(read32(0x583bcc))
for d,y,adj in itertools.product(range(0,256,8),[0,0.125,-0.75],range(-2,3)):
    fx,fy=[value(v) for v in readpoint(0x5b6658+8*d)]
    if abs(fx)>=0.05:vector=[(bits((threshold-fy*y)/fx)+adj)&0xffffffff,bits(y)]
    else:vector=[bits(y),(bits((threshold-fx*y)/fy)+adj)&0xffffffff]
    case(f'dot-boundary-{len(cases)}',function='facing-helper',direction1=d,vector=vector)
for function,v in itertools.product(['location','point-helper','object','facing-helper'],
    [0x00000000,0x80000000,0x7f7fffff,0xff7fffff,0x7f800000,0xff800000,0x7fc12345,0xffc12345]):
    case(f'{function}-edge-{v:08x}',function=function,target_pos=[v,v],vector=[v,v])
for d2,function in itertools.product([0,7,248,252,255,256,0x7fff,0x8000,0xffff],['location','point-helper','object']):
    case(f'{function}-word-{d2:04x}',function=function,direction2=d2,target_pos=point(10,120))
for v in [0,1,0xffffffff,0x13572468]:
    case(f'pop-{v:08x}',function='pop',pop_return=v)
    case(f'point-pop-return-{v:08x}',function='point-helper',pop_return=v)
fixture=dict(control_word=args.control,original_machine_code_executed=True,
             pop_service_body_declared=True,internal_helper_scratch_retained_separately=True,cases=[run(c) for c in cases])
covered={a+i for a,b in traces.items() for i in range(len(b))}
proof=dict(original_sha256=sha(blob),control_word=args.control,contexts=len(cases),
    regions=[dict(name=name,address=hex(a),size=n,sha256=sha(bytes(u.mem_read(a,n))),executed_bytes=len(set(range(a,a+n))&covered),missing_bytes=[hex(p) for p in range(a,a+n) if p not in covered]) for name,a,n in regions],
    data_regions=[dict(name=name,address=hex(a),size=n,sha256=sha(bytes(u.mem_read(a,n)))) for name,a,n in [('directions',0x5b6658,2048),('zero',0x58307c,4),('epsilon',0x583b28,4),('threshold',0x583bcc,4)]],
    entries=dict(entries),services=dict(services),faults=0,register_stack_FPCW_TOP_heap_canaries_preserved=True,
    original_machine_code_executed=True,independent_comparator_NOT_YET_BUILT=True,scope=__doc__)
assert args.game.read_bytes()==blob
for p,v in [(args.fixture,fixture),(args.proof,proof)]:
    with p.open('x') as f:json.dump(v,f,indent=2);f.write('\n')
print(json.dumps(proof))
