package legacy

/*
#include "game2_server_bridge.h"
*/
import "C"
import "unsafe"

//export sub_455A00
func sub_455A00(show C.int) C.int {
	w := C.dword_5d4594_1045604
	if show != 0 && *g2ServerWord(1045608) != 0 && C.nox_xxx_wndGetFlags_46ADA0(w)&C.NOX_WIN_HIDDEN != 0 {
		return C.nox_xxx_wndShowModalMB_46A8C0(w)
	}
	return g2ServerHidden(w, true)
}

//export sub_455A50
func sub_455A50(count C.char) C.char {
	C.nox_xxx_objGetTeamByNetCode_418C80(C.int(C.nox_player_netCode_85319C))
	if g2ServerEngine(C.NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING) {
		return 0
	}
	w := C.dword_5d4594_1045604
	if w == nil {
		v := C.sub_455C30()
		if v == 0 {
			return C.char(v)
		}
		w = C.dword_5d4594_1045604
	}
	var width, height, depth, maxW, maxH C.int
	C.nox_xxx_gameGetScreenBoundaries_43BEB0_get_video_mode(&width, &height, &depth)
	C.nox_xxx_screenGetSize_430C50_get_video_max(&maxW, &maxH)
	if width > maxW {
		width = maxW
	}
	if height > maxH {
		height = maxH
	}
	for _, off := range []uintptr{1045612, 1045616, 1045620} {
		*g2ServerWord(off) = 0
	}
	*g2ServerWord(1045608) = 1
	*g2ServerWord(1045624) = 0
	C.sub_46ACE0(w, 8811, 8826, 1)
	*g2ServerByte(1045628) = byte(count)
	var last *C.nox_window
	if count != 0 {
		for i := 0; i < int(*g2ServerByte(1045628)); i++ {
			last = g2ServerChild(w, i+8811)
			g2ServerHidden(last, false)
		}
	}
	n := *g2ServerByte(1045628)
	x := C.intptr_t(width) - w.width
	if n <= 4 {
		x = C.intptr_t(width) - w.width/2
	}
	w.off_x = x - 91
	w.end_x = w.width + w.off_x
	if n <= 8 {
		if n <= 4 {
			if last == nil {
				*g2ServerWord(1045608) = 0
				return C.char(n)
			}
			w.off_y = C.intptr_t(height) - 40*C.intptr_t(n)
		} else {
			w.off_y = C.intptr_t(height) - w.height/2
		}
	} else {
		w.off_y = C.intptr_t(height) - w.height
	}
	w.end_y = w.height + w.off_y
	return C.char(C.nox_xxx_wndShowModalMB_46A8C0(w))
}

//export sub_455C10
func sub_455C10() C.int { *g2ServerWord(1045608) = 0; return sub_455A00(0) }

//export sub_455CD0
func sub_455CD0(w *C.nox_window, d *C.nox_window_data) C.int {
	var x, y C.uint
	C.nox_client_wndGetPosition_46AA60(w, &x, &y)
	index := uint8(int32(w.id) - 8811)
	image := d.bg_image
	v := *g2ServerByte(1045612 + uintptr(index))
	if v == 1 {
		image = d.en_image
	} else if v == 2 {
		image = d.dis_image
	}
	if image != nil {
		C.nox_client_drawImageAt_47D2C0((*C.nox_video_bag_image_t)(image), C.int(x), C.int(y))
	}
	if w.flags&C.NOX_WIN_LAYER_FRONT != 0 && C.nox_ctf_flag_team_border != nil {
		C.nox_client_drawImageAt_47D2C0(C.nox_ctf_flag_team_border, C.int(x)-4, C.int(y)-4)
	}
	return 1
}

//export sub_455E70
func sub_455E70(index C.uchar) C.int {
	w := C.dword_5d4594_1045604
	for i := 0; i < int(*g2ServerByte(1045628)); i++ {
		C.nox_xxx_wndClearFlag_46AD80(g2ServerChild(w, i+8811), 32)
	}
	return C.nox_xxx_wnd_46AD60(g2ServerChild(w, int(index)+8810), 32)
}

