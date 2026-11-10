#include <stdint.h>
#include "GAME4.h"

extern uintptr_t nox_game4_player_gold_add_go(nox_object_t*, int);
extern uintptr_t nox_game4_player_gold_sub_go(nox_object_t*, unsigned int);
extern void nox_game4_reward_string_go(nox_object_t*, char*, char);

uint32_t* nox_xxx_playerAddGold_4FA590(nox_object_t* unit, int amount) {
	return (uint32_t*)nox_game4_player_gold_add_go(unit, amount);
}
uint32_t* nox_xxx_playerSubGold_4FA5D0(nox_object_t* unit, unsigned int amount) {
	return (uint32_t*)nox_game4_player_gold_sub_go(unit, amount);
}
void sub_4FADD0(nox_object_t* unit, const char* name, char flags) {
	nox_game4_reward_string_go(unit, (char*)name, flags);
}
