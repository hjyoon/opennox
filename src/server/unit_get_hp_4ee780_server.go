package server

func unitGetHPNative4EE780(unit *Object) uint16 {
	return unitGetHP4EE780(unitGetHPHooks4EE780[*Object, *HealthData]{
		loadUnitArg: func() *Object {
			return unit
		},
		loadHealth: func(unit *Object) *HealthData {
			return unit.HealthData
		},
		loadCurrent: func(health *HealthData) uint16 {
			return health.Cur
		},
	})
}

// UnitGetHP4EE780 returns the native-width object's exact current-HP word,
// preserving the two nil gates from GAME.EXE 004EE780.
func UnitGetHP4EE780(unit *Object) uint16 {
	return unitGetHPNative4EE780(unit)
}
