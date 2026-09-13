package server

const (
	spellOvalShieldBuff531490        = int32(27)
	spellOvalShieldBlockedBuff5314F0 = int32(8)
	spellOvalShieldDuration531490    = uint32(20)
)

// spellOvalShieldHooks531490 keeps the original callback's live record reads
// separate from native-width object identities. GAME.EXE stores these fields
// at PE32 offsets; the Go record must never be decoded through those offsets.
type spellOvalShieldHooks531490[Record, Object comparable] struct {
	loadFPS         func() uint32
	loadLevel       func(Record) uint32
	loadTarget      func(Record) Object
	loadClass       func(Object) uint8
	loadFlags       func(Object) uint32
	cancelOffensive func(Object)
	applyBuff       func(Object, int32, int16, int8)
	loadFrame       func() uint32
	storeFrame      func(Record, uint32)
	testBuff        func(Object, int32) int32
	positionDelta   func(Object, Record) int32
	buffOff         func(Object, int32)
}

// spellOvalShieldCreate531490 follows GAME.EXE 00531490. Both arithmetic
// steps wrap in a dword; the buff service receives only the low word of the
// duration and the low byte of the level.
func spellOvalShieldCreate531490[Record, Object comparable](record Record, h spellOvalShieldHooks531490[Record, Object]) int32 {
	duration := spellOvalShieldDuration531490 * h.loadFPS() * h.loadLevel(record)
	target := h.loadTarget(record)
	var nilObject Object
	if target == nilObject || h.loadClass(target)&4 == 0 {
		return 1
	}
	h.cancelOffensive(target)
	// The original reloads Level and Target after cancellation, so a callback
	// that changes either field affects buff application but not duration.
	power := int8(h.loadLevel(record))
	target = h.loadTarget(record)
	h.applyBuff(target, spellOvalShieldBuff531490, int16(duration), power)
	h.storeFrame(record, duration+h.loadFrame())
	return 0
}

// spellOvalShieldUpdate5314F0 follows GAME.EXE 005314F0. The target is
// reloaded after the buff test and after the optional monster position test.
func spellOvalShieldUpdate5314F0[Record, Object comparable](record Record, h spellOvalShieldHooks531490[Record, Object]) int32 {
	var nilObject Object
	target := h.loadTarget(record)
	if target == nilObject || h.testBuff(target, spellOvalShieldBlockedBuff5314F0) != 0 {
		return 1
	}
	target = h.loadTarget(record)
	if target != nilObject && h.loadClass(target)&2 != 0 && h.positionDelta(target, record) != 0 {
		return 1
	}
	target = h.loadTarget(record)
	if target == nilObject {
		// A target cleared during a callback ends the duration safely. The
		// PE32 routine would dereference the now-null pointer here.
		return 1
	}
	if h.loadFlags(target)&0x8020 != 0 {
		return 1
	}
	return 0
}

// spellOvalShieldDestroy531560 follows GAME.EXE 00531560. The original's
// return value is not observed by the duration-spell destroy dispatcher.
func spellOvalShieldDestroy531560[Record, Object comparable](record Record, h spellOvalShieldHooks531490[Record, Object]) {
	target := h.loadTarget(record)
	var nilObject Object
	if target != nilObject {
		h.buffOff(target, spellOvalShieldBuff531490)
	}
}
