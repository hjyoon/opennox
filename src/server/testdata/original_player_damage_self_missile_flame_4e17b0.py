"""Bounded unchanged GAME.EXE PlayerDamage FLAME prefix with the reported
StrongFireball (694, class 0x280001, subclass 1, flags 553665028) as BOTH
source and weapon. Buff/mode/facing/audio/reflection/owner/balance/armor-wear
and DefaultDamage are declared service boundaries. Markers, defenses, signed
raw FLAME, Quest scaling and return values execute sealed PE32 instructions.
DefaultDamage records its input, not Windows HP. C-owned native HP is tested
separately. No stock-map, GUI, item-destruction or whole-world certification.
"""
import hashlib
import json
import pathlib
import struct
import sys
from collections import Counter
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import (UC_X86_REG_EAX, UC_X86_REG_EIP, UC_X86_REG_ESP,
    UC_X86_REG_FPCW, UC_X86_REG_FPSW, UC_X86_REG_FPTAG)

repo, original = map(pathlib.Path, sys.argv[1:3])
# Keep the historical nearest control as the default; pass 0xe3f to replay
# the original startup mode. This changes CPU configuration, never code bytes.
control_word = int(sys.argv[3], 0) if len(sys.argv) > 3 else 0x027f
assert control_word in (0x027f, 0x0e3f)
blob = original.read_bytes()
manifest = json.loads((repo / 'toolchain/oracle/game-exe-functions.json').read_text())
assert hashlib.sha256(blob).hexdigest() == manifest['file_sha256'] == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe = struct.unpack_from('<I', blob, 0x3c)[0]
opt = pe + 24
base = struct.unpack_from('<I', blob, opt + 28)[0]
assert base == 0x400000 and struct.unpack_from('<H', blob, pe + 4)[0] == 0x14c
table = opt + struct.unpack_from('<H', blob, pe + 20)[0]
sections = [struct.unpack_from('<IIII', blob, table + i*40 + 8)
    for i in range(struct.unpack_from('<H', blob, pe + 6)[0])]
size = (struct.unpack_from('<I', blob, opt + 56)[0] + 4095) & ~4095
def read(address, count):
    for _, virtual, raw_size, raw in sections:
        if virtual <= address-base and address-base+count <= virtual+raw_size:
            off = raw+address-base-virtual
            return blob[off:off+count]
    raise AssertionError(hex(address))
addresses = {0x4e17b0, 0x4e20f0, 0x4e1400, 0x4ec580, 0x419a70, 0x4e2330}
ranges = [r for r in manifest['functions'] if int(r['address'], 16) in addresses]
# Later observations can name the same sealed procedure more than once.
# Collapse only identical address/size/hash triples, preserving first metadata.
unique_ranges = {}
for r in ranges:
    unique_ranges.setdefault((int(r['address'], 16), r['size'], r['sha256']), r)
ranges = list(unique_ranges.values())
assert len(ranges) == len(addresses)
for r in ranges:
    assert hashlib.sha256(read(int(r['address'],16),r['size'])).hexdigest() == r['sha256']
cpu = Uc(UC_ARCH_X86, UC_MODE_32)
cpu.mem_map(base, size)
for _, virtual, raw_size, raw in sections:
    if raw_size:
        cpu.mem_write(base+virtual, blob[raw:raw+raw_size])
heap, stack, stop = 0x10000000, 0x20000000, 0x30000000
cpu.mem_map(heap, 0x10000)
cpu.mem_map(stack, 0x10000)
cpu.mem_map(stop, 0x1000)
cpu.mem_write(stop+0x100, b'\xdd\x05'+struct.pack('<I',stop+0x200)+b'\xc3')
v, a, ud, info, item = (heap+x for x in (0x100,0x1100,0x2100,0x4100,0x6100))
get = lambda p: struct.unpack('<I',cpu.mem_read(p,4))[0]
put = lambda p,x: cpu.mem_write(p,struct.pack('<I',x&0xffffffff))
bits = lambda x: struct.unpack('<I',struct.pack('<f',x))[0]
sealed = {int(r['address'],16):bytes(cpu.mem_read(int(r['address'],16),r['size'])) for r in ranges}
entries, services, rows, faults = Counter(), Counter(), [], []
current, events, forward_damage, forward_calls, wear_bits, block_calls = {}, [], 0xfeedface, 0, 0, 0

