package server

type unitGetMaxHPHooks4EE7A0[O, H any] struct {
	loadUnitArg func() O
	loadHealth  func(O) H
	loadMaximum func(H) uint16
}

// unitGetMaxHP4EE7A0 preserves GAME.EXE 004EE7A0. The object and HealthData
// pointers are each loaded exactly once. Either nil gate returns zero without
// performing the following read; otherwise the exact maximum-HP word at
// HealthData offset four is returned from the cached record.
func unitGetMaxHP4EE7A0[O, H comparable](hooks unitGetMaxHPHooks4EE7A0[O, H]) uint16 {
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
	return hooks.loadMaximum(health)
}
