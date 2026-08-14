package server

const (
	barrelCollideDelay4EAAA0 = uint32(3)
	barrelCollideSound4EAAA0 = uint32(281)
)

type barrelCollideHooks4EAAA0[O any] struct {
	loadFrame     func() uint32
	loadLastFrame func(O) uint32
	storeFrame    func(O, uint32)
	audio         func(uint32, O)
}

// barrelCollide4EAAA0 preserves GAME.EXE 004EAAA0. The registered collision
// callback ignores target and collision. It loads the live frame before the
// source timestamp, adds three with uint32 wrapping, and emits sound 281 only
// when frame is strictly greater than that wrapped threshold. The timestamp
// store precedes the audio callback.
func barrelCollide4EAAA0[O, T, C any](
	source O, _ T, _ C,
	hooks barrelCollideHooks4EAAA0[O],
) {
	frame := hooks.loadFrame()
	last := hooks.loadLastFrame(source)
	if frame <= last+barrelCollideDelay4EAAA0 {
		return
	}
	hooks.storeFrame(source, frame)
	hooks.audio(barrelCollideSound4EAAA0, source)
}
