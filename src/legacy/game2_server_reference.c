#include "game2_server_bridge.h"
extern void* game2_server_reference_alloc_go(size_t count, size_t size);
extern void game2_server_reference_free_go(void* ptr);
/* Only the retained diagnostic translation unit redirects these services;
 * the complete reference include and ordinary product allocator stay intact. */
#define calloc game2_server_reference_alloc_go
#define free game2_server_reference_free_go
#define nox_gui_server_list_first_visible game2_server_old_nox_gui_server_list_first_visible
#define sub_4535E0 game2_server_old_sub_4535E0
#define sub_4535F0 game2_server_old_sub_4535F0
#define sub_453600 game2_server_old_sub_453600
#define sub_453610 game2_server_old_sub_453610
#define sub_453620 game2_server_old_sub_453620
#define sub_453640 game2_server_old_sub_453640
#define sub_453660 game2_server_old_sub_453660
#define sub_453690 game2_server_old_sub_453690
#define sub_453F70 game2_server_old_sub_453F70
#define sub_453F90 game2_server_old_sub_453F90
#define sub_453FA0 game2_server_old_sub_453FA0
#define sub_454000 game2_server_old_sub_454000
#define sub_4559B0 game2_server_old_sub_4559B0
#define sub_456E40 game2_server_old_sub_456E40
#define sub_456F10 game2_server_old_sub_456F10
#define nox_gui_server_players_color_value game2_server_old_nox_gui_server_players_color_value
#define sub_4571A0 game2_server_old_sub_4571A0
#define sub_457350 game2_server_old_sub_457350
#define sub_459560 game2_server_old_sub_459560
#define sub_459870 game2_server_old_sub_459870
#define sub_459D50 game2_server_old_sub_459D50
#define sub_459D60 game2_server_old_sub_459D60
#define sub_459D80 game2_server_old_sub_459D80
#define sub_459DA0 game2_server_old_sub_459DA0
#define sub_459DB0 game2_server_old_sub_459DB0
#define nox_xxx_cliNextMinimapObj_459EC0 game2_server_old_nox_xxx_cliNextMinimapObj_459EC0
#define sub_45A010 game2_server_old_sub_45A010
#define nox_drawable_next_45A070 game2_server_old_nox_drawable_next_45A070
#define nox_xxx_spriteSetActiveMB_45A990_drawable game2_server_old_nox_xxx_spriteSetActiveMB_45A990_drawable
#define nox_xxx_spriteSetFrameMB_45AB80_drawable game2_server_old_nox_xxx_spriteSetFrameMB_45AB80_drawable
static int nox_gui_server_list_first_visible(const nox_scrollListBox_data* data);
static int* sub_4535E0(int* a1);
static int sub_4535F0(int a1);
static char* sub_453600();
static int sub_453610();
static uint8_t* sub_453620(uint8_t* a1, char a2, int a3);
static uint32_t* sub_453640(uint32_t* a1, int a2, int a3);
static int sub_453660(int a1);
static int sub_453690(int a1);
static void sub_453F70(const void* a1);
static char* sub_453F90();
static int sub_453FA0(void* a1, int a2, int a3);
static int sub_454000(const void* a1, int a2);
static int sub_4559B0(wchar2_t* a1);
static int sub_456E40(int a1, int a2);
static int sub_456F10(wchar2_t* a1, int a2);
static uint32_t nox_gui_server_players_color_value(int index);
static int sub_4571A0(int a1, int a2);
static void sub_457350(int old_id, int new_id);
static int sub_459560(int a1);
static char* sub_459870();
static int sub_459D50(int a1);
static int sub_459D60();
static int sub_459D80(int a1);
static int sub_459DA0();
static int sub_459DB0(nox_drawable* dr);
static nox_drawable* nox_xxx_cliNextMinimapObj_459EC0(nox_drawable* dr);
static nox_drawable* sub_45A010(nox_drawable* dr);
static nox_drawable* nox_drawable_next_45A070(nox_drawable* a1);
static nox_drawable* nox_xxx_spriteSetActiveMB_45A990_drawable(nox_drawable* dr);
static nox_drawable* nox_xxx_spriteSetFrameMB_45AB80_drawable(nox_drawable* dr, int frame);
#define sub_457140 game2_server_old_sub_457140
static int sub_457140(int a1,wchar2_t* a2);
#include "game2_server_reference.inc"
int sub_457140(int a1, wchar2_t* a2) {
	if (!dword_5d4594_1045684 || !a2) {
		return 0;
	}
	nox_window* players = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1045684, 10501);
	if (!players) {
		return 0;
	}
	nox_gui_server_player_entry* entry = calloc(1, sizeof(*entry));
	if (!entry) {
		return 0;
	}
	sub_425770(&entry->list);
	nox_common_list_append_4258E0(&nox_gui_server_players.players, &entry->list);
	nox_wcscpy(entry->name, a2);
	entry->net_code = a1;
	return (int)nox_window_call_field_94(players, 16397, (uintptr_t)a2, 3);
}
#undef calloc
#undef free
#undef sub_457140
#undef nox_gui_server_list_first_visible
#undef sub_4535E0
#undef sub_4535F0
#undef sub_453600
#undef sub_453610
#undef sub_453620
#undef sub_453640
#undef sub_453660
#undef sub_453690
#undef sub_453F70
#undef sub_453F90
#undef sub_453FA0
#undef sub_454000
#undef sub_4559B0
#undef sub_456E40
#undef sub_456F10
#undef nox_gui_server_players_color_value
#undef sub_4571A0
#undef sub_457350
#undef sub_459560
#undef sub_459870
#undef sub_459D50
#undef sub_459D60
#undef sub_459D80
#undef sub_459DA0
#undef sub_459DB0
#undef nox_xxx_cliNextMinimapObj_459EC0
#undef sub_45A010
#undef nox_drawable_next_45A070
#undef nox_xxx_spriteSetActiveMB_45A990_drawable
#undef nox_xxx_spriteSetFrameMB_45AB80_drawable
uintptr_t game2_server_reference_call(uint32_t entry,void* p,int a,int b) {
 switch(entry){
 case 0x4535E0: return (uintptr_t)(intptr_t)game2_server_old_sub_4535E0((int*)p);
 case 0x4535F0: return (uintptr_t)(intptr_t)game2_server_old_sub_4535F0(a);
 case 0x453600: return (uintptr_t)(intptr_t)game2_server_old_sub_453600();
 case 0x453610: return (uintptr_t)(intptr_t)game2_server_old_sub_453610();
 case 0x453620: return (uintptr_t)(intptr_t)game2_server_old_sub_453620((uint8_t*)p,(char)a,b);
 case 0x453640: return (uintptr_t)(intptr_t)game2_server_old_sub_453640((uint32_t*)p,a,b);
 case 0x453660: return (uintptr_t)(intptr_t)game2_server_old_sub_453660(a);
 case 0x453690: return (uintptr_t)(intptr_t)game2_server_old_sub_453690(a);
 case 0x453F90: return (uintptr_t)(intptr_t)game2_server_old_sub_453F90();
 case 0x453FA0: return (uintptr_t)(intptr_t)game2_server_old_sub_453FA0(p,a,b);
 case 0x454000: return (uintptr_t)(intptr_t)game2_server_old_sub_454000(p,a);
 case 0x4559B0: return (uintptr_t)(intptr_t)game2_server_old_sub_4559B0((wchar2_t*)p);
 case 0x456E40: return (uintptr_t)(intptr_t)game2_server_old_sub_456E40(a,b);
 case 0x456F10: return (uintptr_t)(intptr_t)game2_server_old_sub_456F10((wchar2_t*)p,a);
 case 0x457140: return (uintptr_t)(intptr_t)game2_server_old_sub_457140(a,(wchar2_t*)p);
 case 0x4571A0: return (uintptr_t)(intptr_t)game2_server_old_sub_4571A0(a,b);
 case 0x459560: return (uintptr_t)(intptr_t)game2_server_old_sub_459560(a);
 case 0x459870: return (uintptr_t)(intptr_t)game2_server_old_sub_459870();
 case 0x459D50: return (uintptr_t)(intptr_t)game2_server_old_sub_459D50(a);
 case 0x459D60: return (uintptr_t)(intptr_t)game2_server_old_sub_459D60();
 case 0x459D80: return (uintptr_t)(intptr_t)game2_server_old_sub_459D80(a);
 case 0x459DA0: return (uintptr_t)(intptr_t)game2_server_old_sub_459DA0();
 case 0x459DB0: return (uintptr_t)(intptr_t)game2_server_old_sub_459DB0((nox_drawable*)p);
 case 0x459EC0: return (uintptr_t)(intptr_t)game2_server_old_nox_xxx_cliNextMinimapObj_459EC0((nox_drawable*)p);
 case 0x45A010: return (uintptr_t)(intptr_t)game2_server_old_sub_45A010((nox_drawable*)p);
 case 0x45A070: return (uintptr_t)(intptr_t)game2_server_old_nox_drawable_next_45A070((nox_drawable*)p);
 case 0x45A990: return (uintptr_t)(intptr_t)game2_server_old_nox_xxx_spriteSetActiveMB_45A990_drawable((nox_drawable*)p);
 case 0x45AB80: return (uintptr_t)(intptr_t)game2_server_old_nox_xxx_spriteSetFrameMB_45AB80_drawable((nox_drawable*)p,a);
 case 0x453F70: game2_server_old_sub_453F70(p);return 0;
 case 0x457350: game2_server_old_sub_457350(a,b);return 0;
 default:abort();
 }
}

int game2_server_reference_first_visible(void* p){return game2_server_old_nox_gui_server_list_first_visible(p);}
