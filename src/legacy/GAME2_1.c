#include "GAME2_1.h"
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "GAME3.h"
#include "GAME3_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4_1.h"
#include "GAME5_2.h"
#include "client__draw__debugdraw.h"
#include "client__draw__staticdraw.h"
#include "client__drawable__drawable.h"
#include "client__gui__window.h"
#include "common__net_list.h"
#include "common__object__modifier.h"
#include "common__system__team.h"
#include "operators.h"
#include "secret_wall.h"

#include "client__gui__gamewin__gamewin.h"
#include "client__gui__gui_ctf.h"
#include "client__gui__guicon.h"
#include "client__gui__guiinput.h"
#include "client__gui__guiinv.h"
#include "client__gui__guijourn.h"
#include "client__gui__guimeter.h"
#include "client__gui__guiquit.h"
#include "client__gui__guirank.h"
#include "client__gui__guisave.h"
#include "client__gui__guispell.h"
#include "client__gui__guisumn.h"
#include "client__gui__guitrade.h"
#include "client__gui__tooltip.h"
#include "client__shell__selcolor.h"
#include "client__video__draw_common.h"

#include "client__system__ctrlevnt.h"
#include "common/fs/nox_fs.h"
#include "common__magic__speltree.h"
#include "input_common.h"

#include <limits.h>

extern uintptr_t dword_8531A0_2576;
extern uint32_t dword_8531A0_2572;
extern uint32_t dword_5d4594_1062552;
extern uint32_t dword_5d4594_1049844;
extern uint32_t dword_5d4594_1050008;
extern uint32_t dword_5d4594_1096272;
extern nox_window* dword_5d4594_1049516;
extern uint32_t dword_5d4594_1062520;
extern uint32_t dword_5d4594_1096276;
extern uint32_t dword_5d4594_1049976;
extern uint32_t dword_5d4594_1090284;
extern uint32_t dword_5d4594_1062484;
extern uint32_t dword_5d4594_1049992;
extern uint32_t dword_5d4594_1062556;
extern uint32_t dword_5d4594_1096264;
extern uint32_t dword_5d4594_1062564;
extern uint32_t dword_5d4594_1090280;
extern uint32_t dword_5d4594_1062560;
extern uint32_t dword_5d4594_1096280;
extern uint32_t dword_587000_145672;
extern uint32_t dword_5d4594_1064868;
extern uint32_t dword_5d4594_1049996;
extern uintptr_t dword_5d4594_1062492;
extern uint32_t dword_5d4594_1062496;
extern uint32_t dword_5d4594_1096260;
extern uint32_t nox_client_gui_flag_1556112;
extern uint32_t dword_5d4594_1096284;
extern uint32_t dword_587000_145664;
extern void* dword_5d4594_1096288;
extern uint32_t dword_5d4594_1047936;
extern uint32_t dword_5d4594_1062488;
extern uint32_t dword_5d4594_1062468;
extern nox_window* dword_5d4594_1064860;
extern uint32_t dword_5d4594_1096252;
extern void* dword_5d4594_1064864;
extern uint32_t dword_587000_136184;
extern nox_window* dword_5d4594_1049532;
extern uint32_t dword_5d4594_1049484;
extern uint32_t dword_5d4594_1062516;
extern uint32_t nox_client_translucentFrontWalls_805844;
extern uint64_t qword_581450_9512;
extern uint64_t qword_581450_9544;
extern nox_window* dword_5d4594_1090276;
extern uint32_t dword_5d4594_1062476;
extern uint32_t dword_5d4594_1047932;
extern nox_window* dword_5d4594_1049512;
extern uint32_t nox_client_highResFloors_154952;
extern uint32_t dword_5d4594_1062528;
extern uint32_t dword_5d4594_1062524;
extern nox_window* dword_5d4594_1064856;
extern uint32_t dword_5d4594_1049856;
extern nox_window* dword_5d4594_1049520;
extern uint32_t nox_client_highResFrontWalls_80820;
extern uint32_t dword_5d4594_1049800_inventory_click_row_index;
extern uint32_t nox_xxx_minimap_587000_149232;
extern uint32_t dword_5d4594_1062456;
extern uint32_t dword_5d4594_1063636;
extern uint32_t dword_5d4594_1049796_inventory_click_column_index;
extern nox_window* dword_5d4594_1049508;
extern nox_window* dword_5d4594_1090048;
extern nox_window* dword_5d4594_1049500;
extern uint32_t dword_5d4594_1062512;
extern uint32_t dword_5d4594_1049864;
extern uint32_t dword_5d4594_1062508;
extern nox_window* dword_5d4594_1049504;
extern uint32_t dword_5d4594_1090120;
extern nox_drawable* dword_5d4594_1063116;
extern uintptr_t dword_5d4594_1062480;
extern uint32_t nox_player_netCode_85319C;
extern void* nox_xxx_aClosewoodengat_587000_133480;
extern int nox_win_width;
extern int nox_win_height;
extern uintptr_t array_5D4594_1049872[9];

