package server

const (
	spellDurationProcessDeadOrDestroyed4FEEF0 = uint32(0x8020)
	spellDurationProcessDestroyed4FEEF0       = byte(0x20)
)

// SpellDurationProcessHooks4FEEF0 exposes every observable list, record,
// object, frame, and callback access in GAME.EXE 004FEEF0. Comparable tokens
// preserve native null and pointer identity without inheriting PE32 width.
type SpellDurationProcessHooks4FEEF0[Record, Object, Update comparable] struct {
	LoadFirst              func() Record
	LoadFlagsLowByte       func(Record) byte
	LoadNext               func(Record) Record
	Destroy                func(Record)
	LoadCaster             func(Record) Object
	LoadObjectFlags        func(Object) uint32
	StoreCaster            func(Record, Object)
	LoadObj12              func(Record) Object
	LoadObjectFlagsLowByte func(Object) byte
	StoreObj12             func(Record, Object)
	LoadFlag20             func(Record) uint32
	LoadObj24              func(Record) Object
	StoreObj24             func(Record, Object)
	LoadFrame68            func(Record) uint32
	LoadFrame60            func(Record) uint32
	LoadCurrentFrame       func() uint32
	LoadUpdate             func(Record) Update
	CallUpdate             func(Update, Record) int32
	Cancel                 func(Record)
}

func spellDurationProcessLive4FEEF0[Record, Object, Update comparable](
	record Record,
	h SpellDurationProcessHooks4FEEF0[Record, Object, Update],
) {
	var nilObject Object
	caster := h.LoadCaster(record)
	if caster != nilObject && h.LoadObjectFlags(caster)&spellDurationProcessDeadOrDestroyed4FEEF0 != 0 {
		h.StoreCaster(record, nilObject)
	}

	obj12 := h.LoadObj12(record)
	if obj12 != nilObject && h.LoadObjectFlagsLowByte(obj12)&spellDurationProcessDestroyed4FEEF0 != 0 {
		h.StoreObj12(record, nilObject)
	}

	if h.LoadCaster(record) == nilObject && h.LoadFlag20(record) == 0 {
		h.Cancel(record)
		return
	}

	obj24 := h.LoadObj24(record)
	if obj24 != nilObject && h.LoadObjectFlagsLowByte(obj24)&spellDurationProcessDestroyed4FEEF0 != 0 {
		h.StoreObj24(record, nilObject)
	}

	frame68 := h.LoadFrame68(record)
	frame60 := h.LoadFrame60(record)
	if frame68 != frame60 && frame68 <= h.LoadCurrentFrame() {
		h.Cancel(record)
		return
	}

	update := h.LoadUpdate(record)
	var nilUpdate Update
	if update != nilUpdate && h.CallUpdate(update, record) != 0 {
		h.Cancel(record)
	}
}

// SpellDurationProcess4FEEF0 preserves GAME.EXE 004FEEF0's exact traversal
// and short-circuit order. Each record's low flag byte is loaded before its
// Next link is snapshotted. Pending records go directly to destruction.
// Otherwise dead or destroyed casters and destroyed auxiliary objects are
// cleared in original field order. Caster is reloaded before Flag20, Frame68
// is cached across the Frame60/current-frame comparisons, and Update is
// cached across its nil test and call. Expiry uses unsigned dword ordering;
// equal Frame68/Frame60 skips the current-frame load and still reaches Update.
// Destruction and cancellation cannot redirect traversal from the saved
// successor. No record, object, callback, or cycle guard is added.
func SpellDurationProcess4FEEF0[Record, Object, Update comparable](
	h SpellDurationProcessHooks4FEEF0[Record, Object, Update],
) {
	record := h.LoadFirst()
	var nilRecord Record
	if record == nilRecord {
		return
	}

	for {
		flags := h.LoadFlagsLowByte(record)
		next := h.LoadNext(record)
		if flags&1 != 0 {
			h.Destroy(record)
		} else {
			spellDurationProcessLive4FEEF0(record, h)
		}
		record = next
		if record == nilRecord {
			return
		}
	}
}
