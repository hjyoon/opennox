package server

type playerManaAddHooks4EEB80[O, U, P comparable] struct {
	loadUnitArg         func() (O, uint16)
	loadClassLow        func(O) uint8
	loadUpdateData      func(O) U
	loadAmountArg       func() int32
	loadCurrent         func(U) uint16
	loadMaximum         func(U) uint16
	storePrevious       func(U, uint16)
	storeCurrent        func(U, uint16)
	loadPlayer          func(U) P
	loadProtection      func(P) uint32
	protectMana         func(uint32, int16)
	protectPlayerHPMana func(uint32, uint16) uint16
}

// playerManaAdd4EEB80 preserves GAME.EXE 004EEB80. The second value returned
// by loadUnitArg is AX after the original pointer argument load; it is the
// observable result for a nil or non-Player object. The Player path caches its
// update-data pointer, reads the whole 32-bit amount slot before current and
// maximum mana, stores the wrapping 16-bit sum before its unsigned clamp, and
// then updates the protection token with the signed low word of the amount.
//
// protectMana may mutate the live update data. Consequently maximum and
// current mana are reloaded after that callback, in that order. A value above
// the live maximum reloads Player and its token and returns the low word from
// protectPlayerHPMana; otherwise the live maximum is returned. There are no
// update-data or Player nil guards in the original.
func playerManaAdd4EEB80[O, U, P comparable](
	hooks playerManaAddHooks4EEB80[O, U, P],
) uint16 {
	unit, result := hooks.loadUnitArg()
	var nilObject O
	if unit == nilObject {
		return result
	}
	if hooks.loadClassLow(unit)&0x04 == 0 {
		return result
	}

	update := hooks.loadUpdateData(unit)
	amount := hooks.loadAmountArg()
	current := hooks.loadCurrent(update)
	maximum := hooks.loadMaximum(update)
	hooks.storePrevious(update, current)
	current = uint16(int32(current) + amount)
	hooks.storeCurrent(update, current)
	if current > maximum {
		hooks.storeCurrent(update, maximum)
	}

	player := hooks.loadPlayer(update)
	token := hooks.loadProtection(player)
	hooks.protectMana(token, int16(amount))

	maximum = hooks.loadMaximum(update)
	result = maximum
	if hooks.loadCurrent(update) <= maximum {
		return result
	}
	player = hooks.loadPlayer(update)
	token = hooks.loadProtection(player)
	return hooks.protectPlayerHPMana(token, maximum)
}
