package server

type spellDurationDuplicateNativeDeps4FEE50 struct {
	loadHead         func() *DurSpell
	loadFlag20       func(*DurSpell) uint32
	loadSpell        func(*DurSpell) uint32
	loadCaster       func(*DurSpell) *Object
	loadFlagsLowByte func(*DurSpell) byte
	loadNext         func(*DurSpell) *DurSpell
}

func spellDurationDuplicateNative4FEE50(
	requestedSpell int32,
	caster *Object,
	deps spellDurationDuplicateNativeDeps4FEE50,
) int32 {
	return spellDurationDuplicate4FEE50(spellDurationDuplicateHooks4FEE50[*DurSpell, *Object]{
		loadHead: deps.loadHead,
		loadCasterArg: func() *Object {
			return caster
		},
		loadSpellArg: func() uint32 {
			return uint32(requestedSpell)
		},
		loadFlag20:       deps.loadFlag20,
		loadSpell:        deps.loadSpell,
		loadCaster:       deps.loadCaster,
		loadFlagsLowByte: deps.loadFlagsLowByte,
		loadNext:         deps.loadNext,
	})
}

func spellDurationDuplicateServerDeps4FEE50(sp *SpellsDuration) spellDurationDuplicateNativeDeps4FEE50 {
	return spellDurationDuplicateNativeDeps4FEE50{
		loadHead: func() *DurSpell {
			return sp.List
		},
		loadFlag20: func(record *DurSpell) uint32 {
			return record.Flag20
		},
		loadSpell: func(record *DurSpell) uint32 {
			return record.Spell
		},
		loadCaster: func(record *DurSpell) *Object {
			return record.Caster16
		},
		loadFlagsLowByte: func(record *DurSpell) byte {
			return byte(record.Flags88)
		},
		loadNext: func(record *DurSpell) *DurSpell {
			return record.Next
		},
	}
}

// SpellDurationDuplicate4FEE50 binds GAME.EXE 004FEE50 to native-width
// *DurSpell links and *Object caster identities. Spell and Flag20 remain exact
// dwords, while only the low Flags88 byte is observed. Empty and exhausted
// paths return canonical zero, and a matching live record returns canonical
// one before reading its Next link. Both production callers are Go-owned, so
// no independent C/CGo ABI is retained.
//
//go:noinline
func (sp *SpellsDuration) SpellDurationDuplicate4FEE50(requestedSpell int32, caster *Object) int32 {
	return spellDurationDuplicateNative4FEE50(
		requestedSpell,
		caster,
		spellDurationDuplicateServerDeps4FEE50(sp),
	)
}
