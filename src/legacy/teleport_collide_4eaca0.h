#ifndef NOX_TELEPORT_COLLIDE_4EACA0_H
#define NOX_TELEPORT_COLLIDE_4EACA0_H

#include <stddef.h>
#include <stdint.h>

typedef struct nox_object_t nox_object_t;

typedef struct nox_teleport_collide_data_t {
	int32_t destination_x;
	int32_t destination_y;
} nox_teleport_collide_data_t;

_Static_assert(sizeof(nox_teleport_collide_data_t) == 8, "wrong size of Teleport collide data!");
_Static_assert(offsetof(nox_teleport_collide_data_t, destination_x) == 0, "wrong offset of Teleport X destination!");
_Static_assert(offsetof(nox_teleport_collide_data_t, destination_y) == 4, "wrong offset of Teleport Y destination!");

void sub_4EACA0(nox_object_t* source, nox_object_t* target, float* collision);

#endif // NOX_TELEPORT_COLLIDE_4EACA0_H
