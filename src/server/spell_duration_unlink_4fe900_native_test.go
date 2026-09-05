package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"
)

func requireNativeSpellDurationPointers4FE900(t *testing.T, values ...*DurSpell) {
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

func TestSpellDurationUnlinkNative4FE900PreservesNativePointersAndReloads(t *testing.T) {
	record := new(DurSpell)
	prevA := new(DurSpell)
	nextA := new(DurSpell)
	prevB := new(DurSpell)
	nextB := new(DurSpell)
	record.Prev = prevA
	record.Next = nextA
	requireNativeSpellDurationPointers4FE900(t, record, prevA, nextA, prevB, nextB)

	var events []string
	spellDurationUnlinkNative4FE900(record, spellDurationUnlinkNativeDeps4FE900{
		loadPrev: func(value *DurSpell) *DurSpell {
			events = append(events, "prev")
			return value.Prev
		},
		loadNext: func(value *DurSpell) *DurSpell {
			events = append(events, "next")
			result := value.Next
			if len(events) == 4 {
				value.Prev = prevB
			}
			return result
		},
		storeNext: func(value, next *DurSpell) {
			events = append(events, "store-next")
			if value != prevA || next != nextA {
				t.Fatalf("first store = %p/%p, want %p/%p", value, next, prevA, nextA)
			}
			value.Next = next
			record.Next = nextB
		},
		storeHead: func(*DurSpell) {
			t.Fatal("interior record must not update head")
		},
		storePrev: func(value, prev *DurSpell) {
			events = append(events, "store-prev")
			if value != nextB || prev != prevB {
				t.Fatalf("second store = %p/%p, want live %p/%p", value, prev, nextB, prevB)
			}
			value.Prev = prev
		},
	})

	want := []string{"prev", "next", "store-next", "next", "prev", "store-prev"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
	if prevA.Next != nextA || nextB.Prev != prevB {
		t.Fatalf("native links = prevA.Next %p nextB.Prev %p", prevA.Next, nextB.Prev)
	}
	runtime.KeepAlive(record)
	runtime.KeepAlive(prevA)
	runtime.KeepAlive(nextA)
	runtime.KeepAlive(prevB)
	runtime.KeepAlive(nextB)
}

func TestSpellDurationUnlink4FE900NativeListState(t *testing.T) {
	t.Run("head", func(t *testing.T) {
		var spells SpellsDuration
		record := new(DurSpell)
		next := new(DurSpell)
		record.Next = next
		next.Prev = record
		spells.List = record

		spells.SpellDurationUnlink4FE900(record)

		if spells.List != next || next.Prev != nil {
			t.Fatalf("head/next.Prev = %p/%p, want %p/nil", spells.List, next.Prev, next)
		}
		if record.Prev != nil || record.Next != next {
			t.Fatalf("detached record links = %p/%p, want nil/%p unchanged", record.Prev, record.Next, next)
		}
	})

	t.Run("interior", func(t *testing.T) {
		var spells SpellsDuration
		prev := new(DurSpell)
		record := new(DurSpell)
		next := new(DurSpell)
		prev.Next = record
		record.Prev = prev
		record.Next = next
		next.Prev = record
		spells.List = prev

		spells.SpellDurationUnlink4FE900(record)

		if spells.List != prev || prev.Next != next || next.Prev != prev {
			t.Fatalf("links = head %p prev.Next %p next.Prev %p, want %p/%p/%p", spells.List, prev.Next, next.Prev, prev, next, prev)
		}
		if record.Prev != prev || record.Next != next {
			t.Fatalf("detached record links = %p/%p, want %p/%p unchanged", record.Prev, record.Next, prev, next)
		}
	})

	t.Run("singleton", func(t *testing.T) {
		var spells SpellsDuration
		record := new(DurSpell)
		spells.List = record
		spells.SpellDurationUnlink4FE900(record)
		if spells.List != nil {
			t.Fatalf("head = %p, want nil", spells.List)
		}
	})
}

func TestSpellDurationUnlink4FE900NativeDoesNotGuardNilRecord(t *testing.T) {
	var spells SpellsDuration
	defer func() {
		if recover() == nil {
			t.Fatal("nil duration record did not fault")
		}
	}()
	spells.SpellDurationUnlink4FE900(nil)
}
