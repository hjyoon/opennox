package server

type activeAbilityValueNativeDeps4FC3E0 struct {
	loadExecHead func() *ExecAbilityClass
}

// activeAbilityValueNative4FC3E0 binds the executable's fixed-width fields to
// native Go pointers. Object, PlayerUpdateData, Player, and ExecAbilityClass
// pointers are never narrowed to the PE32 integer slots used by GAME.EXE.
func activeAbilityValueNative4FC3E0(
	unit *Object,
	ability Ability,
	deps activeAbilityValueNativeDeps4FC3E0,
) int32 {
	return activeAbilityValue4FC3E0(activeAbilityValueHooks4FC3E0[
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
				panic("004FC3E0: nil Player")
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
		loadExecActive: func(record *ExecAbilityClass) uint32 {
			return record.Active
		},
	})
}

// ActiveValue4FC3E0 preserves GAME.EXE 004FC3E0. It admits only Player-class
// objects whose available Player record is exactly Warrior, then returns the
// complete 32-bit Active field of the first matching global execution record.
// Object and list links remain native pointers and Ability remains signed
// 32-bit across Go and C boundaries.
//
//go:noinline
func (a *serverAbilities) ActiveValue4FC3E0(unit *Object, ability Ability) int32 {
	return activeAbilityValueNative4FC3E0(
		unit,
		ability,
		activeAbilityValueNativeDeps4FC3E0{
			loadExecHead: func() *ExecAbilityClass {
				return a.execList
			},
		},
	)
}

func (a *serverAbilities) IsActiveVal(unit *Object, ability Ability) bool {
	return a.ActiveValue4FC3E0(unit, ability) != 0
}
