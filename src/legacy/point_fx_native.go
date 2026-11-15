package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

var pointFXRuntimeForCall522FF0 = pointFXDefaultRuntime522FF0

func pointFXDefaultRuntime522FF0() server.PointFXNativeRuntime522FF0 {
	s := GetServer().S()
	return server.PointFXNativeRuntime522FF0{Send: func(point *types.Pointf, packet []byte) int32 {
		s.Nox_xxx_netSendFxAllCli_523030(*point, packet)
		return 0
	}}
}

//export nox_xxx_netSendPointFx_522FF0
func nox_xxx_netSendPointFx_522FF0(op C.char, point *C.float2) C.int {
	return C.int(server.NetSendPointFXNative522FF0(byte(op), (*types.Pointf)(unsafe.Pointer(point)), pointFXRuntimeForCall522FF0()))
}

//export sub_523150
func sub_523150(op, subtype C.char, point *C.float) C.int {
	return C.int(server.NetSendPointSubtypeFXNative523150(byte(op), byte(subtype), (*types.Pointf)(unsafe.Pointer(point)), pointFXRuntimeForCall522FF0()))
}

//export nox_xxx_netSparkExplosionFx_5231B0
func nox_xxx_netSparkExplosionFx_5231B0(point *C.float, subtype C.char) C.int {
	return C.int(server.NetSparkExplosionFXNative5231B0((*types.Pointf)(unsafe.Pointer(point)), byte(subtype), pointFXRuntimeForCall522FF0()))
}

//export nox_xxx_sendGeneratorBreakFX_523200
func nox_xxx_sendGeneratorBreakFX_523200(point *C.float, subtype C.char) {
	server.SendGeneratorBreakFXNative523200((*types.Pointf)(unsafe.Pointer(point)), byte(subtype), pointFXRuntimeForCall522FF0())
}
