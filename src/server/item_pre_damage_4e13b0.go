package server

type itemPreDamageHooks4E13B0[O, I any, M comparable, D any] struct {
	initData       func(O) I
	modifier       func(I, int) M
	hasPreDamage   func(M) bool
	applyPreDamage func(M, O, O, O, D)
}

// itemPreDamage4E13B0 preserves GAME.EXE 004E13B0's four live slot loads.
// Cache only the slot base. The callback receives the original damage address;
// the helper neither reads nor copies its word and adds no object-class gates.
func itemPreDamage4E13B0[O, I any, M comparable, D any](
	target, source, weapon O, damage D, hooks itemPreDamageHooks4E13B0[O, I, M, D],
) int32 {
	initData := hooks.initData(weapon)
	var nilModifier M
	for slot := 0; slot < 4; slot++ {
		modifier := hooks.modifier(initData, slot)
		if modifier != nilModifier && hooks.hasPreDamage(modifier) {
			hooks.applyPreDamage(modifier, weapon, source, target, damage)
		}
	}
	return 0
}
