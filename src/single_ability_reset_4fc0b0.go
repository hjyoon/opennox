package opennox

import "github.com/opennox/opennox/v1/server"

const (
	singleAbilityResetPlayerClass4FC0B0 = uint8(0x04)
	singleAbilityResetWarrior4FC0B0     = uint8(0)
	singleAbilityResetInactive4FC0B0    = int32(0)
)

type singleAbilityResetHooks4FC0B0[
	U comparable,
	D comparable,
	P comparable,
	R comparable,
	A comparable,
] struct {
	loadUnitArg       func() U
	loadUnitClassLow  func(U) uint8
	loadUpdateData    func(U) D
	loadPlayer        func(D) P
	loadPlayerClass   func(P) uint8
	loadAbilityArg    func() server.Ability
	loadPlayerIndex   func(P) uint8
	storeCooldown     func(uint8, server.Ability, int32)
	resetAbility      func(U, server.Ability)
	loadExecHead      func() R
	storeExecHead     func(R)
	loadExecUnit      func(R) U
	loadExecAbility   func(R) server.Ability
	loadExecNext      func(R) R
	loadExecPrev      func(R) R
	storeExecNext     func(R, R)
	storeExecPrev     func(R, R)
	reportActive      func(U, server.Ability, int32)
	loadExecAllocator func() A
	freeExec          func(A, R)
}

// singleAbilityReset4FC0B0 preserves GAME.EXE 004FC0B0. The unit argument is
// cached before the nil and Player-class gates. UpdateData, Player, and the
// exact Warrior class byte are then read without defensive nil checks. Only a
// Warrior reads the signed ability argument and live PlayerInd byte, clears
// the corresponding int32 cooldown cell, and emits the reset callback.
//
// The active-record head is deliberately loaded after that callback. On each
// record Unit and Next are cached immediately, and ability is read only after
// Unit matches. A matching report uses those cached Unit/ability values. The
// unlink that follows deliberately reloads live Next/Prev links, but traversal
// resumes through the Next cached before the report callback. Every matching
// record is released through the allocator loaded after unlinking. Active and
// deadline fields and ability bounds are intentionally not inspected.
func singleAbilityReset4FC0B0[
	U comparable,
	D comparable,
	P comparable,
	R comparable,
	A comparable,
](hooks singleAbilityResetHooks4FC0B0[U, D, P, R, A]) {
	var zeroUnit U
	unit := hooks.loadUnitArg()
	if unit == zeroUnit {
		return
	}
	if hooks.loadUnitClassLow(unit)&singleAbilityResetPlayerClass4FC0B0 == 0 {
		return
	}

	update := hooks.loadUpdateData(unit)
	player := hooks.loadPlayer(update)
	if hooks.loadPlayerClass(player) != singleAbilityResetWarrior4FC0B0 {
		return
	}

	ability := hooks.loadAbilityArg()
	index := hooks.loadPlayerIndex(player)
	hooks.storeCooldown(index, ability, 0)
	hooks.resetAbility(unit, ability)

	var zeroRecord R
	for record := hooks.loadExecHead(); record != zeroRecord; {
		recordUnit := hooks.loadExecUnit(record)
		cachedNext := hooks.loadExecNext(record)
		if recordUnit == unit {
			recordAbility := hooks.loadExecAbility(record)
			if recordAbility == ability {
				hooks.reportActive(recordUnit, recordAbility, singleAbilityResetInactive4FC0B0)

				liveNext := hooks.loadExecNext(record)
				if liveNext != zeroRecord {
					livePrev := hooks.loadExecPrev(record)
					hooks.storeExecPrev(liveNext, livePrev)
				}
				livePrev := hooks.loadExecPrev(record)
				if livePrev != zeroRecord {
					liveNext = hooks.loadExecNext(record)
					hooks.storeExecNext(livePrev, liveNext)
				} else {
					liveNext = hooks.loadExecNext(record)
					hooks.storeExecHead(liveNext)
				}

				allocator := hooks.loadExecAllocator()
				hooks.freeExec(allocator, record)
			}
		}
		record = cachedNext
	}
}
