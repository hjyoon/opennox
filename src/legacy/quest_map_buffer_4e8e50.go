package legacy

/*
#include "quest_map_buffer_4e8e50.h"
*/
import "C"

import "github.com/opennox/opennox/v1/common/memmap"

//export sub_4E8E50
func sub_4E8E50() *C.char {
	return (*C.char)(memmap.PtrOff(0x5D4594, 1567844))
}
