#ifndef NOX_OBJECT_MAP_READ_WRITE_4F4530_H
#define NOX_OBJECT_MAP_READ_WRITE_4F4530_H

#include <stdint.h>

typedef struct nox_object_t nox_object_t;

_Static_assert(sizeof(int32_t) == 4, "object map serializer scalars must be exact 32-bit values");

int32_t nox_xxx_mapReadWriteObjData_4F4530(
	nox_object_t* object,
	int32_t map_version);

#endif // NOX_OBJECT_MAP_READ_WRITE_4F4530_H
