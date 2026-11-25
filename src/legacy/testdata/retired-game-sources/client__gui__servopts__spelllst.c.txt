#include "client__gui__servopts__spelllst.h"
#include "client__gui__window.h"
#include "common__strman.h"

#include "GAME1.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_3.h"
#include "GAME3.h"
#include "GAME5_2.h"
#include "common__magic__speltree.h"
extern nox_window* dword_5d4594_1045508;
extern nox_window* dword_5d4594_1045480;
extern nox_window* dword_5d4594_1045484;
extern uint32_t dword_5d4594_2650652;

static int nox_gui_server_spell_first_visible(const nox_scrollListBox_data* data) {
	if (!data || !data->items || !data->field_11_0) {
		return 0;
	}
	int scroll = data->field_13_1;
	if ((int)data->items[0].field_0 > scroll) {
		return 0;
	}
	for (int i = 1; i < data->field_11_0; ++i) {
		if ((int)data->items[i].field_0 > scroll) {
			return i;
		}
	}
	return 0;
}

//----- (00453850) --------------------------------------------------------
nox_window* nox_xxx_guiSpelllistLoad_453850(nox_window* parent) {
	dword_5d4594_1045484 = nox_new_window_from_file("spelllst.wnd", sub_453C00);
	if (!dword_5d4594_1045484) {
		return NULL;
	}
	nox_xxx_wndSetDrawFn_46B340(dword_5d4594_1045484, sub_453B80);
	sub_46B120(dword_5d4594_1045484, parent);
	nox_xxx_wnd_46B280(dword_5d4594_1045484, parent);
	dword_5d4594_1045480 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1045484, 1110);
	dword_5d4594_1045508 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1045484, 1112);
	if (!dword_5d4594_1045480 || !dword_5d4594_1045508) {
		nox_xxx_windowDestroyMB_46C4E0(dword_5d4594_1045484);
		dword_5d4594_1045484 = NULL;
		dword_5d4594_1045480 = NULL;
		dword_5d4594_1045508 = NULL;
		return NULL;
	}
	sub_453B00();
	nox_window_call_field_94(dword_5d4594_1045480, 16399, 0, 0);
	nox_window_call_field_94(dword_5d4594_1045508, 16399, 0, 0);
	wchar2_t wbuf[64] = {0};
	for (int i = 1; i < NOX_SPELLS_MAX; ++i) {
		if (!nox_xxx_spellIsValid_424B50(i)) {
			continue;
		}
		int flags = nox_xxx_spellFlags_424A70(i);
		if ((flags & 0x15000) != 0) {
			continue;
		}
		if ((flags & 0x1000000) || ((flags & 0x2000000) && (flags & 0x4000000))) {
			wchar2_t* common = nox_strman_loadString_40F1D0(
				"Common", 0, "C:\\NoxPost\\src\\client\\Gui\\ServOpts\\spelllst.c", 307);
			nox_wcscpy(wbuf, common);
		} else {
			if ((flags & 0x6000000) == 0) {
				continue;
			}
			wbuf[0] = 0;
			if (flags & 0x2000000) {
				wchar2_t* wizard = nox_strman_loadString_40F1D0(
					"SpellWizard", 0, "C:\\NoxPost\\src\\client\\Gui\\ServOpts\\spelllst.c", 314);
				nox_wcscat(wbuf, wizard);
			}
			if (flags & 0x4000000) {
				wchar2_t* conjurer = nox_strman_loadString_40F1D0(
					"SpellConjurer", 0, "C:\\NoxPost\\src\\client\\Gui\\ServOpts\\spelllst.c", 318);
				nox_wcscat(wbuf, conjurer);
			}
		}
		nox_window_call_field_94(dword_5d4594_1045508, 16397, (uintptr_t)wbuf, UINTPTR_MAX);
		wchar2_t* spell_title = nox_xxx_spellTitle_424930(i);
		nox_wcsncpy(wbuf, spell_title, sizeof(wbuf) / sizeof(wbuf[0]) - 1);
		wbuf[sizeof(wbuf) / sizeof(wbuf[0]) - 1] = 0;
		nox_window_call_field_94(dword_5d4594_1045480, 16397, (uintptr_t)wbuf, UINTPTR_MAX);
	}
	nox_window* up = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1045484, 1113);
	nox_window_call_field_94(dword_5d4594_1045480, 16408, (uintptr_t)up, 0);
	nox_window_call_field_94(dword_5d4594_1045508, 16408, (uintptr_t)up, 0);
	nox_window* down = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1045484, 1114);
	nox_window_call_field_94(dword_5d4594_1045480, 16409, (uintptr_t)down, 0);
	nox_window_call_field_94(dword_5d4594_1045508, 16409, (uintptr_t)down, 0);
	sub_454040(getMemAt(0x5D4594, 1045488));
	sub_454120();
	if (!nox_common_gameFlags_check_40A5C0(1) || nox_common_gameFlags_check_40A5C0(49152)) {
		sub_46AD20(dword_5d4594_1045484, 1115, 1133, 0);
	}
	return dword_5d4594_1045484;
}

