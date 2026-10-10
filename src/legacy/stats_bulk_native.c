#include <time.h>
#ifdef _WIN32
#include <winsock.h>
#else
#include <arpa/inet.h>
#endif
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "stats_bulk_native.h"

extern uint32_t dword_5d4594_608316;
extern uint32_t dword_5d4594_739392;
extern uint32_t dword_5d4594_600116;
extern uint32_t dword_5d4594_741332;

// NATIVE_STATS_IMPLEMENTATION
typedef struct nox_stats_field_native {
	char name[4];
	uint16_t type;
	uint16_t length;
	void* data;
	struct nox_stats_field_native* next;
} nox_stats_field_native;

static nox_stats_columns_native nox_stats_columns;

static uint32_t nox_stats_read_u32(const uint8_t* at) {
	uint32_t value;
	memcpy(&value, at, sizeof(value));
	return value;
}

static uint16_t nox_stats_read_u16(const uint8_t* at) {
	uint16_t value;
	memcpy(&value, at, sizeof(value));
	return value;
}

static void nox_stats_field_clear(nox_stats_field_native* field) {
	free(field->data);
	strcpy(field->name, (const char*)getMemAt(0x5D4594, 741688));
	field->type = 0;
	field->length = 0;
	field->data = NULL;
	field->next = NULL;
}

static void nox_stats_field_add(nox_stats_field_native** head, const char* name, uint16_t type,
								uint16_t length, const void* data) {
	nox_stats_field_native* field = calloc(1, sizeof(*field));
	if (field) {
		field->data = NULL;
		nox_stats_field_clear(field);
		strncpy(field->name, name, 4);
		field->type = type;
		if (type == 7) length = (uint16_t)(strlen((const char*)data) + 1);
		field->length = length;
		field->data = calloc(1, length);
		memcpy(field->data, data, length);
		field->next = NULL;
	}
	// Preserve 42C360's unguarded link after failed node allocation.
	field->next = *head;
	*head = field;
}

static void nox_stats_add_u32(nox_stats_field_native** head, const char* name, uint32_t value) {
	nox_stats_field_add(head, name, 6, sizeof(value), &value);
}
static void nox_stats_add_u16(nox_stats_field_native** head, const char* name, uint16_t value) {
	nox_stats_field_add(head, name, 3, sizeof(value), &value);
}
static void nox_stats_add_u8(nox_stats_field_native** head, const char* name, uint8_t value) {
	nox_stats_field_add(head, name, 2, sizeof(value), &value);
}

static uint16_t* nox_stats_serialize(nox_stats_field_native* head, uint32_t* length) {
	*length = 4;
	for (nox_stats_field_native* field = head; field; field = field->next) {
		*length += 8;
		*length += field->length + ((0u - (field->length + *length)) & 3u);
	}
	uint16_t* packet = calloc(1, *length);
	packet[0] = htons((uint16_t)*length);
	packet[1] = htons(0);
	uint8_t* cursor = (uint8_t*)(packet + 2);
	for (nox_stats_field_native* field = head; field; field = field->next) {
		switch (field->type) {
		case 3: case 4: *(uint16_t*)field->data = htons(*(uint16_t*)field->data); break;
		case 5: case 6: *(uint32_t*)field->data = htonl(*(uint32_t*)field->data); break;
		}
		field->type = htons(field->type);
		field->length = htons(field->length);
		memcpy(cursor, field->name, 8);
		cursor += 8;
		memcpy(cursor, field->data, ntohs(field->length));
		cursor += ntohs(field->length);
		cursor += (0u - ntohs(field->length)) & 3u;
		field->type = ntohs(field->type);
		field->length = ntohs(field->length);
		switch (field->type) {
		case 3: case 4: *(uint16_t*)field->data = ntohs(*(uint16_t*)field->data); break;
		case 5: case 6: *(uint32_t*)field->data = ntohl(*(uint32_t*)field->data); break;
		}
	}
	return packet;
}

