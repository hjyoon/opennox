#include <stdarg.h>
#include <stdio.h>
#include "../GAME1.h"
#include "../GAME1_1.h"
#include "../GAME1_2.h"
#include "../GAME1_3.h"
#include "../GAME2.h"
#include "../GAME2_2.h"
#include "../GAME2_3.h"
#include "../GAME3_1.h"
#include "../common__magic__speltree.h"

// Formatting remains actual production code, including bare names as format
// strings. Interpose only to reject old-offset format/name pointers safely.
#define nox_swprintf fixture_production_swprintf
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

_Static_assert(offsetof(nox_playerInfo,name_final) == (sizeof(void*) == 4 ? 4704 : 6008), "native final name");
_Static_assert(sizeof(((nox_thing*)0)->pretty_image) == 4, "Thing retains legacy image aliases");

char* nox_itoa(int value, char* dst, int radix) {
	if (radix == 10) sprintf(dst,"%d",value);
	else if (radix == 8) sprintf(dst,"%o",(unsigned int)value);
	else sprintf(dst,"%x",(unsigned int)value);
	return dst;
}
static nox_playerInfo* players[3];
static nox_playerInfo before_players[3];
static nox_thing thing, before_thing;
static void* font_token;
static nox_video_bag_image_t *spell_token, *ability_token, *expected_image;
static wchar2_t expected_text[3][32];
static uint32_t fallback_image, nox_player_netCode_85319C;
static int nox_win_width=640;
static uint32_t dword_8531A0_2572=11111, dword_5d4594_1203832=2;
static uint32_t default_text_color=22222;
static int bad, cases, trace[32], trace_count, measure_count, text_count, image_count;
static int missing_mask, thing_found, mutate_thing, empty_icon, expected_spell;
static int current_color, current_record[6], rect[4], image_position[2];
static int text_role[3], text_position[3][2], text_color[3];

#define CHECK(x) do { if (!(x)) { fprintf(stderr,"case=%d line=%d failed=%s bad=%d\n",cases,__LINE__,#x,bad); return 1; } } while (0)
static void require_service(int ok) { if (!ok) bad++; }
static void event(int code) { if (trace_count == 32) abort(); trace[trace_count++]=code; }
static int utf_len(const wchar2_t* text) { int n=0; while (n<31 && text[n]) n++; return n; }
static void make_name(nox_playerInfo* p, int role, int style) {
	memset(p,0xa5,sizeof(*p));
	if (offsetof(nox_playerInfo,name_final) != 4704) {
		memset((uint8_t*)p+4704,0,56);
		memcpy((uint8_t*)p+4704,(wchar2_t[]){'O','L','D',0},8);
	}
	memset(p->name_final,0,sizeof(p->name_final));
	if (style == 0) {
		p->name_final[0]=(wchar2_t)(0xac00+role); p->name_final[1]=0xd83d; p->name_final[2]=0xde80;
	} else if (style == 1) {
		for (int i=0; i<27; i++) p->name_final[i]=(wchar2_t)(0xb098+role+i);
	} else if (style == 2) {
		memcpy(p->name_final,(wchar2_t[]){'A','%','%','B',0},10);
	}
}
nox_playerInfo* nox_common_playerInfoGetByID_417040(int code) {
	event(code);
	for (int i=0; i<3; i++) if (code == 11*(i+1)) return missing_mask & (1<<i) ? NULL : players[i];
	bad++; return NULL;
}
int nox_swprintf(wchar2_t* dst, const wchar2_t* format, ...) {
	va_list ap; va_start(ap,format); event(40);
	int valid=0, inside_player=0;
	for (int i=0; i<3; i++) {
		valid |= format == players[i]->name_final;
		inside_player |= (uintptr_t)format >= (uintptr_t)players[i] && (uintptr_t)format < (uintptr_t)players[i]+sizeof(*players[i]);
	}
	if (!valid && !inside_player && format && format[0] == '+' && format[1] == '%' && format[2] == 's' && format[3] == 0) {
		va_list copy; va_copy(copy,ap); const wchar2_t* name=va_arg(copy,const wchar2_t*); va_end(copy);
		valid=name == players[1]->name_final;
	}
	if (!valid) { bad++; dst[0]=0; va_end(ap); return 0; }
	int result=nox_vswprintf(dst,format,ap); va_end(ap); return result;
}
void* nox_xxx_guiFontPtrByName_43F360(char* name) { event(1); require_service(strcmp(name,"large") == 0); return font_token; }
void* nox_xxx_spellIcon_424A90(int id) { event(50); require_service(id == expected_spell); return empty_icon ? NULL : spell_token; }
nox_video_bag_image_t* nox_xxx_spellGetAbilityIcon_425310(int id, int which) {
	event(51); require_service(id == 1 && which == 0); return empty_icon ? NULL : ability_token;
}
nox_thing* nox_get_thing(int id) { event(52); require_service(id == current_record[3]); return thing_found ? &thing : NULL; }
uint32_t* sub_4B9650(int id) {
	event(53); require_service(id == current_record[3]);
	if (mutate_thing) thing.pretty_image=UINT32_C(0x87654321);
	return NULL;
}
uint32_t* mem_getU32Ptr(uintptr_t base, uintptr_t off) {
	if (base == 0x5d4594 && off == 1203828) return &fallback_image;
	if (base == 0x5d4594 && off == 2597996) return &default_text_color;
	abort();
}
int32_t* mem_getI32Ptr(uintptr_t base, uintptr_t off) { return (int32_t*)mem_getU32Ptr(base,off); }
int nox_draw_imageMeta_47D5C0(nox_video_bag_image_t* image, uint32_t* off_x, uint32_t* off_y, uint32_t* width, uint32_t* height) {
	// Retain exact four 32-bit output writes into production int locals.
	event(60); require_service(image == expected_image);
	*off_x=2; *off_y=3; *width=20; *height=14; return 1;
}
int nox_xxx_drawGetStringSize_43F840(void* font, wchar2_t* text, int* width, int* height, int max_w) {
	event(61); require_service(font == font_token && max_w == 0 && measure_count < 3);
	int i=measure_count++; if (i > 2) return 0;
	require_service(nox_wcscmp(text,expected_text[i]) == 0);
	*width=utf_len(text)*3; *height=10; return 1;
}
void nox_xxx_drawSetTextColor_434390(int color) { current_color=color; }
void nox_client_drawRectFilledAlpha_49CF10(int x, int y, int width, int height) {
	event(62); rect[0]=x; rect[1]=y; rect[2]=width; rect[3]=height;
}
int nox_xxx_drawString_43F6E0(void* font, wchar2_t* text, int x, int y) {
	event(63); require_service(font == font_token && text_count < 3);
	int index=text_count++; if (index > 2) return x;
	int role=-1, seen=0;
	for (int i=0; i<3; i++) if (current_record[i]) { if (seen++ == index) { role=i; break; } }
	require_service(role >= 0);
	if (role >= 0) require_service(nox_wcscmp(text,expected_text[role]) == 0);
	text_role[index]=role; text_position[index][0]=x; text_position[index][1]=y; text_color[index]=current_color;
	return x+utf_len(text)*3;
}
void nox_client_drawImageAt_47D2C0(nox_video_bag_image_t* image, int x, int y) {
	event(64); image_count++; require_service(image == expected_image); image_position[0]=x; image_position[1]=y;
}

