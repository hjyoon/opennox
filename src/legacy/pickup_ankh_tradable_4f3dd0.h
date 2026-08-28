#ifndef NOX_PORT_PICKUP_ANKH_TRADABLE_4F3DD0_H
#define NOX_PORT_PICKUP_ANKH_TRADABLE_4F3DD0_H

#include <stdint.h>

typedef struct nox_object_t nox_object_t;

// GAME.EXE 004F3DD0..004F3E14. The registered AnkhTradablePickup callback
// receives two native object pointers and two exact 32-bit scalar arguments.
int32_t nox_xxx_pickupAnkhTradable_4F3DD0(
	nox_object_t* owner,
	nox_object_t* item,
	int32_t arg3,
	int32_t arg4);

#endif // NOX_PORT_PICKUP_ANKH_TRADABLE_4F3DD0_H
