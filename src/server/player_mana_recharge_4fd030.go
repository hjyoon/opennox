package server

const playerManaRechargePlayerClass4FD030 = uint8(0x04)

type playerManaRechargeHooks4FD030[Unit any] struct {
	loadUnitArg   func() (Unit, uint16)
	loadClassLow  func(Unit) uint8
	loadAmountArg func() int16
	addMana       func(Unit, int16) uint16
}

// playerManaRecharge4FD030 preserves GAME.EXE 004FD030's exact observation
// order while allowing Unit to retain native pointer width. The object
// argument is loaded once and its low word becomes the default result. The
// class byte is then unconditionally read; there is deliberately no nil
// guard because the original faults at that dereference.
//
// The signed 16-bit amount is observed only on the Player path. Both the
// cached object and cached amount are passed to 004EEB80, and that function's
// uint16 result is returned without reinterpretation.
func playerManaRecharge4FD030[Unit any](
	hooks playerManaRechargeHooks4FD030[Unit],
) uint16 {
	unit, result := hooks.loadUnitArg()
	if hooks.loadClassLow(unit)&playerManaRechargePlayerClass4FD030 == 0 {
		return result
	}
	amount := hooks.loadAmountArg()
	return hooks.addMana(unit, amount)
}
