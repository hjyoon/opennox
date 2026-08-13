package legacy

/*
#include "GAME3_2.h"
#include "GAME3_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

func unitFreezeGateRuntimeRef4E79B0() *uint32 {
	return (*uint32)(unsafe.Pointer(C.nox_xxx_unitFreezeGateRef_4E79B0()))
}

func unitFreezeRuntime4E79C0(obj *server.Object, source uint32) byte {
	return unitFreezeNative4E79C0(obj, source, unitFreezeGateRuntimeRef4E79B0(), unitFreezeRuntimeDeps4E79C0())
}

func unitUnfreezeRuntime4E7A60(obj *server.Object, force uint32) byte {
	return unitUnfreezeNative4E7A60(obj, force, unitFreezeGateRuntimeRef4E79B0(), unitFreezeRuntimeDeps4E79C0())
}

func netReportPlayerStatusNative4D8270(obj *server.Object) int {
	return netReportPlayerStatusNativeWithSend4D8270(obj, func(playerInd byte, packet []byte) int {
		return GetServer().S().NetSendPacketXxx0(int(playerInd), packet, nil, 1)
	})
}

func unitFreezeRuntimeDeps4E79C0() unitFreezeNativeDeps4E79C0 {
	return unitFreezeNativeDeps4E79C0{
		reportStatus: func(obj *server.Object) byte {
			return byte(netReportPlayerStatusNative4D8270(obj))
		},
		setPlayerIdle: func(obj *server.Object) {
			Nox_xxx_playerSetState_4FA020(obj, server.PlayerState13)
		},
		raiseZero: func(obj *server.Object) {
			Nox_xxx_unitRaise_4E46F0(obj, 0)
		},
		resetPaths: func() {
			GetServer().S().AI.Paths.Sub_50B510()
		},
		pushIdle: func(obj *server.Object) byte {
			item := obj.MonsterPushAction(ai.ACTION_IDLE)
			return byte(uintptr(unsafe.Pointer(item)))
		},
		popAction: func(obj *server.Object) byte {
			return byte(obj.MonsterPopAction())
		},
	}
}

//export nox_xxx_netReportPlrStatus_4D8270
func nox_xxx_netReportPlrStatus_4D8270(obj *nox_object_t) C.int {
	return C.int(netReportPlayerStatusNative4D8270(asObjectS(obj)))
}

//export nox_xxx_unitFreeze_4E79C0
func nox_xxx_unitFreeze_4E79C0(obj *nox_object_t, source C.uint32_t) C.char {
	return C.char(unitFreezeRuntime4E79C0(asObjectS(obj), uint32(source)))
}

//export nox_xxx_unitUnFreeze_4E7A60
func nox_xxx_unitUnFreeze_4E7A60(obj *nox_object_t, force C.uint32_t) C.char {
	return C.char(unitUnfreezeRuntime4E7A60(asObjectS(obj), uint32(force)))
}
