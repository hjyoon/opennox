package legacy

/*
#include <math.h>
#include <sys/stat.h>

#include "compat.h"

#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "GAME3.h"
#include "GAME3_1.h"
#include "GAME3_2.h"
#include "GAME4.h"
#include "GAME4_1.h"
#include "GAME5_2.h"
#include "client__draw__animdraw.h"
#include "client__drawable__drawable.h"
#include "client__video__draw_common.h"
#include "common__system__team.h"

#include "client__gui__guicon.h"
#include "client__gui__guiquit.h"
#include "client__gui__window.h"
#include "client__shell__mainmenu.h"
#include "client__shell__noxworld.h"
#include "client__shell__optsback.h"
#include "client__shell__selchar.h"
#include "client__shell__selclass.h"
#include "client__shell__selcolor.h"

#include "client__draw__drawrays.h"
#include "client__draw__fx.h"
#include "client__draw__lvupdraw.h"

#include "common/fs/nox_fs.h"
#include "common__binfile.h"
#include "common__crypt.h"
#include "input.h"
#include "input_common.h"
#include "operators.h"

extern uint32_t dword_8531A0_2572;
extern uint32_t dword_5d4594_1307792;
extern uint32_t dword_5d4594_1313788;
extern nox_window* dword_5d4594_1308136;
extern uint32_t nox_xxx_normalWndBits_587000_172880;
extern nox_window* dword_5d4594_1308104;
extern uint32_t dword_5d4594_1313532;
extern uint32_t dword_5d4594_1313564;
extern nox_window* dword_5d4594_1308096;
extern char* dword_5d4594_1307724;
extern nox_window* dword_5d4594_1308112;
extern uint32_t nox_server_sendMotd_108752;
extern nox_window* dword_5d4594_1308148;
extern uint32_t dword_587000_171388;
extern nox_window* dword_5d4594_1308140;
extern nox_window* dword_5d4594_1308144;
extern nox_window* dword_5d4594_1308152;
extern uint32_t dword_5d4594_1313536;
extern nox_window* dword_5d4594_1308116;
extern uint32_t dword_5d4594_1313740;
extern nox_window* dword_5d4594_1308100;
extern nox_window* dword_5d4594_1308132;
extern nox_window* dword_5d4594_1308108;
extern nox_window* dword_5d4594_1308120;
extern nox_window* dword_5d4594_1308128;
extern nox_window* dword_5d4594_1308124;
extern nox_window* dword_5d4594_1309736;
extern nox_window* dword_5d4594_1309756;
extern nox_window* dword_5d4594_1309832;
extern uint32_t dword_5d4594_1313540;
extern nox_window* dword_5d4594_1309824;
extern uint32_t dword_5d4594_1307720;
extern nox_window* dword_5d4594_1307736;
extern uint32_t nox_server_connectionType_3596;
extern void* dword_587000_122852;
extern nox_window* dword_5d4594_1309828;
extern nox_window* dword_5d4594_1309836;
extern nox_window* dword_5d4594_1309728;
extern nox_window* dword_5d4594_1309732;
extern uint64_t qword_581450_9552;
extern void* dword_587000_93164;
extern nox_window* dword_5d4594_1307716;
extern void* dword_587000_127004;
extern nox_window* dword_5d4594_1308088;
extern nox_window* dword_5d4594_1309748;
extern nox_window* dword_5d4594_1309720;
extern nox_window* dword_5d4594_1308084;
extern nox_window* dword_5d4594_1309812;
extern nox_window* dword_5d4594_1309820;
extern uint32_t dword_5d4594_2650652;
extern char* dword_5d4594_1307784;
extern int nox_win_width;
extern int nox_win_height;

extern uint32_t nox_color_white_2523948;
extern uint32_t nox_color_blue_2650684;
extern uint32_t nox_color_violet_2598268;
extern uint32_t nox_color_black_2650656;
extern uint32_t nox_color_red_2589776;
extern uint32_t nox_color_green_2614268;
extern uint32_t nox_color_cyan_2649820;
extern uint32_t nox_color_orange_2614256;
extern uint32_t nox_color_yellow_2589772;

extern nox_gui_animation* nox_wnd_xxx_1307732;
extern nox_window* nox_wnd_selclass_next;
extern nox_gui_animation* nox_wnd_xxx_1308092;
extern nox_gui_animation* nox_wnd_xxx_1309740;

extern void* dword_5d4594_1308156;
extern void* dword_5d4594_1308160;
extern void* dword_5d4594_1308164;


#include "game3_class_config_options.h"
extern int nox_game_showOptions_native(void);
void nox_video_setMenuOptions(nox_window*);
int nox_gui_menu_proc_ext(int);
void sub_4AA650(void);
static inline void* game3_native_word_ptr(uintptr_t word) { return (void*)word; }
*/
import "C"

import (
	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"strconv"
	"unsafe"
)

