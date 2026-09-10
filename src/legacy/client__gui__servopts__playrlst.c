#include "client__gui__servopts__playrlst.h"
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
#include "common__system__team.h"

nox_gui_server_players_state nox_gui_server_players;

static void nox_gui_server_players_setup_scrollbar(nox_window* root, nox_window* list, int thumb_id, int up_id,
	int down_id, void* slider, void* slider_lit) {
	if (!root || !list) {
		return;
	}
	nox_scrollListBox_data* data = list->widget_data;
	nox_window* thumb = nox_xxx_wndGetChildByID_46B0C0(root, thumb_id);
	nox_window* up = nox_xxx_wndGetChildByID_46B0C0(root, up_id);
	nox_window* down = nox_xxx_wndGetChildByID_46B0C0(root, down_id);
	if (!data || !thumb || !up || !down) {
		return;
	}
	if (thumb->field_100) {
		thumb->field_100->width = 16;
		thumb->field_100->height = 10;
	}
	sub_4B5700(thumb, NULL, NULL, slider, slider_lit, slider_lit);
	nox_xxx_wnd_46B280(thumb, list);
	nox_xxx_wnd_46B280(up, list);
	nox_xxx_wnd_46B280(down, list);
	data->field_9 = thumb;
	data->field_7 = up;
	data->field_8 = down;
}

//----- (00456270) --------------------------------------------------------
nox_window* nox_xxx_guiServerPlayersLoad_456270(nox_window* parent) {
	int lang = nox_strman_get_lang_code();
	if (nox_xxx_guiFontHeightMB_43F320(NULL) > 10) {
		lang = 2;
	}
	if (dword_5d4594_1045684) {
		return NULL;
	}
	dword_5d4594_1045684 = nox_new_window_from_file(
		(char*)getMemPtr(0x587000, 129048 + 4 * lang), sub_4567C0);
	if (!dword_5d4594_1045684) {
		return NULL;
	}
	dword_5d4594_1045688 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1045684, 10507);
	dword_5d4594_1045692 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1045684, 10509);
	sub_46B120(dword_5d4594_1045684, parent);
	nox_xxx_wndSetDrawFn_46B340(dword_5d4594_1045684, sub_456640);
	nox_xxx_wndRetNULL_46A8A0();

	// The PE32 list heads at 0x5D4594+1045652/+1045668 are only 16 bytes
	// apart. A native list head is 24 bytes, so storing it there corrupts the
	// neighboring head and eventually GUI callback pointers.
	nox_common_list_clear_425760(&nox_gui_server_players.players);
	nox_common_list_clear_425760(&nox_gui_server_players.teams);

	void* slider = nox_xxx_gLoadImg_42F970("UISlider");
	void* slider_lit = nox_xxx_gLoadImg_42F970("UISliderLit");
	nox_window* players = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1045684, 10501);
	nox_window* teams = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1045684, 10502);
	nox_gui_server_players_setup_scrollbar(dword_5d4594_1045684, players, 10517, 10515, 10516, slider,
		slider_lit);
	nox_gui_server_players_setup_scrollbar(dword_5d4594_1045684, teams, 10520, 10518, 10519, slider,
		slider_lit);

	sub_456500();
	nox_window* title = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1045684, 10504);
	if (title && nox_common_gameFlags_check_40A5C0(128)) {
		wchar2_t* text =
			nox_strman_loadString_40F1D0("Title1", 0, "C:\\NoxPost\\src\\client\\Gui\\ServOpts\\playrlst.c", 631);
		nox_window_call_field_94(title, 16385, (uintptr_t)text, UINTPTR_MAX);
	}
	return dword_5d4594_1045684;
}

static nox_playerInfo* nox_gui_server_players_player_by_name(const wchar2_t* name) {
	if (!name) {
		return NULL;
	}
	for (nox_playerInfo* player = nox_common_playerInfoGetFirst_416EA0(); player;
		 player = nox_common_playerInfoGetNext_416EE0(player)) {
		if (!_nox_wcsicmp(player->name_final, name)) {
			return player;
		}
	}
	return NULL;
}

