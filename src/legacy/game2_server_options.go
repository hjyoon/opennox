package legacy

/*
#include "game2_server_bridge.h"
*/
import "C"
import (
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"unsafe"
)

func g2ServerAt(p *C.char, off uintptr) unsafe.Pointer { return unsafe.Add(unsafe.Pointer(p), off) }
func g2ServerMode(p *C.char) uint16                    { return *(*uint16)(g2ServerAt(p, 52)) }

//export nox_game2_server_limits_457460
func nox_game2_server_limits_457460(p unsafe.Pointer) C.int {
	g2ServerIntText(C.dword_5d4594_1046516, int32(*(*uint16)(unsafe.Add(p, 54))), 0)
	buf, free := g2ServerTemp(16)
	defer free()
	C.nox_itow(C.int(*(*byte)(unsafe.Add(p, 56))), (*C.wchar2_t)(unsafe.Pointer(&buf[0])), 10)
	return C.int(g2ServerText(C.dword_5d4594_1046520, 16414, unsafe.Pointer(&buf[0]), 0))
}

//export nox_xxx_guiServerOptionsTryHide_4574D0
func nox_xxx_guiServerOptionsTryHide_4574D0() C.int {
	if C.dword_5d4594_1046492 == nil {
		return 0
	}
	nox_xxx_guiServerOptionsHide_4597E0(0)
	C.dword_5d4594_1046492 = nil
	return 1
}

//export sub_457B60
func sub_457B60(input *C.char) C.int {
	buf, free := g2ServerTemp(100)
	defer free()
	C.game2_server_widen((*C.wchar2_t)(unsafe.Pointer(&buf[0])), C.nox_xxx_serverOptionsGetServername_40A4C0())
	g2ServerText(C.dword_5d4594_1046512, 16414, unsafe.Pointer(&buf[0]), 0)
	C.sub_459CD0()
	id := 10119
	if g2ServerGame(128) {
		id = 10122
	}
	g2ServerHidden(g2ServerChild(C.dword_5d4594_1046492, id), true)
	var data *C.char
	if g2ServerGame(1) {
		g2ServerEnabled(C.dword_5d4594_1046508, true)
		if g2ServerGame(128) {
			source := C.nox_xxx_cliGamedataGet_416590(1)
			mode := *(*uint16)(unsafe.Add(unsafe.Pointer(source), 52)) & 0x17F0
			C.nox_client_guiserv_updateMapList_458230(C.int(mode), source, true)
			data = C.sub_4165B0()
			cell := (*uint16)(g2ServerAt(data, 52))
			*cell = mode | (*cell & 0xE80F)
			if C.sub_40A740() != 0 {
				g2ServerEnabled(C.dword_5d4594_1046508, false)
			} else if bool(C.nox_xxx_CheckGameplayFlags_417DA0(4)) {
				g2ServerChild(C.dword_5d4594_1046508, 10330).draw_data.field_0 |= 4
			} else {
				C.sub_46AD20(C.dword_5d4594_1046508, 10331, 10333, 0)
			}
		} else {
			if bool(C.nox_xxx_CheckGameplayFlags_417DA0(4)) {
				C.sub_46AD20(C.dword_5d4594_1046508, 10330, 10331, 0)
			} else {
				C.sub_46AD20(C.dword_5d4594_1046508, 10330, 10333, 0)
			}
			data = input
			C.nox_client_guiserv_updateMapList_458230(C.int(g2ServerMode(input)), input, false)
		}
		g2ServerIntText(C.dword_5d4594_1046516, int32(uint16(C.nox_xxx_servGamedataGet_40A020(C.short(g2ServerMode(data))))), 0)
		g2ServerIntText(C.dword_5d4594_1046520, int32(C.sub_40A180(C.short(g2ServerMode(data)))), 0)
		C.sub_4580E0(data)
	} else {
		g2ServerEnabled(g2ServerChild(C.dword_5d4594_1046492, 10161), false)
		g2ServerEnabled(C.dword_5d4594_1046536, false)
		C.dword_5d4594_1046536.flags |= 8
		g2ServerEnabled(C.dword_5d4594_1046504, true)
		C.sub_46AD20(C.dword_5d4594_1046504, 10134, 10135, 0)
		C.sub_46ACE0(C.dword_5d4594_1046492, 10141, 10141, 1)
		data = input
		C.nox_client_guiserv_updateMapList_458230(C.int(g2ServerMode(input)), input, false)
		nox_game2_server_limits_457460(unsafe.Pointer(input))
	}
	if bool(C.nox_xxx_CheckGameplayFlags_417DA0(2)) {
		g2ServerChild(C.dword_5d4594_1046508, 10331).draw_data.field_0 |= 4
	}
	if bool(C.nox_xxx_CheckGameplayFlags_417DA0(1)) {
		g2ServerChild(C.dword_5d4594_1046508, 10333).draw_data.field_0 |= 4
	}
	name := C.nox_xxx_guiServerOptionsGetGametypeName_4573C0(C.short(g2ServerMode(data)))
	g2ServerText(g2ServerChild(C.dword_5d4594_1046492, 10119), 16385, unsafe.Pointer(name), 0)
	C.sub_459880(data)
	return sub_457F30(C.int(g2ServerMode(data) >> 12 & 1))
}

