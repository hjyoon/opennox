package legacy

/*
#include "GAME3_2.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/server"
)

func Nox_xxx_netPlayerIncomingServ_4DDF60(index int) {
	nox_xxx_netPlayerIncomingServ_4DDF60(C.int(index))
}

//export nox_xxx_netPlayerIncomingServ_4DDF60
func nox_xxx_netPlayerIncomingServ_4DDF60(index C.int) C.int {
	world := GetServer().S()
	player := world.Players.ByInd(ntype.PlayerInd(index))
	if player == nil {
		panic("incoming player has no active slot")
	}
	if noxflags.HasGame(noxflags.GameModeQuest) {
		if index != server.HostPlayerIndex && player.PlayerUnit != nil {
			playerIncomingQuest4DDF60(player.PlayerUnit)
		}
		sub_4D9CF0(index)
		if player.PlayerUnit != nil {
			Sub_4D6000(player.PlayerUnit)
		}
	}
	sub_57B920(unsafe.Pointer(&player.NetData16[0]))
	unit := player.PlayerUnit
	setImportantRecipientMaskC(importantRecipientMaskC() | (uint32(1) << (uint32(index) & 31)))
	position := unit.PosVec
	nox_xxx_newPlayerSendAllPlayers_4DE300(index)
	initializeIncomingPlayer4DDF60(player, position, unit.CallInit)

	if Get_nox_server_sendMotd_108752() != 0 && noxflags.HasGame(noxflags.GameOnline) &&
		!noxflags.HasGame(noxflags.GameModeQuest) {
		nox_xxx_playerSendMOTD_4DD140(int32(index))
	}
	for peer := world.Players.First(); peer != nil; peer = world.Players.Next(peer) {
		if peer.PlayerUnit == nil || peer == player {
			continue
		}
		world.Players.Nox_xxx_netMarkMinimapObject_417190(ntype.PlayerInd(index), peer.PlayerUnit, 1)
		world.Players.Nox_xxx_netMarkMinimapObject_417190(peer.PlayerIndex(), player.PlayerUnit, 1)
		netSendSimpleObjectCall4DF360(int32(peer.Index()), player.PlayerUnit)
		if noxflags.HasGame(noxflags.GameModeQuest) {
			world.NetSendTeamHealth4D8670(peer.Index(), player.PlayerUnit)
			world.NetSendTeamHealth4D8670(int(index), peer.PlayerUnit)
		}
	}
	nox_xxx_servMinimapRevealFlag_4DE380(index)
	sub_4DF2E0(index)
	if noxflags.HasGame(noxflags.GameModeElimination) && sub40AA70Native(player.C()) == 0 {
		world.NeedPlayerStatus4174F0(player, 256)
	}
	sub_4E8110(int32(index))
	sendIncomingObjectiveStatus4DDF60(world, index)
	sendAllStatusNative4175C0(player)
	if sub409F40Native(0x2000) != 0 {
		world.SendIncomingCampingMarkers4DE270(player)
	}
	if noxflags.HasGame(noxflags.GameModeQuest) {
		for peer := world.Players.FirstUnit(); peer != nil; peer = world.NextPlayerUnit4DA7F0(peer) {
			if peer.UpdateDataPlayer().Player.Field4792 == 1 {
				world.QuestNotifyPlayer4D9D20(int(index), peer)
			}
		}
	}
	if noxflags.HasGame(noxflags.GameModeQuest) {
		questPlayerStateMark4D7A60(int(index))
	}
	return C.int(unit.NetCode)
}

func initializeIncomingPlayer4DDF60(player *server.Player, position types.Pointf, initialize func()) {
	player.Field4700 = 0
	initialize()
	player.Field3676 = 3
	if !noxflags.HasGame(noxflags.GameModeCoopTeam) {
		player.SetPos3632(position)
	}
}

func sendIncomingObjectiveStatus4DDF60(world *server.Server, index C.int) {
	if noxflags.HasGame(noxflags.GameModeFlagBall) {
		status := sub_4E8310()
		code, state := status.net_code, status.state
		nox_xxx_netSendBallStatus_4D95F0(index, C.char(state), C.short(code))
	} else if noxflags.HasGame(noxflags.GameModeCTF) {
		for team := world.Teams.First(); team != nil; team = world.Teams.Next(team) {
			status := sub_4E8320(uint8(team.ID()))
			code, flagIndex, state, id := status.carrier_net_code, status.flag_index, status.status, status.team_id
			nox_xxx_netSendFlagStatus_4D95A0(index, C.char(id), C.char(state), C.char(flagIndex), C.short(code))
		}
	}
}

//export nox_xxx_sendAllPlayerIDs_4DE270
func nox_xxx_sendAllPlayerIDs_4DE270(player *nox_playerInfo) C.int {
	return C.int(GetServer().S().SendIncomingCampingMarkers4DE270(asPlayerS(player)))
}
