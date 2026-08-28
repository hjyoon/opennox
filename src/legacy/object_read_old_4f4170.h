#ifndef NOX_OBJECT_READ_OLD_4F4170_H
#define NOX_OBJECT_READ_OLD_4F4170_H

#include <stdint.h>

typedef struct nox_object_t nox_object_t;

_Static_assert(sizeof(int32_t) == 4, "old object record scalars must be exact 32-bit values");

int32_t nox_xxx_readObjectOldVer_4F4170(
	nox_object_t* object,
	int32_t object_version,
	int32_t map_version);

#endif // NOX_OBJECT_READ_OLD_4F4170_H
