package server

const (
	activeAbilityUnitMembershipPlayerClass4FC2B0 = uint8(0x04)
	activeAbilityUnitMembershipWarrior4FC2B0     = uint8(0)
)

type activeAbilityUnitMembershipHooks4FC2B0[
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
	loadExecUnit     func(R) U
	loadExecNext     func(R) R
}

// activeAbilityUnitMembership4FC2B0 preserves GAME.EXE 004FC2B0. The unit
// argument and its low class byte are read first without a defensive nil
// check. Non-Players return before UpdateData is read. A nil UpdateData skips
// the Player-class gate, while a non-nil UpdateData requires a Player pointer
// whose class byte is exactly Warrior; that Player pointer is deliberately not
// checked for nil.
//
// The global execution-list head is loaded only after the class gates. Each
// record's Unit is read and compared before any Next read. A matching record
// returns canonical one immediately; only a mismatch reads the live Next link.
// Ability, Active, deadline, and Prev are deliberately ignored. A list miss
// returns canonical zero.
func activeAbilityUnitMembership4FC2B0[
	U comparable,
	D comparable,
	P comparable,
	R comparable,
](hooks activeAbilityUnitMembershipHooks4FC2B0[U, D, P, R]) int32 {
	unit := hooks.loadUnitArg()
	if hooks.loadUnitClassLow(unit)&activeAbilityUnitMembershipPlayerClass4FC2B0 == 0 {
		return 0
	}

	update := hooks.loadUpdateData(unit)
	var zeroData D
	if update != zeroData {
		player := hooks.loadPlayer(update)
		if hooks.loadPlayerClass(player) != activeAbilityUnitMembershipWarrior4FC2B0 {
			return 0
		}
	}

	var zeroRecord R
	for record := hooks.loadExecHead(); record != zeroRecord; record = hooks.loadExecNext(record) {
		if hooks.loadExecUnit(record) == unit {
			return 1
		}
	}
	return 0
}
