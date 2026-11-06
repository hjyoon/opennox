"""Finite PE32 inputs only; no emulator, native algorithm, or expected outcome.

All arithmetic below constructs addresses, scalar bit patterns, or enumeration
indices. It never evaluates the original/native damage algorithm. Source lane
names describe selected inputs, not the path or result an execution must take.
"""
import argparse
import collections
import copy
import gzip
import hashlib
import json
from pathlib import Path
import struct

R = Path('/Users/ghost/codex-sandbox/worktrees/opennox-r344-20261020')
EXE = Path('/Users/ghost/codex-sandbox/nox-pjt/Nox/GAME.EXE')
ROOT = Path(__file__).resolve().parent
EXE_SHA = '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
SOURCE = '4303b76071b0b20d0bfa9dc67c70ad58516c77ea'
FRAME = 0x0084EA04
BAD = 0x21F00000
DAMAGE_CALLBACK = 0x23001000
SOUND_CALLBACK = 0x23002000
ROLE_NAMES = ['target', 'source', 'weapon', 'parent', 'owner', 'owner_alt',
              'target_update', 'source_update', 'alt_update', 'weapon_use',
              'alt_use', 'soundset', 'sound_resource', 'health']
OBJECTS = ROLE_NAMES[:6]
OFF = dict(type=4, cls=8, sub=12, flags=16, x=56, y=60, px=72,
           py=76, owner=508, attribution=520, damage_type=524,
           last_x=528, last_y=532, frame=536, health=556,
           damage_callback=716, sound_callback=720, use=736, update=748)
SITES = '''004e0b46 004e0b66 004e0bc2 004e0c05 004e0c16 004e0c2c
004e0c45 004e0c63 004e0ca5 004e0cd3 004e0cdb 004e0cec 004e0d18
004e0d9d 004e0dc5 004e0e06 004e0e2e 004e0f33 004e0f51 004e0f77
004e0fef 004e1019 004e102d 004e105c 004e1079 004e1097 004e109f
004e10b0 004e10de 004e1122 004e1131 004e115b 004e1168 004e117c
004e11aa 004e11b7 004e11c2 004e11f2 004e1212 004e0d05 004e1049
004e104f'''.split()
X87_SITES = ['004e0cec', '004e0d9d', '004e0e06', '004e10b0']
# Exact little-endian IEEE payloads, never host floating-point calculations.
F64 = dict(zero='0000000000000000', negative_zero='0000000000000080',
           quarter='000000000000d03f', half='000000000000e03f',
           one='000000000000f03f', one_and_half='000000000000f83f',
           minus_half='000000000000e0bf', above_one='010000000000f03f',
           infinity='000000000000f07f', quiet_nan='010000000000f87f')
U32_VALUES = [0, 1, 19, 20, 0x7FFFFFFF, 0x80000000, 0xFFFFFFFF]


def u32(value):
    return struct.pack('<I', value & 0xFFFFFFFF).hex()


def response(value=0xDEC0ADDE, **kw):
    return dict(eax=value & 0xFFFFFFFF, effects=[], **kw)


