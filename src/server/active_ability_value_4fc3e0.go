package server

const (
	activeAbilityValuePlayerClass4FC3E0 = uint8(0x04)
	activeAbilityValueWarrior4FC3E0     = uint8(0)
)

type activeAbilityValueHooks4FC3E0[
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
	loadExecActive   func(R) uint32
}

// activeAbilityValue4FC3E0 preserves GAME.EXE 004FC3E0. The unit argument and
// its low class byte are read first, without a defensive nil check.
// Non-Players return before UpdateData is read. A nil UpdateData skips the
// Player-class gate, while a non-nil UpdateData requires a Player pointer whose
// class byte is exactly Warrior; that Player pointer is deliberately not
// checked for nil.
//
// The global record head is loaded only after the class gates. An empty list
// returns without reading the signed ability argument. For each record Unit
// and Next are cached before comparison. Ability is read only when the cached
// Unit matches. A full 32-bit Active value is read and returned only when both
// keys match; it is not canonicalized to a boolean. Frame/deadline and Prev are
// deliberately ignored.
func activeAbilityValue4FC3E0[
	U comparable,
	D comparable,
	P comparable,
	R comparable,
](hooks activeAbilityValueHooks4FC3E0[U, D, P, R]) int32 {
	unit := hooks.loadUnitArg()
	if hooks.loadUnitClassLow(unit)&activeAbilityValuePlayerClass4FC3E0 == 0 {
		return 0
	}

	update := hooks.loadUpdateData(unit)
	var zeroData D
	if update != zeroData {
		player := hooks.loadPlayer(update)
		if hooks.loadPlayerClass(player) != activeAbilityValueWarrior4FC3E0 {
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
			return int32(hooks.loadExecActive(record))
		}
		record = next
		if record == zeroRecord {
			return 0
		}
	}
}