// Compile the unchanged baseline's known pointer/int conversions as warnings
// suppressed only around its complete body. The high-token service contracts
// still reject its actual narrowed values; no source rewriting masks the RED.
#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wint-conversion"
#pragma GCC diagnostic ignored "-Wpointer-to-int-cast"
#pragma GCC diagnostic ignored "-Wint-to-pointer-cast"
#pragma GCC diagnostic ignored "-Wpointer-sign"
#pragma GCC diagnostic ignored "-Wincompatible-pointer-types"
#pragma GCC diagnostic ignored "-Wsign-compare"
// PRODUCTION_DEATH_FEED
#pragma GCC diagnostic pop

static int run_case(const int* cfg) {
	cases++; bad=trace_count=measure_count=text_count=image_count=0;
	missing_mask=cfg[3]; thing_found=cfg[7]; mutate_thing=cfg[8]; empty_icon=cfg[9]; expected_spell=cfg[10];
	current_record[0]=cfg[0] ? 11 : 0; current_record[1]=cfg[1] ? 22 : 0; current_record[2]=cfg[2] ? 33 : 0;
	current_record[3]=cfg[5]; current_record[4]=cfg[4]; current_record[5]=(int)UINT32_C(0xf1234567);
	nox_player_netCode_85319C=(uint32_t)cfg[6];
	fallback_image=(uint32_t)cfg[12];
	memset(&thing,0,sizeof(thing)); thing.pri_class=cfg[8] ? 0x1001000 : 0; thing.pretty_image=(uint32_t)cfg[11];
	memset(expected_text,0,sizeof(expected_text));
	for (int i=0; i<3; i++) {
		make_name(players[i],i,cfg[13]); memcpy(&before_players[i],players[i],sizeof(*players[i]));
		if (cfg[i] && !(missing_mask & (1<<i))) {
			if (i == 1 && cfg[0] && !(missing_mask & 1)) fixture_production_swprintf(expected_text[i],L"+%s",players[i]->name_final);
			else fixture_production_swprintf(expected_text[i],players[i]->name_final);
		}
	}
	if (cfg[4] == 2 && cfg[10] >= 0) expected_image=empty_icon ? (nox_video_bag_image_t*)(uintptr_t)fallback_image : cfg[5] == 2 ? ability_token : spell_token;
	else if (cfg[4] == 1 && thing_found) {
		uint32_t alias=mutate_thing ? UINT32_C(0x87654321) : thing.pretty_image;
		expected_image=(nox_video_bag_image_t*)(uintptr_t)(alias ? alias : fallback_image);
	} else expected_image=(nox_video_bag_image_t*)(uintptr_t)fallback_image;
	before_thing=thing;
	int record_before[6]; memcpy(record_before,current_record,sizeof(current_record));
	int* result=sub_495500(current_record);
	CHECK(bad == 0 && measure_count == 3 && image_count == (expected_image != NULL));
	for (int i=0; i<3; i++) CHECK(memcmp(players[i],&before_players[i],sizeof(*players[i])) == 0);
	if (cfg[4] == 1 && thing_found && mutate_thing) before_thing.pretty_image=UINT32_C(0x87654321);
	CHECK(memcmp(&thing,&before_thing,sizeof(thing)) == 0 && memcmp(current_record,record_before,sizeof(current_record)) == 0);
	int widths[3], total=0, draw_count=0;
	for (int i=0; i<3; i++) { widths[i]=utf_len(expected_text[i])*3; total+=widths[i]; draw_count+=cfg[i] != 0; }
	CHECK(text_count == draw_count);
	int left=(640-(total+20+10))/2, index=0, x=left;
	CHECK(rect[0] == left-5 && rect[1] == 72 && rect[2] == total+40 && rect[3] == 36);
	for (int i=0; i<3; i++) {
		if (i == 2) x=left+widths[0]+widths[1]+30;
		if (!cfg[i]) continue;
		CHECK(text_role[index] == i && text_position[index][0] == x && text_position[index][1] == 85);
		CHECK(text_color[index] == (cfg[6] == current_record[i] ? 11111 : 22222));
		x+=widths[i]; index++;
	}
	if (expected_image) CHECK(image_position[0] == left+widths[0]+widths[1]+3 && image_position[1] == 80);
	CHECK(cfg[2] ? (uintptr_t)result == (uintptr_t)x : result == current_record);
	int want[32], n=0; want[n++]=1;
	for (int i=0; i<3; i++) if (cfg[i]) { want[n++]=11*(i+1); if (!(missing_mask & (1<<i))) want[n++]=40; }
	if (cfg[4] == 1) { want[n++]=52; if (thing_found && mutate_thing) want[n++]=53; }
	else if (cfg[4] == 2 && cfg[10] >= 0) want[n++]=cfg[5] == 2 ? 51 : 50;
	want[n++]=60; want[n++]=61; want[n++]=61; want[n++]=61; want[n++]=62;
	if (cfg[0]) want[n++]=63;
	if (cfg[1]) want[n++]=63;
	if (expected_image) want[n++]=64;
	if (cfg[2]) want[n++]=63;
	CHECK(trace_count == n && memcmp(trace,want,(size_t)n*sizeof(*want)) == 0);
	return 0;
}

