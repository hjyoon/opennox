// Suppress unrelated Win32-only assertions while parsing the shared object
// definition, then assert only CrownCollide's native object and callback ABI.
#include <stdio.h>

#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(offsetof(nox_object_t, obj_class) ==
	(sizeof(void*) == 4 ? 8 : 12), "object class offset");
_Static_assert(offsetof(nox_object_t, obj_flags) ==
	(sizeof(void*) == 4 ? 16 : 20), "object flags offset");
_Static_assert(__builtin_types_compatible_p(
	__typeof__(&sub_4EBB50),
	uintptr_t (*)(nox_object_t*, nox_object_t*, float*)),
	"CrownCollide callback pointer width and third argument");

static nox_object_t* seen_crown;
static nox_object_t* seen_target;
static float* seen_collision;

uintptr_t sub_4EBB50(
	nox_object_t* crown,
	nox_object_t* target,
	float* collision) {
	seen_crown = crown;
	seen_target = target;
	seen_collision = collision;
	return (uintptr_t)target;
}

int main(void) {
	nox_object_t crown = {0};
	nox_object_t target = {
		.obj_class = UINT32_C(0x80000004),
		.obj_flags = UINT32_C(0x40000000),
	};
	float collision[2] = {3.5f, -8.25f};

	uintptr_t result = sub_4EBB50(&crown, &target, collision);
	if (seen_crown != &crown || seen_target != &target ||
		seen_collision != collision || result != (uintptr_t)&target ||
		target.obj_class != UINT32_C(0x80000004) ||
		target.obj_flags != UINT32_C(0x40000000)) {
		return 1;
	}
	return 0;
}
