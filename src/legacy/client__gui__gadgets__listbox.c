#include "client__gui__gadgets__listbox.h"

nox_window* nox_gui_newScrollListBox_4A4310_go(nox_window* parent, int flags, int x, int y, int width, int height,
											 nox_window_data* draw, nox_scrollListBox_data* opts);

//----- (004A4310) --------------------------------------------------------
nox_window* nox_gui_newScrollListBox_4A4310(nox_window* parent, int flags, int x, int y, int width, int height,
											 nox_window_data* draw, nox_scrollListBox_data* opts) {
	return nox_gui_newScrollListBox_4A4310_go(parent, flags, x, y, width, height, draw, opts);
}
