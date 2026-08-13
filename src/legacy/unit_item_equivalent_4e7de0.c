#include "GAME3_3.h"

enum {
	item_equivalent_info_book_class_4e7de0 = 0x00000100,
	item_equivalent_modifier_class_mask_4e7de0 = 0x13001000,
};

//----- (004E7DE0) --------------------------------------------------------
int32_t sub_4E7DE0(const nox_object_t* candidate, const nox_object_t* item) {
	if (!candidate || !item) {
		return 0;
	}
	if (candidate->typ_ind != item->typ_ind) {
		return 0;
	}

	const uint32_t object_class = candidate->obj_class;
	if ((object_class & (uint32_t)item_equivalent_modifier_class_mask_4e7de0) != 0) {
		const nox_modifier_attrs_t* candidate_attrs = candidate->init_data;
		const nox_modifier_attrs_t* item_attrs = item->init_data;
		for (int i = 0; i < 4; ++i) {
			if (candidate_attrs->modifiers[i] != item_attrs->modifiers[i]) {
				return 0;
			}
		}
	}

	if ((object_class & (uint32_t)item_equivalent_info_book_class_4e7de0) == 0) {
		return 1;
	}

	const uint32_t object_subclass = candidate->obj_subclass;
	if ((object_subclass & UINT32_C(1)) != 0) {
		const uint8_t* candidate_use = candidate->use_data;
		const uint8_t* item_use = item->use_data;
		return candidate_use[0] == item_use[0];
	}
	if ((object_subclass & UINT32_C(2)) != 0) {
		const uint8_t* item_use = item->use_data;
		const uint8_t* candidate_use = candidate->use_data;
		for (;;) {
			const uint8_t candidate_byte = *candidate_use++;
			const uint8_t item_byte = *item_use++;
			if (candidate_byte != item_byte) {
				return 0;
			}
			if (candidate_byte == 0) {
				return 1;
			}
		}
	}

	const uint8_t* candidate_use = candidate->use_data;
	const uint8_t* item_use = item->use_data;
	return candidate_use[0] == item_use[0];
}
