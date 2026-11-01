package opennox

import (
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/server"
)

func (s *Server) ObjectSpecialRuntimeNative527E50() server.ObjectSpecialRuntime527E50 {
	return server.ObjectSpecialRuntime527E50{
		Animation: func(i int32, unit *server.Object) int32 { s.Nox_xxx_netReportAnimFrame_4D81F0(int(i), unit); return 0 },
		Enemy: func(recipient, unit *server.Object) int32 {
			if s.IsEnemyTo(recipient, unit) {
				return 1
			}
			return 0
		},
		Health: func(i int32, unit *server.Object) int32 { s.CurrentHPReport4D8620(i, unit); return 0 },
		Hidden: func(i int32, unit *server.Object) int32 {
			s.netReportObjectHiddenNative4D8FD0(ntype.PlayerInd(i), unit)
			return 0
		},
		XStatus: func(i int32, unit *server.Object) int32 {
			s.netReportXStatusNative4D8230(ntype.PlayerInd(i), unit)
			return 0
		},
		Height: func(i int32, unit *server.Object) int32 {
			s.netReportUnitHeightNative4D9020(ntype.PlayerInd(i), unit)
			return 0
		},
		Enchant: func(i int32, unit *server.Object) int32 {
			s.netReportEnchant_4D8F90(ntype.PlayerInd(i), unit)
			return 0
		},
		TeamBase: func(i int32, unit *server.Object) int32 {
			s.netReportTeamBaseNative4D92D0(ntype.PlayerInd(i), unit)
			return 0
		},
		NPC: func(i int32, unit *server.Object) int32 {
			s.netSendReportNPCNative4D93A0(ntype.PlayerInd(i), unit)
			return 0
		},
	}
}