//export nox_game_showSelClass_4A4840
func nox_game_showSelClass_4A4840() C.int {
	C.sub_5007E0(internCStr("*:*"))
	C.sub_4A1BE0(1)
	C.dword_5d4594_1307724 = C.nox_xxx_getHostInfoPtr_431770()
	C.nox_game_addStateCode_43BDD0(600)
	C.dword_5d4594_1307736 = C.nox_new_window_from_file(internCStr("SelClass.wnd"), C.sub_4A4A20)
	if C.dword_5d4594_1307736 == nil {
		return 0
	}
	C.nox_xxx_wndSetWindowProc_46B300(C.dword_5d4594_1307736, C.sub_4A18E0)
	C.nox_wnd_xxx_1307732 = C.nox_gui_makeAnimation_43C5B0(C.dword_5d4594_1307736, 0, 0, 0, -460, 0, 20, 0, -40)
	if C.nox_wnd_xxx_1307732 == nil {
		return 0
	}
	a := asGUIAnim(C.nox_wnd_xxx_1307732)
	a.StateID = 600
	a.Func12Ptr = C.sub_4A4970
	a.FncDoneOutPtr = C.sub_4A49A0
	for _, id := range []int{601, 603, 602} {
		w := C.nox_xxx_wndGetChildByID_46B0C0(C.dword_5d4594_1307736, C.int(id))
		C.nox_xxx_wndSetDrawFn_46B340(w, C.sub_4A49D0)
	}
	C.nox_wnd_selclass_next = C.nox_xxx_wndGetChildByID_46B0C0(C.dword_5d4594_1307736, 610)
	C.nox_xxx_wndRetNULL_46A8A0()
	*memmap.PtrUint32(0x5D4594, 1307740) = 0
	C.sub_4A19F0(internCStr("OptsBack.wnd:Back"))
	C.sub_4602F0()
	return 1
}

//export sub_4A4970
func sub_4A4970() C.int {
	asGUIAnim(C.nox_wnd_xxx_1307732).SetState(gui.AnimOut)
	C.sub_43BE40(2)
	C.nox_xxx_clientPlaySoundSpecial_452D80(923, 100)
	return 1
}

//export sub_4A49A0
func sub_4A49A0() C.int {
	done := asGUIAnim(C.nox_wnd_xxx_1307732).Func13Ptr
	C.nox_gui_freeAnimation_43C570(C.nox_wnd_xxx_1307732)
	C.nox_xxx_windowDestroyMB_46C4E0(C.dword_5d4594_1307736)
	C.nox_wnd_xxx_1307732 = nil
	C.dword_5d4594_1307736 = nil
	C.nox_wnd_selclass_next = nil
	ccall.CallIntVoid(done)
	return 1
}

//export sub_4A49D0
func sub_4A49D0(win *C.nox_window, draw *C.nox_window_data) C.int {
	if memmap.Uint32(0x5D4594, 1307740) != uint32(C.nox_xxx_wndGetID_46B0A0(win)) {
		var x, y C.uint
		var w, h C.int
		C.nox_client_wndGetPosition_46AA60(win, &x, &y)
		C.nox_window_get_size(win, &w, &h)
		C.nox_client_drawRectFilledAlpha_49CF10(C.int(x), C.int(y), w, h)
	}
	return 1
}
func game3ColorSlot(i int) **C.nox_window {
	switch i {
	case 0:
		return &C.dword_5d4594_1308096
	case 1:
		return &C.dword_5d4594_1308100
	case 2:
		return &C.dword_5d4594_1308104
	case 3:
		return &C.dword_5d4594_1308108
	case 4:
		return &C.dword_5d4594_1308112
	case 5:
		return &C.dword_5d4594_1308116
	case 6:
		return &C.dword_5d4594_1308120
	case 7:
		return &C.dword_5d4594_1308124
	case 8:
		return &C.dword_5d4594_1308128
	case 9:
		return &C.dword_5d4594_1308132
	case 10:
		return &C.dword_5d4594_1308136
	case 11:
		return &C.dword_5d4594_1308140
	case 12:
		return &C.dword_5d4594_1308144
	case 13:
		return &C.dword_5d4594_1308148
	case 14:
		return &C.dword_5d4594_1308152
	}
	panic("invalid color window slot")
}

