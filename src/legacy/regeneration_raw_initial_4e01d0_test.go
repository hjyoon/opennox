package legacy

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"math"
	"os"
	"os/exec"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

const regenerationRawPEBase4E01D0 = uint32(0x21000000)

type regenerationRawInitial4E01D0 struct {
	Name   string
	Seed   int
	Roles  map[string]uint32
	Shapes map[string]string
	Words  []struct {
		Address uint32
		Width   int
		Value   uint32
	} `json:"initializer_words"`
	Fields []struct {
		Role    uint32
		Address uint32
		Kind    int
		Type    string
		PEWidth int `json:"pe_width"`
		Value   uint32
	} `json:"native_fields"`
	Reads []struct {
		Role uint32
		Kind int
		Want *uint32
	} `json:"independent_reads"`
	InitialSHA string `json:"initial_heap_sha256"`
}

func regenerationRawInitials4E01D0(t *testing.T) []regenerationRawInitial4E01D0 {
	t.Helper()
	var rows []regenerationRawInitial4E01D0
	modifierPredicateFixture4E03D0(t, "regeneration_raw_initial_memory_4e01d0.json.gz", "0eda2dacd7e98102d8aa922088c5c0944d9d38cf00160304e788ee57c4066427", "0a06fbc35a302a823f96c3c188efa2f200a8cd33750953ce4370de5dda7cc652", &rows)
	if len(rows) != 198 {
		t.Fatal("all independent raw Regeneration inputs required")
	}
	return rows
}

type regenerationRawSlot4E01D0 struct {
	pe, native, peWidth, width int
	pointer                    bool
}

type regenerationRawMemory4E01D0 struct {
	t         *testing.T
	heap      unsafe.Pointer
	layout    [11]uintptr
	buf, pe   []byte
	before    []byte
	addresses map[uint32]unsafe.Pointer
	reverse   map[uintptr]uint32
	slots     []regenerationRawSlot4E01D0
	faults    int
}

func (m *regenerationRawMemory4E01D0) add(pe uint32, pointer unsafe.Pointer) {
	m.t.Helper()
	if old, ok := m.addresses[pe]; ok && old != pointer {
		m.t.Fatal("contradictory independent Regeneration pointer correspondence", pe)
	}
	m.addresses[pe], m.reverse[uintptr(pointer)] = pointer, pe
}

func (m *regenerationRawMemory4E01D0) fromPE(pe uint32) unsafe.Pointer {
	m.t.Helper()
	value, ok := m.addresses[pe]
	if !ok {
		m.t.Fatalf("undeclared initial Regeneration pointer %#x", pe)
	}
	return value
}

func (m *regenerationRawMemory4E01D0) toPE(pointer uintptr) uint32 {
	m.t.Helper()
	value, ok := m.reverse[pointer]
	if !ok {
		m.t.Fatalf("unprojected native Regeneration pointer %#x", pointer)
	}
	return value
}