class Case:
    def __init__(self, family, name):
        self.family, self.name = family, name
        self.values = {role: {} for role in ROLE_NAMES}
        self.shapes = {}
        self.arg_roles = ['target', 'source', 'weapon']
        self.damage, self.damage_type, self.frame = 27, 0, 0x10203040
        self.services = {s: {'responses': [response()], 'repeat_last': True} for s in SITES}
        for s in ['004e0b46', '004e0bc2', '004e0c16', '004e0c45',
                  '004e0ca5', '004e1019', '004e1079', '004e1168', '004e11c2']:
            self.ret(s, 0)
        for s in ['004e0c05', '004e0c2c', '004e0c63', '004e11aa', '004e109f']:
            self.ret(s, 1)
        self.ret('004e102d', 0)
        for s in X87_SITES:
            self.fp(s, 'zero')
        self.fp('004e0cec', 'one_and_half')
        self.fp('004e10b0', 'half')
        self.after_reads, self.after_services = [], []
        self.notes = []

    def field(self, role, field, value, width=4):
        offset = OFF.get(field, field)
        self.values[role][offset] = (width, value)
        return self

    def ret(self, site, value):
        self.services[site]['responses'][0]['eax'] = value
        return self

    def fp(self, site, name):
        self.services[site]['responses'][0]['x87'] = {'format': 'f64', 'hex': F64[name]}
        return self

    def effect(self, site, arg_index, value):
        self.services[site]['responses'][0]['effects'].append(
            {'arg_index': arg_index, 'offset': 0, 'hex': u32(value)})
        return self

    def shape(self, role, readable=None, unaligned=0, unreadable=False):
        self.shapes[role] = dict(readable=readable, unaligned=unaligned, unreadable=unreadable)
        return self

    def stimulus(self, when, site, trigger_role, offset, width, effects, occurrence=1):
        # Effects tuples are (role, offset, integer_or_role, width).
        row = dict(occurrence=occurrence, _effects=effects)
        if when == 'read':
            row.update(pc=site, _trigger_role=trigger_role, _offset=offset, width=width)
            self.after_reads.append(row)
        else:
            row['site'] = site
            self.after_services.append(row)
        return self

    def freeze(self, seed):
        roles, regions = {}, []
        for i, role in enumerate(ROLE_NAMES):
            address = 0x21000000 + i * 0x2000
            shape = self.shapes.get(role, {})
            if shape.get('unreadable'):
                roles[role] = BAD + i * 0x2000
                continue
            readable = shape.get('readable')
            roles[role] = address + (0x1000 - readable if readable is not None
                                      else 0x100 + shape.get('unaligned', 0))
            regions.append(dict(name=role, address=address, size=0x1000,
                                fill=(0x51 + 13 * i + seed % 31) & 0xFF,
                                permissions='rw', initializers=[]))
        region_map = {x['name']: x for x in regions}
        default_values = {}
        for i, role in enumerate(OBJECTS):
            default_values[role] = {
                4: (2, 0x8100 + i), 8: (4, 0), 12: (4, 0), 16: (4, 0),
                56: (4, 0x3FC00000 + i * 0x800000),
                60: (4, 0xC0200000 + i * 0x800000),
                72: (4, 0xAA550000 + i), 76: (4, 0x55AA0000 + i),
                508: (4, 0), 520: (4, 0x10203040), 524: (4, 0x89ABCDEF),
                528: (4, 0x13579BDF), 532: (4, 0x2468ACE0),
                536: (4, 0xA1B2C3D4), 556: (4, 0),
                716: (4, DAMAGE_CALLBACK), 720: (4, 0),
                736: (4, 'weapon_use'),
                748: (4, 'target_update' if role == 'target' else 'source_update')}
        for role in ['target_update', 'source_update', 'alt_update']:
            default_values[role] = {88: (1, 0), 160: (2, 0x55AA),
                                    1440: (4, 0x13572468), 2094: (1, 0xAB),
                                    2184: (4, 0xDEADBEEF), 2188: (4, 0xC0FFEE00)}
        default_values.update(weapon_use={96: (4, 0xA1B2C300)},
                              alt_use={96: (4, 0x4D5E6F02)},
                              soundset={32: (4, 'sound_resource')},
                              sound_resource={0: (4, 0xDEAD0101)},
                              health={0: (2, 37), 4: (2, 80)})
        for role in ROLE_NAMES:
            values = dict(default_values[role])
            values.update(self.values[role])
            if role not in region_map:
                continue
            region, base = region_map[role], roles[role]
            for offset, (width, value) in sorted(values.items()):
                if isinstance(value, str):
                    value = roles[value]
                data = int(value).to_bytes(width, 'little', signed=False)
                # Preserve exact partial domain: initialize only accessible bytes.
                start = base + offset - region['address']
                available = max(0, min(len(data), region['size'] - start))
                if available:
                    region['initializers'].append({'offset': start, 'hex': data[:available].hex()})
        def effects(rows):
            result = []
            for role, offset, value, width in rows:
                if isinstance(value, str):
                    value = roles[value]
                result.append(dict(address=roles[role] + offset,
                                   hex=int(value).to_bytes(width, 'little').hex()))
            return result
        stimuli_read, stimuli_service = [], []
        for row in self.after_reads:
            stimuli_read.append(dict(pc=row['pc'], address=roles[row['_trigger_role']] + row['_offset'],
                                     width=row['width'], occurrence=row['occurrence'],
                                     effects=effects(row['_effects'])))
        for row in self.after_services:
            stimuli_service.append(dict(site=row['site'], occurrence=row['occurrence'],
                                        effects=effects(row['_effects'])))
        services = copy.deepcopy(self.services)
        for schedule in services.values():
            for item in schedule['responses']:
                if isinstance(item['eax'], str):
                    item['eax'] = roles[item['eax']]
        return dict(name=self.name, family=self.family, seed=seed,
                    args=[roles[x] if isinstance(x, str) else x for x in self.arg_roles]
                         + [self.damage, self.damage_type], roles=roles, regions=regions,
                    image_initializers=[dict(address=FRAME, hex=u32(self.frame))],
                    services=services, after_reads=stimuli_read, after_services=stimuli_service,
                    domain=dict(shapes=self.shapes, notes=self.notes))


