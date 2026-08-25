#include "client__gui__gui_ctf.h"
#include "client__gui__window.h"
#include "common__strman.h"

#include "client__gui__gamewin__gamewin.h"

#include "GAME1_2.h"
#include "GAME2.h"
#include "GAME2_3.h"
extern nox_window* dword_5d4594_1045604;
nox_video_bag_image_t* nox_ctf_flag_team_border;

//----- (00455C30) --------------------------------------------------------
int sub_455C30() {
	if (dword_5d4594_1045604) {
		return 1;
	}
	dword_5d4594_1045604 = nox_new_window_from_file("GUI_CTF.wnd", 0);
	if (!dword_5d4594_1045604) {
		return 0;
	}
	for (int id = 8811; id <= 8826; id++) {
		nox_window* child = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1045604, id);
		if (!child) {
			return 0;
		}
		nox_window_set_all_funcs(child, 0, sub_455CD0, 0);
		wchar2_t* tooltip = nox_strman_loadString_40F1D0(
			"FlagHomeTT", 0, "C:\\NoxPost\\src\\client\\Gui\\GUI_CTF.c", 201);
		nox_xxx_wndWddSetTooltip_46B000(&child->draw_data, tooltip);
	}
	sub_455A00(0);
	nox_ctf_flag_team_border = nox_xxx_gLoadImg_42F970("FlagTeamBorder");
	return 1;
}

//----- (00455D80) --------------------------------------------------------
void sub_455D80(unsigned char a1, char a2) {
	wchar2_t* v4;

	*getMemU8Ptr(0x5D4594, 1045611 + a1) = a2;
	nox_window* child = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1045604, a1 + 8810);
	if (child) {
		if (child->flags & NOX_WIN_LAYER_FRONT) {
			if (a2) {
				if (a2 == 1) {
					v4 = nox_strman_loadString_40F1D0("YourFlagCarriedTT", 0,
													  "C:\\NoxPost\\src\\client\\Gui\\GUI_CTF.c", 234);
				} else {
					if (a2 != 2) {
						return;
					}
					v4 = nox_strman_loadString_40F1D0("FlagAwayTT", 0, "C:\\NoxPost\\src\\client\\Gui\\GUI_CTF.c", 238);
				}
			} else {
				v4 = nox_strman_loadString_40F1D0("FlagHomeTT", 0, "C:\\NoxPost\\src\\client\\Gui\\GUI_CTF.c", 242);
			}
		} else if (a2) {
			if (a2 == 1) {
				v4 = nox_strman_loadString_40F1D0("TheirFlagCarriedTT", 0, "C:\\NoxPost\\src\\client\\Gui\\GUI_CTF.c",
												  252);
			} else {
				if (a2 != 2) {
					return;
				}
				v4 = nox_strman_loadString_40F1D0("FlagAwayTT", 0, "C:\\NoxPost\\src\\client\\Gui\\GUI_CTF.c", 256);
			}
		} else {
			v4 = nox_strman_loadString_40F1D0("FlagHomeTT", 0, "C:\\NoxPost\\src\\client\\Gui\\GUI_CTF.c", 260);
		}
		nox_xxx_wndWddSetTooltip_46B000(&child->draw_data, v4);
	}
}
