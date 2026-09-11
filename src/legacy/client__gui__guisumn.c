#include "client__gui__guisumn.h"
#include "client__gui__window.h"
#include "common__strman.h"

#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME3_1.h"
#include "GAME3_3.h"
#include "client__gui__gamewin__gamewin.h"
#include "input_common.h"
#include "operators.h"

extern uint32_t dword_5d4594_1321196;
extern nox_window* dword_5d4594_1321036;
extern nox_video_bag_image_t* dword_5d4594_1321024;
extern uint32_t dword_5d4594_1320988;
extern uint32_t dword_5d4594_1321208;
extern nox_window* dword_5d4594_1321032;
extern nox_window* dword_5d4594_1321044;
extern uint32_t nox_xxx_screenWidth_587000_184452;
extern uint32_t dword_5d4594_1320992;
extern nox_window* dword_5d4594_1321040;
extern int nox_win_width;
extern int nox_win_height;

extern uint32_t nox_color_white_2523948;
extern uint32_t nox_color_blue_2650684;
extern uint32_t nox_color_yellow_2589772;
extern uint32_t nox_color_black_2650656;

//----- (004C1D80) --------------------------------------------------------
int nox_xxx_guiSummonCreatureLoad_4C1D80(void) {
	wchar2_t* v0; // eax
	nox_window* v1; // esi
	nox_video_bag_image_t* v2; // eax
	nox_window* v3; // eax
	nox_window* v4; // eax

	*getMemU32Ptr(0x5D4594, 1321004) = 0;
	*getMemU32Ptr(0x5D4594, 1321000) = -145;
	dword_5d4594_1320988 = nox_win_width - 95;
	dword_5d4594_1320992 = -145;
	dword_5d4594_1321032 = nox_window_new(0, 8, nox_win_width - 95, -145, 87, 115, 0);
	nox_window_set_all_funcs(dword_5d4594_1321032, sub_4C2BD0, sub_4C24A0, 0);
	dword_5d4594_1321036 = nox_window_new(dword_5d4594_1321032, 136, 5, 38, 76, 76, 0);
	nox_window_set_all_funcs(dword_5d4594_1321036, nox_xxx_wndSummonProc_4C2B10,
							 nox_xxx_guiDrawSummonBox_4C1FE0, sub_4C2C20);
	v0 = nox_strman_loadString_40F1D0("ToolTipSummon", 0, "C:\\NoxPost\\src\\Client\\Gui\\guisumn.c", 818);
	nox_xxx_wndWddSetTooltip_46B000(&dword_5d4594_1321036->draw_data, v0);
	setMemPtr(0x5D4594, 1320996, nox_xxx_gLoadImg_42F970("CreatureCageBottom"));
	v1 = nox_window_new(dword_5d4594_1321036, 160, 0, 0, 1, 1, 0);
	v2 = nox_xxx_gLoadImg_42F970("CreatureCageTop");
	nox_xxx_wndSetIcon_46AE60(v1, v2);
	nox_xxx_wndSetOffsetMB_46AE40(v1, -5, -38);
	v3 = nox_window_new(dword_5d4594_1321032, 8, 19, 0, 48, 39, 0);
	nox_window_set_all_funcs(v3, sub_4C2BE0, sub_4C24A0, 0);
	setMemPtr(0x5D4594, 1321008, nox_xxx_gLoadImg_42F970("CreatureCageHuntButtonLit"));
	setMemPtr(0x5D4594, 1321012, nox_xxx_gLoadImg_42F970("CreatureCageHuntButton"));
	setMemPtr(0x5D4594, 1321016, nox_xxx_gLoadImg_42F970("CreatureCageGuardButtonLit"));
	setMemPtr(0x5D4594, 1321020, nox_xxx_gLoadImg_42F970("CreatureCageGuardButton"));
	dword_5d4594_1321024 = nox_xxx_gLoadImg_42F970("CreatureCageEscortButtonLit");
	setMemPtr(0x5D4594, 1321028, nox_xxx_gLoadImg_42F970("CreatureCageEscortButton"));
	v4 = nox_window_new(0, 168, dword_5d4594_1320988 + 27, dword_5d4594_1320992 + 12, 34, 34, 0);
	dword_5d4594_1321040 = v4;
	v4->draw_data.group |= 0x100;
	nox_window_set_all_funcs(dword_5d4594_1321040, nox_xxx_wndSummonBigButtonProc_4C24B0, 0, sub_4C2CE0);
	nox_xxx_wndSetIcon_46AE60(dword_5d4594_1321040, getMemPtr(0x5D4594, 1321028));
	sub_46AEC0(dword_5d4594_1321040, dword_5d4594_1321024);
	nox_xxx_wndSetIconLit_46AEA0(dword_5d4594_1321040, dword_5d4594_1321024);
	nox_xxx_wndSetOffsetMB_46AE40(dword_5d4594_1321040, -27, -12);
	*getMemU8Ptr(0x5D4594, 1321200) = 0;
	nox_window_set_hidden(dword_5d4594_1321032, 1);
	nox_window_set_hidden(dword_5d4594_1321040, 1);
	for (int off = 1321060; off < 1321188; off += 32) {
		*getMemU32Ptr(0x5D4594, off) = 0;
	}
	sub_4C2BF0();
	dword_5d4594_1321044 = 0;
	dword_5d4594_1321204 = NULL;
	dword_5d4594_1321196 = 0;
	return 1;
}

