#include <stdarg.h>
#include <stdio.h>
#include "../defs.h"

#define nox_swprintf notices_fixture_swprintf
#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wsign-compare"
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

static nox_playerInfo *player, *other;
static unsigned char settings[2][58];
uintptr_t dword_8531A0_2576;
uint32_t dword_5d4594_1200832, nox_server_sanctuaryHelp_54276;
static int connected, found, acquired, modal, chat, changed, mutation;
static int expected_self, reward_kind, resource_id, queried_code;
static int trace[24], trace_count, prints, bad_name, title_calls;
static unsigned cases;
static wchar2_t title[] = {0xbcf4,0xc0c1,0xd83d,0xde80,0};
static wchar2_t self_format[] = {'S','[','%','s',']',0};
static wchar2_t other_format[] = {'O','[','%','s',']','[','%','s',']',0};
static wchar2_t name_format[] = {'N','[','%','s',']',0};
static wchar2_t message[256];

#define CHECK(x) do { if (!(x)) { fprintf(stderr,"case %u line %d: %s bad-name=%d prints=%d\n",cases,__LINE__,#x,bad_name,prints); return 1; } } while (0)
static void event(int id) { trace[trace_count++]=id; }
static void mutate(void) {
	if (mutation & 1) dword_8531A0_2576=(uintptr_t)other;
	if (mutation & 2) player->name_final[0]=0xc0c8;
}
char* nox_common_playerInfoGetByID_417040(int code) {
	event(1); queried_code=code;
	return found ? (char*)player : NULL;
}
int nox_client_isConnected_43C700(void) { event(2); return connected; }
static wchar2_t* resource(int kind, int id) {
	event(10+kind); title_calls++;
	if (kind!=reward_kind || id!=resource_id) abort();
	mutate();
	return title;
}
wchar2_t* nox_xxx_spellTitle_424930(int id) { return resource(0,id); }
wchar2_t* nox_xxx_guiCreatureGetName_427240(int id) { return resource(1,id); }
wchar2_t* nox_xxx_abilityGetName_0_425260(int id) { return resource(2,id); }
wchar2_t* nox_strman_loadString_40F1D0(char* key, char** out, char* file, int line) {
	(void)out; (void)file;
	if (strcmp(key,"NameChange")==0) {
		if (line!=1470) abort();
		event(30); mutate(); return name_format;
	}
	const char* keys[3][2]={{"plyrspel.c:AwardSpell","plyrspel.c:AwardSpellToOther"},
		{"PlyrGide.c:AwardGuide","PlyrGide.c:AwardGuideToOther"},
		{"ComAblty.c:AwardAbility","ComAblty.c:AwardAbilityToOther"}};
	const int lines[3][2]={{8112,8115},{8145,8148},{8178,8181}};
	int selected=expected_self ? 0 : 1;
	if (strcmp(key,keys[reward_kind][selected])!=0 || line!=lines[reward_kind][selected]) abort();
	event(20+selected); mutate();
	return selected ? other_format : self_format;
}
int nox_swprintf(wchar2_t* out, const wchar2_t* format, ...) {
	event(40);
	va_list args, copy;
	va_start(args,format); va_copy(copy,args);
	const wchar2_t* first=va_arg(copy,const wchar2_t*);
	const wchar2_t* second=format==other_format ? va_arg(copy,const wchar2_t*) : NULL;
	va_end(copy);
	if ((format==self_format && first!=title) ||
		(format==other_format && (first!=player->name_final || second!=title)) ||
		(format==name_format && first!=player->name_final)) {
		bad_name++; out[0]=0; va_end(args); return 0;
	}
	int count=nox_vswprintf(out,format,args);
	va_end(args);
	return count;
}
void nox_xxx_printCentered_445490(wchar2_t* text) { event(41); prints++; nox_wcscpy(message,text); }
char* nox_xxx_cliGamedataGet_416590(int index) { if (index<0 || index>1) abort(); event(50+index); return (char*)settings[index]; }
void sub_4573A0(void) { event(52); }
int sub_4169C0(void) { event(53); return acquired; }
int sub_459DA0(void) { event(54); return modal; }
void sub_4165F0(int from, int to) { if (from!=1 || to!=0) abort(); event(55); memcpy(settings[0],settings[1],58); }
int nox_common_gameFlags_check_40A5C0(int flag) { if (flag!=128) abort(); event(56); return chat; }
void nox_xxx_cliShowHelpGui_49C560(void) { event(57); }
void nox_xxx_cliSetSettingsAcquired_4169D0(int value) { if (value!=1) abort(); event(58); acquired=value; }
void sub_459C30(void) { event(59); }

