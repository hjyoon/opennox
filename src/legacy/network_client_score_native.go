package legacy

/*
#include "client__network__cdecode.h"

extern uint32_t dword_5d4594_1200804;
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

func clientScorePlayer(data []byte) *nox_playerInfo {
	code := binary.LittleEndian.Uint16(data[1:3]) & 0x7fff
	return nox_common_playerInfoGetByID_417040(int32(code))
}

func clientScoreLesson(data []byte) {
	settings := (*server.Settings2)(cliGamedataNative416590(0))
	if nox_client_isConnected_43C700() == 0 {
		return
	}
	player := clientScorePlayer(data)
	if player == nil {
		return
	}
	deaths := binary.LittleEndian.Uint32(data[7:11])
	if !noxflags.HasGame(noxflags.GameHost) {
		player.lessons = C.int(int32(binary.LittleEndian.Uint32(data[3:7])))
		player.field_2140 = C.uint32_t(deaths)
		player.field_2144 = C.uint32_t(gameFrame())
	}
	if noxflags.HasGame(noxflags.GameModeElimination) && int32(deaths) >= int32(settings.Field54) {
		var text [256]wchar2_t
		clientAppendScoreFormat(text[:], "Eliminated", 3452, &player.name_final[0])
		nox_xxx_clientPlaySoundSpecial_452D80(312, 100)
		nox_xxx_printCentered_445490(&text[0])
	}
}

func clientScoreTreasure(data []byte) {
	if nox_client_isConnected_43C700() == 0 {
		return
	}
	player := clientScorePlayer(data)
	if player == nil {
		return
	}
	if !noxflags.HasGame(noxflags.GameHost) && gameFrame() > uint32(player.field_2160) {
		player.field_2152 = C.uint32_t(binary.LittleEndian.Uint16(data[3:5]))
		player.field_2156 = C.uint32_t(binary.LittleEndian.Uint16(data[5:7]))
		player.field_2160 = C.uint32_t(gameFrame())
	}
	if player.field_2152 == player.field_2156-1 {
		var text [256]wchar2_t
		clientAppendScoreFormat(text[:], "SH_NearVictory", 3729, &player.name_final[0])
		nox_xxx_printCentered_445490(&text[0])
	}
}

func clientAppendScoreFormat(out []wchar2_t, key string, line int32, arg *wchar2_t) {
	end := int(nox_wcslen(&out[0]))
	formatWideStrings(out[end:], clientDecodeString(key, line), arg, nil)
}

func clientAppendScoreText(out []wchar2_t, text *wchar2_t) {
	end := int(nox_wcslen(&out[0]))
	length := int(nox_wcslen(text))
	n := copy(out[end:len(out)-1], unsafe.Slice(text, length))
	out[end+n] = 0
}

func clientMatchTimeLimit(op netmsg.Op, reason byte, text []wchar2_t) {
	if reason != 1 {
		return
	}
	switch op {
	case netmsg.MSG_REPORT_FLAG_WINNER:
		clientAppendScoreFormat(text, "TimeLimitReached", 4055, nil)
	case netmsg.MSG_REPORT_DEATHMATCH_WINNER:
		clientAppendScoreFormat(text, "TimeLimitReached", 3823, nil)
	case netmsg.MSG_REPORT_DEATHMATCH_TEAM_WINNER:
		clientAppendScoreFormat(text, "TimeLimitReached", 3945, nil)
	}
}

func clientMatchPlayerText(text []wchar2_t, rawCode uint16, reason byte) C.int {
	code := int32(rawCode & 0x7fff)
	player := nox_common_playerInfoGetByID_417040(code)
	self := int(code) == ClientPlayerNetCode()
	if noxflags.HasGame(noxflags.GameModeElimination) {
		if rawCode == 0 {
			clientAppendScoreFormat(text, "HL_Tie", 3847, nil)
			return 0
		}
		clientAppendScoreFormat(text, "HL_Header", 3852, nil)
		if self {
			you := clientDecodeString("HL_You", 3855)
			clientAppendScoreFormat(text, "HL_Victory", 3855, you)
			return 0
		}
		if player != nil {
			clientAppendScoreFormat(text, "HL_Victory", 3863, &player.name_final[0])
			if reason == 0 {
				sub_4947E0(player)
			}
		}
		return 1
	}
	if rawCode == 0 {
		clientAppendScoreFormat(text, "DM_Tie", 3884, nil)
		return 0
	}
	if player == nil {
		return 1
	}
	if reason == 0 {
		sub_4947E0(player)
	}
	if !self {
		clientAppendScoreFormat(text, "DM_Loss", 3899, &player.name_final[0])
		return 1
	}
	if !asPlayerS(player).Info().IsFemale() {
		clientAppendScoreFormat(text, "DM_MaleVictory", 3905, nil)
	} else {
		clientAppendScoreFormat(text, "DM_FemaleVictory", 3907, nil)
	}
	return 0
}

func clientMatchTeamText(op netmsg.Op, text []wchar2_t, team *nox_team_t, rawCode uint16, reason byte) C.int {
	local := nox_xxx_objGetTeamByNetCode_418C80(C.int(ClientPlayerNetCode()))
	var name *wchar2_t
	var member bool
	if team != nil {
		name = &team.name[0]
		member = nox_xxx_teamCompare2_419180(local, C.uchar(team.field_57)) != 0
	}
	switch op {
	case netmsg.MSG_REPORT_FLAG_BALL_WINNER:
		if nox_xxx_teamCompare2_419180(local, C.uchar(rawCode)) != 0 {
			clientAppendScoreFormat(text, "TeamWon", 3776, nil)
			return 0
		}
		var teamText [256]wchar2_t
		clientAppendScoreFormat(teamText[:], "teamformat", 3781, name)
		clientAppendScoreFormat(text, "FB_Victory", 3782, &teamText[0])
		return 1
	case netmsg.MSG_REPORT_FLAG_WINNER:
		if team == nil {
			clientAppendScoreFormat(text, "CTF_Tie", 4083, nil)
			return 0
		}
		clientAppendScoreFormat(text, "CTF_Victory", 4071, name)
		return C.int(bool2int(!member))
	}

	elimination := noxflags.HasGame(noxflags.GameModeElimination)
	if team == nil {
		if elimination {
			clientAppendScoreFormat(text, "HL_Tie", 3964, nil)
		} else {
			clientAppendScoreFormat(text, "DM_Tie", 3999, nil)
		}
		return 0
	}
	if elimination {
		clientAppendScoreFormat(text, "HL_Header", 3969, nil)
		if member {
			yourTeam := clientDecodeString("HL_YourTeam", 3973)
			clientAppendScoreFormat(text, "HL_Victory", 3973, yourTeam)
		} else {
			var format, named [256]wchar2_t
			teamLabel := clientDecodeString("Team", 3980)
			clientAppendScoreFormat(format[:], "HL_Victory", 3980, teamLabel)
			formatWideStrings(named[:], &format[0], name, nil)
			clientAppendScoreText(text, &named[0])
		}
	} else if member {
		clientAppendScoreFormat(text, "DM_TeamVictory", 4006, nil)
	} else {
		var teamText [256]wchar2_t
		clientAppendScoreFormat(teamText[:], "Team", 4011, name)
		clientAppendScoreFormat(text, "DM_Loss", 4012, &teamText[0])
	}
	if reason == 0 {
		sub_4948B0(team)
	}
	return C.int(bool2int(!member))
}

func clientMatchWinnerPacket(op netmsg.Op, data []byte) {
	code := binary.LittleEndian.Uint16(data[1:3])
	var team *nox_team_t
	if op != netmsg.MSG_REPORT_DEATHMATCH_WINNER {
		team = nox_xxx_getTeamByID_418AB0(int32(code))
	}
	if nox_client_isConnected_43C700() == 0 {
		return
	}
	frame := binary.LittleEndian.Uint32(data[4:8])
	last := uint32(C.dword_5d4594_1200804)
	signedFrame := op == netmsg.MSG_REPORT_FLAG_BALL_WINNER || op == netmsg.MSG_REPORT_DEATHMATCH_WINNER
	if signedFrame && int32(frame) <= int32(last) || !signedFrame && frame <= last {
		return
	}

	var text [256]wchar2_t
	reason := data[3]
	clientMatchTimeLimit(op, reason, text[:])
	nox_xxx_clientPlaySoundSpecial_452D80(309, 100)
	if !noxflags.HasGame(noxflags.GameHost) {
		noxflags.SetGame(noxflags.GameFlag4)
	}
	if op != netmsg.MSG_REPORT_DEATHMATCH_WINNER {
		nox_xxx_initTime_435570()
	}
	if !noxflags.HasEngine(noxflags.EngineNoRendering) {
		sub_470510()
	}
	var style C.int
	if op == netmsg.MSG_REPORT_DEATHMATCH_WINNER {
		nox_xxx_initTime_435570()
		style = clientMatchPlayerText(text[:], code, reason)
	} else {
		style = clientMatchTeamText(op, text[:], team, code, reason)
	}
	sub_435700(&text[0], style)
	nox_xxx_guiServerOptionsHide_4597E0(0)
}

func clientScorePacket(op netmsg.Op, data []byte) int {
	var size int
	switch op {
	case netmsg.MSG_REPORT_LESSON:
		size = 11
	case netmsg.MSG_REPORT_TREASURE_COUNT:
		size = 7
	case netmsg.MSG_REPORT_FLAG_BALL_WINNER, netmsg.MSG_REPORT_FLAG_WINNER,
		netmsg.MSG_REPORT_DEATHMATCH_WINNER, netmsg.MSG_REPORT_DEATHMATCH_TEAM_WINNER:
		size = 8
	default:
		return 0
	}
	if len(data) < size {
		return -1
	}
	switch op {
	case netmsg.MSG_REPORT_LESSON:
		clientScoreLesson(data)
	case netmsg.MSG_REPORT_TREASURE_COUNT:
		clientScoreTreasure(data)
	default:
		clientMatchWinnerPacket(op, data)
	}
	return size
}

//export nox_client_decodeScorePacket
func nox_client_decodeScorePacket(op C.int, data *C.uchar, size C.int) C.int {
	if data == nil || size < 0 {
		return -1
	}
	return C.int(clientScorePacket(netmsg.Op(op), unsafe.Slice((*byte)(unsafe.Pointer(data)), int(size))))
}
