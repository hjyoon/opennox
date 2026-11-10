package legacy

/*
#include <math.h>
#include "audio_ambient_native_45a9b0.h"
#include <stdio.h>
#include <time.h>
#include "client/audio/ail/compat_mss.h"
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "quest_preview_44e110.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "GAME3.h"
#include "GAME3_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME4_1.h"
#include "GAME5_2.h"
#include "client__drawable__drawable.h"
#include "common__system__settings.h"
#include "common__system__team.h"
#include "client__gui__gadgets__listbox.h"
#include "client__gui__gamewin__gamewin.h"
#include "client__gui__gui_ctf.h"
#include "client__gui__guibook.h"
#include "client__gui__guibrief.h"
#include "client__gui__guicon.h"
#include "client__gui__guisave.h"
#include "client__gui__guispell.h"
#include "client__gui__servopts__access.h"
#include "client__gui__servopts__general.h"
#include "client__gui__servopts__guiserv.h"
#include "client__gui__servopts__playrlst.h"
#include "client__gui__window.h"
#include "client__shell__mainmenu.h"
#include "common__gamemech__pausefx.h"
#include "common__net_list.h"
#include "common__strman.h"
#include "common__crypt.h"
#include "operators.h"
#include "client__audio__audevent.h"
#include "client__video__draw_common.h"
#include "client__system__ctrlevnt.h"
#include "common__magic__speltree.h"
#include "input_common.h"
#include <string.h>
#include "game2_quickbar_native.h"
extern uintptr_t dword_8531A0_2576;
extern uint32_t dword_5d4594_1046852;
extern uint32_t dword_5d4594_1047532;
extern void* dword_5d4594_1049692;
extern uint32_t dword_5d4594_1046652;
extern uint32_t dword_5d4594_1047536;
extern uint32_t dword_5d4594_1045420;
extern char* dword_5d4594_831240;
extern uint32_t dword_5d4594_1045640;
extern uint32_t dword_5d4594_831260;
extern uint32_t dword_587000_129656;
extern uint32_t dword_5d4594_832480;
extern void* dword_5d4594_1046360;
extern uint32_t dword_587000_126996;
extern uint32_t dword_5d4594_1045428;
extern uint32_t dword_5d4594_831244;
extern uint32_t dword_5d4594_831076;
extern nox_window* dword_5d4594_1049524;
extern uint32_t dword_5d4594_1047524;
extern void* dword_5d4594_1046356;
extern uint32_t dword_5d4594_1046864;
extern uint32_t dword_5d4594_831256;
extern uint32_t dword_5d4594_1049536;
extern nox_window* dword_5d4594_1046944;
extern uint32_t dword_5d4594_1045424;
extern void* dword_5d4594_1049696;
extern nox_window* dword_5d4594_1046540;
extern uint32_t dword_5d4594_831276;
extern uint32_t dword_5d4594_1046648;
extern uint32_t dword_5d4594_1046640;
extern nox_window* dword_5d4594_1045692;
extern uint32_t dword_5d4594_831084;
extern nox_window* dword_5d4594_1046956;
extern uint32_t dword_5d4594_1047936;
extern float2 obj_5d4594_1046620;
extern nox_window* dword_5d4594_1046536;
extern uint32_t dword_5d4594_1045436;
extern unsigned int dword_5d4594_831220;
extern uint32_t dword_5d4594_1045460;
extern uint32_t dword_587000_122848;
extern uint32_t dword_5d4594_1046932;
extern nox_window* dword_5d4594_1046528;
extern nox_window* dword_5d4594_1046948;
extern uint32_t dword_5d4594_830864;
extern uint32_t dword_5d4594_1047528;
extern uint32_t dword_5d4594_1046636;
extern uintptr_t dword_5d4594_832520;
extern uintptr_t dword_5d4594_832500;
extern uintptr_t dword_5d4594_832528;
extern nox_window* dword_5d4594_1045688;
extern uintptr_t dword_5d4594_832524;
extern uintptr_t dword_5d4594_832512;
extern uintptr_t dword_5d4594_832496;
extern uint32_t nox_xxx_aNox_cfg_0_587000_132136;
extern uintptr_t dword_5d4594_832516;
extern nox_window* dword_5d4594_1046520;
extern nox_window* dword_5d4594_1049532;
extern uintptr_t dword_5d4594_832508;
extern uintptr_t dword_5d4594_832504;
extern uint32_t dword_5d4594_831224;
extern nox_things_imageRef_t* dword_5d4594_1046928;
extern nox_window* dword_5d4594_1045508;
extern nox_window* dword_5d4594_1046504;
extern uintptr_t dword_5d4594_832492;
extern uint32_t dword_5d4594_1047512;
extern uintptr_t dword_5d4594_832532;
extern uint32_t dword_5d4594_1046656;
extern uint32_t dword_5d4594_1049484;
extern nox_window* dword_5d4594_1046516;
extern uintptr_t dword_5d4594_832536;
extern nox_window* dword_5d4594_1046952;
extern void* dword_587000_81128;
extern void* dword_587000_122852;
extern nox_window* dword_5d4594_1046532;
extern uint32_t dword_5d4594_1049496;
extern uint32_t dword_5d4594_1047932;
extern nox_window* dword_5d4594_1049512;
extern nox_window* nox_wnd_briefing_831232;
extern uint32_t dword_5d4594_1045432;
extern nox_things_imageRef_t* dword_5d4594_1046924;
extern void* dword_587000_127004;
extern nox_window* dword_5d4594_1045468;
extern nox_window* dword_5d4594_1046508;
extern nox_window* dword_5d4594_1045480;
extern uint32_t dword_5d4594_1047520;
extern uint32_t nox_xxx_aNox_cfg_0_587000_132132;
extern nox_window* dword_5d4594_1045484;
extern nox_window* dword_5d4594_1045464;
extern nox_window* dword_5d4594_1049520;
extern uint32_t dword_5d4594_1046936;
extern uint32_t dword_5d4594_1047540;
extern nox_window* dword_5d4594_831236;
extern nox_window* dword_5d4594_1046496;
extern uint32_t dword_5d4594_1046872;
extern unsigned int nox_gameDisableMapDraw_5d4594_2650672;
extern nox_window* dword_5d4594_1045636;
extern nox_window* dword_5d4594_1049508;
extern uintptr_t dword_5d4594_1047516;
extern nox_window* dword_5d4594_1049500;
extern nox_window* dword_5d4594_1045604;
extern nox_video_bag_image_t* nox_ctf_flag_team_border;
extern uint32_t dword_5d4594_1046868;
extern nox_window* dword_5d4594_1049504;
extern uintptr_t dword_5d4594_832484;
extern nox_window* dword_5d4594_1045684;
extern void* nox_xxx_aClosewoodengat_587000_133480;
extern nox_window* dword_5d4594_1046492;
extern int nox_win_width;
extern int nox_win_height;
extern uint32_t nox_color_white_2523948;
extern uint32_t nox_color_black_2650656;
extern uint32_t nox_color_violet_2598268;
extern uint32_t nox_color_red_2589776;
extern uint32_t nox_color_green_2614268;
extern uint32_t nox_color_cyan_2649820;
extern uint32_t nox_color_blue_2650684;
extern uint32_t nox_color_orange_2614256;
extern uint32_t nox_color_yellow_2589772;
extern uint32_t dword_8531A0_2572;
extern void nox_xxx_clientPlaySoundSpecial_native_452D80(int a1, int a2);
extern nox_window* nox_win_unk1;
extern nox_window* nox_wnd_briefing_831232;
extern unsigned int nox_gameDisableMapDraw_5d4594_2650672;
extern uint32_t dword_5d4594_831092;
*/
import "C"

