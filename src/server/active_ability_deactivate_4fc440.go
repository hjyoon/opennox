package server

const (
	activeAbilityDeactivatePlayerClass4FC440 = uint8(0x04)
	activeAbilityDeactivateWarrior4FC440     = uint8(0)
)

type activeAbilityDeactivateHooks4FC440[
	U comparable,
	D comparable,
	P comparable,
	R comparable,
] struct {
	loadUnitArg      func() U
	loadUnitClassLow func(U) uint8
	loadUpdateData   func(U) D
	loadPlayer       func(D) P
	loadPlayerClass  func(P) uint8
	loadExecHead     func() R
	loadAbilityArg   func() Ability
	loadExecUnit     func(R) U
	loadExecNext     func(R) R
	loadExecAbility  func(R) Ability
	storeExecActive  func(R, uint32)
}

// activeAbilityDeactivate4FC440 preserves GAME.EXE 004FC440. The unit
// argument and its low class byte are read first, without a defensive nil
// check. Non-Players return before UpdateData is read. A nil UpdateData skips
// the Player-class gate, while a non-nil UpdateData requires a Player pointer
// whose class byte is exactly Warrior; that Player pointer is deliberately not
// checked for nil.
//
// The global record head is loaded only after the class gates. An empty list
// returns without reading the signed ability argument. For each record Unit
// and Next are cached before comparison. Ability is read only when the cached
// Unit matches. The first record matching both keys has its complete 32-bit
// Active field overwritten with zero and the function returns immediately.
// Frame/deadline and Prev are deliberately ignored; no record is unlinked.
func activeAbilityDeactivate4FC440[
	U comparable,
	D comparable,
	P comparable,
	R comparable,
](hooks activeAbilityDeactivateHooks4FC440[U, D, P, R]) {
	unit := hooks.loadUnitArg()
	if hooks.loadUnitClassLow(unit)&activeAbilityDeactivatePlayerClass4FC440 == 0 {
		return
	}

	update := hooks.loadUpdateData(unit)
	var zeroData D
	if update != zeroData {
		player := hooks.loadPlayer(update)
		if hooks.loadPlayerClass(player) != activeAbilityDeactivateWarrior4FC440 {
			return
		}
	}

	var zeroRecord R
	record := hooks.loadExecHead()
	if record == zeroRecord {
		return
	}
	ability := hooks.loadAbilityArg()
	for {
		recordUnit := hooks.loadExecUnit(record)
		next := hooks.loadExecNext(record)
		if recordUnit == unit && hooks.loadExecAbility(record) == ability {
			hooks.storeExecActive(record, 0)
			return
		}
		record = next
		if record == zeroRecord {
			return
		}
	}
}
