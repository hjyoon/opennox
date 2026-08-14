// Suppress unrelated Win32-only assertions while parsing the shared header,
// then assert the GlyphCollide callback and helper native pointer boundaries.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>

_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12),
	"object class offset");
_Static_assert(offsetof(nox_object_t, obj_flags) == (sizeof(void*) == 4 ? 16 : 20),
	"object flags offset");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_collideGlyph_4E9A00),
		void (*)(nox_object_t*, nox_object_t*, float*)),
	"GlyphCollide callback pointer width");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&sub_4E9A30),
		int (*)(nox_object_t*, nox_object_t*)),
	"GlyphCollide helper pointer width");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

void nox_xxx_collideGlyph_4E9A00(
	nox_object_t* source,
	nox_object_t* target,
	float* collision) {
	seen_source = source;
	seen_target = target;
	seen_collision = collision;
}

int sub_4E9A30(nox_object_t* source, nox_object_t* target) {
	seen_source = source;
	seen_target = target;
	return source != NULL && target != NULL;
}

int main(void) {
	nox_object_t source = {0};
	nox_object_t target = {0};
	float collision[2] = {3.5f, -8.25f};

	nox_xxx_collideGlyph_4E9A00(&source, &target, collision);
	if (seen_source != &source || seen_target != &target || seen_collision != collision) {
		return 1;
	}
	if (sub_4E9A30(&source, &target) != 1 || seen_source != &source || seen_target != &target) {
		return 2;
	}
	return 0;
}
