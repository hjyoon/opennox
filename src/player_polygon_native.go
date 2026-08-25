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
	return legacy.Nox_xxx_polygonIsPlayerInPolygon_4217B0(unsafe.Pointer(&point[0]), int(int32(previous)))
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

// audioEventZoneNative501C00 is the native-pointer counterpart of GAME.EXE
// sub_501C00. It intentionally keeps polygon storage and point lookup in the
// legacy fixed-width map oracle.
func (s *Server) audioEventZoneNative501C00(pos types.Pointf, obj *server.Object) byte {
	zone := byte(0)
	if obj != nil {
		if obj.Class().Has(object.ClassPlayer) {
			zone = obj.ControllingPlayer().AudioZone()
			if zone != 0 {
				return zone
			}
		} else if obj.Class().Has(object.ClassMonster) {
			polygon := polygonByIDNative4214A0(obj.UpdateDataMonster().Field0)
			zone = polygonAudioZoneNative501C00(polygon)
			if zone != 0 {
				return zone
			}
		}
	}
	if polygon := polygonAtPointNative4217B0(pos, 0); polygon != nil {
		return polygonAudioZoneNative501C00(polygon)
	}
	return zone
}

func (s *Server) remotePlayerAudioZoneNative501CA0(unit *server.Object) byte {
	update := unit.UpdateDataPlayer()
	player := update.Player
	follow := player.CameraTarget()
	if player.Field3680&3 == 0 || follow == nil {
		if player.CurrentPolygonID() == playerPolygonUninitialized421C70 {
			s.questCheckSecretAreaNative421C70(unit)
		}
		return player.AudioZone()
	}
	if follow.Class().Has(object.ClassPlayer) {
		return follow.ControllingPlayer().AudioZone()
	}
	return polygonAudioZoneNative501C00(polygonAtPointNative4217B0(follow.PosVec, 0))
}

func (s *Server) netUpdateRemotePlayerNative501CA0(unit *server.Object) {
	update := unit.UpdateDataPlayer()
	zone := s.remotePlayerAudioZoneNative501CA0(unit)
	s.netUpdateRemotePlrAudioEventsNative(unit, update, zone)
}
