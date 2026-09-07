package server

import "encoding/binary"

const (
	durationRayStartMessage4FF130   = byte(0x9e)
	durationRayStartRecipient4FF130 = int32(255)
	durationRayStartRemove4FF130    = int32(1)
	durationRayStartUnmark4FF130    = uint32(2)
)

// DurationRayStartHooks4FF130 exposes every observable record, object,
// packet, and callback access in GAME.EXE 004FF130. Comparable tokens retain
// native null and identity semantics without inheriting the PE32 pointer
// width. UnitCode deliberately returns a full dword; the packet stores only
// its low word, exactly like the original MOV stores.
type DurationRayStartHooks4FF130[Record, Object comparable] struct {
	LoadSpell        func(Record) uint32
	LoadLevelLow     func(Record) byte
	LoadCaster       func(Record) Object
	LoadDirectionLow func(Object) byte
	LoadTarget       func(Record) Object
	LoadSub108       func(Record) Record
	LoadNext         func(Record) Record
	UnitCode         func(Object) uint32
	SendPacket       func(int32, [7]byte, Object, int32) int32
	UnmarkMinimap    func(Object, uint32)
}

func durationRayStartSend4FF130[Record, Object comparable](
	record Record,
	typ, value byte,
	h DurationRayStartHooks4FF130[Record, Object],
) {
	var packet [7]byte
	packet[0] = durationRayStartMessage4FF130
	packet[1] = typ
	packet[2] = value

	target := h.LoadTarget(record)
	var nilObject Object
	if target == nilObject {
		return
	}
	binary.LittleEndian.PutUint16(packet[5:7], uint16(h.UnitCode(target)))
	caster := h.LoadCaster(record)
	binary.LittleEndian.PutUint16(packet[3:5], uint16(h.UnitCode(caster)))
	_ = h.SendPacket(
		durationRayStartRecipient4FF130,
		packet,
		nilObject,
		durationRayStartRemove4FF130,
	)
	caster = h.LoadCaster(record)
	h.UnmarkMinimap(caster, durationRayStartUnmark4FF130)
	target = h.LoadTarget(record)
	h.UnmarkMinimap(target, durationRayStartUnmark4FF130)
}

func durationRayStartGreaterHeal4FF130[Record, Object comparable](
	record Record,
	h DurationRayStartHooks4FF130[Record, Object],
) {
	target := h.LoadTarget(record)
	caster := h.LoadCaster(record)
	if caster == target {
		return
	}

	// GreaterHeal has its own executable packet path. It reverses the two
	// code fields used by the ordinary path and delays Level until after both
	// unit-code callbacks.
	var packet [7]byte
	packet[0] = durationRayStartMessage4FF130
	packet[1] = 6
	binary.LittleEndian.PutUint16(packet[3:5], uint16(h.UnitCode(target)))
	caster = h.LoadCaster(record)
	binary.LittleEndian.PutUint16(packet[5:7], uint16(h.UnitCode(caster)))
	packet[2] = h.LoadLevelLow(record)
	var nilObject Object
	_ = h.SendPacket(
		durationRayStartRecipient4FF130,
		packet,
		nilObject,
		durationRayStartRemove4FF130,
	)
	caster = h.LoadCaster(record)
	h.UnmarkMinimap(caster, durationRayStartUnmark4FF130)
	target = h.LoadTarget(record)
	h.UnmarkMinimap(target, durationRayStartUnmark4FF130)
}

// DurationRayStart4FF130 preserves GAME.EXE 004FF130's exact dispatch and
// side-effect order. Spell is the first record access and a full dword; only
// IDs 7, 9, 22, 24, 35, 43, and 59 have cases. Ordinary packets load their
// value before Target48, encode a live Caster16 code in bytes 3:5 and the
// earlier Target48 code in bytes 5:7, then reload both objects around the
// minimap callbacks. Plasma reads Direction1 from the Caster16 value loaded
// before Target48. GreaterHeal loads Target48 before Caster16, compares their
// full native identities, reverses the packet's code fields, and delays Level
// until after both unit-code callbacks.
//
// ChainLightning walks the live Sub108 list. It recurses on each child and
// loads that child's Next link only after recursion returns, so recursive
// mutation can redirect traversal. No record, caster, GreaterHeal target,
// callback, list-cycle, or post-callback validity guard is added. The decoded
// callers all discard EAX, so this restored behavior boundary has no return.
func DurationRayStart4FF130[Record, Object comparable](
	record Record,
	h DurationRayStartHooks4FF130[Record, Object],
) {
	switch h.LoadSpell(record) {
	default:
		return
	case 7: // ChainLightningBolt
		durationRayStartSend4FF130(record, 3, h.LoadLevelLow(record), h)
	case 9: // Charm
		durationRayStartSend4FF130(record, 2, h.LoadLevelLow(record), h)
	case 22: // DrainMana
		durationRayStartSend4FF130(record, 5, h.LoadLevelLow(record), h)
	case 24: // Lightning
		durationRayStartSend4FF130(record, 4, h.LoadLevelLow(record), h)
	case 35: // GreaterHeal
		durationRayStartGreaterHeal4FF130(record, h)
	case 43: // ChainLightning
		var nilRecord Record
		child := h.LoadSub108(record)
		for child != nilRecord {
			DurationRayStart4FF130(child, h)
			child = h.LoadNext(child)
		}
	case 59: // Plasma
		caster := h.LoadCaster(record)
		durationRayStartSend4FF130(record, 1, h.LoadDirectionLow(caster), h)
	}
}
