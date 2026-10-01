package legacy

/*
#include "GAME2_1.h"
*/
import "C"

import "github.com/opennox/opennox/v1/common/memmap"

//export sub_473960
func sub_473960() C.int {
	*memmap.PtrPtr(0x5D4594, 1096456) = nil
	*memmap.PtrPtr(0x5D4594, 1096460) = nil
	return 0
}
