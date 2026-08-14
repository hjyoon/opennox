#ifndef NOX_FIST_COLLIDE_4EADF0_H
#define NOX_FIST_COLLIDE_4EADF0_H

#include <stddef.h>
#include <stdint.h>

typedef struct nox_object_t nox_object_t;

typedef struct nox_fist_update_data_t {
	int32_t damage;
} nox_fist_update_data_t;

_Static_assert(sizeof(nox_fist_update_data_t) == 4,
	"wrong size of Fist update data!");
_Static_assert(offsetof(nox_fist_update_data_t, damage) == 0,
	"wrong offset of Fist damage!");

void nox_xxx_collideFist_4EADF0(
	nox_object_t* source,
	nox_object_t* target,
	float* collision);

#endif // NOX_FIST_COLLIDE_4EADF0_H
