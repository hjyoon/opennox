package legacy

/*
#include "GAME5_2.h"
#include "cgo_const.h"
*/
import "C"

import (
	"github.com/opennox/libs/ifs"
	noxflags "github.com/opennox/opennox/v1/common/flags"
)

var mapRuleModes = [...]struct {
	mode  noxflags.GameFlag
	token string
}{
	{noxflags.GameModeElimination, "[ELIMINATION]"},
	{noxflags.GameModeArena, "[DEATHMATCH]"},
	{noxflags.GameModeCTF, "[CAPTURE_THE_FLAG]"},
	{noxflags.GameModeKOTR, "[KING_OF_THE_REALM]"},
	{noxflags.GameModeFlagBall, "[FLAGBALL]"},
	{noxflags.GameModeMask, "[COMMON]"},
	{noxflags.GameModeQuest, "[QUEST]"},
}

func mapRuleModeName57A1B0(mode noxflags.GameFlag) string {
	mode &= noxflags.GameModeMask
	for _, row := range mapRuleModes {
		if row.mode == mode {
			return row.token
		}
	}
	return ""
}

//export sub_57A1B0
func sub_57A1B0(mode C.short) *C.char {
	name := mapRuleModeName57A1B0(noxflags.GameFlag(uint16(mode)))
	if name == "" {
		return nil
	}
	return internCStr(name)
}

func mapRuleModeByName(token string) noxflags.GameFlag {
	for _, row := range mapRuleModes {
		if token == row.token {
			return row.mode
		}
	}
	return 0
}

//export sub_57AE30
func sub_57AE30(name *C.nox_cgo_const_char) C.int {
	return C.int(mapRuleModeByName(GoString((*C.char)(name))))
}

//export sub_57A9F0
func sub_57A9F0(mapName, filename *C.nox_cgo_const_char) C.int {
	path := "maps\\" + GoString((*C.char)(mapName)) + "\\" + GoString((*C.char)(filename))
	// The OpenNox filesystem API reports successful removal as true.
	return C.int(bool2int(ifs.Remove(path) == nil))
}
