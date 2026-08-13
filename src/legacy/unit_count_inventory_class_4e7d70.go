package legacy

type countInventoryClassHooks4E7D70[O comparable] struct {
	loadFirst func(O) O
	loadClass func(O) uint32
	loadNext  func(O) O
}

func incrementInventoryClassCount4E7D70(count int32) int32 {
	return count + 1
}

// countInventoryClass4E7D70 preserves GAME.EXE 004E7D70. A nil owner or
// zero class mask returns before the inventory-head load. Each node loads its
// complete 32-bit class, counts any non-empty intersection with the mask,
// then loads the successor. The result is the original wrapping 32-bit EAX
// count.
func countInventoryClass4E7D70[O comparable](
	owner O,
	classMask uint32,
	hooks countInventoryClassHooks4E7D70[O],
) int32 {
	var zero O
	if owner == zero || classMask == 0 {
		return 0
	}

	count := int32(0)
	for obj := hooks.loadFirst(owner); obj != zero; obj = hooks.loadNext(obj) {
		if hooks.loadClass(obj)&classMask != 0 {
			count = incrementInventoryClassCount4E7D70(count)
		}
	}
	return count
}
