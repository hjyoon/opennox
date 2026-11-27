package legacy

/*
#include "client__gui__window.h"
#include "cgo_const.h"
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
func nox_new_window_from_file(cname *C.nox_cgo_const_char, fnc unsafe.Pointer) *nox_window {
	return loadWindowFromFile(GoString(cname), fnc)
}

func loadWindowFromFile(name string, fnc unsafe.Pointer) *nox_window {
	win := Nox_new_window_from_file(name, gui.WrapFuncC(fnc))
	if win == nil {
		return nil
	}
	if fnc != nil {
		win.SetFunc94C(fnc)
	}
	return (*nox_window)(win.C())
}
