#include "GAME3_3.h"

#include <stddef.h>
#include <stdint.h>

#ifdef NOX_TEAM_FLAG_STATUS_4E8320_NATIVE_LAYOUT
#undef _Static_assert
_Static_assert(sizeof(nox_team_flag_status_t) == 6, "team flag status must be six bytes");
_Static_assert(offsetof(nox_team_flag_status_t, team_id) == 0, "wrong team ID offset");
_Static_assert(offsetof(nox_team_flag_status_t, flag_index) == 1, "wrong flag index offset");
_Static_assert(offsetof(nox_team_flag_status_t, status) == 2, "wrong status offset");
_Static_assert(offsetof(nox_team_flag_status_t, reserved) == 3, "wrong reserved-byte offset");
_Static_assert(offsetof(nox_team_flag_status_t, carrier_net_code) == 4, "wrong carrier net-code offset");
#endif

static nox_team_flag_status_t records[UINT8_MAX + 1u];

nox_team_flag_status_t* sub_4E8320(uint8_t team_id) { return &records[team_id]; }

static nox_team_flag_status_t* (*const team_flag_status_getter_signature_4e8320)(uint8_t) = sub_4E8320;

int main(void) {
	records[UINT8_MAX].team_id = UINT8_MAX;
	records[UINT8_MAX].flag_index = UINT8_C(0xa7);
	records[UINT8_MAX].status = UINT8_C(0x81);
	records[UINT8_MAX].reserved = UINT8_C(0x7b);
	records[UINT8_MAX].carrier_net_code = UINT16_C(0xbcde);

	nox_team_flag_status_t* got = team_flag_status_getter_signature_4e8320(UINT8_MAX);
	if (got != &records[UINT8_MAX]) {
		return 1;
	}
	if (got->team_id != UINT8_MAX || got->flag_index != UINT8_C(0xa7) ||
		got->status != UINT8_C(0x81) || got->reserved != UINT8_C(0x7b) ||
		got->carrier_net_code != UINT16_C(0xbcde)) {
		return 2;
	}
	got->status = UINT8_MAX;
	got->carrier_net_code = UINT16_MAX;
	if (records[UINT8_MAX].status != UINT8_MAX ||
		records[UINT8_MAX].carrier_net_code != UINT16_MAX ||
		records[UINT8_MAX].reserved != UINT8_C(0x7b)) {
		return 3;
	}
	return 0;
}
