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
static nox_window* nox_inventory_overlay_window = 0;
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

//----- (00461010) --------------------------------------------------------
void sub_461010() {
	if (!dword_5d4594_1049484) {
		return;
	}
	nox_window_set_hidden(nox_quickbar_root(getMemAt(0x5D4594, 1047940)), 1);
	nox_window_set_hidden(dword_5d4594_1049512, 0);
	sub_46AE10(dword_5d4594_1049500, 0);
	nox_xxx_clientPlaySoundSpecial_452D80(797, 100);
	dword_5d4594_1049484 = 0;
}

//----- (00461060) --------------------------------------------------------
void sub_461060() {
	if (dword_5d4594_1049484 == 1) {
		sub_461010();
		return;
	}
	if (*getMemU32Ptr(0x5D4594, 1049476) == 1) {
		nox_xxx_quickBarClose_4606B0();
	}
	nox_window_set_hidden(nox_quickbar_root(getMemAt(0x5D4594, 1047940)), 0);
	nox_window_set_hidden(dword_5d4594_1049512, 1);
	sub_46AE10(dword_5d4594_1049500, 1);
	nox_xxx_clientPlaySoundSpecial_452D80(796, 100);
	dword_5d4594_1049484 = 1;
}

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

//----- (00463430) --------------------------------------------------------
int nox_xxx_inventoryDrawAllMB_463430(nox_window* win, nox_window_data* draw) {
	(void)draw;
	int v1;          // et1
	int v2;          // ebp
	int v3;          // esi
	int v4;          // edi
	int v5;          // eax
	void* v6;        // eax, nox_video_bag_image_t*
	int v7;          // ebp
	void* v8;        // eax, nox_video_bag_image_t*
	int v9;          // edi
	int v10;         // esi
	int v11;         // esi
	int v12;         // edi
	int v13;         // et1
	int v14;         // et1
	int2 v16;        // [esp+10h] [ebp-28h]
	wchar2_t v17[16]; // [esp+18h] [ebp-20h]

	v1 = dword_587000_136184;
	nox_window_setPos_46A9B0(win->parent, 0, v1);
	nox_client_wndGetPosition_46AA60(win, &v16, (unsigned int*)&v16.field_4);
	nox_xxx_guiFontHeightMB_43F320(0);
	v2 = 0;
	v3 = v16.field_0 + 10;
	v4 = v16.field_4 + 234;
	do {
		if ((1 << v2) & *getMemU32Ptr(0x5D4594, 1062540) && v2 != 31 && v2 != 30) {
			v5 = nox_xxx_getEnchantSpell_424920(v2);
			v6 = nox_xxx_spellIcon_424A90(v5);
			nox_client_drawImageAt_47D2C0(v6, v3, v4);
			v3 += 35;
		}
		++v2;
	} while (v2 < 32);
	v7 = 0;
	do {
		if (getMemByte(0x5D4594, 1062536) & (unsigned char)(1 << v7)) {
			v8 = sub_413420(1 << v7);
			nox_client_drawImageAt_47D2C0(v8, v3, v4);
			v3 += 35;
		}
		++v7;
	} while (v7 < 6);
	if (nox_common_gameFlags_check_40A5C0(4096) && nox_inventory_extra_lives_anim) {
		v4 += 5;
		v3 += 6;
		nox_image_ref_anim_t* anim = nox_inventory_extra_lives_anim->field_24;
		if (anim && anim->images_sz) {
			nox_client_drawImageAt_47D2C0(anim->images[gameFrame() % anim->images_sz], v3 - 58, v4 - 53);
		}
		nox_swprintf(v17, L"X %d", *getMemU32Ptr(0x5D4594, 1050012));
		nox_xxx_drawSetTextColor_434390(*getMemIntPtr(0x852978, 0));
		nox_xxx_drawString_43F6E0(nox_inventory_font, v17, v3 + 20, v4 + 9);
		*getMemU32Ptr(0x5D4594, 1049812) = v3 - 30;
		*getMemU32Ptr(0x5D4594, 1049816) = v4 - 20;
		*getMemU32Ptr(0x5D4594, 1049820) = v3 + 30;
		*getMemU32Ptr(0x5D4594, 1049824) = v4 + 20;
	}
	if (nox_common_gameFlags_check_40A5C0(4096) && sub_4BFD30()) {
		v9 = v4 + 5;
		v10 = v3 + 66;
		nox_client_drawImageAt_47D2C0(nox_inventory_images[NOX_INV_IMG_SHARED_KEY_MODE], v10 - 64, v9 - 58);
		*getMemU32Ptr(0x5D4594, 1049828) = v10 - 30;
		*getMemU32Ptr(0x5D4594, 1049832) = v9 - 20;
		*getMemU32Ptr(0x5D4594, 1049836) = v10 + 30;
		*getMemU32Ptr(0x5D4594, 1049840) = v9 + 20;
	}
	if (getMemByte(0x5D4594, 1049868)) {
		v11 = v16.field_4 + 13;
		v12 = v16.field_0 + 254;
		if (v16.field_4 + 163 > 0) {
			nox_xxx_wndDraw_49F7F0();
			nox_client_copyRect_49F6F0(v12, v11, 260, 150);
			if (getMemByte(0x5D4594, 1049869)) {
				if (getMemByte(0x5D4594, 1049869) == 1) {
					nox_xxx_guiDrawJournal_469D40(v12, v11, *(int*)&dword_5d4594_1062512);
				}
			} else {
				nox_xxx_guiDrawInventoryTray_4643B0(v12, v11);
			}
			sub_49F860();
		}
		if (dword_5d4594_1049864 == 5) {
			sub_4627F0(&v16);
			nox_client_drawImageAt_47D2C0(nox_inventory_images[NOX_INV_IMG_IDENTIFY_BASE], v16.field_0, v16.field_4);
		} else {
			if (getMemByte(0x5D4594, 1049870)) {
				if (getMemByte(0x5D4594, 1049870) == 1) {
					nox_client_makePlayerStatsDlg_463880(&v16.field_0);
					nox_client_drawImageAt_47D2C0(nox_inventory_images[NOX_INV_IMG_IDENTIFY_BASE], v16.field_0, v16.field_4);
				}
			} else {
				sub_4BF7E0(&v16);
				nox_client_drawImageAt_47D2C0(nox_inventory_images[NOX_INV_IMG_BASE], v16.field_0, v16.field_4);
			}
			nox_xxx_drawSetTextColor_434390(nox_color_white_2523948);
			nox_xxx_drawStringWrap_43FAF0(0, getMemAt(0x5D4594, 1062588), v16.field_0 + 13, v16.field_4 + 17, 196, 0);
		}
	}
	if (getMemByte(0x5D4594, 1049868) == 1) {
		v14 = dword_587000_136184;
		dword_587000_136184 = v14 + 64;
		if (v14 + 64 > 0) {
			dword_587000_136184 = 0;
			*getMemU8Ptr(0x5D4594, 1049868) = 2;
		}
	} else if (getMemByte(0x5D4594, 1049868) == 3) {
		v13 = dword_587000_136184;
		dword_587000_136184 = v13 - 32;
		if (v13 - 32 <= -225) {
			dword_587000_136184 = -225;
			*getMemU8Ptr(0x5D4594, 1049868) = 0;
			if (getMemByte(0x5D4594, 1049869)) {
				if (getMemByte(0x5D4594, 1049869) == 1) {
					dword_5d4594_1062520 = dword_5d4594_1062512;
				}
			} else {
				dword_5d4594_1062516 = dword_5d4594_1062512;
			}
			*getMemU8Ptr(0x5D4594, 1049869) = 0;
			dword_5d4594_1062512 = dword_5d4594_1062516;
			nox_slider_data_t* slider = nox_inventory_scroll_window->widget_data;
			nox_window_call_field_94(nox_inventory_scroll_window, 16395, 0, 850);
			nox_window_call_field_94(nox_inventory_scroll_window, 16394, slider->max - dword_5d4594_1062512, 0);
			nox_xxx_wndSetIcon_46AE60(nox_inventory_journal_button, 0);
			sub_46AEC0(nox_inventory_journal_button, nox_inventory_images[NOX_INV_IMG_JOURNAL_LIT]);
			nox_xxx_wndSetID_46B080(nox_inventory_journal_button, 9105);
			*getMemU8Ptr(0x5D4594, 1049870) = 0;
			nox_xxx_wndSetIcon_46AE60(nox_inventory_stats_button, nox_inventory_images[NOX_INV_IMG_STATS]);
			sub_46AEC0(nox_inventory_stats_button, nox_inventory_images[NOX_INV_IMG_STATS_LIT]);
			nox_xxx_wndSetID_46B080(nox_inventory_stats_button, 9107);
			nox_window_set_hidden(nox_inventory_current_weapon_window, 0);
		}
	}
	if (sub_467C80() && nox_xxx_playerAnimCheck_4372B0()) {
		sub_467C10();
	}
	return 1;
}

