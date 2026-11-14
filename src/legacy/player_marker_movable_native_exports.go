package legacy

/*
#include "GAME1_1.h"
#include "GAME4_2.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export sub_422140
func sub_422140(player *C.nox_playerInfo) *C.nox_playerInfo {
	(*server.Player)(unsafe.Pointer(player)).Sub422140()
	return player
}

//export nox_xxx_isObjectMovable_52E020
func nox_xxx_isObjectMovable_52E020(obj *C.nox_object_t) C.uint {
	if (*server.Object)(unsafe.Pointer(obj)).IsMovable() {
		return 1
	}
	return 0
}
