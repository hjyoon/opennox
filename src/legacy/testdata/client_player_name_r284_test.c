#include <stdio.h>
#include "../GAME1.h"
#include "../GAME1_1.h"
#include "../GAME1_2.h"
#include "../GAME1_3.h"
#include "../GAME2_1.h"
#include "../GAME2_3.h"
#include "../GAME3_1.h"

// Real production definitions, not packed substitutes for the native structs.
_Static_assert(sizeof(nox_playerInfo) == (sizeof(void*) == 4 ? 4828 : 6160), "Player ABI");
_Static_assert(offsetof(nox_playerInfo, name_final) == (sizeof(void*) == 4 ? 4704 : 6008), "native final name");
_Static_assert(sizeof(((nox_playerInfo*)0)->name_final) == 56, "final name capacity");

static nox_chat_bubble* nox_chat_bubble_head_48D850;
uint32_t nox_color_white_2523948 = UINT32_C(0x123456);
uintptr_t dword_8531A0_2576;
uint32_t dword_5d4594_1320964;
nox_window* dword_5d4594_1320940;

static nox_playerInfo *player, *other, player_before, other_before;
static nox_chat_bubble* bubble;
static nox_drawable* drawable;
static nox_window *root_window, *local_window, *remote_window;
static nox_object_team_t membership;
static nox_team_t team;
static wchar2_t remote_name[25], remote_before[25], packet_name[25];
static uint8_t packet[52];
static uint32_t border_colors[2] = {777, 888};
static int trace[32], trace_count, bad, cases, mode_chat;
static int lookup_found, team_mode, flip_player, copied, finish_calls;
static int text_count, current_text_color, text_colors[2], text_x[2], text_y[2];
static const wchar2_t* text_pointers[2];
static wchar2_t text_values[2][318];
static const wchar2_t* expected_name;

#define CHECK(x) do { if (!(x)) { fprintf(stderr, "case=%d line=%d failed=%s bad=%d\n", cases, __LINE__, #x, bad); return 1; } } while (0)
static void require_service(int ok) { if (!ok) bad++; }
static void record(int value) {
	if (trace_count == (int)(sizeof(trace)/sizeof(trace[0]))) abort();
	trace[trace_count++] = value;
}
static void name_init(nox_playerInfo* p, int style) {
	memset(p, 0xa5, sizeof(*p));
	if (offsetof(nox_playerInfo, name_final) != 4704) {
		memset((uint8_t*)p + 4704, 0, 56);
		memcpy((uint8_t*)p + 4704, (wchar2_t[]){'O','L','D',0}, 8);
	}
	memset(p->name_final, 0, sizeof(p->name_final));
	if (style == 1) {
		p->name_final[0] = 0xac00; p->name_final[1] = 0xd83d;
		p->name_final[2] = 0xde80; p->name_final[3] = 'Z';
	} else if (style == 2) {
		for (int i=0; i<27; i++) p->name_final[i] = (wchar2_t)(0xb098+i);
	}
}
static void reset(int style) {
	name_init(player, style); name_init(other, 1);
	other->name_final[0] = 0xd0c0;
	memcpy(&player_before, player, sizeof(*player));
	memcpy(&other_before, other, sizeof(*other));
	trace_count = bad = copied = finish_calls = text_count = 0;
	current_text_color = 0; expected_name = NULL;
	memset(text_pointers, 0, sizeof(text_pointers));
	memset(text_values, 0, sizeof(text_values));
	memset(bubble, 0, sizeof(*bubble)); memset(drawable, 0, sizeof(*drawable));
	memset(root_window, 0, sizeof(*root_window));
	memset(local_window, 0, sizeof(*local_window));
	memset(remote_window, 0, sizeof(*remote_window));
	dword_5d4594_1320940 = root_window;
	dword_8531A0_2576 = (uintptr_t)player;
	flip_player = 0; nox_chat_bubble_head_48D850 = NULL;
	memset(remote_name, 0, sizeof(remote_name));
	memset(packet, 0, sizeof(packet)); memset(packet_name, 0, sizeof(packet_name));
	packet_name[0]=0xc0c1; packet_name[1]=0xd83d; packet_name[2]=0xde80;
	memcpy(packet+2, packet_name, sizeof(packet_name));
	remote_name[0]='R'; remote_name[1]='0';
	memcpy(remote_before, remote_name, sizeof(remote_name));
}
static int players_unchanged(void) {
	return memcmp(player,&player_before,sizeof(*player)) == 0 && memcmp(other,&other_before,sizeof(*other)) == 0;
}
static int exact_trace(const int* want, size_t size) {
	return trace_count == (int)size && memcmp(trace,want,size*sizeof(*want)) == 0;
}

