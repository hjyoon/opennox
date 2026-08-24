package legacy

/*
#include "defs.h"
#include "client__gui__window.h"
*/
import "C"
import (
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
)

var (
	Nox_gui_makeAnimation_43C5B0 func(win *gui.Window, x1, y1, x2, y2, in_dx, in_dy, out_dx, out_dy int) *gui.Anim
)

var _ = [1]struct{}{}[unsafe.Sizeof(C.nox_gui_animation{})-unsafe.Sizeof(gui.Anim{})]

func asGUIAnim(p *C.nox_gui_animation) *gui.Anim {
	return (*gui.Anim)(unsafe.Pointer(p))
}

//export nox_gui_freeAnimation_43C570
func nox_gui_freeAnimation_43C570(a *C.nox_gui_animation) {
	asGUIAnim(a).Free()
}

//export nox_gui_makeAnimation_43C5B0
func nox_gui_makeAnimation_43C5B0(win *nox_window, x1_cgo, y1_cgo, x2_cgo, y2_cgo, in_dx_cgo, in_dy_cgo, out_dx_cgo, out_dy_cgo int32) *C.nox_gui_animation {
	x1 := int(x1_cgo)
	y1 := int(y1_cgo)
	x2 := int(x2_cgo)
	y2 := int(y2_cgo)
	in_dx := int(in_dx_cgo)
	in_dy := int(in_dy_cgo)
	out_dx := int(out_dx_cgo)
	out_dy := int(out_dy_cgo)
	a := Nox_gui_makeAnimation_43C5B0(asWindow(win), x1, y1, x2, y2, in_dx, in_dy, out_dx, out_dy)
	return (*C.nox_gui_animation)(unsafe.Pointer(a))
}
