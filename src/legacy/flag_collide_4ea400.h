#ifndef NOX_FLAG_COLLIDE_4EA400_H
#define NOX_FLAG_COLLIDE_4EA400_H

#include <stdint.h>

typedef struct nox_object_t nox_object_t;

void sub_4EA400(
	nox_object_t* source,
	nox_object_t* target,
	float* collision);

int32_t sub_4EA7A0(nox_object_t* object);

#endif // NOX_FLAG_COLLIDE_4EA400_H
