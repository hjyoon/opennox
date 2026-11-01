package opennox

import (
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func (s *Server) BasicObjectSendRuntimeNative5187E0() server.BasicObjectSendRuntime5187E0 {
	return server.BasicObjectSendRuntime5187E0{
		Frame:       s.Frame,
		HealthDelta: legacy.PlayerSendHealthDeltaNative4D8760,
		UnitCode:    func(unit *server.Object) uint32 { return uint32(s.GetUnitNetCode(unit)) },
		FloatToInt:  legacy.PlayerSendFloatToIntNative419A70,
		Direction:   legacy.PlayerSendDirectionNative509E20,
		ServerPacket: func(index int32, packet []byte) int32 {
			if nox_netlist_addToMsgListSrv(ntype.PlayerInd(index), packet) {
				return 1
			}
			return 0
		},
	}
}
func (s *Server) netSendSimpleObjectNative5188A0(index int32, unit *server.Object) int32 {
	runtime := s.BasicObjectSendRuntimeNative5187E0()
	runtime.GuardedInputs = true
	return server.SimpleObjectSendNative5188A0(index, unit, runtime)
}
