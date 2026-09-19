package legacy

/*
#include "GAME1.h"
#include "match_state_5096f0.h"
#include "quest_map_buffer_4e8e50.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

var deathmatchHighScoreWinnerCall5098A0 = func() int32 {
	return GetServer().S().DeathmatchHighScoreWinner5098A0(deathmatchWinnerRuntime509x())
}

var teamHighScoreWinnerCall5099B0 = func() int32 {
	return GetServer().S().TeamHighScoreWinner5099B0()
}

var matchLimitStateCall5096F0 = func() int32 {
	outer := GetServer()
	srv := outer.S()
	return srv.MatchLimitState5096F0(server.MatchLimitStateRuntime5096F0{
		LimitExpired: func() int32 {
			return int32(Sub_40A1A0())
		},
		SwitchToNextMap: func() {
			outer.SwitchMap(GoString(C.sub_4E8E50()))
		},
		PrintAutoExit: func() {
			srv.NetPrintLineToAll("chklimit.c:AutoExitToNextMap")
		},
		MatchStateActive: func() int32 {
			if outer.GetFlag3592() {
				return 1
			}
			return 0
		},
		ResolveTeamMode:  teamHighScoreWinnerCall5099B0,
		ResolveHighScore: deathmatchHighScoreWinnerCall5098A0,
		ResolveLowScore:  deathmatchLowScoreWinnerCall5095E0,
		ClearTimer: func() int32 {
			return int32(C.sub_40A1F0(0))
		},
	})
}

func matchLimitStateExportCall5096F0() int32 {
	return int32(C.sub_5096F0())
}

func deathmatchHighScoreWinnerExportCall5098A0() int32 {
	return int32(C.sub_5098A0())
}

func teamHighScoreWinnerExportCall5099B0() int32 {
	return int32(C.sub_5099B0())
}

//export sub_5096F0
func sub_5096F0() C.int32_t {
	return C.int32_t(matchLimitStateCall5096F0())
}

//export sub_5098A0
func sub_5098A0() C.int32_t {
	return C.int32_t(deathmatchHighScoreWinnerCall5098A0())
}

//export sub_5099B0
func sub_5099B0() C.int32_t {
	return C.int32_t(teamHighScoreWinnerCall5099B0())
}