// Controlled services retain production header signatures. Bad name payloads
// are recorded before dereferencing, so old +4704 fails semantically, not by SIGSEGV.
nox_playerInfo* nox_common_playerInfoGetByID_417040(int code) {
	record(15); require_service(mode_chat && code == 0x1234);
	return lookup_found ? player : NULL;
}
int nox_xxx_guiFontHeightMB_43F320(void* font) { require_service(font == NULL); record(10); return 10; }
void sub_437260(void) { record(11); }
void sub_437290(void) { record(13); }
void sub_48DCF0(nox_draw_viewport_t* viewport) { require_service(viewport == NULL); record(12); }
nox_object_team_t* nox_xxx_objGetTeamByNetCode_418C80(int code) {
	record(14); require_service(code == 0x12345678); return team_mode ? &membership : NULL;
}
nox_team_t* nox_xxx_getTeamByID_418AB0(int id) {
	record(16); require_service(id == membership.id); return team_mode == 2 ? &team : NULL;
}
uint32_t nox_xxx_materialGetTeamColor_418D50(nox_team_t* value) {
	record(17); require_service(value == &team); return UINT32_C(0x654321);
}
void nox_client_drawSetColor_434460(int color) { require_service(color == 777 || color == 888); }
void nox_client_drawAddPoint_49F500(int x, int y) { (void)x; (void)y; }
int nox_client_drawLineFromPoints_49E4B0(void) { return 1; }
void nox_xxx_drawSetTextColor_434390(int color) { current_text_color = color; }
void nox_xxx_drawSetColor_4343E0(int color) { require_service(color == 777); }
int nox_xxx_drawStringWrapHL_43FD00(void* font, wchar2_t* text, int x, int y, int max_w, int max_h) {
	record(18); require_service(font == NULL && max_w == 128 && max_h == 0 && text_count < 2);
	if (text_count >= 2) return 0;
	int index=text_count++;
	text_pointers[index]=text; text_x[index]=x; text_y[index]=y; text_colors[index]=current_text_color;
	if ((index == 0 && text == bubble->text) || (index == 1 && text == expected_name)) {
		for (int i=0; i<317 && text[i]; i++) text_values[index][i]=text[i];
	} else bad++;
	return 1;
}
uint32_t* mem_getU32Ptr(uintptr_t base, uintptr_t off) {
	if (base == 0x852978 && off == 4) return &border_colors[0];
	if (base == 0x85b3fc && off == 956) return &border_colors[1];
	abort();
}
int32_t* mem_getI32Ptr(uintptr_t base, uintptr_t off) { return (int32_t*)mem_getU32Ptr(base,off); }
void* mem_getPtr(uintptr_t base, uintptr_t off) {
	if (base == 0x5d4594 && off == 1319844) return remote_name;
	abort();
}
wchar2_t* nox_wcscpy(wchar2_t* dst, const wchar2_t* src) {
	record(101); copied++;
	require_service(dst == remote_name && src == (const wchar2_t*)(packet+2));
	if (dst == remote_name && src == (const wchar2_t*)(packet+2)) memcpy(dst, packet_name, sizeof(packet_name));
	return dst;
}
int sub_4C1410(void) { record(100); if (flip_player) dword_8531A0_2576=(uintptr_t)other; return 1; }
int nox_xxx_wnd_46ABB0(nox_window* win, int value) { record(102); require_service(win == root_window && value == 1); return 1; }
int nox_xxx_showP2PTradeWnd_4C12D0(void) { record(103); require_service(dword_5d4594_1320964 == 1); return 1; }
int nox_window_get_size(nox_window* win, int* width, int* height) {
	record(104); require_service(win == root_window && width && height); *width=321; *height=123; return 1;
}
int nox_window_setPos_46A9B0(nox_window* win, int x, int y) { record(105); require_service(win == root_window && x == 198 && y == 193); return 1; }
nox_window* nox_xxx_wndGetChildByID_46B0C0(nox_window* win, int id) {
	require_service(win == root_window && (id == 3702 || id == 3703)); record(id == 3702 ? 106 : 108);
	return id == 3702 ? local_window : remote_window;
}
uintptr_t nox_window_call_field_94_fnc(nox_window* win, int event, uintptr_t payload, uintptr_t extra, char* file, int line) {
	(void)file; (void)line;
	require_service(event == 16385 && extra == 0 && (win == local_window || win == remote_window));
	if (win == local_window) {
		record(107); require_service(payload == (uintptr_t)player->name_final);
		if (payload == (uintptr_t)player->name_final) {
			memcpy(text_values[0], (void*)payload, sizeof(player->name_final));
			require_service(memcmp(text_values[0],player_before.name_final,sizeof(player->name_final)) == 0);
		}
		return 33;
	}
	record(109); require_service(payload == (uintptr_t)remote_name); return 77;
}
int sub_467BB0(void) { record(110); finish_calls++; return 1; }

