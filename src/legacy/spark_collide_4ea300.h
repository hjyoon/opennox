#ifndef NOX_SPARK_COLLIDE_4EA300_H
#define NOX_SPARK_COLLIDE_4EA300_H

#include <stddef.h>
#include <stdint.h>

typedef struct nox_object_t nox_object_t;

typedef struct nox_spark_update_data_t {
	uint32_t field_0;
	uint32_t field_4;
	uint32_t field_8;
	uint32_t kind;
} nox_spark_update_data_t;
_Static_assert(sizeof(nox_spark_update_data_t) == 16,
	"wrong size of Spark update data structure!");
_Static_assert(offsetof(nox_spark_update_data_t, kind) == 12,
	"wrong offset of Spark update kind!");

void nox_xxx_collideSpark_4EA300(
	nox_object_t* source,
	nox_object_t* target,
	float* collision);

#endif // NOX_SPARK_COLLIDE_4EA300_H
