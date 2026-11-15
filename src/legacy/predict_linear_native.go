package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

var predictLinearRuntimeForCall523530 = predictLinearDefaultRuntime523530

func predictLinearDefaultRuntime523530() server.PredictLinearNativeRuntime523530 {
	return GetServer().S().PredictLinearDefaultRuntime523530()
}

func Nox_xxx_netClientPredictLinear_523530(obj *server.Object) int32 {
	return server.PredictLinearNative523530(obj, predictLinearRuntimeForCall523530())
}

//export nox_xxx_netClientPredictLinear_523530
func nox_xxx_netClientPredictLinear_523530(obj *C.nox_object_t) C.int {
	return C.int(Nox_xxx_netClientPredictLinear_523530(asObjectS((*nox_object_t)(obj))))
}
