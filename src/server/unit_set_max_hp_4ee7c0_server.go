package server

func unitSetMaxHPNative4EE7C0(unit *Object, maximum uint16) *HealthData {
	return unitSetMaxHP4EE7C0(unitSetMaxHPHooks4EE7C0[*Object, *HealthData]{
		loadUnitArg: func() *Object {
			return unit
		},
		loadHealth: func(unit *Object) *HealthData {
			return unit.HealthData
		},
		loadMaximumArg: func() uint16 {
			return maximum
		},
		storeMaximum: func(health *HealthData, maximum uint16) {
			health.Max = maximum
		},
	})
}

// UnitSetMaxHP4EE7C0 stores the exact maximum-HP word through the
// native-width object's cached HealthData pointer and returns that pointer.
// Either nil gate returns nil without performing the following access.
func UnitSetMaxHP4EE7C0(unit *Object, maximum uint16) *HealthData {
	return unitSetMaxHPNative4EE7C0(unit, maximum)
}
