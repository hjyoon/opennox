"""Execute unchanged TeleportToTarget roots, their ray shim and FISTP helper.

Only tile/ray queries, strings/messages, mode/buff reads, balance return,
FX/sounds, wake, teleport and attribution are declared game services.
No caller, argument packing, conversion or control-flow instruction is replaced.
"""
import argparse, collections, copy, hashlib, json, math, struct
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import *
from pathlib import Path

p = argparse.ArgumentParser(description=__doc__)
for key in ['game', 'fixture', 'proof']:
    p.add_argument('--' + key, type=Path, required=True)
p.add_argument('--control', type=lambda value: int(value, 0), choices=[0xe3f, 0x23f], default=0xe3f)
a = p.parse_args()
blob = a.game.read_bytes()
sha = lambda value: hashlib.sha256(value).hexdigest()
assert sha(blob) == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe = struct.unpack_from('<I', blob, 60)[0]
opt = pe + 24
base = struct.unpack_from('<I', blob, opt + 28)[0]
table = opt + struct.unpack_from('<H', blob, pe + 20)[0]
u = Uc(UC_ARCH_X86, UC_MODE_32)
u.mem_map(base, (struct.unpack_from('<I', blob, opt + 56)[0] + 4095) & ~4095)
for n in range(struct.unpack_from('<H', blob, pe + 6)[0]):
    _, va, size, off = struct.unpack_from('<IIII', blob, table + 40*n + 8)
    if size:
        u.mem_write(base + va, blob[off:off+size])
heap, stack, code = 0x21000000, 0x22000000, 0x23000000
for address in [heap, stack, code]:
    u.mem_map(address, 0x10000)
record = heap + 0x1000
objects = [heap + 0x2000 + 0x400*i for i in range(4)]
updates = [heap + 0x5000 + 0x600*i for i in range(4)]
players = [heap + 0x9000 + 0x1000*i for i in range(4)]
sp, stop, init, thunk, scratch, cw, message = stack+0xe000, code+0x100, code+0x200, code+0x300, code+0x400, code+0x500, code+0x600
frameaddr = 0x84ea04
u.mem_write(init, b'\xdb\xe3\xd9\x2d' + struct.pack('<I', cw))
u.mem_write(thunk, b'\xdd\x05' + struct.pack('<I', scratch) + b'\xc3')
get = lambda address: struct.unpack('<I', u.mem_read(address, 4))[0]
put = lambda address, value: u.mem_write(address, struct.pack('<I', value & 0xffffffff))
words = lambda address: [get(address), get(address+4)]
f32 = lambda value: struct.unpack('<I', struct.pack('<f', value))[0]
point = lambda x, y: [f32(x), f32(y)]
f64 = lambda value: struct.pack('<d', value).hex()
ref = lambda index: 0 if index < 0 else objects[index]
index = lambda address: -1 if address == 0 else objects.index(address)
update_index = lambda address: -1 if address == 0 else updates.index(address)
player_index = lambda address: -1 if address == 0 else players.index(address)
cstr = lambda address: bytes(u.mem_read(address, 128)).split(b'\0', 1)[0].decode('ascii')
regions = {'create': (0x530a30, 0x530b6d), 'update': (0x530b70, 0x530c98),
           'converter': (0x419a70, 0x419a81), 'ray': (0x5374b0, 0x5374c4)}
traces = {name: {} for name in regions}
entries, services, mutations = collections.Counter(), collections.Counter(), collections.Counter()
active, current, events, calls, mutated, allowed, faults = False, {}, [], collections.Counter(), False, set(), []

def state():
    return dict(source=index(get(record+12)), caster=index(get(record+16)), target=index(get(record+48)),
                spell=get(record+4), level=get(record+8), flag=get(record+20), flags88=get(record+88),
                destination=words(record+52), timer=get(record+68), frame=get(frameaddr),
                positions=[words(obj+56) for obj in objects], classes=[get(obj+8) for obj in objects],
                flags=[get(obj+16) for obj in objects], buffs=[get(obj+340) for obj in objects],
                netcodes=[get(obj+36) for obj in objects], object_updates=[update_index(get(obj+748)) for obj in objects],
                update_players=[player_index(get(data+276)) for data in updates],
                player_indices=[u.mem_read(data+2064, 1)[0] for data in players])

def allow(address, size):
    allowed.update(range(address-heap, address-heap+size))

def write(address, value):
    allow(address, 4)
    put(address, value)

def writepoint(address, values):
    for n, value in enumerate(values):
        write(address+4*n, value)

