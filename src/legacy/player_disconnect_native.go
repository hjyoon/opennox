package legacy

/*
#include "GAME3_2.h"
#include "quest_map_buffer_4e8e50.h"

#ifdef NOX_SERVER
enum { nox_force_disconnect_dedicated = 1 };
#else
enum { nox_force_disconnect_dedicated = 0 };
#endif
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_playerForceDisconnect_4DE7C0
func nox_xxx_playerForceDisconnect_4DE7C0(index C.int) {
	outer := GetServer()
	world := outer.S()
	player := world.Players.ByInd(ntype.PlayerInd(index))
	outer.Nox_script_event_playerLeave(player)
	sub_4D1250(index)
	if player.Field2068 != 0 {
		if record := sub_425A70(C.int(player.Field2068)); record != nil {
			sub_425B60(unsafe.Pointer(record), index)
		}
	}

	update := player.PlayerUnit.UpdateDataPlayer()
	if update.Trade70 != nil {
		nox_xxx_shopCancelSession_510DC0(unsafe.Pointer(update.Trade70))
	}
	update.Trade70 = nil
	sub_510E20(C.int(player.Index()))
	sub_4FF990(C.uint(uint32(1) << (player.PlayerInd & 31)))
	if C.nox_force_disconnect_dedicated != 0 || !noxflags.HasGame(noxflags.GameClient) {
		clear(unsafe.Slice(&player.Active, 4))
	}
	clearDisconnectedPlayerProtection4DE7C0(player)

	packet := [3]byte{byte(netmsg.MSG_PLAYER_QUIT)}
	binary.LittleEndian.PutUint16(packet[1:], uint16(world.GetUnitNetCode(player.PlayerUnit)))
	sendImportantPacketBytes(index|0x80, packet[:], false, false)
	outer.DelayedDelete(player.PlayerUnit)
	player.PlayerUnit = nil
	clearDisconnectedPeerState4DE7C0(world, int(index))

	if gamePlayAnyNative40A8A0() != 0 {
		if noxflags.HasGame(noxflags.GameModeElimination) && nox_xxx_serverIsClosing_446180() == 0 && sub40A770Native() == 1 {
			sub_5095E0()
		}
	} else {
		sub40A1F0Native(0)
		playerForceLessonsNative416E50(1)
		world.TeamsResetYyy()
		sub40A970Native()
	}
	sub_4E55F0(C.uint8_t(index))
	nox_xxx_playerResetImportantCtr_4E4F40(index)
	sub_4E4F30(index)
	if noxflags.HasGame(noxflags.GameModeQuest) {
		resumeQuestAfterDisconnect4DE7C0(world)
	}
}

func clearDisconnectedPlayerProtection4DE7C0(player *server.Player) {
	tokens := [...]*uint32{
		&player.ProtUnitHPCur, &player.ProtUnitHPMax, &player.ProtUnitManaCur, &player.ProtUnitManaMax,
		&player.ProtUnitExperience, &player.ProtUnitMass, &player.ProtUnitBuffs, &player.ProtPlayerClass,
		&player.ProtPlayerField2235, &player.ProtPlayerField2239, &player.ProtPlayerOrigName,
		&player.Prot4632, &player.Prot4636, &player.ProtPlayerGold, &player.Prot4640, &player.ProtPlayerLevel,
	}
	for _, token := range tokens {
		sub_56F4F0((*C.int)(unsafe.Pointer(token)))
	}
}

func clearDisconnectedPeerState4DE7C0(world *server.Server, index int) {
	for unit := world.Players.FirstUnit(); unit != nil; unit = world.NextPlayerUnit4DA7F0(unit) {
		data := unit.UpdateDataPlayer()
		data.RespawnMarkers[index] = 0
		data.QuestPlayerState[index] = 0
		data.QuestPlayerFlagsA[index] = 0
		data.QuestPlayerFlagsB[index] = 0
	}
}

func resumeQuestAfterDisconnect4DE7C0(world *server.Server) {
	switch {
	case sub_4E9010() == 1:
		nox_xxx_mapLoad_4D2450(sub_4E8E50())
	case nox_server_questMaybeWarp_4E8F60() != 0:
		sub_4D60B0()
		sub_4D76E0(1)
		stage := nox_game_getQuestStage_4E3CC0()
		nox_game_setQuestStage_4E3CD0(nox_server_questNextStageThreshold_4D74F0(stage) - 1)
		nox_xxx_mapLoad_4D2450(sub_4E8E50())
	default:
		for unit := world.Players.FirstUnit(); unit != nil; unit = world.NextPlayerUnit4DA7F0(unit) {
			if unit.UpdateDataPlayer().QuestExit != nil {
				sub_4E8E60()
				break
			}
		}
	}
}
