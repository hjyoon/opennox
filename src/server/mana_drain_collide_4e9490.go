package server

type manaDrainCollideHooks4E9490[O, U, D comparable] struct {
	classLow         func(O) uint8
	loadUpdateData   func(O) U
	loadManaCurrent  func(U) uint16
	loadCollideData  func(O) D
	loadAmount       func(D) uint8
	subtractMana     func(O, uint8)
	loadSharedTimer  func(O) int16
	loadFrame        func() uint32
	loadFPS          func() uint32
	audio            func(O)
	storeSharedTimer func(O, uint16)
}

// manaDrainCollide4E9490 preserves GAME.EXE 004E9490. The target gates run
// before any source field is read. After the mana callback, Object+542 is
// sign-extended, frame arithmetic wraps at 32 bits, and sound 228 is emitted
// only when the unsigned elapsed value is strictly greater than FPS/2. The
// frame stored after sound is a live second read. The collision pointer is not
// observed.
func manaDrainCollide4E9490[O, U, D comparable, C any](
	source, target O,
	collision C,
	hooks manaDrainCollideHooks4E9490[O, U, D],
) {
	_ = collision
	var zeroObject O
	if target == zeroObject {
		return
	}
	if hooks.classLow(target)&0x04 == 0 {
		return
	}
	update := hooks.loadUpdateData(target)
	if hooks.loadManaCurrent(update) == 0 {
		return
	}

	data := hooks.loadCollideData(source)
	amount := hooks.loadAmount(data)
	hooks.subtractMana(target, amount)
	last := hooks.loadSharedTimer(source)
	frame := hooks.loadFrame()
	fps := hooks.loadFPS()
	if frame-uint32(int32(last)) <= fps>>1 {
		return
	}
	hooks.audio(source)
	hooks.storeSharedTimer(source, uint16(hooks.loadFrame()))
}