//export sub_4A5E90
func sub_4A5E90() C.int {
	sub_4A5E90_A()
	for i := 0; i < 10; i++ {
		p := game3ColorSlot(i)
		*p = C.nox_xxx_wndGetChildByID_46B0C0(C.dword_5d4594_1308084, C.int(720+i))
		value := uint32(589825)
		if i == 0 {
			value = 131074
		}
		if i >= 5 {
			value = uint32(memmap.Uint16(0x587000, 171372+uintptr(2*(i-5)))) << 16
		}
		C.nox_selcolor_set_value(*p, C.uint32_t(value))
	}
	for i, id := range []int{711, 712, 713, 714, 751} {
		*game3ColorSlot(10 + i) = C.nox_xxx_wndGetChildByID_46B0C0(C.dword_5d4594_1308084, C.int(id))
	}
	result := C.int(C.dword_587000_171388)
	if C.dword_587000_171388 == 0 {
		C.nox_window_call_field_94_fnc(C.dword_5d4594_1308152, 16414, C.uintptr_t(uintptr(unsafe.Pointer(C.dword_5d4594_1307784))), 0, internCStr("game3_class_config_native.go"), 0)
		for i, off := range []int{71, 68, 74, 77, 80} {
			group := C.int(1)
			if i == 0 {
				group = 2
			}
			game3_selcolor_match_4A61E0(*game3ColorSlot(i), group, (*C.uchar)(unsafe.Add(unsafe.Pointer(C.dword_5d4594_1307784), off)))
		}
		for i := 5; i < 10; i++ {
			result = C.int(uint32(*(*byte)(unsafe.Add(unsafe.Pointer(C.dword_5d4594_1307784), 78+i))) << 16)
			C.nox_selcolor_set_value(*game3ColorSlot(i), C.uint32_t(result))
		}
	}
	return result
}

//export game3_selcolor_match_4A61E0
func game3_selcolor_match_4A61E0(win *C.nox_window, group C.int, rgb *C.uchar) C.uintptr_t {
	index := uint32(0)
	p := memmap.PtrOff(0x5D4594, uintptr(int64(1307797)+96*int64(group)))
	result := uintptr(p)
	for {
		if *(*byte)(unsafe.Add(p, -1)) == byte(*rgb) && *(*byte)(p) == *(*byte)(unsafe.Add(unsafe.Pointer(rgb), 1)) && *(*byte)(unsafe.Add(p, 1)) == *(*byte)(unsafe.Add(unsafe.Pointer(rgb), 2)) {
			break
		}
		index++
		p = unsafe.Add(p, 3)
		result = uintptr(p)
		if index >= 32 {
			break
		}
	}
	if index < 32 && group == 1 {
		id := C.nox_xxx_wndGetID_46B0A0(win)
		w := C.nox_xxx_wndGetChildByID_46B0C0(C.dword_5d4594_1308084, id-10)
		C.nox_xxx_wnd_46ABB0(w, 1)
		w = C.nox_xxx_wndGetChildByID_46B0C0(C.dword_5d4594_1308084, id+10)
		if w != nil {
			w.draw_data.field_0 |= 6
		}
		result = uintptr(unsafe.Pointer(w))
	}
	if index == 32 && group == 1 {
		w := C.nox_xxx_wndGetChildByID_46B0C0(C.dword_5d4594_1308084, C.nox_xxx_wndGetID_46B0A0(win)-10)
		result = uintptr(int(C.nox_xxx_wnd_46ABB0(w, 0)))
		index = 9
	}
	C.nox_selcolor_set_value(win, C.uint32_t(uint32(uint16(group))|uint32(uint16(index))<<16))
	return C.uintptr_t(result)
}

//export sub_4A6890
func sub_4A6890() C.int {
	asGUIAnim(C.nox_wnd_xxx_1308092).SetState(gui.AnimOut)
	C.sub_43BE40(2)
	C.nox_xxx_clientPlaySoundSpecial_452D80(923, 100)
	C.sub_4A68C0()
	return 1
}

//export sub_4A6C90
func sub_4A6C90() C.int {
	done := asGUIAnim(C.nox_wnd_xxx_1308092).Func13Ptr
	C.nox_gui_freeAnimation_43C570(C.nox_wnd_xxx_1308092)
	C.nox_xxx_windowDestroyMB_46C4E0(C.dword_5d4594_1308084)
	C.nox_xxx_windowDestroyMB_46C4E0(C.dword_5d4594_1308088)
	C.nox_wnd_xxx_1308092 = nil
	C.dword_5d4594_1308084 = nil
	C.dword_5d4594_1308088 = nil
	if done != nil {
		ccall.CallIntVoid(done)
	} else {
		C.nox_client_resetScreenParticles_431510()
		C.nox_gui_draw()
		if !bool(C.nox_common_gameFlags_check_40A5C0(0x2000)) {
			C.nox_client_guiXxxDestroy_4A24A0()
			return 1
		}
		C.nox_xxx_serverHost_43B4D0()
	}
	return 1
}
func game3PaletteColor(win *C.nox_window) C.int {
	value := uint32(C.nox_selcolor_value(win))
	index := uint32(uint16(value>>16)) + 32*uint32(uint16(value))
	off := uintptr(1307796 + 3*index)
	return C.int(C.nox_color_rgb_4344A0(C.int(memmap.Uint8(0x5D4594, off)), C.int(memmap.Uint8(0x5D4594, off+1)), C.int(memmap.Uint8(0x5D4594, off+2))))
}
func game3ModifierPalette(mod unsafe.Pointer) {
	for i := 1; i < 8; i++ {
		rgb := uint32(C.nox_modifier_getColorRGB(mod, C.int(i)))
		C.nox_draw_setMaterial_4340A0(C.int(i), C.int(rgb&255), C.int((rgb>>8)&255), C.int((rgb>>16)&255))
	}
}

