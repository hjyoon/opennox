package legacy

/*
#include "client__gui__window.h"
*/
import "C"
import (
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
)

var (
	Sub_4C26F0 func(a1 *gui.Window) int
)

//export sub_4C26F0
func sub_4C26F0(win *nox_window, _ *C.nox_window_data) int32 {
	return int32(Sub_4C26F0(AsWindowP(unsafe.Pointer(win))))
}
