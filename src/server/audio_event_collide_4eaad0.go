package server

const (
	audioEventCollidePlayerClass4EAAD0 = uint8(0x04)
	audioEventCollideDelay4EAAD0       = uint32(30)
)

type audioEventCollideHooks4EAAD0[O comparable, D any] struct {
	classLow        func(O) uint8
	loadFrame       func() uint32
	loadLastFrame   func(O) uint32
	storeFrame      func(O, uint32)
	loadCollideData func(O) D
	loadSound       func(D) uint32
	audio           func(uint32, O)
}

// audioEventCollide4EAAD0 preserves GAME.EXE 004EAAD0. A nil or non-Player
// target returns before touching the source. The live frame is loaded once,
// source timestamp plus 30 wraps at uint32 width, and a successful unsigned
// strict-greater comparison stores the cached frame before loading collide
// data and emitting its sound. The collision pointer is not observed.
func audioEventCollide4EAAD0[O comparable, D, C any](
	source, target O,
	_ C,
	hooks audioEventCollideHooks4EAAD0[O, D],
) {
	var zero O
	if target == zero {
		return
	}
	if hooks.classLow(target)&audioEventCollidePlayerClass4EAAD0 == 0 {
		return
	}

	frame := hooks.loadFrame()
	last := hooks.loadLastFrame(source)
	if frame <= last+audioEventCollideDelay4EAAD0 {
		return
	}
	hooks.storeFrame(source, frame)
	data := hooks.loadCollideData(source)
	soundID := hooks.loadSound(data)
	hooks.audio(soundID, source)
}
