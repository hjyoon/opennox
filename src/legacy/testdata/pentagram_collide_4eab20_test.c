// Suppress unrelated Win32-only assertions while parsing the shared headers,
// then assert PentagramCollide's native object and callback ABI.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(sizeof(nox_pentagram_update_data_prefix_t) == 8,
	"Pentagram update-data prefix size");
_Static_assert(offsetof(nox_pentagram_update_data_prefix_t, triggered) == 4,
	"Pentagram triggered offset");
_Static_assert(offsetof(nox_object_t, data_update) == (sizeof(void*) == 4 ? 748 : 872),
	"object update-data offset");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_collidePentagram_4EAB20),
		void (*)(nox_object_t*, nox_object_t*, float*)),
	"PentagramCollide callback pointer width");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

void nox_xxx_collidePentagram_4EAB20(
	nox_object_t* source,
	nox_object_t* target,
	float* collision) {
	seen_source = source;
	seen_target = target;
	seen_collision = collision;
}

static nox_object_t* pentagram_reference(nox_object_t* source) {
	nox_pentagram_update_data_prefix_t* data = source->data_update;
	data->triggered = UINT32_C(1);
	return source;
}

int main(void) {
	nox_pentagram_update_data_prefix_t data = {
		.reserved_0 = {1, 2, 3, 4},
		.triggered = UINT32_C(0xaabbccdd),
	};
	nox_object_t source = {.data_update = &data};
	nox_object_t target = {.field_188 = UINT32_C(0x11223344)};
	float collision[2] = {3.5f, -8.25f};

	nox_xxx_collidePentagram_4EAB20(&source, &target, collision);
	if (seen_source != &source || seen_target != &target || seen_collision != collision) {
		return 1;
	}
	if (pentagram_reference(&source) != &source) {
		return 2;
	}
	if (data.reserved_0[0] != 1 || data.reserved_0[1] != 2 ||
		data.reserved_0[2] != 3 || data.reserved_0[3] != 4 ||
		data.triggered != UINT32_C(1)) {
		return 3;
	}
	if (target.field_188 != UINT32_C(0x11223344) ||
		collision[0] != 3.5f || collision[1] != -8.25f) {
		return 4;
	}
	return 0;
}
