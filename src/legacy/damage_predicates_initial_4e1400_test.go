package legacy

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

const damagePredicatePEBase4E1400 = uint32(0x21000000)

type damagePredicateInitial4E1400 struct {
	Name, Entry string
	Seed        int
	Roles       map[string]uint32
	Shapes      map[string]string
	Words       []struct {
		Address uint32
		Width   int
		Value   uint32
	} `json:"initializer_words"`
	Fields []struct {
		Kind          int
		Role, Address uint32
		PEWidth       int `json:"pe_width"`
		Pointer       bool
		Value         uint32
	} `json:"native_fields"`
	Reads []struct {
		Role uint32
		Kind int
		Want *uint32
	} `json:"independent_reads"`
	InitialSHA string `json:"initial_heap_sha256"`
}

func damagePredicateInitials4E1400(t *testing.T) []damagePredicateInitial4E1400 {
	t.Helper()
	var rows []damagePredicateInitial4E1400
	modifierPredicateFixture4E03D0(t, "damage_predicates_initial_memory_4e1400.json.gz", "63fc1df78dc74ef62880f559c1feda88c3b2571e6bd2dddc3c8bcec386b4a44d", "b85c65006756fc9ce1714017644eb951f5bd779ab0be010153586dcab68e67f0", &rows)
	if len(rows) != 487 {
		t.Fatal("all487 independent raw damage-predicate initial cases")
	}
	return rows
}

func damagePredicateGoLayout4E1400() [15]uintptr {
	o, w, u := server.Object{}, server.WandUseData{}, server.UseDataPtr{}
	return [15]uintptr{
		unsafe.Offsetof(o.ObjClass), unsafe.Offsetof(o.ObjSubClass), unsafe.Offsetof(o.ObjClass), unsafe.Offsetof(o.ObjSubClass),
		unsafe.Offsetof(o.UseData) + unsafe.Offsetof(u.Ptr), unsafe.Offsetof(w.Flags), unsafe.Sizeof(uintptr(0)),
		unsafe.Offsetof(o.ScriptPickup) + unsafe.Sizeof(o.ScriptPickup), unsafe.Sizeof(w), unsafe.Alignof(o), unsafe.Alignof(w),
		unsafe.Sizeof(o.ObjClass), unsafe.Sizeof(o.ObjSubClass), unsafe.Sizeof(u.Ptr), unsafe.Sizeof(w.Flags),
	}
}

func TestDamagePredicatesNativeLayout4E1400(t *testing.T) {
	want := damagePredicateGoLayout4E1400()
	if got := damagePredicateLayout4E1400(); got != want {
		t.Fatal("actual C/Go object, UseDataPtr and pointer-free Wand payload layouts", got, want)
	}
	if unsafe.Sizeof(server.Object{}) != want[7]+2*want[6] {
		t.Fatal("complete C object prefix plus exactly two Go native handle words")
	}
}

type damagePredicateSlot4E1400 struct {
	pe, native, peWidth, width int
	pointer                    bool
}

type damagePredicateMemory4E1400 struct {
	t               *testing.T
	heap            unsafe.Pointer
	layout          [15]uintptr
	buf, pe, before []byte
	addresses       map[uint32]unsafe.Pointer
	reverse         map[uintptr]uint32
	slots           []damagePredicateSlot4E1400
}

func (m *damagePredicateMemory4E1400) add(pe uint32, pointer unsafe.Pointer) {
	if old, ok := m.addresses[pe]; ok && old != pointer {
		m.t.Fatal("contradictory independent native field correspondence", pe)
	}
	if old, ok := m.reverse[uintptr(pointer)]; ok && old != pe {
		m.t.Fatal("undeclared physical alias in native initial projection", pe, old)
	}
	m.addresses[pe], m.reverse[uintptr(pointer)] = pointer, pe
}

func (m *damagePredicateMemory4E1400) fromPE(pe uint32) unsafe.Pointer {
	pointer, ok := m.addresses[pe]
	if !ok {
		m.t.Fatalf("undeclared independent pointer %#x", pe)
	}
	return pointer
}

