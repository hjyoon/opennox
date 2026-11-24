#include <math.h>

#include "compat.h"

#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "GAME3.h"
#include "GAME3_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME4_1.h"
#include "GAME4_2.h"
#include "GAME4_3.h"
#include "spell_accept_4fd400.h"
#include "GAME5.h"
#include "GAME5_2.h"
#include "client__system__parsecmd.h"
#include "common__log.h"
#include "common__net_list.h"
#include "common__random.h"
#include "common__system__team.h"
#include "server__mapgen__generate__populate.h"
#include "server__system__server.h"

#include "client__gui__chathelp.h"
#include "client__gui__conntype.h"
#include "client__gui__servopts__guiserv.h"
#include "client__gui__window.h"

#include "client__drawable__update__cloud.h"
#include "client__drawable__update__dball.h"

#include "MixPatch.h"
#include "client__video__draw_common.h"
#include "common/fs/nox_fs.h"
#include "common__crypt.h"
#include "common__magic__speltree.h"
#include "defs.h"
#include "mapgen_legacy_ptr.h"
#include "mapgen_theme_native.h"
#include "operators.h"
#include "player_incoming_quest_4ddf60.h"
#include "server__script__builtin.h"

extern uint32_t dword_5d4594_3835368;
extern uint32_t dword_5d4594_3835360;
extern uint32_t dword_5d4594_3835364;
extern uint32_t dword_5d4594_1556128;
extern uint32_t dword_5d4594_1556136;
extern uint32_t dword_5d4594_3835372;
extern nox_window* dword_5d4594_1523040;
extern uint32_t dword_5d4594_3835392;
extern uint32_t nox_server_sendMotd_108752;
extern uint32_t dword_5d4594_1548480;
extern nox_window* dword_5d4594_1523048;
extern nox_window* dword_5d4594_1523044;
extern nox_window* dword_5d4594_1523032;
extern uint32_t dword_5d4594_3835312;
extern uint32_t dword_5d4594_3835388;
extern uint32_t dword_5d4594_3835348;
extern uint32_t dword_5d4594_3835352;
extern uint32_t dword_5d4594_1550912;
extern nox_window* dword_5d4594_1523036;
extern uint32_t dword_5d4594_3835356;

extern uint32_t nox_server_connectionType_3596;
extern uint32_t dword_5d4594_1550916;
extern uint32_t dword_5d4594_2649712;
extern uint32_t dword_5d4594_3835396;
extern nox_window* dword_5d4594_1523024;
extern nox_window* dword_5d4594_1523028;
extern uint32_t dword_5d4594_1548476;

extern obj_5D4594_2650668_t** ptr_5D4594_2650668;
extern int ptr_5D4594_2650668_cap;

nox_list_item_t nox_common_maplist = {0};

//----- (004CDF80) --------------------------------------------------------
#if 0
int nox_xxx_updDrawDBall_4CDF80(int a1, int a2) {
	nox_xxx_updDrawAddRndSpark_4CDFA0(a2, (uint32_t*)3);
	return 1;
}

//----- (004CE0A0) --------------------------------------------------------
int sub_4CE0A0(int a1, int a2) {
	nox_xxx_updDrawAddRndSpark_4CDFA0(a2, (uint32_t*)1);
	return 1;
}
#endif

//----- (004CE1D0) --------------------------------------------------------
#if 0 // R228: original PE32 bodies; native registered entries are exported by Go.
int nox_xxx_updDrawCloud_4CE1D0(int a1, int a2) {
	if ((unsigned char)gameFrame() & 1) {
		sub_4CE200(a1, a2, 1, 75);
	}
	return 1;
}

//----- (004CE340) --------------------------------------------------------
int sub_4CE340(int a1, int a2) {
	*(uint16_t*)(a2 + 104) += *(unsigned char*)(a2 + 432);
	return 1;
}

//----- (004CE360) --------------------------------------------------------
int sub_4CE360(int a1, int a2) {
	if ((unsigned char)gameFrame() & 1) {
		sub_4CE200(a1, a2, 1, 35);
	}
	return 1;
}

#endif // R228 native registered entry

//----- (004CEBA0) --------------------------------------------------------
static void nox_gui_rulelist_clear_windows(void) {
	dword_5d4594_1523024 = NULL;
	dword_5d4594_1523028 = NULL;
	dword_5d4594_1523032 = NULL;
	dword_5d4594_1523036 = NULL;
	dword_5d4594_1523040 = NULL;
	dword_5d4594_1523044 = NULL;
	dword_5d4594_1523048 = NULL;
}

nox_window* sub_4CEBA0(nox_window* parent, char* map_name) {
	nox_gui_rulelist_clear_windows();
	dword_5d4594_1523024 = nox_new_window_from_file("rulelist.wnd", sub_4CF060);
	if (!dword_5d4594_1523024) {
		return NULL;
	}

	dword_5d4594_1523028 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1523024, 10170);
	dword_5d4594_1523032 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1523024, 10171);
	dword_5d4594_1523036 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1523024, 10172);
	dword_5d4594_1523040 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1523024, 10173);
	dword_5d4594_1523044 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1523024, 10174);
	dword_5d4594_1523048 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1523024, 10175);
	nox_window* background = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1523024, 10176);
	nox_window* thumb = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1523024, 10179);
	nox_window* up = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1523024, 10177);
	nox_window* down = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1523024, 10178);
	nox_scrollListBox_data* data = dword_5d4594_1523028 ? dword_5d4594_1523028->widget_data : NULL;
	if (!dword_5d4594_1523028 || !dword_5d4594_1523032 || !dword_5d4594_1523036 ||
		!dword_5d4594_1523040 || !dword_5d4594_1523044 || !dword_5d4594_1523048 || !background || !thumb ||
		!up || !down || !data || !thumb->field_100) {
		nox_window* root = dword_5d4594_1523024;
		nox_gui_rulelist_clear_windows();
		nox_xxx_windowDestroyMB_46C4E0(root);
		return NULL;
	}

	nox_xxx_wndSetDrawFn_46B340(background, sub_4CEED0);
	sub_46B120(dword_5d4594_1523024, parent);
	void* slider = nox_xxx_gLoadImg_42F970("UISlider");
	void* slider_lit = nox_xxx_gLoadImg_42F970("UISliderLit");
	thumb->field_100->width = 16;
	thumb->field_100->height = 10;
	sub_4B5700(thumb, NULL, NULL, slider, slider_lit, slider_lit);
	nox_xxx_wnd_46B280(thumb, dword_5d4594_1523028);
	nox_xxx_wnd_46B280(up, dword_5d4594_1523028);
	nox_xxx_wnd_46B280(down, dword_5d4594_1523028);
	data->field_9 = thumb;
	data->field_7 = up;
	data->field_8 = down;
	sub_4CED40(map_name);
	return dword_5d4594_1523024;
}

//----- (004CED40) --------------------------------------------------------
static void nox_gui_rulelist_add_file(char* map_name, struct _WIN32_FIND_DATAA* find_data) {
	if (!map_name || !find_data || !dword_5d4594_1523028 || (find_data->dwFileAttributes & FILE_ATTRIBUTE_DIRECTORY)) {
		return;
	}
	size_t len = strlen(find_data->cFileName);
	if (len <= 4 || nox_strcmpi(find_data->cFileName + len - 4, ".rul")) {
		return;
	}
	find_data->cFileName[len - 4] = 0;
	if (!nox_strcmpi(map_name, find_data->cFileName) || !nox_strcmpi("user", find_data->cFileName)) {
		return;
	}
	wchar2_t title[100];
	nox_swprintf(title, L"%S", find_data->cFileName);
	nox_window_call_field_94(dword_5d4594_1523028, 16397, (uintptr_t)title, UINTPTR_MAX);
}

void* sub_4CED40(char* a1) {
	if (!a1 || !dword_5d4594_1523028) {
		return (void*)(intptr_t)-1;
	}
	nox_window_call_field_94(dword_5d4594_1523028, 16399, 0, 0);
	char file_name[1024];
	nox_sprintf(file_name, "maps\\%s\\*.rul", a1);
	struct _WIN32_FIND_DATAA find_data;
	HANDLE handle = FindFirstFileA(file_name, &find_data);
	if (handle == (HANDLE)-1) {
		return handle;
	}
	nox_gui_rulelist_add_file(a1, &find_data);
	while (FindNextFileA(handle, &find_data)) {
		nox_gui_rulelist_add_file(a1, &find_data);
	}
	return (void*)(uintptr_t)FindClose(handle);
}

//----- (004CEED0) --------------------------------------------------------
static void nox_gui_rulelist_set_enabled(nox_window* win, int enabled) {
	if (win && (!!(win->flags & 8) != !!enabled)) {
		nox_xxx_wnd_46ABB0(win, enabled);
	}
}

int sub_4CEED0(nox_window* win, nox_window_data* draw) {
	if (!win || !draw || !dword_5d4594_1523028 || !dword_5d4594_1523032 || !dword_5d4594_1523036 ||
		!dword_5d4594_1523040 || !dword_5d4594_1523044 || !dword_5d4594_1523048) {
		return 1;
	}
	unsigned int x_left;
	unsigned int y_top;
	nox_client_wndGetPosition_46AA60(win, &x_left, &y_top);
	if ((int8_t)win->flags >= 0) {
		if (draw->bg_color != 0x80000000) {
			nox_client_drawRectFilledOpaque_49CE30(x_left, y_top, (int)win->width, (int)win->height);
		}
	} else if (draw->bg_image) {
		nox_client_drawImageAt_47D2C0(draw->bg_image, x_left, y_top);
	}
	int selection = (int)(intptr_t)nox_window_call_field_94(dword_5d4594_1523028, 16404, 0, 0);
	int has_selection = selection >= 0;
	nox_gui_rulelist_set_enabled(dword_5d4594_1523040, has_selection);
	nox_gui_rulelist_set_enabled(dword_5d4594_1523044,
		has_selection && !nox_common_gameFlags_check_40A5C0(49152));
	nox_gui_rulelist_set_enabled(dword_5d4594_1523048, has_selection);

	wchar2_t* text = (wchar2_t*)nox_window_call_field_94(dword_5d4594_1523032, 16413, 0, 0);
	nox_window* focus = nox_xxx_wndGetFocus_46B4F0();
	if (focus && focus->id == 10171) {
		nox_gui_rulelist_set_enabled(dword_5d4594_1523036, text && *text);
	}
	return 1;
}

