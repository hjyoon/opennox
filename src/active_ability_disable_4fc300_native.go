package opennox

import "github.com/opennox/opennox/v1/server"

type activeAbilityDisableNativeDeps4FC300 struct {
	breakHarpoon      func(*server.Object, *server.Object)
	disableEnchant    func(*server.Object, server.EnchantID)
	reportActive      func(*server.Object, server.Ability, bool)
	loadExecHead      func() *server.ExecAbilityClass
	storeExecHead     func(*server.ExecAbilityClass)
	loadExecAllocator func() uintptr
	freeExec          func(uintptr, *server.ExecAbilityClass)
}

func activeAbilityDisableNative4FC300(
	unit *server.Object,
	ability server.Ability,
	deps activeAbilityDisableNativeDeps4FC300,
) {
	activeAbilityDisable4FC300(activeAbilityDisableHooks4FC300[
		*server.Object,
		*server.PlayerUpdateData,
		*server.ExecAbilityClass,
		uintptr,
	]{
		loadUnitArg: func() *server.Object {
			return unit
		},
		loadAbilityArg: func() server.Ability {
			return ability
		},
		loadUpdateData: func(unit *server.Object) *server.PlayerUpdateData {
			return (*server.PlayerUpdateData)(unit.UpdateData)
		},
		loadHarpoonBolt: func(update *server.PlayerUpdateData) *server.Object {
			return update.HarpoonBolt
		},
		breakHarpoon: deps.breakHarpoon,
		disableEnchant: func(unit *server.Object, enchant int32) {
			deps.disableEnchant(unit, server.EnchantID(enchant))
		},
		reportActive: func(unit *server.Object, ability server.Ability, active int32) {
			deps.reportActive(unit, ability, active != 0)
		},
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
		loadExecAllocator: deps.loadExecAllocator,
		freeExec:          deps.freeExec,
	})
}

// DisableAbility is the active native-width replacement for GAME.EXE
// 004FC300. Harpoon update data and its bolt are deliberately read directly:
// the original routine neither checks the unit class nor defends a nil update
// pointer before invoking the break callback.
//
//go:noinline
func (a *serverAbilities) DisableAbility(unit *server.Object, ability server.Ability) {
	activeAbilityDisableNative4FC300(unit, ability, activeAbilityDisableNativeDeps4FC300{
		breakHarpoon: a.harpoon.netHarpoonBreak,
		disableEnchant: func(unit *server.Object, enchant server.EnchantID) {
			asObjectS(unit).DisableEnchant(enchant)
		},
		reportActive:      a.netAbilReportActive,
		loadExecHead:      a.s.Abils.ExecHead,
		storeExecHead:     a.s.Abils.SetExecHead,
		loadExecAllocator: func() uintptr { return 0 },
		freeExec: func(_ uintptr, record *server.ExecAbilityClass) {
			*record = server.ExecAbilityClass{}
		},
	})
}
