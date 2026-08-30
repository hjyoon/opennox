#ifndef NOX_XFER_OBELISK_4F6F60_H
#define NOX_XFER_OBELISK_4F6F60_H

#include <stdint.h>

typedef struct nox_object_t nox_object_t;

_Static_assert(sizeof(int32_t) == 4,
	"ObeliskXfer result must remain an exact 32-bit value");

int32_t nox_xxx_XFerObelisk_4F6F60(nox_object_t* object);

#endif // NOX_XFER_OBELISK_4F6F60_H
