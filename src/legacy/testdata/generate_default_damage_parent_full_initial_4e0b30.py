"""Package frozen INPUTS and independent initial descriptors; no algorithm/output.

This does not import/run original/native code, read original recording rows, or
modify the1034/additive4 sources. Existing initial scalar/region certificates
are copied exactly into a compact1038-case native-fixture test input.
"""
from pathlib import Path
import argparse
import gzip
import hashlib
import json

HERE = Path(__file__).resolve().parent
ART = HERE.parents[3]
def sha(data): return hashlib.sha256(data).hexdigest()
def read(path, expected):
    data = path.read_bytes(); assert sha(data) == expected
    return json.loads(gzip.decompress(data)),dict(path=str(path),bytes=len(data),sha256=sha(data))
def fact(path):
    data=path.read_bytes();return dict(path=str(path),bytes=len(data),sha256=sha(data))

def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('--check-only',action='store_true');args=p.parse_args()
    old,oldfact=read(ART/'input-freeze-v1/chosen-independent-inputs-4e0b30-v1.json.gz','7eadc7638c3a9a20d7b01fa19ce05e5c5b15be98b678e923204223967c8ce7cd')
    extra,extrafact=read(ART/'nilweapon-shield-additive-inputs-v1/chosen-additive-nilweapon-shield-inputs-v1.json.gz','e6a1f8280209ce548cee626568cd769ff369e47270f13a5615f090c8eccfa5fe')
    initial,initialfact=read(ART/'initial-memory-drafts-v1/initial-memory-projection-4e0b30-v1.json.gz','704bfc8dc0741bae3913723661103b4c34d55a395e79b550a170ea5373f9bf03')
    additive,additivefact=read(ART/'initial-memory-additive4-drafts-v1/additive4-initial-memory-projection-v1.json.gz','6c5113c87539061163ce2293c3d5602c875007273d61c37895f8f80617d89a16')
    inputs=old['cases']+extra['cases'];descriptors={c['name']:c for c in initial['cases']+additive['cases']}
    assert len(inputs)==len(descriptors)==1038 and len(initial['fields'])==27
    cases=[]
    for case in inputs:
        d=descriptors[case['name']]
        fields=[]
        for role in d['roles']:
            for f in role['selected_fields']:
                fields.append(dict(role=role['role'],field=f['field_id'],hex=f['initial_accessible_storage_hex']))
        assert sum(len(initial['fields'][f['field']]['PE_probe_widths'])for f in fields)==153
        cases.append(dict(input=case,regions={r['name']:r['initial_sha256']for r in d['regions']},fields=fields,
                          reference_input_sha256=sha(json.dumps(case,sort_keys=True,separators=(',',':')).encode())))
    data=dict(schema='opennox-default-damage-parent-full-initial-fixture-v1',cases=cases,case_count=1038,
              fields=initial['fields'],declared_probes=158814,roles_per_case=14,
              sources=dict(base_inputs=oldfact,additive_inputs=extrafact,base_initial=initialfact,additive_initial=additivefact),
              original_or_native_outputs_read=False,parent_algorithm_evaluated=False,
              partial_pointer_policy='Incomplete PE DWORD pointer bytes are opaque literal prefix bytes; no complete identity is invented. Full native uintptr read remains inaccessible.')
    raw=(json.dumps(data,separators=(',',':'),ensure_ascii=True)+'\n').encode()
    target=HERE/'default_damage_parent_full_initial_4e0b30.json.gz'
    zipped=gzip.compress(raw,compresslevel=9,mtime=0)
    if args.check_only:
        assert target.read_bytes()==zipped
        print(json.dumps(dict(cases=1038,probes=158814,regenerated_equal=True,writes=False)));return
    target.write_bytes(zipped)
    receipt=dict(schema='opennox-default-damage-parent-full-initial-fixture-generation-v1',generator=fact(Path(__file__).resolve()),
                 output=fact(target),uncompressed_bytes=len(raw),uncompressed_sha256=sha(raw),cases=1038,probes=158814,
                 immutable_sources=data['sources'],original_outputs_read=False,actual_native_execution=False)
    (HERE/'full-initial-fixture-generation-receipt-v1.json').write_text(json.dumps(receipt,indent=2)+'\n')
    print(json.dumps(receipt))

if __name__=='__main__':main()
