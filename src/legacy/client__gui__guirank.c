#include "client__gui__guirank.h"
#include "client__gui__window.h"
#include "common__strman.h"

#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "client__gui__gadgets__listbox.h"
#include "common__system__team.h"
#include "operators.h"

extern uintptr_t dword_8531A0_2576;
extern uint32_t dword_8531A0_2572;
extern uint32_t dword_587000_145672;
extern nox_window* dword_5d4594_1090108;
extern nox_window* dword_5d4594_1090112;
extern uint32_t dword_587000_145668;
extern uint32_t dword_5d4594_1090040;
extern uint32_t dword_587000_145664;
extern uint32_t dword_5d4594_1090044;
extern nox_window* dword_5d4594_1090100;
extern uint32_t nox_player_netCode_85319C;
extern uint32_t dword_5d4594_1090120;
extern int nox_win_width;
extern uint32_t nox_color_white_2523948;
extern uint32_t nox_color_yellow_2589772;

nox_window* dword_5d4594_1090048 = 0;

enum { NOX_RANK_WINDOW_FIRST = 1090052, NOX_RANK_WINDOW_LAST = 1090112 };
static nox_window* nox_rank_windows[(NOX_RANK_WINDOW_LAST - NOX_RANK_WINDOW_FIRST) / 4 + 1];
static wchar2_t* nox_rank_class_names[3];

nox_window* nox_rank_window_at(int offset) {
	if (offset < NOX_RANK_WINDOW_FIRST || offset > NOX_RANK_WINDOW_LAST || (offset - NOX_RANK_WINDOW_FIRST) % 4) {
		return NULL;
	}
	return nox_rank_windows[(offset - NOX_RANK_WINDOW_FIRST) / 4];
}

const wchar2_t* nox_rank_class_name(unsigned int class_ind) {
	if (class_ind < 3 && nox_rank_class_names[class_ind]) {
		return nox_rank_class_names[class_ind];
	}
	return nox_strman_loadString_40F1D0(
		"InternalError", 0, "C:\\NoxPost\\src\\client\\Gui\\guirank.c", 1050);
}

static void nox_rank_set_window(int offset, nox_window* win) {
	if (offset < NOX_RANK_WINDOW_FIRST || offset > NOX_RANK_WINDOW_LAST || (offset - NOX_RANK_WINDOW_FIRST) % 4) {
		return;
	}
	nox_rank_windows[(offset - NOX_RANK_WINDOW_FIRST) / 4] = win;
}

