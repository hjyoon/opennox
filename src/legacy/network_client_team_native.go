package legacy

/*
#include "common__system__team.h"
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

var clientTeamPacketSizes = [...]uint8{
	0: 18, 1: 10, 2: 6, 3: 10, 4: 46, 5: 6, 6: 6, 7: 2, 8: 10, 9: 2, 12: 5,
}

func clientTeamByPacketID(data []byte) *nox_team_t {
	return nox_xxx_getTeamByID_418AB0(int32(binary.LittleEndian.Uint32(data[2:6])))
}

func clientTeamCreatePacket(data []byte) {
	var name [256]wchar2_t
	count := int(data[15])
	for i := range count {
		name[i] = wchar2_t(binary.LittleEndian.Uint16(data[18+2*i : 20+2*i]))
	}
	color := data[16]
	if data[17] != 0 {
		title := nox_server_teamTitle_418C20(int32(color))
		formatWideStrings(name[:], title, nil, nil)
	}
	team := clientTeamByPacketID(data)
	if team == nil {
		team = nox_xxx_teamCreate_4186D0(C.char(data[2]))
		if team == nil {
			return
		}
	}
	sub_418800(&team.name[0], &name[0], 0)
	sub_418830(team, C.int(binary.LittleEndian.Uint32(data[6:10])))
	nox_xxx_netChangeTeamID_419090(team, C.int(binary.LittleEndian.Uint32(data[10:14])))
	team.def_ind = C.uint8_t(color)
	sub_457230(&name[0])
	if data[14]&1 == 0 {
		return
	}
	code := C.int(ClientPlayerNetCode())
	member := nox_xxx_objGetTeamByNetCode_418C80(code)
	if member == nil {
		return
	}
	if noxflags.HasGame(noxflags.GameHost) {
		nox_xxx_createAtImpl_4191D0(team.field_57, member, 1, code, 1)
	} else {
		sub_419900(team, member, C.short(code))
	}
}

func clientTeamAddMemberPacket(data []byte) {
	rawCode := binary.LittleEndian.Uint16(data[6:8])
	drawable := clientPacketDrawable(rawCode)
	if drawable == nil {
		id := binary.LittleEndian.Uint16(data[8:10])
		drawable = nox_xxx_spriteCreate_48E970(int32(id), rawCode&0x7fff, 0, 0)
	}
	if drawable == nil {
		return
	}
	if team := clientTeamByPacketID(data); team != nil {
		member := (*C.nox_object_team_t)(unsafe.Pointer(&asDrawable(drawable).TeamVal))
		nox_xxx_createAtImpl_4191D0(team.field_57, member, 0, C.int(rawCode), 0)
		sub_4571A0(C.int(rawCode), C.int(team.field_57))
	}
}

func clientTeamUpdatePacket(subtype byte, data []byte) {
	team := clientTeamByPacketID(data)
	if team == nil {
		return
	}
	switch subtype {
	case 4:
		var name [21]wchar2_t
		for i := 0; i < len(name)-1; i++ {
			name[i] = wchar2_t(binary.LittleEndian.Uint16(data[6+2*i : 8+2*i]))
		}
		nox_xxx_teamRenameMB_418CD0(team, &name[0])
	case 5:
		sub_418D80(team)
	case 6:
		name := team.name
		sub_418F20(team, 0)
		sub_456EA0(&name[0])
	case 8:
		nox_xxx_netChangeTeamID_419090(team, C.int(binary.LittleEndian.Uint32(data[6:10])))
	}
}

func clientTeamPacket(data []byte) int {
	if len(data) < 2 || int(data[1]) >= len(clientTeamPacketSizes) {
		return -1
	}
	subtype := data[1]
	size := int(clientTeamPacketSizes[subtype])
	if size == 0 || len(data) < size {
		return -1
	}
	if subtype == 0 {
		size += 2 * int(data[15])
		if len(data) < size {
			return -1
		}
	}
	if nox_client_isConnected_43C700() == 0 {
		return size
	}
	switch subtype {
	case 0:
		clientTeamCreatePacket(data)
	case 1:
		clientTeamAddMemberPacket(data)
	case 2:
		code := C.int(binary.LittleEndian.Uint32(data[2:6]))
		if member := nox_xxx_objGetTeamByNetCode_418C80(code); member != nil {
			nox_xxx_netChangeTeamMb_419570(member, code)
			sub_4571A0(code, 0)
		}
	case 3:
		code := C.int(binary.LittleEndian.Uint16(data[6:8]))
		if member := nox_xxx_objGetTeamByNetCode_418C80(code); member != nil {
			team := clientTeamByPacketID(data)
			if team != nil && sub_4196D0(unsafe.Pointer(member), unsafe.Pointer(team), code, 0) != 0 {
				sub_4571A0(code, C.int(team.field_57))
			}
		}
	case 4, 5, 6, 8:
		clientTeamUpdatePacket(subtype, data)
	case 7:
		nox_server_teamsZzz_419030(0)
		sub_456FA0()
	case 9:
		nox_server_teamsResetYyy_417D00()
	case 12:
		code := binary.LittleEndian.Uint16(data[2:4])
		if player := nox_common_playerInfoGetByID_417040(int32(code)); player != nil {
			player.field_2282 = player.field_2282&0xff00 | C.uint16_t(data[4])
		}
	}
	return size
}
