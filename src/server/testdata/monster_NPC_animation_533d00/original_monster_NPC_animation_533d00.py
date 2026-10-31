"""Execute unchanged NPC animation parent, weapon/frame getters and selector.

Only dictionary lookup4E3830 is declared. It returns a valid type record or
nil and may apply declared input mutations. No predicate/index/cache result
is supplied. Full root/helper code and table coverage is measured separately.
"""
import argparse,collections,gzip,hashlib,itertools,json,struct
from pathlib import Path
from unicorn import Uc,UC_ARCH_X86,UC_MODE_32,UC_HOOK_CODE,UC_HOOK_MEM_WRITE,UC_HOOK_MEM_READ
from unicorn.x86_const import *

parser=argparse.ArgumentParser(description=__doc__)
for name in ['game','fixture','proof']:parser.add_argument('--'+name,type=Path,required=True)
parser.add_argument('--control',type=lambda s:int(s,0),choices=[0xe3f,0x23f],required=True)
args=parser.parse_args();stage=Path(__file__).resolve().parent
audit=json.loads((stage/'original-functions-and-tables-static-v1.json').read_text())
blob=args.game.read_bytes();sha=lambda b:hashlib.sha256(b).hexdigest()
assert sha(blob)==audit['original_sha256']=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe=struct.unpack_from('<I',blob,60)[0];opt=pe+24;base=struct.unpack_from('<I',blob,opt+28)[0]
tab=opt+struct.unpack_from('<H',blob,pe+20)[0]
u=Uc(UC_ARCH_X86,UC_MODE_32);u.mem_map(base,(struct.unpack_from('<I',blob,opt+56)[0]+4095)&~4095)
for i in range(struct.unpack_from('<H',blob,pe+6)[0]):
    _,rva,n,off=struct.unpack_from('<IIII',blob,tab+40*i+8)
    if n:u.mem_write(base+rva,blob[off:off+n])
