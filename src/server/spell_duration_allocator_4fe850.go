package server

const (
	spellDurationAllocatorName4FE850     = "spellDuration"
	spellDurationAllocatorCapacity4FE850 = 512
)

// SpellDurationAllocatorHooks4FE850 separates the allocation contract from
// the native allocation class and pointer-valued global.
type SpellDurationAllocatorHooks4FE850[A any] struct {
	NewClass       func(name string, recordSize uintptr, capacity int) A
	NonZero        func(A) bool
	StoreAllocator func(A)
}

// SpellDurationAllocator4FE850 preserves GAME.EXE 004FE850. The original
// PE32 function requests 512 records named "spellDuration", tests the returned
// class pointer, stores it even when nil, and returns the cached canonical
// zero/one result. recordSize is supplied by the native adapter so records
// containing pointers widen without changing the observable operation order.
func SpellDurationAllocator4FE850[A any](
	recordSize uintptr,
	hooks SpellDurationAllocatorHooks4FE850[A],
) int32 {
	allocator := hooks.NewClass(
		spellDurationAllocatorName4FE850,
		recordSize,
		spellDurationAllocatorCapacity4FE850,
	)
	ok := hooks.NonZero(allocator)
	hooks.StoreAllocator(allocator)
	if ok {
		return 1
	}
	return 0
}
