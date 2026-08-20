package server

type playerManaRefreshHooks4EECF0[O comparable, U, P, R any] struct {
	loadUnitArg    func() (O, R)
	loadClassLow   func(O) uint8
	loadUpdateData func(O) U
	loadCurrent    func(U) uint16
	loadPlayer     func(U) P
	storePrevious  func(U, uint16)
	loadMaximum    func(U) uint16
	storeCurrent   func(U, uint16)
	loadProtection func(P) uint32
	protectMana    func(uint32, int16) R
}

// playerManaRefresh4EECF0 preserves GAME.EXE 004EECF0. R carries the value
// left in EAX without narrowing native pointers to the original ABI32 width.
// Nil and non-Player gates return the loaded unit value. A Player caches its
// UpdateData, reads current mana and Player in that order, stores current into
// previous, then reads maximum and stores it into current. Only after both
// stores does it load the cached Player's protection token and pass the signed
// maximum word to protection. There are no UpdateData or Player nil guards.
func playerManaRefresh4EECF0[O comparable, U, P, R any](
	hooks playerManaRefreshHooks4EECF0[O, U, P, R],
) R {
	unit, result := hooks.loadUnitArg()
	var nilObject O
	if unit == nilObject {
		return result
	}
	if hooks.loadClassLow(unit)&0x04 == 0 {
		return result
	}
	update := hooks.loadUpdateData(unit)
	current := hooks.loadCurrent(update)
	player := hooks.loadPlayer(update)
	hooks.storePrevious(update, current)
	maximum := hooks.loadMaximum(update)
	hooks.storeCurrent(update, maximum)
	token := hooks.loadProtection(player)
	return hooks.protectMana(token, int16(maximum))
}
