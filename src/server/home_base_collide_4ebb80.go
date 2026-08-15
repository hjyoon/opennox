package server

const (
	homeBaseGameBallName4EBB80      = "GameBall"
	homeBaseGameBallStartName4EBB80 = "GameBallStart"
	homeBaseRandomDebugPath4EBB80   = `C:\NoxPost\src\Server\Object\collide\objcoll.c`
	homeBaseRandomDebugLine4EBB80   = int32(0x0f57)
	homeBaseScoreDelta4EBB80        = int32(1)
	homeBaseScoreAudio4EBB80        = uint32(929)
	homeBaseScorePointFX4EBB80      = uint32(154)
	homeBaseRespawnPointFX4EBB80    = uint32(129)
)

type homeBaseCollideHooks4EBB80[O, D, T, U, P comparable] struct {
	lookupType        func(string) uint32
	loadTypeIndex     func(O) uint16
	loadUpdate        func(O) D
	loadCarrier       func(D) O
	hasTeam           func(O) bool
	loadTeamID        func(O) uint8
	teamByID          func(uint8) T
	changeScore       func(O, int32)
	reportLesson      func(O)
	loadTeamLessons   func(T) int32
	changeTeamLessons func(T, int32)
	observerMode      func() uint32
	loadPlayerUpdate  func(O) U
	loadPlayer        func(U) P
	observerUpdate    func(P, P)
	audio             func(uint32, O)
	pointFX           func(uint32, O) uint32
	firstObject       func() O
	nextObject        func(O) O
	randomInt         func(int32, int32, string, int32) int32
	clearOwner        func(O)
	moveToMarker      func(O, O)
	storeVelocityX    func(O, float32)
	storeVelocityY    func(O, float32)
	storeForceX       func(O, float32)
	storePos24Y       func(O, float32)
}

// homeBaseCollide4EBB80 preserves GAME.EXE 004EBB80. Both type lookups run
// before the target guard. A matching GameBall caches its update record, uses
// live carrier loads at every original load site, then performs two separate
// live object-list traversals to choose a GameBallStart marker. The collision
// point is registered but never read.
func homeBaseCollide4EBB80[O, D, T, U, P comparable, C any](
	homeBase, other O,
	_ C,
	hooks homeBaseCollideHooks4EBB80[O, D, T, U, P],
) uint32 {
	gameBallType := hooks.lookupType(homeBaseGameBallName4EBB80)
	startType := hooks.lookupType(homeBaseGameBallStartName4EBB80)

	var zeroObject O
	if other == zeroObject {
		return startType
	}
	otherType := uint32(hooks.loadTypeIndex(other))
	if otherType != gameBallType {
		return otherType
	}

	update := hooks.loadUpdate(other)
	var zeroTeam T
	homeTeam := zeroTeam
	if hooks.hasTeam(homeBase) {
		homeTeam = hooks.teamByID(hooks.loadTeamID(homeBase))
	}

	carrierTeam := zeroTeam
	carrier := hooks.loadCarrier(update)
	if carrier != zeroObject && hooks.hasTeam(carrier) {
		carrier = hooks.loadCarrier(update)
		carrierTeam = hooks.teamByID(hooks.loadTeamID(carrier))
	}
	if homeTeam == carrierTeam {
		carrier = hooks.loadCarrier(update)
		hooks.changeScore(carrier, homeBaseScoreDelta4EBB80)
		carrier = hooks.loadCarrier(update)
		hooks.reportLesson(carrier)
	}

	if homeTeam != zeroTeam {
		lessons := hooks.loadTeamLessons(homeTeam)
		hooks.changeTeamLessons(homeTeam, lessons+homeBaseScoreDelta4EBB80)
		if hooks.observerMode() != 0 {
			carrier = hooks.loadCarrier(update)
			if carrier != zeroObject {
				playerUpdate := hooks.loadPlayerUpdate(carrier)
				var zeroUpdate U
				if playerUpdate != zeroUpdate {
					player := hooks.loadPlayer(playerUpdate)
					var zeroPlayer P
					hooks.observerUpdate(player, zeroPlayer)
				}
			}
		}
		hooks.audio(homeBaseScoreAudio4EBB80, homeBase)
		hooks.pointFX(homeBaseScorePointFX4EBB80, other)
	}

	var count int32
	for obj := hooks.firstObject(); obj != zeroObject; obj = hooks.nextObject(obj) {
		if uint32(hooks.loadTypeIndex(obj)) == startType {
			count++
		}
	}

	selection := hooks.randomInt(
		0,
		count-1,
		homeBaseRandomDebugPath4EBB80,
		homeBaseRandomDebugLine4EBB80,
	)
	for marker := hooks.firstObject(); marker != zeroObject; marker = hooks.nextObject(marker) {
		if uint32(hooks.loadTypeIndex(marker)) != startType {
			continue
		}
		if selection == 0 {
			hooks.clearOwner(other)
			hooks.moveToMarker(other, marker)
			result := hooks.pointFX(homeBaseRespawnPointFX4EBB80, other)
			hooks.storeVelocityX(other, 0)
			hooks.storeVelocityY(other, 0)
			hooks.storeForceX(other, 0)
			hooks.storePos24Y(other, 0)
			return result
		}
		selection--
	}
	return 0
}
