package server

import "github.com/opennox/opennox/v1/legacy/common/alloc"

type spellDurationSelectiveCleanupNativeDeps4FE8A0 struct {
	loadAllocator    func() alloc.ClassT[DurSpell]
	freeAllObjects   func(alloc.ClassT[DurSpell])
	clearList        func()
	loadList         func() *DurSpell
	loadTarget       func(*DurSpell) *Object
	loadNext         func(*DurSpell) *DurSpell
	loadClassLowByte func(*Object) uint8
	unlink           func(*DurSpell)
	freeRecursive    func(*DurSpell)
}

func spellDurationSelectiveCleanupNative4FE8A0(
	mode int32,
	deps spellDurationSelectiveCleanupNativeDeps4FE8A0,
) {
	spellDurationSelectiveCleanup4FE8A0(
		mode,
		spellDurationSelectiveCleanupHooks4FE8A0[*DurSpell, *Object, alloc.ClassT[DurSpell]]{
			loadAllocator:    deps.loadAllocator,
			freeAllObjects:   deps.freeAllObjects,
			clearList:        deps.clearList,
			loadList:         deps.loadList,
			loadTarget:       deps.loadTarget,
			loadNext:         deps.loadNext,
			loadClassLowByte: deps.loadClassLowByte,
			unlink:           deps.unlink,
			freeRecursive:    deps.freeRecursive,
		},
	)
}

// SpellResetDurations4FE8A0 binds GAME.EXE 004FE8A0 to native-width duration
// records and Object pointers. Only the low byte of ObjClass is observed, as
// in the original TEST BYTE instruction.
//
//go:noinline
func (sp *SpellsDuration) SpellResetDurations4FE8A0(mode int32) {
	spellDurationSelectiveCleanupNative4FE8A0(mode, spellDurationSelectiveCleanupNativeDeps4FE8A0{
		loadAllocator: func() alloc.ClassT[DurSpell] {
			return sp.alloc
		},
		freeAllObjects: func(value alloc.ClassT[DurSpell]) {
			value.FreeAllObjects()
		},
		clearList: func() {
			sp.List = nil
		},
		loadList: func() *DurSpell {
			return sp.List
		},
		loadTarget: func(record *DurSpell) *Object {
			return record.Target48
		},
		loadNext: func(record *DurSpell) *DurSpell {
			return record.Next
		},
		loadClassLowByte: func(object *Object) uint8 {
			return uint8(object.ObjClass)
		},
		unlink:        sp.Unlink,
		freeRecursive: sp.FreeRecursive,
	})
}
