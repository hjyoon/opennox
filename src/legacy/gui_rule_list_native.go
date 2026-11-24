package legacy

/*
#include "cgo_const.h"
#include "compat.h"
#include "GAME2.h"
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "GAME3_1.h"
#include "GAME3_2.h"
#include "GAME5_2.h"
#include "client__gui__servopts__guiserv.h"
#include "client__gui__window.h"
#include "client__video__draw_common.h"
extern nox_window *dword_5d4594_1523024, *dword_5d4594_1523028, *dword_5d4594_1523032;
extern nox_window *dword_5d4594_1523036, *dword_5d4594_1523040, *dword_5d4594_1523044, *dword_5d4594_1523048;
*/
import "C"

import (
	"strings"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func clearRuleListWindows() {
	C.dword_5d4594_1523024 = nil
	C.dword_5d4594_1523028 = nil
	C.dword_5d4594_1523032 = nil
	C.dword_5d4594_1523036 = nil
	C.dword_5d4594_1523040 = nil
	C.dword_5d4594_1523044 = nil
	C.dword_5d4594_1523048 = nil
}

//export sub_4CEBA0
func sub_4CEBA0(parent *nox_window, mapName *C.char) *nox_window {
	clearRuleListWindows()
	root := nox_new_window_from_file(internCStr("rulelist.wnd"), C.sub_4CF060)
	C.dword_5d4594_1523024 = root
	if root == nil {
		return nil
	}
	list := g2ServerChild(root, 10170)
	edit := g2ServerChild(root, 10171)
	save := g2ServerChild(root, 10172)
	replace := g2ServerChild(root, 10173)
	load := g2ServerChild(root, 10174)
	remove := g2ServerChild(root, 10175)
	C.dword_5d4594_1523028 = list
	C.dword_5d4594_1523032 = edit
	C.dword_5d4594_1523036 = save
	C.dword_5d4594_1523040 = replace
	C.dword_5d4594_1523044 = load
	C.dword_5d4594_1523048 = remove
	background := g2ServerChild(root, 10176)
	thumb := g2ServerChild(root, 10179)
	up := g2ServerChild(root, 10177)
	down := g2ServerChild(root, 10178)
	if list == nil || edit == nil || save == nil || replace == nil || load == nil || remove == nil ||
		background == nil || thumb == nil || up == nil || down == nil || list.widget_data == nil || thumb.field_100 == nil {
		clearRuleListWindows()
		C.nox_xxx_windowDestroyMB_46C4E0(root)
		return nil
	}
	nox_xxx_wndSetDrawFn_46B340_go(background, C.sub_4CEED0)
	sub_46B120(root, parent)
	slider := unsafe.Pointer(Nox_xxx_gLoadImg("UISlider").C())
	sliderLit := unsafe.Pointer(Nox_xxx_gLoadImg("UISliderLit").C())
	thumb.field_100.width, thumb.field_100.height = 16, 10
	sub_4B5700(thumb, nil, nil, slider, sliderLit, sliderLit)
	C.nox_xxx_wnd_46B280(thumb, list)
	C.nox_xxx_wnd_46B280(up, list)
	C.nox_xxx_wnd_46B280(down, list)
	data := (*C.nox_scrollListBox_data)(list.widget_data)
	data.field_9, data.field_7, data.field_8 = unsafe.Pointer(thumb), unsafe.Pointer(up), unsafe.Pointer(down)
	refreshRuleList4CED40(mapName)
	return C.dword_5d4594_1523024
}

//export nox_gui_refresh_rule_list_4CED40
func nox_gui_refresh_rule_list_4CED40(mapName *C.char) C.uintptr_t {
	return C.uintptr_t(refreshRuleList4CED40(mapName))
}

func refreshRuleList4CED40(mapName *C.char) uintptr {
	list := C.dword_5d4594_1523028
	if mapName == nil || list == nil {
		return ^uintptr(0)
	}
	g2ServerEvent(list, 16399, 0, 0)
	name := GoString(mapName)
	path := CString("maps\\" + name + "\\*.rul")
	defer StrFree(path)
	var entry C.struct__WIN32_FIND_DATAA
	handle := C.FindFirstFileA(path, &entry)
	if uintptr(unsafe.Pointer(handle)) == ^uintptr(0) {
		return ^uintptr(0)
	}
	for {
		if uint32(entry.dwFileAttributes)&C.FILE_ATTRIBUTE_DIRECTORY == 0 {
			filename := GoString(&entry.cFileName[0])
			if title := listedRuleName(name, filename); title != "" {
				addRuleListTitle(list, title)
			}
		}
		if C.FindNextFileA(handle, &entry) == 0 {
			break
		}
	}
	return uintptr(C.FindClose(handle))
}

func listedRuleName(mapName, fileName string) string {
	if len(fileName) <= 4 || !ruleNameEqual(fileName[len(fileName)-4:], ".rul") {
		return ""
	}
	title := fileName[:len(fileName)-4]
	if ruleNameEqual(title, mapName) || ruleNameEqual(title, "user") {
		return ""
	}
	return title
}

func ruleNameEqual(a, b string) bool {
	first, second := CString(a), CString(b)
	defer StrFree(first)
	defer StrFree(second)
	return C.nox_strcmpi(first, second) == 0
}

func addRuleListTitle(list *nox_window, title string) {
	wide, free := g2ServerTemp(len(title) + 1)
	defer free()
	for i := 0; i < len(title); i++ {
		wide[i] = uint16(title[i])
	}
	g2ServerText(list, 16397, unsafe.Pointer(&wide[0]), ^uintptr(0))
}

func setRuleListEnabled(win *nox_window, enabled bool) {
	if win != nil && (win.flags&8 != 0) != enabled {
		g2ServerEnabled(win, enabled)
	}
}

//export sub_4CEED0
func sub_4CEED0(win *nox_window, draw *C.nox_window_data) C.int {
	list, edit := C.dword_5d4594_1523028, C.dword_5d4594_1523032
	save, replace := C.dword_5d4594_1523036, C.dword_5d4594_1523040
	load, remove := C.dword_5d4594_1523044, C.dword_5d4594_1523048
	if win == nil || draw == nil || list == nil || edit == nil || save == nil || replace == nil || load == nil || remove == nil {
		return 1
	}
	var x, y C.uint
	C.nox_client_wndGetPosition_46AA60(win, &x, &y)
	if int8(win.flags) >= 0 {
		if uint32(draw.bg_color) != 0x80000000 {
			C.nox_client_drawRectFilledOpaque_49CE30(C.int(x), C.int(y), C.int(win.width), C.int(win.height))
		}
	} else if draw.bg_image != nil {
		C.nox_client_drawImageAt_47D2C0((*C.nox_video_bag_image_t)(draw.bg_image), C.int(x), C.int(y))
	}
	selected := int32(g2ServerEvent(list, 16404, 0, 0)) >= 0
	setRuleListEnabled(replace, selected)
	setRuleListEnabled(load, selected && !noxflags.HasGame(noxflags.GameFlag15|noxflags.GameFlag16))
	setRuleListEnabled(remove, selected)
	text := (*wchar2_t)(unsafe.Pointer(g2ServerEvent(edit, 16413, 0, 0)))
	if focus := nox_xxx_wndGetFocus_46B4F0(); focus != nil && focus.id == 10171 {
		setRuleListEnabled(save, text != nil && *text != 0)
	}
	return 1
}

func ruleListFilename(title *wchar2_t, suffixOffset uintptr) *C.nox_cgo_const_char {
	words := (*uint16)(unsafe.Pointer(title))
	length := alloc.StrLen(words)
	var filename strings.Builder
	for _, word := range unsafe.Slice(words, length) {
		filename.WriteByte(byte(word))
	}
	filename.WriteString(alloc.GoStringS(memmap.Slice(0x587000, suffixOffset)))
	return (*C.nox_cgo_const_char)(CString(filename.String()))
}

func ruleListText(win *nox_window, event int, index uintptr) *wchar2_t {
	return (*wchar2_t)(unsafe.Pointer(g2ServerEvent(win, event, index, 0)))
}

//export sub_4CF060
func sub_4CF060(win *nox_window, event C.uint, control *nox_window, eventArg C.uintptr_t) C.int {
	if event > 0x4007 {
		if event == 16400 && control != nil && uintptr(unsafe.Pointer(control)) != 1 {
			C.nox_xxx_wndGetID_46B0A0(control)
		}
		return 1
	}
	if event == 16387 {
		return editRuleListTitle(control, int32(eventArg))
	}
	list, edit := C.dword_5d4594_1523028, C.dword_5d4594_1523032
	if event != 16391 || control == nil || uintptr(unsafe.Pointer(control)) == 1 || list == nil || edit == nil {
		return 1
	}
	id := C.nox_xxx_wndGetID_46B0A0(control)
	nox_xxx_clientPlaySoundSpecial_452D80(766, 100)
	switch id {
	case 10172:
		sub_416580()
		title := ruleListText(edit, 16413, 0)
		settings := sub_4165B0()
		if title == nil || *title == 0 || settings == nil {
			return 1
		}
		filename := ruleListFilename(title, 191640)
		defer StrFree(filename)
		nox_game2_server_apply_459AA0(unsafe.Pointer(settings))
		sub_57AAA0(filename, settings, nil)
		if data := (*C.nox_scrollListBox_data)(list.widget_data); data != nil {
			for i := 0; i < int(data.field_11_0); i++ {
				item := ruleListText(list, 16406, uintptr(i))
				if item != nil && C._nox_wcsicmp(title, item) == 0 {
					g2ServerText(edit, 16414, unsafe.Pointer(memmap.PtrUint16(0x5D4594, 1523052)), 0)
					return 1
				}
			}
		}
		g2ServerText(list, 16397, unsafe.Pointer(title), ^uintptr(0))
		g2ServerText(edit, 16414, unsafe.Pointer(memmap.PtrUint16(0x5D4594, 1523056)), 0)
		setRuleListEnabled(C.dword_5d4594_1523036, false)
	case 10173:
		sub_416580()
		selected := int32(g2ServerEvent(list, 16404, 0, 0))
		if selected < 0 {
			return 1
		}
		title := ruleListText(list, 16406, uintptr(selected))
		settings := sub_4165B0()
		if title == nil || settings == nil {
			return 1
		}
		filename := ruleListFilename(title, 191592)
		defer StrFree(filename)
		nox_game2_server_apply_459AA0(unsafe.Pointer(settings))
		sub_57AAA0(filename, settings, nil)
		g2ServerEvent(list, 16403, ^uintptr(0), 0)
	case 10174:
		sub_416580()
		settings := sub_4165B0()
		selected := int32(g2ServerEvent(list, 16404, 0, 0))
		if selected < 0 || settings == nil {
			return 1
		}
		title := ruleListText(list, 16406, uintptr(selected))
		if title == nil {
			return 1
		}
		data := unsafe.Pointer(settings)
		filename := ruleListFilename(title, 191608)
		defer StrFree(filename)
		sub_57A1E0((*C.int)(data), filename, nil, 7, *(*C.short)(unsafe.Add(data, 52)))
		sub_453F70((*C.nox_cgo_const_void)(unsafe.Add(data, 24)))
		sub_4535E0((*C.int)(unsafe.Add(data, 44)))
		sub_4535F0(*(*C.int)(unsafe.Add(data, 48)))
		C.sub_459880(settings)
		g2ServerEvent(list, 16403, ^uintptr(0), 0)
		sub_459D50(1)
	case 10175:
		settings := sub_4165B0()
		selected := int32(g2ServerEvent(list, 16404, 0, 0))
		if selected < 0 || settings == nil {
			return 1
		}
		title := ruleListText(list, 16406, uintptr(selected))
		if title == nil {
			return 1
		}
		filename := ruleListFilename(title, 191624)
		defer StrFree(filename)
		sub_57A9F0((*C.nox_cgo_const_char)(unsafe.Pointer(settings)), filename)
		g2ServerEvent(list, 16398, uintptr(selected), 0)
	}
	return 1
}

func editRuleListTitle(control *nox_window, id int32) C.int {
	root := C.dword_5d4594_1523024
	if root == nil || uintptr(unsafe.Pointer(control)) == 1 {
		return 0
	}
	edit := g2ServerChild(root, int(id))
	if edit == nil {
		return 0
	}
	text := ruleListText(edit, 16413, 0)
	save := C.dword_5d4594_1523036
	if id == 10171 && text != nil && *text != 0 && save != nil {
		name := GoString(sub_4165B0())
		wide, free := g2ServerTemp(len(name) + 1)
		defer free()
		for i := 0; i < len(name); i++ {
			wide[i] = uint16(name[i])
		}
		if C._nox_wcsicmp(text, (*wchar2_t)(unsafe.Pointer(&wide[0]))) == 0 ||
			C._nox_wcsicmp(text, internWStr("user")) == 0 {
			setRuleListEnabled(save, false)
		}
	}
	return 1
}
