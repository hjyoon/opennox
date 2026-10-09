#include <arpa/inet.h>
#include <assert.h>
#include <inttypes.h>
#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

// The standalone fixture uses the original fields accessed by this root.
// The real legacy package also compiles the same source with native defs.h.
typedef struct {
	uint8_t prefix[2064];
	uint8_t playerInd;
	uint8_t rest[4648 - 2065];
	int32_t field_4648;
} nox_playerInfo;

typedef struct { uint32_t count, guard; uint64_t tail; } scalar_box;
static scalar_box *player_count, *pair_count, *length_box;
static uint32_t dword_5d4594_600116, dword_5d4594_741332;
#define dword_5d4594_608316 player_count->count
#define dword_5d4594_739392 pair_count->count
static uint32_t pair_length, sequence, player_index, flags, clock_calls, cases;
static uint8_t *header, *records, *pairs, *tags;
static nox_playerInfo* players;
static uint32_t live_players, variant, allocations;
static int node_pending, fault_recording, json_mode, transport;
static const char* failure;
static char events[32768];
static size_t event_size;
static uint8_t* captured;
static uint32_t captured_length;
static uint8_t empty_name[1];

static void event(char value) {
	assert(event_size + 1 < sizeof(events));
	events[event_size++] = value;
	events[event_size] = 0;
	if (fault_recording) { fputc(value, stderr); fflush(stderr); }
}
static void* high_alloc(size_t bytes) {
	void* pointer = calloc(1, bytes ? bytes : 1);
	assert(pointer && (sizeof(void*) == 4 || (uintptr_t)pointer > UINT32_MAX));
	return pointer;
}
static void put32(uint8_t* to, uint32_t value) { memcpy(to, &value, 4); }
static void put16(uint8_t* to, uint16_t value) { memcpy(to, &value, 2); }
static uint32_t read32(const uint8_t* from) { uint32_t value; memcpy(&value, from, 4); return value; }

static void* getMemAt(uint32_t base, uint32_t offset) {
	if (base == 0x587000) {
		assert(offset >= 71480 && offset < 71944);
		return tags + offset - 71480;
	}
	assert(base == 0x5d4594);
	if (offset == 599476) return header;
	if (offset == 600124) return records;
	if (offset == 608320) return pairs;
	assert(offset == 741688);
	return empty_name;
}
static uint32_t* getMemU32Ptr(uint32_t base, uint32_t offset) {
	assert(base == 0x5d4594);
	if (offset == 599504) return (uint32_t*)(header + 28);
	if (offset == 741308) return &pair_length;
	if (offset == 741312) return &length_box->count;
	if (offset == 741668) return &sequence;
	assert(offset == 741660);
	return &player_index;
}
static uint8_t* getMemU8Ptr(uint32_t base, uint32_t offset) { return getMemAt(base, offset); }
static int16_t* getMemI16Ptr(uint32_t base, uint32_t offset) { return (int16_t*)getMemU32Ptr(base, offset); }

static void* fixture_calloc(size_t count, size_t size, const char* caller) {
	size_t bytes = count * size;
	char kind = 'A';
	if (!strcmp(caller, "nox_stats_serialize")) kind = 'P';
	else if (!strcmp(caller, "nox_stats_field_add")) {
		if (!node_pending) {
			assert(count == 1 && bytes == 8 + 2 * sizeof(void*));
			kind = 'N';
			node_pending = 1;
		} else { kind = 'D'; node_pending = 0; }
	}
	event(kind);
	allocations++;
	if (failure && ((!strcmp(failure, "names_alloc") && allocations == 1) ||
		(!strcmp(failure, "pair_alloc") && kind == 'A') ||
		(!strcmp(failure, "node_alloc") && kind == 'N') ||
		(!strcmp(failure, "data_alloc") && kind == 'D') ||
		(!strcmp(failure, "packet_alloc") && kind == 'P'))) return NULL;
	return high_alloc(bytes);
}
static void fixture_free(void* pointer) { event(pointer ? 'F' : 'Z'); free(pointer); }
static time_t fixture_time(time_t* at) {
	assert(at == NULL);
	event('T');
	return (time_t)(0x12345678u + clock_calls++);
}
static bool nox_common_gameFlags_check_40A5C0(uint32_t flag) {
	event(flag == 0x2000 ? 'G' : 'Q');
	return (flags & flag) != 0;
}
static nox_playerInfo* nox_common_playerInfoGetFirst_416EA0(void) {
	event('L');
	return live_players ? players : NULL;
}
static nox_playerInfo* nox_common_playerInfoGetNext_416EE0(nox_playerInfo* player) {
	event('I');
	ptrdiff_t index = player - players;
	assert(index >= 0 && (uint32_t)index < live_players);
	return (uint32_t)(index + 1) < live_players ? player + 1 : NULL;
}
static nox_playerInfo* nox_common_playerInfoFromNum_417090(int number) {
	event('H');
	assert(number == 31);
	for (uint32_t i = 0; i < live_players; ++i) if (players[i].playerInd == number) return players + i;
	return NULL;
}
static void sub_425F10(nox_playerInfo* player) {
	event('R');
	assert(player->field_4648 == -1);
	player->field_4648 = (int32_t)dword_5d4594_608316;
	records[dword_5d4594_608316 * 32u] = (uint8_t)(player->playerInd + 1);
	dword_5d4594_608316++;
}
static int sub_420360(char* address, uint16_t* port) {
	event('B');
	address[0] = 0;
	*port = 0;
	return transport;
}
static uint16_t* sub_42A8B0(uint8_t* packet, int* length) {
	event('E');
	assert(packet && (uintptr_t)packet > UINT32_MAX);
	assert(length == (int*)&length_box->count);
	free(captured);
	captured_length = (uint32_t)*length;
	captured = high_alloc(captured_length);
	memcpy(captured, packet, captured_length);
	*length = 16;
	return high_alloc(16);
}

