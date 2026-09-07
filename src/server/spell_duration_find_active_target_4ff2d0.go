package server

// SpellDurationFindActiveTargetHooks4FF2D0 exposes every observable list and
// record access in GAME.EXE 004FF2D0. Comparable tokens preserve canonical
// null values and full native pointer identities without inheriting PE32's
// pointer width.
type SpellDurationFindActiveTargetHooks4FF2D0[Record, Object comparable] struct {
	LoadFirst    func() Record
	LoadFlagsLow func(Record) byte
	LoadSpell    func(Record) int32
	LoadTarget   func(Record) Object
	LoadNext     func(Record) Record
}

// SpellDurationFindActiveTarget4FF2D0 returns the first active duration-spell
// record whose full dword spell ID and non-null target match the requested
// values. The original loads Flags88's low byte first and gates every later
// record access: inactive records do not load Spell, spell mismatches do not
// load Target48, and an exact target match returns without loading Next.
//
// First is called exactly once. Every mismatch calls Next with the live
// current record and follows the returned native-width identity. No record,
// callback, or cycle guard is added; empty and exhausted paths return the
// canonical zero Record.
func SpellDurationFindActiveTarget4FF2D0[Record, Object comparable](
	requestedSpell int32,
	target Object,
	h SpellDurationFindActiveTargetHooks4FF2D0[Record, Object],
) Record {
	record := h.LoadFirst()
	var nilRecord Record
	if record == nilRecord {
		return nilRecord
	}

	var nilObject Object
	for {
		if h.LoadFlagsLow(record)&1 == 0 && h.LoadSpell(record) == requestedSpell {
			recordTarget := h.LoadTarget(record)
			if recordTarget != nilObject && recordTarget == target {
				return record
			}
		}
		record = h.LoadNext(record)
		if record == nilRecord {
			return nilRecord
		}
	}
}