import (
	"github.com/opennox/opennox/v1/common/memmap"
	"unsafe"
)

// Serial test seams control external UI/audio services, never the C owners.
type game2QuickbarHooks struct {
	Sound         func(int32, int32)
	Destroy       func(unsafe.Pointer) int32
	Hidden        func(unsafe.Pointer, int32) int32
	Style         func(unsafe.Pointer, int32) int32
	SpellFlags    func(uint32, uint32) int32
	Tooltip       func(unsafe.Pointer, uint32)
	Capture       func() unsafe.Pointer
	SetCapture    func(unsafe.Pointer)
	ClearCapture  func(unsafe.Pointer)
	DragSave      func(uint32, int32)
	DragClear     func()
	Blocked       func() int32
	Enabled       func(uint32) bool
	PointIn       func(unsafe.Pointer, int32, int32) bool
	PointRect     func(int32, int32) bool
	Drop          func(uint32, byte, int32, int32, unsafe.Pointer) uint32
	TargetClick   func(uint32)
	StoreButton   func(int32)
	AbilityCursor func(uint32) int32
	SpellCursor   func(uint32, byte) int32
	EventSwap     func(unsafe.Pointer, int32, int32)
	EventUpdate   func(unsafe.Pointer)
	CancelAbility func()
}

var game2QuickbarTestHooks *game2QuickbarHooks

func g2qSound(id, vol int32) {
	if h := game2QuickbarTestHooks; h != nil && h.Sound != nil {
		h.Sound(id, vol)
		return
	}
	C.nox_xxx_clientPlaySoundSpecial_452D80(C.int(id), C.int(vol))
}
func g2qDestroy(w *C.nox_window) int32 {
	if h := game2QuickbarTestHooks; h != nil && h.Destroy != nil {
		return h.Destroy(unsafe.Pointer(w))
	}
	return int32(C.nox_xxx_windowDestroyMB_46C4E0(w))
}
func g2qHidden(w *C.nox_window, hidden int32) int32 {
	if h := game2QuickbarTestHooks; h != nil && h.Hidden != nil {
		return h.Hidden(unsafe.Pointer(w), hidden)
	}
	return int32(C.nox_window_set_hidden(w, C.int(hidden)))
}
func g2qStyle(w *C.nox_window, on int32) int32 {
	if h := game2QuickbarTestHooks; h != nil && h.Style != nil {
		return h.Style(unsafe.Pointer(w), on)
	}
	return int32(C.sub_46AE10(w, C.int(on)))
}
func g2qFlags(spell uint32, mask uint32) int32 {
	if h := game2QuickbarTestHooks; h != nil && h.SpellFlags != nil {
		return h.SpellFlags(spell, mask)
	}
	if bool(C.nox_xxx_spellHasFlags_424A50(C.int(spell), C.int(mask))) {
		return 1
	}
	return 0
}
func g2qTooltip(w *C.nox_window, spell uint32) {
	if h := game2QuickbarTestHooks; h != nil && h.Tooltip != nil {
		h.Tooltip(unsafe.Pointer(w), spell)
		return
	}
	C.nox_xxx_wndWddSetTooltip_46B000(&w.draw_data, (*C.wchar2_t)(unsafe.Pointer(C.nox_xxx_spellTitle_424930(C.int(spell)))))
}
func g2qB(p unsafe.Pointer, o uintptr) *byte   { return (*byte)(unsafe.Add(p, o)) }
func g2qW(p unsafe.Pointer, o uintptr) *uint32 { return (*uint32)(unsafe.Add(p, o)) }
func g2qRow(p unsafe.Pointer) unsafe.Pointer   { return unsafe.Pointer(C.nox_quickbar_selected_row(p)) }
func g2qMain() unsafe.Pointer                  { return C.nox_xxx_aClosewoodengat_587000_133480 }
func g2qTrap() unsafe.Pointer                  { return memmap.PtrOff(0x5D4594, 1047940) }
func g2qExpanded(i int) unsafe.Pointer         { return memmap.PtrOff(0x5D4594, 1048196+uintptr(256*i)) }
func g2qPlayer() *C.nox_playerInfo {
	return *(**C.nox_playerInfo)(unsafe.Pointer(&C.dword_8531A0_2576))
}
func g2qMirror(p unsafe.Pointer) {
	if unsafe.Sizeof(uintptr(0)) == 4 && p != nil {
		*g2qW(p, 204) = uint32(uintptr(g2qRow(p)))
	}
}
func g2qTrapMirror() {
	if unsafe.Sizeof(uintptr(0)) == 4 {
		*memmap.PtrUint32(0x5D4594, 1048144) = uint32(uintptr(unsafe.Pointer(C.nox_quickbar_trap_selected_row())))
	}
}

// Keep the original eight native C-owned sets. No packed record stores a
// native pointer on a 64-bit host. The original nil-key matching is retained.
func g2qWindows(data unsafe.Pointer, create bool) *C.nox_quickbar_window_set {
	var empty *C.nox_quickbar_window_set
	for i := range C.nox_quickbar_window_sets {
		p := &C.nox_quickbar_window_sets[i]
		if p.data == data {
			return p
		}
		if p.data == nil && empty == nil {
			empty = p
		}
	}
	if create && empty != nil {
		empty.data = data
		return empty
	}
	return nil
}

//export nox_quickbar_root
func nox_quickbar_root(data unsafe.Pointer) *C.nox_window {
	if p := g2qWindows(data, false); p != nil {
		return p.root
	}
	return nil
}

//export nox_quickbar_button
func nox_quickbar_button(data unsafe.Pointer, index C.int) *C.nox_window {
	if p := g2qWindows(data, false); p != nil && index >= 0 && index < 5 {
		return p.buttons[index]
	}
	return nil
}

//export nox_quickbar_nugget
func nox_quickbar_nugget(data unsafe.Pointer, index C.int) *C.nox_window {
	if p := g2qWindows(data, false); p != nil && index >= 0 && index < 5 {
		return p.nuggets[index]
	}
	return nil
}

//export nox_quickbar_nugget_child
func nox_quickbar_nugget_child(data unsafe.Pointer, index C.int) *C.nox_window {
	if p := g2qWindows(data, false); p != nil && index >= 0 && index < 5 {
		return p.nugget_children[index]
	}
	return nil
}

//export nox_quickbar_set_nugget
func nox_quickbar_set_nugget(data unsafe.Pointer, index C.int, win *C.nox_window) {
	if p := g2qWindows(data, true); p != nil && index >= 0 && index < 5 {
		p.nuggets[index] = win
	}
}