extern uint32_t nox_color_white_2523948;
extern uint32_t nox_color_blue_2650684;
extern uint32_t nox_color_yellow_2589772;
extern uint32_t nox_color_violet_2598268;
extern uint32_t nox_color_black_2650656;

nox_window* nox_win_unk5 = 0;
nox_window* dword_5d4594_1062452 = 0;
nox_window* nox_inventory_window = 0;
nox_window* nox_inventory_current_weapon_window = 0;
nox_window* nox_inventory_identify_window = 0;
nox_window* nox_inventory_scroll_window = 0;
nox_window* nox_inventory_journal_button = 0;
nox_window* nox_inventory_stats_button = 0;
nox_window* nox_inventory_scroll_up_button = 0;
nox_window* nox_inventory_scroll_down_button = 0;
nox_window* nox_inventory_close_button = 0;
void* nox_inventory_font = 0;

typedef struct {
	int min;
	int max;
	float scale;
	int value;
} nox_slider_data_t;

nox_video_bag_image_t* nox_inventory_images[NOX_INV_IMG_COUNT];
nox_things_imageRef_t* nox_inventory_extra_lives_anim = 0;

nox_window_yyy nox_windows_arr_1093036[7] = {0};
nox_drawable nox_gui_bottle_drawables[3] = {0};

obj_5D4594_2650668_t** ptr_5D4594_2650668 = 0;
const int ptr_5D4594_2650668_cap = 128;

nox_inventory_cell_t nox_client_inventory_grid_1050020[NOX_INVENTORY_CELLS_MAX] = {0};

//----- (00460D40) --------------------------------------------------------
#if 0 // Restored by global_state_native_entry.go.
int sub_460D40() { return dword_5d4594_1049508 != 0; }
#endif

//----- (004611A0) --------------------------------------------------------
#if 0 // Restored by ui_state_trade_native.go.
int sub_4611A0() { return dword_5d4594_1047932; }
#endif

//----- (00461460) --------------------------------------------------------

//----- (00461600) --------------------------------------------------------
#if 0 // Restored by inventory_state_native_entry.go.
nox_drawable* sub_461600(int thing_type) {
	for (size_t slot = 0; slot < 9; slot++) {
		for (nox_drawable* drawable = (nox_drawable*)array_5D4594_1049872[slot]; drawable;
			 drawable = drawable->field_92) {
			if (drawable->field_27 == (uint32_t)thing_type) {
				return drawable;
			}
		}
	}
	return NULL;
}
#endif

//----- (00461930) --------------------------------------------------------
#if 0 // Restored by inventory_state_native_entry.go.
int sub_461930() {
	for (size_t slot = 0; slot < 9; slot++) {
		for (nox_drawable* drawable = (nox_drawable*)array_5D4594_1049872[slot]; drawable;
			 drawable = drawable->field_92) {
			if (drawable->flags28 & 0x1001000) {
				return 1;
			}
		}
	}
	return 0;
}
#endif

