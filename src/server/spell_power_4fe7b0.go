package server

const (
	spellPowerImaginaryCasterType4FE7B0 = "ImaginaryCaster"
	spellPowerFullStrengthModes4FE7B0   = uint32(0x570)
	spellPowerPlayerClass4FE7B0         = 0x04
	spellPowerMonsterClass4FE7B0        = 0x02
)

// spellPowerHooks4FE7B0 exposes every observable load, store, and callback in
// GAME.EXE 004FE7B0. Pointer-bearing values are generic tokens, while object
// type, class, spell ID, game flags, and returned power retain the widths
// consumed by the original PE32 instruction stream.
type spellPowerHooks4FE7B0[Object comparable, Update, Player any] struct {
	loadImaginaryCasterType  func() uint32
	lookupObjectType         func(string) uint32
	storeImaginaryCasterType func(uint32)

	loadCasterArg  func() Object
	loadCasterType func(Object) uint16
	hasGameFlag    func(uint32) int32
	loadClass      func(Object) uint32

	loadUpdate       func(Object) Update
	loadSpellArg     func() int32
	loadPlayer       func(Update) Player
	loadPlayerPower  func(Player, int32) int32
	loadMonsterPower func(Update) int32
}

// spellPower4FE7B0 preserves GAME.EXE 004FE7B0's exact lazy-cache, argument,
// gate, and field-load order. The caster type is read unconditionally before
// the game-mode and nil gates. Thus an ordinary nil caster faults at the type
// load just as the original does; the later nil return remains represented for
// address models in which that load succeeds.
//
// The class dword is cached and only its low byte is tested, with Player taking
// precedence over Monster. The signed spell argument is observed only on the
// Player path and the loaded power dword is returned without canonicalization.
func spellPower4FE7B0[Object comparable, Update, Player any](
	h spellPowerHooks4FE7B0[Object, Update, Player],
) int32 {
	typeInd := h.loadImaginaryCasterType()
	if typeInd == 0 {
		typeInd = h.lookupObjectType(spellPowerImaginaryCasterType4FE7B0)
		h.storeImaginaryCasterType(typeInd)
	}

	caster := h.loadCasterArg()
	if uint32(h.loadCasterType(caster)) == typeInd {
		return 1
	}
	if h.hasGameFlag(spellPowerFullStrengthModes4FE7B0) != 0 {
		return 3
	}
	var nilObject Object
	if caster == nilObject {
		return 2
	}

	class := uint8(h.loadClass(caster))
	if class&spellPowerPlayerClass4FE7B0 != 0 {
		update := h.loadUpdate(caster)
		spellID := h.loadSpellArg()
		player := h.loadPlayer(update)
		return h.loadPlayerPower(player, spellID)
	}
	if class&spellPowerMonsterClass4FE7B0 == 0 {
		return 3
	}
	update := h.loadUpdate(caster)
	return h.loadMonsterPower(update)
}
