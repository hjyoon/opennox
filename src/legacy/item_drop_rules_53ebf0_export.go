package legacy

/*
#include "item_drop_rules_53ebf0.h"
*/
import "C"

//export nox_xxx_ItemIsDroppable_53EBF0
func nox_xxx_ItemIsDroppable_53EBF0(item *C.nox_object_t) C.int {
	return C.int(GetServer().S().ItemIsDroppable53EBF0(
		asObjectS((*nox_object_t)(item)),
	))
}
