package legacy

/*
#include <stdint.h>
#include "defs.h"
#include "GAME4.h"
_Static_assert(_Generic(&nox_xxx_monsterActionToPlrState_4FABC0,
 char (*)(nox_object_t*):1,default:0),"native one-pointer bot action ABI");
static char bot_action_original_4fabc0(void* p) {
 return nox_xxx_monsterActionToPlrState_4FABC0((nox_object_t*)p);
}
*/
import "C"

import (
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

func playerBotActionOriginalC4FABC0(unit *server.Object) byte {
	return byte(C.bot_action_original_4fabc0(unsafe.Pointer(unit)))
}
