"""Execute unchanged Sentry callers, scan callback, projection and packet helper."""
import argparse,collections,copy,hashlib,json,math,struct
from pathlib import Path
from unicorn import Uc,UC_ARCH_X86,UC_MODE_32,UC_HOOK_CODE,UC_HOOK_MEM_INVALID
from unicorn.x86_const import *
p=argparse.ArgumentParser(description=__doc__)
for n in ('game','fixture','proof','fault-proof'):p.add_argument('--'+n,type=Path,required=True)
p.add_argument('--trig-profile',type=Path)
args=p.parse_args();blob=args.game.read_bytes();sha=lambda b:hashlib.sha256(b).hexdigest()
assert sha(blob)=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,60)[0];opt=pe+24;base=struct.unpack_from('<I',blob,opt+28)[0];size=(struct.unpack_from('<I',blob,opt+56)[0]+4095)&~4095;tab=opt+struct.unpack_from('<H',blob,pe+20)[0]
u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,size)
for n in range(struct.unpack_from('<H',blob,pe+6)[0]):
    _,va,count,off=struct.unpack_from('<IIII',blob,tab+40*n+8)
    if count:u.mem_write(base+va,blob[off:off+count])
stack,heap,stop=0x20000000,0x21000000,0x30000000
for a in (stack,heap,stop):u.mem_map(a,0x10000)
sp=stack+0xe000;objects=[heap+128+1024*n for n in range(4)];data=[heap+0x4000,heap+0x4100];player=heap+0x5000;context=heap+0x7000;ray=heap+0x7100;health=heap+0x7200;damage=stop+0x100;stub=stop+0x3000;head=0x81b018
get=lambda a:struct.unpack('<I',u.mem_read(a,4))[0]
put=lambda a,v:u.mem_write(a,struct.pack('<I',v&0xffffffff))
f32=lambda v:struct.unpack('<I',struct.pack('<f',v))[0]
ref=lambda rows,n:0 if n<0 else rows[n]
idx=lambda rows,a:-1 if a==0 else rows.index(a)
specs=[('reset',0x510e50,11),('update',0x510e60,366),('unlink',0x510fd0,79),('candidate',0x511020,214),('send',0x511100,326),('packet',0x511250,102),('round',0x419a30,62),('projection',0x57c8a0,251)]
entries_addr=[0x510e60,0x510fd0,0x511020,0x511100,0x511250,0x510e50]
regions=[dict(name=n,address=hex(a),size=z,sha256=sha(bytes(u.mem_read(a,z)))) for n,a,z in specs]
events=[];faults=[];trace={};entries=collections.Counter();current={};recording=False;applied=False;calls=collections.Counter();transcendental=[];profile_writes=0
profile=json.loads(args.trig_profile.read_text())['rows'] if args.trig_profile else None
def words(a,n):return [get(a+4*k) for k in range(n)]
def state():
    return dict(objects=[dict(flags=get(a+16),cls=get(a+8),pos=words(a+56,2),end=words(a+156,2),radius=get(a+176),update_ref=idx(data,get(a+748)),next=idx(objects,get(a+496)),prev=idx(objects,get(a+500)),health=get(a+556)!=0) for a in objects],data=[words(a,3) for a in data],head=idx(objects,get(head)),player=[get(player+3632),get(player+3636),struct.unpack('<H',u.mem_read(player+10,2))[0],struct.unpack('<H',u.mem_read(player+12,2))[0]],ray=words(ray,4),context_source=idx(objects,get(context)),context_ray=-1 if get(context+4)==0 else 0)
def event(name,values):events.append(dict(service=name,args=values,**state()))
def mutate(service,ctx=0,rp=0):
    global applied
    number=calls[service];calls[service]+=1
    if applied or current['mutation']!=service+':'+current['action'] or current['at']!=number:return
    applied=True;a=objects[0];action=current['action']
    if action=='position':put(a+56,f32(100.1));put(a+60,f32(-200.3))
    elif action=='angle':put(data[0],f32(0.7));put(data[0]+8,f32(-0.3))
    elif action=='update':put(a+748,data[1])
    elif action=='flags':put(objects[1]+16,0x10000)
    elif action=='candidate-position':put(objects[1]+56,f32(111.1));put(objects[1]+60,f32(52.7))
    elif action=='radius':put(objects[1]+176,f32(99.1))
    elif action=='source':put(ctx,objects[2])
    elif action=='ray':
        for n,v in enumerate(current['write_ray']):put(rp+4*n,v)
    elif action=='next':put(a+496,objects[2])
    elif action=='enabled':put(a+16,get(a+16)^0x1000000)
    else:raise AssertionError(action)
