package legacy

/*
int sub_486640(void*, int);
_Static_assert(_Generic(&sub_486640,int (*)(void*,int):1,default:0),"public pointer and two-word PE32 ABI");
*/
import "C"

import (
	"github.com/opennox/opennox/v1/legacy/timer"
	"unsafe"
)

func timerPercentageOriginalC486640(group *timer.TimerGroup, rate int32) int32 {
	return int32(C.sub_486640(unsafe.Pointer(group), C.int(rate)))
}
