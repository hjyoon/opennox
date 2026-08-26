#include "client__gui__guishop.h"
#include "client__gui__window.h"
#include "common__net_list.h"
#include "common__strman.h"

#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME3_1.h"
#include "client__gui__guimsg.h"
#include "input_common.h"
#include "operators.h"

extern nox_video_bag_image_t* dword_5d4594_1098456;
extern uint32_t dword_5d4594_1098620;
extern uint32_t dword_5d4594_1098596;
extern uint32_t dword_5d4594_1098600;
extern uint32_t dword_5d4594_1098616;
extern uint32_t dword_5d4594_1098604;
extern uint32_t dword_5d4594_1098592;
extern nox_window* dword_5d4594_1098580;
extern uint32_t dword_5d4594_1098624;
extern uint32_t dword_5d4594_1107036;
extern uint32_t dword_5d4594_1098628;
extern nox_window* dword_5d4594_1098576;
extern uint32_t nox_color_white_2523948;

//----- (00478730) --------------------------------------------------------
void sub_478730(int* a1) {
	int column = (*a1 - *getMemU32Ptr(0x5D4594, 1098380)) / 50;
	int row = (a1[1] - *getMemU32Ptr(0x5D4594, 1098384) + dword_5d4594_1107036) / 50;
	if (column >= NOX_SHOP_INVENTORY_COLUMN_COUNT) {
		column = NOX_SHOP_INVENTORY_COLUMN_COUNT - 1;
	}
	if (row >= NOX_SHOP_INVENTORY_ROW_COUNT) {
		row = NOX_SHOP_INVENTORY_ROW_COUNT - 1;
	}
	nox_shop_inventory_cell_t* cell = nox_client_shop_inventory_cell(row, column);
	if (!cell->count) {
		return;
	}
	uint32_t gold = sub_4674A0();
	uint32_t quantity = cell->count;
	if (gold < quantity * cell->price) {
		quantity = gold / cell->price;
	}
	if (!quantity) {
		sub_479520(cell->price - gold);
		return;
	}
	const void* modifiers = NULL;
	if (cell->drawable->flags28 & 0x13001000) {
		modifiers = cell->drawable->item_modifiers;
	}
	sub_4C05F0(1, cell->price);
	uint32_t net_code = cell->net_codes[cell->count - 1];
	wchar2_t* label = nox_strman_loadString_40F1D0("BuyLabel", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIShop.c", 328);
	nox_gui_itemAmountDialog_4C0430(label, a1[0], a1[1], net_code, cell->drawable->field_27, modifiers,
		quantity, 0, sub_478850, 0);
}

//----- (00478880) --------------------------------------------------------
void nox_client_tradeXxxBuyAccept_478880(int a1, short a2) {
	wchar2_t* v2; // eax

	if (sub_467B00(a1, 1)) {
		LOWORD(a1) = 5833;
		HIWORD(a1) = a2;
		nox_netlist_addToMsgListCli_40EBC0(31, 0, &a1, 4);
	} else {
		nox_xxx_clientPlaySoundSpecial_452D80(925, 100);
		v2 = nox_strman_loadString_40F1D0("pickup.c:CarryingTooMuch", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIShop.c",
										  207);
		nox_xxx_printCentered_445490(v2);
	}
}

//----- (004788F0) --------------------------------------------------------
void sub_4788F0(int a1, int a2) {
	wchar2_t* v2; // eax
	char v3[5];  // [esp+8h] [ebp-8h]

	if (sub_467B00(a1, a2)) {
		v3[0] = -55;
		v3[1] = 23;
		*(uint16_t*)&v3[2] = a1;
		v3[4] = a2;
		nox_netlist_addToMsgListCli_40EBC0(31, 0, v3, 5);
	} else {
		nox_xxx_clientPlaySoundSpecial_452D80(925, 100);
		v2 = nox_strman_loadString_40F1D0("pickup.c:CarryingTooMuch", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIShop.c",
										  233);
		nox_xxx_printCentered_445490(v2);
	}
}

//----- (00478B10) --------------------------------------------------------
wchar2_t* sub_478B10(int2* a1) {
	uint32_t* v1;    // esi
	wchar2_t* result; // eax
	int v3;          // [esp+4h] [ebp-10h]
	int v4;          // [esp+8h] [ebp-Ch]
	int v5;          // [esp+Ch] [ebp-8h]
	int v6;          // [esp+10h] [ebp-4h]

	v1 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1098576, 3806);
	nox_client_wndGetPosition_46AA60(v1, &v5, &v6);
	nox_window_get_size((int)v1, &v4, &v3);
	nox_client_drawImageAt_47D2C0(dword_5d4594_1098456, a1->field_0, a1->field_4);
	result = *(wchar2_t**)&dword_5d4594_1098596;
	if (dword_5d4594_1098596 ||
		(result = nox_strman_loadString_40F1D0("SellInstructions", *(uint32_t**)&dword_5d4594_1098596,
											   "C:\\NoxPost\\src\\client\\Gui\\GUIShop.c", 597),
		 (dword_5d4594_1098596 = result) != 0)) {
		result = (wchar2_t*)nox_xxx_drawStringWrap_43FAF0(0, result, v5 + 8, v6 + 8, v4 - 16, v3 - 16);
	}
	return result;
}

