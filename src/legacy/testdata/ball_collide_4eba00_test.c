// Suppress unrelated Win32-only assertions while parsing the shared header,
// then assert only BallCollide's native object, GameBall record, and callback
// ABI.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(offsetof(nox_object_t, typ_ind) ==
	(sizeof(void*) == 4 ? 4 : 8), "object type index offset");
_Static_assert(offsetof(nox_object_t, obj_class) ==
	(sizeof(void*) == 4 ? 8 : 12), "object class offset");
_Static_assert(offsetof(nox_object_t, obj_flags) ==
	(sizeof(void*) == 4 ? 16 : 20), "object flags offset");
_Static_assert(offsetof(nox_object_t, net_code) ==
	(sizeof(void*) == 4 ? 36 : 40), "object net code offset");
_Static_assert(offsetof(nox_object_t, field_12) ==
	(sizeof(void*) == 4 ? 48 : 52), "object team value offset");
_Static_assert(offsetof(nox_object_t, field_13) ==
	(sizeof(void*) == 4 ? 52 : 56), "object team ID offset");
_Static_assert(offsetof(nox_object_t, owner) ==
	(sizeof(void*) == 4 ? 508 : 552), "object owner offset");
_Static_assert(offsetof(nox_object_t, field_128) ==
	(sizeof(void*) == 4 ? 512 : 560), "object owned-next offset");
_Static_assert(offsetof(nox_object_t, field_129) ==
	(sizeof(void*) == 4 ? 516 : 568), "object owned-first offset");
_Static_assert(offsetof(nox_object_t, data_update) ==
	(sizeof(void*) == 4 ? 748 : 872), "object update-data offset");
_Static_assert(sizeof(nox_game_ball_update_data_t) ==
	(sizeof(void*) == 4 ? 32 : 40), "GameBall update-data size");
_Static_assert(offsetof(nox_game_ball_update_data_t, carrier) == 0,
	"GameBall carrier offset");
_Static_assert(offsetof(nox_game_ball_update_data_t, team_id) ==
	(sizeof(void*) == 4 ? 4 : 8), "GameBall team ID offset");
_Static_assert(__builtin_types_compatible_p(
	__typeof__(&nox_xxx_collideBall_4EBA00),
	void (*)(nox_object_t*, nox_object_t*, float*)),
	"BallCollide callback pointer width and third argument");

static nox_object_t* seen_ball;
static nox_object_t* seen_target;
static float* seen_collision;

void nox_xxx_collideBall_4EBA00(
	nox_object_t* ball,
	nox_object_t* target,
	float* collision) {
	seen_ball = ball;
	seen_target = target;
	seen_collision = collision;
}

int main(void) {
	nox_object_t previous = {0};
	nox_game_ball_update_data_t data = {
		.carrier = &previous,
		.team_id = UINT32_C(0xaabbccdd),
		.ticks = UINT64_C(0x0123456789abcdef),
		.carrier_frame = UINT32_C(0x89abcdef),
		.possession_duration = UINT32_C(91),
		.reset_velocity = 12.5f,
		.reserved = UINT32_C(0x76543210),
	};
	nox_object_t owned = {.typ_ind = UINT16_C(0x44)};
	nox_object_t target = {
		.obj_class = UINT32_C(4),
		.net_code = UINT32_C(0xaabb3344),
		.field_13 = UINT32_C(7),
		.field_129 = &owned,
	};
	nox_object_t ball = {
		.typ_ind = UINT16_C(0x44),
		.obj_flags = UINT32_C(0x20),
		.net_code = UINT32_C(0x11223344),
		.field_13 = UINT32_C(3),
		.data_update = &data,
	};
	float collision[2] = {3.5f, -8.25f};

	nox_xxx_collideBall_4EBA00(&ball, &target, collision);
	if (seen_ball != &ball || seen_target != &target ||
		seen_collision != collision || ball.data_update != &data ||
		data.carrier != &previous || target.field_129 != &owned ||
		owned.typ_ind != UINT16_C(0x44) || ball.typ_ind != UINT16_C(0x44) ||
		ball.obj_flags != UINT32_C(0x20) ||
		ball.net_code != UINT32_C(0x11223344) ||
		target.net_code != UINT32_C(0xaabb3344) ||
		ball.field_13 != UINT32_C(3) || target.field_13 != UINT32_C(7)) {
		return 1;
	}
	return 0;
}
