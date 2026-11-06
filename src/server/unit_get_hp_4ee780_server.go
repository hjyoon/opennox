package server

func unitGetHPNative4EE780(unit *Object) uint16 {
	return unitGetHP4EE780(unitGetHPHooks4EE780[*Object, *HealthData]{
		loadUnitArg: func() *Object {
			return unit
		},
		loadHealth: func(unit *Object) *HealthData {
			read := UnitHPRead4EE780{Unit: unit, Field: 0}
			if ObserveUnitHPRead4EE780 != nil {
				ObserveUnitHPRead4EE780(read)
			}
			health := unit.HealthData
			if ObserveUnitHPRead4EE780 != nil {
				read.Complete, read.ValueHealth = true, health
				ObserveUnitHPRead4EE780(read)
			}
			return health
		},
		loadCurrent: func(health *HealthData) uint16 {
			read := UnitHPRead4EE780{Health: health, Field: 1}
			if ObserveUnitHPRead4EE780 != nil {
				ObserveUnitHPRead4EE780(read)
			}
			value := health.Cur
			if ObserveUnitHPRead4EE780 != nil {
				read.Complete, read.Value = true, value
				ObserveUnitHPRead4EE780(read)
			}
			return value
		},
	})
}

// UnitGetHP4EE780 returns the native-width object's exact current-HP word,
// preserving the two nil gates from GAME.EXE 004EE780.
func UnitGetHP4EE780(unit *Object) uint16 {
	return unitGetHPNative4EE780(unit)
}
