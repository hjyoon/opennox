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
	Nox_new_window_from_file func(name string, fnc gui.WindowFunc) *gui.Window
)

//export nox_new_window_from_file
func nox_new_window_from_file(cname *C.char, fnc unsafe.Pointer) *nox_window {
	win := Nox_new_window_from_file(GoString(cname), gui.WrapFuncC(fnc))
	if win == nil {
		return nil
	}
	if fnc != nil {
		win.SetFunc94C(fnc)
	}
	return (*nox_window)(win.C())
}
