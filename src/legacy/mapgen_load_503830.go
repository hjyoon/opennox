package legacy

/*
#include "GAME4.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/server"
)

// The original loader calls an object's transfer callback and then places the
// object. Keep both object pointers native-width across the C/Go boundary.
type mapgenLoadObjectDeps503830 struct {
	newObject   func(string) *server.Object
	xfer        func(*server.Object, unsafe.Pointer) error
	freeObject  func(*server.Object)
	placeObject func(*server.Object, *ntype.Point32) int32
}

func mapgenLoadObjectWithDeps503830(name string, bounds unsafe.Pointer, deps mapgenLoadObjectDeps503830) bool {
	obj := deps.newObject(name)
	if obj == nil {
		return false
	}
	if obj.Xfer == nil || deps.xfer(obj, bounds) != nil {
		deps.freeObject(obj)
		return false
	}
	// GAME.EXE ignores the placement result. Placement owns the object,
	// including its cleanup on rejection.
	deps.placeObject(obj, (*ntype.Point32)(bounds))
	return true
}

var mapgenLoadPlaceObject503830 = func(name string, bounds unsafe.Pointer) bool {
	s := GetServer().S()
	return mapgenLoadObjectWithDeps503830(name, bounds, mapgenLoadObjectDeps503830{
		newObject: s.NewObjectByTypeID,
		xfer: func(obj *server.Object, bounds unsafe.Pointer) error {
			return obj.CallXfer(bounds)
		},
		freeObject: func(obj *server.Object) {
			s.Objs.FreeObject(obj)
		},
		placeObject: func(obj *server.Object, bounds *ntype.Point32) int32 {
			return Nox_xxx_servMapLoadPlaceObj_4F3F50(obj, nil, bounds)
		},
	})
}

//export nox_mapgenLoadPlaceObject_503830
func nox_mapgenLoadPlaceObject_503830(name *C.char, bounds unsafe.Pointer) C.int {
	return C.int(bool2int(mapgenLoadPlaceObject503830(GoString(name), bounds)))
}
