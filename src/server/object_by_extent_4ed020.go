package server

const objectDestroyedFlagLow4ED020 = uint8(0x20)

// objectByExtentHooks4ED020 separates object identities from the original
// ABI32 list and object storage. Extent remains an unsigned 32-bit value on
// every host architecture.
type objectByExtentHooks4ED020[O comparable] struct {
	first         func() O
	loadExtentArg func() uint32
	loadFlagsLow  func(O) uint8
	loadExtent    func(O) uint32
	next          func(O) O
}

// objectByExtent4ED020 preserves GAME.EXE 004ED020. The active-list head is
// requested before the incoming Extent is loaded. Every candidate reads only
// the low flags byte first; destroyed candidates skip the Extent load. A match
// returns the exact candidate without requesting its successor.
func objectByExtent4ED020[O comparable](hooks objectByExtentHooks4ED020[O]) O {
	var zero O
	obj := hooks.first()
	if obj == zero {
		return zero
	}
	extent := hooks.loadExtentArg()
	for {
		if hooks.loadFlagsLow(obj)&objectDestroyedFlagLow4ED020 == 0 && hooks.loadExtent(obj) == extent {
			return obj
		}
		obj = hooks.next(obj)
		if obj == zero {
			return zero
		}
	}
}
