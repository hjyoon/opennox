package server

type unitBuffPowerNativeDeps4FF570 struct {
	loadBuffArg func(int32) int32
	loadUnitArg func(*Object) *Object
	loadPower   func(*Object, int32) uint8
}

func unitBuffPowerNative4FF570(
	unit *Object,
	buff int32,
	deps unitBuffPowerNativeDeps4FF570,
) uint8 {
	return UnitBuffPower4FF570(
		UnitBuffPowerHooks4FF570[*Object]{
			LoadBuffArg: func() int32 {
				return deps.loadBuffArg(buff)
			},
			LoadUnitArg: func() *Object {
				return deps.loadUnitArg(unit)
			},
			LoadPower: deps.loadPower,
		},
	)
}

func unitBuffPowerServerDeps4FF570() unitBuffPowerNativeDeps4FF570 {
	return unitBuffPowerNativeDeps4FF570{
		loadBuffArg: func(buff int32) int32 {
			return buff
		},
		loadUnitArg: func(unit *Object) *Object {
			return unit
		},
		loadPower: func(unit *Object, buff int32) uint8 {
			return unit.BuffsPower[int(buff)]
		},
	}
}

// UnitBuffPower4FF570 binds GAME.EXE 004FF570 to the native-width Object.
// Buff remains an exact signed dword, BuffsPower elements remain exact bytes,
// and only that byte forms the result. The oracle has no null or range guard;
// this binding deliberately lets Go turn invalid native accesses into
// deterministic memory-safety faults instead of returning a fabricated zero.
//
//go:noinline
func (obj *Object) UnitBuffPower4FF570(buff int32) uint8 {
	return unitBuffPowerNative4FF570(obj, buff, unitBuffPowerServerDeps4FF570())
}
