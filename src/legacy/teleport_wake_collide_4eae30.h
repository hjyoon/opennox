#ifndef NOX_TELEPORT_WAKE_COLLIDE_4EAE30_H
#define NOX_TELEPORT_WAKE_COLLIDE_4EAE30_H

#include <stddef.h>

typedef struct nox_object_t nox_object_t;

typedef struct nox_teleport_wake_collide_data_t {
	float destination_x;
	float destination_y;
} nox_teleport_wake_collide_data_t;

_Static_assert(sizeof(nox_teleport_wake_collide_data_t) == 8,
	"wrong size of TeleportWake collide data!");
_Static_assert(offsetof(nox_teleport_wake_collide_data_t, destination_x) == 0,
	"wrong offset of TeleportWake X destination!");
_Static_assert(offsetof(nox_teleport_wake_collide_data_t, destination_y) == 4,
	"wrong offset of TeleportWake Y destination!");

void nox_xxx_collideTeleportWake_4EAE30(
	nox_object_t* source,
	nox_object_t* target,
	float* collision);

#endif // NOX_TELEPORT_WAKE_COLLIDE_4EAE30_H
