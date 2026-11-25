package legacy

/*
#include "common__system__team.h"
#include "client__gui__window.h"

extern nox_window* dword_5d4594_1045604;
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_createAtImpl_4191D0
func nox_xxx_createAtImpl_4191D0(teamID C.uchar, value *C.nox_object_team_t, active, code, flags C.int) {
	world := GetServer().S()
	world.TeamCreateAtImpl4191D0(
		server.TeamID(teamID),
		(*server.ObjectTeam)(unsafe.Pointer(value)),
		int32(active),
		uint32(code),
		int32(flags),
		teamCreateAtRuntimeNative4191D0(world),
	)
}

func teamMemberJoined(team *server.Team, player *nox_playerInfo, questLine, teamLine int32) {
	if player == nil {
		return
	}
	const source = `C:\NoxPost\src\common\System\team.c`
	key, line := "NewMember", teamLine
	var teamName *wchar2_t
	if noxflags.HasGame(noxflags.GameModeQuest) {
		key, line = "GeneralPrint:PlayerJoinQuest", questLine
	} else {
		rawTeam := (*nox_team_t)(team.C())
		teamName = &rawTeam.name[0]
	}
	format := nox_strman_loadString_40F1D0(internCStr(key), nil, internCStr(source), line)
	var text [128]wchar2_t
	formatWideStrings(text[:], format, &player.name_final[0], teamName)
	nox_xxx_printCentered_445490(&text[0])
}

func teamHostAfterAttach(world *server.Server, team *server.Team, active int32, code uint32, flags int32, gameBall uint16) {
	if !noxflags.HasGame(noxflags.GameOnline) {
		return
	}
	unit := world.ObjectFromNetCode4ECCB0(code)
	player := nox_common_playerInfoGetByID_417040(int32(code))
	if player != nil {
		if noxflags.HasGame(noxflags.GameFlag16) {
			sub_425ED0(player, 1)
		}
		if unit != nil && unit.Class().Has(object.ClassPlayer) {
			if flags == 1 && !noxflags.HasGamePlay(noxflags.GameplayFlag2) && noxflags.HasGame(noxflags.GameModeChat) {
				if flag := (*server.Object)(team.Field_72); flag != nil {
					var position C.float2
					center := (*C.float2)(unsafe.Pointer(&flag.PosVec))
					sub_4ED970(50, center, &position)
					nox_xxx_unitMove_4E7010(asObjectC(unit), &position)
				}
			}
			teamMemberJoined(team, nox_common_playerInfoGetByID_417040(int32(code)), 1848, 1850)
		}
	}
	if active != 0 && unit != nil && (player != nil || unit.TypeInd == gameBall) {
		packet := [10]byte{byte(netmsg.MSG_TEAM_MSG), 1}
		binary.LittleEndian.PutUint32(packet[2:6], uint32(team.ID()))
		binary.LittleEndian.PutUint16(packet[6:8], uint16(code))
		binary.LittleEndian.PutUint16(packet[8:10], unit.TypeInd)
		sub_4571A0(C.int(code), C.int(team.ID()))
		world.NetSendPacketXxx1(159, packet[:], nil, 1)
	}
}

func teamCreateAtRuntimeNative4191D0(world *server.Server) server.TeamCreateAtRuntime4191D0 {
	return server.TeamCreateAtRuntime4191D0{
		ClientNetCode: uint32(ClientPlayerNetCode()),
		SelectLocalTeam: func(id server.TeamID) {
			if C.dword_5d4594_1045604 != nil {
				sub_455E70(C.uchar(id))
			}
		},
		AfterAttach: func(team *server.Team, _ *server.ObjectTeam, active int32, code uint32, flags int32, gameBall uint16) {
			if noxflags.HasGame(noxflags.GameHost) {
				teamHostAfterAttach(world, team, active, code, flags, gameBall)
				return
			}
			if GetClient == nil || GetClient() == nil {
				return
			}
			drawable := nox_xxx_netSpriteByCodeDynamic_45A6F0(int32(code))
			if drawable != nil && object.Class(drawable.flags28).Has(object.ClassPlayer) {
				teamMemberJoined(team, nox_common_playerInfoGetByID_417040(int32(code)), 1889, 1891)
			}
		},
	}
}
