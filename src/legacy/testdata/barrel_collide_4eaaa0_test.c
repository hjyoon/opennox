// Suppress unrelated Win32-only assertions while parsing the shared headers,
// then assert BarrelCollide's native object and callback ABI.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(offsetof(nox_object_t, field_34) == (sizeof(void*) == 4 ? 136 : 140),
	"object collision timestamp offset");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&sub_4EAAA0),
		void (*)(nox_object_t*, nox_object_t*, float*)),
	"BarrelCollide callback pointer width");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

void sub_4EAAA0(
	nox_object_t* source,
	nox_object_t* target,
	float* collision) {
	seen_source = source;
	seen_target = target;
	seen_collision = collision;
}

static int barrel_reference(nox_object_t* source, uint32_t frame) {
	uint32_t last = source->field_34;
	if (frame <= last + UINT32_C(3)) {
		return 0;
	}
	source->field_34 = frame;
	return 281;
}

int main(void) {
	nox_object_t source = {0};
	nox_object_t target = {0};
	float collision[2] = {3.5f, -8.25f};

	sub_4EAAA0(&source, &target, collision);
	if (seen_source != &source || seen_target != &target || seen_collision != collision) {
		return 1;
	}

	source.field_34 = 5;
	if (barrel_reference(&source, 8) != 0 || source.field_34 != 5) {
		return 2;
	}
	if (barrel_reference(&source, 9) != 281 || source.field_34 != 9) {
		return 3;
	}
	source.field_34 = UINT32_MAX;
	if (barrel_reference(&source, 2) != 0 || source.field_34 != UINT32_MAX) {
		return 4;
	}
	if (barrel_reference(&source, 3) != 281 || source.field_34 != 3) {
		return 5;
	}
	return 0;
}
