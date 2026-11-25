#ifndef OPENNOX_GAME3_WIDGETS_REFERENCE_H
#define OPENNOX_GAME3_WIDGETS_REFERENCE_H
#include <math.h>
#include "game3_widgets_native.h"
#include "memmap.h"
typedef struct nox_slider_data_4B4860 { int min,max;float scale;int value; } nox_slider_data_4B4860;
static int game3_ref_sub_4A2560(const uint32_t* a1, const nox_gui_server_ent_t* a2) {
	double v2; // st7
	double v3; // st6

	// The PE32 routine subtracts 32-bit coordinates and interprets each
	// result as signed before calculating the distance.
	v2 = (double)(int32_t)((uint32_t)(int32_t)a2->field_11_0 - a1[0]);
	v3 = (double)(int32_t)((uint32_t)(int32_t)a2->field_11_2 - a1[1]);
	return sqrt(v3 * v3 + v2 * v2) <= *getMemDoublePtr(0x581450, 9720);
}

static int game3_ref_sub_4A25C0(uint32_t* a1, nox_list_item_t* a2) {
	int count = 0;
	for (nox_list_item_t* item = nox_common_list_getFirstSafe_425890(a2); item;
		 item = nox_common_list_getNextSafe_4258A0(item)) {
		const nox_gui_server_node_t* node = nox_wol_server_node_from_list_const(item);
		if (game3_ref_sub_4A2560(a1, &node->server)) {
			++count;
		}
	}
	return count;
}

static uint32_t* game3_ref_sub_4A2830(int a1, int a2, uint32_t* a3) {
	uint32_t* result; // eax

	result = a3;
	*a3 = a1 - 100;
	a3[1] = a2 - 20;
	if (a1 - 100 + 200 > 600) {
		*a3 = 400;
	}
	if (a2 - 20 + 200 > 451) {
		a3[1] = 251;
	}
	if (a3[1] < 27) {
		a3[1] = 27;
	}
	if ((int)*a3 < 216) {
		*a3 = 216;
	}
	return result;
}

static int game3_ref_nox_xxx_wndListBoxFirstVisible(nox_window* win) {
	nox_scrollListBox_data* data = win ? win->widget_data : NULL;
	if (!data || !data->items || data->field_11_0 == 0) {
		return 0;
	}
	for (int i = 0; i < data->field_11_0; i++) {
		if (data->items[i].field_0 > data->field_13_1) {
			return i;
		}
	}
	return 0;
}

static int game3_ref_nox_xxx_wndListBox_4A2D10(nox_window* win, int delta, int update_slider) {
	typedef struct {
		int min;
		int max;
		float scale;
		int value;
	} slider_data;

	nox_scrollListBox_data* data = win ? win->widget_data : NULL;
	if (!data || !data->items) {
		return 0;
	}
	int first = game3_ref_nox_xxx_wndListBoxFirstVisible(win);
	first += delta;
	if (first < 0) {
		first = 0;
	} else if (first >= data->field_11_0) {
		first = (int)data->field_11_0 - 1;
	}
	if (update_slider) {
		data->field_13_1 = first <= 0 ? 0 : (uint16_t)(data->items[first - 1].field_0 + 1);
	}

	nox_window* slider = data->field_9;
	slider_data* values = slider ? slider->widget_data : NULL;
	if (!slider || !values) {
		return 0;
	}
	values->max = (int)data->field_10 - (int)data->field_13_0 + 3;
	if (values->max < 0) {
		values->max = 0;
	}
	int thumb_height = slider->field_100 ? (int)slider->field_100->height : 0;
	values->scale = values->max ? (float)((int)slider->height - thumb_height) / (float)values->max : 0;
	if (update_slider) {
		return (int)nox_window_call_field_94(slider, 16394, values->max - data->field_13_1, 0);
	}
	return 0;
}

