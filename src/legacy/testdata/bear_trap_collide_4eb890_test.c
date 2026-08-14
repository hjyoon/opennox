// Suppress unrelated Win32-only assertions while parsing the shared header,
// then assert only BearTrapCollide's native object and callback ABI.
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
_Static_assert(__builtin_types_compatible_p(
	__typeof__(&nox_xxx_collideBearTrap_4EB890),
	void (*)(nox_object_t*, nox_object_t*, float*)),
	"BearTrapCollide callback pointer width");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

void nox_xxx_collideBearTrap_4EB890(
	nox_object_t* source,
	nox_object_t* target,
	float* collision) {
	seen_source = source;
	seen_target = target;
	seen_collision = collision;
}

int main(void) {
	nox_object_t owner = {0};
	nox_object_t source = {
		.x = 12.5f,
		.y = -4.25f,
		.owner = &owner,
	};
	nox_object_t target = {0};
	float collision[2] = {3.5f, -8.25f};

	nox_xxx_collideBearTrap_4EB890(&source, &target, collision);
	if (seen_source != &source || seen_target != &target ||
		seen_collision != collision || source.owner != &owner ||
		source.x != 12.5f || source.y != -4.25f) {
		return 1;
	}
	return 0;
}