func (m *damagePredicateMemory4E1400) toPE(pointer uintptr) uint32 {
	pe, ok := m.reverse[pointer]
	if !ok {
		m.t.Fatalf("unprojected native pointer bits %#x", pointer)
	}
	return pe
}

func newDamagePredicateMemory4E1400(t *testing.T, row damagePredicateInitial4E1400) *damagePredicateMemory4E1400 {
	t.Helper()
	TestDamagePredicatesNativeLayout4E1400(t)
	// Reuse the existing immutable C allocator, which owns64KiB between a
	// protected predecessor and successor. No Go heap pointer enters this map.
	heap, guard, free := regenerationRawHeap4E01D0()
	t.Cleanup(free)
	layout := damagePredicateLayout4E1400()
	if heap == nil || guard < layout[7] || (layout[6] == 8 && uintptr(heap) <= math.MaxUint32) {
		t.Fatal("guarded native high C heap required")
	}
	m := &damagePredicateMemory4E1400{t: t, heap: heap, layout: layout, buf: unsafe.Slice((*byte)(heap), 65536), pe: make([]byte, 65536), addresses: make(map[uint32]unsafe.Pointer), reverse: make(map[uintptr]uint32)}
	for i := range m.pe {
		m.pe[i] = byte((i*37 + row.Seed*19) ^ (i >> 3))
	}
	for _, word := range row.Words {
		offset := int(word.Address - damagePredicatePEBase4E1400)
		for j := 0; j < word.Width; j++ {
			m.pe[offset+j] = byte(word.Value >> (8 * j))
		}
	}
	if fmt.Sprintf("%x", sha256.Sum256(m.pe)) != row.InitialSHA {
		t.Fatal("complete independently initialized PE heap", row.Name)
	}
	copy(m.buf, m.pe)
	m.add(0, nil)
	m.add(0x41000000, regenerationRawBad4E01D0(heap, guard))
	// Weapon precedes source so an explicit source=weapon input reuses the
	// same actual native object, including its already chosen boundary shape.
	for _, role := range []string{"weapon", "source", "use"} {
		pe := row.Roles[role]
		if _, ok := m.addresses[pe]; ok {
			continue
		}
		offset := uintptr(pe - damagePredicatePEBase4E1400)
		switch row.Shapes[role] {
		case "last-class":
			offset = 65536 - layout[0] - 4
		case "last-subclass-byte":
			offset = 65536 - layout[1] - 1
		case "last-subclass":
			offset = 65536 - layout[3] - 4
		case "last-use-pointer":
			offset = 65536 - layout[4] - layout[6]
		case "last-flags-byte":
			offset = 65536 - layout[5] - 1
		}
		m.add(pe, unsafe.Add(heap, offset))
	}
	for _, field := range row.Fields {
		pointer := unsafe.Add(m.fromPE(field.Role), layout[field.Kind])
		width := field.PEWidth
		if field.Pointer {
			width = int(layout[6])
		}
		slot := damagePredicateSlot4E1400{int(field.Address - damagePredicatePEBase4E1400), int(uintptr(pointer) - uintptr(heap)), field.PEWidth, width, field.Pointer}
		if slot.native < 0 || slot.native+width > len(m.buf) {
			t.Fatal("independently readable initial storage needs a readable native field domain")
		}
		for _, old := range m.slots {
			if slot.native < old.native+old.width && old.native < slot.native+slot.width {
				t.Fatal("undeclared overlapping native field storage")
			}
		}
		m.slots = append(m.slots, slot)
		m.add(field.Address, pointer)
	}
	for i, field := range row.Fields {
		slot := m.slots[i]
		value := uint64(field.Value)
		if field.Pointer {
			value = uint64(uintptr(m.fromPE(field.Value)))
		}
		for j := 0; j < slot.width; j++ {
			m.buf[slot.native+j] = byte(value >> (8 * j))
		}
	}
	m.before = bytes.Clone(m.buf)
	if fmt.Sprintf("%x", sha256.Sum256(m.project())) != row.InitialSHA {
		t.Fatal("whole native-to-PE32 initial correspondence")
	}
	return m
}

