"""Add read/write observations to the immutable original producer in memory.

The original producer, inputs and complete outputs stay byte-exact. Stack
scratch/caller words remain in the original golden; the sidecar describes
ordered external-memory accesses, including attempted faulting accesses.
"""
from pathlib import Path
import argparse
import gzip
import hashlib
import json
import sys

parser = argparse.ArgumentParser(description=__doc__, add_help=False)
parser.add_argument('--accesses', type=Path, required=True)
parser.add_argument('--accesses-proof', type=Path, required=True)
options, original_args = parser.parse_known_args()
original = Path(__file__).with_name('original_modifier_defend_predicates_4e03d0.py')
frozen = original.read_bytes()
expected = '82b09332665ebe1f81564d1cab1c411bac2ab757a4862f0efe45a21c55dcb252'
assert hashlib.sha256(frozen).hexdigest() == expected
source = frozen.decode()
def replace(old, new):
    global source
    assert source.count(old) == 1, old
    source = source.replace(old, new)
replace('UC_HOOK_MEM_WRITE, UC_HOOK_MEM_INVALID)',
        'UC_HOOK_MEM_WRITE, UC_HOOK_MEM_INVALID, UC_HOOK_MEM_READ)')
replace('results = []\n', 'results = []\naccesses, boundaries = [], []\n')
observers = '''def observe_read(uc, access, address, length, unused_value, unused):
    if not active or stack <= address < stack+0x10000:
        return
    try:
        value = int.from_bytes(uc.mem_read(address, length), 'little')
    except UcError:
        value = None
    accesses.append(dict(kind='read', address=hex(address), size=length,
                         value=value, complete=value is not None,
                         instruction=hex(uc.reg_read(UC_X86_REG_EIP))))
def observe_write(uc, address, length, value):
    if active and not stack <= address < stack+0x10000:
        accesses.append(dict(kind='write', address=hex(address), size=length,
                             value=value&((1<<(8*length))-1), complete=True,
                             instruction=hex(uc.reg_read(UC_X86_REG_EIP))))
def observe_invalid(uc, access, address, length, value):
    if not active or stack <= address < stack+0x10000:
        return
    assert access in (19,20,23,24)
    kind = 'read' if access in (19,23) else 'write'
    if accesses and accesses[-1]['address']==hex(address) and accesses[-1]['size']==length and accesses[-1]['kind']==kind:
        accesses[-1]['complete'] = False
        if kind=='read': accesses[-1]['value'] = None
    else:
        accesses.append(dict(kind=kind,address=hex(address),size=length,
                             value=None if kind=='read' else value&((1<<(8*length))-1),
                             complete=False,instruction=hex(uc.reg_read(UC_X86_REG_EIP))))
'''
replace('def code(uc, address, length, unused):\n', observers+'def code(uc, address, length, unused):\n')
replace('def write(uc, access, address, length, value, unused):\n',
        'def write(uc, access, address, length, value, unused):\n    observe_write(uc,address,length,value)\n')
replace('def invalid(uc, access, address, length, value, unused):\n',
        'def invalid(uc, access, address, length, value, unused):\n    observe_invalid(uc,access,address,length,value)\n')
replace('u.hook_add(UC_HOOK_MEM_INVALID, invalid)\n',
        'u.hook_add(UC_HOOK_MEM_INVALID, invalid)\nu.hook_add(UC_HOOK_MEM_READ, observe_read)\n')
replace('    writes, fault, calls = [], None, []\n',
        '    writes, fault, calls = [], None, []\n    accesses = []\n')
replace("    results.append(dict(name=case['name'], initial_heap=bytes(initial).hex(),\n",
        "    boundaries.append(dict(name=case['name'],accesses=accesses,entry_calls=calls,fault=fault))\n    results.append(dict(name=case['name'], initial_heap=bytes(initial).hex(),\n")
sys.argv = [str(original), *original_args]
namespace = dict(__file__=str(original), __name__='__main__')
exec(compile(source, str(original), 'exec'), namespace)
assert original.read_bytes() == frozen
assert len(namespace['boundaries']) == 298
output = namespace['args'].output
assert hashlib.sha256(output.read_bytes()).hexdigest() == 'f3cc6af63c5eb0a6b0ced5cf5b9af406cd72b14047562d67222a514020ce9f79'
encoded = json.dumps(namespace['boundaries'], sort_keys=True, separators=(',', ':')).encode()+b'\n'
packed = gzip.compress(encoded, mtime=0)
with options.accesses.open('xb') as stream:
    stream.write(packed)
proof = dict(original_producer_sha256=expected, original_producer_unchanged=True,
             complete_original_output_byte_identical=True, cases=298,
             FPU_control=namespace['args'].control,
             observed_accesses=sum(len(row['accesses']) for row in namespace['boundaries']),
             intentional_function_faults=sum(row['fault'] is not None for row in namespace['boundaries']),
             encoded_sha256=hashlib.sha256(packed).hexdigest(), decoded_sha256=hashlib.sha256(encoded).hexdigest(),
             physical_stack_projection_claim=False, native_implementation_credit=False)
with options.accesses_proof.open('x') as stream:
    json.dump(proof,stream,indent=2);stream.write('\n')
print(json.dumps(proof),flush=True)
