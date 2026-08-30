package server

const (
	networkGauntletPacketSize51BAD0 = int32(2)
	networkGauntletRespawn51BAD0    = uint8(3)
	networkGauntletExit51BAD0       = uint8(27)
	networkGauntletDeadFlag51BAD0   = uint32(0x8000)
)

type networkGauntletHooks51BAD0[O comparable, U, P any] struct {
	loadSubtype    func() uint8
	loadPlayer     func(U) P
	loadPlayerUnit func(P) O
	loadFlags      func(O) uint32
	clearField137  func(U)
	respawn        func(O)
	exit           func(O)
}

// networkGauntlet51BAD0 preserves GAME.EXE 0051CDFD..0051CE43. The Player
// pointer is cached once, while PlayerUnit is deliberately loaded again after
// Field137 is cleared and immediately before the respawn call.
func networkGauntlet51BAD0[O comparable, U, P any](
	unit O,
	update U,
	hooks networkGauntletHooks51BAD0[O, U, P],
) int32 {
	switch hooks.loadSubtype() {
	case networkGauntletRespawn51BAD0:
		player := hooks.loadPlayer(update)
		respawnUnit := hooks.loadPlayerUnit(player)
		var zero O
		if respawnUnit == zero {
			return networkGauntletPacketSize51BAD0
		}
		if hooks.loadFlags(respawnUnit)&networkGauntletDeadFlag51BAD0 == 0 {
			return networkGauntletPacketSize51BAD0
		}
		hooks.clearField137(update)
		hooks.respawn(hooks.loadPlayerUnit(player))
	case networkGauntletExit51BAD0:
		hooks.exit(unit)
	default:
		return -1
	}
	return networkGauntletPacketSize51BAD0
}
