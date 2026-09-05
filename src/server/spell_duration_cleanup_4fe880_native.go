package server

import "github.com/opennox/opennox/v1/legacy/common/alloc"

type spellDurationCleanupNativeDeps4FE880 struct {
	loadAllocator func() alloc.ClassT[DurSpell]
	freeAllocator func(alloc.ClassT[DurSpell])
	clearList     func()
}

func spellDurationCleanupNative4FE880(deps spellDurationCleanupNativeDeps4FE880) {
	SpellDurationCleanup4FE880(SpellDurationCleanupHooks4FE880[alloc.ClassT[DurSpell]]{
		LoadAllocator: deps.loadAllocator,
		FreeAllocator: deps.freeAllocator,
		ClearList:     deps.clearList,
	})
}

// SpellFreeDurations4FE880 binds GAME.EXE 004FE880 to the native-width
// duration-spell allocation class. ClassT.Free has a value receiver, matching
// the original function's observable stale allocator handle after destruction.
func (sp *SpellsDuration) SpellFreeDurations4FE880() {
	spellDurationCleanupNative4FE880(spellDurationCleanupNativeDeps4FE880{
		loadAllocator: func() alloc.ClassT[DurSpell] {
			return sp.alloc
		},
		freeAllocator: func(value alloc.ClassT[DurSpell]) {
			value.Free()
		},
		clearList: func() {
			sp.List = nil
		},
	})
}
