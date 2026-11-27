//go:build amd64 || arm || arm64

package legacy

/*
#include "GAME3_3.h"
#include "GAME4_1.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// These native ABIs copy float arguments as binary32 values.
//
//export nox_xxx_mapDamageUnitsAround_4E25B0
func nox_xxx_mapDamageUnitsAround_4E25B0(center *C.float, outer, inner C.float, damage, typ C.int, source, excluded *nox_object_t) {
	mapDamageUnitsAround4E25B0((*types.Pointf)(unsafe.Pointer(center)), float32(outer), float32(inner), int32(damage), object.DamageType(typ), asObjectS(source), asObjectS(excluded))
}

//export nox_xxx_updateObjectsVelocity_5118A0
func nox_xxx_updateObjectsVelocity_5118A0(step C.float) C.int {
	return C.int(updateObjectsVelocity5118A0(float32(step)))
}
