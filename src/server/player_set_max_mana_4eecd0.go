package server

type playerSetMaxManaHooks4EECD0[O comparable, U, R any] struct {
	loadUnitArg    func() (O, R)
	loadClassLow   func(O) uint8
	loadUpdateData func(O) (U, R)
	loadMaximumArg func() uint16
	storeMaximum   func(U, uint16)
}

// playerSetMaxMana4EECD0 preserves GAME.EXE 004EECD0. R carries the value
// left in EAX without forcing native pointers into the original ABI32 width.
// Nil and non-Player gates return the loaded unit value and do not read the
// maximum argument. A Player replaces EAX with its unguarded UpdateData
// pointer, then reads the low argument word and stores it at ManaMax. Thus a
// nil UpdateData still observes the argument load before the store faults.
func playerSetMaxMana4EECD0[O comparable, U, R any](
	hooks playerSetMaxManaHooks4EECD0[O, U, R],
) R {
	unit, result := hooks.loadUnitArg()
	var nilObject O
	if unit == nilObject {
		return result
	}
	if hooks.loadClassLow(unit)&0x04 == 0 {
		return result
	}
	update, result := hooks.loadUpdateData(unit)
	maximum := hooks.loadMaximumArg()
	hooks.storeMaximum(update, maximum)
	return result
}
