#ifndef NOX_SPARK_EXPLOSION_COLLIDE_4E9AC0_H
#define NOX_SPARK_EXPLOSION_COLLIDE_4E9AC0_H

#include <stddef.h>
#include <stdint.h>

typedef struct nox_object_t nox_object_t;

typedef struct nox_spark_explosion_collide_data_t {
	uint8_t power;
} nox_spark_explosion_collide_data_t;
_Static_assert(sizeof(nox_spark_explosion_collide_data_t) == 1,
	"wrong size of SparkExplosionCollide data structure!");
_Static_assert(offsetof(nox_spark_explosion_collide_data_t, power) == 0,
	"wrong offset of SparkExplosionCollide power!");

void nox_xxx_fireballCollide_4E9AC0(
	nox_object_t* source,
	nox_object_t* target,
	float* collision);
int nox_xxx_collideSparkExplosionLoad_536DE0(
	char* args,
	nox_spark_explosion_collide_data_t* data);

#endif // NOX_SPARK_EXPLOSION_COLLIDE_4E9AC0_H
