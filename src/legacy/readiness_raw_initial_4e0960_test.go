package legacy

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"runtime"
	"runtime/debug"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

const readinessRawPEBase4E0960 = uint32(0x21000000)
const readinessRawCollisionPE4E0960 = uint32(0x004e0951)

type readinessRawInitialRow4E0960 struct {
	Name   string
	Seed   int
	Roles  map[string]uint32
	Shapes map[string]string
	Words  []struct {
		Address, Value uint32
	} `json:"initializer_words"`
	Fields []struct {
		Role, Address uint32
		Kind, Index   int
		Type          string
		Value         uint32
	} `json:"native_fields"`
	Reads []struct {
		Role        uint32
		Kind, Index int
		Want        *uint32
	} `json:"independent_reads"`
	CallbackValues []uint32 `json:"callback_values"`
	InitialSHA     string   `json:"initial_heap_sha256"`
}

func readinessRawInitials4E0960(t *testing.T) []readinessRawInitialRow4E0960 {
	t.Helper()
	encoded, err := os.ReadFile("testdata/readiness_raw_initial_memory_4e0960.json.gz")
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(encoded)) != "fe9c2684a03b84b5feb665eae273931e7ee70dfa9a6ff196af267b0cc3120884" {
		t.Fatal("independent Readiness compressed initial fixture", err)
	}
	z, err := gzip.NewReader(bytes.NewReader(encoded))
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err = z.Close(); err != nil || fmt.Sprintf("%x", sha256.Sum256(decoded)) != "c8df462bf57148a3023a45f86c0b53bd3eb4dd657ab95e319944f03474ed8fa3" {
		t.Fatal("independent Readiness decoded initial fixture", err)
	}
	var rows []readinessRawInitialRow4E0960
	if err := json.Unmarshal(decoded, &rows); err != nil || len(rows) != 113 {
		t.Fatal("all113 independently chosen Readiness inputs", err)
	}
	return rows
}

type readinessRawInitialSlot4E0960 struct {
	pe, native, width int
	pointer           bool
}

type readinessRawMemory4E0960 struct {
	t                  *testing.T
	heap               unsafe.Pointer
	layout             [15]uintptr
	buf, pe            []byte
	before             []byte
	addresses          map[uint32]unsafe.Pointer
	reverse            map[uintptr]uint32
	slots              []readinessRawInitialSlot4E0960
	faults, reads      int
	collisionWitnesses int
}

func (m *readinessRawMemory4E0960) add(pe uint32, native unsafe.Pointer) {
	m.t.Helper()
	if old, ok := m.addresses[pe]; ok && old != native {
		m.t.Fatal("contradictory Readiness native pointer correspondence", pe)
	}
	if old, ok := m.reverse[uintptr(native)]; ok && old != pe {
		m.t.Fatal("aliased Readiness native pointer correspondence", old, pe)
	}
	m.addresses[pe], m.reverse[uintptr(native)] = native, pe
}

func (m *readinessRawMemory4E0960) fromPE(pe uint32) unsafe.Pointer {
	m.t.Helper()
	pointer, ok := m.addresses[pe]
	if !ok {
		m.t.Fatalf("undeclared Readiness PE pointer %#x", pe)
	}
	return pointer
}

func (m *readinessRawMemory4E0960) toPE(native uintptr) uint32 {
	m.t.Helper()
	pe, ok := m.reverse[native]
	if !ok {
		m.t.Fatalf("unprojected Readiness native pointer %#x", native)
	}
	return pe
}

func (m *readinessRawMemory4E0960) fieldOffset(kind, index int) uintptr {
	m.t.Helper()
	switch kind {
	case 0:
		return m.layout[0]
	case 1:
		return m.layout[1]
	case 2:
		if index < 0 || index >= 4 {
			m.t.Fatal("undeclared Readiness modifier slot")
		}
		return m.layout[2+index]
	case 3:
		return m.layout[6]
	case 4:
		return m.layout[7]
	default:
		m.t.Fatal("undeclared Readiness field kind")
		return 0
	}
}