//----- (004CF060) --------------------------------------------------------
int sub_4CF060(nox_window* win, unsigned int event, nox_window* control, uintptr_t event_arg) {
	(void)win;
	if (event > 0x4007) {
		if (event == 16400 && control && (uintptr_t)control != 1) {
			nox_xxx_wndGetID_46B0A0(control);
		}
		return 1;
	}
	if (event == 16387) {
		if (!dword_5d4594_1523024 || (uintptr_t)control == 1) {
			return 0;
		}
		nox_window* edit = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1523024, (int)event_arg);
		if (!edit) {
			return 0;
		}
		const wchar2_t* text = (const wchar2_t*)nox_window_call_field_94(edit, 16413, 0, 0);
		if ((int)event_arg == 10171 && text && *text && dword_5d4594_1523036) {
			wchar2_t current_map[64];
			nox_swprintf(current_map, L"%S", sub_4165B0());
			if (!_nox_wcsicmp(text, current_map) || !_nox_wcsicmp(text, L"user")) {
				nox_gui_rulelist_set_enabled(dword_5d4594_1523036, 0);
			}
		}
		return 1;
	}
	if (event != 16391 || !control || (uintptr_t)control == 1 || !dword_5d4594_1523028 ||
		!dword_5d4594_1523032) {
		return 1;
	}

	int id = nox_xxx_wndGetID_46B0A0(control);
	nox_xxx_clientPlaySoundSpecial_452D80(766, 100);
	char file_name[1024];
	char* settings;
	int selection;
	const wchar2_t* title;
	switch (id) {
	case 10172: {
		sub_416580();
		title = (const wchar2_t*)nox_window_call_field_94(dword_5d4594_1523032, 16413, 0, 0);
		settings = sub_4165B0();
		if (!title || !*title || !settings) {
			return 1;
		}
		nox_sprintf(file_name, "%S%s", title, getMemAt(0x587000, 191640));
		sub_459AA0(settings);
		sub_57AAA0(file_name, settings, NULL);
		nox_scrollListBox_data* data = dword_5d4594_1523028->widget_data;
		int count = data ? data->field_11_0 : 0;
		for (int i = 0; i < count; ++i) {
			const wchar2_t* item =
				(const wchar2_t*)nox_window_call_field_94(dword_5d4594_1523028, 16406, i, 0);
			if (item && !_nox_wcsicmp(title, item)) {
				nox_window_call_field_94(dword_5d4594_1523032, 16414,
					(uintptr_t)getMemAt(0x5D4594, 1523052), 0);
				return 1;
			}
		}
		nox_window_call_field_94(dword_5d4594_1523028, 16397, (uintptr_t)title, UINTPTR_MAX);
		nox_window_call_field_94(dword_5d4594_1523032, 16414, (uintptr_t)getMemAt(0x5D4594, 1523056), 0);
		nox_gui_rulelist_set_enabled(dword_5d4594_1523036, 0);
		return 1;
	}
	case 10173:
		sub_416580();
		selection = (int)(intptr_t)nox_window_call_field_94(dword_5d4594_1523028, 16404, 0, 0);
		if (selection < 0) {
			return 1;
		}
		title = (const wchar2_t*)nox_window_call_field_94(dword_5d4594_1523028, 16406, selection, 0);
		settings = sub_4165B0();
		if (!title || !settings) {
			return 1;
		}
		nox_sprintf(file_name, "%S%s", title, getMemAt(0x587000, 191592));
		sub_459AA0(settings);
		sub_57AAA0(file_name, settings, NULL);
		nox_window_call_field_94(dword_5d4594_1523028, 16403, UINTPTR_MAX, 0);
		return 1;
	case 10174:
		sub_416580();
		settings = sub_4165B0();
		selection = (int)(intptr_t)nox_window_call_field_94(dword_5d4594_1523028, 16404, 0, 0);
		if (!settings || selection < 0) {
			return 1;
		}
		title = (const wchar2_t*)nox_window_call_field_94(dword_5d4594_1523028, 16406, selection, 0);
		if (!title) {
			return 1;
		}
		nox_sprintf(file_name, "%S%s", title, getMemAt(0x587000, 191608));
		sub_57A1E0((int*)settings, file_name, NULL, 7, *((uint16_t*)settings + 26));
		sub_453F70(settings + 24);
		sub_4535E0((int*)settings + 11);
		sub_4535F0(((uint32_t*)settings)[12]);
		sub_459880(settings);
		nox_window_call_field_94(dword_5d4594_1523028, 16403, UINTPTR_MAX, 0);
		sub_459D50(1);
		return 1;
	case 10175:
		settings = sub_4165B0();
		selection = (int)(intptr_t)nox_window_call_field_94(dword_5d4594_1523028, 16404, 0, 0);
		if (!settings || selection < 0) {
			return 1;
		}
		title = (const wchar2_t*)nox_window_call_field_94(dword_5d4594_1523028, 16406, selection, 0);
		if (!title) {
			return 1;
		}
		nox_sprintf(file_name, "%S%s", title, getMemAt(0x587000, 191624));
		sub_57A9F0(settings, file_name);
		nox_window_call_field_94(dword_5d4594_1523028, 16398, selection, 0);
		return 1;
	default:
		return 1;
	}
}

//----- (004CFFA0) --------------------------------------------------------

//----- (004CFFC0) --------------------------------------------------------

//----- (004D0010) --------------------------------------------------------

// GAME.EXE 004D0550 builds map rule paths in Go.

//----- (004D0D50) --------------------------------------------------------
#if 0 // Preserved original for the network table oracle.
int sub_4D0D50(int a1) {
	int result;        // eax
	unsigned char* v2; // ecx

	result = 0;
	v2 = getMemAt(0x587000, 191836);
	while (!(a1 & *(uint32_t*)v2)) {
		v2 += 8;
		++result;
		if (result >= 6) {
			return 0;
		}
	}
	return result;
}
#endif

//----- (004D0D70) --------------------------------------------------------

//----- (004D0F60) --------------------------------------------------------

// GAME.EXE stores this remote-console authorization list in a 12-byte PE32
// sentinel, immediately followed by its initialization flag. Use a native
// sidecar so 64-bit links cannot overwrite that flag or truncate PlayerInfo.
typedef struct nox_remote_admin_entry_t {
	nox_list_item_t list;
	nox_playerInfo* player;
} nox_remote_admin_entry_t;

_Static_assert(offsetof(nox_remote_admin_entry_t, player) == (sizeof(void*) == 4 ? 12 : 24),
	"wrong native offset of remote-admin player");
_Static_assert(sizeof(nox_remote_admin_entry_t) == (sizeof(void*) == 4 ? 16 : 32),
	"wrong native size of remote-admin entry");

static nox_list_item_t nox_remote_admin_list_4D11A0;
static bool nox_remote_admin_initialized;

static void nox_remote_admin_ensure_4D11A0(void) {
	if (nox_remote_admin_initialized) {
		return;
	}
	nox_common_list_clear_425760(&nox_remote_admin_list_4D11A0);
	nox_remote_admin_initialized = true;
}

static nox_remote_admin_entry_t* nox_remote_admin_first_4D11A0(void) {
	nox_remote_admin_ensure_4D11A0();
	return (nox_remote_admin_entry_t*)nox_common_list_getFirstSafe_425890(&nox_remote_admin_list_4D11A0);
}

static nox_remote_admin_entry_t* nox_remote_admin_next_4D11A0(nox_remote_admin_entry_t* entry) {
	return (nox_remote_admin_entry_t*)nox_common_list_getNextSafe_4258A0(&entry->list);
}

//----- (004D11A0) --------------------------------------------------------
void sub_4D11A0() {
	nox_remote_admin_ensure_4D11A0();
	*getMemU32Ptr(0x5D4594, 1548504) = 1;
}

//----- (004D11D0) --------------------------------------------------------
void sub_4D11D0() {
	nox_remote_admin_entry_t* result; // eax
	nox_remote_admin_entry_t* v1;     // esi
	nox_remote_admin_entry_t* v2;     // edi

	result = nox_remote_admin_first_4D11A0();
	v1 = result;
	if (result) {
		do {
			v2 = nox_remote_admin_next_4D11A0(v1);
			nox_common_list_remove_425920(&v1->list);
			free(v1);
			v1 = v2;
		} while (v2);
	}
	nox_common_list_clear_425760(&nox_remote_admin_list_4D11A0);
}

//----- (004D1210) --------------------------------------------------------
void sub_4D1210(int a1) {
	nox_playerInfo* v2;           // esi
	nox_remote_admin_entry_t* v3; // eax

	if (sub_4D12A0(a1)) {
		return;
	}
	v2 = nox_common_playerInfoFromNum_417090(a1);
	if (!v2) {
		return;
	}
	v3 = calloc(1, sizeof(*v3));
	if (!v3) {
		return;
	}
	sub_425770(&v3->list);
	v3->player = v2;
	nox_common_list_append_4258E0(&nox_remote_admin_list_4D11A0, &v3->list);
}

//----- (004D1250) --------------------------------------------------------
int* sub_4D1250(int a1) {
	nox_remote_admin_entry_t* result; // eax
	nox_remote_admin_entry_t* v2;     // esi

	result = nox_remote_admin_first_4D11A0();
	v2 = result;
	if (result) {
		while (v2->player->playerInd != a1) {
			result = nox_remote_admin_next_4D11A0(v2);
			v2 = result;
			if (!result) {
				return 0;
			}
		}
		nox_common_list_remove_425920(&v2->list);
		free(v2);
	}
	return (int*)result;
}

//----- (004D12A0) --------------------------------------------------------
int sub_4D12A0(int a1) {
	nox_remote_admin_entry_t* v1; // eax

	v1 = nox_remote_admin_first_4D11A0();
	if (!v1) {
		return 0;
	}
	while (v1->player->playerInd != a1) {
		v1 = nox_remote_admin_next_4D11A0(v1);
		if (!v1) {
			return 0;
		}
	}
	return 1;
}

//----- (004D6000) --------------------------------------------------------
#if 0 // Restored by server/exit_collide_4e9090_server.go at native width.
int sub_4D6000(nox_object_t* a1p) {
	int a1 = a1p;
	int result; // eax
	int v2;     // esi

	result = 0;
	if (a1) {
		v2 = *(uint32_t*)(a1 + 748);
		*(uint32_t*)(*(uint32_t*)(v2 + 276) + 4652) = 0;
		*(uint32_t*)(*(uint32_t*)(v2 + 276) + 4656) = 0;
		*(uint32_t*)(*(uint32_t*)(v2 + 276) + 4660) = 0;
		*(uint32_t*)(*(uint32_t*)(v2 + 276) + 4664) = 0;
		*(uint32_t*)(*(uint32_t*)(v2 + 276) + 4668) = 0;
		*(uint32_t*)(*(uint32_t*)(v2 + 276) + 4672) = 0;
		*(uint32_t*)(*(uint32_t*)(v2 + 276) + 4676) = 0;
		*(uint32_t*)(*(uint32_t*)(v2 + 276) + 4680) = 0;
		*(uint32_t*)(*(uint32_t*)(v2 + 276) + 4684) = 0;
		*(uint32_t*)(*(uint32_t*)(v2 + 276) + 4688) = nox_game_getQuestStage_4E3CC0();
		result = *(uint32_t*)(v2 + 276);
		*(uint32_t*)(result + 4692) = 63;
	}
	return result;
}
#endif



//----- (004D60B0) --------------------------------------------------------
#if 0 // Restored by server/exit_collide_4e9090_server.go at native width.
int sub_4D60B0() {
	int result; // eax
	int i;      // esi

	result = nox_xxx_getFirstPlayerUnit_4DA7C0();
	for (i = result; result; i = result) {
		sub_4D6000(i);
		result = nox_xxx_getNextPlayerUnit_4DA7F0(i);
	}
	return result;
}
#endif

//----- (004D60E0) --------------------------------------------------------
#if 0 // Restored by pointer_chain_native_entry.go.
uint32_t* sub_4D60E0(int a1) {
	uint32_t* result; // eax
	int v2;           // ecx

	result = (uint32_t*)a1;
	if ((*(uint8_t*)(a1 + 16) & 0x20) != 32) {
		v2 = *(uint32_t*)(a1 + 748);
		result = *(uint32_t**)(v2 + 276);
		if (result[1198] == 1) {
			++result[1163];
			result = *(uint32_t**)(v2 + 276);
			result[1173] |= 1u;
		}
	}
	return result;
}
#endif

//----- (004D6130) --------------------------------------------------------
#if 0
uintptr_t sub_4D6130(nox_object_t* unit) {
	int a1 = (int)(uintptr_t)unit;
	uintptr_t result; // eax
	int v2;     // eax

	result = a1;
	if (a1) {
		if ((*(uint8_t*)(a1 + 16) & 0x20) != 32) {
			v2 = *(uint32_t*)(a1 + 748);
			++*(uint32_t*)(*(uint32_t*)(v2 + 276) + 4660);
			result = *(uint32_t*)(v2 + 276);
			*(uint32_t*)(result + 4692) |= 2u;
		}
	}
	return result;
}
#endif

