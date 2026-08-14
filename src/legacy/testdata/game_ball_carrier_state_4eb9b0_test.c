// Suppress unrelated Win32-only assertions while parsing the shared header,
// then assert only GameBall carrier state's native object, update-data, and
// function ABI.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(offsetof(nox_object_t, obj_class) ==
	(sizeof(void*) == 4 ? 8 : 12), "object class offset");
_Static_assert(offsetof(nox_object_t, field_13) ==
	(sizeof(void*) == 4 ? 52 : 56), "object team ID offset");
_Static_assert(offsetof(nox_object_t, owner) ==
	(sizeof(void*) == 4 ? 508 : 552), "object owner offset");
_Static_assert(offsetof(nox_object_t, data_update) ==
	(sizeof(void*) == 4 ? 748 : 872), "object update-data offset");
_Static_assert(sizeof(nox_game_ball_update_data_t) ==
	(sizeof(void*) == 4 ? 32 : 40), "GameBall update-data size");
_Static_assert(offsetof(nox_game_ball_update_data_t, carrier) == 0,
	"GameBall carrier offset");
_Static_assert(offsetof(nox_game_ball_update_data_t, team_id) ==
	(sizeof(void*) == 4 ? 4 : 8), "GameBall team ID offset");
_Static_assert(offsetof(nox_game_ball_update_data_t, ticks) ==
	(sizeof(void*) == 4 ? 8 : 16), "GameBall ticks offset");
_Static_assert(offsetof(nox_game_ball_update_data_t, carrier_frame) ==
	(sizeof(void*) == 4 ? 16 : 24), "GameBall carrier frame offset");
_Static_assert(offsetof(nox_game_ball_update_data_t, possession_duration) ==
	(sizeof(void*) == 4 ? 20 : 28), "GameBall possession duration offset");
_Static_assert(offsetof(nox_game_ball_update_data_t, reset_velocity) ==
	(sizeof(void*) == 4 ? 24 : 32), "GameBall reset velocity offset");
_Static_assert(offsetof(nox_game_ball_update_data_t, reserved) ==
	(sizeof(void*) == 4 ? 28 : 36), "GameBall reserved offset");
_Static_assert(__builtin_types_compatible_p(
	__typeof__(&sub_4EB9B0),
	nox_object_t* (*)(nox_object_t*, nox_object_t*)),
	"GameBall carrier state pointer width and return type");

static nox_object_t* seen_ball;
static nox_object_t* seen_target;

nox_object_t* sub_4EB9B0(nox_object_t* ball, nox_object_t* target) {
	seen_ball = ball;
	seen_target = target;
	return target;
}

int main(void) {
	nox_object_t previous = {0};
	nox_object_t target = {
		.obj_class = 4,
		.field_13 = 0xab,
	};
	nox_game_ball_update_data_t data = {
		.carrier = &previous,
		.team_id = UINT32_C(0xaabbccdd),
		.ticks = UINT64_C(0x0123456789abcdef),
		.carrier_frame = UINT32_C(0x89abcdef),
		.possession_duration = UINT32_C(91),
		.reset_velocity = 12.5f,
		.reserved = UINT32_C(0x76543210),
	};
	nox_object_t ball = {.data_update = &data};

	nox_object_t* result = sub_4EB9B0(&ball, &target);
	if (result != &target || seen_ball != &ball || seen_target != &target ||
		ball.data_update != &data || target.obj_class != 4 ||
		target.field_13 != 0xab || data.carrier != &previous ||
		data.team_id != UINT32_C(0xaabbccdd) ||
		data.ticks != UINT64_C(0x0123456789abcdef) ||
		data.carrier_frame != UINT32_C(0x89abcdef) ||
		data.possession_duration != UINT32_C(91) ||
		data.reset_velocity != 12.5f ||
		data.reserved != UINT32_C(0x76543210)) {
		return 1;
	}
	return 0;
}