def hook(_cpu,address,_size,_):
    global forward_damage,forward_calls,wear_bits,block_calls
    if address in sealed:
        entries[hex(address)] += 1
    if address == 0x4e1a49:
        events.append('exclude') # Observe inline exclusions; original instructions continue.
    if address == 0x4e2025:
        events.append('god') # Observe original global read, never override it.
    if any(start <= address < start+len(code) for start,code in sealed.items()):
        return
    if stop <= address < stop+0x107:
        return
    sp = cpu.reg_read(UC_X86_REG_ESP)
    arg = lambda i:get(sp+4+4*i)
    result = 0
    services[hex(address)] += 1
    gate = current['gate']
    if address == 0x4ff350:
        assert arg(0) == v
        result = bool(get(v+340) & (1 << arg(1)))
    elif address == 0x40a5c0:
        if arg(0) == 2048:
            events.append('coop')
        else:
            assert arg(0) == 4096
            events.append('quest')
            result = current['scale'] is not None
    elif address == 0x4ddf30:
        assert arg(0) == v
    elif address == 0x4e6e50:
        assert [arg(0),arg(1)] == [v+56,get(v+124)]
        assert struct.unpack('<ff',cpu.mem_read(arg(2),8)) == ((23,29) if gate=='reflect' else (17,19))
        events.append('face')
        result = gate in ('reflect','shield-front')
    elif address == 0x50a020:
        assert arg(0) == v
        result = 21 if gate.startswith('shield-') else 0
    elif address == 0x501960:
        assert [arg(1),arg(2),arg(3)] == [v,0,0]
        assert arg(0) in (71,122,878)
        events.append('audio-'+str(arg(0)))
    elif address == 0x4e0a70:
        assert [arg(0),arg(1)] == [a,v]
        events.append('reflect')
    elif address == 0x4ec300:
        assert arg(0) == a
        events.append('clear-owner')
    elif address == 0x4ec290:
        assert [arg(0),arg(1)] == [v,a]
        events.append('set-owner')
    elif address in (0x419d40,0x4e40b0):
        if address == 0x419d40:
            events.append('balance')
            value = .5
        else:
            events.append('scale')
            value = current['scale']
        cpu.mem_write(stop+0x200,struct.pack('<d',value))
        cpu.reg_write(UC_X86_REG_EIP,stop+0x100)
        return
    elif address == 0x4e16d0:
        assert [arg(0),arg(1),arg(2),arg(3),arg(5)] == [item,v,a,a,1]
        wear_bits,block_calls = arg(4),block_calls+1
        events.append('shield')
    elif address == 0x4e2180:
        assert [arg(0),arg(1),arg(2),arg(3),arg(4)] == [v,a,a,current['damage']&0xffffffff,1]
        # Empty armor loop boundary; input verified, no fabricated wear outcome.
    elif address == 0x4e0b30:
        assert [arg(0),arg(1),arg(2),arg(4)] == [v,a,a,1]
        forward_damage,forward_calls = arg(3),forward_calls+1
        events.append('default')
        result = 1
    else:
        raise AssertionError('undeclared instruction '+hex(address)+' '+current['name'])
    cpu.reg_write(UC_X86_REG_EAX,int(result))
    cpu.reg_write(UC_X86_REG_ESP,sp+4)
    cpu.reg_write(UC_X86_REG_EIP,get(sp))

def invalid(_cpu,access,address,size,value,_):
    faults.append(dict(case=current['name'],address=address,size=size,
        instruction=cpu.reg_read(UC_X86_REG_EIP)))
    print('ORIGINAL_MEMORY_FAULT '+json.dumps(faults[-1],separators=(',',':')),file=sys.stderr)
    return False

cpu.hook_add(UC_HOOK_CODE,hook)
cpu.hook_add(UC_HOOK_MEM_INVALID,invalid)
gates = ['normal','reflect','no-update','dead','invulnerable','invulnerable-mod3',
    'god','shield-front','shield-back','observer-disabled']
