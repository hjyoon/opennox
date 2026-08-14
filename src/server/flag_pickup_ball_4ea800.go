package server

const (
	flagPickupBallGameBallName4EA800    = "GameBall"
	flagPickupBallStartName4EA800       = "GameBallStart"
	flagPickupBallPlayerClass4EA800     = uint8(0x04)
	flagPickupBallInactiveCarrier4EA800 = uint8(0x20)
	flagPickupBallScoreMode4EA800       = uint32(64)
	flagPickupBallScoreDelta4EA800      = int32(1)
	flagPickupBallScoreAudio4EA800      = uint32(929)
	flagPickupBallScoreMessage4EA800    = uint32(9)
	flagPickupBallScorePointFX4EA800    = uint32(154)
	flagPickupBallWinnerGameFlag4EA800  = uint32(8)
	flagPickupBallRespawnPointFX4EA800  = uint32(129)
	flagPickupBallStatusFirstArg4EA800  = uint32(0)
	flagPickupBallStatusSecondArg4EA800 = uint32(0)
)

type flagPickupBallHooks4EA800[O, U, T, P comparable] struct {
	loadBallCache    func() uint32
	lookupType       func(string) uint32
	storeBallCache   func(uint32)
	loadClassLow     func(O) uint8
	unitIsGameBall   func(O) int32
	firstOwned       func(O) O
	nextOwned        func(O) O
	loadTypeInd      func(O) uint16
	loadUpdate       func(O) U
	loadCarrier      func(U) O
	loadFlagsLow     func(O) uint8
	storeCarrier     func(U, O)
	loadTeamID       func(O) uint8
	teamByID         func(uint8) T
	nextTeam         func(T) T
	firstTeam        func() T
	loadTeamIDValue  func(T) uint8
	gameData         func(uint32) uint16
	changeScore      func(O, int32)
	reportLesson     func(O)
	loadTeamScore    func(T) int32
	changeTeamScore  func(T, int32)
	observerMode     func() uint32
	playerFromUpdate func(U) P
	observerUpdate   func(P, P)
	audio            func(uint32, O)
	loadNetCode      func(O) uint32
	informScore      func(uint32, uint32, uint32)
	pointFX          func(uint32, O)
	setGameFlags     func(uint32)
	flagBallWinner   func(T)
	loadStartCache   func() uint32
	storeStartCache  func(uint32)
	firstObject      func() O
	nextObject       func(O) O
	randomInt        func(int32, int32) int32
	clearOwner       func(O)
	dropBall         func(O, O)
	changeObjectTeam func(O, uint32)
	setHPMax         func(O)
	ticks            func() uint64
	storeTicks       func(U, uint64)
	moveToMarker     func(O, O)
	ballStatus       func(uint32, uint32)
	clearMotion      func(O)
}

// flagPickupBall4EA800 preserves GAME.EXE 004EA800. The collision argument is
// accepted for the shared three-argument collision callback ABI but is never
// inspected. The original short return value is only a collection of incidental
// EAX pointer/callback artifacts and its sole caller discards it, so it is not
// exposed as a cross-architecture semantic result.
func flagPickupBall4EA800[O, U, T, P comparable, C any](
	source, target O,
	_ C,
	hooks flagPickupBallHooks4EA800[O, U, T, P],
) {
	ball := flagPickupBallResolve4EA800(target, hooks)
	var zeroObject O
	if ball == zeroObject {
		return
	}
	if !flagPickupBallScore4EA800(source, ball, hooks) {
		return
	}
	flagPickupBallRespawn4EA800(ball, hooks)
}

func flagPickupBallResolve4EA800[O, U, T, P comparable](
	target O,
	hooks flagPickupBallHooks4EA800[O, U, T, P],
) O {
	if hooks.loadBallCache() == 0 {
		ind := hooks.lookupType(flagPickupBallGameBallName4EA800)
		hooks.storeBallCache(ind)
	}

	ball := target
	if hooks.loadClassLow(target)&flagPickupBallPlayerClass4EA800 != 0 {
		var zeroObject O
		if hooks.unitIsGameBall(target) == 0 {
			return zeroObject
		}
		ball = hooks.firstOwned(target)
		if ball == zeroObject {
			return zeroObject
		}
		for {
			ballType := hooks.loadBallCache()
			if uint32(hooks.loadTypeInd(ball)) == ballType {
				break
			}
			ball = hooks.nextOwned(ball)
			if ball == zeroObject {
				return zeroObject
			}
		}
	}
	return ball
}

