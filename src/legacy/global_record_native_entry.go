package legacy

/*
#include "GAME1_1.h"
extern uint32_t dword_587000_60044;
typedef const char global_record_const_char_t;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

//export sub_41FA40
func sub_41FA40() *C.char {
	return (*C.char)(memmap.PtrOff(0x5D4594, 534756))
}

//export sub_41FA50
func sub_41FA50(source *C.global_record_const_char_t) {
	if source == nil {
		return
	}
	text := GoString((*C.char)(unsafe.Pointer(source)))
	destination := memmap.Slice(0x5D4594, 534756)
	copy(destination, text)
	destination[len(text)] = 0
}

//export sub_41FBE0
func sub_41FBE0(first, second **C.global_record_const_char_t) C.int {
	if uint32(C.dword_587000_60044) == ^uint32(0) {
		return 0
	}
	offset := uint32(531660) + 24*uint32(C.dword_587000_60044)
	*first = (*C.global_record_const_char_t)(memmap.PtrOff(0x5D4594, uintptr(offset)))
	// The stock entry reloads the selected index after its first store.
	offset = uint32(531670) + 24*uint32(C.dword_587000_60044)
	*second = (*C.global_record_const_char_t)(memmap.PtrOff(0x5D4594, uintptr(offset)))
	return 1
}

//export sub_4200E0
func sub_4200E0() C.int {
	return C.int(int32(memmap.Uint32(0x587000, 60072)))
}

//export sub_420100
func sub_420100() C.int {
	return C.int(int32(memmap.Uint32(0x587000, 60072)) >> 8)
}
