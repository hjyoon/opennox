#include <stddef.h>

#include "GAME4_1.h"

extern nox_object_t* nox_minimapFirstMonsterObject_50AAE0(void);
extern nox_object_t* nox_minimapNextMonsterObject_50AB10(void);

float* nox_xxx_minimapFirstMonster_50AAE0(void) {
	nox_object_t* obj = nox_minimapFirstMonsterObject_50AAE0();
	return obj != NULL ? &obj->x : NULL;
}

float* nox_xxx_minimapNextMonster_50AB10(void) {
	nox_object_t* obj = nox_minimapNextMonsterObject_50AB10();
	return obj != NULL ? &obj->x : NULL;
}
