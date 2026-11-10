#include "game2_server_bridge.h"
// Only const qualification / historical scalar-token ABI conversion lives here.
extern void nox_game2_server_copySpells_453F70(void*);
extern int nox_game2_server_spellBit_454000(void*,int);
extern int nox_game2_server_applySpells_4540E0(void*);
extern int nox_game2_server_limits_457460(void*);
extern intptr_t nox_game2_server_apply_459AA0(void*);
void sub_453F70(const void* p) { nox_game2_server_copySpells_453F70((void*)p); }
int sub_454000(const void* p,int bit) { return nox_game2_server_spellBit_454000((void*)p,bit); }
int sub_4540E0(const void* p) { return nox_game2_server_applySpells_4540E0((void*)p); }
int sub_457460(const char* p) { return nox_game2_server_limits_457460((void*)p); }
char* sub_459AA0(void* p) { return (char*)nox_game2_server_apply_459AA0(p); }
// Existing typed Go ambient implementation; no second algorithm port claimed.
void sub_45A9B0(nox_drawable* source,nox_drawable* listener) { nox_client_updateAmbient_native_45A9B0(source,listener); }
