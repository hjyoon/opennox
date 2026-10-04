#include <limits.h>
#include <stdint.h>

#include "../monster_look_at_5125a0.h"

typedef float* (*monster_look_at_fn)(nox_object_t*, int);
typedef uintptr_t (*monster_look_at_bridge_fn)(nox_object_t*, int);

_Static_assert(CHAR_BIT == 8 && sizeof(int) == 4,
	"script direction must remain signed 32-bit");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(sizeof(uintptr_t) == sizeof(void*),
	"residual action identity must retain native pointer width");
_Static_assert(_Generic(&nox_xxx_monsterLookAt_5125A0, monster_look_at_fn: 1, default: 0),
	"005125A0 must retain its original pointer-returning C ABI");
_Static_assert(_Generic(&nox_server_monster_look_at_5125a0, monster_look_at_bridge_fn: 1, default: 0),
	"Go bridge must carry a native object/result and signed direction");
