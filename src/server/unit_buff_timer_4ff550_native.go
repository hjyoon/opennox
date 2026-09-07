package server

type unitBuffTimerNativeDeps4FF550 struct {
	loadBuffArg  func(int32) int32
	loadUnitArg  func(*Object) *Object
	loadDuration func(*Object, int32) uint16
}

func unitBuffTimerNative4FF550(
	unit *Object,
	buff int32,
	deps unitBuffTimerNativeDeps4FF550,
) uint32 {
	return UnitBuffTimer4FF550(
		UnitBuffTimerHooks4FF550[*Object]{
			LoadBuffArg: func() int32 {
				return deps.loadBuffArg(buff)
			},
			LoadUnitArg: func() *Object {
				return deps.loadUnitArg(unit)
			},
			LoadDuration: deps.loadDuration,
		},
	)
}

func unitBuffTimerServerDeps4FF550() unitBuffTimerNativeDeps4FF550 {
	return unitBuffTimerNativeDeps4FF550{
		loadBuffArg: func(buff int32) int32 {
			return buff
		},
		loadUnitArg: func(unit *Object) *Object {
			return unit
		},
		loadDuration: func(unit *Object, buff int32) uint16 {
			return unit.BuffsDur[int(buff)]
		},
	}
}

// UnitBuffTimer4FF550 binds GAME.EXE 004FF550 to the native-width Object.
// Buff remains an exact signed dword, BuffsDur elements remain exact words,
// and the returned word is zero-extended. The oracle has no null or range
// guard; this binding deliberately lets Go turn invalid native accesses into
// deterministic memory-safety faults instead of returning a fabricated zero.
//
//go:noinline
func (obj *Object) UnitBuffTimer4FF550(buff int32) uint32 {
	return unitBuffTimerNative4FF550(obj, buff, unitBuffTimerServerDeps4FF550())
}
