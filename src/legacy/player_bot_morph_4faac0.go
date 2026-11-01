package legacy

/*
#include <stdint.h>
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_mobMorphFromPlayer_4FAAC0
func nox_xxx_mobMorphFromPlayer_4FAAC0(unit *C.uint32_t) C.char {
	return C.char((*server.Object)(unsafe.Pointer(unit)).MorphPlayerBotFromPlayer4FAAC0())
}

//export nox_xxx_mobMorphToPlayer_4FAAF0
func nox_xxx_mobMorphToPlayer_4FAAF0(unit *C.uint32_t) C.char {
	return C.char((*server.Object)(unsafe.Pointer(unit)).MorphPlayerBotToPlayer4FAAF0())
}