//----- (00461EF0) --------------------------------------------------------
#if 0 // Restored by inventory_lookup_native_entry.go.
nox_inventory_cell_t* nox_inventory_find_cell_native_461EF0(int net_code, uint32_t* stack_index) {
	for (int row = 0; row < NOX_INVENTORY_ROW_COUNT; row++) {
		for (int column = 0; column < NOX_INVENTORY_COL_COUNT; column++) {
			nox_inventory_cell_t* cell = &nox_client_inventory_grid_1050020[
				row + NOX_INVENTORY_ROW_COUNT * column];
			uint32_t* codes = &cell->field_4;
			for (uint32_t index = 0; index < cell->field_140; index++) {
				if (codes[index] == (uint32_t)net_code) {
					if (stack_index) {
						*stack_index = index;
					}
					return cell;
				}
			}
		}
	}
	return NULL;
}
#endif

#if 0 // Restored by inventory_lookup_native_entry.go.
char* sub_461EF0(int a1) {
	int row_idx = 0;
	do {
		int col_idx = 0;
		do {
			const nox_inventory_cell_t* p_item = &nox_client_inventory_grid_1050020[row_idx + NOX_INVENTORY_ROW_COUNT * col_idx];
			const int field140_val = p_item->field_140;
			if (field140_val > 0) {
				const uint32_t* p_maybe_stack_items = &p_item->field_4;
				for (int maybe_stack_idx = 0; maybe_stack_idx < field140_val; ++maybe_stack_idx) {
					if (*(p_maybe_stack_items + maybe_stack_idx) == a1) {
						*getMemU32Ptr(0x5D4594, 1049792) = maybe_stack_idx;
						char* result = (char*)getMemAt(0x5D4594, 1049788);
						*getMemU32Ptr(0x5D4594, 1049788) = p_item;
						return result;
					}
				}
			}
			++col_idx;
		} while (col_idx < NOX_INVENTORY_COL_COUNT);
		++row_idx;
	} while (row_idx <= NOX_INVENTORY_ROW_COUNT);

	return 0;
}
#endif

// 462878: variable 'v3' is possibly undefined
// 46288E: variable 'v4' is possibly undefined
// 4628A4: variable 'v5' is possibly undefined
// 4628CE: variable 'v6' is possibly undefined
// 4628E8: variable 'v7' is possibly undefined

//----- (00464B40) --------------------------------------------------------
#if 0 // Restored by inventory_state_native_entry.go.
int sub_464B40(int a1, int a2) { return a1 >= 0 && a1 < 4 && a2 >= 0 && a2 < 21; }
#endif

// 464C71: variable 'v4' is possibly undefined
// 464C8B: variable 'v5' is possibly undefined
// 464CD6: variable 'v6' is possibly undefined
// 464CF0: variable 'v7' is possibly undefined
// 464D4C: variable 'v8' is possibly undefined
// 464E24: variable 'v14' is possibly undefined
// 464E64: variable 'v15' is possibly undefined
// 464F2B: variable 'v21' is possibly undefined
// 464F52: variable 'v22' is possibly undefined
// 464F6C: variable 'v23' is possibly undefined
// 465009: variable 'v24' is possibly undefined
// 465020: variable 'v25' is possibly undefined
// 465066: variable 'v26' is possibly undefined
// 465200: variable 'v30' is possibly undefined
// 46521A: variable 'v31' is possibly undefined
// 465281: variable 'v34' is possibly undefined
// 46535F: variable 'v36' is possibly undefined
// 4653B8: variable 'v37' is possibly undefined



// 467460: using guessed type int nox_xxx_j_inventoryNameSignInit_467460(void);


// 466676: variable 'v2' is possibly undefined
// 4666FF: variable 'v6' is possibly undefined
// 466737: variable 'v10' is possibly undefined

// 4668C0: variable 'v11' is possibly undefined
// 466900: variable 'v13' is possibly undefined

//----- (00466BA0) --------------------------------------------------------
int nox_xxx_wndButtonProc_4A7F50(nox_window* win, int ev, int a3, int a4);
//----- (004674A0) --------------------------------------------------------
#if 0 // Restored by ui_state_trade_native.go.
int sub_4674A0() { return dword_5d4594_1062552; }
#endif

