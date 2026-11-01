package legacy

/*
#include "GAME2_2.h"
*/
import "C"

import (
	"github.com/opennox/opennox/v1/legacy/timer"
	"unsafe"
)

//export sub_486640
func sub_486640(group unsafe.Pointer, rate C.int) C.int {
	return C.int(timer.ScalePercentageNative486640((*timer.TimerGroup)(group), int32(rate)))
}
