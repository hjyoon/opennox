package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"
)

func requireNativeSpellDurationInsertPointers4FED40(t *testing.T, values ...*DurSpell) {
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

func TestSpellDurationInsert4FED40NativeLinkLayout(t *testing.T) {
	wantPrev, wantNext := uintptr(112), uintptr(116)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantPrev, wantNext = 168, 176
	}
	if got := unsafe.Offsetof(DurSpell{}.Prev); got != wantPrev {
		t.Fatalf("DurSpell.Prev offset = %d, want %d", got, wantPrev)
	}
	if got := unsafe.Offsetof(DurSpell{}.Next); got != wantNext {
		t.Fatalf("DurSpell.Next offset = %d, want %d", got, wantNext)
	}
}

func TestSpellDurationInsertNative4FED40PreservesPointersAndReloadsHead(t *testing.T) {
	record := new(DurSpell)
	headA := new(DurSpell)
	headB := new(DurSpell)
	previous := new(DurSpell)
	record.Prev = previous
	record.Next = previous
	requireNativeSpellDurationInsertPointers4FED40(t, record, headA, headB, previous)

	head := headA
	var events []string
	spellDurationInsertNative4FED40(record, spellDurationInsertNativeDeps4FED40{
		loadHead: func() *DurSpell {
			events = append(events, "head")
			return head
		},
		storePrev: func(value, prev *DurSpell) {
			if value == headA {
				events = append(events, "old-prev")
				if prev != record {
					t.Fatalf("old head Prev value = %p, want record %p", prev, record)
				}
				value.Prev = prev
				head = headB
				return
			}
			events = append(events, "record-prev")
			if value != record || prev != nil {
				t.Fatalf("record Prev store = %p/%p, want %p/nil", value, prev, record)
			}
			value.Prev = prev
		},
		storeNext: func(value, next *DurSpell) {
			events = append(events, "record-next")
			if value != record || next != headB {
				t.Fatalf("record Next store = %p/%p, want %p/live %p", value, next, record, headB)
			}
			value.Next = next
		},
		storeHead: func(value *DurSpell) {
			events = append(events, "store-head")
			if value != record {
				t.Fatalf("new head = %p, want record %p", value, record)
			}
			head = value
		},
	})

	want := []string{"head", "old-prev", "record-prev", "head", "record-next", "store-head"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
	if headA.Prev != record || record.Prev != nil || record.Next != headB || head != record {
		t.Fatalf("native links = headA.Prev %p record.Prev %p record.Next %p head %p",
			headA.Prev, record.Prev, record.Next, head)
	}
	if headB.Prev != nil {
		t.Fatalf("live replacement head Prev = %p, want untouched nil", headB.Prev)
	}
	runtime.KeepAlive(record)
	runtime.KeepAlive(headA)
	runtime.KeepAlive(headB)
	runtime.KeepAlive(previous)
}

func TestSpellDurationInsert4FED40NativeListState(t *testing.T) {
	t.Run("non-empty", func(t *testing.T) {
		var spells SpellsDuration
		record := new(DurSpell)
		head := new(DurSpell)
		previous := new(DurSpell)
		record.Prev = previous
		record.Next = previous
		spells.List = head

		spells.SpellDurationInsert4FED40(record)

		if spells.List != record || record.Prev != nil || record.Next != head || head.Prev != record {
			t.Fatalf("links = head %p record.Prev %p record.Next %p old.Prev %p, want %p/nil/%p/%p",
				spells.List, record.Prev, record.Next, head.Prev, record, head, record)
		}
	})

	t.Run("empty", func(t *testing.T) {
		var spells SpellsDuration
		record := &DurSpell{Prev: new(DurSpell), Next: new(DurSpell)}
		spells.SpellDurationInsert4FED40(record)
		if spells.List != record || record.Prev != nil || record.Next != nil {
			t.Fatalf("links = head %p Prev %p Next %p, want %p/nil/nil",
				spells.List, record.Prev, record.Next, record)
		}
	})
}

func TestSpellDurationInsert4FED40NativeDoesNotGuardNilRecord(t *testing.T) {
	var spells SpellsDuration
	head := &DurSpell{Prev: new(DurSpell)}
	spells.List = head

	var recovered any
	func() {
		defer func() { recovered = recover() }()
		spells.SpellDurationInsert4FED40(nil)
	}()
	if recovered == nil {
		t.Fatal("nil duration record did not fault")
	}
	if head.Prev != nil {
		t.Fatalf("old head Prev = %p, want nil record stored before fault", head.Prev)
	}
	if spells.List != head {
		t.Fatalf("head = %p, want original %p after fault", spells.List, head)
	}
}

func TestSpellDurationInsert4FED40NativeNilReceiverFaultsBeforeRecordStore(t *testing.T) {
	var spells *SpellsDuration
	record := &DurSpell{Prev: new(DurSpell), Next: new(DurSpell)}
	before := *record
	defer func() {
		if recover() == nil {
			t.Fatal("nil SpellsDuration receiver did not fault")
		}
		if *record != before {
			t.Fatalf("record changed before nil receiver head fault: got %#v want %#v", *record, before)
		}
	}()
	spells.SpellDurationInsert4FED40(record)
}
