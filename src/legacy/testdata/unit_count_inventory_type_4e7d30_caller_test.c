#include "../GAME3_3.h"

int32_t nox_test_count_inventory_all_4e7d30(nox_object_t* owner) {
	return nox_xxx_inventoryCountObjects_4E7D30(owner, INT32_C(0));
}

int32_t nox_test_count_inventory_type_4e7d30(nox_object_t* owner, uint16_t type_ind) {
	return nox_xxx_inventoryCountObjects_4E7D30(owner, (int32_t)(uint32_t)type_ind);
}