//----- (004D6170) --------------------------------------------------------
#if 0 // Restored by pointer_chain_native_entry.go.
int sub_4D6170(int a1) {
	int result; // eax
	int v2;     // eax

	result = a1;
	if (a1) {
		if ((*(uint8_t*)(a1 + 16) & 0x20) != 32) {
			v2 = *(uint32_t*)(a1 + 748);
			++*(uint32_t*)(*(uint32_t*)(v2 + 276) + 4664);
			result = *(uint32_t*)(v2 + 276);
			*(uint32_t*)(result + 4692) |= 4u;
		}
	}
	return result;
}
#endif

//----- (004D61B0) --------------------------------------------------------
#if 0 // Restored by server/generator_credit_raw_4d61b0.go at native width.
void sub_4D61B0(int a1) {
	int v2; // eax
	if (a1) {
		if ((*(uint8_t*)(a1 + 16) & 0x20) != 32) {
			v2 = *(uint32_t*)(a1 + 748);
			++*(uint32_t*)(*(uint32_t*)(v2 + 276) + 4668);
			int result = *(uint32_t*)(v2 + 276);
			*(uint32_t*)(result + 4692) |= 8u;
		}
	}
}
#endif

//----- (004D61F0) --------------------------------------------------------
nox_playerInfo* sub_4D61F0(nox_object_t* player_unit) {
	if (!player_unit || (player_unit->obj_flags & 0x20) == 0x20) {
		return NULL;
	}
	nox_player_update_data_t* update = player_unit->data_update;
	nox_playerInfo* player = update->player;
	++player->field_4672;
	++player->field_4676;
	player->field_4692 |= 0x10u;
	return player;
}

//----- (004D6540) --------------------------------------------------------
#if 0 // Restored by server/quest_stats_4d6540_server.go at native width.
unsigned int sub_4D6540(int a1) {
	int v1;              // edi
	unsigned int* v2;    // eax
	unsigned int* v3;    // ebx
	unsigned int result; // eax
	unsigned int v5;     // ebp
	int i;               // edi
	unsigned int* v7;    // esi
	unsigned int v8;     // eax
	int v9;              // eax
	float v10;           // [esp+0h] [ebp-28h]
	unsigned int v11;    // [esp+14h] [ebp-14h]
	unsigned int v12;    // [esp+18h] [ebp-10h]
	float v13;           // [esp+20h] [ebp-8h]
	unsigned int v14;    // [esp+2Ch] [ebp+4h]

	v1 = nox_xxx_player_4E3CE0();
	v2 = (unsigned int*)nox_common_playerInfoFromNum_417090(a1);
	v3 = v2;
	if (!v2 || !v2[1198]) {
		return 0;
	}
	if (v1 == 1) {
		result = sub_4D66E0(v2[1167], v2[1168], v2[1166], v2[1172]);
	} else {
		v5 = 0;
		v11 = 0;
		v14 = 0;
		v12 = 1;
		v13 = 1.0;
		for (i = nox_xxx_getFirstPlayerUnit_4DA7C0(); i; i = nox_xxx_getNextPlayerUnit_4DA7F0(i)) {
			v7 = *(unsigned int**)(*(uint32_t*)(i + 748) + 276);
			if (v3[1198]) {
				v8 = sub_4D66E0(v7[1167], v7[1168], 0, v7[1172]);
				if (v8 > v5) {
					v5 = v8;
				}
				v11 += v7[1167];
				v12 = v7[1172];
				v14 += v7[1168];
			}
		}
		v9 = sub_4D66E0(v11, v14, 0, v12);
		if (v5 > 0) {
			v13 = (double)(unsigned int)v9 / (double)v5;
		}
		v10 = (double)(unsigned int)sub_4D66E0(v3[1167], v3[1168], v3[1166], v3[1172]) * v13;
		result = nox_float2int(v10);
	}
	if (result > 0x3B9AC9FF) {
		result = 999999999;
	}
	return result;
}

//----- (004D66E0) --------------------------------------------------------
int sub_4D66E0(unsigned int a1, unsigned int a2, unsigned int a3, unsigned int a4) {
	float v5; // [esp+4h] [ebp-10h]

	v5 = nox_double2float(pow((double)a4, *(long double*)getMemAt(0x581450, 10088))) *
		 ((double)a1 * 10.0 + (double)a2 * 35.0 + (double)a3 * 0.1);
	return nox_float2int(v5);
}

//----- (004D6770) --------------------------------------------------------
int sub_4D6770(int a1) {
	char* v1;    // ebp
	int v2;      // ebx
	int v3;      // edi
	char* v4;    // esi
	int v5;      // eax
	char v7[90]; // [esp+Ch] [ebp-5Ch]

	v1 = nox_common_playerInfoFromNum_417090(a1);
	memset(v7, 0, 0x58u);
	*(uint16_t*)&v7[88] = 0;
	v7[0] = -16;
	v7[1] = 12;
	*(uint16_t*)&v7[2] = sub_4D7300();
	v2 = 0;
	v3 = nox_xxx_getFirstPlayerUnit_4DA7C0();
	if (v3) {
		v4 = &v7[8];
		do {
			v5 = *(uint32_t*)(v3 + 748);
			if (*(uint32_t*)(*(uint32_t*)(v5 + 276) + 4792) == 1 && v2 < 6) {
				*((uint16_t*)v4 - 1) = *(uint16_t*)(v3 + 36);
				*(uint16_t*)v4 = *(uint16_t*)(*(uint32_t*)(v5 + 276) + 4668);
				*((uint16_t*)v4 + 3) = *(uint16_t*)(*(uint32_t*)(v5 + 276) + 4664);
				*((uint16_t*)v4 + 1) = *(uint16_t*)(*(uint32_t*)(v5 + 276) + 4672);
				*((uint16_t*)v4 + 2) = *(uint16_t*)(*(uint32_t*)(v5 + 276) + 4680);
				*((uint32_t*)v4 + 2) = sub_4D6540(*(unsigned char*)(*(uint32_t*)(v5 + 276) + 2064));
				++v2;
				v4 += 14;
				*(uint16_t*)&v7[4] = *((uint16_t*)v1 + 2344);
			}
			v3 = nox_xxx_getNextPlayerUnit_4DA7F0(v3);
		} while (v3);
	}
	return nox_xxx_netSendPacket1_4E5390(a1, v7, 90, 0, 1);
}
#endif





//----- (004D70B0) --------------------------------------------------------
char* sub_4D70B0() {
	char* result; // eax

	result = *(char**)getMemAt(0x5D4594, 1556152);
	if (!*getMemU32Ptr(0x5D4594, 1556152)) {
		result = sub_4169F0();
	}
	return result;
}

//----- (004D70C0) --------------------------------------------------------

//----- (004D7100) --------------------------------------------------------

//----- (004D71F0) --------------------------------------------------------
void sub_4D72B0(int a1);

//----- (004D7480) --------------------------------------------------------

//----- (004D7520) --------------------------------------------------------

//----- (004D75E0) --------------------------------------------------------

//----- (004D7600) --------------------------------------------------------

//----- (004D79C0) --------------------------------------------------------
#if 0 // Restored by server/quest_player_state_4d79c0_server.go at native width.
int sub_4D79C0(nox_object_t* a1p) {
	int a1 = a1p;
	int v1;     // esi
	int result; // eax
	int v3;     // ecx

	v1 = *(uint32_t*)(a1 + 748);
	sub_4D9D20(255, a1);
	sub_4D6000(a1);
	for (result = nox_xxx_getFirstPlayerUnit_4DA7C0(); result; result = nox_xxx_getNextPlayerUnit_4DA7F0(result)) {
		v3 = *(uint32_t*)(result + 748);
		*(uint8_t*)(*(unsigned char*)(*(uint32_t*)(v1 + 276) + 2064) + v3 + 452) = 0;
		*(uint32_t*)(v3 + 4 * *(unsigned char*)(*(uint32_t*)(v1 + 276) + 2064) + 324) = 0;
		*(uint8_t*)(*(unsigned char*)(*(uint32_t*)(v1 + 276) + 2064) + v3 + 484) = 0;
		*(uint8_t*)(*(unsigned char*)(*(uint32_t*)(v1 + 276) + 2064) + v3 + 516) = 0;
	}
	return result;
}

//----- (004D7A60) --------------------------------------------------------
int sub_4D7A60(int a1) {
	int result; // eax

	result = a1;
	*getMemU32Ptr(0x5D4594, 1556172 + 4 * a1) = gameFrame();
	return result;
}

//----- (004D7A80) --------------------------------------------------------
int sub_4D7A80() {
	unsigned char* v0; // ebp
	int v1;            // edi
	int v2;            // esi
	char* v3;          // eax
	int v4;            // eax
	int v5;            // edi
	int v6;            // ecx
	int result;        // eax

	v0 = getMemAt(0x5D4594, 1556172);
	v1 = 324 - (uint32_t)getMemAt(0x5D4594, 1556172);
	v2 = 484;
	do {
		v3 = nox_common_playerInfoFromNum_417090(v2 - 484);
		if (v3 && *((uint32_t*)v3 + 523) && *((uint32_t*)v3 + 514) && *((uint32_t*)v3 + 1198) == 1) {
			*(uint32_t*)v0 = 0;
		} else if (*(uint32_t*)v0 && gameFrame() - *(uint32_t*)v0 > (unsigned int)(30 * gameFPS())) {
			v4 = nox_xxx_getFirstPlayerUnit_4DA7C0();
			if (v4) {
				v5 = (int)&v0[v1];
				do {
					v6 = *(uint32_t*)(v4 + 748);
					*(uint8_t*)(v2 + v6 - 32) = 0;
					*(uint32_t*)(v5 + v6) = 0;
					*(uint8_t*)(v2 + v6) = 0;
					*(uint8_t*)(v2 + v6 + 32) = 0;
					v4 = nox_xxx_getNextPlayerUnit_4DA7F0(v4);
				} while (v4);
				v1 = 324 - (uint32_t)getMemAt(0x5D4594, 1556172);
			}
			*(uint32_t*)v0 = 0;
		}
		++v2;
		v0 += 4;
		result = v2 - 484;
	} while (v2 - 484 < 32);
	return result;
}

//----- (004D7B40) --------------------------------------------------------
int sub_4D7B40() {
	int result; // eax

	result = 0;
	memset(getMemAt(0x5D4594, 1556172), 0, 0x80u);
	return result;
}

//----- (004D7BE0) --------------------------------------------------------
int nox_xxx_netSendInterestingId_4D7BE0(int a1) {
	int result; // eax
	int i;      // esi
	char v3[7]; // [esp+4h] [ebp-8h]

	v3[0] = -46;
	*(uint16_t*)&v3[1] = nox_xxx_netGetUnitCodeServ_578AC0((uint32_t*)a1);
	*(uint16_t*)&v3[3] = *(uint16_t*)(a1 + 4);
	v3[5] = 2;
	v3[6] = 2;
	result = nox_xxx_getFirstPlayerUnit_4DA7C0();
	for (i = result; result; i = result) {
		nox_xxx_netSendPacket0_4E5420(*(unsigned char*)(*(uint32_t*)(*(uint32_t*)(i + 748) + 276) + 2064), v3, 7, 0, 1);
		result = nox_xxx_getNextPlayerUnit_4DA7F0(i);
	}
	return result;
}

//----- (004D7E50) --------------------------------------------------------
int sub_4D7E50(nox_object_t* a1p) {
	int a1 = a1p;
	int result;   // eax
	uint32_t* v2; // esi

	result = a1;
	if (*(uint8_t*)(a1 + 8) & 4) {
		v2 = *(uint32_t**)(a1 + 748);
		v2[62] = 0;
		v2[63] = 0;
		v2[64] = gameFrame();
		if (v2[65]) {
			result = nox_xxx_netSendInterestingId_4D7BE0(a1);
		}
		v2[65] = 0;
	}
	return result;
}

