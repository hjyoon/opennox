package server

type unitSetMaxHPHooks4EE7C0[O, H any] struct {
	loadUnitArg    func() O
	loadHealth     func(O) H
	loadMaximumArg func() uint16
	storeMaximum   func(H, uint16)
}

// unitSetMaxHP4EE7C0 preserves GAME.EXE 004EE7C0. The object and HealthData
// pointers are each loaded exactly once. The maximum-HP argument is not read
// until both pointers pass their nil gates. The exact input word is then
// stored through the cached HealthData pointer, which is also returned.
func unitSetMaxHP4EE7C0[O, H comparable](hooks unitSetMaxHPHooks4EE7C0[O, H]) H {
	unit := hooks.loadUnitArg()
	var nilHealth H
	var nilObject O
	if unit == nilObject {
		return nilHealth
	}

	health := hooks.loadHealth(unit)
	if health == nilHealth {
		return nilHealth
	}
	maximum := hooks.loadMaximumArg()
	hooks.storeMaximum(health, maximum)
	return health
}
