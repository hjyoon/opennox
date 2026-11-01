package legacy

/*
#include <stdint.h>
#include "defs.h"
#include "GAME4.h"
_Static_assert(_Generic(&nox_xxx_playerBotCreate_4FA700,
 int (*)(nox_object_t*):1,default:0),"native one-pointer bot Create ABI");
static int32_t bot_create_original_4fa700(void* p) {
 return nox_xxx_playerBotCreate_4FA700((nox_object_t*)p);
}
static uintptr_t bot_create_pud_4fa700(void* p) {
 return (uintptr_t)((nox_object_t*)p)->data_update;
}
static uintptr_t bot_create_mud_4fa700(void* p) {
 return (uintptr_t)((nox_player_update_data_t*)p)->field_73;
}
static size_t bot_create_link_width_4fa700(void) {
 return sizeof(((nox_player_update_data_t*)0)->field_73);
}
static void* bot_create_update_function_4fa700(void) {
 return (void*)nox_xxx_updatePlayerMonsterBot_4FAB20;
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

var playerBotCreateRuntimeOverride4FA700 *server.PlayerBotCreateRuntime4FA700

func playerBotCreateOriginalUpdateFn4FA700() unsafe.Pointer {
	return C.bot_create_update_function_4fa700()
}

func playerBotCreateOriginalC4FA700(unit *server.Object) uint32 {
	return uint32(C.bot_create_original_4fa700(unsafe.Pointer(unit)))
}

func playerBotCreateOriginalControls4FA700(unit *server.Object, pud *server.PlayerUpdateData) (uintptr, uintptr, uintptr) {
	return uintptr(C.bot_create_pud_4fa700(unsafe.Pointer(unit))), uintptr(C.bot_create_mud_4fa700(unsafe.Pointer(pud))), uintptr(C.bot_create_link_width_4fa700())
}