//export nox_quickbar_set_nugget_child
func nox_quickbar_set_nugget_child(data unsafe.Pointer, index C.int, win *C.nox_window) {
	if p := g2qWindows(data, true); p != nil && index >= 0 && index < 5 {
		p.nugget_children[index] = win
	}
}

//export nox_quickbar_data_for_window
func nox_quickbar_data_for_window(win *C.nox_window) unsafe.Pointer {
	for i := range C.nox_quickbar_window_sets {
		p := &C.nox_quickbar_window_sets[i]
		if p.root == win {
			return p.data
		}
		for j := range p.buttons {
			if p.buttons[j] == win || p.nuggets[j] == win || p.nugget_children[j] == win {
				return p.data
			}
		}
	}
	return nil
}

//export nox_quickbar_windows_clear
func nox_quickbar_windows_clear(data unsafe.Pointer) {
	if p := g2qWindows(data, false); p != nil {
		*p = C.nox_quickbar_window_set{}
	}
}

//export nox_quickbar_windows_destroy_all
func nox_quickbar_windows_destroy_all() {
	for i := range C.nox_quickbar_window_sets {
		p := &C.nox_quickbar_window_sets[i]
		g2qDestroy(p.root)
		for j := range p.buttons {
			g2qDestroy(p.buttons[j])
			g2qDestroy(p.nuggets[j])
			g2qDestroy(p.nugget_children[j])
		}
		*p = C.nox_quickbar_window_set{}
	}
}

//export nox_game2_book_insert_native
func nox_game2_book_insert_native(kind C.uintptr_t, spell, slot C.int) C.uintptr_t {
	nox_xxx_spellKeyPackSetSpell_45DC40(g2qMain(), spell, slot)
	result := kind
	if kind == 2 {
		row := g2qRow(g2qMain())
		if row == nil || slot < 0 || slot >= 5 {
			return result
		}
		if spell != 0 {
			result = C.uintptr_t(g2qFlags(uint32(spell), 0x600))
		}
		if spell != 0 && result != 0 {
			*g2qB(row, uintptr(8*slot+4)) = 1
		} else {
			result = C.uintptr_t(uintptr(g2qMain()))
			*g2qB(row, uintptr(8*slot+4)) = 0
		}
	}
	return result
}

//export nox_xxx_spellKeyPackSetSpell_45DC40
func nox_xxx_spellKeyPackSetSpell_45DC40(data unsafe.Pointer, spell, slot C.int) {
	row := g2qRow(data)
	if row == nil || slot < 0 || slot >= 5 {
		return
	}
	*g2qW(row, uintptr(8*slot)) = uint32(spell)
	if w := nox_quickbar_button(data, slot); w != nil {
		g2qTooltip(w, uint32(spell))
	}
	if C.dword_8531A0_2576 != 0 {
		g2qSound(794, 100)
	}
}

//export nox_xxx_bookSpellDrop_45DCA0
func nox_xxx_bookSpellDrop_45DCA0(spell C.int, flags C.char, x, y C.int, source unsafe.Pointer) C.int {
	oldSpell := uint16(137)
	oldFlags := byte(0)
	trap := g2qTrap()
	root := nox_quickbar_root(trap)
	shown := root != nil && C.wndIsShown_nox_xxx_wndIsShown_46ACC0(root) != 0
	if (shown || C.nox_xxx_spellPutInBox_45DEB0(trap, spell, x, y) == 0) && source != trap {
		first := 4
		if memmap.Uint32(0x5D4594, 1049476) != 0 {
			first = 0
		}
		for i := first; i < 5; i++ {
			data := g2qExpanded(i)
			slot := nox_xxx_spellBoxPointToWnd_45DE60(data, x, y)
			row := g2qRow(data)
			if slot >= 0 && row != nil {
				oldSpell = *(*uint16)(unsafe.Add(row, uintptr(8*slot)))
				oldFlags = *g2qB(row, uintptr(8*slot+4))
			}
			if C.nox_xxx_spellPutInBox_45DEB0(data, spell, x, y) == 0 {
				continue
			}
			if slot >= 0 && row != nil {
				v := byte(flags)
				if C.nox_xxx_bookGetSpellDnDType_477670() == 1 {
					v = byte(g2qFlags(uint32(spell), 0x600))
				}
				*g2qB(row, uintptr(8*slot+4)) = v
			}
			nox_xxx_updateSpellIcons_45DDF0(data)
			break
		}
	}
	return C.int(uint32(oldFlags) | (uint32(oldSpell) << 16))
}

//export nox_xxx_updateSpellIcons_45DDF0
func nox_xxx_updateSpellIcons_45DDF0(data unsafe.Pointer) C.int {
	var result C.int
	for i := C.int(0); i < 5; i++ {
		result = nox_xxx_updateSpellIconDir_45DE10(i, data)
	}
	return result
}

//export nox_xxx_updateSpellIconDir_45DE10
func nox_xxx_updateSpellIconDir_45DE10(index C.int, data unsafe.Pointer) C.int {
	row := g2qRow(data)
	if row == nil || index < 0 || index >= 5 {
		return 0
	}
	spell := *g2qW(row, uintptr(8*index))
	flags := *g2qB(row, uintptr(8*index+4))
	on := int32(0)
	if flags&1 != 0 && spell != 0 {
		on = 1
	}
	return C.int(g2qStyle(nox_quickbar_nugget(data, index), on))
}

//export nox_xxx_spellBoxPointToWnd_45DE60
func nox_xxx_spellBoxPointToWnd_45DE60(data unsafe.Pointer, x, y C.int) C.int {
	for i := C.int(0); i < 5; i++ {
		w := nox_quickbar_button(data, i)
		if w != nil && g2qPointIn(w, x, y) {
			return i
		}
	}
	return -1
}

//export nox_xxx_guiSpellSetCursor_45DF60
func nox_xxx_guiSpellSetCursor_45DF60(spell C.int, flags C.char) C.int {
	if spell != 0 {
		g2qSound(766, 100)
		if byte(flags)&1 == 0 || g2qFlags(uint32(spell), 0x2000) != 0 {
			*memmap.PtrUint32(0x5D4594, 1047556) = uint32(spell)
			*memmap.PtrUint32(0x5D4594, 1047928) = 1
			return 0
		}
		C.nox_xxx_guiSpell_45DA10(spell)
		return 1
	}
	*memmap.PtrUint32(0x5D4594, 1047928) = 0
	return 0
}

//export sub_45DFC0
func sub_45DFC0(ability C.int) C.int {
	if ability == 0 {
		C.dword_5d4594_1047932 = 0
		C.dword_5d4594_1047936 = 0
		return 0
	}
	word := memmap.Uint32(0x5D4594, 1047764+uintptr(24*ability)+4)
	if word != 0 {
		C.dword_5d4594_1047932 = 0
		C.dword_5d4594_1047936 = 0
		C.nox_xxx_clientSendAbil_45DAF0(ability)
		g2qSound(766, 100)
		return C.int(word)
	}
	C.dword_5d4594_1047932 = 1
	C.dword_5d4594_1047936 = C.uint32_t(memmap.Uint32(0x5D4594, 1047764+uintptr(24*ability)))
	g2qSound(766, 100)
	return 0
}

