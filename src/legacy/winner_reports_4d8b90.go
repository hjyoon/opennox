package legacy

/*
#include "GAME3_2.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_netSendDMWinner_4D8B90
func nox_xxx_netSendDMWinner_4D8B90(raw *nox_object_t, arg C.char) C.intptr_t {
	unit := asObjectS(raw)
	if unit != nil && !unit.Class().Has(object.ClassPlayer) {
		// The original returns the input pointer for non-players.
		return C.intptr_t(uintptr(unsafe.Pointer(raw)))
	}
	code := uint16(server.ObjectNetCode(unit))
	frame := gameFrame()
	return C.intptr_t(GetServer().S().SendMatchWinner(netmsg.MSG_REPORT_DEATHMATCH_WINNER, code, uint8(arg), frame))
}

func sendTeamWinnerReport(op netmsg.Op, team *nox_team_t, arg uint8, code uint16) C.int {
	if team != nil {
		code = uint16(asTeam(team).ID())
	}
	frame := gameFrame()
	return C.int(GetServer().S().SendMatchWinner(op, code, arg, frame))
}

//export nox_xxx_netSendDMTeamWinner_4D8BF0
func nox_xxx_netSendDMTeamWinner_4D8BF0(team *nox_team_t, arg C.char) C.int {
	return sendTeamWinnerReport(netmsg.MSG_REPORT_DEATHMATCH_TEAM_WINNER, team, uint8(arg), 0)
}

//export nox_xxx_netFlagballWinner_4D8C40
func nox_xxx_netFlagballWinner_4D8C40(team *nox_team_t) C.int {
	return sendTeamWinnerReport(netmsg.MSG_REPORT_FLAG_BALL_WINNER, team, 0, 0)
}

//export nox_xxx_netFlagWinner_4D8C40_4D8C80
func nox_xxx_netFlagWinner_4D8C40_4D8C80(team *nox_team_t, arg C.char) C.int {
	return sendTeamWinnerReport(netmsg.MSG_REPORT_FLAG_WINNER, team, uint8(arg), ^uint16(0))
}
