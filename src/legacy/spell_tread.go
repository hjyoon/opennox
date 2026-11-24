package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import "github.com/opennox/opennox/v1/common/memmap"

func setMonsterSightFrameCopy(increment uint32) C.int {
	frame := gameFrame() + increment
	*memmap.PtrUint32(0x5D4594, 2487684) = gameFrame() + increment
	return C.int(int32(frame))
}

//export nox_xxx_frameCounterSetCopyToNextFrame_5281D0
func nox_xxx_frameCounterSetCopyToNextFrame_5281D0() C.int {
	return setMonsterSightFrameCopy(1)
}

//export nox_xxx_frameCounterSetCopy_5281E0
func nox_xxx_frameCounterSetCopy_5281E0() C.int {
	return setMonsterSightFrameCopy(0)
}

func Nox_xxx_frameCounterSetCopyToNextFrame_5281D0() {
	nox_xxx_frameCounterSetCopyToNextFrame_5281D0()
}
