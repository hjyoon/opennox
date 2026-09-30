package legacy

/*
#include "GAME4_1.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

//export nox_xxx_questGeneratorType_native_51A500
func nox_xxx_questGeneratorType_native_51A500(unit *C.nox_object_t) C.int {
	return C.int(questGeneratorTypeNative51A500(
		asObjectS((*nox_object_t)(unit)), questGeneratorTypeDepsFactory51A500(),
	))
}

func questGeneratorTypeCEntry51A500(unit *server.Object) int32 {
	return int32(C.sub_51A500(asObjectC(unit)))
}
