#ifndef NOX_PORT_CLIENT_GUI_SERVOPTS_PLAYRLST
#define NOX_PORT_CLIENT_GUI_SERVOPTS_PLAYRLST

#include "defs.h"
#include "common__system__team.h"

typedef struct nox_gui_server_player_entry {
	nox_list_item_t list;
	wchar2_t name[24];
	uint32_t net_code;
	uint8_t reserved_64[8];
} nox_gui_server_player_entry;

typedef struct nox_gui_server_team_entry {
	nox_list_item_t list;
	wchar2_t name[24];
	uint32_t team_id;
	uint8_t color;
	uint8_t reserved_65[3];
	uint32_t def_ind;
} nox_gui_server_team_entry;

typedef struct nox_gui_server_players_state {
	nox_list_item_t players;
	nox_list_item_t teams;
} nox_gui_server_players_state;

extern nox_gui_server_players_state nox_gui_server_players;
extern nox_window* dword_5d4594_1045684;
extern nox_window* dword_5d4594_1045688;
extern nox_window* dword_5d4594_1045692;

nox_window* nox_xxx_guiServerPlayersLoad_456270(nox_window* parent);
int sub_4567C0(nox_window* win, int event, nox_window* event_win, uintptr_t event_arg);
int sub_457010(nox_team_t* team, wchar2_t* name);
uintptr_t sub_457230(wchar2_t* name);

#endif // NOX_PORT_CLIENT_GUI_SERVOPTS_PLAYRLST
