package opennox

type spellGestureResetHooks4FCAC0[Allocator, Unit comparable, Update any, Object comparable] struct {
	resetDurations             func(int32)
	loadMagicClass             func() Allocator
	freeAllMagicObjects        func(Allocator)
	clearMagicEntityHead       func()
	firstPlayerUnit            func() Unit
	loadPlayerUpdate           func(Unit) Update
	storeField47LowByte        func(Update, uint8)
	storeSpellCastStart        func(Update, uint32)
	storeTrapSpell             func(Update, int, uint32)
	storeTrapSpellCountLowByte func(Update, uint8)
	nextPlayerUnit             func(Unit) Unit
	newObjectByTypeID          func(string) Object
	storeImaginaryCaster       func(Object)
	createObjectAt             func(object, owner Object, x, y float32)
}

// spellGestureReset4FCAC0 preserves GAME.EXE 004FCAC0's observable callback,
// load, and store order. All pointer-bearing values remain native-width tokens.
// The allocator is deliberately forwarded even when nil, and only the low byte
// of the player's trap-spell count is cleared.
func spellGestureReset4FCAC0[Allocator, Unit comparable, Update any, Object comparable](
	a1, a2 int32,
	hooks spellGestureResetHooks4FCAC0[Allocator, Unit, Update, Object],
) int32 {
	hooks.resetDurations(a1)
	magicClass := hooks.loadMagicClass()
	hooks.freeAllMagicObjects(magicClass)
	hooks.clearMagicEntityHead()

	var nilUnit Unit
	for unit := hooks.firstPlayerUnit(); unit != nilUnit; unit = hooks.nextPlayerUnit(unit) {
		update := hooks.loadPlayerUpdate(unit)
		hooks.storeField47LowByte(update, 0)
		hooks.storeSpellCastStart(update, 0)
		for i := 0; i < 5; i++ {
			hooks.storeTrapSpell(update, i, 0)
		}
		hooks.storeTrapSpellCountLowByte(update, 0)
	}

	if a2 == 0 {
		return 1
	}
	caster := hooks.newObjectByTypeID(spellRuntimeCasterType4FC9B0)
	hooks.storeImaginaryCaster(caster)
	var nilObject Object
	if caster == nilObject {
		return 0
	}
	hooks.createObjectAt(
		caster,
		nilObject,
		spellRuntimeMapCenter4FC9B0,
		spellRuntimeMapCenter4FC9B0,
	)
	return 1
}
