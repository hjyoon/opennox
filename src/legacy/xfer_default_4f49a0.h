#ifndef NOX_XFER_DEFAULT_4F49A0_H
#define NOX_XFER_DEFAULT_4F49A0_H

#include <stdint.h>

typedef struct nox_object_t nox_object_t;

_Static_assert(sizeof(int32_t) == 4, "DefaultXfer result must remain an exact 32-bit value");

int32_t nox_xxx_XFerDefault_4F49A0(
	nox_object_t* object,
	void* context);

#endif // NOX_XFER_DEFAULT_4F49A0_H
