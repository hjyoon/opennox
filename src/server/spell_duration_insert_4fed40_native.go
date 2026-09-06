package server

type spellDurationInsertNativeDeps4FED40 struct {
	loadHead  func() *DurSpell
	storePrev func(*DurSpell, *DurSpell)
	storeNext func(*DurSpell, *DurSpell)
	storeHead func(*DurSpell)
}

func spellDurationInsertNative4FED40(
	record *DurSpell,
	deps spellDurationInsertNativeDeps4FED40,
) {
	spellDurationInsert4FED40(spellDurationInsertHooks4FED40[*DurSpell]{
		loadHead: deps.loadHead,
		loadRecordArg: func() *DurSpell {
			return record
		},
		storePrev: deps.storePrev,
		storeNext: deps.storeNext,
		storeHead: deps.storeHead,
	})
}

func spellDurationInsertServerDeps4FED40(sp *SpellsDuration) spellDurationInsertNativeDeps4FED40 {
	return spellDurationInsertNativeDeps4FED40{
		loadHead: func() *DurSpell {
			return sp.List
		},
		storePrev: func(record, prev *DurSpell) {
			record.Prev = prev
		},
		storeNext: func(record, next *DurSpell) {
			record.Next = next
		},
		storeHead: func(record *DurSpell) {
			sp.List = record
		},
	}
}

// SpellDurationInsert4FED40 binds GAME.EXE 004FED40 to native-width
// *DurSpell list links. The first list-head value remains cached for its Prev
// store, while the record's Next receives an independent live head load.
// There is deliberately no nil-record guard and no retained C ABI.
//
//go:noinline
func (sp *SpellsDuration) SpellDurationInsert4FED40(record *DurSpell) {
	spellDurationInsertNative4FED40(record, spellDurationInsertServerDeps4FED40(sp))
}
