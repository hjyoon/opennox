package legacy

/*
#include <stdint.h>

#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME3.h"

typedef int (*nox_test_widget_event_callback_t)(nox_window*, int, uintptr_t, uintptr_t);
typedef uintptr_t (*nox_test_widget_pre_callback_t)(nox_window*, unsigned int, uintptr_t, uintptr_t);

_Static_assert(_Generic(&nox_xxx_wndEditProc_487D70, nox_test_widget_event_callback_t: 1, default: 0),
	"edit event callback must use native-width arguments");
_Static_assert(_Generic(&sub_46A7E0, nox_test_widget_event_callback_t: 1, default: 0),
	"chat edit event callback must use native-width arguments");
_Static_assert(_Generic(&nox_xxx_wndListboxProcWithoutData10_4A28E0, nox_test_widget_event_callback_t: 1, default: 0),
	"single-select list-box callback must use native-width arguments");
_Static_assert(_Generic(&nox_xxx_wndListboxProcWithData10_4A2DE0, nox_test_widget_event_callback_t: 1, default: 0),
	"multi-select list-box callback must use native-width arguments");
_Static_assert(_Generic(&nox_xxx_wndListboxProcPre_4A30D0, nox_test_widget_pre_callback_t: 1, default: 0),
	"list-box control callback must use native-width arguments");

static int nox_test_edit_event_callback(nox_window* win, int event, uintptr_t a3, uintptr_t a4) {
	return nox_xxx_wndEditProc_487D70(win, event, a3, a4);
}

static int nox_test_listbox_event_callback(nox_window* win, int multi, int event, uintptr_t a3, uintptr_t a4) {
	if (multi) {
		return nox_xxx_wndListboxProcWithData10_4A2DE0(win, event, a3, a4);
	}
	return nox_xxx_wndListboxProcWithoutData10_4A28E0(win, event, a3, a4);
}

static uintptr_t nox_test_listbox_pre_callback(nox_window* win, unsigned int event, uintptr_t a3, uintptr_t a4) {
	return nox_xxx_wndListboxProcPre_4A30D0(win, event, a3, a4);
}

static void nox_test_listbox_init_callback(nox_window* win) {
	nox_xxx_wndListboxInit_4A3C00(win, (nox_scrollListBox_data*)win->widget_data);
}
*/
import "C"

import "github.com/opennox/opennox/v1/client/gui"

func editEventCallbackC(win *gui.Window, event int, a3, a4 uintptr) int {
	return int(C.nox_test_edit_event_callback((*C.nox_window)(win.C()), C.int(event), C.uintptr_t(a3), C.uintptr_t(a4)))
}

func listBoxEventCallbackC(win *gui.Window, multi bool, event int, a3, a4 uintptr) int {
	var cmulti C.int
	if multi {
		cmulti = 1
	}
	return int(C.nox_test_listbox_event_callback((*C.nox_window)(win.C()), cmulti, C.int(event), C.uintptr_t(a3), C.uintptr_t(a4)))
}

func listBoxPreCallbackC(win *gui.Window, event int, a3, a4 uintptr) uintptr {
	return uintptr(C.nox_test_listbox_pre_callback((*C.nox_window)(win.C()), C.uint(event), C.uintptr_t(a3), C.uintptr_t(a4)))
}

func listBoxInitCallbackC(win *gui.Window) {
	C.nox_test_listbox_init_callback((*C.nox_window)(win.C()))
}
