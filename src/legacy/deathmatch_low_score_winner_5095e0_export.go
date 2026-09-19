package legacy

/*
#include <stdint.h>
#include "GAME3_2.h"
#include "deathmatch_low_score_winner_5095e0.h"
*/
import "C"

import (
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func deathmatchWinnerRuntime509x() server.DeathmatchLowScoreWinnerRuntime5095E0 {
	return server.DeathmatchLowScoreWinnerRuntime5095E0{
		SendTeamWinner: func(team *server.Team, flag uint8) int32 {
			var address uintptr
			if team != nil {
				address = uintptr(team.C())
			}
			result := int32(C.nox_xxx_netSendDMTeamWinner_4D8BF0(
				C.intptr_t(address),
				C.char(flag),
			))
			runtime.KeepAlive(team)
			return result
		},
		SendPlayerWinner: func(unit *server.Object, flag uint8) int32 {
			address := uintptr(unsafe.Pointer(asObjectC(unit)))
			result := int32(C.nox_xxx_netSendDMWinner_4D8B90(
				C.intptr_t(address),
				C.char(flag),
			))
			runtime.KeepAlive(unit)
			return result
		},
	}
}

var deathmatchLowScoreWinnerCall5095E0 = func() int32 {
	return GetServer().S().DeathmatchLowScoreWinner5095E0(deathmatchWinnerRuntime509x())
}

func deathmatchLowScoreWinnerExportCall5095E0() int32 {
	return int32(C.sub_5095E0())
}

//export sub_5095E0
func sub_5095E0() C.int32_t {
	return C.int32_t(deathmatchLowScoreWinnerCall5095E0())
}
