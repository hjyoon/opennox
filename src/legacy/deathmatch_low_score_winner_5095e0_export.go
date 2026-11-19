package legacy

/*
#include <stdint.h>
#include "GAME3_2.h"
#include "deathmatch_low_score_winner_5095e0.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func deathmatchWinnerRuntime509x() server.DeathmatchLowScoreWinnerRuntime5095E0 {
	return server.DeathmatchLowScoreWinnerRuntime5095E0{
		SendTeamWinner: func(team *server.Team, flag uint8) int32 {
			return int32(C.nox_xxx_netSendDMTeamWinner_4D8BF0((*nox_team_t)(team.C()), C.char(flag)))
		},
		SendPlayerWinner: func(unit *server.Object, flag uint8) int32 {
			return int32(C.nox_xxx_netSendDMWinner_4D8B90(asObjectC(unit), C.char(flag)))
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