static void nox_stats_fields_destroy(nox_stats_field_native* field) {
	while (field) {
		nox_stats_field_native* next = field->next;
		nox_stats_field_clear(field);
		free(field);
		field = next;
	}
}

static void nox_stats_players_build(nox_stats_columns_native* columns, uint8_t* header,
									uint8_t* records, uint32_t count) {
	if (columns->names) {
		for (uint32_t i = 0; i < dword_5d4594_741332; ++i) {
			free(columns->names[i]);
			columns->names[i] = NULL;
		}
		free(columns->names);
		columns->names = NULL;
	}
	if (columns->ips) { free(columns->ips); columns->ips = NULL; }
	if (columns->teams) { free(columns->teams); columns->teams = NULL; }
	if (columns->classes) { free(columns->classes); columns->classes = NULL; }
	if (columns->active) { free(columns->active); columns->active = NULL; }
	if (columns->durations) { free(columns->durations); columns->durations = NULL; }
	if (columns->participants) { free(columns->participants); columns->participants = NULL; }
	columns->names = calloc(count, sizeof(*columns->names));
	for (uint32_t i = 0; i < count; ++i) columns->names[i] = calloc(1, 10);
	uint32_t bytes = count * 4u;
	columns->ips = calloc(1, bytes);
	columns->teams = calloc(1, bytes);
	columns->classes = calloc(1, count);
	columns->active = calloc(1, count);
	columns->durations = calloc(1, bytes);
	columns->participants = calloc(1, count);
	for (uint32_t i = 0; i < count; ++i) {
		uint8_t* record = records + i * 32u;
		strcpy(columns->names[i], (const char*)record);
		columns->ips[i] = nox_stats_read_u32(record + 12);
		columns->teams[i] = nox_stats_read_u32(record + 16);
		columns->classes[i] = record[20];
		columns->active[i] = record[21];
		columns->participants[i] = record[28];
		uint32_t since = nox_stats_read_u32(record + 24);
		uint32_t elapsed = (uint32_t)time(NULL) - since;
		memcpy(record + 24, &elapsed, sizeof(elapsed));
		columns->durations[i] = elapsed;
	}
	uint16_t players = (uint16_t)count;
	memcpy(header + 6, &players, sizeof(players));
	dword_5d4594_741332 = count;
}

static uint32_t nox_stats_pairs_build(nox_stats_columns_native* columns, const uint8_t* records, uint32_t count) {
	if (columns->pairs) { free(columns->pairs); columns->pairs = NULL; }
	uint32_t length = count * 2u;
	columns->pairs = calloc(1, length);
	uint32_t copied = 0;
	if (length) {
		do {
			columns->pairs[copied] = records[copied];
			columns->pairs[copied + 1] = records[copied + 1];
			copied += 2;
		} while (copied < length);
	}
	*getMemU32Ptr(0x5D4594, 741308) = count;
	return copied;
}

