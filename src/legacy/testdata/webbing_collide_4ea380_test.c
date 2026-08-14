// Suppress unrelated Win32-only assertions while parsing the shared headers,
// then assert WebbingCollide's native object and callback ABI.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>

_Static_assert(sizeof(nox_object_t) == (sizeof(void*) == 4 ? 772 : 912),
	"object size");
_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12),
	"object class offset");
_Static_assert(offsetof(nox_object_t, func_damage) == (sizeof(void*) == 4 ? 716 : 808),
	"object Damage callback offset");
_Static_assert(sizeof(((nox_object_t*)0)->func_damage) == sizeof(void*),
	"object Damage callback width");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(((nox_object_t*)0)->func_damage),
		int (*)(nox_object_t*, nox_object_t*, nox_object_t*, int32_t, int32_t)),
	"object Damage callback signature");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_collideWebbing_4EA380),
		void (*)(nox_object_t*, nox_object_t*, float*)),
	"WebbingCollide callback pointer width");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

static int damage_signature(
	nox_object_t* target,
	nox_object_t* parent,
	nox_object_t* source,
	int32_t damage,
	int32_t damage_type) {
	return target == seen_target && parent == seen_source && source == seen_source &&
		damage == 0 && damage_type == 2 ? -1 : 0;
}

void nox_xxx_collideWebbing_4EA380(
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

	nox_xxx_collideWebbing_4EA380(&source, &target, collision);
	if (seen_source != &source || seen_target != &target || seen_collision != collision) {
		return 1;
	}
	target.func_damage = damage_signature;
	if (target.func_damage(&target, &source, &source, 0, 2) != -1) {
		return 2;
	}
	return 0;
}
