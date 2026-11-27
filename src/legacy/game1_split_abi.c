#include "GAME1_1.h"
#include "GAME1_3.h"
#include "stats_bulk_native.h"

// Keep historical mixed pointer/status returns at the C edge.
extern uintptr_t sub_446490_go(int);
extern uintptr_t nox_xxx_motdAddSomeTextMB_446730_go(uint8_t*);
extern uintptr_t nox_stats_bulk_flush_word(void);

void* sub_446490(int index) { return (void*)sub_446490_go(index); }

uint8_t* nox_xxx_motdAddSomeTextMB_446730(uint8_t* input) {
	return (uint8_t*)nox_xxx_motdAddSomeTextMB_446730_go(input);
}

char* nox_stats_bulk_flush_native(void) {
	return (char*)nox_stats_bulk_flush_word();
}

char* nox_xxx_net_4263C0(void) {
	return nox_stats_bulk_flush_native();
}
