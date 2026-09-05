package server

// SpellDurationCleanupHooks4FE880 separates the cleanup contract from the
// native allocation class and list head.
type SpellDurationCleanupHooks4FE880[A any] struct {
	LoadAllocator func() A
	FreeAllocator func(A)
	ClearList     func()
}

// SpellDurationCleanup4FE880 preserves GAME.EXE 004FE880. The original PE32
// function snapshots the duration-spell allocator, forwards that value to the
// class destructor even when it is nil, and only then clears the list head. It
// does not clear the allocator handle or the last duration-spell identifier.
func SpellDurationCleanup4FE880[A any](hooks SpellDurationCleanupHooks4FE880[A]) {
	allocator := hooks.LoadAllocator()
	hooks.FreeAllocator(allocator)
	hooks.ClearList()
}
