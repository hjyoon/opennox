#ifndef NOX_XFER_MONSTER_GENERATOR_4F7130_H
#define NOX_XFER_MONSTER_GENERATOR_4F7130_H

#include <stdint.h>

typedef struct nox_object_t nox_object_t;

_Static_assert(sizeof(int32_t) == 4,
	"MonsterGeneratorXfer result must remain an exact 32-bit value");

int32_t nox_xxx_XFerMonsterGen_4F7130(nox_object_t* object);

#endif // NOX_XFER_MONSTER_GENERATOR_4F7130_H
