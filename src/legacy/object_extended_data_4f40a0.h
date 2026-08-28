#ifndef NOX_OBJECT_EXTENDED_DATA_4F40A0_H
#define NOX_OBJECT_EXTENDED_DATA_4F40A0_H

#include <stdint.h>

typedef struct nox_object_t nox_object_t;

_Static_assert(sizeof(int8_t) == 1, "object extended-data result must be one byte");
_Static_assert(INT8_MIN == -128 && INT8_MAX == 127,
	"object extended-data result requires an exact signed 8-bit type");

int8_t sub_4F40A0(nox_object_t* object);

#endif // NOX_OBJECT_EXTENDED_DATA_4F40A0_H
