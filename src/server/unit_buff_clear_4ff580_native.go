package server

// UnitBuffClearRuntime4FF580 supplies the protection-table effect that remains
// owned by the legacy layer. SetBuffFlags invokes it only for player objects.
type UnitBuffClearRuntime4FF580 struct {
	ResetPlayerProtection func(*Player, uint32)
}

type unitBuffClearNativeDeps4FF580 struct {
	loadUnitArg   func(*Object) *Object
	setBuffFlags  func(*Object, uint32)
	storeDuration func(*Object, int32, uint16)
	storePower    func(*Object, int32, uint8)
}

func unitBuffClearNative4FF580(unit *Object, deps unitBuffClearNativeDeps4FF580) {
	UnitBuffClear4FF580(UnitBuffClearHooks4FF580[*Object]{
		LoadUnitArg: func() *Object {
			return deps.loadUnitArg(unit)
		},
		SetBuffFlags:  deps.setBuffFlags,
		StoreDuration: deps.storeDuration,
		StorePower:    deps.storePower,
	})
}

func unitBuffClearServerDeps4FF580(runtime UnitBuffClearRuntime4FF580) unitBuffClearNativeDeps4FF580 {
	return unitBuffClearNativeDeps4FF580{
		loadUnitArg: func(unit *Object) *Object {
			return unit
		},
		setBuffFlags: func(unit *Object, flags uint32) {
			unit.SetBuffFlags(flags, runtime.ResetPlayerProtection)
		},
		storeDuration: func(unit *Object, buff int32, value uint16) {
			unit.BuffsDur[buff] = value
		},
		storePower: func(unit *Object, buff int32, value uint8) {
			unit.BuffsPower[buff] = value
		},
	}
}

// UnitBuffClear4FF580 binds GAME.EXE 004FF580 to native-width Object state.
// SetBuffFlags runs before all 32 word/byte pairs are cleared, preserving both
// player-protection and synchronization side effects. The original has no
// null guard, so a nil native object deliberately faults instead of silently
// succeeding.
//
//go:noinline
func (obj *Object) UnitBuffClear4FF580(runtime UnitBuffClearRuntime4FF580) {
	unitBuffClearNative4FF580(obj, unitBuffClearServerDeps4FF580(runtime))
}
