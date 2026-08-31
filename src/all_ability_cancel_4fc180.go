package opennox

import "github.com/opennox/opennox/v1/server"

const (
	allAbilityCancelPlayerClass4FC180 = uint8(0x04)
	allAbilityCancelWarrior4FC180     = uint8(0)
	allAbilityCancelInactive4FC180    = int32(0)
)

type allAbilityCancelHooks4FC180[
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
	loadPlayerIndex   func(P) uint8
	storeCooldown     func(uint8, server.Ability, int32)
	resetAbilities    func(U, server.Ability)
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

// allAbilityCancel4FC180 preserves GAME.EXE 004FC180. The unit argument is
// cached before the nil and Player-class gates. UpdateData, Player, and the
// exact Warrior class byte are then read without defensive nil checks.
//
// A Warrior clears only ability slots 1 through 5. Player is deliberately
// reloaded from the cached UpdateData before every PlayerInd read. Slot zero is
// untouched. The aggregate reset callback receives the AbilityMax sentinel,
// and the active-record head is deliberately loaded only after that callback.
//
// On each record Unit and Next are cached immediately. Every record whose Unit
// matches is reported inactive with its cached Unit/ability values. The unlink
// deliberately reloads live Next/Prev links after the report callback, while
// traversal resumes through the Next cached before it. The allocator is loaded
// after unlinking. Ability, Active, and deadline values do not gate removal.
func allAbilityCancel4FC180[
	U comparable,
	D comparable,
	P comparable,
	R comparable,
	A comparable,
](hooks allAbilityCancelHooks4FC180[U, D, P, R, A]) {
	var zeroUnit U
	unit := hooks.loadUnitArg()
	if unit == zeroUnit {
		return
	}
	if hooks.loadUnitClassLow(unit)&allAbilityCancelPlayerClass4FC180 == 0 {
		return
	}

	update := hooks.loadUpdateData(unit)
	player := hooks.loadPlayer(update)
	if hooks.loadPlayerClass(player) != allAbilityCancelWarrior4FC180 {
		return
	}

	for ability := server.AbilityInvalid + 1; ability < server.AbilityMax; ability++ {
		player = hooks.loadPlayer(update)
		index := hooks.loadPlayerIndex(player)
		hooks.storeCooldown(index, ability, 0)
	}
	hooks.resetAbilities(unit, server.AbilityMax)

	var zeroRecord R
	for record := hooks.loadExecHead(); record != zeroRecord; {
		recordUnit := hooks.loadExecUnit(record)
		cachedNext := hooks.loadExecNext(record)
		if recordUnit == unit {
			recordAbility := hooks.loadExecAbility(record)
			hooks.reportActive(recordUnit, recordAbility, allAbilityCancelInactive4FC180)

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
		record = cachedNext
	}
}
