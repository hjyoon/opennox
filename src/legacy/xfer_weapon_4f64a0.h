#ifndef NOX_XFER_WEAPON_4F64A0_H
#define NOX_XFER_WEAPON_4F64A0_H

#include <stdint.h>

typedef struct nox_object_t nox_object_t;

_Static_assert(sizeof(int32_t) == 4,
	"WeaponXfer result must remain an exact 32-bit value");

int32_t nox_xxx_XFerWeapon_4F64A0(nox_object_t* object);

#endif // NOX_XFER_WEAPON_4F64A0_H
