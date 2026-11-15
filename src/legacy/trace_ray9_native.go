package legacy

/*
#include "GAME4_3.h"
*/
import "C"

// Keep the existing ordinary Go map-trace binding as the default service.
// The original wrapper forwards the raw vector, two null outputs and flags9.
var traceRayMapCall5374B0 = nox_xxx_mapTraceRay_535250

//export nox_xxx_traceRay_5374B0
func nox_xxx_traceRay_5374B0(line *C.float4) C.int {
	return C.int(traceRayMapCall5374B0(line, nil, nil, 9))
}