//----- (004C2560) --------------------------------------------------------
void nox_xxx_wndSummonCreateList_4C2560(int2* a1) {
	int v3;  // esi
	bool v4; // sf
	int v5;  // ecx
	int v6;  // edi
	int v7;  // eax
	int v8;  // edi
	int v12; // [esp+10h] [ebp-8h]
	int v13; // [esp+14h] [ebp-4h]

	nox_xxx_screenWidth_587000_184452 = 0;
	for (int i = 0; i < 6; ++i) {
		if (i != 2) {
			const char* key = getMemPtr(0x587000, 184344 + 4 * i);
			wchar2_t* text = nox_strman_loadString_40F1D0(key, 0, "C:\\NoxPost\\src\\Client\\Gui\\guisumn.c", 588);
			nox_xxx_drawGetStringSize_43F840(0, text, &v12, &v13, nox_win_width);
			if (nox_xxx_screenWidth_587000_184452 < v12) {
				nox_xxx_screenWidth_587000_184452 = v12;
			}
		}
	}
	nox_xxx_screenWidth_587000_184452 += 8;
	v3 = nox_xxx_guiFontHeightMB_43F320(0) + 2;
	v5 = a1->field_0 - nox_xxx_screenWidth_587000_184452 / 2;
	v4 = v5 < 0;
	v12 = a1->field_0 - nox_xxx_screenWidth_587000_184452 / 2;
	v6 = 5 * v3 + 12;
	if (v4) {
		v5 = 0;
	} else {
		if (nox_xxx_screenWidth_587000_184452 + v5 < nox_win_width) {
			goto LABEL_11;
		}
		v5 = nox_win_width - nox_xxx_screenWidth_587000_184452 - 1;
	}
	v12 = v5;
LABEL_11:
	v7 = a1->field_4 - v6 / 2;
	v13 = v7;
	if (v7 < 0) {
		v7 = 0;
		v13 = v7;
		goto LABEL_16;
	}
	if (v6 + v7 >= nox_win_height) {
		v7 = nox_win_width - v6 - 1;
		v13 = v7;
	}
LABEL_16:
	dword_5d4594_1321044 = nox_window_new(0, 40, v5, v7, *(int*)&nox_xxx_screenWidth_587000_184452, 5 * v3 + 12, 0);
	nox_window_set_all_funcs(dword_5d4594_1321044, 0, sub_4C26F0, 0);
	nox_xxx_wndShowModalMB_46A8C0(dword_5d4594_1321044);
	v8 = 0;
	for (int i = 0; i < 6; ++i) {
		if (i != 2) {
			nox_window* entry = nox_window_new(dword_5d4594_1321044, 8, 0, v8,
										  (int)nox_xxx_screenWidth_587000_184452, v3 + 1, 0);
			nox_window_set_all_funcs(entry, nox_xxx_clientOrderCreature_4C2A60, sub_4C27F0, 0);
			entry->widget_data = (void*)(uintptr_t)i;
			v8 += v3 + 2;
		}
	}
	nox_xxx_clientPlaySoundSpecial_452D80(791, 100);
}

