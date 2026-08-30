package server

const (
	winkGameBallTypeName4F7DF0 = "GameBall"
	winkGameBallFlag4F7DF0     = uint32(0x40)
	winkGameBallForce4F7DF0    = float32(100)
	winkGameBallAudio4F7DF0    = uint32(926)
	winkGameBallStatus4F7DF0   = uint8(1)
)

type winkGameBallReleaseHooks4F7DF0[O comparable] struct {
	loadTypeCache  func() uint32
	lookupType     func(string) uint32
	storeTypeCache func(uint32)

	loadFirstOwned func(O) O
	loadTypeInd    func(O) uint16
	loadNextOwned  func(O) O

	loadFlags   func(O) uint32
	storeFlags  func(O, uint32)
	applyForce  func(O, O, float32)
	storeObj130 func(O, O)
	clearOwner  func(O)
	audio       func(uint32, O, int32, uint32)
	ballStatus  func(uint8, uint16) int32
}

// winkGameBallRelease4F7DF0 preserves GAME.EXE 004F7DF0. The private type
// cache is loaded once and initialized before the player is dereferenced. The
// first matching object in the owned-object list is released after its low
// flag bit 0x40 is cleared. All release callbacks retain the original order.
func winkGameBallRelease4F7DF0[O comparable](
	player O,
	hooks winkGameBallReleaseHooks4F7DF0[O],
) int32 {
	typeInd := hooks.loadTypeCache()
	if typeInd == 0 {
		typeInd = hooks.lookupType(winkGameBallTypeName4F7DF0)
		hooks.storeTypeCache(typeInd)
	}

	var zero O
	ball := hooks.loadFirstOwned(player)
	for ball != zero {
		if uint32(hooks.loadTypeInd(ball)) == typeInd {
			break
		}
		ball = hooks.loadNextOwned(ball)
	}
	if ball == zero {
		return 0
	}

	flags := hooks.loadFlags(ball)
	hooks.storeFlags(ball, flags&^winkGameBallFlag4F7DF0)
	hooks.applyForce(player, ball, winkGameBallForce4F7DF0)
	hooks.storeObj130(ball, zero)
	hooks.clearOwner(ball)
	hooks.audio(winkGameBallAudio4F7DF0, player, 0, 0)
	hooks.ballStatus(winkGameBallStatus4F7DF0, 0)
	return 1
}
