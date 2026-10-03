package server

type itemDefendEffectsHooks4E1320[O comparable, I any, M comparable] struct {
	lowDWORD    func(O) int32
	firstItem   func(O) O
	flags       func(O) uint32
	initData    func(O) I
	modifier    func(I, int) M
	hasDefend   func(M) bool
	loadDamage  func() int32
	storeDamage func(int32)
	applyDefend func(M, O, O, O, O, *[2]int32)
	nextItem    func(O) O
}

// itemDefendEffects4E1320 follows each load/call of GAME.EXE 004E1320.
// Equipped flags alone gate the two slots. Cache the slot base, not its
// contents; slot three and the next item are loaded after prior callbacks.
// The incidental EAX result is the target low DWORD for an empty inventory,
// the last non-equipped flags, or the final slot counter zero.
func itemDefendEffects4E1320[O comparable, I any, M comparable](
	target, source, weapon O, typ int32, hooks itemDefendEffectsHooks4E1320[O, I, M],
) int32 {
	result := hooks.lowDWORD(target)
	var nilObject O
	var nilModifier M
	// The original stack has one context address reused by every callback.
	var context [2]int32
	for item := hooks.firstItem(target); item != nilObject; item = hooks.nextItem(item) {
		result = int32(hooks.flags(item))
		if uint32(result)&0x100 == 0 {
			continue
		}
		initData := hooks.initData(item)
		for slot := 2; slot < 4; slot++ {
			modifier := hooks.modifier(initData, slot)
			if modifier != nilModifier && hooks.hasDefend(modifier) {
				context = [2]int32{hooks.loadDamage(), typ}
				hooks.applyDefend(modifier, item, target, weapon, source, &context)
				hooks.storeDamage(context[0])
			}
			result = int32(3 - slot)
		}
	}
	return result
}
