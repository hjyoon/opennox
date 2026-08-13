package server

type playerQuestSpawnHooks4E8210[U, G comparable, D, P any] struct {
	firstUnit          func() U
	nextUnit           func(U) U
	loadSoulGate       func(U) G
	loadCollideData    func(G) D
	loadLastUsedFrame  func(D) uint32
	storeSoulGate      func(U, G)
	loadSoulGatePos    func(G) P
	randomReachablePos func(float32, P) P
}

// playerQuestSpawn4E8210 preserves GAME.EXE 004E8210. It scans live player
// units for the SoulGate with the greatest unsigned last-used frame. Ties keep
// the first gate and frame zero never selects a gate because the running
// maximum starts at zero.
//
// Once selected, the gate is stored on the joining player before its live
// position is passed to the reachable-point search with the exact 60.0f
// radius. A failed selection leaves the joining player untouched and returns
// no point.
func playerQuestSpawn4E8210[U, G comparable, D, P any](
	joining U,
	hooks playerQuestSpawnHooks4E8210[U, G, D, P],
) (P, bool) {
	var (
		maxFrame uint32
		bestGate G
		zeroUnit U
		zeroGate G
		zeroPos  P
	)
	for unit := hooks.firstUnit(); unit != zeroUnit; unit = hooks.nextUnit(unit) {
		gate := hooks.loadSoulGate(unit)
		if gate == zeroGate {
			continue
		}
		data := hooks.loadCollideData(gate)
		frame := hooks.loadLastUsedFrame(data)
		if frame > maxFrame {
			maxFrame = frame
			bestGate = gate
		}
	}
	if bestGate == zeroGate {
		return zeroPos, false
	}
	hooks.storeSoulGate(joining, bestGate)
	pos := hooks.loadSoulGatePos(bestGate)
	return hooks.randomReachablePos(60.0, pos), true
}
