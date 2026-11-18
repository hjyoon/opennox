#include "GAME1_3.h"

// Keep historical mixed pointer/status returns at the C edge.
extern uintptr_t sub_446490_go(int);
extern uintptr_t nox_xxx_motdAddSomeTextMB_446730_go(uint8_t*);

void* sub_446490(int index) { return (void*)sub_446490_go(index); }

uint8_t* nox_xxx_motdAddSomeTextMB_446730(uint8_t* input) {
	return (uint8_t*)nox_xxx_motdAddSomeTextMB_446730_go(input);
}