//----- (004D7EA0) --------------------------------------------------------
int sub_4D7EA0() {
	int result; // eax
	int i;      // esi

	result = nox_xxx_getFirstPlayerUnit_4DA7C0();
	for (i = result; result; i = result) {
		sub_4D7E50(i);
		result = nox_xxx_getNextPlayerUnit_4DA7F0(i);
	}
	return result;
}
#endif

// GAME.EXE 004D8620 is restored by current_hp_report_4d8620.go. Keep the raw
// ABI32 body only as provenance; active callers use the typed CGo export in
// unit_adjust_hp_4ee460.h.
#if 0
//----- (004D8620) --------------------------------------------------------
int nox_xxx_netReportUnitCurrentHP_4D8620(int a1, uint32_t* a2) {
	uint32_t* v2;    // esi
	int result;      // eax
	unsigned int v4; // ecx
	uint32_t* v5;    // [esp-4h] [ebp-8h]

	v2 = a2;
	result = a2[139];
	if (result) {
		v5 = a2;
		LOBYTE(a2) = 65;
		*(uint16_t*)((char*)&a2 + 1) = nox_xxx_netGetUnitCodeServ_578AC0(v5);
		LOWORD(v4) = *(uint16_t*)v2[139];
		HIBYTE(a2) = v4 >> 1;
		result = nox_xxx_netSendPacket1_4E5390(a1, &a2, 4, 0, 1);
	}
	return result;
}
// 4D8656: variable 'v4' is possibly undefined
#endif

//----- (004D8670) --------------------------------------------------------
#if 0 // PE32 reference. The native Go senders use typed Object and Player records.
int nox_xxx_netSendTeam_4D8670(int a1, uint32_t* a2) {
	int result;         // eax
	short v3;           // ax
	unsigned short* v4; // ecx
	char v5[5];         // [esp+4h] [ebp-8h]

	result = a2[139];
	if (result) {
		v5[0] = -60;
		v5[1] = 12;
		v3 = nox_xxx_netGetUnitCodeServ_578AC0(a2);
		v4 = (unsigned short*)a2[139];
		*(uint16_t*)&v5[2] = v3;
		v5[4] = 100 * *v4 / v4[2];
		result = nox_xxx_netSendPacket1_4E5390(a1, v5, 5, 0, 1);
	}
	return result;
}

//----- (004D86E0) --------------------------------------------------------
char* nox_xxx_netSendPlrHealthToTeam_4D86E0(int a1) {
	int v1;       // edi
	char* result; // eax
	char* v3;     // esi

	v1 = a1;
	result = nox_common_playerInfoFromNum_417090(a1);
	v3 = result;
	if (result) {
		result = (char*)*((uint32_t*)result + 514);
		if (result) {
			if (*((uint32_t*)result + 139)) {
				LOBYTE(a1) = 67;
				*(uint16_t*)((char*)&a1 + 1) = **(uint16_t**)(*((uint32_t*)v3 + 514) + 556);
				nox_xxx_netSendPacket1_4E5390(v1, &a1, 3, 0, 1);
				result = (char*)nox_common_gameFlags_check_40A5C0(4096);
				if (result) {
					result = (char*)nox_xxx_netSendTeam_4D8670(v1 | 0x80, *((uint32_t**)v3 + 514));
				}
			}
		}
	}
	return result;
}
#endif

// GAME.EXE 004D8B90 through 004D8C80 use Go winner report entries.

// GAME.EXE 004D8CD0 and 004D8D40 use native Go entries.

//----- (004D8E90) --------------------------------------------------------
#if 0 // Restored by pointer_chain_native_entry.go.
int nox_xxx_changeScore_4D8E90(int a1, int a2) {
	int result; // eax

	result = a1;
	if (*(uint8_t*)(a1 + 8) & 4) {
		result = *(uint32_t*)(*(uint32_t*)(a1 + 748) + 276);
		*(uint32_t*)(result + 2136) += a2;
	}
	return result;
}
#endif

//----- (004D8EC0) --------------------------------------------------------
#if 0 // Restored by pointer_chain_native_entry.go.
int nox_xxx_playerSubLessons_4D8EC0(int a1, int a2) {
	int result; // eax

	result = a1;
	if (*(uint8_t*)(a1 + 8) & 4) {
		result = *(uint32_t*)(*(uint32_t*)(a1 + 748) + 276);
		*(uint32_t*)(result + 2136) -= a2;
	}
	return result;
}
#endif

//----- (004D9020) --------------------------------------------------------
#if 0 // Retained PE32 body; public native entry and ordinary packet builder are implemented in Go.
int nox_xxx_netReportUnitHeight_4D9020(int a1, nox_object_t* a2p) {
	int a2 = a2p;
	int v2;        // esi
	short v3;      // ax
	double v4;     // st7
	long long v5;  // rax
	double v6;     // st7
	long long v7;  // rax
	double v8;     // st7
	int result;    // eax
	short v10;     // ax
	double v11;    // st7
	short v12;     // ax
	double v13;    // st7
	uint32_t* v14; // [esp-4h] [ebp-10h]
	char v15[6];   // [esp+4h] [ebp-8h]

	v2 = a2;
	if (*(uint8_t*)(a2 + 20) & 0x20) {
		v3 = *(uint16_t*)(a2 + 36);
		v15[0] = -97;
		v4 = *(float*)(a2 + 104);
		*(uint16_t*)&v15[1] = v3;
		v5 = (long long)v4;
		v6 = *(float*)(a2 + 108);
		v15[3] = v5;
		v7 = (long long)v6;
		v8 = *(float*)(a2 + 116);
		v15[4] = v7;
		v15[5] = (long long)v8;
		result = nox_netlist_addToMsgListCli_40EBC0(a1, 1, (unsigned char*)v15, 6);
	} else {
		v14 = (uint32_t*)a2;
		if (*(float*)(a2 + 104) < 0.0) {
			LOBYTE(a2) = 95;
			v12 = nox_xxx_netGetUnitCodeServ_578AC0(v14);
			v13 = *(float*)(v2 + 104);
			*(uint16_t*)((char*)&a2 + 1) = v12;
			v11 = -v13;
		} else {
			LOBYTE(a2) = 94;
			v10 = nox_xxx_netGetUnitCodeServ_578AC0(v14);
			v11 = *(float*)(v2 + 104);
			*(uint16_t*)((char*)&a2 + 1) = v10;
		}
		HIBYTE(a2) = (long long)v11;
		result = nox_netlist_addToMsgListCli_40EBC0(a1, 1, &a2, 4);
	}
	return result;
}

#endif

//----- (004D9200) --------------------------------------------------------
#if 0 // PE32 reference. The packet is sent by the native-width Go entry.
int nox_xxx_netFxShield_0_4D9200(int a1, int a2) {
	int v4; // [esp+0h] [ebp-4h]

	LOBYTE(v4) = 109;
	*(uint16_t*)((char*)&v4 + 1) = *(uint16_t*)(a2 + 36);
	if (nox_common_gameFlags_check_40A5C0(0x8000000)) {
		BYTE2(v4) |= 0x80u;
	}
	return nox_xxx_netSendPacket1_4E5390(a1, &v4, 3, 0, 1);
}
#endif

// GAME.EXE 004D92D0 uses a native Go entry.

//----- (004D9900) --------------------------------------------------------
#if 0 // PE32 reference. The native entry shares the Go player-object sender.
void nox_xxx_playerReportAnything_4D9900(int a1) {
	int v1;        // ebp
	int v2;        // esi
	uint16_t* v3;  // edi
	int v4;        // eax
	int v5;        // eax
	int i;         // edi
	char* v7;      // eax
	int j;         // edi
	int v9;        // ebx
	int v10;       // eax
	int k;         // edi
	int v12;       // ebx
	int v13;       // eax
	int v14;       // eax
	uint16_t* v15; // [esp+Ch] [ebp-4h]
	int v16;       // [esp+14h] [ebp+4h]

	v1 = a1;
	if (a1 && *(uint8_t*)(a1 + 8) & 4) {
		v2 = *(uint32_t*)(a1 + 748);
		v3 = *(uint16_t**)(a1 + 556);
		v15 = *(uint16_t**)(a1 + 556);
		v16 = *(int*)(v2 + 228);
		if (*(float*)(v2 + 232) != *(float*)&v16) {
			nox_xxx_netReportArmorVal_4D8A30(*(unsigned char*)(*(uint32_t*)(v2 + 276) + 2064), v16);
			*(uint32_t*)(v2 + 232) = *(uint32_t*)(v2 + 228);
		}
		v4 = *(uint32_t*)(v2 + 276);
		if (*(uint32_t*)(v4 + 2168) != *(uint32_t*)(v4 + 2164)) {
			sub_4D8870(*(unsigned char*)(v4 + 2064), v1);
			*(uint32_t*)(*(uint32_t*)(v2 + 276) + 2168) = *(uint32_t*)(*(uint32_t*)(v2 + 276) + 2164);
		}
		v5 = *(uint32_t*)(v2 + 276);
		if (*(uint8_t*)(v5 + 2172) != *(uint8_t*)(v1 + 440)) {
			sub_4D8840(*(unsigned char*)(v5 + 2064), v1);
			*(uint8_t*)(*(uint32_t*)(v2 + 276) + 2172) = *(uint8_t*)(v1 + 440);
		}
		if (nox_common_gameFlags_check_40A5C0(4096)) {
			for (i = 0; i < 32; ++i) {
				v7 = nox_common_playerInfoFromNum_417090(i);
				if (v7 && *((uint32_t*)v7 + 514) && *(uint32_t*)(v2 + 320) != *(unsigned char*)(i + v2 + 452)) {
					sub_4D9D60(i, (nox_object_t*)(uintptr_t)(uint32_t)v1);
					*(uint8_t*)(i + v2 + 452) = *(uint8_t*)(v2 + 320);
				}
			}
			for (j = 0; j < 32; ++j) {
				v9 = 0;
				if (!*getMemU32Ptr(0x5D4594, 1556324)) {
					*getMemU32Ptr(0x5D4594, 1556324) = nox_xxx_getNameId_4E3AA0("SilverKey");
				}
				if (nox_common_playerInfoFromNum_417090(j)) {
					v10 = nox_xxx_inventoryGetFirst_4E7980(v1);
					if (v10) {
						while (*(unsigned short*)(v10 + 4) != *getMemU32Ptr(0x5D4594, 1556324)) {
							v10 = nox_xxx_inventoryGetNext_4E7990(v10);
							if (!v10) {
								goto LABEL_25;
							}
						}
						v9 = 1;
					}
				LABEL_25:
					if (v9 != *(unsigned char*)(j + v2 + 484)) {
						sub_4D9DF0(j, v1, v9);
						*(uint8_t*)(j + v2 + 484) = v9;
					}
				}
			}
			for (k = 0; k < 32; ++k) {
				v12 = 0;
				if (!*getMemU32Ptr(0x5D4594, 1556328)) {
					*getMemU32Ptr(0x5D4594, 1556328) = nox_xxx_getNameId_4E3AA0("GoldKey");
				}
				if (nox_common_playerInfoFromNum_417090(k)) {
					v13 = nox_xxx_inventoryGetFirst_4E7980(v1);
					if (v13) {
						while (*(unsigned short*)(v13 + 4) != *getMemU32Ptr(0x5D4594, 1556328)) {
							v13 = nox_xxx_inventoryGetNext_4E7990(v13);
							if (!v13) {
								goto LABEL_37;
							}
						}
						v12 = 1;
					}
				LABEL_37:
					if (v12 != *(unsigned char*)(k + v2 + 516)) {
						sub_4D9E30(k, v1, v12);
						*(uint8_t*)(k + v2 + 516) = v12;
					}
				}
			}
			v3 = v15;
		}
		v14 = *(uint32_t*)(v2 + 276);
		if (*(uint8_t*)(v14 + 2184)) {
			nox_xxx_netReportTotalHealth_4D85C0(*(unsigned char*)(v14 + 2064), (uint32_t*)v1);
			nox_xxx_netReportTotalMana_4D88C0(*(unsigned char*)(*(uint32_t*)(v2 + 276) + 2064), v1);
			nox_xxx_servSendStats_4D8990(*(unsigned char*)(*(uint32_t*)(v2 + 276) + 2064), v1,
										 *(uint8_t*)(*(uint32_t*)(v2 + 276) + 3684));
			*(uint8_t*)(*(uint32_t*)(v2 + 276) + 2184) = 0;
		}
		if (0) {
			sub_528030(v1);
		} else {
			if (v3 && *v3 != *(uint16_t*)(v2 + 10)) {
				nox_xxx_netSendPlrHealthToTeam_4D86E0(*(unsigned char*)(*(uint32_t*)(v2 + 276) + 2064));
				*(uint16_t*)(v2 + 10) = *v3;
			}
			if (*(uint16_t*)(v2 + 4) != *(uint16_t*)(v2 + 6)) {
				nox_xxx_netReportMana_4D8930(*(unsigned char*)(*(uint32_t*)(v2 + 276) + 2064), v1);
				*(uint16_t*)(v2 + 6) = *(uint16_t*)(v2 + 4);
			}
		}
	}
}
#endif

