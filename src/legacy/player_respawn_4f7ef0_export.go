package legacy

/*
#include "player_respawn_4f7ef0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

//export nox_xxx_playerRespawn_4F7EF0
func nox_xxx_playerRespawn_4F7EF0(unit *C.nox_object_t) C.int16_t {
	return C.int16_t(playerRespawnCall4F7EF0(
		asObjectS((*nox_object_t)(unit)),
	))
}

//export sub_4F80C0
func sub_4F80C0(gate *C.nox_object_t, output *C.float2) C.int32_t {
	return C.int32_t(GetServer().S().SoulGateRespawnPointInto4F80C0(
		asObjectS((*nox_object_t)(gate)),
		(*types.Pointf)(unsafe.Pointer(output)),
		mapTileAllowTeleport411A90,
	))
}

//export nox_xxx_respawnPlayerImpl_53FBC0
func nox_xxx_respawnPlayerImpl_53FBC0(center *C.float2, direction C.int32_t) {
	respawnPlayerImplCall53FBC0(
		(*types.Pointf)(unsafe.Pointer(center)),
		int16(direction),
	)
}
