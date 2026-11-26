#ifndef OPENNOX_GAME2_BOOK_NATIVE_H
#define OPENNOX_GAME2_BOOK_NATIVE_H
#include <stddef.h>
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


typedef struct {
	uintptr_t player_word;
	float2 reward_velocity;
	uint32_t reward_x_bits;
	uint32_t reward_y_bits;
	uint32_t reward_start_frame;
	uint32_t reward_ability;
	uint32_t line_height;
	uint32_t reward_slot;
	uint32_t open_requested;
	uint32_t page_mode;
	uint32_t guide_mode;
	uint32_t entry_index;
	uint32_t list_page;
	uint32_t hidden_entry_count;
	uint32_t reward_active;
	uint32_t reward_id;
	uint32_t list_drag_id;
	uint32_t list_drag_x;
	uint32_t list_drag_y;
	uint32_t page_drag_id;
	uint32_t show_detail_page;
	uint32_t list_drag_index;
} nox_book_state_t;
_Static_assert(offsetof(nox_book_state_t, reward_velocity) == sizeof(uintptr_t), "book reward velocity offset");
_Static_assert(sizeof(nox_book_state_t) == sizeof(uintptr_t) + 22 * sizeof(uint32_t), "book state size");
nox_book_state_t* nox_book_state_storage(void);

typedef struct nox_book_ui_slots {
	nox_window* list_window;
	nox_video_bag_image_t* arrows[8];
	nox_video_bag_image_t* background;
	nox_video_bag_image_t* guide_tab_lit;
	nox_video_bag_image_t* spell_tab_lit;
} nox_book_ui_slots;
_Static_assert(sizeof(nox_book_ui_slots) == 12 * sizeof(void*), "book UI slots size");
_Static_assert(offsetof(nox_book_ui_slots, background) == 9 * sizeof(void*), "book background offset");
nox_book_ui_slots* nox_book_ui_storage(void);

typedef struct {
	nox_window* left_arrow;
	nox_window* right_arrow;
	nox_window* page_icon;
	nox_window* reward;
	nox_things_imageRef_t* forward;
	nox_things_imageRef_t* backward;
} nox_book_details_t;
_Static_assert(sizeof(nox_book_details_t) == 6 * sizeof(void*),
	"book detail slots must contain six native pointers");
nox_book_details_t* nox_book_details_storage(void);
uint32_t nox_client_player_net_code(void);

// Public mixed pointer/tag ABIs stay in root-owned thin C adapters. The private
// Go exports return words, so Go never manufactures unsafe.Pointer(1/2/3/4).
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
