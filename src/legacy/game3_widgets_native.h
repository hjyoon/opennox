#ifndef OPENNOX_GAME3_WIDGETS_NATIVE_H
#define OPENNOX_GAME3_WIDGETS_NATIVE_H
#include <stdlib.h>
#include <stdint.h>
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "GAME3.h"
#include "GAME3_1.h"
#include "client__gui__window.h"
#include "client__gui__gadgets__listbox.h"
#include "client__video__draw_common.h"
extern uint32_t dword_5d4594_1307720;
extern nox_window* dword_5d4594_1307716;
extern uint32_t nox_color_white_2523948, nox_color_black_2650656;
extern uint32_t nox_color_violet_2598268, nox_color_red_2589776;
extern uint32_t nox_color_green_2614268, nox_color_cyan_2649820;
extern uint32_t nox_color_blue_2650684, nox_color_orange_2614256;
extern uint32_t nox_color_yellow_2589772, dword_8531A0_2572;
// Macro/function-pointer boundaries only; widget algorithms remain in Go.
static nox_gui_server_ent_t* nox_game3_server_record(nox_list_item_t* p) {
 return p ? &nox_wol_server_node_from_list(p)->server : NULL;
}
static void* nox_game3_owner_callback(nox_window* w) { return w ? w->field_94 : NULL; }
#endif
