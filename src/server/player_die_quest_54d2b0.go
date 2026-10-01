package server

import "encoding/binary"

const playerDieQuestLivesKey54D2B0 = "QuestGameStartingExtraLives"

type playerDieQuestHooks54D2B0[O, U, P, R any] struct {
	gameFlag           func(uint32) int32
	loadExtraLives     func(U) uint32
	storeExtraLives    func(U, uint32)
	recordDeath        func(O) R
	frame              func() uint32
	loadPlayer         func(U) P
	storeFrame         func(U, uint32)
	loadStage          func(P) uint16
	loadMonsters       func(P) uint16
	loadSecrets        func(P) uint16
	loadGenerators     func(P) uint16
	loadPlayerIndex    func(P) uint8
	sendStats          func(uint8, [14]byte)
	resetPlayer        func(O)
	penalty            func(O)
	balanceFloat       func(string) float32
	floatToInt         func(float32) int32
	loadExtraLivesByte func(U) uint8
	storeRespawnMarker func(U, uint8, uint8)
}

type playerDieQuestReturnKind54D2B0 uint8

const (
	playerDieQuestFlagReturn54D2B0 playerDieQuestReturnKind54D2B0 = iota
	playerDieQuestRecordReturn54D2B0
	playerDieQuestPlayerReturn54D2B0
)

// EAX is a game-flag scalar, the record-death helper's pointer, or the final
// live Player pointer. Keep these domains separate; the active death callback
// has a void ABI and does not consume the original mixed result.
type playerDieQuestResult54D2B0[P, R any] struct {
	kind     playerDieQuestReturnKind54D2B0
	flag     int32
	recorded R
	player   P
}

// playerDieQuest54D2B0 restores the tail at GAME.EXE 0054D6A6..0054D791.
// The caller supplies the entry-cached update. Frame precedes the first Player
// load, packet statistics use that cached Player, and the final marker uses a
// fresh Player followed by a fresh low-byte life read. No new guards or clamps
// are introduced into this original branch.
func playerDieQuest54D2B0[O, U, P, R any](unit O, update U, h playerDieQuestHooks54D2B0[O, U, P, R]) playerDieQuestResult54D2B0[P, R] {
	flag := h.gameFlag(playerDieQuestMode54D2B0)
	if flag == 0 {
		return playerDieQuestResult54D2B0[P, R]{kind: playerDieQuestFlagReturn54D2B0, flag: flag}
	}
	lives := h.loadExtraLives(update)
	if lives != 0 {
		h.storeExtraLives(update, lives-1)
		recorded := h.recordDeath(unit)
		return playerDieQuestResult54D2B0[P, R]{kind: playerDieQuestRecordReturn54D2B0, recorded: recorded}
	}

	frame := h.frame()
	player := h.loadPlayer(update)
	h.storeFrame(update, frame)
	var packet [14]byte
	packet[0], packet[1] = 0xf0, 2
	binary.LittleEndian.PutUint16(packet[8:], h.loadStage(player))
	binary.LittleEndian.PutUint16(packet[2:], h.loadGenerators(player))
	binary.LittleEndian.PutUint16(packet[6:], h.loadMonsters(player))
	binary.LittleEndian.PutUint16(packet[4:], h.loadSecrets(player))
	index := h.loadPlayerIndex(player)
	h.sendStats(index, packet)
	h.resetPlayer(unit)
	h.penalty(unit)
	value := h.balanceFloat(playerDieQuestLivesKey54D2B0)
	h.storeExtraLives(update, uint32(h.floatToInt(value)))
	player = h.loadPlayer(update)
	lifeByte := h.loadExtraLivesByte(update)
	index = h.loadPlayerIndex(player)
	h.storeRespawnMarker(update, index, lifeByte)
	return playerDieQuestResult54D2B0[P, R]{kind: playerDieQuestPlayerReturn54D2B0, player: player}
}
