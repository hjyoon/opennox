package legacy

/*
#include "local_unit_order_500c70.h"
*/
import "C"

import "github.com/opennox/opennox/v1/common/ntype"

func localUnitOrderExportCall500C70(owner, orderType int32) int32 {
	return int32(C.nox_xxx_orderUnitLocal_500C70(
		C.int32_t(owner),
		C.int32_t(orderType),
	))
}

//export nox_xxx_orderUnitLocal_500C70
func nox_xxx_orderUnitLocal_500C70(owner, orderType C.int32_t) C.int32_t {
	return C.int32_t(int32(GetServer().Nox_xxx_orderUnitLocal_500C70(
		ntype.PlayerInd(int32(owner)),
		uint32(orderType),
	)))
}
