package server

// Sub4FC030 is the native-width replacement for GAME.EXE 004FC030. Object and
// record links remain native pointers while the deadline, current frame, and
// result retain the executable's exact 32-bit widths.
//
//go:noinline
func (a *serverAbilities) Sub4FC030(unit *Object, ability Ability) int32 {
	return activeAbilityDuration4FC030(unit, ability, activeAbilityDurationHooks4FC030[
		*ExecAbilityClass,
		*Object,
	]{
		loadHead: func() *ExecAbilityClass {
			return a.execList
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
		loadDeadline: func(record *ExecAbilityClass) uint32 {
			return record.Frame
		},
		loadFrame: a.s.Frame,
	})
}