//export sub_455EE0
func sub_455EE0() C.int {
	var result C.int
	if C.dword_5d4594_1045604 != nil {
		result = C.nox_xxx_windowDestroyMB_46C4E0(C.dword_5d4594_1045604)
	}
	C.dword_5d4594_1045604 = nil
	*g2ServerWord(1045608) = 0
	return result
}

//export sub_455F10
func sub_455F10(show C.int) C.int {
	w := C.dword_5d4594_1045636
	if show != 0 && C.dword_5d4594_1045640 != 0 && C.nox_xxx_wndGetFlags_46ADA0(w)&C.NOX_WIN_HIDDEN != 0 {
		return C.nox_xxx_wndShowModalMB_46A8C0(w)
	}
	return g2ServerHidden(w, true)
}

//export sub_455F60
func sub_455F60() C.int {
	C.nox_xxx_objGetTeamByNetCode_418C80(C.int(C.nox_player_netCode_85319C))
	if g2ServerEngine(C.NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING) {
		return 0
	}
	w := C.dword_5d4594_1045636
	if w == nil {
		v := sub_456070()
		if v == 0 {
			return v
		}
		w = C.dword_5d4594_1045636
	}
	var width, height, depth, maxW, maxH C.int
	C.nox_xxx_gameGetScreenBoundaries_43BEB0_get_video_mode(&width, &height, &depth)
	C.nox_xxx_screenGetSize_430C50_get_video_max(&maxW, &maxH)
	if width > maxW {
		width = maxW
	}
	if height > maxH {
		height = maxH
	}
	C.dword_5d4594_1045640 = 1
	w.off_x = C.intptr_t(width) - w.width/3 - 91
	w.end_x = w.width + w.off_x
	w.off_y = C.intptr_t(height) - 120
	w.end_y = w.height + w.off_y
	*g2ServerByte(1045644) = 0
	return C.nox_xxx_wndShowModalMB_46A8C0(w)
}

//export sub_456050
func sub_456050() C.int { C.dword_5d4594_1045640 = 0; return sub_455F10(0) }

//export sub_456070
func sub_456070() C.int {
	if C.dword_5d4594_1045636 == nil {
		w := C.nox_new_window_from_file(internCStr("gui_fb.wnd"), nil)
		C.dword_5d4594_1045636 = w
		if w == nil {
			return 0
		}
		C.nox_window_set_all_funcs(w, nil, unsafe.Pointer(C.sub_4560D0), nil)
		sub_455F10(0)
		C.nox_fb_flag_team_border = (*C.nox_video_bag_image_t)(C.nox_xxx_gLoadImg_42F970(internCStr("FlagTeamBorder")))
	}
	return 1
}

//export sub_4560D0
func sub_4560D0(w *C.nox_window, d *C.nox_window_data) C.int {
	var x, y C.uint
	C.nox_client_wndGetPosition_46AA60(w, &x, &y)
	if d.bg_image != nil {
		C.nox_client_drawImageAt_47D2C0((*C.nox_video_bag_image_t)(d.bg_image), C.int(x), C.int(y))
	}
	if w.flags&C.NOX_WIN_LAYER_FRONT != 0 && C.nox_fb_flag_team_border != nil {
		C.nox_client_drawImageAt_47D2C0(C.nox_fb_flag_team_border, C.int(x)-4, C.int(y)-4)
	}
	return 1
}

//export sub_456240
func sub_456240() C.int {
	var result C.int
	if C.dword_5d4594_1045636 != nil {
		result = C.nox_xxx_windowDestroyMB_46C4E0(C.dword_5d4594_1045636)
	}
	C.dword_5d4594_1045636 = nil
	C.dword_5d4594_1045640 = 0
	return result
}