// PRODUCTION_HEADER
#define calloc(count, size) fixture_calloc(count, size, __func__)
#define free fixture_free
#define time fixture_time
// PRODUCTION_NATIVE_BODY
#undef time
#undef free
#undef calloc

// PRODUCTION_ROOT

static void reset(uint32_t choice) {
	variant = choice;
	flags = 0x2000;
	clock_calls = allocations = 0;
	events[0] = 0; event_size = 0;
	node_pending = 0;
	sequence = choice ? 0xfffffffeu : 255u;
	player_index = 0x5a5aa5a5u;
	pair_length = 0;
	dword_5d4594_600116 = 0xfedcba98u;
	memset(header, 0xc7, 656);
	memset(records, 0xa5, 8208);
	memset(pairs, 0x6b, 131088);
	player_count->guard = pair_count->guard = length_box->guard = 0xa5c3e791u;
	player_count->tail = pair_count->tail = length_box->tail = UINT64_C(0x12345678fedcba98);
	length_box->count = 0x80000001u;
	for (uint32_t i = 0; i < 24; ++i) put32(header + i * 4, (i * 0x1234567u) ^ (0x89abcdefu + choice));
	put16(header + 6, 0);
	memset(header + 96, 0, 256);
	memset(header + 352, 0, 256);
	strcpy((char*)header + 96, choice ? "map-x" : "m");
	strcpy((char*)header + 352, choice ? "server-name" : "s");
	memset(header + 608, 0, 32);
	memset(tags, 0, 464);
	const char* names[] = {"CNTL", "LGL?", "IPL?", "CLL?", "CNL?", "CMP?", "DUR?", "PAR?"};
	for (uint32_t i = 0; i < 8; ++i) memcpy(tags + i * 8, names[i], 5);
	memcpy(tags + 392, "PLRS", 5); // 71872
	memcpy(tags + 416, "PLRS", 5); // 71896
	memcpy(tags + 424, "KILS", 5); // 71904
	memcpy(tags + 448, "PLRS", 5); // 71928
	memcpy(tags + 456, "KILS", 5); // 71936
}
static void guards(void) {
	assert(player_count->guard == 0xa5c3e791u && pair_count->guard == 0xa5c3e791u && length_box->guard == 0xa5c3e791u);
	assert(player_count->tail == UINT64_C(0x12345678fedcba98) &&
		pair_count->tail == UINT64_C(0x12345678fedcba98) && length_box->tail == UINT64_C(0x12345678fedcba98));
	for (uint32_t i = 640; i < 656; ++i) assert(header[i] == 0xc7);
	for (uint32_t i = 8192; i < 8208; ++i) assert(records[i] == 0xa5);
	for (uint32_t i = 131072; i < 131088; ++i) assert(pairs[i] == 0x6b);
	for (uint32_t i = 608; i < 640; ++i) assert(header[i] == 0);
}
static void records_fill(uint32_t count) {
	for (uint32_t i = 0; i < count; ++i) {
		uint8_t* record = records + i * 32;
		for (uint32_t j = 0; j < 9; ++j) record[j] = (uint8_t)('a' + (i + j) % 26);
		record[9] = 0;
		put32(record + 12, 0x80123456u ^ (i * 0x10203u));
		put32(record + 16, 0xabcdef00u + i);
		record[20] = (uint8_t)(i + 0x80u);
		record[21] = (uint8_t)(i ^ 0xfeu);
		put32(record + 24, 0xf0123456u + i);
		record[28] = (uint8_t)(i ^ 0x81u);
	}
}
static void columns_release(nox_stats_columns_native* columns) {
	for (uint32_t i = 0; i < dword_5d4594_741332; ++i) free(columns->names[i]);
	free(columns->names); free(columns->ips); free(columns->teams); free(columns->classes);
	free(columns->active); free(columns->durations); free(columns->participants); free(columns->pairs);
	memset(columns, 0, sizeof(*columns));
	dword_5d4594_741332 = 0;
}
static void print_hex(const uint8_t* data, uint32_t length) {
	for (uint32_t i = 0; i < length; ++i) printf("%02x", data[i]);
}

