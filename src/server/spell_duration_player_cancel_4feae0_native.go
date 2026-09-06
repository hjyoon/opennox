package server

type playerCancelSpellsNativeDeps4FEAE0 struct {
	loadFirst  func() *DurSpell
	loadCaster func(*DurSpell) *Object
	loadNext   func(*DurSpell) *DurSpell
	cancel     func(*DurSpell)
}

func playerCancelSpellsNative4FEAE0(
	caster *Object,
	deps playerCancelSpellsNativeDeps4FEAE0,
) int32 {
	return PlayerCancelSpells4FEAE0(caster, PlayerCancelSpellsHooks4FEAE0[*DurSpell, *Object]{
		LoadFirst:  deps.loadFirst,
		LoadCaster: deps.loadCaster,
		LoadNext:   deps.loadNext,
		Cancel:     deps.cancel,
	})
}

// PlayerCancelSpells4FEAE0 binds GAME.EXE 004FEAE0 to native-width duration
// records and Object pointers. Record fields remain live at their original
// access points, while each Next pointer is preserved across cancellation.
//
//go:noinline
func (sp *SpellsDuration) PlayerCancelSpells4FEAE0(caster *Object) int32 {
	return playerCancelSpellsNative4FEAE0(caster, playerCancelSpellsNativeDeps4FEAE0{
		loadFirst: func() *DurSpell {
			return sp.List
		},
		loadCaster: func(record *DurSpell) *Object {
			return record.Caster16
		},
		loadNext: func(record *DurSpell) *DurSpell {
			return record.Next
		},
		cancel: func(record *DurSpell) {
			sp.SpellDurationCancel4FE9D0(record)
		},
	})
}
