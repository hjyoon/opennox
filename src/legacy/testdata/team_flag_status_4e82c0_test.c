#include "GAME3_3.h"

#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#ifdef NOX_TEAM_FLAG_STATUS_4E82C0_NATIVE_LAYOUT
// Native probes suppress unrelated Win32-only assertions while headers are
// parsed, then re-enable exactly the record layout and ABI used by 004E82C0.
#undef _Static_assert
_Static_assert(sizeof(nox_team_flag_status_t) == 6, "team flag status must be six bytes");
_Static_assert(offsetof(nox_team_flag_status_t, team_id) == 0, "wrong team ID offset");
_Static_assert(offsetof(nox_team_flag_status_t, flag_index) == 1, "wrong flag index offset");
_Static_assert(offsetof(nox_team_flag_status_t, status) == 2, "wrong status offset");
_Static_assert(offsetof(nox_team_flag_status_t, reserved) == 3, "wrong reserved-byte offset");
_Static_assert(offsetof(nox_team_flag_status_t, carrier_net_code) == 4, "wrong carrier net-code offset");
#endif

static int32_t (*const team_flag_status_signature_4e82c0)(uint8_t, uint8_t, uint8_t, uint16_t) = sub_4E82C0;

static uint8_t received_team_id;
static uint8_t received_status;
static uint8_t received_flag_index;
static uint16_t received_carrier_net_code;

int32_t sub_4E82C0(uint8_t team_id, uint8_t status, uint8_t flag_index, uint16_t carrier_net_code) {
	received_team_id = team_id;
	received_status = status;
	received_flag_index = flag_index;
	received_carrier_net_code = carrier_net_code;
	return INT32_MIN + 29;
}

static int32_t call_team_flag_status(
	int32_t (*fn)(uint8_t, uint8_t, uint8_t, uint16_t),
	uint32_t team_id, uint32_t status, uint32_t flag_index, uint32_t carrier_net_code) {
	return fn((uint8_t)team_id, (uint8_t)status, (uint8_t)flag_index, (uint16_t)carrier_net_code);
}

int main(void) {
	nox_team_flag_status_t record = {
		.team_id = UINT8_C(0x12),
		.flag_index = UINT8_C(0x34),
		.status = UINT8_C(0x56),
		.reserved = UINT8_C(0x7b),
		.carrier_net_code = UINT16_C(0x89ab),
	};
	record.team_id = UINT8_MAX;
	record.flag_index = UINT8_MAX;
	record.status = UINT8_MAX;
	record.carrier_net_code = UINT16_MAX;
	if (record.reserved != UINT8_C(0x7b)) {
		return 1;
	}
	if (call_team_flag_status(team_flag_status_signature_4e82c0,
			UINT32_C(0x1ff), UINT32_C(0x2ff), UINT32_C(0x3ff), UINT32_C(0x1ffff)) != INT32_MIN + 29) {
		return 2;
	}
	if (received_team_id != UINT8_MAX || received_status != UINT8_MAX ||
		received_flag_index != UINT8_MAX || received_carrier_net_code != UINT16_MAX) {
		return 3;
	}
	return 0;
}
