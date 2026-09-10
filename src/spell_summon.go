package opennox

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

var (
	cheatSummonNoLimit = false
)

func nox_xxx_checkSummonedCreaturesLimit_500D70(u *server.Object, ind int32) bool {
	if cheatSummonNoLimit {
		return true
	}
	return server.CheckSummonedCreaturesLimit500D70(u, ind, func(index int32) int32 {
		return int32(nox_xxx_guideGetUnitSize_427460(int(index)))
	})
}

func nox_xxx_guideGetUnitSize_427460(ind int) int {
	return int(memmap.Uint8(0x5D4594, 740100+28*uintptr(ind)))
}

func summonUnitRuntime5016C0(s *Server) server.SummonUnitRuntime5016C0 {
	return server.SummonUnitRuntime5016C0{
		CreateObjectAt: func(obj, owner *server.Object, pos types.Pointf) {
			s.CreateObjectAt(obj, owner, pos)
		},
		OrderUnit: legacy.Nox_xxx_orderUnit_533900,
		ReportAcquire: func(playerInd uint8, obj *server.Object) {
			legacy.Nox_xxx_netReportAcquireCreature_4D91A0(int(playerInd), obj)
		},
		MarkMinimap: func(playerInd uint8, obj *server.Object, flags uint32) {
			s.Players.Nox_xxx_netMarkMinimapObject_417190(ntype.PlayerInd(playerInd), obj, flags)
		},
		SendSimpleObject: func(playerInd uint8, obj *server.Object) {
			legacy.Nox_xxx_netSendSimpleObject2_4DF360(int(playerInd), obj)
		},
		CreateTeam: func(id server.TeamID, team *server.ObjectTeam, active int32, netCode uint32, flags int32) {
			legacy.Nox_xxx_createAtImpl_4191D0(id, team, int(active), int(netCode), int(flags))
		},
	}
}

func nox_xxx_unitDoSummonAt_5016C0(typID int32, pos *types.Pointf, owner *server.Object, dir uint8) *server.Object {
	s := noxServer
	return s.Server.SummonUnitAt5016C0(typID, pos, owner, dir, summonUnitRuntime5016C0(s))
}
