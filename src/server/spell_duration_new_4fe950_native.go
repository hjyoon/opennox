package server

import "github.com/opennox/opennox/v1/legacy/common/alloc"

type spellDurationNewNativeDeps4FE950 struct {
	loadAllocator func() alloc.ClassT[DurSpell]
	newObject     func(alloc.ClassT[DurSpell]) *DurSpell
	loadLastID    func() uint16
	storeLastID   func(uint16)
	storeRecordID func(*DurSpell, uint16)
}

func spellDurationNewNative4FE950(deps spellDurationNewNativeDeps4FE950) *DurSpell {
	return SpellDurationNew4FE950(SpellDurationNewHooks4FE950[alloc.ClassT[DurSpell], *DurSpell]{
		LoadAllocator: deps.loadAllocator,
		NewObject:     deps.newObject,
		LoadLastID:    deps.loadLastID,
		StoreLastID:   deps.storeLastID,
		StoreRecordID: deps.storeRecordID,
	})
}

// SpellDurationNew4FE950 binds GAME.EXE 004FE950 to the native-width
// duration-spell allocation class and record. The allocator supplies a zeroed
// DurSpell whose pointer fields widen naturally on 64-bit targets, while the
// global and record identifiers remain exact uint16 values.
//
//go:noinline
func (sp *SpellsDuration) SpellDurationNew4FE950() *DurSpell {
	return spellDurationNewNative4FE950(spellDurationNewNativeDeps4FE950{
		loadAllocator: func() alloc.ClassT[DurSpell] {
			return sp.alloc
		},
		newObject: func(allocator alloc.ClassT[DurSpell]) *DurSpell {
			return allocator.NewObject()
		},
		loadLastID: func() uint16 {
			return sp.lastID
		},
		storeLastID: func(id uint16) {
			sp.lastID = id
		},
		storeRecordID: func(record *DurSpell, id uint16) {
			record.ID = id
		},
	})
}
