#ifndef NOX_XFER_AMMO_4F6B20_H
#define NOX_XFER_AMMO_4F6B20_H

#include <stdint.h>

typedef struct nox_object_t nox_object_t;

_Static_assert(sizeof(int32_t) == 4,
	"AmmoXfer result must remain an exact 32-bit value");

int32_t nox_xxx_XFerAmmo_4F6B20(nox_object_t* object);

#endif // NOX_XFER_AMMO_4F6B20_H