func newRegenerationRawMemory4E01D0(t *testing.T, row regenerationRawInitial4E01D0) *regenerationRawMemory4E01D0 {
	t.Helper()
	heap, guard, free := regenerationRawHeap4E01D0()
	t.Cleanup(free)
	layout := regenerationRawLayout4E01D0()
	if heap == nil || guard < layout[8] || (layout[10] == 8 && uintptr(heap) <= math.MaxUint32) {
		t.Fatal("complete native high heap with protected predecessor and successor required")
	}
	m := &regenerationRawMemory4E01D0{t: t, heap: heap, layout: layout, buf: unsafe.Slice((*byte)(heap), 65536), pe: make([]byte, 65536), addresses: make(map[uint32]unsafe.Pointer), reverse: make(map[uintptr]uint32)}
	for i := range m.pe {
		m.pe[i] = byte((i*37 + row.Seed*19) ^ (i >> 3))
	}
	for _, word := range row.Words {
		offset := int(word.Address - regenerationRawPEBase4E01D0)
		for i := 0; i < word.Width; i++ {
			m.pe[offset+i] = byte(word.Value >> (8 * i))
		}
	}
	if fmt.Sprintf("%x", sha256.Sum256(m.pe)) != row.InitialSHA {
		t.Fatal("complete independent PE32 initial heap", row.Name)
	}
	copy(m.buf, m.pe)
	m.add(0, nil)
	m.add(0x41000000, regenerationRawBad4E01D0(heap, guard))
	for _, role := range []string{"item", "owner", "health", "effect"} {
		pe := row.Roles[role]
		if _, ok := m.addresses[pe]; ok {
			continue
		}
		offset := uintptr(pe - regenerationRawPEBase4E01D0)
		switch row.Shapes[role] {
		case "last-holder":
			offset = 65536 - layout[0] - layout[10]
		case "last-health":
			offset = 65536 - layout[2] - layout[10]
		case "last-current":
			offset = 65536 - layout[6] - 2
		case "last-maximum":
			offset = 65536 - layout[7] - 2
		case "last-rate":
			offset = 65536 - layout[5] - 4
		}
		m.add(pe, unsafe.Add(heap, offset))
	}
	for _, field := range row.Fields {
		p := unsafe.Add(m.fromPE(field.Role), layout[field.Kind])
		width := field.PEWidth
		if field.Type == "pointer" {
			width = int(layout[10])
		} else if field.Type != "word" && field.Type != "half" {
			t.Fatal("undeclared independent field type")
		}
		slot := regenerationRawSlot4E01D0{int(field.Address - regenerationRawPEBase4E01D0), int(uintptr(p) - uintptr(heap)), field.PEWidth, width, field.Type == "pointer"}
		if slot.native < 0 || slot.native+width > len(m.buf) {
			t.Fatal("independent readable PE field must have a readable native correspondence")
		}
		for _, previous := range m.slots {
			if slot.native < previous.native+previous.width && previous.native < slot.native+slot.width {
				t.Fatal("overlapping independent native fields")
			}
		}
		m.add(field.Address, p)
		m.slots = append(m.slots, slot)
	}
	for i, field := range row.Fields {
		slot := m.slots[i]
		value := uint64(field.Value)
		if slot.pointer {
			value = uint64(uintptr(m.fromPE(field.Value)))
		}
		for j := 0; j < slot.width; j++ {
			m.buf[slot.native+j] = byte(value >> (8 * j))
		}
	}
	m.before = bytes.Clone(m.buf)
	if fmt.Sprintf("%x", sha256.Sum256(m.project())) != row.InitialSHA {
		t.Fatal("whole native-to-PE32 initial projection")
	}
	return m
}

func (m *regenerationRawMemory4E01D0) project() []byte {
	m.t.Helper()
	result, allowed := bytes.Clone(m.pe), make([]bool, len(m.buf))
	for _, slot := range m.slots {
		var value uint64
		for i := 0; i < slot.width; i++ {
			value |= uint64(m.buf[slot.native+i]) << (8 * i)
			allowed[slot.native+i] = true
		}
		if slot.pointer {
			value = uint64(m.toPE(uintptr(value)))
		}
		for i := 0; i < slot.peWidth; i++ {
			result[slot.pe+i] = byte(value >> (8 * i))
		}
	}
	for i, value := range m.buf {
		if !allowed[i] && value != m.before[i] {
			m.t.Fatal("unrelated native heap byte changed", i)
		}
	}
	return result
}

func (m *regenerationRawMemory4E01D0) initialGo(pointer unsafe.Pointer, kind int) (result uint32, fault bool) {
	fault = modifierPredicateRecover4E03D0(m.t, func() {
		switch kind {
		case 0:
			result = m.toPE(uintptr(unsafe.Pointer((*server.Object)(pointer).InvHolder)))
		case 1:
			result = uint32((*server.Object)(pointer).ObjClass)
		case 2:
			result = m.toPE(uintptr(unsafe.Pointer((*server.Object)(pointer).HealthData)))
		case 3:
			result = (*server.Object)(pointer).Frame134
		case 4:
			result = uint32((*server.Object)(pointer).ObjFlags)
		case 5:
			result = uint32((*server.ModifierEff)(pointer).Update100.Val)
		case 6:
			result = uint32((*server.HealthData)(pointer).Cur)
		case 7:
			result = uint32((*server.HealthData)(pointer).Max)
		default:
			m.t.Fatal("undeclared independent typed read")
		}
	})
	return
}

func checkRegenerationRawInitial4E01D0(m *regenerationRawMemory4E01D0, row regenerationRawInitial4E01D0) {
	m.t.Helper()
	for _, read := range row.Reads {
		pointer := m.fromPE(read.Role)
		got, fault := m.initialGo(pointer, read.Kind)
		if fault != (read.Want == nil) || (!fault && got != *read.Want) {
			m.t.Fatal("independent Go pointer/DWORD/WORD/fault", read.Kind, fault, got, read.Want)
		}
		if fault {
			m.faults++
			continue
		}
		c := regenerationRawInitialC4E01D0(pointer, read.Kind)
		value := uint32(c)
		if read.Kind == 0 || read.Kind == 2 {
			value = m.toPE(c)
		}
		if value != *read.Want {
			m.t.Fatal("independent C typed field", read.Kind, value, *read.Want)
		}
	}
	if !bytes.Equal(m.buf, m.before) || fmt.Sprintf("%x", sha256.Sum256(m.project())) != row.InitialSHA {
		m.t.Fatal("complete initial native/PE32 heaps changed during typed read controls")
	}
}

