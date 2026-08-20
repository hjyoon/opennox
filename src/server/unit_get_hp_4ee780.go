package server

type unitGetHPHooks4EE780[O, H any] struct {
	loadUnitArg func() O
	loadHealth  func(O) H
	loadCurrent func(H) uint16
}

// unitGetHP4EE780 preserves GAME.EXE 004EE780. The object and HealthData
// pointers are each loaded exactly once. Either nil gate returns zero without
// performing the following read; otherwise the exact current-HP word from the
// cached HealthData record is returned.
func unitGetHP4EE780[O, H comparable](hooks unitGetHPHooks4EE780[O, H]) uint16 {
	unit := hooks.loadUnitArg()
	var nilObject O
	if unit == nilObject {
		return 0
	}

	health := hooks.loadHealth(unit)
	var nilHealth H
	if health == nilHealth {
		return 0
	}
	return hooks.loadCurrent(health)
}
