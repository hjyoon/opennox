#ifndef NOX_XFER_ARMOR_4F6860_H
#define NOX_XFER_ARMOR_4F6860_H

#include <stdint.h>

typedef struct nox_object_t nox_object_t;

_Static_assert(sizeof(int32_t) == 4,
	"ArmorXfer result must remain an exact 32-bit value");

int32_t nox_xxx_XFerArmor_4F6860(nox_object_t* object);

#endif // NOX_XFER_ARMOR_4F6860_H
