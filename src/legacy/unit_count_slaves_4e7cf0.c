#include "GAME3_3.h"

#include <string.h>

//----- (004E7CF0) --------------------------------------------------------
int32_t nox_xxx_unitCountSlaves_4E7CF0(
	const nox_object_t* owner,
	uint32_t class_mask,
	uint32_t subclass_mask
) {
	if (!owner || class_mask == 0 || subclass_mask == 0) {
		return 0;
	}

	uint32_t count = 0;
	for (const nox_object_t* obj = owner->field_129; obj; obj = obj->field_128) {
		if ((obj->obj_class & class_mask) != 0 && (obj->obj_subclass & subclass_mask) != 0) {
			++count;
		}
	}

	int32_t result;
	memcpy(&result, &count, sizeof(result));
	return result;
}
