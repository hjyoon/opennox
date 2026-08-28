package legacy

/*
#include <stdint.h>

#include "GAME3_2.h"
#include "pickup_gold_4f3a60.h"

static inline void nox_pickupGoldSendLine_4F3A60(
		nox_object_t* owner, wchar2_t* message, uint32_t amount) {
	(void)nox_xxx_netSendLineMessage_4D9EB0(owner, message, amount);
}
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
// through the existing native-object line-message boundary.
func Nox_xxx_pickupGoldSendLine_4F3A60(
	owner *server.Object,
	message string,
	amount uint32,
) {
	text, free := CWString(message)
	defer free()
	C.nox_pickupGoldSendLine_4F3A60(
		asObjectC(owner),
		text,
		C.uint32_t(amount),
	)
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
