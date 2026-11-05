"""Independent complete initial heaps and native field/read correspondences."""
from pathlib import Path
import argparse
import gzip
import hashlib
import json
import struct

parser=argparse.ArgumentParser(description=__doc__)
parser.add_argument('--output',type=Path,required=True)
options=parser.parse_args()
stage=Path(__file__).resolve().parent
packed=(stage/'inversion_inventory_independent_inputs_4fa4f0.json.gz').read_bytes()
assert hashlib.sha256(packed).hexdigest()=='69c228a2eb2ef67ce89f8d2127766d0b209cb29d80433a3d43fda9ec6e515e1a'
cases=json.loads(gzip.decompress(packed));assert len(cases)==207
base,bad=0x21000000,0x41000000
objects=[base+n for n in [0x3000,0x5000,0x6000,0x7000,0x8000,0xd000,0xe000]]
arrays=[base+n for n in [0x4000,0x9000,0xa000]]
mods=[base+n for n in [0x1000,0x2000,0xb000,0xc000]]
pointer=lambda role,roles:0 if role==-1 else bad if role==-2 else roles[role]
rows=[]
for case in cases:
    memory=bytearray(((number*23+case['seed']*31)^(number>>5))&255 for number in range(0x10000))
    words,fields,reads=[],[],[]
    def put(address,value):
        struct.pack_into('<I',memory,address-base,value&0xffffffff)
        words.append(dict(address=address,value=value&0xffffffff))
    def word(address,width=4):
        return int.from_bytes(memory[address-base:address-base+width],'little') if base<=address<=base+0x10000-width else None
    for number,address in enumerate(objects):
        for offset,value in [(8,case['owner_classes'][number]),(16,0),(496,0),(504,0),
                             (508,pointer(case['owners'][number],objects)),(692,0)]:put(address+offset,value)
    put(objects[0]+504,pointer(case['first'],objects))
    for number,address in enumerate(objects[2:5]):
        for offset,value in [(8,case['classes'][number]),(16,case['flags'][number]),
                             (496,pointer(case['next'][number],objects)),(692,pointer(case['data'][number],arrays))]:put(address+offset,value)
    for number,address in enumerate(arrays):
        put(address,bad);put(address+4,mods[3])
        for index,role in enumerate(case['slots'][number]):put(address+8+4*index,pointer(role,mods))
    for number,address in enumerate(mods):
        put(address+88,{'none':0,'inversion':0x4e03d0,'grip':0x4e0480,'alien':bad}[case['functions'][number]])
        put(address+96,case['strengths'][number])
    def field(role,offset,name,kind):
        fields.append(dict(role=role,pe_address=role+offset,field=name,kind=kind,value=word(role+offset)))
    def probe(role,kind,offset,index=0,width=4):
        reads.append(dict(role=role,kind=kind,index=index,want=word(role+offset,width)))
    for address in objects:
        for offset,name,kind in [(8,'object_class','word'),(16,'object_flags','word'),
                                  (496,'inventory_next','pointer'),(504,'inventory_first','pointer'),
                                  (508,'object_owner','pointer'),(692,'object_init','pointer')]:
            field(address,offset,name,kind);probe(address,name,offset)
        probe(address,'class_byte',8,width=1)
    for address in arrays:
        for index in range(4):
            field(address,4*index,'slot'+str(index),'pointer');probe(address,'slot',4*index,index)
    for address in mods:
        field(address,88,'effect_callback','pointer');field(address,96,'effect_strength','word')
        probe(address,'effect_callback',88);probe(address,'effect_strength',96)
    argument,target=pointer(case['root'],objects),pointer(case['target'],objects)
    probe(argument,'inventory_first',504)
    probe(target,'object_owner',508);probe(target,'class_byte',8,width=1)
    first=word(argument+504)
    if first is not None:
        probe(first,'object_flags',16);probe(first,'object_class',8);probe(first,'object_init',692)
        data=word(first+692)
        if data is not None:
            for index in range(4):
                probe(data,'slot',4*index,index)
                mod=word(data+4*index)
                if index>=2 and mod is not None:
                    probe(mod,'effect_callback',88);probe(mod,'effect_strength',96)
    rows.append(dict(name=case['name'],entry='inventory',seed=case['seed'],argument=argument,target=target,
                     output=0,initializer_words=words,native_fields=fields,independent_reads=reads,
                     initial_heap_sha256=hashlib.sha256(memory).hexdigest()))
encoded=json.dumps(rows,sort_keys=True,separators=(',',':')).encode()+b'\n'
packed=gzip.compress(encoded,mtime=0)
with options.output.open('xb')as out:out.write(packed)
print(json.dumps(dict(cases=len(rows),encoded_sha256=hashlib.sha256(packed).hexdigest(),decoded_sha256=hashlib.sha256(encoded).hexdigest(),
                      independent_read_fault_domains=sum(read['want']is None for row in rows for read in row['independent_reads']),
                      original_or_native_function_executions=0,native_implementation_credit=False)),flush=True)
