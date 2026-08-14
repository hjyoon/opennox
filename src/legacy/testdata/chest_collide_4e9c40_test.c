// Suppress unrelated Win32-only assertions while parsing the shared header,
// then assert ChestCollide's native object and callback ABI.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>

_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12),
	"object class offset");
_Static_assert(offsetof(nox_object_t, obj_subclass) == (sizeof(void*) == 4 ? 12 : 16),
	"object subclass offset");
_Static_assert(offsetof(nox_object_t, obj_flags) == (sizeof(void*) == 4 ? 16 : 20),
	"object flags offset");
_Static_assert(offsetof(nox_object_t, inv_next_item) == (sizeof(void*) == 4 ? 496 : 528),
	"object inventory-next offset");
_Static_assert(offsetof(nox_object_t, inv_first_item) == (sizeof(void*) == 4 ? 504 : 544),
	"object inventory-first offset");
_Static_assert(offsetof(nox_object_t, func_die) == (sizeof(void*) == 4 ? 724 : 824),
	"object Death callback offset");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_collideChest_4E9C40),
		void (*)(nox_object_t*, nox_object_t*, float*)),
	"ChestCollide callback pointer width");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

void nox_xxx_collideChest_4E9C40(
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

	nox_xxx_collideChest_4E9C40(&source, &target, collision);
	if (seen_source != &source || seen_target != &target || seen_collision != collision) {
		return 1;
	}
	return 0;
}
