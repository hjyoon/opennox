package legacy

/*
#include "defs.h"
#include "GAME5.h"
_Static_assert(_Generic(&nox_xxx_monsterCanResumeAttack_545520, int (*)(nox_object_t*):1, default:0), "native resume predicate ABI");
_Static_assert(_Generic(&sub_545580, int (*)(nox_object_t*):1, default:0), "native retreat predicate ABI");
_Static_assert(_Generic(&nox_xxx_monsterCanCast2_5455B0, int (*)(nox_object_t*):1, default:0), "native cast predicate ABI");
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_monsterCanResumeAttack_545520
func nox_xxx_monsterCanResumeAttack_545520(unit *C.nox_object_t) C.int {
	if (*server.Object)(unsafe.Pointer(unit)).MonsterCanResumeAttack545520() {
		return 1
	}
	return 0
}

//export sub_545580
func sub_545580(unit *C.nox_object_t) C.int {
	if (*server.Object)(unsafe.Pointer(unit)).MonsterRetreatActive545580() {
		return 1
	}
	return 0
}

//export nox_xxx_monsterCanCast2_5455B0
func nox_xxx_monsterCanCast2_5455B0(unit *C.nox_object_t) C.int {
	if (*server.Object)(unsafe.Pointer(unit)).MonsterCanCast2_5455B0() {
		return 1
	}
	return 0
}
