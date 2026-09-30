#include "client__gui__guiquit.h"
#include "client__gui__window.h"
#include "common__strman.h"

#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_1.h"
#include "GAME2_3.h"
#include "GAME3.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4_1.h"
#include "client__gui__guisave.h"
#include "client__gui__guivote.h"
#include "client__gui__servopts__guiserv.h"
#include "client__system__parsecmd.h"
#include "common__system__team.h"
#include "common__log.h"
#include "operators.h"

extern uint32_t nox_player_netCode_85319C;


nox_window* nox_wnd_quitMenu_825760 = 0;

// The client-player slot is a packed PE32 pointer in the original binary.
// On native-width builds mem_getPtrValue keeps its full address in a side
// slot, while reading it through getMemU32Ptr truncates (or loses) the pointer.
int nox_xxx_quitMenuCanAutoSave_445830() {
	nox_drawable* local_player = getMemPtr(0x852978, 8);
	return local_player && !(local_player->flags30 & 0x8000);
}

//----- (00445840) --------------------------------------------------------
int nox_xxx_menuGameOnButton_445840(nox_window* root, int event, nox_window* control, uintptr_t event_arg) {
	(void)event_arg;
	if (event != 16391) {
		return 0;
	}
	int id = nox_xxx_wndGetID_46B0A0(control);
	nox_xxx_clientPlaySoundSpecial_452D80(766, 100);
	switch (id) {
	case 9001:
		sub_445C40();
		sub_413A00(1);
		if (!nox_common_gameFlags_check_40A5C0(2048) || nox_xxx_playerAnimCheck_4372B0()) {
			sub_445B40();
		} else {
			wchar2_t* message = nox_strman_loadString_40F1D0(
				"GUIQuit.c:ReallyLoadMessage", 0, "C:\\NoxPost\\src\\client\\Gui\\guiquit.c", 199);
			wchar2_t* title = nox_strman_loadString_40F1D0(
				"SelChar.c:LoadLabel", 0, "C:\\NoxPost\\src\\client\\Gui\\guiquit.c", 198);
			nox_xxx_dialogMsgBoxCreate_449A10(NULL, title, message, 56, sub_445B40, sub_445BA0);
		}
		break;
	case 9002:
		if (nox_xxx_quitMenuCanAutoSave_445830()) {
			sub_445C40();
			if (nox_common_gameFlags_check_40A5C0(2048)) {
				nox_setSaveFileName_4DB130("AUTOSAVE");
				sub_4DB170(1, 0, 0);
			}
			break;
		}
		nox_xxx_wnd_46ABB0(nox_xxx_wndGetChildByID_46B0C0(root, id), 0);
		break;
	case 9003:
		sub_445C40();
		if (nox_common_gameFlags_check_40A5C0(2048)) {
			nox_savegame_sub_46D580();
		} else {
			nox_xxx_netSavePlayer_41CE00();
		}
		if (sub_43C6E0()) {
			break;
		}
		sub_43CF70();
		break;
	case 9004:
		nox_xxx_wndClearCaptureMain_46ADE0(nox_wnd_quitMenu_825760);
		wchar2_t* message = nox_strman_loadString_40F1D0(
			"GUIQuit.c:ReallyQuitMessage", 0, "C:\\NoxPost\\src\\client\\Gui\\guiquit.c", 185);
		wchar2_t* title = nox_strman_loadString_40F1D0(
			"GUIQuit.c:ReallyQuitTitle", 0, "C:\\NoxPost\\src\\client\\Gui\\guiquit.c", 184);
		nox_xxx_dialogMsgBoxCreate_449A10(nox_wnd_quitMenu_825760, title, message, 56,
			nox_xxx_quitDialogYes_445B20, nox_xxx_quitDialogNo_445B30);
		break;
	case 9005:
		sub_445C40();
		sub_4ADA40();
		break;
	case 9006:
		sub_445C40();
		break;
	case 9007:
		if (nox_common_gameFlags_check_40A5C0(1)) {
			nox_playerInfo* player = nox_common_playerInfoGetByID_417040(nox_player_netCode_85319C);
			nox_xxx_serverHandleClientConsole_443E90(player, 0, NULL);
			sub_445C40();
		} else {
			nox_xxx_netServerCmd_440950_empty();
			sub_445C40();
		}
		break;
	case 9008:
		sub_445C40();
		nox_xxx_guiServerOptsLoad_457500();
		break;
	case 9009:
		sub_445C40();
		if (nox_common_gameFlags_check_40A5C0(4096)) {
			sub_48CB10(4);
		} else {
			sub_48CB10(0);
		}
		break;
	default:
		break;
	}
	if (control) {
		control->draw_data.field_0 &= ~UINT32_C(2);
	}
	return 0;
}

