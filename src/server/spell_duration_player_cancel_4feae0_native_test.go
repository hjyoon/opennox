package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"
)

func requirePlayerCancelNativePointers4FEAE0(t *testing.T, values ...unsafe.Pointer) {
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

func TestPlayerCancelSpells4FEAE0NativeLayouts(t *testing.T) {
	wantDurSize := uintptr(120)
	wantCaster := uintptr(16)
	wantNext := uintptr(116)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantDurSize = 184
		wantCaster = 24
		wantNext = 176
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"DurSpell size", unsafe.Sizeof(DurSpell{}), wantDurSize},
		{"DurSpell.Caster16", unsafe.Offsetof(DurSpell{}.Caster16), wantCaster},
		{"DurSpell.Next", unsafe.Offsetof(DurSpell{}.Next), wantNext},
		{"Caster16 width", unsafe.Sizeof(DurSpell{}.Caster16), unsafe.Sizeof(uintptr(0))},
		{"Next width", unsafe.Sizeof(DurSpell{}.Next), unsafe.Sizeof(uintptr(0))},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestPlayerCancelSpellsNative4FEAE0PreservesPointersAndOrder(t *testing.T) {
	recordA := new(DurSpell)
	recordB := new(DurSpell)
	recordC := new(DurSpell)
	caster := new(Object)
	other := new(Object)
	recordA.Caster16 = caster
	recordA.Next = recordB
	recordB.Caster16 = other
	recordB.Next = recordC
	recordC.Caster16 = caster
	requirePlayerCancelNativePointers4FEAE0(t,
		unsafe.Pointer(recordA), unsafe.Pointer(recordB), unsafe.Pointer(recordC),
		unsafe.Pointer(caster), unsafe.Pointer(other),
	)

	var events []string
	var cancelled []*DurSpell
	got := playerCancelSpellsNative4FEAE0(caster, playerCancelSpellsNativeDeps4FEAE0{
		loadFirst: func() *DurSpell {
			events = append(events, "first")
			return recordA
		},
		loadCaster: func(record *DurSpell) *Object {
			events = append(events, "caster")
			if record == recordA {
				record.Next = recordC
			}
			return record.Caster16
		},
		loadNext: func(record *DurSpell) *DurSpell {
			events = append(events, "next")
			return record.Next
		},
		cancel: func(record *DurSpell) {
			events = append(events, "cancel")
			cancelled = append(cancelled, record)
			if record == recordA {
				record.Next = recordB
			}
		},
	})

	wantEvents := []string{"first", "caster", "next", "cancel", "caster", "next", "cancel"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %q, want %q", events, wantEvents)
	}
	wantCancelled := []*DurSpell{recordA, recordC}
	if !reflect.DeepEqual(cancelled, wantCancelled) {
		t.Fatalf("cancelled = %p, want %p", cancelled, wantCancelled)
	}
	if got != 0 {
		t.Fatalf("result = %d, want canonical zero", got)
	}
	runtime.KeepAlive(recordA)
	runtime.KeepAlive(recordB)
	runtime.KeepAlive(recordC)
	runtime.KeepAlive(caster)
	runtime.KeepAlive(other)
}

func TestPlayerCancelSpells4FEAE0ServerBinding(t *testing.T) {
	caster := new(Object)
	other := new(Object)
	recordA := &DurSpell{Caster16: other, Flags88: 0x12345620}
	recordB := &DurSpell{Caster16: caster, Flags88: 0x89abcdee}
	recordC := &DurSpell{Caster16: caster, Flags88: 0xfedcba40}
	recordA.Next = recordB
	recordB.Next = recordC
	durations := SpellsDuration{List: recordA}
	requirePlayerCancelNativePointers4FEAE0(t,
		unsafe.Pointer(caster), unsafe.Pointer(other),
		unsafe.Pointer(recordA), unsafe.Pointer(recordB), unsafe.Pointer(recordC),
	)

	got := durations.PlayerCancelSpells4FEAE0(caster)

	if got != 0 || recordA.Flags88 != 0x12345620 || recordB.Flags88 != 0x89abcdef || recordC.Flags88 != 0xfedcba41 {
		t.Fatalf("result/flags = %d/%#x/%#x/%#x, want 0/0x12345620/0x89abcdef/0xfedcba41",
			got, recordA.Flags88, recordB.Flags88, recordC.Flags88)
	}
	runtime.KeepAlive(caster)
	runtime.KeepAlive(other)
}

func TestPlayerCancelSpells4FEAE0ServerBindingMatchesNilCaster(t *testing.T) {
	record := &DurSpell{Flags88: 0x87654320}
	durations := SpellsDuration{List: record}

	got := durations.PlayerCancelSpells4FEAE0(nil)

	if got != 0 || record.Flags88 != 0x87654321 {
		t.Fatalf("result/flags = %d/%#x, want 0/0x87654321", got, record.Flags88)
	}
}

func TestPlayerCancelSpells4FEAE0ServerBindingEmpty(t *testing.T) {
	var durations SpellsDuration
	if got := durations.PlayerCancelSpells4FEAE0(new(Object)); got != 0 {
		t.Fatalf("empty result = %d, want canonical zero", got)
	}
}
