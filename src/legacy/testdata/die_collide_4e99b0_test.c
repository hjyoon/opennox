// Suppress unrelated Win32-only assertions while parsing the shared header,
// then assert DieCollide and Object death-callback native pointer boundaries.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>

_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12),
	"object class offset");
_Static_assert(offsetof(nox_object_t, obj_flags) == (sizeof(void*) == 4 ? 16 : 20),
	"object flags offset");
_Static_assert(offsetof(nox_object_t, func_die) == (sizeof(void*) == 4 ? 724 : 824),
	"object death callback offset");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(((nox_object_t*)0)->func_die),
		void (*)(nox_object_t*)),
	"object death callback pointer width");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_collideDie_4E99B0),
		void (*)(nox_object_t*, nox_object_t*, float*)),
	"DieCollide callback pointer width");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static nox_object_t* seen_death;
static float* seen_collision;

static void death_callback(nox_object_t* obj) {
	seen_death = obj;
}

void nox_xxx_collideDie_4E99B0(
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
	source.func_die = death_callback;

	nox_xxx_collideDie_4E99B0(&source, &target, collision);
	if (seen_source != &source || seen_target != &target || seen_collision != collision) {
		return 1;
	}
	source.func_die(&source);
	if (seen_death != &source) {
		return 2;
	}
	return 0;
}
