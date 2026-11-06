package server

func unitGetMaxHPNative4EE7A0(unit *Object) uint16 {
	return unitGetMaxHP4EE7A0(unitGetMaxHPHooks4EE7A0[*Object, *HealthData]{
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
		loadMaximum: func(health *HealthData) uint16 {
			read := UnitHPRead4EE780{Health: health, Field: 2}
			if ObserveUnitHPRead4EE780 != nil {
				ObserveUnitHPRead4EE780(read)
			}
			value := health.Max
			if ObserveUnitHPRead4EE780 != nil {
				read.Complete, read.Value = true, value
				ObserveUnitHPRead4EE780(read)
			}
			return value
		},
	})
}

// UnitGetMaxHP4EE7A0 returns the native-width object's exact maximum-HP word,
// preserving the two nil gates from GAME.EXE 004EE7A0.
func UnitGetMaxHP4EE7A0(unit *Object) uint16 {
	return unitGetMaxHPNative4EE7A0(unit)
}
