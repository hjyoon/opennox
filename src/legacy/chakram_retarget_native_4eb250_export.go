package legacy

/*
#include "defs.h"
#include "GAME3_3.h"
#include "GAME4_1.h"
#include "GAME4_3.h"
#include "unit_buff_test_4ff350.h"
_Static_assert(_Generic(&sub_4EB250,nox_object_t* (*)(nox_object_t*):1,default:0),"native Chakram retarget public pointer ABI");
_Static_assert(_Generic(&sub_4EB340,void (*)(nox_object_t*,float2*):1,default:0),"native Chakram callback public pointer ABI");
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func chakramRetargetRuntime4EB250() server.ChakramRetargetRuntime4EB250 {
	if chakramRetargetNativeRuntimeOverride4EB250 != nil {
		return *chakramRetargetNativeRuntimeOverride4EB250
	}
	return server.ChakramRetargetRuntime4EB250{
		EachRect: func(rect types.Rectf, visit func(*server.Object, *types.Pointf), point *types.Pointf) {
			GetServer().S().Map.EachObjInRect(rect, func(candidate *server.Object) bool {
				visit(candidate, point)
				return true
			})
		},
		HasEnchant: func(candidate *server.Object, enchant uint32) bool {
			return nox_xxx_testUnitBuffs_4FF350(asObjectC(candidate), C.int32_t(enchant)) != 0
		},
		CanSee: func(candidate, source *server.Object) bool {
			return nox_xxx_mapCheck_537110(asObjectC(candidate), asObjectC(source)) != 0
		},
	}
}

//export sub_4EB250
func sub_4EB250(source *C.nox_object_t) *C.nox_object_t {
	result := (*server.Object)(unsafe.Pointer(source)).ChakramRetargetNative4EB250(&chakramRetargetNativeScratch4EB250, chakramRetargetRuntime4EB250())
	return (*C.nox_object_t)(unsafe.Pointer(result))
}

//export sub_4EB340
func sub_4EB340(candidate *C.nox_object_t, point *C.float2) {
	(*server.Object)(unsafe.Pointer(candidate)).ChakramRetargetCandidateNative4EB340((*types.Pointf)(unsafe.Pointer(point)), &chakramRetargetNativeScratch4EB250, chakramRetargetRuntime4EB250())
}
