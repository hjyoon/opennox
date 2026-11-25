#include "GAME3_3.h"

//----- (004E7BC0) --------------------------------------------------------
int sub_4E7BC0(const nox_object_t* obj) {
	if (!obj) {
		return 0;
	}
	const uint32_t class = obj->obj_class;
	return (int)((class >> 2) & UINT32_C(1));
}
