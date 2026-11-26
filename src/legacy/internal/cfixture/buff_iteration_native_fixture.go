package cfixture

/*
#include "world_native_storage.h"
#define dword_587000_66116 (nox_world_word_storage()->buff_iteration_count)
#include <stddef.h>
#include "GAME1_1.h"
_Static_assert(_Generic(&sub_424CB0, char (*)(nox_object_t*): 1, default: 0), "native buff counter ABI");
_Static_assert(_Generic(&sub_424D00, int (*)(void): 1, default: 0), "buff first ABI");
_Static_assert(_Generic(&sub_424D20, int (*)(int): 1, default: 0), "buff next token ABI");
static unsigned char buff_iteration_count(nox_object_t* unit) { return (unsigned char)sub_424CB0(unit); }
static int buff_iteration_first(void) { return sub_424D00(); }
static int buff_iteration_next(int buff) { return sub_424D20(buff); }
static size_t buff_iteration_buffs_offset(void) { return offsetof(nox_object_t, buffs); }
static uint32_t* buff_iteration_count_owner(void) { return &dword_587000_66116; }
static uint8_t* buff_iteration_table_owner(void) { return nox_legacy_memory_blob(0x587000, NULL) + 66000; }
static uint8_t* buff_iteration_blob_count_owner(void) { return nox_legacy_memory_blob(0x587000, NULL) + 66116; }
*/
import "C"

import "unsafe"

func BuffIterationLayoutC() (uintptr, uintptr) {
	return C.sizeof_nox_object_t, uintptr(C.buff_iteration_buffs_offset())
}

func BuffIterationCountC(unit unsafe.Pointer) byte {
	return byte(C.buff_iteration_count((*C.nox_object_t)(unit)))
}

func BuffIterationFirstC() uint32 { return uint32(C.buff_iteration_first()) }
func BuffIterationNextC(buff uint32) uint32 {
	return uint32(C.buff_iteration_next(C.int(int32(buff))))
}

func BuffIterationOwnersC() (*uint32, []byte, []byte) {
	return (*uint32)(unsafe.Pointer(C.buff_iteration_count_owner())),
		unsafe.Slice((*byte)(unsafe.Pointer(C.buff_iteration_table_owner())), 116),
		unsafe.Slice((*byte)(unsafe.Pointer(C.buff_iteration_blob_count_owner())), 4)
}
