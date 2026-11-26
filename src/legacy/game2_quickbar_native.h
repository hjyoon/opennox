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
uintptr_t nox_game2_book_insert_native(uintptr_t kind, int spell, int slot);
int nox_client_localPlayerHasFlags30(uint32_t flags);
#endif
