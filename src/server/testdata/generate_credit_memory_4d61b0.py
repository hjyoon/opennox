"""Describe original initial memory independently of any native implementation."""
from pathlib import Path
import argparse
import gzip
import hashlib
import json
import struct

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--output', type=Path, required=True)
options = parser.parse_args()
packed = (Path(__file__).resolve().parent/'credit_independent_inputs_4d61b0.json.gz').read_bytes()
assert hashlib.sha256(packed).hexdigest() == 'a086a617a8b90e951ac06183f77992dfb07f9c6b811757a36c668be3dff1e542'
inputs = json.loads(gzip.decompress(packed))
base = 0x21000000
unit, update, player = base+0x4000, base+0x9000, base+0x1000
rows = []
for case in inputs:
    memory = bytearray((i*17+(i>>7)*29)&255 for i in range(0x10000))
    chosen = {
        'none': unit if case['player'] == 'unit' else player,
        'counter-is-update-player': update+276-4668,
        'counter-is-unit-update': unit+748-4668,
        'counter-is-unit-flags': unit+16-4668,
        'player-flags-is-update-player': update+276-4692,
    }[case['alias']]
    def write(address, value):
        struct.pack_into('<I', memory, address-base, value&0xffffffff)
    def read(address, width=4):
        if not base <= address <= base+len(memory)-width:
            return None
        return int.from_bytes(memory[address-base:address-base+width], 'little')
    write(chosen+4668, case['counter'])
    write(chosen+4692, case['player_flags'])
    selected_player = {'normal':chosen, 'nil':0, 'unmapped':0x41000000, 'unit':unit}[case['player']]
    write(update+276, selected_player)
    selected_update = {'normal':update, 'nil':0, 'unmapped':0x41000000, 'unit':unit}[case['update']]
    write(unit+748, selected_update)
    write(unit+16, case['flags'])
    argument = {'normal':unit, 'nil':0, 'unmapped':0x41000000}[case['unit']]
    # These are independent domain reads, not simulated execution. In
    # particular destroyed flags do not remove invalid-field controls.
    loaded_update = read(argument+748) if argument else None
    loaded_player = read(loaded_update+276) if loaded_update is not None else None
    reads = dict(flags=read(argument+16, 1) if argument else None,
                 update=loaded_update, player=loaded_player,
                 counter=read(loaded_player+4668) if loaded_player is not None else None,
                 player_flags=read(loaded_player+4692) if loaded_player is not None else None)
    rows.append(dict(name=case['name'], chosen_player=chosen, argument=argument,
                     selected_update=selected_update, selected_player=selected_player,
                     initial_heap_sha256=hashlib.sha256(memory).hexdigest(), reads=reads))
assert len(rows) == 286
raw = json.dumps(rows, sort_keys=True, separators=(',', ':')).encode()+b'\n'
encoded = bytearray(gzip.compress(raw, compresslevel=9, mtime=0))
encoded[9] = 0xff
with options.output.open('xb') as stream:
    stream.write(encoded)
print(json.dumps(dict(cases=len(rows), sha256=hashlib.sha256(encoded).hexdigest(),
                      decoded_sha256=hashlib.sha256(raw).hexdigest(),
                      native_source_used=False, function_execution_credit=False)))
