#include "client__system__parsecmd.h"

#include "client__gui__guicon.h"
#include "client__gui__guimsg.h"
#include "client__gui__servopts__general.h"
#include "client__gui__servopts__guiserv.h"
#include "common__log.h"
#include "common__strman.h"
#include "common__system__settings.h"
#include "server__script__script.h"
#include "common__system__team.h"

#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "GAME3.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME5_2.h"
#include "client__video__draw_common.h"
#include "common__magic__speltree.h"
#include "player_level_set_4ef410.h"

extern uint32_t nox_server_connectionType_3596;
extern uint32_t dword_5d4594_2650652;
extern uint32_t nox_player_netCode_85319C;

nox_playerInfo* nox_console_playerWhoSent_823692 = 0;

unsigned int nox_client_consoleIsServer_823684 = 0;

//----- (00440A20) --------------------------------------------------------
void sub_440A20(wchar2_t* a1, ...) {
	va_list va; // [esp+8h] [ebp+8h]

	va_start(va, a1);
	nox_vswprintf((wchar2_t*)getMemAt(0x5D4594, 822660), a1, va);
	nox_xxx_printCentered_445490((wchar2_t*)getMemAt(0x5D4594, 822660));
}

//----- (00443160) --------------------------------------------------------
int nox_cmd_ban(int tokInd, int tokCnt, wchar2_t** tokens) {
	wchar2_t* v4; // edi
	nox_playerInfo* v5; // eax
	nox_playerInfo* v6; // esi
	wchar2_t* v7; // eax
	wchar2_t* v8; // eax
	const wchar2_t* v9; // [esp-Ch] [ebp-Ch]

	if (tokCnt != 2) {
		return 0;
	}
	v4 = tokens[tokInd];
	v5 = nox_xxx_playerByName_4170D0(tokens[tokInd]);
	v6 = v5;
	if (v5) {
		if (v5->playerInd == 31) {
			v8 = nox_strman_loadString_40F1D0("cantbanyourself", 0, "C:\\NoxPost\\src\\Client\\System\\parsecmd.c",
											  3580);
			sub_440A20(v8);
			return 1;
		}
		if (nox_common_gameFlags_check_40A5C0(4096)) {
			sub_4DCFB0(v6->playerUnit);
		} else {
			nox_xxx_playerDisconnByPlrID_4DEB00(v6->playerInd);
		}
		sub_416770(0, v4, v6->serial);
		v9 = v6->name_final;
		v7 = nox_strman_loadString_40F1D0("banned", 0, "C:\\NoxPost\\src\\Client\\System\\parsecmd.c", 3577);
	} else {
		sub_416770(0, v4, 0);
		v9 = v4;
		v7 = nox_strman_loadString_40F1D0("banDisallow", 0, "C:\\NoxPost\\src\\Client\\System\\parsecmd.c", 3585);
	}
	sub_440A20(v7, v9);
	return 1;
}

//----- (00443250) --------------------------------------------------------
int nox_cmd_allow_user(int tokInd, int tokCnt, wchar2_t** tokens) {
	wchar2_t* s =
		nox_strman_loadString_40F1D0("notyetimplemented", 0, "C:\\NoxPost\\src\\Client\\System\\parsecmd.c", 3698);
	nox_gui_console_Printf_450C00(NOX_CONSOLE_RED, s);
	return 1;
}

//----- (00443280) --------------------------------------------------------
int nox_cmd_allow_ip(int tokInd, int tokCnt, wchar2_t** tokens) {
	wchar2_t* s =
		nox_strman_loadString_40F1D0("notyetimplemented", 0, "C:\\NoxPost\\src\\Client\\System\\parsecmd.c", 3704);
	nox_gui_console_Printf_450C00(NOX_CONSOLE_RED, s);
	return 1;
}

