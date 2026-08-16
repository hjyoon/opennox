package server

const (
	objectDropOnlineFlag4ED790 = uint32(0x2000)
	objectDropQuestFlag4ED790  = uint32(0x1000)
	objectDropClassMask4ED790  = uint32(0x03001010)
	objectDropActiveFlag4ED790 = uint32(0x40)
)

// objectDropDispatchHooks4ED790 exposes the delayed argument and field loads
// in GAME.EXE 004ED790. The dropped item is cached at entry, while its Drop
// callback is loaded only after the optional flag store and unit refresh.
type objectDropDispatchHooks4ED790[O comparable, P, F any] struct {
	loadItemArg func() O
	gameFlag    func(uint32) int32
	loadClass   func(O) uint32
	loadFlags   func(O) uint32
	storeFlags  func(O, uint32)
	refreshUnit func(O)

	loadDrop     func(O) F
	hasDrop      func(F) bool
	loadPointArg func() P
	loadOwnerArg func() O
	callDrop     func(F, O, O, P) int32
	defaultDrop  func(O, O, P) int32
}

// objectDropDispatch4ED790 preserves GAME.EXE 004ED790. A nil item faults no
// later dependency. Online, non-Quest Food/Wand/Weapon/Armor objects are marked
// active and refreshed before the live Drop slot is read. A missing handler
// falls back to DefaultDrop, and either handler's full 32-bit result is kept.
func objectDropDispatch4ED790[O comparable, P, F any](hooks objectDropDispatchHooks4ED790[O, P, F]) int32 {
	var zero O
	item := hooks.loadItemArg()
	if item == zero {
		return 0
	}
	if hooks.gameFlag(objectDropOnlineFlag4ED790) != 0 &&
		hooks.gameFlag(objectDropQuestFlag4ED790) == 0 &&
		hooks.loadClass(item)&objectDropClassMask4ED790 != 0 {
		flags := hooks.loadFlags(item)
		hooks.storeFlags(item, flags|objectDropActiveFlag4ED790)
		hooks.refreshUnit(item)
	}

	drop := hooks.loadDrop(item)
	if hooks.hasDrop(drop) {
		point := hooks.loadPointArg()
		owner := hooks.loadOwnerArg()
		return hooks.callDrop(drop, owner, item, point)
	}
	point := hooks.loadPointArg()
	owner := hooks.loadOwnerArg()
	return hooks.defaultDrop(owner, item, point)
}