func TestRegenerationRawNativeLayout4E01D0(t *testing.T) {
	object, effect, health := server.Object{}, server.ModifierEff{}, server.HealthData{}
	// The two Go-only EXT handles follow the C object prefix.
	prefix := unsafe.Offsetof(object.ScriptPickup) + unsafe.Sizeof(object.ScriptPickup)
	if unsafe.Sizeof(object) != prefix+2*unsafe.Sizeof(uintptr(0)) {
		t.Fatal("native object prefix followed by two Go-only handles")
	}
	want := [11]uintptr{unsafe.Offsetof(object.InvHolder), unsafe.Offsetof(object.ObjClass), unsafe.Offsetof(object.HealthData), unsafe.Offsetof(object.Frame134), unsafe.Offsetof(object.ObjFlags), unsafe.Offsetof(effect.Update100) + unsafe.Offsetof(effect.Update100.Val), unsafe.Offsetof(health.Cur), unsafe.Offsetof(health.Max), prefix, unsafe.Sizeof(effect), unsafe.Sizeof(uintptr(0))}
	if got := regenerationRawLayout4E01D0(); got != want {
		t.Fatal("independent actual C/Go raw Regeneration layouts", got, want)
	}
}

func TestRegenerationRawInitialBefore4E01D0(t *testing.T) {
	if os.Getenv("OPENNOX_REGENERATION_RAW_BEFORE") == "1" {
		if regenerationRawABI4E01D0() != 1 {
			t.Fatal("initial memory must freeze before raw Regeneration C signature changes")
		}
		if _, err := os.Stat("regeneration_raw_native_4e01d0.go"); !os.IsNotExist(err) {
			t.Fatal("raw Regeneration engine must remain absent during initial freeze")
		}
	}
}

func TestRegenerationRawInitialMemory4E01D0(t *testing.T) {
	TestRegenerationRawNativeLayout4E01D0(t)
	total := 0
	for _, row := range regenerationRawInitials4E01D0(t) {
		t.Run(row.Name, func(t *testing.T) {
			memory := newRegenerationRawMemory4E01D0(t, row)
			checkRegenerationRawInitial4E01D0(memory, row)
			total += memory.faults
		})
	}
	if total != 36 {
		t.Fatal("all independent raw Regeneration read-fault domains", total)
	}
	t.Logf("REGENERATION_INITIAL_READ_FAULT_DOMAINS=%d", total)
}

func TestRegenerationRawActualCBefore4E01D0(t *testing.T) {
	if os.Getenv("OPENNOX_REGENERATION_RAW_BEFORE") != "1" || unsafe.Sizeof(uintptr(0)) != 8 {
		return
	}
	if os.Getenv("OPENNOX_REGENERATION_RAW_CHILD") == "1" {
		TestRegenerationRawNativeLayout4E01D0(t)
		if regenerationRawABI4E01D0() != 1 {
			t.Fatal("actual unchanged two-integer raw Regeneration required")
		}
		for _, row := range regenerationRawInitials4E01D0(t) {
			if row.Name != "baseline" {
				continue
			}
			memory := newRegenerationRawMemory4E01D0(t, row)
			checkRegenerationRawInitial4E01D0(memory, row)
			t.Logf("REGENERATION_REAL_C_INITIAL_OK item=%p effect=%p independent-layout-C-Go-fields-whole-heap=true", memory.fromPE(row.Roles["item"]), memory.fromPE(row.Roles["effect"]))
			regenerationRawPublic4E01D0(memory.fromPE(row.Roles["effect"]), memory.fromPE(row.Roles["item"]))
			t.Fatal("old two-integer raw Regeneration unexpectedly returned")
		}
		t.Fatal("independent baseline control missing")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestRegenerationRawActualCBefore4E01D0$", "-test.v")
	cmd.Env = append(os.Environ(), "OPENNOX_REGENERATION_RAW_CHILD=1")
	output, err := cmd.CombinedOutput()
	t.Logf("actual unchanged raw Regeneration: %v\n%s", err, output)
	if ctx.Err() != nil || err == nil || !bytes.Contains(output, []byte("REGENERATION_REAL_C_INITIAL_OK")) || bytes.Contains(output, []byte("unexpectedly returned")) {
		t.Fatal("all native typed-field/whole-memory positive controls must precede old-C fault", err)
	}
	if !bytes.Contains(output, []byte("SIGSEGV")) && !bytes.Contains(output, []byte("SIGBUS")) && !bytes.Contains(output, []byte("unexpected fault address")) {
		t.Fatal("actual raw Regeneration high-pointer C fault required")
	}
}
