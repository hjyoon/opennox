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

//----- (0046DC60) --------------------------------------------------------
int sub_46DC60(nox_window* win, unsigned char color, const wchar2_t* text) {
	if (!text) {
		text = nox_strman_loadString_40F1D0("InternalError", 0, "C:\\NoxPost\\src\\client\\Gui\\guirank.c", 1050);
	}
	return text ? sub_46DC00(win, color, text) : 0;
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

static nox_scrollListBox_data* nox_rank_list_data(int offset) {
	nox_window* win = nox_rank_window_at(offset);
	return win ? (nox_scrollListBox_data*)win->widget_data : NULL;
}

static unsigned int nox_rank_list_row(int offset) {
	nox_scrollListBox_data* data = nox_rank_list_data(offset);
	return data ? data->field_11_1 : 0;
}

//----- (0046F080) --------------------------------------------------------
int sub_46F080(nox_window* win, nox_window_data* draw) {
	if (!win || !draw) {
		return 0;
	}
	if (!nox_common_gameFlags_check_40A5C0(8) && dword_587000_145668 != 6) {
		dword_5d4594_1090120 = dword_587000_145668 ? dword_587000_145668 - 1 : 5;
		sub_4703F0();
		dword_587000_145668 = 6;
		if (!dword_5d4594_1090120) {
			return 1;
		}
	}

	unsigned int x = 0;
	unsigned int y = 0;
	nox_client_wndGetPosition_46AA60(win, &x, &y);
	char* game_data = nox_xxx_cliGamedataGet_416590(0);
	int game_limit;
	if (nox_common_gameFlags_check_40A5C0(1)) {
		game_limit = (uint16_t)nox_xxx_servGamedataGet_40A020(nox_common_gameFlags_getVal_40A5B0());
	} else {
		game_limit = *((uint16_t*)game_data + 27);
	}
	if (!(win->flags & 0x80)) {
		if (draw->bg_color != 0x80000000) {
			nox_client_drawRectFilledAlpha_49CF10((int)x, (int)y, (int)win->width, (int)win->height);
		}
	} else {
		nox_client_drawImageAt_47D2C0((nox_video_bag_image_t*)draw->bg_image, (int)x, (int)y);
	}

	if (dword_587000_145664 || gameFrame() > *getMemU32Ptr(0x5D4594, 1090124) + gameFPS()) {
		set_dword_5d4594_3799468(1);
		*getMemU32Ptr(0x5D4594, 1090124) = gameFrame();
		dword_587000_145672 = UINT32_MAX;
		sub_46DB80();
		sub_46DCC0();
		dword_587000_145664 = 0;

		nox_team_t* local_team = NULL;
		nox_object_team_t* membership = nox_xxx_objGetTeamByNetCode_418C80(nox_player_netCode_85319C);
		if (membership) {
			local_team = nox_xxx_getTeamByID_418AB0(membership->id);
		}
		int show_rank = 0;
		nox_playerInfo* local_player = (nox_playerInfo*)dword_8531A0_2576;
		if (local_player && (!(local_player->field_3680 & 1) || local_player->field_3680 & 0x20)) {
			show_rank = 1;
		}

		int mode = dword_5d4594_1090120;
		uint8_t last_color = 9;
		unsigned int team_count = getMemByte(0x5D4594, 1090116);
		if (team_count && (mode == 2 || mode == 3)) {
			sub_46DC60(nox_rank_window_at(1090060), 9,
				nox_strman_loadString_40F1D0("team", 0, "C:\\NoxPost\\src\\client\\Gui\\guirank.c", 1338));
			sub_46DC60(nox_rank_window_at(1090068), 9, (wchar2_t*)getMemAt(0x587000, 146512));
			sub_46DC60(nox_rank_window_at(1090076), 9, (wchar2_t*)getMemAt(0x587000, 146516));
			sub_46DC60(nox_rank_window_at(1090084), 9,
				nox_strman_loadString_40F1D0("score", 0, "C:\\NoxPost\\src\\client\\Gui\\guirank.c", 1341));
			sub_46DC60(nox_rank_window_at(1090092), 9, (wchar2_t*)getMemAt(0x587000, 146564));
			for (unsigned int row = 0; row < team_count; ++row) {
				last_color = sub_46FEB0((uint8_t)row);
				unsigned int off = 56 * row;
				sub_46DC30(nox_rank_window_at(1090060), last_color, (wchar2_t*)getMemAt(0x587000, 146568),
					getMemAt(0x5D4594, 1087204 + off));
				sub_46DC30(nox_rank_window_at(1090068), last_color, (wchar2_t*)getMemAt(0x587000, 146576));
				sub_46DC30(nox_rank_window_at(1090076), last_color, (wchar2_t*)getMemAt(0x587000, 146580));
				sub_46DC30(nox_rank_window_at(1090084), last_color, (wchar2_t*)getMemAt(0x587000, 146584),
					*getMemU32Ptr(0x5D4594, 1087252 + off));
				sub_46DC30(nox_rank_window_at(1090092), last_color, (wchar2_t*)getMemAt(0x587000, 146592));
			}
			sub_46DC30(nox_rank_window_at(1090060), last_color, (wchar2_t*)getMemAt(0x587000, 146596));
			sub_46DC30(nox_rank_window_at(1090068), last_color, (wchar2_t*)getMemAt(0x587000, 146600));
			sub_46DC30(nox_rank_window_at(1090076), last_color, (wchar2_t*)getMemAt(0x587000, 146604));
			sub_46DC30(nox_rank_window_at(1090084), last_color, (wchar2_t*)getMemAt(0x587000, 146608));
			sub_46DC30(nox_rank_window_at(1090092), last_color, (wchar2_t*)getMemAt(0x587000, 146612));
		}

		unsigned int second_page_padding = nox_rank_list_row(1090060);
		unsigned int player_count = getMemByte(0x5D4594, 1090117);
		if (player_count && (mode == 2 || mode == 4 || mode == 5)) {
			sub_46F8F0(0, 0);
			unsigned int rows = mode == 4 && player_count > 3 ? 3 : player_count;
			for (unsigned int row = 0; row < rows; ++row) {
				int page = row >> 4;
				if (row == 16) {
					sub_46F8F0(1, (int)second_page_padding);
				}
				unsigned int off = 80 * row;
				int over_limit = nox_common_gameFlags_check_40A5C0(1024) && game_limit > 0 &&
					*getMemIntPtr(0x5D4594, 1084196 + off) >= game_limit;
				uint32_t player_flags = *getMemU32Ptr(0x5D4594, 1084208 + off);
				uint8_t color;
				if (!(player_flags & 1) || player_flags & 0x20 || over_limit) {
					int team_id = *getMemIntPtr(0x5D4594, 1084184 + off);
					if (team_id == -1) {
						color = (player_flags & 0x20 || over_limit) ? 2 : 3;
					} else {
						color = sub_46FEB0(sub_46FE60((uint32_t)team_id));
						if (player_flags & 0x20 || over_limit) {
							color -= 2;
						}
					}
				} else {
					color = 9;
				}

				if (*getMemU32Ptr(0x5D4594, 1084192 + off) == nox_player_netCode_85319C) {
					dword_587000_145672 = nox_rank_list_row(1090060 + 4 * page);
					*getMemU32Ptr(0x5D4594, 1088996) = page;
				}
				sub_46DC30(nox_rank_window_at(1090060 + 4 * page), color,
					(wchar2_t*)getMemAt(0x587000, 146616), getMemAt(0x5D4594, 1084132 + off));
				uint8_t player_class = getMemByte(0x5D4594, 1084188 + off);
				const wchar2_t* class_name = nox_rank_class_name(player_class);
				sub_46DC30(nox_rank_window_at(1090076 + 4 * page), color,
					(wchar2_t*)getMemAt(0x587000, 146624), class_name);
				uint32_t score = *getMemU32Ptr(0x5D4594, 1084196 + off);
				if (mode != 5 || score > 0) {
					sub_46DC30(nox_rank_window_at(1090084 + 4 * page), color,
						(wchar2_t*)getMemAt(0x587000, 146632), score);
				} else {
					sub_46DC30(nox_rank_window_at(1090084 + 4 * page), color,
						(wchar2_t*)getMemAt(0x587000, 146640));
				}
				sub_46DC30(nox_rank_window_at(1090092 + 4 * page), color,
					(wchar2_t*)getMemAt(0x587000, 146648), *getMemU32Ptr(0x5D4594, 1084200 + off));
				if (mode == 5) {
					nox_playerInfo* player = nox_common_playerInfoGetByID_417040(
						*getMemU32Ptr(0x5D4594, 1084192 + off));
					if (player) {
						sub_46DC30(nox_rank_window_at(1090068 + 4 * page), color,
							(wchar2_t*)getMemAt(0x587000, 146656), player->field_2096);
					}
				} else {
					uint8_t marker_color = 0;
					wchar2_t* marker = sub_46FB50(*getMemU32Ptr(0x5D4594, 1084204 + off), &marker_color);
					sub_46DC60(nox_rank_window_at(1090068 + 4 * page), marker_color, marker);
				}
			}
		} else if (mode == 1) {
			sub_46FFD0();
		}

		const wchar2_t* title = (wchar2_t*)getMemAt(0x587000, 147724);
		switch (mode) {
		case 1:
			title = nox_strman_loadString_40F1D0("Noxworld.c:Quest", 0,
				"C:\\NoxPost\\src\\client\\Gui\\guirank.c", 1487);
			break;
		case 2:
			title = nox_strman_loadString_40F1D0("TeamPlayerRank", 0,
				"C:\\NoxPost\\src\\client\\Gui\\guirank.c", 1475);
			break;
		case 3:
			title = nox_strman_loadString_40F1D0("Teams", 0, "C:\\NoxPost\\src\\client\\Gui\\guirank.c", 1519);
			if (local_team) {
				show_rank = 1;
			}
			break;
		case 4:
			title = nox_strman_loadString_40F1D0("Top3", 0, "C:\\NoxPost\\src\\client\\Gui\\guirank.c", 1479);
			break;
		case 5:
			title = nox_strman_loadString_40F1D0("WolRank", 0, "C:\\NoxPost\\src\\client\\Gui\\guirank.c", 1483);
			break;
		}
		if (!nox_common_gameFlags_check_40A5C0(1) ||
			!nox_common_getEngineFlag(NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING)) {
			if (mode == 1) {
				wchar2_t* label = nox_strman_loadString_40F1D0("Noxworld.c:Stage", 0,
					"C:\\NoxPost\\src\\client\\Gui\\guirank.c", 1499);
				nox_swprintf((wchar2_t*)getMemAt(0x5D4594, 1086692), L"%s %d", label,
					nox_gui_getQuestStage_450B10());
			} else if (mode == 3) {
				unsigned int rank = local_team ? (uint8_t)sub_46FF70(local_team->lessons) : 0;
				wchar2_t* label = nox_strman_loadString_40F1D0("yourteamrank", 0,
					"C:\\NoxPost\\src\\client\\Gui\\guirank.c", 1525);
				nox_swprintf((wchar2_t*)getMemAt(0x5D4594, 1086692), L"%s %d / %d", label, rank,
					getMemByte(0x5D4594, 1090116));
			} else if (mode == 2 || mode == 4 || mode == 5) {
				wchar2_t* label = nox_strman_loadString_40F1D0("yourrank", 0,
					"C:\\NoxPost\\src\\client\\Gui\\guirank.c", 1501);
				nox_swprintf((wchar2_t*)getMemAt(0x5D4594, 1086692), L"%s %d / %d", label,
					(uint8_t)sub_46FEE0(), getMemByte(0x5D4594, 1090118));
			}
		}
		if (show_rank) {
			if (wndIsShown_nox_xxx_wndIsShown_46ACC0(dword_5d4594_1090100)) {
				nox_window_set_hidden(dword_5d4594_1090100, 0);
			}
			nox_window_call_field_94(dword_5d4594_1090100, 16385,
				(uintptr_t)getMemAt(0x5D4594, 1086692), 0);
		} else if (!wndIsShown_nox_xxx_wndIsShown_46ACC0(dword_5d4594_1090100)) {
			nox_window_set_hidden(dword_5d4594_1090100, 1);
		}
		nox_window_call_field_94(nox_rank_window_at(1090104), 16385, (uintptr_t)title, 0);
		sub_46FC50();
		sub_46FD80();
	}
	if ((int32_t)dword_587000_145672 >= 0) {
		sub_46FAE0();
	}
	return 1;
}

//----- (0046F8F0) --------------------------------------------------------
int sub_46F8F0(int page, int empty_rows) {
	for (int row = 0; row < empty_rows; ++row) {
		sub_46DC60(nox_rank_window_at(1090060 + 4 * page), 9, (wchar2_t*)getMemAt(0x587000, 147124));
		sub_46DC60(nox_rank_window_at(1090068 + 4 * page), 9, (wchar2_t*)getMemAt(0x587000, 147128));
		sub_46DC60(nox_rank_window_at(1090076 + 4 * page), 9, (wchar2_t*)getMemAt(0x587000, 147132));
		sub_46DC60(nox_rank_window_at(1090084 + 4 * page), 9, (wchar2_t*)getMemAt(0x587000, 147136));
		sub_46DC60(nox_rank_window_at(1090092 + 4 * page), 9, (wchar2_t*)getMemAt(0x587000, 147140));
	}
	sub_46DC60(nox_rank_window_at(1090060 + 4 * page), 9,
		nox_strman_loadString_40F1D0("player", 0, "C:\\NoxPost\\src\\client\\Gui\\guirank.c", 189));
	sub_46DC60(nox_rank_window_at(1090068 + 4 * page), 9, (wchar2_t*)getMemAt(0x587000, 147188));
	const wchar2_t* class_heading = nox_strman_loadString_40F1D0(
		dword_5d4594_1090120 == 1 ? "LivesHeading" : "class", 0,
		"C:\\NoxPost\\src\\client\\Gui\\guirank.c", dword_5d4594_1090120 == 1 ? 193 : 195);
	sub_46DC60(nox_rank_window_at(1090076 + 4 * page), 9, class_heading);
	const wchar2_t* score_heading;
	if (dword_5d4594_1090120 == 5) {
		score_heading = nox_strman_loadString_40F1D0("rank", 0,
			"C:\\NoxPost\\src\\client\\Gui\\guirank.c", 199);
	} else {
		score_heading = nox_strman_loadString_40F1D0(
			dword_5d4594_1090120 == 1 ? "HealthHeading" : "score", 0,
			"C:\\NoxPost\\src\\client\\Gui\\guirank.c", dword_5d4594_1090120 == 1 ? 201 : 203);
	}
	sub_46DC60(nox_rank_window_at(1090084 + 4 * page), 9, score_heading);
	const wchar2_t* last_heading = nox_strman_loadString_40F1D0(
		dword_5d4594_1090120 == 1 ? "class" : "ping", 0,
		"C:\\NoxPost\\src\\client\\Gui\\guirank.c", dword_5d4594_1090120 == 1 ? 207 : 209);
	return sub_46DC60(nox_rank_window_at(1090092 + 4 * page), 9, last_heading);
}

//----- (0046FB50) --------------------------------------------------------
wchar2_t* sub_46FB50(int a1, uint8_t* a2) {
	wchar2_t* v2;     // eax
	wchar2_t* result; // eax
	wchar2_t* v4;     // eax
	wchar2_t* v5;     // eax
	wchar2_t* v6;     // eax

	switch (a1) {
	case 4:
		v2 = nox_strman_loadString_40F1D0("Ball", 0, "C:\\NoxPost\\src\\client\\Gui\\guirank.c", 244);
		nox_swprintf((wchar2_t*)getMemAt(0x5D4594, 1090024), L"<%s", v2);
		*a2 = 4;
		result = (wchar2_t*)getMemAt(0x5D4594, 1090024);
		break;
	case 1:
		v4 = nox_strman_loadString_40F1D0("King", 0, "C:\\NoxPost\\src\\client\\Gui\\guirank.c", 250);
		nox_swprintf((wchar2_t*)getMemAt(0x5D4594, 1090024), L"<%s", v4);
		result = (wchar2_t*)getMemAt(0x5D4594, 1090024);
		*a2 = 4;
		break;
	case 2:
		v5 = nox_strman_loadString_40F1D0("Flag", 0, "C:\\NoxPost\\src\\client\\Gui\\guirank.c", 256);
		nox_swprintf((wchar2_t*)getMemAt(0x5D4594, 1090024), L"<%s", v5);
		result = (wchar2_t*)getMemAt(0x5D4594, 1090024);
		*a2 = 7;
		break;
	case 3:
		v6 = nox_strman_loadString_40F1D0("Flag", 0, "C:\\NoxPost\\src\\client\\Gui\\guirank.c", 262);
		nox_swprintf((wchar2_t*)getMemAt(0x5D4594, 1090024), L"<%s", v6);
		*a2 = 13;
		result = (wchar2_t*)getMemAt(0x5D4594, 1090024);
		break;
	default:
		result = (wchar2_t*)getMemAt(0x587000, 147724);
		*a2 = 4;
		break;
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
