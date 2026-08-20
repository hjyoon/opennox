package server

// UnitHPSetOnMaxRuntime4EE6F0 supplies the still-legacy 004E4560 HP setter.
// The restored 004EE6F0 routine owns every pointer-bearing read itself; the
// setter remains an explicit dependency until its separate widening audit.
type UnitHPSetOnMaxRuntime4EE6F0 struct {
	SetHP func(*Object, uint16)
}

type unitHPSetOnMaxNativeDeps4EE6F0 struct {
	setHP       func(*Object, uint16)
	informOwner func(*Object)
}

func unitHPSetOnMaxNative4EE6F0(unit *Object, deps unitHPSetOnMaxNativeDeps4EE6F0) {
	unitHPSetOnMax4EE6F0(unitHPSetOnMaxHooks4EE6F0[*Object, *HealthData]{
		loadUnitArg: func() *Object {
			return unit
		},
		loadHealth: func(unit *Object) *HealthData {
			return unit.HealthData
		},
		loadMaximum: func(health *HealthData) uint16 {
			return health.Max
		},
		setHP: deps.setHP,
		loadCurrent: func(health *HealthData) uint16 {
			return health.Cur
		},
		storeField2: func(health *HealthData, value uint16) {
			health.Field2 = value
		},
		loadClassLow: func(unit *Object) uint8 {
			return uint8(unit.ObjClass)
		},
		informOwner: deps.informOwner,
	})
}

// UnitHPSetOnMax4EE6F0 restores maximum HP, synchronizes Field2 from the live
// HealthData record, and reports live Monster HP through native-width objects.
func (s *Server) UnitHPSetOnMax4EE6F0(unit *Object, runtime UnitHPSetOnMaxRuntime4EE6F0) {
	unitHPSetOnMaxNative4EE6F0(unit, unitHPSetOnMaxNativeDeps4EE6F0{
		setHP: runtime.SetHP,
		informOwner: func(unit *Object) {
			s.MobInformOwnerHP4EE4C0(unit)
		},
	})
}
