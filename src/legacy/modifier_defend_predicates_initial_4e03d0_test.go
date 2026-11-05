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
	"reflect"
	"runtime/debug"
	"strconv"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

const modifierPredicatePEBase4E03D0 = uint32(0x21000000)

type modifierPredicateInitial4E03D0 struct {
	Name, Entry string
	Seed        int
	Argument    uint32
	Output      uint32
	OpaqueArgs  [4]uint32 `json:"opaque_args"`
	Words       []struct {
		Address, Value uint32
	} `json:"initializer_words"`
	Fields []struct {
		Role      uint32
		PEAddress uint32 `json:"pe_address"`
		Field     string
		Kind      string
		Value     uint32
	} `json:"native_fields"`
	Reads []struct {
		Role  uint32
		Kind  string
		Index int
		Want  *uint32
	} `json:"independent_reads"`
	InitialSHA string `json:"initial_heap_sha256"`
}

func modifierPredicateFixture4E03D0(t *testing.T, name, encoded, decoded string, target any) {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(data)) != encoded {
		t.Fatal("frozen compressed predicate fixture", name, err)
	}
	z, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err = z.Close(); err != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != decoded {
		t.Fatal("frozen decoded predicate fixture", name, err)
	}
	if err = json.Unmarshal(raw, target); err != nil {
		t.Fatal(err)
	}
}

func modifierPredicateInitials4E03D0(t *testing.T) []modifierPredicateInitial4E03D0 {
	t.Helper()
	var rows []modifierPredicateInitial4E03D0
	modifierPredicateFixture4E03D0(t, "modifier_defend_predicates_initial_memory_4e03d0.json.gz", "8a1ca3db5a6e2c4ded9dd7d33165046f957ccc8355ca8ce2b941dc54b6c2ccc1", "5eea6cc5cdefce04991b2ed96214b4fca33f854e80e08180cd7899c49093a1bc", &rows)
	if len(rows) != 298 {
		t.Fatal("all independent predicate inputs required")
	}
	return rows
}

type modifierPredicateSlot4E03D0 struct {
	pe, native, width int
	pointer           bool
}

type modifierPredicateMemory4E03D0 struct {
	t               *testing.T
	heap            unsafe.Pointer
	buf, before, pe []byte
	off             [13]uintptr
	slots           []modifierPredicateSlot4E03D0
	addresses       map[uint32]unsafe.Pointer
	reverse         map[uintptr]uint32
	argument, out   unsafe.Pointer
	faults          int
}

