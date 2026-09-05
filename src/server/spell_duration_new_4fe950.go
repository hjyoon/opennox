package server

// SpellDurationNewHooks4FE950 separates the duration-spell allocation
// contract from the native allocation class, last identifier, and record.
// Record is comparable so its zero value can model the original nil pointer
// without narrowing a native-width address.
type SpellDurationNewHooks4FE950[Allocator any, Record comparable] struct {
	LoadAllocator func() Allocator
	NewObject     func(Allocator) Record
	LoadLastID    func() uint16
	StoreLastID   func(uint16)
	StoreRecordID func(Record, uint16)
}

// SpellDurationNew4FE950 preserves GAME.EXE 004FE950. The original PE32
// function snapshots the duration-spell allocator and requests one zeroed
// record. Allocation failure returns nil without touching the last identifier.
// Success loads the live uint16 identifier, increments it with 16-bit wrap,
// stores it globally, then stores the same cached value in the record's first
// word. No exhaustion guard, reserved-zero rule, or result canonicalization is
// added.
func SpellDurationNew4FE950[Allocator any, Record comparable](
	hooks SpellDurationNewHooks4FE950[Allocator, Record],
) Record {
	allocator := hooks.LoadAllocator()
	record := hooks.NewObject(allocator)
	var nilRecord Record
	if record == nilRecord {
		return record
	}
	id := hooks.LoadLastID() + 1
	hooks.StoreLastID(id)
	hooks.StoreRecordID(record, id)
	return record
}
