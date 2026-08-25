#include <stdlib.h>
#include <string.h>

#include "client__gui__window.h"
#include "common/alloc/classes/alloc_class.h"

#include "defs.h" // nox_video_bag_section_t, nox_video_bag_image_t
#include "memmap.h"

// TODO: part of the GUI toolkit
void nox_client_drawImageAt_47D2C0(void* a1, int x, int y);
int nox_xxx_drawGetStringSize_43F840(int a1, unsigned short* a2, int* a3, int* a4, int a5);
void nox_client_drawSetColor_434460(int a1);
void nox_client_drawRectFilledOpaque_49CE30(int xLeft, int yTop, int a3, int a4);
nox_video_bag_image_t* nox_xxx_gLoadImg_42F970(const char* a1);

//----- (0046ACE0) --------------------------------------------------------
void sub_46ACE0(nox_window* a1, int a2, int a3, int a4) {
	for (int i = a2; i <= a3; i++) {
		nox_window* v5 = nox_xxx_wndGetChildByID_46B0C0(a1, i);
		nox_window_set_hidden(v5, a4);
	}
}

//----- (0046AD20) --------------------------------------------------------
void sub_46AD20(nox_window* a1, int a2, int a3, int a4) {
	int i;            // esi
	nox_window* v5; // eax

	for (i = a2; i <= a3; ++i) {
		v5 = nox_xxx_wndGetChildByID_46B0C0(a1, i);
		nox_xxx_wnd_46ABB0(v5, a4);
	}
}

//----- (0046AA20) --------------------------------------------------------
int nox_gui_getWindowOffs_46AA20(nox_window* win, unsigned int* px, unsigned int* py) {
	if (!win) {
		*px = 0;
		*py = 0;
		return -2;
	}
	*px = win->off_x;
	*py = win->off_y;
	return 0;
}

//----- (0046AA60) --------------------------------------------------------
int nox_client_wndGetPosition_46AA60(nox_window* win, unsigned int* px, unsigned int* py) {
	if (!win) {
		return -2;
	}

	*px = win->off_x;
	*py = win->off_y;
	for (nox_window* i = win->parent; i; i = i->parent) {
		*px += i->off_x;
		*py += i->off_y;
	}
	return 0;
}

//----- (0046AAB0) --------------------------------------------------------
bool nox_xxx_wndPointInWnd_46AAB0(nox_window* win, int x, int y) {
	unsigned int left = 0;
	unsigned int top = 0;
	int width = 0;
	int height = 0;
	nox_client_wndGetPosition_46AA60(win, &left, &top);
	nox_window_get_size(win, &width, &height);
	return x >= (int)left && x <= (int)left + width && y >= (int)top && y <= (int)top + height;
}

//----- (0046AB20) --------------------------------------------------------
int sub_46AB20(nox_window* win, int width, int height) {
	if (!win) {
		return -2;
	}
	win->width = width;
	win->height = height;
	win->end_x = win->off_x + width;
	win->end_y = win->off_y + height;
	nox_window_call_field_94(win, 16388, width, height);
	return 0;
}

//----- (0046AB70) --------------------------------------------------------
int nox_window_get_size(nox_window* win, int* outW, int* outH) {
	if (!win) {
		*outW = 0;
		*outH = 0;
		return -2;
	}
	*outW = win->width;
	*outH = win->height;
	return 0;
}

//----- (0046ABB0) --------------------------------------------------------
extern int32_t nox_xxx_wnd_46ABB0_go(nox_window* win, int32_t enabled);

int nox_xxx_wnd_46ABB0(nox_window* win, int a2) {
	return nox_xxx_wnd_46ABB0_go(win, a2);
}

//----- (0046AD60) --------------------------------------------------------
int nox_xxx_wnd_46AD60(nox_window* win, int flags) {
	if (!win) {
		return -2;
	}
	int result = win->flags;
	win->flags = flags | result;
	return result;
}

//----- (0046AD80) --------------------------------------------------------
int nox_xxx_wndClearFlag_46AD80(nox_window* win, int flags) {
	if (!win) {
		return -2;
	}
	int result = win->flags;
	win->flags = result & ~flags;
	return result;
}

//----- (0046ADA0) --------------------------------------------------------
int nox_xxx_wndGetFlags_46ADA0(nox_window* win) {
	return win ? win->flags : -2;
}

//----- (0046B250) --------------------------------------------------------
int nox_window_is_child(nox_window* a1, nox_window* a2) {
	if (!a1) {
		return 0;
	}
	if (!a2) {
		return 0;
	}

	nox_window* cur = a2;
	while (1) {
		cur = cur->parent;
		if (a1 == cur) {
			break;
		}
		if (!cur) {
			return 0;
		}
	}
	return 1;
}

//----- (0046B280) --------------------------------------------------------
int nox_xxx_wnd_46B280(nox_window* win, nox_window* draw_owner) {
	if (!win) {
		return -2;
	}
	win->draw_data.win = draw_owner ? draw_owner : win;
	return 0;
}
