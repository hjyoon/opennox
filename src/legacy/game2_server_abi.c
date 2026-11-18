#include "game2_server_bridge.h"

// Preserve the legacy pointer-shaped scalar result at the C boundary.
extern intptr_t nox_game2_server_apply_459AA0(void*);
char* sub_459AA0(void* p) { return (char*)nox_game2_server_apply_459AA0(p); }