int main(void) {
	for (int i=0; i<3; i++) { players[i]=malloc(sizeof(*players[i])); CHECK(players[i]); }
	font_token=malloc(8); spell_token=malloc(8); ability_token=malloc(8); CHECK(font_token && spell_token && ability_token);
	if (sizeof(void*) == 8) CHECK((uintptr_t)font_token > UINT32_MAX && (uintptr_t)spell_token > UINT32_MAX && (uintptr_t)ability_token > UINT32_MAX);
	// first/assistant/victim, lookup-missing mask, icon kind/id, local code,
	// Thing exists/mutates, nil native icon, expected spell ID (-1=fallback),
	// Thing legacy alias, fallback legacy alias, native name style.
	const int configs[][14] = {
		{1,1,1,0,2,1,11,0,0,0,5,0,0x10203040,0},
		{0,1,1,0,2,12,22,0,0,0,5,0,0x10203040,0},
		{1,1,1,1,2,2,33,0,0,0,1,0,0x10203040,0},
		{0,0,1,0,2,4,33,0,0,0,130,0,0x10203040,0},
		{1,1,1,7,2,5,0,0,0,0,60,0,0x10203040,0},
		{1,1,1,0,2,9,0,0,0,0,43,0,0x10203040,1},
		{1,1,1,0,2,17,0,0,0,0,43,0,0x10203040,2},
		{1,1,1,0,2,15,0,0,0,0,56,0,0x10203040,3},
		{1,1,1,0,2,16,0,0,0,0,16,0,0x10203040,0},
		{1,1,1,0,2,99,0,0,0,0,-1,0,0x10203040,0},
		{1,1,1,0,2,1,0,0,0,1,5,0,0x10203040,0},
		{1,1,1,0,2,2,0,0,0,1,1,0,0x10203040,0},
		{1,1,1,0,1,99,0,1,0,0,-1,0x10203040,0x11223344,0},
		{1,1,1,0,1,99,0,1,1,0,-1,0x10203040,0x11223344,0},
		{1,1,1,0,1,99,0,0,0,0,-1,0,0x10203040,0},
		{1,1,1,0,1,99,0,1,0,0,-1,0,0x10203040,0},
		{0,0,0,0,0,99,0,0,0,0,-1,0,0,0},
	};
	for (size_t i=0; i<sizeof(configs)/sizeof(configs[0]); i++) if (run_case(configs[i])) return 1;
	printf("R284 death-feed: %d complete-production-body cases; actual UTF-16 formatter, native Player names/font/native spell icons and unchanged 32-bit Thing/default aliases\n",cases);
	for (int i=0; i<3; i++) free(players[i]);
	free(font_token); free(spell_token); free(ability_token);
	return 0;
}
