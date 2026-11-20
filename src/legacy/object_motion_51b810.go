package legacy

/*
#include "GAME3_3.h"
#include "GAME4_1.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

var objectFallRuntime51B870 = server.ObjectFallRuntime51B870{
	Raise: (*server.Object).Raise,
	Move: func(obj *server.Object, pos types.Pointf) {
		nox_xxx_unitMove_4E7010(asObjectC(obj), (*C.float2)(unsafe.Pointer(&pos)))
	},
	Enqueue: Nox_xxx_unitHasCollideOrUpdateFn_537610,
	LandingSound: func(obj *server.Object) {
		GetServer().S().Audio.EventObj(sound.SoundPlayerFallThud, obj, 0, 0)
	},
}

func Sub_51B810(obj *server.Object) {
	server.IntegrateObjectMotion51B810(obj)
}

//export sub_51B810
func sub_51B810(obj *C.nox_object_t) {
	Sub_51B810(asObjectS(obj))
}

func Nox_xxx_updateFallLogic_51B870(obj *server.Object) {
	server.UpdateObjectFall51B870(obj, objectFallRuntime51B870)
}

//export nox_xxx_updateFallLogic_51B870
func nox_xxx_updateFallLogic_51B870(obj *C.nox_object_t) {
	Nox_xxx_updateFallLogic_51B870(asObjectS(obj))
}

//export nox_xxx_objectUnkUpdateCoords_4E7290
func nox_xxx_objectUnkUpdateCoords_4E7290(obj *C.nox_object_t) *C.nox_object_t {
	if obj != nil {
		asObjectS(obj).Nox_xxx_objectUnkUpdateCoords_4E7290()
	}
	return obj
}

//export sub_4E7350
func sub_4E7350(obj *C.nox_object_t) *C.nox_object_t {
	if obj != nil {
		asObjectS(obj).Sub_4E7350()
	}
	return obj
}

//export sub_4E7410
func sub_4E7410(obj *C.nox_object_t) C.int {
	if obj == nil {
		return 0
	}
	return C.int(asObjectS(obj).Sub_4E7410())
}