func newReadinessRawMemory4E0960(t *testing.T, row readinessRawInitialRow4E0960) *readinessRawMemory4E0960 {
	t.Helper()
	heap, guard, free := readinessRawHeap4E0960()
	t.Cleanup(free)
	layout := readinessRawLayout4E0960()
	if heap == nil || guard < layout[8]+1024 || (layout[11] == 8 && uintptr(heap) <= math.MaxUint32) {
		t.Fatal("native C-owned whole high heap and protected predecessor/successor required")
	}
	m := &readinessRawMemory4E0960{t: t, heap: heap, layout: layout, buf: unsafe.Slice((*byte)(heap), 65536), pe: make([]byte, 65536), addresses: make(map[uint32]unsafe.Pointer), reverse: make(map[uintptr]uint32)}
	for i := range m.pe {
		m.pe[i] = byte((i*37 + row.Seed*19) ^ (i >> 3))
	}
	for _, word := range row.Words {
		offset := int(word.Address - readinessRawPEBase4E0960)
		if offset < 0 || offset+4 > len(m.pe) {
			t.Fatal("initializer outside declared independent PE heap")
		}
		binary.LittleEndian.PutUint32(m.pe[offset:], word.Value)
	}
	if fmt.Sprintf("%x", sha256.Sum256(m.pe)) != row.InitialSHA {
		t.Fatal("complete independent Readiness PE initial heap", row.Name)
	}
	copy(m.buf, m.pe)
	m.add(0, nil)
	m.add(0x41000000, readinessRawBad4E0960(heap, guard))
	callback := readinessRawCallback4E0960()
	if callback == nil || (layout[11] == 8 && uintptr(callback) <= math.MaxUint32) {
		t.Fatal("actual native C.nullsub_22 callback identity required")
	}
	m.add(0x004e0950, callback)
	// Nonmatching original callback values are never invoked. One selected
	// token is a C-created low32 collision; others are opaque C-owned addresses.
	// Both mappings preserve their distinct original PE identities.
	for i, pe := range row.CallbackValues {
		if _, ok := m.addresses[pe]; !ok {
			if pe == readinessRawCollisionPE4E0960 && layout[11] == 8 {
				m.add(pe, readinessRawCollisionCallback4E0960())
			} else {
				m.add(pe, unsafe.Add(heap, 0x1000+i*64))
			}
		}
	}
	for _, role := range []string{"item", "table", "first", "second", "replacement", "alternate_table"} {
		pe := row.Roles[role]
		if _, ok := m.addresses[pe]; ok {
			continue
		}
		offset := uintptr(pe - readinessRawPEBase4E0960)
		switch row.Shapes[role] {
		case "last-init":
			offset = 65536 - layout[0] - layout[11]
		case "last-class":
			// The PE item is only four-byte aligned. Keep its native object
			// aligned while preserving class-readable/InitData-protected
			// domains; the extra native tail bytes remain guarded, unaliased.
			offset = (65536 - layout[1] - 4) &^ (layout[12] - 1)
		case "last-slot2":
			offset = 65536 - layout[4] - layout[11]
		case "last-slot3":
			offset = 65536 - layout[5] - layout[11]
		case "last-callback":
			offset = 65536 - layout[6] - layout[11]
		case "last-strength":
			offset = 65536 - layout[7] - 4
		}
		m.add(pe, unsafe.Add(heap, offset))
	}
	for _, field := range row.Fields {
		pointer := unsafe.Add(m.fromPE(field.Role), m.fieldOffset(field.Kind, field.Index))
		width := 4
		if field.Type == "pointer" {
			width = int(layout[11])
		} else if field.Type != "word" {
			t.Fatal("undeclared Readiness native field type")
		}
		slot := readinessRawInitialSlot4E0960{pe: int(field.Address - readinessRawPEBase4E0960), native: int(uintptr(pointer) - uintptr(heap)), width: width, pointer: field.Type == "pointer"}
		if slot.native < 0 || slot.native+width > len(m.buf) || slot.pe < 0 || slot.pe+4 > len(m.pe) {
			t.Fatal("independent readable fields need complete native correspondences")
		}
		for _, prior := range m.slots {
			if slot.native < prior.native+prior.width && prior.native < slot.native+slot.width {
				t.Fatal("overlapping independent Readiness native fields")
			}
		}
		m.add(field.Address, pointer)
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
		t.Fatal("complete initial native-to-PE32 Readiness projection")
	}
	return m
}

func (m *readinessRawMemory4E0960) project() []byte {
	m.t.Helper()
	pe, allowed := bytes.Clone(m.pe), make([]bool, len(m.buf))
	for _, slot := range m.slots {
		var value uint64
		for j := 0; j < slot.width; j++ {
			value |= uint64(m.buf[slot.native+j]) << (8 * j)
			allowed[slot.native+j] = true
		}
		if slot.pointer {
			value = uint64(m.toPE(uintptr(value)))
		}
		binary.LittleEndian.PutUint32(pe[slot.pe:], uint32(value))
	}
	for i, value := range m.buf {
		if !allowed[i] && value != m.before[i] {
			m.t.Fatal("unrelated native Readiness guard byte changed", i)
		}
	}
	return pe
}

