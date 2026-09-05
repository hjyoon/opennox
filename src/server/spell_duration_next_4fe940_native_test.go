package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"
)

func requireNativeSpellDurationNextPointer4FE940(t *testing.T, value *DurSpell) {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(value)) <= math.MaxUint32 {
		t.Fatalf("pointer = %p, want native address above 4 GiB", value)
	}
}

func TestSpellDurationNextNative4FE940PreservesPointerAndSingleLoad(t *testing.T) {
	record := new(DurSpell)
	next := new(DurSpell)
	replacement := new(DurSpell)
	requireNativeSpellDurationNextPointer4FE940(t, record)
	requireNativeSpellDurationNextPointer4FE940(t, next)
	requireNativeSpellDurationNextPointer4FE940(t, replacement)
	live := next
	loads := 0

	got := spellDurationNextNative4FE940(record, spellDurationNextNativeDeps4FE940{
		loadNext: func(gotRecord *DurSpell) *DurSpell {
			loads++
			if gotRecord != record {
				t.Fatalf("record = %p, want %p", gotRecord, record)
			}
			value := live
			live = replacement
			return value
		},
	})

	if got != next {
		t.Fatalf("result = %p, want original next %p", got, next)
	}
	if loads != 1 {
		t.Fatalf("next loads = %d, want 1", loads)
	}
	if live != replacement {
		t.Fatalf("live next = %p, want replacement %p", live, replacement)
	}
	runtime.KeepAlive(record)
	runtime.KeepAlive(next)
	runtime.KeepAlive(replacement)
}

func TestSpellDurationNextNative4FE940UsesLiveRecordLink(t *testing.T) {
	record := new(DurSpell)
	first := new(DurSpell)
	second := new(DurSpell)

	record.Next = first
	if got := SpellDurationNextNative4FE940(record); got != first {
		t.Fatalf("first result = %p, want %p", got, first)
	}
	record.Next = second
	if got := SpellDurationNextNative4FE940(record); got != second {
		t.Fatalf("second result = %p, want live replacement %p", got, second)
	}
	record.Next = nil
	if got := SpellDurationNextNative4FE940(record); got != nil {
		t.Fatalf("nil-link result = %p, want nil", got)
	}
}

func TestSpellDurationNextNative4FE940NilRecordFaults(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil DurSpell record did not fault")
		}
	}()
	SpellDurationNextNative4FE940(nil)
}