instructions={int(r['address'],16):bytes.fromhex(r['bytes'])for f in audit['functions']for r in f['instructions']}
assert sum(map(len,instructions.values()))==911
for f in audit['functions']:assert sha(bytes(u.mem_read(int(f['entry'],16),f['size'])))==f['sha256']
for d in audit['data']:assert sha(bytes(u.mem_read(int(d['address'],16),d['size'])))==d['sha256']
names=['Mimic','CarnivorousPlant','Zombie','VileZombie'];cache_base=0x833e60
heap,stack,code=0x21000000,0x22000000,0x23000000
for a in [heap,stack,code]:u.mem_map(a,0x10000)
unit,update,other,target=heap+0x1000,heap+0x3000,heap+0x5000,heap+0x7000
animation_arrays={'entry':heap+0x8000,'other':heap+0x9000,'nil':0};scratch=0x833c30
weapon_table=0x5bbb10;frame_table=0x753430
scratch_writes=[];helper_entries=[];dynamic_table_reads=set()
records={name:heap+0xa000+0x100*i for i,name in enumerate(names)}
sp,stop,init,cw=stack+0xe000,code+0x100,code+0x200,code+0x300
u.mem_write(init,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',cw));u.mem_write(cw,struct.pack('<H',args.control))
rd=lambda a:struct.unpack('<I',u.mem_read(a,4))[0]
wr=lambda a,v:u.mem_write(a,struct.pack('<I',v&0xffffffff))
caches=lambda:[rd(cache_base+4*i)for i in range(4)]
saved={UC_X86_REG_EBX:0x12345678,UC_X86_REG_ESI:0x23456789,UC_X86_REG_EDI:0x3456789a,UC_X86_REG_EBP:0x456789ab}
active=False;current=None;visited=set();table_reads={d["name"]:set() for d in audit["data"]};calls=collections.Counter();predicate_entries=[];lookups=[];publications=[];stack_writes=[];mutations=[];call_index=0;lookup_counts=collections.Counter()

def mutate(effect):
    field,value=effect['field'],effect['value']
    if field=='unit_type':a,n,b=unit+4,2,struct.pack('<H',value&0xffff)
    elif field=='unit_class':a,n,b=unit+8,4,struct.pack('<I',value&0xffffffff)
    elif field=='live_update':a,n,b=unit+748,4,struct.pack('<I',other if value=='other'else update)
    elif field in ['entry_status','other_status']:
        a=(update if field=='entry_status'else other)+1440;n=4;b=struct.pack('<I',value&0xffffffff)
    elif field in ['entry_enemy','other_enemy']:
        a=(update if field=='entry_enemy'else other)+1196;n=4;b=struct.pack('<I',target if value else 0)
    elif field=='entry_action':a,n,b=update+552+24*current['head'],4,struct.pack('<I',value&0xffffffff)
    elif field=='entry_animation':a,n,b=update+476,4,struct.pack('<I',animation_arrays[value])
    elif field.startswith('cache_'):
        a=cache_base+4*int(field.removeprefix('cache_'));n=4;b=struct.pack('<I',value&0xffffffff)
    else:raise AssertionError(field)
    u.mem_write(a,b);mutations.append({'address':a,'bytes':b.hex(),'field':field,'value':value,'call':call_index})

def on_code(_u,pc,size,_):
    if not active:return
    if pc==0x4e3830:
        pointer=u.reg_read(UC_X86_REG_ESP);assert rd(pointer)==0x4e3aaa
        name=bytes(u.mem_read(rd(pointer+4),64)).split(b'\0',1)[0].decode('ascii');assert name in names
        lookup_counts[name]+=1
        event={'name':name,'call':call_index,'occurrence':lookup_counts[name],'caches_before':caches()}
        mutation=current.get('mutation')
        if mutation and mutation['name']==name and mutation.get('occurrence',1)==lookup_counts[name]:
            for effect in mutation['effects']:mutate(effect)
        found=current['found'][names.index(name)];index=current['lookup'][names.index(name)]
        event.update(found=found,index=index if found else None,caches_after_declared_mutations=caches())
        lookups.append(event)
        u.reg_write(UC_X86_REG_EAX,records[name]if found else 0)
        u.reg_write(UC_X86_REG_ECX,0x87654321);u.reg_write(UC_X86_REG_EDX,0x98765432)
        u.reg_write(UC_X86_REG_ESP,pointer+4);u.reg_write(UC_X86_REG_EIP,rd(pointer));return
    assert pc in instructions,(current['name'],hex(pc))
    assert bytes(u.mem_read(pc,size))==instructions[pc]
    visited.add(pc)
    if pc in [0x533d00,0x4fa280,0x4f9f90,0x533790,0x534840,0x534a10,0x534a40,0x4e3aa0]:calls[hex(pc)]+=1
    if pc in [0x4fa280,0x4f9f90]:helper_entries.append({'entry':hex(pc),'argument':rd(u.reg_read(UC_X86_REG_ESP)+4),'call':call_index})
    if pc in [0x534840,0x534a10,0x534a40]:
        predicate_entries.append({'name':{0x534840:'Mimic',0x534a10:'CarnivorousPlant',0x534a40:'ZombiePair'}[pc],'call':call_index,'caches_before':caches(),'unit_type':struct.unpack('<H',u.mem_read(unit+4,2))[0],'unit_class':rd(unit+8),'entry_status':rd(update+1440),'live_update':'other'if rd(unit+748)==other else 'entry'})

def on_write(_u,_access,a,n,value,_):
    if not active:return
    if stack<=a and a+n<=stack+0x10000:
        assert (sp-64<=a and a+n<=sp) or (sp+4<=a and a+n<=sp+8),(current['name'],hex(a),n)
        stack_writes.append((a,n,value));return
    if scratch<=a and a+n<=scratch+16:
        assert n in [1,4];scratch_writes.append({'offset':a-scratch,'size':n,'value':value&((1<<(n*8))-1),'call':call_index});return
    assert n==4 and a in [cache_base+4*i for i in range(4)],(current['name'],hex(a),n)
    publications.append({'index':(a-cache_base)//4,'value':value&0xffffffff,'call':call_index})

def on_read(_u,_access,a,n,_value,_):
    if active:
        for address in range(a,a+n):
            if weapon_table<=address<weapon_table+108 or frame_table<=address<frame_table+424:dynamic_table_reads.add(address)
        for d in audit['data']:
            start=int(d['address'],16);end=start+d['size']
            for address in range(max(a,start),min(a+n,end)):table_reads[d['name']].add(address)
u.hook_add(UC_HOOK_CODE,on_code);u.hook_add(UC_HOOK_MEM_WRITE,on_write);u.hook_add(UC_HOOK_MEM_READ,on_read)
defaults={'head':0,'action':0,'status':0,'enemy':False,'other_status':0xc4000,'other_enemy':True,'class':2,'flags':0x1000000,'type_index':117,'caches':[11,22,33,44],'lookup':[11,22,33,44],'found':[True]*4,'repeat':1,'subclass':0,'animation':'entry','equip':0,'attack':0,'table_profile':0}
cases=[]
def add(group,**changes):
    c=dict(defaults,**changes);c['name']=f'{len(cases):05d}/{group}';cases.append(c)
actions=list(range(39))+[61,0x7fffffff,0x80000000,0xffffffff]
statuses=[(0x4000 if a else 0)|(0x40000 if b else 0)|(0x80000 if c else 0)for a,b,c in itertools.product([0,1],repeat=3)]
for head,action,status,enemy,typ in itertools.product([-1,0,7,23],actions,statuses,[False,True],[117,11,22,33,44]):
    add('actions',head=head,action=action,status=status,enemy=enemy,type_index=typ)
for cls,action,status,typ in itertools.product([0,1,2,4,6,0x10000,0xffffffff],[0,7,30,37],[0,0xc4000],[11,22,33,117]):
    add('class',**{'class':cls},action=action,status=status,type_index=typ)
for cache,action,status,enemy in itertools.product([[11]*4,[11,11,11,11],[11,11,33,44],[11,22,11,11],[0x1000b]*4,[0xffffffff]*4],[0,7,30,37],[0,0x40000,0x80000],[False,True]):
    add('cache-alias',caches=cache,type_index=11,action=action,status=status,enemy=enemy)
for cache,typ,action,repeat,found in itertools.product([[0]*4,[0,22,33,44],[11,0,0,44],[11,22,0,44]],[0,11,22,33,44,0x7fff,0xffff],[0,7,30,37],[1,2],[False,True]):
    add('cache-lookup',caches=cache,type_index=typ,action=action,repeat=repeat,found=[found]*4)
effects=[{'field':'unit_type','value':11},{'field':'unit_type','value':22},{'field':'live_update','value':'other'},{'field':'unit_class','value':0},{'field':'entry_status','value':0xc4000},{'field':'entry_status','value':0},{'field':'entry_enemy','value':True},{'field':'entry_enemy','value':False},{'field':'entry_action','value':30},{'field':'cache_2','value':0xffffffff}]
for name,effect,action,status in itertools.product(names,effects,[0,7,30,37],[0,0xc4000]):
    cache=[11,22,33,44];cache[names.index(name)]=0
    if name=='VileZombie':cache[2]=0
    add('lookup-mutation',caches=cache,type_index=117,action=action,status=status,mutation={'name':name,'effects':[effect]})
unique=[];seen=set()
for case in cases:
    key=json.dumps({k:v for k,v in case.items()if k!='name'},sort_keys=True,separators=(',',':'))
    if key in seen:continue
    seen.add(key);case['name']=f"{len(unique):05d}/"+case['name'].split('/',1)[1];unique.append(case)
cases=unique
# All inherited selector contexts remain inputs to the parent. Some begin
# with nil animation storage, which must skip the selector and all lookups.
for i,c in enumerate(cases):c['animation']=['entry','other','nil'][i%3]
for head,action,equip,attack,profile in itertools.product([-1,0,7,23],actions,[0,3,4,1<<17,1<<26,1<<31,(1<<2)|(1<<17)|(1<<31)],[0,52],[0,1]):
    add('NPC-actions',head=head,action=action,equip=equip,attack=attack,table_profile=profile,subclass=0x10,repeat=2 if action==16 else 1)
for name,role in itertools.product(names,['entry','other','nil']):
    cache=[11,22,33,44];cache[names.index(name)]=0
    if name=='VileZombie':cache[2]=0
    add('animation-reload-after-lookup',caches=cache,animation='entry',mutation={'name':name,'effects':[{'field':'entry_animation','value':role}]})
seen=set()
for c in cases:
    key=json.dumps({k:v for k,v in c.items()if k!='name'},sort_keys=True,separators=(',',':'))
    assert key not in seen,c['name'];seen.add(key)
rows=[];initial_globals=bytes(u.mem_read(cache_base-64,144))
for current in cases:
    active=False;u.mem_write(heap,b'\xa5'*0x10000);u.mem_write(stack,b'\xcc'*0x10000)
    u.mem_write(cache_base-64,b'\x6b'*144)
    u.mem_write(unit+4,struct.pack('<H',current['type_index']));wr(unit+8,current['class']);wr(unit+12,current['subclass']);wr(unit+16,current['flags']);wr(unit+748,update)
    u.mem_write(scratch-32,b'\x6d'*80)
    profile=current['table_profile']
    frames=[[((i*257+3)&0xffffffff),((i*259+5)&0xffffffff)]if not profile else [[0xffffffff,0x80000000,0,256][i%4],[0x80000001,0xffffffff,511,1][i%4]]for i in range(53)]
    weapon=[(i*7+3)%53 for i in range(27)]
    for i,pair in enumerate(frames):
        for j,value in enumerate(pair):wr(frame_table+8*i+4*j,value)
    for i,value in enumerate(weapon):wr(weapon_table+4*i,value)
    tables_before=[bytes(u.mem_read(weapon_table,108)),bytes(u.mem_read(frame_table,424))]
    for role,address in animation_arrays.items():
        if address:u.mem_write(address,bytes((i+7*(role=='other'))%256 for i in range(256)))
    for record,status,enemy in [(update,current['status'],current['enemy']),(other,current['other_status'],current['other_enemy'])]:
        u.mem_write(record+544,struct.pack('<b',current['head']))
        wr(record+552+24*current['head'],current['action'])
        wr(record+476,animation_arrays[current['animation']]);wr(record+2056,current['equip']);u.mem_write(record+2068,bytes([current['attack']]))
        wr(record+1440,status);wr(record+1196,target if enemy else 0)
    for i,name in enumerate(names):wr(cache_base+4*i,current['caches'][i]);u.mem_write(records[name],struct.pack('<H',current['lookup'][i]))
    heap_before=bytearray(u.mem_read(heap,0x10000));global_before=bytearray(u.mem_read(cache_base-64,144))
    predicate_entries=[];lookups=[];publications=[];mutations=[];returns=[];lookup_counts=collections.Counter();scratch_writes=[];helper_entries=[];scratch_before=bytearray(u.mem_read(scratch-32,80))
    for call_index in range(current['repeat']):
        u.mem_write(stack,b'\xcc'*0x10000);wr(sp,stop);wr(sp+4,unit)
        stack_before=bytearray(u.mem_read(stack,0x10000));stack_writes=[]
        for r,value in saved.items():u.reg_write(r,value)
        u.reg_write(UC_X86_REG_ESP,sp);u.emu_start(init,init+8,count=4)
        fpu=[u.reg_read(r)for r in [UC_X86_REG_FPCW,UC_X86_REG_FPSW,UC_X86_REG_FPTAG]];assert fpu[0]==args.control
        active=True;u.emu_start(0x533d00,stop,count=2000);active=False
        assert u.reg_read(UC_X86_REG_EIP)==stop and u.reg_read(UC_X86_REG_ESP)==sp+4
        assert all(u.reg_read(r)==value for r,value in saved.items())
        assert [u.reg_read(r)for r in [UC_X86_REG_FPCW,UC_X86_REG_FPSW,UC_X86_REG_FPTAG]]==fpu
        for a,n,value in stack_writes:stack_before[a-stack:a-stack+n]=int(value).to_bytes(n,'little',signed=False)
        assert bytes(u.mem_read(stack,0x10000))==stack_before
        value=u.reg_read(UC_X86_REG_EAX)
        if current['subclass']&0x10:assert value==scratch;returns.append({'role':'scratch','offset':0})
        else:
            pointer=rd(update+476)
            if not value:returns.append({'role':'nil','offset':0})
            else:
                role=next(k for k,v in animation_arrays.items()if v==pointer)
                assert 0<=value-pointer<=240 and (value-pointer)%16==0
                returns.append({'role':role,'offset':value-pointer})
    for m in mutations:
        a=m['address'];b=bytes.fromhex(m['bytes'])
        if heap<=a<heap+0x10000:heap_before[a-heap:a-heap+len(b)]=b
    assert bytes(u.mem_read(heap,0x10000))==heap_before
    for event in scratch_writes:
        a=32+event['offset'];n=event['size'];scratch_before[a:a+n]=event['value'].to_bytes(n,'little')
    assert bytes(u.mem_read(scratch-32,80))==scratch_before
    assert [bytes(u.mem_read(weapon_table,108)),bytes(u.mem_read(frame_table,424))]==tables_before
    for i,value in enumerate(caches()):struct.pack_into('<I',global_before,64+4*i,value)
    assert bytes(u.mem_read(cache_base-64,144))==global_before
    for f in audit['functions']:assert sha(bytes(u.mem_read(int(f['entry'],16),f['size'])))==f['sha256']
    for d in audit['data']:assert sha(bytes(u.mem_read(int(d['address'],16),d['size'])))==d['sha256']
    rows.append({'name':current['name'],'input':{k:v for k,v in current.items()if k!='name'},'output':{'predicate_entries':predicate_entries,'helper_entries':helper_entries,'returns':returns,'scratch_after':bytes(u.mem_read(scratch,16)).hex(),'scratch_writes':scratch_writes,'entry_animation_after':next(k for k,v in animation_arrays.items()if v==rd(update+476)),'caches':caches(),'lookups':lookups,'publications':publications,'declared_mutations':[{k:v for k,v in m.items()if k not in ['address','bytes']}for m in mutations],'unit_type_after':struct.unpack('<H',u.mem_read(unit+4,2))[0],'unit_class_after':rd(unit+8),'live_update_after':'other'if rd(unit+748)==other else 'entry','entry_status_after':rd(update+1440),'other_status_after':rd(other+1440),'entry_enemy_after':bool(rd(update+1196)),'other_enemy_after':bool(rd(other+1196)),'entry_action_after':rd(update+552+24*current['head'])if current['head']>=0 else None,'complete_remaining_heap_stack_global_guards_code_tables_nonvolatile_and_FPU_preserved':True}})
u.mem_write(cache_base-64,initial_globals)
coverage={f['name']:sum(len(instructions[a])for a in visited if int(f['entry'],16)<=a<int(f['entry'],16)+f['size'])for f in audit['functions']}
assert coverage=={f['name']:f['size']for f in audit['functions']},coverage
assert {d['name']:len(table_reads[d['name']])for d in audit['data']}=={d['name']:d['size']for d in audit['data']}
fixture={'schema':'original-monster-NPC-animation-533D00-v1','original_sha256':sha(blob),'control_word':args.control,'entry':0x533d00,'functions':[{k:v for k,v in f.items()if k!='instructions'}for f in audit['functions']],'tables':audit['data'],'declared_services':['dictionary_type_record_by_name_4E3830; actual WORD helper4E3AA0 and all predicates execute'],'rows':rows}
raw=(json.dumps(fixture,sort_keys=True,separators=(',',':'))+'\n').encode()
with args.fixture.open('xb')as out:
    with gzip.GzipFile(filename='',mode='wb',fileobj=out,compresslevel=9,mtime=0)as z:z.write(raw)
assert sha(args.game.read_bytes())==sha(blob)
proof={'status':'passed','original_sha256':sha(blob),'control_word':hex(args.control),'contexts':len(rows),'root_calls':calls['0x533d00'],'NPC_contexts':sum(bool(r['input']['subclass']&0x10)for r in rows),'nonNPC_contexts':sum(not(r['input']['subclass']&0x10)for r in rows),'dynamic_input_table_bytes_read':len(dynamic_table_reads),'root_scratch_write_count':sum(len(r['output']['scratch_writes'])for r in rows),'executed_body_bytes':coverage,'executed_total_body_bytes':sum(coverage.values()),'root_tables_bytes_readonly_and_preserved':83,'actual_table_bytes_read':{d['name']:len(table_reads[d['name']])for d in audit['data']},'all_inputs_unique_before_freeze':True,'per_context_predicate_entry_trace_observed':True,'actual_function_entries':dict(calls),'declared_dictionary_only':True,'no_predicate_index_or_cache_result_substitution':True,'full_native_role_heap_stack_global_guards_code_tables_nonvolatile_and_FPU_preserved':True,'fixture_sha256':sha(args.fixture.read_bytes()),'raw_fixture_sha256':sha(raw),'raw_fixture_bytes':len(raw),'original_unchanged':True,'overall_porting_100_percent':False}
with args.proof.open('x')as f:json.dump(proof,f,indent=2);f.write('\n')
print(json.dumps(proof))
