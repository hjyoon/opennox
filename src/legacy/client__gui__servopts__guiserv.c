#include "client__gui__servopts__guiserv.h"
#include "client__gui__window.h"
#include "common__strman.h"
#include "common__system__team.h"

#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "GAME3.h"
#include "GAME3_1.h"
#include "GAME3_2.h"
#include "GAME5_2.h"
#include "MixPatch.h"
#include "client__video__draw_common.h"
#include "common__magic__speltree.h"
#include "defs.h"
#include "operators.h"
#include "server__script__builtin.h"

#include "client__gui__gamewin__gamewin.h"
#include "client__gui__servopts__playrlst.h"
#include "client__system__parsecmd.h"

extern nox_window* dword_5d4594_1046512;
extern uint32_t dword_587000_129656;
extern void* dword_5d4594_1046360;
extern nox_window* dword_5d4594_1046524;
extern void* dword_5d4594_1046356;
extern nox_window* dword_5d4594_1046540;
extern nox_window* dword_5d4594_1046536;
extern nox_window* dword_5d4594_1046528;
extern nox_window* dword_5d4594_1046500;
extern nox_window* dword_5d4594_1046520;
extern nox_window* dword_5d4594_1046504;
extern nox_window* dword_5d4594_1046516;
extern nox_window* dword_5d4594_1046532;
extern nox_window* dword_5d4594_1046508;
extern nox_window* dword_5d4594_1046496;
extern nox_window* dword_5d4594_1046492;
extern int nox_win_width;
extern uint32_t nox_color_white_2523948;

