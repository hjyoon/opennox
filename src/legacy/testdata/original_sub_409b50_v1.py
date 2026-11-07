#!/usr/bin/env python3
"""Record complete41B original inline REP setter; no substituted helpers.

DF=0 cdecl contexts only. Original-only guard cases are retained observations,
not native fault equivalence. No Go result or expected memory is consulted.
"""
from pathlib import Path
import argparse,gc,gzip,hashlib,json,struct
EXE_SHA='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
BODY_SHA='39f4f1088bbfd476b125556405887c8220aa5912fb70137f4620cc0ab6bf5ee4'
ENTRY,DST,SRC,STACK,SP,STOP=0x409b50,0x5d5310,0x23000000,0x22000000,0x22000800,0x24000100
OFFSETS={0:1,1:1,2:4,6:3,9:2,11:2,13:2,15:2,17:2,19:2,21:5,26:3,29:2,31:2,33:3,36:2,38:1,39:1,40:1}
def sha(b):return hashlib.sha256(b).hexdigest()
def require(v,msg):
 if not v:raise ValueError(msg)
def rec(p):
 b=p.read_bytes();return {'path':str(p.resolve()),'bytes':len(b),'sha256':sha(b)}
def json_write(p,j):
 with p.open('x')as f:json.dump(j,f,indent=2,sort_keys=True);f.write('\n')
def pe_bytes(b,va,n):
 require(b[:2]==b'MZ','MZ');pe=struct.unpack_from('<I',b,60)[0];require(b[pe:pe+4]==b'PE\0\0','PE');opt=pe+24;require(struct.unpack_from('<H',b,opt)[0]==0x10b,'PE32');base=struct.unpack_from('<I',b,opt+28)[0];tab=opt+struct.unpack_from('<H',b,pe+20)[0]
 for i in range(struct.unpack_from('<H',b,pe+6)[0]):
  rva,size,off=struct.unpack_from('<3I',b,tab+40*i+12);delta=va-base-rva
  if 0<=delta and delta+n<=size:return b[off+delta:off+delta+n]
 raise ValueError('file-backed PE extent')
def prepare(a):
 binds={k:rec(getattr(a,k))for k in ('exe','manifest','inputs','main_manifest')};b=a.exe.read_bytes();require(sha(b)==EXE_SHA,'wholePE');body=pe_bytes(b,ENTRY,41);require(sha(body)==BODY_SHA,'complete original41B')
 m=json.loads(a.manifest.read_bytes());r=m['functions'];require(m['file_sha256']==EXE_SHA and len(r)==1 and r[0]['size']==41 and int(r[0]['address'],16)==ENTRY and r[0]['sha256']==BODY_SHA and not m['data'],'dedicated41B root')
 inputs=json.loads(gzip.decompress(a.inputs.read_bytes()));require(inputs['schema']=='opennox-sub409b50-independent-inputs-v1' and inputs['DF']==0 and inputs['FPU_controls']==[0xe3f,0x23f] and len(inputs['cases'])==25,'fixed independent DF0 cases')
 require(len({c['id']for c in inputs['cases']})==25 and sum(c['native_consumer']for c in inputs['cases'])==20,'exact case domains')
 for c in inputs['cases']:require(c['eflags']&0x400==0,'DF0 actual cdecl domain')
 seal=json.loads(a.execution_seal.read_bytes());require(seal.get('reviewed_before_execution')is True and seal.get('new_native_setter_executed')is False and seal['producer']==rec(Path(__file__)) and seal['bindings']==binds,'root reviewed input/producer binding')
 return binds,b,body,inputs

