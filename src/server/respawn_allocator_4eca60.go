package server

const (
	respawnAllocatorName4ECA60     = "Respawn"
	respawnAllocatorCapacity4ECA60 = 384
)

// RespawnAllocatorHooks4ECA60 separates the portable allocation contract from
// the legacy allocation class and pointer-valued global.
type RespawnAllocatorHooks4ECA60[A any] struct {
	NewClass       func(name string, recordSize uintptr, capacity int) A
	NonZero        func(A) bool
	StoreAllocator func(A)
}

// RespawnAllocator4ECA60 preserves GAME.EXE 004ECA60. The original 32-bit
// function requests 384 records named "Respawn", tests the returned class
// pointer before storing it, stores that pointer even when it is nil, and
// returns the cached zero/nonzero result. recordSize is supplied by the native
// adapter so pointer-bearing records widen without changing this order.
func RespawnAllocator4ECA60[A any](recordSize uintptr, hooks RespawnAllocatorHooks4ECA60[A]) bool {
	allocator := hooks.NewClass(respawnAllocatorName4ECA60, recordSize, respawnAllocatorCapacity4ECA60)
	ok := hooks.NonZero(allocator)
	hooks.StoreAllocator(allocator)
	return ok
}