//----- (004643B0) --------------------------------------------------------
int nox_xxx_guiDrawInventoryTray_4643B0(int a1, int a2) {
	int text_width;
	wchar2_t text[16];

	nox_client_drawImageAt_47D2C0(nox_inventory_images[NOX_INV_IMG_TRAY_SPECIAL], a1, a2);
	nox_itow(*(int*)&dword_5d4594_1062552, text, 10);
	nox_xxx_drawSetTextColor_434390(nox_color_yellow_2589772);
	nox_xxx_drawGetStringSize_43F840(nox_inventory_font, text, &text_width, 0, 0);
	nox_xxx_drawString_43F6E0(nox_inventory_font, text, a1 - text_width + 43, a2 + 36);
	if (dword_5d4594_1049864 == 5) {
		nox_client_drawImageAt_47D2C0(nox_inventory_images[NOX_INV_IMG_TRAY_IDENTIFY_LIT], a1, a2 + 50);
	}
	if (sub_473670()) {
		nox_client_drawImageAt_47D2C0(nox_inventory_images[NOX_INV_IMG_TRAY_MAP_LIT], a1, a2 + 100);
	}

	for (int row = 0; row < NOX_INVENTORY_ROW_COUNT; row++) {
		int y = a2 - dword_5d4594_1062512 + 50 * row;
		if (y <= a2 - 50) {
			continue;
		}
		if (y > a2 + 150) {
			break;
		}
		nox_client_drawImageAt_47D2C0(nox_inventory_images[NOX_INV_IMG_TRAY1 + row % 3], a1 + 60, y);
		for (int column = 0; column < NOX_INVENTORY_COL_COUNT; column++) {
			int x = a1 + 60 + 50 * column;
			nox_inventory_cell_t* cell =
				&nox_client_inventory_grid_1050020[row + NOX_INVENTORY_ROW_COUNT * column];
			nox_drawable* drawable = cell->field_140 ? cell->field_0 : NULL;
			if (!drawable) {
				continue;
			}

			bool depleted_ammo = false;
			nox_client_drawEnableAlpha_434560(1);
			nox_client_drawSetAlpha_434580(0x40u);
			double amount = (double)drawable->field_73_1;
			double capacity = (double)drawable->field_73_2;
			uint32_t color = 0x80000000;
			if (amount < capacity * *getMemDoublePtr(0x581450, 9608)) {
				color = *getMemU32Ptr(0x85B3FC, 940);
			} else if (amount < capacity * *(double*)&qword_581450_9544) {
				color = nox_color_yellow_2589772;
			}
			if (color != 0x80000000) {
				nox_client_drawSetColor_434460(color);
				nox_client_drawRectFilledOpaque_49CE30(x, y, 50, 50);
			}

			nox_client_drawSetAlpha_434580(0x80u);
			if (cell->field_132) {
				nox_client_drawImageAt_47D2C0(nox_inventory_images[NOX_INV_IMG_EQUIP_RING], x, y);
			} else if (cell->field_136) {
				nox_client_drawImageAt_47D2C0(nox_inventory_images[NOX_INV_IMG_QUICK_ITEM_RING], x, y);
			} else if (cell->field_4 == dword_5d4594_1062488) {
				nox_inventory_cell_t* alternate = (nox_inventory_cell_t*)dword_5d4594_1062480;
				nox_drawable* alternate_drawable = alternate ? alternate->field_0 : NULL;
				if (alternate_drawable && (alternate_drawable->flags28 & 0x1000000) &&
					(alternate_drawable->flags29 & 0xC)) {
					nox_client_drawImageAt_47D2C0(nox_inventory_images[NOX_INV_IMG_QUICK_ITEM_RING], x, y);
				}
			}
			nox_client_drawEnableAlpha_434560(0);

			drawable->pos.x = x + 25;
			drawable->pos.y = y + 25;
			if (drawable->draw_func) {
				drawable->draw_func((uint32_t*)nox_client_inventory_viewport_native(), drawable);
			}
			if (dword_5d4594_1049864 == 6) {
				if ((drawable->flags28 & 0x13001000) && (drawable->flags28 & 0x1000) &&
					drawable->item_field_112_2 && drawable->item_field_112_0 < drawable->item_field_112_2) {
					depleted_ammo = true;
				}
				if ((drawable->field_73_1 == drawable->field_73_2 || !drawable->field_73_2) && !depleted_ammo) {
					nox_client_drawRectFilledAlpha_49CF10(x, y, 50, 50);
				}
			}
			if (cell->field_140 > 1) {
				nox_swprintf(text, L"%d", cell->field_140);
				nox_xxx_drawSetTextColor_434390(nox_color_white_2523948);
				nox_xxx_drawString_43F6E0(nox_inventory_font, text, x + 6, y + 6);
			}
			if ((drawable->flags28 & 0x13001000) && drawable->item_field_112_0 >= 0) {
				nox_swprintf(text, L"%d", drawable->item_field_112_0);
				nox_xxx_drawSetTextColor_434390(nox_color_blue_2650684);
				nox_xxx_drawGetStringSize_43F840(nox_inventory_font, text, &text_width, 0, 0);
				nox_xxx_drawString_43F6E0(nox_inventory_font, text, x - text_width + 44, y + 6);
			}
		}
	}
	return a2 + 150;
}

