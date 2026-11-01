package legacy

/*
#include <stdint.h>
#include "defs.h"
#include "GAME4.h"
_Static_assert(_Generic(&nox_xxx_respawnPlayerBot_4FAC70,
 int (*)(nox_object_t*):1,default:0),"native one-pointer bot Respawn ABI");
static uint32_t bot_remaining_original_4fab20(int root, void* p) {
 if(root==0)return nox_xxx_updatePlayerMonsterBot_4FAB20((uint32_t*)p);
 return nox_xxx_respawnPlayerBot_4FAC70((nox_object_t*)p);
}
static void* bot_regular_player_update_4fab20(void) { return (void*)nox_xxx_updatePlayer_4F8100; }
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

var playerBotRuntimeOverride4FAB20 *server.PlayerBotRuntime4FAB20

func playerBotRemainingOriginalC4FAB20(root int, unit *server.Object) uint32 {
	return uint32(C.bot_remaining_original_4fab20(C.int(root), unsafe.Pointer(unit)))
}

func playerBotRegularUpdate4FAB20() unsafe.Pointer {
	return C.bot_regular_player_update_4fab20()
}
