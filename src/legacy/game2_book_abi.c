#include "game2_book_native.h"

// Preserve the old public pointer/tag return ABIs without making a scalar into
// a Go pointer. These adapters contain no book algorithm and never dereference
// the returned word. Native addresses returned by Go retain their full width.
int nox_xxx_guiSpellSortFn_45ABC0(const void* a, const void* b) {
    return game2_book_compare((void*)a, (void*)b);
}
int* nox_xxx_bookSetForward_45D200(int* kind, int id, int2* pos) {
    return (int*)(uintptr_t)game2_book_forward_word((uintptr_t)kind, id, pos);
}
char* nox_xxx_clientQuestDisableAbility_45D4A0(int id) {
    return (char*)(uintptr_t)game2_book_disable_ability_word(id);
}
uint32_t* sub_45D550(uint32_t* pos) {
    return (uint32_t*)(uintptr_t)game2_book_position_word(pos);
}

void nox_xxx_abilityReward_45D290(int id, char* notify, int extra) {
    game2_book_ability_reward(id, (uintptr_t)notify, extra);
}
