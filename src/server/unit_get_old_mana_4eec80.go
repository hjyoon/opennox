package server

type unitGetOldManaHooks4EEC80[O, U any] struct {
	loadUnitArg    func() O
	loadClass      func(O) uint32
	loadUpdateData func(O) U
	loadCurrent    func(U) uint16
}

// unitGetOldMana4EEC80 preserves GAME.EXE 004EEC80. The object argument is
// loaded once and nil is the only entry guard. The full class dword is then
// loaded once, while both decisions use only its low byte. Player takes
// precedence over Monster when both bits are set. A Player unconditionally
// loads update data and its current-mana word; there is no update-data nil
// guard. A non-Player Monster returns exactly 1000 and every other class
// returns zero. Only the 16-bit AX result is part of this function's ABI.
func unitGetOldMana4EEC80[O comparable, U any](hooks unitGetOldManaHooks4EEC80[O, U]) uint16 {
	unit := hooks.loadUnitArg()
	var nilObject O
	if unit == nilObject {
		return 0
	}

	classLow := uint8(hooks.loadClass(unit))
	if classLow&0x04 != 0 {
		update := hooks.loadUpdateData(unit)
		return hooks.loadCurrent(update)
	}
	if classLow&0x02 != 0 {
		return 1000
	}
	return 0
}
