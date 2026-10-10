// Reuse the R62 declared services, not its expected results.
static void* getMemAt(uint32_t base, uint32_t offset);
static uint32_t* getMemU32Ptr(uint32_t base, uint32_t offset);
typedef struct {
	uint8_t prefix[2064], playerInd, gap[31];
	char field_2096[12];
	uint8_t to_info[77];
	struct { uint8_t prefix[66], playerClass, tail[30]; } info;
	uint8_t to_stats[4648-2282];
	int32_t field_4648;
	uint32_t field_4652, field_4656, field_4660, field_4664, field_4668, field_4672;
	uint32_t field_4676, field_4680, field_4684, field_4688;
	uint8_t to_state[4792-4692];
	uint32_t field_4792;
	uint8_t tail[32];
} nox_playerInfo;
// BULK_FIXTURE_SERVICES

static uint8_t* quest;
static scalar_box* quest_length;
static uint32_t quest_sequence, quest_index, count_value, initial_sequence;
static uint32_t ip_args[8], score_args[8], tracking_args[8][2], ip_count, score_count, tracking_count;
static uint32_t column_sizes[40000], column_size_count;
static int mutate_count, mutate_ip, mutate_free;
static void* owned[40000];
static uint32_t owned_count;
static void* end_calloc(size_t count,size_t size,const char* caller) {
	if (!strcmp(caller,"nox_stats_quest_players_build")) {
		assert(column_size_count < 40000);
		// Only the first pointer-array stride is normalized back to PE32.
		uint32_t bytes=(uint32_t)(column_size_count ? count*size : count*4u);
		column_sizes[column_size_count++]=bytes;
	}
	void* pointer=fixture_calloc(count,size,caller);
	if (pointer) { assert(owned_count<40000); owned[owned_count++]=pointer; }
	if (mutate_count && !strcmp(caller,"nox_stats_quest_players_build")) {
		put16(quest,0); mutate_count=0;
	}
	return pointer;
}
static void end_free(void* pointer) {
	fixture_free(pointer);
	for (uint32_t i=0;i<owned_count;++i) if (owned[i]==pointer) owned[i]=NULL;
	if (mutate_free) { quest_sequence=0xfedcba98u;mutate_free=0; }
}
static void* getMemAt(uint32_t base,uint32_t offset) {
	if (base==0x587000) { assert(offset>=71480 && offset<72016); return tags+offset-71480; }
	if (base==0x5d4594) {
		if (offset>=739396 && offset<739976) return quest+offset-739396;
		if (offset>=599476 && offset<600116) return header+offset-599476;
	}
	return bulk_getMemAt(base,offset);
}
static uint32_t* getMemU32Ptr(uint32_t base,uint32_t offset) {
	if (base==0x5d4594) {
		if (offset>=739396 && offset<739976) return (uint32_t*)(quest+offset-739396);
		if (offset==741300) return &quest_length->count;
		if (offset==741672) return &quest_sequence;
		if (offset==741664) return &quest_index;
	}
	return bulk_getMemU32Ptr(base,offset);
}
static uint8_t* getMemU8Ptr(uint32_t base,uint32_t offset) { return bulk_getMemU8Ptr(base,offset); }
static int16_t* getMemI16Ptr(uint32_t base,uint32_t offset) { return bulk_getMemI16Ptr(base,offset); }
static int nox_xxx_player_4E3CE0(void) { event('K');return (int)count_value; }
static uint32_t nox_xxx_net_getIP_554200(uint32_t connection) {
	assert(ip_count<8);ip_args[ip_count++]=connection;event('U');
	if (mutate_ip) { players[0].field_4688=0xfedcba98u;mutate_ip=0; }
	return 0x80123456u^connection;
}
static uint32_t sub_4D6540(uint32_t index) {
	assert(score_count<8);score_args[score_count++]=index;event('V');return 0xfedcba98u^index;
}
static int sub_425E90(nox_playerInfo* player,char state) {
	ptrdiff_t index=player-players;
	assert(index>=0 && (uint32_t)index<live_players && state==1 && tracking_count<8);
	tracking_args[tracking_count][0]=(uint32_t)index;tracking_args[tracking_count++][1]=(uint8_t)state;event('W');return 0;
}
static uint16_t* sub_42A8B0(uint8_t* packet,int* length) {
	if (length==(int*)&length_box->count) return bulk_envelope(packet,length);
	assert(length==(int*)&quest_length->count);
	event('E'); assert(packet && (sizeof(void*)==4 || (uintptr_t)packet>UINT32_MAX));
	free(captured);captured_length=(uint32_t)*length;captured=high_alloc(captured_length);
	memcpy(captured,packet,captured_length);*length=16;return high_alloc(16);
}
// PRODUCTION_BULK_HEADER
// PRODUCTION_START_HEADER
// PRODUCTION_END_HEADER
static nox_stats_quest_columns_native nox_stats_quest_columns;
// PRODUCTION_GETTER
#define calloc(count,size) end_calloc(count,size,__func__)
#define free end_free
#define time fixture_time
// PRODUCTION_BULK_BODY
// PRODUCTION_END_BODY
// PRODUCTION_ROOT
#undef time
#undef free
#undef calloc

