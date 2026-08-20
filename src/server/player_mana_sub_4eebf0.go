package server

type playerManaSubHooks4EEBF0[O comparable, U, P, R any] struct {
	loadUnitArg       func() (O, R)
	loadClassLow      func(O) uint8
	loadEngineGodMode func() bool
	loadUpdateData    func(O) (U, R)
	loadCurrent       func(U) uint16
	loadAmountArg     func() int32
	storePrevious     func(U, uint16)
	storeCurrent      func(U, uint16)
	loadPlayer        func(U) P
	loadProtection    func(P) uint32
	protectMana       func(uint32, int16) R
}

// playerManaSub4EEBF0 preserves GAME.EXE 004EEBF0. R carries the value that
// the original leaves in EAX without imposing an ABI32 pointer representation
// on native-width code. Entry gates return the loaded unit value. The Player
// path caches the GodMode byte before loading update data, but it tests that
// byte only after the update pointer has replaced EAX; a GodMode exit therefore
// returns the update value and does not read current mana or amount.
//
// The active path reads the unsigned current word before the whole signed
// 32-bit amount. Both comparisons promote the mana word to signed int32. The
// first comparison selects a wrapping low-word subtraction or zero. After a
// live current-mana reload, the second comparison deliberately sends -amount
// only when new mana is greater than amount; otherwise it sends -newMana.
// Protection observes a cached Player pointer and its return becomes the
// function result. There are no update-data or Player nil guards.
func playerManaSub4EEBF0[O comparable, U, P, R any](
	hooks playerManaSubHooks4EEBF0[O, U, P, R],
) R {
	unit, result := hooks.loadUnitArg()
	var nilObject O
	if unit == nilObject {
		return result
	}
	if hooks.loadClassLow(unit)&0x04 == 0 {
		return result
	}

	godMode := hooks.loadEngineGodMode()
	update, result := hooks.loadUpdateData(unit)
	if godMode {
		return result
	}

	current := hooks.loadCurrent(update)
	amount := hooks.loadAmountArg()
	hooks.storePrevious(update, current)
	if int32(current) > amount {
		hooks.storeCurrent(update, uint16(int32(current)-amount))
	} else {
		hooks.storeCurrent(update, 0)
	}

	current = hooks.loadCurrent(update)
	delta := int16(-int32(current))
	if int32(current) > amount {
		delta = int16(-amount)
	}
	player := hooks.loadPlayer(update)
	token := hooks.loadProtection(player)
	return hooks.protectMana(token, delta)
}
