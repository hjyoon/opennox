package legacy

/*
#include "GAME3_2.h"
#include "GAME4_1.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_tradeSetPlayer_50F370
func nox_xxx_tradeSetPlayer_50F370(session *C.uint32_t, owner *C.nox_object_t) *C.uint32_t {
	return (*C.uint32_t)(unsafe.Pointer(pointerChainAttach50F370((*server.TradeSession)(unsafe.Pointer(session)), asObjectS((*nox_object_t)(owner)))))
}

//export sub_50F490
func sub_50F490(session *C.uint32_t, owner *C.nox_object_t) unsafe.Pointer {
	return pointerChainDetach50F490((*server.TradeSession)(unsafe.Pointer(session)), asObjectS((*nox_object_t)(owner)))
}

//export sub_4D60E0
func sub_4D60E0(owner *C.nox_object_t) *C.uint32_t {
	return (*C.uint32_t)(pointerChainStat4D60E0(asObjectS((*nox_object_t)(owner))))
}

//export sub_4D6170
func sub_4D6170(owner *C.nox_object_t) unsafe.Pointer {
	return pointerChainKill4D6170(asObjectS((*nox_object_t)(owner)))
}

//export sub_510DE0
func sub_510DE0(owner *C.nox_object_t, code C.int) *C.nox_object_t {
	return (*C.nox_object_t)(unsafe.Pointer(pointerChainLookup510DE0(asObjectS((*nox_object_t)(owner)), uint32(code))))
}

//export nox_xxx_changeScore_4D8E90
func nox_xxx_changeScore_4D8E90(owner *C.nox_object_t, delta C.int) unsafe.Pointer {
	return pointerChainLessons(asObjectS((*nox_object_t)(owner)), int32(delta), false)
}

//export nox_xxx_playerSubLessons_4D8EC0
func nox_xxx_playerSubLessons_4D8EC0(owner *C.nox_object_t, delta C.int) unsafe.Pointer {
	return pointerChainLessons(asObjectS((*nox_object_t)(owner)), int32(delta), true)
}
