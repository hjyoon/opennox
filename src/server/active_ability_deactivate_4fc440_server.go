package server

type activeAbilityDeactivateNativeDeps4FC440 struct {
	loadExecHead func() *ExecAbilityClass
}

// activeAbilityDeactivateNative4FC440 binds the executable's fixed-width
// fields to native Go pointers. Object, PlayerUpdateData, Player, and
// ExecAbilityClass pointers are never narrowed to the PE32 integer slots used
// by GAME.EXE.
func activeAbilityDeactivateNative4FC440(
	unit *Object,
	ability Ability,
	deps activeAbilityDeactivateNativeDeps4FC440,
) {
	activeAbilityDeactivate4FC440(activeAbilityDeactivateHooks4FC440[
		*Object,
		*PlayerUpdateData,
		*Player,
		*ExecAbilityClass,
	]{
		loadUnitArg: func() *Object {
			return unit
		},
		loadUnitClassLow: func(unit *Object) uint8 {
			return uint8(unit.ObjClass)
		},
		loadUpdateData: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadPlayerClass: func(player *Player) uint8 {
			// Player.PlayerClass deliberately treats nil as Warrior. GAME.EXE
			// dereferences Player+2251 here and therefore faults instead.
			if player == nil {
				panic("004FC440: nil Player")
			}
			return uint8(player.Info().PlayerClass())
		},
		loadExecHead: deps.loadExecHead,
		loadAbilityArg: func() Ability {
			return ability
		},
		loadExecUnit: func(record *ExecAbilityClass) *Object {
			return record.Unit
		},
		loadExecNext: func(record *ExecAbilityClass) *ExecAbilityClass {
			return record.Next
		},
		loadExecAbility: func(record *ExecAbilityClass) Ability {
			return record.Abil
		},
		storeExecActive: func(record *ExecAbilityClass, active uint32) {
			record.Active = active
		},
	})
}

// Sub4FC440 preserves GAME.EXE 004FC440. It admits only Player-class objects
// whose available Player record is exactly Warrior, then overwrites the
// complete 32-bit Active field of the first matching global execution record
// with zero. Object and list links remain native pointers and Ability remains
// signed 32-bit across Go and C boundaries.
//
//go:noinline
func (a *serverAbilities) Sub4FC440(unit *Object, ability Ability) {
	activeAbilityDeactivateNative4FC440(
		unit,
		ability,
		activeAbilityDeactivateNativeDeps4FC440{
			loadExecHead: func() *ExecAbilityClass {
				return a.execList
			},
		},
	)
}