def selected_cases():
    cases = []
    def add(family, name):
        c = Case(family, family + '-' + name)
        cases.append(c)
        return c
    for source, weapon in [(0, 0), ('source', 0), (0, 'weapon'), ('source', 'weapon')]:
        c = add('early', f'nil-target-{source}-{weapon}'); c.arg_roles = [0, source, weapon]
    for value in [1, 0x100, 0x80000000, 0xFFFFFFFF]:
        for frame in [0, 1, 3, 4, 0x100, 0xFFFFFFFF]:
            c = add('early', f'buff23-{value:08x}-frame-{frame:08x}')
            c.ret('004e0b46', value); c.frame = frame
            c.shape('target', unreadable=True)
    for cls in [0, 2, 4]:
        for zombie in [0, 1, 0x80000000]:
            for source, weapon in [(0, 0), ('source', 0), ('source', 'weapon')]:
                c = add('early', f'dead-class-{cls}-zombie-{zombie:08x}-{source}-{weapon}')
                c.field('target', 'cls', cls).field('target', 'flags', 0x8000).ret('004e0bc2', zombie)
                c.arg_roles = ['target', source, weapon]; c.damage_type = 14
    for parent in [0, 'parent', 'target']:
        for cls in [0, 2, 4, 6]:
            for flags in [0, 1]:
                c = add('quest', f'parent-{parent}-class-{cls}-flag4096-{flags}')
                c.ret('004e0c05', 0).ret('004e0c16', parent).ret('004e0c2c', 0).ret('004e0c45', flags)
                c.field(parent if parent else 'parent', 'cls', cls)
    for cls in [0, 2, 4]:
        for source in [0, 'source', BAD]:
            for weapon in [0, 'weapon', BAD + 0x2000]:
                for enemy in [0, 1]:
                    c = add('no-update', f'class-{cls}-source-{source}-weapon-{weapon}-enemy-{enemy}')
                    c.field('target', 'cls', cls).field('target', 'flags', 2).ret('004e0c63', enemy)
                    c.field('weapon', 'cls', 2)
                    c.arg_roles = ['target', source, weapon]
    for typ in list(range(19)) + [0xFFFFFFFF, 0x80000000]:
        for damage in U32_VALUES:
            c = add('physical', f'type-{typ:08x}-damage-{damage:08x}')
            c.damage_type, c.damage = typ, damage
            c.arg_roles[2] = 0
    for cls in [0, 2, 4, 6, 0x100, 0x01000000, 0x80000000]:
        for typ in [0, 2, 7, 9, 10, 17]:
            c = add('class-width', f'class-{cls:08x}-type-{typ}')
            c.field('target', 'cls', cls); c.damage_type = typ
    for cls, sub in [(0, 0), (2, 0), (0x1000, 0), (0x1000, 0x10000),
                     (0x1000, 0x47F0000), (0x01000000, 0),
                     (0x01000000, 0x4000), (0x01000000, 2)]:
        for target_cls in [0, 2, 4]:
            for enemy in [0, 1]:
                c = add('children', f'weapon-{cls:08x}-{sub:08x}-target-{target_cls}-enemy-{enemy}')
                c.field('target', 'cls', target_cls).field('weapon', 'cls', cls).field('weapon', 'sub', sub)
                c.ret('004e0c63', enemy); c.field('weapon_use', 96, 0x89ABCD02)
    for source_cls in [0, 2, 4, 6]:
        for weapon_cls in [0, 2, 0x1000, 0x01000000]:
            for shock in ['half', 'one_and_half', 'quiet_nan']:
                c = add('shock', f'source-{source_cls}-weapon-{weapon_cls:08x}-payload-{shock}')
                c.field('source', 'cls', source_cls).field('weapon', 'cls', weapon_cls)
                c.ret('004e0ca5', 1).fp('004e0cec', shock)
    for callback in [0, BAD]:
        c = add('shock', f'damage-callback-{callback:08x}')
        c.field('source', 'cls', 4).field('weapon', 'cls', 2)
        c.field('source', 'damage_callback', callback).ret('004e0ca5', 1)
    for subclass in [0, 0x200, 0x400, 0x800, 0xC00, 0xE10]:
        for typ in [1, 5, 7, 9, 12, 17]:
            c = add('immunity', f'subclass-{subclass:08x}-type-{typ}')
            c.field('target', 'cls', 2).field('target', 'sub', subclass); c.damage_type = typ
            c.damage = 0xFFFFFFFD
    for typ in [1, 7, 9, 12, 17]:
        for payload in F64:
            for damage in [0, 3, 0x80000000]:
                c = add('elemental', f'type-{typ}-payload-{payload}-damage-{damage:08x}')
                c.damage_type, c.damage = typ, damage
                c.fp('004e0d9d', payload).fp('004e0e06', payload)
                c.arg_roles[1:] = [0, 0]
    for typ in [1, 9]:
        for frame in [0, 1, 3, 4, 0xFFFFFFFF]:
            c = add('clock', f'type-{typ}-frame-{frame:08x}'); c.damage_type, c.frame = typ, frame
            c.fp('004e0d9d', 'half').fp('004e0e06', 'half')
    for source, weapon in [(0, 'weapon'), ('source', 0), ('source', 'weapon')]:
        for weapon_cls in [0, 0x1000, 0x01000000]:
            for typ in [0, 2, 10, 12]:
                c = add('attribution', f'{source}-{weapon}-class-{weapon_cls:08x}-type-{typ}')
                c.arg_roles = ['target', source, weapon]; c.damage_type = typ
                c.field('target', 'cls', 2).field('weapon', 'cls', weapon_cls)
    for target_cls, target_sub in [(4, 0), (2, 0x10), (2, 0)]:
        for weapon_cls in [0, 0x1000, 0x01000000]:
            for value in [0, 20, 0xFFFFFFFF]:
                c = add('damage-services', f'class-{target_cls}-sub-{target_sub}-weapon-{weapon_cls:08x}-write-{value:08x}')
                c.field('target', 'cls', target_cls).field('target', 'sub', target_sub).field('weapon', 'cls', weapon_cls)
                c.effect('004e0f77', 3, value).effect('004e0fef', 3, value)
    for callback in [0, SOUND_CALLBACK, BAD]:
        for soundset in [0, 'soundset']:
            for count in [0xFFFFFFFF, 0, 1, 2]:
                c = add('sound', f'callback-{callback:08x}-soundset-{soundset}-count-{count:08x}')
                c.field('source', 'cls', 2).field('target', 'sound_callback', callback)
                c.ret('004e1019', soundset).ret('004e102d', count)
    for callback in [0, SOUND_CALLBACK, BAD]:
        for source in [0, 'source']:
            c = add('sound', f'nil-weapon-source-{source}-callback-{callback:08x}')
            c.arg_roles = ['target', source, 0]
            c.field('target', 'sound_callback', callback)
    for power in [0, 1, 2, 255, 0x12340001]:
        for coefficient in ['zero', 'half', 'minus_half', 'infinity', 'quiet_nan']:
            c = add('vampirism', f'power-{power:08x}-coefficient-{coefficient}')
            c.field('target', 'cls', 4).ret('004e1079', 1).ret('004e109f', power).fp('004e10b0', coefficient)
            c.damage = 0x7FFFFFFF
    for damage in [0xFFFF, 0x10000, 0x10001, 0x1FFFF, 0xFFFF0001]:
        c = add('vampirism', f'full-conversion-word-damage-{damage:08x}')
        c.field('target', 'cls', 4).ret('004e1079', 1).fp('004e10b0', 'one')
        c.damage = damage
    for state in [0, 1, 15, 30, 255]:
        for damage in [19, 20, 0xFFFFFFFF]:
            c = add('hurt-state', f'state-{state}-damage-{damage:08x}')
            c.field('target', 'cls', 4).field('target_update', 88, state, 1); c.damage = damage
    for flags in [0, 1, 2, 0xFFFFFFFF]:
        for value in [0, 17, 0x80000000]:
            c = add('field-guide', f'flags-{flags:08x}-write-{value:08x}')
            c.ret('004e1168', flags).effect('004e117c', 2, value)
    for source_cls in [0, 2, 4]:
        for owner in [0, 'owner', 'target']:
            for enemy in [0, 1]:
                c = add('owner', f'source-{source_cls}-owner-{owner}-enemy-{enemy}')
                c.field('source', 'cls', source_cls).field('source', 'owner', owner).field('owner', 'cls', 2)
                c.ret('004e11aa', enemy)
    for typ in [0, 5, 15, 0xFFFFFFFF]:
        for source in ['source', 'target', 0]:
            for value in [0, 1, 0xFFFFFFFF]:
                c = add('shield', f'type-{typ:08x}-source-{source}-write-{value:08x}')
                c.damage_type = typ; c.arg_roles[1] = source
                c.ret('004e11c2', 1).effect('004e11f2', 1, value)
    for damage in [0, 1, 0xFFFFFFFF]:
        for membership in [0, 1]:
            c = add('shield', f'self-type15-initial-{damage:08x}-buff26-{membership}')
            c.arg_roles[1] = 'target'; c.damage, c.damage_type = damage, 15
            c.ret('004e11c2', membership)
    c = add('shield', 'type5-initial-zero-buff26-one')
    c.damage, c.damage_type = 0, 5; c.ret('004e11c2', 1)
    # All five set partitions of three argument identities plus nil combinations.
    for args in [('target', 'source', 'weapon'), ('target', 'target', 'weapon'),
                 ('target', 'source', 'target'), ('target', 'source', 'source'),
                 ('target', 'target', 'target'), ('target', 0, 'target'),
                 ('target', 'target', 0), ('target', 0, 0)]:
        for cls in [0, 2, 4, 0x1000]:
            c = add('aliases', '-'.join(map(str, args)) + f'-class-{cls:08x}')
            c.arg_roles = list(args); c.field('target', 'cls', cls)
    # Prefixes are physical guarded-page domains, with no inaccessible padding.
    for role, prefixes in [('target', [8, 9, 12, 13, 16, 17, 20, 520, 524, 528, 532, 536, 540, 720, 724, 748, 752]),
                           ('source', [8, 9, 12, 13, 17, 56, 60, 64, 72, 76, 80, 508, 512, 716, 720, 748, 752]),
                           ('weapon', [4, 5, 6, 8, 9, 12, 13, 16, 72, 76, 80, 736, 737, 740]),
                           ('target_update', [88, 89, 160, 161, 162, 1440, 1441, 1444, 2094, 2095, 2184, 2188, 2192]),
                           ('weapon_use', [96, 97, 100]), ('soundset', [32, 33, 36])]:
        for prefix in prefixes:
            c = add('read-domain', f'{role}-prefix-{prefix}').shape(role, readable=prefix)
            c.field('target', 'cls', 2).field('weapon', 'cls', 0x1000).field('weapon', 'sub', 0x10000)
            c.field('source', 'cls', 2).ret('004e0c63', 0).ret('004e1019', 'soundset')
            c.damage_type = 9
            if role == 'target':
                c.field('target', 'cls', 0).ret('004e0c63', 1)
            if role == 'source':
                c.field('source', 'cls', 2 if prefix in [748, 752] else 4)
                c.field('target', 'cls', 4).ret('004e0c63', 1).ret('004e1079', 1)
                if prefix <= 64:
                    c.field('weapon', 'cls', 0)
            if role == 'target_update':
                if prefix <= 89:
                    c.field('target', 'cls', 4); c.damage_type = 0
                elif prefix <= 162:
                    c.field('target', 'cls', 4)
                else:
                    # Mandatory early monster write uses the other full record;
                    # a declared external-service effect later selects the exact
                    # partial domain. This does not widen its inaccessible tail.
                    c.field('target', 'update', 'alt_update').ret('004e0c63', 1)
                    if prefix in [2094, 2095]:
                        c.field('target', 'sub', 0x10)
                        site = '004e0e06'
                    else:
                        c.damage_type = 0
                        site = '004e0f33'
                    c.stimulus('service', site, None, 0, 0,
                               [('target', 748, 'target_update', 4)])
    for prefix in [520, 524, 528, 532, 536, 540, 720, 724, 748, 752]:
        c = add('late-read-domain', f'world-target-prefix-{prefix}')
        c.shape('target', readable=prefix).field('target', 'cls', 0)
        c.damage_type = 0
    for prefix in [1440, 1441, 1444, 2188, 2189, 2192]:
        c = add('late-read-domain', f'post-defend-alternate-update-prefix-{prefix}')
        c.shape('alt_update', readable=prefix).field('target', 'cls', 2).field('target', 'sub', 0x10)
        c.damage_type = 0
        c.stimulus('service', '004e0f77', None, 0, 0, [('target', 748, 'alt_update', 4)])
    for role in ['target', 'source', 'weapon', 'target_update', 'source_update', 'weapon_use', 'soundset', 'parent', 'owner']:
        c = add('unreadable', role).shape(role, unreadable=True)
        c.field('target', 'cls', 2).field('weapon', 'cls', 0x1000).field('weapon', 'sub', 0x10000)
        c.field('source', 'cls', 2).ret('004e0c63', 0).ret('004e1019', 'soundset')
        if role == 'parent': c.ret('004e0c05', 0).ret('004e0c16', 'parent')
        if role == 'owner': c.field('source', 'cls', 0).field('source', 'owner', 'owner')
    for role in ['target', 'source', 'weapon', 'target_update', 'weapon_use', 'soundset']:
        for alignment in [1, 2, 3]:
            c = add('unaligned', f'{role}-plus-{alignment}').shape(role, unaligned=alignment)
            c.field('target', 'cls', 4).field('weapon', 'cls', 0x1000).field('weapon', 'sub', 0x10000)
            c.field('source', 'cls', 2).ret('004e0c63', 0).ret('004e1019', 'soundset').ret('004e1079', 1)
            c.ret('004e0ca5', 1); c.damage_type = 9
    # Each trigger is a statically decoded original memory read or service site.
    mutations = [
        ('early-class', 'read', '004e0b7b', 'target', 8, 1, [('target', 8, 4, 4)]),
        ('initial-update-cache', 'read', '004e0b86', 'target', 748, 4, [('target', 748, 'alt_update', 4)]),
        ('dead-flags-cache', 'read', '004e0bb9', 'target', 16, 4, [('target', 16, 0x8000, 4)]),
        ('child-weapon-class-cache', 'read', '004e142c', 'weapon', 8, 4, [('weapon', 8, 0x01000000, 4)]),
        ('child-use-pointer-cache', 'read', '004e143d', 'weapon', 736, 4, [('weapon', 736, 'alt_use', 4)]),
        ('child-use-byte-cache', 'read', '004e1443', 'weapon_use', 96, 1, [('weapon_use', 96, 0x01020300, 4)]),
        ('weapon-reload-between-children', 'read', '004e1456', 'weapon', 12, 4, [('weapon', 12, 0x4000, 4)]),
        ('shock-source-class-reload', 'service', '004e0d05', None, 0, 0, [('source', 8, 4, 4)]),
        ('immunity-subclass-cache', 'read', '004e0d29', 'target', 12, 4, [('target', 12, 0x800, 4)]),
        ('electric-class-cache', 'read', '004e0e60', 'target', 8, 4, [('target', 8, 0, 4)]),
        ('attribution-weapon-class-cache', 'read', '004e0ea1', 'weapon', 8, 4, [('weapon', 8, 0, 4)]),
        ('attribution-source-px', 'read', '004e0eaa', 'source', 72, 4, [('source', 72, 0x80000000, 4), ('source', 76, 0x7FC00001, 4)]),
        ('attribution-weapon-px', 'read', '004e0ebe', 'weapon', 72, 4, [('weapon', 72, 0x80000000, 4), ('weapon', 76, 0x7FC00001, 4)]),
        ('defend-later-class', 'service', '004e0f77', None, 0, 0, [('target', 8, 2, 4)]),
        ('generic-update-cache', 'read', '004e0faa', 'target', 748, 4, [('target', 748, 'alt_update', 4)]),
        ('status-cache', 'read', '004e0fb0', 'target_update', 1440, 4, [('target_update', 1440, 0x80000000, 4)]),
        ('injury-marker-cache', 'read', '004e0fb6', 'target_update', 2188, 4, [('target_update', 2188, 0, 4)]),
        ('pre-damage-later-class', 'service', '004e0fef', None, 0, 0, [('source', 8, 2, 4)]),
        ('source-update-lazy', 'read', '004e1008', 'source', 8, 1, [('source', 748, 0, 4)]),
        ('soundset-result-field', 'service', '004e1019', None, 0, 0, [('soundset', 32, 0, 4)]),
        ('sound-callback-cache', 'read', '004e1039', 'target', 720, 4, [('target', 720, 0, 4)]),
        ('vamp-late-pos', 'service', '004e10de', None, 0, 0, [('source', 56, 0x40200000, 4)]),
        ('vamp-source-x-cache', 'read', '004e10e3', 'source', 56, 4, [('source', 56, 0x7FC00001, 4), ('source', 60, 0xBF000000, 4)]),
        ('vamp-target-x-cache', 'read', '004e10f9', 'target', 56, 4, [('target', 56, 0x7FC00001, 4), ('target', 60, 0xBF000000, 4)]),
        ('ball-later-class', 'service', '004e1131', None, 0, 0, [('target', 8, 4, 4)]),
        ('state-byte-cache', 'read', '004e114d', 'target_update', 88, 1, [('target_update', 88, 15, 1)]),
        ('owner-pointer-cache', 'read', '004e1192', 'source', 508, 4, [('source', 508, 'owner_alt', 4)]),
        ('owner-class-pointer-cache', 'read', '004e119c', 'owner', 8, 1, [('source', 508, 'owner_alt', 4)]),
        ('enemy-late-shield', 'service', '004e11aa', None, 0, 0, [('target', 16, 2, 4)]),
        ('shield-end-attribution', 'service', '004e11f2', None, 0, 0, [('target', 520, 'source', 4)]),
    ]
    for name, when, pc, role, offset, width, effects in mutations:
        for variant in [0, 1]:
            c = add('stimulus', f'{name}-variant-{variant}')
            c.field('target', 'cls', 2 if variant == 0 else 4).field('target', 'sub', 0x10)
            c.field('source', 'cls', 4 if variant == 0 else 0).field('source', 'owner', 'owner')
            c.field('owner', 'cls', 2).field('owner_alt', 'cls', 0)
            c.field('weapon', 'cls', 0x1000 if variant == 0 else 0).field('weapon', 'sub', 0x10000)
            c.field('weapon_use', 96, 0xBADCFE02).field('target', 'sound_callback', SOUND_CALLBACK)
            c.ret('004e0c63', 1).ret('004e0ca5', 1).ret('004e1079', 1)
            c.ret('004e1019', 'soundset').ret('004e11c2', 1); c.damage_type = 9
            if pc.startswith('004e14'):
                c.ret('004e0c63', 0)
                if name == 'weapon-reload-between-children':
                    c.field('weapon', 'cls', 0x01000000).field('weapon', 'sub', 0)
                elif name in ['child-use-pointer-cache', 'child-use-byte-cache']:
                    c.field('weapon', 'cls', 0x1000)
            if name in ['source-update-lazy', 'soundset-result-field']:
                c.field('source', 'cls', 2)
            if name in ['initial-update-cache', 'immunity-subclass-cache',
                        'generic-update-cache', 'status-cache', 'injury-marker-cache']:
                c.field('target', 'cls', 2 if variant == 0 else 6)
            if name == 'shock-source-class-reload':
                c.field('source', 'cls', 2 if variant == 0 else 4)
                c.field('weapon', 'cls', 0x1000).field('weapon_use', 96, 0xBADCFE02)
            if name == 'pre-damage-later-class':
                c.field('weapon', 'cls', 0x1000 if variant == 0 else 0x01000000)
            if name == 'attribution-source-px':
                c.field('weapon', 'cls', 0x1000 if variant == 0 else 0x01000000)
            if name == 'attribution-weapon-px':
                c.field('weapon', 'cls', 0)
            if name == 'state-byte-cache':
                c.field('target', 'cls', 4)
            if name == 'electric-class-cache':
                c.damage_type = 9
            c.stimulus(when, pc, role, offset, width, effects)
    for pc, width in [('004e0b52', 1), ('004e0bea', 4), ('004e0db6', 1), ('004e0e1f', 1), ('004e0f9a', 4)]:
        c = add('clock-stimulus', pc); c.frame = 0
        c.damage_type = 1 if pc == '004e0db6' else 9
        c.fp('004e0d9d', 'half').fp('004e0e06', 'half')
        if pc == '004e0b52': c.ret('004e0b46', 1)
        if pc == '004e0bea': c.field('target', 'flags', 0x8000).ret('004e0bc2', 1)
        c.after_reads.append(dict(pc=pc, _trigger_role=None, _offset=0, width=width,
                                 occurrence=1, _effects=[]))
        # Absolute image mutation is intentionally distinct from graph effects.
        c.notes.append('clock mutation after original read; old loaded value remains in CPU')
        c._clock_stimulus = (pc, width)
    result = []
    for index, c in enumerate(cases):
        # Clock rows are resolved separately because FRAME is an image global.
        if hasattr(c, '_clock_stimulus'):
            c.after_reads.pop()
        row = c.freeze(index + 1)
        if hasattr(c, '_clock_stimulus'):
            pc, width = c._clock_stimulus
            row['after_reads'].append(dict(pc=pc, address=FRAME, width=width, occurrence=1,
                                           effects=[dict(address=FRAME, hex=u32(0x89ABCDEF))]))
        result.append(row)
    return result