//export sub_4A6D20
func sub_4A6D20(win *C.nox_window, draw *C.nox_window_data) C.int {
	var x, y C.uint
	var w, h C.int
	C.nox_client_wndGetPosition_46AA60(win, &x, &y)
	C.nox_window_get_size(win, &w, &h)
	C.nox_client_drawSetColor_434460(game3PaletteColor(win))
	C.nox_client_drawRectFilledOpaque_49CE30(C.int(x), C.int(y), w, h)
	return 1
}

//export sub_4A6DC0
func sub_4A6DC0(win *C.nox_window, draw *C.nox_window_data) C.int {
	var x, y C.uint
	C.nox_client_wndGetPosition_46AA60(win, &x, &y)
	body := game3PaletteColor(C.dword_5d4594_1308096)
	C.nox_draw_setMaterial_4341D0(6, body)
	C.nox_draw_setMaterial_4341D0(1, game3PaletteColor(C.dword_5d4594_1308100))
	color := body
	if C.dword_5d4594_1308144.flags&8 != 0 {
		color = game3PaletteColor(C.dword_5d4594_1308108)
	}
	C.nox_draw_setMaterial_4341D0(2, color)
	C.nox_draw_setMaterial_4341D0(3, body)
	color = body
	if C.dword_5d4594_1308148.flags&8 != 0 {
		color = game3PaletteColor(C.dword_5d4594_1308112)
	}
	C.nox_draw_setMaterial_4341D0(4, color)
	color = body
	if C.dword_5d4594_1308140.flags&8 != 0 {
		color = game3PaletteColor(C.dword_5d4594_1308104)
	}
	C.nox_draw_setMaterial_4341D0(5, color)
	variant := uint32(*(*byte)(unsafe.Add(unsafe.Pointer(C.dword_5d4594_1307784), 67)))
	off := uintptr(24 + 4*variant)
	if C.dword_5d4594_1308136.flags&8 != 0 {
		off = uintptr(16 + 4*variant)
	}
	C.nox_client_drawImageAt_47D2C0((*C.nox_video_bag_image_t)(C.mem_getPtrValue(0x973A20, C.uintptr_t(off))), C.int(x), C.int(y))
	for layer := 0; layer < 26; layer++ {
		var mod unsafe.Pointer
		switch uint32(1) << layer {
		case 4:
			mod = C.dword_5d4594_1308156
			game3ModifierPalette(mod)
			C.nox_draw_setMaterial_4341D0(C.nox_modifier_getMaterial(mod), game3PaletteColor(C.dword_5d4594_1308116))
		case 1024:
			mod = C.dword_5d4594_1308160
			game3ModifierPalette(mod)
			C.nox_draw_setMaterial_4341D0(C.nox_modifier_getMaterial(mod), game3PaletteColor(C.dword_5d4594_1308120))
			C.nox_draw_setMaterial_4341D0(C.nox_modifier_getPriEnchant(mod), game3PaletteColor(C.dword_5d4594_1308124))
		case 1:
			mod = C.dword_5d4594_1308164
			game3ModifierPalette(mod)
			C.nox_draw_setMaterial_4341D0(C.nox_modifier_getMaterial(mod), game3PaletteColor(C.dword_5d4594_1308128))
			C.nox_draw_setMaterial_4341D0(C.nox_modifier_getEffectiveness(mod), game3PaletteColor(C.dword_5d4594_1308132))
		default:
			continue
		}
		C.nox_client_drawImageAt_47D2C0((*C.nox_video_bag_image_t)(C.mem_getPtrValue(0x973A20, C.uintptr_t(32+104*variant+4*uint32(layer)))), C.int(x), C.int(y))
	}
	return 1
}

//export sub_4A7270
func sub_4A7270(win *C.nox_window, event C.int, a3, a4 C.uintptr_t) C.int {
	if event != 5 {
		return 0
	}
	if bool(C.nox_xxx_wndPointInWnd_46AAB0(win, C.int(uint16(a3)), C.int(uintptr(a3)>>16))) {
		return 0
	}
	sub_4A72D0(0xDEAD)
	return 1
}

//export sub_4A72D0
func sub_4A72D0(index C.ushort) *C.nox_window {
	C.nox_xxx_wnd_46C6E0(C.dword_5d4594_1308088)
	C.nox_window_set_hidden(C.dword_5d4594_1308088, 1)
	if index < 32 {
		w := C.nox_xxx_wndGetChildByID_46B0C0(C.dword_5d4594_1308088, C.int(C.dword_5d4594_1307792))
		if w != nil {
			C.nox_selcolor_set_value(w, C.uint32_t(uint32(memmap.Uint16(0x5D4594, 1307788))|uint32(index)<<16))
		}
		return w
	}
	return nil
}

