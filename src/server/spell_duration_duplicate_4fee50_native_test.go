package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"
)

func requireSpellDurationDuplicateNativePointers4FEE50(t *testing.T, values ...unsafe.Pointer) {
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

func TestSpellDurationDuplicate4FEE50NativeLayouts(t *testing.T) {
	type layout struct {
		durSize uintptr
		caster  uintptr
		flag20  uintptr
		flags88 uintptr
		next    uintptr
		spSize  uintptr
		list    uintptr
	}
	wants := map[uintptr]layout{
		4: {durSize: 120, caster: 16, flag20: 20, flags88: 88, next: 116, spSize: 16, list: 8},
		8: {durSize: 184, caster: 24, flag20: 32, flags88: 120, next: 176, spSize: 32, list: 16},
	}
	ptrSize := unsafe.Sizeof(uintptr(0))
	want, ok := wants[ptrSize]
	if !ok {
		t.Fatalf("unsupported pointer size %d", ptrSize)
	}
	got := layout{
		durSize: unsafe.Sizeof(DurSpell{}),
		caster:  unsafe.Offsetof(DurSpell{}.Caster16),
		flag20:  unsafe.Offsetof(DurSpell{}.Flag20),
		flags88: unsafe.Offsetof(DurSpell{}.Flags88),
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

func TestSpellDurationDuplicateNative4FEE50PreservesPointersAndShortCircuitOrder(t *testing.T) {
	caster := new(Object)
	other := new(Object)
	recordA := &DurSpell{Flag20: 0x100}
	recordB := &DurSpell{Spell: 0x80000000, Caster16: other}
	recordC := &DurSpell{Spell: 0x80000000, Caster16: caster, Flags88: 0xffffff80}
	recordA.Next = recordB
	recordB.Next = recordC
	requireSpellDurationDuplicateNativePointers4FEE50(t,
		unsafe.Pointer(caster), unsafe.Pointer(other), unsafe.Pointer(recordA),
		unsafe.Pointer(recordB), unsafe.Pointer(recordC),
	)

	var events []string
	name := func(record *DurSpell) string {
		switch record {
		case recordA:
			return "A"
		case recordB:
			return "B"
		case recordC:
			return "C"
		default:
			return "nil"
		}
	}
	got := spellDurationDuplicateNative4FEE50(math.MinInt32, caster, spellDurationDuplicateNativeDeps4FEE50{
		loadHead: func() *DurSpell {
			events = append(events, "head")
			return recordA
		},
		loadFlag20: func(record *DurSpell) uint32 {
			events = append(events, "flag:"+name(record))
			return record.Flag20
		},
		loadSpell: func(record *DurSpell) uint32 {
			events = append(events, "spell:"+name(record))
			return record.Spell
		},
		loadCaster: func(record *DurSpell) *Object {
			events = append(events, "caster:"+name(record))
			return record.Caster16
		},
		loadFlagsLowByte: func(record *DurSpell) byte {
			events = append(events, "flags:"+name(record))
			return byte(record.Flags88)
		},
		loadNext: func(record *DurSpell) *DurSpell {
			events = append(events, "next:"+name(record))
			return record.Next
		},
	})
	if got != 1 {
		t.Fatalf("result = %d, want canonical one", got)
	}
	wantEvents := []string{
		"head", "flag:A", "next:A",
		"flag:B", "spell:B", "caster:B", "next:B",
		"flag:C", "spell:C", "caster:C", "flags:C",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %q, want %q", events, wantEvents)
	}
	runtime.KeepAlive(caster)
	runtime.KeepAlive(other)
	runtime.KeepAlive(recordA)
	runtime.KeepAlive(recordB)
	runtime.KeepAlive(recordC)
}

func TestSpellDurationDuplicate4FEE50ServerAndCompatibilityBindings(t *testing.T) {
	caster := new(Object)
	other := new(Object)
	recordA := &DurSpell{Spell: 59, Caster16: caster, Flags88: 0x12345601}
	recordB := &DurSpell{Spell: 59, Caster16: other, Flags88: 0x89abcde0}
	recordC := &DurSpell{Spell: 59, Caster16: caster, Flags88: 0xffffff80}
	recordA.Next = recordB
	recordB.Next = recordC
	durations := SpellsDuration{List: recordA}
	requireSpellDurationDuplicateNativePointers4FEE50(t,
		unsafe.Pointer(caster), unsafe.Pointer(other), unsafe.Pointer(recordA),
		unsafe.Pointer(recordB), unsafe.Pointer(recordC),
	)

	if got := durations.SpellDurationDuplicate4FEE50(59, caster); got != 1 {
		t.Fatalf("server result = %d, want canonical one", got)
	}
	if !durations.Sub4FEE50(spell.ID(59), caster) {
		t.Fatal("compatibility binding did not use restored predicate")
	}
	if got := durations.SpellDurationDuplicate4FEE50(59, new(Object)); got != 0 {
		t.Fatalf("different caster result = %d, want canonical zero", got)
	}

	deps := spellDurationCreateServerDeps4FEBA0(&durations, SpellDurationCreateRuntime4FEBA0{})
	if got := deps.findDuplicate(59, caster); got != 1 {
		t.Fatalf("creation duplicate binding = %d, want canonical one", got)
	}
	runtime.KeepAlive(caster)
	runtime.KeepAlive(other)
	runtime.KeepAlive(recordA)
	runtime.KeepAlive(recordB)
	runtime.KeepAlive(recordC)
}

func TestSpellDurationDuplicate4FEE50ServerNilCasterDisabledAndEmpty(t *testing.T) {
	record := &DurSpell{Spell: 43, Flags88: 0xfffffffe}
	durations := SpellsDuration{List: record}
	if got := durations.SpellDurationDuplicate4FEE50(43, nil); got != 1 {
		t.Fatalf("nil-caster result = %d, want canonical one", got)
	}
	record.Flag20 = 1
	if got := durations.SpellDurationDuplicate4FEE50(43, nil); got != 0 {
		t.Fatalf("disabled-record result = %d, want canonical zero", got)
	}
	var empty SpellsDuration
	if got := empty.SpellDurationDuplicate4FEE50(43, nil); got != 0 {
		t.Fatalf("empty result = %d, want canonical zero", got)
	}
}
