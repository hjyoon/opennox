package server

type playerGetMaxManaHooks4EECB0[O, U any] struct {
	loadUnitArg    func() O
	loadClassLow   func(O) uint8
	loadUpdateData func(O) U
	loadMaximum    func(U) uint16
}

// playerGetMaxMana4EECB0 preserves GAME.EXE 004EECB0. The object argument is
// loaded once and nil is the first gate. The original then reads only the low
// class byte; a non-Player returns zero without touching update data. A Player
// unconditionally loads update data and its maximum-mana word, with no
// update-data nil guard. Only the 16-bit AX result is part of the ABI.
func playerGetMaxMana4EECB0[O comparable, U any](hooks playerGetMaxManaHooks4EECB0[O, U]) uint16 {
	unit := hooks.loadUnitArg()
	var nilObject O
	if unit == nilObject {
		return 0
	}
	if hooks.loadClassLow(unit)&0x04 == 0 {
		return 0
	}
	update := hooks.loadUpdateData(unit)
	return hooks.loadMaximum(update)
}