//----- (00464770) --------------------------------------------------------
int sub_464770(int a1, int a2, unsigned int a3) {
	(void)a1;
	nox_drawable* dragged = NULL;

	if (dword_5d4594_1049864 == 6) {
		return 1;
	}
	switch (a2) {
	case 5:
	case 8:
		return 1;
	case 6:
		dragged = nox_client_inventory_get_dragged();
		if (!dragged) {
			goto LABEL_25;
		}
		if (!nox_xxx_wndPointInWnd_46AAB0((unsigned int*)nox_inventory_window, (unsigned short)a3, a3 >> 16)) {
			goto LABEL_22;
		}
		if (dword_5d4594_1049856) {
			if (dragged->flags28 & 0x1001000) {
				if (dword_5d4594_1062480) {
					nox_client_invAlterWeapon_4672C0();
				} else {
					dword_5d4594_1062492 = (uintptr_t)dragged;
					nox_xxx_clientDequip_464B70(dragged);
				}
			}
		} else if (!(dragged->flags28 & 0x1001000) ||
				   nox_client_inventory_grid_1050020[dword_5d4594_1049800_inventory_click_row_index +
									   NOX_INVENTORY_ROW_COUNT * dword_5d4594_1049796_inventory_click_column_index]
					   .field_132) {
			sub_4649B0(dragged, *(int*)&dword_5d4594_1049796_inventory_click_column_index,
					   *(int*)&dword_5d4594_1049800_inventory_click_row_index);
		} else {
			if (nox_xxx_ammoCheck_415880(dragged->field_27) == 2) {
				nox_drawable* ammo4 = sub_461600(sub_415840(4));
				nox_drawable* ammo8 = sub_461600(sub_415840(8));
				if (!ammo4 && !ammo8) {
					sub_4649B0(dragged,
							   *(int*)&dword_5d4594_1049796_inventory_click_column_index,
							   *(int*)&dword_5d4594_1049800_inventory_click_row_index);
					nox_xxx_cursorResetDraggedItem_4776A0();
					if (!dword_5d4594_1049856) {
						nox_xxx_spriteDelete_45A4B0(dragged);
					}
					nox_client_inventory_set_dragged(NULL);
					dword_5d4594_1049856 = 0;
					return 1;
				}
			}
			nox_inventory_cell_t* alternate = (nox_inventory_cell_t*)dword_5d4594_1062480;
			if (alternate) {
				alternate->field_136 = 0;
			}
			sub_4649B0(dragged, *(int*)&dword_5d4594_1049796_inventory_click_column_index,
					   *(int*)&dword_5d4594_1049800_inventory_click_row_index);
			alternate = &nox_client_inventory_grid_1050020[dword_5d4594_1049800_inventory_click_row_index +
														 NOX_INVENTORY_ROW_COUNT *
															 dword_5d4594_1049796_inventory_click_column_index];
			nox_xxx_clientSetAltWeapon_461550(alternate);
			alternate->field_136 = 1;
		}
	LABEL_22:
		nox_xxx_cursorResetDraggedItem_4776A0();
		if (!dword_5d4594_1049856) {
			nox_xxx_spriteDelete_45A4B0(dragged);
		}
		nox_client_inventory_set_dragged(NULL);
		dword_5d4594_1049856 = 0;
	LABEL_25:
		nox_xxx_wndClearCaptureMain_46ADE0(nox_inventory_window);
		return 1;
	case 7:
		if (dword_5d4594_1062480) {
			nox_client_invAlterWeapon_4672C0();
		}
		return 1;
	default:
		return 0;
	}
}

//----- (00464B40) --------------------------------------------------------
#if 0 // Restored by inventory_state_native_entry.go.
int sub_464B40(int a1, int a2) { return a1 >= 0 && a1 < 4 && a2 >= 0 && a2 < 21; }
#endif

//----- (00464BA0) --------------------------------------------------------
int nox_xxx_XorEaxEaxSub_464BA0() { return 0; }

//----- (00464BB0) --------------------------------------------------------
int nox_xxx_inventoryWndProc_464BB0(int a1, int a2) { return a2 != 8 && a2 != 12 && a2 != 16; }

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

//----- (004657B0) --------------------------------------------------------
int nox_xxx_trade_4657B0(short a1) {
	char v2[4]; // [esp+0h] [ebp-4h]

	v2[0] = -55;
	v2[1] = 30;
	*(uint16_t*)&v2[2] = a1;
	return nox_netlist_addToMsgListCli_40EBC0(31, 0, (unsigned char*)v2, 4);
}

char nox_xxx_clientTradeMB_4657E0(uint32_t* a1) {
	int v1; // eax

	v1 = nox_xxx_pointInRect_4281F0((int2*)a1, (int4*)getMemAt(0x587000, 136352));
	if (v1) {
		int i = (a1[1] + dword_5d4594_1062512 - 13) / 50 + NOX_INVENTORY_ROW_COUNT * ((*a1 - 314) / 50);
		uint8_t n = nox_client_inventory_grid_1050020[i].field_140;
		LOBYTE(v1) = n;
		if ((uint8_t)v1) {
			// GAME.EXE's PE32 cell + 4*n selects the final net code in the
			// stack. Use the native field layout so the widened drawable pointer
			// at field_0 cannot be mistaken for the first code on 64-bit hosts.
			LOBYTE(v1) = nox_xxx_clientTrade_465870(
				nox_inventory_cell_last_net_code_4657E0(&nox_client_inventory_grid_1050020[i]));
		}
	}
	return v1;
}
// 4657F5: variable 'v1' is possibly undefined

//----- (00465870) --------------------------------------------------------
int nox_xxx_clientTrade_465870(short a1) {
	char v2[4]; // [esp+0h] [ebp-4h]

	v2[0] = -55;
	v2[1] = 28;
	*(uint16_t*)&v2[2] = a1;
	return nox_netlist_addToMsgListCli_40EBC0(31, 0, (unsigned char*)v2, 4);
}

//----- (004658A0) --------------------------------------------------------
void sub_4658A0(nox_window* win, int2* a2) {
	(void)win;
	if (getMemByte(0x5D4594, 1049868) == 2) {
		if (nox_xxx_pointInRect_4281F0(a2, (int4*)getMemAt(0x587000, 136336))) {
			int slot = sub_465990(a2);
			nox_drawable* dragged = slot >= 0 ? (nox_drawable*)array_5D4594_1049872[slot] : NULL;
			nox_client_inventory_set_dragged(dragged);
			dword_5d4594_1049856 = dragged != NULL;
		} else {
			dword_5d4594_1049856 = 0;
			if (nox_xxx_pointInRect_4281F0(a2, (int4*)getMemAt(0x587000, 136368))) {
				if ((a2->field_4 - 13) / 50 == 2) {
					nox_client_toggleMap_473610();
				}
			} else if (nox_xxx_pointInRect_4281F0(a2, (int4*)getMemAt(0x587000, 136352))) {
				dword_5d4594_1049796_inventory_click_column_index = (a2->field_0 - 314) / 50;
				dword_5d4594_1049800_inventory_click_row_index = (a2->field_4 + *(int*)&dword_5d4594_1062512 - 13) / 50;
				if (sub_464B40(*(int*)&dword_5d4594_1049796_inventory_click_column_index,
							   *(int*)&dword_5d4594_1049800_inventory_click_row_index)) {
					nox_xxx_cliInventorySpriteUpd_465A30();
				}
			}
		}
	}
}
// 4658C2: variable 'v2' is possibly undefined
// 4658FF: variable 'v3' is possibly undefined
// 465934: variable 'v5' is possibly undefined

//----- (00465BE0) --------------------------------------------------------
int nox_xxx_clientDrop_465BE0(int2* a1) {
	nox_drawable* dragged = nox_client_inventory_get_dragged();
	if (!dragged) {
		return 0;
	}
	short v2; // dx
	char v3[7]; // [esp+0h] [ebp-8h]

	v3[0] = 114;
	*(uint16_t*)&v3[1] = nox_xxx_netGetUnitCodeCli_578B00(dragged);
	v2 = a1->field_4;
	*(uint16_t*)&v3[3] = a1->field_0;
	*(uint16_t*)&v3[5] = v2;
	return nox_netlist_addToMsgListCli_40EBC0(31, 0, (unsigned char*)v3, 7);
}

