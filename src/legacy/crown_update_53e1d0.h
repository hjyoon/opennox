#ifndef NOX_CROWN_UPDATE_53E1D0_H
#define NOX_CROWN_UPDATE_53E1D0_H

#include <stddef.h>
#include <stdint.h>

typedef struct nox_object_t nox_object_t;

typedef struct nox_crown_update_data_t {
	nox_object_t* field_0;
	nox_object_t* pickup_target;
	uint32_t field_2;
} nox_crown_update_data_t;

_Static_assert(offsetof(nox_crown_update_data_t, field_0) == 0,
	"wrong offset of Crown field 0!");
_Static_assert(offsetof(nox_crown_update_data_t, pickup_target) == sizeof(void*),
	"wrong offset of Crown pickup target!");
_Static_assert(offsetof(nox_crown_update_data_t, field_2) == 2 * sizeof(void*),
	"wrong offset of Crown field 2!");
_Static_assert(sizeof(nox_crown_update_data_t) == (sizeof(void*) == 4 ? 12 : 24),
	"wrong size of Crown update data!");

void nox_xxx_updateCrown_53E1D0(nox_object_t* crown);

void nox_server_crownUpdateDataSetPickupTarget_53E1D0(
	nox_object_t* crown,
	nox_object_t* target);

#endif // NOX_CROWN_UPDATE_53E1D0_H
