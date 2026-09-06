package server

// SpellDurationCancelSelectedHooks4FEE90 exposes every observable list,
// argument, record, and callback access in GAME.EXE 004FEE90. Comparable
// record and object tokens preserve native null and pointer identity without
// inheriting the executable's PE32 pointer width.
type SpellDurationCancelSelectedHooks4FEE90[Record, Object comparable] struct {
	LoadFirst     func() Record
	LoadCasterArg func() Object
	LoadNext      func(Record) Record
	LoadCaster    func(Record) Object
	LoadSpell     func(Record) uint32
	Cancel        func(Record)
}

func spellDurationCancelSelectedID4FEE90(spell uint32) bool {
	return spell == 24 ||
		spell == 43 ||
		spell == 35 ||
		spell == 8 ||
		spell == 22 ||
		spell == 59 ||
		spell == 67
}

// SpellDurationCancelSelected4FEE90 preserves GAME.EXE 004FEE90's exact
// traversal and short-circuit order. The list head is loaded before the
// caster argument, so an empty list returns without reading that argument.
// On a non-empty list the caster identity is cached once. Each record's Next
// link is snapshotted before its caster and spell fields; a caster mismatch
// skips the spell load. Matching records load the full spell dword once and
// cancel only IDs 24, 43, 35, 8, 22, 59, and 67 in that comparison order.
// Traversal always continues through the saved successor, even if Cancel
// mutates the record or list. Nil is a valid caster identity, and no record,
// callback, or cycle guard is added. The original residual EAX is outside the
// function's void contract.
func SpellDurationCancelSelected4FEE90[Record, Object comparable](
	h SpellDurationCancelSelectedHooks4FEE90[Record, Object],
) {
	record := h.LoadFirst()
	var nilRecord Record
	if record == nilRecord {
		return
	}

	caster := h.LoadCasterArg()
	for {
		next := h.LoadNext(record)
		if h.LoadCaster(record) == caster {
			spell := h.LoadSpell(record)
			if spellDurationCancelSelectedID4FEE90(spell) {
				h.Cancel(record)
			}
		}
		record = next
		if record == nilRecord {
			return
		}
	}
}
