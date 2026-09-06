package server

type spellCancelDurSpellNativeDeps4FEB10 struct {
	loadFirst  func() *DurSpell
	loadSpell  func(*DurSpell) int32
	loadNext   func(*DurSpell) *DurSpell
	loadCaster func(*DurSpell) *Object
	cancel     func(*DurSpell)
}

func spellCancelDurSpellNative4FEB10(
	requestedSpell int32,
	caster *Object,
	deps spellCancelDurSpellNativeDeps4FEB10,
) int32 {
	return SpellCancelDurSpell4FEB10(requestedSpell, caster, SpellCancelDurSpellHooks4FEB10[*DurSpell, *Object]{
		LoadFirst:  deps.loadFirst,
		LoadSpell:  deps.loadSpell,
		LoadNext:   deps.loadNext,
		LoadCaster: deps.loadCaster,
		Cancel:     deps.cancel,
	})
}

// SpellCancelDurSpell4FEB10 binds GAME.EXE 004FEB10 to the native-width
// duration list and Object pointers. Spell remains an exact signed dword,
// while record links and caster identities retain the host pointer width.
// Field loads remain live at their original access points and cancellation
// traverses through the successor snapshotted before the callback.
//
//go:noinline
func (sp *SpellsDuration) SpellCancelDurSpell4FEB10(requestedSpell int32, caster *Object) int32 {
	return spellCancelDurSpellNative4FEB10(requestedSpell, caster, spellCancelDurSpellNativeDeps4FEB10{
		loadFirst: func() *DurSpell {
			return sp.List
		},
		loadSpell: func(record *DurSpell) int32 {
			return int32(record.Spell)
		},
		loadNext: func(record *DurSpell) *DurSpell {
			return record.Next
		},
		loadCaster: func(record *DurSpell) *Object {
			return record.Caster16
		},
		cancel: func(record *DurSpell) {
			sp.SpellDurationCancel4FE9D0(record)
		},
	})
}
