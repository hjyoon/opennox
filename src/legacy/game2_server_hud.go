package legacy

/*
#include "game2_server_bridge.h"
*/
import "C"
import "unsafe"

var flagballTeamBorder *C.nox_video_bag_image_t

//export sub_455A00
func sub_455A00(show C.int) C.int {
	w := serverSharedWindows.ctf_hud
	if show != 0 && *g2ServerWord(1045608) != 0 && nox_xxx_wndGetFlags_46ADA0(w)&C.NOX_WIN_HIDDEN != 0 {
		return C.int(nox_xxx_wndShowModalMB_46A8C0(w))
	}
	return g2ServerHidden(w, true)
}

//export sub_455A50
func sub_455A50(count C.char) C.char {
	C.nox_xxx_objGetTeamByNetCode_418C80(C.int(clientPlayerNetCode))
	if g2ServerEngine(C.NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING) {
		return 0
	}
	w := serverSharedWindows.ctf_hud
	if w == nil {
		v := sub_455C30()
		if v == 0 {
			return C.char(v)
		}
		w = serverSharedWindows.ctf_hud
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
	sub_46ACE0(w, 8811, 8826, 1)
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
	return C.char(C.int(nox_xxx_wndShowModalMB_46A8C0(w)))
}

//export sub_455C10
func sub_455C10() C.int { *g2ServerWord(1045608) = 0; return sub_455A00(0) }

//export sub_455CD0
func sub_455CD0(w *C.nox_window, d *C.nox_window_data) C.int {
	var x, y C.uint
	nox_client_wndGetPosition_46AA60(w, &x, &y)
	index := uint8(int32(w.id) - 8811)
	image := d.bg_image
	v := *g2ServerByte(1045612 + uintptr(index))
	if v == 1 {
		image = d.en_image
	} else if v == 2 {
		image = d.dis_image
	}
	if image != nil {
		nox_client_drawImageAt_47D2C0((*C.nox_video_bag_image_t)(image), int32(x), int32(y))
	}
	if w.flags&C.NOX_WIN_LAYER_FRONT != 0 && legacyPointers.flagTeamBorder != nil {
		nox_client_drawImageAt_47D2C0(legacyPointers.flagTeamBorder, int32(C.int(x)-4), int32(C.int(y)-4))
	}
	return 1
}

//export sub_455E70
func sub_455E70(index C.uchar) C.int {
	w := serverSharedWindows.ctf_hud
	for i := 0; i < int(*g2ServerByte(1045628)); i++ {
		nox_xxx_wndClearFlag_46AD80(g2ServerChild(w, i+8811), 32)
	}
	return nox_xxx_wnd_46AD60(g2ServerChild(w, int(index)+8810), 32)
}

//export sub_455EE0
func sub_455EE0() C.int {
	var result C.int
	if serverSharedWindows.ctf_hud != nil {
		result = C.int(nox_xxx_windowDestroyMB_46C4E0(serverSharedWindows.ctf_hud))
	}
	serverSharedWindows.ctf_hud = nil
	*g2ServerWord(1045608) = 0
	return result
}

//export sub_455F10
func sub_455F10(show C.int) C.int {
	w := serverSharedWindows.flagball_hud
	if show != 0 && legacyWords.dword_5d4594_1045640 != 0 && nox_xxx_wndGetFlags_46ADA0(w)&C.NOX_WIN_HIDDEN != 0 {
		return C.int(nox_xxx_wndShowModalMB_46A8C0(w))
	}
	return g2ServerHidden(w, true)
}

//export sub_455F60
func sub_455F60() C.int {
	C.nox_xxx_objGetTeamByNetCode_418C80(C.int(clientPlayerNetCode))
	if g2ServerEngine(C.NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING) {
		return 0
	}
	w := serverSharedWindows.flagball_hud
	if w == nil {
		v := sub_456070()
		if v == 0 {
			return v
		}
		w = serverSharedWindows.flagball_hud
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
	legacyWords.dword_5d4594_1045640 = 1
	w.off_x = C.intptr_t(width) - w.width/3 - 91
	w.end_x = w.width + w.off_x
	w.off_y = C.intptr_t(height) - 120
	w.end_y = w.height + w.off_y
	*g2ServerByte(1045644) = 0
	return C.int(nox_xxx_wndShowModalMB_46A8C0(w))
}

//export sub_456050
func sub_456050() C.int { legacyWords.dword_5d4594_1045640 = 0; return sub_455F10(0) }

//export sub_456070
func sub_456070() C.int {
	if serverSharedWindows.flagball_hud == nil {
		w := loadWindowFromFile("gui_fb.wnd", nil)
		serverSharedWindows.flagball_hud = w
		if w == nil {
			return 0
		}
		nox_window_set_all_funcs(w, nil, unsafe.Pointer(C.sub_4560D0), nil)
		sub_455F10(0)
		flagballTeamBorder = loadImageHandle("FlagTeamBorder")
	}
	return 1
}

//export sub_4560D0
func sub_4560D0(w *C.nox_window, d *C.nox_window_data) C.int {
	var x, y C.uint
	nox_client_wndGetPosition_46AA60(w, &x, &y)
	if d.bg_image != nil {
		nox_client_drawImageAt_47D2C0((*C.nox_video_bag_image_t)(d.bg_image), int32(x), int32(y))
	}
	if w.flags&C.NOX_WIN_LAYER_FRONT != 0 && flagballTeamBorder != nil {
		nox_client_drawImageAt_47D2C0(flagballTeamBorder, int32(C.int(x)-4), int32(C.int(y)-4))
	}
	return 1
}

//export sub_456240
func sub_456240() C.int {
	var result C.int
	if serverSharedWindows.flagball_hud != nil {
		result = C.int(nox_xxx_windowDestroyMB_46C4E0(serverSharedWindows.flagball_hud))
	}
	serverSharedWindows.flagball_hud = nil
	legacyWords.dword_5d4594_1045640 = 0
	return result
}

const clientCTFSource = `C:\\NoxPost\\src\\client\\Gui\\GUI_CTF.c`

//export sub_455C30
func sub_455C30() C.int {
	if serverSharedWindows.ctf_hud != nil {
		return 1
	}
	root := loadWindowFromFile("GUI_CTF.wnd", nil)
	serverSharedWindows.ctf_hud = root
	if root == nil {
		return 0
	}
	for id := 8811; id <= 8826; id++ {
		child := g2ServerChild(serverSharedWindows.ctf_hud, id)
		if child == nil {
			return 0
		}
		nox_window_set_all_funcs(child, nil, unsafe.Pointer(C.sub_455CD0), nil)
		nox_xxx_wndWddSetTooltip_46B000(&child.draw_data, g2ServerLoad("FlagHomeTT", clientCTFSource, 201))
	}
	sub_455A00(0)
	legacyPointers.flagTeamBorder = nox_xxx_gLoadImg_42F970(internCStr("FlagTeamBorder"))
	return 1
}

//export sub_455D80
func sub_455D80(index C.uchar, status C.char) {
	if index > 16 {
		return
	}
	*g2ServerByte(1045611 + uintptr(index)) = byte(status)
	child := g2ServerChild(serverSharedWindows.ctf_hud, int(index)+8810)
	if child == nil {
		return
	}
	key, line := "FlagHomeTT", 260
	switch status {
	case 0:
		if child.flags&C.NOX_WIN_LAYER_FRONT != 0 {
			line = 242
		}
	case 1:
		key, line = "TheirFlagCarriedTT", 252
		if child.flags&C.NOX_WIN_LAYER_FRONT != 0 {
			key, line = "YourFlagCarriedTT", 234
		}
	case 2:
		key, line = "FlagAwayTT", 256
		if child.flags&C.NOX_WIN_LAYER_FRONT != 0 {
			line = 238
		}
	default:
		return
	}
	nox_xxx_wndWddSetTooltip_46B000(&child.draw_data, g2ServerLoad(key, clientCTFSource, line))
}

//export sub_456140
func sub_456140(status C.uchar) {
	*g2ServerByte(1045644) = byte(status)
	root := serverSharedWindows.flagball_hud
	if root == nil {
		return
	}
	image, tooltip, line := "BallAtHome", "BallHomeTT", 165
	switch status {
	case 0:
	case 1:
		image, tooltip, line = "BallAway", "BallAwayTT", 170
	case 2:
		image, tooltip, line = "BallRed", "BallRedTT", 175
	case 4:
		image, tooltip, line = "BallBlue", "BallBlueTT", 180
	default:
		return
	}
	draw := &root.draw_data
	draw.bg_image = unsafe.Pointer(nox_xxx_gLoadImg_42F970(internCStr(image)))
	const source = `C:\\NoxPost\\src\\client\\Gui\\guifb.c`
	nox_xxx_wndWddSetTooltip_46B000(draw, g2ServerLoad(tooltip, source, line))
}