def mutate(kind):
    global mutated
    at = calls[kind]
    calls[kind] += 1
    mutation = current['mutation']
    if mutated or mutation['stage'] != kind or mutation['at'] != at:
        return
    mutated = True
    action = mutation['action']
    mutations[kind+'/'+str(at)+'/'+action] += 1
    if action == 'target-b': write(record+48, objects[3])
    elif action == 'caster-b': write(record+16, objects[0])
    elif action == 'caster-nil': write(record+16, 0)
    elif action == 'spell': write(record+4, 43)
    elif action == 'destination': writepoint(record+52, point(-37.5, 43.25))
    elif action == 'position-target': writepoint(objects[2]+56, point(-17.5, 23.25))
    elif action == 'invisible-target': write(objects[2]+340, get(objects[2]+340) | 1)
    elif action == 'visible-target': write(objects[2]+340, get(objects[2]+340) & ~1)
    elif action == 'netcode-target': write(objects[2]+36, 0xdeadbeef)
    elif action == 'update-caster': write(objects[1]+748, updates[0])
    elif action == 'player-caster': write(updates[1]+276, players[0])
    elif action == 'frame': put(frameaddr, 0xfffffffd)
    else: raise AssertionError(action)

def event(kind, arguments):
    events.append(dict(kind=kind, args=arguments, state=state()))
    mutate(kind)

def ret(value=0):
    esp = u.reg_read(UC_X86_REG_ESP)
    u.reg_write(UC_X86_REG_EAX, value & 0xffffffff)
    u.reg_write(UC_X86_REG_EIP, get(esp))
    u.reg_write(UC_X86_REG_ESP, esp+4)

def hook(_u, pc, length, _):
    if not active: return
    if pc == stop:
        u.emu_stop()
        return
    for name, (lo, hi) in regions.items():
        if lo <= pc < hi:
            traces[name][pc] = bytes(u.mem_read(pc, length))
            if pc == lo: entries[name] += 1
            return
    if thunk <= pc < thunk+7: return
    esp = u.reg_read(UC_X86_REG_ESP)
    arg = lambda n: get(esp+4*n)
    services[hex(pc)] += 1
    if pc == 0x411a90:
        assert arg(1) == record+52
        event('tile', words(arg(1)))
        ret(int(current['blocked']))
    elif pc == 0x40f1d0:
        assert cstr(arg(1)) == 'UnseenTarget' and arg(2) == 0 and cstr(arg(3)) == 'C:\\NoxPost\\src\\Server\\Magic\\Spell\\ExecDur.c' and arg(4) == 2912
        event('lookup', [])
        ret(message)
    elif pc == 0x4d9eb0:
        assert arg(2) == message
        event('unseen', [index(arg(1))])
        ret(1)
    elif pc == 0x535250:
        assert [arg(2), arg(3), arg(4)] == [0, 0, 9]
        event('ray', [get(arg(1)+4*n) for n in range(4)])
        ret(int(current['clear']))
    elif pc == 0x4da0f0:
        assert arg(2) == 0 and get(arg(3)) == 2
        event('inform', [arg(1), 0, 2])
        ret(0x24681357)
    elif pc == 0x40a5c0:
        assert arg(1) == 2048
        ret(int(current['coop']))
    elif pc == 0x419d70:
        assert cstr(arg(1)) == 'TeleportDelay'
        event('delay', [arg(2)])
        u.mem_write(scratch, bytes.fromhex(current['delay_bits64']))
        u.reg_write(UC_X86_REG_EIP, thunk)
    elif pc == 0x4ff350:
        assert arg(2) in [0, 14]
        ret(int(bool(get(arg(1)+340) & (1 << arg(2)))))
    elif pc == 0x522ff0:
        assert arg(1) == 137
        event('fx', [137, *words(arg(2))])
        ret(0x24681357)
    elif pc == 0x424800:
        assert arg(2) == 0
        event('sound', [arg(1), 0])
        ret(current['sound'] + calls['sound'] - 1)
    elif pc == 0x501960:
        event('audio', [arg(1), index(arg(2)), arg(3), arg(4)])
        ret(0x13572468)
    elif pc == 0x530560:
        assert arg(2) == arg(1)+56 and arg(3) == record+52
        event('wake', [index(arg(1)), *words(arg(2)), *words(arg(3))])
        ret(0x13572468)
    elif pc == 0x4e7190:
        assert arg(2) == record+52
        target, destination = arg(1), words(arg(2))
        event('teleport', [index(target), *destination])
        writepoint(target+56, destination)
        ret(1)
    elif pc == 0x4e7540:
        event('attribution', [index(arg(1)), index(arg(2))])
        ret(0x24681357)
    else:
        raise AssertionError((current['name'], 'undeclared instruction/service', hex(pc)))

def invalid(_u, access, address, size, value, _):
    faults.append(dict(name=current['name'], pc=hex(u.reg_read(UC_X86_REG_EIP)), address=hex(address)))
    return False

