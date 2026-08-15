package server

// RespawnResetHooks4EC5B0 separates the portable reset contract from the
// legacy storage that still owns the respawn allocator, list head, and gate.
type RespawnResetHooks4EC5B0[A any] struct {
	LoadAllocator func() A
	ClearHead     func()
	FreeAll       func(A)
	Enable        func()
}

// RespawnReset4EC5B0 preserves GAME.EXE 004EC5B0. The allocator is captured
// before the list head is cleared. FreeAll therefore receives that captured
// value while observing an already-empty list and the old allow flag. The
// allow flag is set only after FreeAll returns normally.
func RespawnReset4EC5B0[A any](hooks RespawnResetHooks4EC5B0[A]) {
	allocator := hooks.LoadAllocator()
	hooks.ClearHead()
	hooks.FreeAll(allocator)
	hooks.Enable()
}
