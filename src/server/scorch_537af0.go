package server

import "github.com/opennox/libs/types"

const scorchQuestFlag537AF0 = uint32(0x1000)

var scorchTypeIDs537AF0 = [...]string{
	"ScorchMarkFloorSmallA",
	"ScorchMarkFloorMediumA",
	"ScorchMarkFloorLargeB",
}

type makeScorchHooks537AF0[O comparable] struct {
	newObject      func(string) O
	createObjectAt func(O, types.Pointf)
	gameFlag       func(uint32) int32
	randomInt      func(int, int) int
	loadGameFPS    func() uint32
	setDecay       func(O, uint32)
}

// makeScorch537AF0 preserves GAME.EXE 00537AF0 without using the original
// PE32 type-ID cache. The three cache entries map directly to stable object
// type names, which also avoids overlapping native pointers with 32-bit IDs
// in the legacy memory blob on 64-bit hosts.
func makeScorch537AF0[O comparable](pos types.Pointf, kind int, hooks makeScorchHooks537AF0[O]) {
	if kind < 0 || kind >= len(scorchTypeIDs537AF0) {
		return
	}
	// The original indexes a one-element table with Random(0, 0). It still
	// advances the deterministic logic RNG, so retain that otherwise redundant
	// call before allocating the object.
	_ = hooks.randomInt(0, 0)
	obj := hooks.newObject(scorchTypeIDs537AF0[kind])
	var zero O
	if obj == zero {
		return
	}
	hooks.createObjectAt(obj, pos)

	minimum, maximum := 10, 20
	if hooks.gameFlag(scorchQuestFlag537AF0) != 0 {
		minimum, maximum = 5, 8
	}
	seconds := hooks.randomInt(minimum, maximum)
	hooks.setDecay(obj, hooks.loadGameFPS()*uint32(seconds))
}