def one(b,body,c,cw):
 import unicorn as U
 import unicorn.x86_const as X
 u=U.Uc(U.UC_ARCH_X86,U.UC_MODE_32);handles=[];regions=[('code',ENTRY&~4095,pe_bytes(b,ENTRY&~4095,4096)),('stack',STACK,bytes((i*17+c['seed'])&255 for i in range(4096))),('stop',STOP&~4095,b'\xcc'*4096)]
 data=bytearray(b'\xa5'*4096);source=bytearray(b'\x6d'*4096);payload=bytes.fromhex(c['bytes_hex'])
 if c['source']=='alias':
  pointer=DST+c['source_delta'];at=pointer-(DST&~4095);data[at:at+len(payload)]=payload
 elif c['source']=='separate':
  at=0x100+c['source_alignment_offset'];pointer=SRC+at;source[at:at+len(payload)]=payload
 elif c['source']=='source-end-guard':pointer=SRC+4096-len(payload);source[-len(payload):]=payload
 elif c['source']=='null':pointer=0
 elif c['source']=='unmapped':pointer=0x41000000
 else:raise ValueError('source domain')
 if c['destination']!='unmapped':regions.append(('destination',DST&~4095,bytes(data)))
 if c['source']in('separate','source-end-guard'):regions.append(('source',SRC,bytes(source)))
 regs={n:getattr(X,'UC_X86_REG_'+n.upper())for n in ('eax','ebx','ecx','edx','esi','edi','ebp','esp','eip','eflags')}
 def registers():return{n:int(u.reg_read(r))for n,r in regs.items()}
 def fpu():return{'control':u.reg_read(X.UC_X86_REG_FPCW),'status':u.reg_read(X.UC_X86_REG_FPSW),'tag':u.reg_read(X.UC_X86_REG_FPTAG),'mxcsr':u.reg_read(X.UC_X86_REG_MXCSR),'registers':[u.reg_read(getattr(X,'UC_X86_REG_FP'+str(i)))for i in range(8)]}
 events=[];instructions=[];pending=[];faults=[]
 def mapped(address,width):return any(base<=address and address+width<=base+len(raw)for _,base,raw in regions)
 def complete():
  while pending:
   e=pending.pop(0)
   if mapped(e['address'],e['width']):
    actual=bytes(u.mem_read(e['address'],e['width']));e['after_hex']=actual.hex();e['complete']=actual==int(e['value']).to_bytes(e['width'],'little')
   else:e['complete']=False
 def code(uc,pc,size,user):
  complete();off=pc-ENTRY;require(off in OFFSETS and OFFSETS[off]==size and bytes(uc.mem_read(pc,size))==body[off:off+size],'only complete sealed41B instructions execute');instructions.append(pc)
 def read(uc,access,address,width,value,user):
  complete()
  try:v=bytes(uc.mem_read(address,width)).hex()
  except U.UcError:v=None
  events.append({'kind':'read','PC':uc.reg_read(X.UC_X86_REG_EIP),'address':address,'width':width,'value_hex':v,'complete':v is not None})
 def write(uc,access,address,width,value,user):
  complete();e={'kind':'write','PC':uc.reg_read(X.UC_X86_REG_EIP),'address':address,'width':width,'value':value&((1<<(8*width))-1),'complete':False};events.append(e);pending.append(e)
 def invalid(uc,access,address,width,value,user):faults.append({'PC':uc.reg_read(X.UC_X86_REG_EIP),'access':access,'address':address,'width':width});return False
 try:
  for _,base,raw in regions:u.mem_map(base,4096);u.mem_write(base,raw)
  u.mem_write(SP,struct.pack('<II',STOP,pointer))
  for n,v in c['registers'].items():u.reg_write(regs[n],v)
  u.reg_write(regs['esp'],SP);u.reg_write(regs['eip'],ENTRY);u.reg_write(regs['eflags'],c['eflags']);u.reg_write(X.UC_X86_REG_FPCW,cw)
  before={n:bytes(u.mem_read(base,4096))for n,base,_ in regions};rb=registers();fb=fpu();caller=bytes(u.mem_read(SP,96))
  handles=[u.hook_add(U.UC_HOOK_CODE,code),u.hook_add(U.UC_HOOK_MEM_READ,read),u.hook_add(U.UC_HOOK_MEM_WRITE,write),u.hook_add(U.UC_HOOK_MEM_INVALID,invalid)]
  error=None
  try:u.emu_start(ENTRY,STOP,timeout=2_000_000,count=100_000)
  except Exception as exc:error=repr(exc)
  complete();ra=registers();fa=fpu();memory=[];replay_ok=True
  for name,base,_ in regions:
   actual=bytes(u.mem_read(base,4096));replay=bytearray(before[name])
   for e in events:
    if e['kind']=='write' and e['complete'] and base<=e['address'] and e['address']+e['width']<=base+4096:at=e['address']-base;replay[at:at+e['width']]=int(e['value']).to_bytes(e['width'],'little')
   replay_ok &= bytes(replay)==actual
   memory.append({'name':name,'address':base,'before_hex':before[name].hex(),'after_hex':actual.hex(),'before_sha256':sha(before[name]),'after_sha256':sha(actual)})
  returned=error is None and ra['eip']==STOP
  preserved=not returned or(ra['esp']==SP+4 and ra['ecx']==0 and all(ra[r]==rb[r]for r in ('ebx','ebp','edx','esi','edi')))
  valid=(returned or bool(faults)) and preserved and replay_ok and fb==fa and fb['control']==cw and caller==bytes(u.mem_read(SP,96))
  return{'case_id':c['id'],'FPU_control':cw,'source_argument':pointer,'returned':returned,'error':error,'actual_result_u32':ra['eax'],'certificate_valid':valid,'memory_replay_exact':bool(replay_ok),'registers_before':rb,'registers_after':ra,'fpu_before':fb,'fpu_after':fa,'caller_before_hex':caller.hex(),'caller_after_hex':bytes(u.mem_read(SP,96)).hex(),'executed_instructions':instructions,'events':events,'faults':faults,'memory':memory}
 finally:
  for h in handles:u.hook_del(h)
  u=None;gc.collect()

