#ifndef NOX_POISON_GAS_TRAP_COLLIDE_4EB910_H
#define NOX_POISON_GAS_TRAP_COLLIDE_4EB910_H

#include <stddef.h>
#include <stdint.h>

typedef struct nox_object_t nox_object_t;

typedef struct nox_toxic_cloud_update_data_t {
	int32_t duration;
} nox_toxic_cloud_update_data_t;

_Static_assert(sizeof(nox_toxic_cloud_update_data_t) == 4,
	"wrong size of nox_toxic_cloud_update_data_t structure");
_Static_assert(offsetof(nox_toxic_cloud_update_data_t, duration) == 0,
	"wrong offset of ToxicCloud duration");

void nox_xxx_collidePoisonGasTrap_4EB910(
	nox_object_t* source,
	nox_object_t* target,
	float* collision);

#endif // NOX_POISON_GAS_TRAP_COLLIDE_4EB910_H