func (m *damagePredicateMemory4E1400) project() []byte {
	result, allowed := bytes.Clone(m.pe), make([]bool, len(m.buf))
	for _, slot := range m.slots {
		var value uint64
		for j := 0; j < slot.width; j++ {
			value |= uint64(m.buf[slot.native+j]) << (8 * j)
			allowed[slot.native+j] = true
		}
		if slot.pointer {
			value = uint64(m.toPE(uintptr(value)))
		}
		for j := 0; j < slot.peWidth; j++ {
			result[slot.pe+j] = byte(value >> (8 * j))
		}
	}
	for i, value := range m.buf {
		if !allowed[i] && value != m.before[i] {
			m.t.Fatal("unrelated complete native heap guard byte changed", i)
		}
	}
	return result
}

// These independent read controls use actual named Go field offsets while
// retaining partial/unaligned field domains. The pointer slot is read as
// uintptr-sized scalar storage and immediately projected; all allocations and
// pointees here are C-owned. This is not a Go-heap pointer/GC implementation,
// a forged misaligned *server.Object cast, or a future raw gameplay getter.
func damagePredicateInitialGo4E1400(base unsafe.Pointer, kind int) uintptr {
	layout := damagePredicateGoLayout4E1400()
	// Keep an actual nil scalar read. unsafe.Add(nil, offset) can terminate
	// under checkptr before the recoverable memory-fault domain is observed.
	if base == nil {
		if kind == 1 || kind == 5 {
			return uintptr(*(*uint8)(base))
		}
		if kind == 4 && layout[6] == 8 {
			return uintptr(*(*uint64)(base))
		}
		return uintptr(*(*uint32)(base))
	}
	address := unsafe.Add(base, layout[kind])
	if kind == 1 || kind == 5 {
		return uintptr(*(*uint8)(address))
	}
	if kind == 4 {
		if layout[6] == 8 {
			return uintptr(*(*uint64)(address))
		}
		return uintptr(*(*uint32)(address))
	}
	return uintptr(*(*uint32)(address))
}

func checkDamagePredicateInitial4E1400(m *damagePredicateMemory4E1400, row damagePredicateInitial4E1400) int {
	faults := 0
	for _, read := range row.Reads {
		var got uintptr
		failure := modifierPredicateRecover4E03D0(m.t, func() { got = damagePredicateInitialGo4E1400(m.fromPE(read.Role), read.Kind) })
		if failure != (read.Want == nil) {
			m.t.Fatal("independent partial/unaligned native field read domain", row.Name, read.Kind, failure)
		}
		if failure {
			faults++
			continue
		}
		value := uint32(got)
		if read.Kind == 4 {
			value = m.toPE(got)
		}
		if value != *read.Want {
			m.t.Fatal("independent Go scalar/byte/native-pointer-storage value", row.Name, read.Kind, value, *read.Want)
		}
		c := damagePredicateInitialC4E1400(m.fromPE(read.Role), read.Kind)
		cValue := uint32(c)
		if read.Kind == 4 {
			cValue = m.toPE(c)
		}
		if cValue != *read.Want {
			m.t.Fatal("independently readable unaligned-safe C field", row.Name, read.Kind, cValue, *read.Want)
		}
	}
	if !bytes.Equal(m.buf, m.before) || fmt.Sprintf("%x", sha256.Sum256(m.project())) != row.InitialSHA {
		m.t.Fatal("complete native/PE initial heaps changed during independent controls")
	}
	return faults
}

func TestDamagePredicatesInitialBefore4E1400(t *testing.T) {
	if os.Getenv("OPENNOX_DAMAGE_PREDICATES_BEFORE") != "1" {
		return
	}
	for entry := 0; entry < 2; entry++ {
		if damagePredicateABI4E1400(entry) != 1 {
			t.Fatal("both original public C ABIs required during initial freeze")
		}
	}
	for _, name := range []string{"damage_predicates_raw_native_4e1400.go", "damage_predicates_native_4e1400.go"} {
		if _, err := os.Stat(name); !os.IsNotExist(err) {
			t.Fatal("raw engines must remain absent during initial freeze", name)
		}
	}
}

