"""Four additive inputs selected from original static branch conditions only.

Read frozen independent inputs; never import a producer/native algorithm or read
original execution rows. No original or native execution. Canonical gzip only.
"""
from pathlib import Path
import argparse,copy,gzip,hashlib,json
OUT=Path(__file__).resolve().parent
FROZEN=OUT.parent/'input-freeze-v1'
OLD_GZIP=FROZEN/'chosen-independent-inputs-4e0b30-v1.json.gz'
OLD_GENERATOR=FROZEN/'generate-independent-inputs-4e0b30-v1.py'
OLD_GZIP_SHA='7eadc7638c3a9a20d7b01fa19ce05e5c5b15be98b678e923204223967c8ce7cd'
OLD_GENERATOR_SHA='f45e6fe09a38b44087656f7825530b24ebfe371a4401dcbd7d674276194674a7'
OLD_PLAIN_SHA='f0e5b6c18a402925390b707bbf945be89a19e2a60fb53648ce0183bc8b006688'
BASE='physical-type-00000000-damage-00000014'
sha=lambda b:hashlib.sha256(b).hexdigest()
def fact(path):
 b=path.read_bytes();return dict(path=str(path),bytes=len(b),sha256=sha(b))
def generate():
 raw=OLD_GZIP.read_bytes();assert sha(raw)==OLD_GZIP_SHA
 assert sha(OLD_GENERATOR.read_bytes())==OLD_GENERATOR_SHA
 plain=gzip.decompress(raw);assert sha(plain)==OLD_PLAIN_SHA
 old=json.loads(plain);assert old['case_count']==len(old['cases'])==1034
 assert old['inputs_only']is True and old['execution_performed']is False
 base=[c for c in old['cases']if c['name']==BASE];assert len(base)==1;base=base[0]
 assert base['args']==[base['roles']['target'],base['roles']['source'],0,20,0]
 selected=[('source-nonnil-type0-damage20','source',20,0),
           ('source-nil-type0-damage20',0,20,0),
           ('source-alias-target-type15-damage0','target',0,15),
           ('source-nonnil-type5-damage0-control','source',0,5)]
 cases=[]
 for index,(name,source,damage,typ)in enumerate(selected,1):
  c=copy.deepcopy(base);c['name']='parent4e0b30-additive-shield-nilweapon-v1/'+name
  c['family']='additive-shield-nilweapon-v1'
  c['args']=[c['roles']['target'],c['roles'][source]if isinstance(source,str)else source,0,damage,typ]
  c['services']['004e11c2']['responses'][0]['eax']=1
  c['after_reads']=[];c['after_services']=[]
  c['domain']['notes'].append('Additive independent nil-weapon/Shield inputs only; cloned frozen input declaration, no original rows or native outcomes read.')
  assert c['name']not in {r['name']for r in old['cases']}
  cases.append(c)
 payload={k:copy.deepcopy(v)for k,v in old.items()if k not in ['cases','families','case_count']}
 payload.update(case_count=4,families={'additive-shield-nilweapon-v1':4},cases=cases)
 data=(json.dumps(payload,separators=(',',':'),ensure_ascii=True)+'\n').encode()
 assert sha(OLD_GZIP.read_bytes())==OLD_GZIP_SHA and sha(OLD_GENERATOR.read_bytes())==OLD_GENERATOR_SHA
 return old,base,cases,data,gzip.compress(data,compresslevel=9,mtime=0)
def main():
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('--check-only',action='store_true');args=p.parse_args()
 old,base,cases,data,compressed=generate()
 target=OUT/'chosen-additive-nilweapon-shield-inputs-v1.json.gz'
 if args.check_only:
  assert target.read_bytes()==compressed;print(json.dumps(dict(check_only=True,cases=4,old1034_exact=True,original_calls=0)));return
 with target.open('xb')as f:f.write(compressed)
 receipt=dict(schema='opennox-additive-parent-nilweapon-shield-input-generation-v1',generator=fact(Path(__file__).resolve()),
  frozen1034_input=fact(OLD_GZIP),frozen1034_generator=fact(OLD_GENERATOR),frozen1034_uncompressed_sha256=OLD_PLAIN_SHA,
  base_independent_input_name=BASE,base_input_sha256=sha(json.dumps(base,sort_keys=True,separators=(',',':')).encode()),
  output=fact(target),uncompressed_bytes=len(data),uncompressed_sha256=sha(data),case_count=4,
  names=[c['name']for c in cases],all_names_additive_and_distinct_from_frozen1034=True,
  inputs_only=True,no_original_rows_or_native_algorithm_read=True,original_execution_performed=False,
  old1034_input_and_generator_byte_exact=True,root_review_REQUIRED_before_any_extra_original_execution=True,
  overall100=False)
 with(OUT/'additive-input-generation-receipt-v1.json').open('x')as f:json.dump(receipt,f,indent=2);f.write('\n')
 print(json.dumps(dict(cases=4,output=str(target),gzip_sha256=sha(compressed),old1034_exact=True,original_calls=0)))
if __name__=='__main__':main()
