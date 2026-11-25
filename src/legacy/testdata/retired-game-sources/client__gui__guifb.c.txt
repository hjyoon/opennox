#include "common__strman.h"

#include "GAME1_2.h"
#include "client__gui__window.h"
#include "client__gui__gamewin__gamewin.h"
extern nox_window* dword_5d4594_1045636;

//----- (00456140) --------------------------------------------------------
void sub_456140(unsigned char a1) {
	wchar2_t* v3; // eax
	wchar2_t* v4; // eax
	wchar2_t* v5; // eax
	wchar2_t* v6; // eax

	*getMemU8Ptr(0x5D4594, 1045644) = a1;
	nox_window_data* draw = &dword_5d4594_1045636->draw_data;
	switch (a1) {
	case 0:
		draw->bg_image = nox_xxx_gLoadImg_42F970("BallAtHome");
		v3 = nox_strman_loadString_40F1D0("BallHomeTT", 0, "C:\\NoxPost\\src\\client\\Gui\\guifb.c", 165);
		nox_xxx_wndWddSetTooltip_46B000(draw, v3);
		break;
	case 1:
		draw->bg_image = nox_xxx_gLoadImg_42F970("BallAway");
		v4 = nox_strman_loadString_40F1D0("BallAwayTT", 0, "C:\\NoxPost\\src\\client\\Gui\\guifb.c", 170);
		nox_xxx_wndWddSetTooltip_46B000(draw, v4);
		break;
	case 2:
		draw->bg_image = nox_xxx_gLoadImg_42F970("BallRed");
		v5 = nox_strman_loadString_40F1D0("BallRedTT", 0, "C:\\NoxPost\\src\\client\\Gui\\guifb.c", 175);
		nox_xxx_wndWddSetTooltip_46B000(draw, v5);
		break;
	case 4:
		draw->bg_image = nox_xxx_gLoadImg_42F970("BallBlue");
		v6 = nox_strman_loadString_40F1D0("BallBlueTT", 0, "C:\\NoxPost\\src\\client\\Gui\\guifb.c", 180);
		nox_xxx_wndWddSetTooltip_46B000(draw, v6);
		break;
	}
}
