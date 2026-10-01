package server

import "unsafe"

// PlayerDieQuestRuntime54D2B0 contains the existing native services reached
// by the Quest branch of PlayerDie, not replacement loss algorithms.
type PlayerDieQuestRuntime54D2B0 struct {
	SendStats    func(uint8, [14]byte)
	RecordDeath  func(*Object) unsafe.Pointer
	ResetPlayer  func(*Object)
	Penalty      func(*Object)
	BalanceFloat func(string) float32
}

func playerDieQuestRuntimeReady54D2B0(runtime PlayerDieRuntime54D2B0) bool {
	q := runtime.Quest
	return q != nil && runtime.Frame != nil && q.SendStats != nil &&
		q.RecordDeath != nil && q.ResetPlayer != nil && q.Penalty != nil &&
		q.BalanceFloat != nil
}

func playerDieQuestNative54D2B0(unit *Object, update *PlayerUpdateData, runtime PlayerDieRuntime54D2B0) playerDieQuestResult54D2B0[*Player, unsafe.Pointer] {
	// Service closures are deliberately invoked only on their original branch:
	// an ordinary death must not dereference an absent Quest runtime.
	return playerDieQuest54D2B0(unit, update, playerDieQuestHooks54D2B0[*Object, *PlayerUpdateData, *Player, unsafe.Pointer]{
		gameFlag: func(flag uint32) int32 {
			if runtime.GameFlag(flag) {
				return 1
			}
			return 0
		},
		loadExtraLives:     func(update *PlayerUpdateData) uint32 { return update.ExtraLives },
		storeExtraLives:    func(update *PlayerUpdateData, value uint32) { update.ExtraLives = value },
		recordDeath:        func(unit *Object) unsafe.Pointer { return runtime.Quest.RecordDeath(unit) },
		frame:              runtime.Frame,
		loadPlayer:         func(update *PlayerUpdateData) *Player { return update.Player },
		storeFrame:         func(update *PlayerUpdateData, value uint32) { update.Field137 = value },
		loadStage:          func(player *Player) uint16 { return uint16(player.field4688) },
		loadGenerators:     func(player *Player) uint16 { return uint16(player.field4668) },
		loadMonsters:       func(player *Player) uint16 { return uint16(player.field4664) },
		loadSecrets:        func(player *Player) uint16 { return uint16(player.field4672) },
		loadPlayerIndex:    func(player *Player) uint8 { return uint8(player.PlayerInd) },
		sendStats:          func(index uint8, packet [14]byte) { runtime.Quest.SendStats(index, packet) },
		resetPlayer:        func(unit *Object) { runtime.Quest.ResetPlayer(unit) },
		penalty:            func(unit *Object) { runtime.Quest.Penalty(unit) },
		balanceFloat:       func(key string) float32 { return runtime.Quest.BalanceFloat(key) },
		floatToInt:         playerUnitInitFloatToInt4EFE80,
		loadExtraLivesByte: func(update *PlayerUpdateData) uint8 { return uint8(update.ExtraLives) },
		storeRespawnMarker: func(update *PlayerUpdateData, index, value uint8) { update.RespawnMarkers[index] = value },
	})
}
