#include "GAME2_3.h"

extern uintptr_t nox_client_chat_candidate_word(char direction, uint32_t* moving, uint32_t* earlier, int* x, int* y);

int* sub_48E6A0(char direction, uint32_t* moving, uint32_t* earlier, int* x, int* y) {
	return (int*)nox_client_chat_candidate_word(direction, moving, earlier, x, y);
}
