package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func pentagramUpdateRuntime53BEF0() server.PentagramUpdateRuntime53BEF0 {
	return server.PentagramUpdateRuntime53BEF0{
		Teleport: func(obj *server.Object, destination *types.Pointf) {
			teleportToMBObject4E7190(obj, func(got *server.Object) {
				Nox_xxx_unitMove_4E7010(got, *destination)
			})
		},
	}
}

//export nox_xxx_updateTeleportPentagram_53BEF0
func nox_xxx_updateTeleportPentagram_53BEF0(pentagram *C.nox_object_t) C.int {
	return C.int(GetServer().S().TeleportPentagramUpdate53BEF0(
		asObjectS((*nox_object_t)(pentagram)),
		pentagramUpdateRuntime53BEF0(),
	))
}

//export nox_xxx_fnPentagramTeleport_53C060
func nox_xxx_fnPentagramTeleport_53C060(unit *C.nox_object_t, destination unsafe.Pointer) {
	GetServer().S().PentagramTeleportUnit53C060(
		asObjectS((*nox_object_t)(unit)),
		(*types.Pointf)(destination),
		pentagramUpdateRuntime53BEF0(),
	)
}

//export nox_xxx_updateInvisiblePentagram_53C0C0
func nox_xxx_updateInvisiblePentagram_53C0C0(pentagram *C.nox_object_t) C.int {
	return C.int(GetServer().S().InvisiblePentagramUpdate53C0C0(
		asObjectS((*nox_object_t)(pentagram)),
		pentagramUpdateRuntime53BEF0(),
	))
}

//export sub_53C140
func sub_53C140(unit *C.nox_object_t, destination unsafe.Pointer) {
	GetServer().S().PentagramTeleportUnitInvisible53C140(
		asObjectS((*nox_object_t)(unit)),
		(*types.Pointf)(destination),
		pentagramUpdateRuntime53BEF0(),
	)
}
