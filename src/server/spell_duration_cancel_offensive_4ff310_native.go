package server

import "github.com/opennox/libs/spell"

type spellDurationCancelOffensiveNativeDeps4FF310 struct {
	loadFirst      func() *DurSpell
	loadCasterArg  func() *Object
	loadCaster     func(*DurSpell) *Object
	loadNext       func(*DurSpell) *DurSpell
	loadSpell      func(*DurSpell) int32
	loadSpellFlags func(int32) uint32
	cancel         func(*DurSpell)
}

func spellDurationCancelOffensiveNative4FF310(
	deps spellDurationCancelOffensiveNativeDeps4FF310,
) {
	SpellDurationCancelOffensive4FF310(
		SpellDurationCancelOffensiveHooks4FF310[*DurSpell, *Object]{
			LoadFirst:      deps.loadFirst,
			LoadCasterArg:  deps.loadCasterArg,
			LoadCaster:     deps.loadCaster,
			LoadNext:       deps.loadNext,
			LoadSpell:      deps.loadSpell,
			LoadSpellFlags: deps.loadSpellFlags,
			Cancel:         deps.cancel,
		},
	)
}

func spellDurationCancelOffensiveServerDeps4FF310(
	sp *SpellsDuration,
	caster *Object,
) spellDurationCancelOffensiveNativeDeps4FF310 {
	return spellDurationCancelOffensiveNativeDeps4FF310{
		loadFirst: sp.SpellDurationFirst4FE930,
		loadCasterArg: func() *Object {
			return caster
		},
		loadCaster: func(record *DurSpell) *Object {
			return record.Caster16
		},
		loadNext: SpellDurationNextNative4FE940,
		loadSpell: func(record *DurSpell) int32 {
			return int32(record.Spell)
		},
		loadSpellFlags: func(spellID int32) uint32 {
			return uint32(sp.s.Spells.Flags(spell.ID(spellID)))
		},
		cancel: func(record *DurSpell) {
			_ = sp.SpellDurationCancel4FE9D0(record)
		},
	}
}

// SpellDurationCancelOffensive4FF310 binds GAME.EXE 004FF310 to native-width
// duration-spell and caster identities. Spell is passed to the flags lookup as
// an exact signed dword, while only the returned low byte's offensive bit is
// tested. Cancellation uses the restored 004FE9D0 implementation and traversal
// retains the successor loaded before that callback.
//
//go:noinline
func (sp *SpellsDuration) SpellDurationCancelOffensive4FF310(caster *Object) {
	spellDurationCancelOffensiveNative4FF310(
		spellDurationCancelOffensiveServerDeps4FF310(sp, caster),
	)
}