for target,type_id,target_class,sub in [('NewPlayer',713,2621444,0),('NPC',1343,2097154,69650)]:
    player = target == 'NewPlayer'
    for gate in gates:
        if not player and gate == 'observer-disabled':
            continue
        for damage in [-3,0,3,96]:
            for scale in [None,0,.5]:
                name = f'{target}/{gate}/raw-{damage}/scale-{scale}'
                current = dict(name=name,target=target,type=type_id,class_bits=target_class,
                    subclass=sub,gate=gate,damage=damage,scale=scale)
                events,forward_damage,forward_calls,wear_bits,block_calls = [],0xfeedface,0,0,0
                cpu.mem_write(heap,bytes([0xa7])*0x10000)
                for pointer,size_zero in [(v,0x300),(a,0x300),(ud,0xa00),(info,0xf00),(item,0x300)]:
                    cpu.mem_write(pointer,bytes(size_zero))
                flags = 16777732 | (2 if gate=='no-update' else 0) | (0x8000 if gate=='dead' else 0)
                buffs = (1 << 27 if gate=='reflect' else 0) | (1 << 23 if gate.startswith('invulnerable') else 0)
                frame = 1401 if gate=='invulnerable-mod3' else 1400
                cpu.mem_write(v+4,struct.pack('<H',type_id))
                put(v+8,target_class);put(v+12,sub);put(v+16,flags);put(v+340,buffs);put(v+748,ud)
                put(ud+276,info)
                marker,kind,carry,armor = (304,300,84,228) if player else (2188,2184,4,2072)
                put(ud+marker,99);put(ud+kind,91);put(ud+carry,bits(.25));put(ud+armor,bits(.5))
                cpu.mem_write(ud+88,bytes([16 if gate.startswith('shield-') else 13]))
                if gate.startswith('shield-'):
                    put(info if player else ud+2060,0x1000000)
                    put(v+504,item);put(item+8,0x2000000);put(item+12,2);put(item+16,0x100)
                if gate=='observer-disabled':
                    put(info+3680,1)
                cpu.mem_write(a+4,struct.pack('<H',694))
                put(a+8,2621441);put(a+12,1);put(a+16,553665028)
                cpu.mem_write(a+56,struct.pack('<ff',23,29))
                cpu.mem_write(a+72,struct.pack('<ff',17,19))
                for i in range(6):
                    put(0x75204c+4*i,900+i)
                put(0x84ea04,frame);put(0x85b7a0,0x20 if gate=='god' else 0)
                initial_heap = bytes(cpu.mem_read(heap,0x10000))
                initial_sp = stack+0x8000
                cpu.mem_write(initial_sp,struct.pack('<IIIIII',stop,v,a,a,damage&0xffffffff,1))
                cpu.reg_write(UC_X86_REG_ESP,initial_sp)
                cpu.reg_write(UC_X86_REG_FPCW,control_word);cpu.reg_write(UC_X86_REG_FPSW,0);cpu.reg_write(UC_X86_REG_FPTAG,0xffff)
                cpu.emu_start(0x4e17b0,stop,count=12000)
                assert cpu.reg_read(UC_X86_REG_EIP)==stop and cpu.reg_read(UC_X86_REG_ESP)==initial_sp+4
                guarded = bytearray(cpu.mem_read(heap,0x10000))
                for off in (marker,kind):
                    index = ud-heap+off
                    guarded[index:index+4] = initial_heap[index:index+4]
                assert bytes(guarded)==initial_heap
                rows.append(dict(**current,flags=flags,buffs=buffs,frame=frame,
                    result=cpu.reg_read(UC_X86_REG_EAX),marker=get(ud+marker),marker_type=get(ud+kind),
                    carry=get(ud+carry),armor=get(ud+armor),state=cpu.mem_read(ud+88,1)[0] if player else 0,
                    forward_damage=forward_damage,forward_calls=forward_calls,wear_bits=wear_bits,
                    block_calls=block_calls,events=events.copy(),source_and_guards_unchanged=True))
assert len(rows)==228 and not faults
assert all(bytes(cpu.mem_read(start,len(code)))==code for start,code in sealed.items())
print(json.dumps(dict(original_sha256=manifest['file_sha256'],
    control_word=control_word,
    function_sha256=hashlib.sha256(read(0x4e17b0,2368)).hexdigest(),
    original_ranges=ranges,source=dict(type=694,class_bits=2621441,subclass=1,flags=553665028),
    rows=rows,unhooked_entries=dict(entries),declared_service_entries=dict(services),
    memory_faults=faults,executable_bytes_unchanged=True,scope=__doc__.strip()),separators=(',',':')))
