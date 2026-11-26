package legacy

/*
#include "GAME2_1.h"
#include "client__gui__guimeter.h"

extern uint32_t dword_5d4594_1096252, dword_5d4594_1096256, dword_5d4594_1096260, dword_5d4594_1096264;
extern uint32_t dword_5d4594_1096272, dword_5d4594_1096276, dword_5d4594_1096280, dword_5d4594_1096284;
extern void* dword_5d4594_1096288;
extern int nox_win_width, nox_win_height;
*/
import "C"

import (
	"strconv"
	"unsafe"

	"github.com/opennox/libs/player"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const clientMeterSource = `C:\NoxPost\src\Client\Gui\guimeter.c`

var clientHealthManaImages [10]*nox_video_bag_image_t
var clientPoisonTube *nox_video_bag_image_t
var clientMeterRoot *nox_window
var clientMeterRecords, _ = alloc.New([7]C.nox_window_yyy{})
var clientPotionDrawables, _ = alloc.New([3]C.nox_drawable{})

func clientMeterString(key string, line int32) *wchar2_t {
	return nox_strman_loadString_40F1D0(internCStr(key), nil, internCStr(clientMeterSource), line)
}

//export nox_client_guiHealthManaState
func nox_client_guiHealthManaState(index C.int, state *C.nox_client_gui_health_mana_state) C.int {
	root := clientMeterRoot
	if state == nil || root == nil || index < 0 || index >= 2 {
		return 0
	}
	meter := &clientMeterRecords[index]
	if meter.win == nil {
		return 0
	}
	state.current, state.maximum = meter.field_1, meter.field_2
	state.primary_color, state.secondary_color = meter.color_1, meter.color_2
	state.poisoned = C.int(bool2int(C.dword_5d4594_1096264 != 0))
	state.meter_x, state.meter_y = meterWindowPosition(meter.win)
	state.root_x, state.root_y = meterWindowPosition(root)
	state.poison_tube_ready = C.int(bool2int(clientPoisonTube != nil))
	return 1
}

//export sub_4710B0
func sub_4710B0() C.int {
	current := sub_4615C0()
	if current == nil {
		nox_xxx_cursorSetTooltip_4776B0(clientMeterString("ToolTipCurWeapon", 881))
		return 1
	}
	text := (*wchar2_t)(memmap.PtrOff(0x5D4594, 1091968))
	nox_wcscpy(text, nox_xxx_clientAskInfoMb_4BF050(current))
	if current.flags29&0xc != 0 {
		quiverType := memmap.PtrUint32(0x5D4594, 1096292)
		if *quiverType == 0 {
			*quiverType = uint32(nox_xxx_getNameId_4E3AA0(internCStr("Quiver")))
		}
		if quiver := sub_461600(C.int(*quiverType)); quiver != nil {
			nox_wcscat(text, internWStr("\n"))
			nox_wcscat(text, nox_xxx_clientAskInfoMb_4BF050(quiver))
		}
	}
	nox_xxx_cursorSetTooltip_4776B0(text)
	return 1
}

func configureClientMeterWindow(win *nox_window, index int, procedure, draw unsafe.Pointer, label string, line int32) {
	if win == nil {
		return
	}
	nox_window_set_all_funcs(win, procedure, draw, nil)
	if label != "" {
		nox_xxx_wndWddSetTooltip_46B000(&win.draw_data, clientMeterString(label, line))
	}
	*(*uintptr)(unsafe.Pointer(&win.widget_data)) = uintptr(index)
}

func clientMeterWindow(parent *nox_window, flags, x, y, w, h C.int, index int,
	procedure, draw unsafe.Pointer, label string, line int32) *nox_window {
	win := nox_window_new(parent, flags, x, y, w, h, nil)
	configureClientMeterWindow(win, index, procedure, draw, label, line)
	return win
}

//export sub_471160
func sub_471160(parent *C.nox_window, x, y, w, h C.int) C.int {
	charges := &clientMeterRecords[5]
	icon := &clientMeterRecords[6]
	storeWindowReference(&charges.win, nox_window_new(parent, 1032, x, y, w, h, nil))
	storeWindowReference(&icon.win, nox_window_new(parent, 1032, x-17, y-15, 15, 15, nil))
	configureClientMeterWindow(charges.win, 5, nil, C.sub_471250, "ToolTipCharges", 921)
	configureClientMeterWindow(icon.win, 6, nil, C.sub_471450, "ToolTipCharges", 925)
	return C.int(bool2int(charges.win != nil))
}

func clientMeterBottle(root *nox_window, x C.int, index int, label string, line int32) {
	clientMeterWindow(root, 8, x, 166, 28, 30, index,
		C.nox_xxx_guiBottleSlotProc_471B90, C.nox_xxx_guiBottleSlotDrawFn_471A80, label, line)
}

func clientMeterAssignPotion(slot int, countOffset, typeOffset uintptr, thingType C.uint32_t) {
	*memmap.PtrUint16(0x5D4594, countOffset) = 0
	if thing := nox_get_thing(int32(thingType)); thing != nil {
		drawable := &clientPotionDrawables[slot]
		nox_drawable_link_thing(drawable, int32(thing.field_1c))
		drawable.flags30 |= 0x40000000
	}
	*memmap.PtrUint32(0x5D4594, typeOffset) = uint32(thingType)
}

//export nox_xxx_guiHealthManaInit_4714E0
func nox_xxx_guiHealthManaInit_4714E0() C.int {
	pl := Get_dword_8531A0_2576()
	if pl == nil {
		return 0
	}
	C.dword_5d4594_1096288 = nox_xxx_guiFontPtrByName_43F360(internCStr("small"))
	C.dword_5d4594_1096264, C.dword_5d4594_1096256, C.dword_5d4594_1096260 = 0, 0, 0
	width, height := C.nox_win_width, C.nox_win_height
	*memmap.PtrT[[6]uint32](0x5D4594, 1091908) = [6]uint32{0, 0, uint32(width), uint32(height), 0, 0}
	*memmap.PtrUint32(0x5D4594, 1091940) = uint32(width)
	*memmap.PtrUint32(0x5D4594, 1091944) = uint32(height)
	for i := range clientHealthManaImages {
		clientHealthManaImages[i] = nox_xxx_gLoadImg_42F970(internCStr("HealthMana" + strconv.Itoa(i+1)))
	}
	if red := memmap.PtrUint32(0x5D4594, 1096268); *red == 0 {
		*red = uint32(nox_xxx_getTTByNameSpriteMB_44CFC0(internCStr("RedPotion")))
		C.dword_5d4594_1096272 = C.uint32_t(nox_xxx_getTTByNameSpriteMB_44CFC0(internCStr("BluePotion")))
		C.dword_5d4594_1096276 = C.uint32_t(nox_xxx_getTTByNameSpriteMB_44CFC0(internCStr("CurePoisonPotion")))
		C.dword_5d4594_1096280 = C.uint32_t(nox_xxx_getTTByNameSpriteMB_44CFC0(internCStr("RedApple")))
		C.dword_5d4594_1096284 = C.uint32_t(nox_xxx_getTTByNameSpriteMB_44CFC0(internCStr("Meat")))
	}
	root := nox_window_new(nil, 136, width-91, height-201, 91, 201, nil)
	storeWindowReference(&clientMeterRoot, root)
	if root == nil {
		return 0
	}
	nox_xxx_wndSetIcon_46AE60(root, clientHealthManaImages[0])
	clientMeterBottle(root, 6, 2, "CurePoisonSlotTT", 1029)
	clientMeterAssignPotion(2, 1091384, 1091380, C.dword_5d4594_1096276)
	clientMeterBottle(root, 34, 0, "HealthSlotTT", 1060)
	*memmap.PtrUint16(0x5D4594, 1090312) = 0
	*memmap.PtrUint32(0x5D4594, 1090296) = 0
	*memmap.PtrUint32(0x5D4594, 1090308) = 0

	mana := pl.PlayerClass() != player.Warrior
	parentIcon, poisonIcon := "WarriorHealthTube", "WarriorPoisonTube"
	healthTipLine := int32(1173)
	floatingFlags := C.int(24)
	if mana {
		clientMeterBottle(root, 62, 1, "ManaSlotTT", 1083)
		clientMeterAssignPotion(1, 1090848, 1090844, C.dword_5d4594_1096272)
		parentIcon, poisonIcon = "HealthManaTubes", "PoisonTube"
		healthTipLine = 1135
		floatingFlags = 8
	}
	clientPoisonTube = nox_xxx_gLoadImg_42F970(internCStr(poisonIcon))
	panel := nox_window_new(root, 136, 0, 0, 91, 159, nil)
	nox_xxx_wndSetIcon_46AE60(panel, nox_xxx_gLoadImg_42F970(internCStr(parentIcon)))
	if mana {
		window := clientMeterWindow(panel, 8, 60, 34, 25, 125, 1,
			C.nox_xxx_guiHealthManaTubeProc_472100, C.nox_xxx_guiHealthManaTubeDraw_471D10,
			"ToolTipMana", 1122)
		storeWindowReference(&clientMeterRecords[1].win, window)
	} else {
		clientMeterRecords[1].win = nil
	}
	healthTube := clientMeterWindow(panel, 8, 34, 34, 25, 125, 0,
		C.nox_xxx_guiHealthManaTubeProc_472100, C.nox_xxx_guiHealthManaTubeDraw_471D10,
		"ToolTipHealth", healthTipLine)
	storeWindowReference(&clientMeterRecords[0].win, healthTube)
	healthBar := clientMeterWindow(nil, floatingFlags, 0, 0, 0, 0, 0,
		nil, C.nox_xxx_drawHealthManaBar_471C00, "", 0)
	storeWindowReference(&clientMeterRecords[2].win, healthBar)
	if mana {
		window := clientMeterWindow(nil, 8, 0, 0, 0, 0, 1,
			nil, C.nox_xxx_drawHealthManaBar_471C00, "", 0)
		storeWindowReference(&clientMeterRecords[3].win, window)
		*memmap.PtrUint32(0x5D4594, 1093176) = 1
	} else {
		clientMeterRecords[3].win = nil
	}
	sub_472280()
	nox_xxx_guiHealthManaColorInit_470B00()
	if noxflags.HasGame(noxflags.GameModeQuest) {
		C.dword_5d4594_1096252 = 1
	} else {
		C.dword_5d4594_1096252 = 0
		nox_window_set_hidden(clientMeterRecords[2].win, 1)
		nox_window_set_hidden(clientMeterRecords[3].win, 1)
	}
	return 1
}

type clientMeterBubble struct {
	x, y   int32
	size   int32
	speed  int32
	active uint32
	color  uint32
}

func clientMeterDrawBubbles(index int, left, top, level int32, color uint32) {
	bubbles := memmap.PtrT[[64]clientMeterBubble](0x5D4594, 1093180+1536*uintptr(index))
	for i := range bubbles {
		b := &bubbles[i]
		if b.active == 0 {
			continue
		}
		y := b.y >> 4
		if y < 125-level {
			b.active = 0
			continue
		}
		drawColor := b.color
		if C.dword_5d4594_1096264 != 0 {
			drawColor = color
		}
		nox_client_drawSetColor_434460(int32(drawColor))
		x := left + b.x
		if b.size <= 2 {
			nox_client_drawRectFilledOpaque_49CE30(x, top+y, b.size, b.size)
		} else {
			sub_4720C0(C.int(x), C.int(top+y))
		}
		b.y -= b.speed
	}
	if level <= 1 {
		return
	}
	file := internCStr(clientMeterSource)
	for i := range bubbles {
		b := &bubbles[i]
		if b.active != 0 {
			continue
		}
		switch size := nox_common_randomIntMinMax_415FF0(1, 100, file, 668); {
		case size >= 95:
			b.size = 3
		case size >= 80:
			b.size = 2
		default:
			b.size = 1
		}
		b.x = min(15-b.size, nox_common_randomIntMinMax_415FF0(0, 14, file, 676))
		b.y = 16 * (125 - b.size)
		b.speed = nox_common_randomIntMinMax_415FF0(4, 48, file, 684)
		b.active = 1
		shade := nox_common_randomIntMinMax_415FF0(0, 64, file, 688)
		if index != 0 {
			b.color = uint32(nox_color_rgb_4344A0(shade, shade, 255))
		} else {
			b.color = uint32(nox_color_rgb_4344A0(255, shade, shade))
		}
	}
}

func clientMeterDrawHealthIcon() {
	index := int32(C.dword_5d4594_1096256 >> 3)
	if int32(C.dword_5d4594_1096260) > 0 {
		C.dword_5d4594_1096256 += C.dword_5d4594_1096260
		C.dword_5d4594_1096260--
		index = int32(C.dword_5d4594_1096256) >> 3
		if index >= int32(len(clientHealthManaImages)) {
			index = 0
			C.dword_5d4594_1096256 = 0
		}
	}
	nox_xxx_wndSetIcon_46AE60(clientMeterRoot, clientHealthManaImages[index])
}

//export nox_xxx_guiHealthManaTubeDraw_471D10
func nox_xxx_guiHealthManaTubeDraw_471D10(win *C.nox_window, draw *C.nox_window_data) C.int {
	index := int(uintptr(win.widget_data))
	meter := &clientMeterRecords[index]
	if index == 0 && C.dword_5d4594_1096264 != 0 {
		x, y := meterWindowPosition(clientMeterRoot)
		nox_client_drawImageAt_47D2C0(clientPoisonTube, int32(x), int32(y))
	}
	x, y := meterWindowPosition(win)
	left, top := int32(x)+5, int32(y)
	if get_dword_5d4594_3799468() != 0 {
		nox_client_drawSetColor_434460(int32(Get_nox_color_black_2650656()))
		nox_client_drawRectFilledOpaque_49CE30(left, top, 15, 125)
	}
	if meter.field_2 == 0 {
		nox_client_drawRectFilledAlpha_49CF10(left, top, 15, 125)
		return 1
	}
	level := int32(125 * meter.field_1 / meter.field_2)
	surface := top + 125 - level
	nox_client_drawRectFilledAlpha_49CF10(left, top, 15, 125-level)
	nox_client_drawSetColor_434460(int32(meter.color_1))
	nox_client_drawEnableAlpha_434560(1)
	nox_client_drawRectFilledOpaque_49CE30(left, surface, 15, level)
	nox_client_drawEnableAlpha_434560(0)
	nox_client_drawAddPoint_49F500(left, surface)
	nox_xxx_rasterPointRel_49F570(14, 0)
	nox_client_drawLineFromPoints_49E4B0()
	if index < 2 && Get_nox_client_renderBubbles_80844() == 1 {
		clientMeterDrawBubbles(index, left, top, level, uint32(meter.color_1))
	}
	if index == 0 {
		clientMeterDrawHealthIcon()
	}
	sub_472080()
	return 1
}
