#ifndef NOX_XFER_READABLE_4F4AB0_H
#define NOX_XFER_READABLE_4F4AB0_H

#include <stdint.h>

typedef struct nox_object_t nox_object_t;

_Static_assert(sizeof(int32_t) == 4,
	"ReadableXfer result must remain an exact 32-bit value");

int32_t nox_xxx_XFerReadable_4F4AB0(
	nox_object_t* object,
	void* context);

#endif // NOX_XFER_READABLE_4F4AB0_H
