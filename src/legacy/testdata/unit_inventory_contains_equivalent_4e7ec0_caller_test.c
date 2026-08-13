#include "../GAME3_3.h"

#include <assert.h>

#include "../unit_item_equivalent_4e7de0.c"
#include "../unit_inventory_contains_equivalent_4e7ec0.c"

static int32_t (*const inventory_contains_equivalent_signature_4e7ec0)(
	const nox_object_t*, const nox_object_t*
) = sub_4E7EC0;

int main(void) {
	const nox_object_t* invalid = (const nox_object_t*)(uintptr_t)1;
	assert(inventory_contains_equivalent_signature_4e7ec0(NULL, invalid) == 0);
	assert(inventory_contains_equivalent_signature_4e7ec0(invalid, NULL) == 0);

	nox_object_t owner = {0};
	nox_object_t item = {.typ_ind = UINT16_C(17)};
	assert(inventory_contains_equivalent_signature_4e7ec0(&owner, &item) == 0);

	nox_object_t first = {.typ_ind = UINT16_C(18)};
	nox_object_t second = {.typ_ind = UINT16_C(17)};
	first.inv_next_item = &second;
	second.inv_next_item = (nox_object_t*)(uintptr_t)1;
	owner.inv_first_item = &first;
	assert(inventory_contains_equivalent_signature_4e7ec0(&owner, &item) == 1);

	second.typ_ind = UINT16_C(19);
	second.inv_next_item = NULL;
	assert(inventory_contains_equivalent_signature_4e7ec0(&owner, &item) == 0);

	int modifiers[5] = {0};
	nox_modifier_attrs_t candidate_attrs = {
		.modifiers = {&modifiers[0], &modifiers[1], &modifiers[2], &modifiers[3]},
	};
	nox_modifier_attrs_t item_attrs = candidate_attrs;
	second.typ_ind = item.typ_ind;
	second.obj_class = UINT32_C(0x01000000);
	second.init_data = &candidate_attrs;
	item.init_data = &item_attrs;
	owner.inv_first_item = &second;
	assert(inventory_contains_equivalent_signature_4e7ec0(&owner, &item) == 1);
	item_attrs.modifiers[3] = &modifiers[4];
	assert(inventory_contains_equivalent_signature_4e7ec0(&owner, &item) == 0);

	return 0;
}
