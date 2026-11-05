"""Independently describe the frozen raw three-entry initial PE32 heap.

No body or native implementation is executed. Native relocation and complete
consumers require separate verification before the raw native engines.
"""
from pathlib import Path
import argparse
import gzip
import hashlib
import json
import struct

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--inputs', type=Path, default=Path(__file__).with_name('marker_boulder_generator_independent_inputs_54e460.json.gz'))
parser.add_argument('--output', type=Path, required=True)
options = parser.parse_args()
encoded = options.inputs.read_bytes()
assert hashlib.sha256(encoded).hexdigest() == '33513f4d4be79a95526c408703bbe11ecb71bd7dfa69283c3d29f84d3dbf2cf8'
inputs = json.loads(gzip.decompress(encoded))
base = 0x21000000
roles = [base+x for x in (0x1000,0x2000,0x3000,0x4000,0x5000,0x6000,0x9000,0xa000)]
def pointer(role):
    return 0 if role == -1 else 0x41000000 if role == -2 else roles[role]
rows = []
for case in inputs:
    heap = bytearray((i*17+(i>>7)*29)&255 for i in range(0x10000))
    def write(address, value):
        struct.pack_into('<I',heap,address-base,value&0xffffffff)
    def read(address, width=4):
        if not base <= address <= base+len(heap)-width:
            return None
        return int.from_bytes(heap[address-base:address-base+width], 'little')
    for role in range(6):
        for offset, value in ((8,2),(16,0x11223344+role),(56,0x40e00000+role),(60,0x41000000+role),(520,pointer(1)),(748,pointer(7))):
            write(roles[role]+offset,value)
    write(roles[0]+748,pointer(case['source_update']))
    write(roles[0]+520,pointer(case['source_owner']))
    write(roles[1]+748,pointer(case['parent_update']))
    write(roles[1]+8,case['parent_class'])
    for index,target in enumerate(case['slots']):
        write(roles[6]+116+4*index,pointer(target))
    def object_fields(address):
        return {name:read(address+offset,width) for name,offset,width in (
            ('class_byte',8,1),('class_word',8,4),('flags_byte',16,1),('flags_word',16,4),
            ('x_bits',56,4),('y_bits',60,4),('vertical_bits',108,4),('mass_bits',116,4),('owner',520,4),('update',748,4))}
    argument = pointer(case['unit'])
    parent = pointer(case['owner_return'])
    parent_update = read(parent+748)
    rows.append(dict(name=case['name'],entry=case['entry'],argument=argument,
                     initial_heap_sha256=hashlib.sha256(heap).hexdigest(),
                     source_reads=object_fields(argument),parent_reads=object_fields(parent),
                     object_role_reads=[object_fields(address) for address in roles[:6]],
                     marker_slot_domains={str(role):[read(roles[role]+116+4*i) for i in range(4)] for role in (6,7)},
                     returned_parent_slot_reads=[read(parent_update+116+4*i) if parent_update is not None else None for i in range(4)],
                     initial_world=dict(index=case['index']&0xffffffff,divisor=case['divisor']&0xffffffff,fps=case['fps']&0xffffffff,frame=case['frame']&0xffffffff,quest=case['quest'],owner_return=case['owner_return'])))
assert len(rows) == 200
raw = json.dumps(rows,sort_keys=True,separators=(',',':')).encode()+b'\n'
packed=bytearray(gzip.compress(raw,compresslevel=9,mtime=0));packed[9]=0xff
with options.output.open('xb') as stream:stream.write(packed)
print(json.dumps(dict(cases=len(rows),sha256=hashlib.sha256(packed).hexdigest(),decoded_sha256=hashlib.sha256(raw).hexdigest(),native_source_used=False,function_execution_credit=False,native_relocation_credit=False)),flush=True)