//export sub_4A7530
func sub_4A7530(group C.ushort) *C.nox_window {
	*memmap.PtrUint32(0x5D4594, 1307788) = uint32(group)
	var w *C.nox_window
	for id := 761; id <= 792; id++ {
		w = C.nox_xxx_wndGetChildByID_46B0C0(C.dword_5d4594_1308088, C.int(id))
		if w != nil {
			C.nox_selcolor_set_value(w, C.uint32_t(uint32(group)|uint32(id-761)<<16))
		}
	}
	return w
}

//export sub_4A7580
func sub_4A7580(x, y C.int) C.int {
	C.nox_xxx_wndShowModalMB_46A8C0(C.dword_5d4594_1308088)
	C.sub_46C690(C.dword_5d4594_1308088)
	return C.nox_window_setPos_46A9B0(C.dword_5d4594_1308088, C.int(int64(x)-int64(C.dword_5d4594_1308088.width)), C.int(int64(y)-int64(C.dword_5d4594_1308088.height)/2))
}

//export sub_4A7330
func sub_4A7330(win *C.nox_window, event C.int, a3, a4 C.uintptr_t) C.uintptr_t {
	if event == 16389 {
		C.nox_xxx_clientPlaySoundSpecial_452D80(920, 100)
		return 1
	}
	if event != 16391 {
		return 0
	}
	child := (*C.nox_window)(C.game3_native_word_ptr(a3))
	id := C.nox_xxx_wndGetID_46B0A0(child)
	switch {
	case id >= 720 && id <= 729:
		C.dword_5d4594_1307792 = C.uint32_t(id)
		group := C.ushort(0)
		if id == 720 {
			group = 2
		} else if id <= 724 {
			group = 1
		}
		sub_4A7530(group)
		sub_4A7580(C.int(uint16(a4)), C.int(uintptr(a4)>>16))
	case id >= 731 && id <= 734:
		w := C.nox_xxx_wndGetChildByID_46B0C0(C.dword_5d4594_1308084, id-20)
		if w != nil {
			C.nox_xxx_wnd_46ABB0(w, C.int((^uint32(w.flags)>>3)&1))
		}
	case id >= 761 && id <= 792:
		sub_4A72D0(C.ushort(id - 761))
	case id == 799:
		if memmap.Uint32(0x5D4594, 1308168) == 1 {
			C.nox_game_decStateInd_43BDC0()
		}
		C.nox_game_decStateInd_43BDC0()
		C.nox_game_decStateInd_43BDC0()
		C.dword_587000_171388 = 1
		if C.sub_4A75C0() != 0 {
			switch *(*byte)(unsafe.Add(unsafe.Pointer(C.dword_5d4594_1307784), 66)) {
			case 0:
				C.nox_xxx_gameSetMapPath_409D70(internCStr("war01a.map"))
			case 1:
				C.nox_xxx_gameSetMapPath_409D70(internCStr("wiz01a.map"))
			case 2:
				C.nox_xxx_gameSetMapPath_409D70(internCStr("con01a.map"))
			}
			C.sub_4A24C0(0)
			sub_4A6890()
			asGUIAnim(C.nox_wnd_xxx_1308092).Func13Ptr = nil
		}
	}
	C.nox_xxx_clientPlaySoundSpecial_452D80(921, 100)
	return 1
}

//export nox_xxx_compassGenStrings_4A9C80
func nox_xxx_compassGenStrings_4A9C80() C.int {
	*memmap.PtrUint32(0x5D4594, 1309664) = 0
	for i := 1; i <= 4; i++ {
		h := C.nox_xxx_gLoadImg_42F970(internCStr("Compass" + strconv.Itoa(i)))
		*memmap.PtrUint32(0x5D4594, 1309644+uintptr(4*(i-1))) = uint32(uintptr(unsafe.Pointer(h)))
	}
	for i := 1; i <= 32; i++ {
		h := C.nox_xxx_gLoadImg_42F970(internCStr("CompassMainArrow" + strconv.Itoa(i)))
		*memmap.PtrUint32(0x5D4594, 1309516+uintptr(4*(i-1))) = uint32(uintptr(unsafe.Pointer(h)))
	}
	return 1
}

//export nox_xxx_wndRadioButtonProc_4A84E0
func nox_xxx_wndRadioButtonProc_4A84E0(w *C.nox_window, e C.int, a, b C.uintptr_t) C.int {
	var event gui.WindowEvent
	switch e {
	case 5, 6, 7, 8:
		event = gui.AsWindowEvent(int(e), uintptr(a), 0)
	case 17, 18:
		event = &gui.RawEvent{Event: int(e), Arg1: uintptr(a)}
	case 21:
		event = gui.WindowKeyPress{Key: keybind.Key(a), Pressed: b == 2}
	default:
		return 0
	}
	return C.int(gui.EventRespInt(gui.RadioButtonProcNative(asWindow(w), event)))
}

