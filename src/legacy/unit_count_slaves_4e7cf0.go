package legacy

type countSlavesHooks4E7CF0[O comparable] struct {
	loadFirst    func(O) O
	loadClass    func(O) uint32
	loadSubclass func(O) uint32
	loadNext     func(O) O
}

func incrementSlaveCount4E7CF0(count int32) int32 {
	return count + 1
}

// countSlaves4E7CF0 preserves GAME.EXE 004E7CF0. Nil owner, zero class mask,
// and zero subclass mask return in that order before the owned-head load.
// Each node loads its complete 32-bit class first; only a class intersection
// permits the complete 32-bit subclass load. The successor is loaded after
// either path. The result is the original wrapping 32-bit EAX count.
func countSlaves4E7CF0[O comparable](
	owner O,
	classMask uint32,
	subclassMask uint32,
	hooks countSlavesHooks4E7CF0[O],
) int32 {
	var zero O
	if owner == zero || classMask == 0 || subclassMask == 0 {
		return 0
	}

	count := int32(0)
	for obj := hooks.loadFirst(owner); obj != zero; obj = hooks.loadNext(obj) {
		if hooks.loadClass(obj)&classMask != 0 && hooks.loadSubclass(obj)&subclassMask != 0 {
			count = incrementSlaveCount4E7CF0(count)
		}
	}
	return count
}
