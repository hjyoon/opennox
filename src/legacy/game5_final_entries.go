package legacy

/*
#include "GAME5_2.h"
#include "client__gui__guiinv.h"
*/
import "C"

//export nox_xxx_j_inventoryNameSignInit_467460
func nox_xxx_j_inventoryNameSignInit_467460() C.int {
	return nox_xxx_inventoryNameSignInit_4671E0()
}

//export nox_xxx_j_allocHitArray_511840
func nox_xxx_j_allocHitArray_511840() {
	nox_xxx_allocHitArray_5486D0()
}

//export nullsub_8
func nullsub_8(a1, a2 C.int) C.int { return 0 }

//export nullsub_9
func nullsub_9(a1 C.uint32_t) {}

//export nullsub_10
func nullsub_10(a1 C.uint32_t) {}

//export nullsub_28
func nullsub_28(a1 C.uint32_t) {}

//export nullsub_29
func nullsub_29() {}

//export nullsub_30
func nullsub_30(a1 C.uint32_t) {}

//export nullsub_31
func nullsub_31(a1 C.uint32_t) {}

//export nullsub_35
func nullsub_35(a1, a2 C.uint32_t) {}
