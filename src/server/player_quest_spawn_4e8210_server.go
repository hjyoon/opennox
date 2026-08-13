package server

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

// SoulGateCollideData is the four-byte collide record registered by
// SoulGateCollide. GAME.EXE 004E8210 compares LastUsedFrame as uint32.
type SoulGateCollideData struct {
	LastUsedFrame uint32
}

func (obj *Object) collideDataSoulGate() *SoulGateCollideData {
	return (*SoulGateCollideData)(obj.CollideData)
}

func playerUpdateDataRaw4E8210(obj *Object) *PlayerUpdateData {
	return (*PlayerUpdateData)(obj.UpdateData)
}

type playerQuestSpawnNativeDeps4E8210 struct {
	firstUnit          func() *Object
	nextUnit           func(*Object) *Object
	randomReachablePos func(float32, types.Pointf) types.Pointf
}

func playerQuestSpawnNative4E8210(
	joining *Object,
	deps playerQuestSpawnNativeDeps4E8210,
) (types.Pointf, bool) {
	return playerQuestSpawn4E8210(joining, playerQuestSpawnHooks4E8210[*Object, *Object, *SoulGateCollideData, types.Pointf]{
		firstUnit: deps.firstUnit,
		nextUnit:  deps.nextUnit,
		loadSoulGate: func(unit *Object) *Object {
			return playerUpdateDataRaw4E8210(unit).SoulGate
		},
		loadCollideData: func(gate *Object) *SoulGateCollideData {
			return gate.collideDataSoulGate()
		},
		loadLastUsedFrame: func(data *SoulGateCollideData) uint32 {
			return data.LastUsedFrame
		},
		storeSoulGate: func(unit, gate *Object) {
			playerUpdateDataRaw4E8210(unit).SoulGate = gate
		},
		loadSoulGatePos: func(gate *Object) types.Pointf {
			return gate.PosVec
		},
		randomReachablePos: deps.randomReachablePos,
	})
}

func (s *Server) Sub4E8210(joining *Object) (types.Pointf, bool) {
	return playerQuestSpawnNative4E8210(joining, playerQuestSpawnNativeDeps4E8210{
		firstUnit:          s.Players.FirstUnit,
		nextUnit:           s.Players.NextUnit,
		randomReachablePos: s.RandomReachablePointAround,
	})
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(SoulGateCollideData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(SoulGateCollideData{}.LastUsedFrame)]
)