u.hook_add(UC_HOOK_CODE, hook)
u.hook_add(UC_HOOK_MEM_INVALID, invalid)
seed = bytes((n*31+19) & 255 for n in range(0x10000))
registers = {UC_X86_REG_EBX:0x13572468, UC_X86_REG_ESI:0x24681357, UC_X86_REG_EDI:0x98765432, UC_X86_REG_EBP:0xfedcba98}

def run(case):
    global current, events, calls, mutated, active, allowed
    active, current, events, calls, mutated, allowed = False, case, [], collections.Counter(), False, set()
    u.mem_write(heap, seed)
    for offset, value in [(4,126),(8,case['level']),(12,ref(case['source'])),(16,ref(case['caster'])),(20,case['flag']),(48,ref(case['target'])),(68,case['timer']),(88,case['flags88'])]: put(record+offset, value)
    for n, value in enumerate(case['destination']): put(record+52+4*n, value)
    for n, obj in enumerate(objects):
        for offset, value in [(8,case['classes'][n]),(16,case['flags'][n]),(36,case['netcodes'][n]),(340,case['buffs'][n]),(748,0 if case['object_updates'][n] < 0 else updates[case['object_updates'][n]])]: put(obj+offset, value)
        for j, value in enumerate(case['positions'][n]): put(obj+56+4*j, value)
        put(updates[n]+276, players[case['update_players'][n]])
        u.mem_write(players[n]+2064, bytes([case['player_indices'][n]]))
    u.mem_write(frameaddr-16, bytes(range(36)))
    put(frameaddr, case['frame'])
    global_before = bytes(u.mem_read(frameaddr-16, 36))
    u.mem_write(cw, struct.pack('<H', a.control))
    u.emu_start(init, init+8, count=2)
    before = bytes(u.mem_read(heap, len(seed)))
    u.mem_write(sp-512, bytes(1024))
    put(sp, stop)
    put(sp+4, record)
    for reg, value in registers.items(): u.reg_write(reg, value)
    u.reg_write(UC_X86_REG_EFLAGS, 0x202)
    u.reg_write(UC_X86_REG_ESP, sp)
    active = True
    u.emu_start(regions[case['kind']][0], stop+1, count=20000)
    active = False
    assert u.reg_read(UC_X86_REG_EIP) == stop and u.reg_read(UC_X86_REG_ESP) == sp+4 and get(sp) == stop
    assert all(u.reg_read(reg) == value for reg, value in registers.items())
    assert bytes(u.mem_read(sp-512, 256)) == bytes(256) and bytes(u.mem_read(sp+8, 256)) == bytes(256)
    assert u.reg_read(UC_X86_REG_FPCW) == a.control and u.reg_read(UC_X86_REG_FPSW) & 0x3800 == 0
    if case['kind'] == 'create':
        allow(record+48, 4)
        allow(record+68, 4)
    after = bytes(u.mem_read(heap, len(seed)))
    assert all(x == y or n in allowed for n, (x,y) in enumerate(zip(before, after)))
    after = bytes(u.mem_read(frameaddr-16, 36))
    assert all(x == y or 16 <= n < 20 for n, (x,y) in enumerate(zip(global_before, after)))
    if case['mutation']['stage']: assert mutated, (case['name'], 'mutation not reached')
    return dict(input=copy.deepcopy(case), result=dict(ret=u.reg_read(UC_X86_REG_EAX), state=state(), events=copy.deepcopy(events)))

template = dict(name='', kind='update', coop=False, blocked=False, clear=True, flag=0, flags88=0,
                level=3, timer=101, frame=100, source=0, caster=1, target=2,
                classes=[4,4,4,2], flags=[0]*4, buffs=[0]*4, netcodes=[0x80000001,0x81234567,0xf000cdef,0x12345678],
                positions=[point(7,11),point(100.25,-100.5),point(110.25,-80.5),point(104,-90)],
                destination=point(300.75,-400.5), object_updates=[0,1,2,3], update_players=[0,1,2,3],
                player_indices=[30,31,32,33], delay_bits64=f64(2.5), sound=351, mutation=dict(stage='',at=0,action=''))
cases = []
def add(name, **changes):
    case = copy.deepcopy(template)
    case.update(changes)
    case['name'] = case['kind']+'/'+name
    cases.append(case)
def create(name, **changes):
    changes.update(kind='create', timer=0x87654321)
    add(name, **changes)
create('normal'); create('default-target',target=-1); create('blocked',blocked=True); create('blocked-default',blocked=True,target=-1)
create('no-ray',clear=False)
classes = list(template['classes']); classes[1] = 2
create('no-ray-monster',clear=False,classes=classes)
object_updates = list(template['object_updates']); object_updates[1] = -1
create('no-ray-null-update',clear=False,object_updates=object_updates)
for value in [.49,.51,1.5,2.5,3.5,-1.5,-2.5,-3.5,1.999999999999,2.999999999999,2147483647.,-2147483647.,math.inf,math.nan]:
    create('delay-'+f64(value),coop=True,delay_bits64=f64(value))
