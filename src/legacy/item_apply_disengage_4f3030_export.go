package legacy

/*
#include "item_apply_disengage_4f3030.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func Nox_xxx_itemApplyDisengageEffect_4F3030(item, owner *server.Object) {
	server.ItemApplyDisengageEffect4F3030(item, owner)
}

func itemApplyDisengageExportCall4F3030(item, owner *server.Object) {
	C.nox_xxx_itemApplyDisengageEffect_4F3030(asObjectC(item), asObjectC(owner))
}

//export nox_xxx_itemApplyDisengageEffect_4F3030_go
func nox_xxx_itemApplyDisengageEffect_4F3030_go(item, owner *C.nox_object_t) {
	Nox_xxx_itemApplyDisengageEffect_4F3030(
		asObjectS((*nox_object_t)(item)),
		asObjectS((*nox_object_t)(owner)),
	)
}
