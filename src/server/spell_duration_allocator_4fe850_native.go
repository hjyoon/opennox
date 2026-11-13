package server

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type spellDurationAllocatorNativeDeps4FE850 struct {
	newClass       func(name string, recordSize uintptr, capacity int) alloc.ClassT[DurSpell]
	storeAllocator func(alloc.ClassT[DurSpell])
}

func spellDurationAllocatorNative4FE850(deps spellDurationAllocatorNativeDeps4FE850) int32 {
	return SpellDurationAllocator4FE850(
		unsafe.Sizeof(DurSpell{}),
		SpellDurationAllocatorHooks4FE850[alloc.ClassT[DurSpell]]{
			NewClass: deps.newClass,
			NonZero: func(value alloc.ClassT[DurSpell]) bool {
				return value.Class != nil
			},
			StoreAllocator: deps.storeAllocator,
		},
	)
}

// SpellCreateDurations4FE850 binds GAME.EXE 004FE850 to a native-width
// DurSpell allocation class. The original PE32 record is 120 bytes; pointer
// fields widen the record on 64-bit targets while scalar fields retain their
// original widths.
func (sp *SpellsDuration) SpellCreateDurations4FE850() int32 {
	return sp.spellCreateDurations4FE850(sp.s != nil && sp.s.highAddresses)
}

// SpellCreateDurationsHigh4FE850 explicitly selects high foreign records for
// this allocation class without changing any other duration-spell owner.
func (sp *SpellsDuration) SpellCreateDurationsHigh4FE850() int32 {
	return sp.spellCreateDurations4FE850(true)
}

func (sp *SpellsDuration) spellCreateDurations4FE850(highAddresses bool) int32 {
	return spellDurationAllocatorNative4FE850(spellDurationAllocatorNativeDeps4FE850{
		newClass: func(name string, recordSize uintptr, capacity int) alloc.ClassT[DurSpell] {
			if highAddresses {
				return alloc.ClassT[DurSpell]{Class: alloc.NewClassHigh(name, recordSize, capacity)}
			}
			return alloc.ClassT[DurSpell]{Class: alloc.NewClass(name, recordSize, capacity)}
		},
		storeAllocator: func(value alloc.ClassT[DurSpell]) {
			sp.alloc = value
		},
	})
}
