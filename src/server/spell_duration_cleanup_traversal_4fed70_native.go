package server

type spellDurationCleanupTraversalNativeDeps4FED70 struct {
	loadFirst        func() *DurSpell
	loadFlagsLowByte func(*DurSpell) byte
	loadNext         func(*DurSpell) *DurSpell
	destroy          func(*DurSpell)
}

func spellDurationCleanupTraversalNative4FED70(
	deps spellDurationCleanupTraversalNativeDeps4FED70,
) {
	SpellDurationCleanupTraversal4FED70(
		SpellDurationCleanupTraversalHooks4FED70[*DurSpell]{
			LoadFirst:        deps.loadFirst,
			LoadFlagsLowByte: deps.loadFlagsLowByte,
			LoadNext:         deps.loadNext,
			Destroy:          deps.destroy,
		},
	)
}

func spellDurationCleanupTraversalServerDeps4FED70(
	sp *SpellsDuration,
	destroy func(*DurSpell),
) spellDurationCleanupTraversalNativeDeps4FED70 {
	return spellDurationCleanupTraversalNativeDeps4FED70{
		loadFirst: func() *DurSpell {
			return sp.List
		},
		loadFlagsLowByte: func(record *DurSpell) byte {
			return byte(record.Flags88)
		},
		loadNext: func(record *DurSpell) *DurSpell {
			return record.Next
		},
		destroy: destroy,
	}
}

// SpellDurationCleanupTraversal4FED70 binds GAME.EXE 004FED70 to native-width
// *DurSpell list links and the native record accepted by the restored 004FEDA0
// destruction callback. The low flag byte is loaded before Next, and Next is
// snapshotted before destruction. There is deliberately no retained C ABI.
//
//go:noinline
func (sp *SpellsDuration) SpellDurationCleanupTraversal4FED70(
	destroy func(*DurSpell),
) {
	spellDurationCleanupTraversalNative4FED70(
		spellDurationCleanupTraversalServerDeps4FED70(sp, destroy),
	)
}
