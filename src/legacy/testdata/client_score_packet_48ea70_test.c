#include <stdarg.h>
#include <stdio.h>
#include "../defs.h"

#define nox_swprintf score_fixture_swprintf
#pragma GCC diagnostic push
#ifdef __clang__
#pragma GCC diagnostic ignored "-Wincompatible-pointer-types-discards-qualifiers"
#else
#pragma GCC diagnostic ignored "-Wdiscarded-qualifiers"
#endif
#include "../noxstring.c"
#pragma GCC diagnostic pop
#undef nox_swprintf

char* nox_itoa(int value, char* dest, int radix) {
	if (radix == 10) sprintf(dest, "%d", value);
	else if (radix == 8) sprintf(dest, "%o", (unsigned int)value);
	else sprintf(dest, "%x", (unsigned int)value);
	return dest;
}

static nox_playerInfo* player;
static uint8_t settings[58];
static int connected, found, host, elimination, debug;
static uint32_t frame, frame_calls;
static int trace[16], trace_count, bad_name, prints, sounds;
static int mutation;
static unsigned int cases;
static wchar2_t format[] = {'[','%','s',']',0};
static wchar2_t message[256];

#define CHECK(x) do { if (!(x)) { \
	fprintf(stderr, "case %u line %d: %s; bad-name=%d fields=%x/%x/%x/%x/%x/%x\n", \
		cases, __LINE__, #x, bad_name, (uint32_t)player->lessons, player->field_2140, player->field_2144, \
		player->field_2152, player->field_2156, player->field_2160); return 1; \
} } while (0)

int nox_xxx_netClearHighBit_578B30(short code) { trace[trace_count++] = 1; return code & 0x7fff; }
char* nox_xxx_cliGamedataGet_416590(int index) { if (index) abort(); trace[trace_count++] = 2; return (char*)settings; }
int nox_client_isConnected_43C700(void) { trace[trace_count++] = 3; return connected; }
char* nox_common_playerInfoGetByID_417040(int code) {
	if (code != 0x1234) abort();
	trace[trace_count++] = 4;
	return found ? (char*)player : NULL;
}
int nox_common_gameFlags_check_40A5C0(int flag) {
	trace[trace_count++] = flag == 1 ? 5 : 6;
	if (flag == 1) return host;
	if (flag != 1024) abort();
	return elimination;
}
int nox_common_getEngineFlag(int flag) { (void)flag; trace[trace_count++] = 7; return debug; }
int nox_xxx_netTestHighBit_578B70(unsigned short code) { if (code != 0x9234) abort(); trace[trace_count++] = 8; return 1; }
uint32_t gameFrame(void) { trace[trace_count++] = 9; frame_calls++; return frame; }
wchar2_t* nox_strman_loadString_40F1D0(char* name, char** out, char* file, int line) {
	(void)out; (void)file;
	if (line == 3452 && strcmp(name,"Eliminated") == 0) trace[trace_count++] = 10;
	else if (line == 3729 && strcmp(name,"SH_NearVictory") == 0) trace[trace_count++] = 11;
	else abort();
	if (mutation) player->name_final[0] = 0xc0c8;
	return format;
}
int nox_swprintf(wchar2_t* out, const wchar2_t* fmt, ...) {
	trace[trace_count++] = 12;
	va_list args, copy;
	va_start(args,fmt); va_copy(copy,args);
	const wchar2_t* name = va_arg(copy,const wchar2_t*);
	va_end(copy);
	if (name != player->name_final) { bad_name++; out[0]=0; va_end(args); return 0; }
	int result = nox_vswprintf(out,fmt,args);
	va_end(args);
	return result;
}
void nox_xxx_clientPlaySoundSpecial_452D80(int id, int volume) {
	if (id != 312 || volume != 100) abort();
	trace[trace_count++] = 13;
	sounds++;
}
void nox_xxx_printCentered_445490(wchar2_t* value) {
	trace[trace_count++] = 14;
	nox_wcscpy(message,value);
	prints++;
}

#define NOX_ENGINE_FLAG_ENABLE_NET_DEBUG 0x8000
// PRODUCTION_SCORE_READERS

static int production_score_packet(int op, unsigned char* data) {
	int k, v124, v125, v140, v141;
	char* v126;
	wchar2_t *v128, *v143, v395[256], v406[256];
	long long v5;
	// PRODUCTION_PLAYER_v127
	// PRODUCTION_PLAYER_v142
	// The original treasure arm uses these decompiler word slots. Their bit
	// layout, not LP64 integer arithmetic, is fixed to the original DWORDs.
#define HIWORD(x) (((uint16_t*)&(x))[1])
#define WORD2(x) (((uint16_t*)&(x))[2])
	switch (op) {
	// PRODUCTION_SCORE_ARM_78
	// PRODUCTION_SCORE_ARM_85
	default: return -1;
	}
}

static void write16(uint8_t* p, uint16_t value) { p[0]=value; p[1]=value>>8; }
static void write32(uint8_t* p, uint32_t value) { for (int i=0;i<4;i++) p[i]=value>>(8*i); }

static void reset(void) {
	memset(player,0x5a,sizeof(*player));
	player->active=1;
	player->lessons=-17;
	player->field_2140=0x77778888;
	player->field_2144=0xaabbccdd;
	player->field_2152=8;
	player->field_2156=9;
	player->field_2160=10;
	if (offsetof(nox_playerInfo,name_final)!=4704) memset((uint8_t*)player+4704,0,56);
	player->name_final[0]=0xac00; player->name_final[1]=0xd83d; player->name_final[2]=0xde80; player->name_final[3]=0;
	memset(settings,0, sizeof(settings));
	memset(message,0,sizeof(message));
	trace_count=bad_name=prints=sounds=0;
	frame_calls=0;
}