//----- (00467590) --------------------------------------------------------
#if 0 // Restored by global_state_native_entry.go.
int sub_467590() {
	int result; // eax

	if (dword_8531A0_2576) {
		result = ((nox_playerInfo*)dword_8531A0_2576)->field_3684;
	} else {
		result = 1;
	}
	return result;
}
#endif

//----- (004675B0) --------------------------------------------------------
#if 0 // Restored by ui_state_trade_native.go.
int sub_4675B0() { return dword_5d4594_1049864; }
#endif

//----- (00467680) --------------------------------------------------------
#if 0 // Restored by inventory_state_native_entry.go.
void sub_467680() {
	if (dword_5d4594_1049864 == 6) {
		dword_5d4594_1049864 = 0;
	}
}
#endif

//----- (004676A0) --------------------------------------------------------
#if 0 // Restored by inventory_state_native_entry.go.
nox_window* nox_xxx_wndGetHandle_4676A0() { return dword_5d4594_1062452; }
#endif

//----- (00467740) --------------------------------------------------------
#if 0 // Restored by inventory_state_native_entry.go.
int sub_467740(int a1) {
	dword_5d4594_1062488 = a1;
	return a1;
}
#endif

//----- (004678B0) --------------------------------------------------------
#if 0 // Restored by global_state_native_entry.go.
int sub_4678B0() {
	nox_inventory_cell_t* alternate = (nox_inventory_cell_t*)dword_5d4594_1062480;
	return alternate ? alternate->field_4 : 0;
}
#endif

//----- (004678C0) --------------------------------------------------------
#if 0 // Restored by global_state_native_entry.go.
int sub_4678C0() { return dword_5d4594_1062488; }
#endif

//----- (0046A4A0) --------------------------------------------------------
#if 0 // Restored by ui_state_trade_native.go.
int sub_46A4A0() { return dword_5d4594_1064868; }
#endif

//----- (0046AE10) --------------------------------------------------------
#if 0 // Restored by gui_geometry_native_entry.go.
int sub_46AE10(nox_window* win, int enabled) {
	if (!win) {
		return 0;
	}
	if (enabled) {
		win->draw_data.field_0 |= 2;
	} else {
		win->draw_data.field_0 &= ~2u;
	}
	return 1;
}
#endif

//----- (0046AE40) --------------------------------------------------------
#if 0 // Restored by window_field_native_entry.go.
int nox_xxx_wndSetOffsetMB_46AE40(nox_window* win, int x, int y) {
	if (win) {
		win->draw_data.img_px = x;
		win->draw_data.img_py = y;
	}
	return (intptr_t)win;
}
#endif

//----- (0046AE60) --------------------------------------------------------
#if 0 // Restored by window_field_native_entry.go.
int nox_xxx_wndSetIcon_46AE60(nox_window* win, nox_video_bag_image_t* image) {
	if (!win) {
		return -2;
	}
	win->draw_data.bg_image = image;
	return 0;
}
#endif

//----- (0046AEA0) --------------------------------------------------------
#if 0 // Restored by window_field_native_entry.go.
int nox_xxx_wndSetIconLit_46AEA0(nox_window* win, nox_video_bag_image_t* image) {
	if (!win) {
		return -2;
	}
	win->draw_data.hl_image = image;
	return 0;
}
#endif

//----- (0046AEC0) --------------------------------------------------------
#if 0 // Restored by window_field_native_entry.go.
int sub_46AEC0(nox_window* win, nox_video_bag_image_t* image) {
	if (!win) {
		return -2;
	}
	win->draw_data.sel_image = image;
	return 0;
}
#endif

//----- (0046AF40) --------------------------------------------------------
#if 0 // Restored by window_field_native_entry.go.
void* sub_46AF40(nox_window* win) {
	if (!win) {
		return NULL;
	}
	return win->draw_data.font;
}
#endif

//----- (00470A90) --------------------------------------------------------
#if 0 // Restored by ui_state_trade_native.go.
int nox_xxx_playerGet_470A90() { return dword_5d4594_1096252; }
#endif

