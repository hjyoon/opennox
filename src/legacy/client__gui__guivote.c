#include "client__gui__guivote.h"
#include "client__gui__window.h"
#include "common__strman.h"
#include "common__system__team.h"

#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_3.h"
#include "GAME2_1.h"
extern uint32_t dword_5d4594_1197308;
extern uint32_t dword_5d4594_1197332;
extern nox_window* dword_5d4594_1197316;
extern nox_window* dword_5d4594_1197320;
extern uint32_t dword_5d4594_1197324;
extern nox_window* dword_5d4594_1197312;
extern uintptr_t dword_8531A0_2576;
extern uint32_t nox_player_netCode_85319C;

//----- (0048CB10) --------------------------------------------------------
uintptr_t sub_48CB10(int a1) {
	static const wchar2_t space[] = {' ', 0};
	int list_index = 0;
	wchar2_t text[256];

	nox_window_call_field_94(dword_5d4594_1197316, 16399, 0, 0);
	nox_window_call_field_94(dword_5d4594_1197320, 16399, 0, 0);
	dword_5d4594_1197308 = a1;
	switch (a1) {
	case 4:
		if (dword_8531A0_2576 && !((nox_playerInfo*)dword_8531A0_2576)->field_4792) {
			wchar2_t* message = nox_strman_loadString_40F1D0(
				"GUIVote.c:NotAllowedVote", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIVote.c", 452);
			wchar2_t* title = nox_strman_loadString_40F1D0(
				"guiquit.c:Vote", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIVote.c", 451);
			return (uintptr_t)nox_xxx_dialogMsgBoxCreate_449A10(NULL, title, message, 33, NULL, NULL);
		}
		nox_window_set_hidden(dword_5d4594_1197316, 1);
		nox_window_set_hidden(dword_5d4594_1197320, 0);
		nox_window* prompt = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1197312, 4301);
		sub_46AEE0(prompt, nox_strman_loadString_40F1D0(
			"SelectVoteTopic", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIVote.c", 465));
		nox_wcscpy(text, nox_strman_loadString_40F1D0(
			"VoteTopicLabel", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIVote.c", 468));
		nox_wcscat(text, space);
		nox_wcscat(text, nox_strman_loadString_40F1D0(
			"VoteResetServer", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIVote.c", 470));
		nox_window_call_field_94(dword_5d4594_1197320, 16397, (uintptr_t)text, 4);
		nox_wcscpy(text, nox_strman_loadString_40F1D0(
			"VoteTopicLabel", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIVote.c", 474));
		nox_wcscat(text, space);
		nox_wcscat(text, nox_strman_loadString_40F1D0(
			"VoteKickPlayer", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIVote.c", 476));
		nox_window_call_field_94(dword_5d4594_1197320, 16397, (uintptr_t)text, 4);
		return (uintptr_t)nox_xxx_wndShowModalMB_46A8C0(dword_5d4594_1197312);
	case 2:
		nox_window_set_hidden(dword_5d4594_1197316, 1);
		nox_window_set_hidden(dword_5d4594_1197320, 0);
		prompt = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1197312, 4301);
		sub_46AEE0(prompt, nox_strman_loadString_40F1D0(
			"Vote:ResetQuest", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIVote.c", 491));
		nox_window_call_field_94(dword_5d4594_1197320, 16397,
			(uintptr_t)nox_strman_loadString_40F1D0(
				"WindowDir:Yes", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIVote.c", 495),
			4);
		nox_window_call_field_94(dword_5d4594_1197320, 16397,
			(uintptr_t)nox_strman_loadString_40F1D0(
				"WindowDir:No", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIVote.c", 498),
			4);
		if (dword_5d4594_1197332 == 1) {
			nox_window_call_field_94(dword_5d4594_1197320, 16403, 0, 0);
		} else {
			nox_window_call_field_94(dword_5d4594_1197320, 16403, 1, 0);
		}
		return (uintptr_t)nox_xxx_wndShowModalMB_46A8C0(dword_5d4594_1197312);
	case 0:
	case 1:
	case 3:
		nox_window_set_hidden(dword_5d4594_1197316, 0);
		nox_window_set_hidden(dword_5d4594_1197320, 1);
		prompt = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1197312, 4301);
		sub_46AEE0(prompt, nox_strman_loadString_40F1D0(
			"VoteKickPlayer", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIVote.c", 520));
		nox_playerInfo* local_player = (nox_playerInfo*)dword_8531A0_2576;
		if (nox_xxx_getTeamCounter_417DD0()) {
			nox_object_team_t* membership = nox_xxx_objGetTeamByNetCode_418C80(nox_player_netCode_85319C);
			nox_team_t* team = membership ? nox_xxx_getTeamByID_418AB0(membership->id) : NULL;
			if (team) {
				for (nox_playerInfo* player = nox_common_playerInfoGetFirst_416EA0(); player;
					 player = nox_common_playerInfoGetNext_416EE0(player)) {
					if (player != local_player) {
						membership = nox_xxx_objGetTeamByNetCode_418C80(player->netCode);
						if (membership && nox_xxx_teamCompare2_419180(membership, team->field_57)) {
							nox_window_call_field_94(
								dword_5d4594_1197316, 16397, (uintptr_t)player->name_final,
								*getMemU32Ptr(0x587000, 156400 + 8 * (team->field_57 % 10)));
							int excluded_index = 0;
							if (dword_5d4594_1197324 > 0) {
								const wchar2_t* excluded = (const wchar2_t*)getMemAt(0x5D4594, 1193720);
								do {
									if (!nox_wcscmp(excluded, player->name_final)) {
										nox_window_call_field_94(dword_5d4594_1197316, 16405, list_index, 0);
									}
									++excluded_index;
									excluded += 28;
								} while (excluded_index < (int)dword_5d4594_1197324);
							}
							++list_index;
						}
					}
				}
			}
		} else {
			for (nox_playerInfo* player = nox_common_playerInfoGetFirst_416EA0(); player;
				 player = nox_common_playerInfoGetNext_416EE0(player)) {
				if (player != local_player) {
					nox_window_call_field_94(
						dword_5d4594_1197316, 16397, (uintptr_t)player->name_final, 4);
					int excluded_index = 0;
					if (dword_5d4594_1197324 > 0) {
						const wchar2_t* excluded = (const wchar2_t*)getMemAt(0x5D4594, 1193720);
						do {
							if (!nox_wcscmp(excluded, player->name_final)) {
								nox_window_call_field_94(dword_5d4594_1197316, 16405, list_index, 0);
							}
							++excluded_index;
							excluded += 28;
						} while (excluded_index < (int)dword_5d4594_1197324);
					}
					++list_index;
				}
			}
		}
		return (uintptr_t)nox_xxx_wndShowModalMB_46A8C0(dword_5d4594_1197312);
	}
	return (uintptr_t)a1;
}
