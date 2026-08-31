package opennox

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

const (
	playerExecuteAbilityPlayerClass4FBB70 = uint8(0x04)
	playerExecuteAbilityWarrior4FBB70     = uint8(0x00)

	playerExecuteAbilityCTF4FBB70    = uint32(0x0020)
	playerExecuteAbilityCoop4FBB70   = uint32(0x0800)
	playerExecuteAbilityQuest4FBB70  = uint32(0x1000)
	playerExecuteAbilityOnline4FBB70 = uint32(0x2000)

	playerExecuteAbilityBlockedState4FBB70 = uint8(12)
	playerExecuteAbilityReportKind4FBB70   = uint8(2)
	playerExecuteAbilityFizzleSound4FBB70  = int32(0xe7)
	playerExecuteAbilityStartSound4FBB70   = int32(0)

	playerExecuteAbilityReportBusy4FBB70       = int32(2)
	playerExecuteAbilityReportUnknown4FBB70    = int32(3)
	playerExecuteAbilityReportFlag4FBB70       = int32(5)
	playerExecuteAbilityReportMovement4FBB70   = int32(6)
	playerExecuteAbilityReportOverweight4FBB70 = int32(7)
)

type playerExecuteAbilityHooks4FBB70[
	U comparable,
	D any,
	P any,
	I comparable,
	R comparable,
] struct {
	loadFlags          func(U) object.Flags
	loadClassLow       func(U) uint8
	loadUpdateData     func(U) D
	loadPlayer         func(D) P
	loadPlayerClassLow func(P) uint8
	gameFlag           func(uint32) int32
	firstItem          func(U) I
	nextItem           func(I) I
	loadItemClass      func(I) object.Class
	isActive           func(U, server.Ability) int32
	isActiveVal        func(U, server.Ability) int32
	loadState          func(D) uint8
	loadSpellLevel     func(P, server.Ability) uint32
	loadBerserkBlock   func(P) uint32
	loadPlayerIndex    func(P) uint8
	reportText         func(uint8, uint8, *int32)
	loadCooldown       func(uint8, server.Ability) int32
	storeCooldown      func(uint8, server.Ability, int32)
	loadDelay          func(server.Ability) int32
	reportState        func(U, server.Ability, uint8)
	loadDuration       func(server.Ability) int32
	allocExec          func() R
	loadFrame          func() uint32
	storeExecUnit      func(R, U)
	storeExecAbility   func(R, server.Ability)
	storeExecFrame     func(R, uint32)
	loadExecHead       func() R
	storeExecNext      func(R, R)
	storeExecActive    func(R, uint32)
	storeExecPrev      func(R, R)
	storeExecHead      func(R)
	invoke             func(U, server.Ability)
	loadSound          func(server.Ability, int32) int32
	audio              func(int32, U, int32, uint32)
}

