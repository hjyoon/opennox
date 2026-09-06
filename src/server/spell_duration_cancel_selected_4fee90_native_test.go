package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"
)

func requireSpellDurationCancelSelectedNativePointers4FEE90(t *testing.T, values ...unsafe.Pointer) {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) != 8 {
		return
	}
	for i, value := range values {
		if value == nil || uintptr(value) <= math.MaxUint32 {
			t.Fatalf("pointer %d = %p, want native address above 4 GiB", i, value)
		}
	}
}

func TestSpellDurationCancelSelected4FEE90NativeLayouts(t *testing.T) {
	type layout struct {
		durSize uintptr
		caster  uintptr
		next    uintptr
		spSize  uintptr
		list    uintptr
	}
	wants := map[uintptr]layout{
		4: {durSize: 120, caster: 16, next: 116, spSize: 16, list: 8},
		8: {durSize: 184, caster: 24, next: 176, spSize: 32, list: 16},
	}
	ptrSize := unsafe.Sizeof(uintptr(0))
	want, ok := wants[ptrSize]
	if !ok {
		t.Fatalf("unsupported pointer size %d", ptrSize)
	}
	got := layout{
		durSize: unsafe.Sizeof(DurSpell{}),
		caster:  unsafe.Offsetof(DurSpell{}.Caster16),
		next:    unsafe.Offsetof(DurSpell{}.Next),
		spSize:  unsafe.Sizeof(SpellsDuration{}),
		list:    unsafe.Offsetof(SpellsDuration{}.List),
	}
	if got != want {
		t.Fatalf("native layout on %s/%s = %+v, want %+v", runtime.GOOS, runtime.GOARCH, got, want)
	}
	if got := unsafe.Offsetof(DurSpell{}.Spell); got != 4 {
		t.Fatalf("DurSpell.Spell offset = %d, want 4", got)
	}
	if got := unsafe.Sizeof(DurSpell{}.Caster16); got != ptrSize {
		t.Fatalf("DurSpell.Caster16 width = %d, want native pointer width %d", got, ptrSize)
	}
}

func TestSpellDurationCancelSelectedNative4FEE90PreservesPointersAndOrder(t *testing.T) {
	recordA := &DurSpell{Spell: 24}
	recordB := &DurSpell{Spell: math.MaxUint32}
	caster := new(Object)
	other := new(Object)
	recordA.Caster16 = other
	recordA.Next = recordB
	recordB.Caster16 = caster
	recordB.Spell = 59
	requireSpellDurationCancelSelectedNativePointers4FEE90(t,
		unsafe.Pointer(recordA), unsafe.Pointer(recordB), unsafe.Pointer(caster), unsafe.Pointer(other),
	)

	var events []string
	var canceled []*DurSpell
	name := func(record *DurSpell) string {
		if record == recordA {
			return "A"
		}
		if record == recordB {
			return "B"
		}
		return "nil"
	}
	spellDurationCancelSelectedNative4FEE90(caster, spellDurationCancelSelectedNativeDeps4FEE90{
		loadFirst: func() *DurSpell {
			events = append(events, "head")
			return recordA
		},
		loadNext: func(record *DurSpell) *DurSpell {
			events = append(events, "next:"+name(record))
			return record.Next
		},
		loadCaster: func(record *DurSpell) *Object {
			events = append(events, "caster:"+name(record))
			return record.Caster16
		},
		loadSpell: func(record *DurSpell) uint32 {
			events = append(events, "spell:"+name(record))
			return record.Spell
		},
		cancel: func(record *DurSpell) {
			events = append(events, "cancel:"+name(record))
			canceled = append(canceled, record)
		},
	})
	wantEvents := []string{
		"head", "next:A", "caster:A",
		"next:B", "caster:B", "spell:B", "cancel:B",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %q, want %q", events, wantEvents)
	}
	if !reflect.DeepEqual(canceled, []*DurSpell{recordB}) {
		t.Fatalf("canceled = %p, want [%p]", canceled, recordB)
	}
	runtime.KeepAlive(recordA)
	runtime.KeepAlive(recordB)
	runtime.KeepAlive(caster)
	runtime.KeepAlive(other)
}

func TestSpellDurationCancelSelected4FEE90ServerBinding(t *testing.T) {
	caster := new(Object)
	other := new(Object)
	recordA := &DurSpell{Spell: 24, Caster16: caster, Flags88: 0x12345620}
	recordB := &DurSpell{Spell: 23, Caster16: caster, Flags88: 0x89abcdee}
	recordC := &DurSpell{Spell: 59, Caster16: other, Flags88: 0xfedcba40}
	recordD := &DurSpell{Spell: 67, Caster16: caster, Flags88: 0x76543280}
	recordA.Next = recordB
	recordB.Next = recordC
	recordC.Next = recordD
	durations := SpellsDuration{List: recordA}
	requireSpellDurationCancelSelectedNativePointers4FEE90(t,
		unsafe.Pointer(recordA), unsafe.Pointer(recordB), unsafe.Pointer(recordC), unsafe.Pointer(recordD),
		unsafe.Pointer(caster), unsafe.Pointer(other),
	)

	durations.SpellDurationCancelSelected4FEE90(caster)
	if recordA.Flags88 != 0x12345621 {
		t.Fatalf("selected A flags = %#x, want 0x12345621", recordA.Flags88)
	}
	if recordB.Flags88 != 0x89abcdee {
		t.Fatalf("unselected B flags = %#x, want unchanged", recordB.Flags88)
	}
	if recordC.Flags88 != 0xfedcba40 {
		t.Fatalf("other-caster C flags = %#x, want unchanged", recordC.Flags88)
	}
	if recordD.Flags88 != 0x76543281 {
		t.Fatalf("selected D flags = %#x, want 0x76543281", recordD.Flags88)
	}
	runtime.KeepAlive(recordA)
	runtime.KeepAlive(recordB)
	runtime.KeepAlive(recordC)
	runtime.KeepAlive(recordD)
	runtime.KeepAlive(caster)
	runtime.KeepAlive(other)
}

func TestSpellDurationCancelSelected4FEE90ServerNilCasterIdentity(t *testing.T) {
	recordA := &DurSpell{Spell: 35, Flags88: 0xabcdef40}
	recordB := &DurSpell{Spell: 43, Flags88: 0x12345680}
	recordA.Next = recordB
	durations := SpellsDuration{List: recordA}
	durations.SpellDurationCancelSelected4FEE90(nil)
	if recordA.Flags88 != 0xabcdef41 || recordB.Flags88 != 0x12345681 {
		t.Fatalf("nil-caster flags = %#x/%#x, want 0xabcdef41/0x12345681", recordA.Flags88, recordB.Flags88)
	}
}

func TestSpellDurationCancelSelected4FEE90ServerNilReceiverFaultsAtHead(t *testing.T) {
	var durations *SpellsDuration
	defer func() {
		if recover() == nil {
			t.Fatal("nil SpellsDuration receiver did not fault at the head load")
		}
	}()
	durations.SpellDurationCancelSelected4FEE90(nil)
}
