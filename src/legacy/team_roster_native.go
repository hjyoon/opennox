package legacy

/*
#include "GAME1_1.h"
#include "GAME3_2.h"
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//export sub_418800
func sub_418800(name, source *C.wchar2_t, mode C.int) {
	if name == nil {
		return
	}
	team := (*nox_team_t)(unsafe.Pointer(name))
	i := 0
	for ; i < 20; i++ {
		v := *(*C.wchar2_t)(unsafe.Add(unsafe.Pointer(source), uintptr(i)*unsafe.Sizeof(*source)))
		if v == 0 {
			break
		}
		team.name[i] = v
	}
	clear(team.name[i:])
	team.field_68 = C.uint32_t(mode)
}

//export sub_418830
func sub_418830(team *nox_team_t, id C.int) {
	if team != nil {
		team.field_60 = C.uint32_t(id)
	}
}

//export sub_4189D0
func sub_4189D0() *nox_team_t {
	world := GetServer().S()
	minimum := 32
	var selected *server.Team
	for team := world.Teams.First(); team != nil; team = world.Teams.Next(team) {
		count := countTeamPlayers418BC0(world, (*nox_team_t)(team.C()))
		if count < minimum {
			minimum = count
			selected = team
		}
	}
	return (*nox_team_t)(selected.C())
}

//export sub_418A10
func sub_418A10() *nox_team_t {
	world := GetServer().S()
	for team := world.Teams.First(); team != nil; team = world.Teams.Next(team) {
		if team.Ind60() == 0 {
			return (*nox_team_t)(team.C())
		}
	}
	return (*nox_team_t)(world.Teams.Create(0).C())
}

//export sub_418A40
func sub_418A40(name *wchar2_t) *nox_team_t {
	if name == nil {
		return nil
	}
	teams := &GetServer().S().Teams
	for team := teams.First(); team != nil; team = teams.Next(team) {
		record := (*nox_team_t)(team.C())
		if C._nox_wcsicmp(&record.name[0], name) == 0 {
			return record
		}
	}
	return nil
}

//export sub_418A80
func sub_418A80(color C.int) *nox_team_t {
	world := GetServer().S()
	for team := world.Teams.First(); team != nil; team = world.Teams.Next(team) {
		if C.int(team.ColorInd) == color {
			return (*nox_team_t)(team.C())
		}
	}
	return nil
}

//export sub_4190F0
func sub_4190F0(name *wchar2_t) C.int {
	return C.int(bool2int(sub_418A40(name) != nil))
}

func teamObjectByNetCode418C80(code uint32) *server.ObjectTeam {
	if noxflags.HasGame(noxflags.GameHost) {
		if object := GetServer().S().ObjectFromNetCode4ECCB0(code); object != nil {
			return &object.TeamVal
		}
	} else if drawable := GetClient().Cli().Objs.ByNetCodeDynamic(int(int32(code))); drawable != nil {
		return &drawable.TeamVal
	}
	return nil
}

func countTeamPlayers418BC0(world *server.Server, team *nox_team_t) int {
	count := 0
	for player := world.Players.First(); player != nil; player = world.Players.Next(player) {
		member := teamObjectByNetCode418C80(uint32(player.NetCode()))
		if world.Teams.ContainsObject(member, server.TeamID(team.field_57)) {
			count++
		}
	}
	return count
}

//export sub_418BC0
func sub_418BC0(team *nox_team_t) C.int {
	if team == nil {
		return 0
	}
	return C.int(countTeamPlayers418BC0(GetServer().S(), team))
}

//export nox_xxx_objGetTeamByNetCode_418C80
func nox_xxx_objGetTeamByNetCode_418C80(code C.int) *C.nox_object_team_t {
	return (*C.nox_object_team_t)(unsafe.Pointer(teamObjectByNetCode418C80(uint32(code))))
}

func teamDescriptorHeader(team *nox_team_t, coopTeam func() bool) (header [18]byte, size int) {
	// GAME.EXE reads these fields before checking the mode and scanning the name.
	binary.LittleEndian.PutUint32(header[6:10], uint32(team.field_60))
	binary.LittleEndian.PutUint32(header[2:6], uint32(team.field_57))
	header[0] = byte(netmsg.MSG_TEAM_MSG)
	binary.LittleEndian.PutUint32(header[10:14], uint32(team.lessons))
	header[16], header[17] = byte(team.def_ind), byte(team.field_68)
	if coopTeam() {
		header[14] = 1
	}
	header[15] = byte(alloc.StrLen((*uint16)(unsafe.Pointer(team))))
	return header, len(header) + 2*int(header[15])
}

func sendTeamDescriptor4197C0(team *nox_team_t, recipient C.int) {
	header, size := teamDescriptorHeader(team, func() bool {
		return noxflags.HasGame(noxflags.GameModeCoopTeam)
	})
	var packet [18 + 2*255]byte
	copy(packet[:], header[:])
	copy(packet[len(header):size], unsafe.Slice((*byte)(unsafe.Pointer(team)), size-len(header)))
	sendImportantPacketBytes(recipient, packet[:size], true, true)
}

//export sub_4197C0
func sub_4197C0(name *C.wchar2_t, recipient C.int) {
	if name != nil {
		sendTeamDescriptor4197C0((*nox_team_t)(unsafe.Pointer(name)), recipient)
	}
}

func sendTeamMember4198A0(world *server.Server, member *server.ObjectTeam, recipient C.int, code uint32) {
	if object := world.ObjectFromNetCode4ECCB0(code); object != nil {
		packet := [10]byte{byte(netmsg.MSG_TEAM_MSG), 1}
		binary.LittleEndian.PutUint32(packet[2:6], uint32(member.ID))
		binary.LittleEndian.PutUint16(packet[6:8], uint16(code))
		binary.LittleEndian.PutUint16(packet[8:10], object.TypeInd)
		sendImportantPacketBytes(recipient, packet[:], true, true)
	}
}

//export sub_4198A0
func sub_4198A0(member *C.nox_object_team_t, recipient, code C.int) {
	if member != nil {
		sendTeamMember4198A0(GetServer().S(), (*server.ObjectTeam)(unsafe.Pointer(member)), recipient, uint32(code))
	}
}

//export sub_4DF2E0
func sub_4DF2E0(recipient C.int) {
	if recipient == server.HostPlayerIndex {
		return
	}
	world := GetServer().S()
	for team := world.Teams.First(); team != nil; team = world.Teams.Next(team) {
		sendTeamDescriptor4197C0((*nox_team_t)(team.C()), recipient)
		for unit := world.Players.FirstUnit(); unit != nil; unit = world.NextPlayerUnit4DA7F0(unit) {
			if world.Teams.ContainsObject(&unit.TeamVal, team.ID()) {
				sendTeamMember4198A0(world, &unit.TeamVal, recipient, unit.NetCode)
			}
		}
	}
}
