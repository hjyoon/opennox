package server

type spellDurationFindActiveTargetNativeDeps4FF2D0 struct {
	loadFirst    func() *DurSpell
	loadFlagsLow func(*DurSpell) byte
	loadSpell    func(*DurSpell) int32
	loadTarget   func(*DurSpell) *Object
	loadNext     func(*DurSpell) *DurSpell
}

func spellDurationFindActiveTargetNative4FF2D0(
	requestedSpell int32,
	target *Object,
	deps spellDurationFindActiveTargetNativeDeps4FF2D0,
) *DurSpell {
	return SpellDurationFindActiveTarget4FF2D0(
		requestedSpell,
		target,
		SpellDurationFindActiveTargetHooks4FF2D0[*DurSpell, *Object]{
			LoadFirst:    deps.loadFirst,
			LoadFlagsLow: deps.loadFlagsLow,
			LoadSpell:    deps.loadSpell,
			LoadTarget:   deps.loadTarget,
			LoadNext:     deps.loadNext,
		},
	)
}

func spellDurationFindActiveTargetServerDeps4FF2D0(
	sp *SpellsDuration,
) spellDurationFindActiveTargetNativeDeps4FF2D0 {
	return spellDurationFindActiveTargetNativeDeps4FF2D0{
		loadFirst: sp.SpellDurationFirst4FE930,
		loadFlagsLow: func(record *DurSpell) byte {
			return byte(record.Flags88)
		},
		loadSpell: func(record *DurSpell) int32 {
			return int32(record.Spell)
		},
		loadTarget: func(record *DurSpell) *Object {
			return record.Target48
		},
		loadNext: SpellDurationNextNative4FE940,
	}
}

// SpellDurationFindActiveTarget4FF2D0 binds GAME.EXE 004FF2D0 to the
// native-width duration-spell list and object identities. Spell remains an
// exact signed dword view, only Flags88's low byte is tested, and both the
// returned *DurSpell and target comparison retain the host pointer width.
//
//go:noinline
func (sp *SpellsDuration) SpellDurationFindActiveTarget4FF2D0(
	requestedSpell int32,
	target *Object,
) *DurSpell {
	return spellDurationFindActiveTargetNative4FF2D0(
		requestedSpell,
		target,
		spellDurationFindActiveTargetServerDeps4FF2D0(sp),
	)
}
