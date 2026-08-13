#include "../GAME3_3.h"

#include <assert.h>

#include "../unit_item_equivalent_4e7de0.c"

static int32_t (*const item_equivalent_signature_4e7de0)(
	const nox_object_t*, const nox_object_t*
) = sub_4E7DE0;

int main(void) {
	const nox_object_t* invalid = (const nox_object_t*)(uintptr_t)1;
	assert(item_equivalent_signature_4e7de0(NULL, invalid) == 0);
	assert(item_equivalent_signature_4e7de0(invalid, NULL) == 0);

	nox_object_t candidate = {0};
	nox_object_t item = {0};
	candidate.typ_ind = UINT16_C(0xffff);
	item.typ_ind = UINT16_C(0xfffe);
	assert(item_equivalent_signature_4e7de0(&candidate, &item) == 0);
	item.typ_ind = candidate.typ_ind;
	item.obj_class = UINT32_MAX;
	item.obj_subclass = UINT32_MAX;
	assert(item_equivalent_signature_4e7de0(&candidate, &item) == 1);

	int modifiers[5] = {0};
	nox_modifier_attrs_t candidate_attrs = {
		.modifiers = {&modifiers[0], &modifiers[1], &modifiers[2], &modifiers[3]},
	};
	nox_modifier_attrs_t item_attrs = candidate_attrs;
	candidate.obj_class = UINT32_C(0x01000000);
	candidate.init_data = &candidate_attrs;
	item.init_data = &item_attrs;
	assert(item_equivalent_signature_4e7de0(&candidate, &item) == 1);
	for (int i = 0; i < 4; ++i) {
		item_attrs = candidate_attrs;
		item_attrs.modifiers[i] = &modifiers[4];
		assert(item_equivalent_signature_4e7de0(&candidate, &item) == 0);
	}

	uint8_t candidate_use[] = {'G', 'u', 'i', 'd', 'e', 0, 'x'};
	uint8_t item_use[] = {'G', 'u', 'i', 'd', 'e', 0, 'y'};
	candidate.obj_class = UINT32_C(0x00000100);
	candidate.obj_subclass = UINT32_C(2);
	candidate.use_data = candidate_use;
	item.use_data = item_use;
	assert(item_equivalent_signature_4e7de0(&candidate, &item) == 1);
	item_use[4] = 'X';
	assert(item_equivalent_signature_4e7de0(&candidate, &item) == 0);

	candidate.obj_subclass = UINT32_C(3);
	candidate_use[0] = UINT8_C(77);
	item_use[0] = UINT8_C(77);
	assert(item_equivalent_signature_4e7de0(&candidate, &item) == 1);
	item_use[0] = UINT8_C(78);
	assert(item_equivalent_signature_4e7de0(&candidate, &item) == 0);

	candidate.obj_subclass = 0;
	item_use[0] = candidate_use[0];
	assert(item_equivalent_signature_4e7de0(&candidate, &item) == 1);
	item_use[0] ^= UINT8_C(1);
	assert(item_equivalent_signature_4e7de0(&candidate, &item) == 0);

	return 0;
}