static int game3_ref_nox_xxx_wndListBox_4A3A70(nox_window* win) {
	nox_scrollListBox_data* data = win ? win->widget_data : NULL;
	if (!data || !data->items) {
		return 0;
	}
	int total_height = 0;
	for (int i = 0; i < data->field_11_0; i++) {
		total_height += *(uint8_t*)&data->items[i].field_130 + 1;
		data->items[i].field_0 = (uint32_t)total_height;
	}
	data->field_10 = (uint32_t)total_height;
	return game3_ref_nox_xxx_wndListBox_4A2D10(win, 0, 1);
}

static uint32_t game3_ref_nox_xxx_wndListBoxColor(int color_index) {
	switch (color_index) {
	case 0: return nox_color_black_2650656;
	case 1: return *getMemU32Ptr(0x852978, 4);
	case 2: return *getMemU32Ptr(0x85B3FC, 956);
	case 3: return *getMemU32Ptr(0x5D4594, 2597996);
	case 4: return nox_color_white_2523948;
	case 5: return nox_color_violet_2598268;
	case 6: return *getMemU32Ptr(0x85B3FC, 940);
	case 7: return nox_color_red_2589776;
	case 8: return *getMemU32Ptr(0x85B3FC, 984);
	case 9: return dword_8531A0_2572;
	case 10: return nox_color_green_2614268;
	case 11: return *getMemU32Ptr(0x85B3FC, 944);
	case 12: return nox_color_cyan_2649820;
	case 13: return nox_color_blue_2650684;
	case 14: return nox_color_orange_2614256;
	case 15: return nox_color_yellow_2589772;
	case 16: return *getMemU32Ptr(0x852978, 0);
	default: return 0;
	}
}

static int game3_ref_sub_4A4800(const nox_scrollListBox_data* data) {
	if (!data || !data->items || data->field_11_0 == 0 || data->items[0].field_0 > data->field_13_1) {
		return 0;
	}
	for (int i = 1; i < data->field_11_0; i++) {
		if (data->items[i].field_0 > data->field_13_1) {
			return i;
		}
	}
	return 0;
}

static int game3_ref_sub_4B4860(nox_window* win, int ev, uintptr_t a3, uintptr_t a4) {
	nox_slider_data_4B4860* data = win->widget_data;
	switch (ev) {
	case 5:
		return 1;
	case 6:
	case 7: {
		uintptr_t packed = a3;
		int cursor_x = (uint16_t)a3;
		unsigned int left;
		unsigned int top;
		nox_client_wndGetPosition_46AA60(win, &left, &top);
		if (cursor_x > (int)win->width + (int)left - 10) {
			cursor_x = (int)win->width + (int)left - 10;
		}
		nox_window_setPos_46A9B0(win->field_100, cursor_x - (int)left - 5, 0);
		nox_window_call_field_94(win, 0x4000, 0, packed);
		return 1;
	}
	case 8:
		if (!(win->draw_data.style & 0x100)) {
			return 1;
		}
		nox_window_call_field_94(win->draw_data.win, 0x4000, (uintptr_t)win, 0);
		return 1;
	case 17:
		if (!(win->draw_data.style & 0x100)) {
			return 1;
		}
		nox_xxx_wndSetRectColor2MB_46AFE0(win, win->draw_data.hl_color);
		nox_window_call_field_94(win->draw_data.win, 16389, (uintptr_t)win, 0);
		nox_xxx_windowFocus_46B500(win);
		return 1;
	case 18:
		if (win->draw_data.style & 0x100) {
			nox_xxx_wndSetRectColor2MB_46AFE0(win, win->draw_data.en_color);
			nox_window_call_field_94(win->draw_data.win, 16390, (uintptr_t)win, 0);
		}
		return 1;
	case 21:
		switch (a3) {
		case 15:
		case 208:
			if (a4 != 2) {
				return 1;
			}
			nox_xxx_wndRetNULL_46A8A0();
			return 1;
		case 200:
			if (a4 == 2) {
				nox_xxx_wndRetNULL_0_46A8B0();
			}
			return 1;
		case 203:
			if (a4 != 2) {
				return 1;
			}
			if (data->value >= data->max - 1) {
				return 1;
			}
			data->value += 2;
			nox_window_call_field_94(win->draw_data.win, 16393, (uintptr_t)win, data->value);
			nox_window_setPos_46A9B0(win->field_100, (int)((double)(data->value - data->min) * data->scale), 0);
			return 1;
		case 205:
			if (a4 != 2) {
				return 1;
			}
			if (data->value <= data->min + 1) {
				return 1;
			}
			data->value -= 2;
			nox_window_call_field_94(win->draw_data.win, 16393, (uintptr_t)win, data->value);
			nox_window_setPos_46A9B0(win->field_100, (int)((double)(data->value - data->min) * data->scale), 0);
			return 1;
		default:
			return 0;
		}
	default:
		return 0;
	}
}

