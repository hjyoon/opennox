#!/usr/bin/env python3
"""Run the complete six original getter bytes; no native implementation/result.

Prepared only. Eight independently chosen caller/data contexts, two inherited
FPU controls, no original helper/caller/setter execution. A root-reviewed seal
and --execute-reviewed-original are mandatory before any Unicorn import/run.
"""
from pathlib import Path
import argparse,gc,gzip,hashlib,json,struct

EXE_SHA='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
BODY_SHA='2171b26dc3b3aa41e6ec1720367d741f96c119ebcb5d114d8851f6e0c3f8ff34'
ENTRY,STACK,SP,STOP,DATA=0x409b80,0x21000000,0x21000800,0x22000100,0x5d5000

def sha(b):return hashlib.sha256(b).hexdigest()
def record(path):
 b=path.read_bytes();return {'path':str(path.resolve()),'bytes':len(b),'sha256':sha(b)}
def require(v,msg):
 if not v:raise ValueError(msg)
def write_json(path,value):
 with path.open('x') as f:json.dump(value,f,indent=2,sort_keys=True);f.write('\n')
def pe_bytes(raw,va,n):
 require(raw[:2]==b'MZ','MZ');pe=struct.unpack_from('<I',raw,0x3c)[0]
 require(raw[pe:pe+4]==b'PE\0\0','PE');opt=pe+24
 require(struct.unpack_from('<H',raw,opt)[0]==0x10b,'PE32')
 base=struct.unpack_from('<I',raw,opt+28)[0];table=opt+struct.unpack_from('<H',raw,pe+20)[0]
 for i in range(struct.unpack_from('<H',raw,pe+6)[0]):
  _,rva,size,off=struct.unpack_from('<4I',raw,table+40*i+8)
  delta=va-base-rva
  if 0<=delta and delta+n<=size:return raw[off+delta:off+delta+n]
 raise ValueError('requested actual file-backed PE extent')

def prepare(a):
 bindings={k:record(getattr(a,k)) for k in ('exe','manifest','inputs','main_manifest')}
 raw=a.exe.read_bytes();require(sha(raw)==EXE_SHA,'whole original EXE')
 body=pe_bytes(raw,ENTRY,6);require(sha(body)==BODY_SHA and body.hex()=='b810535d00c3','complete original six bytes')
 manifest=json.loads(a.manifest.read_bytes());require(manifest['file_sha256']==EXE_SHA and len(manifest['functions'])==1 and not manifest['data'],'minimal dedicated leaf seal')
 row=manifest['functions'][0];require(int(row['address'],16)==ENTRY and row['size']==6 and row['sha256']==BODY_SHA,'root tuple')
 inputs=json.loads(gzip.decompress(a.inputs.read_bytes()));require(inputs['schema']=='opennox-sub409b80-independent-inputs-v1','input schema')
 cases=inputs['original_accessor_cases'];require(len(cases)==8 and len({x['id'] for x in cases})==8,'eight independent contexts')
 require(inputs['FPU_controls']==[0xe3f,0x23f] and len(inputs['native_strings'])==6,'two FPU modes and separate native string cases')
 require({x['data_mapped'] for x in cases}=={False,True},'mapped/unmapped return-address contexts')
 for c in cases:
  require(set(c['registers'])=={'eax','ebx','ecx','edx','esi','edi','ebp'},'complete caller input registers')
  require(all(type(v)is int and 0<=v<=0xffffffff for v in c['registers'].values()),'DWORD caller inputs')
 seal=json.loads(a.execution_seal.read_bytes())
 require(seal.get('reviewed_before_execution')is True and seal.get('new_native_getter_executed')is False,'explicit root pre-execution review')
 require(seal['producer']==record(Path(__file__)) and seal['bindings']==bindings,'whole producer/input/PE/seal binding')
 return bindings,raw,body,inputs

