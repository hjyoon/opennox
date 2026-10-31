package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

func animationSelectorPublicC533790() unsafe.Pointer { return C.nox_xxx_mobActionToAnimation_533790 }

func animationPredicatePublicC533790(index int, unit *server.Object) uint32 {
	var fn unsafe.Pointer
	switch index {
	case 0:
		fn = C.nox_xxx_unitIsMimic_534840
	case 1:
		fn = C.nox_xxx_unitIsPlant_534A10
	case 2:
		fn = C.nox_xxx_unitIsZombie_534A40
	default:
		panic("invalid actual animation predicate")
	}
	return uint32(ccall.CallIntPtr(fn, unit.CObj()))
}
