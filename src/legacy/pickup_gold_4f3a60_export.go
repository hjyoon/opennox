package legacy

/*
#include <stdint.h>

#include "GAME3_2.h"
#include "pickup_gold_4f3a60.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func pickupGoldCall4F3A60(
	owner, item *server.Object,
	arg3, arg4 int32,
) int32 {
	return Nox_xxx_pickupGold_4F3A60(owner, item, arg3, arg4)
}

func pickupGoldExportCall4F3A60(
	owner, item *server.Object,
	arg3, arg4 int32,
) int32 {
	return int32(C.nox_xxx_pickupGold_4F3A60(
		asObjectC(owner),
		asObjectC(item),
		C.int32_t(arg3),
		C.int32_t(arg4),
	))
}

// Nox_xxx_pickupGoldSendLine_4F3A60 formats the localized GoldPickup message
// with the shared line-message formatter.
func Nox_xxx_pickupGoldSendLine_4F3A60(
	owner *server.Object,
	message string,
	amount uint32,
) {
	text, free := CWString(message)
	defer free()
	sendNumberedLineMessage(owner, text, amount)
}

//export nox_xxx_pickupGold_4F3A60
func nox_xxx_pickupGold_4F3A60(
	owner, item *C.nox_object_t,
	arg3, arg4 C.int32_t,
) C.int32_t {
	return C.int32_t(pickupGoldCall4F3A60(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(item)),
		int32(arg3),
		int32(arg4),
	))
}
