package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/types"
)

// AudioEventZoneRuntime501C00 supplies polygon services that remain owned by
// package legacy. Polygon handles stay native-width across every callback.
type AudioEventZoneRuntime501C00 struct {
	PolygonByID    func(uint32) unsafe.Pointer
	PolygonAtPoint func([2]int32, uint32) unsafe.Pointer
	PolygonZone    func(unsafe.Pointer) uint8
}

func audioEventZoneFloatToInt501C00(value float32) int32 {
	if math.IsNaN(float64(value)) || value >= 2147483648 || value < -2147483648 {
		return math.MinInt32
	}
	return int32(math.RoundToEven(float64(value)))
}

func audioEventZoneNative501C00(
	position *types.Pointf,
	object *Object,
	runtime AudioEventZoneRuntime501C00,
) uint8 {
	return audioEventZone501C00(position, object, audioEventZoneHooks501C00[
		*types.Pointf,
		*Object,
		unsafe.Pointer,
		*Player,
		unsafe.Pointer,
	]{
		loadClassLow: func(object *Object) uint8 {
			return uint8(object.ObjClass)
		},
		loadUpdate: func(object *Object) unsafe.Pointer {
			return object.UpdateData
		},
		loadPlayer: func(update unsafe.Pointer) *Player {
			return (*PlayerUpdateData)(update).Player
		},
		loadPlayerZone: func(player *Player) uint8 {
			return uint8(player.field3668)
		},
		loadMonsterPolygonID: func(update unsafe.Pointer) uint32 {
			return (*MonsterUpdateData)(update).Field0
		},
		polygonByID:     runtime.PolygonByID,
		loadPolygonZone: runtime.PolygonZone,
		loadPositionX: func(position *types.Pointf) float32 {
			return position.X
		},
		floatToInt: audioEventZoneFloatToInt501C00,
		loadPositionY: func(position *types.Pointf) float32 {
			return position.Y
		},
		polygonAtPoint: runtime.PolygonAtPoint,
	})
}

// AudioEventZone501C00 binds GAME.EXE 00501C00 to native Object, Player, and
// Pointf layouts while delegating legacy polygon storage and lookup.
func (s *Server) AudioEventZone501C00(
	position *types.Pointf,
	object *Object,
	runtime AudioEventZoneRuntime501C00,
) uint8 {
	return audioEventZoneNative501C00(position, object, runtime)
}
