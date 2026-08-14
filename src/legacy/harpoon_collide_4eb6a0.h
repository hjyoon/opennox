#ifndef NOX_HARPOON_COLLIDE_4EB6A0_H
#define NOX_HARPOON_COLLIDE_4EB6A0_H

#include <stddef.h>
#include <stdint.h>

typedef struct nox_object_t nox_object_t;

typedef struct nox_harpoon_collide_data_t {
	uint32_t field_0;
	nox_object_t* owner;
} nox_harpoon_collide_data_t;

_Static_assert(offsetof(nox_harpoon_collide_data_t, field_0) == 0,
	"wrong offset of Harpoon collide field zero!");
_Static_assert(offsetof(nox_harpoon_collide_data_t, owner) == sizeof(void*),
	"wrong offset of Harpoon collide owner!");
_Static_assert(sizeof(nox_harpoon_collide_data_t) == 2 * sizeof(void*),
	"wrong size of Harpoon collide data!");

void nox_xxx_collideHarpoon_4EB6A0(
	nox_object_t* source,
	nox_object_t* target,
	float* collision);

#endif // NOX_HARPOON_COLLIDE_4EB6A0_H
