#include "../client__gui__gadgets__listbox.h"

#include <limits.h>
#include <stdint.h>

typedef nox_window* (*scroll_list_box_fn)(nox_window*, int, int, int, int, int,
										 nox_window_data*, nox_scrollListBox_data*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(
	_Generic(&nox_gui_newScrollListBox_4A4310,
		scroll_list_box_fn: 1, default: 0),
	"scroll-list constructor must preserve its pointer-width arguments");

static nox_window* expected_parent;
static nox_window_data* expected_draw;
static nox_scrollListBox_data* expected_opts;
static nox_window* expected_result;

nox_window* nox_gui_newScrollListBox_4A4310_go(
		nox_window* parent, int flags, int x, int y, int width, int height,
		nox_window_data* draw, nox_scrollListBox_data* opts) {
	if (parent != expected_parent || draw != expected_draw || opts != expected_opts)
		return 0;
	if (flags != INT32_C(0x12345678) || x != -31 || y != 47 ||
		width != 640 || height != 480)
		return 0;
	return expected_result;
}

int main(void) {
	nox_window parent = {0};
	nox_window result = {0};
	nox_window_data draw = {0};
	nox_scrollListBox_data opts = {0};
	expected_parent = &parent;
	expected_draw = &draw;
	expected_opts = &opts;
	expected_result = &result;
	if (nox_gui_newScrollListBox_4A4310(&parent, INT32_C(0x12345678),
			-31, 47, 640, 480, &draw, &opts) != &result)
		return __LINE__;
	return 0;
}
