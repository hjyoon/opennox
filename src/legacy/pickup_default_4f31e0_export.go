package legacy

/*
#include "pickup_default_4f31e0.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func pickupDefaultCall4F31E0(
	owner, item *server.Object,
	report, ignored int32,
) int32 {
	return int32(bool2int(Nox_xxx_pickupDefault_4F31E0(
		owner,
		item,
		int(report),
		int(ignored),
	)))
}

func pickupDefaultExportCall4F31E0(
	owner, item *server.Object,
	report, ignored int32,
) int32 {
	return int32(C.nox_xxx_pickupDefault_4F31E0(
		asObjectC(owner),
		asObjectC(item),
		C.int32_t(report),
		C.int32_t(ignored),
	))
}

//export nox_xxx_pickupDefault_4F31E0
func nox_xxx_pickupDefault_4F31E0(
	owner, item *C.nox_object_t,
	report, ignored C.int32_t,
) C.int32_t {
	return C.int32_t(pickupDefaultCall4F31E0(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(item)),
		int32(report),
		int32(ignored),
	))
}
