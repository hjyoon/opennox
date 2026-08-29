#ifndef NOX_XFER_FIELD_GUIDE_4F6390_H
#define NOX_XFER_FIELD_GUIDE_4F6390_H

#include <stdint.h>

typedef struct nox_object_t nox_object_t;

_Static_assert(sizeof(int32_t) == 4,
	"FieldGuideXfer result must remain an exact 32-bit value");

int32_t nox_xxx_XFerFieldGuide_4F6390(nox_object_t* object);

#endif // NOX_XFER_FIELD_GUIDE_4F6390_H
