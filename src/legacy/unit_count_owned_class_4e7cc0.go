package legacy

type countOwnedClassHooks4E7CC0[O comparable] struct {
	loadFirst func(O) O
	loadClass func(O) uint32
	loadNext  func(O) O
}

func incrementOwnedClassCount4E7CC0(count int32) int32 {
	return count + 1
}

// countOwnedClass4E7CC0 preserves GAME.EXE 004E7CC0. A nil owner or zero
// class mask returns before the owned-head load. Each node loads its complete
// 32-bit class, counts any non-empty intersection with the mask, then loads
// the successor. The result is the original wrapping 32-bit EAX count.
func countOwnedClass4E7CC0[O comparable](
	owner O,
	classMask uint32,
	hooks countOwnedClassHooks4E7CC0[O],
) int32 {
	var zero O
	if owner == zero || classMask == 0 {
		return 0
	}

	count := int32(0)
	for obj := hooks.loadFirst(owner); obj != zero; obj = hooks.loadNext(obj) {
		if hooks.loadClass(obj)&classMask != 0 {
			count = incrementOwnedClassCount4E7CC0(count)
		}
	}
	return count
}
