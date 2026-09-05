// Parse only the production types used by this fixture so strict-warning
// builds retain every relevant layout assertion.
#if defined(__clang__)
#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wstrict-prototypes"
#pragma clang diagnostic ignored "-Wpedantic"
#elif defined(__GNUC__)
#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wstrict-prototypes"
#pragma GCC diagnostic ignored "-Wpedantic"
#endif
#include "../defs.h"
#if defined(__clang__)
#pragma clang diagnostic pop
#elif defined(__GNUC__)
#pragma GCC diagnostic pop
#endif
#include "../pentagram_collide_4eab20.h"

#include <stddef.h>
#include <stdint.h>

_Static_assert(sizeof(nox_pentagram_update_data_prefix_t) == 8,
	"Pentagram update-data prefix size");
_Static_assert(offsetof(nox_pentagram_update_data_prefix_t, triggered) == 4,
	"Pentagram triggered offset");
_Static_assert(sizeof(nox_pentagram_update_data_t) == 24,
	"Pentagram fixed update-data size");
_Static_assert(offsetof(nox_pentagram_update_data_t, destination_pe32) == 12,
	"Pentagram fixed PE32 destination offset");
_Static_assert(offsetof(nox_pentagram_update_data_t, destination_extent) == 16,
	"Pentagram fixed destination extent offset");
_Static_assert(offsetof(nox_pentagram_update_data_t, animation_step) == 20,
	"Pentagram fixed animation-step offset");
_Static_assert(offsetof(nox_object_t, data_update) == (sizeof(void*) == 4 ? 748 : 872),
	"object update-data offset");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_collidePentagram_4EAB20),
		void (*)(nox_object_t*, nox_object_t*, float*)),
	"PentagramCollide callback pointer width");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_updateTeleportPentagram_53BEF0),
		int (*)(nox_object_t*)),
	"TeleportPentagram update ABI");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_fnPentagramTeleport_53C060),
		void (*)(nox_object_t*, void*)),
	"TeleportPentagram callback ABI");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_updateInvisiblePentagram_53C0C0),
		int (*)(nox_object_t*)),
	"InvisibleTeleportPentagram update ABI");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&sub_53C140),
		void (*)(nox_object_t*, void*)),
	"InvisibleTeleportPentagram callback ABI");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

void nox_xxx_collidePentagram_4EAB20(
	nox_object_t* source,
	nox_object_t* target,
	float* collision) {
	seen_source = source;
	seen_target = target;
	seen_collision = collision;
}

static nox_object_t* pentagram_reference(nox_object_t* source) {
	nox_pentagram_update_data_prefix_t* data = source->data_update;
	data->triggered = UINT32_C(1);
	return source;
}

int main(void) {
	nox_pentagram_update_data_prefix_t data = {
		.reserved_0 = {1, 2, 3, 4},
		.triggered = UINT32_C(0xaabbccdd),
	};
	nox_object_t source = {.data_update = &data};
	nox_object_t target = {.field_188 = UINT32_C(0x11223344)};
	float collision[2] = {3.5f, -8.25f};

	nox_xxx_collidePentagram_4EAB20(&source, &target, collision);
	if (seen_source != &source || seen_target != &target || seen_collision != collision) {
		return 1;
	}
	if (pentagram_reference(&source) != &source) {
		return 2;
	}
	if (data.reserved_0[0] != 1 || data.reserved_0[1] != 2 ||
		data.reserved_0[2] != 3 || data.reserved_0[3] != 4 ||
		data.triggered != UINT32_C(1)) {
		return 3;
	}
	if (target.field_188 != UINT32_C(0x11223344) ||
		collision[0] != 3.5f || collision[1] != -8.25f) {
		return 4;
	}
	nox_pentagram_update_data_t complete = {
		.destination_pe32 = UINT32_C(0xffffffff),
		.destination_extent = UINT32_C(0xf1234567),
		.animation_step = UINT8_C(8),
	};
	if (complete.destination_pe32 != UINT32_C(0xffffffff) ||
		complete.destination_extent != UINT32_C(0xf1234567) ||
		complete.animation_step != UINT8_C(8)) {
		return 5;
	}
	return 0;
}