//----- (004432B0) --------------------------------------------------------
int nox_cmd_kick(int tokInd, int tokCnt, wchar2_t** tokens) {
	nox_playerInfo* v4; // eax
	nox_playerInfo* v5; // esi
	wchar2_t* v6; // eax
	wchar2_t* v7; // eax

	if (tokCnt != 2) {
		return 0;
	}
	v4 = nox_xxx_playerByName_4170D0(tokens[tokInd]);
	v5 = v4;
	if (!v4) {
		return 1;
	}
	if (v4->playerInd == 31) {
		v7 = nox_strman_loadString_40F1D0("cantkickyourself", 0, "C:\\NoxPost\\src\\Client\\System\\parsecmd.c", 3745);
		sub_440A20(v7);
		return 1;
	}
	if (nox_common_gameFlags_check_40A5C0(4096)) {
		sub_4DCFB0(v5->playerUnit);
	} else {
		nox_xxx_playerCallDisconnect_4DEAB0(v5->playerInd, 4);
		v6 = nox_strman_loadString_40F1D0("kicked", 0, "C:\\NoxPost\\src\\Client\\System\\parsecmd.c", 3739);
		sub_440A20(v6, v5->name_final);
	}
	return 1;
}

//----- (00443E90) --------------------------------------------------------
int nox_xxx_serverHandleClientConsole_443E90(nox_playerInfo* pl, char a2, wchar2_t* a3) {
	nox_playerInfo* v3; // ecx
	int result;    // eax
	int v5;        // edi
	wchar2_t* v6;   // ebx
	int v7;        // eax
	nox_object_t* v8; // eax
	wchar2_t* v9;   // eax
	wchar2_t* v10;  // esi
	int v11;       // edi
	wchar2_t* v12;  // eax
	wchar2_t* v13;  // eax
	nox_playerInfo* v14; // esi
	nox_object_t* v15; // eax
	wchar2_t* v16;  // eax
	nox_playerInfo* v17; // esi
	wchar2_t* v18;  // eax
	wchar2_t* v19; // [esp-8h] [ebp-94h]
	wchar2_t* v20; // [esp-8h] [ebp-94h]
	wchar2_t* v21;  // [esp-4h] [ebp-90h]
	char v22[128]; // [esp+Ch] [ebp-80h]

	if (!pl || !pl->playerUnit) {
		return 0;
	}
	if (a3) {
		nox_wcscpy((wchar2_t*)getMemAt(0x5D4594, 818228), a3);
	} else {
		*getMemU16Ptr(0x5D4594, 818228) = 0;
	}
	v3 = pl;
	nox_console_playerWhoSent_823692 = pl;
	if (a2 != 4 && a2 != 5 && a2) {
		if (nox_common_gameFlags_check_40A5C0(49152)) {
			return 1;
		}
	}
	switch (a2) {
	case 0:
		if (nox_common_gameFlags_check_40A5C0(8) || nox_common_gameFlags_check_40A5C0(4096) ||
			pl->field_3680 & 1) {
			nox_console_playerWhoSent_823692 = 0;
			return 1;
		}
		v5 = *getMemI16Ptr(0x5D4594, 818228) == -4083 && *getMemI16Ptr(0x5D4594, 818230) == -3923 &&
			 !*getMemU16Ptr(0x5D4594, 818232);
		if (nox_xxx_playerGoObserver_4E6860(pl, v5, 0) != 1) {
			nox_console_playerWhoSent_823692 = 0;
			return 1;
		}
		v6 = nox_strman_loadString_40F1D0("set", 0, "C:\\NoxPost\\src\\Client\\System\\parsecmd.c", 4287);
		v7 = nox_xxx_gamePlayIsAnyPlayers_40A8A0();
		if (v7) {
			if (!v5) {
				nox_xxx_netNeedTimestampStatus_4174F0(pl, 256);
			}
			if (nox_common_gameFlags_check_40A5C0(1024) && sub_40A770() == 1) {
				sub_5095E0();
			}
		}
		v8 = pl->playerUnit;
		if (v8) {
			nox_xxx_netChangeTeamMb_419570((nox_object_team_t*)&v8->field_12, pl->netCode);
		}
		v21 = v6;
		v9 = nox_strman_loadString_40F1D0("observermode", 0, "C:\\NoxPost\\src\\Client\\System\\parsecmd.c", 4308);
		nox_gui_console_Printf_450C00(NOX_CONSOLE_RED, v9, v21);
		nox_console_playerWhoSent_823692 = 0;
		return 1;
	case 1:
		if (!sub_4D12A0(pl->playerInd) && pl->playerInd != nox_player_netCode_85319C &&
			!nox_common_gameFlags_check_40A5C0(2048)) {
			nox_console_playerWhoSent_823692 = 0;
			return 1;
		}
		nox_wcstok((wchar2_t*)getMemAt(0x5D4594, 818228), L" ");
		v10 = nox_wcstok(0, L" ");
		nox_sprintf(v22, "%S", v10);
		v11 = nox_script_indexByEvent(v22);
		if (v11 != -1 && nox_console_playerWhoSent_823692) {
			v12 = nox_strman_loadString_40F1D0("ExecutingFunction", 0, "C:\\NoxPost\\src\\Client\\System\\parsecmd.c",
											   4332);
			nox_gui_console_Printf_450C00(NOX_CONSOLE_RED, v12, v10);
			nox_script_callByIndex_507310(v11, nox_console_playerWhoSent_823692->playerUnit,
										  nox_console_playerWhoSent_823692->playerUnit);
			nox_console_playerWhoSent_823692 = 0;
		} else {
			v21 = v10;
			v9 = nox_strman_loadString_40F1D0("InvalidFunction", 0, "C:\\NoxPost\\src\\Client\\System\\parsecmd.c",
											  4329);
			nox_gui_console_Printf_450C00(NOX_CONSOLE_RED, v9, v21);
			nox_console_playerWhoSent_823692 = 0;
		}
		return 1;
	case 2:
		if (!sub_4D12A0(pl->playerInd) && pl->playerInd != nox_player_netCode_85319C &&
			!nox_common_gameFlags_check_40A5C0(2048)) {
			nox_console_playerWhoSent_823692 = 0;
			return 1;
		}
		v19 = nox_console_playerWhoSent_823692->name_final;
		v13 = nox_strman_loadString_40F1D0("RemoteSysop", 0, "C:\\NoxPost\\src\\Client\\System\\parsecmd.c", 4345);
		nox_gui_console_Printf_450C00(NOX_CONSOLE_RED, v13, v19, a3);
		nox_server_parseCmdText_443C80(a3, 0);
		nox_console_playerWhoSent_823692 = 0;
		return 1;
	case 3:
		nox_xxx_printToAll_4D9FD0(0, a3);
		nox_console_playerWhoSent_823692 = 0;
		return 1;
	case 4:
		if (!(v3->field_3680 & 1) && !(nox_common_getEngineFlag(NOX_ENGINE_FLAG_REPLAY_READ))) {
			if (nox_common_gameFlags_check_40A5C0(1)) {
				v16 = nox_strman_loadString_40F1D0("notinobserver", 0, "C:\\NoxPost\\src\\Client\\System\\parsecmd.c",
												   4392);
				nox_gui_console_Printf_450C00(NOX_CONSOLE_RED, v16);
				nox_console_playerWhoSent_823692 = 0;
				return 1;
			}
			nox_console_playerWhoSent_823692 = 0;
			return 1;
		}
		if (!*a3) {
			nox_xxx_playerCameraUnlock_4E6040(v3->playerUnit);
			nox_console_playerWhoSent_823692 = 0;
			return 1;
		}
		v17 = nox_common_playerInfoGetFirst_416EA0();
		if (!v17) {
			nox_console_playerWhoSent_823692 = 0;
			return 1;
		}
		do {
			if (!_nox_wcsicmp(a3, v17->name_final)) {
				nox_xxx_playerCameraFollow_4E6060(nox_console_playerWhoSent_823692->playerUnit,
												  v17->playerUnit);
			}
			v17 = nox_common_playerInfoGetNext_416EE0(v17);
		} while (v17);
		nox_console_playerWhoSent_823692 = 0;
		return 1;
	case 5:
		nox_xxx_printToAll_4D9FD0(16, a3);
		v14 = nox_common_playerInfoGetFirst_416EA0();
		if (!v14) {
			nox_console_playerWhoSent_823692 = 0;
			return 1;
		}
		do {
			v15 = v14->playerUnit;
			if (v15) {
				nox_xxx_aud_501960(902, v15, 0, 0);
			}
			v14 = nox_common_playerInfoGetNext_416EE0(v14);
		} while (v14);
		nox_console_playerWhoSent_823692 = 0;
		return 1;
	default:
		v20 = v3->name_final;
		v18 = nox_strman_loadString_40F1D0("invalidattempt", 0, "C:\\NoxPost\\src\\Client\\System\\parsecmd.c", 4400);
		nox_gui_console_Printf_450C00(NOX_CONSOLE_RED, v18, v20, a3);
		nox_console_playerWhoSent_823692 = 0;
		return 1;
	}
	return result;
}
// 443FC8: variable 'v7' is possibly undefined
