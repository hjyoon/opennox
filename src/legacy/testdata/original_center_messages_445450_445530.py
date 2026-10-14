"""Execute the two complete original HUD message functions.

Viewport, font metrics and drawing are declared services. UTF-16 copying runs
the unhooked original 565CA3 body. This is not a Windows font/pixel oracle.
"""
import hashlib
import json
import struct
import sys
from pathlib import Path
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import UC_X86_REG_EAX, UC_X86_REG_EIP, UC_X86_REG_ESP

pe = Path(sys.argv[1]).read_bytes()
assert hashlib.sha256(pe).hexdigest() == "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a"
h = struct.unpack_from("<I", pe, 0x3c)[0]
optional = h + 24
base = struct.unpack_from("<I", pe, optional + 28)[0]
size = struct.unpack_from("<I", pe, optional + 56)[0]
sections = optional + struct.unpack_from("<H", pe, h + 20)[0]
cpu = Uc(UC_ARCH_X86, UC_MODE_32)
cpu.mem_map(base, (size + 4095) & ~4095)
for i in range(struct.unpack_from("<H", pe, h + 6)[0]):
    _, virtual, length, raw = struct.unpack_from("<IIII", pe, sections + i*40 + 8)
    cpu.mem_write(base + virtual, pe[raw:raw+length])
cpu.mem_map(0x20000000, 0x20000)
STACK, STOP, VIEW = 0x20008000, 0x20009000, 0x20010000
CENTER, INDEX, EMPTY, OFFSETS = 0x69d790, 0x69df1c, 0x69df20, 0x5a1548
FRAME, WIDTH, BLACK, WHITE, OLDER = 0x84ea04, 0x9756cc, 0x85b7b4, 0x83c8c0, 0x84ea00
ranges = {0x445450: 62, 0x445530: 287, 0x565ca3: 37}
code = {address: bytes(cpu.mem_read(address, length)) for address, length in ranges.items()}
assert all(data[-1] == 0xc3 for data in code.values())
offsets = bytes(cpu.mem_read(OFFSETS, 32))
shadow_offsets = list(struct.unpack("<8i", offsets))
services = {0x437250: ("viewport", 0), 0x434390: ("color", 1),
            0x43f840: ("measure", 5), 0x43f6e0: ("draw", 4), 0x43f320: ("height", 1)}
events, faults, current = [], [], {}
counts = {hex(address): 0 for address in ranges}

def word(at):
    return struct.unpack("<I", cpu.mem_read(at, 4))[0]

def put(at, value):
    cpu.mem_write(at, struct.pack("<I", value & 0xffffffff))

def signed(value):
    return value if value < 0x80000000 else value - 0x100000000

def kind(address):
    delta = address - CENTER
    assert delta in (0, 644, 1288), "wrong message text address"
    return delta // 644

def hook(_cpu, address, length, _):
    if address in ranges:
        counts[hex(address)] += 1
        if address == 0x565ca3:
            sp = cpu.reg_read(UC_X86_REG_ESP)
            assert word(sp+8) == EMPTY
            events.append(["copy", kind(word(sp+4))])
    if any(start <= address < start+len(data) for start, data in code.items()) or address == STOP:
        return
    name, count = services[address]
    sp = cpu.reg_read(UC_X86_REG_ESP)
    args = list(struct.unpack("<"+"I"*count, cpu.mem_read(sp+4, count*4))) if count else []
    result = 0
    if name == "viewport":
        events.append([name])
        result = VIEW
    elif name == "measure":
        assert args[0] == args[3] == args[4] == 0
        row = kind(args[1])
        events.append([name, row])
        put(args[2], current["measured"] + 8*row)
        result = 1
    elif name == "draw":
        assert args[0] == 0
        events.append([name, kind(args[1]), signed(args[2]), signed(args[3])])
        result = 0xdeadbeef
    elif name == "height":
        assert args[0] == 0
        events.append([name])
        result = current["font"]
    else:
        events.append([name, args[0]])
    cpu.reg_write(UC_X86_REG_EAX, result & 0xffffffff)
    cpu.reg_write(UC_X86_REG_EIP, word(sp))
    cpu.reg_write(UC_X86_REG_ESP, sp+4)

