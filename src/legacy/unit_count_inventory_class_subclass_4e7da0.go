package legacy

type countInventoryClassSubclassHooks4E7DA0[O comparable] struct {
	loadFirst    func(O) O
	loadClass    func(O) uint32
	loadSubclass func(O) uint32
	loadNext     func(O) O
}

func incrementInventoryClassSubclassCount4E7DA0(count int32) int32 {
	return count + 1
}

// countInventoryClassSubclass4E7DA0 preserves GAME.EXE 004E7DA0. Nil owner,
// zero class mask, and zero subclass mask return in that order before the
// inventory-head load. Each node loads its complete 32-bit class first; only
// a class intersection permits the complete 32-bit subclass load. The
// successor is loaded after either path. The result is the original wrapping
// 32-bit EAX count.
func countInventoryClassSubclass4E7DA0[O comparable](
	owner O,
	classMask uint32,
	subclassMask uint32,
	hooks countInventoryClassSubclassHooks4E7DA0[O],
) int32 {
	var zero O
	if owner == zero || classMask == 0 || subclassMask == 0 {
		return 0
	}

	count := int32(0)
	for obj := hooks.loadFirst(owner); obj != zero; obj = hooks.loadNext(obj) {
		if hooks.loadClass(obj)&classMask != 0 && hooks.loadSubclass(obj)&subclassMask != 0 {
			count = incrementInventoryClassSubclassCount4E7DA0(count)
		}
	}
	return count
}
