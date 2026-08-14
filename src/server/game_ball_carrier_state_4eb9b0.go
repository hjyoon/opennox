package server

type gameBallCarrierStateHooks4EB9B0[O comparable, D any] struct {
	loadUpdateData func(O) D
	findPlayer     func(O) O
	loadClassLow   func(O) uint8
	storeCarrier   func(D, O)
	loadTeamID     func(O) uint8
	storeTeamID    func(D, uint32)
	loadFrame      func() uint32
	storeFrame     func(D, uint32)
}

// gameBallCarrierState4EB9B0 preserves GAME.EXE 004EB9B0. The ball's
// UpdateData pointer is cached before the target is examined. A nil target,
// failed owner-chain lookup, or non-Player result clears only the carrier and
// zero-extended team fields, retaining the previous frame. The non-Player
// terminal result is still returned. Success stores the carrier before reading
// its live team ID, then reads and stores the current frame.
func gameBallCarrierState4EB9B0[O comparable, D any](
	ball, target O,
	hooks gameBallCarrierStateHooks4EB9B0[O, D],
) O {
	data := hooks.loadUpdateData(ball)
	var zero O
	result := target
	if target != zero {
		result = hooks.findPlayer(target)
		if result != zero && hooks.loadClassLow(result)&0x4 != 0 {
			hooks.storeCarrier(data, result)
			hooks.storeTeamID(data, uint32(hooks.loadTeamID(result)))
			hooks.storeFrame(data, hooks.loadFrame())
			return result
		}
	}
	hooks.storeCarrier(data, zero)
	hooks.storeTeamID(data, 0)
	return result
}
