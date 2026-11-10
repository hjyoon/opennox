package legacy

/*
#include "game2_server_bridge.h"
*/
import "C"
import "unsafe"

//export nox_xxx_guiServerAccessLoad_4541D0
func nox_xxx_guiServerAccessLoad_4541D0(parent *C.nox_window) *C.nox_window {
	s := &C.nox_gui_server_access
	if s.root != nil {
		return nil
	}
	lang := int(C.nox_strman_get_lang_code())
	if C.nox_xxx_guiFontHeightMB_43F320(nil) > 10 {
		lang = 2
	}
	path := (*C.char)(g2ServerPointerSlot(0x587000, uintptr(127824+4*lang)))
	s.root = C.nox_new_window_from_file(path, unsafe.Pointer(C.nox_xxx_windowAccessProc_454BA0))
	if s.root == nil {
		return nil
	}
	C.nox_draw_setTabWidth_43FE20(100)
	C.sub_46B120(s.root, parent)
	s.wnd_10102 = g2ServerChild(s.root, 10102)
	s.wnd_10103 = g2ServerChild(s.root, 10103)
	s.wnd_10104 = g2ServerChild(s.root, 10104)
	s.wnd_10105 = g2ServerChild(s.root, 10105)
	s.wnd_10109 = g2ServerChild(s.root, 10109)
	s.wnd_10111 = g2ServerChild(s.root, 10111)
	s.wnd_10112 = g2ServerChild(s.root, 10112)
	s.wnd_10113 = g2ServerChild(s.root, 10113)
	s.wnd_10123 = g2ServerChild(s.root, 10123)
	s.wnd_10125 = g2ServerChild(s.root, 10125)
	s.wnd_10126 = g2ServerChild(s.root, 10126)
	s.wnd_10127 = g2ServerChild(s.root, 10127)
	s.wnd_10128 = g2ServerChild(s.root, 10128)
	s.wnd_10129 = g2ServerChild(s.root, 10129)
	s.wnd_10130 = g2ServerChild(s.root, 10130)
	s.wnd_10131 = g2ServerChild(s.root, 10131)
	s.wnd_10132 = g2ServerChild(s.root, 10132)
	s.wnd_10133 = g2ServerChild(s.root, 10133)
	s.wnd_10200 = g2ServerChild(s.root, 10200)
	slider := C.nox_xxx_gLoadImg_42F970(internCStr("UISlider"))
	lit := C.nox_xxx_gLoadImg_42F970(internCStr("UISliderLit"))
	for _, row := range []struct {
		list            *C.nox_window
		thumb, up, down int
	}{{s.wnd_10109, 10190, 10188, 10189}, {s.wnd_10105, 10187, 10185, 10186}, {s.wnd_10200, 10203, 10201, 10202}} {
		d := g2ServerList(row.list)
		thumb, up, down := g2ServerChild(s.root, row.thumb), g2ServerChild(s.root, row.up), g2ServerChild(s.root, row.down)
		if thumb.field_100 != nil {
			thumb.field_100.width = 16
			thumb.field_100.height = 10
		}
		C.sub_4B5700(thumb, nil, nil, unsafe.Pointer(slider), unsafe.Pointer(lit), unsafe.Pointer(lit))
		C.nox_xxx_wnd_46B280(thumb, row.list)
		C.nox_xxx_wnd_46B280(up, row.list)
		C.nox_xxx_wnd_46B280(down, row.list)
		d.Field_9 = unsafe.Pointer(thumb)
		d.Field_7 = unsafe.Pointer(up)
		d.Field_8 = unsafe.Pointer(down)
	}
	C.sub_454740()
	C.sub_454640()
	if g2ServerGame(1) {
		C.nox_xxx_wndSetDrawFn_46B340(s.root, unsafe.Pointer(C.sub_454A90))
	}
	C.nox_xxx_wndRetNULL_46A8A0()
	return s.root
}