//----- (00465C30) --------------------------------------------------------
int nox_xxx_clientKeyEquip_465C30(int a1, int a2) {
	dword_5d4594_1049796_inventory_click_column_index = a1;
	dword_5d4594_1049800_inventory_click_row_index = a2;
	nox_xxx_cliInventorySpriteUpd_465A30();
	nox_drawable* dragged = nox_client_inventory_get_dragged();
	if (!dragged) {
		return 0;
	}
	nox_xxx_clientEquip_4623B0(dragged);
	return sub_4649B0(dragged, a1, a2);
}

//----- (00465C70) --------------------------------------------------------
void nox_xxx_clientUse_465C70(nox_drawable* drawable) {
	if (drawable) {
		uint8_t message[3];
		message[0] = 116;
		*(uint16_t*)&message[1] = nox_xxx_netGetUnitCodeCli_578B00(drawable);
		nox_netlist_addToMsgListCli_40EBC0(31, 0, message, sizeof(message));
	}
}

//----- (00465CA0) --------------------------------------------------------
int sub_465CA0() {
	nox_window_set_hidden(nox_inventory_identify_window, 0);
	dword_5d4594_1049864 = 5;
	nox_client_setCursorType_477610(6);
	return nox_xxx_wndSetCaptureMain_46ADC0(nox_inventory_window);
}

//----- (00465CD0) --------------------------------------------------------
void sub_465CD0(int2* position, uint32_t item_id, uint32_t thing_type, uint32_t amount, uint32_t extra) {
	int2 v7;  // [esp+8h] [ebp-8h]

	(void)thing_type;
	(void)extra;
	if (amount > 0) {
		sub_473970(position, &v7);
		nox_inventory_cell_t* cell = nox_inventory_find_cell_native_461EF0(item_id, NULL);
		if (cell && cell->field_0) {
			for (uint32_t i = 0; i < amount; i++) {
				nox_drawable* dragged = cell->field_0;
				nox_client_inventory_set_dragged(dragged);
				dragged->field_32 = (&cell->field_4)[i];
				if (!sub_4C12C0()) {
					nox_xxx_clientDrop_465BE0(&v7);
				}
				nox_client_inventory_set_dragged(NULL);
			}
		}
	}
}

// 467460: using guessed type int nox_xxx_j_inventoryNameSignInit_467460(void);


//----- (00465E00) --------------------------------------------------------
nox_window* nox_xxx_wndCreateInventoryMB_465E00() {
	nox_window* v0; // eax
	int result;     // eax
	nox_window* v5; // eax
	int v6;       // eax
	int v7;       // eax

	nox_xxx_inventoryLoadImages_467050();
	nox_xxx_inventoryNameSignInit_4671E0();
	nox_inventory_font = nox_xxx_guiFontPtrByName_43F360("small");
	*getMemU32Ptr(0x5D4594, 1049732) = 0;
	*getMemU32Ptr(0x5D4594, 1049736) = 0;
	*getMemU32Ptr(0x5D4594, 1049740) = nox_win_width;
	*getMemU32Ptr(0x5D4594, 1049744) = nox_win_height;
	*getMemU32Ptr(0x5D4594, 1049764) = nox_win_width;
	*getMemU32Ptr(0x5D4594, 1049768) = nox_win_height;
	*getMemU32Ptr(0x5D4594, 1049748) = 0;
	*getMemU32Ptr(0x5D4594, 1049752) = 0;
	dword_5d4594_1062452 = nox_window_new(0, 552, 0, 0, 563, 264, 0);
	if (!dword_5d4594_1062452) {
		return 0;
	}
	nox_window_set_all_funcs(dword_5d4594_1062452, 0, nox_xxx_movEax1Sub_4661C0, 0);
	v0 = nox_window_new(dword_5d4594_1062452, 8, 0, 224, nox_win_width, 40, 0);
	nox_window_set_all_funcs(v0, nox_xxx_XorEaxEaxSub_464BA0, nox_xxx_movEax1Sub_4661C0,
							 nox_xxx_inventroryOnHovewerSub_4667E0);
	nox_inventory_window = nox_window_new(dword_5d4594_1062452, 40, 0, 0, 563, 224, sub_466220);
	if (!nox_inventory_window) {
		return 0;
	}
	nox_window_set_all_funcs(nox_inventory_window, sub_464BD0, nox_xxx_inventoryDrawAllMB_463430, sub_466620);
	nox_inventory_window->draw_data.style |= 0x100u;
	nox_inventory_overlay_window = nox_window_new(dword_5d4594_1062452, 40, 0, 0, 1, 1, 0);
	nox_window_set_all_funcs(nox_inventory_overlay_window, sub_464BD0, nox_xxx_movEax1Sub_4661C0, 0);
	nox_inventory_current_weapon_window = nox_window_new(nox_inventory_window, 40, 173, 174, 50, 50, 0);
	nox_window_set_all_funcs(nox_inventory_current_weapon_window, sub_464770, sub_4625D0, sub_4661D0);
	nox_inventory_current_weapon_window->draw_data.style |= 0x100u;
	result = sub_466950(nox_inventory_window);
	if (!result) {
		return 0;
	}

	result = sub_466C40(nox_inventory_window);
	if (!result) {
		return 0;
	}

	result = sub_466ED0(nox_inventory_window);
	if (!result) {
		return 0;
	}

	nox_win_unk5 = nox_window_new(0, 0x408 | NOX_WIN_LAYER_BACK, -1, nox_win_height - 127, 111, 127, 0);
	if (!nox_win_unk5) {
		return 0;
	}

	nox_window_set_all_funcs(nox_win_unk5, nox_xxx_inventoryWndProc_464BB0, nox_xxx_inventoryDrawProc_466580, 0);
	nox_xxx_wndSetIcon_46AE60(nox_win_unk5, nox_xxx_gLoadImg_42F970("CurrentWeapon"));
	nox_xxx_wndSetIconLit_46AEA0(nox_win_unk5, nox_xxx_gLoadImg_42F970("CurrentWeaponLit"));
	nox_xxx_wndSetOffsetMB_46AE40(nox_win_unk5, -1, 0);
	nox_win_init_cur_weapon(nox_win_unk5, 24, 51, 53, 53);
	sub_471160(nox_win_unk5, 79, 40, 20, 127);
	sub_470D70();
	v5 = nox_window_new(nox_win_unk5, 8, 5, 11, 28, 29, 0);
	nox_window_set_all_funcs(v5, sub_466550, nox_xxx_movEax1Sub_4661C0, sub_466160);
	memset(nox_client_inventory_grid_1050020, 0, sizeof(nox_inventory_cell_t) * NOX_INVENTORY_CELLS_MAX);
	if (!dword_5d4594_1062560) {
		dword_5d4594_1062560 = nox_xxx_getTTByNameSpriteMB_44CFC0("Gold");
		*getMemU32Ptr(0x5D4594, 1049728) = nox_xxx_getTTByNameSpriteMB_44CFC0("QuestGoldPile");
		*getMemU32Ptr(0x5D4594, 1049724) = nox_xxx_getTTByNameSpriteMB_44CFC0("QuestGoldChest");
	}
	nox_client_inventory_grid_1050020[1 * NOX_INVENTORY_ROW_COUNT - 1].field_0 =
		nox_new_drawable_for_thing(*(int*)&dword_5d4594_1062560);
	if (nox_client_inventory_grid_1050020[NOX_INVENTORY_ROW_COUNT - 1].field_0) {
		nox_client_inventory_grid_1050020[NOX_INVENTORY_ROW_COUNT - 1].field_140 = 1;
	}
	v6 = dword_5d4594_1062564;
	if (!dword_5d4594_1062564) {
		v6 = nox_xxx_getTTByNameSpriteMB_44CFC0("Identify");
		dword_5d4594_1062564 = v6;
	}
	nox_client_inventory_grid_1050020[2 * NOX_INVENTORY_ROW_COUNT - 1].field_0 = nox_new_drawable_for_thing(v6);
	if (nox_client_inventory_grid_1050020[2 * NOX_INVENTORY_ROW_COUNT - 1].field_0) {
		nox_client_inventory_grid_1050020[2 * NOX_INVENTORY_ROW_COUNT - 1].field_140 = 1;
	}
	v7 = dword_5d4594_1062556;
	if (!dword_5d4594_1062556) {
		v7 = nox_xxx_getTTByNameSpriteMB_44CFC0("AutoMap");
		dword_5d4594_1062556 = v7;
	}
	nox_client_inventory_grid_1050020[3 * NOX_INVENTORY_ROW_COUNT - 1].field_0 = nox_new_drawable_for_thing(v7);
	if (nox_client_inventory_grid_1050020[3 * NOX_INVENTORY_ROW_COUNT - 1].field_0) {
		nox_client_inventory_grid_1050020[3 * NOX_INVENTORY_ROW_COUNT - 1].field_140 = 1;
	}

	return nox_inventory_window;
}

