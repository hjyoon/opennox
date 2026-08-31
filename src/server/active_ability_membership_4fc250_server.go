package server

type activeAbilityMembershipNativeDeps4FC250 struct {
	loadExecHead func() *ExecAbilityClass
}

// activeAbilityMembershipNative4FC250 binds the executable's fixed-width
// fields to native Go pointers. In particular, it does not turn Object or
// ExecAbilityClass pointers into the 32-bit integer slots used by GAME.EXE.
func activeAbilityMembershipNative4FC250(
	unit *Object,
	ability Ability,
	deps activeAbilityMembershipNativeDeps4FC250,
) int32 {
	return activeAbilityMembership4FC250(activeAbilityMembershipHooks4FC250[
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
			// Player.PlayerClass deliberately treats nil as Warrior. The
			// executable instead dereferences Player+2251 and therefore faults.
			if player == nil {
				panic("004FC250: nil Player")
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
	})
}

// IsActive preserves GAME.EXE 004FC250. It admits only Player-class objects
// whose available Player record is exactly Warrior, then tests membership in
// the global execution list. Record Active, deadline, and Prev are ignored.
// Object and list links remain native pointers and Ability remains signed
// 32-bit across the Go and C boundaries.
//
//go:noinline
func (a *serverAbilities) IsActive(unit *Object, ability Ability) bool {
	return activeAbilityMembershipNative4FC250(
		unit,
		ability,
		activeAbilityMembershipNativeDeps4FC250{
			loadExecHead: func() *ExecAbilityClass {
				return a.execList
			},
		},
	) != 0
}
