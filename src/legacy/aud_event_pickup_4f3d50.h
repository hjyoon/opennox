#ifndef NOX_PORT_AUD_EVENT_PICKUP_4F3D50_H
#define NOX_PORT_AUD_EVENT_PICKUP_4F3D50_H

#include <stdint.h>

typedef struct nox_object_t nox_object_t;

// GAME.EXE 004F3D50..004F3DC8. The registered AudEventPickup callback
// receives two native object pointers and two exact 32-bit scalar arguments.
int32_t nox_objectPickupAudEvent_4F3D50(
	nox_object_t* owner,
	nox_object_t* item,
	int32_t arg3,
	int32_t arg4);

#endif // NOX_PORT_AUD_EVENT_PICKUP_4F3D50_H
