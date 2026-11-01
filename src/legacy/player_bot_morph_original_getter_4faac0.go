package legacy

/*
#include <stdint.h>
char nox_xxx_mobMorphFromPlayer_4FAAC0(uint32_t*);
char nox_xxx_mobMorphToPlayer_4FAAF0(uint32_t*);
_Static_assert(_Generic(&nox_xxx_mobMorphFromPlayer_4FAAC0,char (*)(uint32_t*):1,default:0),"original char and one-pointer PE32 ABI");
_Static_assert(_Generic(&nox_xxx_mobMorphToPlayer_4FAAF0,char (*)(uint32_t*):1,default:0),"original char and one-pointer PE32 ABI");
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func playerBotMorphOriginalC4FAAC0(unit *server.Object, from bool) byte {
	if from {
		return byte(C.nox_xxx_mobMorphFromPlayer_4FAAC0((*C.uint32_t)(unsafe.Pointer(unit))))
	}
	return byte(C.nox_xxx_mobMorphToPlayer_4FAAF0((*C.uint32_t)(unsafe.Pointer(unit))))
}
