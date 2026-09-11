package legacy

/*
#include <stddef.h>
#include <stdint.h>

#include "GAME3.h"

typedef int (*nox_test_slider_event_callback_t)(nox_window*, int, uintptr_t, uintptr_t);
typedef int (*nox_test_slider_draw_callback_t)(nox_window*, nox_window_data*);

_Static_assert(_Generic(&sub_4B4860, nox_test_slider_event_callback_t: 1, default: 0),
	"horizontal slider event callback must use native-width arguments");
_Static_assert(_Generic(&nox_xxx_wndScrollBoxDraw_4B4BA0, nox_test_slider_event_callback_t: 1, default: 0),
	"vertical slider event callback must use native-width arguments");
_Static_assert(_Generic(&sub_4B51E0, nox_test_slider_draw_callback_t: 1, default: 0),
	"horizontal color draw callback must use typed native pointers");
_Static_assert(_Generic(&sub_4B52C0, nox_test_slider_draw_callback_t: 1, default: 0),
	"horizontal image draw callback must use typed native pointers");

static int nox_test_slider_event_callback_contract(int vertical, uintptr_t* address) {
	nox_window win = {0};
	uint32_t data[4] = {0};
	win.widget_data = data;
	*address = (uintptr_t)&win;

	int event5;
	int event17;
	if (vertical) {
		event5 = nox_xxx_wndScrollBoxDraw_4B4BA0(&win, 5, 0, 0);
		event17 = nox_xxx_wndScrollBoxDraw_4B4BA0(&win, 17, 0, 0);
	} else {
		event5 = sub_4B4860(&win, 5, 0, 0);
		event17 = sub_4B4860(&win, 17, 0, 0);
	}
	return event5 == 1 && event17 == 1 && win.widget_data == data;
}

static int nox_test_slider_pointer_slots(uintptr_t value) {
	nox_window win = {0};
	win.widget_data = (void*)value;
	win.draw_data.win = (nox_window*)value;
	win.field_100 = (nox_window*)value;
	return (uintptr_t)win.widget_data == value &&
		(uintptr_t)win.draw_data.win == value &&
		(uintptr_t)win.field_100 == value;
}

static size_t nox_test_slider_window_size(void) { return sizeof(nox_window); }
static size_t nox_test_slider_widget_data_offset(void) { return offsetof(nox_window, widget_data); }
static size_t nox_test_slider_style_offset(void) {
	return offsetof(nox_window, draw_data) + offsetof(nox_window_data, style);
}
static size_t nox_test_slider_owner_offset(void) {
	return offsetof(nox_window, draw_data) + offsetof(nox_window_data, win);
}
static size_t nox_test_slider_enabled_color_offset(void) {
	return offsetof(nox_window, draw_data) + offsetof(nox_window_data, en_color);
}
static size_t nox_test_slider_highlight_color_offset(void) {
	return offsetof(nox_window, draw_data) + offsetof(nox_window_data, hl_color);
}
static size_t nox_test_slider_thumb_offset(void) { return offsetof(nox_window, field_100); }
static size_t nox_test_slider_draw_bg_image_offset(void) { return offsetof(nox_window_data, bg_image); }
static size_t nox_test_slider_draw_disabled_image_offset(void) { return offsetof(nox_window_data, dis_image); }
*/
import "C"

func sliderEventCallbackContract(vertical bool) (uintptr, bool) {
	var address C.uintptr_t
	verticalC := C.int(0)
	if vertical {
		verticalC = 1
	}
	ok := C.nox_test_slider_event_callback_contract(verticalC, &address) != 0
	return uintptr(address), ok
}

func sliderPointerSlotsContract(value uintptr) bool {
	return C.nox_test_slider_pointer_slots(C.uintptr_t(value)) != 0
}

func sliderNativeLayout() (windowSize, widgetData, style, owner, enabledColor, highlightColor, thumb, bgImage, disabledImage uintptr) {
	return uintptr(C.nox_test_slider_window_size()),
		uintptr(C.nox_test_slider_widget_data_offset()),
		uintptr(C.nox_test_slider_style_offset()),
		uintptr(C.nox_test_slider_owner_offset()),
		uintptr(C.nox_test_slider_enabled_color_offset()),
		uintptr(C.nox_test_slider_highlight_color_offset()),
		uintptr(C.nox_test_slider_thumb_offset()),
		uintptr(C.nox_test_slider_draw_bg_image_offset()),
		uintptr(C.nox_test_slider_draw_disabled_image_offset())
}
