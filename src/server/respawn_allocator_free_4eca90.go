package server

// RespawnAllocatorFreeHooks4ECA90 separates the portable free contract from
// the legacy pointer-valued allocator global and allocation-class runtime.
type RespawnAllocatorFreeHooks4ECA90[A any] struct {
	LoadAllocator func() A
	FreeClass     func(A)
}

// RespawnAllocatorFree4ECA90 preserves GAME.EXE 004ECA90. The allocator is
// loaded exactly once and that captured value, including nil, is forwarded to
// the class-free boundary. The original does not clear the allocator global.
func RespawnAllocatorFree4ECA90[A any](hooks RespawnAllocatorFreeHooks4ECA90[A]) {
	allocator := hooks.LoadAllocator()
	hooks.FreeClass(allocator)
}