// playerExecuteAbility4FBB70 preserves GAME.EXE 004FBB70. The PE32 routine
// explicitly accepts only signed ability IDs 1..5 after a nil-unit guard. It
// caches UpdateData once, but deliberately reloads UpdateData.Player for
// failure reports, the spell-level snapshot, and the Berserk overweight gate.
// The cooldown read and write each reload the byte index from the same spell-
// level Player snapshot.
//
// Game predicates retain their original integer tests: CTF and Coop accept
// any nonzero result, while the Online fast path requires exactly one. The
// first delay result is stored and a second independent delay result controls
// the state report. Positive-duration execution records preserve the original
// store and live-head reload order. No defensive nil checks are introduced
// beyond the unit check present in the executable.
func playerExecuteAbility4FBB70[
	U comparable,
	D any,
	P any,
	I comparable,
	R comparable,
](
	unit U,
	ability server.Ability,
	hooks playerExecuteAbilityHooks4FBB70[U, D, P, I, R],
) {
	var zeroUnit U
	if unit == zeroUnit {
		return
	}
	if ability < server.AbilityBerserk || ability > server.AbilityInfravis {
		return
	}
	if hooks.loadFlags(unit).HasAny(object.FlagDestroyed | object.FlagDead) {
		return
	}
	if hooks.loadClassLow(unit)&playerExecuteAbilityPlayerClass4FBB70 == 0 {
		return
	}

	update := hooks.loadUpdateData(unit)
	player := hooks.loadPlayer(update)
	if hooks.loadPlayerClassLow(player) != playerExecuteAbilityWarrior4FBB70 {
		return
	}

	report := func(code int32, fizzle bool) {
		livePlayer := hooks.loadPlayer(update)
		index := hooks.loadPlayerIndex(livePlayer)
		hooks.reportText(index, playerExecuteAbilityReportKind4FBB70, &code)
		if fizzle {
			hooks.audio(playerExecuteAbilityFizzleSound4FBB70, unit, 0, 0)
		}
	}

	if hooks.gameFlag(playerExecuteAbilityCTF4FBB70) != 0 && ability == server.AbilityBerserk {
		var zeroItem I
		for item := hooks.firstItem(unit); item != zeroItem; item = hooks.nextItem(item) {
			if hooks.loadItemClass(item).Has(object.ClassFlag) {
				report(playerExecuteAbilityReportFlag4FBB70, true)
				return
			}
		}
	}

	switch ability {
	case server.AbilityBerserk:
		if hooks.isActiveVal(unit, server.AbilityWarcry) != 0 ||
			hooks.isActive(unit, server.AbilityHarpoon) != 0 {
			report(playerExecuteAbilityReportBusy4FBB70, true)
			return
		}
	case server.AbilityWarcry:
		if hooks.isActive(unit, server.AbilityBerserk) != 0 ||
			hooks.isActive(unit, server.AbilityHarpoon) != 0 {
			report(playerExecuteAbilityReportBusy4FBB70, true)
			return
		}
	case server.AbilityHarpoon:
		if hooks.isActiveVal(unit, server.AbilityWarcry) != 0 ||
			hooks.isActive(unit, server.AbilityBerserk) != 0 {
			report(playerExecuteAbilityReportBusy4FBB70, true)
			return
		}
	}
	if hooks.isActive(unit, ability) != 0 {
		report(playerExecuteAbilityReportBusy4FBB70, true)
		return
	}

	if hooks.loadState(update) == playerExecuteAbilityBlockedState4FBB70 {
		report(playerExecuteAbilityReportMovement4FBB70, true)
		return
	}
	if hooks.gameFlag(playerExecuteAbilityCoop4FBB70) == 0 &&
		hooks.loadFlags(unit).Has(object.FlagAirborne) {
		report(playerExecuteAbilityReportMovement4FBB70, true)
		return
	}

	player = hooks.loadPlayer(update)
	skipLevel := false
	if hooks.gameFlag(playerExecuteAbilityOnline4FBB70) == 1 {
		skipLevel = hooks.gameFlag(playerExecuteAbilityQuest4FBB70) == 0
	}
	if !skipLevel && hooks.loadSpellLevel(player, ability) == 0 {
		report(playerExecuteAbilityReportUnknown4FBB70, true)
		return
	}

	if ability == server.AbilityBerserk {
		livePlayer := hooks.loadPlayer(update)
		if hooks.loadBerserkBlock(livePlayer) == 1 {
			report(playerExecuteAbilityReportOverweight4FBB70, false)
			return
		}
	}

	index := hooks.loadPlayerIndex(player)
	if hooks.loadCooldown(index, ability) != 0 {
		report(playerExecuteAbilityReportBusy4FBB70, true)
		return
	}
	delay := hooks.loadDelay(ability)
	index = hooks.loadPlayerIndex(player)
	hooks.storeCooldown(index, ability, delay)
	if hooks.loadDelay(ability) != 0 {
		hooks.reportState(unit, ability, 0)
	}

	duration := hooks.loadDuration(ability)
	if duration > 0 {
		var zeroExec R
		exec := hooks.allocExec()
		if exec != zeroExec {
			frame := hooks.loadFrame()
			hooks.storeExecUnit(exec, unit)
			hooks.storeExecAbility(exec, ability)
			hooks.storeExecFrame(exec, frame+uint32(duration))
			head := hooks.loadExecHead()
			hooks.storeExecNext(exec, head)
			hooks.storeExecActive(exec, 1)
			hooks.storeExecPrev(exec, zeroExec)
			head = hooks.loadExecHead()
			if head != zeroExec {
				hooks.storeExecPrev(head, exec)
			}
			hooks.storeExecHead(exec)
		}
	}

	hooks.invoke(unit, ability)
	soundID := hooks.loadSound(ability, playerExecuteAbilityStartSound4FBB70)
	hooks.audio(soundID, unit, 0, 0)
}
