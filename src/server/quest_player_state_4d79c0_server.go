package server

import (
	"encoding/binary"

	"github.com/opennox/opennox/v1/common/ntype"
)

const questPlayerStateSlots4D79C0 = 32

// QuestPlayerStateRuntime4D79C0 supplies the timestamp storage and legacy
// integration calls that are still owned by the Quest runtime. Object and
// update-data pointers remain native-width throughout the restored routines.
type QuestPlayerStateRuntime4D79C0 struct {
	LoadTimestamp  func(index int) uint32
	StoreTimestamp func(index int, value uint32)
	Notify         func(recipient int, unit *Object) int
	Reset          func(unit *Object)
}

func clearQuestPlayerStateSlot4D79C0(update *PlayerUpdateData, index int) {
	update.RespawnMarkers[index] = 0
	update.QuestPlayerState[index] = 0
	update.QuestPlayerFlagsA[index] = 0
	update.QuestPlayerFlagsB[index] = 0
}

// QuestPlayerStateRemove4D79C0 preserves GAME.EXE 004D79C0. The source
// update-data pointer is cached before notification and reset callbacks, while
// its Player and PlayerInd are deliberately reloaded afterwards.
func (s *Server) QuestPlayerStateRemove4D79C0(unit *Object, runtime QuestPlayerStateRuntime4D79C0) int32 {
	update := (*PlayerUpdateData)(unit.UpdateData)
	runtime.Notify(255, unit)
	runtime.Reset(unit)
	index := int(update.Player.PlayerInd)

	for current := s.Players.FirstUnit(); current != nil; current = s.questNextPlayerUnit4DA7F0(current) {
		clearQuestPlayerStateSlot4D79C0((*PlayerUpdateData)(current.UpdateData), index)
	}
	return 0
}

// QuestPlayerStateMark4D7A60 preserves GAME.EXE 004D7A60.
func (s *Server) QuestPlayerStateMark4D7A60(index int, runtime QuestPlayerStateRuntime4D79C0) int32 {
	runtime.StoreTimestamp(index, s.Frame())
	return int32(index)
}

// QuestPlayerStateExpire4D7A80 preserves GAME.EXE 004D7A80. Active Quest
// players clear their disconnect timestamp immediately; inactive slots expire
// after a strict 30-second unsigned frame interval.
func (s *Server) QuestPlayerStateExpire4D7A80(runtime QuestPlayerStateRuntime4D79C0) int32 {
	threshold := uint32(30) * s.TickRate()
	for index := 0; index < questPlayerStateSlots4D79C0; index++ {
		player := s.Players.ByIndRaw(ntype.PlayerInd(index))
		if player != nil && player.Active != 0 && player.PlayerUnit != nil && player.Field4792 == 1 {
			runtime.StoreTimestamp(index, 0)
			continue
		}

		timestamp := runtime.LoadTimestamp(index)
		if timestamp == 0 || s.Frame()-timestamp <= threshold {
			continue
		}
		for current := s.Players.FirstUnit(); current != nil; current = s.questNextPlayerUnit4DA7F0(current) {
			clearQuestPlayerStateSlot4D79C0((*PlayerUpdateData)(current.UpdateData), index)
		}
		runtime.StoreTimestamp(index, 0)
	}
	return questPlayerStateSlots4D79C0
}

// QuestPlayerStateClear4D7B40 preserves GAME.EXE 004D7B40.
func (s *Server) QuestPlayerStateClear4D7B40(runtime QuestPlayerStateRuntime4D79C0) int32 {
	for index := 0; index < questPlayerStateSlots4D79C0; index++ {
		runtime.StoreTimestamp(index, 0)
	}
	return 0
}

// QuestNotifyPlayer4D9D20 preserves the reliable four-byte Quest player
// notification emitted by GAME.EXE 004D9D20.
func (s *Server) QuestNotifyPlayer4D9D20(recipient int, unit *Object) int {
	var packet [4]byte
	packet[0] = 0xf0
	packet[1] = 0x01
	binary.LittleEndian.PutUint16(packet[2:], uint16(unit.NetCode))
	return s.NetSendPacketXxx1(recipient, packet[:], nil, 1)
}
