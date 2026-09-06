package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"
)

func requireSpellCancelDurSpellNativePointers4FEB10(t *testing.T, values ...unsafe.Pointer) {
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

func TestSpellCancelDurSpell4FEB10NativeLayouts(t *testing.T) {
	wantDurSize := uintptr(120)
	wantCaster := uintptr(16)
	wantNext := uintptr(116)
	wantSpellsDurationSize := uintptr(16)
	wantList := uintptr(8)
	wantLastID := uintptr(12)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantDurSize = 184
		wantCaster = 24
		wantNext = 176
		wantSpellsDurationSize = 32
		wantList = 16
		wantLastID = 24
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"DurSpell size", unsafe.Sizeof(DurSpell{}), wantDurSize},
		{"DurSpell.Spell", unsafe.Offsetof(DurSpell{}.Spell), 4},
		{"DurSpell.Caster16", unsafe.Offsetof(DurSpell{}.Caster16), wantCaster},
		{"DurSpell.Next", unsafe.Offsetof(DurSpell{}.Next), wantNext},
		{"Caster16 width", unsafe.Sizeof(DurSpell{}.Caster16), unsafe.Sizeof(uintptr(0))},
		{"Next width", unsafe.Sizeof(DurSpell{}.Next), unsafe.Sizeof(uintptr(0))},
		{"SpellsDuration size", unsafe.Sizeof(SpellsDuration{}), wantSpellsDurationSize},
		{"SpellsDuration.List", unsafe.Offsetof(SpellsDuration{}.List), wantList},
		{"SpellsDuration.lastID", unsafe.Offsetof(SpellsDuration{}.lastID), wantLastID},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestSpellCancelDurSpellNative4FEB10PreservesPointersAndRepeatedCasterLoad(t *testing.T) {
	recordA := &DurSpell{Spell: 75}
	recordB := &DurSpell{Spell: 114}
	recordC := &DurSpell{Spell: 74}
	caster := new(Object)
	other := new(Object)
	recordA.Caster16 = other
	recordA.Next = recordB
	recordB.Caster16 = caster
	recordB.Next = recordC
	recordC.Caster16 = caster
	requireSpellCancelDurSpellNativePointers4FEB10(t,
		unsafe.Pointer(recordA), unsafe.Pointer(recordB), unsafe.Pointer(recordC),
		unsafe.Pointer(caster), unsafe.Pointer(other),
	)

	var events []string
	var cancelled []*DurSpell
	casterReadsA := 0
	got := spellCancelDurSpellNative4FEB10(75, caster, spellCancelDurSpellNativeDeps4FEB10{
		loadFirst: func() *DurSpell {
			events = append(events, "first")
			return recordA
		},
		loadSpell: func(record *DurSpell) int32 {
			events = append(events, "spell")
			return int32(record.Spell)
		},
		loadNext: func(record *DurSpell) *DurSpell {
			events = append(events, "next")
			return record.Next
		},
		loadCaster: func(record *DurSpell) *Object {
			events = append(events, "caster")
			value := record.Caster16
			if record == recordA {
				casterReadsA++
				if casterReadsA == 1 {
					record.Caster16 = caster
				}
			}
			return value
		},
		cancel: func(record *DurSpell) {
			events = append(events, "cancel")
			cancelled = append(cancelled, record)
		},
	})

	wantEvents := []string{
		"first",
		"spell", "next", "caster", "caster", "cancel",
		"spell", "next", "caster", "cancel",
		"spell", "next",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %q, want %q", events, wantEvents)
	}
	wantCancelled := []*DurSpell{recordA, recordB}
	if !reflect.DeepEqual(cancelled, wantCancelled) {
		t.Fatalf("cancelled = %p, want %p", cancelled, wantCancelled)
	}
	if got != 0 || casterReadsA != 2 {
		t.Fatalf("result/caster reads = %d/%d, want 0/2", got, casterReadsA)
	}
	runtime.KeepAlive(recordA)
	runtime.KeepAlive(recordB)
	runtime.KeepAlive(recordC)
	runtime.KeepAlive(caster)
	runtime.KeepAlive(other)
}

func TestSpellCancelDurSpell4FEB10ServerBinding(t *testing.T) {
	caster := new(Object)
	other := new(Object)
	recordA := &DurSpell{Spell: 67, Caster16: caster, Flags88: 0x12345620}
	recordB := &DurSpell{Spell: 75, Caster16: other, Flags88: 0x89abcdee}
	recordC := &DurSpell{Spell: 114, Caster16: caster, Flags88: 0xfedcba40}
	recordA.Next = recordB
	recordB.Next = recordC
	durations := SpellsDuration{List: recordA}
	requireSpellCancelDurSpellNativePointers4FEB10(t,
		unsafe.Pointer(caster), unsafe.Pointer(other),
		unsafe.Pointer(recordA), unsafe.Pointer(recordB), unsafe.Pointer(recordC),
	)

	got := durations.SpellCancelDurSpell4FEB10(75, caster)

	if got != 0 || recordA.Flags88 != 0x12345620 || recordB.Flags88 != 0x89abcdee || recordC.Flags88 != 0xfedcba41 {
		t.Fatalf("result/flags = %d/%#x/%#x/%#x, want 0/0x12345620/0x89abcdee/0xfedcba41",
			got, recordA.Flags88, recordB.Flags88, recordC.Flags88)
	}
	runtime.KeepAlive(caster)
	runtime.KeepAlive(other)
}

func TestSpellCancelDurSpell4FEB10ServerBindingSignedDword(t *testing.T) {
	caster := new(Object)
	record := &DurSpell{Spell: uint32(0x80000000), Caster16: caster, Flags88: 0x76543210}
	durations := SpellsDuration{List: record}

	if got := durations.SpellCancelDurSpell4FEB10(math.MinInt32, caster); got != 0 || record.Flags88 != 0x76543211 {
		t.Fatalf("result/flags = %d/%#x, want 0/0x76543211", got, record.Flags88)
	}
	runtime.KeepAlive(caster)
}

func TestSpellCancelDurSpell4FEB10CancelForBinding(t *testing.T) {
	caster := new(Object)
	record := &DurSpell{Spell: 67, Caster16: caster, Flags88: 0xabcdef20}
	durations := SpellsDuration{List: record}

	durations.CancelFor(spell.ID(67), caster)

	if record.Flags88 != 0xabcdef21 {
		t.Fatalf("flags = %#x, want 0xabcdef21", record.Flags88)
	}
	runtime.KeepAlive(caster)
}

func TestSpellCancelDurSpell4FEB10ServerBindingMatchesNilCasterAndEmpty(t *testing.T) {
	record := &DurSpell{Spell: 67, Flags88: 0x87654320}
	durations := SpellsDuration{List: record}

	if got := durations.SpellCancelDurSpell4FEB10(67, nil); got != 0 || record.Flags88 != 0x87654321 {
		t.Fatalf("result/flags = %d/%#x, want 0/0x87654321", got, record.Flags88)
	}
	var empty SpellsDuration
	if got := empty.SpellCancelDurSpell4FEB10(75, new(Object)); got != 0 {
		t.Fatalf("empty result = %d, want canonical zero", got)
	}
}