def ret(v=0):
    esp=u.reg_read(UC_X86_REG_ESP);u.reg_write(UC_X86_REG_EAX,v&0xffffffff);u.reg_write(UC_X86_REG_EIP,get(esp));u.reg_write(UC_X86_REG_ESP,esp+4)
def hook(_u,address,length,_):
    global profile_writes
    if address==stop:u.emu_stop();return
    if not recording:return
    if address in (0x510ee2,0x510f01):
        status=u.reg_read(UC_X86_REG_FPSW);top=(status>>11)&7
        register=globals()['UC_X86_REG_FP'+str(top)]
        sample=dict(instruction=hex(address-2),ST0=list(u.reg_read(register)),FPSW=status)
        if profile is not None:
            key=f"{u.reg_read(UC_X86_REG_FPCW):x}/{current['data'][0][0]:08x}/{1 if address==0x510ee2 else 0}"
            value=profile[key];sample['independent_profile']=value
            u.reg_write(register,(value['significand'],value['exponent']))
            u.reg_write(UC_X86_REG_FPSW,(status&~1024)|(value['FPSW']&1024));profile_writes+=1
        transcendental.append(sample)
    for n,a,z in specs:
        if a<=address<a+z:
            trace[address]=bytes(u.mem_read(address,length)).hex()
            if address==a:entries[n]+=1
            if address!=0x57c8a0:return
            break
    if stub<=address<stub+1024:return
    esp=u.reg_read(UC_X86_REG_ESP);arg=lambda n:get(esp+4*n)
    if address==0x535250:
        rp,hp=arg(1),arg(2);assert [arg(3),arg(4)]==[0,5]
        event('trace',words(rp,4)+[0,5]);mutate('trace',rp=rp)
        for n,v in enumerate(current['hit']):put(hp+4*n,v)
        ret(current['trace_result'])
    elif address==0x517c10:
        rp,cb,ctx=arg(1),arg(2),arg(3);assert cb==0x511020
        cp=get(ctx+4);event('each',words(rp,4)+[idx(objects,get(ctx))]+words(cp,4));mutate('each',ctx=ctx,rp=cp)
        code=bytearray()
        for candidate in current['candidates']:
            code+=b'\x68'+struct.pack('<I',ctx)+b'\x68'+struct.pack('<I',ref(objects,candidate))
            callsite=stub+len(code);code+=b'\xe8'+struct.pack('<i',cb-callsite-5)+b'\x83\xc4\x08'
        code+=b'\xc3';u.mem_write(stub,bytes(code));u.reg_write(UC_X86_REG_EIP,stub)
    elif address==0x40a5c0:
        assert arg(1)==4096;event('game',[4096]);mutate('game');ret(current['game_flags'])
    elif address==0x57c8a0:
        rp,pp,out=arg(1),arg(2),arg(3)
        event('project',([-1]*4 if rp==0 else words(rp,4))+[idx(objects,pp-56)]+words(pp,2));mutate('project',ctx=context,rp=rp)
        if not current['real_projection']:
            for n,v in enumerate(current['projected']):put(out+4*n,v)
            ret(current['project_result'])
    elif address==0x4ec580:
        event('parent',[idx(objects,arg(1))]);mutate('parent',ctx=context);ret(ref(objects,current['parent']))
    elif address==damage:
        event('damage',[idx(objects,arg(1)),idx(objects,arg(2)),idx(objects,arg(3)),arg(4),arg(5)]);mutate('damage',ctx=context);ret(17)
    elif address==0x501960:
        assert arg(1)==298 and [arg(3),arg(4)]==[0,0];event('audio',[298,idx(objects,arg(2)),0,0]);mutate('audio');ret()
    elif address==0x417090:
        event('player',[arg(1)]);mutate('player');ret(player if current['player_present'] else 0)
    elif address==0x40ebc0:
        assert arg(2)==1 and arg(4)==9
        event('packet',[arg(1),1]+list(u.mem_read(arg(3),9)));mutate('packet');ret(current['packet_result'])
    else:raise AssertionError((hex(address),current['name']))