// The legacy chat body independently mis-types the ObjectTeam service return.
// Keep that unrelated production code intact; suppress only its existing warning.
#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wincompatible-pointer-types"
// PRODUCTION_CHAT_NAME
#pragma GCC diagnostic pop
// PRODUCTION_TRADE_START_NAME
// PRODUCTION_TRADE_PREPARE_NAME

static int run_chat(void) {
	// empty, hidden, no drawable, non-player, missing Player, UTF-16, empty name,
	// max-capacity name, unresolved/resolved team, and both bubble tail states.
	const int configs[][8] = {
		{1,0,0,0,0,0,0,1}, {0,0,1,1,1,0,0,1}, {0,1,0,0,1,0,0,1},
		{0,1,1,0,1,0,0,1}, {0,1,1,1,0,0,0,1}, {0,1,1,1,1,0,0,1},
		{0,1,1,1,1,0,0,0}, {0,1,1,1,1,0,0,2}, {0,1,1,1,1,1,0,1},
		{0,1,1,1,1,2,0,1}, {0,1,1,1,1,0,1,1}, {0,1,1,1,1,2,1,2},
	};
	mode_chat=1;
	for (size_t c=0; c<sizeof(configs)/sizeof(configs[0]); c++) {
		cases++; const int* cfg=configs[c]; reset(cfg[7]);
		lookup_found=cfg[4]; team_mode=cfg[5]; membership.id=7;
		bubble->visible=cfg[1]; bubble->drawable=cfg[2] ? drawable : NULL;
		drawable->flags28=cfg[3] ? 4 : 0; drawable->field_32=0x12345678;
		bubble->net_code=0x1234; bubble->x=11; bubble->y=22; bubble->width=50; bubble->height=20;
		bubble->draw_tail=cfg[6]; bubble->text[0]=0xc548; bubble->text[1]=0xb155;
		nox_chat_bubble_head_48D850=cfg[0] ? NULL : bubble;
		int visible=!cfg[0] && cfg[1], is_player=visible && cfg[2] && cfg[3];
		int named=is_player && cfg[4]; expected_name=named ? player->name_final : NULL;
		nox_chat_bubble bubble_before=*bubble; nox_drawable drawable_before=*drawable;
		sub_48D990(NULL);
		CHECK(bad == 0 && players_unchanged());
		CHECK(memcmp(bubble,&bubble_before,sizeof(*bubble)) == 0 && memcmp(drawable,&drawable_before,sizeof(*drawable)) == 0);
		int want[16], n=0; want[n++]=10; want[n++]=11; want[n++]=12;
		if (is_player) { want[n++]=14; want[n++]=15; if (cfg[5]) { want[n++]=16; if (cfg[5] == 2) want[n++]=17; } }
		if (visible) { want[n++]=18; if (named) want[n++]=18; }
		want[n++]=13; CHECK(exact_trace(want,(size_t)n));
		CHECK(text_count == (visible ? 1+named : 0));
		if (visible) CHECK(text_pointers[0] == bubble->text && text_x[0] == 10 && text_y[0] == 21 && text_colors[0] == (int)nox_color_white_2523948 && memcmp(text_values[0],bubble->text,sizeof(bubble->text)) == 0);
		if (named) CHECK(text_pointers[1] == player->name_final && text_x[1] == 10 && text_y[1] == 10 && text_colors[1] == (cfg[5] == 2 ? 0x654321 : (int)nox_color_white_2523948) && memcmp(text_values[1],player->name_final,sizeof(player->name_final)) == 0);
	}
	return 0;
}