//export sub_454A90
func sub_454A90(w *C.nox_window, d *C.nox_window_data) C.int {
	s := &C.nox_gui_server_access
	g2ServerDrawBackground(w, d, false)
	if s.wnd_10111.flags&8 != 0 {
		p := (*C.wchar2_t)(unsafe.Pointer(g2ServerEvent(s.wnd_10111, 16413, 0, 0)))
		has := p != nil && *p != 0
		if has && s.wnd_10112.flags&8 == 0 {
			g2ServerEnabled(s.wnd_10112, true)
		} else if !has && s.wnd_10112.flags&8 != 0 {
			g2ServerEnabled(s.wnd_10112, false)
		}
	}
	selected := int32(g2ServerEvent(s.active_list, 16404, 0, 0))
	if selected < 0 {
		if s.wnd_10113.flags&8 != 0 {
			g2ServerEnabled(s.wnd_10113, false)
		}
	} else if s.wnd_10113.flags&8 == 0 {
		g2ServerEnabled(s.wnd_10113, true)
	}
	return 1
}
func g2ServerAccessEdit(edit *C.nox_window, id int) {
	controls := unsafe.Pointer(C.sub_416630())
	settings := C.sub_416640()
	p := (*C.wchar2_t)(unsafe.Pointer(g2ServerEvent(edit, 16413, 0, 0)))
	var value int32
	if p != nil && *p != 0 {
		value = int32(C.nox_wcstol(p, nil, 10))
		if value < 0 {
			value = 0
		}
	}
	switch id {
	case 10104:
		if p != nil {
			C.nox_wcscpy((*C.wchar2_t)(unsafe.Add(settings, 78)), p)
		}
	case 10126, 10128:
		if value > 14 {
			value = 14
			g2ServerIntText(edit, value, ^uintptr(0))
		}
		b := (*byte)(unsafe.Add(controls, 1))
		if id == 10126 {
			*b = byte(value) | (*b & 0xF0)
		} else {
			*b = byte(value*16) | (*b & 0x0F)
		}
	case 10130:
		*(*uint16)(unsafe.Add(controls, 5)) = uint16(value)
	case 10132:
		*(*uint16)(unsafe.Add(controls, 7)) = uint16(value)
	case 10133:
		if value < 1 {
			value = 1
			g2ServerIntText(edit, value, ^uintptr(0))
		} else if value > 32 {
			value = 32
			g2ServerIntText(edit, value, ^uintptr(0))
		}
		C.nox_xxx_servSetPlrLimit_409F80(C.int(value))
		*(*byte)(unsafe.Add(controls, 4)) = byte(value)
	case 10136:
		C.nox_xxx_sysopSetPass_40A610(p)
	}
}

