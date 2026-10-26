"""Execute unchanged DefaultDamage, real elemental protection and FISTP bodies.
Only named hash lookup and integer world services are substituted. No floating
arithmetic, protection return, balance float load or conversion is substituted.
Nil-health UnitDamageClear executes its original gate after observing damage.
"""
from pathlib import Path
from collections import Counter
import hashlib
import json
import struct
import sys
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import *

if len(sys.argv) not in (4, 5) or (len(sys.argv) == 5 and sys.argv[4] != '--historical'):
    raise SystemExit('usage: python original_default_damage_elemental_startup_4e0b30.py REPO GAME.EXE OUTPUT.json [--historical]')
REPO = Path(sys.argv[1]).resolve()
EXE = Path(sys.argv[2]).resolve()
OUT = Path(sys.argv[3]).resolve()
HISTORICAL = len(sys.argv) == 5
if OUT.exists():
    raise FileExistsError(OUT)
blob = EXE.read_bytes()
sha = lambda data: hashlib.sha256(data).hexdigest()
assert sha(blob) == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe = struct.unpack_from('<I', blob, 60)[0]
opt = pe + 24
base = struct.unpack_from('<I', blob, opt+28)[0]
tab = opt + struct.unpack_from('<H', blob, pe+20)[0]
sections = [struct.unpack_from('<IIII', blob, tab+40*i+8) for i in range(struct.unpack_from('<H', blob, pe+6)[0])]

def code(address, count):
    for _, virtual, length, offset in sections:
        if base+virtual <= address and address+count <= base+virtual+length:
            return blob[offset+address-base-virtual:offset+address-base-virtual+count]
    raise AssertionError(hex(address))

regions = {
    'DefaultDamage-complete': (0x4e0b30, 1792),
    'GameBallOnDamage-complete': (0x4e1230, 240),
    'Defend-complete': (0x4e1320, 143),
    'PreDamage-complete': (0x4e13b0, 76),
    'AttackQualifies-complete': (0x4e1400, 112),
    'DefendUnit-complete': (0x4e1470, 34),
    'OwnerChain-complete': (0x4ec580, 38),
    'MonsterHitLatch-complete': (0x532880, 33),
    'FireProtection-complete': (0x4dfe40, 255),
    'ElectricProtection-complete': (0x4dff40, 255),
    'buff-membership-complete': (0x4ff350, 35),
    'buff-power-complete': (0x4ff570, 16),
    'balance-float-table-complete': (0x419d70, 50),
    'signed-FISTP-complete': (0x419a70, 17),
    'UnitDamageClear-nil-health-gate': (0x4ee5e0, 264),
    'startup-control-call': (0x4014b5, 15),
    'actual-CRT-control-word': (0x40209e, 358),
}
manifest = json.loads((REPO/'toolchain/oracle/game-exe-functions.json').read_text())
for address, size in regions.values():
    matches = [r for r in manifest['functions'] if int(r['address'], 16) == address and r['size'] == size]
    if matches:
        assert sha(code(address, size)) == matches[0]['sha256']

cpu = Uc(UC_ARCH_X86, UC_MODE_32)
cpu.mem_map(base, (struct.unpack_from('<I', blob, opt+56)[0]+4095) & ~4095)
for _, virtual, size, offset in sections:
    if size:
        cpu.mem_write(base+virtual, blob[offset:offset+size])
heap, stack, stop = 0x10000000, 0x20000000, 0x30000000
cpu.mem_map(heap, 0x20000)
cpu.mem_map(stack, 0x10000)
cpu.mem_map(stop, 0x1000)
v, a, first, second, attrs0, attrs1, attrs2 = [heap+0x1000*i for i in range(1, 8)]
effects = [heap+0x8000+0x100*i for i in range(12)]
record, table = heap+0x9000, heap+0xa004
hash_handle = 0x12345678
get = lambda p: struct.unpack('<I', cpu.mem_read(p, 4))[0]
put = lambda p, value: cpu.mem_write(p, struct.pack('<I', value & 0xffffffff))
f32bits = lambda value: struct.unpack('<I', struct.pack('<f', value))[0]
trace, entries, intercepted, faults, rows = {}, Counter(), Counter(), [], []
events, clear_damage, current = [], [], {}