func newModifierPredicateMemory4E03D0(t *testing.T, row modifierPredicateInitial4E03D0) *modifierPredicateMemory4E03D0 {
	t.Helper()
	heap, guard, free := generatorCreditOriginalHeap4D61B0()
	t.Cleanup(free)
	off := modifierPredicatesLayout4E03D0()
	if heap == nil || guard < off[4]+1024 || (off[12] == 8 && uintptr(heap) <= math.MaxUint32) {
		t.Fatal("C-owned whole high heap and protected invalid domain required")
	}
	m := &modifierPredicateMemory4E03D0{t: t, heap: heap, off: off, buf: unsafe.Slice((*byte)(heap), 0x10000), pe: make([]byte, 0x10000), addresses: make(map[uint32]unsafe.Pointer), reverse: make(map[uintptr]uint32)}
	for i := range m.pe {
		m.pe[i] = byte((i*37 + row.Seed*19) ^ (i >> 3))
	}
	for _, word := range row.Words {
		binary.LittleEndian.PutUint32(m.pe[word.Address-modifierPredicatePEBase4E03D0:], word.Value)
	}
	if fmt.Sprintf("%x", sha256.Sum256(m.pe)) != row.InitialSHA {
		t.Fatal("independent complete PE32 initial heap", row.Name)
	}
	copy(m.buf, m.pe)
	add := func(pe uint32, native unsafe.Pointer) {
		if old, ok := m.addresses[pe]; ok && old != native {
			t.Fatal("contradictory declared pointer correspondence")
		}
		m.addresses[pe], m.reverse[uintptr(native)] = native, pe
	}
	add(0, nil)
	add(0x41000000, generatorCreditOriginalBad4D61B0(heap, guard, 0))
	add(0x4e03d0, modifierPredicatesCallback4E03D0(0))
	add(0x4e0480, modifierPredicatesCallback4E03D0(1))
	for _, role := range []int{0x1000, 0x2000, 0x3000, 0x4000, 0x5000} {
		add(modifierPredicatePEBase4E03D0+uint32(role), unsafe.Add(heap, role))
	}
	// Only the observed strength word must be readable in this boundary case.
	// Neither original nor native claims a full ModifierEff past the heap end.
	add(modifierPredicatePEBase4E03D0+0x10000-100, unsafe.Add(heap, 0x10000-int(off[2])-4))
	fieldIDs := map[string]int{"effect_index": 3, "effect_callback": 1, "effect_strength": 2, "object_class": 5, "object_init": 6, "slot0": 8, "slot1": 9, "slot2": 10, "slot3": 11}
	for _, field := range row.Fields {
		id, ok := fieldIDs[field.Field]
		if !ok {
			t.Fatal("undeclared native field", field.Field)
		}
		native := unsafe.Add(m.fromPE(field.Role), m.off[id])
		add(field.PEAddress, native)
		width := 4
		if field.Kind == "pointer" {
			width = int(off[12])
		} else if field.Kind != "word" {
			t.Fatal("undeclared native field width")
		}
		slot := modifierPredicateSlot4E03D0{int(field.PEAddress - modifierPredicatePEBase4E03D0), int(uintptr(native) - uintptr(heap)), width, field.Kind == "pointer"}
		for _, old := range m.slots {
			if slot.native < old.native+old.width && old.native < slot.native+slot.width {
				t.Fatal("overlapping declared native fields", field.Field)
			}
		}
		m.slots = append(m.slots, slot)
	}
	for i, field := range row.Fields {
		slot := m.slots[i]
		value := uintptr(field.Value)
		if slot.pointer {
			value = uintptr(m.fromPE(field.Value))
		}
		for j := 0; j < slot.width; j++ {
			m.buf[slot.native+j] = byte(value >> (8 * j))
		}
	}
	m.argument, m.out = m.fromPE(row.Argument), m.fromPE(row.Output)
	m.before = bytes.Clone(m.buf)
	if fmt.Sprintf("%x", sha256.Sum256(m.project(row.Output))) != row.InitialSHA {
		t.Fatal("complete native-to-PE32 initial projection", row.Name)
	}
	return m
}

func (m *modifierPredicateMemory4E03D0) fromPE(pe uint32) unsafe.Pointer {
	if p, ok := m.addresses[pe]; ok {
		return p
	}
	if pe >= modifierPredicatePEBase4E03D0 && pe < modifierPredicatePEBase4E03D0+0x10000 {
		p := unsafe.Add(m.heap, pe-modifierPredicatePEBase4E03D0)
		m.addresses[pe], m.reverse[uintptr(p)] = p, pe
		return p
	}
	m.t.Fatalf("undeclared PE32 pointer %#x", pe)
	return nil
}

func (m *modifierPredicateMemory4E03D0) toPE(p uintptr) uint32 {
	if pe, ok := m.reverse[p]; ok {
		return pe
	}
	m.t.Fatalf("undeclared native pointer %#x", p)
	return 0
}

