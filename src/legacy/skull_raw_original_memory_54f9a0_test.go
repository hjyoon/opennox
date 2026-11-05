package legacy

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type skullRawField54F9A0 struct{ pe, native, width int }

// Only original destinations are projected from native execution. Read-only
// object fields are initialized independently and guarded in their entirety.
type skullRawMemory54F9A0 struct {
	t                   *testing.T
	heap                unsafe.Pointer
	page                uintptr
	buf, before, pe     []byte
	skull, collision    map[int]bool
	outputs             []skullRawField54F9A0
	objectFields        []skullRawField54F9A0
	updateAt, collideAt int
}

func (m *skullRawMemory54F9A0) pointer(role int) unsafe.Pointer {
	if role < 0 {
		if role != -1 && role != -2 {
			m.t.Fatal("undeclared bad role", role)
		}
		return simplePolypOriginalBad54CAE0(m.heap, m.page, role)
	}
	if role >= 8 {
		m.t.Fatal("undeclared mapped role", role)
	}
	return unsafe.Add(m.heap, (role+1)*0x1000)
}

func (m *skullRawMemory54F9A0) role(p unsafe.Pointer) int {
	for n := -2; n < 8; n++ {
		if m.pointer(n) == p {
			return n
		}
	}
	m.t.Fatalf("undeclared native pointer %p", p)
	return -99
}

func skullRawPEPointer54F9A0(role int) uint32 {
	if role == -1 {
		return 0
	}
	if role == -2 {
		return 0x41000000
	}
	return 0x21000000 + uint32((role+1)*0x1000)
}

func skullRawPut54F9A0(buf []byte, at, width int, value uint64) {
	for n := 0; n < width; n++ {
		buf[at+n] = byte(value >> (8 * n))
	}
}

func (m *skullRawMemory54F9A0) setPointer(role, offset, target int) {
	at := (role + 1) * 0x1000
	native := m.updateAt
	if offset == 700 {
		native = m.collideAt
	} else if offset != 748 {
		m.t.Fatal("undeclared pointer field", offset)
	}
	skullRawPut54F9A0(m.pe, at+offset, 4, uint64(skullRawPEPointer54F9A0(target)))
	skullRawPut54F9A0(m.buf, at+native, int(unsafe.Sizeof(uintptr(0))), uint64(uintptr(m.pointer(target))))
}

func (m *skullRawMemory54F9A0) setWord(role, offset, width int, value uint64) {
	at := (role + 1) * 0x1000
	skullRawPut54F9A0(m.pe, at+offset, width, value)
	native := offset
	if role < 4 {
		found := false
		for _, f := range m.objectFields {
			if offset >= f.pe && offset+width <= f.pe+f.width {
				native, found = f.native+offset-f.pe, true
				break
			}
		}
		if !found {
			m.t.Fatal("undeclared object mutation", role, offset, width)
		}
		// The frozen corpus never combines an object-field callback mutation
		// with a conflicting raw-data alias. This is declared correspondence,
		// not a general native named-field alias certificate.
		if m.skull[role] || m.collision[role] {
			m.t.Fatal("additional alias mutation correspondence required", role, offset)
		}
	}
	skullRawPut54F9A0(m.buf, at+native, width, value)
}

func (m *skullRawMemory54F9A0) checkpoint() string {
	expected := bytes.Clone(m.before)
	for role := 0; role < 4; role++ {
		at := (role + 1) * 0x1000
		for _, f := range m.outputs {
			copy(expected[at+f.native:at+f.native+f.width], m.buf[at+f.native:at+f.native+f.width])
			copy(m.pe[at+f.pe:at+f.pe+f.width], m.buf[at+f.native:at+f.native+f.width])
		}
	}
	for role := 0; role < 8; role++ {
		at := (role + 1) * 0x1000
		var fields []skullRawField54F9A0
		if m.skull[role] {
			fields = append(fields, skullRawField54F9A0{0, 0, 4}, skullRawField54F9A0{4, 4, 4}, skullRawField54F9A0{8, 8, 1}, skullRawField54F9A0{48, 48, 1})
		}
		if m.collision[role] {
			fields = append(fields, skullRawField54F9A0{0, 0, 4}, skullRawField54F9A0{4, 4, 4})
		}
		for _, f := range fields {
			copy(expected[at+f.native:at+f.native+f.width], m.buf[at+f.native:at+f.native+f.width])
			copy(m.pe[at+f.pe:at+f.pe+f.width], m.buf[at+f.native:at+f.native+f.width])
		}
	}
	if !bytes.Equal(expected, m.buf) {
		m.t.Fatal("complete64KiB native heap changed outside original destinations and declared callbacks")
	}
	m.before = bytes.Clone(m.buf)
	return fmt.Sprintf("%x", sha256.Sum256(m.pe))
}