//----- (00466220) --------------------------------------------------------
int sub_466220(nox_window* win, int a2, uintptr_t a3, uintptr_t a4) {
	(void)win;
	int result; // eax
	int v5;     // ecx
	int v6;     // ecx
	int v7;     // ecx
	int v8;     // eax
	nox_slider_data_t* slider;

	if (a2 == 16391) {
		switch (nox_xxx_wndGetID_46B0A0((nox_window*)a3)) {
		case 9102:
			if (*(int*)&dword_5d4594_1062512 - 25 >= 0) {
				v5 = *(int*)&dword_5d4594_1062512 - 25 - (*(int*)&dword_5d4594_1062512 - 25) % 50;
			} else {
				v5 = 0;
			}
			dword_5d4594_1062512 = v5;
			slider = nox_inventory_scroll_window->widget_data;
			nox_window_call_field_94(nox_inventory_scroll_window, 16394, slider->max - v5, 0);
			nox_xxx_clientPlaySoundSpecial_452D80(766, 100);
			result = 0;
			break;
		case 9103:
			v6 = dword_5d4594_1062512 + 50;
			dword_5d4594_1062512 = v6;
			slider = nox_inventory_scroll_window->widget_data;
			if (v6 <= slider->max) {
				v7 = v6 - v6 % 50;
			} else {
				v7 = slider->max;
			}
			dword_5d4594_1062512 = v7;
			nox_window_call_field_94(nox_inventory_scroll_window, 16394, slider->max - v7, 0);
			nox_xxx_clientPlaySoundSpecial_452D80(766, 100);
			result = 0;
			break;
		case 9105:
			v8 = sub_469FA0() - 150;
			if (dword_5d4594_1049864 == 5) {
				return 0;
			}
			if (v8 < 0) {
				v8 = 0;
			}
			*getMemU8Ptr(0x5D4594, 1049869) = 1;
			dword_5d4594_1062516 = dword_5d4594_1062512;
			dword_5d4594_1062512 = dword_5d4594_1062520;
			slider = nox_inventory_scroll_window->widget_data;
			nox_window_call_field_94(nox_inventory_scroll_window, 16395, 0, v8);
			nox_window_call_field_94(nox_inventory_scroll_window, 16394, slider->max - dword_5d4594_1062512, 0);
			nox_xxx_wndSetIcon_46AE60(nox_inventory_journal_button, nox_inventory_images[NOX_INV_IMG_INVENTORY]);
			sub_46AEC0(nox_inventory_journal_button, nox_inventory_images[NOX_INV_IMG_INVENTORY_LIT]);
			nox_xxx_wndSetID_46B080(nox_inventory_journal_button, 9106);
			result = 0;
			break;
		case 9106:
			*getMemU8Ptr(0x5D4594, 1049869) = 0;
			dword_5d4594_1062520 = dword_5d4594_1062512;
			dword_5d4594_1062512 = dword_5d4594_1062516;
			slider = nox_inventory_scroll_window->widget_data;
			nox_window_call_field_94(nox_inventory_scroll_window, 16395, 0, 850);
			nox_window_call_field_94(nox_inventory_scroll_window, 16394, slider->max - dword_5d4594_1062512, 0);
			nox_xxx_wndSetIcon_46AE60(nox_inventory_journal_button, 0);
			sub_46AEC0(nox_inventory_journal_button, nox_inventory_images[NOX_INV_IMG_JOURNAL_LIT]);
			nox_xxx_wndSetID_46B080(nox_inventory_journal_button, 9105);
			result = 0;
			break;
		case 9107:
			if (dword_5d4594_1049864 == 5) {
				return 0;
			}
			*getMemU8Ptr(0x5D4594, 1049870) = 1;
			nox_xxx_wndSetIcon_46AE60(nox_inventory_stats_button, 0);
			sub_46AEC0(nox_inventory_stats_button, nox_inventory_images[NOX_INV_IMG_DOLL_LIT]);
			nox_xxx_wndSetID_46B080(nox_inventory_stats_button, 9108);
			nox_window_set_hidden(nox_inventory_current_weapon_window, 1);
			result = 0;
			break;
		case 9108:
			if (dword_5d4594_1049864 != 5) {
				*getMemU8Ptr(0x5D4594, 1049870) = 0;
				nox_xxx_wndSetIcon_46AE60(nox_inventory_stats_button, nox_inventory_images[NOX_INV_IMG_STATS]);
				sub_46AEC0(nox_inventory_stats_button, nox_inventory_images[NOX_INV_IMG_STATS_LIT]);
				nox_xxx_wndSetID_46B080(nox_inventory_stats_button, 9107);
				nox_window_set_hidden(nox_inventory_current_weapon_window, 0);
			}
			return 0;
		case 9111:
			sub_467C10();
			result = 0;
			break;
		default:
			return 0;
		}
	} else if (a2 == 16393) {
		result = 0;
		slider = nox_inventory_scroll_window->widget_data;
		dword_5d4594_1062512 = slider->max - a4;
	} else {
		return 0;
	}
	return result;
}

//----- (00466550) --------------------------------------------------------
int sub_466550(nox_window* win, int a2, uintptr_t a3, uintptr_t a4) {
	(void)win;
	(void)a3;
	(void)a4;
	if (a2 >= 5) {
		if (a2 <= 6) {
			return 1;
		}
		if (a2 == 7) {
			nox_client_toggleInventory_467C60();
			return 1;
		}
	}
	return 0;
}

// 466676: variable 'v2' is possibly undefined
// 4666FF: variable 'v6' is possibly undefined
// 466737: variable 'v10' is possibly undefined

// 4668C0: variable 'v11' is possibly undefined
// 466900: variable 'v13' is possibly undefined

