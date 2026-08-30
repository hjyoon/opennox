// Suppress unrelated Win32-only assertions while parsing the shared object
// definition, then assert only HomeBaseCollide's native object, GameBall
// update-data and callback ABI.
#include <stdio.h>

#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(offsetof(nox_object_t, typ_ind) ==
	(sizeof(void*) == 4 ? 4 : 8), "object type index offset");
_Static_assert(offsetof(nox_object_t, field_12) ==
	(sizeof(void*) == 4 ? 48 : 52), "object team value offset");
_Static_assert(offsetof(nox_object_t, field_13) ==
	(sizeof(void*) == 4 ? 52 : 56), "object team ID offset");
_Static_assert(offsetof(nox_object_t, x) ==
	(sizeof(void*) == 4 ? 56 : 60), "object position offset");
_Static_assert(offsetof(nox_object_t, vel_x) ==
	(sizeof(void*) == 4 ? 80 : 84), "object velocity offset");
_Static_assert(offsetof(nox_object_t, force_x) ==
	(sizeof(void*) == 4 ? 88 : 92), "object force offset");
_Static_assert(offsetof(nox_object_t, float_25) ==
	(sizeof(void*) == 4 ? 100 : 104), "object Pos24.Y offset");
_Static_assert(offsetof(nox_object_t, object_next) ==
	(sizeof(void*) == 4 ? 444 : 448), "object next offset");
_Static_assert(offsetof(nox_object_t, data_update) ==
	(sizeof(void*) == 4 ? 748 : 872), "object update-data offset");
_Static_assert(sizeof(nox_game_ball_update_data_t) ==
	(sizeof(void*) == 4 ? 32 : 40), "GameBall update-data size");
_Static_assert(offsetof(nox_game_ball_update_data_t, carrier) == 0,
	"GameBall carrier offset");
_Static_assert(offsetof(nox_player_update_data_t, player) ==
	(sizeof(void*) == 4 ? 276 : 336), "PlayerUpdate player offset");
_Static_assert(__builtin_types_compatible_p(
	__typeof__(&nox_xxx_collideHomeBase_4EBB80),
	uint32_t (*)(nox_object_t*, nox_object_t*, float*)),
	"HomeBaseCollide callback return width and three pointer arguments");

static nox_object_t* seen_home_base;
static nox_object_t* seen_other;
static float* seen_collision;

uint32_t nox_xxx_collideHomeBase_4EBB80(
	nox_object_t* home_base,
	nox_object_t* other,
	float* collision) {
	seen_home_base = home_base;
	seen_other = other;
	seen_collision = collision;
	return UINT32_C(0xf1234567);
}

int main(void) {
	nox_object_t carrier = {.field_13 = UINT32_C(8)};
	nox_player_update_data_t player_update = {0};
	nox_game_ball_update_data_t ball_update = {.carrier = &carrier};
	nox_object_t marker = {
		.typ_ind = UINT16_C(0x3333),
		.x = 30.0f,
		.y = 40.0f,
	};
	nox_object_t ball = {
		.typ_ind = UINT16_C(0x2222),
		.x = 10.0f,
		.y = 20.0f,
		.vel_x = 1.0f,
		.vel_y = 2.0f,
		.force_x = 3.0f,
		.force_y = 4.0f,
		.float_24 = 5.0f,
		.float_25 = 6.0f,
		.object_next = &marker,
		.data_update = &ball_update,
	};
	nox_object_t home_base = {.field_13 = UINT32_C(7)};
	float collision[2] = {3.5f, -8.25f};

	carrier.data_update = &player_update;
	uint32_t result = nox_xxx_collideHomeBase_4EBB80(
		&home_base, &ball, collision);
	if (result != UINT32_C(0xf1234567) ||
		seen_home_base != &home_base || seen_other != &ball ||
		seen_collision != collision || ball.data_update != &ball_update ||
		ball_update.carrier != &carrier || carrier.data_update != &player_update ||
		ball.object_next != &marker || ball.typ_ind != UINT16_C(0x2222) ||
		marker.typ_ind != UINT16_C(0x3333) || home_base.field_13 != UINT32_C(7) ||
		carrier.field_13 != UINT32_C(8) || ball.x != 10.0f || ball.y != 20.0f ||
		ball.vel_x != 1.0f || ball.vel_y != 2.0f || ball.force_x != 3.0f ||
		ball.force_y != 4.0f || ball.float_24 != 5.0f || ball.float_25 != 6.0f) {
		return 1;
	}
	return 0;
}
