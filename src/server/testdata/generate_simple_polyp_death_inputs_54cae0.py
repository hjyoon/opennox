"""Independent PE inputs for ImpEgg, Polyp and Potion; no native source is read."""
import argparse
import copy
import gzip
import hashlib
import json

from pathlib import Path

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument("--output", type=Path, required=True)
options = parser.parse_args()
rows = []
base = dict(entry="54cb10", unit=0, cloud=2, update=4, cache=0,
            name_result=73, fps=30, lifetime_bits=0x3e800000,
            flags=0x12345678, script={})


def add(name, **changes):
    row = copy.deepcopy(base)
    row.update(changes)
    row["name"] = name
    rows.append(row)


for value in [0, 1, 0x40, 0x80, 0x12345678, 0x80000000, 0xffffffff]:
    add(f"egg-flags-{value:08x}", entry="54cae0", flags=value)
for entry in ["54cae0", "54cb10", "54cbb0"]:
    for unit in [-4, -3, -2, -1, 0, 1, 2, 3]:
        add(f"{entry}-source-{unit}", entry=entry, unit=unit)
        if entry == "54cb10":
            add(f"polyp-source-{unit}-allocation-nil", unit=unit, cloud=-1)
    for nested in ["54cae0", "54cb10", "54cbb0"]:
        add(f"{entry}-audio-reentry-{nested}", entry=entry,
            script={"audio:0": {"reenter": [nested, 1]}})
    add(f"{entry}-audio-live-flags", entry=entry,
        script={"audio:0": {"flags": [0, 0xabcdef81]}})
for cache in [1, 73, 0x7fffffff, 0x80000000, 0xffffffff]:
    add(f"polyp-cache-{cache:08x}", cache=cache)
for value in [0, 1, 0x80000000, 0xffffffff]:
    add(f"polyp-name-return-{value:08x}", name_result=value)
for cloud in [-2, -1, 0, 1, 2, 3]:
    add(f"polyp-cloud-{cloud}", cloud=cloud)
for update in [-2, -1, 0, 1, 2, 3, 4, 5, 6]:
    add(f"polyp-update-{update}", update=update)
for unit in [-4, -3, -2, -1]:
    add(f"polyp-source-{unit}-bad-cloud", unit=unit, cloud=-2)
for value in [0, 0x80000000, 1, 0x007fffff, 0x00800000,
              0x3dcccccd, 0x3e800000, 0xbe800000, 0x3f800001,
              0x4effffff, 0x4f000000, 0xcf000000, 0xcf000001,
              0x7f7fffff, 0xff7fffff, 0x7f800000, 0xff800000,
              0x7fc01234]:
    add(f"polyp-lifetime-{value:08x}", lifetime_bits=value)
for fps in [0, 1, 2, 0x7fffffff, 0x80000000, 0xffffffff]:
    add(f"polyp-signed-fps-{fps:08x}", fps=fps)
for kind in ["name", "new", "create", "lifetime", "audio", "delete"]:
    add(f"polyp-live-source-at-{kind}", script={kind + ":0": {
        "position": [0, 0x447a0001, 0xc47a0001]}})
    add(f"polyp-live-state-at-{kind}", script={kind + ":0": {
        "cache": 0x87654321, "fps": 0xffffffff}})
    add(f"polyp-replace-update-at-{kind}", script={kind + ":0": {
        "update": [2, 5]}})
    add(f"polyp-null-update-at-{kind}", script={kind + ":0": {
        "update": [2, -1]}})
    for nested in ["54cae0", "54cb10", "54cbb0"]:
        add(f"polyp-reenter-{nested}-at-{kind}", script={kind + ":0": {
            "reenter": [nested, 1]}})
add("egg-same-object-reentry-and-live-flags", entry="54cae0",
    script={"audio:0": {"reenter": ["54cae0", 0]},
            "audio:1": {"flags": [0, 0xfeed0081]}})
add("polyp-name-outer-result-after-reentry", name_result=0,
    script={"name:0": {"reenter": ["54cb10", 1]},
            "name:1": {"result": 99}})
add("polyp-allocation-nil-after-reentry", cloud=-1,
    script={"new:0": {"reenter": ["54cb10", 1]},
            "new:1": {"result": 3}})
add("polyp-cached-update-replaced-during-nested-call", script={
    "create:0": {"reenter": ["54cbb0", 1]},
    "audio:0": {"update": [2, -1]}})
assert len({row["name"] for row in rows}) == len(rows)
decoded = json.dumps(rows, sort_keys=True, separators=(",", ":")).encode() + b"\n"
encoded = bytearray(gzip.compress(decoded, compresslevel=9, mtime=0))
encoded[9] = 0x13
with options.output.open("xb") as stream:
    stream.write(encoded)
print(json.dumps(dict(cases=len(rows), sha256=hashlib.sha256(encoded).hexdigest(),
                      decoded_sha256=hashlib.sha256(decoded).hexdigest(), native_source_used=False)))
