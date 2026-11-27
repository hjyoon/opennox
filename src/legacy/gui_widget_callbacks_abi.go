package legacy

/*
#include <stdint.h>

#include "GAME2.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME3.h"
#include "GAME3_1.h"
#include "client__gui__guiinput.h"
#include "client__shell__inputcfg__inputcfg.h"

typedef uintptr_t (*nox_test_widget_pre_callback_t)(nox_window*, unsigned int, uintptr_t, uintptr_t);
typedef int (*nox_test_widget_draw_callback_t)(nox_window*, nox_window_data*);
typedef nox_window* (*nox_test_window_constructor_t)(nox_window*, int, int, int, int, int,
	uintptr_t (*)(nox_window*, int, uintptr_t, uintptr_t));

_Static_assert(_Generic(&nox_window_new, nox_test_window_constructor_t: 1, default: 0),
	"window constructor must take a native-width callback");
_Static_assert(_Generic(&nox_xxx_wndProc_44E6E0, nox_window_init_proc: 1, default: 0),
	"briefing initializer must use the window constructor callback ABI");
_Static_assert(_Generic(&sub_466220, nox_window_init_proc: 1, default: 0),
	"inventory initializer must use the window constructor callback ABI");

_Static_assert(_Generic(&nox_xxx_wndEditProc_487D70, nox_window_proc: 1, default: 0),
	"edit event callback must use native-width arguments");
_Static_assert(_Generic(&nox_xxx_wndEditProcPre_488710, nox_test_widget_pre_callback_t: 1, default: 0),
	"edit control callback must return native-width arguments");
_Static_assert(_Generic(&nox_xxx_wndEditDrawNoImage_488160, nox_test_widget_draw_callback_t: 1, default: 0),
	"edit draw callback must use native-width arguments");
_Static_assert(_Generic(&sub_46A7E0, nox_window_proc: 1, default: 0),
	"chat edit event callback must use native-width arguments");
_Static_assert(_Generic(&nox_xxx_wndListboxProcWithoutData10_4A28E0, nox_window_proc: 1, default: 0),
	"single-select list-box callback must use native-width arguments");
_Static_assert(_Generic(&nox_xxx_wndListboxProcWithData10_4A2DE0, nox_window_proc: 1, default: 0),
	"multi-select list-box callback must use native-width arguments");
_Static_assert(_Generic(&nox_xxx_wndListboxProcPre_4A30D0, nox_test_widget_pre_callback_t: 1, default: 0),
	"list-box control callback must use native-width arguments");
_Static_assert(_Generic(&sub_4C3A60, nox_window_proc: 1, default: 0),
	"in-game input-config list callback must use native-width arguments");
_Static_assert(_Generic(&sub_4C3A90, nox_window_proc: 1, default: 0),
	"in-game input-config root callback must use native-width arguments");
_Static_assert(_Generic(&sub_4C3CD0, nox_window_proc: 1, default: 0),
	"in-game input-config control callback must use native-width arguments");
_Static_assert(_Generic(&sub_4C3EB0, nox_window_proc: 1, default: 0),
	"in-game input-config modal callback must use native-width arguments");
_Static_assert(_Generic(&sub_4CBE70, nox_window_proc: 1, default: 0),
	"shell input-config root callback must use native-width arguments");
_Static_assert(_Generic(&sub_4CBF60, nox_window_proc: 1, default: 0),
	"shell input-config control callback must use native-width arguments");
_Static_assert(_Generic(&sub_4CC140, nox_window_proc: 1, default: 0),
	"shell input-config list callback must use native-width arguments");
_Static_assert(_Generic(&sub_4CC170, nox_window_proc: 1, default: 0),
	"shell input-config modal callback must use native-width arguments");

static int nox_test_edit_event_callback(nox_window* win, int event, uintptr_t a3, uintptr_t a4) {
	return nox_xxx_wndEditProc_487D70(win, event, a3, a4);
}

static uintptr_t nox_test_edit_pre_callback(nox_window* win, unsigned int event, uintptr_t a3, uintptr_t a4) {
	return nox_xxx_wndEditProcPre_488710(win, event, a3, a4);
}

static int nox_test_edit_draw_callback(nox_window* win) {
	return nox_xxx_wndEditDrawNoImage_488160(win, &win->draw_data);
}

static nox_window* nox_test_edit_constructor(nox_window* parent, nox_window_data* draw, void* data) {
	return nox_gui_newEntryField_488500(parent, 8, 10, 20, 120, 20, draw, data);
}

static uintptr_t nox_test_window_initializer(nox_window* win, int event, uintptr_t a3, uintptr_t a4) {
	if (event == 1) {
		win->field_92++;
	}
	if (event == 0x5005 && a3 == (uintptr_t)win && a4 == ~a3) {
		return a3;
	}
	return 0;
}

static nox_window* nox_test_window_constructor(void) {
	return nox_window_new(NULL, 8, 10, 20, 120, 20, nox_test_window_initializer);
}

static int nox_test_window_signed_proc(nox_window* win, int event, uintptr_t a, uintptr_t b) {
	if (event == 0x5005 && a == (uintptr_t)win && b == ~a) {
		return -3;
	}
	return 0;
}

static int nox_test_window_set_signed_proc(nox_window* win) {
	nox_window_proc proc = nox_test_window_signed_proc;
	return nox_xxx_wndSetWindowProc_46B300(win, proc) == 0 &&
		nox_xxx_wndSetProc_46B2C0(win, proc) == 0 &&
		win->field_93 == (void*)proc && win->field_94 == (void*)proc;
}

static int nox_test_window_draw_proc(nox_window* win, nox_window_data* draw) {
	if (draw == &win->draw_data) {
		win->field_92 = 1;
	}
	return -7;
}

static int nox_test_window_tooltip_proc(nox_window* win, nox_window_data* draw, int pos) {
	if (draw == &win->draw_data && (uint32_t)pos == 0x80010003) {
		win->field_92 = 2;
	}
	return -9;
}

static int nox_test_window_set_draw_and_tooltip(nox_window* win) {
	nox_window_draw_proc draw = nox_test_window_draw_proc;
	nox_window_tooltip_proc tooltip = nox_test_window_tooltip_proc;
	return nox_window_set_all_funcs(win, NULL, draw, tooltip) == 0 &&
		win->draw_func == draw && win->tooltip_func == tooltip;
}

static int nox_test_window_null_setters(void) {
	return nox_window_set_all_funcs(NULL, NULL, NULL, NULL) == -2 &&
		nox_xxx_wndSetProc_46B2C0(NULL, NULL) == -2 &&
		nox_xxx_wndSetWindowProc_46B300(NULL, NULL) == -2 &&
		nox_xxx_wndSetDrawFn_46B340(NULL, NULL) == -2;
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

static int nox_test_input_config_list_callback(nox_window* win, int shell, int event,
		uintptr_t a3, uintptr_t a4) {
	if (shell) {
		return sub_4CC140(win, event, a3, a4);
	}
	return sub_4C3A60(win, event, a3, a4);
}

static int nox_test_input_config_control_callback(nox_window* win, int shell, int event,
		uintptr_t a3, uintptr_t a4) {
	if (shell) {
		return sub_4CBF60(win, event, a3, a4);
	}
	return sub_4C3CD0(win, event, a3, a4);
}

static int nox_test_input_config_root_callback(nox_window* win, int shell, int event,
		uintptr_t a3, uintptr_t a4) {
	if (shell) {
		return sub_4CBE70(win, event, a3, a4);
	}
	return sub_4C3A90(win, event, a3, a4);
}
*/
import "C"

