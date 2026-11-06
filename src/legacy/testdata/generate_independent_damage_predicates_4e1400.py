"""Choose raw damage-predicate input domains before native implementation.

This producer chooses only pointers, class/subclass words, flag bytes and
declared after-read input stimuli. It computes no predicate return, branch,
memory-fault prediction or native expected result, and imports no game code.
"""
from pathlib import Path
import argparse, gzip, hashlib, itertools, json

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--output', type=Path, required=True)
args = parser.parse_args()
rows = []

def add(entry, name, **changes):
    row = dict(entry=entry, name=entry+'/'+name, source='valid', weapon='valid',
               use='valid', source_class=0, source_subclass=0,
               weapon_class=0, weapon_subclass=0, use_flags=0, mutations=[])
    row.update(changes)
    row['seed'] = len(rows)+1
    rows.append(row)

WAND, WEAPON, MONSTER, PLAYER = 0x1000, 0x01000000, 2, 4
for pointer in ['valid', 'null', 'bad', 'last-class', 'last-subclass-byte']:
    for cls, sub in itertools.product([0, MONSTER, PLAYER, MONSTER|PLAYER], [0, 0x10]):
        add('attack', f'source-pointer/{pointer}/{cls:x}/{sub:x}', weapon='null',
            source=pointer, source_class=cls, source_subclass=sub)
for pointer, cls in itertools.product(
        ['bad', 'last-class', 'last-subclass', 'last-use-pointer', 'unaligned1', 'unaligned3'],
        [0, MONSTER, WAND, WEAPON]):
    add('attack', f'weapon-pointer/{pointer}/{cls:x}', source='bad', weapon=pointer,
        weapon_class=cls, weapon_subclass=0x10000, use_flags=2)
for pointer, flag, cls in itertools.product(['null', 'bad', 'last-flags-byte'], [0, 2], [WAND, WAND|MONSTER]):
    add('attack', f'use-pointer/{pointer}/{flag:x}/{cls:x}', source='bad',
        weapon_class=cls, weapon_subclass=0x10000, use=pointer, use_flags=flag)
classes = list(dict.fromkeys([0]+[1 << bit for bit in range(32)]+[
    WAND|WEAPON, WAND|MONSTER, WEAPON|MONSTER, WAND|WEAPON|MONSTER,
    WAND|PLAYER, WEAPON|PLAYER, 0x80000002, 0xffffffff]))
for cls in classes:
    add('attack', f'class/{cls:08x}', source='bad', weapon_class=cls,
        weapon_subclass=0x47f00fe, use_flags=0)
subclasses = list(dict.fromkeys([0]+[1 << bit for bit in range(32)]+[
    0xffffffff, 0x047f0000, 0x047f00fe, 0xfb80ffff, 0xfb80ff01]))
for cls, sub in itertools.product([WAND, WEAPON], subclasses):
    add('attack', f'subclass/{cls:08x}/{sub:08x}', source='bad',
        weapon_class=cls, weapon_subclass=sub, use_flags=0)
flags = list(dict.fromkeys([1 << bit for bit in range(8)]+[
    0, 1, 2, 3, 0x7f, 0x80, 0xfe, 0xff, 0x102, 0x80000002, 0xfffffffd, 0xffffffff]))
for cls, flag in itertools.product([WAND, WAND|MONSTER], flags):
    add('attack', f'wand-flags/{cls:08x}/{flag:08x}', source='bad',
        weapon_class=cls, weapon_subclass=0x10000, use_flags=flag)
for cls, sub in itertools.product([0, 2, 4, 6, 0x80000000, 0x80000002, 0xffffffff],
        [0, 1, 0x10, 0x11, 0x80, 0x90, 0x100, 0x110, 0x80000010, 0xffffffef, 0xffffffff]):
    add('attack', f'unit-byte/{cls:08x}/{sub:08x}', weapon='null',
        source_class=cls, source_subclass=sub)
for name, changes in [
    ('nonwand-protects-use', dict(weapon_class=0, use='bad')),
    ('wand-subclass-protects-use', dict(weapon_class=WAND, weapon_subclass=0, use='bad')),
    ('weapon-subclass-protects-use', dict(weapon_class=WEAPON, weapon_subclass=0, use='bad')),
    ('source-player-protects-subclass', dict(weapon='null', source='last-class', source_class=PLAYER)),
    ('source-other-protects-subclass', dict(weapon='null', source='last-class', source_class=0)),
    ('source-argument-aliases-weapon', dict(source='weapon', source_class=MONSTER, weapon_class=MONSTER)),
]:
    add('attack', name, **changes)

def after(role, field, value):
    return dict(when='after-read', role=role, field=field, occurrence=1, value=value)

for name, changes, stimulus in [
    ('cached-unit-class', dict(weapon='null', source_class=MONSTER, source_subclass=0x10), after('source', 'class', 0)),
    ('cached-unit-subclass-byte', dict(weapon='null', source_class=MONSTER, source_subclass=0x10), after('source', 'subclass', 0)),
    ('cached-weapon-class', dict(weapon_class=MONSTER), after('weapon', 'class', 0)),
    ('cached-wand-priority', dict(weapon_class=WAND|WEAPON, weapon_subclass=0x10000), after('weapon', 'class', WEAPON)),
    ('cached-wand-subclass-test', dict(weapon_class=WAND, weapon_subclass=0x10000), after('weapon', 'subclass', 0)),
    ('cached-use-pointer', dict(weapon_class=WAND, weapon_subclass=0x10000, use_flags=2), after('weapon', 'use', 'bad')),
    ('cached-flags-byte-test', dict(weapon_class=WAND, weapon_subclass=0x10000, use_flags=2), after('use', 'flags', 0)),
]:
    add('attack', 'stimulus/'+name, mutations=[stimulus], **changes)

for pointer, cls, sub in itertools.product(
        ['null', 'bad', 'valid', 'last-class', 'last-subclass', 'unaligned1'],
        [0, WAND, WEAPON, WEAPON|MONSTER], [0, 0x4000]):
    add('friendly', f'weapon-pointer/{pointer}/{cls:08x}/{sub:08x}',
        source='bad', weapon=pointer, weapon_class=cls, weapon_subclass=sub)
for cls, sub in itertools.product(classes, [0, 0x4000]):
    add('friendly', f'class/{cls:08x}/{sub:08x}', source='bad',
        weapon_class=cls, weapon_subclass=sub)
for sub in subclasses:
    add('friendly', f'subclass/{sub:08x}', source='bad',
        weapon_class=WEAPON, weapon_subclass=sub)
add('friendly', 'stimulus/cached-class', weapon_class=WEAPON, weapon_subclass=0x4000,
    mutations=[after('weapon', 'class', 0)])
add('friendly', 'stimulus/cached-subclass', weapon_class=WEAPON, weapon_subclass=0x4000,
    mutations=[after('weapon', 'subclass', 0)])

assert len({r['name'] for r in rows}) == len(rows)
data = json.dumps(rows, sort_keys=True, separators=(',', ':')).encode()+b'\n'
encoded = gzip.compress(data, mtime=0)
with args.output.open('xb') as out:
    out.write(encoded)
print(json.dumps(dict(cases=len(rows), entries={e:sum(r['entry']==e for r in rows) for e in ['attack', 'friendly']},
    encoded_sha256=hashlib.sha256(encoded).hexdigest(), decoded_sha256=hashlib.sha256(data).hexdigest(),
    expected_results_calculated=False, native_code_imported=False)), flush=True)
