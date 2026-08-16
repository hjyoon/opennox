package server

const objectDeadFlagLow4ECF10 = uint8(0x20)

// objectByScriptIDHooks4ECF10 separates the active-object, inventory,
// pending-object, and missile-object lists traversed by GAME.EXE 004ECF10.
// Object identities remain pointer-width values, while ScriptID remains the
// original signed 32-bit field on every host architecture.
type objectByScriptIDHooks4ECF10[O comparable] struct {
	firstActive     func() O
	loadScriptIDArg func() int32
	nextActive      func(O) O
	firstInventory  func(O) O
	nextInventory   func(O) O
	firstPending    func() O
	nextPending     func(O) O
	firstMissile    func() O
	nextMissile     func(O) O
	loadFlagsLow    func(O) uint8
	loadScriptID    func(O) int32
}

func objectMatchesScriptID4ECF10[O comparable](
	obj O,
	scriptID int32,
	hooks objectByScriptIDHooks4ECF10[O],
) bool {
	if hooks.loadFlagsLow(obj)&objectDeadFlagLow4ECF10 != 0 {
		return false
	}
	return hooks.loadScriptID(obj) == scriptID
}

// objectByScriptID4ECF10 preserves GAME.EXE 004ECF10.
//
// The first active object is requested before the incoming ScriptID is loaded
// and cached. Active top-level objects and each active object's inventory are
// searched first, followed by pending objects and then missile objects. Every
// candidate reads its low flags byte before ScriptID; dead candidates skip the
// ScriptID load. A match returns the exact cached candidate without reading its
// next pointer or consulting any later domain.
func objectByScriptID4ECF10[O comparable](hooks objectByScriptIDHooks4ECF10[O]) O {
	var zero O
	obj := hooks.firstActive()
	scriptID := hooks.loadScriptIDArg()

	for ; obj != zero; obj = hooks.nextActive(obj) {
		if objectMatchesScriptID4ECF10(obj, scriptID, hooks) {
			return obj
		}
		for item := hooks.firstInventory(obj); item != zero; item = hooks.nextInventory(item) {
			if objectMatchesScriptID4ECF10(item, scriptID, hooks) {
				return item
			}
		}
	}

	for obj = hooks.firstPending(); obj != zero; obj = hooks.nextPending(obj) {
		if objectMatchesScriptID4ECF10(obj, scriptID, hooks) {
			return obj
		}
	}

	for obj = hooks.firstMissile(); obj != zero; obj = hooks.nextMissile(obj) {
		if objectMatchesScriptID4ECF10(obj, scriptID, hooks) {
			return obj
		}
	}
	return zero
}
