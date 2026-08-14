// Suppress unrelated Win32-only assertions while parsing the shared headers,
// then assert the native OwnCollide callback and observed object fields.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>

_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12),
	"object class offset");
_Static_assert(offsetof(nox_object_t, field_34) == (sizeof(void*) == 4 ? 136 : 140),
	"object owner frame offset");
_Static_assert(offsetof(nox_object_t, owner) == (sizeof(void*) == 4 ? 508 : 552),
	"object owner offset");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&sub_4EA2C0),
		void (*)(nox_object_t*, nox_object_t*, float*)),
	"OwnCollide callback pointer width");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

void sub_4EA2C0(
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

	sub_4EA2C0(&source, &target, collision);
	if (seen_source != &source || seen_target != &target || seen_collision != collision) {
		return 1;
	}
	return 0;
}
