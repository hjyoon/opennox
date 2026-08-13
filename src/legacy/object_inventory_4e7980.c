#include "GAME3_3.h"

//----- (004E7980) --------------------------------------------------------
nox_object_t* nox_xxx_inventoryGetFirst_4E7980(nox_object_t* obj) {
	return obj->inv_first_item;
}

//----- (004E7990) --------------------------------------------------------
nox_object_t* nox_xxx_inventoryGetNext_4E7990(nox_object_t* obj) {
	if (obj) {
		return obj->inv_next_item;
	}
	return NULL;
}
