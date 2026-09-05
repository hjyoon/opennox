package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestSpellDurationSelectiveCleanupNative4FE8A0PreservesNativePointers(t *testing.T) {
	recordA := new(DurSpell)
	recordB := new(DurSpell)
	player := &Object{ObjClass: object.ClassPlayer}
	allocatorClass := new(alloc.Class)
	allocator := alloc.ClassT[DurSpell]{Class: allocatorClass}
	recordA.Next = recordB
	recordB.Target48 = player

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"record A":  unsafe.Pointer(recordA),
			"record B":  unsafe.Pointer(recordB),
			"player":    unsafe.Pointer(player),
			"allocator": unsafe.Pointer(allocatorClass),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want native address above 4 GiB", name, pointer)
			}
		}
	}

	var events []string
	spellDurationSelectiveCleanupNative4FE8A0(1, spellDurationSelectiveCleanupNativeDeps4FE8A0{
		loadAllocator: func() alloc.ClassT[DurSpell] {
			events = append(events, "allocator")
			return allocator
		},
		freeAllObjects: func(value alloc.ClassT[DurSpell]) {
			events = append(events, "free-all")
			if value.Class != allocatorClass {
				t.Fatalf("allocator = %p, want %p", value.Class, allocatorClass)
			}
		},
		clearList: func() { events = append(events, "clear-list") },
		loadList: func() *DurSpell {
			events = append(events, "list")
			return recordA
		},
		loadTarget: func(record *DurSpell) *Object {
			events = append(events, "target")
			return record.Target48
		},
		loadNext: func(record *DurSpell) *DurSpell {
			events = append(events, "next")
			next := record.Next
			record.Next = nil
			return next
		},
		loadClassLowByte: func(value *Object) uint8 {
			events = append(events, "class")
			if value != player {
				t.Fatalf("class object = %p, want %p", value, player)
			}
			return uint8(value.ObjClass)
		},
		unlink: func(record *DurSpell) {
			events = append(events, "unlink")
			if record != recordA {
				t.Fatalf("unlinked record = %p, want %p", record, recordA)
			}
		},
		freeRecursive: func(record *DurSpell) {
			events = append(events, "free")
			if record != recordA {
				t.Fatalf("freed record = %p, want %p", record, recordA)
			}
		},
	})

	want := []string{"list", "target", "next", "unlink", "free", "target", "next", "class"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
	runtime.KeepAlive(recordA)
	runtime.KeepAlive(recordB)
	runtime.KeepAlive(player)
	runtime.KeepAlive(allocatorClass)
}

func TestSpellResetDurations4FE8A0NativeSelectiveState(t *testing.T) {
	var spells SpellsDuration
	if got := spells.SpellCreateDurations4FE850(); got != 1 {
		t.Fatalf("allocator result = %d, want canonical 1", got)
	}
	t.Cleanup(spells.SpellFreeDurations4FE880)

	player, freePlayer := alloc.New(Object{})
	monster, freeMonster := alloc.New(Object{})
	t.Cleanup(freePlayer)
	t.Cleanup(freeMonster)
	player.ObjClass = object.ClassPlayer | object.Class(0x80000000)
	monster.ObjClass = object.ClassMonster | object.Class(0x40000000)

	nilTarget := spells.NewRaw()
	keep := spells.NewRaw()
	nonPlayer := spells.NewRaw()
	if nilTarget == nil || keep == nil || nonPlayer == nil {
		t.Fatal("native duration allocation returned nil")
	}
	keep.Target48 = player
	nonPlayer.Target48 = monster
	spells.Add(nilTarget)
	spells.Add(keep)
	spells.Add(nonPlayer)

	spells.SpellResetDurations4FE8A0(1)

	if spells.List != keep || keep.Prev != nil || keep.Next != nil {
		t.Fatalf("kept list = head %p prev %p next %p, want %p/nil/nil", spells.List, keep.Prev, keep.Next, keep)
	}
}

func TestSpellResetDurations4FE8A0NativeZeroModeKeepsAllocatorAndID(t *testing.T) {
	var spells SpellsDuration
	if got := spells.SpellCreateDurations4FE850(); got != 1 {
		t.Fatalf("allocator result = %d, want canonical 1", got)
	}
	t.Cleanup(spells.SpellFreeDurations4FE880)
	record := spells.NewRaw()
	if record == nil {
		t.Fatal("native duration allocation returned nil")
	}
	spells.List = record
	class := spells.alloc.Class

	spells.SpellResetDurations4FE8A0(0)

	if spells.List != nil || spells.alloc.Class != class || spells.lastID != 1 {
		t.Fatalf("list/allocator/lastID = %p/%p/%d, want nil/%p/1", spells.List, spells.alloc.Class, spells.lastID, class)
	}
	next := spells.NewRaw()
	if next == nil || next.ID != 2 {
		t.Fatalf("post-reset allocation = %p ID %d, want nonnil ID 2", next, func() uint16 {
			if next == nil {
				return 0
			}
			return next.ID
		}())
	}
}

func TestSpellResetDurations4FE8A0NativeNilAllocator(t *testing.T) {
	spells := SpellsDuration{List: &DurSpell{ID: 0x1234}, lastID: 0xabcd}
	spells.SpellResetDurations4FE8A0(0)
	if spells.List != nil || spells.alloc.Class != nil || spells.lastID != 0xabcd {
		t.Fatalf("list/allocator/lastID = %p/%p/%#x, want nil/nil/%#x", spells.List, spells.alloc.Class, spells.lastID, uint16(0xabcd))
	}
}
