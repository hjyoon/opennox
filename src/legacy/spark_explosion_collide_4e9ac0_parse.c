#include "spark_explosion_collide_4e9ac0.h"

#include <stdint.h>
#include <stdio.h>

int nox_xxx_collideSparkExplosionLoad_536DE0(
	char* args,
	nox_spark_explosion_collide_data_t* data) {
	// GAME.EXE scans into the 32-bit stack slot that initially contains the
	// input pointer. On conversion failure its low byte therefore becomes the
	// stored value. Preserve that observable fallback without aliasing a
	// pointer as int on 64-bit targets.
	int value = (int)(uint8_t)(uintptr_t)args;
	(void)sscanf(args, "%d", &value);
	data->power = (uint8_t)value;
	return 1;
}