def invalid(_u,access,address,length,value,_):faults.append(dict(address=address,size=length,instruction=u.reg_read(UC_X86_REG_EIP)));return False
u.hook_add(UC_HOOK_CODE,hook);u.hook_add(UC_HOOK_MEM_INVALID,invalid)
u.reg_write(UC_X86_REG_ESP,sp);u.reg_write(UC_X86_REG_FPCW,0x27f);u.emu_start(0x4014b5,0x4014c4,count=20000)
assert u.reg_read(UC_X86_REG_FPCW)==0xe3f
initial=u.context_save();recording=True;cases=[]
def add(kind,name,**kw):
    objs=[dict(flags=0x1000000,cls=0,pos=[f32(100.1),f32(50.7)],end=[f32(150.2),f32(50.7)],radius=f32(3.3),update_ref=0,next=-1,prev=-1,health=True) for _ in range(4)]
    objs[1]['pos']=[f32(125.3),f32(51.7)];objs[2]['flags']=0x80000000;objs[2]['next']=3;objs[3]['flags']=0x80000000;objs[3]['prev']=2
    c=dict(name=specs[[1,2,3,4,5,0][kind]][0]+'/'+name,kind=kind,objects=objs,data=[[f32(0.3),f32(0.7),f32(0.01)],[f32(1.3),f32(1.7),f32(0.11)]],head=2,player=[f32(110.3),f32(50.7),80,60],ray=[f32(100.1),f32(50.7),f32(150.2),f32(50.7)],context_source=0,context_ray=0,source_ref=0,context_present=True,player_present=True,index=2,trace_result=1,hit=[f32(140.7),f32(51.3)],candidates=[1],game_flags=4096,real_projection=False,project_result=1,projected=[f32(125.3),f32(50.7)],parent=3,packet_result=0x12345678,mutation='',action='',at=0,write_ray=[f32(30.7),f32(-50.3),f32(151.2),f32(51.7)])
    c.update(kw);cases.append(c);return c
for flags in (0,0x80000000,0x20,0x80000020,0x1000000,0x81000000,0x1000020,0x81000020):
    for h in (-1,2):
        c=add(0,f'flags/{flags:x}/head/{h}',head=h);c['objects'][0]['flags']=flags
for n in range(256):
    c=add(0,'angle/'+str(n));c['data'][0][0]=f32((n-128)*0.1)
for bits in (0,0x80000000,1,0x80000001,0x3f800001,0xbf800001,0x4b000001,0x4f000000,0x5effffff,0x5f000000,0x7f7fffff,0xff7fffff,0x7f800000,0xff800000,0x7fc12345):
    for field in range(5):
        c=add(0,f'word/{field}/{bits:x}')
        if field<3:c['data'][0][field]=bits
        else:c['objects'][0]['pos'][field-3]=bits
for result in (0,1,0x100,-1):add(0,'trace-result/'+str(result),trace_result=result)
for action in ('position','angle','update','ray','enabled'):add(0,'trace-mutate/'+action,mutation='trace:'+action,action=action)
for action in ('source','ray'):add(0,'each-mutate/'+action,mutation='each:'+action,action=action)
for pos in ('only','head','middle','tail','unmarked'):
    c=add(1,pos);c['objects'][0]['flags']=0x80000000;c['head']=0;c['objects'][0]['next']=-1
    if pos=='head':c['objects'][0]['next']=2;c['objects'][2]['prev']=0
    if pos=='middle':c['head']=2;c['objects'][2]['next']=0;c['objects'][0]['prev']=2;c['objects'][0]['next']=3;c['objects'][3]['prev']=0
    if pos=='tail':c['objects'][2]['next']=0;c['objects'][0]['prev']=2;c['head']=2
    if pos=='unmarked':c['objects'][0]['flags']=0
for flags in (0,1,0x40,0x41,0x10,0x8010,0x10000000):
    for cls in (0,2,0x1000002):
        for game in (0,1,0x1000):
            c=add(2,f'gate/{flags:x}/{cls:x}/{game:x}',source_ref=1,game_flags=game);c['objects'][1]['flags']=flags;c['objects'][1]['cls']=cls
for health_present in (False,True):
    for result in (0,1,0x100,-1):
        c=add(2,f'project-gate/{health_present}/{result}',source_ref=1,project_result=result);c['objects'][1]['health']=health_present
for bits in (0,0x80000000,1,0x3f800001,0x40000000,0x40400000,0x7f7fffff,0x7f800000,0x7fc12345):
    for field in range(5):
        c=add(2,f'distance/{field}/{bits:x}',source_ref=1)
        if field<2:c['objects'][1]['pos'][field]=bits
        elif field<4:c['projected'][field-2]=bits
        else:c['objects'][1]['radius']=bits
for action in ('position','candidate-position','radius','flags','source'):
    for service in ('project','parent','damage'):add(2,f'mutate/{service}/{action}',source_ref=1,mutation=service+':'+action,action=action)
