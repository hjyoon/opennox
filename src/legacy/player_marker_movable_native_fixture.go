package legacy

/*
#include <stddef.h>
#include "GAME1_1.h"
#include "GAME4_2.h"

_Static_assert(_Generic(&sub_422140, nox_playerInfo* (*)(nox_playerInfo*): 1, default: 0),
    "player marker reset must retain native argument and return pointers");
_Static_assert(_Generic(&nox_xxx_isObjectMovable_52E020, unsigned int (*)(nox_object_t*): 1, default: 0),
    "movability must take a native object pointer");

static size_t marker_movable_player_local(void) { return offsetof(nox_playerInfo, field_3660); }
static size_t marker_movable_player_current(void) { return offsetof(nox_playerInfo, field_3664); }
static size_t marker_movable_object_class(void) { return offsetof(nox_object_t, obj_class); }
static size_t marker_movable_object_flags(void) { return offsetof(nox_object_t, obj_flags); }

static nox_playerInfo* marker_movable_reset(nox_playerInfo* player) { return sub_422140(player); }
static unsigned int marker_movable_is_movable(nox_object_t* obj) { return nox_xxx_isObjectMovable_52E020(obj); }
*/
import "C"

import "unsafe"

type playerMarkerMovableLayoutNative struct {
	playerSize, local, current uintptr
	objectSize, class, flags   uintptr
}

func playerMarkerMovableLayoutC() playerMarkerMovableLayoutNative {
	return playerMarkerMovableLayoutNative{
		C.sizeof_nox_playerInfo, uintptr(C.marker_movable_player_local()),
		uintptr(C.marker_movable_player_current()), C.sizeof_nox_object_t,
		uintptr(C.marker_movable_object_class()), uintptr(C.marker_movable_object_flags()),
	}
}

func playerMarkerResetC(player unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(C.marker_movable_reset((*C.nox_playerInfo)(player)))
}

func objectMovableC(obj unsafe.Pointer) uint32 {
	return uint32(C.marker_movable_is_movable((*C.nox_object_t)(obj)))
}
