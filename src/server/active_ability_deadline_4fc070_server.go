package server

// Sub4FC070 is the native-width replacement for GAME.EXE 004FC070. Object and
// record links remain native pointers while the ability, delta, deadline, and
// current frame retain the executable's exact signed or unsigned 32-bit widths.
//
//go:noinline
func (a *serverAbilities) Sub4FC070(unit *Object, ability Ability, delta int32) {
	activeAbilityDeadline4FC070(activeAbilityDeadlineHooks4FC070[
		*ExecAbilityClass,
		*Object,
	]{
		loadHead: func() *ExecAbilityClass {
			return a.execList
		},
		loadAbilityArg: func() Ability {
			return ability
		},
		loadUnitArg: func() *Object {
			return unit
		},
		loadUnit: func(record *ExecAbilityClass) *Object {
			return record.Unit
		},
		loadAbility: func(record *ExecAbilityClass) Ability {
			return record.Abil
		},
		loadNext: func(record *ExecAbilityClass) *ExecAbilityClass {
			return record.Next
		},
		loadDeltaArg: func() int32 {
			return delta
		},
		loadFrame: a.s.Frame,
		storeDeadline: func(record *ExecAbilityClass, deadline uint32) {
			record.Frame = deadline
		},
	})
}
