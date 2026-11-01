package opennox

import (
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func (s *Server) PlayerObjectSendRuntimeNative518C30() server.PlayerObjectSendRuntime518C30 {
	word := func(value bool) int32 {
		if value {
			return 1
		}
		return 0
	}
	return server.PlayerObjectSendRuntime518C30{
		Frame:       s.Frame,
		HealthDelta: legacy.PlayerSendHealthDeltaNative4D8760,
		Self:        s.playerReportSelfNative4D9900,
		UnitCode:    func(unit *server.Object) uint32 { return uint32(s.GetUnitNetCode(unit)) },
		FloatToInt:  legacy.PlayerSendFloatToIntNative419A70,
		Action:      func(unit *server.Object) uint32 { return uint32(s.Server.PlayerActionState4FA2B0(unit)) },
		Random:      func(min, max int32) int32 { return int32(s.Rand.Logic.IntClamp(int(min), int(max))) },
		Frames: func(index int32) (int32, int32) {
			count, delay := s.PlayerAnimFrames(int(index))
			return int32(count), int32(delay)
		},
		Direction: legacy.PlayerSendDirectionNative509E20,
		ServerPacket: func(index int32, packet [12]byte) int32 {
			return word(nox_netlist_addToMsgListSrv(ntype.PlayerInd(index), packet[:]))
		},
		ClientPacket: func(index, kind int32, packet [12]byte) int32 {
			return word(s.NetList.AddToMsgListCli(ntype.PlayerInd(index), netlist.Kind(kind), packet[:]))
		},
	}
}
