#include <stdio.h>
#include <limits.h>
#ifdef _WIN32
#include <windows.h>
#else
#include <sys/mman.h>
#endif
#include "../defs.h"
#include "../GAME1_1.h"
#include "../GAME1_2.h"
#include "../GAME1_3.h"
#include "../common__strman.h"
#include "../quest_win_screen_450770.h"

// Use the actual formatter, including native-width variadic string arguments.
#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wsign-compare"
#ifdef __clang__
#pragma GCC diagnostic ignored "-Wincompatible-pointer-types-discards-qualifiers"
#else
#pragma GCC diagnostic ignored "-Wdiscarded-qualifiers"
#endif
#include "../noxstring.c"
#pragma GCC diagnostic pop
char* nox_itoa(int value, char* dest, int radix) {
	if (radix == 10) sprintf(dest, "%d", value);
	else if (radix == 8) sprintf(dest, "%o", (unsigned int)value);
	else sprintf(dest, "%x", (unsigned int)value);
	return dest;
}

uintptr_t dword_8531A0_2576;
uint32_t dword_587000_122956, dword_5d4594_832476;
uint32_t nox_color_white_2523948 = 1, nox_color_orange_2614256 = 2;
uint32_t nox_color_green_2614268 = 3, nox_color_blue_2650684 = 4;
int nox_win_width, nox_win_height;
nox_quest_stats_row_450770 nox_quest_stats_450770[6];
static nox_playerInfo* players;
static nox_window_data* draw_data;
static int32_t* coordinates;
static int32_t stage;
static uint32_t total_secrets, frame;
static int title_mode, failure, draw_count, measure_count, locale_count, current_color;
static unsigned int cases;
static int prompt_result = INT_MIN + 37;
static const char* source_file = "C:\\NoxPost\\src\\client\\Gui\\GUIBrief.c";

// Independently sealed GAME.EXE 0044E8E0 (2057 bytes):
// a6a1bbb66d529bde80cd58e40d160d8b3b595049227f00559e522ee4996a1f4e.
// Its wcscat at 0044E99B loads 005A52B4: 20 00 00 00, NOT " XX1 ".
// Its 0044EE17..0044EE2B loop visits 005A5054..005A507C in 8-byte
// increments, stopping at 005A5084: exactly six rows, irrespective of ASLR.
static const int32_t original_xy[12] = {120,55,290,55,460,55,120,235,290,235,460,235};
static wchar2_t* locale[] = {
	(wchar2_t*)L"결과", (wchar2_t*)L"단계",
	(wchar2_t*)L"생성기", (wchar2_t*)L"비밀", (wchar2_t*)L"처치", (wchar2_t*)L"점수",
	(wchar2_t*)L"합계 %d", (wchar2_t*)L"발견 %d", (wchar2_t*)L"발견 없음",
	(wchar2_t*)L"친구 %d", (wchar2_t*)L"친구 없음", (wchar2_t*)L"계속"
};
static const char* keys[] = {
	"GUIBrief.c:GauntletStatTitle", "Noxworld.c:Stage",
	"GUIBrief.c:GeneratorsDestroyed", "GUIBrief.c:numSecretsFound", "GUIBrief.c:Kills", "GUIBrief.c:TotalScore",
	"GeneralPrint:SecretsTotal", "GeneralPrint:SecretsFound", "GeneralPrint:SecretsNoneFound",
	"GeneralPrint:SecretsFoundByFriends", "GeneralPrint:SecretsNoneFoundByFriends", "GeneralPrint:QuestSplash12"
};
static const int lines[] = {449,451,529,541,553,565,583,593,596,606,612,642};
typedef struct draw_event {
	wchar2_t text[257];
	int wrap, color, x, y, width, height;
} draw_event;
static draw_event actual[64], expected[64];
static int expected_count, expected_locales, expected_measures;

#define CHECK(x) do { if (!(x)) { fprintf(stderr, "case %u line %d: %s failure=%d draws=%d/%d coordinates=%p\n", \
	cases,__LINE__,#x,failure,draw_count,expected_count,(void*)coordinates); return 1; } } while (0)