//----- (004567C0) --------------------------------------------------------
int sub_4567C0(nox_window* win, int event, nox_window* event_win, uintptr_t event_arg) {
	(void)win;
	(void)event_arg;
	wchar2_t team_name[56];

	if (!dword_5d4594_1045684 || !event_win) {
		return 0;
	}
	if (event == 16400 && nox_xxx_wndGetID_46B0A0(event_win) == 10502) {
		char* settings = sub_4165B0();
		if ((int)nox_window_call_field_94(event_win, 16404, 0, 0) < 0 ||
			nox_common_gameFlags_check_40A5C0(0x8000) || !settings || (int8_t)settings[53] < 0) {
			nox_xxx_wnd_46ABB0(dword_5d4594_1045688, 0);
			nox_xxx_wnd_46ABB0(dword_5d4594_1045692, 0);
		} else {
			if (nox_common_gameFlags_check_40A5C0(1)) {
				nox_xxx_wnd_46ABB0(dword_5d4594_1045692, 1);
			}
			nox_xxx_wnd_46ABB0(dword_5d4594_1045688,
				nox_common_gameFlags_check_40A5C0(128) || !*getMemU32Ptr(0x5D4594, 1045696));
		}
		if (nox_common_gameFlags_check_40A5C0(1) &&
			nox_common_getEngineFlag(NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING)) {
			nox_xxx_wnd_46ABB0(dword_5d4594_1045688, 0);
		}
	}

	if (event != 16391 && event != 16400) {
		return 0;
	}
	int id = nox_xxx_wndGetID_46B0A0(event_win);
	nox_xxx_clientPlaySoundSpecial_452D80(766, 100);
	switch (id) {
	case 10509: {
		wchar2_t* prompt =
			nox_strman_loadString_40F1D0("NewName", 0, "C:\\NoxPost\\src\\client\\Gui\\ServOpts\\playrlst.c", 504);
		wchar2_t* title =
			nox_strman_loadString_40F1D0("Rename", 0, "C:\\NoxPost\\src\\client\\Gui\\ServOpts\\playrlst.c", 504);
		nox_xxx_dialogMsgBoxCreate_449A10(dword_5d4594_1045684, title, prompt, 163, NULL, NULL);
		return 0;
	}
	case 10507: {
		nox_window* teams = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1045684, 10502);
		if (!teams) {
			return 0;
		}
		int selected = (int)nox_window_call_field_94(teams, 16404, 0, 0);
		if (!sub_456D00(selected, team_name)) {
			return 0;
		}
		nox_team_t* team = sub_418A40(team_name);
		if (team) {
			sub_456BB0(team);
		}
		return 0;
	}
	case 4001: {
		nox_window* teams = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1045684, 10502);
		if (!teams) {
			return 0;
		}
		int selected = (int)nox_window_call_field_94(teams, 16404, 0, 0);
		if (!sub_456D00(selected, team_name)) {
			return 0;
		}
		nox_team_t* team = sub_418A40(team_name);
		wchar2_t* new_name = (wchar2_t*)(uintptr_t)sub_449E60(168);
		if (team && new_name && !sub_4190F0(new_name)) {
			nox_xxx_teamRenameMB_418CD0(team, new_name);
		}
		return 0;
	}
	case 10503: {
		nox_window* teams = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1045684, 10502);
		if (!teams) {
			return 0;
		}
		int selected = (int)nox_window_call_field_94(teams, 16404, 0, 0);
		if (selected >= 0) {
			if (!sub_456D00(selected, team_name)) {
				return 0;
			}
			nox_team_t* team = sub_418A40(team_name);
			nox_window* players = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1045684, 10501);
			if (!players) {
				return 0;
			}
			int* selections = (int*)(uintptr_t)nox_window_call_field_94(players, 16404, 0, 0);
			if (team && selections) {
				for (int* selection = selections; *selection >= 0; ++selection) {
					const wchar2_t* name = (const wchar2_t*)nox_window_call_field_94(players, 16406, *selection, 0);
					nox_playerInfo* player = nox_gui_server_players_player_by_name(name);
					if (!player || (player->field_3680 & 1) || (player->field_4 & 1)) {
						continue;
					}
					nox_object_team_t* value = nox_xxx_objGetTeamByNetCode_418C80(player->netCode);
					if (!value) {
						continue;
					}
					if (nox_xxx_servObjectHasTeam_419130(value)) {
						sub_4196D0(value, team, player->netCode, 1);
					} else {
						nox_xxx_createAtImpl_4191D0(team->field_57, value, 1, player->netCode, 1);
					}
				}
			}
		}
		nox_window_call_field_94(teams, 16403, UINTPTR_MAX, 0);
		nox_window* players = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1045684, 10501);
		if (players) {
			nox_window_call_field_94(players, 16403, UINTPTR_MAX, 0);
		}
		return 0;
	}
	default:
		return 0;
	}
}