//export nox_xxx_windowAccessProc_454BA0
func nox_xxx_windowAccessProc_454BA0(unused1 *C.nox_window, event C.int, a, b C.uintptr_t) C.int {
	s := &C.nox_gui_server_access
	switch event {
	case 16387:
		if uint16(a) == 1 {
			return 0
		}
		edit := g2ServerChild(s.root, int(b))
		if edit == nil {
			return 0
		}
		g2ServerAccessEdit(edit, int(b))
		return 0
	case 16415:
		if a == 0 {
			return 0
		}
		edit := (*C.nox_window)(unsafe.Pointer(uintptr(a)))
		g2ServerAccessEdit(edit, int(edit.id))
		return 0
	case 16400:
		if a == 0 {
			return 0
		}
		id := (*C.nox_window)(unsafe.Pointer(uintptr(a))).id
		if id == 10200 {
			C.sub_46AD20(s.root, 10191, 10192, g2ServerBool(sub_455770() != 0))
		} else if id == 10123 {
			settings := C.sub_416640()
			*(*byte)(unsafe.Add(settings, 100)) ^= byte(uint32(1) << uint(uint32(b)&31))
		}
		return 0
	case 16391:
		if a == 0 {
			return 0
		}
		clicked := (*C.nox_window)(unsafe.Pointer(uintptr(a)))
		id := int(clicked.id)
		C.nox_xxx_clientPlaySoundSpecial_452D80(766, 100)
		var controls unsafe.Pointer
		switch id {
		case 10102:
			controls = unsafe.Pointer(C.sub_416630())
			*(*byte)(controls) ^= 0x10
			if *(*byte)(controls)&0x10 != 0 {
				g2ServerEnabled(g2ServerChild(s.root, 10206), true)
			} else {
				ban, allow := g2ServerChild(s.root, 10206), g2ServerChild(s.root, 10207)
				g2ServerEnabled(ban, false)
				ban.draw_data.field_0 &^= 4
				allow.draw_data.field_0 |= 4
				g2ServerHidden(s.wnd_10109, true)
				C.sub_46ACE0(s.root, 10188, 10190, 1)
				g2ServerHidden(s.wnd_10105, false)
				C.sub_46ACE0(s.root, 10185, 10187, 0)
			}
		case 10103:
			controls = unsafe.Pointer(C.sub_416630())
			g2ServerEnabled(s.wnd_10104, (s.wnd_10103.draw_data.field_0&4) == 0)
			*(*byte)(controls) ^= 0x20
		case 10112:
			text := (*C.wchar2_t)(unsafe.Pointer(g2ServerEvent(s.wnd_10111, 16413, 0, 0)))
			if C.wndIsShown_nox_xxx_wndIsShown_46ACC0(s.wnd_10105) != 0 {
				C.sub_4168A0(text)
			} else {
				C.sub_416770(0, text, nil)
			}
			g2ServerText(s.wnd_10111, 16414, g2ServerMem(1045600), 0)
		case 10113:
			if s.wnd_10102.draw_data.field_0&4 != 0 {
				selected := int32(g2ServerEvent(s.wnd_10109, 16404, 0, 0))
				g2ServerEvent(s.wnd_10109, 16398, uintptr(selected), 0)
				C.sub_416860(C.int(selected))
			} else {
				selected := int32(g2ServerEvent(s.wnd_10105, 16404, 0, 0))
				g2ServerEvent(s.wnd_10105, 16398, uintptr(selected), 0)
				C.sub_416820(C.int(selected))
			}
		case 10124:
			controls = unsafe.Pointer(C.sub_416630())
			*(*byte)(unsafe.Add(controls, 2)) ^= 0x80
		case 10125, 10127:
			controls = unsafe.Pointer(C.sub_416630())
			check, edit := s.wnd_10125, s.wnd_10126
			mask := byte(0x0F)
			if id == 10127 {
				check, edit = s.wnd_10127, s.wnd_10128
				mask = 0xF0
			}
			cell := (*byte)(unsafe.Add(controls, 1))
			if check.draw_data.field_0&4 != 0 {
				g2ServerEnabled(edit, false)
				*cell |= mask
				return 0
			}
			g2ServerEnabled(edit, true)
			text := (*C.wchar2_t)(unsafe.Pointer(g2ServerEvent(edit, 16413, 0, 0)))
			if text == nil || *text == 0 {
				return 0
			}
			v := int32(C.nox_wcstol(text, nil, 10))
			*cell = byte((uint32(*cell & ^mask)) | uint32(v))
		case 10129, 10131:
			controls = unsafe.Pointer(C.sub_416630())
			check, edit := s.wnd_10129, s.wnd_10130
			off := uintptr(5)
			if id == 10131 {
				check, edit = s.wnd_10131, s.wnd_10132
				off = 7
			}
			cell := (*uint16)(unsafe.Add(controls, off))
			if check.draw_data.field_0&4 != 0 {
				g2ServerEnabled(edit, false)
				*cell = 0xFFFF
				return 0
			}
			g2ServerEnabled(edit, true)
			text := (*C.wchar2_t)(unsafe.Pointer(g2ServerEvent(edit, 16413, 0, 0)))
			if text == nil || *text == 0 {
				return 0
			}
			*cell = uint16(C.nox_wcstol(text, nil, 10))
		case 10191, 10192:
			selections := (*int32)(unsafe.Pointer(g2ServerEvent(s.wnd_10200, 16404, 0, 0)))
			if selections == nil || *selections < 0 {
				return 0
			}
			for p := selections; *p >= 0; p = (*int32)(unsafe.Add(unsafe.Pointer(p), 4)) {
				name := (*C.wchar2_t)(unsafe.Pointer(g2ServerEvent(s.wnd_10200, 16406, uintptr(*p), 0)))
				player := g2ServerPlayerByName(name)
				if player != nil && player.playerInd != 31 {
					if g2ServerGame(4096) {
						C.sub_4DCFB0(player.playerUnit)
					} else if id == 10191 {
						C.nox_xxx_playerCallDisconnect_4DEAB0(C.int(player.playerInd), 4)
					} else {
						C.nox_xxx_playerDisconnByPlrID_4DEB00(C.int(player.playerInd))
					}
					if id == 10192 {
						C.sub_416770(0, name, &player.serial[0])
					}
				}
			}
		case 10206:
			g2ServerHidden(s.wnd_10109, false)
			C.sub_46ACE0(s.root, 10188, 10190, 0)
			g2ServerHidden(s.wnd_10105, true)
			C.sub_46ACE0(s.root, 10185, 10187, 1)
			s.active_list = s.wnd_10109
		case 10207:
			g2ServerHidden(s.wnd_10109, true)
			C.sub_46ACE0(s.root, 10188, 10190, 1)
			g2ServerHidden(s.wnd_10105, false)
			C.sub_46ACE0(s.root, 10185, 10187, 0)
			s.active_list = s.wnd_10105
		}
	}
	return 0
}