func (m *modifierPredicateMemory4E03D0) project(output uint32) []byte {
	pe := bytes.Clone(m.pe)
	for _, slot := range m.slots {
		if slot.pointer {
			var value uintptr
			for j := 0; j < slot.width; j++ {
				value |= uintptr(m.buf[slot.native+j]) << (8 * j)
			}
			binary.LittleEndian.PutUint32(pe[slot.pe:], m.toPE(value))
		} else {
			copy(pe[slot.pe:slot.pe+4], m.buf[slot.native:slot.native+4])
		}
	}
	if output >= modifierPredicatePEBase4E03D0 && output <= modifierPredicatePEBase4E03D0+0x10000-4 {
		copy(pe[output-modifierPredicatePEBase4E03D0:], unsafe.Slice((*byte)(m.fromPE(output)), 4))
	}
	return pe
}

var modifierPredicateReadKinds4E03D0 = map[string]int{"strength": 0, "class": 1, "init": 2, "slot": 3, "callback": 4, "word": 5}

func (m *modifierPredicateMemory4E03D0) initialGo(p unsafe.Pointer, kind, index int) (out *uint32) {
	old := debug.SetPanicOnFault(true)
	defer debug.SetPanicOnFault(old)
	defer func() {
		if failure := recover(); failure != nil {
			if _, ok := failure.(interface{ RuntimeError() }); !ok {
				m.t.Fatalf("initial typed read did not produce a memory fault: %v", failure)
			}
			m.faults++
			out = nil
		}
	}()
	var value uint32
	switch kind {
	case 0:
		value = uint32((*server.ModifierEff)(p).DefendCollide88.Val)
	case 1:
		value = uint32((*server.Object)(p).ObjClass)
	case 2:
		value = m.toPE(uintptr((*server.Object)(p).InitData))
	case 3:
		value = m.toPE(uintptr(unsafe.Pointer((*server.ModifierInitData)(p).Modifiers[index])))
	case 4:
		value = m.toPE(uintptr((*server.ModifierEff)(p).DefendCollide88.Fnc))
	case 5:
		value = *(*uint32)(p)
	default:
		m.t.Fatal("undeclared typed read")
	}
	return &value
}

func (m *modifierPredicateMemory4E03D0) checkInitial(row modifierPredicateInitial4E03D0) {
	m.t.Helper()
	for _, read := range row.Reads {
		kind, ok := modifierPredicateReadKinds4E03D0[read.Kind]
		if !ok {
			m.t.Fatal("undeclared independent field read")
		}
		p := m.fromPE(read.Role)
		got := m.initialGo(p, kind, read.Index)
		if (got == nil) != (read.Want == nil) || (got != nil && *got != *read.Want) {
			m.t.Fatal("independent typed initial word/pointer/fault domain", read.Kind, read.Index, got, read.Want)
		}
		if read.Want != nil {
			c := modifierPredicatesInitialC4E03D0(p, kind, read.Index)
			value := uint32(c)
			if kind == 2 || kind == 3 || kind == 4 {
				value = m.toPE(c)
			}
			if value != *read.Want {
				m.t.Fatal("independent typed C initial field", read.Kind)
			}
		}
	}
	if !bytes.Equal(m.buf, m.before) || fmt.Sprintf("%x", sha256.Sum256(m.project(row.Output))) != row.InitialSHA {
		m.t.Fatal("whole native heap and complete PE32 projection changed during independent reads")
	}
}

func TestModifierDefendPredicatesNativeLayout4E03D0(t *testing.T) {
	e, o, d := server.ModifierEff{}, server.Object{}, server.ModifierInitData{}
	index, ok := reflect.TypeOf(e).FieldByName("ind4")
	if !ok || index.Type.Kind() != reflect.Uint32 {
		t.Fatal("modifier index DWORD")
	}
	word := unsafe.Sizeof(uintptr(0))
	want := [13]uintptr{unsafe.Sizeof(e), unsafe.Offsetof(e.DefendCollide88) + unsafe.Offsetof(e.DefendCollide88.Fnc), unsafe.Offsetof(e.DefendCollide88) + unsafe.Offsetof(e.DefendCollide88.Val), index.Offset,
		unsafe.Offsetof(o.ScriptPickup) + unsafe.Sizeof(o.ScriptPickup), unsafe.Offsetof(o.ObjClass), unsafe.Offsetof(o.InitData), unsafe.Sizeof(d), 0, word, 2 * word, 3 * word, word}
	if got := modifierPredicatesLayout4E03D0(); got != want {
		t.Fatal("independent complete C/Go layouts", got, want)
	}
}

