package server

const (
	flagCollideCTFMode4EA400      = uint32(0x20)
	flagCollideBallMode4EA400     = uint32(0x40)
	flagCollideDeadFlag4EA400     = uint32(0x8000)
	flagCollidePlayerClass4EA400  = uint8(0x04)
	flagCollideGameBallName4EA400 = "GameBall"
)

type flagCollideHooks4EA400[O comparable, C any] struct {
	loadFlags         func(O) uint32
	hasGameFlag       func(uint32) int32
	loadGameBallCache func() uint32
	lookupGameBall    func(string) uint32
	storeGameBall     func(uint32)
	loadTypeInd       func(O) uint16
	loadClassLow      func(O) uint8
	pickupCTF         func(O, O, C)
	pickupGameBall    func(O, O, C)
}

// flagCollide4EA400 preserves the routing callback at GAME.EXE 004EA400.
// A nil target returns before the source or collision argument is inspected.
// CTF has strict priority over FlagBall. The FlagBall path populates its type
// cache before reading the target TypeInd and reads class only after a type
// mismatch. Both handlers receive the original three callback arguments.
func flagCollide4EA400[O comparable, C any](
	source, target O,
	collision C,
	hooks flagCollideHooks4EA400[O, C],
) {
	var zero O
	if target == zero {
		return
	}
	if hooks.loadFlags(target)&flagCollideDeadFlag4EA400 != 0 {
		return
	}
	if hooks.hasGameFlag(flagCollideCTFMode4EA400) != 0 {
		if hooks.loadClassLow(target)&flagCollidePlayerClass4EA400 != 0 {
			hooks.pickupCTF(source, target, collision)
		}
		return
	}
	if hooks.hasGameFlag(flagCollideBallMode4EA400) == 0 {
		return
	}
	gameBall := hooks.loadGameBallCache()
	if gameBall == 0 {
		gameBall = hooks.lookupGameBall(flagCollideGameBallName4EA400)
		hooks.storeGameBall(gameBall)
	}
	if uint32(hooks.loadTypeInd(target)) != gameBall &&
		hooks.loadClassLow(target)&flagCollidePlayerClass4EA400 == 0 {
		return
	}
	hooks.pickupGameBall(source, target, collision)
}
