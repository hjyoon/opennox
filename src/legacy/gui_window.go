package legacy

/*
#include <stdlib.h>
#include "client__gui__window.h"

static int nox_window_call_draw_func_go(int (*fnc)(nox_window*, nox_window_data*), nox_window* win, nox_window_data* data) {
	return fnc(win, data);
}
static uintptr_t nox_window_call_func_go(uintptr_t (*fnc)(nox_window*, int, uintptr_t, uintptr_t), nox_window* win, int a2, uintptr_t a3, uintptr_t a4) {
	return fnc(win, a2, a3, a4);
}
static void nox_window_call_tooltip_func(nox_window* win, nox_window_data* data, int a3) {
	if (!win || !win->tooltip_func)
		return;
	win->tooltip_func(win, data, a3);
}
*/
import "C"
import (
	"image"
	"log/slog"
	"runtime/debug"
	"unsafe"

	noxcolor "github.com/opennox/libs/color"

	"github.com/opennox/opennox/v1/client/gui"
)

type nox_window = C.nox_window

func AsWindowP(win unsafe.Pointer) *gui.Window {
	w := (*gui.Window)(win)
	if false && cgoSafe && w.ID() == DeadWord {
		slog.Error("memory corruption detected")
		debug.PrintStack()
		C.abort()
	}
	return w
}

func asWindow(win *nox_window) *gui.Window {
	return AsWindowP(unsafe.Pointer(win))
}

//export get_dword_5d4594_3799468
func get_dword_5d4594_3799468() int32 {
	return int32(GetClient().Cli().GUI.ValXXX)
}

//export set_dword_5d4594_3799468
func set_dword_5d4594_3799468(v_cgo int32) {
	v := int(v_cgo)
	GetClient().Cli().GUI.ValYYY = v
}

//export nox_window_new_go
func nox_window_new_go(par *nox_window, flags, a3, a4, w, h C.int, fnc unsafe.Pointer) *nox_window {
	return (*nox_window)(GetClient().Cli().GUI.NewWindowRaw(asWindow(par), gui.StatusFlags(flags), int(a3), int(a4), int(w), int(h), gui.WrapFuncC(fnc)).C())
}

//export nox_xxx_wndGetID_46B0A0
func nox_xxx_wndGetID_46B0A0(win *nox_window) int32 {
	if win == nil {
		return int32(-2)
	}
	return int32(int(asWindow(win).ID()))
}

//export nox_xxx_wndSetID_46B080
func nox_xxx_wndSetID_46B080(win *nox_window, id_cgo int32) int32 {
	id := int(id_cgo)
	if win == nil {
		return int32(-2)
	}
	asWindow(win).SetID(uint(id))
	return int32(0)
}

//export nox_window_set_all_funcs_go
func nox_window_set_all_funcs_go(p *nox_window, a2 unsafe.Pointer, draw unsafe.Pointer, tooltip unsafe.Pointer) int32 {
	if p == nil {
		return int32(-2)
	}
	win := asWindow(p)
	win.SetFunc93(gui.WrapFuncC(a2))
	win.SetDraw(gui.WrapDrawFuncC(draw))
	win.SetTooltipFunc(tooltip)
	return int32(0)
}

//export nox_xxx_wndSetWindowProc_46B300_go
func nox_xxx_wndSetWindowProc_46B300_go(win *nox_window, fnc unsafe.Pointer) int32 {
	if win == nil {
		return int32(-2)
	}
	asWindow(win).SetFunc93(gui.WrapFuncC(fnc))
	return int32(0)
}

//export nox_xxx_wndSetProc_46B2C0_go
func nox_xxx_wndSetProc_46B2C0_go(win *nox_window, fnc unsafe.Pointer) int32 {
	if win == nil {
		return int32(-2)
	}
	asWindow(win).SetFunc94(gui.WrapFuncC(fnc))
	return int32(0)
}

//export nox_xxx_wndSetDrawFn_46B340_go
func nox_xxx_wndSetDrawFn_46B340_go(win *nox_window, fnc unsafe.Pointer) int32 {
	if win == nil {
		return int32(-2)
	}
	asWindow(win).SetDraw(gui.WrapDrawFuncC(fnc))
	return int32(0)
}

