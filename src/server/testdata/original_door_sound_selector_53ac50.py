"""Execute the two unchanged sound-selection blocks of GAME.EXE 0053AC50.

This is a sound-selector proof, not whole door-motion/collision execution.
AudioEvent and removal from the update list are declared call boundaries.
"""
from pathlib import Path
import hashlib, itertools, json, struct
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE
from unicorn.x86_const import *

r = Path(__file__).resolve().parent
game = Path('/Users/ghost/codex-sandbox/tmp/opennox-warrior-r318-20261019/save-load-v2-normal-virtual/runtime-data/GAME.EXE')
blob = game.read_bytes()
sha = lambda raw: hashlib.sha256(raw).hexdigest()
assert sha(blob) == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe = struct.unpack_from('<I', blob, 60)[0]
op = pe + 24
table = op + struct.unpack_from('<H', blob, pe+20)[0]
base = struct.unpack_from('<I', blob, op+28)[0]
u = Uc(UC_ARCH_X86, UC_MODE_32)
u.mem_map(base, (struct.unpack_from('<I', blob, op+56)[0]+4095)&~4095)
for i in range(struct.unpack_from('<H', blob, pe+6)[0]):
    _, va, count, off = struct.unpack_from('<IIII', blob, table+i*40+8)
    if count:
        u.mem_write(base+va, blob[off:off+count])
heap, stack, harness = 0x21000000, 0x22000000, 0x23000000
for address in [heap, stack, harness]:
    u.mem_map(address, 0x10000)
unit, sp = heap+0x1000, stack+0xe000
seed = bytes((i*31+19)&255 for i in range(0x10000))
ranges = {0x53ac6f:0x53ace9-0x53ac6f, 0x53acf2:0x53ad4a-0x53acf2}
original = {pc:bytes(u.mem_read(pc,n)) for pc,n in ranges.items()}
traces, events, current = {}, [], None
rd = lambda address: struct.unpack('<I',u.mem_read(address,4))[0]
wr = lambda address,value: u.mem_write(address,struct.pack('<I',value&0xffffffff))
u.mem_write(harness,b'\xdb\xe3\xd9\x2d'+struct.pack('<I',harness+0x100))

def hook(_u, pc, size, _):
    if harness <= pc < harness+8:
        return
    if pc == 0x53ad4a:
        u.emu_stop()
        return
    if any(start <= pc < start+n for start,n in ranges.items()):
        traces[pc] = bytes(u.mem_read(pc,size))
        return
    assert pc in [0x501960,0x4da920], (current,hex(pc))
    esp = u.reg_read(UC_X86_REG_ESP)
    if pc == 0x501960:
        assert rd(esp+8) == unit and rd(esp+12) == rd(esp+16) == 0
        events.append({'kind':'audio','sound':rd(esp+4)})
    else:
        assert rd(esp+4) == unit
        events.append({'kind':'remove-updatable'})
    u.reg_write(UC_X86_REG_EAX,0x87654321)
    u.reg_write(UC_X86_REG_ECX,0xc1c2c3c4)
    u.reg_write(UC_X86_REG_EDX,0xd1d2d3d4)
    u.reg_write(UC_X86_REG_ESP,esp+4)
    u.reg_write(UC_X86_REG_EIP,rd(esp))

u.hook_add(UC_HOOK_CODE,hook)
profiles = []
for control in [0xe3f,0x23f]:
    rows = []
    for cls,subclass,material,reached in itertools.product(
        [0x80,0x84,1,4,0x1000,0xa5],
        [0,1,4,0x1000,5,0x1004,0x1005,0x80001000],
        [0,8,0x100,0x108],[False,True]):
        current = dict(cls=cls,subclass=subclass,material=material,reached=reached)
        events = []
        u.mem_write(heap,seed)
        wr(unit+8,cls)
        wr(unit+12,subclass)
        u.mem_write(unit+24,struct.pack('<H',material))
        before = bytes(u.mem_read(heap,len(seed)))
        wr(harness+0x100,control)
        u.emu_start(harness,harness+8,count=2)
        u.mem_write(sp-512,bytes(1024))
        registers = {UC_X86_REG_EBX:0x12345678,UC_X86_REG_ESI:unit,
                     UC_X86_REG_EDI:0x3456789a,UC_X86_REG_EBP:0x456789ab}
        for reg,value in registers.items():
            u.reg_write(reg,value)
        u.reg_write(UC_X86_REG_ESP,sp)
        u.reg_write(UC_X86_REG_EFLAGS,0x202)
        start = 0x53acf2 if reached else 0x53ac6f
        u.emu_start(start,0x53ad4b,count=200)
        assert u.reg_read(UC_X86_REG_EIP) == 0x53ad4a
        assert u.reg_read(UC_X86_REG_ESP) == sp
        assert all(u.reg_read(reg) == value for reg,value in registers.items())
        assert bytes(u.mem_read(heap,len(seed))) == before
        assert bytes(u.mem_read(sp-512,256)) == bytes(256)
        assert bytes(u.mem_read(sp,256)) == bytes(256)
        assert u.reg_read(UC_X86_REG_FPCW) == control and u.reg_read(UC_X86_REG_FPSW)&0x3800 == 0
        assert len(events) == (2 if reached else 1)
        assert events[0]['kind'] == 'audio'
        if reached:
            assert events[1]['kind'] == 'remove-updatable'
        rows.append(dict(input=current,sound=events[0]['sound'],events=events.copy()))
    assert len(rows) == 384
    profiles.append(rows)
assert profiles[0] == profiles[1]
visited = {pc+i for pc,raw in traces.items() for i in range(len(raw))}
expected = {pc+i for pc,n in ranges.items() for i in range(n)}
assert visited == expected
assert all(bytes(u.mem_read(pc,len(raw))) == raw for pc,raw in original.items())
reference = dict(original_sha256=sha(blob),rows=profiles[0])
proof = dict(passed=True,original_sha256=sha(blob),contexts_per_profile=384,total_contexts=768,
    profiles=[0xe3f,0x23f],profile_differences=0,executed_selector_bytes=len(visited),
    selected_blocks={hex(pc):dict(bytes=len(raw),sha256=sha(raw)) for pc,raw in original.items()},
    original_type_load_offset=12,original_material_load_offset=24,
    declared_services=['AudioEvent501960','RemoveFromUpdatable4DA920'],
    register_stack_FPCW_TOP_and_heap_canaries_preserved=True,original_code_unmodified=True,
    whole_door_motion_collision_or_audio_playback_execution=False)
for name,value in [('original-door-sound-selector-v1.json',reference),('original-door-sound-selector-qualified-v1.json',proof)]:
    with (r/name).open('x') as f:
        json.dump(value,f,indent=2)
        f.write('\n')
print(json.dumps(proof))
