#include "GAME2.h"
#include "game2_quickbar_native.h"
/* Scalar kind/result tokens cross Go as uintptr_t, never fake Go pointers. */
void* nox_xxx_book_45DBE0(void* kind, int spell, int slot) {
    return (void*)nox_game2_book_insert_native((uintptr_t)kind, spell, slot);
}
