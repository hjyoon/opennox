package legacy

/*
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2_1.h"
#include "GAME2_3.h"
#include "GAME3.h"
#include "GAME3_1.h"
#include "GAME3_2.h"
#include "client__gui__window.h"
#include "client__shell__optsback.h"
#include "client__shell__selcolor.h"
#include "common__savegame.h"
#include "noxstring.h"

extern nox_window* dword_5d4594_1308084;
extern nox_window* dword_5d4594_1308088;
extern nox_window* dword_5d4594_1308096;
extern nox_window* dword_5d4594_1308100;
extern nox_window* dword_5d4594_1308104;
extern nox_window* dword_5d4594_1308108;
extern nox_window* dword_5d4594_1308112;
extern nox_window* dword_5d4594_1308116;
extern nox_window* dword_5d4594_1308120;
extern nox_window* dword_5d4594_1308124;
extern nox_window* dword_5d4594_1308128;
extern nox_window* dword_5d4594_1308132;
extern nox_window* dword_5d4594_1308136;
extern nox_window* dword_5d4594_1308140;
extern nox_window* dword_5d4594_1308144;
extern nox_window* dword_5d4594_1308148;
extern nox_window* dword_5d4594_1308152;
extern nox_gui_animation* nox_wnd_xxx_1308092;
extern uint32_t dword_587000_171388;
extern char* dword_5d4594_1307784;
*/
import "C"

