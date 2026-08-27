package server

const (
	playerCheckStrengthPlayerClass4F3180 = uint8(0x04)
	playerCheckStrengthArmorClass4F3180  = uint32(0x02000000)
)

// playerCheckStrengthHooks4F3180 exposes every observable load and call in
// GAME.EXE 004F3180. It deliberately has no nil or cheat hook: the original
// reads the player's class byte unconditionally and contains no allow-all
// branch.
type playerCheckStrengthHooks4F3180[O, M comparable] struct {
	loadPlayerClassLow func(O) uint8
	getUnitStrength    func(O) int32
	loadItemClass      func(O) uint32
	loadItemType       func(O) uint16
	findArmorDef       func(uint16) M
	findWeaponDef      func(uint16) M
	loadRequired       func(M) uint16
}

// playerCheckStrength4F3180 preserves the exact observable order of
// GAME.EXE 004F3180. A non-Player returns before strength or item access. A
// Player calls the strength service before loading any item field, selects an
// armor or weapon definition from the live full class and zero-extended type
// index, then performs a signed comparison against the zero-extended uint16
// requirement.
func playerCheckStrength4F3180[O, M comparable](
	player, item O,
	hooks playerCheckStrengthHooks4F3180[O, M],
) int32 {
	if hooks.loadPlayerClassLow(player)&playerCheckStrengthPlayerClass4F3180 == 0 {
		return 0
	}

	strength := hooks.getUnitStrength(player)
	itemClass := hooks.loadItemClass(item)
	itemType := hooks.loadItemType(item)

	var def M
	if itemClass&playerCheckStrengthArmorClass4F3180 != 0 {
		def = hooks.findArmorDef(itemType)
	} else {
		def = hooks.findWeaponDef(itemType)
	}
	var zero M
	if def == zero {
		return 0
	}
	if strength >= int32(hooks.loadRequired(def)) {
		return 1
	}
	return 0
}