//----- (00445C40) --------------------------------------------------------
void sub_445C40() {
	nox_drawable* local_player = getMemPtr(0x852978, 8);
	nox_window* v1;  // eax
	nox_window* v2;  // eax
	nox_window* v3;  // eax
	nox_window* v4;  // eax
	nox_window* v5;  // eax
	nox_window* v6;  // esi
	nox_window* v7;  // eax
	nox_window* v8;  // eax
	nox_window* v9;  // eax
	nox_window* v10; // eax
	nox_window* v11; // esi
	nox_window* v12; // eax
	nox_window* v13; // esi
	nox_window* v14; // eax
	nox_window* v15; // eax
	nox_window* v16; // eax
	nox_window* v17; // eax
	nox_window* v18; // eax
	nox_window* v19; // eax
	wchar2_t* v20;  // [esp-4h] [ebp-8h]
	wchar2_t* v21;  // [esp-4h] [ebp-8h]

	if (nox_xxx_wndGetFlags_46ADA0(nox_wnd_quitMenu_825760) & 0x10) {
		if (!local_player || !nox_common_gameFlags_check_40A5C0(2048) ||
			(local_player->field_69 != 2 && local_player->field_69 != 1 && local_player->field_69 != 51)) {
			if (sub_45D9B0() != 1) {
				if (nox_xxx_checkGameFlagPause_413A50() != 1) {
					nox_xxx_clientPlaySoundSpecial_452D80(921, 100);
					nox_xxx_wndShowModalMB_46A8C0(nox_wnd_quitMenu_825760);
					nox_wnd_quitMenu_825760->flags |= 8u;
					nox_xxx_wndSetCaptureMain_46ADC0(nox_wnd_quitMenu_825760);
					if (nox_common_gameFlags_check_40A5C0(2048)) {
						v20 = nox_strman_loadString_40F1D0("SoloSaveLabel", 0,
														   "C:\\NoxPost\\src\\client\\Gui\\guiquit.c", 396);
						v1 = nox_xxx_wndGetChildByID_46B0C0(nox_wnd_quitMenu_825760, 9003);
						sub_46AEE0(v1, v20);
						v2 = nox_xxx_wndGetChildByID_46B0C0(nox_wnd_quitMenu_825760, 9001);
						nox_window_set_hidden(v2, 0);
						v3 = nox_xxx_wndGetChildByID_46B0C0(nox_wnd_quitMenu_825760, 9002);
						nox_window_set_hidden(v3, 0);
						v4 = nox_xxx_wndGetChildByID_46B0C0(nox_wnd_quitMenu_825760, 9007);
						nox_window_set_hidden(v4, 1);
						v5 = nox_xxx_wndGetChildByID_46B0C0(nox_wnd_quitMenu_825760, 9008);
						nox_window_set_hidden(v5, 1);
						v6 = nox_xxx_wndGetChildByID_46B0C0(nox_wnd_quitMenu_825760, 9009);
						nox_window_set_hidden(v6, 1);
						v7 = nox_xxx_wndGetChildByID_46B0C0(nox_wnd_quitMenu_825760, 9004);
						nox_window_setPos_46A9B0(v7, (int)v6->off_x, (int)v6->off_y);
						sub_413A00(1);
						sub_46AB20(nox_wnd_quitMenu_825760, 220, 285);
					} else {
						v21 = nox_strman_loadString_40F1D0("MultiplayerSaveLabel", 0,
														   "C:\\NoxPost\\src\\client\\Gui\\guiquit.c", 427);
						v8 = nox_xxx_wndGetChildByID_46B0C0(nox_wnd_quitMenu_825760, 9003);
						sub_46AEE0(v8, v21);
						v9 = nox_xxx_wndGetChildByID_46B0C0(nox_wnd_quitMenu_825760, 9001);
						nox_window_set_hidden(v9, 1);
						v10 = nox_xxx_wndGetChildByID_46B0C0(nox_wnd_quitMenu_825760, 9002);
						nox_window_set_hidden(v10, 1);
						v11 = nox_xxx_wndGetChildByID_46B0C0(nox_wnd_quitMenu_825760, 9007);
						nox_window_set_hidden(v11, 0);
						nox_xxx_wnd_46ABB0(v11, 1);
						v12 = nox_xxx_wndGetChildByID_46B0C0(nox_wnd_quitMenu_825760, 9008);
						nox_window_set_hidden(v12, 0);
						v13 = nox_xxx_wndGetChildByID_46B0C0(nox_wnd_quitMenu_825760, 9009);
						// fix demo crash -- see QuitMenu.wnd -- there is no child with id 9009
						if (v13) {
							nox_window_set_hidden(v13, 0);
							sub_46AEE0(v13, (const wchar2_t*)getMemAt(0x5D4594, 825772));
							if (nox_common_gameFlags_check_40A5C0(49152) || !nox_xxx_getTeamCounter_417DD0()) {
								nox_xxx_wnd_46ABB0(v13, 0);
							} else {
								nox_xxx_wnd_46ABB0(v13, 1);
							}
							v14 = nox_xxx_wndGetChildByID_46B0C0(nox_wnd_quitMenu_825760, 9004);
							nox_window_setPos_46A9B0(v14, (int)v13->off_x, (int)v13->off_y + 45);
						}

						sub_46AB20(nox_wnd_quitMenu_825760, 220, 330);
						if (nox_common_gameFlags_check_40A5C0(4096)) {
							v15 = nox_xxx_wndGetChildByID_46B0C0(nox_wnd_quitMenu_825760, 9007);
							nox_xxx_wnd_46ABB0(v15, 0);
							v16 = nox_xxx_wndGetChildByID_46B0C0(nox_wnd_quitMenu_825760, 9003);
							nox_xxx_wnd_46ABB0(v16, 0);
						}
						if (nox_common_getEngineFlag(NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING)) {
							v17 = nox_xxx_wndGetChildByID_46B0C0(nox_wnd_quitMenu_825760, 9007);
							nox_xxx_wnd_46ABB0(v17, 0);
							v18 = nox_xxx_wndGetChildByID_46B0C0(nox_wnd_quitMenu_825760, 9005);
							nox_xxx_wnd_46ABB0(v18, 0);
							v19 = nox_xxx_wndGetChildByID_46B0C0(nox_wnd_quitMenu_825760, 9003);
							nox_xxx_wnd_46ABB0(v19, 0);
						}
					}
				}
			}
		}
	} else {
		nox_xxx_windowFocus_46B500(0);
		nox_xxx_wndClearCaptureMain_46ADE0(nox_wnd_quitMenu_825760);
		nox_window_set_hidden(nox_wnd_quitMenu_825760, 1);
		nox_wnd_quitMenu_825760->flags &= 0xFFFFFFF7;
		sub_413A00(0);
	}
}
