package legacy

/*
#include "defs.h"
#include "GAME5.h"
_Static_assert(_Generic(&sub_545790,void* (*)(nox_object_t*):1,default:0), "native ROAM start object and pointer-return ABI");
_Static_assert(_Generic(&sub_5457C0,void* (*)(nox_object_t*):1,default:0), "native ROAM cancel object and pointer-return ABI");
*/
import "C"

import (
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

//export sub_545790
func sub_545790(unit *C.nox_object_t) unsafe.Pointer {
	return unsafe.Pointer((*server.Object)(unsafe.Pointer(unit)).MonsterActionRoamStartNative545790())
}

//export sub_5457C0
func sub_5457C0(unit *C.nox_object_t) unsafe.Pointer {
	return unsafe.Pointer((*server.Object)(unsafe.Pointer(unit)).MonsterActionRoamCancelNative5457C0())
}
