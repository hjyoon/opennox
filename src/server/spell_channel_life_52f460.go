package server

// channelLifeHooks52F460 separates the original callback's observable loads
// from object identity. The PE32 record offsets are not valid on a 64-bit host.
type channelLifeHooks52F460[Record, Object comparable] struct {
	loadTarget    func(Record) Object
	loadFlags     func(Object) uint32
	loadMode      func(Record) uint32
	addMana       func(Object, int16)
	clearDamage   func(Object, int32)
	loadCaster    func(Record) Object
	testBuff      func(Object, int32) int32
	loadClass     func(Object) uint8
	positionDelta func(Object, Record) int32
	maxMana       func(Object) uint16
	currentMana   func(Object) uint16
	getHP         func(Object) uint16
	loadFraction  func(Record) float32
	loadLevel     func(Record) uint32
	coefficient   func(uint32) float64
	storeFraction func(Record, float32)
}

// spellChannelLifeUpdate52F460 follows GAME.EXE 0052F460's live read order.
// Target and caster are reloaded after calls that can mutate the record. The
// float accumulator is stored in Field72's bit pattern by the native adapter.
func spellChannelLifeUpdate52F460[Record, Object comparable](
	record Record, h channelLifeHooks52F460[Record, Object],
) int32 {
	var nilObject Object
	target := h.loadTarget(record)
	if target == nilObject || h.loadFlags(target)&0x8020 != 0 {
		return 1
	}
	if h.loadMode(record) != 0 {
		h.addMana(target, 20)
		h.clearDamage(h.loadTarget(record), 20)
		return 1
	}

	caster := h.loadCaster(record)
	if caster != nilObject && h.testBuff(caster, 8) != 0 {
		return 1
	}
	target = h.loadTarget(record)
	if h.loadClass(target)&2 != 0 && h.positionDelta(target, record) != 0 {
		return 1
	}
	if h.maxMana(h.loadTarget(record)) == h.currentMana(h.loadTarget(record)) {
		return 1
	}
	if h.getHP(h.loadCaster(record)) <= 1 {
		return 1
	}
	if h.getHP(h.loadCaster(record)) != 0 {
		fraction := h.loadFraction(record)
		value := float32(h.coefficient(h.loadLevel(record)-1) + float64(fraction))
		amount := spellDurationRoundNearestEven(value)
		h.storeFraction(record, float32(float64(value)-float64(amount)))
		h.addMana(h.loadTarget(record), int16(amount))
		h.clearDamage(h.loadCaster(record), 1)
	}
	return 0
}
