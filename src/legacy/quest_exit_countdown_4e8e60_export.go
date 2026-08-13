package legacy

/*
#include "GAME1.h"
#include "GAME3_3.h"
*/
import "C"

import (
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/server"
)

//export sub_4E8E60
func sub_4E8E60() int32 {
	return GetServer().S().QuestExitCountdown4E8E60(server.QuestExitCountdownRuntime4E8E60{
		TimerActive: func() int32 {
			return int32(C.sub_40A220())
		},
		TimerRemainingMillis: func() int32 {
			return int32(C.sub_40A230())
		},
		StopTimer: func(value int32) int32 {
			return int32(C.sub_40A1F0(C.int(value)))
		},
		CountdownStarted: func() int32 {
			if GetServer().GetFlag3592() {
				return 1
			}
			return 0
		},
		StartCountdown: func(seconds int32, id string) {
			GetServer().ServStartCountdown(int(seconds), strman.ID(id))
		},
	})
}
