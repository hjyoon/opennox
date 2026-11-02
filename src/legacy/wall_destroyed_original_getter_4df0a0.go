package legacy

/*
#include <stdint.h>
#include "GAME3_2.h"
#include "map_temp_lists_503f40.h"
static int wall_destroyed_old_4df0a0(int (*fn)(int,int), nox_map_wall_504290* wall, int recipient) {
 return fn((int)(uintptr_t)wall,recipient);
}
static int wall_destroyed_native_4df0a0(int (*fn)(nox_map_wall_504290*,int), nox_map_wall_504290* wall, int recipient) {
 return fn(wall,recipient);
}
static int wall_destroyed_original_public_4df0a0(nox_map_wall_504290* wall, int recipient) {
 return _Generic(&nox_xxx_wallSendDestroyed_4DF0A0,int (*)(int,int):wall_destroyed_old_4df0a0,int (*)(nox_map_wall_504290*,int):wall_destroyed_native_4df0a0)(&nox_xxx_wallSendDestroyed_4DF0A0,wall,recipient);
}
static int wall_destroyed_original_typed_4df0a0(void) {
 return _Generic(&nox_xxx_wallSendDestroyed_4DF0A0,int (*)(int,int):0,int (*)(nox_map_wall_504290*,int):1);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func wallDestroyedOriginalC4DF0A0(wall *server.Wall, recipient int32) int32 {
	return int32(C.wall_destroyed_original_public_4df0a0((*C.nox_map_wall_504290)(unsafe.Pointer(wall)), C.int(recipient)))
}

func wallDestroyedOriginalTyped4DF0A0() bool {
	return C.wall_destroyed_original_typed_4df0a0() != 0
}
