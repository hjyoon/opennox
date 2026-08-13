#include "GAME3_3.h"

#include <stddef.h>
#include <stdint.h>

#ifdef NOX_GAME_BALL_STATUS_4E8310_NATIVE_LAYOUT
#undef _Static_assert
_Static_assert(sizeof(nox_game_ball_status_t) == 4, "GameBall status must be four bytes");
_Static_assert(offsetof(nox_game_ball_status_t, state) == 0, "wrong GameBall state offset");
_Static_assert(offsetof(nox_game_ball_status_t, reserved) == 1, "wrong GameBall reserved-byte offset");
_Static_assert(offsetof(nox_game_ball_status_t, net_code) == 2, "wrong GameBall net-code offset");
#endif

static nox_game_ball_status_t record = {
	.state = UINT8_C(0x81),
	.reserved = UINT8_C(0x7b),
	.net_code = UINT16_C(0xbcde),
};

nox_game_ball_status_t* sub_4E8310(void) { return &record; }

static nox_game_ball_status_t* (*const game_ball_status_getter_signature_4e8310)(void) = sub_4E8310;

int main(void) {
	nox_game_ball_status_t* got = game_ball_status_getter_signature_4e8310();
	if (got != &record) {
		return 1;
	}
	if (got->state != UINT8_C(0x81) || got->reserved != UINT8_C(0x7b) ||
		got->net_code != UINT16_C(0xbcde)) {
		return 2;
	}
	got->state = UINT8_MAX;
	got->net_code = UINT16_MAX;
	if (record.state != UINT8_MAX || record.net_code != UINT16_MAX ||
		record.reserved != UINT8_C(0x7b)) {
		return 3;
	}
	return 0;
}