static void expect(int wrap, int color, const wchar2_t* text, int x, int y, int width, int height) {
	if (expected_count == 64) abort();
	draw_event* e = &expected[expected_count++];
	memset(e,0,sizeof(*e)); nox_wcscpy(e->text,text);
	e->wrap=wrap; e->color=color; e->x=x; e->y=y; e->width=width; e->height=height;
}
static int text_width(const wchar2_t* text) { return (int)nox_wcslen(text)*8; }
static void record(int wrap, void* font, wchar2_t* text, int x, int y, int width, int height) {
	if (font!=draw_data->font || draw_count==64) abort();
	draw_event* e=&actual[draw_count++];
	memset(e,0,sizeof(*e)); nox_wcscpy(e->text,text);
	e->wrap=wrap; e->color=current_color; e->x=x; e->y=y; e->width=width; e->height=height;
}
void* mem_getPtr(uintptr_t base, uintptr_t off) {
	if (base==0x587000 && off>=122960 && off<=123012) return (uint8_t*)coordinates+off-122960;
	if (base==0x5D4594 && off==831228) return &stage;
	if (base==0x5D4594 && off==832356) return &total_secrets;
	abort();
}
uint32_t* mem_getU32Ptr(uintptr_t base, uintptr_t off) { return mem_getPtr(base,off); }
int32_t* mem_getI32Ptr(uintptr_t base, uintptr_t off) { return mem_getPtr(base,off); }
uint32_t gameFrame(void) { return frame; }
int nox_float2int(float value) { return (int)value; }
wchar2_t* nox_strman_loadString_40F1D0(char* key, char** out, char* file, int line) {
	if (out || strcmp(file,source_file)) abort();
	for (int i=0;i<12;i++) if (strcmp(key,keys[i])==0) {
		if (line!=lines[i]) abort();
		locale_count++; return locale[i];
	}
	abort();
}
void nox_xxx_drawSetTextColor_434390(int color) { current_color=color; }
int nox_xxx_drawGetStringSize_43F840(void* font, wchar2_t* text, int* width, int* height, int flags) {
	if (font!=draw_data->font || flags) abort();
	measure_count++;
	// A deterministic font metric lets the row test isolate pointer iteration,
	// even before the independent title regression has been corrected.
	if (width) *width=draw_count==0 ? 200 : text_width(text);
	if (height) *height=12;
	return 0;
}
int nox_xxx_drawString_43F6E0(void* font, wchar2_t* text, int x, int y) {
	record(0,font,text,x,y,0,0); return prompt_result;
}
int nox_xxx_drawStringWrap_43FAF0(void* font, wchar2_t* text, int x, int y, int width, int height) {
	record(1,font,text,x,y,width,height); return INT_MIN + 19;
}

// The unrelated retained signed wchar casts and unused locals are decompiler
// artifacts; permit the baseline's pointer narrowing so it compiles and fails
// on observable six-row output, rather than failing at compilation.
#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wunused-parameter"
#pragma GCC diagnostic ignored "-Wunused-variable"
#pragma GCC diagnostic ignored "-Wpointer-sign"
#pragma GCC diagnostic ignored "-Wpointer-to-int-cast"
// COMPLETE_RENDERER
#pragma GCC diagnostic pop