def invalid(_cpu, access, address, length, value, _):
    faults.append(dict(address=address, pc=cpu.reg_read(UC_X86_REG_EIP)))
    return False

cpu.hook_add(UC_HOOK_CODE, hook)
cpu.hook_add(UC_HOOK_MEM_INVALID, invalid)

def begin(pattern):
    global events
    events = []
    cpu.mem_write(STACK-256, bytes([pattern])*512)
    put(STACK, STOP)
    cpu.reg_write(UC_X86_REG_ESP, STACK)

def execute(address):
    cpu.emu_start(address, STOP, count=10000)
    assert cpu.reg_read(UC_X86_REG_EIP) == STOP and cpu.reg_read(UC_X86_REG_ESP) == STACK+4
    assert not faults and all(bytes(cpu.mem_read(at, len(data))) == data for at, data in code.items())
    assert bytes(cpu.mem_read(OFFSETS, 32)) == offsets

clear_rows = []
for index in range(3):
    for text in ([], [0x41], [0xac00, 0xd83d, 0xde80, 0x42]):
        observed = []
        for pattern in (0, 0x5a, 0xff):
            cpu.mem_write(CENTER-32, bytes([pattern])*(1932+64))
            source = struct.pack("<"+"H"*(len(text)+1), *(text+[0]))
            cpu.mem_write(EMPTY, source)
            put(INDEX, index)
            before = bytes(cpu.mem_read(CENTER-32, 1932+64))
            begin(pattern)
            execute(0x445450)
            after = bytes(cpu.mem_read(CENTER-32, 1932+64))
            expected = bytearray(before)
            for row in range(3):
                at = 32+644*row
                expected[at:at+len(source)] = source
                expected[at+636:at+641] = bytes(5)
            expected[32+1932:32+1932+4] = bytes(4)  # Original final INDEX store.
            assert after == expected and word(INDEX) == 0
            assert cpu.reg_read(UC_X86_REG_EAX) == CENTER+1288
            observed.append(dict(events=events, returned_slot=2, index=0))
        assert observed[0] == observed[1] == observed[2]
        clear_rows.append(dict(input=dict(index=index, text=text), expected=observed[0]))

draw_rows = []
for index in range(3):
    for mask in range(8):
        for frame in (0, 37, 0x80000000, 0xffffffff):
            for width, height, top, measured, font in ((640,480,0,19,13), (641,481,7,20,0), (-25,-31,-11,23,7)):
                current = dict(index=index, mask=mask, frame=frame, width=width, height=height,
                               top=top, measured=measured, font=font)
                observed = []
                for pattern in (0, 0x5a, 0xff):
                    cpu.mem_write(CENTER-32, bytes([pattern])*(1932+64))
                    for row in range(3):
                        cpu.mem_write(CENTER+644*row, struct.pack("<HH", 0x41+row, 0))
                        put(CENTER+644*row+636, frame if mask & (1<<row) else (frame-1)&0xffffffff)
                    put(INDEX,index);put(FRAME,frame);put(WIDTH,width)
                    put(BLACK,0);put(WHITE,0x7fff);put(OLDER,0x03e0)
                    put(VIEW+4,top);put(VIEW+36,height)
                    before = bytes(cpu.mem_read(CENTER-32,1932+64))
                    begin(pattern)
                    execute(0x445530)
                    assert bytes(cpu.mem_read(CENTER-32,1932+64)) == before and word(INDEX) == index
                    observed.append(dict(events=events, returned=cpu.reg_read(UC_X86_REG_EAX)))
                assert observed[0] == observed[1] == observed[2]
                draw_rows.append(dict(input=current, expected=observed[0]))
print(json.dumps(dict(original_sha256=hashlib.sha256(pe).hexdigest(),
    code=[dict(entry=hex(at), bytes=len(data), sha256=hashlib.sha256(data).hexdigest()) for at,data in code.items()],
    entries=counts, faults=faults, code_table_message_guards_unchanged=True,
    shadow_offsets=shadow_offsets, clear_rows=clear_rows, draw_rows=draw_rows,
    scope="complete original HUD entries and unhooked UTF-16 copy; viewport/font/draw declared services, not Windows pixels"),separators=(",",":")))