//----- (00466950) --------------------------------------------------------
int sub_466950(nox_window* parent) {
	nox_window_data draw = {0};
	draw.style = 1;
	draw.win = parent;
	draw.en_image = nox_inventory_images[NOX_INV_IMG_UP];
	draw.hl_image = nox_inventory_images[NOX_INV_IMG_UP_LIT];
	draw.sel_image = nox_inventory_images[NOX_INV_IMG_UP_LIT];
	draw.text_color = 0x80000000;
	nox_inventory_scroll_up_button = nox_gui_newButtonOrCheckbox_4A91A0(parent, 1161, 522, 2, 20, 25, &draw);
	if (!nox_inventory_scroll_up_button) {
		return 0;
	}
	nox_xxx_wndSetID_46B080(nox_inventory_scroll_up_button, 9102);
	draw.en_image = nox_inventory_images[NOX_INV_IMG_DOWN];
	draw.hl_image = nox_inventory_images[NOX_INV_IMG_DOWN_LIT];
	draw.sel_image = nox_inventory_images[NOX_INV_IMG_DOWN_LIT];
	nox_inventory_scroll_down_button = nox_gui_newButtonOrCheckbox_4A91A0(parent, 1161, 522, 148, 20, 25, &draw);
	if (!nox_inventory_scroll_down_button) {
		return 0;
	}
	nox_xxx_wndSetID_46B080(nox_inventory_scroll_down_button, 9103);
	memset(&draw, 0, sizeof(draw));
	draw.style = 8;
	draw.win = parent;
	draw.bg_color = 0x80000000;
	draw.en_color = 0x80000000;
	draw.hl_color = 0x80000000;
	draw.dis_color = 0x80000000;
	draw.sel_color = 0x80000000;
	nox_slider_data_t values = {.min = 0, .max = 850};
	nox_inventory_scroll_window = nox_gui_newSlider_4B4EE0(parent, 1033, 524, 42, 16, 91, &draw, (float*)&values);
	if (!nox_inventory_scroll_window) {
		return 0;
	}
	nox_xxx_wndSetWindowProc_46B300(nox_inventory_scroll_window, sub_466BF0);
	nox_xxx_wndSetWindowProc_46B300(nox_inventory_scroll_window->field_100, sub_466BA0);
	nox_inventory_scroll_window->field_100->width = 16;
	nox_inventory_scroll_window->field_100->height = 16;
	nox_xxx_wndSetOffsetMB_46AE40(nox_inventory_scroll_window->field_100, 0, -15);
	sub_4B5700(nox_inventory_scroll_window, 0, 0, nox_inventory_images[NOX_INV_IMG_SLIDER],
			   nox_inventory_images[NOX_INV_IMG_SLIDER_LIT], nox_inventory_images[NOX_INV_IMG_SLIDER_LIT]);
	return 1;
}

//----- (00466BA0) --------------------------------------------------------
int nox_xxx_wndButtonProc_4A7F50(nox_window* win, int ev, int a3, int a4);
//----- (00466C40) --------------------------------------------------------
int sub_466C40(nox_window* parent) {
	nox_window_data draw = {0};
	draw.style = 1;
	draw.win = parent;
	draw.sel_image = nox_inventory_images[NOX_INV_IMG_JOURNAL_LIT];
	draw.img_px = -243;
	draw.img_py = -170;
	nox_inventory_journal_button = nox_gui_newButtonOrCheckbox_4A91A0(parent, 1161, 243, 170, 34, 34, &draw);
	if (!nox_inventory_journal_button) {
		return 0;
	}
	nox_xxx_wndSetID_46B080(nox_inventory_journal_button, 9105);
	nox_gui_winSetFunc96_46B070(nox_inventory_journal_button, sub_466E20);
	memset(&draw, 0, sizeof(draw));
	draw.style = 1;
	draw.win = parent;
	draw.bg_image = nox_inventory_images[NOX_INV_IMG_STATS];
	draw.sel_image = nox_inventory_images[NOX_INV_IMG_STATS_LIT];
	draw.img_px = -5;
	draw.img_py = -186;
	nox_inventory_stats_button = nox_gui_newButtonOrCheckbox_4A91A0(parent, 1161, 5, 186, 34, 34, &draw);
	if (!nox_inventory_stats_button) {
		return 0;
	}
	nox_xxx_wndSetID_46B080(nox_inventory_stats_button, 9107);
	nox_gui_winSetFunc96_46B070(nox_inventory_stats_button, sub_466E20);
	memset(&draw, 0, sizeof(draw));
	draw.style = 1;
	draw.win = parent;
	draw.sel_image = nox_inventory_images[NOX_INV_IMG_CLOSE_LIT];
	draw.img_px = -547;
	draw.img_py = -2;
	nox_inventory_close_button = nox_gui_newButtonOrCheckbox_4A91A0(parent, 1161, 547, 2, 16, 16, &draw);
	if (!nox_inventory_close_button) {
		return 0;
	}
	nox_xxx_wndSetID_46B080(nox_inventory_close_button, 9111);
	nox_gui_winSetFunc96_46B070(nox_inventory_close_button, sub_466E20);
	return 1;
}

//----- (00466ED0) --------------------------------------------------------
int sub_466ED0(nox_window* parent) {
	nox_inventory_identify_window = nox_new_window_from_file("identify.wnd", 0);
	if (nox_inventory_identify_window) {
		sub_46AB20(nox_inventory_identify_window, 200, 200);
		sub_46B120(nox_inventory_identify_window, parent);
		nox_window_setPos_46A9B0(nox_inventory_identify_window, 51, 15);
		nox_window* child = nox_xxx_wndGetChildByID_46B0C0(nox_inventory_identify_window, 9155);
		nox_xxx_wndSetDrawFn_46B340(child, sub_466F50);
		return 1;
	}
	return 0;
}

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
uint32_t nox_xxx_wallFlags(int i);
void nox_client_wall_screen_position_473C10(const nox_draw_viewport_t* vp, const unsigned char* wall, int* x,
	int* y) {
	*x = (int)vp->x1 + 23 * wall[5] - (int)vp->field_4;
	*y = (int)vp->y1 + 23 * wall[6] - (int)vp->field_5;
}

