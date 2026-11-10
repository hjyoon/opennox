#ifndef OPENNOX_GAME2_BOOK_NATIVE_H
#define OPENNOX_GAME2_BOOK_NATIVE_H
#include <stdint.h>
#include <stdlib.h>
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
#include "GAME5_2.h"
#include "client__drawable__drawable.h"
#include "client__gui__guibook.h"
#include "client__gui__window.h"
#include "common__magic__speltree.h"
#include "common__net_list.h"
#include "common__strman.h"
#include "client__video__draw_common.h"
#include "input_common.h"
#include "operators.h"
#include "client__gui__guispell.h"
#include "common__gamemech__pausefx.h"

extern uintptr_t dword_8531A0_2576;
extern uintptr_t dword_5d4594_1047516;
extern uint32_t dword_5d4594_1046656, dword_5d4594_1047512;
extern uint32_t dword_5d4594_1046864, dword_5d4594_1046868, dword_5d4594_1046872;
extern uint32_t dword_5d4594_1046932, dword_5d4594_1046936;
extern uint32_t dword_5d4594_1047520, dword_5d4594_1047524, dword_5d4594_1047528;
extern uint32_t dword_5d4594_1047532, dword_5d4594_1047536, dword_5d4594_1047540;
extern uint32_t dword_5d4594_1046852, dword_5d4594_1046636, dword_5d4594_1046640;
extern uint32_t dword_5d4594_1046652, dword_5d4594_1046648;
extern uint32_t nox_xxx_aNox_cfg_0_587000_132132;
extern uint32_t nox_xxx_aNox_cfg_0_587000_132136;
extern unsigned int nox_player_netCode_85319C;
extern nox_window *nox_win_unk1, *dword_5d4594_1046944, *dword_5d4594_1046948;
extern nox_window *dword_5d4594_1046952, *dword_5d4594_1046956;
extern nox_things_imageRef_t *dword_5d4594_1046924, *dword_5d4594_1046928;
extern nox_video_bag_image_t *nox_book_arrow_images[8];
extern nox_video_bag_image_t *nox_book_background_image, *nox_book_guide_tab_lit_image, *nox_book_spell_tab_lit_image;
extern void* nox_xxx_aClosewoodengat_587000_133480;
extern float2 obj_5d4594_1046620;
extern int nox_win_width, nox_win_height;

// Native typed access to existing C-owned uintptr_t globals. These getters do
// not reconstruct a pointer from any PE32 packed DWORD or retain Go memory.
static nox_playerInfo* game2_book_player(void) { return (nox_playerInfo*)dword_5d4594_1047516; }
static nox_playerInfo* game2_book_current_player(void) { return (nox_playerInfo*)dword_8531A0_2576; }
static void game2_book_reward_scalar(uintptr_t token, int id, int extra) {
    nox_xxx_bookRewardCli_499CF0((int*)token, id, extra);
}

// Public mixed pointer/tag ABIs stay in root-owned thin C adapters. The private
// Go exports return words, so Go never manufactures unsafe.Pointer(1/2/3/4).
int game2_book_compare(void* a, void* b);
int nox_xxx_book_45BD30(nox_window* win, nox_window_data* draw);
uintptr_t game2_book_forward_word(uintptr_t kind, int id, int2* pos);
void game2_book_ability_reward(int id, uintptr_t notify, int extra);
uintptr_t game2_book_disable_ability_word(int id);
uintptr_t game2_book_position_word(uint32_t* pos);

nox_image_ref_anim_t* nox_book_anim_data(nox_things_imageRef_t* ref);
void nox_book_anim_set_end(nox_things_imageRef_t* ref, void* fnc);
void nox_book_anim_start(nox_things_imageRef_t* ref);
int nox_client_localPlayerHasFlags30(uint32_t flags);

#endif