static void packet_case(int mode, uint16_t count, int16_t kill_count, uint32_t seq, uint32_t choice) {
	reset(choice);
	records_fill(count);
	nox_stats_columns_native columns = {0};
	nox_stats_players_build(&columns, header, records, count);
	nox_stats_pairs_build(&columns, pairs, 32768);
	put16(header + 6, count);
	sequence = seq;
	events[0] = 0; event_size = 0;
	uint16_t* result = nox_stats_packet_build(&columns, header, mode, kill_count, &length_box->count);
	guards();
	assert(ntohs(result[0]) == (uint16_t)length_box->count && result[1] == 0);
	if (mode == 0) assert(sequence == 0);
	else if (mode == 1 || mode == 2) assert(sequence == seq + 1u);
	else assert(sequence == seq);
	if (json_mode) {
		printf("{\"mode\":%d,\"count\":%u,\"pairs\":%d,\"sequence\":%u,\"variant\":%u,\"result_sequence\":%u,\"length\":%u,\"events\":\"%s\",\"output\":\"",
			mode, count, kill_count, seq, choice, sequence, length_box->count, events);
		print_hex((uint8_t*)result, length_box->count); puts("\"}");
	}
	free(result); columns_release(&columns); cases++;
}

static void flush_case(uint32_t count, uint32_t kills, uint32_t game_flags, uint32_t nplayers, uint32_t choice) {
	reset(choice);
	records_fill(count);
	for (uint32_t i = 0; i < 131072; ++i) pairs[i] = (uint8_t)(i * 37u + choice);
	live_players = nplayers;
	for (uint32_t i = 0; i < live_players; ++i) {
		players[i].playerInd = (uint8_t)(i == 1 ? 31 : i * 7u);
		players[i].field_4648 = (int32_t)(42u + i);
	}
	player_count->count = count; pair_count->count = kills;
	flags = game_flags;
	char* result = nox_xxx_net_4263C0();
	assert((uintptr_t)result == (game_flags & 0x2000 ? !!(game_flags & 0x1000) : 0));
	guards();
	if ((flags & 0x2000) && !(flags & 0x1000)) {
		assert(player_count->count == nplayers && pair_count->count == 0);
		assert(length_box->count == 16 && dword_5d4594_741332 == count);
		assert(pair_length == kills && clock_calls == count + 1);
		for (uint32_t i = 0; i < live_players; ++i) {
			int32_t expected = players[i].playerInd == 31 ? 0 : (int32_t)(i + (nplayers > 1 && i < 1));
			assert(players[i].field_4648 == expected);
		}
		for (uint32_t i = 0; i < 131072; ++i) assert(pairs[i] == 0);
		for (uint32_t i = 0; i < 8192; ++i) if (i % 32 != 0 || i / 32 >= nplayers) assert(records[i] == 0);
	} else {
		assert(player_count->count == count && pair_count->count == kills);
		assert(clock_calls == 0 && length_box->count == 0x80000001u);
	}
	if (json_mode) {
		printf("{\"count\":%u,\"pairs\":%u,\"flags\":%u,\"players\":%u,\"variant\":%u,\"result\":%u,\"sequence\":%u,\"duration\":%u,\"player_count\":%u,\"pair_count\":%u,\"events\":\"%s\",\"indices\":[",
			count,kills,game_flags,nplayers,choice,(unsigned)(uintptr_t)result,sequence,read32(header+28),player_count->count,pair_count->count,events);
		for (uint32_t i = 0; i < live_players; ++i) printf("%s%d", i ? "," : "", players[i].field_4648);
		printf("],\"record_prefix\":\""); print_hex(records,nplayers*32);
		printf("\",\"output\":\"");
		if ((flags & 0x2000) && !(flags & 0x1000)) print_hex(captured,captured_length);
		puts("\"}");
	}
	if (nox_stats_columns.names) columns_release(&nox_stats_columns);
	cases++;
}

