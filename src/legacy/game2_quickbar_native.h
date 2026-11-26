#ifndef NOX_GAME2_QUICKBAR_NATIVE_H
#define NOX_GAME2_QUICKBAR_NATIVE_H
#include "client__gui__window.h"
#include <stddef.h>
#include <stdint.h>
typedef struct nox_quickbar_window_set {
    void* data;
    nox_window* root;
    nox_window* buttons[5];
    nox_window* nuggets[5];
    nox_window* nugget_children[5];
} nox_quickbar_window_set;
_Static_assert(sizeof(nox_quickbar_window_set) == 17 * sizeof(void*), "quickbar window set size");
_Static_assert(offsetof(nox_quickbar_window_set, nugget_children) == 12 * sizeof(void*),
               "quickbar window set children offset");
nox_quickbar_window_set* nox_quickbar_window_storage(void);

typedef struct {
	nox_window* trap_construct;
	nox_window* right;
	nox_window* left;
	nox_window* title;
	nox_window* drag_window;
	nox_window* trap_activate;
	nox_window* book_button;
	nox_window* dragged_slot;
	void* drag_source_data;
	void* drag_source_row;
	void* font;
	uint32_t expanded;
	uint32_t trap_visible;
	uint32_t targeting_ability;
	uint32_t target_ability_id;
} nox_quickbar_controls_t;
_Static_assert(offsetof(nox_quickbar_controls_t, expanded) == 11 * sizeof(void*),
	"quickbar state word offset");
_Static_assert(sizeof(nox_quickbar_controls_t) == 11 * sizeof(void*) + 4 * sizeof(uint32_t),
	"quickbar controls size");
nox_quickbar_controls_t* nox_quickbar_controls_storage(void);
uintptr_t nox_game2_book_insert_native(uintptr_t kind, int spell, int slot);
int nox_client_localPlayerHasFlags30(uint32_t flags);
#endif
