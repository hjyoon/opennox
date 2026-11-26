package legacy

/*
#include <stdlib.h>
#include "client__gui__window.h"
*/
import "C"
import (
	"image"
	"log/slog"
	"runtime/debug"
	"sync"
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

// Store a GUI-thread reference until its window is returned to the pool.
func storeWindowReference(slot **nox_window, win *nox_window) {
	*slot = win
	asWindow(win).AddCleanup(func() {
		if *slot == win {
			*slot = nil
		}
	})
}

var (
	legacyWindowRecoveryWarning sync.Once
	legacyWindowRejectWarning   sync.Once
)

func resolveLegacyWindow(p *nox_window, field string, file *C.char, line int) *gui.Window {
	if p == nil {
		return nil
	}
	addr := uintptr(unsafe.Pointer(p))
	win := GetClient().Cli().GUI.ResolveLegacyWindow(addr)
	if win == nil {
		legacyWindowRejectWarning.Do(func() {
			source := "<unknown>"
			if file != nil {
				source = GoString(file)
			}
			guiLog.Printf("rejected invalid legacy GUI window pointer %p at %s %s:%d", p, field, source, line)
		})
		return nil
	}
	if uintptr(unsafe.Pointer(win)) != addr {
		legacyWindowRecoveryWarning.Do(func() {
			source := "<unknown>"
			if file != nil {
				source = GoString(file)
			}
			guiLog.Printf("recovered truncated legacy GUI window pointer %p as %p at %s %s:%d", p, win, field, source, line)
		})
	}
	return win
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

//export nox_window_new
func nox_window_new(par *nox_window, flags, a3, a4, w, h C.int, fnc C.nox_window_init_proc) *nox_window {
	win := GetClient().Cli().GUI.NewWindowRaw(asWindow(par), gui.StatusFlags(flags), int(a3), int(a4), int(w), int(h), nil)
	if fnc != nil {
		win.SetInitializerC(unsafe.Pointer(fnc))
		win.Func94(gui.WindowInit{})
	}
	return (*nox_window)(win.C())
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

//export nox_window_set_all_funcs
func nox_window_set_all_funcs(p *nox_window, a2 unsafe.Pointer, draw unsafe.Pointer, tooltip unsafe.Pointer) C.int {
	if p == nil {
		return -2
	}
	win := asWindow(p)
	win.SetFunc93C(a2)
	win.SetDrawC(draw)
	win.SetTooltipFunc(tooltip)
	return 0
}

//export nox_xxx_wndSetWindowProc_46B300
func nox_xxx_wndSetWindowProc_46B300(win *nox_window, fnc unsafe.Pointer) C.int {
	if win == nil {
		return -2
	}
	asWindow(win).SetFunc93C(fnc)
	return 0
}

//export nox_xxx_wndSetProc_46B2C0
func nox_xxx_wndSetProc_46B2C0(win *nox_window, fnc unsafe.Pointer) C.int {
	if win == nil {
		return -2
	}
	asWindow(win).SetFunc94C(fnc)
	return 0
}

//export nox_xxx_wndSetDrawFn_46B340
func nox_xxx_wndSetDrawFn_46B340(win *nox_window, fnc unsafe.Pointer) C.int {
	if win == nil {
		return -2
	}
	asWindow(win).SetDrawC(fnc)
	return 0
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
	win := resolveLegacyWindow(p, "field94", file, line)
	if win == nil {
		return 0
	}
	if guiDebug {
		guiLog.Printf("nox_window_call_field_94(%p, %x, %x, %x): %s:%d", p, a2, a3, a4, GoString(file), line)
	}
	r := win.Func94(gui.AsWindowEvent(a2, a3, a4))
	if r == nil {
		return 0
	}
	return r.EventRespC()
}

//export nox_window_call_field_93
func nox_window_call_field_93(p *nox_window, a2_cgo int32, a3, a4 uintptr) uintptr {
	a2 := int(a2_cgo)
	win := resolveLegacyWindow(p, "field93", nil, 0)
	if win == nil {
		return 0
	}
	r := win.Func93(gui.AsWindowEvent(a2, a3, a4))
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
	if !GetClient().Cli().GUI.OwnsWindow(win) {
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

//export nox_xxx_wnd_46ABB0_go
func nox_xxx_wnd_46ABB0_go(p *nox_window, enabled int32) int32 {
	if p == nil {
		return int32(-2)
	}
	return int32(asWindow(p).SetEnabled(enabled != 0))
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
	if p == nil {
		return -2
	}
	return p.SetEnabled(v != 0)
}
