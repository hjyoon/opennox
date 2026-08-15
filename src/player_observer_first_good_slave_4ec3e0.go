package opennox

const (
	playerObserverMonsterClassBit4EC3E0 = uint8(0x02)
	playerObserverSummonedStatus4EC3E0  = uint8(0x80)
)

type playerObserverFindGoodSlave2Hooks4EC3E0[O comparable, D any] struct {
	loadFirstOwned func(O) O
	loadClassByte  func(O) uint8
	loadUpdateData func(O) D
	loadStatusByte func(D) uint8
	loadNextOwned  func(O) O
}

// playerObserverFindGoodSlave2Contract4EC3E0 preserves GAME.EXE 004EC3E0.
// Only Monster candidates load update data. A rejected candidate's next-owned
// link is read after its class and optional status, while an accepted candidate
// returns without reading that link.
func playerObserverFindGoodSlave2Contract4EC3E0[O comparable, D any](
	owner O,
	hooks playerObserverFindGoodSlave2Hooks4EC3E0[O, D],
) O {
	var zero O
	if owner == zero {
		return zero
	}
	candidate := hooks.loadFirstOwned(owner)
	for candidate != zero {
		if hooks.loadClassByte(candidate)&playerObserverMonsterClassBit4EC3E0 != 0 {
			data := hooks.loadUpdateData(candidate)
			if hooks.loadStatusByte(data)&playerObserverSummonedStatus4EC3E0 != 0 {
				return candidate
			}
		}
		candidate = hooks.loadNextOwned(candidate)
	}
	return zero
}
