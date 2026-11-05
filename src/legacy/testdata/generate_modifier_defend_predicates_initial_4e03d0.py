"""Describe initial memory independently of all predicate/traversal bodies."""
from pathlib import Path
import argparse
import gzip
import hashlib
import json
import struct

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--output', type=Path, required=True)
args = parser.parse_args()
stage = Path(__file__).resolve().parent
packed = (stage/'modifier_defend_predicates_independent_inputs_4e03d0.json.gz').read_bytes()
assert hashlib.sha256(packed).hexdigest() == 'adc9b7435ce18b9006cf77e7e84057d18b1ed42272bda5432dd81d99b99ad5c7'
cases = json.loads(gzip.decompress(packed))
base, bad = 0x21000000, 0x41000000
effect0, effect1, item, data, output = [base+n for n in [0x1000,0x2000,0x3000,0x4000,0x5000]]
pointer = lambda kind, good:{'valid':good,'null':0,'bad':bad}[kind]
rows = []
for case in cases:
    memory = bytearray(((number*37+case['seed']*19)^(number>>3))&255 for number in range(0x10000))
    writes, slots, reads = [], [], []
    def initialize(address, value):
        struct.pack_into('<I',memory,address-base,value&0xffffffff)
        writes.append(dict(address=address,value=value&0xffffffff))
    def read(address):
        return struct.unpack_from('<I',memory,address-base)[0] if base<=address<=base+0x10000-4 else None
    def field(role, offset, name, kind):
        record = dict(role=role,pe_address=role+offset,field=name,kind=kind,value=read(role+offset))
        assert record['value'] is not None
        if not any(record['pe_address']==old['pe_address'] for old in slots):
            slots.append(record)
    def probe(role, kind, offset, index=0):
        reads.append(dict(role=role,kind=kind,index=index,want=read(role+offset)))
    if case['entry']!='scan':
        argument = base+0x10000-100 if case['effect']=='last-field' else pointer(case['effect'],effect0)
        if base<=argument<base+0x10000:
            initialize(argument+96,case['strength'])
            field(argument,4,'effect_index','word')
            field(argument,96,'effect_strength','word')
        out = {'separate':output,'field':effect0+96,'header':effect0+4,
               'unaligned':output+1,'null':0,'bad':bad}[case['output']]
        probe(argument,'strength',96)
        probe(out,'word',0)
    else:
        argument = pointer(case['item'],item)
        out = 0
        initialize(item+8,case['obj_class'])
        initialize(item+692,pointer(case['data'],data))
        initialize(data,bad);initialize(data+4,bad)
        for number,identity in enumerate(case['slots']):
            role = effect0 if case['alias'] or number==0 else effect1
            callbacks = {'grip':0x4e0480,'inversion':0x4e03d0,'null-fn':0,'alien-fn':bad}
            if identity in callbacks:
                initialize(role+88,callbacks[identity]);initialize(role+96,case['strengths'][number])
            initialize(data+8+4*number,0 if identity=='none' else bad if identity=='bad' else role)
        field(item,8,'object_class','word');field(item,692,'object_init','pointer')
        for index in range(4):
            field(data,4*index,'slot'+str(index),'pointer')
        for number,identity in enumerate(case['slots']):
            if identity in callbacks:
                role = effect0 if case['alias'] or number==0 else effect1
                field(role,88,'effect_callback','pointer');field(role,96,'effect_strength','word')
        probe(argument,'class',8);probe(argument,'init',692)
        array = pointer(case['data'],data)
        for index in range(4):
            probe(array,'slot',4*index,index)
            mod = read(array+4*index)
            if index>=2 and mod is not None:
                probe(mod,'callback',88);probe(mod,'strength',96)
    rows.append(dict(name=case['name'],entry=case['entry'],seed=case['seed'],argument=argument,
                     output=out,opaque_args=case['opaque_args'],initializer_words=writes,native_fields=slots,
                     independent_reads=reads,initial_heap_sha256=hashlib.sha256(memory).hexdigest()))
assert len(rows)==298
encoded = json.dumps(rows,sort_keys=True,separators=(',',':')).encode()+b'\n'
compressed = gzip.compress(encoded,mtime=0)
with args.output.open('xb') as stream:
    stream.write(compressed)
print(json.dumps(dict(cases=len(rows),encoded_sha256=hashlib.sha256(compressed).hexdigest(),
                      decoded_sha256=hashlib.sha256(encoded).hexdigest(),
                      independent_read_fault_domains=sum(read['want'] is None for row in rows for read in row['independent_reads']),
                      original_or_native_bodies_executed=0,native_engine_credit=False)),flush=True)