//----- (004D9D20) --------------------------------------------------------
#if 0 // Restored by server/quest_player_state_4d79c0_server.go at native width.
int sub_4D9D20(int a1, nox_object_t* a2p) {
	int a2 = a2p;
	short v3;   // cx
	char v5[4]; // [esp+0h] [ebp-4h]

	v3 = *(uint16_t*)(a2 + 36);
	*(uint16_t*)v5 = 496;
	*(uint16_t*)&v5[2] = v3;
	return nox_xxx_netSendPacket1_4E5390(a1, v5, 4, 0, 1);
}
#endif



//----- (004D9D60) --------------------------------------------------------

//----- (004D9DF0) --------------------------------------------------------
#if 0 // PE32 references. Both key packets share the Go implementation.
int sub_4D9DF0(int a1, int a2, char a3) {
	char v4[5]; // [esp+0h] [ebp-8h]

	*(uint16_t*)&v4[3] = *(uint16_t*)(a2 + 36);
	v4[0] = -16;
	v4[1] = 22;
	v4[2] = a3;
	return nox_xxx_netSendPacket0_4E5420(a1, v4, 5, 0, 1);
}

//----- (004D9E30) --------------------------------------------------------
int sub_4D9E30(int a1, int a2, char a3) {
	char v4[5]; // [esp+0h] [ebp-8h]

	*(uint16_t*)&v4[3] = *(uint16_t*)(a2 + 36);
	v4[0] = -16;
	v4[1] = 23;
	v4[2] = a3;
	return nox_xxx_netSendPacket0_4E5420(a1, v4, 5, 0, 1);
}
#endif

//----- (004D9FD0) --------------------------------------------------------
int nox_xxx_printToAll_4D9FD0(char a1, wchar2_t* a2, ...) {
	char v2;         // al
	int v3;          // edi
	wchar2_t v6[516]; // [esp+Ch] [ebp-408h]
	va_list va;      // [esp+420h] [ebp+Ch]

	va_start(va, a2);
	nox_vswprintf(&v6[260], a2, va);
	LOBYTE(v6[0]) = -88; // MSG_TEXT_MESSAGE
	*(wchar2_t*)((char*)v6 + 1) = 0;
	HIBYTE(v6[1]) = a1;
	if (nox_xxx_cliCanTalkMB_4100F0((short*)&v6[260])) {
		v2 = HIBYTE(v6[1]) | 2;
	} else {
		v2 = HIBYTE(v6[1]) | 4;
	}
	HIBYTE(v6[1]) = v2;
	v6[2] = 0;
	v6[3] = 0;
	LOBYTE(v6[5]) = 0;
	v6[4] = (unsigned char)(nox_wcslen(&v6[260]) + 1);
	if (v6[1] & 0x400) {
		nox_wcscpy((wchar2_t*)((char*)&v6[5] + 1), &v6[260]);
		v3 = 2;
	} else {
		nox_sprintf((char*)&v6[5] + 1, "%S", &v6[260]);
		v3 = 1;
	}
	for (nox_object_t* unit = nox_xxx_getFirstPlayerUnit_4DA7C0(); unit;
		 unit = nox_xxx_getNextPlayerUnit_4DA7F0(unit)) {
		nox_player_update_data_t* update = unit->data_update;
		if (update && update->player) {
			nox_netlist_addToMsgListCli_40EBC0(update->player->playerInd, 1, (unsigned char*)v6,
										   v3 * LOBYTE(v6[4]) + 11);
		}
	}
	return 0;
}

//----- (004DDF60) --------------------------------------------------------
void nox_xxx_playerSendMOTD_4DD140(int a1);
int nox_xxx_netPlayerIncomingServ_4DDF60(int a1) {
	int v1;             // ebx
	nox_playerInfo* v2; // esi
	float v4;           // edi
	float v5;           // ebp
	const volatile nox_game_ball_status_t* v8; // eax
	nox_team_t* j;      // edi
	const volatile nox_team_flag_status_t* v10; // eax
	nox_object_t* v13;  // [esp+Ch] [ebp+4h]
	uint16_t game_ball_net_code;
	uint8_t game_ball_state;
	uint16_t team_flag_carrier_net_code;
	uint8_t team_flag_index;
	uint8_t team_flag_status;
	uint8_t team_flag_id;

	v1 = a1;
	v2 = nox_common_playerInfoFromNum_417090(a1);
	if (!v2) {
		abort();
	}
	if (nox_common_gameFlags_check_40A5C0(4096)) {
		if (a1 != 31) {
			if (v2->playerUnit) {
				nox_xxx_playerIncomingQuest_native_4DDF60(v2->playerUnit);
			}
		}
		sub_4D9CF0(a1);
		if (v2->playerUnit) {
			sub_4D6000(v2->playerUnit);
		}
	}
	sub_57B920(v2->netData16);
	v13 = v2->playerUnit;
	dword_5d4594_2649712 |= 1 << v1;
	v4 = v13->x;
	v5 = v13->y;
	nox_xxx_newPlayerSendAllPlayers_4DE300(v1);
	v2->field_4700 = 0;
	((void (*)(nox_object_t*, uint32_t))v13->func_init)(v13, 0);
	v2->field_3676 = 3;
	if (!nox_common_gameFlags_check_40A5C0(512)) {
		v2->pos_x_3632 = v4;
		v2->pos_y_3636 = v5;
	}
	if (nox_server_sendMotd_108752 && nox_common_gameFlags_check_40A5C0(0x2000) &&
		!nox_common_gameFlags_check_40A5C0(4096)) {
		nox_xxx_playerSendMOTD_4DD140(v1);
	}
	for (nox_playerInfo* i = nox_common_playerInfoGetFirst_416EA0(); i; i = nox_common_playerInfoGetNext_416EE0(i)) {
		if (i->playerUnit) {
			if (i != v2) {
				nox_xxx_netMarkMinimapObject_417190(v1, i->playerUnit, 1);
				nox_xxx_netMarkMinimapObject_417190(i->playerInd, v2->playerUnit, 1);
				nox_xxx_netSendSimpleObject2_4DF360(i->playerInd, v2->playerUnit);
				if (nox_common_gameFlags_check_40A5C0(4096)) {
					nox_xxx_netSendTeam_4D8670(i->playerInd, v2->playerUnit);
					nox_xxx_netSendTeam_4D8670(v1, i->playerUnit);
				}
			}
		}
	}
	nox_xxx_servMinimapRevealFlag_4DE380(v1);
	sub_4DF2E0(v1);
	if (nox_common_gameFlags_check_40A5C0(1024) && !sub_40AA70(v2)) {
		nox_xxx_netNeedTimestampStatus_4174F0(v2, 256);
	}
	if (0) {
		sub_4161E0();
		sub_416690();
	}
	sub_4E8110(v1);
	if (nox_common_gameFlags_check_40A5C0(64)) {
		v8 = sub_4E8310();
		game_ball_net_code = v8->net_code;
		game_ball_state = v8->state;
		nox_xxx_netSendBallStatus_4D95F0(v1, game_ball_state, game_ball_net_code);
	} else if (nox_common_gameFlags_check_40A5C0(32)) {
		for (j = nox_server_teamFirst_418B10(); j; j = nox_server_teamNext_418B60(j)) {
			v10 = sub_4E8320(j->field_57);
			team_flag_carrier_net_code = v10->carrier_net_code;
			team_flag_index = v10->flag_index;
			team_flag_status = v10->status;
			team_flag_id = v10->team_id;
			nox_xxx_netSendFlagStatus_4D95A0(v1, team_flag_id, team_flag_status, team_flag_index,
											 team_flag_carrier_net_code);
		}
	}
	nox_xxx_sendAllClientStatus_4175C0(v2);
	if (sub_409F40(0x2000)) {
		nox_xxx_sendAllPlayerIDs_4DE270(v2);
	}
	if (nox_common_gameFlags_check_40A5C0(4096)) {
		for (nox_object_t* k = nox_xxx_getFirstPlayerUnit_4DA7C0(); k; k = nox_xxx_getNextPlayerUnit_4DA7F0(k)) {
			nox_player_update_data_t* update = k->data_update;
			if (update->player->field_4792 == 1) {
				sub_4D9D20(v1, k);
			}
		}
	}
	if (nox_common_gameFlags_check_40A5C0(4096)) {
		sub_4D7A60(v1);
	}
	return v13->net_code;
}

//----- (004DE270) --------------------------------------------------------
int nox_xxx_sendAllPlayerIDs_4DE270(nox_playerInfo* a1) {
	int result; // eax
	int v3;     // [esp-1Ch] [ebp-28h]
	char v4[7]; // [esp+4h] [ebp-8h]

	result = 0;
	for (nox_object_t* i = nox_xxx_getFirstPlayerUnit_4DA7C0(); i; i = nox_xxx_getNextPlayerUnit_4DA7F0(i)) {
		if (i->net_code != a1->netCode) {
			nox_player_update_data_t* update = i->data_update;
			if (*(uint32_t*)&update->reserved_1[16]) {
				v4[0] = -46;
				*(uint16_t*)&v4[1] = nox_xxx_netGetUnitCodeServ_578AC0((uint32_t*)i);
				v3 = a1->playerInd;
				*(uint16_t*)&v4[3] = i->typ_ind;
				v4[5] = 1;
				v4[6] = 2;
				result = nox_xxx_netSendPacket0_4E5420(v3, v4, 7, 0, 1);
			}
		}
	}
	return result;
}

// GAME.EXE 004DE4D0 uses a native Go entry.

//----- (004DE7C0) --------------------------------------------------------
void nox_script_event_playerLeave(nox_playerInfo* pl);
int sub_4FF990(unsigned int a1);
// Only the prefix is named in defs.h. The disconnect routine also clears four
// per-client arrays in the tail of the Go PlayerUpdateData allocation.
typedef struct nox_player_disconnect_update_t {
	nox_player_update_data_t prefix;
	uint32_t extra_lives;
	uint32_t client_status[32];
	uint8_t respawn_markers[32];
	uint8_t field121_bytes[32];
	uint8_t field129_bytes[32];
} nox_player_disconnect_update_t;
_Static_assert(offsetof(nox_player_disconnect_update_t, client_status) ==
	(sizeof(void*) == 4 ? 324 : 420), "wrong disconnect client-status offset");
_Static_assert(offsetof(nox_player_disconnect_update_t, respawn_markers) ==
	(sizeof(void*) == 4 ? 452 : 548), "wrong disconnect respawn-marker offset");
_Static_assert(offsetof(nox_player_disconnect_update_t, field121_bytes) ==
	(sizeof(void*) == 4 ? 484 : 580), "wrong disconnect field121 offset");
