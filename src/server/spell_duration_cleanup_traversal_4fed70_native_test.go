package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"
)

func requireNativeSpellDurationCleanupPointers4FED70(t *testing.T, values ...*DurSpell) {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) != 8 {
		return
	}
	for i, value := range values {
		if uintptr(unsafe.Pointer(value)) <= math.MaxUint32 {
			t.Fatalf("pointer %d = %p, want native address above 4 GiB", i, value)
		}
	}
}

func TestSpellDurationCleanupTraversal4FED70NativeLayout(t *testing.T) {
	wantFlags, wantNext := uintptr(88), uintptr(116)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantFlags, wantNext = 120, 176
	}
	if got := unsafe.Offsetof(DurSpell{}.Flags88); got != wantFlags {
		t.Fatalf("DurSpell.Flags88 offset = %d, want %d", got, wantFlags)
	}
	if got := unsafe.Offsetof(DurSpell{}.Next); got != wantNext {
		t.Fatalf("DurSpell.Next offset = %d, want %d", got, wantNext)
	}
}

func TestSpellDurationCleanupTraversal4FED70NativeListAndSavedSuccessor(t *testing.T) {
	recordA := &DurSpell{Flags88: 0x12345601}
	recordB := &DurSpell{Flags88: 0x89abcdee}
	recordC := &DurSpell{Flags88: 0xfedcba81}
	recordA.Next = recordB
	recordB.Prev = recordA
	recordB.Next = recordC
	recordC.Prev = recordB
	requireNativeSpellDurationCleanupPointers4FED70(t, recordA, recordB, recordC)

	spells := SpellsDuration{List: recordA}
	var destroyed []*DurSpell
	spells.SpellDurationCleanupTraversal4FED70(func(record *DurSpell) {
		destroyed = append(destroyed, record)
		spells.SpellDurationUnlink4FE900(record)
		record.Prev = nil
		record.Next = nil
	})

	if !reflect.DeepEqual(destroyed, []*DurSpell{recordA, recordC}) {
		t.Fatalf("destroyed = %p, want [%p %p]", destroyed, recordA, recordC)
	}
	if spells.List != recordB || recordB.Prev != nil || recordB.Next != nil {
		t.Fatalf("surviving list = head %p Prev %p Next %p, want %p/nil/nil",
			spells.List, recordB.Prev, recordB.Next, recordB)
	}
	if recordA.Flags88 != 0x12345601 || recordB.Flags88 != 0x89abcdee || recordC.Flags88 != 0xfedcba81 {
		t.Fatalf("flags changed = %#x/%#x/%#x", recordA.Flags88, recordB.Flags88, recordC.Flags88)
	}
	runtime.KeepAlive(recordA)
	runtime.KeepAlive(recordB)
	runtime.KeepAlive(recordC)
}

func TestSpellDurationCleanupTraversal4FED70NativeNilReceiverFaultsAtHead(t *testing.T) {
	var spells *SpellsDuration
	called := false
	defer func() {
		if recover() == nil {
			t.Fatal("nil SpellsDuration receiver did not fault")
		}
		if called {
			t.Fatal("destroy callback ran before the nil receiver head fault")
		}
	}()
	spells.SpellDurationCleanupTraversal4FED70(func(*DurSpell) {
		called = true
	})
}
