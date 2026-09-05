package server

type spellDurationNextNativeDeps4FE940 struct {
	loadNext func(*DurSpell) *DurSpell
}

func spellDurationNextNative4FE940(record *DurSpell, deps spellDurationNextNativeDeps4FE940) *DurSpell {
	return SpellDurationNext4FE940(record, SpellDurationNextHooks4FE940[*DurSpell]{
		LoadNext: deps.loadNext,
	})
}

// SpellDurationNextNative4FE940 binds GAME.EXE 004FE940 to the native-width
// DurSpell.Next field. Neither the input nor the returned *DurSpell is copied
// or converted through the original PE32 pointer width.
//
//go:noinline
func SpellDurationNextNative4FE940(record *DurSpell) *DurSpell {
	return spellDurationNextNative4FE940(record, spellDurationNextNativeDeps4FE940{
		loadNext: func(value *DurSpell) *DurSpell {
			return value.Next
		},
	})
}
