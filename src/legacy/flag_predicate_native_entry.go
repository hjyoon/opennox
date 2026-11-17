package legacy

/*
#include "client__gui__window.h"
#include "GAME5_2.h"
*/
import "C"

//export nox_xxx_wndGetFlags_46ADA0
func nox_xxx_wndGetFlags_46ADA0(win *C.nox_window) C.int {
	return C.int(windowFlags46ADA0(asWindow(win)))
}

//export sub_579EE0
func sub_579EE0(wp *C.nox_waypoint_t, mask C.uchar) C.int {
	return C.int(bool2int(waypointFlagMask579EE0(asWaypointS(wp), byte(mask))))
}
