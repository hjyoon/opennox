#include "client__gui__chathelp.h"
#include "client__gui__window.h"

#include "GAME1.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_3.h"
#include "GAME3_2.h"
#include "client__gui__servopts__guiserv.h"
#include "client__system__ctrlevnt.h"
#include "common__strman.h"

extern nox_window* dword_5d4594_1305680;
extern int nox_win_width;
extern int nox_win_height;

//----- (0049C560) --------------------------------------------------------
void nox_xxx_cliShowHelpGui_49C560() {
	int v0;            // esi
	nox_window* result; // eax
	nox_window* v2;     // esi
	wchar2_t* v3;      // eax
	wchar2_t* v4;      // eax
	wchar2_t* v5;      // eax
	char* v6;         // [esp-4h] [ebp-8h]
	char* v7;         // [esp-4h] [ebp-8h]
	char* v8;         // [esp-4h] [ebp-8h]

	v0 = nox_strman_get_lang_code();
	if (nox_xxx_guiFontHeightMB_43F320(0) > 10) {
		v0 = 2;
	}
	result = nox_new_window_from_file((const char*)getMemPtr(0x587000, 164512 + 4 * v0), nox_xxx_wnd_49C760);
	dword_5d4594_1305680 = result;
	if (result) {
		sub_46B120(result, 0);
		nox_xxx_wndShowModalMB_46A8C0(dword_5d4594_1305680);
		sub_46C690(dword_5d4594_1305680);
		nox_xxx_windowFocus_46B500(dword_5d4594_1305680);
		nox_window_setPos_46A9B0(dword_5d4594_1305680,
								 (nox_win_width - dword_5d4594_1305680->width) / 2,
								 (nox_win_height - dword_5d4594_1305680->height) / 2);
		if (nox_common_gameFlags_check_40A5C0(1)) {
			v2 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1305680, 4102);
			v6 = sub_42E8E0(45, 1);
			v3 = nox_strman_loadString_40F1D0("Sanchlp.wnd:Help", 0, "C:\\NoxPost\\src\\client\\Gui\\chathelp.c", 85);
			nox_swprintf((wchar2_t*)getMemAt(0x5D4594, 1304656), v3, v6);
			nox_wcscat((wchar2_t*)getMemAt(0x5D4594, 1304656), L" ");
			v7 = sub_42E8E0(8, 1);
			v4 =
				nox_strman_loadString_40F1D0("cdecode.c:KeyToChat", 0, "C:\\NoxPost\\src\\client\\Gui\\chathelp.c", 87);
		} else {
			v2 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1305680, 4102);
			v8 = sub_42E8E0(45, 1);
			v5 = nox_strman_loadString_40F1D0("Sanchlp.wnd:ClientHelp", 0, "C:\\NoxPost\\src\\client\\Gui\\chathelp.c",
											  94);
			nox_swprintf((wchar2_t*)getMemAt(0x5D4594, 1304656), v5, v8);
			nox_wcscat((wchar2_t*)getMemAt(0x5D4594, 1304656), L" ");
			v7 = sub_42E8E0(8, 1);
			v4 =
				nox_strman_loadString_40F1D0("cdecode.c:KeyToChat", 0, "C:\\NoxPost\\src\\client\\Gui\\chathelp.c", 96);
		}
		nox_swprintf((wchar2_t*)getMemAt(0x5D4594, 1304400), v4, v7);
		nox_wcscat((wchar2_t*)getMemAt(0x5D4594, 1304656), (const wchar2_t*)getMemAt(0x5D4594, 1304400));
		nox_window_call_field_94(v2, 16385, (uintptr_t)getMemAt(0x5D4594, 1304656), 0);
		if (nox_common_gameFlags_check_40A5C0(1)) {
			if (!sub_459DA0()) {
				nox_xxx_guiServerOptsLoad_457500();
			}
			sub_459D80(1);
		}
		if (nox_common_gameFlags_check_40A5C0(4096) || nox_xxx_isQuest_4D6F50()) {
			sub_49C7A0();
		}
	}
}
