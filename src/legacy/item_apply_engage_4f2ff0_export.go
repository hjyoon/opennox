package legacy

/*
#include "item_apply_engage_4f2ff0.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func Nox_xxx_itemApplyEngageEffect_4F2FF0(item, owner *server.Object) {
	server.ItemApplyEngageEffect4F2FF0(item, owner)
}

func itemApplyEngageExportCall4F2FF0(item, owner *server.Object) {
	C.nox_xxx_itemApplyEngageEffect_4F2FF0(asObjectC(item), asObjectC(owner))
}

//export nox_xxx_itemApplyEngageEffect_4F2FF0
func nox_xxx_itemApplyEngageEffect_4F2FF0(item, owner *C.nox_object_t) {
	Nox_xxx_itemApplyEngageEffect_4F2FF0(
		asObjectS((*nox_object_t)(item)),
		asObjectS((*nox_object_t)(owner)),
	)
}
