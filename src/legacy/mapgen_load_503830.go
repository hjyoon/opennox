package legacy

/*
#include "GAME4.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/ntype"
)

// The original loader calls an object's transfer callback and then places the
// object. Keep both object pointers native-width across the C/Go boundary.
var mapgenLoadPlaceObject503830 = func(name string, bounds unsafe.Pointer) bool {
	s := GetServer().S()
	obj := s.NewObjectByTypeID(name)
	if obj == nil {
		return false
	}
	if obj.Xfer == nil || obj.CallXfer(bounds) != nil {
		s.Objs.FreeObject(obj)
		return false
	}
	// GAME.EXE ignores the placement result. Placement owns the object,
	// including its cleanup on rejection.
	Nox_xxx_servMapLoadPlaceObj_4F3F50(obj, nil, (*ntype.Point32)(bounds))
	return true
}

//export nox_mapgenLoadPlaceObject_503830
func nox_mapgenLoadPlaceObject_503830(name *C.char, bounds unsafe.Pointer) C.int {
	return C.int(bool2int(mapgenLoadPlaceObject503830(GoString(name), bounds)))
}
