#ifdef _WIN32
#include <winsock.h>
#else
#include <arpa/inet.h>
#endif
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME5_2.h"
#include "stats_end_native.h"

extern uint32_t dword_5d4594_741332;

// NATIVE_STATS_END_IMPLEMENTATION
static uint16_t nox_stats_quest_read_u16(const uint8_t* data) {
	uint16_t value;
	memcpy(&value, data, sizeof(value));
	return value;
}
static uint32_t nox_stats_quest_read_u32(const uint8_t* data) {
	uint32_t value;
	memcpy(&value, data, sizeof(value));
	return value;
}
static uint32_t nox_stats_quest_column_bytes(const uint8_t* header) {
	return (uint32_t)(int32_t)(int16_t)nox_stats_quest_read_u16(header) * 4u;
}

static void nox_stats_quest_players_build(nox_stats_quest_columns_native* columns, uint8_t* header) {
	uint16_t count = (uint16_t)nox_xxx_player_4E3CE0();
	memcpy(header, &count, sizeof(count));
	if ((int16_t)count > 0) {
		// 4285C0 overwrites the old slots without freeing them. Allocation and
		// live WORD reads stay in machine order; only pointer-array stride grows.
		columns->names = calloc((uint32_t)(int16_t)count, sizeof(*columns->names));
		uint32_t i = 0;
		while (i < (uint32_t)(int32_t)(int16_t)nox_stats_quest_read_u16(header)) {
			columns->names[i++] = calloc(1, 10);
		}
		void* value = calloc(1, nox_stats_quest_column_bytes(header));
		uint32_t next = (uint32_t)(int32_t)(int16_t)nox_stats_quest_read_u16(header);
		columns->ips = value;
		value = calloc(1, next);
		next = nox_stats_quest_column_bytes(header);
		columns->classes = value;
		columns->scores[0] = calloc(1, next);
		value = calloc(1, nox_stats_quest_column_bytes(header));
		next = nox_stats_quest_column_bytes(header);
		columns->scores[1] = value;
		value = calloc(1, next);
		next = nox_stats_quest_column_bytes(header);
		columns->scores[2] = value;
		columns->scores[3] = calloc(1, next);
		value = calloc(1, nox_stats_quest_column_bytes(header));
		next = nox_stats_quest_column_bytes(header);
		columns->scores[4] = value;
		value = calloc(1, next);
		next = nox_stats_quest_column_bytes(header);
		columns->scores[5] = value;
		columns->scores[6] = calloc(1, next);
		columns->scores[7] = calloc(1, nox_stats_quest_column_bytes(header));
		i = 0;
		for (nox_playerInfo* player = nox_common_playerInfoGetFirst_416EA0(); player;
			 player = nox_common_playerInfoGetNext_416EE0(player)) {
			if (!player->field_4792) continue;
			strcpy(columns->names[i], player->field_2096);
			columns->ips[i] = htonl(nox_xxx_net_getIP_554200((uint32_t)player->playerInd + 1u));
			columns->classes[i] = player->info.playerClass;
			columns->scores[0][i] = player->field_4688;
			columns->scores[1][i] = player->field_4688;
			columns->scores[2][i] = player->field_4664;
			columns->scores[3][i] = player->field_4660;
			columns->scores[4][i] = player->field_4668;
			columns->scores[5][i] = player->field_4672;
			columns->scores[6][i] = player->field_4684;
			columns->scores[7][i++] = sub_4D6540(player->playerInd);
		}
	}
	dword_5d4594_741332 = (uint32_t)(int32_t)(int16_t)nox_stats_quest_read_u16(header);
}

static uint16_t* nox_stats_quest_packet_build(const nox_stats_quest_columns_native* columns,
											const uint8_t* header, uint32_t* length) {
	static const struct { uint32_t name; uint16_t type, offset; } fields[] = {
		{71944, 6, 4}, {71952, 6, 8}, {71960, 6, 12}, {71968, 7, 24},
		{71976, 7, 280}, {71984, 6, 20}, {71992, 2, 16}, {72000, 3, 0},
	};
	void* head = NULL;
	for (size_t i = 0; i < sizeof(fields) / sizeof(fields[0]); ++i) {
		const char* name = getMemAt(0x587000, fields[i].name);
		const uint8_t* data = header + fields[i].offset;
		if (fields[i].type == 6) {
			uint32_t value = nox_stats_quest_read_u32(data);
			head = nox_stats_field_native_prepend(head, name, 6, 4, &value);
		} else if (fields[i].type == 3) {
			uint16_t value = nox_stats_quest_read_u16(data);
			head = nox_stats_field_native_prepend(head, name, 3, 2, &value);
		} else head = nox_stats_field_native_prepend(head, name, fields[i].type, 1, data);
	}
	uint32_t value = *getMemU32Ptr(0x5D4594, 741672);
	head = nox_stats_field_native_prepend(head, getMemAt(0x587000, 72008), 6, 4, &value);
	uint32_t index = 0;
	*getMemU32Ptr(0x5D4594, 741664) = 0;
	while ((int32_t)index < (int16_t)nox_stats_quest_read_u16(header)) {
		*getMemU8Ptr(0x587000, 71547) = (uint8_t)(index + 48u);
		head = nox_stats_field_native_prepend(head, getMemAt(0x587000, 71544), 7, 0, columns->names[index]);
		index = *getMemU32Ptr(0x5D4594, 741664);
		*getMemU8Ptr(0x587000, 71555) = (uint8_t)(index + 48u);
		value = columns->ips[index];
		head = nox_stats_field_native_prepend(head, getMemAt(0x587000, 71552), 6, 4, &value);
		index = *getMemU32Ptr(0x5D4594, 741664);
		*getMemU8Ptr(0x587000, 71563) = (uint8_t)(index + 48u);
		uint8_t player_class = columns->classes[index];
		head = nox_stats_field_native_prepend(head, getMemAt(0x587000, 71560), 2, 1, &player_class);
		for (uint32_t column = 0; column < 8; ++column) {
			index = *getMemU32Ptr(0x5D4594, 741664);
			uint32_t name = 71568 + column * 8;
			*getMemU8Ptr(0x587000, name + 3) = (uint8_t)(index + 48u);
			value = columns->scores[column][index];
			head = nox_stats_field_native_prepend(head, getMemAt(0x587000, name), 6, 4, &value);
		}
		index = *getMemU32Ptr(0x5D4594, 741664) + 1u;
		*getMemU32Ptr(0x5D4594, 741664) = index;
	}
	return nox_stats_packet_native_finish(head, length, getMemU32Ptr(0x5D4594, 741672));
}

int nox_stats_quest_end_native(void) {
	uint8_t* header = getMemAt(0x5D4594, 739396);
	nox_stats_quest_columns_native* columns = nox_stats_quest_columns_native_get();
	nox_stats_quest_players_build(columns, header);
	uint32_t* length = getMemU32Ptr(0x5D4594, 741300);
	uint16_t* packet = nox_stats_quest_packet_build(columns, header, length);
	uint16_t* encoded = sub_42A8B0((uint8_t*)packet, (int*)length);
	free(packet);
	char address[72];
	uint16_t port = 0;
	// Keep the existing unsupported transport boundary, not a fake send.
	if (sub_420360(address, &port)) abort();
	free(encoded);
	return 1;
}
