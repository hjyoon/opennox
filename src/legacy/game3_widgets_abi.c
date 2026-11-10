#include "game3_widgets_native.h"
extern int nox_game3_distance_go(uint32_t*, nox_gui_server_ent_t*);
extern int nox_game3_first_data_go(nox_scrollListBox_data*);
// Const qualification adapters; neither changes the pointer representation.
int sub_4A2560(const uint32_t* xy,const nox_gui_server_ent_t* server) {
 return nox_game3_distance_go((uint32_t*)xy,(nox_gui_server_ent_t*)server);
}
int sub_4A4800(const nox_scrollListBox_data* data) {
 return nox_game3_first_data_go((nox_scrollListBox_data*)data);
}
