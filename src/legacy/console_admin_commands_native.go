package legacy

/*
#include "client__gui__guicon.h"
*/
import "C"

import (
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

//export nox_cmd_ban
func nox_cmd_ban(i, n C.int, tokens **C.wchar2_t) C.int {
	if n != 2 {
		return 0
	}
	name := consoleLegacyToken(tokens, i, n)
	pl := nox_xxx_playerByName_4170D0(name)
	if pl == nil {
		sub_416770(0, name, nil)
		centerMessageWithName(consoleLegacyString("banDisallow", 3585), name)
		return 1
	}
	if pl.playerInd == server.HostPlayerIndex {
		nox_xxx_printCentered_445490(consoleLegacyString("cantbanyourself", 3580))
		return 1
	}
	if noxflags.HasGame(noxflags.GameModeQuest) {
		sub_4DCFB0(pl.playerUnit)
	} else {
		nox_xxx_playerDisconnByPlrID_4DEB00(int32(pl.playerInd))
	}
	sub_416770(0, name, &pl.serial[0])
	centerMessageWithName(consoleLegacyString("banned", 3577), &pl.name_final[0])
	return 1
}

//export nox_cmd_allow_user
func nox_cmd_allow_user(i, n C.int, tokens **C.wchar2_t) C.int {
	consolePrintRed(consoleLegacyString("notyetimplemented", 3698))
	return 1
}

//export nox_cmd_allow_ip
func nox_cmd_allow_ip(i, n C.int, tokens **C.wchar2_t) C.int {
	consolePrintRed(consoleLegacyString("notyetimplemented", 3704))
	return 1
}

//export nox_cmd_kick
func nox_cmd_kick(i, n C.int, tokens **C.wchar2_t) C.int {
	if n != 2 {
		return 0
	}
	pl := nox_xxx_playerByName_4170D0(consoleLegacyToken(tokens, i, n))
	if pl == nil {
		return 1
	}
	if pl.playerInd == server.HostPlayerIndex {
		nox_xxx_printCentered_445490(consoleLegacyString("cantkickyourself", 3745))
		return 1
	}
	if noxflags.HasGame(noxflags.GameModeQuest) {
		sub_4DCFB0(pl.playerUnit)
	} else {
		nox_xxx_playerCallDisconnect_4DEAB0(int32(pl.playerInd), 4)
		centerMessageWithName(consoleLegacyString("kicked", 3739), &pl.name_final[0])
	}
	return 1
}
