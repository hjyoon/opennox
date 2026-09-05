package server

const (
	spellDurationCancelPlayerClass4FE9D0    = byte(4)
	spellDurationCancelChainLightning4FE9D0 = uint32(43)
)

// SpellDurationCancelHooks4FE9D0 exposes every observable pointer, field,
// callback, and low-byte flag access in GAME.EXE 004FE9D0. Comparable record
// and object tokens preserve a native null value without imposing PE32 width.
type SpellDurationCancelHooks4FE9D0[Record, Object comparable, Update, Player any] struct {
	LoadCaster        func(Record) Object
	LoadClassLowByte  func(Object) byte
	LoadSpell         func(Record) uint32
	LoadUpdate        func(Object) Update
	LoadPlayer        func(Update) Player
	LoadPlayerIndex   func(Player) byte
	ReportSpellStat   func(byte, uint32, byte)
	LoadSub108        func(Record) Record
	LoadTarget        func(Record) Object
	StopRay           func(Record, Object)
	LoadNext          func(Record) Record
	LoadFlagsLowByte  func(Record) byte
	StoreFlagsLowByte func(Record, byte)
}

// SpellDurationCancel4FE9D0 preserves GAME.EXE 004FE9D0's callback and live-
// reload order. A player caster reports the spell value loaded before the
// callback. The ray branch reloads Spell afterward. Chain Lightning loads each
// current node's Next only after an optional StopRay callback, while the final
// flag byte is loaded after every ray callback and returned after bit zero is
// stored. No record, update, player, or cycle guard is added; the original
// caster and target null checks remain the only guards.
func SpellDurationCancel4FE9D0[Record, Object comparable, Update, Player any](
	record Record,
	h SpellDurationCancelHooks4FE9D0[Record, Object, Update, Player],
) byte {
	var nilObject Object
	caster := h.LoadCaster(record)
	if caster != nilObject && h.LoadClassLowByte(caster)&spellDurationCancelPlayerClass4FE9D0 != 0 {
		initialSpell := h.LoadSpell(record)
		update := h.LoadUpdate(caster)
		status := byte(15)
		if initialSpell == spellDurationCancelChainLightning4FE9D0 {
			status = 0
		}
		player := h.LoadPlayer(update)
		index := h.LoadPlayerIndex(player)
		h.ReportSpellStat(index, initialSpell, status)
	}

	liveSpell := h.LoadSpell(record)
	if liveSpell == spellDurationCancelChainLightning4FE9D0 {
		var nilRecord Record
		for current := h.LoadSub108(record); current != nilRecord; current = h.LoadNext(current) {
			target := h.LoadTarget(current)
			if target != nilObject {
				h.StopRay(current, target)
			}
		}
	} else {
		target := h.LoadTarget(record)
		if target != nilObject {
			h.StopRay(record, target)
		}
	}

	flags := h.LoadFlagsLowByte(record) | 1
	h.StoreFlagsLowByte(record, flags)
	return flags
}