def validate(cases):
    assert len({x['name'] for x in cases}) == len(cases)
    for c in cases:
        assert len(c['args']) == 5 and all(isinstance(x, int) and 0 <= x <= 0xFFFFFFFF for x in c['args'])
        assert set(c['services']) == set(SITES)
        occupied = set()
        for region in c['regions']:
            assert region['address'] % 4096 == 0 and region['size'] == 4096
            assert 0x21000000 <= region['address'] < 0x22000000
            assert region['address'] not in occupied; occupied.add(region['address'])
            for init in region['initializers']:
                assert 0 <= init['offset'] < 4096
                assert init['offset'] + len(bytes.fromhex(init['hex'])) <= 4096
        for site, schedule in c['services'].items():
            assert len(schedule['responses']) == 1 and schedule['repeat_last'] is True
            item = schedule['responses'][0]
            assert 0 <= item['eax'] <= 0xFFFFFFFF
            if site in X87_SITES:
                assert item['x87']['format'] == 'f64' and len(bytes.fromhex(item['x87']['hex'])) == 8
        for key in ['after_reads', 'after_services']:
            for s in c[key]:
                assert s['occurrence'] == 1 and s['effects']
                for effect in s['effects']:
                    assert len(bytes.fromhex(effect['hex'])) in [1, 2, 4]
        # Explicitly prohibit behavior-oracle metadata anywhere in the input.
        forbidden = {'expected', 'expected_return', 'expected_fault', 'expected_branch', 'native_output', 'golden'}
        def walk(x):
            if isinstance(x, dict):
                assert not forbidden.intersection(x)
                for y in x.values(): walk(y)
            elif isinstance(x, list):
                for y in x: walk(y)
        walk(c)


