#include "client__gui__guitrade.h"

#include <string.h>

#include "client__gui__window.h"
#include "common__strman.h"

#include "GAME1_2.h"
#include "GAME2.h"
#include "GAME2_3.h"
#include "GAME3_1.h"
#include "client__gui__gamewin__gamewin.h"
#include "client__gui__guimsg.h"
#include "client__drawable__drawable.h"

extern uint32_t dword_5d4594_1320964;
extern nox_window* dword_5d4594_1320940;
extern int nox_win_width;
extern int nox_win_height;

static nox_gui_trade_slot nox_gui_trade_local_slots[NOX_GUI_TRADE_SLOT_COUNT];
static nox_gui_trade_slot nox_gui_trade_remote_slots[NOX_GUI_TRADE_SLOT_COUNT];

nox_gui_trade_slot* nox_gui_trade_slot_at(int remote, int index) {
	if (index < 0 || index >= NOX_GUI_TRADE_SLOT_COUNT) {
		return NULL;
	}
	return remote ? &nox_gui_trade_remote_slots[index] : &nox_gui_trade_local_slots[index];
}

void nox_gui_trade_slots_reset(void) {
	for (int remote = 0; remote < 2; remote++) {
		for (int index = 0; index < NOX_GUI_TRADE_SLOT_COUNT; index++) {
			nox_gui_trade_slot* slot = nox_gui_trade_slot_at(remote, index);
			if (slot->drawable) {
				nox_xxx_spriteDelete_45A4B0(slot->drawable);
			}
			memset(slot, 0, sizeof(*slot));
		}
	}
}

static uint16_t nox_gui_trade_read_u16(const uint8_t* data) {
	return (uint16_t)data[0] | (uint16_t)data[1] << 8;
}

