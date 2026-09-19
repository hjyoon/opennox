package server

import "encoding/binary"

const (
	matchLimitQuestMask5096F0       = uint32(0x00001000)
	matchLimitTeamModeMask5096F0    = uint32(0x00000060)
	matchLimitHighScoreMask5096F0   = uint32(0x00000110)
	matchLimitLowScoreMask5096F0    = uint32(0x00000400)
	matchLimitSuddenDeathFlag5096F0 = uint32(0x04000000)
	matchLimitQuestActive5096F0     = uint32(1)
	matchLimitPositionOp5096F0      = byte(0x9a)
	matchLimitAudioID5096F0         = uint32(582)
	matchLimitAudioKind5096F0       = int32(2)
)

type matchLimitStateHooks5096F0[O, U, P comparable] struct {
	limitExpired        func() int32
	hasGameFlags        func(uint32) bool
	switchToNextMap     func()
	printAutoExit       func()
	firstPlayer         func() O
	nextPlayer          func(O) O
	loadUpdate          func(O) U
	loadQuestExit       func(U) O
	loadPlayer          func(U) P
	loadQuestState      func(P) uint32
	recordQuestProgress func(O)
	matchStateActive    func() int32
	resolveTeamMode     func() int32
	resolveHighScore    func() int32
	resolveLowScore     func() int32
	setGameFlags        func(uint32)
	loadPositionX       func(P) float32
	loadPositionY       func(P) float32
	loadPlayerIndex     func(P) uint8
	sendPosition        func(uint8, [5]byte)
	loadNetCode         func(O) uint32
	audioEvent          func(uint32, O, int32, uint32)
	clearTimer          func() int32
}

func matchLimitPositionPacket5096F0(x, y float32) [5]byte {
	var packet [5]byte
	packet[0] = matchLimitPositionOp5096F0
	binary.LittleEndian.PutUint16(packet[1:], uint16(x87TruncSignedQwordLow566DCC(float64(x))))
	binary.LittleEndian.PutUint16(packet[3:], uint16(x87TruncSignedQwordLow566DCC(float64(y))))
	return packet
}

// matchLimitState5096F0 preserves GAME.EXE 005096F0, including its live
// Player reloads for X, Y, and recipient index and the exact winner-mode test
// order. Pointer-bearing storage is abstracted so the server binding can use
// native layouts on every architecture.
func matchLimitState5096F0[O, U, P comparable](hooks matchLimitStateHooks5096F0[O, U, P]) int32 {
	result := hooks.limitExpired()
	if result == 0 {
		return result
	}

	var nilObject O
	var nilUpdate U
	if hooks.hasGameFlags(matchLimitQuestMask5096F0) {
		hooks.switchToNextMap()
		hooks.printAutoExit()
		for unit := hooks.firstPlayer(); unit != nilObject; unit = hooks.nextPlayer(unit) {
			update := hooks.loadUpdate(unit)
			if hooks.loadQuestExit(update) == nilObject {
				player := hooks.loadPlayer(update)
				if hooks.loadQuestState(player) == matchLimitQuestActive5096F0 {
					hooks.recordQuestProgress(unit)
				}
			}
		}
		return hooks.clearTimer()
	}

	if hooks.matchStateActive() == 0 {
		if hooks.hasGameFlags(matchLimitTeamModeMask5096F0) {
			_ = hooks.resolveTeamMode()
			return hooks.clearTimer()
		}
		if hooks.hasGameFlags(matchLimitHighScoreMask5096F0) {
			_ = hooks.resolveHighScore()
			return hooks.clearTimer()
		}
		if hooks.hasGameFlags(matchLimitLowScoreMask5096F0) {
			_ = hooks.resolveLowScore()
		}
		return hooks.clearTimer()
	}

	hooks.setGameFlags(matchLimitSuddenDeathFlag5096F0)
	for unit := hooks.firstPlayer(); unit != nilObject; unit = hooks.nextPlayer(unit) {
		update := hooks.loadUpdate(unit)
		if update != nilUpdate {
			player := hooks.loadPlayer(update)
			x := hooks.loadPositionX(player)
			player = hooks.loadPlayer(update)
			y := hooks.loadPositionY(player)
			player = hooks.loadPlayer(update)
			index := hooks.loadPlayerIndex(player)
			hooks.sendPosition(index, matchLimitPositionPacket5096F0(x, y))
		}
		hooks.audioEvent(
			matchLimitAudioID5096F0,
			unit,
			matchLimitAudioKind5096F0,
			hooks.loadNetCode(unit),
		)
	}
	return hooks.clearTimer()
}