def sha(data):
    return hashlib.sha256(data).hexdigest()


def fact(path):
    data = path.read_bytes()
    return dict(path=str(path), bytes=len(data), sha256=sha(data))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output-dir', type=Path, default=ROOT)
    parser.add_argument('--check-only', action='store_true',
                        help='regenerate inputs in memory and compare canonical gzip; write nothing')
    args = parser.parse_args()
    assert sha(EXE.read_bytes()) == EXE_SHA
    cases = selected_cases(); validate(cases)
    payload = dict(schema='opennox-default-damage-parent-independent-inputs-v1',
                   source_reference=SOURCE, original_EXE_sha256=EXE_SHA,
                   entry='004e0b30', case_count=len(cases),
                   inputs_only=True, execution_performed=False,
                   families=dict(sorted(collections.Counter(c['family'] for c in cases).items())),
                   allocation_contract=dict(graph_range=[0x21000000, 0x22000000],
                       stack_address=0x22000000, stack_bytes=65536, caller_esp=0x2200E000,
                       auxiliary_address=0x23000000, auxiliary_bytes=65536,
                       damage_callback=DAMAGE_CALLBACK, sound_callback=SOUND_CALLBACK,
                       unspecified_pages='unmapped', region_fill='single explicit byte'),
                   cases=cases)
    data = (json.dumps(payload, separators=(',', ':'), ensure_ascii=True) + '\n').encode()
    plain = args.output_dir / 'chosen-independent-inputs-4e0b30-v1.json'
    zipped = args.output_dir / 'chosen-independent-inputs-4e0b30-v1.json.gz'
    if args.check_only:
        assert gzip.decompress(zipped.read_bytes()) == data
        if plain.exists():
            assert plain.read_bytes() == data
        print(json.dumps(dict(case_count=len(cases), regenerated_in_memory_equal=True,
                              uncompressed_bytes=len(data), uncompressed_sha256=sha(data),
                              writes_performed=False)))
        return
    args.output_dir.mkdir(parents=True, exist_ok=True)
    plain.write_bytes(data); zipped.write_bytes(gzip.compress(data, compresslevel=9, mtime=0))
    receipt = dict(schema='opennox-default-damage-parent-input-generation-receipt-v1',
                   generator=fact(Path(__file__).resolve()), output=fact(plain), compressed=fact(zipped),
                   original_EXE=fact(EXE), case_count=len(cases), families=payload['families'],
                   after_read_stimuli=sum(len(c['after_reads']) for c in cases),
                   after_service_stimuli=sum(len(c['after_services']) for c in cases),
                   external_schedule_entries_per_case=len(SITES),
                   no_outcome_fields=True, no_original_or_native_execution=True,
                   independent_input_only=True)
    (args.output_dir / 'input-generation-receipt-v1.json').write_text(json.dumps(receipt, indent=2) + '\n')
    print(json.dumps({k:receipt[k] for k in ['case_count', 'families', 'after_read_stimuli', 'after_service_stimuli']}))


if __name__ == '__main__':
    main()
