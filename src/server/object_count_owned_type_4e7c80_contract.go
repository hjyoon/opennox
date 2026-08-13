package server

const objectDestroyedFlagLow4E7C80 uint8 = 0x20

type countOwnedTypeHooks4E7C80[O comparable] struct {
	loadFirst    func(O) O
	loadType     func(O) uint16
	loadFlagsLow func(O) uint8
	loadNext     func(O) O
}

func incrementOwnedTypeCount4E7C80(count int32) int32 {
	return count + 1
}

// countOwnedType4E7C80 preserves GAME.EXE 004E7C80. A nil owner returns
// before the first-owned load. Each non-nil node loads its zero-extended
// 16-bit type first, loads only the low flags byte on a type match, excludes
// Destroyed, and loads the successor after processing the current node. The
// result is the original wrapping 32-bit EAX count.
func countOwnedType4E7C80[O comparable](
	owner O,
	typeInd int32,
	hooks countOwnedTypeHooks4E7C80[O],
) int32 {
	var zero O
	if owner == zero {
		return 0
	}

	count := int32(0)
	for obj := hooks.loadFirst(owner); obj != zero; obj = hooks.loadNext(obj) {
		if int32(hooks.loadType(obj)) != typeInd {
			continue
		}
		if hooks.loadFlagsLow(obj)&objectDestroyedFlagLow4E7C80 == 0 {
			count = incrementOwnedTypeCount4E7C80(count)
		}
	}
	return count
}