//export nox_xxx_clientUpdateButtonRow_45E110
func nox_xxx_clientUpdateButtonRow_45E110(index C.int) C.int {
	if index >= 0 && index < 5 && memmap.Uint32(0x5D4594, 1049476) == 0 && C.dword_5d4594_1049496 == 0 {
		data := g2qMain()
		if data == nil {
			return index
		}
		*g2qB(data, 200) = byte(index)
		if C.dword_8531A0_2576 != 0 {
			g2qSound(798, 100)
		}
		g2qMirror(data)
		return nox_xxx_updateSpellIcons_45DDF0(data)
	}
	return index
}

//export nox_xxx_buttonsGetSelectedRow_45E180
func nox_xxx_buttonsGetSelectedRow_45E180() C.int {
	if p := g2qMain(); p != nil {
		return C.int(*g2qB(p, 200))
	}
	return 0
}

//export nox_xxx_quickbarButtonBookDraw_45EF30
func nox_xxx_quickbarButtonBookDraw_45EF30() C.int { return 1 }

//export sub_45EF40
func sub_45EF40() C.int { return 0 }

//export nox_xxx_quickBarWnd_45EF50
func nox_xxx_quickBarWnd_45EF50(win *C.nox_window, event C.int, pos C.uint) C.int {
	data := nox_quickbar_data_for_window(win)
	player := g2qPlayer()
	if data == nil {
		return 0
	}
	if C.nox_client_localPlayerHasFlags30(2) != 0 {
		return 1
	}
	slot := C.int(-1)
	for i := C.int(0); i < 5; i++ {
		if nox_quickbar_button(data, i) == win {
			slot = i
			break
		}
	}
	if slot < 0 {
		return 0
	}
	row := g2qRow(data)
	if row == nil {
		return 0
	}
	spell := g2qW(row, uintptr(8*slot))
	x, y := C.int(uint16(pos)), C.int(uint32(pos)>>16)
	switch event {
	case 5:
		if *spell == 0 || memmap.Uint32(0x5D4594, 1047928) != 0 || C.dword_5d4594_1047932 != 0 || g2qBlocked() != 0 || (player != nil && game2PlayerClass(player) != 0 && !g2qEnabled(*spell)) {
			return 1
		}
		g2qCaptureSet(win)
		g2qDragSave(*spell, 2)
		C.dword_5d4594_1049692 = data
		C.dword_5d4594_1049696 = row
		g2qSound(793, 100)
		return 1
	case 6, 7:
		if g2qCapture() == nil {
			return 1
		}
		target := nox_xxx_spellBoxPointToWnd_45DE60(data, x, y)
		if g2qBlocked() == 0 && memmap.Uint32(0x5D4594, 1047928) == 0 && C.dword_5d4594_1047932 == 0 {
			if target == slot {
				C.dword_5d4594_1049532 = win
				if data != g2qTrap() {
					if player != nil && game2PlayerClass(player) != 0 {
						if g2qEnabled(*spell) {
							g2qCursor(*spell, *g2qB(row, uintptr(8*slot+4)))
							g2qStore(slot)
						}
					} else {
						g2qAbility(*spell)
						g2qStore(slot)
					}
				}
			} else if target >= 0 {
				g2qEventSwap(data, slot, target)
				g2qSound(794, 100)
				g2qEventUpdate(data)
			} else {
				dropped := g2qEventDrop(*spell, *g2qB(row, uintptr(8*slot+4)), x, y, data)
				replacement := dropped >> 16
				if replacement == 137 {
					if !g2qPointRect(x, y) && (C.dword_5d4594_1049696 == nil || C.dword_5d4594_1049696 == row) {
						replacement = 0
						dropped = 0
					}
				}
				if replacement != 137 {
					*spell = replacement
					*g2qB(row, uintptr(8*slot+4)) = byte(dropped)
					if data != g2qTrap() {
						g2qEventUpdate(data)
					}
				}
			}
		} else if C.dword_5d4594_1047932 != 0 && target < 0 {
			g2qCancelAbility()
			C.dword_5d4594_1047932 = 0
		} else if memmap.Uint32(0x5D4594, 1047928) != 0 {
			if target < 0 {
				g2qTarget(memmap.Uint32(0x5D4594, 1047556))
			}
			*memmap.PtrUint32(0x5D4594, 1047928) = 0
		}
		if memmap.Uint32(0x5D4594, 1047928) == 0 && C.dword_5d4594_1047932 == 0 {
			g2qCaptureClear(win)
			C.dword_5d4594_1049532 = nil
		}
		g2qDragClear()
		return 1
	case 8, 12, 16:
		return 0
	default:
		return 1
	}
}

//export nox_xxx_clientSwapQuickbarKeys_45F300
func nox_xxx_clientSwapQuickbarKeys_45F300(data unsafe.Pointer, a, b C.int) {
	row := g2qRow(data)
	if row == nil || a < 0 || a >= 5 || b < 0 || b >= 5 {
		return
	}
	p, q := unsafe.Add(row, uintptr(8*a)), unsafe.Add(row, uintptr(8*b))
	tmp := *(*[8]byte)(p)
	*(*[8]byte)(p) = *(*[8]byte)(q)
	*(*[8]byte)(q) = tmp
	if w := nox_quickbar_button(data, a); w != nil {
		g2qTooltip(w, *g2qW(p, 0))
	}
	if w := nox_quickbar_button(data, b); w != nil {
		g2qTooltip(w, *g2qW(q, 0))
	}
}

//export sub_45F390
func sub_45F390(a, b unsafe.Pointer, sa, sb C.int) {
	p, q := unsafe.Add(a, uintptr(8*sa)), unsafe.Add(b, uintptr(8*sb))
	tmp := *(*[8]byte)(q)
	*(*[8]byte)(q) = *(*[8]byte)(p)
	*(*[8]byte)(p) = tmp
	if w := nox_quickbar_button(g2qMain(), sb); w != nil {
		g2qTooltip(w, *g2qW(q, 0))
	}
}

//export nox_xxx_quickbarButtonBookWnd_45F450
func nox_xxx_quickbarButtonBookWnd_45F450(unused1 *C.nox_window, event C.uint) C.int {
	if event >= 5 {
		if event <= 6 {
			return 1
		}
		if event == 7 {
			C.nox_client_toggleSpellbook_45AC70()
			return 1
		}
	}
	return 0
}

//export sub_45F500
func sub_45F500(slot C.int, data unsafe.Pointer) C.int {
	if w := nox_quickbar_nugget(data, slot); w != nil {
		return C.int((w.draw_data.field_0 >> 1) & 1)
	}
	return 0
}

