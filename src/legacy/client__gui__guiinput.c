#include "client__gui__guiinput.h"
#include "client__gui__window.h"
#include "client__system__ctrlevnt.h"
#include "common__strman.h"

#include "GAME1_2.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "GAME3.h"
#include "GAME3_1.h"

nox_window* dword_5d4594_1321236 = 0;
nox_window* dword_5d4594_1321240 = 0;
nox_window* dword_5d4594_1321248 = 0;
nox_window* dword_5d4594_1321244 = 0;

extern nox_window* dword_5d4594_1321252;
extern nox_window* dword_5d4594_1321232;
extern nox_window* dword_5d4594_1321228;
extern int nox_win_width;

//----- (004C3620) --------------------------------------------------------
char* sub_4C3620() {
	char v12[256]; // [esp+8h] [ebp-100h]

	sub_42CD90();
	nox_scrollListBox_data* data = dword_5d4594_1321240 ? dword_5d4594_1321240->widget_data : NULL;
	int count = data ? data->field_11_0 : 0;
	for (int i = 0; i < count; ++i) {
		wchar2_t* event_title = (wchar2_t*)nox_window_call_field_94(dword_5d4594_1321240, 16406, i, 0);
		char* event_name = nox_xxx_bindevent_bindNameByTitle_42EA40(event_title);
		wchar2_t* primary_title = (wchar2_t*)nox_window_call_field_94(dword_5d4594_1321244, 16406, i, 0);
		char* primary_name = nox_xxx_keybind_nameByTitle_42E960(primary_title);
		wchar2_t* secondary_title = (wchar2_t*)nox_window_call_field_94(dword_5d4594_1321248, 16406, i, 0);
		char* secondary_name = nox_xxx_keybind_nameByTitle_42E960(secondary_title);
		if (secondary_name) {
			nox_sprintf(v12, "%s = %s", secondary_name, event_name);
			nox_client_parseConfigHotkeysLine_42CF50(v12);
		}
		if (primary_name) {
			nox_sprintf(v12, "%s = %s", primary_name, event_name);
			nox_client_parseConfigHotkeysLine_42CF50(v12);
		}
	}
	wchar2_t* event_title = nox_strman_loadString_40F1D0(
		"bindevent:ToggleQuitMenu", 0, "C:\\NoxPost\\src\\client\\Gui\\GuiInput.c", 191);
	char* event_name = nox_xxx_bindevent_bindNameByTitle_42EA40(event_title);
	wchar2_t* key_title = nox_strman_loadString_40F1D0(
		"keybind:Esc", 0, "C:\\NoxPost\\src\\client\\Gui\\GuiInput.c", 192);
	char* result = nox_xxx_keybind_nameByTitle_42E960(key_title);
	if (result) {
		nox_sprintf(v12, "%s = %s", result, event_name);
		result = (char*)nox_client_parseConfigHotkeysLine_42CF50(v12);
	}
	return result;
}

