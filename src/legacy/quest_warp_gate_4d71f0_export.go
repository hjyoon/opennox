package legacy

/*
#include "defs.h"
#include "GAME3_2.h"
#include "GAME3_3.h"

extern uint32_t dword_5d4594_1556136;
*/
import "C"

import (
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

const (
	questStateBase4D71F0       = uintptr(0x5D4594)
	questWarpFrameOffset4D7600 = uintptr(1556108)
	questWarpOnOffset4D7520    = uintptr(1556120)
)

func questWarpGateRuntime4D71F0() server.QuestWarpGateRuntime4D71F0 {
	s := GetServer().S()
	return server.QuestWarpGateRuntime4D71F0{
		ExitCountdownStart: func() uint32 {
			return uint32(C.dword_5d4594_1556136)
		},
		StoreExitCountdownStart: func(value uint32) {
			C.sub_4D71E0(C.int(int32(value)))
		},
		WarpEnabled: func() uint32 {
			return memmap.Uint32(questStateBase4D71F0, questWarpOnOffset4D7520)
		},
		StoreWarpEnabled: func(value uint32) {
			*memmap.PtrUint32(questStateBase4D71F0, questWarpOnOffset4D7520) = value
		},
		WarpFrame: func() uint32 {
			return memmap.Uint32(questStateBase4D71F0, questWarpFrameOffset4D7600)
		},
		LeaveObserver: Nox_xxx_playerLeaveObserver_0_4E6AA0,
		CameraUnlock:  Nox_xxx_playerCameraUnlock_4E6040,
		Move:          Nox_xxx_unitMove_4E7010,
		Audio: func(id sound.ID, obj *server.Object, kind int, code uint32) {
			s.Audio.EventObj(id, obj, kind, code)
		},
		PointFX:      s.Nox_xxx_netSendPointFx_522FF0,
		ObjectSetOff: objectSetOffRuntime4E7600,
		MaybeWarp: func() int32 {
			return s.QuestMaybeWarp4E8F60(server.QuestMaybeWarpRuntime4E8F60{
				CurrentQuestStage: func() uint32 {
					return uint32(C.nox_game_getQuestStage_4E3CC0())
				},
				NextStageThreshold: func(stage uint32) uint32 {
					return uint32(Nox_server_questNextStageThreshold_4D74F0(int32(stage)))
				},
			})
		},
		PriMessage: func(obj *server.Object, id strman.ID, value byte) {
			s.NetPriMsgToPlayer(obj, id, value)
		},
	}
}

func Sub_4D71F0() {
	GetServer().S().QuestExitTimeout4D71F0(questWarpGateRuntime4D71F0())
}

func Sub_4D7480(obj *server.Object) {
	GetServer().S().QuestLeaveWarpGate4D7480(obj, questWarpGateRuntime4D71F0())
}

func Sub_4D7520(enabled int) {
	GetServer().S().QuestSetWarpEnabled4D7520(int32(enabled), questWarpGateRuntime4D71F0())
}

func Nox_server_checkWarpGate_4D7600() {
	GetServer().S().QuestCheckWarpGate4D7600(questWarpGateRuntime4D71F0())
}

//export nox_xxx_questExitTimeout_native_4D71F0
func nox_xxx_questExitTimeout_native_4D71F0() C.uint {
	return C.uint(GetServer().S().QuestExitTimeout4D71F0(questWarpGateRuntime4D71F0()))
}

//export nox_xxx_questLeaveWarpGate_native_4D7480
func nox_xxx_questLeaveWarpGate_native_4D7480(obj *nox_object_t) {
	Sub_4D7480(asObjectS(obj))
}

//export nox_xxx_questWarpEnabled_native_4D7520
func nox_xxx_questWarpEnabled_native_4D7520(enabled C.int) C.uchar {
	return C.uchar(GetServer().S().QuestSetWarpEnabled4D7520(
		int32(enabled),
		questWarpGateRuntime4D71F0(),
	))
}

//export nox_xxx_questCheckWarpGate_native_4D7600
func nox_xxx_questCheckWarpGate_native_4D7600() {
	Nox_server_checkWarpGate_4D7600()
}