func (m *skullRawMemory54F9A0) acceptWorldWrites() { m.before = bytes.Clone(m.buf) }

func newSkullRawMemory54F9A0(t *testing.T, in skullRawInput54F9A0) (*skullRawMemory54F9A0, func()) {
	t.Helper()
	heap, page, free := simplePolypOriginalHeap54CAE0()
	if heap == nil || page <= unsafe.Sizeof(server.Object{}) {
		t.Fatal("complete native C arena and guard page")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(heap) <= 0xffffffff {
		t.Fatal("native arena must exceed4GiB")
	}
	m := &skullRawMemory54F9A0{t: t, heap: heap, page: page, buf: unsafe.Slice((*byte)(heap), 0x10000), skull: map[int]bool{4: true, 6: true}, collision: map[int]bool{5: true, 7: true}}
	if in.Update >= 0 {
		m.skull[in.Update] = true
	}
	if in.Collide >= 0 {
		m.collision[in.Collide] = true
	}
	for i := range m.buf {
		m.buf[i] = byte(i*17 + (i>>7)*29)
	}
	m.pe = bytes.Clone(m.buf)
	o := server.Object{}
	m.updateAt, m.collideAt = int(unsafe.Offsetof(o.UpdateData)), int(unsafe.Offsetof(o.CollideData))
	m.objectFields = []skullRawField54F9A0{
		{4, int(unsafe.Offsetof(o.TypeInd)), 2}, {8, int(unsafe.Offsetof(o.ObjClass)), 4},
		{16, int(unsafe.Offsetof(o.ObjFlags)), 4}, {56, int(unsafe.Offsetof(o.PosVec)), 4}, {60, int(unsafe.Offsetof(o.PosVec)) + 4, 4},
		{124, int(unsafe.Offsetof(o.Direction1)), 2}, {126, int(unsafe.Offsetof(o.Direction2)), 2},
		{176, int(unsafe.Offsetof(o.Shape)) + int(unsafe.Offsetof(o.Shape.Circle)), 4}, {544, int(unsafe.Offsetof(o.SpeedCur)), 4},
		{80, int(unsafe.Offsetof(o.VelVec)), 4}, {84, int(unsafe.Offsetof(o.VelVec)) + 4, 4},
	}
	m.outputs = append(m.outputs, m.objectFields[5], m.objectFields[6], m.objectFields[9], m.objectFields[10])
	for role := 0; role < 4; role++ {
		at := (role + 1) * 0x1000
		flags := in.Flags
		if role == 1 {
			flags = in.TargetFlags
		}
		for _, field := range []struct {
			offset, width int
			value         uint32
		}{
			{4, 2, uint32(in.SourceType)}, {8, 4, in.TargetClass}, {16, 4, flags},
			{56, 4, in.Position[0] + uint32(role)}, {60, 4, in.Position[1] + uint32(role)},
			{124, 2, uint32(in.Direction)}, {176, 4, in.RadiusBits}, {544, 4, in.SpeedBits},
		} {
			skullRawPut54F9A0(m.pe, at+field.offset, field.width, uint64(field.value))
		}
		update := in.Update
		if role == 3 {
			update = 6
		}
		skullRawPut54F9A0(m.pe, at+748, 4, uint64(skullRawPEPointer54F9A0(update)))
		skullRawPut54F9A0(m.pe, at+700, 4, uint64(skullRawPEPointer54F9A0(in.Collide)))
	}
	for _, field := range []struct {
		role, offset, width int
		value               uint32
	}{
		{4, 0, 4, in.Timers[0]}, {4, 4, 4, in.Timers[1]}, {4, 8, 1, uint32(in.State)},
		{4, 12, 4, in.ProjectileType}, {4, 48, 1, uint32(in.Enabled)},
		{6, 0, 4, 17}, {6, 4, 4, 19}, {6, 8, 1, 1}, {6, 12, 4, 73}, {6, 48, 1, 1},
	} {
		skullRawPut54F9A0(m.pe, (field.role+1)*0x1000+field.offset, field.width, uint64(field.value))
	}
	copy(m.buf, m.pe)
	for role := 0; role < 4; role++ {
		at := (role + 1) * 0x1000
		for _, f := range m.objectFields {
			copy(m.buf[at+f.native:at+f.native+f.width], m.pe[at+f.pe:at+f.pe+f.width])
		}
		update := in.Update
		if role == 3 {
			update = 6
		}
		m.setPointer(role, 748, update)
		m.setPointer(role, 700, in.Collide)
	}
	// Raw data aliases have their declared fixed-width view. Read-only native
	// object fields can physically overlap it; the frozen corpus observes no
	// conflicting subsequent object-field read in these malformed cases.
	// Whole native bytes remain guarded, and no arbitrary alias claim is made.
	for role := 0; role < 4; role++ {
		at := (role + 1) * 0x1000
		fields := []skullRawField54F9A0{}
		if m.skull[role] {
			fields = append(fields, skullRawField54F9A0{0, 0, 4}, skullRawField54F9A0{4, 4, 4}, skullRawField54F9A0{8, 8, 1}, skullRawField54F9A0{12, 12, 4}, skullRawField54F9A0{48, 48, 1})
		}
		if m.collision[role] {
			fields = append(fields, skullRawField54F9A0{0, 0, 4}, skullRawField54F9A0{4, 4, 4})
		}
		for _, f := range fields {
			copy(m.buf[at+f.native:at+f.native+f.width], m.pe[at+f.pe:at+f.pe+f.width])
		}
	}
	for _, plan := range in.Script {
		for _, p := range plan.Pointers {
			if len(p) != 3 {
				t.Fatal("pointer mutation tuple")
			}
			if p[2] >= 0 {
				if p[1] == 748 {
					m.skull[int(p[2])] = true
				} else if p[1] == 700 {
					m.collision[int(p[2])] = true
				}
			}
		}
	}
	m.before = bytes.Clone(m.buf)
	return m, free
}

func (m *skullRawMemory54F9A0) object(role int) *server.Object {
	return (*server.Object)(m.pointer(role))
}
func (m *skullRawMemory54F9A0) word(role, offset int) uint32 {
	return binary.LittleEndian.Uint32(m.pe[(role+1)*0x1000+offset:])
}

func TestSkullRawInitialMemoryCorrespondence54F9A0(t *testing.T) {
	TestSkullRawNativeLayout54F9A0(t)
	var inputs []skullRawInput54F9A0
	trollDeathFixture54A270(t, "skull_raw_independent_inputs_54f9a0.json.gz", "2fa912b92071a20903b0596adce24a80942874efeb6e80fe5b8d8f6def6dcc36", "8a004ff539f042e37bfa144916780ab9042d5a6ed48889229bb6d6e18bce8825", &inputs)
	var expected []struct {
		Name string
		Heap string `json:"heap_sha256"`
	}
	trollDeathFixture54A270(t, "skull_raw_initial_heap_54f9a0.json.gz", "b4c2147a417a21e7583c8a049ab57c9c9d337f2eb5a9fe9b8a0f6eb0aebf3e7c", "5fe35c7ac9568302f4a2cf4b4dde164a1662e752f56bc826309f05ef8d2f6e23", &expected)
	if len(inputs) != 753 || len(expected) != 753 {
		t.Fatal("complete independent initial memory domain")
	}
	for i, in := range inputs {
		t.Run(in.Name, func(t *testing.T) {
			if expected[i].Name != in.Name {
				t.Fatal("independent memory input identity")
			}
			memory, free := newSkullRawMemory54F9A0(t, in)
			defer free()
			before := bytes.Clone(memory.buf)
			if got := memory.checkpoint(); got != expected[i].Heap {
				t.Fatal("complete initial PE/native declared correspondence", got, expected[i].Heap)
			}
			if !bytes.Equal(before, memory.buf) {
				t.Fatal("initial correspondence control changed native bytes")
			}
		})
	}
}
