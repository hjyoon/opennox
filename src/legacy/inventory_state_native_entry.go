package legacy

/*
#include "GAME2_1.h"
#include "GAME3_3.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

//export sub_461600
func sub_461600(thingType C.int) *C.nox_drawable {
	return (*C.nox_drawable)(inventoryStateFindType461600(uint32(thingType)).C())
}

//export sub_461930
func sub_461930() C.int {
	return C.int(inventoryStateHasClass461930())
}

//export sub_464B40
func sub_464B40(column, row C.int) C.int {
	return C.int(inventoryStateBounds464B40(int32(column), int32(row)))
}

//export sub_467680
func sub_467680() {
	inventoryStateReset467680()
}

//export nox_xxx_wndGetHandle_4676A0
func nox_xxx_wndGetHandle_4676A0() *C.nox_window {
	return (*C.nox_window)(Get_dword_5d4594_1062452().C())
}

//export sub_467740
func sub_467740(value C.int) C.int {
	return C.int(inventoryStateSet467740(int32(value)))
}

//export nox_xxx_inventoryGetFirst_4E7980
func nox_xxx_inventoryGetFirst_4E7980(owner *C.nox_object_t) *C.nox_object_t {
	first := objectInventoryFirst4E7980(asObjectS((*nox_object_t)(owner)), func(obj *server.Object) *server.Object {
		return obj.InvFirstItem
	})
	return (*C.nox_object_t)(asObjectC(first))
}