static void* seed_alloc(size_t size) {
	void* pointer=high_alloc(size);assert(owned_count<40000);owned[owned_count++]=pointer;return pointer;
}
static void seed_columns(uint32_t count,int is_quest) {
	char** names=seed_alloc((count?count:1)*sizeof(*names));
	for (uint32_t i=0;i<count;++i) {
		names[i]=seed_alloc(10);
		for (uint32_t j=0;j<9;++j) names[i][j]=(char)(97+(i+j)%26);
	}
	if (is_quest) nox_stats_quest_columns.names=names;else nox_stats_columns.names=names;
	for (uint32_t column=1;column<(is_quest?11u:7u);++column) {
		int byte_column=column==2 || (!is_quest && (column==4 || column==6));
		void* pointer=seed_alloc((count?count:1)*(byte_column?1u:4u));
		for (uint32_t i=0;i<count;++i) {
			if (byte_column) ((uint8_t*)pointer)[i]=(uint8_t)(0x80+column+i);
			else ((uint32_t*)pointer)[i]=0x80123456u+column*100+i;
		}
		if (is_quest) {
			if (column==1) nox_stats_quest_columns.ips=pointer;
			else if (column==2) nox_stats_quest_columns.classes=pointer;
			else nox_stats_quest_columns.scores[column-3]=pointer;
		} else {
			switch(column) {
			case 1:nox_stats_columns.ips=pointer;break;case 2:nox_stats_columns.classes=pointer;break;
			case 3:nox_stats_columns.teams=pointer;break;case 4:nox_stats_columns.active=pointer;break;
			case 5:nox_stats_columns.durations=pointer;break;case 6:nox_stats_columns.participants=pointer;break;
			}
		}
	}
}
static void reset_end(uint32_t mode,uint32_t choice,uint32_t count,uint32_t sequ) {
	for (uint32_t i=0;i<owned_count;++i) free(owned[i]);owned_count=0;
	free(captured);captured=NULL;captured_length=0;
	flags=mode;variant=choice;count_value=count;initial_sequence=sequ;live_players=count?3u:0u;
	clock_calls=allocations=column_size_count=ip_count=score_count=tracking_count=0;
	node_pending=mutate_count=mutate_ip=mutate_free=0;
	memset(&nox_stats_columns,0,sizeof(nox_stats_columns));memset(&nox_stats_quest_columns,0,sizeof(nox_stats_quest_columns));
	memset(header,0xc7,656);memset(records,0xa5,8208);memset(pairs,0x6b,131088);memset(quest,0xd6,596);
	for (uint32_t i=0;i<24;++i) put32(header+i*4,(i*0x1234567u)^(0x89abcdefu+choice));
	memset(header+96,0,256);memset(header+352,0,256);memset(header+608,0,32);memset(quest+536,0,44);
	strcpy((char*)header+96,choice?"map-x":"m");strcpy((char*)header+352,choice?"server-name":"s");
	memset(quest+24,0,512);strcpy((char*)quest+24,choice?"quest-map":"q");strcpy((char*)quest+280,choice?"quest-name":"n");
	const char* table[]={"CNTL","LGL?","IPL?","CLL?","CNL?","CMP?","DUR?","PAR?","LGLS","IPLS","CLLS","CSTS","HSTS","MKLS","ANKS","GNDS","SECS","BPTS","SCRS","MXPL","IDNO","GSKU","GSTY","CLGM","LIMT","TLMT","RSTC","MINE","MAXE","MINP","MAXP","VIDM","SVRS","NTMS","SCEN","GNAM","SPL1","SPL2","SPL3","ARMR","WPN1","WPN2","WPN3","STAF","DURA","FINI","TRNY","SEQU","ENDF","PLRS","SEQU","ENDF","PLRS","KILS","SEQU","ENDF","PLRS","KILS","IDNO","GSKU","GSTY","SCEN","GNAM","DURA","TRNY","PLRS","SEQU"};
	memset(tags,0,536);for(uint32_t i=0;i<67;++i) memcpy(tags+i*8,table[i],4);
	for (uint32_t i=0;i<3;++i) {
		nox_playerInfo* p=players+i;memset(p,0,sizeof(*p));
		const uint8_t indices[]={255,128,31};p->playerInd=choice?indices[i]:(i==2?31:(uint8_t)(i*7));
		for(uint32_t j=0;j<9;++j) p->field_2096[j]=(char)(97+(i+j)%26);
		p->info.playerClass=(uint8_t)(128+i);p->field_4648=(int32_t)(42+i);
		p->field_4660=0x80123456u+4660+i;p->field_4664=0x80123456u+4664+i;
		p->field_4668=0x80123456u+4668+i;p->field_4672=0x80123456u+4672+i;
		p->field_4684=0x80123456u+4684+i;p->field_4688=0x80123456u+4688+i;
		p->field_4792=i<count?(choice?0x80000001u:1u):0u;
	}
	quest_sequence=sequ;sequence=choice?0xfffffffeu:255u;quest_index=player_index=0x5a5aa5a5u;
	dword_5d4594_741332=7;pair_length=0;dword_5d4594_600116=0xfedcba98u;
	length_box->count=quest_length->count=0x80000001u;
	length_box->guard=quest_length->guard=0xa5c3e791u;length_box->tail=quest_length->tail=UINT64_C(0x12345678fedcba98);
	seed_columns(2,0);put16(header+6,2);events[0]=0;event_size=0;
}
static void print_hex(const uint8_t* data,uint32_t length) { for(uint32_t i=0;i<length;++i) printf("%02x",data[i]); }
static void print_end(const char* kind) {
	cases++;
	assert(length_box->guard==0xa5c3e791u && quest_length->guard==0xa5c3e791u && length_box->tail==UINT64_C(0x12345678fedcba98) && quest_length->tail==length_box->tail);
	for(uint32_t i=640;i<656;++i) assert(header[i]==0xc7);
	for(uint32_t i=580;i<596;++i) assert(quest[i]==0xd6);
	assert(read32(header+608)==0 && read32(header+636)==0 && read32(quest+536)==0 && read32(quest+576)==0);
	if(!json_mode)return;
	uint32_t mask=!!nox_stats_quest_columns.names|(!!nox_stats_quest_columns.ips<<1)|(!!nox_stats_quest_columns.classes<<2);
	for(uint32_t i=0;i<8;++i)mask|=!!nox_stats_quest_columns.scores[i]<<(i+3);
	printf("{\"kind\":\"%s\",\"flags\":%u,\"variant\":%u,\"count\":%u,\"initial_sequence\":%u,\"column_count\":%u,\"mask\":%u,\"clock\":%u,\"sequence\":%u,\"quest_sequence\":%u,\"index\":%u,\"online_length\":%u,\"quest_length\":%u,\"ip_args\":[",kind,flags,variant,count_value,initial_sequence,dword_5d4594_741332,mask,clock_calls,sequence,quest_sequence,quest_index,length_box->count,quest_length->count);
	for(uint32_t i=0;i<ip_count;++i)printf("%s%u",i?",":"",ip_args[i]);printf("],\"score_args\":[");
	for(uint32_t i=0;i<score_count;++i)printf("%s%u",i?",":"",score_args[i]);printf("],\"tracking_args\":[");
	for(uint32_t i=0;i<tracking_count;++i)printf("%s[%u,%u]",i?",":"",tracking_args[i][0],tracking_args[i][1]);printf("],\"column_sizes\":[");
	for(uint32_t i=0;i<column_size_count;++i)printf("%s%u",i?",":"",column_sizes[i]);printf("],\"columns\":[");
	int16_t count;memcpy(&count,quest,2);
	for(uint32_t i=0;i<(uint32_t)(count>0?(count<257?count:257):0);++i) {
		printf("%s[\"",i?",":"");print_hex((uint8_t*)nox_stats_quest_columns.names[i],10);
		printf("\",%u,%u",nox_stats_quest_columns.ips[i],nox_stats_quest_columns.classes[i]);
		for(uint32_t c=0;c<8;++c)printf(",%u",nox_stats_quest_columns.scores[c][i]);printf("]");
	}
	printf("],\"header\":\"");print_hex(header,640);printf("\",\"quest\":\"");print_hex(quest,580);
	printf("\",\"events\":\"%s\",\"output\":\"",events);print_hex(captured,captured_length);printf("\"}\n");
}
int main(int argc,char** argv) {
	assert(argc>=2);json_mode=argc>2&&!strcmp(argv[2],"--json");
	header=high_alloc(656);quest=high_alloc(596);records=high_alloc(8208);pairs=high_alloc(131088);tags=high_alloc(536);players=high_alloc(3*sizeof(*players));
	player_count=high_alloc(sizeof(*player_count));pair_count=high_alloc(sizeof(*pair_count));length_box=high_alloc(sizeof(*length_box));quest_length=high_alloc(sizeof(*quest_length));
	const uint32_t flag_cases[]={0,0x1000,0x2000,0x3000}, sequences[]={255,256,0xffffffffu};
	if(!strcmp(argv[1],"--root")) {
		const uint32_t counts[]={0,1,3};
		for(uint32_t f=0;f<4;++f)for(uint32_t v=0;v<2;++v)for(uint32_t n=0;n<3;++n)for(uint32_t s=0;s<3;++s) {
			reset_end(flag_cases[f],v,counts[n],sequences[s]);assert(sub_4264D0()==1);print_end("root");
		}
	}else if(!strcmp(argv[1],"--builder")) {
		const uint32_t counts[]={0,1,3,32767,32768,65535,0x10001};
		for(uint32_t v=0;v<2;++v)for(uint32_t n=0;n<7;++n) {
			reset_end(0x1000,v,counts[n],256);live_players=counts[n]==1||counts[n]==3?3u:0u;
			nox_stats_quest_players_build(&nox_stats_quest_columns,quest);print_end("builder");
		}
		for(uint32_t m=0;m<2;++m) {
			reset_end(0x1000,0,3,256);mutate_count=m==0;mutate_ip=m==1;if(m==0)live_players=0;
			nox_stats_quest_players_build(&nox_stats_quest_columns,quest);print_end(m==0?"builder-count":"builder-ip");
		}
	}else if(!strcmp(argv[1],"--packet")) {
		const uint32_t counts[]={0,1,3,208,256,257,32768,65535};
		for(uint32_t v=0;v<2;++v)for(uint32_t n=0;n<8;++n)for(uint32_t s=0;s<3;++s) {
			reset_end(0x1000,v,counts[n],sequences[s]);seed_columns(counts[n]<257?counts[n]:257,1);put16(quest,(uint16_t)counts[n]);events[0]=0;event_size=0;
			uint16_t* packet=nox_stats_quest_packet_build(&nox_stats_quest_columns,quest,&quest_length->count);
			captured_length=quest_length->count;captured=high_alloc(captured_length);memcpy(captured,packet,captured_length);print_end("packet");
		}
		reset_end(0x1000,0,0,0xffffffffu);put16(quest,0);mutate_free=1;
		uint16_t* packet=nox_stats_quest_packet_build(&nox_stats_quest_columns,quest,&quest_length->count);
		captured_length=quest_length->count;captured=high_alloc(captured_length);memcpy(captured,packet,captured_length);print_end("packet-free");
	}else {
		assert(!strncmp(argv[1],"--fault=",8));failure=argv[1]+8;transport=!strcmp(failure,"transport");
		reset_end(0x1000,0,1,256);fault_recording=1;sub_4264D0();assert(!"original fault became success");
	}
	if(!json_mode)printf("%u native statistics end %s cases passed\n",cases,argv[1]+2);
	for(uint32_t i=0;i<owned_count;++i)free(owned[i]);free(captured);
	free(header);free(quest);free(records);free(pairs);free(tags);free(players);free(player_count);free(pair_count);free(length_box);free(quest_length);
	return 0;
}
