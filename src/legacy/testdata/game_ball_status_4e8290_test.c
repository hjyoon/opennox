#include "GAME3_3.h"

#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#ifdef NOX_GAME_BALL_STATUS_4E8290_NATIVE_LAYOUT
// Native probes suppress unrelated Win32-only assertions while headers are
// parsed, then re-enable exactly the record layout and ABI used by 004E8290.
#undef _Static_assert
_Static_assert(sizeof(nox_game_ball_status_t) == 4, "GameBall status must be four bytes");
_Static_assert(offsetof(nox_game_ball_status_t, state) == 0, "wrong GameBall state offset");
_Static_assert(offsetof(nox_game_ball_status_t, reserved) == 1, "wrong GameBall reserved-byte offset");
_Static_assert(offsetof(nox_game_ball_status_t, net_code) == 2, "wrong GameBall net-code offset");
#endif

static int32_t (*const game_ball_status_signature_4e8290)(uint8_t, uint16_t) = sub_4E8290;

static uint8_t received_state;
static uint16_t received_net_code;

int32_t sub_4E8290(uint8_t state, uint16_t net_code) {
	received_state = state;
	received_net_code = net_code;
	return INT32_MIN + 17;
}

static int32_t call_game_ball_status(int32_t (*fn)(uint8_t, uint16_t), uint32_t state, uint32_t net_code) {
	return fn((uint8_t)state, (uint16_t)net_code);
}

int main(void) {
	nox_game_ball_status_t record = {
		.state = UINT8_C(0x12),
		.reserved = UINT8_C(0x7b),
		.net_code = UINT16_C(0x3456),
	};
	record.state = UINT8_MAX;
	record.net_code = UINT16_MAX;
	if (record.reserved != UINT8_C(0x7b)) {
		return 1;
	}
	if (call_game_ball_status(game_ball_status_signature_4e8290, UINT32_C(0x1ff), UINT32_C(0x1ffff)) !=
		INT32_MIN + 17) {
		return 2;
	}
	if (received_state != UINT8_MAX || received_net_code != UINT16_MAX) {
		return 3;
	}
	return 0;
}