//----- (00470C40) --------------------------------------------------------
#if 0 // Restored by gui_bar_state_native_entry.go.
int sub_470C40(int a1) {
	int result; // eax

	dword_5d4594_1096264 = a1;
	if (a1) {
		result = *getMemU32Ptr(0x5D4594, 1091964);
		nox_windows_arr_1093036[0].color_1 = *getMemU32Ptr(0x5D4594, 1091964);
		nox_windows_arr_1093036[0].color_2 = *getMemU32Ptr(0x5D4594, 1092992);
	} else {
		result = dword_5d4594_1090280;
		nox_windows_arr_1093036[0].color_1 = dword_5d4594_1090284;
		nox_windows_arr_1093036[0].color_2 = dword_5d4594_1090280;
	}
	return result;
}
#endif

//----- (00470C80) --------------------------------------------------------
#if 0 // Restored by gui_bar_state_native_entry.go.
int nox_xxx_cliSetTotalHealth_470C80(int a1, int a2) {
	int result; // eax

	if (dword_8531A0_2576) {
		((nox_playerInfo*)dword_8531A0_2576)->info.field_2247 = a2;
	}
	result = a1;
	nox_windows_arr_1093036[0].field_2 = a2;
	nox_windows_arr_1093036[0].field_1 = a1;
	dword_5d4594_1096260 = 32;
	return result;
}
#endif

//----- (00470CB0) --------------------------------------------------------
#if 0 // Restored by gui_bar_state_native_entry.go.
int sub_470CB0(int a1) {
	int result; // eax

	result = a1;
	nox_windows_arr_1093036[0].field_1 = a1;
	return result;
}
#endif

//----- (00470CC0) --------------------------------------------------------
#if 0 // Restored by global_state_native_entry.go.
int sub_470CC0() { return nox_windows_arr_1093036[0].field_1; }
#endif

//----- (00470CD0) --------------------------------------------------------
#if 0 // Restored by global_state_native_entry.go.
int sub_470CD0() { return nox_windows_arr_1093036[0].field_2; }
#endif

//----- (00470CE0) --------------------------------------------------------
#if 0 // Restored by gui_bar_state_native_entry.go.
int nox_xxx_cliSetManaAndMax_470CE0(int a1, int a2) {
	int result; // eax

	if (dword_8531A0_2576) {
		((nox_playerInfo*)dword_8531A0_2576)->info.field_2243 = a2;
	}
	result = a1;
	nox_windows_arr_1093036[1].field_2 = a2;
	nox_windows_arr_1093036[1].field_1 = a1;
	dword_5d4594_1096260 = 32;
	return result;
}
#endif

//----- (00470D10) --------------------------------------------------------
#if 0 // Restored by gui_bar_state_native_entry.go.
int nox_xxx_cliSetMana_470D10(int a1) {
	int result; // eax

	result = a1;
	nox_windows_arr_1093036[1].field_1 = a1;
	return result;
}
#endif

//----- (00470DD0) --------------------------------------------------------
#if 0 // Restored by gui_bar_state_native_entry.go.
int nox_xxx_cliGetMana_470DD0() { return nox_windows_arr_1093036[1].field_1; }
#endif

//----- (00471C00) --------------------------------------------------------
extern uint32_t nox_gameDisableMapDraw_5d4594_2650672;
//----- (004724E0) --------------------------------------------------------
#if 0 // Restored by global_state_native_entry.go.
void nox_client_mapZoomIn_4724E0() {
	nox_xxx_minimap_587000_149232 -= 10;
	if (*(int*)&nox_xxx_minimap_587000_149232 < 500) {
		nox_xxx_minimap_587000_149232 = 500;
	}
}
#endif

//----- (00472500) --------------------------------------------------------
#if 0 // Restored by global_state_native_entry.go.
void nox_client_mapZoomOut_472500() {
	nox_xxx_minimap_587000_149232 += 10;
	if (nox_xxx_minimap_587000_149232 > 4000) {
		nox_xxx_minimap_587000_149232 = 4000;
	}
}
#endif