//----- (00457010) --------------------------------------------------------
int sub_457010(nox_team_t* team, wchar2_t* name) {
	if (!dword_5d4594_1045684 || !team || !name) {
		return 0;
	}
	for (nox_gui_server_team_entry* entry =
			 (nox_gui_server_team_entry*)nox_common_list_getFirstSafe_425890(&nox_gui_server_players.teams);
		 entry; entry = (nox_gui_server_team_entry*)nox_common_list_getNextSafe_4258A0(&entry->list)) {
		if (entry->team_id == team->field_57) {
			nox_wcscpy(entry->name, name);
			break;
		}
	}
	char* settings = sub_4165B0();
	nox_window* teams = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1045684, 10502);
	if (!teams) {
		return 0;
	}
	int selected = (int)nox_window_call_field_94(teams, 16404, 0, 0);
	if (selected < 0) {
		return 0;
	}
	wchar2_t line[56];
	nox_wcscpy(line, name);
	if (nox_common_gameFlags_check_40A5C0(96) || (settings && (settings[52] & 0x60))) {
		if (team->field_57 < 3) {
			wchar2_t* suffix = nox_strman_loadString_40F1D0(team->field_57 == 1 ? "RedFlag" : "BlueFlag", 0,
				"C:\\NoxPost\\src\\client\\Gui\\ServOpts\\playrlst.c", team->field_57 == 1 ? 778 : 782);
			nox_wcscat(line, suffix);
		}
	}
	nox_window_call_field_94(teams, 16398, selected, 0);
	nox_window_call_field_94(teams, 16402, selected, 0);
	nox_window_call_field_94(teams, 16397, (uintptr_t)line, sub_457120(team));
	return (int)nox_window_call_field_94(teams, 16403, selected, 0);
}

//----- (00457230) --------------------------------------------------------
uintptr_t sub_457230(wchar2_t* name) {
	if (!dword_5d4594_1045684 || !name) {
		return 0;
	}
	nox_window* teams = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1045684, 10502);
	if (!teams) {
		return 0;
	}
	nox_team_t* team = sub_418A40(name);
	if (!team) {
		return 0;
	}
	nox_gui_server_team_entry* entry = calloc(1, sizeof(*entry));
	if (!entry) {
		return 0;
	}
	entry->team_id = team->field_57;
	entry->def_ind = team->def_ind;
	entry->color = sub_457120(team);
	sub_425770(&entry->list);
	nox_common_list_append_4258E0(&nox_gui_server_players.teams, &entry->list);
	nox_wcscpy(entry->name, name);

	wchar2_t line[56];
	nox_wcscpy(line, name);
	char* settings = sub_4165B0();
	if (nox_common_gameFlags_check_40A5C0(96) || (settings && (settings[52] & 0x60))) {
		if (team->field_57 < 3) {
			wchar2_t* suffix = nox_strman_loadString_40F1D0(team->field_57 == 1 ? "RedFlag" : "BlueFlag", 0,
				"C:\\NoxPost\\src\\client\\Gui\\ServOpts\\playrlst.c", team->field_57 == 1 ? 893 : 897);
			nox_wcscat(line, suffix);
		}
	}
	return nox_window_call_field_94(teams, 16397, (uintptr_t)line, entry->color);
}
