package legacy

/*
#include "inventory_lookup_4f78e0.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func inventoryContainsExportCall4F78E0(holder, item *server.Object) int32 {
	return int32(C.nox_xxx_inventoryContains_4F78E0(
		asObjectC(holder),
		asObjectC(item),
	))
}

func equippedItemByCodeExportCall4F7920(holder *server.Object, code uint32) *server.Object {
	return asObjectS((*nox_object_t)(C.nox_xxx_equipedItemByCode_4F7920(
		asObjectC(holder),
		C.uint32_t(code),
	)))
}

//export nox_xxx_inventoryContains_4F78E0
func nox_xxx_inventoryContains_4F78E0(holder, item *C.nox_object_t) C.int32_t {
	return C.int32_t(server.InventoryContains4F78E0(
		asObjectS((*nox_object_t)(holder)),
		asObjectS((*nox_object_t)(item)),
	))
}

//export nox_xxx_equipedItemByCode_4F7920
func nox_xxx_equipedItemByCode_4F7920(holder *C.nox_object_t, code C.uint32_t) *C.nox_object_t {
	return (*C.nox_object_t)(server.EquippedItemByCode4F7920(
		asObjectS((*nox_object_t)(holder)),
		uint32(code),
	).CObj())
}