static uint16_t* nox_stats_packet_build(const nox_stats_columns_native* columns, const uint8_t* header,
										int mode, int16_t pairs, uint32_t* length) {
	static const struct { char name[5]; uint16_t type, offset; } fields[] = {
		{"MXPL",6,8}, {"IDNO",6,12}, {"GSKU",6,16}, {"GSTY",6,20}, {"CLGM",2,24},
		{"LIMT",6,32}, {"TLMT",6,36}, {"RSTC",2,40}, {"MINE",6,44}, {"MAXE",6,48},
		{"MINP",6,52}, {"MAXP",6,56}, {"VIDM",2,60}, {"SVRS",2,61}, {"NTMS",2,25},
		{"SCEN",7,96}, {"GNAM",7,352}, {"SPL1",6,64}, {"SPL2",6,68}, {"SPL3",6,72},
		{"ARMR",6,88}, {"WPN1",2,84}, {"WPN2",2,85}, {"WPN3",2,86}, {"STAF",6,92},
		{"DURA",6,28},
	};
	nox_stats_field_native* head = NULL;
	for (size_t i = 0; i < sizeof(fields) / sizeof(fields[0]); ++i) {
		const uint8_t* data = header + fields[i].offset;
		if (fields[i].type == 6) nox_stats_add_u32(&head, fields[i].name, nox_stats_read_u32(data));
		else if (fields[i].type == 2) nox_stats_add_u8(&head, fields[i].name, *data);
		else nox_stats_field_add(&head, fields[i].name, 7, 0, data);
	}
	nox_stats_add_u8(&head, "FINI", 1);
	nox_stats_add_u8(&head, "TRNY", header[26]);
	if (mode == 0) {
		*getMemU32Ptr(0x5D4594, 741668) = 0;
		nox_stats_add_u32(&head, "SEQU", 0);
		nox_stats_add_u8(&head, "ENDF", 0);
		nox_stats_add_u16(&head, getMemAt(0x587000, 71872), UINT16_MAX);
	} else if (mode == 1 || mode == 2) {
		uint32_t sequence = ++*getMemU32Ptr(0x5D4594, 741668);
		nox_stats_add_u32(&head, "SEQU", sequence);
		nox_stats_add_u8(&head, "ENDF", mode == 2);
		nox_stats_add_u16(&head, getMemAt(0x587000, mode == 1 ? 71928 : 71896), nox_stats_read_u16(header + 6));
		uint32_t index = 0;
		*getMemU32Ptr(0x5D4594, 741660) = 0;
		while ((int32_t)index < (int16_t)nox_stats_read_u16(header + 6)) {
			*getMemU8Ptr(0x587000, 71491) = (uint8_t)(index + 48);
			nox_stats_field_add(&head, getMemAt(0x587000, 71488), 7, 0, columns->names[index]);
			*getMemU8Ptr(0x587000, 71499) = (uint8_t)(*getMemU32Ptr(0x5D4594, 741660) + 48);
			nox_stats_add_u32(&head, getMemAt(0x587000, 71496), columns->ips[*getMemU32Ptr(0x5D4594, 741660)]);
			*getMemU8Ptr(0x587000, 71515) = (uint8_t)(*getMemU32Ptr(0x5D4594, 741660) + 48);
			nox_stats_add_u32(&head, getMemAt(0x587000, 71512), columns->teams[*getMemU32Ptr(0x5D4594, 741660)]);
			*getMemU8Ptr(0x587000, 71507) = (uint8_t)(*getMemU32Ptr(0x5D4594, 741660) + 48);
			nox_stats_add_u8(&head, getMemAt(0x587000, 71504), columns->classes[*getMemU32Ptr(0x5D4594, 741660)]);
			*getMemU8Ptr(0x587000, 71523) = (uint8_t)(*getMemU32Ptr(0x5D4594, 741660) + 48);
			nox_stats_add_u8(&head, getMemAt(0x587000, 71520), columns->active[*getMemU32Ptr(0x5D4594, 741660)]);
			*getMemU8Ptr(0x587000, 71531) = (uint8_t)(*getMemU32Ptr(0x5D4594, 741660) + 48);
			nox_stats_add_u32(&head, getMemAt(0x587000, 71528), columns->durations[*getMemU32Ptr(0x5D4594, 741660)]);
			*getMemU8Ptr(0x587000, 71539) = (uint8_t)(*getMemU32Ptr(0x5D4594, 741660) + 48);
			nox_stats_add_u8(&head, getMemAt(0x587000, 71536), columns->participants[*getMemU32Ptr(0x5D4594, 741660)]);
			index = *getMemU32Ptr(0x5D4594, 741660) + 1;
			*getMemU32Ptr(0x5D4594, 741660) = index;
		}
		nox_stats_field_add(&head, getMemAt(0x587000, mode == 1 ? 71936 : 71904), 20,
							(uint16_t)(2 * pairs), columns->pairs);
	}
	uint16_t* packet = nox_stats_serialize(head, length);
	nox_stats_fields_destroy(head);
	return packet;
}

