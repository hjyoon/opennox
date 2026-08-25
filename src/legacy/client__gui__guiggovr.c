#include "client__gui__guiggovr.h"
#include "client__gui__window.h"
#include "common__strman.h"

#include "GAME2.h"
#include "GAME2_1.h"

extern nox_window* dword_5d4594_1303452;
extern uintptr_t dword_8531A0_2576;
extern int nox_win_width;
extern int nox_win_height;

//----- (0049B4B0) --------------------------------------------------------
int sub_49B4B0(unsigned short* a1) {
	wchar2_t* v1;  // eax
	wchar2_t* v2;  // eax
	wchar2_t* v3;  // eax
	nox_window* v4; // eax
	nox_window* v5; // eax
	nox_window* v6; // eax
	nox_window* v7; // eax
	nox_window* v8; // eax
	nox_window* v9; // eax
	int result;   // eax
	int v11;      // [esp-4h] [ebp-10h]
	int v12;      // [esp-4h] [ebp-10h]
	int v13;      // [esp-4h] [ebp-10h]
	int v14;      // [esp+4h] [ebp-8h]
	int v15;      // [esp+8h] [ebp-4h]

	nox_window_set_hidden(dword_5d4594_1303452, 0);
	nox_xxx_wnd_46ABB0(dword_5d4594_1303452, 1);
	nox_xxx_clientPlaySoundSpecial_452D80(1007, 100);
	nox_window_get_size(dword_5d4594_1303452, &v15, &v14);
	nox_window_setPos_46A9B0(dword_5d4594_1303452, nox_win_width / 2 - v15 / 2,
							 nox_win_height / 2 - v14 / 2);
	v11 = a1[1];
	v1 = nox_strman_loadString_40F1D0("GGOver.wnd:GeneratorsDestroyed", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIGGOvr.c",
									  178);
	nox_swprintf((wchar2_t*)getMemAt(0x5D4594, 1302172), v1, v11);
	v12 = a1[2];
	v2 =
		nox_strman_loadString_40F1D0("GGOver.wnd:NumSecretsFound", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIGGOvr.c", 181);
	nox_swprintf((wchar2_t*)getMemAt(0x5D4594, 1301916), v2, v12);
	v13 = a1[3];
	v3 = nox_strman_loadString_40F1D0("GGOver.wnd:Kills", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIGGOvr.c", 183);
	nox_swprintf((wchar2_t*)getMemAt(0x5D4594, 1302428), v3, v13);
	nox_swprintf((wchar2_t*)getMemAt(0x5D4594, 1303196), (const wchar2_t*)getMemAt(0x5D4594, 1303460));
	v4 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1303452, 10710);
	nox_window_call_field_94(v4, 16385, (uintptr_t)getMemAt(0x5D4594, 1302172), 0);
	v5 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1303452, 10705);
	nox_window_call_field_94(v5, 16385, (uintptr_t)getMemAt(0x5D4594, 1302940), 0);
	v6 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1303452, 10706);
	nox_window_call_field_94(v6, 16385, (uintptr_t)getMemAt(0x5D4594, 1302684), 0);
	v7 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1303452, 10707);
	nox_window_call_field_94(v7, 16385, (uintptr_t)getMemAt(0x5D4594, 1301916), 0);
	v8 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1303452, 10708);
	nox_window_call_field_94(v8, 16385, (uintptr_t)getMemAt(0x5D4594, 1302428), 0);
	v9 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1303452, 10711);
	nox_window_call_field_94(v9, 16385, (uintptr_t)getMemAt(0x5D4594, 1303196), 0);
	result = gameFrame();
	*getMemU32Ptr(0x5D4594, 1303456) = gameFrame();
	return result;
}

//----- (0049B6E0) --------------------------------------------------------
int sub_49B6E0() {
	int result;   // eax
	int v1;       // eax
	wchar2_t* v2;  // eax
	nox_window* v3; // eax
	int v4;       // [esp-4h] [ebp-4h]

	result = dword_5d4594_1303452 != 0;
	if (dword_5d4594_1303452) {
		result = wndIsShown_nox_xxx_wndIsShown_46ACC0(dword_5d4594_1303452);
		if (!result) {
			v1 = *getMemU32Ptr(0x5D4594, 1303456) + 30 * gameFPS() - gameFrame();
			if (v1 < 0) {
				v1 = 0;
			}
			if (dword_8531A0_2576 && *(uint8_t*)(dword_8531A0_2576 + 2064) == 31) {
				nox_wcscpy((wchar2_t*)getMemAt(0x5D4594, 1301852), (const wchar2_t*)getMemAt(0x5D4594, 1303464));
			} else {
				v4 = (unsigned int)v1 / gameFPS();
				v2 = nox_strman_loadString_40F1D0("Rules.c:Time", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIGGOvr.c", 265);
				nox_swprintf((wchar2_t*)getMemAt(0x5D4594, 1301852), L"%s - %d", v2, v4);
			}
			v3 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1303452, 10712);
			result = (int)nox_window_call_field_94(v3, 16385, (uintptr_t)getMemAt(0x5D4594, 1301852), 0);
		}
	}
	return result;
}
