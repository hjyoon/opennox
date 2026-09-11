package opennox

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

const (
	playerPolygonUninitialized421C70 = uint32(0xDEADFACE)
	playerPolygonRecordSize421C70    = uintptr(140)
	playerPolygonRecordCount421C70   = uint32(255)
)

func polygonFloatToIntNative4217B0(value float32) int32 {
	if math.IsNaN(float64(value)) || value >= 2147483648 || value < -2147483648 {
		return math.MinInt32
	}
	return int32(math.RoundToEven(float64(value)))
}

func polygonByIDNative4214A0(id uint32) *legacy.Nox_player_polygon_check_data {
	if id == playerPolygonUninitialized421C70 || id >= playerPolygonRecordCount421C70 {
		return nil
	}
	return (*legacy.Nox_player_polygon_check_data)(memmap.PtrOff(0x5D4594, 552228+playerPolygonRecordSize421C70*uintptr(id)))
}

func polygonAtPointNative4217B0(pos types.Pointf, previous uint32) *legacy.Nox_player_polygon_check_data {
	point := [2]int32{
		polygonFloatToIntNative4217B0(pos.X),
		polygonFloatToIntNative4217B0(pos.Y),
	}
	return polygonAtIntPointNative4217B0(point, previous)
}

func polygonAtIntPointNative4217B0(point [2]int32, previous uint32) *legacy.Nox_player_polygon_check_data {
	return legacy.Nox_xxx_polygonIsPlayerInPolygon_4217B0(unsafe.Pointer(&point[0]), int(int32(previous)))
}

func monsterPolygonAtPointNative421F10(pos types.Pointf, previous uint32) *legacy.Nox_player_polygon_check_data {
	point := [2]int32{
		polygonFloatToIntNative4217B0(pos.X),
		polygonFloatToIntNative4217B0(pos.Y),
	}
	return legacy.Sub_421F10(unsafe.Pointer(&point[0]), int(int32(previous)))
}

func polygonIDNative421C70(polygon *legacy.Nox_player_polygon_check_data) uint32 {
	if polygon == nil {
		return 0
	}
	return polygon.Field_0[20]
}

func polygonAudioZoneNative501C00(polygon *legacy.Nox_player_polygon_check_data) byte {
	if polygon == nil {
		return 0
	}
	return byte(polygon.Field_0[32] >> 16)
}

func (s *Server) polygonScriptCallbackNative421C70(polygon *legacy.Nox_player_polygon_check_data, word int, unit *server.Object, event server.ScriptEventType) {
	callback := (*server.ScriptCallback)(unsafe.Pointer(&polygon.Field_0[word]))
	if callback.Func != -1 {
		s.noxScript.ScriptCallback(callback, unit, nil, event)
	}
}

type monsterPolygonHooks421FF0 struct {
	find     func(types.Pointf, uint32) *legacy.Nox_player_polygon_check_data
	byID     func(uint32) *legacy.Nox_player_polygon_check_data
	callback func(*legacy.Nox_player_polygon_check_data, int, *server.Object, server.ScriptEventType)
}

func monsterPolygonEnterNative421FF0(unit *server.Object, hooks monsterPolygonHooks421FF0) {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) {
		return
	}
	update := unit.UpdateDataMonster()
	current := update.Field0
	if current != playerPolygonUninitialized421C70 && unit.PosVec == unit.PrevPos {
		return
	}

	polygon := hooks.find(unit.PosVec, current)
	if polygon != nil {
		id := polygonIDNative421C70(polygon)
		if current == id {
			return
		}
		if current != playerPolygonUninitialized421C70 {
			if current != 0 {
				if old := hooks.byID(current); old != nil && old.Field_0[31] != ^uint32(0) {
					hooks.callback(old, 30, unit, server.NoxEventPolygonEnterZZZ)
				}
			}
			if polygon.Field_0[29] != ^uint32(0) {
				hooks.callback(polygon, 28, unit, server.NoxEventPolygonEnterYYY)
			}
		}
		update.Field0 = id
		return
	}
	if current == 0 || current == playerPolygonUninitialized421C70 {
		return
	}
	if old := hooks.byID(current); old != nil && old.Field_0[21] != 0 && old.Field_0[31] != ^uint32(0) {
		hooks.callback(old, 30, unit, server.NoxEventPolygonEnterXXX)
	}
	update.Field0 = 0
}

