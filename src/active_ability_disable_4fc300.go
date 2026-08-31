package opennox

import "github.com/opennox/opennox/v1/server"

const (
	activeAbilityDisableSneakEnchant4FC300 = int32(31)
	activeAbilityDisableInactive4FC300     = int32(0)
)

type activeAbilityDisableHooks4FC300[
	U comparable,
	D comparable,
	R comparable,
	A comparable,
] struct {
	loadUnitArg       func() U
	loadAbilityArg    func() server.Ability
	loadUpdateData    func(U) D
	loadHarpoonBolt   func(D) U
	breakHarpoon      func(U, U)
	disableEnchant    func(U, int32)
	reportActive      func(U, server.Ability, int32)
	loadExecHead      func() R
	storeExecHead     func(R)
	loadExecUnit      func(R) U
	loadExecAbility   func(R) server.Ability
	loadExecNext      func(R) R
	loadExecPrev      func(R) R
	storeExecNext     func(R, R)
	storeExecPrev     func(R, R)
	loadExecAllocator func() A
	freeExec          func(A, R)
}

// activeAbilityDisable4FC300 preserves GAME.EXE 004FC300. The unit argument
// is cached and nil-checked before the signed ability argument is read. Only
// abilities 1..5 are accepted. Harpoon directly reads UpdateData and its bolt
// before emitting the break callback; Tread Lightly removes enchant 31; Eye
// of the Wolf returns without reporting or touching the active-record list.
//
// The inactive report for abilities 1..4 is emitted exactly once, before the
// list head is loaded. On each record Unit and Next are cached immediately,
// and Ability is read only when Unit matches. Matching records are unlinked
// with freshly reloaded live links, while traversal resumes through the Next
// cached before the match test. All matches are released through an allocator
// loaded after unlinking. Player class, Active, and deadline fields are not
// inspected.
func activeAbilityDisable4FC300[
	U comparable,
	D comparable,
	R comparable,
	A comparable,
](hooks activeAbilityDisableHooks4FC300[U, D, R, A]) {
	var zeroUnit U
	unit := hooks.loadUnitArg()
	if unit == zeroUnit {
		return
	}

	ability := hooks.loadAbilityArg()
	if ability <= server.AbilityInvalid || ability >= server.AbilityMax {
		return
	}

	switch ability {
	case server.AbilityHarpoon:
		update := hooks.loadUpdateData(unit)
		bolt := hooks.loadHarpoonBolt(update)
		hooks.breakHarpoon(unit, bolt)
	case server.AbilityTreadLightly:
		hooks.disableEnchant(unit, activeAbilityDisableSneakEnchant4FC300)
	case server.AbilityInfravis:
		return
	}

	hooks.reportActive(unit, ability, activeAbilityDisableInactive4FC300)

	var zeroRecord R
	for record := hooks.loadExecHead(); record != zeroRecord; {
		recordUnit := hooks.loadExecUnit(record)
		cachedNext := hooks.loadExecNext(record)
		if recordUnit == unit {
			recordAbility := hooks.loadExecAbility(record)
			if recordAbility == ability {
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