import (
	"fmt"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const colorSelectionStringsFile = "C:\\NoxPost\\src\\client\\shell\\SelColor.c"

func colorSelectionEquipment() [5]*nox_window {
	return [5]*nox_window{
		C.dword_5d4594_1308116, C.dword_5d4594_1308120, C.dword_5d4594_1308124,
		C.dword_5d4594_1308128, C.dword_5d4594_1308132,
	}
}

//export nox_client_selColorSetEquipmentPalette
func nox_client_selColorSetEquipmentPalette(index C.uchar) C.int {
	if index >= 32 {
		return 0
	}
	equipment := colorSelectionEquipment()
	for _, window := range equipment {
		if window == nil {
			return 0
		}
	}
	for _, window := range equipment {
		value := C.nox_selcolor_value(window)
		C.nox_selcolor_set_value(window, value&0xffff|C.uint32_t(index)<<16)
	}
	return 1
}

//export nox_game_showSelColor_4A5D00
func nox_game_showSelColor_4A5D00() C.int {
	nox_game_addStateCode_43BDD0(700)
	profile := nox_xxx_getHostInfoPtr_431770()
	C.dword_5d4594_1307784 = profile
	*(*byte)(unsafe.Add(unsafe.Pointer(profile), 67)) = 0
	events := unsafe.Pointer(C.sub_4A7330)
	root := nox_new_window_from_file(internCStr("SelColor.wnd"), events)
	C.dword_5d4594_1308084 = root
	if root == nil {
		return 0
	}
	nox_xxx_wndSetWindowProc_46B300_go(root, unsafe.Pointer(C.sub_4A18E0))
	ani := nox_gui_makeAnimation_43C5B0(root, 0, 0, 0, -440, 0, 20, 0, -40)
	C.nox_wnd_xxx_1308092 = ani
	if ani == nil {
		return 0
	}
	ani.field_0 = 700
	ani.field_12 = (*[0]byte)(C.sub_4A6890)
	ani.fnc_done_out = (*[0]byte)(C.sub_4A6C90)
	sub_4A5E90()
	for _, ids := range [...][2]int32{{720, 729}, {761, 792}} {
		for id := ids[0]; id <= ids[1]; id++ {
			window := nox_xxx_wndGetChildByID_46B0C0(root, id)
			nox_xxx_wndSetDrawFn_46B340_go(window, unsafe.Pointer(C.sub_4A6D20))
		}
	}
	if C.dword_587000_171388 != 0 {
		name := g2ServerLoad("DefaultName", colorSelectionStringsFile, 1138)
		g2ServerText(C.dword_5d4594_1308152, 16414, unsafe.Pointer(name), 0)
	}
	nox_xxx_wndRetNULL_46A8A0()
	palette := nox_xxx_wndGetChildByID_46B0C0(root, 760)
	C.dword_5d4594_1308088 = palette
	nox_xxx_wndSetProc_46B2C0_go(palette, events)
	nox_xxx_wndSetWindowProc_46B300_go(palette, unsafe.Pointer(C.sub_4A7270))
	sub_46B120(palette, nil)
	sub_4BFAD0()
	model := nox_xxx_wndGetChildByID_46B0C0(root, 740)
	nox_xxx_wndSetDrawFn_46B340_go(model, unsafe.Pointer(C.sub_4A6DC0))
	return 1
}

func colorSelectionName(destination *C.wchar2_t, emptyLine, invalidLine int) {
	window := C.dword_5d4594_1308152
	name := (*C.wchar2_t)(unsafe.Pointer(g2ServerEvent(window, 16413, 0, 0)))
	if name == nil {
		name = g2ServerLoad("DefaultName", colorSelectionStringsFile, emptyLine)
	} else if *name == 0 {
		C.nox_wcscpy(name, g2ServerLoad("DefaultName", colorSelectionStringsFile, emptyLine))
	}
	C.nox_wcscpy(destination, name)
	if sub_4A6B50(destination) == 0 {
		C.nox_wcscpy(destination, g2ServerLoad("DefaultName", colorSelectionStringsFile, invalidLine))
	}
}

func colorSelectionRGB(window *nox_window) [3]byte {
	index := C.nox_selcolor_palette_index(window)
	data := memmap.Slice(0x5D4594, 1307796+3*uintptr(index))
	return [3]byte{data[0], data[1], data[2]}
}

func captureColorSelection() [20]byte {
	var colors [20]byte
	primary := colorSelectionRGB(C.dword_5d4594_1308096)
	copy(colors[:3], primary[:])
	custom := [...]*nox_window{C.dword_5d4594_1308136, C.dword_5d4594_1308140, C.dword_5d4594_1308144, C.dword_5d4594_1308148}
	palette := [...]*nox_window{C.dword_5d4594_1308100, C.dword_5d4594_1308104, C.dword_5d4594_1308108, C.dword_5d4594_1308112}
	for i, toggle := range custom {
		color := primary
		if toggle != nil && toggle.flags&8 != 0 {
			color = colorSelectionRGB(palette[i])
		}
		copy(colors[3+3*i:6+3*i], color[:])
	}
	for i, window := range colorSelectionEquipment() {
		colors[15+i] = byte(C.nox_selcolor_value(window) >> 16)
	}
	return colors
}

//export sub_4A68C0
func sub_4A68C0() *C.wchar2_t {
	base := unsafe.Pointer(C.dword_5d4594_1307784)
	name := (*C.wchar2_t)(base)
	colorSelectionName(name, 225, 232)
	profile := unsafe.Slice((*byte)(base), 88)
	colors := captureColorSelection()
	for i, off := range [...]int{71, 68, 74, 77, 80} {
		copy(profile[off:off+3], colors[3*i:3*i+3])
	}
	copy(profile[83:88], colors[15:])
	return name
}

func colorSelectionSaveName(save *C.nox_savegame_xxx) string {
	if noxflags.HasGame(noxflags.GameModeCoop) {
		return "Player.plr"
	}
	prefix := alloc.GoStringS(unsafe.Slice((*byte)(unsafe.Pointer(&save.player_name[0])), 6))
	for i := range 100 {
		name := fmt.Sprintf("%s%02d.plr", prefix, i)
		encoded := CString(name)
		file := nox_fs_open(encoded)
		StrFree(encoded)
		if file == nil {
			return name
		}
		nox_fs_close(file)
	}
	return ""
}

//export sub_4A75C0
func sub_4A75C0() C.int {
	var save C.nox_savegame_xxx
	if noxflags.HasGame(noxflags.GameModeCoop) {
		nox_savegame_rm_4DBE10(internCStr("WORKING"), 0)
	}
	colorSelectionName(&save.player_name[0], 605, 612)
	save.field_1276 = 1
	save.player_class = *(*C.uint8_t)(unsafe.Add(unsafe.Pointer(C.dword_5d4594_1307784), 66))
	colors := captureColorSelection()
	for i, value := range colors {
		save.field_1204[i] = C.uint8_t(value)
	}

	root := GoString(nox_fs_root())
	path := root + "\\Save\\"
	if noxflags.HasGame(noxflags.GameModeCoop) {
		path += "WORKING\\"
	}
	pathBuffer := unsafe.Slice((*byte)(unsafe.Pointer(&save.path[0])), len(save.path))
	if len(path) >= len(pathBuffer) {
		return 0
	}
	alloc.StrCopy(pathBuffer, path)
	nox_fs_mkdir(&save.path[0])
	nox_fs_set_workdir(&save.path[0])

	filename := colorSelectionSaveName(&save)
	nox_fs_set_workdir(nox_fs_root())
	if filename == "" || len(path)+len(filename) >= len(pathBuffer) {
		return 0
	}
	alloc.StrCopy(pathBuffer, path+filename)

	if noxflags.HasGame(noxflags.GameModeCoop) {
		firstMap := [...]string{"War01a.map", "Wiz01a.map", "Con01a.map"}
		if class := int(save.player_class); class < len(firstMap) {
			nox_xxx_gameSetMapPath_409D70(internCStr(firstMap[class]))
		}
	}
	return sub_41CEE0(unsafe.Pointer(&save), 1)
}
