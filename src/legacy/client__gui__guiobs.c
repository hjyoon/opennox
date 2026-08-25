#include "client__gui__window.h"
#include "common__strman.h"

#include "GAME1_2.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "client__gui__gamewin__gamewin.h"

extern nox_window* dword_5d4594_1193712;
extern uintptr_t dword_8531A0_2576;
extern int nox_win_width;
extern int nox_win_height;

//----- (0048C9F0) --------------------------------------------------------
int sub_48C9F0(nox_window* win, void* draw) {
	wchar2_t* v2; // eax
	unsigned int x;
	unsigned int y;

	(void)draw;
	nox_client_wndGetPosition_46AA60(win, &x, &y);
	if (dword_8531A0_2576) {
		v2 = nox_strman_loadString_40F1D0("observermode", 0, "C:\\NoxPost\\src\\client\\Gui\\guiobs.c", 41);
		nox_xxx_wndWddSetTooltip_46B000(&dword_5d4594_1193712->draw_data, v2);
		nox_client_drawImageAt_47D2C0(win->draw_data.bg_image, x + win->draw_data.img_px, y + win->draw_data.img_py);
	}
	return 1;
}

//----- (0048C980) --------------------------------------------------------
int sub_48C980() {
	setMemPtr(0x5D4594, 1193716, nox_xxx_gLoadImg_42F970("ObserverIcon"));
	dword_5d4594_1193712 = nox_window_new(0, 136, nox_win_width - 50, nox_win_height / 2 - 100, 50, 50, 0);
	nox_xxx_wndSetIcon_46AE60(dword_5d4594_1193712, getMemPtr(0x5D4594, 1193716));
	nox_window_set_all_funcs(dword_5d4594_1193712, 0, sub_48C9F0, 0);
	return 1;
}
