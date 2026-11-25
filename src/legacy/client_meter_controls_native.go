package legacy

/*
#include "GAME2_1.h"
#include "client__gui__guimeter.h"
#include "client__system__ctrlevnt.h"
#include "GAME2_3.h"

extern nox_window_yyy nox_windows_arr_1093036[7];
extern nox_drawable nox_gui_bottle_drawables[3];
extern nox_window* dword_5d4594_1090276;
extern uint32_t dword_5d4594_1096252;
extern uint32_t dword_5d4594_1090280, dword_5d4594_1090284;
extern uint32_t dword_5d4594_1096272, dword_5d4594_1096276;
extern uint32_t dword_5d4594_1096280, dword_5d4594_1096284;
extern unsigned int nox_player_netCode_85319C;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

const (
	healthQuickItem = 1090308
	manaQuickItem   = 1090844
	cureQuickItem   = 1091380
)

//export nox_xxx_cliShowHideTubes_470AA0
func nox_xxx_cliShowHideTubes_470AA0(visible C.int) {
	C.dword_5d4594_1096252 = C.uint32_t(visible)
	if memmap.Uint32(0x5D4594, 1093176) != 0 {
		hidden := int32(bool2int(visible == 0))
		nox_window_set_hidden(C.nox_windows_arr_1093036[2].win, hidden)
		nox_window_set_hidden(C.nox_windows_arr_1093036[3].win, hidden)
	}
}

//export nox_xxx_guiHealthManaColorInit_470B00
func nox_xxx_guiHealthManaColorInit_470B00() *C.uchar {
	red := nox_color_rgb_4344A0(255, 0, 0)
	darkRed := nox_color_rgb_4344A0(100, 0, 0)
	C.dword_5d4594_1090284, C.dword_5d4594_1090280 = red, darkRed
	*memmap.PtrUint32(0x5D4594, 1091964) = uint32(nox_color_rgb_4344A0(0, 255, 0))
	*memmap.PtrUint32(0x5D4594, 1092992) = uint32(nox_color_rgb_4344A0(0, 100, 0))

	health := &C.nox_windows_arr_1093036[0]
	health.color_1, health.color_2 = red, darkRed
	mana := &C.nox_windows_arr_1093036[1]
	mana.color_1 = nox_color_rgb_4344A0(0, 0, 255)
	mana.color_2 = nox_color_rgb_4344A0(0, 0, 100)
	for i := 4; i <= 6; i++ {
		level := int32(255)
		if i == 4 {
			level = 240
		}
		meter := &C.nox_windows_arr_1093036[i]
		meter.color_1 = nox_color_rgb_4344A0(level, 0, level)
		meter.color_2 = nox_color_rgb_4344A0(50, 0, 50)
	}
	for row := range uintptr(64) {
		*memmap.PtrUint32(0x5D4594, 1093196+24*row) = 0
		*memmap.PtrUint32(0x5D4594, 1094732+24*row) = 0
	}
	return (*C.uchar)(memmap.PtrOff(0x5D4594, 1096268))
}

//export sub_470D20
func sub_470D20(current, maximum C.int) C.int {
	meter := &C.nox_windows_arr_1093036[4]
	meter.field_1, meter.field_2 = C.uint32_t(current), C.uint32_t(maximum)
	if current != maximum {
		return C.int(nox_xxx_setKeybTimeout_4160D0(17))
	}
	return current
}

//export sub_470D70
func sub_470D70() {
	for i := 5; i <= 6; i++ {
		nox_window_set_hidden(C.nox_windows_arr_1093036[i].win, 1)
	}
}

//export sub_470D90
func sub_470D90(current, maximum C.int) C.int {
	for i := 5; i <= 6; i++ {
		nox_window_set_hidden(C.nox_windows_arr_1093036[i].win, 0)
	}
	for i := 5; i <= 6; i++ {
		meter := &C.nox_windows_arr_1093036[i]
		meter.field_1, meter.field_2 = C.uint32_t(current), C.uint32_t(maximum)
	}
	return current
}

//export sub_470DE0
func sub_470DE0() C.int {
	netcode := C.nox_player_netCode_85319C
	health := &C.nox_windows_arr_1093036[0]
	current, maximum := uint32(health.field_1), uint32(health.field_2)
	if netcode == 0 || current == 0 {
		return C.int(netcode)
	}
	threshold := 2 * maximum / 5
	if current >= threshold {
		return C.int(maximum) * C.int(-858993458)
	}
	rate := gameFPS()
	interval := rate/3 + current*((3*rate)>>2)/threshold
	*memmap.PtrUint32(0x5D4594, 1091960) = interval
	if !nox_xxx_checkKeybTimeout_4160F0(4, C.uint(interval-1)) {
		return 0
	}
	volume := 66*(threshold-current)/threshold + 33
	nox_xxx_clientPlaySoundSpecial_452D80(896, C.int(volume))
	return C.int(nox_xxx_setKeybTimeout_4160D0(4))
}

//export sub_470E90
func sub_470E90(_ *C.nox_window, event C.int) C.int {
	switch event {
	case 5:
		C.nox_client_invAlterWeapon_4672C0()
		return 1
	case 8, 12, 16:
		return 0
	default:
		return 1
	}
}

//export nox_win_init_cur_weapon
func nox_win_init_cur_weapon(parent *C.nox_window, x, y, width, height C.int) {
	win := nox_window_new_go(parent, 0x408, x, y, width, height, nil)
	C.nox_windows_arr_1093036[4].win = win
	if win != nil {
		nox_window_set_all_funcs_go(win, C.sub_470E90, C.sub_470F40_draw, C.sub_4710B0)
		*(*uintptr)(unsafe.Pointer(&win.widget_data)) = 4
	}
}

//export sub_472080
func sub_472080() C.int {
	meter := &C.nox_windows_arr_1093036[4]
	if meter.field_1 == meter.field_2 {
		return C.int(meter.field_1)
	}
	if !sub_416120(17) {
		return 0
	}
	add := 100 / gameFPS()
	meter.field_1 += C.uint32_t(add)
	return C.int(add)
}

//export sub_4720C0
func sub_4720C0(x, y C.int) C.int {
	C.nox_client_drawPixel_49EFA0(x+1, y)
	C.nox_client_drawRectFilledOpaque_49CE30(x, y+1, 3, 1)
	C.nox_client_drawPixel_49EFA0(x+1, y+2)
	return 0
}

//export nox_xxx_guiHealthManaTubeProc_472100
func nox_xxx_guiHealthManaTubeProc_472100(_ *C.nox_window, event C.int) C.int {
	switch event {
	case 7:
		previous := C.dword_5d4594_1096252
		C.dword_5d4594_1096252 = 1 - previous
		nox_window_set_hidden(C.nox_windows_arr_1093036[2].win, int32(bool2int(previous == 1)))
		if memmap.Uint8(0x85B3FC, 12254) != 0 {
			nox_window_set_hidden(C.nox_windows_arr_1093036[3].win, int32(bool2int(C.dword_5d4594_1096252 == 0)))
		}
		nox_xxx_clientPlaySoundSpecial_452D80(901, 100)
		return 1
	case 8, 12, 16:
		return 0
	default:
		return 1
	}
}

//export sub_4721A0
func sub_4721A0(visible C.int) C.int {
	return C.int(nox_window_set_hidden(C.dword_5d4594_1090276, int32(bool2int(visible == 0))))
}

//export nox_xxx_cliPrepareGameplay2_4721D0
func nox_xxx_cliPrepareGameplay2_4721D0() C.int {
	nox_xxx_windowDestroyMB_46C4E0(C.dword_5d4594_1090276)
	nox_xxx_windowDestroyMB_46C4E0(C.nox_windows_arr_1093036[2].win)
	if win := C.nox_windows_arr_1093036[3].win; win != nil {
		nox_xxx_windowDestroyMB_46C4E0(win)
	}
	C.nox_xxx_guiHealthManaInit_4714E0()
	sub_472310()
	return sub_4721A0(C.int(nox_client_getRenderGUI()))
}

func useQuickPotion(slot uintptr) {
	if nox_xxx_guiCursor_477600() == 0 {
		if item := memmap.Int32(0x5D4594, slot); item != 0 {
			C.nox_xxx_cliUseCurePoison_4674E0(C.int(item))
		}
	}
}

//export nox_client_quickHealthPotion_472220
func nox_client_quickHealthPotion_472220() {
	useQuickPotion(healthQuickItem)
}

//export nox_client_quickManaPotion_472240
func nox_client_quickManaPotion_472240() {
	useQuickPotion(manaQuickItem)
}

//export nox_client_quickCurePoisonPotion_472260
func nox_client_quickCurePoisonPotion_472260() {
	useQuickPotion(cureQuickItem)
}

func copyPotionShortcutLabel(key C.int, offset uintptr) *C.wchar2_t {
	label := C.sub_42E8E0(key, 1)
	dst := memmap.PtrT[[4]uint16](0x5D4594, offset)
	*dst = [4]uint16{}
	if label != nil {
		src := unsafe.Slice((*uint16)(unsafe.Pointer(label)), 3)
		copy(dst[:], src[:alloc.StrLenS(src)])
	}
	return (*C.wchar2_t)(unsafe.Pointer(dst))
}

//export sub_472280
func sub_472280() *C.wchar2_t {
	player := (*server.Player)(globalStateNativePointer(0))
	if player == nil {
		return nil
	}
	copyPotionShortcutLabel(38, cureQuickItem-8)
	copyPotionShortcutLabel(36, healthQuickItem-8)
	if player.PlayerClass() != 0 {
		return copyPotionShortcutLabel(37, manaQuickItem-8)
	}
	return (*C.wchar2_t)(unsafe.Pointer(player))
}

func linkQuickPotionDrawable(index int, thingID C.int) {
	drawable := &C.nox_gui_bottle_drawables[index]
	*drawable = C.nox_drawable{}
	if thingID == 0 {
		return
	}
	if thing := nox_get_thing(int32(thingID)); thing != nil {
		nox_drawable_link_thing(drawable, int32(thing.field_1c))
		drawable.flags30 |= 0x40000000
	}
}

//export sub_472310
func sub_472310() *C.uchar {
	cure := C.int(C.dword_5d4594_1096276)
	mana := C.int(C.dword_5d4594_1096272)
	*memmap.PtrUint16(0x5D4594, cureQuickItem+4) = uint16(C.sub_467850(cure))
	*memmap.PtrUint16(0x5D4594, manaQuickItem+4) = uint16(C.sub_467850(mana))

	healthCount := memmap.PtrUint16(0x5D4594, healthQuickItem+4)
	var health C.int
	for _, id := range [...]C.int{C.int(memmap.Uint32(0x5D4594, 1096268)), C.int(C.dword_5d4594_1096284), C.int(C.dword_5d4594_1096280)} {
		*healthCount = uint16(C.sub_467850(id))
		if *healthCount != 0 {
			health = id
			break
		}
	}
	*memmap.PtrUint32(0x5D4594, healthQuickItem) = uint32(health)
	*memmap.PtrUint32(0x5D4594, manaQuickItem) = uint32(mana)
	*memmap.PtrUint32(0x5D4594, cureQuickItem) = uint32(cure)
	linkQuickPotionDrawable(0, health)
	linkQuickPotionDrawable(1, mana)
	linkQuickPotionDrawable(2, cure)
	if health == 0 {
		return nil
	}
	return (*C.uchar)(unsafe.Pointer(&C.nox_gui_bottle_drawables[0]))
}