//----- (004C09D0) --------------------------------------------------------
int sub_4C09D0() {
	nox_window* v0;     // eax
	wchar2_t* v2;        // eax
	nox_window* v3;      // esi
	wchar2_t* v4;        // eax
	nox_window* v5;      // esi
	wchar2_t* v6;        // eax
	nox_window* v7;      // eax
	nox_window* v8;      // eax
	nox_window* v9;      // esi
	wchar2_t* v10;       // eax
	nox_window* v11;     // esi
	wchar2_t* v12;       // eax
	nox_window* v13;     // esi
	wchar2_t* v14;       // eax
	wchar2_t* v21;       // eax

	v0 = nox_new_window_from_file("Trade.wnd", sub_4C0C90);
	dword_5d4594_1320940 = v0;
	if (!v0) {
		return 0;
	}
	nox_window_set_all_funcs(v0, sub_4C0630, sub_4C0D00, 0);
	v2 = nox_strman_loadString_40F1D0("TradeMain", 0, "C:\\NoxPost\\src\\client\\Gui\\GUITrade.c", 692);
	nox_xxx_wndWddSetTooltip_46B000(&dword_5d4594_1320940->draw_data, v2);
	v3 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1320940, 3702);
	v4 = nox_strman_loadString_40F1D0("TradePlayerName", 0, "C:\\NoxPost\\src\\client\\Gui\\GUITrade.c", 695);
	nox_xxx_wndWddSetTooltip_46B000(&v3->draw_data, v4);
	v5 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1320940, 3703);
	v6 = nox_strman_loadString_40F1D0("TradeVendorName", 0, "C:\\NoxPost\\src\\client\\Gui\\GUITrade.c", 698);
	nox_xxx_wndWddSetTooltip_46B000(&v5->draw_data, v6);
	v7 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1320940, 3704);
	nox_gui_winSetFunc96_46B070(v7, sub_4C1120);
	v8 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1320940, 3705);
	nox_gui_winSetFunc96_46B070(v8, sub_4C1120);
	v9 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1320940, 3708);
	v10 = nox_strman_loadString_40F1D0("TradePlayerAccept", 0, "C:\\NoxPost\\src\\client\\Gui\\GUITrade.c", 709);
	nox_xxx_wndWddSetTooltip_46B000(&v9->draw_data, v10);
	v11 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1320940, 3709);
	v12 = nox_strman_loadString_40F1D0("TradeVendorAccept", 0, "C:\\NoxPost\\src\\client\\Gui\\GUITrade.c", 712);
	nox_xxx_wndWddSetTooltip_46B000(&v11->draw_data, v12);
	v13 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1320940, 3710);
	v14 = nox_strman_loadString_40F1D0("TradeCancel", 0, "C:\\NoxPost\\src\\client\\Gui\\GUITrade.c", 715);
	nox_xxx_wndWddSetTooltip_46B000(&v13->draw_data, v14);
	nox_window_set_hidden(dword_5d4594_1320940, 1);
	nox_xxx_wnd_46ABB0(dword_5d4594_1320940, 0);
	nox_gui_trade_slots_reset();
	v21 = nox_strman_loadString_40F1D0("TotalValueLabel", 0, "C:\\NoxPost\\src\\client\\Gui\\GUITrade.c", 749);
	nox_wcscpy((wchar2_t*)getMemAt(0x5D4594, 1319972), v21);
	*getMemU32Ptr(0x5D4594, 1320188) = 0;
	*getMemU32Ptr(0x5D4594, 1320192) = 0;
	*getMemU32Ptr(0x5D4594, 1320196) = nox_win_width;
	*getMemU32Ptr(0x5D4594, 1320200) = nox_win_height;
	*getMemU32Ptr(0x5D4594, 1320220) = nox_win_width;
	*getMemU32Ptr(0x5D4594, 1320224) = nox_win_height;
	*getMemU32Ptr(0x5D4594, 1320204) = 0;
	*getMemU32Ptr(0x5D4594, 1320208) = 0;
	*getMemU32Ptr(0x5D4594, 1320164) = nox_xxx_gLoadImg_42F970("TradeBase");
	*getMemU32Ptr(0x5D4594, 1320168) = nox_xxx_gLoadImg_42F970("TradeLeftAcceptPushed");
	*getMemU32Ptr(0x5D4594, 1320172) = nox_xxx_gLoadImg_42F970("TradeLeftAcceptLit");
	*getMemU32Ptr(0x5D4594, 1320176) = nox_xxx_gLoadImg_42F970("TradeRightAcceptLit");
	*getMemU32Ptr(0x5D4594, 1320180) = nox_xxx_gLoadImg_42F970("TradeCancelLit");
	*getMemU32Ptr(0x5D4594, 1320184) = nox_xxx_gLoadImg_42F970("TradeGold");
	return 1;
}

//----- (004C15D0) --------------------------------------------------------
int sub_4C15D0(const uint8_t* data) {
	if (!dword_5d4594_1320964 || !data) {
		return 0;
	}
	uint32_t item_id = nox_gui_trade_read_u16(data + 2);
	nox_gui_trade_slot* found = NULL;
	for (int remote = 0; remote < 2 && !found; remote++) {
		for (int index = 0; index < NOX_GUI_TRADE_SLOT_COUNT; index++) {
			nox_gui_trade_slot* slot = nox_gui_trade_slot_at(remote, index);
			if (sub_4C1760(slot, item_id)) {
				found = slot;
				break;
			}
		}
	}
	if (!found) {
		wchar2_t* message =
			nox_strman_loadString_40F1D0("TradeGUIItemNotFound", 0, "C:\\NoxPost\\src\\client\\Gui\\GUITrade.c", 1141);
		nox_xxx_printCentered_445490(message);
		return 0;
	}
	found->total_cost -= found->total_cost / found->count;
	sub_4C1710(found, item_id);
	found->count--;
	if (found->count == 0) {
		nox_xxx_spriteDelete_45A4B0(found->drawable);
		found->drawable = NULL;
		found->total_cost = 0;
	}
	return (int)found->count;
}
