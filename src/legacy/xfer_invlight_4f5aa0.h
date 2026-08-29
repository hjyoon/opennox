#ifndef NOX_XFER_INVLIGHT_4F5AA0_H
#define NOX_XFER_INVLIGHT_4F5AA0_H

#include <stdint.h>

typedef struct nox_object_t nox_object_t;

_Static_assert(sizeof(int32_t) == 4,
	"InvisibleLightXfer result must remain an exact 32-bit value");

int32_t nox_xxx_XFerInvLight_4F5AA0(nox_object_t* object);

#endif // NOX_XFER_INVLIGHT_4F5AA0_H