//export sub_455770
func sub_455770() C.int {
	s := &C.nox_gui_server_access
	if s.wnd_10200 == nil {
		return 0
	}
	p := (*int32)(unsafe.Pointer(g2ServerEvent(s.wnd_10200, 16404, 0, 0)))
	if p == nil {
		return 0
	}
	for ; *p >= 0; p = (*int32)(unsafe.Add(unsafe.Pointer(p), 4)) {
		name := (*C.wchar2_t)(unsafe.Pointer(g2ServerEvent(s.wnd_10200, 16406, uintptr(*p), 0)))
		player := g2ServerPlayerByName(name)
		if player != nil && player.playerInd != 31 {
			return 1
		}
	}
	return 0
}

//export sub_4557D0
func sub_4557D0(v C.int) C.int {
	s := &C.nox_gui_server_access
	result := v
	if v != 0 && s.root != nil {
		result = C.nox_xxx_windowDestroyMB_46C4E0(s.root)
	}
	*s = C.nox_gui_server_access_state{}
	return result
}

//export sub_455800
func sub_455800() *C.int {
	s := &C.nox_gui_server_access
	if s.root == nil {
		return nil
	}
	g2ServerIntText(s.wnd_10133, int32(C.nox_xxx_servGetPlrLimit_409FA0()), 0)
	if g2ServerGame(1) {
		g2ServerEvent(s.wnd_10105, 16399, 0, 0)
		g2ServerEvent(s.wnd_10109, 16399, 0, 0)
		for p := C.sub_4168E0(); p != nil; p = C.sub_4168F0(p) {
			g2ServerText(s.wnd_10109, 16397, unsafe.Pointer(&p.name[0]), ^uintptr(0))
		}
		for p := C.sub_416900(); p != nil; p = C.sub_416910(p) {
			if p.serial[0] != 0 {
				g2ServerText(s.wnd_10105, 16397, unsafe.Pointer(&p.name[0]), ^uintptr(0))
			} else {
				buf, free := g2ServerTemp(26)
				C.game2_server_format_wide((*C.wchar2_t)(unsafe.Pointer(&buf[0])), &p.name[0], 1)
				g2ServerText(s.wnd_10105, 16397, unsafe.Pointer(&buf[0]), ^uintptr(0))
				free()
			}
		}
	}
	return nil
}

//export sub_455920
func sub_455920(name *C.wchar2_t) C.int {
	s := &C.nox_gui_server_access
	if s.root == nil || s.wnd_10200 == nil {
		return 0
	}
	return C.int(g2ServerText(s.wnd_10200, 16397, unsafe.Pointer(name), 3))
}

//export sub_455950
func sub_455950(name *C.wchar2_t) {
	s := &C.nox_gui_server_access
	if s.root != nil {
		i := sub_4559B0(name)
		if i != -1 {
			g2ServerEvent(s.wnd_10200, 16398, uintptr(i), 0)
			if sub_455770() == 0 {
				C.sub_46AD20(s.root, 10191, 10192, 0)
			}
		}
	}
}

//export sub_4559B0
func sub_4559B0(name *C.wchar2_t) C.int {
	s := &C.nox_gui_server_access
	if s.wnd_10200 == nil {
		return -1
	}
	d := g2ServerList(s.wnd_10200)
	if d == nil || d.Field_11_0 == 0 {
		return -1
	}
	for i := 0; i < int(d.Field_11_0); i++ {
		p := (*C.wchar2_t)(unsafe.Pointer(g2ServerEvent(s.wnd_10200, 16406, uintptr(i), 0)))
		if p != nil && C.nox_wcscmp(p, name) == 0 {
			return C.int(i)
		}
	}
	return -1
}
