#ifndef NOX_INVENTORY_LOOKUP_4F78E0_H
#define NOX_INVENTORY_LOOKUP_4F78E0_H

#include <stdint.h>

typedef struct nox_object_t nox_object_t;

int32_t nox_xxx_inventoryContains_4F78E0(
	nox_object_t* holder,
	nox_object_t* item);

nox_object_t* nox_xxx_equipedItemByCode_4F7920(
	nox_object_t* holder,
	uint32_t code);

#endif // NOX_INVENTORY_LOOKUP_4F78E0_H