_Static_assert(offsetof(nox_player_disconnect_update_t, field129_bytes) ==
	(sizeof(void*) == 4 ? 516 : 612), "wrong disconnect field129 offset");
void nox_xxx_playerForceDisconnect_4DE7C0(int ind) {
	nox_playerInfo* plr = nox_common_playerInfoFromNum_417090(ind);
	nox_script_event_playerLeave(plr);
	if (sub_4D12A0(ind)) {
		sub_4D1250(ind);
	}
	if (plr->field_2068) {
		int* v3 = sub_425A70(plr->field_2068);
		if (v3) {
			sub_425B60(v3, ind);
		}
	}
	nox_player_update_data_t* update = plr->playerUnit->data_update;
	if (update->trade_70) {
		nox_xxx_shopCancelSession_510DC0(update->trade_70);
	}
	update->trade_70 = NULL;
	sub_510E20(plr->playerInd);
	sub_4FF990(1 << plr->playerInd);

#ifndef NOX_SERVER
	if (!nox_common_gameFlags_check_40A5C0(2))
#endif // NOX_SERVER
	{
		memset(&plr->active, 0, 4);
	}

	sub_56F4F0((int*)&plr->prot_unit_hp_cur);
	sub_56F4F0((int*)&plr->prot_unit_hp_max);
	sub_56F4F0((int*)&plr->prot_unit_mana_cur);
	sub_56F4F0((int*)&plr->prot_unit_mana_max);
	sub_56F4F0((int*)&plr->prot_unit_experience);
	sub_56F4F0((int*)&plr->prot_unit_mass);
	sub_56F4F0((int*)&plr->prot_unit_buffs);
	sub_56F4F0((int*)&plr->prot_player_class);
	sub_56F4F0((int*)&plr->prot_player_field_2235);
	sub_56F4F0((int*)&plr->prot_player_field_2239);
	sub_56F4F0((int*)&plr->prot_player_orig_name);
	sub_56F4F0((int*)&plr->prot_4632);
	sub_56F4F0((int*)&plr->prot_4636);
	sub_56F4F0((int*)&plr->prot_player_gold);
	sub_56F4F0((int*)&plr->prot_4640);
	sub_56F4F0((int*)&plr->prot_player_level);

	char buf[3];
	buf[0] = 46;
	uint16_t unit_code = nox_xxx_netGetUnitCodeServ_578AC0(plr->playerUnit);
	memcpy(&buf[1], &unit_code, sizeof(unit_code));
	nox_xxx_netSendPacket0_4E5420(ind | 0x80, buf, 3, 0, 0);
	nox_xxx_delayedDeleteObject_4E5CC0(plr->playerUnit);
	plr->playerUnit = 0;
	for (nox_object_t* unit = nox_xxx_getFirstPlayerUnit_4DA7C0(); unit;
		 unit = nox_xxx_getNextPlayerUnit_4DA7F0(unit)) {
		nox_player_disconnect_update_t* data = unit->data_update;
		data->respawn_markers[ind] = 0;
		data->client_status[ind] = 0;
		data->field121_bytes[ind] = 0;
		data->field129_bytes[ind] = 0;
	}
	if (nox_xxx_gamePlayIsAnyPlayers_40A8A0()) {
		if (nox_common_gameFlags_check_40A5C0(1024) && !nox_xxx_serverIsClosing_446180() && sub_40A770() == 1) {
			sub_5095E0();
		}
	} else {
		sub_40A1F0(0);
		nox_xxx_playerForceSendLessons_416E50(1);
		nox_server_teamsResetYyy_417D00();
		sub_40A970();
	}
	sub_4E55F0(ind);
	nox_xxx_playerResetImportantCtr_4E4F40(ind);
	sub_4E4F30(ind);
	if (0) {
		sub_4161E0();
		sub_416690();
	}
	if (!nox_common_gameFlags_check_40A5C0(4096)) {
		return;
	}
	if (sub_4E9010() == 1) {
		nox_xxx_mapLoad_4D2450(sub_4E8E50());
	} else if (nox_server_questMaybeWarp_4E8F60()) {
		sub_4D60B0();
		sub_4D76E0(1);
		int v12 = nox_game_getQuestStage_4E3CC0();
		int v13 = nox_server_questNextStageThreshold_4D74F0(v12);
		nox_game_setQuestStage_4E3CD0(v13 - 1);
		nox_xxx_mapLoad_4D2450(sub_4E8E50());
	} else {
		nox_object_t* unit = nox_xxx_getFirstPlayerUnit_4DA7C0();
		if (unit) {
			while (!((nox_player_update_data_t*)unit->data_update)->quest_exit) {
				unit = nox_xxx_getNextPlayerUnit_4DA7F0(unit);
				if (!unit) {
					return;
				}
			}
			(void)sub_4E8E60();
		}
	}
}

//----- (004DF020) --------------------------------------------------------
char* sub_4DF020() {
	char* result;      // eax
	int v1;            // ecx
	unsigned char* v2; // edi
	char* v3;          // esi
	bool v4;           // zf
	char* i;           // esi
	char v6[60];       // [esp+8h] [ebp-3Ch]

	result = sub_459AA0(v6);
	v1 = 29;
	v2 = getMemAt(0x5D4594, 1563214);
	v3 = v6;
	v4 = 1;
	do {
		if (!v1) {
			break;
		}
		v4 = *(uint16_t*)v3 == *(uint16_t*)v2;
		v3 += 2;
		v2 += 2;
		--v1;
	} while (v4);
	if (!v4) {
		nox_playerInfo* player = nox_common_playerInfoGetFirst_416EA0();
		for (; player; player = nox_common_playerInfoGetNext_416EE0(player)) {
			if (player->playerInd != 31) {
				nox_xxx_netGuiGameSettings_4DD9B0(1, v6, player->playerInd);
			}
		}
		result = NULL;
		memcpy(getMemAt(0x5D4594, 1563214), v6, 0x38u);
		*getMemU16Ptr(0x5D4594, 1563270) = *(uint16_t*)&v6[56];
	}
	return result;
}

//----- (004DF0A0) --------------------------------------------------------
#if 0 // Preserved PE32 body; the public C entry uses native wall/player pointers.
int nox_xxx_wallSendDestroyed_4DF0A0(int a1, int a2) {
	int result; // eax
	int i;      // esi
	char v4[3]; // [esp+4h] [ebp-4h]

	v4[0] = 58;
	*(uint16_t*)&v4[1] = *(uint16_t*)(a1 + 10);
	if (a2 != 32) {
		return nox_xxx_netSendPacket0_4E5420(a2, v4, 3, 0, 1);
	}
	result = nox_xxx_getFirstPlayerUnit_4DA7C0();
	for (i = result; result; i = result) {
		nox_xxx_netSendPacket0_4E5420(*(unsigned char*)(*(uint32_t*)(*(uint32_t*)(i + 748) + 276) + 2064), v4, 3, 0, 1);
		result = nox_xxx_getNextPlayerUnit_4DA7F0(i);
	}
	return result;
}
#endif

//----- (004DF120) --------------------------------------------------------
#if 0 // Preserved PE32 body; the public C entry uses native wall/player pointers.
int sub_4DF120(void* a1) {
	int result; // eax
	int i;      // esi
	char v3[3]; // [esp+4h] [ebp-4h]

	v3[0] = 59;
	*(uint16_t*)&v3[1] = *(uint16_t*)((uint32_t)a1 + 10);
	result = nox_xxx_getFirstPlayerUnit_4DA7C0();
	for (i = result; result; i = result) {
		nox_xxx_netSendPacket0_4E5420(*(unsigned char*)(*(uint32_t*)(*(uint32_t*)(i + 748) + 276) + 2064), v3, 3, 0, 1);
		result = nox_xxx_getNextPlayerUnit_4DA7F0(i);
	}
	return result;
}
#endif

//----- (004DF180) --------------------------------------------------------
#if 0 // Preserved PE32 body; the public C entry uses native wall/player pointers.
int sub_4DF180(void* a1) {
	int result; // eax
	int i;      // esi
	char v3[3]; // [esp+4h] [ebp-4h]

	v3[0] = 60;
	*(uint16_t*)&v3[1] = *(uint16_t*)((uint32_t)a1 + 10);
	result = nox_xxx_getFirstPlayerUnit_4DA7C0();
	for (i = result; result; i = result) {
		nox_xxx_netSendPacket0_4E5420(*(unsigned char*)(*(uint32_t*)(*(uint32_t*)(i + 748) + 276) + 2064), v3, 3, 0, 1);
		result = nox_xxx_getNextPlayerUnit_4DA7F0(i);
	}
	return result;
}
#endif

#if 0 // Preserved original for the team report oracle.
//----- (004DF2E0) --------------------------------------------------------
void sub_4DF2E0(int a1) {
	nox_team_t* i;   // ebx
	nox_object_t* j; // esi

	if (a1 != 31) {
		for (i = nox_server_teamFirst_418B10(); i; i = nox_server_teamNext_418B60(i)) {
			sub_4197C0((wchar2_t*)i, a1);
			for (j = nox_xxx_getFirstPlayerUnit_4DA7C0(); j; j = nox_xxx_getNextPlayerUnit_4DA7F0(j)) {
				nox_object_team_t* member = (nox_object_team_t*)&j->field_12;
				if (nox_xxx_teamCompare2_419180(member, i->field_57)) {
					sub_4198A0(member, a1, j->net_code);
				}
			}
		}
	}
}
#endif

//----- (004DF3C0) --------------------------------------------------------
void sub_4DF3C0(nox_playerInfo* pl) {
	if (!pl || !pl->playerUnit) {
		return;
	}
	nox_object_t* unit = pl->playerUnit;
	nox_object_team_t* value = (nox_object_team_t*)&unit->field_12;
	char* settings = sub_416640();
	if (!sub_40A740() && !nox_common_gameFlags_check_40A5C0(0x8000)) {
		if (nox_xxx_getTeamCounter_417DD0()) {
			nox_team_t* team = sub_4189D0();
			if (team && !nox_xxx_servObjectHasTeam_419130(value)) {
				nox_xxx_createAtImpl_4191D0(team->field_57, value, 1, unit->net_code, 1);
			}
		}
		return;
	}
	if (!pl->field_2068) {
		return;
	}
	nox_team_t* team = nox_server_teamByXxx_418AE0(pl->field_2068);
	int has_team;
	if (team) {
		has_team = nox_xxx_servObjectHasTeam_419130(value);
	} else {
		int maximum = (unsigned char)settings[52];
		if ((nox_common_gameFlags_check_40A5C0(96) ||
			 nox_common_gameFlags_check_40A5C0(16) && nox_xxx_CheckGameplayFlags_417DA0(4)) &&
			maximum > 2) {
			maximum = 2;
		}
		int count = (unsigned char)sub_417DE0();
		if (count >= maximum) {
			return;
		}
		if (!nox_common_gameFlags_check_40A5C0(96) ||
			(count = (unsigned char)sub_417DE0(), count < sub_417DC0())) {
			team = sub_418A10();
			if (!team) {
				return;
			}
			sub_418800(team->name, pl->field_2072, 0);
			sub_418830(team, pl->field_2068);
			sub_4184D0(team->name);
			has_team = nox_xxx_servObjectHasTeam_419130(value);
		} else {
			return;
		}
	}
	if (has_team) {
		sub_4196D0(value, team, unit->net_code, 0);
	} else {
		nox_xxx_createAtImpl_4191D0(team->field_57, value, 1, unit->net_code, 0);
	}
}

// GAME.EXE 004DFE40, 004DFF40 and 004E0040 use typed Go entries.

