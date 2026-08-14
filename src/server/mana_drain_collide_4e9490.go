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

type manaDrainManaSubHooks4EEBF0[O, U, P comparable] struct {
	classLow         func(O) uint8
	loadUpdateData   func(O) U
	godMode          func() bool
	loadManaCurrent  func(U) uint16
	storeManaPrev    func(U, uint16)
	storeManaCurrent func(U, uint16)
	loadPlayer       func(U) P
	loadProtection   func(P) uint32
	protectMana      func(uint32, int16)
}

// manaDrainManaSub4EEBF0 preserves the portion of GAME.EXE 004EEBF0 reached
// by ManaDrainCollide. Its amount is the zero-extended byte from collide data.
// In particular, the original reloads the new mana and updates the protection
// token by -amount only when newMana > amount; otherwise it uses -newMana.
// That counterintuitive branch is intentional and matches the machine code.
func manaDrainManaSub4EEBF0[O, U, P comparable](
	unit O,
	amount uint8,
	hooks manaDrainManaSubHooks4EEBF0[O, U, P],
) {
	var zeroObject O
	if unit == zeroObject {
		return
	}
	if hooks.classLow(unit)&0x04 == 0 {
		return
	}
	update := hooks.loadUpdateData(unit)
	if hooks.godMode() {
		return
	}

	current := hooks.loadManaCurrent(update)
	hooks.storeManaPrev(update, current)
	if int32(current) > int32(amount) {
		hooks.storeManaCurrent(update, current-uint16(amount))
	} else {
		hooks.storeManaCurrent(update, 0)
	}

	current = hooks.loadManaCurrent(update)
	delta := -int16(current)
	if int32(current) > int32(amount) {
		delta = -int16(amount)
	}
	player := hooks.loadPlayer(update)
	token := hooks.loadProtection(player)
	hooks.protectMana(token, delta)
}
