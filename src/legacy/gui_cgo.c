#include "client__gui__window.h"

int nox_window_set_all_funcs_go(nox_window* win, void* a2, void* draw, void* a4);
int nox_window_set_all_funcs(nox_window* win, void* a2, void* draw, void* a4) {
	return nox_window_set_all_funcs_go(win, a2, draw, a4);
}

nox_window* nox_window_new_go(nox_window* parent, int flags, int x, int y, int w, int h, nox_window_init_proc init);
nox_window* nox_window_new(nox_window* parent, int flags, int x, int y, int w, int h, nox_window_init_proc init) {
	return nox_window_new_go(parent, flags, x, y, w, h, init);
}

int nox_xxx_wndSetProc_46B2C0_go(nox_window* a1, void* a2);
int nox_xxx_wndSetProc_46B2C0(nox_window* a1, void* a2) { return nox_xxx_wndSetProc_46B2C0_go(a1, a2); }

int nox_xxx_wndSetWindowProc_46B300_go(nox_window* a1, void* a2);
int nox_xxx_wndSetWindowProc_46B300(nox_window* a1, void* a2) {
	return nox_xxx_wndSetWindowProc_46B300_go(a1, a2);
}

int nox_xxx_wndSetDrawFn_46B340_go(nox_window* a1, void* a2);
int nox_xxx_wndSetDrawFn_46B340(nox_window* a1, void* a2) { return nox_xxx_wndSetDrawFn_46B340_go(a1, a2); }