//----- (00472600) --------------------------------------------------------
void* sub_4106A0(int a1);
void* nox_server_wallNextByY_4106B0(void* wall);
int sub_50CB00();
void* sub_50CB10();
//----- (004734F0) --------------------------------------------------------
void nox_video_drawCircle_4B0B90(int a1, int a2, int a3);
// sub_473930 (00473930) is implemented in client_status_animations_473930.go.

// sub_473960 (00473960) is implemented in client_status_animations_473960.go.

//----- (004739E0) --------------------------------------------------------
#if 0 // Restored by gui_geometry_native_entry.go.
int sub_4739E0(uint32_t* a1, int2* a2, int2* a3) {
	int result; // eax

	a3->field_0 = a2->field_0 + *a1 - a1[4];
	result = a2->field_4;
	a3->field_4 = result + a1[1] - a1[5];
	return result;
}
#endif

//----- (00473A10) --------------------------------------------------------
#if 0 // Restored by gui_geometry_native_entry.go.
int sub_473A10(uint32_t* a1, int2* a2, uint32_t* a3) {
	int result; // eax

	*a3 = a2->field_0 + a1[4] - *a1;
	result = a2->field_4;
	a3[1] = result + a1[5] - a1[1];
	return result;
}
#endif

//----- (00473C10) --------------------------------------------------------

//----- (004756E0) --------------------------------------------------------
#if 0 // Restored by drawable_predicate_native_entry.go.
int nox_xxx_sprite_4756E0_drawable(nox_drawable* dr) {
	int result;           // eax
	int (*v2)(uint32_t*, nox_drawable*); // esi
	int v3;               // edx
	int v4;               // ecx

	result = 0;
	v2 = dr->draw_func;
	if (v2) {
		v3 = dr->flags30;
		v4 = dr->flags28;
		if (!(v3 & 0x1000) && v3 & 1 && (v2 == nox_thing_static_draw || v2 == nox_thing_static_random_draw) &&
			!(v4 & 0x80800000) && (v3 & 0x48 || v4 & 0x400000) && !(v3 & 0x800)) {
			result = 1;
		}
	}
	return result;
}
#endif

//----- (00475740) --------------------------------------------------------
#if 0 // Restored by drawable_predicate_native_entry.go.
int nox_xxx_sprite_475740_drawable(nox_drawable* dr) {
	int result;           // eax
	int (*v2)(uint32_t*, nox_drawable*); // edx
	int v3;               // ebx
	int v4;               // ecx

	result = 0;
	v2 = dr->draw_func;
	if (v2) {
		v3 = dr->flags30;
		v4 = dr->flags28;
		if (!(v3 & 0x1000)) {
			if (v3 & 1) {
				result = 1;
				if ((v2 == nox_thing_static_draw || v2 == nox_thing_static_random_draw) && !(v4 & 0x80800000) &&
					!(v3 & 0x800) && (v3 & 0x48 || v4 & 0x400000)) {
					result = 0;
				}
			}
		}
	}
	return result;
}
#endif

//----- (004757A0) --------------------------------------------------------
#if 0 // Restored by drawable_predicate_native_entry.go.
int nox_xxx_sprite_4757A0_drawable(nox_drawable* dr) {
	int a1 = dr;
	int result; // eax
	int v2;     // ecx

	result = 0;
	if (*(uint32_t*)(a1 + 300)) {
		v2 = *(uint32_t*)(a1 + 120);
		if (!(v2 & 0x1000)) {
			if (v2 & 0x4000) {
				if (v2 & 0x40) {
					result = 1;
				}
			}
		}
	}
	return result;
}
#endif

//----- (004757D0) --------------------------------------------------------
#if 0 // Restored by drawable_predicate_native_entry.go.
int sub_4757D0_drawable(nox_drawable* dr) {
	uint32_t* a1 = dr;
	int result; // eax
	int v2;     // ecx

	result = 0;
	if (a1[75]) {
		v2 = a1[30];
		if (!(v2 & 1) && (!(a1[28] & 0x2000) || v2 & 0x1000000) && !(v2 & 0x1000)) {
			result = 1;
		}
	}
	return result;
}
#endif