//----- (00453C00) --------------------------------------------------------
int sub_453C00(nox_window* win, int event, nox_window* control, uintptr_t event_arg) {
	(void)win;
	(void)event_arg;
	if (!dword_5d4594_1045484 || !dword_5d4594_1045480 || !dword_5d4594_1045508 || !control) {
		return 0;
	}
	if (event == 0x4000) {
		if (control == nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1045484, 1113) ||
			control == nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1045484, 1114)) {
			nox_window_call_field_94(dword_5d4594_1045480, 0x4000, (uintptr_t)control, 0);
			nox_window_call_field_94(dword_5d4594_1045508, 0x4000, (uintptr_t)control, 0);
			sub_454120();
		}
		return 0;
	}
	if (event != 16391) {
		return 0;
	}

	int id = nox_xxx_wndGetID_46B0A0(control);
	switch (id) {
	case 1113:
	case 1114:
		nox_window_call_field_94(dword_5d4594_1045480, 0x4000, (uintptr_t)control, 0);
		nox_window_call_field_94(dword_5d4594_1045508, 0x4000, (uintptr_t)control, 0);
		sub_454120();
		return 0;
	case 1115:
	case 1116: {
		nox_scrollListBox_data* data = dword_5d4594_1045480->widget_data;
		char* settings = sub_4165B0();
		if (data && data->items) {
			for (int i = 0; i < data->field_11_0; ++i) {
				if (!data->items[i].text[0]) {
					continue;
				}
				int spell = nox_xxx_spellByTitle_424960(data->items[i].text);
				if (!spell) {
					continue;
				}
				if (id == 1115) {
					if ((!nox_common_gameFlags_check_40A5C0(64) && (!settings || !(settings[52] & 0x40))) ||
						spell != 132) {
						sub_453FA0(getMemAt(0x5D4594, 1045488), spell, 1);
					}
				} else {
					sub_453FA0(getMemAt(0x5D4594, 1045488), spell, 0);
				}
			}
		}
		if (dword_5d4594_2650652) {
			int online[15];
			sub_57A1E0(online, 0, 0, 4, 6128);
			for (int i = 0; i < 5; ++i) {
				*getMemU32Ptr(0x5D4594, 1045488 + i * 4) &= online[i + 6];
			}
		}
		sub_454120();
		sub_459D50(1);
		nox_xxx_clientPlaySoundSpecial_452D80(766, 100);
		return 0;
	}
	default:
		break;
	}

	if (id >= 1120 && id <= 1133) {
		nox_scrollListBox_data* data = dword_5d4594_1045480->widget_data;
		int index = nox_gui_server_spell_first_visible(data) + id - 1120;
		if (!data || !data->items || index < 0 || index >= data->field_11_0 || !data->items[index].text[0]) {
			sub_459D50(1);
			nox_xxx_clientPlaySoundSpecial_452D80(766, 100);
			return 0;
		}
		int spell = nox_xxx_spellByTitle_424960(data->items[index].text);
		if (!spell) {
			sub_459D50(1);
			nox_xxx_clientPlaySoundSpecial_452D80(766, 100);
			return 0;
		}
		int online[15];
		if (dword_5d4594_2650652) {
			sub_57A1E0(online, 0, 0, 4, 6128);
			if (!sub_454000(&online[6], spell)) {
				control->draw_data.field_0 ^= 4u;
				wchar2_t* message = nox_strman_loadString_40F1D0(
					"NotInternet", 0, "C:\\NoxPost\\src\\client\\Gui\\ServOpts\\spelllst.c", 211);
				wchar2_t* caption = nox_strman_loadString_40F1D0(
					"Notice", 0, "C:\\NoxPost\\src\\client\\Gui\\ServOpts\\spelllst.c", 210);
				nox_xxx_dialogMsgBoxCreate_449A10(dword_5d4594_1045484, caption, message, 33, NULL, NULL);
				sub_44A360(1);
				nox_xxx_clientPlaySoundSpecial_452D80(766, 100);
				return 0;
			}
		}
		char* settings = sub_4165B0();
		if ((nox_common_gameFlags_check_40A5C0(64) || (settings && (settings[52] & 0x40))) && spell == 132) {
			control->draw_data.field_0 ^= 4u;
			wchar2_t* message = nox_strman_loadString_40F1D0(
				"plyrspel.c:Illegal", 0, "C:\\NoxPost\\src\\client\\Gui\\ServOpts\\spelllst.c", 226);
			wchar2_t* caption = nox_strman_loadString_40F1D0(
				"Notice", 0, "C:\\NoxPost\\src\\client\\Gui\\ServOpts\\spelllst.c", 225);
			nox_xxx_dialogMsgBoxCreate_449A10(dword_5d4594_1045484, caption, message, 33, NULL, NULL);
			sub_44A360(1);
		} else {
			sub_453FA0(getMemAt(0x5D4594, 1045488), spell, (control->draw_data.field_0 & 4) ? 0 : 1);
			sub_459D50(1);
		}
		nox_xxx_clientPlaySoundSpecial_452D80(766, 100);
		return 0;
	}

	nox_xxx_clientPlaySoundSpecial_452D80(766, 100);
	return 0;
}
