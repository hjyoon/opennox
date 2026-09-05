package server

type spellDurationUnlinkNativeDeps4FE900 struct {
	loadPrev  func(*DurSpell) *DurSpell
	loadNext  func(*DurSpell) *DurSpell
	storeNext func(*DurSpell, *DurSpell)
	storeHead func(*DurSpell)
	storePrev func(*DurSpell, *DurSpell)
}

func spellDurationUnlinkNative4FE900(
	record *DurSpell,
	deps spellDurationUnlinkNativeDeps4FE900,
) {
	spellDurationUnlink4FE900(record, spellDurationUnlinkHooks4FE900[*DurSpell]{
		loadPrev:  deps.loadPrev,
		loadNext:  deps.loadNext,
		storeNext: deps.storeNext,
		storeHead: deps.storeHead,
		storePrev: deps.storePrev,
	})
}

// SpellDurationUnlink4FE900 binds GAME.EXE 004FE900 to native-width
// *DurSpell links. The detached record's own Prev and Next fields remain
// untouched, matching the original intrusive-list helper.
//
//go:noinline
func (sp *SpellsDuration) SpellDurationUnlink4FE900(record *DurSpell) {
	spellDurationUnlinkNative4FE900(record, spellDurationUnlinkNativeDeps4FE900{
		loadPrev: func(value *DurSpell) *DurSpell {
			return value.Prev
		},
		loadNext: func(value *DurSpell) *DurSpell {
			return value.Next
		},
		storeNext: func(value, next *DurSpell) {
			value.Next = next
		},
		storeHead: func(value *DurSpell) {
			sp.List = value
		},
		storePrev: func(value, prev *DurSpell) {
			value.Prev = prev
		},
	})
}
