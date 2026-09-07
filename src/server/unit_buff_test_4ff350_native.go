package server

type unitBuffTestNativeDeps4FF350 struct {
	loadBuffArg func() int32
	loadBuffs   func(*Object) uint32
}

func unitBuffTestNative4FF350(
	unit *Object,
	deps unitBuffTestNativeDeps4FF350,
) int32 {
	return UnitBuffTest4FF350(
		unit,
		UnitBuffTestHooks4FF350[*Object]{
			LoadBuffArg: deps.loadBuffArg,
			LoadBuffs:   deps.loadBuffs,
		},
	)
}

func unitBuffTestServerDeps4FF350(buff int32) unitBuffTestNativeDeps4FF350 {
	return unitBuffTestNativeDeps4FF350{
		loadBuffArg: func() int32 {
			return buff
		},
		loadBuffs: func(unit *Object) uint32 {
			return unit.Buffs
		},
	}
}

// UnitBuffTest4FF350 binds GAME.EXE 004FF350 to the native-width Object.
// The object identity is never converted through a PE32 integer, Buffs stays
// an exact uint32 word, and the signed dword buff argument retains the
// original x86 modulo-32 shift behavior.
//
//go:noinline
func (obj *Object) UnitBuffTest4FF350(buff int32) int32 {
	return unitBuffTestNative4FF350(obj, unitBuffTestServerDeps4FF350(buff))
}
