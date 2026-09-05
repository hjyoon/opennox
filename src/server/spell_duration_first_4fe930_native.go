package server

type spellDurationFirstNativeDeps4FE930 struct {
	loadHead func() *DurSpell
}

func spellDurationFirstNative4FE930(deps spellDurationFirstNativeDeps4FE930) *DurSpell {
	return SpellDurationFirst4FE930(SpellDurationFirstHooks4FE930[*DurSpell]{
		LoadHead: deps.loadHead,
	})
}

// SpellDurationFirst4FE930 binds GAME.EXE 004FE930 to the native-width
// duration-spell list head. The returned *DurSpell is neither copied nor
// converted through the original PE32 pointer width.
//
//go:noinline
func (sp *SpellsDuration) SpellDurationFirst4FE930() *DurSpell {
	return spellDurationFirstNative4FE930(spellDurationFirstNativeDeps4FE930{
		loadHead: func() *DurSpell {
			return sp.List
		},
	})
}