//export sub_457F30
func sub_457F30(value C.int) C.int {
	enabled := value != 1
	g2ServerEnabled(g2ServerChild(C.dword_5d4594_1046504, 10152), enabled)
	g2ServerEnabled(g2ServerChild(C.dword_5d4594_1046504, 10141), enabled)
	g2ServerEnabled(C.dword_5d4594_1046516, enabled)
	g2ServerEnabled(C.dword_5d4594_1046520, enabled)
	if value == 1 {
		w := g2ServerChild(C.dword_5d4594_1046492, 10122)
		if C.wndIsShown_nox_xxx_wndIsShown_46ACC0(w) == 0 {
			g2ServerHidden(w, true)
		}
	}
	return C.nox_window_set_hidden(g2ServerChild(C.dword_5d4594_1046492, 10183), value)
}

//export nox_xxx_windowServerOptionsDrawProc_458500
func nox_xxx_windowServerOptionsDrawProc_458500(w *C.nox_window, unused1 C.int) C.int {
	pos := C.nox_client_getMousePos_4309F0()
	var x, y C.uint
	C.nox_client_wndGetPosition_46AA60(w, &x, &y)
	C.nox_client_drawRectFilledAlpha_49CF10(C.int(x), C.int(y)+25, C.int(w.width), C.int(w.height)-25)
	v := g2ServerChild(C.dword_5d4594_1046492, 10120)
	if C.wndIsShown_nox_xxx_wndIsShown_46ACC0(v) == 0 && !bool(C.nox_xxx_wndPointInWnd_46AAB0(v, pos.x, pos.y)) {
		C.nox_xxx_wndClearCaptureMain_46ADE0(v)
		g2ServerHidden(v, true)
	}
	return 1
}

//export nox_xxx_guiServerOptionsProc_458590
func nox_xxx_guiServerOptionsProc_458590(unused1 *C.nox_window, event C.int, a, b C.uintptr_t) C.int {
	if event == 21 {
		if a != 1 {
			return 0
		}
		if b == 2 {
			C.nox_xxx_clientPlaySoundSpecial_452D80(231, 100)
			nox_xxx_guiServerOptionsHide_4597E0(0)
		}
	}
	return 1
}

