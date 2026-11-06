"""Choose raw Readiness inputs before any raw native implementation.

No branch, fault, return or native result is calculated. The callback at
004E0950 is an identity compared by the original root, never an invoked helper.
"""
from pathlib import Path
import argparse, gzip, hashlib, json

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--output', type=Path, required=True)
options = parser.parse_args()
rows = []


def add(name, **changes):
    row = dict(name=name, seed=len(rows)+1, item='valid', table='valid',
               first='valid', second='valid', replacement='valid',
               obj_class=0x01000000, slots=['bad', 'bad', 'first', 'second'],
               callbacks=dict(first=0x004E0950, second=0x004E0950, replacement=0),
               strengths=dict(first=37, second=91, replacement=0x89abcdef),
               mutations=[])
    row.update(changes)
    rows.append(row)


add('baseline')
for role, shapes in [
    ('item', ['null', 'bad', 'last-init', 'last-class']),
    ('table', ['null', 'bad', 'last-slot2', 'last-slot3']),
    ('first', ['null', 'bad', 'last-callback', 'last-strength']),
    ('second', ['null', 'bad', 'last-callback', 'last-strength']),
]:
    for shape in shapes:
        changes = {role: shape}
        if role == 'second':
            changes['slots'] = ['bad', 'bad', 'null', 'second']
        add('pointer/'+role+'/'+shape, **changes)

for bits in [0, 1, 2, 4, 0x1000, 0x01000000, 0x02000000, 0x10000000,
             0x13001000, 0x08000000, 0x20000000, 0x80000000, 0xffffffff]:
    add('class/'+format(bits, '08x'), obj_class=bits)

for slots in [
    ['bad', 'bad', 'null', 'null'], ['bad', 'bad', 'null', 'second'],
    ['bad', 'bad', 'first', 'null'], ['bad', 'bad', 'first', 'first'],
    ['bad', 'bad', 'second', 'first'], ['bad', 'bad', 'bad', 'second'],
    ['bad', 'bad', 'null', 'bad'], ['first', 'second', 'null', 'null'],
]:
    add('slots/'+'/'.join(slots), slots=slots)

for callback in [0, 1, 0x004E0950, 0x004E0951, 0x004E0940, 0x41000000, 0xffffffff]:
    for index in [2, 3]:
        slots = ['bad', 'bad', 'null', 'null']
        slots[index] = 'first'
        add('callback/'+str(index)+'/'+format(callback, '08x'), slots=slots,
            callbacks=dict(first=callback, second=0x004E0950, replacement=0))

for strength in [0, 1, 37, 0x7fffffff, 0x80000000, 0x80000001, 0xfffffffe, 0xffffffff]:
    for index in [2, 3]:
        slots = ['bad', 'bad', 'null', 'null']
        slots[index] = 'first'
        add('strength/'+str(index)+'/'+format(strength, '08x'), slots=slots,
            strengths=dict(first=strength, second=91, replacement=0x89abcdef))

for first, second in [(0, 0), (0, 0x004E0950), (0x004E0950, 0),
                      (0x004E0950, 0x004E0950), (0x004E0951, 0x004E0950),
                      (0x41000000, 0x004E0950)]:
    add('priority/'+format(first, '08x')+'/'+format(second, '08x'),
        callbacks=dict(first=first, second=second, replacement=0))

# Earlier admission must protect unreachable table/descriptor pages, but item
# InitData is read even when the class will reject the item.
for bits in [0, 4, 0x80000000]:
    for table in ['null', 'bad']:
        add('protected-table/'+format(bits, '08x')+'/'+table, obj_class=bits, table=table)
add('boundary/table-slot2-hit', table='last-slot2')
add('boundary/table-slot2-miss', table='last-slot2',
    callbacks=dict(first=0, second=0x004E0950, replacement=0))
add('boundary/first-callback-miss', first='last-callback',
    callbacks=dict(first=0, second=0x004E0950, replacement=0))
add('boundary/second-callback-miss', second='last-callback',
    slots=['bad', 'bad', 'null', 'second'],
    callbacks=dict(first=0, second=0, replacement=0))


def after(role, field, occurrence, value, target_role, target_field, index=None, target_index=None):
    row = dict(when='after-read', role=role, field=field, occurrence=occurrence,
               target_role=target_role, target_field=target_field, value=value)
    if index is not None:
        row['index'] = index
    if target_index is not None:
        row['target_index'] = target_index
    return row


for value in ['null', 'bad', 'alternate_table']:
    add('mutation/cached-initial-table/'+value, mutations=[
        after('item', 'init-data', 1, value, 'item', 'init-data')])
add('mutation/cached-class', mutations=[after('item', 'class', 1, 0, 'item', 'class')])
for index in [2, 3]:
    slots = ['bad', 'bad', 'null', 'null']
    slots[index] = 'first'
    for value in ['replacement', 'null', 'bad']:
        add('mutation/slot-reload/'+str(index)+'/'+value, slots=slots, mutations=[
            after('modifier', 'callback', 1, value, 'table', 'slot', target_index=index)])
        add('mutation/first-slot-load/'+str(index)+'/'+value, slots=slots, mutations=[
            after('table', 'slot', 1, value, 'table', 'slot', index=index, target_index=index)])
    add('mutation/cached-matching-callback/'+str(index), slots=slots, mutations=[
        after('modifier', 'callback', 1, 0, 'first', 'callback')])
    add('mutation/live-callback/'+str(index), slots=slots, mutations=[
        after('table', 'slot', 1, 0, 'first', 'callback', index=index)])
    add('mutation/cached-slot-reload/'+str(index), slots=slots, mutations=[
        after('table', 'slot', 2, 'bad', 'table', 'slot', index=index, target_index=index)])
    add('mutation/live-strength/'+str(index), slots=slots, mutations=[
        after('modifier', 'callback', 1, 0x87654321, 'first', 'strength')])
    add('mutation/cached-strength/'+str(index), slots=slots, mutations=[
        after('modifier', 'strength', 1, 0, 'first', 'strength')])
add('mutation/cached-slot-null', slots=['bad', 'bad', 'null', 'second'], mutations=[
    after('table', 'slot', 1, 'first', 'table', 'slot', index=2, target_index=2)])
add('mutation/cached-callback-miss', callbacks=dict(first=0, second=0x004E0950, replacement=0), mutations=[
    after('modifier', 'callback', 1, 0x004E0950, 'first', 'callback')])
add('mutation/reload-replacement-callback-ignored', mutations=[
    after('modifier', 'callback', 1, 'replacement', 'table', 'slot', target_index=2)])

assert len({row['name'] for row in rows}) == len(rows)
data = json.dumps(rows, sort_keys=True, separators=(',', ':')).encode()+b'\n'
encoded = gzip.compress(data, mtime=0)
with options.output.open('xb') as stream:
    stream.write(encoded)
print(json.dumps(dict(cases=len(rows), encoded_sha256=hashlib.sha256(encoded).hexdigest(),
                     decoded_sha256=hashlib.sha256(data).hexdigest(),
                     expected_results_calculated=False, helper_invocations_defined=0)), flush=True)
