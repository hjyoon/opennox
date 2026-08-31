package opennox

import "github.com/opennox/opennox/v1/server"

type allAbilityCancelNativeDeps4FC180 struct {
	storeCooldown     func(uint8, server.Ability, int32)
	resetAbilities    func(*server.Object, server.Ability)
	loadExecHead      func() *server.ExecAbilityClass
	storeExecHead     func(*server.ExecAbilityClass)
	reportActive      func(*server.Object, server.Ability, bool)
	loadExecAllocator func() uintptr
	freeExec          func(uintptr, *server.ExecAbilityClass)
}

func allAbilityCancelNative4FC180(
	unit *server.Object,
	deps allAbilityCancelNativeDeps4FC180,
) {
	allAbilityCancel4FC180(allAbilityCancelHooks4FC180[
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
				panic("004FC180: nil Player")
			}
			return uint8(player.PlayerClass())
		},
		loadPlayerIndex: func(player *server.Player) uint8 {
			return player.PlayerInd
		},
		storeCooldown:  deps.storeCooldown,
		resetAbilities: deps.resetAbilities,
		loadExecHead:   deps.loadExecHead,
		storeExecHead:  deps.storeExecHead,
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

// CancelAbilities is the active native-width replacement for GAME.EXE
// 004FC180. It clears only Warrior ability slots 1 through 5; slot zero stays
// untouched. The generic model retains the executable's raw byte index, while
// this live binding deliberately lets Go's fixed cooldown matrix trap invalid
// PlayerInd values instead of reproducing an out-of-bounds write.
//
//go:noinline
func (a *serverAbilities) CancelAbilities(unit *server.Object) {
	allAbilityCancelNative4FC180(unit, allAbilityCancelNativeDeps4FC180{
		storeCooldown:     a.s.Abils.SetPlayerAbilityCooldownAt,
		resetAbilities:    a.netAbilReset,
		loadExecHead:      a.s.Abils.ExecHead,
		storeExecHead:     a.s.Abils.SetExecHead,
		reportActive:      a.netAbilReportActive,
		loadExecAllocator: func() uintptr { return 0 },
		freeExec: func(_ uintptr, record *server.ExecAbilityClass) {
			*record = server.ExecAbilityClass{}
		},
	})
}