//export sub_45F520
func sub_45F520(win *C.nox_window, event C.uint) C.int {
	if win == nil {
		return 0
	}
	ri, slot := uint32(win.field_92)>>16, uint32(win.field_92)&0xffff
	if ri >= 5 || slot >= 5 {
		return 0
	}
	data := g2qExpanded(int(ri))
	row := g2qRow(data)
	if row == nil {
		return 0
	}
	entry := unsafe.Add(row, uintptr(8*slot))
	spell := *g2qW(entry, 0)
	if spell == 0 {
		return 0
	}
	if event != 5 {
		if event > 5 && event <= 7 {
			return 1
		}
		return 0
	}
	if g2qFlags(spell, 0x200400) != 0 {
		g2qSound(925, 100)
	} else {
		*g2qB(entry, 4) ^= 1
		nox_xxx_updateSpellIconDir_45DE10(C.int(slot), data)
		g2qSound(921, 100)
	}
	return 1
}

//export sub_45F5D0
func sub_45F5D0(win *C.nox_window, unused1 *C.nox_window_data) C.int {
	if win == nil {
		return 0
	}
	var x, y C.uint
	C.nox_client_wndGetPosition_46AA60(win, &x, &y)
	if text := C.sub_42E8E0(C.int(win.field_92+28), 1); text != nil {
		C.nox_xxx_drawSetTextColor_434390(C.int(C.nox_color_white_2523948))
		C.nox_xxx_drawString_43F6E0(C.nox_quickbar_font, (*C.wchar2_t)(unsafe.Pointer(text)), C.int(x), C.int(y))
	}
	return 1
}

//export nox_xxx_quickbarTrapUpDownProc_45F630
func nox_xxx_quickbarTrapUpDownProc_45F630(win *C.nox_window, event C.uint) C.int {
	if win == nil {
		return 0
	}
	control := uint32(win.field_92)
	if C.nox_client_localPlayerHasFlags30(2) == 0 && C.sub_4AE3D0() == 0 {
		if event == 5 {
			if control <= 2 && C.dword_5d4594_1049508 != nil {
				C.dword_5d4594_1049508.draw_data.bg_image = nil
			}
			switch control {
			case 0:
				nox_client_spellSetPrev_460540()
			case 1:
				nox_client_spellSetNext_4604F0()
			case 2:
				sub_460920()
			case 3:
				nox_client_trapSetPrev_4603F0()
			case 4:
				nox_client_trapSetNext_4603A0()
			}
			*memmap.PtrUint32(0x5D4594, 1049700) = 2
			*memmap.PtrUint32(0x5D4594, 1049704) = control
		} else if event <= 5 || event > 7 {
			return 0
		}
	}
	return 1
}

//export nox_xxx_quickbarTrapUpDownDraw_45F6F0
func nox_xxx_quickbarTrapUpDownDraw_45F6F0(win *C.nox_window, data *C.nox_window_data) C.int {
	if win == nil {
		return 0
	}
	if data == nil {
		data = &win.draw_data
	}
	var x, y C.uint
	C.nox_client_wndGetPosition_46AA60(win, &x, &y)
	dx, dy := C.int(int32(x)+int32(data.img_px)), C.int(int32(y)+int32(data.img_py))
	if memmap.Int32(0x5D4594, 1049700) <= 0 || memmap.Int32(0x5D4594, 1049704) != int32(win.field_92) {
		if data.bg_image != nil {
			C.nox_client_drawImageAt_47D2C0((*C.nox_video_bag_image_t)(data.bg_image), dx, dy)
		}
	} else {
		if data.hl_image != nil {
			C.nox_client_drawImageAt_47D2C0((*C.nox_video_bag_image_t)(data.hl_image), dx, dy)
		}
		p := memmap.PtrUint32(0x5D4594, 1049700)
		*p--
		if *p == 0 && memmap.Uint32(0x5D4594, 1049704) <= 2 && C.dword_5d4594_1049508 != nil {
			C.dword_5d4594_1049508.draw_data.bg_image = C.dword_5d4594_1049508.draw_data.hl_image
		}
	}
	g2qStyle(C.dword_5d4594_1049524, int32(C.sub_45CFC0()))
	return 1
}

//export nox_xxx_quickbarTrapButtonProc_45F7A0
func nox_xxx_quickbarTrapButtonProc_45F7A0(win *C.nox_window, event, pos C.uint) C.int {
	if win == nil {
		return 0
	}
	if event < 5 {
		g2qStyle(C.dword_5d4594_1049504, 0)
		return 0
	}
	if event <= 6 {
		return 1
	}
	if event != 7 {
		g2qStyle(C.dword_5d4594_1049504, 0)
		return 0
	}
	if memmap.Uint32(0x5D4594, 1047928) != 0 {
		if !bool(C.nox_xxx_wndPointInWnd_46AAB0(win, C.int(uint16(pos)), C.int(uint32(pos)>>16))) {
			C.nox_client_buildTrap_45E040()
		}
		C.nox_xxx_wndClearCaptureMain_46ADE0(win)
		C.dword_5d4594_1049532 = nil
		*memmap.PtrUint32(0x5D4594, 1047928) = 0
		return 1
	}
	g2qStyle(C.dword_5d4594_1049504, 1)
	player := g2qPlayer()
	if player == nil {
		return 1
	}
	if game2PlayerClass(player) == 1 && sub_45F890() != 0 {
		C.nox_xxx_wndSetCaptureMain_46ADC0(win)
		C.dword_5d4594_1049532 = win
		*memmap.PtrUint32(0x5D4594, 1047928) = 1
	} else {
		C.nox_client_buildTrap_45E040()
	}
	return 1
}

//export sub_45F890
func sub_45F890() C.int {
	row := unsafe.Pointer(C.nox_quickbar_trap_selected_row())
	if row == nil {
		return 0
	}
	for i := 0; i < 3; i++ {
		spell := *g2qW(row, uintptr(i*8))
		if spell != 0 && uint32(C.nox_xxx_spellFlags_424A70(C.int(spell)))&8 == 8 {
			return 1
		}
	}
	return 0
}

//export nox_xxx_quickbar_45F8D0
func nox_xxx_quickbar_45F8D0(unused1 *C.nox_window, event C.int, unused2 C.uintptr_t, unused3 C.uintptr_t) C.int {
	if event != 8 && event != 12 && event != 16 {
		return 1
	}
	return 0
}

//export sub_45F8F0
func sub_45F8F0() C.int { sub_45F900(); return 1 }

//export sub_45F900
func sub_45F900() {
	data := g2qMain()
	if data == nil {
		return
	}
	sequence := uint32(C.nox_xxx_bookGet_430B40_get_mouse_prev_seq())
	old := memmap.Uint32(0x5D4594, 1049492)
	if C.dword_5d4594_1049496 != 0 {
		if old == sequence {
			return
		}
		C.dword_5d4594_1049496 = 0
	} else {
		if old != sequence {
			return
		}
		C.dword_5d4594_1049496 = 1
	}
	row := g2qRow(data)
	if row == nil {
		return
	}
	for i := 0; i < 5; i++ {
		spell := *g2qW(row, uintptr(8*i))
		if spell != 0 && g2qFlags(spell, 0x200400) == 0 {
			*g2qB(row, uintptr(8*i+4)) ^= 1
			nox_xxx_updateSpellIconDir_45DE10(C.int(i), data)
		}
	}
	nox_xxx_updateSpellIcons_45DDF0(data)
	g2qSound(921, 100)
}