//export sub_459150
func sub_459150() C.char {
	data := C.sub_4165B0()
	if *(*byte)(g2ServerAt(data, 52))&0x60 != 0 {
		for i := C.nox_xxx_getTeamCounter_417DD0(); i < 2; i++ {
			team := C.nox_xxx_teamCreate_4186D0(0)
			which := C.int(1)
			if i != 0 {
				which = 2
			}
			if C.sub_40A740() != 0 {
				C.sub_418800((*C.wchar2_t)(unsafe.Pointer(team)), (*C.wchar2_t)(g2ServerMem(1046564)), 0)
			} else {
				C.sub_418800((*C.wchar2_t)(unsafe.Pointer(team)), C.nox_server_teamTitle_418C20(which), 1)
			}
			team.def_ind = C.uint8_t(which)
			C.sub_4184D0(team)
		}
		C.nox_xxx_SetGameplayFlag_417D50(4)
	}
	result := C.char(*(*byte)(g2ServerAt(data, 57)))
	current := data
	if result != 0 {
		current = C.sub_409B80()
	}
	if *current == 0 {
		return result
	}
	different := C.nox_strcmpi(current, C.nox_xxx_mapGetMapName_409B40())
	C.sub_4165F0(1, 0)
	settings := C.sub_4165D0(0)
	C.nox_xxx_gameSetServername_40A440((*C.char)(g2ServerAt(settings, 9)))
	mode := g2ServerMode(settings)
	if mode&0x1000 == 0 {
		C.sub_409FB0_settings(C.short(mode), C.ushort(*(*uint16)(g2ServerAt(settings, 54))))
		C.sub_40A040_settings(C.short(mode), C.uchar(*(*byte)(g2ServerAt(settings, 56))))
	}
	if g2ServerGame(128) {
		C.nox_common_gameFlags_unset_40A540(49152)
		C.nox_xxx_setGameFlags_40A4D0(C.uint(mode & 0xC000))
	}
	*(*byte)(g2ServerAt(settings, 57)) = 0
	if different != 0 {
		filename, free := alloc.Make([]byte{}, 16)
		defer free()
		p := (*C.char)(unsafe.Pointer(&filename[0]))
		C.strcpy(p, current)
		off := uintptr(C.strlen(p))
		*(*uint32)(unsafe.Add(unsafe.Pointer(p), off)) = *(*uint32)(g2ServerMemOff(0x587000, 131668))
		*(*byte)(unsafe.Add(unsafe.Pointer(p), off+4)) = *(*byte)(g2ServerMemOff(0x587000, 131672))
		C.nox_xxx_mapLoad_4D2450(p)
		C.sub_416690()
		C.sub_4165D0(1)
		return C.char(uintptr(unsafe.Pointer(nox_xxx_guiServerOptionsHide_4597E0(0))))
	}
	C.nox_xxx_spellEnableAll_424BD0()
	C.sub_4537F0()
	C.sub_57AAA0(internCStr("user.rul"), C.nox_xxx_cliGamedataGet_416590(1), (*C.int)(unsafe.Pointer(&C.nox_gui_server_options_rules)))
	C.sub_57A950(C.nox_server_currentMapGetFilename_409B30())
	flags := C.short(C.nox_common_gameFlags_getVal_40A5B0())
	C.sub_57A1E0((*C.int)(unsafe.Pointer(C.nox_xxx_cliGamedataGet_416590(0))), internCStr("user.rul"), (*C.int)(unsafe.Pointer(&C.nox_gui_server_options_rules)), 3, flags)
	return C.char(uintptr(unsafe.Pointer(nox_xxx_guiServerOptionsHide_4597E0(0))))
}

//export sub_4593B0
func sub_4593B0(tab C.int) C.int {
	root := C.dword_5d4594_1046492
	if root == nil {
		return 0
	}
	v1, v2, v3 := g2ServerChild(root, 10163), g2ServerChild(root, 10161), g2ServerChild(root, 10162)
	if v1 == nil || v2 == nil || v3 == nil {
		return 0
	}
	if tab == 1 {
		v1.draw_data.bg_image = C.dword_5d4594_1046360
		v3.draw_data.bg_image = C.dword_5d4594_1046356
		if C.dword_5d4594_1046532 != nil {
			sub_456D60(1)
			C.dword_5d4594_1046532 = nil
		}
		if C.dword_5d4594_1046540 != nil {
			C.sub_4AD820()
			C.dword_5d4594_1046540 = nil
		}
		C.dword_5d4594_1046528 = nox_xxx_guiServerAccessLoad_4541D0(root)
		return sub_459560(10161)
	}
	if tab == 2 {
		v1.draw_data.bg_image = C.dword_5d4594_1046360
		if g2ServerGame(1) {
			v2.draw_data.bg_image = C.dword_5d4594_1046360
		} else {
			v2.draw_data.dis_image = C.dword_5d4594_1046360
		}
		if C.dword_5d4594_1046528 != nil {
			sub_4557D0(1)
			C.dword_5d4594_1046528 = nil
		}
		if C.dword_5d4594_1046540 != nil {
			C.sub_4AD820()
			C.dword_5d4594_1046540 = nil
		}
		C.dword_5d4594_1046532 = C.nox_xxx_guiServerPlayersLoad_456270(root)
		return sub_459560(10162)
	}
	if tab != 0 {
		return tab - 2
	}
	if g2ServerGame(1) {
		v2.draw_data.bg_image = C.dword_5d4594_1046356
	} else {
		v2.draw_data.dis_image = C.dword_5d4594_1046356
	}
	v3.draw_data.bg_image = C.dword_5d4594_1046356
	if C.dword_5d4594_1046532 != nil {
		sub_456D60(1)
		C.dword_5d4594_1046532 = nil
	} else if C.dword_5d4594_1046528 != nil {
		sub_4557D0(1)
		C.dword_5d4594_1046528 = nil
	}
	C.dword_5d4594_1046540 = C.nox_xxx_gui_4AD320(root)
	return sub_459560(10163)
}

