package server

const (
	controlledCreatureSmallBit500D50  = uint8(0x01)
	controlledCreatureMediumBit500D50 = uint8(0x02)
)

// controlledCreatureCountHooks500D10 exposes each observable load and call in
// GAME.EXE 00500D10. Object handles remain native-width values; only the low
// subclass byte interpreted by the private helper at 00500D50 is narrowed.
type controlledCreatureCountHooks500D10[O comparable] struct {
	loadFirstOwned  func(O) O
	isMonitored     func(owner, unit O) bool
	loadSubclassLow func(O) uint8
	loadNextOwned   func(O) O
}

// controlledCreatureSize500D50 preserves the bit precedence in GAME.EXE
// 00500D50: small wins over medium, and every other subclass consumes four
// controlled-creature slots.
func controlledCreatureSize500D50(subclassLow uint8) uint32 {
	if subclassLow&controlledCreatureSmallBit500D50 != 0 {
		return 1
	}
	if subclassLow&controlledCreatureMediumBit500D50 != 0 {
		return 2
	}
	return 4
}

func controlledCreatureCountAdd500D10(count, size uint32) uint32 {
	return count + size
}

// controlledCreatureCount500D10 walks the live owned-unit links in the same
// order as GAME.EXE 00500D10. It deliberately has no null-owner preflight,
// loads each successor after the monitor and size work, and returns the exact
// signed bit pattern produced by the original wrapping PE32 accumulator.
func controlledCreatureCount500D10[O comparable](
	owner O,
	hooks controlledCreatureCountHooks500D10[O],
) int32 {
	var zero O
	count := uint32(0)
	for unit := hooks.loadFirstOwned(owner); unit != zero; unit = hooks.loadNextOwned(unit) {
		if hooks.isMonitored(owner, unit) {
			size := controlledCreatureSize500D50(hooks.loadSubclassLow(unit))
			count = controlledCreatureCountAdd500D10(count, size)
		}
	}
	return int32(count)
}
