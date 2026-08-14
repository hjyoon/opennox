// Suppress unrelated Win32-only assertions while parsing the shared headers,
// then assert SignCollide's native object, Use callback, and collision ABI.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12),
	"object class offset");
_Static_assert(offsetof(nox_object_t, func_use) == (sizeof(void*) == 4 ? 732 : 840),
	"object Use offset");
_Static_assert(
	__builtin_types_compatible_p(
		nox_object_use_func_t,
		int (*)(nox_object_t*, nox_object_t*)),
	"object Use callback pointer width");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_collideSign_4EAB40),
		void (*)(nox_object_t*, nox_object_t*, float*)),
	"SignCollide callback pointer width");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;
static nox_object_t* used_target;
static nox_object_t* used_source;
static uint32_t use_calls;

void nox_xxx_collideSign_4EAB40(
	nox_object_t* source,
	nox_object_t* target,
	float* collision) {
	seen_source = source;
	seen_target = target;
	seen_collision = collision;
}

static int sign_use(nox_object_t* target, nox_object_t* source) {
	used_target = target;
	used_source = source;
	use_calls++;
	return -7;
}

static void sign_reference(nox_object_t* source, nox_object_t* target) {
	if (target != NULL && ((uint8_t)target->obj_class & UINT8_C(4)) != 0) {
		(void)source->func_use(target, source);
	}
}

int main(void) {
	nox_object_t source = {
		.func_use = sign_use,
		.field_188 = UINT32_C(0x11223344),
	};
	nox_object_t target = {
		.obj_class = UINT32_C(0x80000004),
		.field_188 = UINT32_C(0x55667788),
	};
	float collision[2] = {3.5f, -8.25f};

	nox_xxx_collideSign_4EAB40(&source, &target, collision);
	if (seen_source != &source || seen_target != &target || seen_collision != collision) {
		return 1;
	}

	sign_reference(&source, NULL);
	if (use_calls != 0) {
		return 2;
	}
	target.obj_class = UINT32_C(0xffffff80);
	sign_reference(&source, &target);
	if (use_calls != 0) {
		return 3;
	}
	target.obj_class = UINT32_C(0x80000004);
	sign_reference(&source, &target);
	if (use_calls != 1 || used_target != &target || used_source != &source) {
		return 4;
	}
	if (source.field_188 != UINT32_C(0x11223344) ||
		target.field_188 != UINT32_C(0x55667788) ||
		collision[0] != 3.5f || collision[1] != -8.25f) {
		return 5;
	}
	return 0;
}