for level in [0,19,0xffffffff]: create('level-'+str(level),coop=True,level=level)
create('frame-wrap',frame=0xffffffff)
for stage, actions, options in [
    ('tile',['target-b','caster-b','caster-nil','destination','position-target','update-caster','player-caster'],{}),
    ('lookup',['target-b','caster-b','destination'],dict(blocked=True)),
    ('unseen',['target-b','caster-b'],dict(blocked=True)),
    ('ray',['target-b','caster-b','destination'],{}),
    ('tile',['update-caster','player-caster','caster-b'],dict(clear=False)),
    ('ray',['update-caster','player-caster','caster-b'],dict(clear=False)),
    ('delay',['frame','target-b','caster-b','destination'],dict(coop=True))]:
    for action in actions:
        create(stage+'-'+action+('-no-ray' if options.get('clear') is False else ''),mutation=dict(stage=stage,at=0,action=action),**options)
for caster, target in [(-1,2),(1,-1)]: add('nil-'+str(caster)+'-'+str(target),caster=caster,target=target)
for unit in [1,2]:
    for flag in [0x20,0x8000]:
        flags=[0]*4; flags[unit]=flag
        add('dead-'+str(unit)+'-'+str(flag),flags=flags)
add('wrong-frame',timer=103); add('zero-timer',timer=0); add('frame-wrap',timer=0,frame=0xffffffff)
buffs=[0]*4; buffs[2]=1<<14; add('anchor',buffs=buffs)
add('visible')
for cls in [2,4]:
    classes=list(template['classes']); classes[2]=cls
    buffs=[0]*4; buffs[2]=1
    add('invisible-'+str(cls),classes=classes,buffs=buffs)
for stage, at, actions in [
    ('fx',0,['target-b','caster-b','destination','position-target','invisible-target','spell']),
    ('sound',0,['target-b','caster-b','destination','spell']),
    ('audio',0,['target-b','caster-b','destination','position-target','invisible-target']),
    ('wake',0,['target-b','caster-b','destination','position-target']),
    ('teleport',0,['target-b','caster-b','destination','position-target','invisible-target']),
    ('fx',1,['target-b','caster-b','destination','position-target','spell']),
    ('sound',1,['target-b','caster-b','spell']),
    ('audio',1,['target-b','caster-b'])]:
    for action in actions: add(stage+'-'+str(at)+'-'+action,mutation=dict(stage=stage,at=at,action=action))
for stage, at, action in [('sound',1,'target-b'),('sound',1,'netcode-target'),('sound',1,'spell'),('teleport',0,'visible-target')]:
    buffs=[0]*4; buffs[2]=1
    add('invisible-'+stage+'-'+str(at)+'-'+action,buffs=buffs,mutation=dict(stage=stage,at=at,action=action))
buffs=[0]*4; buffs[2]=1<<14
add('anchor-audio-target',buffs=buffs,mutation=dict(stage='audio',at=0,action='target-b'))
assert len({case['name'] for case in cases}) == len(cases)
rows = [run(case) for case in cases]
sealed = []
for name, (lo,hi) in regions.items():
    instructions = traces[name]
    covered = {pc+n for pc, body in instructions.items() for n in range(len(body))}
    assert covered == set(range(lo,hi)), (name, sorted(set(range(lo,hi))-covered))
    sealed.append(dict(name=name,address=hex(lo),end_exclusive=hex(hi),bytes=hi-lo,sha256=sha(bytes(u.mem_read(lo,hi-lo))),
                       instructions=[dict(pc=hex(pc),bytes=body.hex()) for pc,body in sorted(instructions.items())]))
encoded = (json.dumps(dict(original_sha256=sha(blob),control_word=a.control,rows=rows),indent=2)+'\n').encode()
a.fixture.write_bytes(encoded)
proof = dict(original_sha256=sha(blob),control_word=a.control,contexts=len(rows),entries=dict(entries),regions=sealed,
             all_original_instruction_bytes_executed=True,zero_caller_packing_arithmetic_or_conversion_replacements=True,
             callee_saved_registers_stack_control_word_TOP_and_nonoutput_heap_global_canaries_preserved=True,
             services=dict(services),mutations=dict(mutations),faults=faults,fixture_sha256=sha(encoded),
             scope='Whole original TeleportToTarget roots with actual ray shim and FISTP converter under declared game services. Wake is a declared service; no transitive-callee or full-campaign equivalence claim.')
assert not faults
a.proof.write_text(json.dumps(proof,indent=2)+'\n')
print(json.dumps({key:value for key,value in proof.items() if key not in ['regions','services','mutations']},indent=2))
print('mutations',len(mutations),sum(mutations.values()))
