#include "client__gui__servopts__access.h"
#include "client__gui__window.h"
#include "common__strman.h"

#include "GAME1.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME3_2.h"

nox_gui_server_access_state nox_gui_server_access;

//----- (00454640) --------------------------------------------------------
int sub_454640(void) {
	nox_window* list = nox_gui_server_access.wnd_10123;
	if (!list) {
		return 0;
	}
	int line_height = nox_xxx_guiFontHeightMB_43F320(list->draw_data.font) + 1;
	list->height = 4 * line_height + 2;
	list->end_y = list->off_y + list->height;
	int width = 0;
	int candidate = 0;
	const wchar2_t* text =
		nox_strman_loadString_40F1D0("WARRIOR", 0, "C:\\NoxPost\\src\\client\\Gui\\ServOpts\\access.c", 88);
	nox_xxx_drawGetStringSize_43F840(list->draw_data.font, text, &width, 0, 0);
	text = nox_strman_loadString_40F1D0("WIZARD", 0, "C:\\NoxPost\\src\\client\\Gui\\ServOpts\\access.c", 89);
	nox_xxx_drawGetStringSize_43F840(list->draw_data.font, text, &candidate, 0, 0);
	if (candidate > width) {
		width = candidate;
	}
	text = nox_strman_loadString_40F1D0("CONJURER", 0, "C:\\NoxPost\\src\\client\\Gui\\ServOpts\\access.c", 94);
	nox_xxx_drawGetStringSize_43F840(list->draw_data.font, text, &candidate, 0, 0);
	if (candidate > width) {
		width = candidate;
	}
	list->width = width + 7;
	list->end_x = list->off_x + list->width;
	return (int)list->end_x;
}

//----- (00454740) --------------------------------------------------------
int* sub_454740(void) {
	wchar2_t WideCharStr[18]; // [esp+Ch] [ebp-24h]

	char* settings = sub_416640();
	nox_window* password = nox_xxx_wndGetChildByID_46B0C0(nox_gui_server_access.root, 10136);
	nox_window_call_field_94(password, 16414, (uintptr_t)nox_xxx_sysopGetPass_40A630(), 0);
	if (*(short*)(settings + 105) != -1) {
		nox_xxx_wnd_46ABB0(nox_gui_server_access.wnd_10130, 1);
		nox_gui_server_access.wnd_10129->draw_data.field_0 |= 4u;
		nox_itow(*(unsigned short*)(settings + 105), WideCharStr, 10);
		nox_window_call_field_94(nox_gui_server_access.wnd_10130, 16414, (uintptr_t)WideCharStr, 0);
	}
	if (*(short*)(settings + 107) != -1) {
		nox_xxx_wnd_46ABB0(nox_gui_server_access.wnd_10132, 1);
		nox_gui_server_access.wnd_10131->draw_data.field_0 |= 4u;
		nox_itow(*(unsigned short*)(settings + 107), WideCharStr, 10);
		nox_window_call_field_94(nox_gui_server_access.wnd_10132, 16414, (uintptr_t)WideCharStr, 0);
	}
	nox_window* latency = nox_xxx_wndGetChildByID_46B0C0(nox_gui_server_access.root, 10124);
	if ((int)settings[102] < 0) {
		latency->draw_data.field_0 |= 4u;
	}
	if (settings[100] & 0x20) {
		nox_xxx_wnd_46ABB0(nox_gui_server_access.wnd_10104, 1);
		nox_gui_server_access.wnd_10103->draw_data.field_0 |= 4u;
	}
	nox_window_call_field_94(nox_gui_server_access.wnd_10104, 16414, (uintptr_t)(settings + 78), 0);
	if (sub_4D6F30()) {
		nox_xxx_wnd_46ABB0(nox_gui_server_access.wnd_10102, 0);
	} else {
		nox_xxx_wnd_46ABB0(nox_gui_server_access.wnd_10102, 1);
		if (settings[100] & 0x10) {
			nox_gui_server_access.wnd_10102->draw_data.field_0 = 4;
		} else {
			nox_xxx_wnd_46ABB0(nox_xxx_wndGetChildByID_46B0C0(nox_gui_server_access.root, 10206), 0);
		}
	}
	nox_xxx_wndGetChildByID_46B0C0(nox_gui_server_access.root, 10207)->draw_data.field_0 |= 4u;
	nox_gui_server_access.active_list = nox_gui_server_access.wnd_10105;
	const wchar2_t* text =
		nox_strman_loadString_40F1D0("WARRIOR", 0, "C:\\NoxPost\\src\\client\\Gui\\ServOpts\\access.c", 242);
	nox_window_call_field_94(nox_gui_server_access.wnd_10123, 16397, (uintptr_t)text, -1);
	text = nox_strman_loadString_40F1D0("WIZARD", 0, "C:\\NoxPost\\src\\client\\Gui\\ServOpts\\access.c", 243);
	nox_window_call_field_94(nox_gui_server_access.wnd_10123, 16397, (uintptr_t)text, -1);
	text = nox_strman_loadString_40F1D0("CONJURER", 0, "C:\\NoxPost\\src\\client\\Gui\\ServOpts\\access.c", 244);
	nox_window_call_field_94(nox_gui_server_access.wnd_10123, 16397, (uintptr_t)text, -1);
	if (settings[100] & 0x10) {
		nox_window_set_hidden(nox_gui_server_access.wnd_10109, 0);
		nox_window_set_hidden(nox_gui_server_access.wnd_10105, 1);
	}
	uint8_t class_mask = settings[100];
	if (class_mask) {
		int count = 0;
		nox_scrollListBox_data* list = nox_gui_server_access.wnd_10123->widget_data;
		uint32_t* selection = list ? list->field_12 : NULL;
		if (selection && (class_mask & 1)) {
			selection[0] = 0;
			count = 1;
			selection[1] = UINT32_MAX;
		}
		if (selection && (class_mask & 2)) {
			selection[count++] = 1;
			selection[count] = UINT32_MAX;
		}
		if (selection && (class_mask & 4)) {
			selection[count] = 2;
			selection[count + 1] = UINT32_MAX;
		}
	}
	nox_itow((unsigned char)settings[104], WideCharStr, 10);
	nox_window_call_field_94(nox_gui_server_access.wnd_10133, 16414, (uintptr_t)WideCharStr, 0);
	for (nox_playerInfo* player = nox_common_playerInfoGetFirst_416EA0(); player;
		 player = nox_common_playerInfoGetNext_416EE0(player)) {
		if (player->playerInd != 31 || !nox_common_getEngineFlag(NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING)) {
			sub_455920(player->name_final);
		}
	}
	return sub_455800();
}