func TestModifierDefendPredicatesInitialBefore4E03D0(t *testing.T) {
	if os.Getenv("OPENNOX_MODIFIER_PREDICATES_BEFORE") != "1" {
		return
	}
	for i := 0; i < 3; i++ {
		if modifierPredicatesABI4E03D0(i) != 1 {
			t.Fatal("initial memory freeze must precede all three native C entries")
		}
	}
}

func TestModifierDefendPredicatesInitialMemory4E03D0(t *testing.T) {
	TestModifierDefendPredicatesNativeLayout4E03D0(t)
	total := 0
	for _, row := range modifierPredicateInitials4E03D0(t) {
		t.Run(row.Name, func(t *testing.T) {
			m := newModifierPredicateMemory4E03D0(t, row)
			m.checkInitial(row)
			total += m.faults
		})
	}
	if total != 136 {
		t.Fatal("all independent null/protected read domains", total)
	}
	t.Logf("INITIAL_MODIFIER_NATIVE_READ_FAULT_DOMAINS=%d", total)
}

func TestModifierDefendPredicatesActualCBefore4E03D0(t *testing.T) {
	if os.Getenv("OPENNOX_MODIFIER_PREDICATES_BEFORE") != "1" {
		return
	}
	if entry := os.Getenv("OPENNOX_MODIFIER_PREDICATES_CHILD"); entry != "" {
		i, err := strconv.Atoi(entry)
		if err != nil || i < 0 || i > 2 || unsafe.Sizeof(uintptr(0)) != 8 || modifierPredicatesABI4E03D0(i) != 1 {
			t.Fatal("actual unchanged integer C entry required")
		}
		name := []string{"inversion/00000001/separate", "grip/00000000/separate", "scan/only-slot-2-00000001"}[i]
		for _, row := range modifierPredicateInitials4E03D0(t) {
			if row.Name != name {
				continue
			}
			m := newModifierPredicateMemory4E03D0(t, row)
			m.checkInitial(row)
			t.Logf("MODIFIER_REAL_C_INITIAL_OK entry=%d argument=%p output=%p whole-native-heap-layout-C-Go-getters=true", i, m.argument, m.out)
			modifierPredicatesPublic4E03D0(i, m.argument, row.OpaqueArgs, m.out)
			t.Fatal("old integer C entry unexpectedly returned")
		}
		t.Fatal("required independent positive control missing")
	}
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Fatal("pointer truncation reproduction requires a 64-bit host")
	}
	for i, name := range []string{"Inversion", "Grip", "GripTraversal"} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestModifierDefendPredicatesActualCBefore4E03D0$", "-test.v")
			cmd.Env = append(os.Environ(), "OPENNOX_MODIFIER_PREDICATES_CHILD="+strconv.Itoa(i))
			out, err := cmd.CombinedOutput()
			t.Logf("actual unchanged %s C entry: %v\n%s", name, err, out)
			if ctx.Err() != nil || err == nil || !bytes.Contains(out, []byte("MODIFIER_REAL_C_INITIAL_OK")) || bytes.Contains(out, []byte("unexpectedly returned")) {
				t.Fatal("independent native controls must precede actual old C fault", err)
			}
			if !bytes.Contains(out, []byte("SIGSEGV")) && !bytes.Contains(out, []byte("SIGBUS")) && !bytes.Contains(out, []byte("unexpected fault address")) {
				t.Fatal("actual C pointer memory fault required")
			}
		})
	}
}