static int run_lesson(uint32_t score, uint32_t deaths, uint16_t limit) {
	cases++;
	reset();
	write16(settings+54,limit);
	uint8_t bytes[15]={0x97,0x76}, *packet=bytes+2;
	packet[0]=78; write16(packet+1,0x9234); write32(packet+3,score); write32(packet+7,deaths);
	bytes[13]=0x55; bytes[14]=0xaa;
	uint8_t packet_before[15]; memcpy(packet_before,bytes,sizeof(bytes));
	nox_playerInfo expected; memcpy(&expected,player,sizeof(expected));
	if (connected && found && !host) {
		memcpy(&expected.lessons,&score,4);
		expected.field_2140=deaths; expected.field_2144=frame;
	}
	int announce=connected&&found&&elimination&&(int32_t)deaths >= (int32_t)limit;
	if (announce && mutation) expected.name_final[0]=0xc0c8;
	CHECK(production_score_packet(78,packet)==11);
	CHECK(memcmp(player,&expected,sizeof(expected))==0);
	CHECK(memcmp(bytes,packet_before,sizeof(bytes))==0 && memcmp(settings+54,&limit,2)==0);
	CHECK(bad_name==0 && prints==announce && sounds==announce);
	int want[16]={1,2,3}, n=3;
	if (connected) { want[n++]=4; if (found) { want[n++]=5; if (!host) want[n++]=9; want[n++]=6;
		if (announce) { want[n++]=10; want[n++]=12; want[n++]=13; want[n++]=14; }
	} }
	CHECK(trace_count==n && memcmp(trace,want,n*sizeof(int))==0);
	if (announce) { wchar2_t text[256]; score_fixture_swprintf(text,format,expected.name_final); CHECK(nox_wcscmp(message,text)==0); }
	return 0;
}

static int run_treasure(uint16_t collected, uint16_t total, uint32_t stamp, int almost) {
	cases++;
	reset();
	player->field_2160=stamp;
	player->field_2152=almost ? UINT32_MAX : 5;
	player->field_2156=almost ? 0 : 9;
	uint8_t bytes[11]={0x97,0x76}, *packet=bytes+2;
	packet[0]=85; write16(packet+1,0x9234); write16(packet+3,collected); write16(packet+5,total);
	bytes[9]=0x55; bytes[10]=0xaa;
	uint8_t before[11]; memcpy(before,bytes,sizeof(bytes));
	nox_playerInfo expected; memcpy(&expected,player,sizeof(expected));
	int update=connected&&found&&!host&&frame>stamp;
	if (update) { expected.field_2152=collected; expected.field_2156=total; expected.field_2160=frame; }
	int announce=connected&&found&&expected.field_2152==expected.field_2156-UINT32_C(1);
	if (announce&&mutation) expected.name_final[0]=0xc0c8;
	CHECK(production_score_packet(85,packet)==7);
	CHECK(memcmp(player,&expected,sizeof(expected))==0 && memcmp(bytes,before,sizeof(bytes))==0);
	CHECK(bad_name==0 && prints==announce && sounds==0);
	int want[16]={1,3}, n=2;
	if (connected) { want[n++]=7; if (debug) want[n++]=8; want[n++]=4;
		if (found) { want[n++]=5; if (!host) { want[n++]=9; if (update) want[n++]=9; }
			if (announce) { want[n++]=11; want[n++]=12; want[n++]=14; }
		}
	}
	CHECK(trace_count==n && memcmp(trace,want,n*sizeof(int))==0);
	if (announce) { wchar2_t text[256]; score_fixture_swprintf(text,format,expected.name_final); CHECK(nox_wcscmp(message,text)==0); }
	return 0;
}

int main(void) {
	player=malloc(sizeof(*player));
	if (!player) return 2;
	/* Layout-check diagnostics run before the per-case pattern fills. */
	memset(player,0,sizeof(*player));
	if (sizeof(void*)==8) CHECK((uintptr_t)player>UINT32_MAX && offsetof(nox_playerInfo,lessons)==2140 && offsetof(nox_playerInfo,name_final)==6008);
	const uint32_t values[]={0,1,6,0x7fffffff,0x80000000,UINT32_MAX};
	const uint16_t limits[]={0,1,6,UINT16_MAX};
	frame=UINT32_C(0xf1234567);
	for (connected=0;connected<2;connected++) for (found=0;found<2;found++) for (host=0;host<2;host++) for (elimination=0;elimination<2;elimination++)
		for (unsigned a=0;a<6;a++) for (unsigned b=0;b<6;b++) for (unsigned l=0;l<4;l++) {
			mutation=(a+b+l)&1;
			CHECK(run_lesson(values[a],values[b],limits[l])==0);
		}
	const uint16_t totals[][2]={{0,1},{5,6},{6,6},{0,0},{UINT16_MAX-1,UINT16_MAX},{UINT16_MAX,0}};
	const uint32_t times[]={0,9,10,11,0x7fffffff,0x80000000,UINT32_MAX};
	for (connected=0;connected<2;connected++) for (found=0;found<2;found++) for (host=0;host<2;host++) for (debug=0;debug<2;debug++)
		for (unsigned a=0;a<6;a++) for (unsigned f=0;f<7;f++) for (int almost=0;almost<2;almost++) {
			frame=times[f]; mutation=(a+f)&1;
			CHECK(run_treasure(totals[a][0],totals[a][1],10,almost)==0);
		}
	printf("Score packets: %u native cases; Player=%p lessons-offset=%zu name-offset=%zu; signed elimination, unsigned strict treasure timestamp, host authority, original helper order and whole Player/packet bytes\n",
		cases,(void*)player,offsetof(nox_playerInfo,lessons),offsetof(nox_playerInfo,name_final));
	free(player);
	return 0;
}