static int game3_ref_nox_xxx_wndScrollBoxDraw_4B4BA0(nox_window* win, int ev, uintptr_t a3, uintptr_t a4) {
	nox_slider_data_4B4860* data = win->widget_data;
	switch (ev) {
	case 5:
		return 1;
	case 6:
	case 7: {
		uintptr_t packed = a3;
		int cursor_y = (uint16_t)(a3 >> 16);
		unsigned int left;
		unsigned int top;
		nox_client_wndGetPosition_46AA60(win, &left, &top);
		if (cursor_y > (int)win->height + (int)top - 10) {
			cursor_y = (int)win->height + (int)top - 10;
		}
		nox_window_setPos_46A9B0(win->field_100, 0, cursor_y - (int)top - 5);
		nox_window_call_field_94(win, 0x4000, 0, packed);
		return 1;
	}
	case 8:
		if (!(win->draw_data.style & 0x100)) {
			return 1;
		}
		nox_window_call_field_94(win->draw_data.win, 0x4000, (uintptr_t)win, 0);
		return 1;
	case 17:
		if (!(win->draw_data.style & 0x100)) {
			return 1;
		}
		nox_xxx_wndSetRectColor2MB_46AFE0(win, win->draw_data.hl_color);
		nox_window_call_field_94(win->draw_data.win, 16389, (uintptr_t)win, 0);
		nox_xxx_windowFocus_46B500(win);
		return 1;
	case 18:
		if (win->draw_data.style & 0x100) {
			nox_xxx_wndSetRectColor2MB_46AFE0(win, win->draw_data.en_color);
			nox_window_call_field_94(win->draw_data.win, 16390, (uintptr_t)win, 0);
		}
		return 1;
	case 21:
		switch (a3) {
		case 0xFu:
		case 0xCDu:
			if (a4 != 2) {
				return 1;
			}
			nox_xxx_wndRetNULL_46A8A0();
			return 1;
		case 0xC8u:
			if (a4 != 2) {
				return 1;
			}
			if (data->value >= data->max - 1) {
				return 1;
			}
			data->value += 2;
			nox_window_call_field_94(win->draw_data.win, 16393, (uintptr_t)win, data->value);
			nox_window_setPos_46A9B0(win->field_100, 0, (int)((double)(data->max - data->value) * data->scale));
			return 1;
		case 0xCBu:
			if (a4 == 2) {
				nox_xxx_wndRetNULL_0_46A8B0();
			}
			return 1;
		case 0xD0u:
			if (a4 != 2) {
				return 1;
			}
			if (data->value <= data->min + 1) {
				return 1;
			}
			data->value -= 2;
			nox_window_call_field_94(win->draw_data.win, 16393, (uintptr_t)win, data->value);
			nox_window_setPos_46A9B0(win->field_100, 0, (int)((double)(data->max - data->value) * data->scale));
			return 1;
		default:
			return 0;
		}
	default:
		return 0;
	}
}

