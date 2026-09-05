package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func requireNativeSpellDurationNewPointer4FE950(t *testing.T, name string, value unsafe.Pointer) {
	t.Helper()
	if value == nil {
		t.Fatalf("%s pointer is nil", name)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(value) <= math.MaxUint32 {
		t.Fatalf("%s pointer = %p, want native address above 4 GiB", name, value)
	}
}

func TestSpellDurationNewNative4FE950PreservesOrderPointersAndWrap(t *testing.T) {
	originalClass := new(alloc.Class)
	replacementClass := new(alloc.Class)
	record, freeRecord := alloc.New(DurSpell{})
	t.Cleanup(freeRecord)
	requireNativeSpellDurationNewPointer4FE950(t, "original allocator", unsafe.Pointer(originalClass))
	requireNativeSpellDurationNewPointer4FE950(t, "replacement allocator", unsafe.Pointer(replacementClass))
	requireNativeSpellDurationNewPointer4FE950(t, "record", unsafe.Pointer(record))

	liveAllocator := alloc.ClassT[DurSpell]{Class: originalClass}
	lastID := uint16(0xffff)
	var events []string
	got := spellDurationNewNative4FE950(spellDurationNewNativeDeps4FE950{
		loadAllocator: func() alloc.ClassT[DurSpell] {
			events = append(events, "load-allocator")
			return liveAllocator
		},
		newObject: func(allocator alloc.ClassT[DurSpell]) *DurSpell {
			events = append(events, "new-object")
			if allocator.Class != originalClass {
				t.Fatalf("allocator = %p, want snapshot %p", allocator.Class, originalClass)
			}
			liveAllocator.Class = replacementClass
			return record
		},
		loadLastID: func() uint16 {
			events = append(events, "load-last-id")
			if liveAllocator.Class != replacementClass {
				t.Fatalf("live allocator = %p, want replacement %p", liveAllocator.Class, replacementClass)
			}
			return lastID
		},
		storeLastID: func(id uint16) {
			events = append(events, "store-last-id")
			lastID = id
		},
		storeRecordID: func(gotRecord *DurSpell, id uint16) {
			events = append(events, "store-record-id")
			if gotRecord != record {
				t.Fatalf("record = %p, want %p", gotRecord, record)
			}
			gotRecord.ID = id
		},
	})

	if got != record {
		t.Fatalf("result = %p, want exact record %p", got, record)
	}
	if lastID != 0 || record.ID != 0 {
		t.Fatalf("last/record ID = (%#x, %#x), want uint16 wrap to zero", lastID, record.ID)
	}
	want := []string{"load-allocator", "new-object", "load-last-id", "store-last-id", "store-record-id"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
	runtime.KeepAlive(originalClass)
	runtime.KeepAlive(replacementClass)
	runtime.KeepAlive(record)
}

func TestSpellDurationNewNative4FE950AllocationFailurePreservesID(t *testing.T) {
	spells := SpellsDuration{lastID: 0xabcd}
	if got := spells.SpellDurationNew4FE950(); got != nil {
		t.Fatalf("result = %p, want nil without an allocator", got)
	}
	if spells.lastID != 0xabcd {
		t.Fatalf("last ID = %#x, want preserved 0xabcd", spells.lastID)
	}
}

func TestSpellDurationNewNative4FE950ReusesZeroedNativeRecord(t *testing.T) {
	var spells SpellsDuration
	if got := spells.SpellCreateDurations4FE850(); got != 1 {
		t.Fatalf("allocator result = %d, want canonical 1", got)
	}
	t.Cleanup(spells.Free)

	record := spells.SpellDurationNew4FE950()
	if record == nil {
		t.Fatal("first allocation returned nil")
	}
	requireNativeSpellDurationNewPointer4FE950(t, "record", unsafe.Pointer(record))
	record.Spell = 0xaaaaaaaa
	record.Caster16 = &Object{}
	record.Field76 = ^uintptr(0)
	record.Next = &DurSpell{ID: 0xbbbb}
	spells.alloc.FreeObjectFirst(record)

	spells.lastID = 0xffff
	reused := spells.NewRaw()
	if reused != record {
		t.Fatalf("reused record = %p, want same first-freed record %p", reused, record)
	}
	if got, want := *reused, (DurSpell{ID: 0}); got != want {
		t.Fatalf("reused record = %#v, want allocator-zeroed record %#v", got, want)
	}
	if unsafe.Offsetof(DurSpell{}.ID) != 0 {
		t.Fatalf("DurSpell.ID offset = %d, want first field at 0", unsafe.Offsetof(DurSpell{}.ID))
	}

	next := spells.NewRaw()
	if next == nil {
		t.Fatal("next allocation returned nil")
	}
	if next.ID != 1 || spells.lastID != 1 {
		t.Fatalf("next record/last ID = (%p/%#x, %#x), want non-nil/1 and 1", next, next.ID, spells.lastID)
	}
}