//----- (0046E870) --------------------------------------------------------
nox_window* nox_xxx_guiDrawRank_46E870() {
	int v0;              // ebx
	unsigned short* v1;  // eax
	int v2;              // eax
	unsigned short* v3;  // eax
	unsigned short* v4;  // eax
	int v5;              // eax
	unsigned short* v6;  // eax
	int v7;              // eax
	unsigned short* v8;  // eax
	unsigned short* v9;  // eax
	unsigned short* v10; // eax
	unsigned short* v11; // eax
	unsigned short* v12; // eax
	int v13;             // eax
	int v14;             // ecx
	int v16;             // edi
	int v17;             // esi
	int v18;             // ebp
	int v19;             // edi
	int v20;             // ebx
	nox_window* v21;     // eax
	int v22;             // edx
	int v23;             // ecx
	nox_window* v24;     // eax
	int v25;             // edx
	nox_window* v26;     // eax
	int v27;             // edx
	nox_window* v28;     // eax
	int v29;             // ecx
	int v30;             // esi
	nox_window* result;  // eax
	int v32;             // [esp-4h] [ebp-1DCh]
	int v33;             // [esp-4h] [ebp-1DCh]
	int v34;             // [esp-4h] [ebp-1DCh]
	int v35;             // [esp-4h] [ebp-1DCh]
	int v36;             // [esp-4h] [ebp-1DCh]
	int v37;             // [esp-4h] [ebp-1DCh]
	int v38;             // [esp-4h] [ebp-1DCh]
	int v39;             // [esp-4h] [ebp-1DCh]
	int v40;             // [esp-4h] [ebp-1DCh]
	int v41;             // [esp+10h] [ebp-1C8h]
	int v42;             // [esp+14h] [ebp-1C4h]
	int v43;             // [esp+18h] [ebp-1C0h]
	int v44;             // [esp+1Ch] [ebp-1BCh]
	int v46;             // [esp+2Ch] [ebp-1ACh]
	int v47;             // [esp+30h] [ebp-1A8h]
	nox_scrollListBox_data opts;
	nox_window_data draw;
	nox_staticText_data text_data;

	dword_587000_145668 = 6;
	sub_46F030();
	v0 = nox_xxx_guiFontHeightMB_43F320(0);
	v46 = v0;
	*getMemU32Ptr(0x5D4594, 1084036) = 80;

	v32 = nox_win_width;
	v1 = nox_strman_loadString_40F1D0("Flag", 0, "C:\\NoxPost\\src\\client\\Gui\\guirank.c", 1641);
	nox_xxx_drawGetStringSize_43F840(0, v1, &v41, &v44, v32);

	v2 = v41;
	if (v41 < 18) {
		v2 = 18;
		v41 = 18;
	}
	*getMemU32Ptr(0x5D4594, 1084040) = v2 + 14;
	v33 = nox_win_width;
	v3 = nox_strman_loadString_40F1D0("Score", 0, "C:\\NoxPost\\src\\client\\Gui\\guirank.c", 1649);
	nox_xxx_drawGetStringSize_43F840(0, v3, &v41, &v44, v33);
	v34 = nox_win_width;
	v4 = nox_strman_loadString_40F1D0("HealthHeading", 0, "C:\\NoxPost\\src\\client\\Gui\\guirank.c", 1650);
	nox_xxx_drawGetStringSize_43F840(0, v4, &v42, &v43, v34);
	v5 = v41;
	if (v42 > v41) {
		v5 = v42;
		v41 = v42;
	}
	*getMemU32Ptr(0x5D4594, 1084048) = v5 + 7;
	v35 = nox_win_width;
	v6 = nox_strman_loadString_40F1D0("Ping", 0, "C:\\NoxPost\\src\\client\\Gui\\guirank.c", 1657);
	nox_xxx_drawGetStringSize_43F840(0, v6, &v41, &v44, v35);
	nox_xxx_drawGetStringSize_43F840(0, getMemU16Ptr(0x587000, 145972), &v42, &v43, nox_win_width);
	v7 = v41;
	if (v42 > v41) {
		v7 = v42;
		v41 = v42;
	}
	*getMemU32Ptr(0x5D4594, 1084052) = v7 + 7;
	v36 = nox_win_width;
	v8 = nox_strman_loadString_40F1D0("Class", 0, "C:\\NoxPost\\src\\client\\Gui\\guirank.c", 1667);
	nox_xxx_drawGetStringSize_43F840(0, v8, &v41, &v44, v36);
	v37 = nox_win_width;
	v9 = nox_strman_loadString_40F1D0("Warrior", 0, "C:\\NoxPost\\src\\client\\Gui\\guirank.c", 1668);
	nox_xxx_drawGetStringSize_43F840(0, v9, &v42, &v43, v37);
	if (v42 > v41) {
		v41 = v42;
	}
	v38 = nox_win_width;
	v10 = nox_strman_loadString_40F1D0("Wizard", 0, "C:\\NoxPost\\src\\client\\Gui\\guirank.c", 1671);
	nox_xxx_drawGetStringSize_43F840(0, v10, &v42, &v43, v38);
	if (v42 > v41) {
		v41 = v42;
	}
	v39 = nox_win_width;
	v11 = nox_strman_loadString_40F1D0("Conjurer", 0, "C:\\NoxPost\\src\\client\\Gui\\guirank.c", 1674);
	nox_xxx_drawGetStringSize_43F840(0, v11, &v42, &v43, v39);
	if (v42 > v41) {
		v41 = v42;
	}
	v40 = nox_win_width;
	v12 = nox_strman_loadString_40F1D0("LivesHeading", 0, "C:\\NoxPost\\src\\client\\Gui\\guirank.c", 1677);
	nox_xxx_drawGetStringSize_43F840(0, v12, &v42, &v43, v40);
	v13 = v41;
	if (v42 > v41) {
		v13 = v42;
		v41 = v42;
	}
	*getMemU32Ptr(0x5D4594, 1084044) = v13 + 7;
	dword_5d4594_1090040 = 0;
	for (int i = 0; i < 5; ++i) {
		dword_5d4594_1090040 += *getMemU32Ptr(0x5D4594, 1084036 + 4 * i);
	}
	dword_5d4594_1090044 = 439 - v0;
	dword_5d4594_1090048 = nox_window_new(0, 1560, 0, v0 + 40, 1, 1, 0);
	if (!dword_5d4594_1090048) {
		return NULL;
	}
	nox_window_set_all_funcs(dword_5d4594_1090048, sub_46F060, sub_46F080, 0);
	dword_5d4594_1090048->draw_data.bg_color = 0x80000000;
	dword_5d4594_1090048->draw_data.en_color = 0x80000000;
	dword_5d4594_1090048->draw_data.hl_color = 0x80000000;
	dword_5d4594_1090048->draw_data.dis_color = 0x80000000;
	dword_5d4594_1090048->draw_data.sel_color = 0x80000000;
	memset(&draw, 0, sizeof(draw));
	nox_wcscpy(draw.text, (const wchar2_t*)getMemAt(0x5D4594, 1090136));
	draw.text_color = *getMemU32Ptr(0x85B3FC, 940);
	draw.bg_color = 0x80000000;
	draw.en_color = 0x80000000;
	draw.hl_color = 0x80000000;
	draw.dis_color = 0x80000000;
	draw.sel_color = 0x80000000;
	memset(&opts, 0, sizeof(opts));
	opts.count = 64;
	opts.line_height = (uint16_t)(v0 + 1);
	opts.field_2 = 1;
	v17 = 0;
	v47 = 3 * v0;
	v18 = 3 * v0 + 1;
	draw.style = 32;
	v19 = 2 * (v0 + 1);
	v20 = 2 * v0;
	do {
		v21 = nox_gui_newScrollListBox_4A4310(dword_5d4594_1090048, 1088, v17 * dword_5d4594_1090040,
												  v18, dword_5d4594_1090040, dword_5d4594_1090044 - v19, &draw, &opts);
		v22 = *getMemU32Ptr(0x5D4594, 1084036);
		v23 = dword_5d4594_1090044 - v19;
		nox_rank_set_window(1090052 + 4 * v17, v21);
		v24 = nox_gui_newScrollListBox_4A4310(v21, 1088, 0, v20, v22, v23, &draw, &opts);
		v25 = dword_5d4594_1090044;
		nox_rank_set_window(1090060 + 4 * v17, v24);
		v26 = nox_gui_newScrollListBox_4A4310(nox_rank_window_at(1090052 + 4 * v17), 1088,
												  *getMemIntPtr(0x5D4594, 1084036), v20, *getMemIntPtr(0x5D4594, 1084040),
												  v25 - v19, &draw, &opts);
		v27 = dword_5d4594_1090044;
		nox_rank_set_window(1090068 + 4 * v17, v26);
		nox_rank_set_window(
			1090076 + 4 * v17,
			nox_gui_newScrollListBox_4A4310(nox_rank_window_at(1090052 + 4 * v17), 1088,
													*getMemU32Ptr(0x5D4594, 1084036) + *getMemU32Ptr(0x5D4594, 1084040), v20,
													*getMemIntPtr(0x5D4594, 1084044), v27 - v19, &draw, &opts));
		v28 = nox_gui_newScrollListBox_4A4310(nox_rank_window_at(1090052 + 4 * v17), 1088,
			*getMemU32Ptr(0x5D4594, 1084036) + *getMemU32Ptr(0x5D4594, 1084040) + *getMemU32Ptr(0x5D4594, 1084044), v20,
			*getMemIntPtr(0x5D4594, 1084048), dword_5d4594_1090044 - v19, &draw, &opts);
		v29 = dword_5d4594_1090044;
		nox_rank_set_window(1090084 + 4 * v17, v28);
		nox_rank_set_window(
			1090092 + 4 * v17,
			nox_gui_newScrollListBox_4A4310(nox_rank_window_at(1090052 + 4 * v17), 1088,
													*getMemU32Ptr(0x5D4594, 1084036) + *getMemU32Ptr(0x5D4594, 1084040) +
														*getMemU32Ptr(0x5D4594, 1084048) + *getMemU32Ptr(0x5D4594, 1084044),
													v20, *getMemIntPtr(0x5D4594, 1084052), v29 - v19, &draw, &opts));
		nox_xxx_wndSetProc_46B2C0(nox_rank_window_at(1090052 + 4 * v17), nox_xxx_Proc_46F070);
		for (int offset = 1090060; offset <= 1090092; offset += 8) {
			sub_46B120(nox_rank_window_at(offset + 4 * v17), nox_rank_window_at(1090052 + 4 * v17));
		}
		++v17;
	} while (v17 < 2);
	draw.style = 2048;
	draw.text_color = nox_color_yellow_2589772;
	memset(&text_data, 0, sizeof(text_data));
	text_data.text = nox_strman_loadString_40F1D0("yourrank", 0, "C:\\NoxPost\\src\\client\\Gui\\guirank.c", 1772);
	v30 = v46 + 1;
	dword_5d4594_1090100 = nox_gui_newStaticText_489300(dword_5d4594_1090048, 1088, 0, v46,
																dword_5d4594_1090040, v46 + 1, &draw, &text_data);
	nox_rank_set_window(1090100, dword_5d4594_1090100);
	draw.text_color = nox_color_white_2523948;
	text_data.text = nox_strman_loadString_40F1D0("WindowDir:Empty", 0, "C:\\NoxPost\\src\\client\\Gui\\guirank.c", 1782);
	dword_5d4594_1090112 = nox_gui_newStaticText_489300(dword_5d4594_1090048, 1088, 0, v20,
																dword_5d4594_1090040, v30, &draw, &text_data);
	nox_rank_set_window(1090112, dword_5d4594_1090112);
	draw.text_color = *getMemU32Ptr(0x85B3FC, 940);
	text_data.text = nox_strman_loadString_40F1D0("WindowDir:Empty", 0, "C:\\NoxPost\\src\\client\\Gui\\guirank.c", 1790);
	dword_5d4594_1090108 = nox_gui_newStaticText_489300(dword_5d4594_1090048, 1088, 0, v47,
																dword_5d4594_1090040, v30, &draw, &text_data);
	nox_rank_set_window(1090108, dword_5d4594_1090108);
	draw.text_color = dword_8531A0_2572;
	text_data.text = nox_strman_loadString_40F1D0("TeamPlayerRank", 0, "C:\\NoxPost\\src\\client\\Gui\\guirank.c", 1798);
	nox_rank_set_window(1090104, nox_gui_newStaticText_489300(dword_5d4594_1090048, 1088, 0, 0,
																					 dword_5d4594_1090040, v30, &draw, &text_data));
	result = dword_5d4594_1090048;
	dword_587000_145664 = 1;
	return result;
}

