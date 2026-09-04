package server

type pixieTeleportHooks4FD050[Object any] struct {
	loadOwnerArg      func() Object
	loadPixieArg      func() Object
	loadOwnerXBits    func(Object) uint32
	loadOwnerYBits    func(Object) uint32
	storeNewPosXBits  func(Object, uint32)
	storeNewPosYBits  func(Object, uint32)
	storePosXBits     func(Object, uint32)
	storePosYBits     func(Object, uint32)
	storePrevPosXBits func(Object, uint32)
	storePrevPosYBits func(Object, uint32)
	moveUpdate        func(Object)
}

// pixieTeleport4FD050 preserves GAME.EXE 004FD050's exact argument,
// coordinate-load, coordinate-store, and delegated-call order. The owner and
// Pixie arguments are cached once, but the owner's position components are
// loaded independently before each of the six stores. Coordinates use raw
// uint32 bits so signed zero and NaN payloads cross the copy unchanged.
//
// There are deliberately no nil guards. The original first faults while
// loading owner X, or while storing NewPos.X when only the Pixie is nil. The
// move-update callback runs only after NewPos, PosVec, and PrevPos have each
// received X followed by Y.
func pixieTeleport4FD050[Object any](hooks pixieTeleportHooks4FD050[Object]) {
	owner := hooks.loadOwnerArg()
	pixie := hooks.loadPixieArg()

	x := hooks.loadOwnerXBits(owner)
	hooks.storeNewPosXBits(pixie, x)
	y := hooks.loadOwnerYBits(owner)
	hooks.storeNewPosYBits(pixie, y)

	x = hooks.loadOwnerXBits(owner)
	hooks.storePosXBits(pixie, x)
	y = hooks.loadOwnerYBits(owner)
	hooks.storePosYBits(pixie, y)

	x = hooks.loadOwnerXBits(owner)
	hooks.storePrevPosXBits(pixie, x)
	y = hooks.loadOwnerYBits(owner)
	hooks.storePrevPosYBits(pixie, y)

	hooks.moveUpdate(pixie)
}