// Independent expected draw operations/geometry from the original six-column
// table and renderer instructions, not extracted from the current C body.
static void build_expected(unsigned mask, int local, int width, int height, int32_t level, uint32_t tick, unsigned blink) {
	wchar2_t text[257], mine[128], friends[128];
	int dx=(width-640)/2, dy=(height-480)/2;
	expected_count=0; expected_locales=2; expected_measures=1;
	nox_swprintf(text,L"결과 - 단계 %d",level);
	expect(0,1,text,width/2-100,height/2-228,0,0);
	unsigned mine_count=0, friend_count=0, present=0;
	for (int i=0;i<6;i++) if (mask & (1u<<i)) {
		int x=dx+original_xy[2*i], y=dy+original_xy[2*i+1], measures=1;
		wchar2_t full[257]; nox_swprintf(full,L"%d) %s",i+1,players[i].name_final);
		int n=(int)nox_wcslen(full);
		// 0044EB0E..0044EB28: second X - first X - 16 = 154.
		while (x+text_width(full)>=x+154 && n>5) { full[--n]=0; measures++; }
		expect(1,2,full,x,y,227,12);
		expected_measures+=measures;
		const uint32_t values[4]={nox_quest_stats_450770[i].generators,nox_quest_stats_450770[i].secrets,
			nox_quest_stats_450770[i].kills,nox_quest_stats_450770[i].score};
		for (int j=0;j<4;j++) {
			int py=y+27+18*j;
			expect(1,1,locale[2+j],x,py,(int)dword_5d4594_832476,12);
			nox_swprintf(text,L" %d",(int32_t)values[j]);
			expect(1,j==3 ? 4 : 3,text,x+(int)dword_5d4594_832476,py,227-(int)dword_5d4594_832476,12);
		}
		present++; expected_locales+=4;
		if (i==local) mine_count=nox_quest_stats_450770[i].coop_secrets;
		else friend_count+=nox_quest_stats_450770[i].coop_secrets;
	}
	nox_swprintf(text,L"합계 %d",(int32_t)total_secrets);
	expect(0,2,text,dx+320-text_width(text)/2,dy+414,0,0);
	if (mine_count) nox_swprintf(mine,L"발견 %d",mine_count);
	else nox_wcscpy(mine,L"발견 없음");
	if (present>1) {
		if (friend_count) nox_swprintf(friends,L"친구 %d",friend_count);
		else nox_wcscpy(friends,L"친구 없음");
		nox_swprintf(text,L"%s - %s",mine,friends); expected_locales++;
	} else nox_wcscpy(text,mine);
	expect(0,2,text,dx+320-text_width(text)/2,dy+426,0,0);
	expected_locales+=2; expected_measures+=2;
	if ((tick%30 && blink==1) || (! (tick%30) && blink!=1)) {
		expect(0,1,locale[11],dx+320-text_width(locale[11])/2,dy+450,0,0);
		expected_locales++; expected_measures++;
	}
}
static int run_case(unsigned mask, int local, int width, int height, int32_t level, uint32_t tick, unsigned blink) {
	cases++; failure=draw_count=measure_count=locale_count=0;
	memset(actual,0,sizeof(actual));
	for (int i=0;i<6;i++) {
		nox_quest_stats_450770[i].player=(mask & (1u<<i)) ? &players[i] : NULL;
		nox_quest_stats_450770[i].generators=(uint16_t)(0x8000+i);
		nox_quest_stats_450770[i].secrets=(uint16_t)(0xffff-i);
		nox_quest_stats_450770[i].kills=(uint16_t)(300+i);
		nox_quest_stats_450770[i].coop_secrets=(uint16_t)(i%2 ? 0 : i+1);
		nox_quest_stats_450770[i].score=i%2 ? UINT_MAX-i : (unsigned)i;
	}
	dword_8531A0_2576=local<0 ? 0 : (uintptr_t)&players[local];
	nox_win_width=width; nox_win_height=height; stage=level; total_secrets=0xffff; frame=tick;
	dword_587000_122956=blink;
	dword_5d4594_832476=64;
	memcpy(coordinates,original_xy,sizeof(original_xy));
	int32_t coords_before[12]; memcpy(coords_before,coordinates,sizeof(coords_before));
	nox_quest_stats_row_450770 rows_before[6]; memcpy(rows_before,nox_quest_stats_450770,sizeof(rows_before));
	build_expected(mask,local,width,height,level,tick,blink);
	int result=sub_44E8E0(NULL,draw_data);
	CHECK(draw_count==expected_count);
	for (int i=0;i<expected_count;i++) {
		if (i!=0 || title_mode) CHECK(nox_wcscmp(actual[i].text,expected[i].text)==0);
		CHECK(memcmp(&actual[i].wrap,&expected[i].wrap,sizeof(draw_event)-offsetof(draw_event,wrap))==0);
	}
	CHECK(locale_count==expected_locales && measure_count==expected_measures);
	int want_prompt=(tick%30 && blink==1) || (!(tick%30) && blink!=1);
	CHECK(result==(want_prompt ? prompt_result : (tick%30 ? (int)(tick/30) : 1)));
	CHECK(dword_587000_122956==(tick%30 ? blink : (blink==1 ? 0 : 1)));
	CHECK(memcmp(coordinates,coords_before,sizeof(coords_before))==0);
	CHECK(memcmp(nox_quest_stats_450770,rows_before,sizeof(rows_before))==0);
	CHECK(stage==level && total_secrets==0xffff && frame==tick);
	return 0;
}
static void* map_near(uintptr_t address) {
#ifdef _WIN32
	return VirtualAlloc((void*)address,0x80000,MEM_RESERVE|MEM_COMMIT,PAGE_READWRITE);
#else
	void* ptr=mmap((void*)address,0x80000,PROT_READ|PROT_WRITE,MAP_PRIVATE|MAP_ANON,-1,0);
	return ptr==MAP_FAILED ? NULL : ptr;
#endif
}
static void release_map(void* ptr) {
#ifdef _WIN32
	VirtualFree(ptr,0,MEM_RELEASE);
#else
	munmap(ptr,0x80000);
#endif
}
int main(int argc, char** argv) {
	CHECK(argc==2 && (strcmp(argv[1],"title")==0 || strcmp(argv[1],"rows")==0));
	title_mode=strcmp(argv[1],"title")==0;
	players=calloc(6,sizeof(*players)); draw_data=calloc(1,sizeof(*draw_data));
	CHECK(players && draw_data);
	draw_data->font=malloc(8); CHECK(draw_data->font);
	for (int i=0;i<6;i++) {
		for (int j=0;j<27;j++) players[i].name_final[j]=(wchar2_t)(0xac00+i+j);
		players[i].name_final[27]=0;
		if (offsetof(nox_playerInfo,name_final)!=4704) memset((uint8_t*)&players[i]+4704,0,56);
	}
	if (sizeof(void*)>4) CHECK((uintptr_t)players>UINT32_MAX && (uintptr_t)draw_data>UINT32_MAX && (uintptr_t)draw_data->font>UINT32_MAX);
	const uintptr_t bases[]={(uintptr_t)0x37ffc0000ull,(uintptr_t)0x3fffc0000ull,(uintptr_t)0x440000000ull};
	const unsigned masks[]={0,1,3,0x15,0x2a,0x20,0x3f};
	const int locals[]={-1,0,5};
	const int32_t stages[]={0,1,0x8001,-1};
	const uint32_t ticks[]={0,1,29,30,31,UINT_MAX};
	for (unsigned map=0;map<3;map++) {
		// Hint only: never MAP_FIXED or overwrite an existing mapping.
		void* ptr=map_near(bases[map]); CHECK(ptr);
		coordinates=(int32_t*)((uint8_t*)ptr+(map<2 ? 0x40000-20 : 64));
		if (sizeof(void*)>4) {
			CHECK((uintptr_t)ptr==bases[map]);
			if (map==0) CHECK((uintptr_t)(coordinates+1)<0x380000000ull && (uintptr_t)(coordinates+13)>0x380000000ull);
			if (map==1) CHECK((uintptr_t)(coordinates+1)<0x400000000ull && (uintptr_t)(coordinates+13)>0x400000000ull);
		}
		for (unsigned m=0;m<7;m++) for (unsigned l=0;l<3;l++) for (int resolution=0;resolution<2;resolution++)
			for (unsigned st=0;st<4;st++) for (unsigned tick=0;tick<6;tick++) for (unsigned blink=0;blink<2;blink++) {
				if (run_case(masks[m],locals[l],resolution ? 1025 : 640,resolution ? 769 : 480,stages[st],ticks[tick],blink)) return 1;
			}
		release_map(ptr);
	}
	printf("%s: %u complete native score-renderer cases; six rows, original title, draws, blink and storage checked\n",argv[1],cases);
	free(draw_data->font); free(draw_data); free(players);
	return 0;
}