//export nox_gui_winSetFunc96_46B070
func nox_gui_winSetFunc96_46B070(win *nox_window, fnc unsafe.Pointer) {
	asWindow(win).SetTooltipFunc(fnc)
}

//export nox_xxx_wndSetRectColor2MB_46AFE0
func nox_xxx_wndSetRectColor2MB_46AFE0(win *nox_window, a2_cgo int32) int32 {
	a2 := int(a2_cgo)
	if win == nil {
		return int32(-2)
	}
	// TODO: not sure if the color format is correct
	asWindow(win).DrawData().SetBackgroundColor(noxcolor.RGBA5551(a2))
	return int32(0)
}

//export nox_window_call_field_94_fnc
func nox_window_call_field_94_fnc(p *nox_window, a2_cgo int32, a3, a4 uintptr, file *C.char, line_cgo int32) uintptr {
	a2 := int(a2_cgo)
	line := int(line_cgo)
	if p == nil {
		return 0
	}
	if guiDebug {
		guiLog.Printf("nox_window_call_field_94(%p, %x, %x, %x): %s:%d", p, a2, a3, a4, GoString(file), line)
	}
	r := asWindow(p).Func94(gui.AsWindowEvent(a2, a3, a4))
	if r == nil {
		return 0
	}
	return r.EventRespC()
}

//export nox_window_call_field_93
func nox_window_call_field_93(p *nox_window, a2_cgo int32, a3, a4 uintptr) uintptr {
	a2 := int(a2_cgo)
	if p == nil {
		return 0
	}
	r := asWindow(p).Func93(gui.AsWindowEvent(a2, a3, a4))
	if r == nil {
		return 0
	}
	return r.EventRespC()
}

//export nox_xxx_wndGetChildByID_46B0C0
func nox_xxx_wndGetChildByID_46B0C0(root *nox_window, id_cgo int32) *nox_window {
	id := int(id_cgo)
	return (*nox_window)(asWindow(root).ChildByID(uint(id)).C())
}

//export nox_xxx_windowDestroyMB_46C4E0
func nox_xxx_windowDestroyMB_46C4E0(a1 *nox_window) int32 {
	win := asWindow(a1)
	if win == nil {
		return int32(-2)
	}
	win.Destroy()
	return int32(0)
}

//export nox_window_set_hidden
func nox_window_set_hidden(p *nox_window, hidden_cgo int32) int32 {
	hidden := int(hidden_cgo)
	if p == nil {
		return int32(-2)
	}
	win := asWindow(p)
	if hidden != 0 {
		win.Hide()
	} else {
		win.Show()
	}
	return int32(0)
}

//export nox_xxx_wndShowModalMB_46A8C0
func nox_xxx_wndShowModalMB_46A8C0(p *nox_window) int32 {
	return int32(asWindow(p).ShowModal())
}

//export nox_window_setPos_46A9B0
func nox_window_setPos_46A9B0(p *nox_window, x_cgo, y_cgo int32) int32 {
	x := int(x_cgo)
	y := int(y_cgo)
	win := asWindow(p)
	if win == nil {
		return int32(-2)
	}
	win.SetPos(image.Point{X: x, Y: y})
	return int32(0)
}

//export wndIsShown_nox_xxx_wndIsShown_46ACC0
func wndIsShown_nox_xxx_wndIsShown_46ACC0(p *nox_window) int32 {
	if p == nil {
		return int32(1)
	}
	win := asWindow(p)
	is := win.GetFlags().IsHidden()
	return int32(bool2int(is))
}

//export nox_xxx_wnd_46C6E0
func nox_xxx_wnd_46C6E0(p *nox_window) int32 {
	return int32(asWindow(p).StackPop())
}

//export sub_46C690
func sub_46C690(p *nox_window) int32 {
	return int32(asWindow(p).StackPush())
}

func Nox_xxx_wnd_46ABB0(p *gui.Window, v int) int {
	return int(C.nox_xxx_wnd_46ABB0((*nox_window)(p.C()), C.int(v)))
}
