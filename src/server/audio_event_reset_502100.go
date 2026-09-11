package server

// audioEventResetHooks502100 describes the only observable operations made by
// GAME.EXE 00502100. The allocation-class handle remains native-width.
type audioEventResetHooks502100[C any] struct {
	loadClass      func() C
	freeAllObjects func(C)
	clearHead      func()
}

// resetAudioEvents502100 preserves GAME.EXE 00502100's exact load, call, and
// store order. It returns all active events to the existing allocation class,
// then clears the global event-list head. It does not free the class or touch
// the initialized flag, bitmap, per-sound descriptors, or queue lifecycle.
func resetAudioEvents502100[C any](hooks audioEventResetHooks502100[C]) {
	class := hooks.loadClass()
	hooks.freeAllObjects(class)
	hooks.clearHead()
}