def hook(_cpu, address, count, _):
    if address == 0x4014c4:
        cpu.emu_stop()
        return
    sp = cpu.reg_read(UC_X86_REG_ESP)
    arg = lambda index: get(sp+4+4*index)
    if address == 0x4e0b30:
        entries['DefaultDamage'] += 1
    if address in (0x4dfe40, 0x4dff40):
        assert arg(0) == v
        entries['FireProtection' if address == 0x4dfe40 else 'ElectricProtection'] += 1
        events.append('protection')
    if address == 0x419d70:
        key = bytes(cpu.mem_read(arg(0), 64)).split(b'\0', 1)[0].decode('ascii')
        index = arg(1)
        assert key == current['key'] and index == ((current['input']['power']-1) & 0xffffffff)
        events.append(f'balance:{key}:{index if index < 0x80000000 else index-0x100000000}')
    if address == 0x419a70:
        entries['FISTP'] += 1
    if address == 0x4ee5e0:
        assert arg(0) == v
        value = arg(1)
        clear_damage.append(value if value < 0x80000000 else value-0x100000000)
        events.append('clear')
    if any(start <= address < start+size for start, size in regions.values()):
        trace[address] = bytes(cpu.mem_read(address, count)).hex()
        return
    result = 0
    if address == 0x426890:
        key = bytes(cpu.mem_read(arg(1), 64)).split(b'\0', 1)[0].decode('ascii')
        assert arg(0) == hash_handle and key == current['key']
        result = record
    elif address == 0x417da0:
        assert arg(0) == 1
        result = 1
    elif address == 0x40a5c0:
        assert arg(0) == 6144
    elif address == 0x5330c0:
        assert (arg(0), arg(1)) == (v, a)
        events.append('enemy')
    elif address == 0x501960:
        assert (arg(0), arg(1), arg(2), arg(3)) == (104 if current['function'] == 'FireProtection' else 108, v, 0, 0)
        events.append('audio')
    elif address == 0x4ff5b0:
        assert (arg(0), arg(1)) == (v, 0)
        put(v+340, get(v+340) & ~1)
        events.append('buff')
    elif address == 0x532e20:
        assert (arg(0), arg(1)) == (v, a)
        events.append('sound')
    else:
        raise AssertionError(f'undeclared instruction {address:x} {current.get("name")}')
    intercepted[hex(address)] += 1
    cpu.reg_write(UC_X86_REG_EAX, result)
    cpu.reg_write(UC_X86_REG_ESP, sp+4)
    cpu.reg_write(UC_X86_REG_EIP, get(sp))

def invalid(_cpu, access, address, size, value, _):
    faults.append(dict(name=current.get('name'), instruction=hex(cpu.reg_read(UC_X86_REG_EIP)), address=hex(address), size=size))
    return False

cpu.hook_add(UC_HOOK_CODE, hook)
cpu.hook_add(UC_HOOK_MEM_INVALID, invalid)
sp = stack+0x8000
cpu.reg_write(UC_X86_REG_ESP, sp)
cpu.reg_write(UC_X86_REG_FPCW, 0x27f)
cpu.emu_start(0x4014b5, 0x4014c4, count=20000)
assert cpu.reg_read(UC_X86_REG_FPCW) == 0xe3f
initial = cpu.context_save()
contexts = []
if HISTORICAL:
    fixture = json.loads((REPO/'src/server/testdata/default_damage_elemental_historical_startup.json').read_text())
    for row in fixture['results']:
        protection = struct.unpack('<d', struct.pack('<Q', row['protection_bits']))[0]
        stored_bits = f32bits(protection)
        stored = struct.unpack('<f', struct.pack('<I', stored_bits))[0]
        if abs(stored) > abs(protection):
            stored_bits -= 1
        ctx = dict(name='historical-projection', inventory=[], unit=[], buff=True, power=1, balance_bits=stored_bits,
                   unit_class=0, item_flags=0x100, item_class=0x1000, matching=True, null_unit=False)
        contexts.append(dict(name=f'historical-projection/{row["protection_bits"]:016x}/{row["damage"]}',
                             function='FireProtection', input=ctx, damage=row['damage'], damage_type=12,
                             requested_protection_bits=row['protection_bits'], expected_startup_spill_bits=stored_bits))
    assert len(contexts) == 552
else:
    fixture = json.loads((REPO/'src/server/testdata/default_damage_elemental_startup_original.json').read_text())
    # Read declared inputs only. Recompute all returns, events, stores and HP
    # arguments through the unchanged PE32 code, without reading expected results.
    for row in fixture['cases']:
        contexts.append({key: row[key] for key in ('name', 'function', 'input', 'damage', 'damage_type')})
    assert len(contexts) == 622

