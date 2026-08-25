#include "client__gui__gadgets__listbox.h"
#include "client__gui__window.h"

#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2_1.h"
#include "GAME3.h"
#include "common__strman.h"

extern uint32_t nox_color_white_2523948;
extern uint32_t nox_color_yellow_2589772;
extern uint32_t nox_color_black_2650656;
extern uint32_t nox_color_orange_2614256;

//----- (004A4310) --------------------------------------------------------
nox_window* nox_gui_newScrollListBox_4A4310(nox_window* parent, int flags, int x, int y, int width, int height,
											 nox_window_data* draw, nox_scrollListBox_data* opts) {
	int has_heading = draw->text[0] != 0;
	int content_height = height;
	int font_height = nox_xxx_guiFontHeightMB_43F320(draw->font);
	if (opts->line_height < font_height) {
		opts->line_height = font_height;
	}
	if (!(draw->style & 0x20)) {
		return NULL;
	}
	nox_window* win = nox_window_new(parent, flags, x, y, width, height, nox_xxx_wndListboxProcPre_4A30D0);
	if (!win) {
		return NULL;
	}
	nox_xxx_wndListboxInit_4A3C00(win, opts);
	if (!draw->win) {
		draw->win = win;
	}
	nox_gui_windowCopyDrawData_46AF80(win, draw);
	opts->items = calloc(opts->count, sizeof(*opts->items));
	if (!opts->items) {
		return NULL;
	}
	opts->field_13_0 = height;
	if (has_heading) {
		opts->field_13_0 -= font_height;
	}
	opts->field_13_1 = 0;
	opts->field_12 = (uint32_t*)(uintptr_t)UINTPTR_MAX;
	opts->field_11_1 = 0;
	opts->field_11_0 = 0;
	opts->field_10 = 0;
	if (opts->field_4) {
		opts->field_12 = calloc(opts->count, sizeof(*opts->field_12));
		if (!opts->field_12) {
			free(opts->items);
			return NULL;
		}
		memset(opts->field_12, 0xFF, opts->count * sizeof(*opts->field_12));
	}
	if (opts->field_3) {
		float slider_values[4] = {0};
		unsigned int child_flags = flags & 0xFFFFEFEF;
		int child_y;
		if (has_heading) {
			child_y = font_height + 1;
			content_height -= font_height + 1;
		} else {
			child_y = 0;
		}
		nox_window_data child_draw = {0};
		int button_height;
		if (!(win->flags & 0x80)) {
			child_draw.bg_color = nox_color_black_2650656;
			child_draw.dis_color = nox_color_black_2650656;
			child_draw.en_color = nox_color_orange_2614256;
			child_draw.hl_color = nox_color_white_2523948;
			child_draw.sel_color = nox_color_yellow_2589772;
			child_draw.text_color = nox_color_orange_2614256;
			nox_wcscpy(child_draw.text, nox_strman_loadString_40F1D0(
				"WindowDir:Up", 0, "C:\\NoxPost\\src\\Client\\Gui\\Gadgets\\listbox.c", 1483));
			button_height = 10;
		} else {
			child_draw.bg_image = nox_xxx_gLoadImg_42F970("DefaultLBUpButton");
			child_draw.hl_image = nox_xxx_gLoadImg_42F970("DefaultLBUpButtonLit");
			child_draw.dis_image = nox_xxx_gLoadImg_42F970("DefaultLBUpButtonDis");
			child_draw.sel_image = nox_xxx_gLoadImg_42F970("DefaultLBUpButtonLit");
			button_height = 13;
		}
		int control_flags = child_flags | 9;
		child_draw.win = win;
		child_draw.style = 1;
		opts->field_7 = nox_gui_newButtonOrCheckbox_4A91A0(win, child_flags | 9, width - 10, child_y, 10,
																 button_height, &child_draw);

		memset(&child_draw, 0, sizeof(child_draw));
		if (!(win->flags & 0x80)) {
			child_draw.bg_color = nox_color_black_2650656;
			child_draw.dis_color = nox_color_black_2650656;
			child_draw.en_color = nox_color_orange_2614256;
			child_draw.hl_color = nox_color_white_2523948;
			child_draw.sel_color = nox_color_yellow_2589772;
			child_draw.text_color = nox_color_orange_2614256;
			nox_wcscpy(child_draw.text, nox_strman_loadString_40F1D0(
				"WindowDir:Down", 0, "C:\\NoxPost\\src\\Client\\Gui\\Gadgets\\listbox.c", 1519));
		} else {
			child_draw.bg_image = nox_xxx_gLoadImg_42F970("DefaultLBDownButton");
			child_draw.hl_image = nox_xxx_gLoadImg_42F970("DefaultLBDownButtonLit");
			child_draw.dis_image = nox_xxx_gLoadImg_42F970("DefaultLBDownButtonDis");
			child_draw.sel_image = nox_xxx_gLoadImg_42F970("DefaultLBDownButtonLit");
		}
		child_draw.win = win;
		child_draw.style = 1;
		opts->field_8 = nox_gui_newButtonOrCheckbox_4A91A0(win, control_flags, width - 10,
																 child_y + content_height - button_height, 10, button_height,
																 &child_draw);

		memset(&child_draw, 0, sizeof(child_draw));
		int slider_width;
		if (!(win->flags & 0x80)) {
			child_draw.bg_color = nox_color_black_2650656;
			child_draw.dis_color = nox_color_black_2650656;
			child_draw.hl_color = nox_color_black_2650656;
			child_draw.en_color = nox_color_orange_2614256;
			child_draw.sel_color = nox_color_orange_2614256;
			slider_width = 10;
		} else {
			child_draw.bg_image = nox_xxx_gLoadImg_42F970("DefaultSliderThumb");
			child_draw.hl_image = nox_xxx_gLoadImg_42F970("DefaultSliderThumbLit");
			child_draw.dis_image = nox_xxx_gLoadImg_42F970("DefaultSliderThumbDis");
			child_draw.sel_image = nox_xxx_gLoadImg_42F970("DefaultSliderThumbLit");
			slider_width = 9;
		}
		child_draw.win = win;
		child_draw.style = 8;
		opts->field_9 = nox_gui_newSlider_4B4EE0(win, control_flags, width - slider_width,
													 child_y + button_height, slider_width,
													 content_height - 2 * button_height, &child_draw, slider_values);
	}
	nox_scrollListBox_data* copy = calloc(1, sizeof(*copy));
	if (!copy) {
		return NULL;
	}
	memcpy(copy, opts, sizeof(*copy));
	win->widget_data = copy;
	return win;
}
