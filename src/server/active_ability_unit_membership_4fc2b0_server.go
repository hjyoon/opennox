package server

type activeAbilityUnitMembershipNativeDeps4FC2B0 struct {
	loadExecHead func() *ExecAbilityClass
}

// activeAbilityUnitMembershipNative4FC2B0 binds the executable's fixed-width
// fields to native Go pointers. Object, Player, and ExecAbilityClass pointers
// are never narrowed through the 32-bit integer slots used by GAME.EXE.
func activeAbilityUnitMembershipNative4FC2B0(
	unit *Object,
	deps activeAbilityUnitMembershipNativeDeps4FC2B0,
) int32 {
	return activeAbilityUnitMembership4FC2B0(activeAbilityUnitMembershipHooks4FC2B0[
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
				panic("004FC2B0: nil Player")
			}
			return uint8(player.Info().PlayerClass())
		},
		loadExecHead: deps.loadExecHead,
		loadExecUnit: func(record *ExecAbilityClass) *Object {
			return record.Unit
		},
		loadExecNext: func(record *ExecAbilityClass) *ExecAbilityClass {
			return record.Next
		},
	})
}

// IsAnyActive preserves GAME.EXE 004FC2B0. It admits only Player-class
// objects whose available Player record is exactly Warrior, then tests whether
// any global execution-list record has the same native unit pointer. Ability,
// Active, deadline, and Prev are ignored.
//
//go:noinline
func (a *serverAbilities) IsAnyActive(unit *Object) bool {
	return activeAbilityUnitMembershipNative4FC2B0(
		unit,
		activeAbilityUnitMembershipNativeDeps4FC2B0{
			loadExecHead: func() *ExecAbilityClass {
				return a.execList
			},
		},
	) != 0
}
