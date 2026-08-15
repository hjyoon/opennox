package opennox

const (
	playerObserverMonsterClassBit4EC420 = uint8(0x02)
	playerObserverSummonedStatus4EC420  = uint8(0x80)
)

type playerObserverFindGoodSlaveHooks4EC420[O comparable, D any] struct {
	loadOwner      func(O) O
	loadNextOwned  func(O) O
	loadClassByte  func(O) uint8
	loadUpdateData func(O) D
	loadStatusByte func(D) uint8
}

// playerObserverFindGoodSlaveContract4EC420 preserves GAME.EXE 004EC420.
// The current object's owner is used only as a non-nil guard. Traversal starts
// at the current object's next-owned link and otherwise follows the original
// low-byte Monster/summoned tests and live-link load ordering.
func playerObserverFindGoodSlaveContract4EC420[O comparable, D any](
	current O,
	hooks playerObserverFindGoodSlaveHooks4EC420[O, D],
) O {
	var zero O
	if current == zero || hooks.loadOwner(current) == zero {
		return zero
	}
	candidate := hooks.loadNextOwned(current)
	for candidate != zero {
		if hooks.loadClassByte(candidate)&playerObserverMonsterClassBit4EC420 != 0 {
			data := hooks.loadUpdateData(candidate)
			if hooks.loadStatusByte(data)&playerObserverSummonedStatus4EC420 != 0 {
				return candidate
			}
		}
		candidate = hooks.loadNextOwned(candidate)
	}
	return zero
}
