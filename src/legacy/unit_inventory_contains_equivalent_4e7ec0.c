#include "GAME3_3.h"

//----- (004E7EC0) --------------------------------------------------------
int32_t sub_4E7EC0(const nox_object_t* owner, const nox_object_t* item) {
	if (!owner || !item) {
		return 0;
	}

	for (const nox_object_t* candidate = owner->inv_first_item;
		 candidate;
		 candidate = candidate->inv_next_item) {
		if (sub_4E7DE0(candidate, item)) {
			return 1;
		}
	}
	return 0;
}
