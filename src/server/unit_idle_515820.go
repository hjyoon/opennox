package server

const (
	unitIdleMonsterClassLow515820 = uint8(0x02)
	unitIdleBlockedFlag515820     = uint32(0x00008000)
	unitIdleAction515820          = uint32(0)
)

// unitIdleHooks515820 exposes every observable load and call made by
// GAME.EXE 00515820. A generic object handle keeps the semantic contract
// independent of the original PE32 pointer width.
type unitIdleHooks515820[O comparable] struct {
	loadClassLow     func(O) uint8
	loadFlags        func(O) uint32
	clearActionStack func(O)
	pushAction       func(O, uint32)
}

// unitIdle515820 preserves the original null, class, and flag short-circuit
// order before replacing an eligible monster's action stack with action zero.
// The return value of the original push routine is deliberately unobserved.
func unitIdle515820[O comparable](unit O, hooks unitIdleHooks515820[O]) {
	var nilUnit O
	if unit == nilUnit {
		return
	}
	if hooks.loadClassLow(unit)&unitIdleMonsterClassLow515820 == 0 {
		return
	}
	if hooks.loadFlags(unit)&unitIdleBlockedFlag515820 != 0 {
		return
	}
	hooks.clearActionStack(unit)
	hooks.pushAction(unit, unitIdleAction515820)
}
