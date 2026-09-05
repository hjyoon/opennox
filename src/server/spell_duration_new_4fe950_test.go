package server

import (
	"fmt"
	"reflect"
	"testing"
)

func TestSpellDurationNew4FE950SuccessOrderSnapshotsAndWrap(t *testing.T) {
	const (
		originalAllocator    = uint64(0x100000101)
		replacementAllocator = uint64(0x200000202)
		record               = uint64(0x300000303)
	)
	liveAllocator := originalAllocator
	lastID := uint16(0xffff)
	storedRecord := uint64(0)
	storedRecordID := uint16(0xaaaa)
	var events []string

	got := SpellDurationNew4FE950(SpellDurationNewHooks4FE950[uint64, uint64]{
		LoadAllocator: func() uint64 {
			events = append(events, "load-allocator")
			return liveAllocator
		},
		NewObject: func(allocator uint64) uint64 {
			events = append(events, "new-object")
			if allocator != originalAllocator {
				t.Fatalf("allocator = %#x, want snapshot %#x", allocator, originalAllocator)
			}
			liveAllocator = replacementAllocator
			return record
		},
		LoadLastID: func() uint16 {
			events = append(events, "load-last-id")
			if liveAllocator != replacementAllocator {
				t.Fatalf("live allocator = %#x, want callback replacement %#x", liveAllocator, replacementAllocator)
			}
			return lastID
		},
		StoreLastID: func(id uint16) {
			events = append(events, "store-last-id")
			lastID = id
		},
		StoreRecordID: func(gotRecord uint64, id uint16) {
			events = append(events, "store-record-id")
			storedRecord = gotRecord
			storedRecordID = id
		},
	})

	if got != record {
		t.Fatalf("result = %#x, want exact record %#x", got, record)
	}
	if lastID != 0 {
		t.Fatalf("last ID = %#x, want uint16 wrap to zero", lastID)
	}
	if storedRecord != record || storedRecordID != 0 {
		t.Fatalf("record/ID = (%#x, %#x), want (%#x, 0)", storedRecord, storedRecordID, record)
	}
	want := []string{"load-allocator", "new-object", "load-last-id", "store-last-id", "store-record-id"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestSpellDurationNew4FE950AllocationFailurePreservesID(t *testing.T) {
	const allocator = uint64(0x100000101)
	lastID := uint16(0xabcd)
	var events []string

	got := SpellDurationNew4FE950(SpellDurationNewHooks4FE950[uint64, uint64]{
		LoadAllocator: func() uint64 {
			events = append(events, "load-allocator")
			return allocator
		},
		NewObject: func(got uint64) uint64 {
			events = append(events, "new-object")
			if got != allocator {
				t.Fatalf("allocator = %#x, want %#x", got, allocator)
			}
			return 0
		},
		LoadLastID: func() uint16 {
			t.Fatal("last ID loaded after allocation failure")
			return 0
		},
		StoreLastID: func(uint16) {
			t.Fatal("last ID stored after allocation failure")
		},
		StoreRecordID: func(uint64, uint16) {
			t.Fatal("record ID stored after allocation failure")
		},
	})

	if got != 0 {
		t.Fatalf("result = %#x, want nil token", got)
	}
	if lastID != 0xabcd {
		t.Fatalf("last ID = %#x, want preserved 0xabcd", lastID)
	}
	want := []string{"load-allocator", "new-object"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestSpellDurationNew4FE950CachesIncrementedIDAcrossStores(t *testing.T) {
	const record = uint64(0x100000101)
	lastID := uint16(0x1234)
	var recordID uint16

	SpellDurationNew4FE950(SpellDurationNewHooks4FE950[uint64, uint64]{
		LoadAllocator: func() uint64 { return 7 },
		NewObject:     func(uint64) uint64 { return record },
		LoadLastID:    func() uint16 { return lastID },
		StoreLastID: func(id uint16) {
			if id != 0x1235 {
				t.Fatalf("stored last ID = %#x, want 0x1235", id)
			}
			lastID = 0xbeef
		},
		StoreRecordID: func(gotRecord uint64, id uint16) {
			if gotRecord != record {
				t.Fatalf("record = %#x, want %#x", gotRecord, record)
			}
			recordID = id
		},
	})

	if recordID != 0x1235 {
		t.Fatalf("record ID = %#x, want cached incremented ID 0x1235", recordID)
	}
	if lastID != 0xbeef {
		t.Fatalf("instrumented live last ID = %#x, want 0xbeef", lastID)
	}
}

func TestSpellDurationNew4FE950FaultPrefixes(t *testing.T) {
	stop := &struct{}{}
	want := []string{"load-allocator", "new-object", "load-last-id", "store-last-id", "store-record-id"}
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			var events []string
			step := func(event string) {
				events = append(events, event)
				if len(events) == faultAt {
					panic(stop)
				}
			}
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				SpellDurationNew4FE950(SpellDurationNewHooks4FE950[uint64, uint64]{
					LoadAllocator: func() uint64 {
						step("load-allocator")
						return 7
					},
					NewObject: func(uint64) uint64 {
						step("new-object")
						return 9
					},
					LoadLastID: func() uint16 {
						step("load-last-id")
						return 11
					},
					StoreLastID: func(uint16) {
						step("store-last-id")
					},
					StoreRecordID: func(uint64, uint16) {
						step("store-record-id")
					},
				})
			}()
			if recovered != stop {
				t.Fatalf("recovered = %#v, want sentinel", recovered)
			}
			if prefix := want[:faultAt]; !reflect.DeepEqual(events, prefix) {
				t.Fatalf("events = %#v, want fault prefix %#v", events, prefix)
			}
		})
	}
}
