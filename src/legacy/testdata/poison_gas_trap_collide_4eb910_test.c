// Suppress unrelated Win32-only assertions while parsing the shared header,
// then assert only PoisonGasTrapCollide's native object, update-data, and
// callback ABI.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>

_Static_assert(offsetof(nox_object_t, x) ==
	(sizeof(void*) == 4 ? 56 : 60), "object X position offset");
_Static_assert(offsetof(nox_object_t, y) ==
	(sizeof(void*) == 4 ? 60 : 64), "object Y position offset");
_Static_assert(offsetof(nox_object_t, owner) ==
	(sizeof(void*) == 4 ? 508 : 552), "object owner offset");
_Static_assert(offsetof(nox_object_t, data_update) ==
	(sizeof(void*) == 4 ? 748 : 872), "object update-data offset");
_Static_assert(sizeof(nox_toxic_cloud_update_data_t) == 4,
	"ToxicCloud update-data size");
_Static_assert(offsetof(nox_toxic_cloud_update_data_t, duration) == 0,
	"ToxicCloud duration offset");
_Static_assert(__builtin_types_compatible_p(
	__typeof__(&nox_xxx_collidePoisonGasTrap_4EB910),
	void (*)(nox_object_t*, nox_object_t*, float*)),
	"PoisonGasTrapCollide callback pointer width");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

void nox_xxx_collidePoisonGasTrap_4EB910(
	nox_object_t* source,
	nox_object_t* target,
	float* collision) {
	seen_source = source;
	seen_target = target;
	seen_collision = collision;
}

int main(void) {
	nox_object_t owner = {0};
	nox_toxic_cloud_update_data_t data = {.duration = -75};
	nox_object_t source = {
		.x = 12.5f,
		.y = -4.25f,
		.owner = &owner,
		.data_update = &data,
	};
	nox_object_t target = {0};
	float collision[2] = {3.5f, -8.25f};

	nox_xxx_collidePoisonGasTrap_4EB910(&source, &target, collision);
	if (seen_source != &source || seen_target != &target ||
		seen_collision != collision || source.owner != &owner ||
		source.data_update != &data || data.duration != -75 ||
		source.x != 12.5f || source.y != -4.25f) {
		return 1;
	}
	return 0;
}