//export sub_459560
func sub_459560(id C.int) C.int {
	root := C.dword_5d4594_1046492
	if root == nil {
		return 0
	}
	selected := g2ServerChild(root, int(id))
	if selected == nil || selected.parent == nil {
		return 0
	}
	parent := selected.parent
	var first, second *C.nox_window
	switch id {
	case 10161:
		first = g2ServerChild(root, 10163)
		second = g2ServerChild(root, 10162)
	case 10162:
		first = g2ServerChild(root, 10163)
		second = g2ServerChild(root, 10161)
	case 10163:
		first = g2ServerChild(root, 10162)
		second = g2ServerChild(root, 10161)
	default:
		return 0
	}
	if first == nil || second == nil {
		return 0
	}
	C.sub_46B120(first, nil)
	C.sub_46B120(first, parent)
	C.sub_46B120(second, nil)
	C.sub_46B120(second, parent)
	C.sub_46B120(selected, nil)
	return C.sub_46B120(selected, parent)
}

//export sub_459700
func sub_459700() C.int {
	C.sub_416580()
	data := C.sub_4165B0()
	index := int(g2ServerEvent(C.dword_5d4594_1046496, 16404, 0, 0))
	line := (*C.wchar2_t)(unsafe.Pointer(g2ServerEvent(C.dword_5d4594_1046496, 16406, uintptr(index), 0)))
	buf, free := g2ServerTemp(256)
	defer free()
	if line != nil {
		C.nox_wcsncpy((*C.wchar2_t)(unsafe.Pointer(&buf[0])), line, 255)
		buf[255] = 0
	}
	g2ServerEvent(C.dword_5d4594_1046496, 16398, uintptr(int32(index)), 0)
	g2ServerEvent(C.dword_5d4594_1046496, 16402, uintptr(int32(index)), 0)
	g2ServerText(C.dword_5d4594_1046496, 16397, unsafe.Pointer(&buf[0]), ^uintptr(0))
	g2ServerEvent(C.dword_5d4594_1046496, 16403, uintptr(int32(index)), 0)
	C.sub_57A9F0(data, internCStr("user.rul"))
	C.sub_57A1E0((*C.int)(unsafe.Pointer(data)), internCStr("user.rul"), nil, 5, C.short(g2ServerMode(data)))
	C.sub_459880(data)
	return sub_459D50(1)
}

//export nox_xxx_guiServerOptionsHide_4597E0
func nox_xxx_guiServerOptionsHide_4597E0(save C.int) *C.int {
	root := C.dword_5d4594_1046492
	if root != nil {
		focus := C.nox_xxx_wndGetFocus_46B4F0()
		if C.nox_window_is_child(root, focus) != 0 {
			C.nox_xxx_windowFocus_46B500(nil)
		}
		C.nox_xxx_windowDestroyMB_46C4E0(root)
		C.dword_5d4594_1046492 = nil
		sub_456D60(0)
		C.sub_4BE610()
		C.dword_5d4594_1046532 = nil
		sub_4557D0(0)
		C.dword_5d4594_1046528 = nil
		C.sub_4AD820()
		C.dword_5d4594_1046536 = nil
	}
	if save != 0 {
		result := C.sub_57ADF0((*C.int)(unsafe.Pointer(&C.nox_gui_server_options_rules)))
		C.dword_587000_129656 = 1
		return result
	}
	return nil
}

//export sub_459870
func sub_459870() *C.char { return (*C.char)(unsafe.Pointer(&C.nox_gui_server_options_rules)) }

//export sub_459A40
func sub_459A40(name *C.char) C.int {
	narrow, free := alloc.Make([]byte{}, 100)
	defer free()
	wide, freeWide := g2ServerTemp(100)
	defer freeWide()
	C.strncpy((*C.char)(unsafe.Pointer(&narrow[0])), name, 15)
	narrow[15] = 0
	C.game2_server_widen((*C.wchar2_t)(unsafe.Pointer(&wide[0])), (*C.char)(unsafe.Pointer(&narrow[0])))
	return C.int(g2ServerText(C.dword_5d4594_1046512, 16414, unsafe.Pointer(&wide[0]), 0))
}