//----- (004E01D0) --------------------------------------------------------
#if 0
// PE32 provenance; the public raw symbol is now the native-width Go export.
void nox_xxx_effectRegeneration_4E01D0(int a1, int a2) {
	uint32_t* v2;      // esi
	unsigned short v3; // di
	unsigned int v4;   // edi
	int v5;            // eax

	if (a2) {
		v2 = *(uint32_t**)(a2 + 492);
		if (v2) {
			if (v2[139]) {
				if ((unsigned int)(gameFrame() - v2[134]) >= (int)gameFPS() && !(v2[4] & 0x8020)) {
					v3 = nox_xxx_unitGetMaxHP_4EE7A0(*(uint32_t*)(a2 + 492));
					if ((unsigned short)nox_xxx_unitGetHP_4EE780((int)v2) < v3) {
						v4 = *(uint32_t*)(a1 + 108);
						if (*(uint32_t*)(a2 + 8) & 0x2000000) {
							v5 = nox_xxx_unitArmorInventoryEquipFlags_415C70(a2);
							if (v5 & 0x4000) {
								v4 /= 3u;
							}
						}
						if (!(gameFrame() %
							  (v4 * gameFPS() / (unsigned short)nox_xxx_unitGetMaxHP_4EE7A0((int)v2)))) {
							nox_xxx_unitAdjustHP_4EE460((int)v2, 1);
						}
					}
				}
			}
		}
	}
}
#endif

//----- (004E02C0) --------------------------------------------------------
#if 0
// PE32 provenance. The registered symbol is the native-width Go export in
// effect_continual_replenishment_4e02c0.go, including its original byte wrap.
void nox_xxx_attribContinualReplen_4E02C0(int a1, uint32_t* a2) {
	int v2;           // eax
	int v3;           // ecx
	unsigned char v4; // dl
	unsigned char v5; // al
	int v6;           // esi
	int v7;           // edx
	int v8;           // ecx

	if (a2) {
		if (!(gameFrame() % *(uint32_t*)(a1 + 108))) {
			v2 = a2[2];
			v3 = a2[184];
			if (v2 & 0x1000) {
				v4 = *(uint8_t*)(v3 + 108);
				v5 = *(uint8_t*)(v3 + 109);
				v6 = v4++;
				*(uint8_t*)(v3 + 108) = v4;
				if (v4 > v5) {
					*(uint8_t*)(v3 + 108) = v5;
				}
				v7 = *(unsigned char*)(v3 + 108);
				if (v6 != v7) {
					v8 = a2[123];
					if (v8) {
						if (*(uint8_t*)(v8 + 8) & 4) {
							nox_xxx_netReportCharges_4D82B0(
								*(unsigned char*)(*(uint32_t*)(*(uint32_t*)(v8 + 748) + 276) + 2064), a2, v7, v5);
						}
					}
				}
			}
		}
	}
}
#endif

//----- (004E03D0) --------------------------------------------------------
#if 0 // Original bodies retained; native entries are in modifier_defend_predicates_native_4e03d0.go.
int nox_xxx_inversionEffect_4E03D0(int a1, int a2, int a3, int a4, int a5, int* a6) {
	int result; // eax

	result = *(uint32_t*)(a1 + 96) >= 1;
	*a6 = result;
	return result;
}

//----- (004E03F0) --------------------------------------------------------
int nox_xxx_unusedCheckGripEffect_4E03F0(int a1, int a2, int a3, int a4) {
	int v4;                                   // esi
	int v5;                                   // edi
	uint32_t* v6;                             // eax
	int v7;                                   // eax
	int (*v8)(int, int, int, int, int, int*); // ecx
	bool v9;                                  // cc
	uint32_t* i;                              // [esp+10h] [ebp-4h]
	int v12;                                  // [esp+20h] [ebp+Ch]

	v4 = a3;
	if (!a3 || !(*(uint32_t*)(a3 + 8) & 0x3001000)) {
		return 0;
	}
	v5 = a4;
	v6 = (uint32_t*)(*(uint32_t*)(a3 + 692) + 8);
	v12 = 2;
	for (i = v6;; ++i) {
		v7 = *v6;
		if (v7) {
			v8 = *(int (**)(int, int, int, int, int, int*))(v7 + 88);
			if (v8 == nox_xxx_gripEffect_4E0480) {
				v8(v7, v4, a2, v5, a1, &a4);
				if (!a4) {
					break;
				}
			}
		}
		v6 = i + 1;
		v9 = ++v12 < 4;
		if (!v9) {
			return 0;
		}
	}
	return 1;
}

//----- (004E0480) --------------------------------------------------------
int nox_xxx_gripEffect_4E0480(int a1, int a2, int a3, int a4, int a5, int* a6) {
	int result; // eax

	result = *(uint32_t*)(a1 + 96) < 1;
	*a6 = result;
	return result;
}
#endif

//----- (004E0960) --------------------------------------------------------
// The native-pointer export is implemented in readiness_raw_native_4e0960.go.
// Preserve the ABI32 transcription verbatim as original provenance.
#if 0
int nox_xxx_itemCheckReadinessEffect_4E0960(int a1) {
	int v1; // esi
	int v2; // edx
	int i;  // ecx

	v1 = *(uint32_t*)(a1 + 692);
	if (!a1 || !(*(uint32_t*)(a1 + 8) & 0x13001000)) {
		return 0;
	}
	v2 = 2;
	for (i = v1 + 8; !*(uint32_t*)i || *(void (**)())(*(uint32_t*)i + 40) != nullsub_22; i += 4) {
		if (++v2 >= 4) {
			return 0;
		}
	}
	return *(uint32_t*)(*(uint32_t*)(v1 + 4 * v2) + 48);
}
// 4E0950: using guessed type void nullsub_22();

#endif

// Projectile reflection at GAME.EXE 004E0A70 uses a typed Go entry.