//----- (0046F030) --------------------------------------------------------
wchar2_t* sub_46F030() {
	int i;           // esi
	wchar2_t* result; // eax

	for (i = 0; i < 12; i += 4) {
		result = nox_strman_loadString_40F1D0((char*)getMemPtr(0x587000, 145676 + i), 0,
											  "C:\\NoxPost\\src\\client\\Gui\\guirank.c", 167);
		nox_rank_class_names[i / 4] = result;
	}
	return result;
}

//----- (0046FC50) --------------------------------------------------------
char sub_46FC50() {
	int result = 0;
	if (!dword_5d4594_1090108) {
		return 0;
	}
	if (sub_40A220() && (!nox_common_gameFlags_check_40A5C0(1) || sub_40A300() ||
						 sub_40A180(nox_common_gameFlags_getVal_40A5B0()))) {
		if (!nox_common_gameFlags_check_40A5C0(1) || sub_40A300() ||
			sub_40A180(nox_common_gameFlags_getVal_40A5B0())) {
			if (wndIsShown_nox_xxx_wndIsShown_46ACC0(dword_5d4594_1090108)) {
				nox_window_set_hidden(dword_5d4594_1090108, 0);
			}
			int remaining = sub_40A230();
			wchar2_t* format = nox_strman_loadString_40F1D0(
				"TimeRemaining", 0, "C:\\NoxPost\\src\\client\\Gui\\guirank.c", 352);
			nox_swprintf((wchar2_t*)getMemAt(0x5D4594, 1084068), format, remaining / 60000,
				remaining % 60000 / 1000);
			result = (int)nox_window_call_field_94(dword_5d4594_1090108, 16385,
				(uintptr_t)getMemAt(0x5D4594, 1084068), 0);
		}
	} else {
		result = wndIsShown_nox_xxx_wndIsShown_46ACC0(dword_5d4594_1090108);
		if (!result) {
			result = nox_window_set_hidden(dword_5d4594_1090108, 1);
		}
	}
	return (char)result;
}