import "github.com/opennox/opennox/v1/client/gui"

func editEventCallbackC(win *gui.Window, event int, a3, a4 uintptr) int {
	return int(C.nox_test_edit_event_callback((*C.nox_window)(win.C()), C.int(event), C.uintptr_t(a3), C.uintptr_t(a4)))
}

func editPreCallbackC(win *gui.Window, event int, a3, a4 uintptr) uintptr {
	return uintptr(C.nox_test_edit_pre_callback((*C.nox_window)(win.C()), C.uint(event), C.uintptr_t(a3), C.uintptr_t(a4)))
}

func editDrawCallbackC(win *gui.Window) int {
	return int(C.nox_test_edit_draw_callback((*C.nox_window)(win.C())))
}

func editConstructorC(parent *gui.Window, draw *gui.WindowData, data *gui.EntryFieldData) *gui.Window {
	return asWindow(C.nox_test_edit_constructor(
		(*C.nox_window)(parent.C()), (*C.nox_window_data)(draw.C()), data.CWidgetData(),
	))
}

func windowWithInitializerC() (*gui.Window, uint32) {
	win := C.nox_test_window_constructor()
	if win == nil {
		return nil, 0
	}
	return asWindow(win), uint32(win.field_92)
}

func windowNullSettersC() bool {
	return C.nox_test_window_null_setters() != 0
}

func windowChildByIDC(root *gui.Window, id int32) *gui.Window {
	return asWindow(C.nox_xxx_wndGetChildByID_46B0C0((*nox_window)(root.C()), C.int(id)))
}

func windowSetSignedProcC(win *gui.Window) bool {
	return C.nox_test_window_set_signed_proc((*nox_window)(win.C())) != 0
}

func windowSetDrawAndTooltipC(win *gui.Window) bool {
	return C.nox_test_window_set_draw_and_tooltip((*nox_window)(win.C())) != 0
}

func windowCallbackStateC(win *gui.Window) uint32 {
	return uint32((*nox_window)(win.C()).field_92)
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

func inputConfigListCallbackC(win *gui.Window, shell bool, event int, a3, a4 uintptr) int {
	var cshell C.int
	if shell {
		cshell = 1
	}
	return int(C.nox_test_input_config_list_callback(
		(*C.nox_window)(win.C()), cshell, C.int(event), C.uintptr_t(a3), C.uintptr_t(a4),
	))
}

func inputConfigControlCallbackC(win *gui.Window, shell bool, event int, a3, a4 uintptr) int {
	var cshell C.int
	if shell {
		cshell = 1
	}
	return int(C.nox_test_input_config_control_callback(
		(*C.nox_window)(win.C()), cshell, C.int(event), C.uintptr_t(a3), C.uintptr_t(a4),
	))
}

func inputConfigRootCallbackC(win *gui.Window, shell bool, event int, a3, a4 uintptr) int {
	var cshell C.int
	if shell {
		cshell = 1
	}
	return int(C.nox_test_input_config_root_callback(
		(*C.nox_window)(win.C()), cshell, C.int(event), C.uintptr_t(a3), C.uintptr_t(a4),
	))
}
