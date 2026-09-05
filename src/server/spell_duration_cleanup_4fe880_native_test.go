package server

import (
	"reflect"
	"testing"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestSpellDurationCleanupNative4FE880PreservesAllocatorIdentityAndOrder(t *testing.T) {
	class := new(alloc.Class)
	allocator := alloc.ClassT[DurSpell]{Class: class}
	var events []string

	spellDurationCleanupNative4FE880(spellDurationCleanupNativeDeps4FE880{
		loadAllocator: func() alloc.ClassT[DurSpell] {
			events = append(events, "load-allocator")
			return allocator
		},
		freeAllocator: func(value alloc.ClassT[DurSpell]) {
			events = append(events, "free-allocator")
			if value.Class != class {
				t.Fatalf("allocator pointer = %p, want exact %p", value.Class, class)
			}
		},
		clearList: func() {
			events = append(events, "clear-list")
		},
	})

	want := []string{"load-allocator", "free-allocator", "clear-list"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestSpellFreeDurations4FE880NativeStateEffects(t *testing.T) {
	var spells SpellsDuration
	if got := spells.SpellCreateDurations4FE850(); got != 1 {
		t.Fatalf("allocator result = %d, want canonical 1", got)
	}
	record := spells.NewRaw()
	if record == nil {
		t.Fatal("native allocation returned nil")
	}
	spells.List = record
	class := spells.alloc.Class

	spells.SpellFreeDurations4FE880()

	if spells.alloc.Class != class {
		t.Fatalf("allocator handle = %p, want preserved stale handle %p", spells.alloc.Class, class)
	}
	if spells.List != nil {
		t.Fatalf("duration list = %p, want nil", spells.List)
	}
	if spells.lastID != 1 {
		t.Fatalf("last duration ID = %d, want preserved 1", spells.lastID)
	}
}

func TestSpellFreeDurations4FE880NativeNilAllocator(t *testing.T) {
	spells := SpellsDuration{
		List:   &DurSpell{ID: 0x1234},
		lastID: 0xabcd,
	}
	spells.SpellFreeDurations4FE880()
	if spells.alloc.Class != nil || spells.List != nil || spells.lastID != 0xabcd {
		t.Fatalf(
			"allocator/list/lastID = (%p, %p, %#x), want (nil, nil, %#x)",
			spells.alloc.Class, spells.List, spells.lastID, uint16(0xabcd),
		)
	}
}
