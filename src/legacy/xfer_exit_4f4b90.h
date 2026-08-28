#ifndef NOX_XFER_EXIT_4F4B90_H
#define NOX_XFER_EXIT_4F4B90_H

#include <stdint.h>

typedef struct nox_object_t nox_object_t;

_Static_assert(sizeof(int32_t) == 4,
	"ExitXfer result must remain an exact 32-bit value");

int32_t nox_xxx_XFerExit_4F4B90(
	nox_object_t* object,
	void* context);

#endif // NOX_XFER_EXIT_4F4B90_H