static void lifecycle_snapshot(uint32_t count) {
	guards();
	if (json_mode) {
		printf("{\"count\":%u,\"pairs\":%u,\"players\":%u,\"column_count\":%u,\"clock_calls\":%u,\"events\":\"%s\",\"record_prefix\":\"",
			count, pair_length, (unsigned)nox_stats_read_u16(header + 6), dword_5d4594_741332, clock_calls, events);
		print_hex(records, 64);
		printf("\",\"output\":\"");
		for (uint32_t i = 0; i < count; ++i) print_hex((uint8_t*)nox_stats_columns.names[i], 10);
		print_hex((uint8_t*)nox_stats_columns.ips, count * 4);
		print_hex((uint8_t*)nox_stats_columns.teams, count * 4);
		print_hex(nox_stats_columns.classes, count);
		print_hex(nox_stats_columns.active, count);
		print_hex((uint8_t*)nox_stats_columns.durations, count * 4);
		print_hex(nox_stats_columns.participants, count);
		print_hex(nox_stats_columns.pairs, pair_length * 2u);
		puts("\"}");
	}
	cases++;
}

int main(int argc, char** argv) {
	header=high_alloc(656); records=high_alloc(8208); pairs=high_alloc(131088); tags=high_alloc(464);
	player_count=high_alloc(sizeof(*player_count)); pair_count=high_alloc(sizeof(*pair_count)); length_box=high_alloc(sizeof(*length_box));
	players=high_alloc(3*sizeof(*players));
	json_mode=argc>2&&!strcmp(argv[2],"--json");
	if (argc > 1 && !strncmp(argv[1],"--fault=",8)) {
		failure=argv[1]+8; fault_recording=1; reset(0);
		if (!strcmp(failure,"names_alloc")) nox_stats_players_build(&nox_stats_columns,header,records,1);
		else if (!strcmp(failure,"pair_alloc")) nox_stats_pairs_build(&nox_stats_columns,pairs,1);
		else if (!strcmp(failure,"transport")) {transport=1; nox_stats_report(&nox_stats_columns,header,0);}
		else {
			nox_stats_packet_build(&nox_stats_columns,header,0,0,!strcmp(failure,"nil_length")?NULL:&length_box->count);
		}
		return 99;
	}
	if (argc > 1 && !strcmp(argv[1],"--packet")) {
		const uint16_t counts[]={0,1,2,9,32,209,256};
		const uint32_t sequences[]={0,127,255,256,0xfffffffeu,0xffffffffu};
		for (uint32_t i=0;i<sizeof(counts)/sizeof(counts[0]);++i)
			for (int mode=0;mode<4;++mode) packet_case(mode,counts[i],(int16_t)(i%2?1:0),sequences[i%6],i%2);
		const int16_t lengths[]={32767,INT16_MIN,-1};
		for (uint32_t i=0;i<3;++i) for(int mode=1;mode<=2;++mode) packet_case(mode,0,lengths[i],0x12345678u,i%2);
	} else if (argc>1&&!strcmp(argv[1],"--flush")) {
		const uint32_t counts[]={0,1,2,9,32,128,256};
		for(uint32_t i=0;i<7;++i) for(uint32_t n=0;n<4;++n) flush_case(counts[i],i%3,0x2000,n,i%2);
		flush_case(1,1,0,3,0); flush_case(2,2,0x1000,3,0); flush_case(2,2,0x3000,3,1);
		flush_case(0,65536,0x2000,0,0);
	} else {
		reset(0);
		for(uint32_t cycle=0;cycle<4;++cycle) {
			uint32_t count=cycle%2?0:2; records_fill(count);
			events[0] = 0; event_size = 0;
			nox_stats_players_build(&nox_stats_columns,header,records,count);
			nox_stats_pairs_build(&nox_stats_columns,pairs,cycle%2?0:3);
			assert(dword_5d4594_741332==count);
			lifecycle_snapshot(count);
		}
		events[0] = 0; event_size = 0;
		assert(nox_stats_pairs_build(&nox_stats_columns,pairs,0x80000000u)==0 && pair_length==0x80000000u);
		lifecycle_snapshot(0);
		columns_release(&nox_stats_columns); guards();
	}
	if(!json_mode) printf("%u native stats bulk cases passed\n",cases);
	free(header);free(records);free(pairs);free(tags);free(player_count);free(pair_count);free(length_box);free(players);free(captured);
	return 0;
}
