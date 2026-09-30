#include "client__gui__conntype.h"

#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_3.h"
#include "client__gui__servopts__guiserv.h"
#include "client__gui__window.h"
#include "common__strman.h"

extern nox_window* dword_5d4594_1305684;
extern int nox_win_width;
extern int nox_win_height;

enum { nox_connection_type_count = 4 };

static char* nox_connection_type_key(int ind) {
	return (char*)getMemPtr(0x587000, 164928 + 4 * ind);
}

//----- (0049C820) --------------------------------------------------------
int sub_49C820() {
	dword_5d4594_1305684 = nox_new_window_from_file("conntype.wnd", sub_49CA60);
	if (!dword_5d4594_1305684) {
		return 0;
	}
	sub_46B120(dword_5d4594_1305684, 0);
	nox_xxx_wndShowModalMB_46A8C0(dword_5d4594_1305684);
	sub_46C690(dword_5d4594_1305684);
	nox_xxx_windowFocus_46B500(dword_5d4594_1305684);
	sub_49C910();
	nox_window_setPos_46A9B0(dword_5d4594_1305684,
							 nox_win_width / 2 - (int)dword_5d4594_1305684->width / 2,
							 nox_win_height / 2 - (int)dword_5d4594_1305684->height / 2);
	nox_xxx_guiServerOptsLoad_457500();
	sub_459D80(1);
	nox_window* list = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1305684, 10352);
	if (!list) {
		return 0;
	}
	for (int i = 0; i < nox_connection_type_count; i++) {
		wchar2_t* text =
			nox_strman_loadString_40F1D0(nox_connection_type_key(i), 0, "C:\\NoxPost\\src\\client\\Gui\\conntype.c", 158);
		nox_window_call_field_94(list, 16397, (uintptr_t)text, UINTPTR_MAX);
	}
	return (int)nox_window_call_field_94(list, 16403, 0, 0);
}

//----- (0049C910) --------------------------------------------------------
nox_window* sub_49C910() {
	if (!dword_5d4594_1305684) {
		return 0;
	}
	nox_window* list = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1305684, 10352);
	if (!list) {
		return 0;
	}
	int height = 5 * (nox_xxx_guiFontHeightMB_43F320(list->draw_data.font) + 1);
	list->end_y = list->off_y + height + 2;
	list->height = height + 2;

	int max_width = 0;
	for (int i = 0; i < nox_connection_type_count; i++) {
		wchar2_t* text =
			nox_strman_loadString_40F1D0(nox_connection_type_key(i), 0, "C:\\NoxPost\\src\\client\\Gui\\conntype.c", 53);
		int width = 0;
		nox_xxx_drawGetStringSize_43F840(list->draw_data.font, text, &width, 0, 0);
		if (width > max_width) {
			max_width = width;
		}
	}
	int width = max_width + 7;
	int title_width = 0;
	nox_xxx_drawGetStringSize_43F840(list->draw_data.font, list->draw_data.text, &title_width, 0, 0);
	if (width <= title_width) {
		width = title_width;
	}

	list->width = width;
	list->end_x = list->off_x + width;
	dword_5d4594_1305684->off_x = list->off_x - 40;
	dword_5d4594_1305684->off_y = list->off_y - 40;
	dword_5d4594_1305684->end_x = list->end_x + 40;
	dword_5d4594_1305684->end_y = list->end_y + 40;
	dword_5d4594_1305684->width = list->width + 80;
	dword_5d4594_1305684->height = list->height + 80;

	nox_window* button = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1305684, 10353);
	if (!button) {
		return 0;
	}
	button->off_y = list->end_y + 10;
	button->off_x = list->end_x - button->width;
	button->end_y = button->off_y + button->height;
	button->end_x = button->off_x + button->width;
	return button;
}
