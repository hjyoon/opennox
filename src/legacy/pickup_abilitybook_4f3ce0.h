#ifndef NOX_PORT_PICKUP_ABILITYBOOK_4F3CE0_H
#define NOX_PORT_PICKUP_ABILITYBOOK_4F3CE0_H

#include <stdint.h>

typedef struct nox_object_t nox_object_t;

// GAME.EXE 004F3CE0..004F3D43. The registered AbilityBookPickup callback
// receives two native object pointers and two exact 32-bit scalar arguments.
int32_t nox_xxx_pickupAbilitybook_4F3CE0(
	nox_object_t* owner,
	nox_object_t* item,
	int32_t arg3,
	int32_t arg4);

#endif // NOX_PORT_PICKUP_ABILITYBOOK_4F3CE0_H
