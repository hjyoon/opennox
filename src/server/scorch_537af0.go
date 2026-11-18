package server

import "github.com/opennox/libs/types"

const scorchQuestFlag537AF0 = uint32(0x1000)

var scorchTypeIDs537AF0 = [...]string{
	"ScorchMarkFloorSmallA",
	"ScorchMarkFloorMediumA",
	"ScorchMarkFloorLargeB",
}

type scorchTypeCache537BD0 struct {
	ready uint32
	ids   [len(scorchTypeIDs537AF0)]uint32
}

func (cache *scorchTypeCache537BD0) initialize(typeIndex func(string) int) int {
	for i, name := range scorchTypeIDs537AF0 {
		cache.ids[i] = uint32(typeIndex(name))
	}
	cache.ready = 1
	return int(cache.ids[len(cache.ids)-1])
}

type makeScorchHooks537AF0[O comparable] struct {
	typeIndex      func(string) int
	newObject      func(int) O
	createObjectAt func(O, types.Pointf)
	gameFlag       func(uint32) int32
	randomInt      func(int, int) int
	loadGameFPS    func() uint32
	setDecay       func(O, uint32)
}

func makeScorch537AF0[O comparable](pos types.Pointf, kind int, cache *scorchTypeCache537BD0, hooks makeScorchHooks537AF0[O]) {
	if cache.ready == 0 {
		cache.initialize(hooks.typeIndex)
	}
	if uint(kind) >= uint(len(cache.ids)) {
		return
	}
	// The original indexes a one-element table with Random(0, 0). It still
	// advances the deterministic logic RNG, so retain that otherwise redundant
	// call before allocating the object.
	_ = hooks.randomInt(0, 0)
	obj := hooks.newObject(int(cache.ids[kind]))
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
