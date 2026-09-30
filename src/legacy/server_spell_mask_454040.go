package legacy

/*
#include <string.h>
#include "GAME2.h"

static uintptr_t nox_server_spell_mask_stack_call_454040(uint32_t* out) {
	struct {
		uint32_t before;
		uint32_t mask[5];
		uint32_t after;
	} snapshot = { .before = 0xa1b2c3d4u, .after = 0x5e6f7081u };
	sub_454040(snapshot.mask);
	out[0] = snapshot.before;
	memcpy(out + 1, snapshot.mask, sizeof(snapshot.mask));
	out[6] = snapshot.after;
	return (uintptr_t)snapshot.mask;
}
*/
import "C"

import "unsafe"

func serverSpellMaskCall454040(mask *uint32) {
	C.sub_454040((*C.uint32_t)(unsafe.Pointer(mask)))
}

func serverSpellMaskStackCall454040() (words [7]uint32, address uintptr) {
	address = uintptr(C.nox_server_spell_mask_stack_call_454040((*C.uint32_t)(unsafe.Pointer(&words[0]))))
	return words, address
}
