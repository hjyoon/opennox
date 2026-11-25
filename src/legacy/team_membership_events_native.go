package legacy

/*
#include "GAME1_1.h"
#include "GAME2.h"
#include "GAME3_3.h"
#include "client__gui__servopts__playrlst.h"
#include "common__system__team.h"

extern unsigned int nox_player_netCode_85319C;
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func sendTeamMembershipPacket(packet []byte) {
	nox_xxx_netSendPacket1_4E5390(159, C.nox_important_payload_t(unsafe.Pointer(&packet[0])), C.int(len(packet)), nil, 1)
}

//export nox_xxx_teamRenameMB_418CD0
func nox_xxx_teamRenameMB_418CD0(team *C.nox_team_t, name *C.wchar2_t) {
	if team == nil || name == nil {
		return
	}
	C.sub_457010(team, name)
	team.field_68 = 0

	var units [22]C.wchar2_t
	n := 0
	for ; n < len(units)-1; n++ {
		value := *(*C.wchar2_t)(unsafe.Add(unsafe.Pointer(name), uintptr(n)*unsafe.Sizeof(*name)))
		if value == 0 {
			break
		}
		units[n] = value
	}
	if noxflags.HasGame(noxflags.GameHost) {
		packet := [46]byte{byte(netmsg.MSG_TEAM_MSG), 4}
		binary.LittleEndian.PutUint32(packet[2:], uint32(team.field_57))
		for i, unit := range units[:min(n, (len(packet)-6)/2)] {
			binary.LittleEndian.PutUint16(packet[6+2*i:], uint16(unit))
		}
		sendTeamMembershipPacket(packet[:])
	}
	copy(team.name[:], units[:n+1])
}

//export sub_418D80
func sub_418D80(team *C.nox_team_t) {
	if sub_418BC0(team) == 0 {
		return
	}
	world := GetServer().S()
	if noxflags.HasGame(noxflags.GameHost) {
		packet := [6]byte{byte(netmsg.MSG_TEAM_MSG), 5}
		binary.LittleEndian.PutUint32(packet[2:], uint32(team.field_57))
		sendTeamMembershipPacket(packet[:])
	}
	for player := world.Players.First(); player != nil; player = world.Players.Next(player) {
		member := teamObjectByNetCode418C80(player.NetCodeVal)
		if member != nil && member.ID == server.TeamID(team.field_57) {
			sub_4571A0(C.int(player.NetCodeVal), 0)
			sub_418E40(team, (*C.nox_object_team_t)(member.C()))
		}
	}
}

//export sub_418E40
func sub_418E40(team *C.nox_team_t, value *C.nox_object_team_t) *C.nox_object_team_t {
	if team == nil || value == nil {
		return value
	}
	world := GetServer().S()
	member := (*server.ObjectTeam)(unsafe.Pointer(value))
	if !world.Teams.ContainsObject(member, server.TeamID(team.field_57)) {
		return value
	}
	world.Teams.DetachObject(member)
	for unit := world.Players.FirstUnit(); unit != nil; unit = world.NextPlayerUnit4DA7F0(unit) {
		if &unit.TeamVal != member {
			continue
		}
		if player := nox_common_playerInfoGetByID_417040(int32(unit.NetCode)); player != nil {
			sub_4D97E0(C.int(player.playerInd))
			world.RebuildObjectPlayerMasks4E8110(int32(player.playerInd))
		}
		world.Nox_xxx_monsterMarkUpdate_4E8020(unit)
		for owned := unit.FirstOwned516(); owned != nil; owned = owned.NextOwned512() {
			if owned.ObjClass&(object.ClassMonster|object.ClassPlayer) != 0 {
				world.Nox_xxx_monsterMarkUpdate_4E8020(owned)
			}
		}
		break
	}
	return value
}

//export nox_xxx_netChangeTeamID_419090
func nox_xxx_netChangeTeamID_419090(team *C.nox_team_t, lessons C.int) {
	if team == nil {
		return
	}
	team.lessons = lessons
	if noxflags.HasGame(noxflags.GameHost) {
		packet := [10]byte{byte(netmsg.MSG_TEAM_MSG), 8}
		binary.LittleEndian.PutUint32(packet[2:], uint32(team.field_57))
		binary.LittleEndian.PutUint32(packet[6:], uint32(lessons))
		sendTeamMembershipPacket(packet[:])
	}
}

//export nox_xxx_netChangeTeamMb_419570
func nox_xxx_netChangeTeamMb_419570(value *C.nox_object_team_t, netCode C.int) {
	if value == nil {
		return
	}
	world := GetServer().S()
	member := (*server.ObjectTeam)(unsafe.Pointer(value))
	team := (*C.nox_team_t)(world.Teams.ByID(member.ID).C())
	if team == nil || !world.Teams.ContainsObject(member, member.ID) {
		return
	}
	if noxflags.HasGame(noxflags.GameHost) && noxflags.HasGame(noxflags.GameOnline) {
		nox_common_playerInfoGetByID_417040(int32(netCode))
		sub_4571A0(netCode, 0)
		packet := [6]byte{byte(netmsg.MSG_TEAM_MSG), 2}
		binary.LittleEndian.PutUint32(packet[2:], uint32(netCode))
		sendTeamMembershipPacket(packet[:])
	}
	sub_418E40(team, value)
	if (sub_40A740() != 0 || noxflags.HasGame(noxflags.GameFlag16)) && sub_418BC0(team) == 0 {
		if noxflags.HasGame(noxflags.GameModeCTF|noxflags.GameModeFlagBall) ||
			(noxflags.HasGame(noxflags.GameModeKOTR) && bool(nox_xxx_CheckGameplayFlags_417DA0(4))) {
			team.field_60 = 0
			sub_418800(&team.name[0], (*C.wchar2_t)(memmap.PtrOff(0x5D4594, 527664)), 0)
		} else {
			sub_418F20(team, 1)
		}
	}
}

//export sub_4196D0
func sub_4196D0(rawMember, rawTeam unsafe.Pointer, netCode, flags C.int) C.int {
	value, team := (*C.nox_object_team_t)(rawMember), (*C.nox_team_t)(rawTeam)
	if value == nil || team == nil || nox_xxx_teamCompare2_419180(value, value.id) == 0 {
		return 0
	}
	nox_common_playerInfoGetByID_417040(int32(netCode))
	if noxflags.HasGame(noxflags.GameHost) && noxflags.HasGame(noxflags.GameOnline) {
		packet := [10]byte{byte(netmsg.MSG_TEAM_MSG), 3}
		binary.LittleEndian.PutUint32(packet[2:], uint32(team.field_57))
		binary.LittleEndian.PutUint16(packet[6:], uint16(netCode))
		sendTeamMembershipPacket(packet[:])
		sub_4571A0(netCode, C.int(team.field_57))
	}
	previous := nox_xxx_getTeamByID_418AB0(int32(value.id))
	if previous == nil {
		return 0
	}
	if previous.field_48 > 0 {
		previous.field_48--
	}
	sub_418E40(previous, value)
	nox_xxx_createAtImpl_4191D0(team.field_57, value, 0, netCode, flags)
	if C.uint(netCode) == C.nox_player_netCode_85319C {
		sub_455E70(team.field_57)
	}
	return 1
}

func requestTeamChange(team *C.nox_team_t, value *C.nox_object_team_t, netCode C.short, operation byte) C.char {
	if team == nil || value == nil {
		return 0
	}
	if value.id == team.field_57 {
		return C.char(team.field_57)
	}
	packet := [10]byte{byte(netmsg.MSG_TEAM_MSG), operation}
	binary.LittleEndian.PutUint32(packet[2:], uint32(team.field_57))
	binary.LittleEndian.PutUint16(packet[6:], uint16(netCode))
	return C.char(nox_xxx_netClientSend2_4E53C0(31, C.nox_important_payload_t(unsafe.Pointer(&packet[0])), C.int(len(packet)), nil, 1))
}

//export sub_419900
func sub_419900(team *C.nox_team_t, value *C.nox_object_team_t, netCode C.short) C.char {
	return requestTeamChange(team, value, netCode, 10)
}

//export sub_419960
func sub_419960(team *C.nox_team_t, value *C.nox_object_team_t, netCode C.short) C.char {
	return requestTeamChange(team, value, netCode, 11)
}

//export nox_xxx_countObserverPlayers_425BF0
func nox_xxx_countObserverPlayers_425BF0() C.int {
	if !noxflags.HasGame(noxflags.GameFlag16) {
		return 0
	}
	players := &GetServer().S().Players
	var count C.int
	for player := players.First(); player != nil; player = players.Next(player) {
		if player.Field3680&1 != 0 && player.Field3680&0x20 == 0 && player.PlayerInd != 31 {
			count++
		}
	}
	return count
}
