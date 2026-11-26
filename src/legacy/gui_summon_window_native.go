package legacy

/*
#include "cgo_const.h"
#include "GAME1_2.h"
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "GAME3_1.h"
#include "client__gui__window.h"
#include "common__strman.h"
extern int nox_win_width, nox_win_height;
extern uint32_t nox_xxx_screenWidth_587000_184452;
extern uint32_t dword_5d4594_1320988, dword_5d4594_1320992, dword_5d4594_1321196, dword_8531A0_2572;
extern nox_window *dword_5d4594_1321032, *dword_5d4594_1321036, *dword_5d4594_1321040, *dword_5d4594_1321044;
extern nox_video_bag_image_t* dword_5d4594_1321024;
extern uint32_t nox_color_white_2523948, nox_color_blue_2650684, nox_color_yellow_2589772;
extern uint32_t nox_color_black_2650656, nox_color_violet_2598268;
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

const summonSourceFile = "C:\\NoxPost\\src\\Client\\Gui\\guisumn.c"

func summonText(name *C.char, line C.int) *wchar2_t {
	return C.nox_strman_loadString_40F1D0(name, nil, internCStr(summonSourceFile), line)
}

func summonCommandText(command uintptr, line C.int) *wchar2_t {
	return summonText((*C.char)(g2ServerPointerSlot(0x587000, 184344+4*command)), line)
}

func summonImage(offset uintptr) *nox_video_bag_image_t {
	return (*nox_video_bag_image_t)(*memmap.PtrPtr(0x5D4594, offset))
}

func summonWindowPosition(win *nox_window) (x, y C.int) {
	var px, py C.uint
	C.nox_client_wndGetPosition_46AA60(win, &px, &py)
	return C.int(px), C.int(py)
}

func setSummonButtonImages(normal, lit *nox_video_bag_image_t) C.int {
	button := C.dword_5d4594_1321040
	nox_xxx_wndSetIcon_46AE60(button, normal)
	sub_46AEC0(button, lit)
	return nox_xxx_wndSetIconLit_46AEA0(button, lit)
}

//export sub_4C1CA0
func sub_4C1CA0(command C.int) C.int {
	*memmap.PtrUint32(0x587000, 184448) = uint32(command)
	switch command {
	case 3:
		return setSummonButtonImages(summonImage(1321020), summonImage(1321016))
	case 4:
		return setSummonButtonImages(summonImage(1321028), C.dword_5d4594_1321024)
	case 5:
		return setSummonButtonImages(summonImage(1321012), summonImage(1321008))
	default:
		return command - 5
	}
}

//export nox_xxx_guiSummonCreatureLoad_4C1D80
func nox_xxx_guiSummonCreatureLoad_4C1D80() C.int {
	x, y := C.nox_win_width-95, C.int(-145)
	*memmap.PtrInt32(0x5D4594, 1321004) = 0
	*memmap.PtrInt32(0x5D4594, 1321000) = int32(y)
	C.dword_5d4594_1320988, C.dword_5d4594_1320992 = C.uint32_t(x), C.uint32_t(y)
	root := nox_window_new(nil, 8, x, y, 87, 115, nil)
	C.dword_5d4594_1321032 = root
	nox_window_set_all_funcs(root, C.sub_4C2BD0, C.sub_4C24A0, nil)
	grid := nox_window_new(root, 136, 5, 38, 76, 76, nil)
	C.dword_5d4594_1321036 = grid
	nox_window_set_all_funcs(grid, C.nox_xxx_wndSummonProc_4C2B10, C.nox_xxx_guiDrawSummonBox_4C1FE0, C.sub_4C2C20)
	nox_xxx_wndWddSetTooltip_46B000(&grid.draw_data, summonText(internCStr("ToolTipSummon"), 818))
	*memmap.PtrPtr(0x5D4594, 1320996) = unsafe.Pointer(Nox_xxx_gLoadImg("CreatureCageBottom").C())
	top := nox_window_new(grid, 160, 0, 0, 1, 1, nil)
	nox_xxx_wndSetIcon_46AE60(top, nox_xxx_gLoadImg_42F970(internCStr("CreatureCageTop")))
	nox_xxx_wndSetOffsetMB_46AE40(top, -5, -38)
	cage := nox_window_new(root, 8, 19, 0, 48, 39, nil)
	nox_window_set_all_funcs(cage, C.sub_4C2BE0, C.sub_4C24A0, nil)
	names := [...]string{
		"CreatureCageHuntButtonLit", "CreatureCageHuntButton",
		"CreatureCageGuardButtonLit", "CreatureCageGuardButton",
		"CreatureCageEscortButtonLit", "CreatureCageEscortButton",
	}
	for i, name := range names {
		img := Nox_xxx_gLoadImg(name).C()
		if i == 4 {
			C.dword_5d4594_1321024 = (*nox_video_bag_image_t)(img)
		} else {
			*memmap.PtrPtr(0x5D4594, 1321008+4*uintptr(i)) = unsafe.Pointer(img)
		}
	}
	button := nox_window_new(nil, 168, x+27, y+12, 34, 34, nil)
	C.dword_5d4594_1321040 = button
	button.draw_data.group |= 0x100
	nox_window_set_all_funcs(button, C.nox_xxx_wndSummonBigButtonProc_4C24B0, nil, C.sub_4C2CE0)
	setSummonButtonImages(summonImage(1321028), C.dword_5d4594_1321024)
	nox_xxx_wndSetOffsetMB_46AE40(button, -27, -12)
	*memmap.PtrUint8(0x5D4594, 1321200) = 0
	g2ServerHidden(root, true)
	g2ServerHidden(button, true)
	for i := 0; i < summonRecordCount; i++ {
		summonRecordAt(i).active = 0
	}
	sub_4C2BF0()
	C.dword_5d4594_1321044, C.dword_5d4594_1321204, C.dword_5d4594_1321196 = nil, nil, 0
	return 1
}

func stepSummonCage(state byte, y, shown, hidden int32) (byte, int32) {
	// The original animation compares signed screen coordinates.
	switch state {
	case 1:
		y += 20
		if y >= shown {
			return 2, shown
		}
	case 3:
		y -= 20
		if y <= hidden {
			return 0, hidden
		}
	}
	return state, y
}

func updateSummonCage() (x, y int32) {
	root, button := C.dword_5d4594_1321032, C.dword_5d4594_1321040
	state := memmap.PtrUint8(0x5D4594, 1321200)
	prev := *state
	next, y := stepSummonCage(prev, int32(C.dword_5d4594_1320992), memmap.Int32(0x5D4594, 1321004), memmap.Int32(0x5D4594, 1321000))
	C.dword_5d4594_1320992 = C.uint32_t(y)
	*state = next
	if prev == 3 && next == 0 {
		g2ServerHidden(root, true)
		g2ServerHidden(button, true)
	}
	x = int32(C.dword_5d4594_1320988)
	nox_window_setPos_46A9B0(root, x, y)
	nox_window_setPos_46A9B0(button, x+27, y+12)
	return x, y
}

func drawSummonCell(record *summonRecord, windowX, windowY C.int) {
	if record.size != 1 && record.size != 2 && record.size != 4 {
		return
	}
	w, h := summonGridSpan(int32(record.size))
	x := windowX + 38*C.int(record.grid_x) + 2
	y := windowY + 38*C.int(record.grid_y) + 2
	width, height := 38*C.int(w)-4, 38*C.int(h)-4
	if record.highlight != 0 {
		C.nox_client_drawSetColor_434460(C.int(C.nox_color_yellow_2589772))
		record.highlight = 0
		C.nox_client_drawRectFilledOpaque_49CE30(x, y, width, height)
	} else if img := nox_xxx_guiDrawSummon_4C2440(C.int(record.thing_id)); img != nil {
		nox_client_drawImageAt_47D2C0(img, int32(x), int32(y))
	} else {
		cx, cy := int32(x+width/2), int32(y+height/2)
		C.nox_client_drawSetColor_434460(C.int(memmap.Int32(0x85B3FC, 956)))
		nox_client_drawPoint_4B0BC0(cx, cy, 9)
		nox_video_drawCircleColored_4C3270(cx, cy, 9, memmap.Int32(0x852978, 4))
	}
	var hp, maxHP C.uint16_t
	var alternate C.uint8_t
	if sub_495180(C.int(record.net_code), &hp, &maxHP, &alternate) == 0 {
		return
	}
	fill := C.int(0)
	if maxHP != 0 {
		fill = height * C.int(hp) / C.int(maxHP)
	}
	emptyColor, fillColor := C.int(C.nox_color_violet_2598268), C.int(memmap.Int32(0x85B3FC, 940))
	if alternate != 0 {
		emptyColor, fillColor = C.int(memmap.Int32(0x85B3FC, 984)), C.int(C.dword_8531A0_2572)
	}
	C.nox_client_drawSetColor_434460(emptyColor)
	C.nox_client_drawRectFilledOpaque_49CE30(x+width-2, y, 2, height)
	C.nox_client_drawSetColor_434460(fillColor)
	C.nox_client_drawRectFilledOpaque_49CE30(x+width-2, y+height-fill, 2, fill)
}

func summonRecordUnder(x, y, windowX, windowY C.int) *summonRecord {
	cell := summonGridCell(int32((x-windowX)/38), int32((y-windowY)/38))
	if cell == nil {
		return nil
	}
	return *cell
}

//export nox_xxx_guiDrawSummonBox_4C1FE0
func nox_xxx_guiDrawSummonBox_4C1FE0(win *nox_window, draw *C.nox_window_data) C.int {
	mouse := nox_client_getMousePos_4309F0()
	x, y := updateSummonCage()
	wx, wy := summonWindowPosition(win)
	nox_xxx_guiFontHeightMB_43F320(nil)
	if img := summonImage(1320996); img != nil {
		nox_client_drawImageAt_47D2C0(img, x, y)
	}
	for i := 0; i < summonRecordCount; i++ {
		if record := summonRecordAt(i); record.active != 0 {
			drawSummonCell(record, wx, wy)
		}
	}
	hoverFlag := memmap.PtrUint32(0x5D4594, 1321212)
	overButton := bool(C.nox_xxx_wndPointInWnd_46AAB0(C.dword_5d4594_1321040, mouse.x, mouse.y))
	if bool(C.nox_xxx_wndPointInWnd_46AAB0(win, mouse.x, mouse.y)) || overButton || *hoverFlag == 1 {
		hovered := summonRecordUnder(mouse.x, mouse.y, wx, wy)
		*hoverFlag = 0
		for record := sub_4C2D60(); record != nil; record = sub_4C2D90(record) {
			if sprite := nox_xxx_netSpriteByCodeDynamic_45A6F0(int32(record.net_code)); sprite != nil {
				if record == hovered || overButton {
					sprite.flags30 |= 0x40000000
					*hoverFlag = 1
				} else {
					sprite.flags30 &^= 0x40000000
				}
			}
		}
	}
	if popup := C.dword_5d4594_1321044; popup != nil && !C.nox_xxx_wndPointInWnd_46AAB0(popup, mouse.x, mouse.y) {
		nox_xxx_guiHideSummonWindow_4C2470()
	}
	return 1
}

//export nox_xxx_guiDrawSummon_4C2440
func nox_xxx_guiDrawSummon_4C2440(thing C.int) *nox_video_bag_image_t {
	name := nox_get_thing_name(int32(thing))
	if name == nil {
		return nil
	}
	return sub_427430(nox_xxx_guide_427010((*C.nox_cgo_const_char)(name)))
}

//export nox_xxx_guiHideSummonWindow_4C2470
func nox_xxx_guiHideSummonWindow_4C2470() C.int {
	nox_xxx_windowDestroyMB_46C4E0(C.dword_5d4594_1321044)
	C.dword_5d4594_1321044, C.dword_5d4594_1321204 = nil, nil
	nox_xxx_clientPlaySoundSpecial_452D80(920, 100)
	C.dword_5d4594_1321040.draw_data.field_0 &^= 2
	return 1
}

//export sub_4C24A0
func sub_4C24A0(win *nox_window, draw *C.nox_window_data) C.int { return 1 }

func summonPointerPosition(packed C.uintptr_t) C.int2 {
	return C.int2{field_0: C.int(uint16(packed)), field_4: C.int(uint16(packed >> 16))}
}

//export nox_xxx_wndSummonBigButtonProc_4C24B0
func nox_xxx_wndSummonBigButtonProc_4C24B0(win *nox_window, event C.int, packed, unused C.uintptr_t) C.int {
	switch event {
	case 5, 6:
		return 1
	case 7:
		C.dword_5d4594_1321204 = nil
		p := summonPointerPosition(packed)
		nox_xxx_wndSummonCreateList_4C2560(&p)
		C.dword_5d4594_1321040.draw_data.field_0 |= 2
		return 1
	case 17:
		C.dword_5d4594_1321040.draw_data.field_0 |= 2
	case 18:
		C.dword_5d4594_1321040.draw_data.field_0 &^= 2
	}
	return 0
}

var summonCommands = [...]uintptr{0, 1, 3, 4, 5}

//export nox_xxx_wndSummonCreateList_4C2560
func nox_xxx_wndSummonCreateList_4C2560(pos *C.int2) {
	C.nox_xxx_screenWidth_587000_184452 = 0
	for _, command := range summonCommands {
		text := summonCommandText(command, 588)
		var width, height C.int
		nox_xxx_drawGetStringSize_43F840(nil, text, &width, &height, int32(C.nox_win_width))
		if w := C.uint32_t(width); w > C.nox_xxx_screenWidth_587000_184452 {
			C.nox_xxx_screenWidth_587000_184452 = w
		}
	}
	C.nox_xxx_screenWidth_587000_184452 += 8
	row := C.int(nox_xxx_guiFontHeightMB_43F320(nil) + 2)
	width, height := C.int(C.nox_xxx_screenWidth_587000_184452), 5*row+12
	x := pos.field_0 - width/2
	y := pos.field_4 - height/2
	if x < 0 {
		x = 0
	} else if x+width >= C.nox_win_width {
		x = C.nox_win_width - width - 1
	}
	if y < 0 {
		y = 0
	} else if y+height >= C.nox_win_height {
		y = C.nox_win_width - height - 1
	}
	popup := nox_window_new(nil, 40, x, y, width, height, nil)
	C.dword_5d4594_1321044 = popup
	nox_window_set_all_funcs(popup, nil, C.sub_4C26F0, nil)
	nox_xxx_wndShowModalMB_46A8C0(popup)
	for i, command := range summonCommands {
		button := nox_window_new(popup, 8, 0, C.int(i)*(row+2), width, row+1, nil)
		nox_window_set_all_funcs(button, C.nox_xxx_clientOrderCreature_4C2A60, C.sub_4C27F0, nil)
		*(*uintptr)(unsafe.Pointer(&button.widget_data)) = command
	}
	nox_xxx_clientPlaySoundSpecial_452D80(791, 100)
}

func drawSummonOutlinedText(x, y, color, shadow C.int, text *wchar2_t) int32 {
	nox_xxx_drawSetTextColor_434390(int32(shadow))
	for i := uintptr(0); i < 4; i++ {
		dx := memmap.Int32(0x587000, 184520+8*i)
		dy := memmap.Int32(0x587000, 184524+8*i)
		nox_xxx_drawString_43F6E0(nil, text, int32(x)+dx, int32(y)+dy)
	}
	nox_xxx_drawSetTextColor_434390(int32(color))
	return nox_xxx_drawString_43F6E0(nil, text, int32(x), int32(y))
}

//export sub_4C2A00
func sub_4C2A00(x, y, color, shadow C.int, text *C.short) C.int {
	return C.int(drawSummonOutlinedText(x, y, color, shadow, (*wchar2_t)(unsafe.Pointer(text))))
}

//export sub_4C27F0
func sub_4C27F0(win *nox_window, draw *C.nox_window_data) C.int {
	command := uintptr(win.widget_data)
	summonCarnivorousPlant()
	selected := C.dword_5d4594_1321204
	if selected == nil && command == 1 {
		return 1
	}

	text := summonCommandText(command, 446)
	x, y := summonWindowPosition(win)
	var textWidth C.int
	nox_xxx_drawGetStringSize_43F840(nil, text, &textWidth, nil, 0)
	mouse := nox_client_getMousePos_4309F0()
	nox_xxx_guiFontHeightMB_43F320(nil)
	x += (C.int(C.nox_xxx_screenWidth_587000_184452)-textWidth)/2 + 1
	y += 3
	shadow := C.int(C.nox_color_black_2650656)
	if C.nox_xxx_wndPointInWnd_46AAB0(win, mouse.x, mouse.y) {
		drawSummonOutlinedText(x, y, C.int(C.nox_color_yellow_2589772), shadow, text)
		last := memmap.PtrUint32(0x587000, 184552)
		if command != uintptr(*last) {
			*last = uint32(command)
			nox_xxx_clientPlaySoundSpecial_452D80(920, 100)
		}
		return 1
	}
	color := C.int(memmap.Int32(0x85B3FC, 956))
	if selected != nil {
		if sub_4C2DD0((*C.nox_summon_const_record_t)(selected)) != 0 || (command != 4 && command != 5) {
			color = C.int(C.nox_color_white_2523948)
		}
	} else if (command != 4 && command != 5) || sub_4C2E00() != 0 {
		color = C.int(C.nox_color_blue_2650684)
	}
	drawSummonOutlinedText(x, y, color, shadow, text)
	return 1
}

//export nox_client_orderCreature
func nox_client_orderCreature(creature *summonRecord, command C.int) {
	packet := [4]byte{0x78, 0, 0, byte(command)}
	if creature != nil {
		binary.LittleEndian.PutUint16(packet[1:3], uint16(creature.net_code))
	} else if command == 1 {
		return
	}
	nox_netlist_addToMsgListCli_40EBC0(31, 0, (*C.uchar)(&packet[0]), int32(len(packet)))
	nox_xxx_guiHideSummonWindow_4C2470()
	if command == 0 {
		nox_xxx_clientPlaySoundSpecial_452D80(777, 100)
	} else {
		nox_xxx_clientPlaySoundSpecial_452D80(898, 100)
	}
}

//export nox_xxx_clientOrderCreature_4C2A60
func nox_xxx_clientOrderCreature_4C2A60(win *nox_window, event C.int, a, b C.uintptr_t) C.int {
	switch event {
	case 5, 6:
		return 1
	case 7:
		command := C.int(uintptr(win.widget_data))
		selected := C.dword_5d4594_1321204
		if command != 2 && (selected != nil || command != 1) {
			nox_client_orderCreature(selected, command)
		}
		return 1
	default:
		return 0
	}
}

//export nox_xxx_wndSummonProc_4C2B10
func nox_xxx_wndSummonProc_4C2B10(win *nox_window, event C.int, packed, unused C.uintptr_t) C.int {
	switch event {
	case 5, 6:
		return 1
	case 7:
		pos := summonPointerPosition(packed)
		wx, wy := summonWindowPosition(win)
		record := summonRecordUnder(pos.field_0, pos.field_4, wx, wy)
		C.dword_5d4594_1321204 = record
		if record != nil {
			nox_xxx_wndSummonCreateList_4C2560(&pos)
		}
		return 1
	default:
		return 0
	}
}

//export sub_4C2BD0
func sub_4C2BD0(win *nox_window, event C.int, a, b C.uintptr_t) C.int { return 0 }

//export sub_4C2BE0
func sub_4C2BE0(win *nox_window, event C.int, a, b C.uintptr_t) C.int { return 1 }

//export sub_4C2C20
func sub_4C2C20(win *nox_window, draw *C.nox_window_data, packed C.uintptr_t) C.int {
	pos := summonPointerPosition(packed)
	nox_xxx_cursorSetTooltip_4776B0(sub_4C2C60(win, (*C.nox_summon_const_position_t)(&pos)))
	return 1
}

//export sub_4C2C60
func sub_4C2C60(win *nox_window, pos *C.nox_summon_const_position_t) *wchar2_t {
	wx, wy := summonWindowPosition(win)
	if record := summonRecordUnder(pos.field_0, pos.field_4, wx, wy); record != nil {
		return nox_get_thing_pretty_name(int32(record.thing_id))
	}
	return nil
}

//export sub_4C2CE0
func sub_4C2CE0(win *nox_window, draw *C.nox_window_data, packed C.uintptr_t) C.int {
	switch memmap.Uint32(0x587000, 184448) {
	case 3:
		nox_xxx_cursorSetTooltip_4776B0(summonText(internCStr("ccs:GUARD"), 785))
	case 4:
		nox_xxx_cursorSetTooltip_4776B0(summonText(internCStr("ccs:ESCORT"), 782))
	case 5:
		nox_xxx_cursorSetTooltip_4776B0(summonText(internCStr("ccs:HUNT"), 788))
	}
	return 1
}
