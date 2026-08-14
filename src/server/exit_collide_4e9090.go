package server

const (
	exitCollidePlayerClassByte4E9090 = uint8(0x04)
	exitCollideWarriorClass4E9090    = uint8(1)

	exitCollideSubtypeExit4E9090 = uint8(0x01)
	exitCollideSubtypeWarp4E9090 = uint8(0x02)

	exitCollideGameCoop4E9090  = uint32(0x0800)
	exitCollideGameQuest4E9090 = uint32(0x1000)

	exitCollidePlayerState4E9090 = int32(13)
	exitCollideSound4E9090       = uint32(1003)
	exitCollideMessageExit4E9090 = uint8(18)
	exitCollideMessageWarp4E9090 = uint8(19)

	exitCollideWorkingSave4E9090 = "WORKING"
	exitCollideWarpMessage4E9090 = "objcoll.c:PlayerEntersWarp"
)

type exitCollideHooks4E9090[O comparable, U, P any, M comparable, C any] struct {
	glyphType            func() uint32
	loadClassByte        func(O) uint8
	defaultCollide       func(O, O, C)
	loadUpdateData       func(O) U
	loadMap              func(O) M
	gameFlag             func(uint32) int32
	loadQuestExit        func(U) O
	loadQuestWarpGate    func(U) O
	loadSubclassByte     func(O) uint8
	warpEnabled          func() int32
	saveBusy             func() int32
	exitAllowed          func(O) int32
	paused               func() int32
	mapFirstByte         func(M) byte
	loadPlayer           func(U) P
	loadPlayerClass      func(P) uint8
	firstOwned           func(O) O
	nextOwned            func(O) O
	loadTypeIndex        func(O) uint16
	loadInventoryHolder  func(O) O
	delayedDelete        func(O)
	loadCurTrapsByte     func(U) uint8
	storeCurTrapsByte    func(U, uint8)
	setMapLoadRequired   func(int32)
	setSaveFileName      func(string)
	saveCoop             func(int32, O, int32)
	questMapFile         func() M
	abilityActive        func(O, uint32) int32
	disableAbility       func(O, uint32)
	currentQuestStage    func() uint32
	recordExitProgress   func(O)
	loadQuestStage       func(P) uint32
	storeQuestStage      func(P, uint32)
	loadPlayerIndex      func(P) uint8
	sendQuestStage       func(uint8, uint32)
	storeQuestExit       func(U, O)
	storeQuestWarpGate   func(U, O)
	setPlayerState       func(O, int32)
	goObserver           func(P, int32, int32)
	broadcastUnitMessage func(uint8, O)
	allPlayersExited     func() int32
	frame                func() uint32
	storeWarpFrame       func(uint32)
	firstPlayerUnit      func() O
	nextPlayerUnit       func(O) O
	sendUnitMessage      func(uint8, O, O)
	priorityMessage      func(O, string, uint8)
	maybeWarp            func() int32
	audio                func(uint32, O, int32, int32)
	exitCountdown        func() int32
	copyNextMap          func(M)
	nextStageThreshold   func(uint32) uint32
	setCurrentQuestStage func(uint32)
	setQuestWarping      func(int32)
	resetQuestPlayers    func()
	mapLoad              func(M)
}

