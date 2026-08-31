package opennox

import "github.com/opennox/opennox/v1/server"

type singleAbilityResetNativeDeps4FC0B0 struct {
	storeCooldown     func(uint8, server.Ability, int32)
	resetAbility      func(*server.Object, server.Ability)
	loadExecHead      func() *server.ExecAbilityClass
	storeExecHead     func(*server.ExecAbilityClass)
	reportActive      func(*server.Object, server.Ability, bool)
	loadExecAllocator func() uintptr
	freeExec          func(uintptr, *server.ExecAbilityClass)
}

func singleAbilityResetNative4FC0B0(
	unit *server.Object,
	ability server.Ability,
	deps singleAbilityResetNativeDeps4FC0B0,
) {
	singleAbilityReset4FC0B0(singleAbilityResetHooks4FC0B0[
		*server.Object,
		*server.PlayerUpdateData,
		*server.Player,
		*server.ExecAbilityClass,
		uintptr,
	]{
		loadUnitArg: func() *server.Object {
			return unit
		},
		loadUnitClassLow: func(unit *server.Object) uint8 {
			return uint8(unit.ObjClass)
		},
		loadUpdateData: func(unit *server.Object) *server.PlayerUpdateData {
			return (*server.PlayerUpdateData)(unit.UpdateData)
		},
		loadPlayer: func(update *server.PlayerUpdateData) *server.Player {
			return update.Player
		},
		loadPlayerClass: func(player *server.Player) uint8 {
			// PlayerClass is intentionally nil-safe elsewhere in the port. The
			// original direct Player+2251 byte load is not, so retain its fault.
			if player == nil {
				panic("004FC0B0: nil Player")
			}
			return uint8(player.PlayerClass())
		},
		loadAbilityArg: func() server.Ability {
			return ability
		},
		loadPlayerIndex: func(player *server.Player) uint8 {
			return player.PlayerInd
		},
		storeCooldown: deps.storeCooldown,
		resetAbility:  deps.resetAbility,
		loadExecHead:  deps.loadExecHead,
		storeExecHead: deps.storeExecHead,
		loadExecUnit: func(record *server.ExecAbilityClass) *server.Object {
			return record.Unit
		},
		loadExecAbility: func(record *server.ExecAbilityClass) server.Ability {
			return record.Abil
		},
		loadExecNext: func(record *server.ExecAbilityClass) *server.ExecAbilityClass {
			return record.Next
		},
		loadExecPrev: func(record *server.ExecAbilityClass) *server.ExecAbilityClass {
			return record.Prev
		},
		storeExecNext: func(record, next *server.ExecAbilityClass) {
			record.Next = next
		},
		storeExecPrev: func(record, prev *server.ExecAbilityClass) {
			record.Prev = prev
		},
		reportActive: func(unit *server.Object, ability server.Ability, active int32) {
			deps.reportActive(unit, ability, active != 0)
		},
		loadExecAllocator: deps.loadExecAllocator,
		freeExec:          deps.freeExec,
	})
}

// ResetAbility is the active native-width replacement for GAME.EXE 004FC0B0.
// The original callers only pass the Warrior Berserk ability. The generic
// model retains the executable's unbounded signed index semantics; this live
// binding deliberately lets Go's fixed cooldown matrix trap invalid indices
// instead of reproducing an out-of-bounds write.
//
//go:noinline
func (a *serverAbilities) ResetAbility(unit *server.Object, ability server.Ability) {
	singleAbilityResetNative4FC0B0(unit, ability, singleAbilityResetNativeDeps4FC0B0{
		storeCooldown:     a.s.Abils.SetPlayerAbilityCooldownAt,
		resetAbility:      a.netAbilReset,
		loadExecHead:      a.s.Abils.ExecHead,
		storeExecHead:     a.s.Abils.SetExecHead,
		reportActive:      a.netAbilReportActive,
		loadExecAllocator: func() uintptr { return 0 },
		freeExec: func(_ uintptr, record *server.ExecAbilityClass) {
			*record = server.ExecAbilityClass{}
		},
	})
}
