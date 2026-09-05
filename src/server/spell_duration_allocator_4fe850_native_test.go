package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestSpellDurationAllocatorNative4FE850LayoutResultAndStateEffects(t *testing.T) {
	wantRecordSize := uintptr(120)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantRecordSize = 184
	}
	if got := unsafe.Sizeof(DurSpell{}); got != wantRecordSize {
		t.Fatalf("DurSpell size = %d, want native size %d", got, wantRecordSize)
	}

	oldClass := new(alloc.Class)
	newClass := new(alloc.Class)
	list := &DurSpell{ID: 0x1234}
	tests := []struct {
		name      string
		allocator *alloc.Class
		want      int32
	}{
		{name: "success", allocator: newClass, want: 1},
		{name: "failure", allocator: nil, want: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spells := SpellsDuration{
				alloc:  alloc.ClassT[DurSpell]{Class: oldClass},
				List:   list,
				lastID: 0xabcd,
			}
			got := spellDurationAllocatorNative4FE850(spellDurationAllocatorNativeDeps4FE850{
				newClass: func(name string, recordSize uintptr, capacity int) alloc.ClassT[DurSpell] {
					if name != "spellDuration" || recordSize != wantRecordSize || capacity != 512 {
						t.Fatalf(
							"allocation request = (%q, %d, %d), want (spellDuration, %d, 512)",
							name, recordSize, capacity, wantRecordSize,
						)
					}
					return alloc.ClassT[DurSpell]{Class: tc.allocator}
				},
				storeAllocator: func(value alloc.ClassT[DurSpell]) {
					spells.alloc = value
				},
			})

			if got != tc.want {
				t.Fatalf("result = %d, want %d", got, tc.want)
			}
			if spells.alloc.Class != tc.allocator {
				t.Fatalf("stored allocator = %p, want %p", spells.alloc.Class, tc.allocator)
			}
			if spells.List != list || spells.lastID != 0xabcd {
				t.Fatalf(
					"list/lastID = %p/%#x, want preserved %p/%#x",
					spells.List, spells.lastID, list, uint16(0xabcd),
				)
			}
		})
	}
}

func TestSpellCreateDurations4FE850AllocatesNativeRecords(t *testing.T) {
	var spells SpellsDuration
	if got := spells.SpellCreateDurations4FE850(); got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}
	defer spells.Free()

	first := spells.NewRaw()
	second := spells.NewRaw()
	if first == nil || second == nil {
		t.Fatalf("allocated records = (%p, %p), want two non-nil records", first, second)
	}
	if first.ID != 1 || second.ID != 2 {
		t.Fatalf("allocated IDs = (%d, %d), want (1, 2)", first.ID, second.ID)
	}
}
