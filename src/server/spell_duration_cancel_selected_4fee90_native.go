package server

type spellDurationCancelSelectedNativeDeps4FEE90 struct {
	loadFirst  func() *DurSpell
	loadNext   func(*DurSpell) *DurSpell
	loadCaster func(*DurSpell) *Object
	loadSpell  func(*DurSpell) uint32
	cancel     func(*DurSpell)
}

func spellDurationCancelSelectedNative4FEE90(
	caster *Object,
	deps spellDurationCancelSelectedNativeDeps4FEE90,
) {
	SpellDurationCancelSelected4FEE90(
		SpellDurationCancelSelectedHooks4FEE90[*DurSpell, *Object]{
			LoadFirst: deps.loadFirst,
			LoadCasterArg: func() *Object {
				return caster
			},
			LoadNext:   deps.loadNext,
			LoadCaster: deps.loadCaster,
			LoadSpell:  deps.loadSpell,
			Cancel:     deps.cancel,
		},
	)
}

func spellDurationCancelSelectedServerDeps4FEE90(
	sp *SpellsDuration,
) spellDurationCancelSelectedNativeDeps4FEE90 {
	return spellDurationCancelSelectedNativeDeps4FEE90{
		loadFirst: func() *DurSpell {
			return sp.SpellDurationFirst4FE930()
		},
		loadNext: SpellDurationNextNative4FE940,
		loadCaster: func(record *DurSpell) *Object {
			return record.Caster16
		},
		loadSpell: func(record *DurSpell) uint32 {
			return record.Spell
		},
		cancel: func(record *DurSpell) {
			_ = sp.SpellDurationCancel4FE9D0(record)
		},
	}
}

// SpellDurationCancelSelected4FEE90 binds GAME.EXE 004FEE90 to native-width
// *DurSpell links and *Object caster identities. The head and successor loads
// use the restored 004FE930 and 004FE940 accessors, and selected records pass
// to the restored native 004FE9D0 cancellation boundary. Spell remains an
// exact dword. All three decoded callers are Go-owned, so no independent
// C/CGo ABI is retained.
//
//go:noinline
func (sp *SpellsDuration) SpellDurationCancelSelected4FEE90(caster *Object) {
	spellDurationCancelSelectedNative4FEE90(
		caster,
		spellDurationCancelSelectedServerDeps4FEE90(sp),
	)
}
