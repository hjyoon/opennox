package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"
)

func requireNativeSpellDurationFirstPointer4FE930(t *testing.T, value *DurSpell) {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(value)) <= math.MaxUint32 {
		t.Fatalf("pointer = %p, want native address above 4 GiB", value)
	}
}

func TestSpellDurationFirstNative4FE930PreservesPointerAndSingleLoad(t *testing.T) {
	head := new(DurSpell)
	replacement := new(DurSpell)
	requireNativeSpellDurationFirstPointer4FE930(t, head)
	requireNativeSpellDurationFirstPointer4FE930(t, replacement)
	live := head
	loads := 0

	got := spellDurationFirstNative4FE930(spellDurationFirstNativeDeps4FE930{
		loadHead: func() *DurSpell {
			loads++
			value := live
			live = replacement
			return value
		},
	})

	if got != head {
		t.Fatalf("result = %p, want original head %p", got, head)
	}
	if loads != 1 {
		t.Fatalf("head loads = %d, want 1", loads)
	}
	if live != replacement {
		t.Fatalf("live head = %p, want replacement %p", live, replacement)
	}
	runtime.KeepAlive(head)
	runtime.KeepAlive(replacement)
}

func TestSpellDurationFirst4FE930NativeListState(t *testing.T) {
	var spells SpellsDuration
	first := new(DurSpell)
	second := new(DurSpell)

	spells.List = first
	if got := spells.SpellDurationFirst4FE930(); got != first {
		t.Fatalf("first result = %p, want %p", got, first)
	}
	spells.List = second
	if got := spells.SpellDurationFirst4FE930(); got != second {
		t.Fatalf("second result = %p, want live replacement %p", got, second)
	}
	spells.List = nil
	if got := spells.SpellDurationFirst4FE930(); got != nil {
		t.Fatalf("nil-list result = %p, want nil", got)
	}
}

func TestSpellDurationFirst4FE930NativeNilReceiverFaults(t *testing.T) {
	var spells *SpellsDuration
	defer func() {
		if recover() == nil {
			t.Fatal("nil SpellsDuration receiver did not fault")
		}
	}()
	spells.SpellDurationFirst4FE930()
}