for current in contexts:
    ctx = current['input']
    fire = current['function'] == 'FireProtection'
    enchant, engage = (17, 0x4dfd10) if fire else (20, 0x4dfd80)
    current['key'] = 'FireSpellProtection' if fire else 'ElectricitySpellProtection'
    for mode, cw in [('actual-startup-inherited', 0xe3f), ('nearest-control', 0x27f), ('down-control', 0x67f), ('up-control', 0xa7f)]:
        cpu.context_restore(initial)
        cpu.reg_write(UC_X86_REG_ESP, sp)
        cpu.reg_write(UC_X86_REG_FPCW, cw)
        cpu.reg_write(UC_X86_REG_FPSW, 0)
        cpu.reg_write(UC_X86_REG_FPTAG, 0xffff)
        cpu.mem_write(heap, b'\xa7'*0x20000)
        for obj in (v, a, first, second):
            cpu.mem_write(obj, bytes(0x300))
        for attrs in (attrs0, attrs1, attrs2):
            cpu.mem_write(attrs, bytes(16))
        for effect in effects:
            cpu.mem_write(effect, bytes(0x100))
        cpu.mem_write(v+4, struct.pack('<H', 714))
        put(v+8, ctx['unit_class'])
        put(v+16, 0x1000000)
        put(v+340, 1 | ((1 << enchant) if ctx['buff'] else 0))
        cpu.mem_write(v+408+enchant, bytes([ctx['power']]))
        put(v+504, first if ctx['inventory'] else 0)
        put(v+692, attrs2)
        put(v+524, 19)
        put(v+536, 77)
        cpu.mem_write(v+528, struct.pack('<ff', 31, 47))
        cpu.mem_write(a+4, struct.pack('<H', 706))
        put(a+8, 1572865)
        put(a+12, 1)
        put(a+16, 553665028)
        cpu.mem_write(a+72, struct.pack('<ff', 17, 19))
        for j, (obj, attrs) in enumerate(((first, attrs0), (second, attrs1))):
            put(obj+8, ctx['item_class'])
            put(obj+16, ctx['item_flags'])
            put(obj+496, second if j == 0 and len(ctx['inventory']) > 1 else 0)
            put(obj+692, attrs)
        for group, attrs in enumerate((attrs0, attrs1, attrs2)):
            values = ctx['unit'] if group == 2 else ctx['inventory'][group] if group < len(ctx['inventory']) else []
            for i in range(4):
                effect = effects[group*4+i]
                put(attrs+4*i, effect if i < len(values) else 0)
                put(effect+112, engage if ctx['matching'] else 0x401000)
                put(effect+120, values[i] if i < len(values) else 0)
        put(0x6552d8, hash_handle)
        put(record, table)
        cpu.mem_write(record+4, b'\xff')
        put(table-4, 0)
        for i in range(255):
            put(table+4*i, ctx['balance_bits'])
        put(0x84ea04, 1400)
        before = bytes(cpu.mem_read(heap, 0x20000))
        args = struct.pack('<6I', stop, v, a, a, current['damage'] & 0xffffffff, current['damage_type'])
        cpu.mem_write(sp, args)
        events.clear()
        clear_damage.clear()
        cpu.emu_start(0x4e0b30, stop, count=20000)
        assert cpu.reg_read(UC_X86_REG_EIP) == stop and cpu.reg_read(UC_X86_REG_ESP) == sp+4
        assert cpu.reg_read(UC_X86_REG_FPCW) == cw and len(clear_damage) == 1
        after = bytearray(cpu.mem_read(heap, 0x20000))
        allowed = set(range(340, 344)) | set(range(520, 540))
        changed = [i for i, (x, y) in enumerate(zip(before[0x1000:0x1300], after[0x1000:0x1300])) if x != y]
        assert set(changed) <= allowed, changed
        after[0x1000:0x1300] = before[0x1000:0x1300]
        assert bytes(after) == before
        caller_after = bytearray(cpu.mem_read(sp, len(args)))
        # 004E0DA2/004E0E0B reuse the cached source argument slot for v46's
        # protection FST; the incoming damage slot is also reused in place.
        protection_spill = get(sp+8)
        caller_after[8:12] = args[8:12]
        caller_after[16:20] = args[16:20]
        assert bytes(caller_after) == args, (current['name'], mode, [(i, x, y) for i, (x, y) in enumerate(zip(args, caller_after)) if x != y])
        rows.append(dict(**current, mode=mode, fpcw=cw, result=cpu.reg_read(UC_X86_REG_EAX),
            clear_damage=clear_damage[0], protection_spill_bits=protection_spill, buffs=get(v+340), attributed=get(v+520) == a,
            damage_type_after=get(v+524), frame=get(v+536), position_bits=[get(v+528), get(v+532)],
            events=list(events), all_nonoutput_bytes_preserved=True, stack_balanced=True, rounding_mode_unchanged=True))

assert not faults and EXE.read_bytes() == blob
for address, size in regions.values():
    assert bytes(cpu.mem_read(address, size)) == code(address, size)
proof = dict(original_sha256=sha(blob), original_unchanged=True, contexts=len(contexts), raw_rows=len(rows),
    original_complete_calls=len(rows), rows=rows, regions=[dict(name=name, address=hex(address), size=size, sha256=sha(code(address, size))) for name, (address, size) in regions.items()],
    executed_original_instruction_records=[dict(address=hex(a), bytes=b) for a, b in sorted(trace.items())],
    unhooked_entries=dict(entries), declared_external_service_counts=dict(intercepted),
    declared_external_services=['426890 named hash lookup returns declared table record; actual419D70 float load executes', '417da0 gameplay flag1=true', '40a5c0 quest flags=false', '5330c0 enemy=false', '501960 protection sound', '4ff5b0 invisibility removal', '532e20 damage sound'],
    numeric_or_protection_return_substitutions=0, memory_faults=faults,
    scope=__doc__.strip(), overall_porting_100_percent=False)
out = OUT
with out.open('x') as stream:
    json.dump(proof, stream, indent=2)
    stream.write('\n')
print(json.dumps(dict(path=str(out), contexts=len(contexts), complete_calls=len(rows), executed_instruction_addresses=len(trace), services=len(intercepted))))
