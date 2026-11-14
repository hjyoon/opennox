#include "GAME1_3.h"

// Keep const-qualified inputs and the historical mixed pointer/status return
// types at the C edge. Algorithms and retained pointers live in Go.
extern int nox_sprintAddrPort_43BC80_go(char*, unsigned short, char*);
extern uintptr_t sub_446490_go(int);
extern uintptr_t nox_xxx_motdAddSomeTextMB_446730_go(uint8_t*);
extern void sub_4417E0_go(wchar2_t*, char*);

int nox_sprintAddrPort_43BC80(const char* addr, unsigned short port, char* dst) {
	return nox_sprintAddrPort_43BC80_go((char*)addr, port, dst);
}

void* sub_446490(int index) { return (void*)sub_446490_go(index); }

void sub_4417E0(wchar2_t* dst, const char* name) { sub_4417E0_go(dst, (char*)name); }

uint8_t* nox_xxx_motdAddSomeTextMB_446730(uint8_t* input) {
	return (uint8_t*)nox_xxx_motdAddSomeTextMB_446730_go(input);
}
