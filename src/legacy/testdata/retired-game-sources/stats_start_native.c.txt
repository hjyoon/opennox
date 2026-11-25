#include "GAME1.h"
#include "GAME1_1.h"
#include "common__system__team.h"
#include "stats_start_native.h"

extern uint32_t dword_5d4594_741332;

// NATIVE_STATS_START_IMPLEMENTATION
static nox_stats_quest_columns_native nox_stats_quest_columns;

static uint32_t nox_stats_start_read_u32(const uint8_t* data) {
	uint32_t value;
	memcpy(&value, data, sizeof(value));
	return value;
}

static uint16_t nox_stats_start_read_u16(const uint8_t* data) {
	uint16_t value;
	memcpy(&value, data, sizeof(value));
	return value;
}

void nox_stats_start_header_native(void) {
	if (!nox_common_gameFlags_check_40A5C0(0x2000)) return;
	if (nox_common_gameFlags_check_40A5C0(4096)) return;
	const uint8_t* settings = (const uint8_t*)sub_416640();
	const uint8_t* game = (const uint8_t*)nox_xxx_cliGamedataGet_416590(0);
	*getMemU16Ptr(0x5D4594, 599482) = settings[103];
	*getMemU32Ptr(0x5D4594, 599484) = settings[104];
	*getMemU32Ptr(0x5D4594, 599488) = nox_stats_start_read_u32(settings + 40);
	*getMemU32Ptr(0x5D4594, 599492) = sub_4200E0();
	*getMemU8Ptr(0x5D4594, 599502) = (game[53] & 0xC0) != 0;
	uint16_t mode = nox_stats_start_read_u16(game + 52);
	// In GAME.EXE this DWORD is part of the packet header, not a detached
	// extracted C global. With no matching mode the existing value is kept.
	if (mode & 0x100) *getMemU32Ptr(0x5D4594, 599496) = 0;
	else if (mode & 0x20) *getMemU32Ptr(0x5D4594, 599496) = 1;
	else if (mode & 0x40) *getMemU32Ptr(0x5D4594, 599496) = 2;
	else if (mode & 0x10) *getMemU32Ptr(0x5D4594, 599496) = 3;
	else if (mode & 0x400) *getMemU32Ptr(0x5D4594, 599496) = 4;
	*getMemU8Ptr(0x5D4594, 599500) = (game[53] & 0x40) == 0;
	*getMemU32Ptr(0x5D4594, 599508) = nox_stats_start_read_u16(game + 54);
	*getMemU32Ptr(0x5D4594, 599512) = game[56];
	*getMemU8Ptr(0x5D4594, 599516) = settings[100];
	*getMemU32Ptr(0x5D4594, 599520) = settings[101] & 0xF;
	*getMemU32Ptr(0x5D4594, 599524) = settings[101] >> 4;
	*getMemU32Ptr(0x5D4594, 599528) = nox_stats_start_read_u16(settings + 105);
	*getMemU32Ptr(0x5D4594, 599532) = nox_stats_start_read_u16(settings + 107);
	*getMemU8Ptr(0x5D4594, 599536) = settings[102];
	*getMemU8Ptr(0x5D4594, 599537) = settings[100] & 0x30;
	uint8_t teams = 0;
	for (nox_team_t* team = nox_server_teamFirst_418B10(); team;
		 team = nox_server_teamNext_418B60(team)) {
		if (team->field_60) ++teams;
	}
	*getMemU8Ptr(0x5D4594, 599501) = teams;
	strncpy((char*)getMemAt(0x5D4594, 599828), (const char*)game + 9, 15);
	*getMemU8Ptr(0x5D4594, 599843) = 0;
	memcpy(getMemAt(0x5D4594, 599540), game + 24, 100);
	*getMemU32Ptr(0x5D4594, 599564) = nox_stats_start_read_u32(game + 48);
	*getMemU32Ptr(0x5D4594, 599560) = nox_stats_start_read_u32(game + 44);
	strncpy((char*)getMemAt(0x5D4594, 599572), (const char*)game, 8);
	*getMemU8Ptr(0x5D4594, 599580) = 0;
	*getMemU32Ptr(0x5D4594, 600112) = 0;
	*getMemU32Ptr(0x5D4594, 600084) = 0;
	*getMemU32Ptr(0x5D4594, 600088) = 0;
	*getMemU32Ptr(0x5D4594, 600092) = 0;
	*getMemU32Ptr(0x5D4594, 600096) = 0;
}

void* nox_stats_quest_columns_clear_native(void) {
	nox_stats_quest_columns_native* columns = &nox_stats_quest_columns;
	if (columns->names) {
		for (uint32_t i = 0; i < dword_5d4594_741332; ++i) {
			free(columns->names[i]);
			columns->names[i] = NULL;
		}
		free(columns->names);
		columns->names = NULL;
	}
	if (columns->ips) { free(columns->ips); columns->ips = NULL; }
	if (columns->classes) { free(columns->classes); columns->classes = NULL; }
	for (uint32_t i = 0; i < 7; ++i) {
		if (columns->scores[i]) { free(columns->scores[i]); columns->scores[i] = NULL; }
	}
	void* result = columns->scores[7];
	if (result) { free(columns->scores[7]); columns->scores[7] = NULL; }
	return result;
}

// The end root shares the native columns owned by Quest startup/cleanup.
nox_stats_quest_columns_native* nox_stats_quest_columns_native_get(void) {
	return &nox_stats_quest_columns;
}
