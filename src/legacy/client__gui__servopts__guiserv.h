#ifndef NOX_PORT_CLIENT_GUI_SERVOPTS_GUISERV
#define NOX_PORT_CLIENT_GUI_SERVOPTS_GUISERV

#include "defs.h"

typedef struct {
	nox_window* object_filter;
	nox_window* object_list;
	nox_window* spell_filter;
	nox_window* spell_names;
	nox_window* spell_classes;
	nox_window* players;
	nox_window* remove_team;
	nox_window* rename_team;
	nox_window* options;
	nox_window* ctf_hud;
	nox_window* flagball_hud;
} nox_game2_server_windows_t;

_Static_assert(sizeof(nox_game2_server_windows_t) == 11 * sizeof(nox_window*),
	"server GUI storage must contain 11 native window pointers");

nox_game2_server_windows_t* nox_game2_server_windows_storage(void);

void nox_gui_gamemode_load_457410(void);
void sub_457A10(void);
int sub_459650(wchar2_t* title);
void nox_xxx_windowServerOptionsFillGametypeList_4596A0(void);

wchar2_t* nox_xxx_guiServerOptionsGetGametypeName_4573C0(short mode);
int nox_xxx_guiServerOptsLoad_457500();
int nox_xxx_options_457AA0(int a1, uint8_t* a2);
int nox_xxx_options_457B00(int a1, uint8_t* a2);
int sub_457FE0();
int sub_4580E0(char* a1);
void nox_client_guiserv_updateMapList_458230(int a1, char* name, bool a3);
int nox_xxx_guiServerOptionsProcPre_4585D0(nox_window* win, unsigned int event, uintptr_t a3, uintptr_t a4);
int sub_459880(char* a1);
int sub_459CD0();
int sub_459C10();

#endif // NOX_PORT_CLIENT_GUI_SERVOPTS_GUISERV