func readinessRawRecover4E0960(t *testing.T, fn func()) (fault bool) {
	t.Helper()
	old := debug.SetPanicOnFault(true)
	defer debug.SetPanicOnFault(old)
	defer func() {
		if value := recover(); value != nil {
			if _, ok := value.(runtime.Error); !ok {
				t.Fatalf("unexpected non-runtime Readiness initial read failure: %v", value)
			}
			fault = true
		}
	}()
	fn()
	return
}

func (m *readinessRawMemory4E0960) initialGo(base unsafe.Pointer, kind, index int) (value uint32, fault bool) {
	fault = readinessRawRecover4E0960(m.t, func() {
		switch kind {
		case 0:
			value = m.toPE(uintptr((*server.Object)(base).InitData))
		case 1:
			value = uint32((*server.Object)(base).ObjClass)
		case 2:
			value = m.toPE(uintptr(unsafe.Pointer((*server.ModifierInitData)(base).Modifiers[index])))
		case 3:
			value = m.toPE(uintptr((*server.ModifierEff)(base).Attack40.Fnc))
		case 4:
			value = uint32((*server.ModifierEff)(base).Attack40.Val)
		default:
			m.t.Fatal("undeclared Readiness initial Go field")
		}
	})
	return
}

func checkReadinessRawInitial4E0960(m *readinessRawMemory4E0960, row readinessRawInitialRow4E0960) {
	m.t.Helper()
	for _, read := range row.Reads {
		base := m.fromPE(read.Role)
		value, fault := m.initialGo(base, read.Kind, read.Index)
		m.reads++
		if fault != (read.Want == nil) || (!fault && value != *read.Want) {
			m.t.Fatal("independent Readiness Go pointer/DWORD/fault", read.Kind, read.Index, fault, value, read.Want)
		}
		if fault {
			m.faults++
			continue
		}
		actual := readinessRawInitialC4E0960(base, read.Kind, read.Index)
		c := uint32(actual)
		if read.Kind == 0 || read.Kind == 2 || read.Kind == 3 {
			c = m.toPE(actual)
		}
		if c != *read.Want {
			m.t.Fatal("independent Readiness actual C typed field", read.Kind, read.Index, c, *read.Want)
		}
		if m.layout[11] == 8 && read.Kind == 3 && *read.Want == readinessRawCollisionPE4E0960 {
			real := uintptr(readinessRawCallback4E0960())
			token := uintptr(readinessRawCollisionCallback4E0960())
			goPointer := (*server.ModifierEff)(base).Attack40.Fnc
			if real <= math.MaxUint32 || token <= math.MaxUint32 || token == real || uint32(token) != uint32(real) ||
				uintptr(goPointer) != token || actual != token || m.toPE(uintptr(goPointer)) != readinessRawCollisionPE4E0960 ||
				!readinessRawTruncationNegativeControl4E0960() {
				m.t.Fatal("native nonmatch token must collide only under the deliberately wrong both-sides uint32 comparison", real, token, goPointer, actual)
			}
			m.collisionWitnesses++
			m.t.Logf("READINESS_NATIVE_CALLBACK_LOW32_COLLISION_OK case=%s real=%#x nonmatch=%#x low32=%#x token=%#x full-pointer-different=true bad-uint32-comparison=true never-invoked=true", row.Name, real, token, uint32(real), readinessRawCollisionPE4E0960)
		}
	}
	if !bytes.Equal(m.buf, m.before) || fmt.Sprintf("%x", sha256.Sum256(m.project())) != row.InitialSHA {
		m.t.Fatal("complete native/PE heaps changed during Readiness initial controls")
	}
}

func TestReadinessRawNativeLayout4E0960(t *testing.T) {
	obj, table, effect, call := server.Object{}, server.ModifierInitData{}, server.ModifierEff{}, server.ModifierEffFnc{}
	prefix := unsafe.Offsetof(obj.ScriptPickup) + unsafe.Sizeof(obj.ScriptPickup)
	if unsafe.Sizeof(obj) != prefix+2*unsafe.Sizeof(uintptr(0)) {
		t.Fatal("complete C object prefix plus exactly two Go native handle words")
	}
	want := [15]uintptr{unsafe.Offsetof(obj.InitData), unsafe.Offsetof(obj.ObjClass),
		unsafe.Offsetof(table.Modifiers), unsafe.Offsetof(table.Modifiers) + unsafe.Sizeof(table.Modifiers[0]),
		unsafe.Offsetof(table.Modifiers) + 2*unsafe.Sizeof(table.Modifiers[0]), unsafe.Offsetof(table.Modifiers) + 3*unsafe.Sizeof(table.Modifiers[0]),
		unsafe.Offsetof(effect.Attack40) + unsafe.Offsetof(call.Fnc), unsafe.Offsetof(effect.Attack40) + unsafe.Offsetof(call.Val),
		prefix, unsafe.Sizeof(table), unsafe.Sizeof(effect), unsafe.Sizeof(uintptr(0)),
		unsafe.Alignof(obj), unsafe.Alignof(table), unsafe.Alignof(effect)}
	if got := readinessRawLayout4E0960(); got != want {
		t.Fatal("actual independent Readiness C/Go offsets sizes alignments", got, want)
	}
}

