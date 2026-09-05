package server

const spellDurationPlayerClassLowByte4FE8A0 = uint8(0x04)

// spellDurationSelectiveCleanupHooks4FE8A0 exposes every observable load,
// store, and callback in GAME.EXE 004FE8A0. Pointer-bearing values are generic
// tokens so the semantic core cannot inherit the original PE32 pointer width.
type spellDurationSelectiveCleanupHooks4FE8A0[Record, Object comparable, Allocator any] struct {
	loadAllocator    func() Allocator
	freeAllObjects   func(Allocator)
	clearList        func()
	loadList         func() Record
	loadTarget       func(Record) Object
	loadNext         func(Record) Record
	loadClassLowByte func(Object) uint8
	unlink           func(Record)
	freeRecursive    func(Record)
}

// spellDurationSelectiveCleanup4FE8A0 preserves GAME.EXE 004FE8A0. A zero
// mode snapshots the duration allocator, frees all records (including through
// a nil allocator), and only then clears the list head. Any nonzero mode walks
// the original list and removes records with a nil target or a target whose
// class low byte lacks Player.
//
// The successor is deliberately loaded before the target is tested or its
// class byte is read. Unlink and recursive-free callbacks therefore cannot
// redirect the current iteration. No guards are added at callback boundaries.
func spellDurationSelectiveCleanup4FE8A0[Record, Object comparable, Allocator any](
	mode int32,
	h spellDurationSelectiveCleanupHooks4FE8A0[Record, Object, Allocator],
) {
	if mode == 0 {
		allocator := h.loadAllocator()
		h.freeAllObjects(allocator)
		h.clearList()
		return
	}

	record := h.loadList()
	var nilRecord Record
	if record == nilRecord {
		return
	}
	var nilObject Object
	for {
		target := h.loadTarget(record)
		next := h.loadNext(record)
		if target == nilObject || h.loadClassLowByte(target)&spellDurationPlayerClassLowByte4FE8A0 == 0 {
			h.unlink(record)
			h.freeRecursive(record)
		}
		record = next
		if record == nilRecord {
			return
		}
	}
}
