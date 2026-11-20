package legacy

/*
#include "GAME4_1.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

func enqueueUnitCollision5117F0(obj *server.Object) byte {
	if obj == nil {
		return 0
	}
	if obj.ObjClass.Has(object.ClassMissile) {
		// On this branch GAME.EXE leaves the object's address low byte in AL.
		return byte(uintptr(unsafe.Pointer(obj)))
	}
	return pendingCollisionEnqueue537610(obj)
}

func Sub_5117F0(obj *server.Object) {
	enqueueUnitCollision5117F0(obj)
}

//export sub_5117F0
func sub_5117F0(obj *C.nox_object_t) C.char {
	return C.char(enqueueUnitCollision5117F0(asObjectS(obj)))
}

func Nox_xxx_unit_511810(obj *server.Object) {
	if obj == nil || obj.ObjClass.Has(object.ClassMissile) || pendingCollisionQueue537580.Contains(obj) == 0 {
		return
	}
	pendingCollisionRemove5375A0(obj)
}

//export nox_xxx_unit_511810
func nox_xxx_unit_511810(obj *C.nox_object_t) {
	Nox_xxx_unit_511810(asObjectS(obj))
}

//export sub_51B860
func sub_51B860(obj *C.nox_object_t) C.char {
	if obj == nil {
		return 0
	}
	return C.char(pendingCollisionEnqueue537610(asObjectS(obj)))
}

//export nox_enqueue_wall_collision_51B860
func nox_enqueue_wall_collision_51B860(obj *C.nox_object_t, _ unsafe.Pointer) {
	sub_51B860(obj)
}