//----- (004585D0) --------------------------------------------------------
void sub_4537F0();
int nox_xxx_guiServerOptionsProcPre_4585D0(nox_window* a1, unsigned int a2, uintptr_t a3, uintptr_t a4) {
	nox_window* v4;         // esi
	char* v5;               // ebp
	int result;             // eax
	wchar2_t* v7;           // eax
	int v8;                 // ebx
	int v9;                 // edi
	int* v10;               // eax
	nox_window* v11;        // edi
	nox_window* v12;        // esi
	nox_window* v13;        // eax
	nox_window* v14;        // eax
	nox_window* v15;        // eax
	nox_window* v16;        // edi
	nox_window* v17;        // eax
	char* v18;              // esi
	int v19;                // edi
	short v20;              // ax
	bool v21;               // cc
	wchar2_t* v22;           // eax
	wchar2_t* v23;           // eax
	wchar2_t* v24;           // eax
	char* v25;              // esi
	nox_window* v26;        // esi
	int v27;                // edi
	int v28;                // eax
	char* v29;              // esi
	nox_window* v30;        // edi
	intptr_t v31;           // eax
	wchar2_t* v32;           // ebp
	int v33;                // edi
	unsigned int v34;       // ecx
	wchar2_t* v35;          // eax
	char* v36;              // edi
	char* v37;              // ebx
	wchar2_t* v38;           // [esp-10h] [ebp-F0h]
	wchar2_t* v39;           // [esp-10h] [ebp-F0h]
	wchar2_t WideCharStr[4]; // [esp+10h] [ebp-D0h]
	char v41[100];          // [esp+18h] [ebp-C8h]
	char v42[100];          // [esp+7Ch] [ebp-64h]
	nox_window* event_win = (nox_window*)a3;
	(void)a1;

	if (a2 > 16391) {
		if (a2 == 16400) {
			v28 = nox_xxx_wndGetID_46B0A0(event_win) - 10114;
			if (v28) {
				if (v28 == 6) {
					v29 = sub_4165B0();
					v30 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1046492, 10119);
					v31 = (intptr_t)nox_window_call_field_94(event_win, 16404, 0, 0);
					if (v31 >= 0 && v31 < ((nox_scrollListBox_data*)event_win->widget_data)->field_11_0) {
						v32 = (wchar2_t*)nox_window_call_field_94(event_win, 16406, a4, 0);
						nox_window_call_field_94(v30, 16385, (uintptr_t)v32, -1);
						v33 = *((uint16_t*)v29 + 26) & 0x17F0;
						*((uint16_t*)v29 + 26) &= 0xE80Fu;
						*((uint16_t*)v29 + 26) |= sub_459650(v32);
						if (v33 != (*((uint16_t*)v29 + 26) & 0x17F0)) {
							nox_client_guiserv_updateMapList_458230(*((unsigned short*)v29 + 26),
																	(char*)getMemAt(0x5D4594, 1046556), 1);
						}
						nox_window_set_hidden(event_win, 1);
						nox_xxx_wndClearCaptureMain_46ADE0(event_win);
						*((uint16_t*)v29 + 27) = nox_xxx_servGamedataGet_40A020(*((uint16_t*)v29 + 26));
						v29[56] = sub_40A180(*((uint16_t*)v29 + 26));
						sub_457460(v29);
						sub_459D50(1);
						if (v29[53] & 0x10) {
							nox_xxx_spellEnableAll_424BD0();
							sub_4537F0();
							sub_459700();
						}
						LOWORD(v34) = *((uint16_t*)v29 + 26);
						sub_457F30((v34 >> 12) & 1);
						return 1;
					}
				}
			} else if ((intptr_t)nox_window_call_field_94(event_win, 16404, 0, 0) >= 0) {
				v35 = (wchar2_t*)nox_window_call_field_94(event_win, 16406, a4, 0);
				nox_sprintf(v42, "%S", v35);
				v36 = strtok(v42, "\t");
				v37 = sub_4165B0();
				if (!v36) {
					v36 = "";
				}
				strcpy(v37, v36);
				sub_57A1E0((int*)v37, "user.rul", 0, 7, *((uint16_t*)v37 + 26));
				sub_459880(v37);
				sub_459D50(1);
			}
		}
		return 1;
	}
	if (a2 == 16391) {
		v9 = nox_xxx_wndGetID_46B0A0(event_win);
		nox_xxx_clientPlaySoundSpecial_452D80(766, 100);
		switch (v9) {
		case 10119:
			v26 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1046492, 10120);
			nox_window_set_hidden(v26, 0);
			nox_xxx_wndSetCaptureMain_46ADC0(v26);
			nox_xxx_wndShowModalMB_46A8C0(v26);
			nox_xxx_windowServerOptionsFillGametypeList_4596A0();
			return 1;
		case 10122:
			v27 = (event_win->draw_data.field_0 & 4) != 0;
			if (!nox_common_gameFlags_check_40A5C0(49152)) {
				nox_xxx_wnd_46ABB0(dword_5d4594_1046500, v27);
				nox_xxx_wnd_46ABB0(dword_5d4594_1046504, v27);
			}
			sub_459D50(1);
			return 1;
		case 10141:
			sub_459700();
			return 1;
		case 10145:
			v18 = nox_xxx_cliGamedataGet_416590(1);
			sub_459AA0(v18);
			sub_4165F0(1, 0);
			v19 = nox_xxx_getTeamCounter_417DD0();
			if (nox_common_gameFlags_check_40A5C0(128) && (v20 = *((uint16_t*)v18 + 26), v20 & 0x60)) {
				if (v20 < 0) {
					v21 = v19 <= 2;
				} else {
					v21 = v19 <= 2;
					if (v19 < 2) {
						v22 = nox_strman_loadString_40F1D0("NeedTeams", 0,
														   "C:\\NoxPost\\src\\client\\Gui\\ServOpts\\guiserv.c", 1363);
						nox_xxx_dialogMsgBoxCreate_449A10(dword_5d4594_1046492,
													  (wchar2_t*)getMemAt(0x5D4594, 1046560), v22, 56, sub_459150,
													  0);
						sub_44A360(1);
						return 1;
					}
				}
				if (!v21) {
					v38 = nox_strman_loadString_40F1D0("TooManyTeams", 0,
													   "C:\\NoxPost\\src\\client\\Gui\\ServOpts\\guiserv.c", 1371);
					v23 = nox_strman_loadString_40F1D0("Notice", 0,
													   "C:\\NoxPost\\src\\client\\Gui\\ServOpts\\guiserv.c", 1370);
					nox_xxx_dialogMsgBoxCreate_449A10(dword_5d4594_1046492, v23, v38, 33, 0, 0);
					sub_44A360(1);
					return 1;
				}
			} else if (nox_common_gameFlags_check_40A5C0(128) && nox_common_gameFlags_check_40A5C0(16) &&
					   nox_xxx_CheckGameplayFlags_417DA0(4) && v19 > 2) {
				v39 = nox_strman_loadString_40F1D0("TooManyTeams", 0,
												   "C:\\NoxPost\\src\\client\\Gui\\ServOpts\\guiserv.c", 1384);
				v24 = nox_strman_loadString_40F1D0("Notice", 0, "C:\\NoxPost\\src\\client\\Gui\\ServOpts\\guiserv.c",
												   1383);
				nox_xxx_dialogMsgBoxCreate_449A10(dword_5d4594_1046492, v24, v39, 33, 0, 0);
				sub_44A360(1);
				return 1;
			}
			sub_459150();
			result = 1;
			break;
		case 10146:
			v25 = sub_4165B0();
			if (nox_common_gameFlags_check_40A5C0(128)) {
				if ((int)v25[53] < 0) {
					nox_server_teamsZzz_419030(1);
				}
				*((uint16_t*)v25 + 26) &= 0x3FFFu;
			}
			nox_xxx_guiServerOptionsHide_4597E0(0);
			return 1;
		case 10149:
			nox_xxx_guiServerOptionsHide_4597E0(0);
			return 1;
		case 10152:
			v10 = (int*)nox_xxx_cliGamedataGet_416590(1);
			nox_xxx_loadAdvancedWnd_4BDC10(v10);
			return 1;
		case 10159:
			v11 = event_win->parent;
			sub_46B120(event_win, 0);
			sub_46B120(event_win, v11);
			v12 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1046492, 10196);
			if (nox_common_gameFlags_check_40A5C0(1)) {
				nox_window_set_hidden(v12, 0);
				sub_46ACE0(dword_5d4594_1046492, 10161, 10163, 0);
				sub_46AD20(dword_5d4594_1046492, 10161, 10163, 1);
				sub_4593B0(0);
				v13 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1046492, 10163);
				v13->draw_data.field_0 |= 4u;
				v14 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1046492, 10161);
				v14->draw_data.field_0 &= 0xFFFFFFFB;
				v15 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1046492, 10162);
				v15->draw_data.field_0 &= 0xFFFFFFFB;
				sub_46ACE0(dword_5d4594_1046492, 10141, 10141, 1);
			} else {
				nox_window_set_hidden(v12, 1);
				dword_5d4594_1046532 = nox_xxx_guiServerPlayersLoad_456270(dword_5d4594_1046492);
				sub_46ACE0(dword_5d4594_1046492, 10141, 10141, 1);
			}
			nox_window_set_hidden(dword_5d4594_1046524, 1);
			return 1;
		case 10160:
			v16 = event_win->parent;
			sub_46B120(event_win, 0);
			sub_46B120(event_win, v16);
			if (dword_5d4594_1046532) {
				sub_456D60(1);
				dword_5d4594_1046532 = 0;
			}
			if (dword_5d4594_1046528) {
				sub_4557D0(1);
				dword_5d4594_1046528 = 0;
			}
			if (dword_5d4594_1046536) {
				sub_4AD820();
				dword_5d4594_1046540 = 0;
			}
			sub_46ACE0(dword_5d4594_1046492, 10161, 10163, 1);
			if (nox_common_gameFlags_check_40A5C0(1)) {
				sub_46ACE0(dword_5d4594_1046492, 10141, 10141, 0);
			}
			v17 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1046492, 10196);
			nox_window_set_hidden(v17, 1);
			nox_window_set_hidden(dword_5d4594_1046524, 0);
			return 1;
		case 10161:
			sub_4593B0(1);
			return 1;
		case 10162:
			sub_4593B0(2);
			return 1;
		case 10163:
			sub_4593B0(0);
			return 1;
		case 10330:
			if (nox_xxx_CheckGameplayFlags_417DA0(4)) {
				nox_xxx_toggleAllTeamFlags_418690(0);
				nox_server_teamsZzz_419030(1);
				sub_46AD20(dword_5d4594_1046508, 10331, 10333, 0);
				result = 1;
			} else {
				nox_xxx_wndGuiTeamCreate_4185B0();
				if (nox_xxx_CheckGameplayFlags_417DA0(2)) {
					sub_4181F0(0);
				} else {
					nox_xxx_toggleAllTeamFlags_418690(1);
				}
				sub_46AD20(dword_5d4594_1046508, 10331, 10333, 1);
				result = 1;
			}
			return result;
		case 10331:
			if (nox_xxx_CheckGameplayFlags_417DA0(2)) {
				nox_xxx_UnsetGameplayFlags_417D70(2);
				nox_xxx_toggleAllTeamFlags_418690(1);
			} else {
				nox_xxx_toggleAllTeamFlags_418690(0);
				sub_418390();
			}
			return 1;
		case 10332:
			sub_4181F0(1);
			return 1;
		case 10333:
			if (nox_xxx_CheckGameplayFlags_417DA0(1)) {
				nox_xxx_UnsetGameplayFlags_417D70(1);
			} else {
				nox_xxx_SetGameplayFlag_417D50(1);
			}
			return 1;
		default:
			return 1;
		}
		return result;
	}
	if (a2 == 23) {
		return 0;
	}
	if (a2 != 16387) {
		return 1;
	}
	v4 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1046492, (int)a4);
	v5 = sub_4165B0();
	if (!v4) {
		return 0;
	}
	if ((unsigned short)a3 == 1) {
		if (a4 == 10101) {
			v4->draw_data.text_color = nox_color_white_2523948;
			return 1;
		}
		return 1;
	}
	v7 = (wchar2_t*)nox_window_call_field_94(v4, 16413, 0, 0);
	nox_sprintf(v41, "%S", v7);
	if (!v41[0]) { // TODO: was if (!v41 || ...
		return 1;
	}
	v8 = atoi(v41);
	if (v8 < 0) {
		v8 = 0;
	}
	if (a4 == 10101) {
		v4->draw_data.text_color = nox_color_rgb_4344A0(230, 165, 65);
		strcpy(v5 + 9, v41);
		nox_xxx_gameSetServername_40A440(v41);
		return 1;
	}
	if (a4 == 10134) {
		*((uint16_t*)v5 + 27) = v8;
		sub_459D50(1);
		return 1;
	}
	if (a4 != 10135) {
		return 1;
	}
	if (v8 > 255) {
		LOBYTE(v8) = -1;
		nox_itow(255, WideCharStr, 10);
		nox_window_call_field_94(v4, 16414, (uintptr_t)WideCharStr, -1);
	}
	v5[56] = v8;
	sub_459D50(1);
	return 1;
}