//export nox_game2_server_apply_459AA0
func nox_game2_server_apply_459AA0(output unsafe.Pointer) C.intptr_t {
	narrow, free := alloc.Make([]byte{}, 100)
	defer free()
	buf := (*C.char)(unsafe.Pointer(&narrow[0]))
	name := (*C.wchar2_t)(unsafe.Pointer(g2ServerEvent(C.dword_5d4594_1046512, 16413, 0, 0)))
	C.game2_server_narrow_wide(buf, name)
	C.strncpy((*C.char)(unsafe.Add(output, 9)), buf, 15)
	*(*uint16)(unsafe.Add(output, 52)) = uint16(C.sub_459C10())
	copy(unsafe.Slice((*byte)(unsafe.Add(output, 24)), 20), unsafe.Slice((*byte)(unsafe.Pointer(sub_453F90())), 20))
	*(*uint32)(unsafe.Add(output, 44)) = *g2ServerWord(1045452)
	*(*uint32)(unsafe.Add(output, 48)) = *g2ServerWord(1045456)
	p := (*C.wchar2_t)(unsafe.Pointer(g2ServerEvent(C.dword_5d4594_1046516, 16413, 0, 0)))
	if *p != 0 {
		*(*uint16)(unsafe.Add(output, 54)) = uint16(C.nox_wcstol(p, nil, 10))
	}
	p = (*C.wchar2_t)(unsafe.Pointer(g2ServerEvent(C.dword_5d4594_1046520, 16413, 0, 0)))
	if *p != 0 {
		*(*byte)(unsafe.Add(output, 56)) = byte(C.nox_wcstol(p, nil, 10))
	}
	*(*byte)(unsafe.Add(output, 57)) = byte(g2ServerChild(C.dword_5d4594_1046492, 10122).draw_data.field_0>>2) & 1
	result := int(g2ServerEvent(C.dword_5d4594_1046496, 16404, 0, 0))
	if result >= 0 {
		p = (*C.wchar2_t)(unsafe.Pointer(g2ServerEvent(C.dword_5d4594_1046496, 16406, uintptr(result), 0)))
		C.game2_server_narrow_wide(buf, p)
		token := C.strtok(buf, internCStr("\t"))
		if token != nil {
			C.strcpy((*C.char)(output), token)
			return 0
		}
		*(*byte)(output) = 0
		return 0
	}
	*(*byte)(output) = 0
	return C.intptr_t(result)
}

//export sub_459C30
func sub_459C30() C.int {
	data := C.nox_xxx_cliGamedataGet_416590(1)
	nox_game2_server_copySpells_453F70(unsafe.Add(unsafe.Pointer(data), 24))
	sub_4535E0((*C.int)(unsafe.Add(unsafe.Pointer(data), 44)))
	sub_4535F0(C.int(*(*uint32)(unsafe.Add(unsafe.Pointer(data), 48))))
	if C.dword_5d4594_1046492 == nil {
		return 0
	}
	C.sub_459880(data)
	C.sub_4BDF70((*C.int)(unsafe.Pointer(data)))
	C.nox_client_guiserv_updateMapList_458230(C.int(*(*uint16)(unsafe.Add(unsafe.Pointer(data), 52))), data, false)
	nox_game2_server_limits_457460(unsafe.Pointer(data))
	name := C.nox_xxx_guiServerOptionsGetGametypeName_4573C0(C.short(*(*uint16)(unsafe.Add(unsafe.Pointer(data), 52))))
	return C.int(g2ServerText(g2ServerChild(C.dword_5d4594_1046492, 10119), 16385, unsafe.Pointer(name), 0))
}

//export sub_459D50
func sub_459D50(v C.int) C.int { *g2ServerWord(1046544) = uint32(v); return v }

//export sub_459D60
func sub_459D60() C.int { return C.int(*g2ServerWord(1046544)) }

//export sub_459D80
func sub_459D80(hidden C.int) C.int {
	if C.dword_5d4594_1046492 != nil {
		return C.nox_window_set_hidden(C.dword_5d4594_1046492, hidden)
	}
	return 0
}

//export sub_459DA0
func sub_459DA0() C.int { return g2ServerBool(C.dword_5d4594_1046492 != nil) }