static int game3_ref_sub_4B5010(nox_window* win, int ev, uintptr_t a3, uintptr_t a4) {
	nox_slider_data_4B4860* data = win ? win->widget_data : 0;
	if (!win) {
		return 0;
	}
	if (ev == 2) {
		free(data);
		win->widget_data = 0;
		return 0;
	}
	if (!data) {
		return 0;
	}
	if (ev == 16394) {
		int value = (int)a3;
		if (value >= data->min && value <= data->max) {
			data->value = value;
			nox_window_setPos_46A9B0(win->field_100, (int)((double)value * data->scale), 0);
		}
		return 0;
	}
	if (ev == 16395) {
		data->min = (int)a3;
		data->max = (int)a4;
		data->value = data->min;
		int range = data->max - data->min;
		data->scale = range ? (float)(win->width - 10) / (float)range : 0;
		nox_window_setPos_46A9B0(win->field_100, 0, 0);
		return 0;
	}
	if (ev == 0x4000) {
		int cursor_x = (int)(uint16_t)a4;
		unsigned int left;
		unsigned int top;
		nox_client_wndGetPosition_46AA60(win, &left, &top);
		if (cursor_x < (int)left) {
			nox_window_setPos_46A9B0(win->field_100, 0, 0);
			data->value = data->min;
		} else if (cursor_x < (int)(left + win->width)) {
			int value = data->scale ? (int)((double)(cursor_x - (int)left) / data->scale) : data->min;
			if (value > data->max) {
				value = data->max;
			}
			data->value = value;
		} else {
			nox_window_setPos_46A9B0(win->field_100, win->width - win->field_100->width, 0);
			data->value = data->max;
		}
		nox_window_call_field_94(win->parent, 16393, (uintptr_t)win, data->value);
		return 0;
	}
	if (ev != 23) {
		return 0;
	}
	if (a3) {
		win->draw_data.field_0 |= 2;
	} else {
		win->draw_data.field_0 &= ~2u;
	}
	nox_window_call_field_94(win->parent, 16387, a3, nox_xxx_wndGetID_46B0A0(win));
	return 1;
}

static int game3_ref_nox_xxx_wndScrollBoxProc_4B5320(nox_window* win, int ev, uintptr_t a3, uintptr_t a4) {
	nox_slider_data_4B4860* data = win ? win->widget_data : 0;
	if (!win) {
		return 0;
	}
	if (ev == 2) {
		free(data);
		win->widget_data = 0;
		return 0;
	}
	if (!data) {
		return 0;
	}
	if (ev == 16394) {
		int value = (int)a3;
		if (value >= data->min && value <= data->max) {
			data->value = value;
			int inverted = data->max - value;
			nox_window_setPos_46A9B0(win->field_100, 0, (int)((double)inverted * data->scale));
		}
		return 0;
	}
	if (ev == 16395) {
		data->min = (int)a3;
		data->max = (int)a4;
		data->value = data->min;
		int range = data->max - data->min;
		data->scale = range ? (float)(win->height - 10) / (float)range : 0;
		nox_window_setPos_46A9B0(win->field_100, 0, (int)((double)range * data->scale));
		return 0;
	}
	if (ev == 0x4007) {
		nox_window_call_field_94(win->parent, 16396, (uintptr_t)win, data->value);
		return 0;
	}
	if (ev == 0x17) {
		if (a3) {
			win->draw_data.field_0 |= 2;
		} else {
			win->draw_data.field_0 &= ~2u;
		}
		nox_window_call_field_94(win->parent, 16387, a3, nox_xxx_wndGetID_46B0A0(win));
		return 1;
	}
	if (ev != 0x4000) {
		return 0;
	}
	int cursor_y = (int)(a4 >> 16);
	unsigned int left;
	unsigned int top;
	nox_client_wndGetPosition_46AA60(win, &left, &top);
	if (cursor_y >= (int)top) {
		if (cursor_y < (int)(top + win->height)) {
			int raw = data->scale ? (int)((double)(cursor_y - (int)top) / data->scale) : 0;
			if (raw > data->max) {
				raw = data->max;
			}
			data->value = data->max - raw;
		} else {
			nox_window_setPos_46A9B0(win->field_100, 0, win->height - win->field_100->height);
			data->value = data->min;
		}
	} else {
		nox_window_setPos_46A9B0(win->field_100, 0, 0);
		data->value = data->max;
	}
	nox_window_call_field_94(win->parent, 16393, (uintptr_t)win, data->value);
	return 0;
}
#endif
