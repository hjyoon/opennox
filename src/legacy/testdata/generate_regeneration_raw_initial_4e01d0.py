"""Independent initial bytes/field/read correspondence; no engine executions.

Only the separately chosen input file is read. Neither original outputs nor
any Go/C port implementation is imported or consulted.
"""
from pathlib import Path
import argparse,gzip,hashlib,json

parser=argparse.ArgumentParser(description=__doc__)
parser.add_argument('--output',type=Path,required=True)
options=parser.parse_args();root=Path(__file__).resolve().parent
data=(root/'regeneration_raw_independent_inputs_4e01d0.json.gz').read_bytes()
assert hashlib.sha256(data).hexdigest()=='e7a9dc92b938fd023aef24b062e14b0e7127259350ec67cd2bdf8b14c030b713'
inputs=json.loads(gzip.decompress(data));assert len(inputs)==198
base,bad=0x21000000,0x41000000
rows=[]
for case in inputs:
 def address(role):
  kind=case[role]
  if kind=='null':return 0
  if kind=='bad':return bad
  if kind=='last-holder':return base+65536-496
  if kind=='last-health':return base+65536-560
  if kind=='last-current':return base+65536-2
  if kind=='last-maximum':return base+65536-6
  if kind=='last-rate':return base+65536-112
  if kind=='item':return address('item')
  return base+dict(item=0x3000,owner=0x5000,health=0x7000,effect=0x1000)[role]
 roles={role:address(role) for role in ['item','owner','health','effect']}
 memory=bytearray(((i*37+case['seed']*19)^(i>>3))&255 for i in range(65536))
 specs=[('item',492,'pointer',roles['owner']),('item',8,'word',case['obj_class']),
        ('owner',556,'pointer',roles['health']),('owner',536,'word',case['last_damage']),
        ('owner',16,'word',case['owner_flags']),('effect',108,'word',case['rate']),
        ('health',0,'half',case['current_hp']),('health',4,'half',case['maximum_hp'])]
 fields,initializers,reads=[],[],[]
 for kind,(role,offset,typ,value) in enumerate(specs):
  pointer=roles[role];where=pointer+offset;width=2 if typ=='half' else 4
  if base<=where and where+width<=base+65536:
   memory[where-base:where-base+width]=value.to_bytes(width,'little')
   fields.append(dict(role=pointer,address=where,kind=kind,type=typ,pe_width=width,value=value))
   initializers.append(dict(address=where,width=width,value=value))
 for kind,(role,offset,typ,value) in enumerate(specs):
  where=roles[role]+offset;width=2 if typ=='half' else 4
  want=int.from_bytes(memory[where-base:where-base+width],'little') if base<=where and where+width<=base+65536 else None
  reads.append(dict(role=roles[role],kind=kind,want=want))
 rows.append(dict(name=case['name'],seed=case['seed'],roles=roles,shapes={role:case[role] for role in roles},
                  initializer_words=initializers,native_fields=fields,independent_reads=reads,
                  initial_heap_sha256=hashlib.sha256(memory).hexdigest()))
raw=json.dumps(rows,sort_keys=True,separators=(',',':')).encode()+b'\n'
encoded=gzip.compress(raw,mtime=0)
with options.output.open('xb') as stream:stream.write(encoded)
print(json.dumps(dict(cases=len(rows),encoded_sha256=hashlib.sha256(encoded).hexdigest(),decoded_sha256=hashlib.sha256(raw).hexdigest(),independent_read_fault_domains=sum(read['want'] is None for row in rows for read in row['independent_reads']),original_output_files_read=0,original_or_native_engines_executed=0)),flush=True)
