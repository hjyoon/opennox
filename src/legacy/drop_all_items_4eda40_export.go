package legacy

/*
#include "drop_all_items_4eda40.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func dropAllItemsCall4EDA40(owner *server.Object) int32 {
	return GetServer().S().DropAllItems4EDA40(owner, server.DropAllItemsRuntime4EDA40{
		Dispatch: objectDropDispatchCall4ED790,
	})
}

//export nox_xxx_dropAllItems_4EDA40
func nox_xxx_dropAllItems_4EDA40(owner *C.nox_object_t) C.int32_t {
	return C.int32_t(dropAllItemsCall4EDA40(
		asObjectS((*nox_object_t)(owner)),
	))
}
