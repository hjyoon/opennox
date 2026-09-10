#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../GAME1_1.h"
#include "../GAME1_3.h"
#include "../GAME2.h"
#include "../client__gui__servopts__playrlst.h"

typedef nox_window* (*server_players_load_fn)(nox_window*);
typedef int (*server_players_draw_fn)(nox_window*, nox_window_data*);
typedef int (*server_players_event_fn)(nox_window*, int, nox_window*, uintptr_t);
typedef int (*server_players_rename_fn)(nox_team_t*, wchar2_t*);
typedef uintptr_t (*server_players_append_team_fn)(wchar2_t*);
typedef void (*server_players_change_team_id_fn)(int, int);
typedef uintptr_t (*dialog_response_fn)(char);
typedef void (*team_set_id_fn)(nox_team_t*, int);
typedef nox_team_t* (*team_select_fn)(void);
typedef void (*team_set_lessons_fn)(nox_team_t*, int);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == sizeof(uintptr_t),
	"server-player pointers must remain native-width");
_Static_assert(
	_Generic(&nox_xxx_guiServerPlayersLoad_456270,
		server_players_load_fn: 1, default: 0),
	"00456270 must accept and return native window pointers");
_Static_assert(_Generic(&sub_456640, server_players_draw_fn: 1, default: 0),
	"00456640 must preserve the native draw callback ABI");
_Static_assert(_Generic(&sub_4567C0, server_players_event_fn: 1, default: 0),
	"004567C0 must preserve native-width GUI event payloads");
_Static_assert(_Generic(&sub_457010, server_players_rename_fn: 1, default: 0),
	"00457010 must accept a native team pointer");
_Static_assert(_Generic(&sub_457230, server_players_append_team_fn: 1, default: 0),
	"00457230 must preserve its native-width GUI response");
_Static_assert(_Generic(&sub_457350, server_players_change_team_id_fn: 1, default: 0),
	"00457350 must preserve its two fixed-width team-ID arguments");
_Static_assert(_Generic(&sub_449E60, dialog_response_fn: 1, default: 0),
	"00449E60 must preserve its native-width dialog response");
_Static_assert(_Generic(&sub_418830, team_set_id_fn: 1, default: 0),
	"00418830 must accept a native team pointer");
_Static_assert(_Generic(&sub_4189D0, team_select_fn: 1, default: 0),
	"004189D0 must return a native team pointer");
_Static_assert(_Generic(&sub_418A10, team_select_fn: 1, default: 0),
	"00418A10 must return a native team pointer");
_Static_assert(_Generic(&nox_xxx_netChangeTeamID_419090,
		team_set_lessons_fn: 1, default: 0),
	"00419090 must accept a native team pointer");
_Static_assert(offsetof(nox_gui_server_player_entry, name) == sizeof(nox_list_item_t),
	"player names must not overlap native list links");
_Static_assert(offsetof(nox_gui_server_player_entry, net_code) == sizeof(nox_list_item_t) + 48,
	"player net codes must retain the PE32 record offset after native links");
_Static_assert(sizeof(((nox_gui_server_player_entry*)0)->net_code) == 4,
	"player net codes must remain 32-bit values");
_Static_assert(offsetof(nox_gui_server_team_entry, name) == sizeof(nox_list_item_t),
	"team names must not overlap native list links");
_Static_assert(offsetof(nox_gui_server_team_entry, team_id) == sizeof(nox_list_item_t) + 48,
	"team IDs must retain the PE32 record offset after native links");
_Static_assert(sizeof(((nox_gui_server_team_entry*)0)->team_id) == 4,
	"team IDs must remain 32-bit values");
_Static_assert(offsetof(nox_gui_server_team_entry, color) == sizeof(nox_list_item_t) + 52,
	"team colors must retain the PE32 record offset after native links");
_Static_assert(sizeof(((nox_gui_server_team_entry*)0)->color) == 1,
	"team colors must remain byte values");
_Static_assert(offsetof(nox_gui_server_team_entry, def_ind) == sizeof(nox_list_item_t) + 56,
	"team definitions must retain the PE32 record offset after native links");
_Static_assert(sizeof(((nox_gui_server_team_entry*)0)->def_ind) == 4,
	"team definition indices must remain 32-bit values");
_Static_assert(offsetof(nox_gui_server_players_state, teams) == sizeof(nox_list_item_t),
	"native player and team list heads must not overlap");
_Static_assert(sizeof(nox_gui_server_players_state) == 2 * sizeof(nox_list_item_t),
	"server-player list state must contain two independent native heads");
_Static_assert(sizeof(dword_5d4594_1045684) == sizeof(void*) &&
		sizeof(dword_5d4594_1045688) == sizeof(void*) &&
		sizeof(dword_5d4594_1045692) == sizeof(void*),
	"server-player window globals must preserve native pointers");
#if UINTPTR_MAX == UINT32_MAX
_Static_assert(sizeof(nox_gui_server_player_entry) == 72,
	"PE32 player list records must remain 72 bytes");
_Static_assert(sizeof(nox_gui_server_team_entry) == 72,
	"PE32 team list records must remain 72 bytes");
#endif
_Static_assert(offsetof(nox_drawable, field_7) ==
		offsetof(nox_drawable, field_6) + sizeof(uint32_t),
	"drawable object-team records must use native structure offsets");

int main(void) {
	return 0;
}
