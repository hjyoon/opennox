package server

const objectDestroyedFlagLow4E7D30 uint8 = 0x20

type countInventoryTypeHooks4E7D30[O comparable] struct {
	loadFirst    func(O) O
	loadType     func(O) uint16
	loadFlagsLow func(O) uint8
	loadNext     func(O) O
}

func incrementInventoryTypeCount4E7D30(count int32) int32 {
	return count + 1
}

// countInventoryType4E7D30 preserves GAME.EXE 004E7D30. A nil owner returns
// before the first-inventory load. A zero query counts every node without
// reading its type or flags. A nonzero query reads the zero-extended 16-bit
// type first, reads only the low flags byte on a type match, and excludes
// Destroyed. Every visited node loads its live successor after processing.
// The result is the original wrapping 32-bit EAX count.
func countInventoryType4E7D30[O comparable](
	owner O,
	typeInd int32,
	hooks countInventoryTypeHooks4E7D30[O],
) int32 {
	var zero O
	if owner == zero {
		return 0
	}

	count := int32(0)
	for obj := hooks.loadFirst(owner); obj != zero; obj = hooks.loadNext(obj) {
		if typeInd == 0 {
			count = incrementInventoryTypeCount4E7D30(count)
			continue
		}
		if int32(hooks.loadType(obj)) != typeInd {
			continue
		}
		if hooks.loadFlagsLow(obj)&objectDestroyedFlagLow4E7D30 == 0 {
			count = incrementInventoryTypeCount4E7D30(count)
		}
	}
	return count
}
