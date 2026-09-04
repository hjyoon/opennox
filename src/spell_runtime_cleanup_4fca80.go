package opennox

type spellRuntimeCleanupHooks4FCA80[Allocator, Object any] struct {
	freeDurations        func()
	loadMagicClass       func() Allocator
	freeMagicClass       func(Allocator)
	loadImaginaryCaster  func() Object
	clearMagicEntityHead func()
	delayedDelete        func(Object)
	clearImaginaryCaster func()
}

// spellRuntimeCleanup4FCA80 preserves GAME.EXE 004FCA80's observable load,
// callback, and store order. Allocator and Object remain native-width tokens;
// nil values are deliberately forwarded to the original nil-tolerant frees.
func spellRuntimeCleanup4FCA80[Allocator, Object any](
	hooks spellRuntimeCleanupHooks4FCA80[Allocator, Object],
) int32 {
	hooks.freeDurations()
	magicClass := hooks.loadMagicClass()
	hooks.freeMagicClass(magicClass)
	caster := hooks.loadImaginaryCaster()
	hooks.clearMagicEntityHead()
	hooks.delayedDelete(caster)
	hooks.clearImaginaryCaster()
	return 1
}