// exitCollide4E9090 preserves GAME.EXE 004E9090. The Glyph type cache is
// populated before any argument check. Player collisions then retain the
// original gate order, callback-visible reloads, Quest exit/warp transitions,
// live owned/player traversal, and next-map copy/load ordering. The third
// collision argument is observed only by the non-Player default path.
func exitCollide4E9090[O comparable, U, P any, M comparable, C any](
	exit, unit O,
	collision C,
	hooks exitCollideHooks4E9090[O, U, P, M, C],
) {
	glyphType := hooks.glyphType()

	var zeroObject O
	if unit == zeroObject {
		return
	}
	if hooks.loadClassByte(unit)&exitCollidePlayerClassByte4E9090 == 0 {
		hooks.defaultCollide(exit, unit, collision)
		return
	}

	update := hooks.loadUpdateData(unit)
	mapName := hooks.loadMap(exit)
	if hooks.gameFlag(exitCollideGameQuest4E9090) != 0 {
		if hooks.loadQuestExit(update) != zeroObject {
			return
		}
		if hooks.loadQuestWarpGate(update) != zeroObject {
			return
		}
	}
	if hooks.loadSubclassByte(exit)&exitCollideSubtypeWarp4E9090 != 0 && hooks.warpEnabled() == 0 {
		return
	}
	if hooks.saveBusy() == 1 {
		return
	}
	if hooks.exitAllowed(unit) == 0 {
		return
	}
	if hooks.paused() == 1 {
		return
	}
	if hooks.mapFirstByte(mapName) == 0 && hooks.gameFlag(exitCollideGameQuest4E9090) == 0 {
		return
	}

	player := hooks.loadPlayer(update)
	if hooks.loadPlayerClass(player) == exitCollideWarriorClass4E9090 {
		for owned := hooks.firstOwned(unit); owned != zeroObject; owned = hooks.nextOwned(owned) {
			if uint32(hooks.loadTypeIndex(owned)) == glyphType && hooks.loadInventoryHolder(owned) == zeroObject {
				hooks.delayedDelete(owned)
				if count := hooks.loadCurTrapsByte(update); count != 0 {
					hooks.storeCurTrapsByte(update, count-1)
				}
			}
		}
	}

	hooks.setMapLoadRequired(1)
	if hooks.gameFlag(exitCollideGameCoop4E9090) != 0 {
		hooks.setSaveFileName(exitCollideWorkingSave4E9090)
		hooks.saveCoop(1, exit, 0)
		return
	}

	result := int32(1)
	if hooks.gameFlag(exitCollideGameQuest4E9090) != 0 {
		mapName = hooks.questMapFile()
	}
	if hooks.gameFlag(exitCollideGameQuest4E9090) != 0 {
		for ability := uint32(1); ability < 6; ability++ {
			if hooks.abilityActive(unit, ability) != 0 {
				hooks.disableAbility(unit, ability)
			}
		}

		subclass := hooks.loadSubclassByte(exit)
		switch {
		case subclass&exitCollideSubtypeExit4E9090 != 0:
			stage := hooks.currentQuestStage() + 1
			hooks.recordExitProgress(unit)
			player = hooks.loadPlayer(update)
			if hooks.loadQuestStage(player) < stage {
				hooks.storeQuestStage(player, stage)
				player = hooks.loadPlayer(update)
				reportedStage := hooks.loadQuestStage(player)
				playerIndex := hooks.loadPlayerIndex(player)
				hooks.sendQuestStage(playerIndex, reportedStage)
			}
			hooks.storeQuestExit(update, exit)
			hooks.storeQuestWarpGate(update, zeroObject)
			hooks.setPlayerState(unit, exitCollidePlayerState4E9090)
			hooks.goObserver(hooks.loadPlayer(update), 0, 0)
			hooks.broadcastUnitMessage(exitCollideMessageExit4E9090, unit)
			result = hooks.allPlayersExited()

		case subclass&exitCollideSubtypeWarp4E9090 != 0:
			hooks.storeQuestExit(update, zeroObject)
			hooks.storeQuestWarpGate(update, exit)
			hooks.storeWarpFrame(hooks.frame())
			hooks.setPlayerState(unit, exitCollidePlayerState4E9090)
			hooks.goObserver(hooks.loadPlayer(update), 0, 0)
			for other := hooks.firstPlayerUnit(); other != zeroObject; other = hooks.nextPlayerUnit(other) {
				if other != unit {
					hooks.sendUnitMessage(exitCollideMessageWarp4E9090, other, unit)
				}
			}
			hooks.priorityMessage(unit, exitCollideWarpMessage4E9090, 0)
			result = hooks.maybeWarp()
		}

		hooks.audio(exitCollideSound4E9090, exit, 0, 0)
		if hooks.loadSubclassByte(exit)&exitCollideSubtypeWarp4E9090 == 0 && result == 0 {
			_ = hooks.exitCountdown()
		}
		hooks.copyNextMap(mapName)
		if result != 1 {
			return
		}
	}

	var zeroMap M
	if mapName == zeroMap || hooks.mapFirstByte(mapName) == 0 {
		return
	}
	if hooks.loadSubclassByte(exit)&exitCollideSubtypeWarp4E9090 != 0 {
		stage := hooks.currentQuestStage()
		hooks.setCurrentQuestStage(hooks.nextStageThreshold(stage) - 1)
		hooks.setQuestWarping(1)
		hooks.resetQuestPlayers()
	}
	hooks.mapLoad(mapName)
}