//----- (00478BC0) --------------------------------------------------------
wchar2_t* sub_478BC0(int* a1) {
	uint32_t* v1;    // esi
	wchar2_t* result; // eax
	int v3;          // [esp+4h] [ebp-10h]
	int v4;          // [esp+8h] [ebp-Ch]
	int v5;          // [esp+Ch] [ebp-8h]
	int v6;          // [esp+10h] [ebp-4h]

	v1 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1098576, 3806);
	nox_client_wndGetPosition_46AA60(v1, &v5, &v6);
	nox_window_get_size((int)v1, &v4, &v3);
	nox_xxx_drawSetTextColor_434390(nox_color_white_2523948);
	nox_client_drawImageAt_47D2C0(dword_5d4594_1098456, *a1, a1[1]);
	result = *(wchar2_t**)&dword_5d4594_1098600;
	if (dword_5d4594_1098600 ||
		(result = nox_strman_loadString_40F1D0("RepairInstructions", *(uint32_t**)&dword_5d4594_1098600,
											   "C:\\NoxPost\\src\\client\\Gui\\GUIShop.c", 628),
		 (dword_5d4594_1098600 = result) != 0)) {
		result = (wchar2_t*)nox_xxx_drawStringWrap_43FAF0(0, result, v5 + 8, v6 + 8, v4 - 16, v3 - 16);
	}
	return result;
}

//----- (00478FD0) --------------------------------------------------------
int nox_xxx_cliStartShopDlg_478FD0(const wchar2_t* a1, char* a2, int a3) {
	uint32_t* v3; // esi

	v3 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1098576, 3810);
	sub_445C20();
	dword_5d4594_1098624 = 1;
	dword_5d4594_1098628 = 1;
	nox_window_set_hidden(dword_5d4594_1098576, 0);
	nox_xxx_wnd_46ABB0(dword_5d4594_1098576, 1);
	nox_xxx_wndShowModalMB_46A8C0(dword_5d4594_1098576);
	*getMemU32Ptr(0x5D4594, 1098612) = nox_client_getRenderGUI();
	nox_client_setRenderGUI(0);
	sub_467BB0();
	if (a1) {
		nox_wcscpy((wchar2_t*)getMemAt(0x5D4594, 1097300), a1);
	} else {
		nox_wcscpy((wchar2_t*)getMemAt(0x5D4594, 1097300), (const wchar2_t*)getMemAt(0x5D4594, 1107044));
	}
	sub_46AEE0((int)v3, (int)getMemAt(0x5D4594, 1097300));
	if (strlen(a2)) {
		dword_5d4594_1098604 = nox_strman_loadString_40F1D0(a2, getMemAt(0x5D4594, 1098608),
															"C:\\NoxPost\\src\\client\\Gui\\GUIShop.c", 1128);
	} else {
		dword_5d4594_1098604 = 0;
		*getMemU32Ptr(0x5D4594, 1098608) = 0;
	}
	nox_xxx_getShopPic_4790F0(a3);
	if (*getMemU32Ptr(0x5D4594, 1098608)) {
		nox_xxx_playDialogFile_44D900(*getMemIntPtr(0x5D4594, 1098608), 100);
	}
	dword_5d4594_1107036 = 0;
	return nox_window_call_field_94(dword_5d4594_1098580, 16394, dword_5d4594_1098592, 0);
}