//export nox_xxx_wndRadioButtonProcPre_4A93C0
func nox_xxx_wndRadioButtonProcPre_4A93C0(w *C.nox_window, e C.int, a, b C.uintptr_t) C.int {
	var event gui.WindowEvent
	switch e {
	case 23:
		event = gui.WindowFocus(a != 0)
	case 16385:
		// Preserve raw UTF-16 and the original63-word read bound.
		src := (*uint16)(C.game3_native_word_ptr(a))
		ended := false
		for i := 0; i < 63; i++ {
			var value uint16
			if !ended {
				value = *(*uint16)(unsafe.Add(unsafe.Pointer(src), 2*i))
				ended = value == 0
			}
			w.draw_data.text[i] = C.wchar2_t(value)
		}
		w.draw_data.text[63] = 0
		return 0
	case 16392:
		event = &gui.RawEvent{Event: int(e), Arg1: uintptr(a)}
	default:
		return 0
	}
	return C.int(gui.EventRespInt(gui.RadioButtonProcPre(asWindow(w), event)))
}

//export nox_xxx_wndRadioButtonSetAllFn_4A87E0
func nox_xxx_wndRadioButtonSetAllFn_4A87E0(w *C.nox_window) C.int {
	if w == nil {
		return -2
	}
	if uint32(w.flags)&0x80 != 0 {
		return C.nox_window_set_all_funcs(w, C.nox_xxx_wndRadioButtonProc_4A84E0, C.nox_xxx_wndRadioButtonDraw_4A8A20, nil)
	}
	return C.nox_window_set_all_funcs(w, C.nox_xxx_wndRadioButtonProc_4A84E0, C.nox_xxx_wndRadioButtonDrawNoImg_4A8820, nil)
}

//export nox_xxx_wndRadioButtonDrawNoImg_4A8820
func nox_xxx_wndRadioButtonDrawNoImg_4A8820(w *C.nox_window, d *C.nox_window_data) C.int {
	return C.int(gui.RadioButtonDrawNoImageNative(asWindow(w), asWindowData(d)))
}

//export nox_xxx_wndRadioButtonDraw_4A8A20
func nox_xxx_wndRadioButtonDraw_4A8A20(w *C.nox_window, d *C.nox_window_data) C.int {
	return C.int(gui.RadioButtonDrawImageNative(asWindow(w), asWindowData(d)))
}

//export nox_game_showOptions_4AA6B0
func nox_game_showOptions_4AA6B0() C.int { return C.nox_game_showOptions_native() }

//export sub_4AAA70
func sub_4AAA70() C.uintptr_t {
	if C.dword_5d4594_1309720 == nil {
		return 0
	}
	C.nox_video_setMenuOptions(C.dword_5d4594_1309720)
	id := C.int(331)
	if C.nox_video_getFullScreen() != 0 {
		id = 332
	}
	if w := C.nox_xxx_wndGetChildByID_46B0C0(C.dword_5d4594_1309720, id); w != nil {
		C.nox_window_call_field_94_fnc(w, 16392, 1, 0, internCStr("game3_options"), 0)
	}
	cut := C.nox_video_getCutSize_4766D0()
	*memmap.PtrUint32(0x587000, 172884) = uint32(cut)
	switch {
	case cut <= 69:
		id = 311
	case cut <= 79:
		id = 312
	case cut <= 89:
		id = 313
	default:
		id = 314
	}
	if w := C.nox_xxx_wndGetChildByID_46B0C0(C.dword_5d4594_1309720, id); w != nil {
		return C.nox_window_call_field_94_fnc(w, 16392, 1, 0, internCStr("game3_options"), 0)
	}
	return 0
}
func game3OptionsToggle(w *C.nox_window) {
	if w != nil {
		C.nox_window_call_field_93(w, 21, 28, 2)
	}
}