// WIRE_READERS
static int production_reward(int sub, unsigned char* data) {
	// PLAYER_v303
	// PLAYER_v306
	// PLAYER_v309
	wchar2_t *v304,*v305,*v307,*v308,*v310,*v311;
	wchar2_t *v348,*v349,*v350,*v351,*v352,*v353;
	wchar2_t v403[256],v404[256],v405[256];
	switch (sub) {
	// QUEST_ARM_30
	// QUEST_ARM_31
	// QUEST_ARM_32
	default: return -1;
	}
}
static int production_settings(unsigned char* data) {
	char* v39;
	unsigned int v40=0xdeadbeef;
	wchar2_t *v330,*v41,v401[256];
#define LOBYTE(x) (((uint8_t*)&(x))[0])
	switch (177) {
	// NAME_CHANGE_ARM
	default: return -1;
	}
}

static void reset(void) {
	memset(player,0x5a,sizeof(*player)); memset(other,0x3c,sizeof(*other));
	player->name_final[0]=0xac00; player->name_final[1]=0xd83d; player->name_final[2]=0xde80; player->name_final[3]=0;
	other->name_final[0]=0xb2e4; other->name_final[1]=0;
	if (offsetof(nox_playerInfo,name_final)!=4704) memset((uint8_t*)player+4704,0,56);
	trace_count=prints=bad_name=title_calls=0; queried_code=-1;
	memset(message,0,sizeof(message)); memset(settings,0x5a,sizeof(settings));
}
static int rewards(void) {
	const uint16_t codes[]={0,1,0x7fff,0x8000,0xffff};
	const uint8_t ids[]={0,1,255};
	for (reward_kind=0;reward_kind<3;reward_kind++) for (connected=0;connected<2;connected++) for (found=0;found<2;found++)
		for (int local=0;local<3;local++) for (unsigned c=0;c<5;c++) for (unsigned id=0;id<3;id++) for (mutation=0;mutation<4;mutation++) {
			cases++; reset(); expected_self=local==0; resource_id=ids[id];
			dword_8531A0_2576=local==0 ? (uintptr_t)player : local==1 ? (uintptr_t)other : 0;
			unsigned char bytes[9]={0x97,0x76,240,(unsigned char)(30+reward_kind),ids[id],codes[c],codes[c]>>8,0x55,0xaa};
			unsigned char before[9]; memcpy(before,bytes,9);
			nox_playerInfo expected; memcpy(&expected,player,sizeof(expected));
			nox_playerInfo other_before; memcpy(&other_before,other,sizeof(other_before));
			int announce=connected && found;
			if (announce && (mutation&2)) expected.name_final[0]=0xc0c8;
			CHECK(production_reward(30+reward_kind,bytes+2)==5);
			CHECK(queried_code==codes[c] && bad_name==0 && prints==announce && title_calls==announce);
			CHECK(memcmp(player,&expected,sizeof(expected))==0 && memcmp(other,&other_before,sizeof(other_before))==0 && memcmp(bytes,before,9)==0);
			int want[24]={1,2}, n=2;
			if (announce) { want[n++]=10+reward_kind; want[n++]=20+(expected_self ? 0 : 1); want[n++]=40; want[n++]=41; }
			CHECK(trace_count==n && memcmp(trace,want,n*sizeof(int))==0);
			if (announce) { wchar2_t text[256]; if (expected_self) notices_fixture_swprintf(text,self_format,title);
				else notices_fixture_swprintf(text,other_format,expected.name_final,title);
				CHECK(nox_wcscmp(message,text)==0); }
		}
	return 0;
}
static int names(void) {
	for (int init=0;init<2;init++) for (modal=0;modal<2;modal++) for (int index=0;index<2;index++) for (chat=0;chat<2;chat++)
		for (int pending=0;pending<2;pending++) for (int help=0;help<2;help++) for (changed=0;changed<2;changed++) for (mutation=0;mutation<4;mutation++) {
			cases++; reset(); acquired=init; dword_8531A0_2576=(uintptr_t)player; dword_5d4594_1200832=pending; nox_server_sanctuaryHelp_54276=help;
			settings[index][52]=0x1f;
			unsigned char bytes[64]; memset(bytes,0x97,sizeof(bytes)); unsigned char* data=bytes+2;
			data[0]=177; data[1]=index; for (int i=0;i<58;i++) data[i+2]=(unsigned char)(0x81+i);
			data[54]=changed ? 0x3f : 0x1f;
			unsigned char packet_before[64]; memcpy(packet_before,bytes,64);
			unsigned char expected_settings[2][58]; memcpy(expected_settings,settings,sizeof(settings));
			memcpy(expected_settings[index],data+2,58);
			if (!init && !modal && index==1) memcpy(expected_settings[0],expected_settings[1],58);
			int announce=!init && !modal && chat && pending;
			nox_playerInfo expected,other_before; memcpy(&expected,player,sizeof(expected)); memcpy(&other_before,other,sizeof(other_before));
			if (announce && (mutation&2)) expected.name_final[0]=0xc0c8;
			CHECK(production_settings(data)==60);
			CHECK(bad_name==0 && prints==announce && title_calls==0);
			CHECK(memcmp(bytes,packet_before,64)==0 && memcmp(settings,expected_settings,sizeof(settings))==0);
			CHECK(memcmp(player,&expected,sizeof(expected))==0 && memcmp(other,&other_before,sizeof(other_before))==0);
			CHECK(acquired==1 && dword_5d4594_1200832==(unsigned)(announce ? 0 : pending));
			int want[24]={50+index},n=1;
			if (changed) want[n++]=52;
			want[n++]=53;
			if (!init) { want[n++]=54; if (!modal) { if (index==1) want[n++]=55; want[n++]=56;
				if (chat) { if (pending) { want[n++]=30; want[n++]=40; want[n++]=41; } if (help) want[n++]=57; }
			} want[n++]=58; }
			want[n++]=59;
			CHECK(trace_count==n && memcmp(trace,want,n*sizeof(int))==0);
			if (announce) { wchar2_t text[256]; notices_fixture_swprintf(text,name_format,expected.name_final); CHECK(nox_wcscmp(message,text)==0); }
		}
	return 0;
}
int main(int argc,char** argv) {
	player=malloc(sizeof(*player)); other=malloc(sizeof(*other));
	CHECK(player && other);
	if (sizeof(void*)==8) CHECK((uintptr_t)player>UINT32_MAX && (uintptr_t)other>UINT32_MAX && offsetof(nox_playerInfo,name_final)==6008);
	CHECK(argc==2);
	int result=strcmp(argv[1],"rewards")==0 ? rewards() : names();
	if (result) return result;
	printf("Quest notices %s: %u native cases; Player=%p name-offset=%zu; complete actual arms, original helper order, UTF-16, cached identity/name pointer and whole Player/packet/settings bytes\n",argv[1],cases,(void*)player,offsetof(nox_playerInfo,name_final));
	free(player); free(other);
	return 0;
}
