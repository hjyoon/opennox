#include "client__gui__servopts__general.h"
#include "client__gui__window.h"
#include "common__strman.h"

#include "GAME1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2_1.h"
#include "GAME2_3.h"
#include "GAME3.h"
#include "GAME3_2.h"
extern uint32_t nox_server_sendMotd_108752;
extern uint32_t nox_server_connectionType_3596;
extern nox_window* dword_5d4594_1309812;
extern uint32_t dword_5d4594_2650652;

//----- (004AD320) --------------------------------------------------------
nox_window* nox_xxx_gui_4AD320(nox_window* parent) {
	int lang = nox_strman_get_lang_code();
	if (nox_xxx_guiFontHeightMB_43F320(0) > 10) {
		lang = 2;
	}
	if (dword_5d4594_1309812) {
		return NULL;
	}
	int resource_table = dword_5d4594_2650652 == 1 ? 173596 : 173556;
	char* resource = (char*)getMemPtr(0x587000, resource_table + 4 * lang);
	if (!resource) {
		return NULL;
	}
	dword_5d4594_1309812 =
		nox_new_window_from_file(resource, nox_xxx_windowServerOptionsGeneralProc_4AD5D0);
	if (!dword_5d4594_1309812) {
		return NULL;
	}
	sub_46B120(dword_5d4594_1309812, parent);
	nox_xxx_wndSetDrawFn_46B340(dword_5d4594_1309812, sub_4AD570);
	nox_window* spells = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1309812, 10306);
	if (spells && nox_common_gameFlags_check_40A5C0(1056)) {
		nox_xxx_wnd_46ABB0(spells, 0);
	}
	nox_xxx_wndRetNULL_46A8A0();
	sub_4AD840();
	if (dword_5d4594_2650652 == 1) {
		sub_4AD4B0();
		nox_window* owner = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1309812, 10310);
		if (owner) {
			nox_xxx_wnd_46B280(owner, dword_5d4594_1309812);
			nox_xxx_wndSetProc_46B2C0(owner, nox_xxx_windowServerOptionsGeneralProc_4AD5D0);
		}
		nox_window* list = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1309812, 10317);
		for (int off = 173540; list && off < 173556; off += 4) {
			const char* key = (const char*)getMemPtr(0x587000, off);
			if (!key) {
				continue;
			}
			wchar2_t* text = nox_strman_loadString_40F1D0(
				key, 0, "C:\\NoxPost\\src\\client\\Gui\\ServOpts\\general.c", 308);
			if (!text) {
				continue;
			}
			nox_window_call_field_94(list, 16397, (uintptr_t)text, UINTPTR_MAX);
		}
	} else {
		nox_window* advanced = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1309812, 10319);
		if (advanced) {
			nox_window_set_hidden(advanced, 1);
		}
	}
	if (nox_common_getEngineFlag(NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING)) {
		nox_window* video = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1309812, 10304);
		if (video) {
			nox_xxx_wnd_46ABB0(video, 0);
		}
	}
	return dword_5d4594_1309812;
}

//----- (004AD4B0) --------------------------------------------------------
int sub_4AD4B0() {
	if (!dword_5d4594_1309812) {
		return 0;
	}
	nox_window* list = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1309812, 10317);
	nox_window* button = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1309812, 10316);
	if (!list || !button) {
		return 0;
	}
	int line_height = nox_xxx_guiFontHeightMB_43F320(list->draw_data.font) + 1;
	list->height = 4 * line_height + 2;
	list->end_y = list->off_y + list->height;
	int max_width = 0;
	for (int off = 173540; off < 173556; off += 4) {
		const char* key = (const char*)getMemPtr(0x587000, off);
		if (!key) {
			continue;
		}
		wchar2_t* text = nox_strman_loadString_40F1D0(
			key, 0, "C:\\NoxPost\\src\\client\\Gui\\ServOpts\\general.c", 92);
		if (!text) {
			continue;
		}
		int width = 0;
		nox_xxx_drawGetStringSize_43F840(list->draw_data.font, text, &width, 0, 0);
		if (width > max_width) {
			max_width = width;
		}
	}
	int width = max_width + 7;
	if (width < button->width) {
		width = (int)button->width;
	}
	list->width = width;
	list->off_x = list->end_x - width;
	return (int)button->width;
}

//----- (004AD840) --------------------------------------------------------
int sub_4AD840() {
	if (!dword_5d4594_1309812) {
		return 0;
	}
	if (nox_server_doPlayersAutoRespawn_40A5F0()) {
		nox_window* control = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1309812, 10301);
		if (control) {
			control->draw_data.field_0 |= 4u;
			if (nox_common_gameFlags_check_40A5C0(1024)) {
				nox_xxx_wnd_46ABB0(control, 0);
			}
		}
	}
	if (nox_server_sendMotd_108752) {
		nox_window* control = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1309812, 10302);
		if (control) {
			control->draw_data.field_0 |= 4u;
		}
	}
	if (sub_4D0D70()) {
		nox_window* control = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1309812, 10304);
		if (control) {
			control->draw_data.field_0 |= 4u;
		}
	}
	if (sub_409F40(2)) {
		nox_window* control = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1309812, 10305);
		if (control) {
			control->draw_data.field_0 |= 4u;
		}
	}
	if (sub_409F40(0x2000)) {
		nox_window* control = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1309812, 10306);
		if (control) {
			control->draw_data.field_0 |= 4u;
		}
	}
	if (dword_5d4594_2650652 == 1 && nox_server_connectionType_3596 >= 1 &&
		nox_server_connectionType_3596 <= 4) {
		nox_window* type = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1309812, 10316);
		nox_window* rate = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1309812, 10312);
		if (!type || !rate) {
			return 0;
		}
		const char* key = (const char*)getMemPtr(0x587000, 173536 + 4 * nox_server_connectionType_3596);
		if (!key) {
			return 0;
		}
		wchar2_t* text = nox_strman_loadString_40F1D0(
			key, 0, "C:\\NoxPost\\src\\client\\Gui\\ServOpts\\general.c", 391);
		if (!text) {
			return 0;
		}
		nox_window_call_field_94(type, 16385, (uintptr_t)text, UINTPTR_MAX);
		return (int)nox_window_call_field_94(rate, 16394, 4 - nox_xxx_rateGet_40A6C0(), 0);
	}
	return 0;
}
