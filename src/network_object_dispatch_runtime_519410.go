package opennox

import (
	"math"

	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// ObjectDispatchRuntimeNative519410 binds every reached child to its native
// production service. The remaining PE32 C sender bodies are not reentered.
func (s *Server) ObjectDispatchRuntimeNative519410() server.ObjectDispatchNativeRuntime519410 {
	word := func(value bool) uint32 {
		if value {
			return 1
		}
		return 0
	}
	return server.ObjectDispatchNativeRuntime519410{
		Friend: func(index int32, unit *server.Object, add uint32) uint32 {
			s.netFriendAddRemoveNative4D97A0(ntype.PlayerInd(index), unit, add != 0)
			return 0
		},
		Owner: func(unit, recipient *server.Object) uint32 { return word(unit.HasOwner(recipient)) },
		Game:  func(flag uint32) uint32 { return word(noxflags.HasGame(noxflags.GameFlag(flag))) },
		Tracks: func(index int32, unit *server.Object) uint32 {
			return word(s.Players.ByIndRaw(ntype.PlayerInd(index)).MinimapTracks(unit))
		},
		Ray: func(bits [4]uint32, _, _ int32, flags int32) uint32 {
			from := types.Pointf{X: math.Float32frombits(bits[0]), Y: math.Float32frombits(bits[1])}
			to := types.Pointf{X: math.Float32frombits(bits[2]), Y: math.Float32frombits(bits[3])}
			return word(s.MapTraceRayAt(from, to, nil, nil, server.MapTraceFlags(flags)))
		},
		Sprite: func(recipient *server.Object, index int32, unit *server.Object) uint32 {
			return uint32(legacy.NetSpriteUpdate518AE0(recipient, index, unit))
		},
		Complex: func(index int32, unit *server.Object, health int32) uint32 {
			return uint32(legacy.NetSendComplexObject518960(index, unit, health))
		},
		Player: func(recipient, unit *server.Object, stream, _ int32) uint32 {
			return word(s.netPlayerObjectSendNative518C30(recipient, unit, stream != 0))
		},
		Phantom: func(index int32, unit *server.Object) uint32 {
			packet := s.phantomObjectPacketNative5187E0(unit)
			return word(nox_netlist_addToMsgListSrv(ntype.PlayerInd(index), packet[:]))
		},
		Simple: func(index int32, unit *server.Object) uint32 {
			if uint32(unit.ObjClass)&0x20000 != 0 {
				s.reportUnitHealthDeltaNative4D8760(int(index), unit)
			}
			packet := s.simpleObjectPacketNative5188A0(unit)
			return word(nox_netlist_addToMsgListSrv(ntype.PlayerInd(index), packet[:]))
		},
		Special: func(recipient, unit *server.Object) uint32 {
			return word(s.netUpdateObjectSpecialNative527E50(recipient, unit))
		},
		Out: func(index int32, unit *server.Object) uint32 {
			return uint32(s.Nox_xxx_netObjectOutOfSight_528A60(int(index), unit))
		},
		Shadow: func(index int32, unit *server.Object) uint32 {
			return uint32(s.Nox_xxx_netObjectInShadows_528A90(int(index), unit))
		},
	}
}