//----- (0046FD80) --------------------------------------------------------
int sub_46FD80() {
	if (!dword_5d4594_1090112) {
		return 0;
	}
	int result;
	if (nox_common_gameFlags_check_40A5C0(4224)) {
		result = wndIsShown_nox_xxx_wndIsShown_46ACC0(dword_5d4594_1090112);
		if (!result) {
			result = nox_window_set_hidden(dword_5d4594_1090112, 1);
		}
	} else {
		if (wndIsShown_nox_xxx_wndIsShown_46ACC0(dword_5d4594_1090112)) {
			nox_window_set_hidden(dword_5d4594_1090112, 0);
		}
		int limit;
		if (nox_common_gameFlags_check_40A5C0(1)) {
			limit = (uint16_t)nox_xxx_servGamedataGet_40A020(nox_common_gameFlags_getVal_40A5B0());
		} else {
			limit = *((uint16_t*)nox_xxx_cliGamedataGet_416590(0) + 27);
		}
		wchar2_t* format = nox_strman_loadString_40F1D0(
			"LessonLimit", 0, "C:\\NoxPost\\src\\client\\Gui\\guirank.c", 390);
		nox_swprintf((wchar2_t*)getMemAt(0x5D4594, 1083972), format, limit);
		result = (int)nox_window_call_field_94(dword_5d4594_1090112, 16385,
			(uintptr_t)getMemAt(0x5D4594, 1083972), 0);
	}
	return result;
}