def main():
 p=argparse.ArgumentParser(description=__doc__)
 for n in ('exe','manifest','inputs','main-manifest','execution-seal','output','proof'):p.add_argument('--'+n,type=Path,required=True)
 p.add_argument('--validate-inputs-only',action='store_true');p.add_argument('--execute-reviewed-original',action='store_true');a=p.parse_args();rows=[];producer=rec(Path(__file__))
 try:
  binds,b,body,inputs=prepare(a)
  if a.validate_inputs_only:print(json.dumps({'validated':True,'original_calls':0,'bindings':binds}));return 0
  require(a.execute_reviewed_original,'explicit execute flag required')
  for cw in inputs['FPU_controls']:
   for case in inputs['cases']:rows.append(one(b,body,case,cw))
  require(all(rec(getattr(a,k))==v for k,v in binds.items())and rec(Path(__file__))==producer,'source/input binding after recording')
  result={'schema':'opennox-sub409b50-original-reference-v1','source':{'Exe':EXE_SHA,'Body':BODY_SHA,'Inputs':binds['inputs']['sha256'],'producer':producer,'bindings':binds},'rows':rows}
  with a.output.open('xb')as f:f.write(gzip.compress((json.dumps(result,sort_keys=True,separators=(',',':'))+'\n').encode(),mtime=0))
  covered={pc for r in rows for pc in r['executed_instructions']};proof={'schema':'opennox-sub409b50-original-proof-v1','qualified':len(rows)==50 and all(r['certificate_valid']for r in rows) and covered=={ENTRY+o for o in OFFSETS},'actual_calls':len(rows),'returned':sum(r['returned']for r in rows),'original_faults':sum(bool(r['faults'])for r in rows),'whole_original_bytes':41,'actual_callees':0,'padding_executed':0,'DF0_only':True,'new_native_executions':0,'source':result['source'],'output':rec(a.output)};json_write(a.proof,proof);print(json.dumps(proof));return 0 if proof['qualified']else 1
 except Exception as exc:
  target=a.proof if not a.proof.exists()else a.proof.with_name(a.proof.name+'.failure.json');json_write(target,{'qualified':False,'error':repr(exc),'producer':producer,'completed_rows':rows,'new_native_executions':0});raise
if __name__=='__main__':raise SystemExit(main())
