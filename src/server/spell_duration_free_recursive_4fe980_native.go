package server

import "github.com/opennox/opennox/v1/legacy/common/alloc"

type spellDurationFreeRecursiveNativeDeps4FE980 struct {
	loadSub108      func(*DurSpell) *DurSpell
	loadSub104      func(*DurSpell) *DurSpell
	loadNext        func(*DurSpell) *DurSpell
	loadAllocator   func() alloc.ClassT[DurSpell]
	freeObjectFirst func(alloc.ClassT[DurSpell], *DurSpell)
}

func spellDurationFreeRecursiveNative4FE980(
	record *DurSpell,
	deps spellDurationFreeRecursiveNativeDeps4FE980,
) {
	SpellDurationFreeRecursive4FE980(
		SpellDurationFreeRecursiveHooks4FE980[alloc.ClassT[DurSpell], *DurSpell]{
			LoadSub108:      deps.loadSub108,
			LoadSub104:      deps.loadSub104,
			LoadNext:        deps.loadNext,
			LoadAllocator:   deps.loadAllocator,
			FreeObjectFirst: deps.freeObjectFirst,
		},
		record,
	)
}

// SpellDurationFreeRecursive4FE980 binds GAME.EXE 004FE980 to native-width
// DurSpell child and sibling links and the native allocation class. The
// original PE32 offsets widen through typed fields without truncating record
// identity on 64-bit targets.
//
//go:noinline
func (sp *SpellsDuration) SpellDurationFreeRecursive4FE980(record *DurSpell) {
	spellDurationFreeRecursiveNative4FE980(record, spellDurationFreeRecursiveNativeDeps4FE980{
		loadSub108: func(value *DurSpell) *DurSpell {
			return value.Sub108
		},
		loadSub104: func(value *DurSpell) *DurSpell {
			return value.Sub104
		},
		loadNext: func(value *DurSpell) *DurSpell {
			return value.Next
		},
		loadAllocator: func() alloc.ClassT[DurSpell] {
			return sp.alloc
		},
		freeObjectFirst: func(allocator alloc.ClassT[DurSpell], value *DurSpell) {
			allocator.FreeObjectFirst(value)
		},
	})
}