//export nox_xxx_quickbarTrapProc_45FB90
func nox_xxx_quickbarTrapProc_45FB90(win *C.nox_window, event C.uint, unused1 C.uintptr_t, unused2 C.uintptr_t) C.int {
	if event >= 5 {
		if event <= 6 {
			return 1
		}
		if event == 7 {
			if win != nil && C.wndIsShown_nox_xxx_wndIsShown_46ACC0(win) == 0 {
				C.sub_461060()
			}
			return 1
		}
	}
	return 0
}

//export nox_xxx_quickbarDrawFn_460000
func nox_xxx_quickbarDrawFn_460000(unused1 *C.nox_window, unused2 *C.nox_window_data) C.int {
	win := C.dword_5d4594_1049504
	if win == nil {
		return 0
	}
	y := int32(win.off_y)
	delta := y - int32(C.dword_5d4594_1049536)
	if delta >= 0 {
		if delta <= 0 {
			if uint32(C.dword_5d4594_1049536) > uint32(C.nox_win_height) {
				C.sub_460070()
				C.dword_5d4594_1049536 = C.uint32_t(C.nox_win_height - 74)
			}
		} else {
			C.nox_window_setPos_46A9B0(win, C.int(win.off_x), C.int(y-1))
		}
	} else {
		C.nox_window_setPos_46A9B0(win, C.int(win.off_x), C.int(y+1))
	}
	return 1
}

//export nox_xxx_quickBarInitWindow_4601F0
func nox_xxx_quickBarInitWindow_4601F0(data unsafe.Pointer, x, y, count, spacing C.int, proc, draw unsafe.Pointer) C.int {
	set := g2qWindows(data, true)
	if data == nil || set == nil || count < 0 || count > 5 {
		return 0
	}
	set.root = C.nox_window_new(nil, 1224, x-10, y-15, 199, 59, nil)
	if set.root == nil {
		nox_quickbar_windows_clear(data)
		return 0
	}
	C.nox_xxx_wndSetIcon_46AE60(set.root, C.nox_xxx_gLoadImg_42F970(internCStr("QuickBarBase")))
	C.nox_xxx_wndSetOffsetMB_46AE40(set.root, -61, -18)
	C.nox_xxx_wndSetWindowProc_46B300(set.root, unsafe.Pointer(C.nox_xxx_quickbar_45F8D0))
	bx := spacing + 10
	for i := C.int(0); i < count; i++ {
		set.buttons[i] = C.nox_window_new(set.root, 1032, bx, 15, 34, 34, nil)
		if set.buttons[i] == nil {
			return 0
		}
		C.nox_window_set_all_funcs(set.buttons[i], proc, draw, nil)
		if unsafe.Sizeof(uintptr(0)) == 4 {
			*(*unsafe.Pointer)(unsafe.Add(data, uintptr(212+4*i))) = unsafe.Pointer(set.buttons[i])
			set.buttons[i].field_92 = C.uint(uintptr(data))
		}
		bx += C.int(memmap.Uint32(0x587000, 133488+uintptr(4*i)))
	}
	if unsafe.Sizeof(uintptr(0)) == 4 {
		*(*unsafe.Pointer)(unsafe.Add(data, 208)) = unsafe.Pointer(set.root)
		g2qMirror(data)
	}
	return C.int(*g2qB(data, 200))
}

//export sub_4602F0
func sub_4602F0() unsafe.Pointer {
	nox_xxx_quickBarClose_4606B0()
	main := g2qMain()
	if main == nil {
		return nil
	}
	for row := 0; row < 5; row++ {
		expanded := g2qExpanded(row)
		for slot := 0; slot < 5; slot++ {
			dst := unsafe.Add(main, uintptr(40*row+8*slot))
			src := unsafe.Add(expanded, uintptr(8*slot))
			*g2qW(dst, 0) = 0
			*g2qB(dst, 4) = 0
			*g2qW(src, 0) = 0
			*g2qB(src, 4) = 0
		}
	}
	for row := 0; row < 3; row++ {
		for slot := 0; slot < 3; slot++ {
			p := unsafe.Add(g2qTrap(), uintptr(40*row+8*slot))
			*g2qW(p, 0) = 0
			*g2qB(p, 4) = 0
		}
	}
	nox_xxx_updateSpellIcons_45DDF0(main)
	return main
}

//export sub_460380
func sub_460380() unsafe.Pointer {
	for offset := 1047764 + 24 + 16; offset < 1047924; offset += 24 {
		*memmap.PtrUint32(0x5D4594, uintptr(offset)) = 0
	}
	return sub_4602F0()
}

//export nox_client_trapSetNext_4603A0
func nox_client_trapSetNext_4603A0() {
	p := memmap.PtrUint8(0x5D4594, 1048140)
	if *p == 2 {
		*p = 0
	} else {
		*p++
	}
	g2qTrapMirror()
	g2qSound(798, 100)
}

//export nox_client_trapSetPrev_4603F0
func nox_client_trapSetPrev_4603F0() {
	p := memmap.PtrUint8(0x5D4594, 1048140)
	if *p != 0 {
		*p--
	} else {
		*p = 2
	}
	g2qTrapMirror()
	g2qSound(798, 100)
}

//export nox_client_trapSetSelect_4604B0
func nox_client_trapSetSelect_4604B0(index C.int) C.int {
	if index >= 0 && index < 3 {
		*memmap.PtrUint8(0x5D4594, 1048140) = byte(index)
		g2qTrapMirror()
		return C.int(5 * byte(index))
	}
	return index
}

//export sub_4604E0
func sub_4604E0() C.int { return C.int(memmap.Uint8(0x5D4594, 1048140)) }

//export nox_client_spellSetNext_4604F0
func nox_client_spellSetNext_4604F0() C.int {
	result := C.int(memmap.Uint32(0x5D4594, 1049476))
	if result == 0 {
		result = C.int(C.dword_5d4594_1049496)
		if result == 0 {
			result = C.nox_xxx_get_57AF20()
			if result == 0 {
				C.nox_xxx_clientStoreLastButton_45DAD0(-1)
				row := nox_xxx_buttonsGetSelectedRow_45E180() + 1
				if row > 4 {
					row = 0
				}
				result = nox_xxx_clientUpdateButtonRow_45E110(row)
			}
		}
	}
	return result
}

//export nox_client_spellSetPrev_460540
func nox_client_spellSetPrev_460540() C.int {
	result := C.int(memmap.Uint32(0x5D4594, 1049476))
	if result == 0 {
		result = C.int(C.dword_5d4594_1049496)
		if result == 0 {
			result = C.nox_xxx_get_57AF20()
			if result == 0 {
				C.nox_xxx_clientStoreLastButton_45DAD0(-1)
				row := nox_xxx_buttonsGetSelectedRow_45E180() - 1
				if row < 0 {
					row = 4
				}
				result = nox_xxx_clientUpdateButtonRow_45E110(row)
			}
		}
	}
	return result
}