static int nox_stats_report(nox_stats_columns_native* columns, const uint8_t* header, int mode) {
	uint32_t* length = getMemU32Ptr(0x5D4594, 741312);
	uint16_t* packet = nox_stats_packet_build(columns, header, mode, *getMemI16Ptr(0x5D4594, 741308), length);
	uint16_t* encoded = sub_42A8B0((uint8_t*)packet, (int*)length);
	free(packet);
	char address[72];
	// Transport remains unsupported, with the existing explicit abort.
	// Building a native packet must not silently turn this into send success.
	if (sub_420360(address, (uint16_t*)&mode)) abort();
	free(encoded);
	return 1;
}

char* nox_stats_bulk_flush_native(void) {
	char* result = (char*)(uintptr_t)nox_common_gameFlags_check_40A5C0(0x2000);
	if (result) {
		result = (char*)(uintptr_t)nox_common_gameFlags_check_40A5C0(4096);
		if (!result) {
			uint8_t* header = getMemAt(0x5D4594, 599476);
			nox_stats_players_build(&nox_stats_columns, header, getMemAt(0x5D4594, 600124), dword_5d4594_608316);
			nox_stats_pairs_build(&nox_stats_columns, getMemAt(0x5D4594, 608320), dword_5d4594_739392);
			*getMemU32Ptr(0x5D4594, 599504) = (uint32_t)time(NULL) - dword_5d4594_600116;
			nox_stats_report(&nox_stats_columns, header, 1);
			memset(getMemAt(0x5D4594, 600124), 0, 0x2000);
			memset(getMemAt(0x5D4594, 608320), 0, 0x20000);
			dword_5d4594_608316 = 0;
			dword_5d4594_739392 = 0;
			for (nox_playerInfo* player = nox_common_playerInfoGetFirst_416EA0(); player;
				 player = nox_common_playerInfoGetNext_416EE0(player)) player->field_4648 = -1;
			nox_playerInfo* host = nox_common_playerInfoFromNum_417090(31);
			if (host) sub_425F10(host);
			result = (char*)nox_common_playerInfoGetFirst_416EA0();
			while (result) {
				nox_playerInfo* player = (nox_playerInfo*)result;
				if (player->playerInd != 31) sub_425F10(player);
				result = (char*)nox_common_playerInfoGetNext_416EE0(player);
			}
		}
	}
	return result;
}

// New bridge for the startup/end roots. Mode zero forgets exactly the five
// original header slots cleared by 426150, without adding frees. The other
// three player columns and the shared count intentionally survive startup.
int nox_stats_session_report_native(int mode) {
	if (mode == 0) {
		nox_stats_columns.names = NULL;
		nox_stats_columns.ips = NULL;
		nox_stats_columns.teams = NULL;
		nox_stats_columns.classes = NULL;
		nox_stats_columns.pairs = NULL;
	}
	return nox_stats_report(&nox_stats_columns, getMemAt(0x5D4594, 599476), mode);
}

// Quest uses the same native field nodes and original serializer. These new
// bridges leave the existing online packet/bulk/start bodies unchanged.
void* nox_stats_field_native_prepend(void* head, const char* name, uint16_t type,
									uint16_t length, const void* data) {
	nox_stats_field_native* field = head;
	nox_stats_field_add(&field, name, type, length, data);
	return field;
}

uint16_t* nox_stats_packet_native_finish(void* head, uint32_t* length, uint32_t* sequence) {
	uint16_t* packet = nox_stats_serialize(head, length);
	// GAME.EXE 42B810 increments the full DWORD after serialization and before
	// destroying the list, including on zero/signed-negative player counts.
	++*sequence;
	nox_stats_fields_destroy(head);
	return packet;
}