//----- (00479520) --------------------------------------------------------
void sub_479520(int a1) {
	wchar2_t* v1; // eax
	wchar2_t* v2; // eax

	v1 = nox_strman_loadString_40F1D0("NotEnoughGold", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIShop.c", 1346);
	nox_swprintf((wchar2_t*)getMemAt(0x5D4594, 1097352), v1, a1);
	v2 = nox_strman_loadString_40F1D0("ShopInformationTitle", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIShop.c", 1350);
	nox_xxx_dialogMsgBoxCreate_449A10(dword_5d4594_1098576, v2, (wchar2_t*)getMemAt(0x5D4594, 1097352), 33, 0, 0);
	nox_xxx_clientPlaySoundSpecial_452D80(925, 100);
}

void sub_479680() { dword_5d4594_1098616 = 0; }

//----- (004795E0) --------------------------------------------------------
int sub_4795E0(int a1, int a2) {
	const void* v2; // ebp
	int result;     // eax
	int v9;         // [esp-18h] [ebp-28h]
	int v10;        // [esp-10h] [ebp-20h]

	v2 = 0;
	nox_point mpos = nox_client_getMousePos_4309F0();
	result = dword_5d4594_1098616;
	if (dword_5d4594_1098616 != 1) {
		nox_drawable* drawable = sub_4676D0(a1);
		if (drawable) {
			if (drawable->flags28 & 0x13001000) {
				v2 = drawable->item_modifiers;
			}
			sub_4C05F0(1, a2);
			v10 = sub_467700(a1);
			v9 = drawable->field_27;
			wchar2_t* str =
				nox_strman_loadString_40F1D0("SellLabel", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIShop.c", 1504);
			result = nox_gui_itemAmountDialog_4C0430(str, mpos.x, mpos.y, a1, v9, v2, v10, 0, sub_479690, sub_479680);
			dword_5d4594_1098616 = 1;
		}
	}
	return result;
}

//----- (00479740) --------------------------------------------------------
void sub_479740(int a1, unsigned int a2) {
	const void* v2;   // ebp
	wchar2_t* v6;      // eax
	int v7;           // [esp-24h] [ebp-38h]
	int v8;           // [esp-20h] [ebp-34h]
	int v9;           // [esp-18h] [ebp-2Ch]
	unsigned int v10; // [esp+10h] [ebp-4h]

	v2 = 0;
	nox_point mpos = nox_client_getMousePos_4309F0();
	v10 = sub_4674A0();
	if (dword_5d4594_1098620 != 1) {
		nox_drawable* drawable = sub_4676D0(a1);
		if (drawable) {
			if (drawable->flags28 & 0x13001000) {
				v2 = drawable->item_modifiers;
			}
			sub_4C05F0(1, a2);
			if (a2 > v10) {
				sub_479520(a2 - v10);
				sub_467680();
			} else {
				v9 = drawable->field_27;
				v8 = mpos.y;
				v7 = mpos.x;
				v6 = nox_strman_loadString_40F1D0("RepairLabel", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIShop.c", 1580);
				nox_gui_itemAmountDialog_4C0430(v6, v7, v8, a1, v9, v2, 1, 0, sub_479820, sub_479810);
				dword_5d4594_1098620 = 1;
			}
		}
	}
}
