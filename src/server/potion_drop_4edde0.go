package server

const (
	potionDropAudio4EDDE0     = uint32(833)
	potionDropCoopFlag4EDDE0  = uint32(0x00000800)
	potionDropQuestFlag4EDDE0 = uint32(0x00001000)
	potionDropSeconds4EDDE0   = uint32(25)
)

// potionDropHooks4EDDE0 exposes GAME.EXE 004EDDE0's exact argument-load and
// callback order. The point, owner, and item arguments are cached before
// DefaultDrop; game flags and FPS remain live callback reads after the audio
// event.
type potionDropHooks4EDDE0[O, P comparable] struct {
	loadPointArg func() P
	loadOwnerArg func() O
	loadItemArg  func() O

	defaultDrop func(O, O, P) int32
	audio       func(uint32, O, int32, uint32)
	gameFlag    func(uint32) int32
	loadGameFPS func() uint32
	setDecay    func(O, uint32)
}

// potionDrop4EDDE0 preserves GAME.EXE 004EDDE0. Every condition tests the
// whole 32-bit callback result and every public result is canonical zero or
// one. The x86 LEA sequence computes 25*FPS modulo 2^32.
func potionDrop4EDDE0[O, P comparable](hooks potionDropHooks4EDDE0[O, P]) int32 {
	point := hooks.loadPointArg()
	owner := hooks.loadOwnerArg()
	item := hooks.loadItemArg()
	if hooks.defaultDrop(owner, item, point) == 0 {
		return 0
	}

	hooks.audio(potionDropAudio4EDDE0, item, 0, 0)
	if hooks.gameFlag(potionDropCoopFlag4EDDE0) != 0 {
		return 1
	}
	if hooks.gameFlag(potionDropQuestFlag4EDDE0) != 0 {
		return 1
	}
	hooks.setDecay(item, hooks.loadGameFPS()*potionDropSeconds4EDDE0)
	return 1
}
