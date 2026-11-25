package legacy

/*
#include "GAME2_1.h"
typedef const wchar2_t nox_const_window_text_unit;
typedef const void* nox_const_window_draw_data;
*/
import "C"

import (
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
)

const clientWindowDataSource = "client_window_data_native.go"

//export sub_46AEE0
func sub_46AEE0(win *C.nox_window, text *C.nox_const_window_text_unit) C.int {
	nox_window_call_field_94_fnc(win, 16385, uintptr(unsafe.Pointer(text)), 0, internCStr(clientWindowDataSource), 0)
	runtime.KeepAlive(text)
	return 0
}

//export sub_46AF00
func sub_46AF00(win *C.nox_window) *C.wchar2_t {
	if win == nil {
		return nil
	}
	var event int32
	switch style := win.draw_data.style; {
	case style&0x800 != 0:
		event = 16386
	case style&0x80 != 0:
		event = 16413
	default:
		return nil
	}
	result := nox_window_call_field_94_fnc(win, event, 0, 0, internCStr(clientWindowDataSource), 0)
	return (*C.wchar2_t)(unsafe.Pointer(result))
}

//export nox_gui_windowCopyDrawData_46AF80
func nox_gui_windowCopyDrawData_46AF80(win *C.nox_window, data C.nox_const_window_draw_data) C.int {
	return C.int(asWindow(win).CopyDrawData((*gui.WindowData)(unsafe.Pointer(data))))
}