static int run_trade(int start) {
	// presence, packet presence, window state, reset callback changes global,
	// native name style. Every name-bearing case uses the cached original Player.
	const int configs[][5] = {
		{0,1,0,0,1}, {1,0,0,0,1}, {1,1,1,0,1}, {1,1,0,0,1},
		{1,1,2,0,2}, {1,1,0,1,1}, {1,1,2,1,0}, {0,1,2,0,1},
		{1,1,1,1,2}, {1,1,1,0,0},
	};
	mode_chat=0;
	for (size_t c=0; c<sizeof(configs)/sizeof(configs[0]); c++) {
		cases++; const int* cfg=configs[c]; reset(cfg[4]);
		dword_8531A0_2576=cfg[0] ? (uintptr_t)player : 0;
		dword_5d4594_1320964=(uint32_t)cfg[2]; flip_player=cfg[3];
		int active=start ? cfg[0] && cfg[1] && cfg[2] != 1 : cfg[0] && cfg[2] != 0;
		int result=start ? nox_xxx_netP2PStartTrade_4C1320(cfg[1] ? packet : NULL) : nox_xxx_prepareP2PTrade_4C1BF0();
		CHECK(bad == 0 && players_unchanged());
		CHECK(result == (active ? (start ? 1 : 77) : (start ? 0 : cfg[2])));
		CHECK(dword_5d4594_1320964 == (uint32_t)(active && start ? 1 : cfg[2]));
		CHECK(dword_8531A0_2576 == (active && cfg[3] ? (uintptr_t)other : cfg[0] ? (uintptr_t)player : 0));
		if (!active) CHECK(trace_count == 0);
		else if (start) {
			const int want[]={100,101,102,103,104,105,106,107,108,109,110}; CHECK(exact_trace(want,sizeof(want)/sizeof(want[0])));
		} else {
			const int want[]={100,106,107,108,109}; CHECK(exact_trace(want,sizeof(want)/sizeof(want[0])));
		}
		CHECK(copied == (active && start) && finish_calls == (active && start));
		CHECK(memcmp(remote_name,active && start ? packet_name : remote_before,sizeof(remote_name)) == 0);
	}
	return 0;
}

int main(int argc, char** argv) {
	if (argc != 2) return 2;
	player=malloc(sizeof(*player)); other=malloc(sizeof(*other));
	bubble=malloc(sizeof(*bubble)); drawable=malloc(sizeof(*drawable));
	root_window=malloc(sizeof(*root_window)); local_window=malloc(sizeof(*local_window)); remote_window=malloc(sizeof(*remote_window));
	CHECK(player && other && bubble && drawable && root_window && local_window && remote_window);
	if (sizeof(void*) == 8) CHECK((uintptr_t)player > UINT32_MAX && (uintptr_t)other > UINT32_MAX);
	int result;
	if (strcmp(argv[1],"chat") == 0) result=run_chat();
	else if (strcmp(argv[1],"start") == 0) result=run_trade(1);
	else if (strcmp(argv[1],"prepare") == 0) result=run_trade(0);
	else return 2;
	if (!result) printf("R284 %s: %d complete-production-body controlled-service cases; Player name offset=%zu; native payload identity, UTF-16, guards, order and whole Player preservation\n",argv[1],cases,offsetof(nox_playerInfo,name_final));
	free(player); free(other); free(bubble); free(drawable); free(root_window); free(local_window); free(remote_window);
	return result;
}