// flagPickupBallScore4EA800 returns true only after the scoring branch reaches
// the unconditional respawn tail at 004EA994.
func flagPickupBallScore4EA800[O, U, T, P comparable](
	source, ball O,
	hooks flagPickupBallHooks4EA800[O, U, T, P],
) bool {
	var zeroObject O
	update := hooks.loadUpdate(ball)
	carrier := hooks.loadCarrier(update)
	if carrier != zeroObject && hooks.loadFlagsLow(carrier)&flagPickupBallInactiveCarrier4EA800 != 0 {
		hooks.storeCarrier(update, zeroObject)
	}

	team := hooks.teamByID(hooks.loadTeamID(source))
	team = hooks.nextTeam(team)
	var zeroTeam T
	if team == zeroTeam {
		team = hooks.firstTeam()
	}

	carrier = hooks.loadCarrier(update)
	if carrier == zeroObject {
		return false
	}
	if hooks.loadTeamID(carrier) != hooks.loadTeamIDValue(team) {
		return false
	}

	threshold := int32(hooks.gameData(flagPickupBallScoreMode4EA800))
	hooks.changeScore(carrier, flagPickupBallScoreDelta4EA800)
	hooks.reportLesson(carrier)
	teamScore := hooks.loadTeamScore(team)
	hooks.changeTeamScore(team, teamScore+flagPickupBallScoreDelta4EA800)
	if hooks.observerMode() != 0 {
		carrierUpdate := hooks.loadUpdate(carrier)
		var zeroUpdate U
		if carrierUpdate != zeroUpdate {
			player := hooks.playerFromUpdate(carrierUpdate)
			var zeroPlayer P
			hooks.observerUpdate(player, zeroPlayer)
		}
	}

	hooks.audio(flagPickupBallScoreAudio4EA800, source)
	netCode := hooks.loadNetCode(carrier)
	teamID := uint32(hooks.loadTeamIDValue(team))
	hooks.informScore(flagPickupBallScoreMessage4EA800, netCode, teamID)
	hooks.pointFX(flagPickupBallScorePointFX4EA800, ball)

	if threshold > 0 {
		for candidate := hooks.firstTeam(); candidate != zeroTeam; candidate = hooks.nextTeam(candidate) {
			if hooks.loadTeamScore(candidate) >= threshold {
				hooks.setGameFlags(flagPickupBallWinnerGameFlag4EA800)
				hooks.flagBallWinner(candidate)
				break
			}
		}
	}
	return true
}

func flagPickupBallRespawn4EA800[O, U, T, P comparable](
	ball O,
	hooks flagPickupBallHooks4EA800[O, U, T, P],
) {
	if hooks.loadStartCache() == 0 {
		ind := hooks.lookupType(flagPickupBallStartName4EA800)
		hooks.storeStartCache(ind)
	}

	var zeroObject O
	var count int32
	for obj := hooks.firstObject(); obj != zeroObject; obj = hooks.nextObject(obj) {
		startType := hooks.loadStartCache()
		if uint32(hooks.loadTypeInd(obj)) == startType {
			count++
		}
	}

	selection := hooks.randomInt(0, count-1)
	for marker := hooks.firstObject(); marker != zeroObject; marker = hooks.nextObject(marker) {
		startType := hooks.loadStartCache()
		if uint32(hooks.loadTypeInd(marker)) != startType {
			continue
		}
		if selection == 0 {
			flagPickupBallPlaceAtMarker4EA800(ball, marker, hooks)
			return
		}
		selection--
	}
}

func flagPickupBallPlaceAtMarker4EA800[O, U, T, P comparable](
	ball, marker O,
	hooks flagPickupBallHooks4EA800[O, U, T, P],
) {
	var zeroObject O
	update := hooks.loadUpdate(ball)
	hooks.clearOwner(ball)
	hooks.dropBall(ball, zeroObject)
	netCode := hooks.loadNetCode(ball)
	hooks.changeObjectTeam(ball, netCode)
	hooks.setHPMax(ball)
	ticks := hooks.ticks()
	hooks.storeTicks(update, ticks)
	hooks.moveToMarker(ball, marker)
	hooks.ballStatus(flagPickupBallStatusFirstArg4EA800, flagPickupBallStatusSecondArg4EA800)
	hooks.pointFX(flagPickupBallRespawnPointFX4EA800, ball)
	hooks.clearMotion(ball)
}
