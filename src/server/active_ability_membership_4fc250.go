package server

const (
	activeAbilityMembershipPlayerClass4FC250 = uint8(0x04)
	activeAbilityMembershipWarrior4FC250     = uint8(0)
)

type activeAbilityMembershipHooks4FC250[
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
}

// activeAbilityMembership4FC250 preserves GAME.EXE 004FC250. The unit
// argument and its low class byte are read first, without a defensive nil
// check. Non-Players return before UpdateData is read. A nil UpdateData skips
// the Player-class gate, while a non-nil UpdateData requires a Player pointer
// whose class byte is exactly Warrior; that Player pointer is deliberately not
// checked for nil.
//
// The global record head is loaded only after the class gates. An empty list
// returns without reading the signed ability argument. For each record Unit
// and Next are cached before comparison. Ability is read only when the cached
// Unit matches, after Next has already been cached. Active, deadline, and Prev
// are deliberately ignored. A match and a miss return canonical one and zero.
func activeAbilityMembership4FC250[
	U comparable,
	D comparable,
	P comparable,
	R comparable,
](hooks activeAbilityMembershipHooks4FC250[U, D, P, R]) int32 {
	unit := hooks.loadUnitArg()
	if hooks.loadUnitClassLow(unit)&activeAbilityMembershipPlayerClass4FC250 == 0 {
		return 0
	}

	update := hooks.loadUpdateData(unit)
	var zeroData D
	if update != zeroData {
		player := hooks.loadPlayer(update)
		if hooks.loadPlayerClass(player) != activeAbilityMembershipWarrior4FC250 {
			return 0
		}
	}

	var zeroRecord R
	record := hooks.loadExecHead()
	if record == zeroRecord {
		return 0
	}
	ability := hooks.loadAbilityArg()
	for {
		recordUnit := hooks.loadExecUnit(record)
		next := hooks.loadExecNext(record)
		if recordUnit == unit && hooks.loadExecAbility(record) == ability {
			return 1
		}
		record = next
		if record == zeroRecord {
			return 0
		}
	}
}
