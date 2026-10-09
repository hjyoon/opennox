// Reuse the already checked R62 external services, not its case expectations.
// Native startup/header/cleanup and packet code is inserted from production.
static void* getMemAt(uint32_t base, uint32_t offset);
static uint32_t* getMemU32Ptr(uint32_t base, uint32_t offset);
// BULK_FIXTURE_SERVICES

typedef struct { uint8_t prefix[60]; uint32_t field_60; } nox_team_t;
static nox_team_t* teams;
static uint32_t team_count, quest_mask;
static uint8_t *quest, *settings, *game, *ip_output, *ip_source;
static scalar_box* port_box;
static uint16_t port;
static int live_free;

static void* getMemAt(uint32_t base, uint32_t offset) {
	assert(base == 0x5d4594 || base == 0x587000);
	if (base == 0x5d4594) {
		if (offset >= 599476 && offset < 600116) return header + offset - 599476;
		if (offset >= 739396 && offset < 739976) return quest + offset - 739396;
		if (offset == 741316) return ip_output;
	}
	return bulk_getMemAt(base, offset);
}
static uint32_t* getMemU32Ptr(uint32_t base, uint32_t offset) {
	if (base == 0x5d4594) {
		if (offset >= 599476 && offset < 600116) return (uint32_t*)(header + offset - 599476);
		if (offset >= 739396 && offset < 739976) return (uint32_t*)(quest + offset - 739396);
		if (offset == 741304) return &port_box->count;
	}
	return bulk_getMemU32Ptr(base, offset);
}
static uint16_t* getMemU16Ptr(uint32_t base, uint32_t offset) { return (uint16_t*)getMemAt(base, offset); }
static uint8_t* getMemU8Ptr(uint32_t base, uint32_t offset) { return bulk_getMemU8Ptr(base, offset); }
static int16_t* getMemI16Ptr(uint32_t base, uint32_t offset) { return bulk_getMemI16Ptr(base, offset); }
static char* sub_416640(void) { event('C'); return failure && !strcmp(failure,"settings_nil") ? NULL : (char*)settings; }
static char* nox_xxx_cliGamedataGet_416590(int index) {
	assert(index == 0); event('J');
	return failure && !strcmp(failure,"game_nil") ? NULL : (char*)game;
}
static uint32_t sub_4200E0(void) { event('V'); return 0x80000123u; }
static uint32_t sub_40A770(void) { event('K'); return 0x80000003u; }
static int16_t sub_5545A0(void) { event('O'); return (int16_t)port; }
static char* sub_554230(void) { event('U'); return (char*)ip_source; }
static nox_team_t* nox_server_teamFirst_418B10(void) { event('X'); return team_count ? teams : NULL; }
static nox_team_t* nox_server_teamNext_418B60(nox_team_t* team) {
	event('Y'); ptrdiff_t index = team - teams;
	assert(index >= 0 && (uint32_t)index < team_count);
	return (uint32_t)(index + 1) < team_count ? team + 1 : NULL;
}
static void startup_free(void* pointer) {
	fixture_free(pointer);
	// The original cleanup reads the live DWORD again after every free.
	if (live_free) { dword_5d4594_741332 = 0; live_free = 0; }
}

// PRODUCTION_BULK_HEADER
// PRODUCTION_START_HEADER
#define calloc(count, size) fixture_calloc(count, size, __func__)
#define free startup_free
#define time fixture_time
// PRODUCTION_BULK_BODY
// PRODUCTION_START_BODY
// PRODUCTION_IP_ROOT
// PRODUCTION_START_ROOT
#undef time
#undef free
#undef calloc

