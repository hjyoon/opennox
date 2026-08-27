package server

// itemApplyEngageEffectHooks4F2FF0 keeps the object, modifier-init-data,
// modifier, and callback pointer domains distinct. This exposes every PE32
// load in GAME.EXE 004F2FF0 without representing a native pointer as uint32.
type itemApplyEngageEffectHooks4F2FF0[O, D, M, C comparable] struct {
	loadInitData func(O) D
	loadModifier func(D, int) M
	loadEngage   func(M) C
	callEngage   func(C, M, O, O)
}

// itemApplyEngageEffect4F2FF0 preserves GAME.EXE 004F2FF0. Item InitData is
// loaded exactly once, then modifier slots two and three are loaded in order.
// A nil modifier skips its callback load; a nil callback skips invocation.
// Every non-nil callback receives modifier, owner, and item in that order.
//
// The original has no item or InitData guard. A callback may mutate the
// cached modifier data, so the second slot and its callback remain live loads.
func itemApplyEngageEffect4F2FF0[O, D, M, C comparable](
	item, owner O,
	hooks itemApplyEngageEffectHooks4F2FF0[O, D, M, C],
) {
	data := hooks.loadInitData(item)
	var nilModifier M
	var nilCallback C
	for slot := 2; slot < 4; slot++ {
		modifier := hooks.loadModifier(data, slot)
		if modifier == nilModifier {
			continue
		}
		callback := hooks.loadEngage(modifier)
		if callback == nilCallback {
			continue
		}
		hooks.callEngage(callback, modifier, owner, item)
	}
}