//export sub_4AABE0
func sub_4AABE0(win *C.nox_window, event C.int, sender *C.nox_window, arg C.uintptr_t) C.int {
	switch event {
	case 16389:
		C.nox_xxx_clientPlaySoundSpecial_452D80(920, 100)
		return 1
	case 16391:
		id := C.nox_xxx_wndGetID_46B0A0(sender)
		if id >= 380 {
			return C.nox_gui_menu_proc_ext(id)
		}
		if id > 341 {
			if id > 363 {
				if id == 2099 {
					sub_4AAA70()
				}
			} else if id == 363 {
				if C.sub_43DC30() == 1 {
					C.sub_43DC00()
				} else {
					C.sub_43DC10()
					value := *(*uint32)(unsafe.Add(C.dword_587000_93164, 4))
					C.sub_486320(C.dword_587000_93164, C.int(value>>16))
				}
			} else {
				switch id - 361 {
				case 0:
					if C.sub_453070() == 1 {
						C.sub_453050()
					} else {
						C.nox_xxx____setargv_9_453060()
					}
				case 1:
					if C.sub_44D990() == 1 {
						C.sub_44D960()
					} else {
						C.sub_44D970()
					}
				}
			}
		} else if id == 341 {
			C.sub_4AA9C0()
			a := asGUIAnim(C.nox_wnd_xxx_1309740)
			a.FncDoneOutPtr = C.sub_4AB0C0
			a.Func13Ptr = C.sub_4CB880
		} else {
			switch id {
			case 311:
				C.nox_video_setCutSize_4766A0(65)
			case 312:
				C.nox_video_setCutSize_4766A0(75)
			case 313:
				C.nox_video_setCutSize_4766A0(85)
			case 314:
				C.nox_video_setCutSize_4766A0(100)
			case 332:
				C.nox_xxx_normalWndBits_587000_172880 = 16
				C.nox_video_setFullScreen(1)
			case 333:
				C.nox_video_setFullScreen(1)
			case 331:
				C.nox_xxx_normalWndBits_587000_172880 = 8
				C.nox_video_setFullScreen(0)
			case 334:
				C.nox_video_setFullScreen(0)
			}
		}
		C.nox_xxx_clientPlaySoundSpecial_452D80(921, 100)
		return 1
	case 16393:
		id := C.nox_xxx_wndGetID_46B0A0(sender)
		active := C.nox_client_getWin1064916_46C720()
		switch id {
		case 351:
			C.sub_486320(C.dword_587000_127004, C.int(arg))
			if active != nil && active.parent == sender {
				return 0
			}
			if arg != 0 {
				if C.sub_453070() != 0 {
					C.nox_xxx_clientPlaySoundSpecial_452D80(768, 100)
					return 0
				}
				game3OptionsToggle(C.dword_5d4594_1309728)
				C.nox_xxx_clientPlaySoundSpecial_452D80(768, 100)
				return 0
			}
			if C.sub_453070() != 1 {
				return 0
			}
			game3OptionsToggle(C.dword_5d4594_1309728)
			return 0
		case 352:
			C.sub_486320(C.dword_587000_122852, C.int(arg))
			if arg != 0 {
				if C.sub_44D990() != 0 {
					C.sub_4AA650()
					return 0
				}
				game3OptionsToggle(C.dword_5d4594_1309732)
				C.sub_4AA650()
				return 0
			}
			if C.sub_44D990() != 1 {
				return 0
			}
			game3OptionsToggle(C.dword_5d4594_1309732)
			return 0
		case 353:
			C.sub_486320(C.dword_587000_93164, C.int(arg))
			if arg != 0 {
				if C.sub_43DC30() != 0 {
					return 0
				}
				game3OptionsToggle(C.dword_5d4594_1309736)
			} else {
				if C.sub_43DC30() != 1 {
					return 0
				}
				game3OptionsToggle(C.dword_5d4594_1309736)
			}
			return 0
		case 316:
			C.nox_video_setGammaSlider(C.int(arg))
			return 0
		case 318:
			C.nox_input_setSensitivity(C.powf(10, C.float(C.int(arg))/50-1))
			return 0
		}
		return 0
	case 16396:
		id := C.nox_xxx_wndGetID_46B0A0(sender) - 351
		if id != 0 {
			if id != 1 {
				return 0
			}
			C.sub_486320(C.dword_587000_122852, C.int(arg))
			if arg != 0 {
				if C.sub_44D990() == 0 {
					game3OptionsToggle(C.dword_5d4594_1309732)
				}
				C.sub_4AA650()
				return 0
			}
			if C.sub_44D990() != 1 {
				return 0
			}
			game3OptionsToggle(C.dword_5d4594_1309732)
			return 0
		}
		C.sub_486320(C.dword_587000_127004, C.int(arg))
		if arg != 0 {
			if C.sub_453070() != 0 {
				C.nox_xxx_clientPlaySoundSpecial_452D80(768, 100)
				return 0
			}
			game3OptionsToggle(C.dword_5d4594_1309728)
			C.nox_xxx_clientPlaySoundSpecial_452D80(768, 100)
			return 0
		}
		if C.sub_453070() != 1 {
			return 0
		}
		game3OptionsToggle(C.dword_5d4594_1309728)
		return 0
	}
	return 0
}

//export sub_4AB0C0
func sub_4AB0C0() C.int {
	done := asGUIAnim(C.nox_wnd_xxx_1309740).Func13Ptr
	C.nox_gui_freeAnimation_43C570(C.nox_wnd_xxx_1309740)
	C.nox_xxx_windowDestroyMB_46C4E0(C.dword_5d4594_1309720)
	C.nox_wnd_xxx_1309740 = nil
	C.dword_5d4594_1309720 = nil
	C.dword_5d4594_1309728 = nil
	C.dword_5d4594_1309732 = nil
	C.dword_5d4594_1309736 = nil
	ccall.CallIntVoid(done)
	return 1
}