//----- (004E0B30) --------------------------------------------------------
#if 0 // Preserved PE32 body; native pointer entry is in raw_default_damage_export_4e0b30.go.
int nox_xxx_damageDefaultProc_4E0B30(int a1, int a2, int a3, int a4, int a5) {
	int v6;                // edi
	int v7;                // eax
	int v8;                // eax
	int v9;                // eax
	int v10;               // edi
	int v11;               // eax
	int v12;               // ebp
	int v13;               // eax
	int v14;               // eax
	int v15;               // ebp
	double v16;            // st7
	double v17;            // st7
	int v18;               // eax
	int v19;               // ebx
	int v20;               // eax
	int v21;               // edx
	int v22;               // eax
	char v23;              // al
	uint32_t* v24;         // eax
	int v25;               // edx
	int v26;               // ecx
	int v27;               // eax
	int v28;               // eax
	void (*v29)(int, int); // eax
	unsigned char v30;     // al
	double v31;            // st7
	unsigned short v32;    // bp
	int v33;               // eax
	float v34;             // ecx
	int v35;               // eax
	float v36;             // edx
	char v37;              // al
	int v38;               // ebp
	int v39;               // eax
	int v40;               // eax
	float v41;             // [esp+0h] [ebp-28h]
	float v42;             // [esp+4h] [ebp-24h]
	float v43;             // [esp+4h] [ebp-24h]
	float v44;             // [esp+4h] [ebp-24h]
	int v45[4];            // [esp+18h] [ebp-10h]
	float v46;             // [esp+30h] [ebp+8h]
	float v47;             // [esp+30h] [ebp+8h]

	if (!a1) {
		return 1;
	}
	if (nox_xxx_testUnitBuffs_4FF350(a1, 23)) {
		if (!((unsigned char)gameFrame() & 3)) {
			nox_xxx_aud_501960(71, a1, 0, 0);
			return 1;
		}
		return 1;
	}
	v6 = a5;
	if (*(uint8_t*)(a1 + 8) & 2) {
		v7 = *(uint32_t*)(a1 + 748);
		*(uint32_t*)(v7 + 2188) = 0;
		if (v6 == 1 || v6 == 12 || v6 == 7 || v6 == 14 || v6 == 6) {
			*(uint32_t*)(v7 + 1440) |= 0x80000u;
		}
	}
	v8 = *(uint32_t*)(a1 + 16);
	if ((v8 & 0x8000) != 0) {
		if (!nox_xxx_unitIsZombie_534A40(a1)) {
			return 1;
		}
		v9 = a3;
		if (!a3) {
			v9 = a2;
		}
		*(uint32_t*)(a1 + 520) = v9;
		*(uint32_t*)(a1 + 524) = v6;
		*(uint32_t*)(a1 + 536) = gameFrame();
		return 1;
	}
	v10 = a2;
	if (!nox_xxx_CheckGameplayFlags_417DA0(1)) {
		v11 = nox_xxx_findParentChainPlayer_4EC580(a2);
		v12 = v11;
		if (v11) {
			if (*(uint8_t*)(v11 + 8) & 6 && !nox_xxx_unitIsEnemyTo_5330C0(a1, v11) &&
				(a1 != v12 || nox_common_gameFlags_check_40A5C0(4096))) {
				return 1;
			}
		}
	}
	if (a2 && a3 && !nox_xxx_unitIsEnemyTo_5330C0(a1, a2) && *(uint8_t*)(a1 + 8) & 6 &&
			sub_4E1400((nox_object_t*)(uintptr_t)(uint32_t)a2, (nox_object_t*)(uintptr_t)(uint32_t)a3) && !sub_4E1470((nox_object_t*)(uintptr_t)(uint32_t)a3) ||
		*(uint8_t*)(a1 + 16) & 2) {
		return 1;
	}
	if (a2) {
		if (nox_xxx_testUnitBuffs_4FF350(a1, 22)) {
			if (*(uint8_t*)(a2 + 8) & 6) {
				if (a3) {
					if (sub_4E1400((nox_object_t*)(uintptr_t)(uint32_t)a2, (nox_object_t*)(uintptr_t)(uint32_t)a3)) {
						nox_xxx_aud_501960(135, a2, 0, 0);
						nox_xxx_spellBuffOff_4FF5B0((nox_object_t*)(uintptr_t)(uint32_t)a1, 22);
						v41 = nox_xxx_gamedataGetFloatTable_419D70("ShockDamage", 4);
						v13 = nox_float2int(v41);
						(*(void (**)(int, int, uint32_t, int, int))(a2 + 716))(a2, a1, 0, v13, 9);
						if (*(uint8_t*)(a2 + 8) & 4) {
							nox_xxx_playerSetState_4FA020((uint32_t*)a2, 23);
						}
					}
				}
			}
		}
	}
	if (*(uint8_t*)(a1 + 8) & 2) {
		v14 = *(uint32_t*)(a1 + 12);
		v15 = a5;
		if (v14 & 0x200 && a5 == 5) {
			return 1;
		}
		if (v14 & 0x400) {
			if (a5 == 1 || a5 == 12) {
				return 1;
			}
			if (a5 == 7) {
				a4 /= 2;
				goto LABEL_53;
			}
		} else if (v14 & 0x800) {
			if (a5 == 9) {
				return 1;
			}
			if (a5 == 17) {
				return 1;
			}
		}
	} else {
		v15 = a5;
	}
	if (v15 != 1 && v15 != 12 && v15 != 7) {
		goto LABEL_58;
	}
LABEL_53:
	v16 = nox_xxx_checkFireProtect_4DFE40((uint32_t*)a1);
	if (v16 != 0.0 && !((unsigned char)gameFrame() & 3)) {
		nox_xxx_aud_501960(104, a1, 0, 0);
	}
	v46 = v16;
	v42 = (1.0 - v46) * (double)a4;
	a4 = nox_float2int(v42);
	if (!a4) {
		a4 = 1;
	}
LABEL_58:
	if (v15 == 9 || v15 == 17) {
		v17 = nox_xxx_checkElectrProtect_4DFF40((uint32_t*)a1);
		if (v17 != 0.0 && !((unsigned char)gameFrame() & 3)) {
			nox_xxx_aud_501960(108, a1, 0, 0);
		}
		v47 = v17;
		v43 = (1.0 - v47) * (double)a4;
		a4 = nox_float2int(v43);
		if (!a4) {
			a4 = 1;
		}
		v18 = *(uint32_t*)(a1 + 8);
		if (v18 & 4) {
			*(uint16_t*)(*(uint32_t*)(a1 + 748) + 160) = 2;
		} else if (v18 & 2 && *(uint8_t*)(a1 + 12) & 0x10) {
			*(uint8_t*)(*(uint32_t*)(a1 + 748) + 2094) = 2;
		}
	}
	if (!v10) {
		*(uint32_t*)(a1 + 528) = 0;
		*(uint32_t*)(a1 + 532) = 0;
		if (v15 == 12) {
			nox_xxx_spellBuffOff_4FF5B0((nox_object_t*)(uintptr_t)(uint32_t)a1, 0);
		}
		v19 = a3;
		goto LABEL_87;
	}
	v19 = a3;
	if (a3) {
		if (*(uint32_t*)(a3 + 8) & 0x1001000) {
			*(uint32_t*)(a1 + 528) = *(uint32_t*)(v10 + 72);
			*(uint32_t*)(a1 + 532) = *(uint32_t*)(v10 + 76);
		} else {
			*(uint32_t*)(a1 + 528) = *(uint32_t*)(a3 + 72);
			*(uint32_t*)(a1 + 532) = *(uint32_t*)(a3 + 76);
		}
		if (a3 == v10 || !(*(uint8_t*)(a1 + 8) & 2)) {
			goto LABEL_83;
		}
		v20 = *(uint32_t*)(a1 + 748);
		HIWORD(v21) = 0;
		*(uint32_t*)(v20 + 2188) = 1;
		LOWORD(v21) = *(uint16_t*)(a3 + 4);
	} else {
		*(uint32_t*)(a1 + 528) = *(uint32_t*)(v10 + 72);
		*(uint32_t*)(a1 + 532) = *(uint32_t*)(v10 + 76);
		if (v15 != 10 && v15 != 2 || !(*(uint8_t*)(a1 + 8) & 2)) {
			goto LABEL_83;
		}
		v20 = *(uint32_t*)(a1 + 748);
		HIWORD(v21) = 0;
		*(uint32_t*)(v20 + 2188) = 1;
		LOWORD(v21) = *(uint16_t*)(v10 + 4);
	}
	*(uint32_t*)(v20 + 2184) = v21;
LABEL_83:
	nox_xxx_spellBuffOff_4FF5B0((nox_object_t*)(uintptr_t)(uint32_t)a1, 0);
LABEL_87:
	v22 = *(uint32_t*)(a1 + 8);
	if (v22 & 4 || v22 & 2 && *(uint8_t*)(a1 + 12) & 0x10) {
		nox_xxx_itemApplyDefendEffect2_4E1320((nox_object_t*)(uintptr_t)(uint32_t)a1,
			(nox_object_t*)(uintptr_t)(uint32_t)v10, (nox_object_t*)(uintptr_t)(uint32_t)v19, &a4, v15);
	}
	if (v19) {
		*(uint32_t*)(a1 + 520) = v19;
	} else {
		*(uint32_t*)(a1 + 520) = v10;
	}
	v23 = *(uint8_t*)(a1 + 8);
	*(uint32_t*)(a1 + 524) = v15;
	*(uint32_t*)(a1 + 536) = gameFrame();
	if (v23 & 2) {
		v24 = *(uint32_t**)(a1 + 748);
		v25 = v24[360];
		v26 = v24[547];
		BYTE1(v25) |= 2u;
		v24[360] = v25;
		if (!v26) {
			v24[547] = 2;
			v24[546] = v15;
		}
	}
	if (v19 && *(uint32_t*)(v19 + 8) & 0x1001000) {
		nox_xxx_itemApplyPreDamageEffect_4E13B0((nox_object_t*)(uintptr_t)(uint32_t)a1,
			(nox_object_t*)(uintptr_t)(uint32_t)v10, (nox_object_t*)(uintptr_t)(uint32_t)v19, &a4);
	}
	if (a1 != v19 || !(*(uint32_t*)(a1 + 8) & 0x1001000)) {
		if (!v10 || !(*(uint8_t*)(v10 + 8) & 2) || !*(uint32_t*)(v10 + 748) ||
			(v27 = nox_xxx_monsterGetSoundSet_424300(v10)) == 0 ||
			(v28 = *(uint32_t*)(v27 + 32)) == 0 || nox_xxx_getSevenDwords3_501940(v28) <= 0) {
			v29 = *(void (**)(int, int))(a1 + 720);
			if (v29) {
				if (v19) {
					v29(a1, v19);
				} else {
					v29(a1, v10);
				}
			} else if (v19) {
				nox_xxx_soundDefaultDamageSound_532E20(a1, v19);
			} else {
				nox_xxx_soundDefaultDamageSound_532E20(a1, v10);
			}
		}
	}
	if (v10 && *(uint8_t*)(a1 + 8) & 6 && nox_xxx_testUnitBuffs_4FF350(v10, 13)) {
		nox_xxx_aud_501960(163, a3, 0, 0);
		v30 = nox_xxx_buffGetPower_4FF570(v10, 13);
		v31 = nox_xxx_gamedataGetFloatTable_419D70("VampirismCoeff", (unsigned int)v30 - 1);
		v44 = v31 * (double)a4;
		v32 = nox_float2int(v44);
		if (v32 < 1u) {
			v32 = 1;
		}
		nox_xxx_unitAdjustHP_4EE460(v10, v32);
		v45[0] = nox_float2int(*(float*)(v10 + 56));
		v33 = nox_float2int(*(float*)(v10 + 60));
		v34 = *(float*)(a1 + 56);
		v45[1] = v33;
		v35 = nox_float2int(v34);
		v36 = *(float*)(a1 + 60);
		v45[2] = v35;
		v45[3] = nox_float2int(v36);
		nox_xxx_netSendVampFx_523270((char)0xA2, (short*)v45, v32);
	}
	nox_xxx_gameballOnPlayerDamage_4E1230(v10, a1, a4);
	if (*(uint8_t*)(a1 + 8) & 4) {
		if (a4 >= 20) {
			v37 = *(uint8_t*)(*(uint32_t*)(a1 + 748) + 88);
			if (v37 != 1 && v37 != 15) {
				nox_xxx_playerSetState_4FA020((uint32_t*)a1, 30);
			}
		}
	}
	if (nox_common_gameFlags_check_40A5C0(6144)) {
		sub_4FB050(v10, a1, &a4);
	}
	if (!v10) {
		goto LABEL_137;
	}
	if (*(uint8_t*)(v10 + 8) & 2) {
		v38 = v10;
	} else {
		v39 = *(uint32_t*)(v10 + 508);
		if (!v39 || !(*(uint8_t*)(v39 + 8) & 2)) {
			goto LABEL_137;
		}
		v38 = *(uint32_t*)(v10 + 508);
	}
	if (v38 && nox_xxx_unitIsEnemyTo_5330C0(a1, v38)) {
		// Retained PE32 int-ABI parent; its native-width migration remains separate.
		sub_532880((nox_object_t*)(uintptr_t)(uint32_t)v38);
	}
LABEL_137:
	if (nox_xxx_testUnitBuffs_4FF350(a1, 26) && a5 != 5) {
		v40 = a3;
		if (!a3) {
			v40 = v10;
		}
		if (a5 != 15 || v10 != a1) {
			nox_xxx_unitShieldReduceDamage_52F710(a1, &a4, a5, v40);
		}
		if (!a4) {
			return 0;
		}
	}
	nox_xxx_unitDamageClear_4EE5E0(a1, a4);
	return 1;
}
#endif

// GameBall damage at GAME.EXE 004E1230 uses a typed Go entry.

//----- (004E1320) --------------------------------------------------------
#if 0 // Restored from sealed GAME.EXE 004E1320 in item_defend_effects_4e1320_export.go.
int nox_xxx_itemApplyDefendEffect2_4E1320(nox_object_t* a1p, nox_object_t* a2p, nox_object_t* a3p, int* a4, int a5) {
	// Public ABI restoration only. Remove these PE32 temporaries when the
	// original body is moved to the native-pointer implementation.
	int a1 = (int)(uintptr_t)a1p;
	int a2 = (int)(uintptr_t)a2p;
	int a3 = (int)(uintptr_t)a3p;
	int result;   // eax
	uint32_t* v6; // esi
	int v7;       // ebp
	int* v8;      // ebx
	int v9;       // eax
	int2 v10;     // [esp+4h] [ebp-8h]
	int v11;      // [esp+20h] [ebp+14h]

	result = a1;
	v6 = *(uint32_t**)(a1 + 504);
	if (v6) {
		v7 = a5;
		do {
			result = v6[4];
			if (result & 0x100) {
				v11 = 2;
				v8 = (int*)(v6[173] + 8);
				do {
					v9 = *v8;
					if (*v8) {
						if (*(uint32_t*)(v9 + 76)) {
							v10.field_0 = *a4;
							v10.field_4 = v7;
							(*(void (**)(int, uint32_t*, int, int, int, int2*))(v9 + 76))(v9, v6, a1, a3, a2, &v10);
							*a4 = v10.field_0;
						}
					}
					++v8;
					result = --v11;
				} while (v11);
			}
			v6 = (uint32_t*)v6[124];
		} while (v6);
	}
	return result;
}
#endif

//----- (004E13B0) --------------------------------------------------------
#if 0 // Restored from sealed GAME.EXE 004E13B0 in item_pre_damage_4e13b0_export.go.
int nox_xxx_itemApplyPreDamageEffect_4E13B0(nox_object_t* a1p, nox_object_t* a2p, nox_object_t* a3p, int32_t* a4p) {
	// Public ABI restoration only. Keep the original PE32 temporaries until
	// the separate native-pointer body restoration.
	int a1 = (int)(uintptr_t)a1p;
	int a2 = (int)(uintptr_t)a2p;
	int a3 = (int)(uintptr_t)a3p;
	int a4 = (int)(uintptr_t)a4p;
	int v4;                              // edi
	int* v5;                             // esi
	int v6;                              // eax
	void (*v7)(int, int, int, int, int); // ecx
	int result;                          // eax
	int v9;                              // [esp+1Ch] [ebp+Ch]

	v4 = a3;
	v9 = 4;
	v5 = *(int**)(v4 + 692);
	do {
		v6 = *v5;
		if (*v5) {
			v7 = *(void (**)(int, int, int, int, int))(v6 + 64);
			if (v7) {
				v7(v6, v4, a2, a1, a4);
			}
		}
		++v5;
		result = --v9;
	} while (v9);
	return result;
}
#endif

//----- (004E1400) --------------------------------------------------------
#if 0 // Original PE32 transcription; native pointer entry is exported from Go.
int sub_4E1400(int a1, uint32_t* a2) {
	int v2;     // eax
	int result; // eax
	int v4;     // eax

	if (a2) {
		v4 = a2[2];
		if (v4 & 0x1000) {
			if (!(a2[3] & 0x47F0000)) {
				return 1;
			}
			if (*(uint8_t*)(a2[184] + 96) & 2) {
				return 1;
			}
		} else if (v4 & 0x1000000 && !(a2[3] & 0x47F00FE)) {
			return 1;
		}
		result = ((unsigned char)v4 >> 1) & 1;
	} else {
		v2 = *(uint32_t*)(a1 + 8);
		if (v2 & 4) {
			return 1;
		}
		result = v2 & 2 && *(uint8_t*)(a1 + 12) & 0x10;
	}
	return result;
}
#endif

//----- (004E1470) --------------------------------------------------------
#if 0 // Original PE32 transcription; native pointer entry is exported from Go.
int sub_4E1470(int a1) {
	int v1;     // ecx
	int result; // eax

	result = 0;
	if (a1) {
		if (*(uint32_t*)(a1 + 8) & 0x1000000) {
			v1 = *(uint32_t*)(a1 + 12);
			if (v1 & 0x4000) {
				result = 1;
			}
		}
	}
	return result;
}
#endif
