package legacy

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"os/exec"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type inversionInventoryInitial4FA4F0 struct {
	modifierPredicateInitial4E03D0
	Target uint32
}

func inversionInventoryInitials4FA4F0(t *testing.T) []inversionInventoryInitial4FA4F0 {
	t.Helper()
	var rows []inversionInventoryInitial4FA4F0
	modifierPredicateFixture4E03D0(t, "inversion_inventory_initial_memory_4fa4f0.json.gz", "1b79975ddc7aeac9f60850ada1cd975fdbe105a52556c433e8e35b1e499755a2", "6e21535d557aeb9ff7a6b9a7d0b6cd77eb8c2fde6b889f696909a6158bd2117e", &rows)
	if len(rows) != 207 {
		t.Fatal("all independent raw inventory inputs")
	}
	return rows
}

func newInversionInventoryMemory4FA4F0(t *testing.T, row inversionInventoryInitial4FA4F0) *modifierPredicateMemory4E03D0 {
	t.Helper()
	heap, guard, free := generatorCreditOriginalHeap4D61B0()
	t.Cleanup(free)
	off, inventory := modifierPredicatesLayout4E03D0(), inversionInventoryLayout4FA4F0()
	if heap == nil || guard < off[4]+1024 || (off[12] == 8 && uintptr(heap) <= math.MaxUint32) {
		t.Fatal("complete C-owned high heap and protected invalid domain")
	}
	m := &modifierPredicateMemory4E03D0{t: t, heap: heap, off: off, buf: unsafe.Slice((*byte)(heap), 0x10000), pe: make([]byte, 0x10000), addresses: make(map[uint32]unsafe.Pointer), reverse: make(map[uintptr]uint32)}
	for i := range m.pe {
		m.pe[i] = byte((i*23 + row.Seed*31) ^ (i >> 5))
	}
	for _, word := range row.Words {
		binary.LittleEndian.PutUint32(m.pe[word.Address-modifierPredicatePEBase4E03D0:], word.Value)
	}
	if fmt.Sprintf("%x", sha256.Sum256(m.pe)) != row.InitialSHA {
		t.Fatal("independent complete PE32 inventory initial heap", row.Name)
	}
	copy(m.buf, m.pe)
	add := func(pe uint32, p unsafe.Pointer) {
		if old, ok := m.addresses[pe]; ok && old != p {
			t.Fatal("contradictory inventory pointer correspondence")
		}
		m.addresses[pe], m.reverse[uintptr(p)] = p, pe
	}
	add(0, nil)
	add(0x41000000, generatorCreditOriginalBad4D61B0(heap, guard, 0))
	add(0x4e03d0, modifierPredicatesCallback4E03D0(0))
	add(0x4e0480, modifierPredicatesCallback4E03D0(1))
	add(0x4fa4f0, inversionInventoryPointer4FA4F0(false))
	add(0x4ec580, inversionInventoryPointer4FA4F0(true))
	for _, role := range []int{0x1000, 0x2000, 0x3000, 0x4000, 0x5000, 0x6000, 0x7000, 0x8000, 0x9000, 0xa000, 0xb000, 0xc000, 0xd000, 0xe000} {
		add(modifierPredicatePEBase4E03D0+uint32(role), unsafe.Add(heap, role))
	}
	fields := map[string]uintptr{"effect_callback": off[1], "effect_strength": off[2], "object_class": off[5], "object_init": off[6], "slot0": off[8], "slot1": off[9], "slot2": off[10], "slot3": off[11],
		"inventory_first": inventory[0], "inventory_next": inventory[1], "object_owner": inventory[2], "object_flags": inventory[3]}
	for _, field := range row.Fields {
		offset, ok := fields[field.Field]
		if !ok {
			t.Fatal("undeclared inventory field", field.Field)
		}
		p := unsafe.Add(m.fromPE(field.Role), offset)
		add(field.PEAddress, p)
		width := 4
		if field.Kind == "pointer" {
			width = int(off[12])
		} else if field.Kind != "word" {
			t.Fatal("undeclared inventory field width")
		}
		slot := modifierPredicateSlot4E03D0{int(field.PEAddress - modifierPredicatePEBase4E03D0), int(uintptr(p) - uintptr(heap)), width, field.Kind == "pointer"}
		for _, old := range m.slots {
			if slot.native < old.native+old.width && old.native < slot.native+slot.width {
				t.Fatal("overlapping independent inventory fields", field.Field)
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
	m.argument = m.fromPE(row.Argument)
	m.before = bytes.Clone(m.buf)
	if fmt.Sprintf("%x", sha256.Sum256(m.project(0))) != row.InitialSHA {
		t.Fatal("complete independent inventory native-to-PE32 initial projection")
	}
	return m
}

var inversionInventoryReadKinds4FA4F0 = map[string]int{"effect_strength": 0, "object_class": 1, "object_init": 2, "slot": 3, "effect_callback": 4,
	"inventory_first": 6, "inventory_next": 7, "object_owner": 8, "object_flags": 9, "class_byte": 10}

func inversionInventoryPointerKind4FA4F0(kind int) bool {
	return kind == 2 || kind == 3 || kind == 4 || kind == 6 || kind == 7 || kind == 8
}

func inversionInventoryInitialGo4FA4F0(m *modifierPredicateMemory4E03D0, p unsafe.Pointer, kind, index int) (value *uint32) {
	if kind < 6 {
		return m.initialGo(p, kind, index)
	}
	var result uint32
	fault := modifierPredicateRecover4E03D0(m.t, func() {
		o := (*server.Object)(p)
		switch kind {
		case 6:
			result = m.toPE(uintptr(unsafe.Pointer(o.InvFirstItem)))
		case 7:
			result = m.toPE(uintptr(unsafe.Pointer(o.InvNextItem)))
		case 8:
			result = m.toPE(uintptr(unsafe.Pointer(o.ObjOwner)))
		case 9:
			result = uint32(o.ObjFlags)
		case 10:
			result = uint32(uint8(o.ObjClass))
		default:
			m.t.Fatal("undeclared inventory typed read")
		}
	})
	if fault {
		m.faults++
		return nil
	}
	return &result
}

func checkInversionInventoryInitial4FA4F0(m *modifierPredicateMemory4E03D0, row inversionInventoryInitial4FA4F0) {
	m.t.Helper()
	for _, read := range row.Reads {
		kind, ok := inversionInventoryReadKinds4FA4F0[read.Kind]
		if !ok {
			m.t.Fatal("undeclared independent inventory read")
		}
		p := m.fromPE(read.Role)
		got := inversionInventoryInitialGo4FA4F0(m, p, kind, read.Index)
		if (got == nil) != (read.Want == nil) || (got != nil && *got != *read.Want) {
			m.t.Fatal("independent inventory Go word/pointer/byte/fault", read.Kind)
		}
		if read.Want != nil {
			c := inversionInventoryInitialC4FA4F0(p, kind, read.Index)
			value := uint32(c)
			if inversionInventoryPointerKind4FA4F0(kind) {
				value = m.toPE(c)
			}
			if value != *read.Want {
				m.t.Fatal("independent typed C inventory word/pointer/byte", read.Kind)
			}
		}
	}
	if !bytes.Equal(m.buf, m.before) || fmt.Sprintf("%x", sha256.Sum256(m.project(0))) != row.InitialSHA {
		m.t.Fatal("whole native inventory heap and complete original projection changed during typed reads")
	}
}

func TestInversionInventoryNativeLayout4FA4F0(t *testing.T) {
	TestModifierDefendPredicatesNativeLayout4E03D0(t)
	o := server.Object{}
	want := [5]uintptr{unsafe.Offsetof(o.InvFirstItem), unsafe.Offsetof(o.InvNextItem), unsafe.Offsetof(o.ObjOwner), unsafe.Offsetof(o.ObjFlags), unsafe.Offsetof(o.ObjClass)}
	if got := inversionInventoryLayout4FA4F0(); got != want {
		t.Fatal("independent C/Go inventory/owner/flags/class layouts", got, want)
	}
}

func TestInversionInventoryInitialBefore4FA4F0(t *testing.T) {
	if os.Getenv("OPENNOX_INVERSION_INVENTORY_BEFORE") == "1" {
		if inversionInventoryABI4FA4F0() != 1 {
			t.Fatal("independent memory must freeze before raw checker prototype changes")
		}
		if _, err := os.Stat("inversion_inventory_native_4fa4f0.go"); !os.IsNotExist(err) {
			t.Fatal("raw checker engine must remain absent")
		}
	}
}

func TestInversionInventoryInitialMemory4FA4F0(t *testing.T) {
	TestInversionInventoryNativeLayout4FA4F0(t)
	total := 0
	for _, row := range inversionInventoryInitials4FA4F0(t) {
		t.Run(row.Name, func(t *testing.T) {
			m := newInversionInventoryMemory4FA4F0(t, row)
			checkInversionInventoryInitial4FA4F0(m, row)
			total += m.faults
		})
	}
	if total != 242 {
		t.Fatal("all independent inventory read-fault domains", total)
	}
	t.Logf("INVENTORY_INITIAL_READ_FAULT_DOMAINS=%d", total)
}

type inversionInventoryParentOriginal4FA4F0 struct {
	Name  string
	Calls []struct {
		Function string
		Args     []uint32
	}
	Returns []struct {
		Function string
		Result   uint32
	}
	Fault *struct{ Instruction string }
}

func inversionInventoryParentOriginals4FA4F0(t *testing.T) []inversionInventoryParentOriginal4FA4F0 {
	t.Helper()
	var originals []inversionInventoryParentOriginal4FA4F0
	modifierPredicateFixture4E03D0(t, "inversion_inventory_original_4fa4f0.json.gz", "0bb44fa16475b2311e8d1ec5aa862fbeb87b7b94141719efe681e44310269853", "df0ec5b5b7bb81bc8c36e7e7ff2f7e936f3d741949926d91c2c4f86e6dcea609", &originals)
	if len(originals) != 207 {
		t.Fatal("complete actual-original parent observations")
	}
	return originals
}

func TestInversionInventoryActualParent4FA4F0(t *testing.T) {
	initial, originals := inversionInventoryInitials4FA4F0(t), inversionInventoryParentOriginals4FA4F0(t)
	calls, faults := 0, 0
	for i, row := range initial {
		t.Run(row.Name, func(t *testing.T) {
			if originals[i].Name != row.Name {
				t.Fatal("independent input/original identity")
			}
			m := newInversionInventoryMemory4FA4F0(t, row)
			var returned []uint32
			for _, result := range originals[i].Returns {
				if result.Function == "parent" {
					returned = append(returned, result.Result)
				}
			}
			index := 0
			for _, call := range originals[i].Calls {
				if call.Function != "parent" {
					continue
				}
				if len(call.Args) != 1 || call.Args[0] != row.Target {
					t.Fatal("actual original owner-service argument")
				}
				var got unsafe.Pointer
				fault := modifierPredicateRecover4E03D0(t, func() { got = inversionInventoryParent4FA4F0(m.fromPE(row.Target)) })
				if fault != (index >= len(returned)) {
					t.Fatal("actual default C parent-service original fault domain")
				}
				if fault {
					if originals[i].Fault == nil || (originals[i].Fault.Instruction != "0x4ec588" && originals[i].Fault.Instruction != "0x4ec59b") {
						t.Fatal("original parent read fault required")
					}
					faults++
				} else if m.toPE(uintptr(got)) != returned[index] {
					t.Fatal("actual default C parent return against original machine body")
				}
				index++
				calls++
			}
			if !bytes.Equal(m.buf, m.before) || fmt.Sprintf("%x", sha256.Sum256(m.project(0))) != row.InitialSHA {
				t.Fatal("whole native heap and original projection changed by actual default parent service")
			}
		})
	}
	if calls != 267 || faults != 4 {
		t.Fatal("every original parent call and fault", calls, faults)
	}
	t.Logf("INVENTORY_ACTUAL_DEFAULT_PARENT_CALLS=%d MEMORY_FAULTS=%d", calls, faults)
}

func TestInversionInventoryActualCBefore4FA4F0(t *testing.T) {
	if os.Getenv("OPENNOX_INVERSION_INVENTORY_BEFORE") != "1" {
		return
	}
	if os.Getenv("OPENNOX_INVERSION_INVENTORY_CHILD") == "1" {
		if unsafe.Sizeof(uintptr(0)) != 8 || inversionInventoryABI4FA4F0() != 1 {
			t.Fatal("actual old two-integer raw checker required")
		}
		for _, row := range inversionInventoryInitials4FA4F0(t) {
			if row.Name != "signed-00000001-00000001" {
				continue
			}
			m := newInversionInventoryMemory4FA4F0(t, row)
			checkInversionInventoryInitial4FA4F0(m, row)
			target := m.fromPE(row.Target)
			if inversionInventoryParent4FA4F0(target) != target || !bytes.Equal(m.buf, m.before) {
				t.Fatal("actual typed parent-service positive control")
			}
			t.Logf("INVENTORY_REAL_C_INITIAL_OK root=%p target=%p independent-layout-C-Go-fields-parent-whole-heap=true", m.argument, target)
			inversionInventoryPublic4FA4F0(m.argument, target)
			t.Fatal("old two-integer raw checker unexpectedly returned")
		}
		t.Fatal("required independent positive control missing")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestInversionInventoryActualCBefore4FA4F0$", "-test.v")
	cmd.Env = append(os.Environ(), "OPENNOX_INVERSION_INVENTORY_CHILD=1")
	out, err := cmd.CombinedOutput()
	t.Logf("actual unchanged raw checker: %v\n%s", err, out)
	if ctx.Err() != nil || err == nil || !bytes.Contains(out, []byte("INVENTORY_REAL_C_INITIAL_OK")) || bytes.Contains(out, []byte("unexpectedly returned")) {
		t.Fatal("full independent native controls must precede actual old-C fault", err)
	}
	if !bytes.Contains(out, []byte("SIGSEGV")) && !bytes.Contains(out, []byte("SIGBUS")) && !bytes.Contains(out, []byte("unexpected fault address")) {
		t.Fatal("actual raw-checker high-pointer C fault required")
	}
}