//export sub_4AB260
func sub_4AB260() C.int {
	C.mem_setPtrValue(0x5D4594, 1309752, unsafe.Pointer(C.nox_xxx_gLoadImg_42F970(internCStr("DisconnectIcon"))))
	C.dword_5d4594_1309756 = C.nox_window_new(nil, 136, C.nox_win_width-50, C.nox_win_height/2+3, 50, 50, nil)
	C.nox_xxx_wndSetIcon_46AE60(C.dword_5d4594_1309756, (*C.nox_video_bag_image_t)(C.mem_getPtrValue(0x5D4594, 1309752)))
	C.nox_window_set_all_funcs(C.dword_5d4594_1309756, nil, C.sub_4AB420, nil)
	C.dword_5d4594_1309748 = C.nox_new_window_from_file(internCStr("discon.wnd"), C.sub_4AB390)
	C.nox_xxx_wndSetWindowProc_46B300(C.dword_5d4594_1309748, C.sub_4AB340)
	C.sub_46B120(C.dword_5d4594_1309748, nil)
	C.nox_window_setPos_46A9B0(C.dword_5d4594_1309748, C.int(int64(C.nox_win_width)/2-int64(C.dword_5d4594_1309748.end_x)/2), C.int(int64(C.nox_win_height)/2-int64(C.dword_5d4594_1309748.end_y)/2))
	return 1
}

//export sub_4AB340
func sub_4AB340(win *C.nox_window, event, a3, a4 C.int) C.int {
	if event != 21 {
		return 0
	}
	if a3 == 1 {
		return 1
	}
	if a3 == 57 {
		p := C.nox_client_getMousePos_4309F0()
		C.nox_window_call_field_93(win, 5, C.uintptr_t(uint32(p.x)|uint32(p.y)<<16), 0)
	}
	return 0
}

//export sub_4AB390
func sub_4AB390(win *C.nox_window, event C.int, sender *C.nox_window, arg C.int) C.int {
	if event == 23 {
		return 1
	}
	if event != 16391 {
		return 0
	}
	id := C.nox_xxx_wndGetID_46B0A0(sender) - 576
	if id == 0 {
		C.sub_43CF40()
		return 0
	}
	if id != 1 {
		return 0
	}
	C.sub_446380()
	if C.dword_5d4594_2650652 != 0 && C.sub_41E2F0() == 9 {
		C.sub_41F4B0()
		C.sub_41EC30()
		C.sub_446490(0)
		C.nox_xxx____setargv_4_44B000()
		sub_4AB4D0(0)
	} else {
		C.sub_43B750()
		sub_4AB4D0(0)
	}
	return 0
}

//export sub_4AB420
func sub_4AB420(win *C.nox_window, draw unsafe.Pointer) C.int {
	var x, y C.uint
	C.nox_client_wndGetPosition_46AA60(win, &x, &y)
	C.nox_client_drawImageAt_47D2C0((*C.nox_video_bag_image_t)(win.draw_data.bg_image), C.int(int64(x)+int64(win.draw_data.img_px)), C.int(int64(y)+int64(win.draw_data.img_py)))
	return 1
}

//export sub_4AB470
func sub_4AB470() C.int {
	C.nox_xxx_windowDestroyMB_46C4E0(C.dword_5d4594_1309748)
	C.nox_xxx_windowDestroyMB_46C4E0(C.dword_5d4594_1309756)
	C.dword_5d4594_1309756 = nil
	C.dword_5d4594_1309748 = nil
	return 0
}

//export sub_4AB4A0
func sub_4AB4A0(show C.int) C.int {
	if show != 0 {
		return C.nox_window_set_hidden(C.dword_5d4594_1309756, 0)
	}
	return C.nox_window_set_hidden(C.dword_5d4594_1309756, 1)
}

//export sub_4AB4D0
func sub_4AB4D0(show C.int) C.int {
	if show != 0 {
		C.nox_video_stopAllFades_44E040()
		C.nox_window_set_hidden(C.dword_5d4594_1309748, 0)
		C.nox_xxx_wndShowModalMB_46A8C0(C.dword_5d4594_1309748)
		C.sub_46C690(C.dword_5d4594_1309748)
		C.nox_xxx_windowFocus_46B500(C.dword_5d4594_1309748)
		return C.nox_xxx_wnd_46ABB0(C.dword_5d4594_1309748, 1)
	}
	C.nox_window_set_hidden(C.dword_5d4594_1309748, 1)
	C.nox_xxx_wnd_46C6E0(C.dword_5d4594_1309748)
	C.nox_xxx_windowFocus_46B500(nil)
	return C.nox_xxx_wnd_46ABB0(C.dword_5d4594_1309748, 0)
}