// MonsterPolygonEnterNative421FF0 preserves the original monster polygon
// transition callbacks while keeping the monster and its update pointer native.
// Polygon records and point-in-polygon lookup remain fixed-width oracle data.
func (s *Server) MonsterPolygonEnterNative421FF0(unit *server.Object) {
	monsterPolygonEnterNative421FF0(unit, monsterPolygonHooks421FF0{
		find: func(pos types.Pointf, previous uint32) *legacy.Nox_player_polygon_check_data {
			if previous == playerPolygonUninitialized421C70 {
				return polygonAtPointNative4217B0(pos, 0)
			}
			return monsterPolygonAtPointNative421F10(pos, previous)
		},
		byID:     polygonByIDNative4214A0,
		callback: s.polygonScriptCallbackNative421C70,
	})
}

// questCheckSecretAreaNative421C70 preserves GAME.EXE 0x421C70's polygon
// transition order while keeping Object, PlayerUpdateData, and Player pointers
// native-width. Polygon records remain fixed 140-byte oracle data.
func (s *Server) questCheckSecretAreaNative421C70(unit *server.Object) {
	if unit == nil || !unit.Class().Has(object.ClassPlayer) {
		return
	}
	update := unit.UpdateDataPlayer()
	player := update.Player
	if player == nil {
		return
	}

	current := player.CurrentPolygonID()
	var polygon *legacy.Nox_player_polygon_check_data
	if player.Index() == server.HostPlayerIndex {
		id := player.LocalPolygonID()
		if id == 0 {
			player.SetCurrentPolygonID(0)
			return
		}
		if id != playerPolygonUninitialized421C70 {
			polygon = polygonByIDNative4214A0(id)
		}
	} else {
		if current != playerPolygonUninitialized421C70 && unit.PosVec == unit.PrevPos {
			return
		}
		polygon = polygonAtPointNative4217B0(unit.PosVec, current)
		if polygon == nil && current != 0 && current != playerPolygonUninitialized421C70 {
			polygon = polygonByIDNative4214A0(current)
		}
	}

	if polygon == nil {
		if current != 0 && current != playerPolygonUninitialized421C70 {
			old := polygonByIDNative4214A0(current)
			if old != nil {
				s.polygonScriptCallbackNative421C70(old, 30, unit, server.NoxEventPolygonPlayerXXX)
			}
			player.SetCurrentPolygonID(0)
			player.SetAudioZone(1)
		}
		return
	}

	id := polygonIDNative421C70(polygon)
	if current == id {
		return
	}
	if current != playerPolygonUninitialized421C70 {
		if current != 0 {
			old := polygonByIDNative4214A0(current)
			if old != nil {
				s.polygonScriptCallbackNative421C70(old, 30, unit, server.NoxEventPolygonPlayerZZZ)
			}
		}

		playerBit := uint32(1) << uint(player.PlayerInd)
		if polygon.Field_0[34]&playerBit == 0 && byte(polygon.Field_0[33])&1 != 0 && noxflags.HasGame(noxflags.GameModeQuest) {
			if !unit.Flags().Has(object.FlagDestroyed) {
				player.RecordSecretFound()
			}
			s.NetPriMsgToPlayer(unit, "GeneralPrint:SecretFound", 0)
			s.Audio.EventObj(sound.SoundSecretFound, unit, 0, 0)
			for _, other := range s.Players.ListUnits() {
				if other != unit {
					s.NetInformTextMsg(other.ControllingPlayer().PlayerIndex(), 20, int(unit.NetCode))
				}
			}
			polygon.Field_0[33] &^= 1
		}
		polygon.Field_0[34] |= playerBit
		s.polygonScriptCallbackNative421C70(polygon, 28, unit, server.NoxEventPolygonPlayerEnter)
	}
	player.SetCurrentPolygonID(id)
	player.SetAudioZone(polygonAudioZoneNative501C00(polygon))
}

func audioEventZoneRuntimeNative501C00() server.AudioEventZoneRuntime501C00 {
	return server.AudioEventZoneRuntime501C00{
		PolygonByID: func(id uint32) unsafe.Pointer {
			return unsafe.Pointer(polygonByIDNative4214A0(id))
		},
		PolygonAtPoint: func(point [2]int32, previous uint32) unsafe.Pointer {
			return unsafe.Pointer(polygonAtIntPointNative4217B0(point, previous))
		},
		PolygonZone: func(polygon unsafe.Pointer) uint8 {
			return uint8((*legacy.Nox_player_polygon_check_data)(polygon).Field_0[32] >> 16)
		},
	}
}

// audioEventZonePtrNative501C00 is the native-pointer counterpart of
// GAME.EXE sub_501C00. It preserves the position pointer until the original
// control flow reaches the polygon fallback.
func (s *Server) audioEventZonePtrNative501C00(pos *types.Pointf, obj *server.Object) byte {
	return byte(s.Server.AudioEventZone501C00(pos, obj, audioEventZoneRuntimeNative501C00()))
}

func (s *Server) audioEventZoneNative501C00(pos types.Pointf, obj *server.Object) byte {
	return s.audioEventZonePtrNative501C00(&pos, obj)
}
