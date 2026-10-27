#ifndef NOX_MAP_PUSH_UNITS_52E040_H
#define NOX_MAP_PUSH_UNITS_52E040_H
#include "defs.h"

typedef struct nox_map_push_args_52E0E0 {
	float2* origin;
	float radius;
	float inner;
	float scaled_force;
	nox_object_t* source;
	uintptr_t callback;
	uintptr_t callback_arg;
} nox_map_push_args_52E0E0;

_Static_assert(sizeof(nox_map_push_args_52E0E0) == (sizeof(void*) == 4 ? 28 : 48), "radial push argument size");

void nox_xxx_mapPushUnitsAround_52E040(void* origin, float outer, float inner, float force, nox_object_t* source, uintptr_t callback, uintptr_t callback_arg);
void nox_xxx_unitPushAroundFn_52E0E0(nox_object_t* candidate, nox_map_push_args_52E0E0* args);
#endif
