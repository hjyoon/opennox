// Suppress unrelated Win32-only assertions while parsing the shared headers,
// then assert SparkCollide's native object, update-data, and callback ABI.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>

_Static_assert(sizeof(nox_spark_update_data_t) == 16,
	"Spark update data size");
_Static_assert(offsetof(nox_spark_update_data_t, kind) == 12,
	"Spark update kind offset");
_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12),
	"object class offset");
_Static_assert(offsetof(nox_object_t, field_541) == (sizeof(void*) == 4 ? 541 : 601),
	"object slow-count offset");
_Static_assert(offsetof(nox_object_t, field_542) == (sizeof(void*) == 4 ? 542 : 602),
	"object slow-timer offset");
_Static_assert(offsetof(nox_object_t, collide_data) == (sizeof(void*) == 4 ? 700 : 776),
	"object collide-data offset");
_Static_assert(offsetof(nox_object_t, data_update) == (sizeof(void*) == 4 ? 748 : 872),
	"object update-data offset");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_collideSpark_4EA300),
		void (*)(nox_object_t*, nox_object_t*, float*)),
	"SparkCollide callback pointer width");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

void nox_xxx_collideSpark_4EA300(
	nox_object_t* source,
	nox_object_t* target,
	float* collision) {
	seen_source = source;
	seen_target = target;
	seen_collision = collision;
}

int main(void) {
	nox_object_t source = {0};
	nox_object_t target = {0};
	float collision[2] = {3.5f, -8.25f};

	nox_xxx_collideSpark_4EA300(&source, &target, collision);
	if (seen_source != &source || seen_target != &target || seen_collision != collision) {
		return 1;
	}
	return 0;
}