def run_case(a,raw,body,case,cw):
 import unicorn as uni
 import unicorn.x86_const as x
 u=uni.Uc(uni.UC_ARCH_X86,uni.UC_MODE_32);handles=[]
 regions=[('code',ENTRY&~4095,pe_bytes(raw,ENTRY&~4095,4096)),
  ('stack',STACK,bytes(((i*17+case['seed'])^(i>>4))&255 for i in range(4096))),
  ('return-stop',STOP&~4095,b'\xcc'*4096)]
 if case['data_mapped']:
  data=bytearray(((i*29+case['seed'])^(i>>3))&255 for i in range(4096));prefix=bytes.fromhex(case['data_prefix_hex']);data[0x310:0x310+len(prefix)]=prefix;regions.append(('returned-data',DATA,bytes(data)))
 events=[];reads=[];writes=[];faults=[]
 regs={name:getattr(x,'UC_X86_REG_'+name.upper()) for name in ('eax','ebx','ecx','edx','esi','edi','ebp','esp','eip','eflags')}
 def registers():return {name:int(u.reg_read(reg)) for name,reg in regs.items()}
 def fpu():return {'control':u.reg_read(x.UC_X86_REG_FPCW),'status':u.reg_read(x.UC_X86_REG_FPSW),'tag':u.reg_read(x.UC_X86_REG_FPTAG),'mxcsr':u.reg_read(x.UC_X86_REG_MXCSR),'registers':[u.reg_read(getattr(x,'UC_X86_REG_FP'+str(i))) for i in range(8)]}
 def code(uc,pc,size,user):
  offset=pc-ENTRY;require((pc,size)in((ENTRY,5),(ENTRY+5,1)),'no padding/helper/caller instruction may execute')
  require(bytes(uc.mem_read(pc,size))==body[offset:offset+size],'actual unchanged original instruction bytes');events.append(pc)
 def read(uc,access,address,width,value,user):reads.append({'address':address,'width':width,'PC':uc.reg_read(x.UC_X86_REG_EIP)})
 def write(uc,access,address,width,value,user):writes.append({'address':address,'width':width,'value':value,'PC':uc.reg_read(x.UC_X86_REG_EIP)})
 def invalid(uc,access,address,width,value,user):faults.append({'access':access,'address':address,'width':width,'PC':uc.reg_read(x.UC_X86_REG_EIP)});return False
 try:
  for name,address,data in regions:u.mem_map(address,4096);u.mem_write(address,data)
  u.mem_write(SP,struct.pack('<I',STOP))
  for name,value in case['registers'].items():u.reg_write(regs[name],value)
  u.reg_write(regs['esp'],SP);u.reg_write(regs['eip'],ENTRY);u.reg_write(regs['eflags'],case['eflags']);u.reg_write(x.UC_X86_REG_FPCW,cw)
  before={name:bytes(u.mem_read(address,4096)) for name,address,data in regions};rb=registers();fb=fpu();caller=bytes(u.mem_read(SP,96))
  handles=[u.hook_add(uni.UC_HOOK_CODE,code),u.hook_add(uni.UC_HOOK_MEM_READ,read),u.hook_add(uni.UC_HOOK_MEM_WRITE,write),u.hook_add(uni.UC_HOOK_MEM_INVALID,invalid)]
  error=None
  try:u.emu_start(ENTRY,STOP,timeout=1_000_000,count=8)
  except Exception as exc:error=repr(exc)
  ra=registers();fa=fpu();memory=[]
  for name,address,data in regions:
   after=bytes(u.mem_read(address,4096));memory.append({'Name':name,'Address':address,'Before':before[name].hex(),'After':after.hex(),'SHA':sha(before[name])})
  returned=ra['eip']==STOP and error is None
  controls=(returned and ra['esp']==SP+4 and all(ra[k]==rb[k] for k in ('ebx','ecx','edx','esi','edi','ebp','eflags')) and fb==fa and fb['control']==cw and len(events)==2 and events==[ENTRY,ENTRY+5] and not writes and all(m['Before']==m['After'] for m in memory) and reads==[{'address':SP,'width':4,'PC':ENTRY+5}])
  return {'case_id':case['id'],'FPU_control':cw,'returned':returned,'controls_passed':controls,'error':error,'registers_before':rb,'registers_after':ra,'fpu_before':fb,'fpu_after':fa,'caller_before_hex':caller.hex(),'caller_after_hex':bytes(u.mem_read(SP,96)).hex(),'executed_instructions':events,'reads':reads,'writes':writes,'faults':faults,'memory':memory,'returned_data_was_mapped':case['data_mapped']}
 finally:
  for h in handles:u.hook_del(h)
  u=None;gc.collect()

def main():
 p=argparse.ArgumentParser(description=__doc__)
 for n in ('exe','manifest','inputs','main-manifest','execution-seal','output','proof'):p.add_argument('--'+n,type=Path,required=True)
 p.add_argument('--validate-inputs-only',action='store_true');p.add_argument('--execute-reviewed-original',action='store_true');a=p.parse_args()
 initial=record(Path(__file__));rows=[]
 try:
  bindings,raw,body,inputs=prepare(a)
  if a.validate_inputs_only:print(json.dumps({'validated':True,'actual_original_calls':0,'bindings':bindings}));return 0
  require(a.execute_reviewed_original,'explicit --execute-reviewed-original required')
  for cw in inputs['FPU_controls']:
   for case in inputs['original_accessor_cases']:rows.append(run_case(a,raw,body,case,cw))
  require(initial==record(Path(__file__)) and all(v==record(getattr(a,k)) for k,v in bindings.items()),'inputs/source unchanged after recording')
  result={'schema':'opennox-sub409b80-original-reference-v1','source':{'Exe':EXE_SHA,'Body':BODY_SHA,'Inputs':bindings['inputs']['sha256'],'producer':initial,'bindings':bindings},'rows':rows}
  encoded=(json.dumps(result,sort_keys=True,separators=(',',':'))+'\n').encode()
  with a.output.open('xb') as f:f.write(gzip.compress(encoded,mtime=0))
  proof={'schema':'opennox-sub409b80-original-proof-v1','qualified':len(rows)==16 and all(r['controls_passed'] for r in rows),'actual_calls':len(rows),'returned':sum(r['returned'] for r in rows),'source':result['source'],'output':record(a.output),'observed_return_words':sorted({r['registers_after']['eax'] for r in rows if r['returned']}),'original_body_bytes':6,'padding_bytes_executed':0,'original_helpers_or_setter_or_GUI_executed':0,'new_native_execution':0}
  write_json(a.proof,proof);print(json.dumps({'proof':record(a.proof),'qualified':proof['qualified'],'actual_calls':len(rows)}));return 0 if proof['qualified'] else 1
 except Exception as exc:
  failure={'schema':'opennox-sub409b80-original-failure-v1','error':repr(exc),'producer':initial,'completed_rows':rows,'actual_calls_completed':len(rows),'output_exists':a.output.exists(),'native_execution':0}
  target=a.proof if not a.proof.exists() else a.proof.with_name(a.proof.name+'.failure.json');write_json(target,failure);raise

if __name__=='__main__':raise SystemExit(main())