for points in [(100.1,50.7,150.2,50.7),(100.1,50.7,150.2,75.3),(150.2,75.3,100.1,50.7),(100.1,50.7,100.1,50.7),(0.0,0.0,600.0,600.0)]:
    add(2,'real-projection/'+','.join(x.hex() for x in points),source_ref=1,real_projection=True,ray=[f32(x) for x in points])
for bits in (0,1,0x3f7fffff,0x3f800000,0x3f800001,0x40000000,0x7f7fffff,0x7f800000,0xff800000,0x7fc12345):
    c=add(2,f'real-radius/{bits:x}',source_ref=1,real_projection=True);c['objects'][1]['radius']=bits
for enabled in (False,True):
    for width in (0,1,20,80,65535):
        c=add(3,f'viewport/{enabled}/{width}',head=0);c['objects'][0]['flags']=0x80000000|(0x1000000 if enabled else 0);c['player'][2]=width;c['objects'][0]['next']=2
for ex in (80.3,150.2):
    for ey in (30.7,75.3):
        c=add(3,f'beam-order/{ex}/{ey}',head=0);c['objects'][0]['end']=[f32(ex),f32(ey)]
for bits in (0,0x80000000,1,0x42c80001,0xc2c80001,0x7f800000,0xff800000,0x7fc12345):
    for field in range(6):
        c=add(3,f'viewport-word/{field}/{bits:x}',head=0)
        if field<2:c['player'][field]=bits
        else:c['objects'][0]['pos' if field<4 else 'end'][field%2]=bits
for bits in (0,0x80000000,1,0x3f800001,0x42bf3333,0x42b4cccd,0x4affffff,0x4b000000,0x4f000000,0x7f800000,0xff800000,0x7fc12345):
    for field in range(4):
        c=add(4,f'coordinate/{field}/{bits:x}');c['objects'][0]['pos' if field<2 else 'end'][field%2]=bits
for result in (0,1,-1,0x100):add(4,'return/'+str(result),packet_result=result)
for kind in (3,4):
    c=add(kind,'existing-public-packet',head=0);c['objects'][0]['pos']=[f32(90.4),f32(95.6)];c['objects'][0]['end']=[f32(110.6),f32(105.4)];c['player']=[f32(100),f32(100),20,20]
    c['objects'][0]['next']=2;c['objects'][2]['prev']=0;c['objects'][2]['next']=-1;c['objects'][2]['update_ref']=1
for action in ('position','next','enabled'):add(3,'packet-mutate/'+action,head=0,mutation='packet:'+action,action=action)
add(5,'populated');add(5,'empty',head=-1)
def prepare(c,cw):
    global applied
    u.context_restore(initial);u.reg_write(UC_X86_REG_FPCW,cw);u.mem_write(heap,b'\xa5'*0x10000);u.mem_write(stack,bytes(0x10000));put(head,ref(objects,c['head']))
    for a,v in zip(objects,c['objects']):
        for off,val in [(8,v['cls']),(16,v['flags']),(176,v['radius']),(496,ref(objects,v['next'])),(500,ref(objects,v['prev'])),(556,health if v['health'] else 0),(716,damage),(748,ref(data,v['update_ref']))]:put(a+off,val)
        for off,values in ((56,v['pos']),(156,v['end'])):
            for k,value in enumerate(values):put(a+off+4*k,value)
    for a,values in zip(data,c['data']):
        for n,v in enumerate(values):put(a+4*n,v)
    for off,v in zip((3632,3636),c['player'][:2]):put(player+off,v)
    u.mem_write(player+10,struct.pack('<H',c['player'][2]));u.mem_write(player+12,struct.pack('<H',c['player'][3]))
    for n,v in enumerate(c['ray']):put(ray+4*n,v)
    put(context,ref(objects,c['context_source']));put(context+4,ray if c['context_ray']>=0 else 0)
    put(sp,stop);put(sp+4,c['index'] if c['kind'] in (3,4) else ref(objects,c['source_ref']))
    if c['kind']==2:put(sp+8,context if c['context_present'] else 0)
    if c['kind']==4:put(sp+8,ref(objects,c['source_ref']))
    u.reg_write(UC_X86_REG_ESP,sp);events.clear();faults.clear();calls.clear();transcendental.clear();applied=False
def preserved(before):
    expected=bytearray(before)
    for a in objects:
        for off,z in ((8,4),(16,4),(56,8),(156,8),(176,4),(496,8),(556,4),(748,4)):expected[a-heap+off:a-heap+off+z]=u.mem_read(a+off,z)
    for a in data:expected[a-heap:a-heap+12]=u.mem_read(a,12)
    for a,z in ((player+10,4),(player+3632,8),(context,8),(ray,16)):expected[a-heap:a-heap+z]=u.mem_read(a,z)
    return expected==bytes(u.mem_read(heap,0x10000))