//----- (004C3760) --------------------------------------------------------
int sub_4C3760() {
	dword_5d4594_1321228 = nox_new_window_from_file("InputCfg.wnd", sub_4C3A90);
	if (!dword_5d4594_1321228) {
		return 0;
	}
	dword_5d4594_1321236 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1321228, 910);
	dword_5d4594_1321240 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1321228, 911);
	dword_5d4594_1321244 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1321228, 912);
	dword_5d4594_1321248 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1321228, 913);
	if (!dword_5d4594_1321236 || !dword_5d4594_1321240 || !dword_5d4594_1321244 ||
		!dword_5d4594_1321248) {
		return 0;
	}

	nox_scrollListBox_data* parent_data = dword_5d4594_1321236->widget_data;
	if (!parent_data || !parent_data->field_7 || !parent_data->field_8 || !parent_data->field_9) {
		return 0;
	}
	nox_xxx_wndSetID_46B080(parent_data->field_7, 921);
	nox_xxx_wndSetID_46B080(parent_data->field_8, 922);
	nox_xxx_wndSetID_46B080(parent_data->field_9, 920);
	nox_xxx_wndSetProc_46B2C0(dword_5d4594_1321236, sub_4C3CD0);
	sub_46B120(dword_5d4594_1321240, dword_5d4594_1321236);
	sub_46B120(dword_5d4594_1321244, dword_5d4594_1321236);
	sub_46B120(dword_5d4594_1321248, dword_5d4594_1321236);
	nox_xxx_wndSetWindowProc_46B300(dword_5d4594_1321244, sub_4C3A60);
	nox_xxx_wndSetWindowProc_46B300(dword_5d4594_1321248, sub_4C3A60);

	nox_window* scroll_up = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1321228, 921);
	nox_window* scroll_down = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1321228, 922);
	nox_window* lists[] = {dword_5d4594_1321240, dword_5d4594_1321244, dword_5d4594_1321248};
	for (int i = 0; i < 3; ++i) {
		nox_window_call_field_94(lists[i], 16408, (uintptr_t)scroll_up, 0);
		nox_window_call_field_94(lists[i], 16409, (uintptr_t)scroll_down, 0);
	}

	for (int id = 971, end = sub_47DBC0() + 971; id < end; ++id) {
		nox_xxx_wnd_46ABB0(nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1321228, id), 1);
	}
	nox_window* primary = nox_xxx_wndGetChildByID_46B0C0(
		dword_5d4594_1321228, nox_client_mousePriKey_430AF0() + 971);
	if (primary) {
		nox_window_call_field_94(primary, 16392, 1, 0);
	}
	nox_window_setPos_46A9B0(dword_5d4594_1321228,
							 (nox_win_width - dword_5d4594_1321228->width) / 2, 0);

	dword_5d4594_1321232 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1321228, 980);
	if (!dword_5d4594_1321232) {
		return 0;
	}
	sub_46B120(dword_5d4594_1321232, 0);
	nox_xxx_wndSetProc_46B2C0(dword_5d4594_1321232, sub_4C3A90);
	nox_xxx_wndSetWindowProc_46B300(dword_5d4594_1321232, sub_4C3EB0);
	nox_window_set_hidden(dword_5d4594_1321232, 1);
	nox_window_setPos_46A9B0(dword_5d4594_1321232,
							 (nox_win_width - dword_5d4594_1321232->width) / 2,
							 dword_5d4594_1321232->off_y);
	nox_window* prompt = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1321232, 981);
	if (prompt) {
		nox_window_call_field_94(prompt, 16385, (uintptr_t)getMemAt(0x5D4594, 1321256), 0);
	}
	nox_xxx_wnd_46ABB0(nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1321228, 932), 1);
	nox_window_set_hidden(dword_5d4594_1321228, 1);
	return 1;
}

//----- (004C3CD0) --------------------------------------------------------
int sub_4C3CD0(nox_window* win, int event, uintptr_t event_arg, uintptr_t event_arg2) {
	if (event > 0x4007) {
		if (event == 16393) {
			(void)nox_xxx_wndListboxProcPre_4A30D0(win, 0x4009u, event_arg, event_arg2);
			int first = nox_xxx_wndListBoxFirstVisible(win);
			nox_window_call_field_94(dword_5d4594_1321240, 16412, first, 0);
			nox_window_call_field_94(dword_5d4594_1321244, 16412, first, 0);
			nox_window_call_field_94(dword_5d4594_1321248, 16412, first, 0);
		} else if (event == 16400) {
			nox_window* event_win = (nox_window*)event_arg;
			int selection = (int)(intptr_t)nox_window_call_field_94(event_win, 16404, 0, 0);
			if (selection >= 0) {
				dword_5d4594_1321252 = event_win;
				const wchar2_t* item = (const wchar2_t*)nox_window_call_field_94(
					dword_5d4594_1321240, 16406, selection, 0);
				wchar2_t* prompt = nox_strman_loadString_40F1D0(
					"InputCfg.wnd:PressKey", 0, "C:\\NoxPost\\src\\client\\Gui\\GuiInput.c", 436);
				nox_swprintf((wchar2_t*)getMemAt(0x5D4594, 1321256), L"%s\n'%s'", prompt, item);
				nox_xxx_wndShowModalMB_46A8C0(dword_5d4594_1321232);
				nox_xxx_windowFocus_46B500(dword_5d4594_1321232);
				sub_46C690(dword_5d4594_1321232);
				return (int)nox_xxx_wndListboxProcPre_4A30D0(win, 0x4010u, event_arg, event_arg2);
			}
		}
	} else {
		if (event != 16391) {
			if (event == 23) {
				return 1;
			}
			if (event != 0x4000) {
				return (int)nox_xxx_wndListboxProcPre_4A30D0(win, event, event_arg, event_arg2);
			}
		}
		nox_window* event_win = (nox_window*)event_arg;
		if (event_win == nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1321228, 921) ||
			event_win == nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1321228, 922)) {
			nox_window_call_field_94(dword_5d4594_1321240, event, event_arg, 0);
			nox_window_call_field_94(dword_5d4594_1321244, event, event_arg, 0);
			nox_window_call_field_94(dword_5d4594_1321248, event, event_arg, 0);
			return (int)nox_xxx_wndListboxProcPre_4A30D0(win, event, event_arg, event_arg2);
		}
	}
	return (int)nox_xxx_wndListboxProcPre_4A30D0(win, event, event_arg, event_arg2);
}