void nox_xxx_drawWalls_473C10(nox_draw_viewport_t* vp, void* data) {
	unsigned char* a2 = data;
	unsigned char* v3; // esi
	unsigned char v4;  // dl
	int v6;            // ebx
	int v7;            // ebp
	int v8;            // eax
	int v9;            // edi
	int v10;           // eax
	int v11;           // ecx
	int v12;           // eax
	int v13;           // ecx
	int v14;           // eax
	int v15;           // edx
	int v16;           // eax
	int v17;           // ecx
	int v18;           // eax
	int v19;           // ecx
	bool v20;          // zf
	int v21;           // edx
	int v22;           // ebx
	int v23;           // ecx
	int v24;           // edx
	int v25;           // eax
	unsigned char v26; // al
	unsigned char v27; // al
	char v28;          // cl
	int v29;           // eax
	int v30;           // eax
	int v31;           // eax
	int* v32;          // edi
	int v33;           // eax
	int* wall_light_end; // eax
	int v35;           // eax
	int v36;           // edx
	int v37;           // eax
	int v38;           // eax
	int v39;           // eax
	int v40;           // edx
	int v41;           // eax
	int v42;           // eax
	int v43;           // eax
	int v44;           // edx
	int v45x;         // eax
	int v45y;         // eax
	int v46;           // ebx
	int v47;           // ebp
	int v48;           // eax
	int v49;           // ecx
	int v50;           // edx
	nox_video_bag_image_t* v51; // eax
	nox_video_bag_image_t* v52; // eax
	int v53;           // eax
	uint8_t* v54;      // edi
	int v55x;         // eax
	int v55y;         // eax
	int v56;           // ebx
	int v57;           // ebp
	int v58;           // eax
	int v59;           // edx
	int v60;           // ecx
	nox_video_bag_image_t* v61; // eax
	int v64;           // [esp-14h] [ebp-7Ch]
	int v65;           // [esp-10h] [ebp-78h]
	int v66;           // [esp-Ch] [ebp-74h]
	int v67;           // [esp-8h] [ebp-70h]
	int v68;           // [esp-4h] [ebp-6Ch]
	int v69;           // [esp-4h] [ebp-6Ch]
	int a3;            // [esp+10h] [ebp-58h]
	int a4;            // [esp+14h] [ebp-54h]
	int v72;           // [esp+18h] [ebp-50h]
	int v73;           // [esp+1Ch] [ebp-4Ch]
	int v74;           // [esp+20h] [ebp-48h]
	int v75;           // [esp+24h] [ebp-44h]
	int v76;           // [esp+28h] [ebp-40h]
	int2 v77;          // [esp+2Ch] [ebp-3Ch]
	int2 a2a;          // [esp+34h] [ebp-34h]
	int2 a1a;          // [esp+3Ch] [ebp-2Ch]
	int2 v80;          // [esp+44h] [ebp-24h]
	int2 v81;          // [esp+4Ch] [ebp-1Ch]
	int v82;           // [esp+54h] [ebp-14h]
	int v83[3];        // [esp+5Ch] [ebp-Ch]
	int v84;           // [esp+70h] [ebp+8h]

	_Static_assert(_Generic(v51, nox_video_bag_image_t*: 1, default: 0),
		"wall image pointer must keep native width");
	_Static_assert(_Generic(v52, nox_video_bag_image_t*: 1, default: 0),
		"wall edge image pointer must keep native width");
	_Static_assert(_Generic(v61, nox_video_bag_image_t*: 1, default: 0),
		"wall image pointer must keep native width");
	_Static_assert(_Generic(v32, int*: 1, default: 0),
		"wall light pointer must keep native width");
	_Static_assert(_Generic(wall_light_end, int*: 1, default: 0),
		"wall light pointer must keep native width");

	v3 = a2;
	a4 = nox_win_width;
	v72 = 0;
	a3 = 0;
	if (!a2) {
		return;
	}
	v4 = a2[4];
	if (!(v4 & 1)) {
		return;
	}
	nox_client_wall_screen_position_473C10(vp, a2, &v6, &v7);
	v82 = v6;
	v74 = *getMemU32Ptr(0x587000, 149364 + 4 * a2[3]);
	v8 = v74;
	if (v74 == -1) {
		v8 = *a2;
		v74 = *a2;
	}
	v84 = v8;
	if (v8) {
		if (v8 == 1 && v4 & 0x40) {
			v84 = 12;
		}
	} else if (v4 & 0x40) {
		v84 = 11;
	}
	if (*getMemU32Ptr(0x587000, 80808)) {
		v9 = 16 * v74;
		v10 = *getMemU32Ptr(0x587000, 85440 + 16 * v74);
		v11 = *getMemU32Ptr(0x587000, 85448 + 16 * v74);
		a1a.field_4 = v7 + *getMemU32Ptr(0x587000, 85444 + 16 * v74);
		v12 = v6 + v10;
		a2a.field_4 = v7 + *getMemU32Ptr(0x587000, 85452 + 16 * v74);
		v13 = v6 + v11;
		a1a.field_0 = v12;
		a2a.field_0 = v13;
		if (v74 == 7 || v74 == 9) {
			if (sub_4C42A0(&a1a, &a2a, &a3, &a4)) {
				v22 = 1;
			} else {
				v22 = 0;
				a3 = a2a.field_0;
			}
			v23 = *getMemU32Ptr(0x587000, 85508 + v9);
			a1a.field_0 = v82 + *getMemU32Ptr(0x587000, 85504 + v9);
			v24 = v82 + *getMemU32Ptr(0x587000, 85512 + v9);
			v25 = *getMemU32Ptr(0x587000, 85516 + v9);
			a1a.field_4 = v7 + v23;
			a2a.field_0 = v24;
			a2a.field_4 = v7 + v25;
			if (sub_4C42A0(&a1a, &a2a, &a3, &a4)) {
				v19 = a3;
			} else {
				if (!v22) {
					return;
				}
				if (a4 > a1a.field_0) {
					a4 = a1a.field_0;
				}
				v19 = a3;
			}
		} else {
			if (v74 != 8 && v74 != 10) {
				if (!sub_4C42A0(&a1a, &a2a, &a3, &a4)) {
					return;
				}
				v19 = a3;
			} else {
				v76 = v13;
				v75 = v12;
				if (sub_4C42A0(&a1a, &a2a, &v75, &v76)) {
					v73 = v76 - v75 >= 3;
				} else {
					v73 = 0;
				}
				v14 = *getMemU32Ptr(0x587000, 85504 + 16 * v74);
				v15 = *getMemU32Ptr(0x587000, 85516 + 16 * v74);
				v80.field_4 = v7 + *getMemU32Ptr(0x587000, 85508 + 16 * v74);
				v16 = v6 + v14;
				v17 = v6 + *getMemU32Ptr(0x587000, 85512 + 16 * v74);
				v80.field_0 = v16;
				a3 = v16;
				v81.field_0 = v17;
				a4 = v17;
				v81.field_4 = v7 + v15;
				v18 = sub_4C42A0(&v80, &v81, &a3, &a4);
				v19 = a3;
				v20 = v18 == 0;
				if (v20) {
					v21 = 0;
				} else {
					v21 = a4 - a3 >= 3;
				}
				if (v73) {
					if (v21) {
						if (a3 > v75) {
							v19 = v75;
							a3 = v75;
						}
						if (v19 <= v80.field_0) {
							v19 = 0;
							a3 = 0;
						}
						if (a4 < v76) {
							a4 = v76;
						}
						if (a4 >= v81.field_0) {
							a4 = nox_win_width;
						}
					} else {
						v19 = v75;
						a3 = v75;
						a4 = v76;
						if (v74 != 8) {
							v84 = 1;
							if (v19 == v80.field_0) {
								v19 = 0;
								a3 = 0;
							}
						} else {
							v84 = 0;
							if (v76 == v81.field_0) {
								a4 = nox_win_width;
							}
						}
					}
				} else {
					if (!v21) {
						return;
					}
					v84 = (v74 != 8) + 13;
					if (a4 == v81.field_0) {
						a4 = nox_win_width;
					}
					if (v19 == v80.field_0) {
						v19 = 0;
						a3 = 0;
					}
				}
			}
		}
		if (v19 >= a4) {
			v26 = v3[4];
			v3[3] = 0;
			v3[4] = v26 & 0xFC;
			return;
		}
	}
	v27 = v3[4];
	v28 = v3[4] & 2;
	if (!v28) {
		v29 = (v3[4] >> 2) & 2;
		goto LABEL_64;
	}
	if (*getMemU32Ptr(0x5D4594, 805848) && nox_client_translucentFrontWalls_805844) {
		if (!nox_client_highResFrontWalls_80820 && nox_client_highResFloors_154952) {
			v72 |= 4u;
			goto LABEL_61;
		}
		v72 = 8;
	}
	if (!nox_client_highResFrontWalls_80820) {
		v72 |= 4u;
	}
LABEL_61:
	v29 = (v27 & 8 | 4u) >> 2;
LABEL_64:
	v73 = v29;
	if (v28 && nox_client_translucentFrontWalls_805844 && !(nox_xxx_wallFlags(v3[1]) & 4)) {
		v30 = v72;
		LOBYTE(v30) = v72 | 2;
		v72 = v30;
	} else {
		v72 |= 1u;
	}
	if (*getMemU32Ptr(0x587000, 80816)) {
		switch (v74) {
		case 0:
		case 3:
			v31 = v3[6];
			v77.field_0 = 23 * v3[5];
			v77.field_4 = 23 * (v31 + 1);
			v32 = (int*)sub_469920(&v77);
			if (v32 != (int*)31) {
				v83[0] = *v32;
				v83[1] = v32[1];
				v33 = v32[2];
				v32 = v83;
				v83[2] = v33;
			}
			v77.field_0 += 23;
			v77.field_4 -= 23;
			wall_light_end = (int*)sub_469920(&v77);
			break;
		case 1:
		case 4:
			v35 = v3[6];
			v77.field_0 = 23 * v3[5];
			v77.field_4 = 23 * v35;
			v32 = (int*)sub_469920(&v77);
			if (v32 != (int*)31) {
				v83[0] = *v32;
				v83[1] = v32[1];
				v36 = v32[2];
				v32 = v83;
				v83[2] = v36;
			}
			v77.field_0 += 23;
			v77.field_4 += 23;
			wall_light_end = (int*)sub_469920(&v77);
			break;
		case 7:
			v37 = v3[6];
			v77.field_0 = 23 * v3[5];
			v77.field_4 = 23 * v37;
			v32 = (int*)sub_469920(&v77);
			if (v32 != (int*)31) {
				v83[0] = *v32;
				v83[1] = v32[1];
				v38 = v32[2];
				v32 = v83;
				v83[2] = v38;
			}
			v77.field_0 += 23;
			wall_light_end = (int*)sub_469920(&v77);
			break;
		case 8:
			v39 = v3[6];
			v77.field_0 = 23 * v3[5] + 11;
			v77.field_4 = 23 * v39 + 11;
			v32 = (int*)sub_469920(&v77);
			if (v32 != (int*)31) {
				v83[0] = *v32;
				v83[1] = v32[1];
				v40 = v32[2];
				v32 = v83;
				v83[2] = v40;
			}
			v77.field_0 -= 34;
			v77.field_4 -= 34;
			wall_light_end = (int*)sub_469920(&v77);
			break;
		case 10:
			v41 = v3[6];
			v77.field_0 = 23 * v3[5];
			v77.field_4 = 23 * (v41 + 1);
			v32 = (int*)sub_469920(&v77);
			if (v32 != (int*)31) {
				v83[0] = *v32;
				v83[1] = v32[1];
				v42 = v32[2];
				v32 = v83;
				v83[2] = v42;
			}
			v77.field_0 += 11;
			v77.field_4 -= 11;
			wall_light_end = (int*)sub_469920(&v77);
			break;
		default:
			v43 = v3[6];
			v77.field_0 = 23 * v3[5];
			v77.field_4 = 23 * (v43 + 1);
			v32 = (int*)sub_469920(&v77);
			if (v32 != (int*)31) {
				v83[0] = *v32;
				v83[1] = v32[1];
				v44 = v32[2];
				v32 = v83;
				v83[2] = v44;
			}
			v77.field_0 += 23;
			wall_light_end = (int*)sub_469920(&v77);
			break;
		}
		nox_xxx_getWallDrawOffset_46A3F0(v3[1], v84, v3[2], v73, &v45x, &v45y);
		v46 = v82 + v45x - 51;
		v47 = -73 - v45y + v7;
		sub_4345F0(1);
		v48 = *((uint8_t*)v32 + 8);
		LOBYTE(v49) = *((uint8_t*)v32 + 4);
		LOBYTE(v50) = *(uint8_t*)v32;
		nox_draw_setColorMultAndIntensityRGB_433CD0(v50, v49, v48);
		if (!(v72 & 2)) {
			v69 = v72;
			v66 = a4;
			v65 = a3;
			v64 = nox_win_height;
			v52 = nox_xxx_getWallSprite_46A3B0(v3[1], v84, v3[2], v73);
			nox_xxx_edgeDraw_480EF0(v52, v46, v47, v32, wall_light_end, v64, v65, v66, 0, v69);
			goto LABEL_106;
		}
		if (!sub_47D380(a3, a4)) {
			goto LABEL_106;
		}
		nox_client_drawEnableAlpha_434560(1);
		nox_client_drawSetAlpha_434580(0x80u);
		sub_47D400(nox_client_highResFrontWalls_80820 == 0, (int)vp->field_5);
		v68 = v47;
		v67 = v46;
		v51 = nox_xxx_getWallSprite_46A3B0(v3[1], v84, v3[2], v73);
	} else {
		v53 = v3[6];
		v77.field_0 = 23 * v3[5] + 11;
		v77.field_4 = 23 * v53 + 11;
		v54 = (uint8_t*)(sub_469920(&v77));
		nox_xxx_getWallDrawOffset_46A3F0(v3[1], v84, v3[2], v73, &v55x, &v55y);
		v56 = v82 + v55x - 50;
		v57 = -72 - v55y + v7;
		sub_4345F0(1);
		LOBYTE(v59) = v54[8];
		v58 = v54[4];
		LOBYTE(v60) = *v54;
		nox_draw_setColorMultAndIntensityRGB_433CD0(v60, v58, v59);
		if (!(v72 & 2)) {
			if (sub_47D380(a3, a4)) {
				sub_47D400(nox_client_highResFrontWalls_80820 == 0, (int)vp->field_5);
				v61 = nox_xxx_getWallSprite_46A3B0(v3[1], v84, v3[2], v73);
				nox_client_drawImageAt_47D2C0(v61, v56, v57);
				sub_47D400(0, 0);
			}
			goto LABEL_106;
		}
		if (!sub_47D380(a3, a4)) {
			goto LABEL_106;
		}
		nox_client_drawEnableAlpha_434560(1);
		nox_client_drawSetAlpha_434580(0x80u);
		sub_47D400(nox_client_highResFrontWalls_80820 == 0, (int)vp->field_5);
		v68 = v57;
		v67 = v56;
		v51 = nox_xxx_getWallSprite_46A3B0(v3[1], v84, v3[2], v73);
	}
	nox_client_drawImageAt_47D2C0(v51, v67, v68);
	sub_47D400(0, 0);
	nox_client_drawEnableAlpha_434560(0);
LABEL_106:
	sub_4345F0(0);
	v3[3] = 0;
	v3[4] &= 0xFC;
	*((uint32_t*)v3 + 3) = 1;
	return;
}
// 474366: variable 'v50' is possibly undefined
// 474366: variable 'v49' is possibly undefined
// 4744A3: variable 'v60' is possibly undefined
// 4744A3: variable 'v59' is possibly undefined

