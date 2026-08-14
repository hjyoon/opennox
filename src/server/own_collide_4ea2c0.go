package server

const ownCollidePlayerClassBit4EA2C0 = uint32(0x04)

type ownCollideHooks4EA2C0[O comparable] struct {
	loadTargetClass  func(O) uint32
	loadSourceOwner  func(O) O
	loadFrame        func() uint32
	storeSourceFrame func(O, uint32)
	setOwner         func(O, O)
}

// ownCollide4EA2C0 preserves GAME.EXE 004EA2C0. It does not inspect the
// registered callback's collision argument. The target and its low Player
// class bit are checked before the source is touched. A source is adopted
// only when its one cached owner load is nil; the frame store precedes the
// owner callback.
func ownCollide4EA2C0[O comparable](source, target O, hooks ownCollideHooks4EA2C0[O]) {
	var zero O
	if target == zero {
		return
	}
	if hooks.loadTargetClass(target)&ownCollidePlayerClassBit4EA2C0 == 0 {
		return
	}

	owner := hooks.loadSourceOwner(source)
	if owner == target {
		return
	}
	if owner != zero {
		return
	}

	frame := hooks.loadFrame()
	hooks.storeSourceFrame(source, frame)
	hooks.setOwner(target, source)
}
