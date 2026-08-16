#include "../inventory_detach_4ed0c0.h"

static void (*const inventory_detach_abi)(
	nox_object_t*, nox_object_t*) = sub_4ED0C0;

int nox_inventory_detach_abi_test(void) {
	return inventory_detach_abi != 0;
}
