#include "client__gui__servopts__objlst.h"
#include "client__gui__window.h"
#include "common__strman.h"

#include "GAME1.h"
#include "GAME2.h"
#include "GAME2_3.h"
extern uint32_t dword_5d4594_1045460;
extern nox_window* dword_5d4594_1045468;
extern nox_window* dword_5d4594_1045464;

//----- (004530C0) --------------------------------------------------------
// The original table and window fields were PE32 values. The window tree and
// localized title pointers are native-width in the port, so keep them typed
// throughout this host rather than round-tripping them through int/uint32_t.
nox_window* nox_xxx_guiObjlistLoad_4530C0(nox_window* parent, int class_mask) {
	int count = 0;
	wchar2_t title[66];

	dword_5d4594_1045468 = nox_new_window_from_file("objlst.wnd", sub_4533D0);
	if (!dword_5d4594_1045468) {
		return NULL;
	}
	nox_xxx_wndSetDrawFn_46B340(dword_5d4594_1045468, sub_453350);
	sub_46B120(dword_5d4594_1045468, parent);
	nox_xxx_wnd_46B280(dword_5d4594_1045468, parent);
	dword_5d4594_1045464 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1045468, 1510);
	if (!dword_5d4594_1045464) {
		nox_xxx_windowDestroyMB_46C4E0(dword_5d4594_1045468);
		dword_5d4594_1045468 = NULL;
		return NULL;
	}
	sub_4532E0();
	nox_window_call_field_94(dword_5d4594_1045464, 16399, 0, 0);
	if (class_mask == 0x1000000) {
		dword_5d4594_1045460 = 0;
		wchar2_t* text =
			nox_strman_loadString_40F1D0("Weapons", 0, "C:\\NoxPost\\src\\client\\Gui\\ServOpts\\objlst.c", 321);
		nox_wcscpy(title, text);
		int flag = 4;
		for (int i = 0; i < 25; ++i, flag *= 2) {
			wchar2_t* name = sub_4159F0(flag);
			if (name) {
				nox_window_call_field_94(dword_5d4594_1045464, 16397, (uintptr_t)name, UINTPTR_MAX);
				++count;
			}
		}
	} else if (class_mask == 0x2000000) {
		dword_5d4594_1045460 = 1;
		wchar2_t* text = nox_strman_loadString_40F1D0("servopts.wnd:Armor", 0,
			"C:\\NoxPost\\src\\client\\Gui\\ServOpts\\objlst.c", 308);
		nox_wcscpy(title, text);
		int flag = 1;
		for (int i = 0; i < 26; ++i, flag *= 2) {
			wchar2_t* name = sub_415E80(flag);
			if (name) {
				nox_window_call_field_94(dword_5d4594_1045464, 16397, (uintptr_t)name, UINTPTR_MAX);
				++count;
			}
		}
	} else {
		title[0] = 0;
	}
	nox_window_call_field_94(dword_5d4594_1045464, 16385, (uintptr_t)title, 0);
	nox_window* up = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1045468, 1513);
	nox_window_call_field_94(dword_5d4594_1045464, 16408, (uintptr_t)up, 0);
	nox_window* down = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1045468, 1514);
	nox_window_call_field_94(dword_5d4594_1045464, 16409, (uintptr_t)down, 0);
	*getMemU32Ptr(0x5D4594, 1045472 + 4 * dword_5d4594_1045460) = count;
	sub_453750();
	if (!nox_common_gameFlags_check_40A5C0(1) || nox_common_gameFlags_check_40A5C0(49152)) {
		sub_46AD20(dword_5d4594_1045468, 1515, 1533, 0);
	}
	return dword_5d4594_1045468;
}