func TestDamagePredicatesInitialMemory4E1400(t *testing.T) {
	faults, attack, friendly := 0, 0, 0
	for _, row := range damagePredicateInitials4E1400(t) {
		t.Run(row.Name, func(t *testing.T) {
			m := newDamagePredicateMemory4E1400(t, row)
			faults += checkDamagePredicateInitial4E1400(m, row)
			if row.Entry == "attack" {
				attack++
			} else if row.Entry == "friendly" {
				friendly++
			} else {
				t.Fatal("independent entry identity")
			}
		})
	}
	if faults != 1222 || attack != 317 || friendly != 170 {
		t.Fatal("all independent typed read domains and both input families", faults, attack, friendly)
	}
	t.Logf("DAMAGE_PREDICATES_INITIAL_FIELD_FAULT_DOMAINS=%d inputs=%d/%d full64KiB-native-and-PE-guards=true", faults, attack, friendly)
}

func TestDamagePredicatesActualCBefore4E1400(t *testing.T) {
	if os.Getenv("OPENNOX_DAMAGE_PREDICATES_BEFORE") != "1" || unsafe.Sizeof(uintptr(0)) != 8 {
		return
	}
	if child := os.Getenv("OPENNOX_DAMAGE_PREDICATES_CHILD"); child != "" {
		TestDamagePredicatesNativeLayout4E1400(t)
		TestDamagePredicatesInitialBefore4E1400(t)
		names := map[string]string{"attack-source": "attack/source-pointer/valid/4/0", "attack-use": "attack/class/ffffffff", "friendly": "friendly/weapon-pointer/valid/01000000/00004000"}
		name, ok := names[child]
		if !ok {
			t.Fatal("unknown old C positive-control child")
		}
		for _, row := range damagePredicateInitials4E1400(t) {
			if row.Name == name {
				m := newDamagePredicateMemory4E1400(t, row)
				checkDamagePredicateInitial4E1400(m, row)
				entry := 0
				if row.Entry == "friendly" {
					entry = 1
				}
				if damagePredicateABI4E1400(entry) != 1 {
					t.Fatal("actual unchanged raw C signature required")
				}
				t.Logf("DAMAGE_PREDICATES_REAL_C_INITIAL_OK condition=%s source=%p weapon=%p use=%p independent-layout-byte-DWORD-native-pointer-storage-fields-whole-heaps=true", child, m.fromPE(row.Roles["source"]), m.fromPE(row.Roles["weapon"]), m.fromPE(row.Roles["use"]))
				result := damagePredicatePublic4E1400(entry, m.fromPE(row.Roles["source"]), m.fromPE(row.Roles["weapon"]))
				t.Fatal("old raw C unexpectedly returned", result)
			}
		}
		t.Fatal("independently chosen child control missing", name)
	}
	for _, condition := range []string{"attack-source", "attack-use", "friendly"} {
		t.Run(condition, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestDamagePredicatesActualCBefore4E1400$", "-test.v")
			cmd.Env = append(os.Environ(), "OPENNOX_DAMAGE_PREDICATES_CHILD="+condition)
			output, err := cmd.CombinedOutput()
			t.Logf("actual unchanged raw C condition=%s error=%v\n%s", condition, err, output)
			if ctx.Err() != nil || err == nil || !bytes.Contains(output, []byte("DAMAGE_PREDICATES_REAL_C_INITIAL_OK")) || bytes.Contains(output, []byte("unexpectedly returned")) {
				t.Fatal("independent native positive controls must precede actual unchanged-C fault", err)
			}
			if !strings.Contains(string(output), "SIGSEGV") && !strings.Contains(string(output), "SIGBUS") && !strings.Contains(string(output), "unexpected fault address") {
				t.Fatal("actual old C memory fault required")
			}
		})
	}
}
