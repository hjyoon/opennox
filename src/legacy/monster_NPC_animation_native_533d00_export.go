package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
_Static_assert(_Generic(&nox_xxx_unitNPCActionToAnim_533D00, unsigned char* (*)(nox_object_t*):1,default:0), "native NPC animation pointer ABI");
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_unitNPCActionToAnim_533D00
func nox_xxx_unitNPCActionToAnim_533D00(unit *C.nox_object_t) *C.uchar {
	return (*C.uchar)(unsafe.Pointer(server.MonsterNPCAnimationNative533D00(asObjectS((*nox_object_t)(unit)), monsterNPCAnimationRuntimeNative533D00())))
}
