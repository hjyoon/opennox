package legacy

/*
#include "GAME1_1.h"
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
	globalRecordStore(unsafe.Pointer(source))
}

func globalRecordStore(source unsafe.Pointer) {
	if source == nil {
		return
	}
	text := GoString((*C.char)(source))
	destination := memmap.Slice(0x5D4594, 534756)
	copy(destination, text)
	destination[len(text)] = 0
}

//export sub_41FBE0
func sub_41FBE0(first, second **C.global_record_const_char_t) C.int {
	return C.int(globalRecordOutput(unsafe.Pointer(first), unsafe.Pointer(second)))
}

func globalRecordOutput(first, second unsafe.Pointer) int32 {
	if uint32(worldWordStorage.global_record_word_60044) == ^uint32(0) {
		return 0
	}
	offset := uint32(531660) + 24*uint32(worldWordStorage.global_record_word_60044)
	// These C-owned output slots may be uninitialized. A Go pointer write
	// barrier would interpret their old bytes as Go heap references.
	*(*uintptr)(first) = uintptr(memmap.PtrOff(0x5D4594, uintptr(offset)))
	// The stock entry reloads the selected index after its first store.
	offset = uint32(531670) + 24*uint32(worldWordStorage.global_record_word_60044)
	*(*uintptr)(second) = uintptr(memmap.PtrOff(0x5D4594, uintptr(offset)))
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
