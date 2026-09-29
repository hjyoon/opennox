#ifndef NOX_PORT_CLIENT_SHELL_NOXWORLD
#define NOX_PORT_CLIENT_SHELL_NOXWORLD

#include "defs.h"

void sub_4373A0();
void nox_client_refreshServerList_4378B0();
int nox_game_showGameSel_4379F0();
int sub_4383A0();
int sub_438770();
int sub_438BD0();
void nox_client_gui_serverInfoBlockCheckExp_439370(int2* a1, const nox_gui_server_ent_t* a2);
void nox_client_gui_serverInfoBlock_4394D0(const nox_gui_server_ent_t* server);
int nox_xxx_windowMultiplayerSub_439E70(nox_window* win, unsigned int event, uintptr_t event_arg,
									 uintptr_t event_arg2);
void sub_43A810();
uint32_t* sub_43B630();
void sub_43B6E0();
void sub_43B750();
void nox_gui_wol_newServerLine_43B7C0(nox_gui_server_ent_t* srv);
wchar2_t* nox_gui_wol_gameModeString_43BCB0(short a1);
nox_video_bag_image_t* nox_wol_server_icon(int small, int tier, int lit);
nox_video_bag_image_t* nox_wol_map_icon(int quadrant);

#endif // NOX_PORT_CLIENT_SHELL_NOXWORLD