//export nox_client_spellSetSelect_460590
func nox_client_spellSetSelect_460590() {
	if memmap.Uint32(0x5D4594, 1049476) != 0 || C.dword_5d4594_1049496 != 0 {
		return
	}
	data := g2qMain()
	if data == nil {
		return
	}
	var row uint32
	if bool(C.nox_xxx_checkKeybTimeout_4160F0(7, C.gameFPS()>>1)) {
		*memmap.PtrUint32(0x5D4594, 1049712) = 0
	} else {
		p := memmap.PtrUint32(0x5D4594, 1049712)
		*p++
		row = *p
		if int32(*p) >= 5 {
			row = 4
		}
	}
	*g2qB(data, 200) = byte(row)
	g2qMirror(data)
	C.nox_xxx_setKeybTimeout_4160D0(7)
	g2qSound(798, 100)
	nox_xxx_updateSpellIcons_45DDF0(data)
}

//export sub_460630
func sub_460630() {
	if memmap.Uint32(0x5D4594, 1049476) == 0 {
		*memmap.PtrUint32(0x5D4594, 1049492) = uint32(C.nox_xxx_bookGet_430B40_get_mouse_prev_seq())
	}
}

//export nox_xxx_guiSpell_460650
func nox_xxx_guiSpell_460650() C.int { return C.int(memmap.Uint32(0x5D4594, 1047928)) }

//export sub_460660
func sub_460660() C.int {
	if C.dword_5d4594_1049532 == nil && memmap.Uint32(0x5D4594, 1047928) == 0 && C.dword_5d4594_1047932 == 0 {
		return 0
	}
	capture := C.dword_5d4594_1049532
	if capture == nil {
		capture = C.nox_xxx_wndGetCaptureMain_46AE00()
	}
	if capture != nil {
		C.nox_xxx_wndClearCaptureMain_46ADE0(capture)
	}
	C.dword_5d4594_1049532 = nil
	*memmap.PtrUint32(0x5D4594, 1047928) = 0
	C.dword_5d4594_1047932 = 0
	return 1
}

//export nox_xxx_quickBarClose_4606B0
func nox_xxx_quickBarClose_4606B0() C.int {
	if memmap.Uint32(0x5D4594, 1049476) != 1 {
		return 0
	}
	main := g2qMain()
	if main == nil {
		return 0
	}
	for row := 0; row < 5; row++ {
		expanded := g2qExpanded(row)
		for slot := 0; slot < 5; slot++ {
			src := unsafe.Add(expanded, uintptr(slot*8))
			dst := unsafe.Add(main, uintptr(row*40+slot*8))
			*g2qW(dst, 0) = *g2qW(src, 0)
			*g2qB(dst, 4) = *g2qB(src, 4)
			g2qHidden(nox_quickbar_nugget(expanded, C.int(slot)), 1)
		}
		g2qHidden(nox_quickbar_root(expanded), 1)
	}
	selected := memmap.Uint32(0x5D4594, 1047912)
	if selected >= 5 {
		selected = 0
	}
	*g2qB(main, 200) = byte(selected)
	g2qMirror(main)
	nox_xxx_updateSpellIcons_45DDF0(main)
	g2qHidden(C.dword_5d4594_1049512, 0)
	g2qSound(800, 100)
	*memmap.PtrUint32(0x5D4594, 1049476) = 0
	return 1
}

//export sub_460920
func sub_460920() {
	if memmap.Uint32(0x5D4594, 1049476) != 0 {
		nox_xxx_quickBarClose_4606B0()
		return
	}
	if C.dword_5d4594_1049496 != 0 || C.nox_xxx_get_57AF20() != 0 {
		return
	}
	if C.dword_5d4594_1049484 == 1 {
		C.sub_461010()
	}
	main := g2qMain()
	if main == nil {
		return
	}
	for row := 0; row < 5; row++ {
		expanded := g2qExpanded(row)
		for slot := 0; slot < 5; slot++ {
			src := unsafe.Add(main, uintptr(row*40+slot*8))
			dst := unsafe.Add(expanded, uintptr(slot*8))
			*g2qW(dst, 0) = *g2qW(src, 0)
			*g2qB(dst, 4) = *g2qB(src, 4)
			g2qHidden(nox_quickbar_nugget(expanded, C.int(slot)), 0)
		}
		g2qHidden(nox_quickbar_root(expanded), 0)
		nox_xxx_updateSpellIcons_45DDF0(expanded)
	}
	*memmap.PtrUint32(0x5D4594, 1047912) = uint32(*g2qB(main, 200))
	*g2qB(main, 200) = 4
	g2qMirror(main)
	nox_xxx_updateSpellIcons_45DDF0(main)
	g2qHidden(C.dword_5d4594_1049512, 1)
	g2qSound(799, 100)
	*memmap.PtrUint32(0x5D4594, 1049476) = 1
}

//export sub_460940
func sub_460940(unused1 unsafe.Pointer) C.int {
	if C.nox_crypt_IsReadOnly() == 0 && C.sub_461450() == 1 {
		C.sub_461400()
		C.sub_461440(0)
	}
	var kind C.uint8_t
	if p := g2qPlayer(); p != nil {
		kind = C.uint8_t(game2PlayerClass(p))
	} else {
		kind = *(*C.uint8_t)(unsafe.Add(unsafe.Pointer(C.nox_xxx_getHostInfoPtr_431770()), 66))
	}
	g2qCrypt((*C.uint8_t)(unsafe.Pointer(&kind)), 1)
	if kind == 0 {
		if sub_460A10(g2qMain(), 5, 5, 0) == 0 {
			return 0
		}
		return 1
	}
	result := sub_460A10(g2qMain(), 5, 5, C.char(kind))
	if result != 0 {
		result = sub_460A10(g2qTrap(), 3, 3, C.char(kind))
		if result != 0 {
			if nox_quickbar_nugget(g2qMain(), 0) != nil {
				nox_xxx_updateSpellIcons_45DDF0(g2qMain())
			}
			return 1
		}
	}
	return result
}

//export sub_460A10
func sub_460A10(data unsafe.Pointer, rows, slots C.int, kind C.char) C.int {
	if rows <= 0 {
		return 1
	}
	// The C byte/buffer persist across records, including a later short read.
	// Their initial zero value defines only the old uninitialized-input edge.
	var length C.uint8_t
	var name [256]C.char
	for row := C.int(0); row < rows; row++ {
		for slot := C.int(0); slot < slots; slot++ {
			p := unsafe.Add(data, uintptr(row*40+slot*8))
			if C.nox_crypt_IsReadOnly() == 1 {
				g2qCrypt(&length, 1)
				g2qCrypt((*C.uint8_t)(unsafe.Pointer(&name[0])), C.size_t(length))
				name[int(length)] = 0
				if kind != 0 {
					*g2qW(p, 0) = uint32(C.nox_xxx_spellNameToN_4243F0(&name[0]))
				} else {
					*g2qW(p, 0) = uint32(C.nox_xxx_abilityNameToN_424D80(&name[0]))
				}
			} else {
				var name *C.char
				if kind != 0 {
					name = C.nox_xxx_spellNameByN_424870(C.int(*g2qW(p, 0)))
				} else {
					name = C.nox_xxx_abilityGetName_425250(C.int(*g2qW(p, 0)))
				}
				length = C.uint8_t(C.strlen(name))
				g2qCrypt(&length, 1)
				if kind != 0 {
					name = C.nox_xxx_spellNameByN_424870(C.int(*g2qW(p, 0)))
				} else {
					name = C.nox_xxx_abilityGetName_425250(C.int(*g2qW(p, 0)))
				}
				g2qCrypt((*C.uint8_t)(unsafe.Pointer(name)), C.size_t(length))
			}
			g2qCrypt((*C.uint8_t)(unsafe.Add(p, 4)), 1)
		}
	}
	return 1
}

