package legacy

/*
#include "item_apply_disengage_4f3030.h"
#include "cgo_const.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func Nox_xxx_itemApplyDisengageEffect_4F3030(item, owner *server.Object) {
	server.ItemApplyDisengageEffect4F3030(item, owner)
}

func itemApplyDisengageExportCall4F3030(item, owner *server.Object) {
	C.nox_xxx_itemApplyDisengageEffect_4F3030(asObjectC(item), asObjectC(owner))
}

//export nox_xxx_itemApplyDisengageEffect_4F3030
func nox_xxx_itemApplyDisengageEffect_4F3030(item *C.nox_cgo_const_object_t, owner *nox_object_t) {
	Nox_xxx_itemApplyDisengageEffect_4F3030(
		asObjectS(item),
		asObjectS(owner),
	)
}