func TestReadinessRawInitialBefore4E0960(t *testing.T) {
	if os.Getenv("OPENNOX_READINESS_RAW_BEFORE") == "1" {
		if readinessRawABI4E0960() != 1 {
			t.Fatal("raw Readiness C signature must remain int before initial freeze")
		}
		if _, err := os.Stat("readiness_raw_native_4e0960.go"); !os.IsNotExist(err) {
			t.Fatal("raw Readiness native engine must remain absent during initial freeze")
		}
	}
}

func TestReadinessRawInitialMemory4E0960(t *testing.T) {
	TestReadinessRawNativeLayout4E0960(t)
	reads, faults := 0, 0
	witnesses, witnessCases := 0, 0
	for _, row := range readinessRawInitials4E0960(t) {
		t.Run(row.Name, func(t *testing.T) {
			memory := newReadinessRawMemory4E0960(t, row)
			checkReadinessRawInitial4E0960(memory, row)
			reads += memory.reads
			faults += memory.faults
			witnesses += memory.collisionWitnesses
			if memory.collisionWitnesses > 0 {
				witnessCases++
			}
		})
	}
	if reads != 1808 || faults != 52 {
		t.Fatal("all independent Readiness field and fault domains", reads, faults)
	}
	t.Logf("READINESS_INITIAL_READ_DOMAINS=%d FAULT_DOMAINS=%d", reads, faults)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if witnesses != 3 || witnessCases != 3 {
			t.Fatal("all three independently chosen nonmatch callback fields must carry a native low32-collision witness", witnesses, witnessCases)
		}
		t.Logf("READINESS_NATIVE_CALLBACK_LOW32_COLLISION_WITNESSES=%d selected-cases=%d", witnesses, witnessCases)
	}
}

func TestReadinessRawActualCBefore4E0960(t *testing.T) {
	if os.Getenv("OPENNOX_READINESS_RAW_BEFORE") != "1" || unsafe.Sizeof(uintptr(0)) != 8 {
		return
	}
	if os.Getenv("OPENNOX_READINESS_RAW_CHILD") == "1" {
		if readinessRawABI4E0960() != 1 {
			t.Fatal("actual unchanged int Readiness C entry required")
		}
		TestReadinessRawNativeLayout4E0960(t)
		for _, row := range readinessRawInitials4E0960(t) {
			if row.Name != "baseline" {
				continue
			}
			memory := newReadinessRawMemory4E0960(t, row)
			checkReadinessRawInitial4E0960(memory, row)
			if memory.toPE(uintptr(readinessRawCallback4E0960())) != 0x004e0950 || memory.faults != 0 {
				t.Fatal("real full-width C callback identity and positive field controls")
			}
			item := memory.fromPE(row.Roles["item"])
			t.Logf("READINESS_REAL_C_INITIAL_OK item=%p table=%p first=%p callback=%p independent-layout-C-Go-fields-whole-heap=true", item, memory.fromPE(row.Roles["table"]), memory.fromPE(row.Roles["first"]), readinessRawCallback4E0960())
			readinessRawPublic4E0960(item)
			t.Fatal("old int Readiness C unexpectedly returned")
		}
		t.Fatal("independent Readiness baseline missing")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestReadinessRawActualCBefore4E0960$", "-test.v")
	cmd.Env = append(os.Environ(), "OPENNOX_READINESS_RAW_CHILD=1")
	output, err := cmd.CombinedOutput()
	t.Logf("actual unchanged raw Readiness: %v\n%s", err, output)
	if ctx.Err() != nil || err == nil || !bytes.Contains(output, []byte("READINESS_REAL_C_INITIAL_OK")) || bytes.Contains(output, []byte("unexpectedly returned")) {
		t.Fatal("all native field/layout/whole-memory positive controls must precede old-C fault", err)
	}
	if !bytes.Contains(output, []byte("SIGSEGV")) && !bytes.Contains(output, []byte("SIGBUS")) && !bytes.Contains(output, []byte("unexpected fault address")) {
		t.Fatal("actual raw Readiness high-pointer C fault required")
	}
}
