package server

func unitGetMaxHPNative4EE7A0(unit *Object) uint16 {
	return unitGetMaxHP4EE7A0(unitGetMaxHPHooks4EE7A0[*Object, *HealthData]{
		loadUnitArg: func() *Object {
			return unit
		},
		loadHealth: func(unit *Object) *HealthData {
			return unit.HealthData
		},
		loadMaximum: func(health *HealthData) uint16 {
			return health.Max
		},
	})
}

// UnitGetMaxHP4EE7A0 returns the native-width object's exact maximum-HP word,
// preserving the two nil gates from GAME.EXE 004EE7A0.
func UnitGetMaxHP4EE7A0(unit *Object) uint16 {
	return unitGetMaxHPNative4EE7A0(unit)
}
