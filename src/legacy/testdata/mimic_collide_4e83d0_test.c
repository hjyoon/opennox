// Suppress unrelated Win32-only declarations while the shared legacy headers
// are parsed, then restore and assert every C field exposed by 004E83D0.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(sizeof(((nox_object_t*)0)->obj_class) == 4, "object class width");
_Static_assert(sizeof(((nox_object_t*)0)->obj_flags) == 4, "object flags width");
_Static_assert(sizeof(((nox_object_t*)0)->x) == 4, "object X width");
_Static_assert(sizeof(((nox_object_t*)0)->y) == 4, "object Y width");
_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12),
	"object class offset");
_Static_assert(offsetof(nox_object_t, obj_flags) == (sizeof(void*) == 4 ? 16 : 20),
	"object flags offset");
_Static_assert(offsetof(nox_object_t, x) == (sizeof(void*) == 4 ? 56 : 60),
	"object X offset");
_Static_assert(offsetof(nox_object_t, y) == (sizeof(void*) == 4 ? 60 : 64),
	"object Y offset");

static nox_object_t* seen_mimic;
static nox_object_t* seen_other;
static float* seen_collision;
static uint32_t result_word;

void* nox_xxx_collideMimic_4E83D0(nox_object_t* mimic, nox_object_t* other, float* collision) {
	seen_mimic = mimic;
	seen_other = other;
	seen_collision = collision;
	return &result_word;
}

static void* (*const mimic_signature)(nox_object_t*, nox_object_t*, float*) =
	nox_xxx_collideMimic_4E83D0;

int main(void) {
	nox_object_t mimic = {0};
	nox_object_t other = {0};
	float collision[2] = {-3.0f, 9.0f};
	if (mimic_signature(&mimic, &other, collision) != &result_word) {
		return 1;
	}
	if (seen_mimic != &mimic || seen_other != &other || seen_collision != collision) {
		return 2;
	}
	if (mimic_signature(0, 0, 0) != &result_word || seen_mimic != 0 || seen_other != 0 || seen_collision != 0) {
		return 3;
	}
	return 0;
}
