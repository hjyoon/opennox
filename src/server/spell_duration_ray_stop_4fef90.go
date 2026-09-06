package server

import "encoding/binary"

const (
	durationRayStopMessage4FEF90   = byte(0x9e)
	durationRayStopRecipient4FEF90 = int32(255)
	durationRayStopRemove4FEF90    = int32(1)
	durationRayStopUnmark4FEF90    = uint32(2)
)

// DurationRayStopHooks4FEF90 exposes every observable record, object,
// packet, and callback access in GAME.EXE 004FEF90. Comparable tokens retain
// native null and identity semantics without inheriting the PE32 pointer
// width. UnitCode deliberately returns a full dword; the packet stores only
// its low word, exactly like the original MOV stores.
type DurationRayStopHooks4FEF90[Record, Object comparable] struct {
	LoadCaster       func(Record) Object
	LoadSpell        func(Record) uint32
	LoadLevelLow     func(Record) byte
	LoadDirectionLow func(Object) byte
	LoadTarget       func(Record) Object
	LoadSub108       func(Record) Record
	LoadNext         func(Record) Record
	UnitCode         func(Object) uint32
	SendPacket       func(int32, [7]byte, Object, int32) int32
	UnmarkMinimap    func(Object, uint32)
}

func durationRayStopSend4FEF90[Record, Object comparable](
	record Record,
	who Object,
	typ, value byte,
	h DurationRayStopHooks4FEF90[Record, Object],
) {
	var packet [7]byte
	packet[0] = durationRayStopMessage4FEF90
	packet[1] = typ
	packet[2] = value
	binary.LittleEndian.PutUint16(packet[3:5], uint16(h.UnitCode(who)))
	caster := h.LoadCaster(record)
	binary.LittleEndian.PutUint16(packet[5:7], uint16(h.UnitCode(caster)))
	var nilObject Object
	_ = h.SendPacket(
		durationRayStopRecipient4FEF90,
		packet,
		nilObject,
		durationRayStopRemove4FEF90,
	)
	caster = h.LoadCaster(record)
	h.UnmarkMinimap(caster, durationRayStopUnmark4FEF90)
	h.UnmarkMinimap(who, durationRayStopUnmark4FEF90)
}

func durationRayStopGreaterHeal4FEF90[Record, Object comparable](
	record Record,
	who, entryCaster Object,
	h DurationRayStopHooks4FEF90[Record, Object],
) {
	if entryCaster == h.LoadTarget(record) {
		return
	}

	// GreaterHeal has its own executable packet path. Unlike every other
	// level-valued case, it loads Level only after both unit-code callbacks.
	var packet [7]byte
	packet[0] = durationRayStopMessage4FEF90
	packet[1] = 13
	binary.LittleEndian.PutUint16(packet[3:5], uint16(h.UnitCode(who)))
	caster := h.LoadCaster(record)
	binary.LittleEndian.PutUint16(packet[5:7], uint16(h.UnitCode(caster)))
	packet[2] = h.LoadLevelLow(record)
	var nilObject Object
	_ = h.SendPacket(
		durationRayStopRecipient4FEF90,
		packet,
		nilObject,
		durationRayStopRemove4FEF90,
	)
	caster = h.LoadCaster(record)
	h.UnmarkMinimap(caster, durationRayStopUnmark4FEF90)
	h.UnmarkMinimap(who, durationRayStopUnmark4FEF90)
}

// DurationRayStop4FEF90 preserves GAME.EXE 004FEF90's exact dispatch and
// side-effect order. Record, entry caster, and target are rejected in that
// order. Spell is a full dword; only IDs 7, 9, 22, 24, 35, 43, and 59 have
// cases. Ordinary packets always encode the target unit code before a live
// reload of the caster code. Plasma reads Direction1 from the cached entry
// caster, while GreaterHeal compares that cached caster with a live Target48
// and delays its Level load until after both unit-code callbacks. Send reloads
// Caster16 again before the first minimap callback.
//
// ChainLightning walks the live Sub108 list. Each child Target48 is loaded
// before recursion and the child's Next link is loaded only after recursion
// returns, so recursive mutation can redirect traversal. No record, object,
// callback, list-cycle, or post-callback validity guard is added.
func DurationRayStop4FEF90[Record, Object comparable](
	record Record,
	who Object,
	h DurationRayStopHooks4FEF90[Record, Object],
) {
	var nilRecord Record
	if record == nilRecord {
		return
	}
	entryCaster := h.LoadCaster(record)
	var nilObject Object
	if entryCaster == nilObject || who == nilObject {
		return
	}

	switch h.LoadSpell(record) {
	default:
		return
	case 7: // ChainLightningBolt
		durationRayStopSend4FEF90(record, who, 10, h.LoadLevelLow(record), h)
	case 9: // Charm
		durationRayStopSend4FEF90(record, who, 9, h.LoadLevelLow(record), h)
	case 22: // DrainMana
		durationRayStopSend4FEF90(record, who, 12, h.LoadLevelLow(record), h)
	case 24: // Lightning
		durationRayStopSend4FEF90(record, who, 11, h.LoadLevelLow(record), h)
	case 35: // GreaterHeal
		durationRayStopGreaterHeal4FEF90(record, who, entryCaster, h)
	case 43: // ChainLightning
		child := h.LoadSub108(record)
		for child != nilRecord {
			target := h.LoadTarget(child)
			DurationRayStop4FEF90(child, target, h)
			child = h.LoadNext(child)
		}
	case 59: // Plasma
		durationRayStopSend4FEF90(record, who, 8, h.LoadDirectionLow(entryCaster), h)
	}
}
