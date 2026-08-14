#ifndef NOX_GAME_BALL_CARRIER_STATE_4EB9B0_H
#define NOX_GAME_BALL_CARRIER_STATE_4EB9B0_H

#include <stddef.h>
#include <stdint.h>

typedef struct nox_object_t nox_object_t;

typedef struct nox_game_ball_update_data_t {
	nox_object_t* carrier;
	uint32_t team_id;
	uint64_t ticks;
	uint32_t carrier_frame;
	uint32_t possession_duration;
	float reset_velocity;
	uint32_t reserved;
} nox_game_ball_update_data_t;

_Static_assert(sizeof(nox_game_ball_update_data_t) ==
	(sizeof(void*) == 4 ? 32 : 40),
	"wrong size of nox_game_ball_update_data_t structure");
_Static_assert(offsetof(nox_game_ball_update_data_t, carrier) == 0,
	"wrong offset of GameBall carrier");
_Static_assert(offsetof(nox_game_ball_update_data_t, team_id) ==
	(sizeof(void*) == 4 ? 4 : 8), "wrong offset of GameBall team ID");
_Static_assert(offsetof(nox_game_ball_update_data_t, ticks) ==
	(sizeof(void*) == 4 ? 8 : 16), "wrong offset of GameBall ticks");
_Static_assert(offsetof(nox_game_ball_update_data_t, carrier_frame) ==
	(sizeof(void*) == 4 ? 16 : 24), "wrong offset of GameBall carrier frame");
_Static_assert(offsetof(nox_game_ball_update_data_t, possession_duration) ==
	(sizeof(void*) == 4 ? 20 : 28), "wrong offset of GameBall possession duration");
_Static_assert(offsetof(nox_game_ball_update_data_t, reset_velocity) ==
	(sizeof(void*) == 4 ? 24 : 32), "wrong offset of GameBall reset velocity");
_Static_assert(offsetof(nox_game_ball_update_data_t, reserved) ==
	(sizeof(void*) == 4 ? 28 : 36), "wrong offset of GameBall reserved dword");

nox_object_t* sub_4EB9B0(nox_object_t* ball, nox_object_t* target);

#endif // NOX_GAME_BALL_CARRIER_STATE_4EB9B0_H
