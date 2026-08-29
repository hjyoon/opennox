#ifndef NOX_XFER_SENTRY_4F5E50_H
#define NOX_XFER_SENTRY_4F5E50_H

#include <stdint.h>

typedef struct nox_object_t nox_object_t;

_Static_assert(sizeof(int32_t) == 4,
	"SentryXfer result must remain an exact 32-bit value");

int32_t nox_xxx_XFerSentry_4F5E50(nox_object_t* object);

#endif // NOX_XFER_SENTRY_4F5E50_H