//----- (00474B40) --------------------------------------------------------
int sub_474B40(nox_drawable* dr) {
	nox_object_team_t* v1; // edi
	nox_object_team_t* v2; // eax
	nox_drawable* local_player;

	v1 = nox_xxx_objGetTeamByNetCode_418C80(nox_player_netCode_85319C);
	if (v1) {
		v2 = nox_xxx_objGetTeamByNetCode_418C80(dr->field_32);
		if (v2) {
			if (nox_player_netCode_85319C == dr->field_32 || nox_xxx_servCompareTeams_419150(v1, v2)) {
				return 1;
			}
		}
	}
	local_player = getMemPtr(0x852978, 8);
	if (dr == local_player) {
		return 1;
	}
	if (local_player && nox_client_drawable_testBuff_4356C0(local_player, 21)) {
		return 1;
	}
	if (dr->flags28 & 4) {
		if (dr != local_player) {
			nox_common_playerInfoGetByID_417040(dr->field_32);
		}
	}
	return 0;
}

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

int nox_xxx_drawAllMB_475810_draw_B(nox_draw_viewport_t* vp) {
	int v10 = 1;
	int v11;
	float2 v38;
	if (nox_common_getEngineFlag(NOX_ENGINE_FLAG_DISABLE_FLOOR_RENDERING) ||
		(v38.field_0 = (double)vp->field_6, v38.field_4 = (double)vp->field_7,
		 v11 = nox_xxx_tileNFromPoint_411160(&v38), v11 == 255) ||
		v11 == -1) {
		v10 = 0;
	}
	return v10;
}
