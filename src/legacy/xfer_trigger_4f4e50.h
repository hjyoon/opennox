#ifndef NOX_XFER_TRIGGER_4F4E50_H
#define NOX_XFER_TRIGGER_4F4E50_H

#include <stdint.h>

typedef struct nox_object_t nox_object_t;

_Static_assert(sizeof(int32_t) == 4,
	"TriggerXfer result must remain an exact 32-bit value");

int32_t nox_xxx_unitTriggerXfer_4F4E50(
	nox_object_t* object,
	void* context);

#endif // NOX_XFER_TRIGGER_4F4E50_H
