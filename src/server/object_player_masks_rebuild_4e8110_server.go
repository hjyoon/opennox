package server

import "github.com/opennox/opennox/v1/common/ntype"

type objectPlayerMasksRebuildNativeDeps4E8110 struct {
	playerByInd func(int32) *Player
	firstObject func() *Object
	nextObject  func(*Object) *Object
	isHostile   func(*Object, *Object) int32
}

func objectPlayerMasksRebuildNative4E8110(
	playerInd int32,
	deps objectPlayerMasksRebuildNativeDeps4E8110,
) *Object {
	return objectPlayerMasksRebuild4E8110(playerInd, objectPlayerMasksRebuildHooks4E8110[*Object, *Player]{
		playerByInd: deps.playerByInd,
		firstObject: deps.firstObject,
		loadField36: func(obj *Object) uint32 {
			return obj.Field36
		},
		loadField35: func(obj *Object) uint32 {
			return obj.Field35
		},
		loadClassLow: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		storeField36: func(obj *Object, value uint32) {
			obj.Field36 = value
		},
		storeField35: func(obj *Object, value uint32) {
			obj.Field35 = value
		},
		loadPlayerUnit: func(player *Player) *Object {
			return player.PlayerUnit
		},
		isHostile:  deps.isHostile,
		nextObject: deps.nextObject,
	})
}

// RebuildObjectPlayerMasks4E8110 clears and recomputes one player's relation
// bit across the active object list, retaining the terminal null EAX artifact.
func (s *Server) RebuildObjectPlayerMasks4E8110(playerInd int32) *Object {
	return objectPlayerMasksRebuildNative4E8110(playerInd, objectPlayerMasksRebuildNativeDeps4E8110{
		playerByInd: func(ind int32) *Player {
			return s.Players.ByInd(ntype.PlayerInd(ind))
		},
		firstObject: s.Objs.First,
		nextObject: func(obj *Object) *Object {
			return obj.Next()
		},
		isHostile: s.isHostileMimicResult4E7F90,
	})
}
