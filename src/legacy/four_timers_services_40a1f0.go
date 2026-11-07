package legacy

/*
#include "GAME3_2.h"
#include <stdint.h>
_Static_assert(sizeof(int)==4 && sizeof(long long)==8, "timer public ABI widths");
typedef int (*timer_notify_native_fn_40a1f0)(int,int);
_Static_assert(_Generic(&nox_xxx_netTimerStatus_4D8F50,timer_notify_native_fn_40a1f0:1,default:0), "actual raw Notify int(int,int)");
*/
import "C"

func timerNativeNotify40A1F0(recipient, active int32) int32 {
	return int32(C.nox_xxx_netTimerStatus_4D8F50(C.int(recipient), C.int(active)))
}
