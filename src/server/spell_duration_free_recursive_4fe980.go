package server

// SpellDurationFreeRecursiveHooks4FE980 exposes every observable record-link
// and allocator access in GAME.EXE 004FE980. Record is comparable so its zero
// value can model the original nil pointer without narrowing native addresses.
type SpellDurationFreeRecursiveHooks4FE980[Allocator any, Record comparable] struct {
	LoadSub108      func(Record) Record
	LoadSub104      func(Record) Record
	LoadNext        func(Record) Record
	LoadAllocator   func() Allocator
	FreeObjectFirst func(Allocator, Record)
}

// SpellDurationFreeRecursive4FE980 preserves GAME.EXE 004FE980. It walks the
// Sub108 sibling chain before loading and walking the live Sub104 chain. Each
// sibling's Next link is deliberately cached before recursively freeing that
// sibling. The allocator is loaded only after both child chains have returned,
// immediately before the current record is freed. No nil or cycle guard is
// added.
func SpellDurationFreeRecursive4FE980[Allocator any, Record comparable](
	hooks SpellDurationFreeRecursiveHooks4FE980[Allocator, Record],
	record Record,
) {
	var nilRecord Record
	for child := hooks.LoadSub108(record); child != nilRecord; {
		next := hooks.LoadNext(child)
		SpellDurationFreeRecursive4FE980(hooks, child)
		child = next
	}
	for child := hooks.LoadSub104(record); child != nilRecord; {
		next := hooks.LoadNext(child)
		SpellDurationFreeRecursive4FE980(hooks, child)
		child = next
	}
	allocator := hooks.LoadAllocator()
	hooks.FreeObjectFirst(allocator, record)
}