//export sub_460B90
func sub_460B90(show C.int) C.int {
	if show == 0 && memmap.Uint32(0x5D4594, 1049476) != 0 {
		sub_460920()
	}
	hidden := int32(1)
	if show != 0 {
		hidden = 0
	}
	g2qHidden(C.dword_5d4594_1049500, hidden)
	g2qHidden(C.dword_5d4594_1049504, hidden)
	g2qHidden(C.dword_5d4594_1049520, hidden)
	g2qHidden(C.dword_5d4594_1049508, hidden)
	g2qHidden(C.dword_5d4594_1049512, hidden)
	main := g2qMain()
	g2qHidden(nox_quickbar_root(main), hidden)
	for i := C.int(0); i < 5; i++ {
		g2qHidden(nox_quickbar_button(main, i), hidden)
		g2qHidden(nox_quickbar_nugget(main, i), hidden)
	}
	trapHidden := int32(0)
	if show == 0 || C.dword_5d4594_1049484 == 0 {
		trapHidden = 1
	}
	g2qHidden(nox_quickbar_root(g2qTrap()), trapHidden)
	if show != 0 && C.dword_5d4594_1049484 != 0 {
		for i := C.int(0); i < 3; i++ {
			g2qHidden(nox_quickbar_button(g2qTrap(), i), 0)
			g2qHidden(nox_quickbar_nugget(g2qTrap(), i), 0)
		}
		g2qHidden(C.dword_5d4594_1049512, 1)
	}
	return 1
}

func g2qCrypt(p *C.uint8_t, n C.size_t) {
	C.nox_xxx_fileReadWrite_426AC0_file3_fread_impl((*C.uchar)(unsafe.Pointer(p)), n, nil, 0)
}

func g2qCapture() *C.nox_window {
	if h := game2QuickbarTestHooks; h != nil && h.Capture != nil {
		return (*C.nox_window)(h.Capture())
	}
	return C.nox_xxx_wndGetCaptureMain_46AE00()
}
func g2qCaptureSet(w *C.nox_window) {
	if h := game2QuickbarTestHooks; h != nil && h.SetCapture != nil {
		h.SetCapture(unsafe.Pointer(w))
		return
	}
	C.nox_xxx_wndSetCaptureMain_46ADC0(w)
}
func g2qCaptureClear(w *C.nox_window) {
	if h := game2QuickbarTestHooks; h != nil && h.ClearCapture != nil {
		h.ClearCapture(unsafe.Pointer(w))
		return
	}
	C.nox_xxx_wndClearCaptureMain_46ADE0(w)
}
func g2qDragSave(spell uint32, kind int32) {
	if h := game2QuickbarTestHooks; h != nil && h.DragSave != nil {
		h.DragSave(spell, kind)
		return
	}
	C.nox_xxx_bookSaveSpellForDragDrop_477640(C.int(spell), C.int(kind))
}
func g2qDragClear() {
	if h := game2QuickbarTestHooks; h != nil && h.DragClear != nil {
		h.DragClear()
		return
	}
	C.nox_xxx_bookSpellDnDclear_477660()
}
func g2qBlocked() C.int {
	if h := game2QuickbarTestHooks; h != nil && h.Blocked != nil {
		return C.int(h.Blocked())
	}
	return C.nox_xxx_get_57AF20()
}
func g2qEnabled(spell uint32) bool {
	if h := game2QuickbarTestHooks; h != nil && h.Enabled != nil {
		return h.Enabled(spell)
	}
	return bool(C.nox_xxx_spellIsEnabled_424B70(C.int(spell)))
}
func g2qPointIn(w *C.nox_window, x, y C.int) bool {
	if h := game2QuickbarTestHooks; h != nil && h.PointIn != nil {
		return h.PointIn(unsafe.Pointer(w), int32(x), int32(y))
	}
	return bool(C.nox_xxx_wndPointInWnd_46AAB0(w, x, y))
}
func g2qPointRect(x, y C.int) bool {
	if h := game2QuickbarTestHooks; h != nil && h.PointRect != nil {
		return h.PointRect(int32(x), int32(y))
	}
	point := C.int2{field_0: x, field_4: y}
	return C.nox_xxx_pointInRect_4281F0(&point, (*C.int4)(memmap.PtrOff(0x587000, 133656))) != 0
}
func g2qEventDrop(spell uint32, flag byte, x, y C.int, data unsafe.Pointer) uint32 {
	if h := game2QuickbarTestHooks; h != nil && h.Drop != nil {
		return h.Drop(spell, flag, int32(x), int32(y), data)
	}
	return uint32(nox_xxx_bookSpellDrop_45DCA0(C.int(spell), C.char(flag), x, y, data))
}
func g2qTarget(spell uint32) {
	if h := game2QuickbarTestHooks; h != nil && h.TargetClick != nil {
		h.TargetClick(spell)
		return
	}
	C.nox_xxx_guiSpellTargetClickSet_45D9D0(C.int(spell))
}
func g2qStore(slot C.int) {
	if h := game2QuickbarTestHooks; h != nil && h.StoreButton != nil {
		h.StoreButton(int32(slot))
		return
	}
	C.nox_xxx_clientStoreLastButton_45DAD0(slot)
}
func g2qAbility(spell uint32) C.int {
	if h := game2QuickbarTestHooks; h != nil && h.AbilityCursor != nil {
		return C.int(h.AbilityCursor(spell))
	}
	return sub_45DFC0(C.int(spell))
}
func g2qCursor(spell uint32, flag byte) C.int {
	if h := game2QuickbarTestHooks; h != nil && h.SpellCursor != nil {
		return C.int(h.SpellCursor(spell, flag))
	}
	return nox_xxx_guiSpellSetCursor_45DF60(C.int(spell), C.char(flag))
}
func g2qEventSwap(data unsafe.Pointer, a, b C.int) {
	if h := game2QuickbarTestHooks; h != nil && h.EventSwap != nil {
		h.EventSwap(data, int32(a), int32(b))
		return
	}
	nox_xxx_clientSwapQuickbarKeys_45F300(data, a, b)
}
func g2qEventUpdate(data unsafe.Pointer) {
	if h := game2QuickbarTestHooks; h != nil && h.EventUpdate != nil {
		h.EventUpdate(data)
		return
	}
	nox_xxx_updateSpellIcons_45DDF0(data)
}

func g2qCancelAbility() {
	if h := game2QuickbarTestHooks; h != nil && h.CancelAbility != nil {
		h.CancelAbility()
		return
	}
	C.sub_4611B0()
}
