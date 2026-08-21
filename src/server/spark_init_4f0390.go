package server

const sparkInitLifetime4F0390 = uint32(32)

type sparkInitHooks4F0390[O, D any] struct {
	loadUpdateData         func(O) D
	storeLifetimeRemaining func(D, uint32)
	storeLifetimeInitial   func(D, uint32)
}

// sparkInit4F0390 preserves the exact observable order of GAME.EXE 004F0390.
// The update-data pointer is cached once. Remaining lifetime at offset +4 is
// written before initial lifetime at offset +0, and the cached pointer is
// returned. The original has no nil guards.
func sparkInit4F0390[O, D any](unit O, hooks sparkInitHooks4F0390[O, D]) D {
	update := hooks.loadUpdateData(unit)
	hooks.storeLifetimeRemaining(update, sparkInitLifetime4F0390)
	hooks.storeLifetimeInitial(update, sparkInitLifetime4F0390)
	return update
}
