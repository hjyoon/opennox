#include "client__shell__inputcfg__inputcfg.h"
#include "client__system__ctrlevnt.h"

#include "GAME3.h"
#include "client__gui__window.h"
#include "common__strman.h"
extern nox_window* dword_5d4594_1522620;
extern nox_window* dword_5d4594_1522624;
extern nox_window* dword_5d4594_1522628;
extern nox_window* dword_5d4594_1522632;
extern nox_window* dword_5d4594_1522612;
extern nox_window* dword_5d4594_1522604;

//----- (004CBD30) --------------------------------------------------------
char* sub_4CBD30() {
	char v12[256]; // [esp+8h] [ebp-100h]

	sub_42CD90();
	nox_scrollListBox_data* data = dword_5d4594_1522620 ? dword_5d4594_1522620->widget_data : NULL;
	int count = data ? data->field_11_0 : 0;
	for (int i = 0; i < count; ++i) {
		wchar2_t* event_title = (wchar2_t*)nox_window_call_field_94(dword_5d4594_1522620, 16406, i, 0);
		char* event_name = nox_xxx_bindevent_bindNameByTitle_42EA40(event_title);
		wchar2_t* primary_title = (wchar2_t*)nox_window_call_field_94(dword_5d4594_1522624, 16406, i, 0);
		char* primary_name = nox_xxx_keybind_nameByTitle_42E960(primary_title);
		wchar2_t* secondary_title = (wchar2_t*)nox_window_call_field_94(dword_5d4594_1522628, 16406, i, 0);
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
		"bindevent:ToggleQuitMenu", 0, "C:\\NoxPost\\src\\Client\\shell\\InputCfg\\inputcfg.c", 192);
	char* event_name = nox_xxx_bindevent_bindNameByTitle_42EA40(event_title);
	wchar2_t* key_title = nox_strman_loadString_40F1D0(
		"keybind:Esc", 0, "C:\\NoxPost\\src\\Client\\shell\\InputCfg\\inputcfg.c", 193);
	char* result = nox_xxx_keybind_nameByTitle_42E960(key_title);
	if (result) {
		nox_sprintf(v12, "%s = %s", result, event_name);
		result = (char*)nox_client_parseConfigHotkeysLine_42CF50(v12);
	}
	return result;
}

//----- (004CBF60) --------------------------------------------------------
int sub_4CBF60(nox_window* win, int event, uintptr_t event_arg, uintptr_t event_arg2) {
	if (event > 0x4007) {
		if (event == 16393) {
			(void)nox_xxx_wndListboxProcPre_4A30D0(win, 0x4009u, event_arg, event_arg2);
			int first = nox_xxx_wndListBoxFirstVisible(win);
			nox_window_call_field_94(dword_5d4594_1522620, 16412, first, 0);
			nox_window_call_field_94(dword_5d4594_1522624, 16412, first, 0);
			nox_window_call_field_94(dword_5d4594_1522628, 16412, first, 0);
		} else if (event == 16400) {
			nox_window* event_win = (nox_window*)event_arg;
			int selection = (int)(intptr_t)nox_window_call_field_94(event_win, 16404, 0, 0);
			if (selection >= 0) {
				dword_5d4594_1522632 = event_win;
				const wchar2_t* item = (const wchar2_t*)nox_window_call_field_94(
					dword_5d4594_1522620, 16406, selection, 0);
				wchar2_t* prompt = nox_strman_loadString_40F1D0(
					"InputCfg.wnd:PressKey", 0,
					"C:\\NoxPost\\src\\Client\\shell\\InputCfg\\inputcfg.c", 424);
				nox_swprintf((wchar2_t*)getMemAt(0x5D4594, 1522636), L"%s\n'%s'", prompt, item);
				nox_xxx_wndShowModalMB_46A8C0(dword_5d4594_1522612);
				nox_xxx_windowFocus_46B500(dword_5d4594_1522612);
				sub_46C690(dword_5d4594_1522612);
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
		if (event_win == nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1522604, 921) ||
			event_win == nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1522604, 922)) {
			nox_window_call_field_94(dword_5d4594_1522620, event, event_arg, 0);
			nox_window_call_field_94(dword_5d4594_1522624, event, event_arg, 0);
			nox_window_call_field_94(dword_5d4594_1522628, event, event_arg, 0);
			return (int)nox_xxx_wndListboxProcPre_4A30D0(win, event, event_arg, event_arg2);
		}
	}
	return (int)nox_xxx_wndListboxProcPre_4A30D0(win, event, event_arg, event_arg2);
}