//----- (004C27F0) --------------------------------------------------------
int sub_4C27F0(nox_window* win, nox_window_data* draw) {
	(void)draw;
	uintptr_t command = (uintptr_t)win->widget_data;
	unsigned int v10; // [esp+0h] [ebp-Ch]
	unsigned int v11; // [esp+4h] [ebp-8h]
	int v12;          // [esp+8h] [ebp-4h]

	if (!dword_5d4594_1321208) {
		dword_5d4594_1321208 = nox_xxx_getNameId_4E3AA0("CarnivorousPlant");
	}
	if (!dword_5d4594_1321204 && command == 1) {
		return 1;
	}
	const char* key = getMemPtr(0x587000, 184344 + 4 * command);
	wchar2_t* text = nox_strman_loadString_40F1D0(key, 0, "C:\\NoxPost\\src\\Client\\Gui\\guisumn.c", 446);
	nox_client_wndGetPosition_46AA60(win, &v11, &v10);
	nox_xxx_drawGetStringSize_43F840(0, text, &v12, 0, 0);
	nox_point mpos = nox_client_getMousePos_4309F0();
	nox_xxx_guiFontHeightMB_43F320(0);
	int text_x = ((int)nox_xxx_screenWidth_587000_184452 - v12) / 2 + 1;
	if (nox_xxx_wndPointInWnd_46AAB0(win, mpos.x, mpos.y)) {
		sub_4C2A00((int)v11 + text_x, (int)v10 + 3, nox_color_yellow_2589772, nox_color_black_2650656,
				   (short*)text);
		if (command != *getMemU32Ptr(0x587000, 184552)) {
			*getMemU32Ptr(0x587000, 184552) = (uint32_t)command;
			nox_xxx_clientPlaySoundSpecial_452D80(920, 100);
		}
		return 1;
	}
	if (dword_5d4594_1321204) {
		if (sub_4C2DD0(dword_5d4594_1321204) || (command != 4 && command != 5)) {
			sub_4C2A00((int)v11 + text_x, (int)v10 + 3, nox_color_white_2523948, nox_color_black_2650656,
					   (short*)text);
			return 1;
		}
		sub_4C2A00((int)v11 + text_x, (int)v10 + 3, *getMemU32Ptr(0x85B3FC, 956), nox_color_black_2650656,
				   (short*)text);
		return 1;
	}
	if (command != 4 && command != 5) {
		sub_4C2A00((int)v11 + text_x, (int)v10 + 3, nox_color_blue_2650684, nox_color_black_2650656,
				   (short*)text);
		return 1;
	}
	uint32_t color = sub_4C2E00() ? nox_color_blue_2650684 : *getMemU32Ptr(0x85B3FC, 956);
	sub_4C2A00((int)v11 + text_x, (int)v10 + 3, color, nox_color_black_2650656, (short*)text);
	return 1;
}

//----- (004C2CE0) --------------------------------------------------------
int sub_4C2CE0(nox_window* win, nox_window_data* draw, uintptr_t packed_position) {
	(void)win;
	(void)draw;
	(void)packed_position;
	wchar2_t* v0; // eax
	wchar2_t* v2; // eax
	wchar2_t* v3; // eax

	switch (*getMemU32Ptr(0x587000, 184448)) {
	case 3:
		v3 = nox_strman_loadString_40F1D0("ccs:GUARD", 0, "C:\\NoxPost\\src\\Client\\Gui\\guisumn.c", 785);
		nox_xxx_cursorSetTooltip_4776B0(v3);
		break;
	case 4:
		v2 = nox_strman_loadString_40F1D0("ccs:ESCORT", 0, "C:\\NoxPost\\src\\Client\\Gui\\guisumn.c", 782);
		nox_xxx_cursorSetTooltip_4776B0(v2);
		return 1;
	case 5:
		v0 = nox_strman_loadString_40F1D0("ccs:HUNT", 0, "C:\\NoxPost\\src\\Client\\Gui\\guisumn.c", 788);
		nox_xxx_cursorSetTooltip_4776B0(v0);
		return 1;
	}
	return 1;
}