static void* online_owned[8];
static void* quest_owned[14];
static uint32_t quest_owned_count;
static uint32_t online_mask(void) {
	return !!nox_stats_columns.names | (!!nox_stats_columns.ips << 1) |
		(!!nox_stats_columns.teams << 2) | (!!nox_stats_columns.classes << 3) |
		(!!nox_stats_columns.active << 4) | (!!nox_stats_columns.durations << 5) |
		(!!nox_stats_columns.participants << 6) | (!!nox_stats_columns.pairs << 7);
}
static uint32_t quest_present(void) {
	uint32_t mask = !!nox_stats_quest_columns.names | (!!nox_stats_quest_columns.ips << 1) |
		(!!nox_stats_quest_columns.classes << 2);
	for (uint32_t i = 0; i < 8; ++i) mask |= !!nox_stats_quest_columns.scores[i] << (i+3);
	return mask;
}
static void seed_quest(uint32_t mask, uint32_t count) {
	quest_mask = mask; quest_owned_count = 0;
	memset(&nox_stats_quest_columns, 0, sizeof(nox_stats_quest_columns));
	if (mask & 1) {
		nox_stats_quest_columns.names = high_alloc((count ? count : 1) * sizeof(char*));
		for (uint32_t i = 0; i < count; ++i) {
			if (i != 1) nox_stats_quest_columns.names[i] = quest_owned[quest_owned_count++] = high_alloc(10);
		}
		quest_owned[quest_owned_count++] = nox_stats_quest_columns.names;
	}
	if (mask & 2) nox_stats_quest_columns.ips = quest_owned[quest_owned_count++] = high_alloc(16);
	if (mask & 4) nox_stats_quest_columns.classes = quest_owned[quest_owned_count++] = high_alloc(16);
	for (uint32_t i = 0; i < 8; ++i) if (mask & (1u << (i+3)))
		nox_stats_quest_columns.scores[i] = quest_owned[quest_owned_count++] = high_alloc(16);
	dword_5d4594_741332 = count;
}
static void reset_start(uint32_t choice, uint32_t mode, uint32_t count, uint32_t mask, uint32_t names) {
	variant = choice; flags = mode; team_count = count;
	clock_calls = allocations = 0; node_pending = 0; live_free = 0;
	sequence = choice ? 0xfffffffeu : 255u; player_index = 0x5a5aa5a5u;
	pair_length = 0; dword_5d4594_600116 = 0xfedcba98u;
	player_count->count = 19; pair_count->count = 5; length_box->count = 0x80000001u;
	player_count->guard = pair_count->guard = length_box->guard = port_box->guard = 0xa5c3e791u;
	player_count->tail = pair_count->tail = length_box->tail = port_box->tail = UINT64_C(0x12345678fedcba98);
	memset(header,0xc7,656); memset(quest,0xd6,596);
	put32(header+20,0x89abcdefu+choice);
	memset(records,0xa5,8208); memset(pairs,0x6b,131088);
	memset(header+608,0,32); memset(quest+536,0,44);
	for (uint32_t i = 0; i < 128; ++i) { settings[i]=(uint8_t)(i*13+7+choice); game[i]=(uint8_t)(i*17+3+choice); }
	memset(settings+128,0xa8,16); memset(game+128,0xb9,16);
	memset(game,0,24);
	memcpy(game,choice ? "map-1234567" : "m",choice ? 9 : 2);
	memcpy(game+9,choice ? "server-12345678" : "s",choice ? 15 : 2);
	put16(game+54,0x8001); game[56]=0xfe;
	put16(settings+105,0xfedc); put16(settings+107,0x8001);
	memset(ip_source,0,32); memcpy(ip_source,choice ? "123456789012345678901" : "203.0.113.250",choice ? 21 : 13);
	memset(ip_output,0xe4,32);
	memset(tags,0,464); memcpy(tags+392,"PLRS",5); memcpy(tags+416,"PLRS",5);
	memcpy(tags+424,"KILS",5); memcpy(tags+448,"PLRS",5); memcpy(tags+456,"KILS",5);
	for (uint32_t i = 0; i < count; ++i) teams[i].field_60 = choice || i!=1 ? 0x80000001u : 0;
	for (uint32_t i = 0; i < 3; ++i) { players[i].playerInd = i==1 ? 31 : (uint8_t)(i*7); players[i].field_4648 = (int32_t)(42+i); }
	for (uint32_t i = 0; i < 8; ++i) online_owned[i] = high_alloc(16);
	nox_stats_columns.names=online_owned[0]; nox_stats_columns.ips=online_owned[1];
	nox_stats_columns.teams=online_owned[2]; nox_stats_columns.classes=online_owned[3];
	nox_stats_columns.active=online_owned[4]; nox_stats_columns.durations=online_owned[5];
	nox_stats_columns.participants=online_owned[6]; nox_stats_columns.pairs=online_owned[7];
	seed_quest(mask,names);
	free(captured); captured=NULL; captured_length=0;
	events[0]=0; event_size=0;
}
static void guards_start(void) {
	assert(player_count->guard==0xa5c3e791u && pair_count->guard==0xa5c3e791u && length_box->guard==0xa5c3e791u && port_box->guard==0xa5c3e791u);
	assert(player_count->tail==UINT64_C(0x12345678fedcba98) && pair_count->tail==player_count->tail && length_box->tail==player_count->tail && port_box->tail==player_count->tail);
	for (uint32_t i=640;i<656;++i) assert(header[i]==0xc7);
	for (uint32_t i=580;i<596;++i) assert(quest[i]==0xd6);
	for (uint32_t i=128;i<144;++i) { assert(settings[i]==0xa8); assert(game[i]==0xb9); }
	for (uint32_t i=16;i<32;++i) assert(ip_output[i]==0xe4);
	for (uint32_t i=8192;i<8208;++i) assert(records[i]==0xa5);
	for (uint32_t i=131072;i<131088;++i) assert(pairs[i]==0x6b);
	assert(read32(header+608)==0 && read32(header+636)==0);
}
static void print_hex(const uint8_t* data,uint32_t length) { for (uint32_t i=0;i<length;++i) printf("%02x",data[i]); }
static void print_start(const char* kind,uint32_t mode,uint32_t choice,uint32_t count,uint32_t names,int last) {
	guards_start(); cases++;
	if (json_mode) {
		printf("{\"kind\":\"%s\",\"flags\":%u,\"variant\":%u,\"teams\":%u,\"mode\":%u,\"players\":%u,\"mask\":%u,\"names\":%u,\"live\":%d,\"result\":%d,\"port\":%u,\"port_value\":%u,\"sequence\":%u,\"clock\":%u,\"start\":%u,\"count\":%u,\"pairs\":%u,\"column_count\":%u,\"online_mask\":%u,\"quest_mask\":%u,\"events\":\"%s\",\"indices\":[%d,%d,%d],\"ip\":\"",
			kind,flags,choice,count,mode,live_players,quest_mask,names,live_free,last,port,port_box->count,sequence,clock_calls,dword_5d4594_600116,player_count->count,pair_count->count,dword_5d4594_741332,online_mask(),quest_present(),events,players[0].field_4648,players[1].field_4648,players[2].field_4648);
		print_hex(ip_output,16); printf("\",\"header\":\""); print_hex(header,640);
		printf("\",\"quest\":\""); print_hex(quest,580); printf("\",\"record_prefix\":\""); print_hex(records,96);
		printf("\",\"output\":\""); print_hex(captured,captured_length); printf("\"}\n");
	}
	for (uint32_t i=0;i<8;++i) free(online_owned[i]);
	// Cleared columns were already freed by production. On untouched paths
	// (or a live-count mutation) release only the still-owned allocations.
	if (quest_present()) { for (uint32_t i=0;i<quest_owned_count;++i) free(quest_owned[i]); }
}
int main(int argc,char** argv) {
	assert(argc>=2); json_mode=argc>2 && !strcmp(argv[2],"--json");
	header=high_alloc(656); quest=high_alloc(596); settings=high_alloc(144); game=high_alloc(144);
	records=high_alloc(8208); pairs=high_alloc(131088); tags=high_alloc(464); players=high_alloc(3*sizeof(*players)); teams=high_alloc(257*sizeof(*teams));
	player_count=high_alloc(sizeof(*player_count)); pair_count=high_alloc(sizeof(*pair_count)); length_box=high_alloc(sizeof(*length_box)); port_box=high_alloc(sizeof(*port_box));
	ip_output=high_alloc(32); ip_source=high_alloc(32);
	const uint32_t flag_cases[]={0,0x1000,0x2000,0x3000};
	const uint16_t mode_cases[]={0,0x100,0x20,0x40,0x10,0x400,0x4320,0x57f,0x8000,0xc000};
	const uint32_t team_cases[]={0,3,256,257};
	port=0x8000;
	if (!strcmp(argv[1],"--header")) {
		live_players=0;
		for (uint32_t f=0;f<4;++f) for (uint32_t v=0;v<2;++v) for (uint32_t m=0;m<10;++m) for (uint32_t c=0;c<4;++c) {
			reset_start(v,flag_cases[f],team_cases[c],0,7); put16(game+52,mode_cases[m]);
			memset(&nox_stats_columns,0,sizeof(nox_stats_columns));
			nox_stats_start_header_native(); print_start("header",mode_cases[m],v,team_cases[c],7,0);
		}
	} else if (!strcmp(argv[1],"--root")) {
		const uint16_t ports[]={32767,32768,65535}; const uint32_t player_cases[]={0,1,3};
		for (uint32_t f=0;f<4;++f) for (uint32_t v=0;v<2;++v) for (uint32_t p=0;p<3;++p) for (uint32_t n=0;n<3;++n) {
			live_players=player_cases[n]; port=ports[p]; reset_start(v,flag_cases[f],v ? 257 : 3,2047,3); put16(game+52,v ? 0x4320 : 0x20);
			sub_426060(); print_start("root",v ? 0x4320 : 0x20,v,v ? 257 : 3,3,0);
		}
	} else if (!strcmp(argv[1],"--cleanup")) {
		live_players=0; const uint32_t masks[]={0,1,2,4,8,16,32,64,128,256,512,1024,2047};
		const uint32_t counts[]={0,1,3};
		for (uint32_t m=0;m<13;++m) for (uint32_t n=0;n<3;++n) {
			reset_start(0,0x1000,0,masks[m],counts[n]);
			void* result=nox_stats_quest_columns_clear_native(); assert(quest_present()==0);
			print_start("cleanup",0,0,0,counts[n],result!=NULL);
		}
		for (uint32_t v=0;v<2;++v) {
			reset_start(v,0x1000,0,2047,3); live_free=1;
			void* result=nox_stats_quest_columns_clear_native(); assert(dword_5d4594_741332==0);
			// The early live-count stop deliberately leaves name[2] allocated.
			free(quest_owned[1]); print_start("cleanup-live",0,v,0,3,result!=NULL);
		}
	} else if (!strncmp(argv[1],"--fault=",8)) {
		failure=argv[1]+8; live_players=0; reset_start(0,!strcmp(failure,"quest_names_nil") ? 0x1000 : 0x2000,3,0,3); put16(game+52,0x20);
		if (!strcmp(failure,"quest_names_nil")) nox_stats_quest_columns.names=(char**)(uintptr_t)1;
		if (!strcmp(failure,"transport")) transport=1;
		fault_recording=1; sub_426060(); return 91;
	} else return 2;
	if (!json_mode) printf("%u native stats start %s cases passed\n",cases,argv[1]+2);
	free(captured); free(header); free(quest); free(settings); free(game); free(records); free(pairs); free(tags); free(players); free(teams);
	free(player_count); free(pair_count); free(length_box); free(port_box); free(ip_output); free(ip_source);
	return 0;
}