def run(c,cw):
    prepare(c,cw);before=bytes(u.mem_read(heap,0x10000));error=None
    try:u.emu_start(entries_addr[c['kind']],stop,count=20000)
    except Exception as e:error=str(e)
    assert preserved(before)
    if not faults:assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4 and u.reg_read(UC_X86_REG_FPCW)==cw
    eax=u.reg_read(UC_X86_REG_EAX);result=idx(objects,eax) if c['kind']==1 and not faults else struct.unpack('<i',struct.pack('<I',eax))[0] if c['kind']==4 and not faults else 0
    return dict(input=copy.deepcopy(c),output=state(),events=copy.deepcopy(events),result=result,fault=bool(faults),faults=list(faults),error=error,nonoutput_bytes_preserved=True,transcendental=copy.deepcopy(transcendental))
rows=[];counterfactual=[]
for current in cases:
    actual=run(current,0xe3f);nearest=run(current,0x23f);assert not actual['fault'] and not nearest['fault'],current['name'];rows.append(actual);counterfactual.append(nearest)
fault_rows=[]
for kind in range(6):
    if kind not in (3,5):
        c=copy.deepcopy(cases[0]);c.update(kind=kind,name='nil-source/'+str(kind),source_ref=-1);current=c;fault_rows.append(run(c,0xe3f))
for name,changes in [('nil-update-disabled',{'kind':0}),('nil-update-enabled',{'kind':0}),('nil-context-rejected',{'kind':2,'context_present':False,'source_ref':1}),('nil-context-eligible',{'kind':2,'context_present':False,'source_ref':1}),('nil-ray',{'kind':2,'source_ref':1,'context_ray':-1,'real_projection':True}),('nil-player-empty-list',{'kind':3,'player_present':False,'head':-1})]:
    c=copy.deepcopy(cases[0]);c.update(changes,name=name)
    if name.startswith('nil-update'):c['objects'][0]['update_ref']=-1;c['objects'][0]['flags']=0 if name.endswith('disabled') else 0x1000000
    if name=='nil-context-rejected':c['objects'][1]['flags']=1
    current=c;fault_rows.append(run(c,0xe3f))
assert args.game.read_bytes()==blob
proof=dict(original_sha256=sha(blob),original_file_unchanged=True,original_startup_FPCW=0xe3f,counterfactual_FPCW=0x23f,contexts=len(rows),executions=2*len(rows),regions=regions,entries=dict(entries),rows=rows,counterfactual_rows=counterfactual,actual_vs_nearest_state_differences=sum(a['output']!=b['output'] for a,b in zip(rows,counterfactual)),actual_vs_nearest_trace_differences=sum(a['events']!=b['events'] for a,b in zip(rows,counterfactual)),executed_instructions=[dict(address=hex(a),bytes=b) for a,b in sorted(trace.items())],arithmetic_or_conversion_substitutions=profile_writes,independent_trig_profile_sha256=sha(args.trig_profile.read_bytes()) if profile else None,scope='Complete original callers, original scan callback, original coordinate helper and selected composed original projection calls. Original FSIN/FCOS opcodes execute. Without --trig-profile, all emulator results remain untouched. With that option, each post-instruction ST0 and C2 is explicitly replaced by the separately sealed independent Rosetta x87 profile; this is a profiled composition, not unchanged-Unicorn or physical x86 hardware evidence. Both profiles are retained independently. World services are declared. No full native original-game/world composition or general signaling-NaN claim.')
if profile is None:
    del proof['independent_trig_profile_sha256']
    proof['scope']='Complete original callers, original scan callback, original coordinate helper and selected composed original projection calls. Original inline FSIN/FCOS execute in Unicorn, whose transcendental implementation is emulator evidence, not a native x87 hardware accuracy oracle. World trace/scan/quest/player/parent/damage/audio/network services are declared. No full native original-game/world composition or general signaling-NaN claim.'
with args.proof.open('x') as f:json.dump(proof,f,separators=(',',':'))
with args.fixture.open('x') as f:json.dump({k:v for k,v in proof.items() if k not in ('executed_instructions','counterfactual_rows')},f,separators=(',',':'))
with args.fault_proof.open('x') as f:json.dump(dict(original_sha256=sha(blob),rows=fault_rows),f,separators=(',',':'))
print(json.dumps({k:proof[k] for k in ('contexts','executions','entries','actual_vs_nearest_state_differences','actual_vs_nearest_trace_differences')}))
