#include <assert.h>

#include "../object_inventory_4e7980.c"

static nox_object_t* (*const first_signature)(nox_object_t*) = nox_xxx_inventoryGetFirst_4E7980;
static nox_object_t* (*const next_signature)(nox_object_t*) = nox_xxx_inventoryGetNext_4E7990;

int main(void) {
	nox_object_t owner = {0};
	nox_object_t first = {0};
	nox_object_t second = {0};

	owner.inv_first_item = &first;
	first.inv_next_item = &second;
	assert(first_signature(&owner) == &first);
	assert(next_signature(&first) == &second);
	assert(next_signature(&second) == NULL);
	assert(next_signature(NULL) == NULL);
	return 0;
}
