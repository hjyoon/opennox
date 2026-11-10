package legacy

/*
#include "game2_book_native.h"
*/
import "C"

import "unsafe"

func game2BookInputBlocked() bool {
	return C.dword_5d4594_1047520 == 1 || (nox_client_localPlayerHasFlags30(2) != 0 && C.sub_478030() == 0 && C.sub_47A260() == 0)
}

//export nox_xxx_bookWndProc_45B070
func nox_xxx_bookWndProc_45B070(unused1 *C.nox_window, event C.int) C.int {
	if C.dword_5d4594_1047520 == 1 {
		return 1
	}
	if event == 10 {
		nox_client_toggleSpellbook_45AC70()
		return 1
	}
	if nox_client_localPlayerHasFlags30(2) != 0 && C.sub_478030() == 0 && C.sub_47A260() == 0 {
		return 1
	}
	if event != 5 {
		return 0
	}
	game2BookHide(C.dword_5d4594_1046944, 0)
	if C.nox_xxx_aNox_cfg_0_587000_132132 != 0 {
		// DWORD addition occurs before the unsigned C comparison.
		if uint32(C.dword_5d4594_1046936)+1 < uint32(*game2BookU(1046940)) {
			C.dword_5d4594_1046936++
		} else {
			C.dword_5d4594_1046932 = 0
			C.nox_xxx_aNox_cfg_0_587000_132132 = 0
			game2BookHide(C.dword_5d4594_1046952, 0)
		}
		game2BookAnimChoose(C.dword_5d4594_1046924, C.dword_5d4594_1046872 != 0)
		C.dword_5d4594_1046868 = 2
		nox_book_anim_start(C.dword_5d4594_1046924)
		game2BookSound(788)
		return 1
	}
	if uint32(C.dword_5d4594_1046932) < *game2BookU(1047508)-uint32(C.dword_5d4594_1047512)-1 {
		C.dword_5d4594_1046932++
		game2BookAnimChoose(C.dword_5d4594_1046924, C.dword_5d4594_1046872 != 0)
		C.dword_5d4594_1046868 = 2
		nox_book_anim_start(C.dword_5d4594_1046924)
		game2BookSound(788)
	}
	return 1
}

//export nox_xxx_book_45B210
func nox_xxx_book_45B210(unused1 *C.nox_window, event C.int) C.int {
	if C.dword_5d4594_1047520 == 1 {
		return 1
	}
	if event == 10 {
		nox_client_toggleSpellbook_45AC70()
		return 1
	}
	if nox_client_localPlayerHasFlags30(2) != 0 && C.sub_478030() == 0 && C.sub_47A260() == 0 {
		return 1
	}
	if event != 5 {
		return 0
	}
	game2BookHide(C.dword_5d4594_1046948, 0)
	if C.nox_xxx_aNox_cfg_0_587000_132132 == 0 {
		if int32(C.dword_5d4594_1046932) < int32(*game2BookU(1047508))-int32(C.dword_5d4594_1047512) {
			if int32(C.dword_5d4594_1046932) <= 0 {
				C.nox_xxx_aNox_cfg_0_587000_132132 = 1
				C.dword_5d4594_1046936 = C.uint32_t(*game2BookU(1046940) - 1)
				game2BookHide(C.dword_5d4594_1046952, 1)
			} else {
				C.dword_5d4594_1046932--
			}
		} else {
			C.nox_xxx_aNox_cfg_0_587000_132132 = 1
			C.dword_5d4594_1046936 = 0
			game2BookHide(C.dword_5d4594_1046952, 1)
		}
	} else {
		if C.dword_5d4594_1046936 == 0 {
			return 1
		}
		C.dword_5d4594_1046936--
	}
	game2BookAnimChoose(C.dword_5d4594_1046928, C.dword_5d4594_1046872 != 0)
	C.dword_5d4594_1046868 = 3
	nox_book_anim_start(C.dword_5d4594_1046928)
	game2BookSound(788)
	return 1
}

//export nox_xxx_bookChildWndProcMB_45B360
func nox_xxx_bookChildWndProcMB_45B360(window *C.uint32_t, event C.uint) C.int {
	player := C.game2_book_player()
	changed := false
	if game2BookInputBlocked() {
		return 1
	}
	if event != 5 {
		if event <= 5 || event > 7 {
			return 0
		}
		return 1
	}
	if *window == 1310 {
		if C.dword_5d4594_1046872 != 0 {
			nox_xxx_bookClickSpell_45B1F0()
			if nox_xxx_guiSpellSortList_45ADF0(C.int(game2PlayerClass(player))) == 0 {
				nox_xxx_bookClickCreature_45B200()
				nox_xxx_guiSpellSortList_45ADF0(C.int(game2PlayerClass(player)))
				game2BookSound(925)
				return 1
			}
			changed = true
		}
		if C.nox_xxx_aNox_cfg_0_587000_132132 == 0 || C.dword_5d4594_1046936 != 0 || changed {
			C.nox_xxx_aNox_cfg_0_587000_132132 = 1
			C.dword_5d4594_1046936 = 0
			game2BookSound(788)
			C.dword_5d4594_1046868 = 3
			nox_book_anim_start(C.dword_5d4594_1046928)
			nox_book_anim_set_end(C.dword_5d4594_1046928, unsafe.Pointer(C.nox_xxx_bookClickSpell_45B1F0))
			game2BookHide(C.dword_5d4594_1046952, 1)
			game2BookHide(C.dword_5d4594_1046948, 0)
			C.nox_xxx_wndClearCaptureMain_46ADE0(C.nox_win_unk1)
			C.nox_xxx_bookSpellDnDclear_477660()
		}
		return 1
	}
	if *window != 1320 {
		return 1
	}
	if C.dword_5d4594_1046872 != 1 {
		nox_xxx_bookClickCreature_45B200()
		if nox_xxx_guiSpellSortList_45ADF0(C.int(game2PlayerClass(player))) == 0 {
			nox_xxx_bookClickSpell_45B1F0()
			nox_xxx_guiSpellSortList_45ADF0(C.int(game2PlayerClass(player)))
			game2BookSound(925)
			return 1
		}
		changed = true
	}
	if C.nox_xxx_aNox_cfg_0_587000_132132 != 0 && C.dword_5d4594_1046936 == 0 && !changed {
		return 1
	}
	C.nox_xxx_aNox_cfg_0_587000_132132 = 1
	C.dword_5d4594_1046936 = 0
	game2BookSound(788)
	ref := C.dword_5d4594_1046928
	C.dword_5d4594_1046868 = 3
	if changed {
		ref = C.dword_5d4594_1046924
		C.dword_5d4594_1046868 = 2
	}
	nox_book_anim_start(ref)
	nox_book_anim_set_end(ref, unsafe.Pointer(C.nox_xxx_bookClickCreature_45B200))
	game2BookHide(C.dword_5d4594_1046952, 1)
	game2BookHide(C.dword_5d4594_1046948, 0)
	C.nox_xxx_wndClearCaptureMain_46ADE0(C.nox_win_unk1)
	C.nox_xxx_bookSpellDnDclear_477660()
	return 1
}

//export nox_xxx_bookListWndProc_45B5F0
func nox_xxx_bookListWndProc_45B5F0(win *C.nox_window, event, pos C.uint) C.int {
	player := C.game2_book_player()
	y := int32(uint32(pos) >> 16)
	x := int32(uint16(pos))
	if game2BookInputBlocked() {
		return 1
	}
	if C.nox_xxx_aNox_cfg_0_587000_132132 == 0 {
		if event >= 5 {
			if event <= 8 {
				return 1
			}
			if event == 11 {
				nox_client_toggleSpellbook_45AC70()
				return 1
			}
		}
		return 0
	}
	C.dword_5d4594_1046656 = C.uint32_t(C.nox_xxx_guiFontHeightMB_43F320(nil) + 2)
	rows := int32(uint32(141)/uint32(C.dword_5d4594_1046656) - 1)
	var wx, wy C.uint
	C.nox_gui_getWindowOffs_46AA20(win, &wx, &wy)
	switch event {
	case 5:
		if nox_client_localPlayerHasFlags30(2) != 0 {
			return 1
		}
		C.dword_5d4594_1047536 = C.uint32_t(y)
		dy := int32(uint32(y) - uint32(wy) - 19)
		C.dword_5d4594_1047532 = C.uint32_t(x)
		if dy < 0 {
			C.nox_xxx_aNox_cfg_0_587000_132136 = ^C.uint32_t(0)
			return 1
		}
		row := dy / int32(C.dword_5d4594_1046656)
		if row >= rows {
			C.nox_xxx_aNox_cfg_0_587000_132136 = ^C.uint32_t(0)
			return 1
		}
		if x-int32(wx) > 145 {
			row += rows
		}
		C.nox_xxx_aNox_cfg_0_587000_132136 = C.uint32_t(uint32(2*rows)*uint32(C.dword_5d4594_1046936) + uint32(row))
		C.nox_xxx_wndSetCaptureMain_46ADC0(win)
		index := uint32(C.nox_xxx_aNox_cfg_0_587000_132136)
		if index >= *game2BookU(1047508)-uint32(C.dword_5d4594_1047512) {
			C.nox_xxx_aNox_cfg_0_587000_132136 = ^C.uint32_t(0)
			return 1
		}
		class := game2PlayerClass(player)
		if class == 0 && C.dword_5d4594_1046868 == 0 {
			C.dword_5d4594_1047528 = C.uint32_t(game2BookID(index))
		} else if class != 2 || C.dword_5d4594_1046868 != 1 {
			if C.dword_5d4594_1046868 != 0 {
				return 1
			}
			C.dword_5d4594_1047528 = C.uint32_t(game2BookID(index))
		} else {
			if player.spell_lvl[134] == 0 && (!game2BookGameFlag(0x2000) || game2BookGameFlag(4096)) {
				C.nox_xxx_wndClearCaptureMain_46ADE0(win)
				nox_xxx_bookMoveToPage_45B930(C.int(index))
				C.dword_5d4594_1047528 = 0
				C.nox_xxx_aNox_cfg_0_587000_132136 = ^C.uint32_t(0)
				return 1
			}
			C.dword_5d4594_1047528 = C.uint32_t(game2BookID(index) + 74)
		}
		C.nox_xxx_bookSaveSpellForDragDrop_477640(C.int(C.dword_5d4594_1047528), 1)
		game2BookSound(793)
		return 1
	case 6, 7:
		if int32(C.nox_xxx_aNox_cfg_0_587000_132136) >= 0 {
			dx := int32(uint32(C.dword_5d4594_1047532) - uint32(x))
			if dx < 0 {
				dx = int32(uint32(x) - uint32(C.dword_5d4594_1047532))
			}
			dy := int32(0)
			if dx < 5 {
				dy = int32(uint32(C.dword_5d4594_1047536) - uint32(y))
				if dy < 0 {
					dy = int32(uint32(y) - uint32(C.dword_5d4594_1047536))
				}
			}
			if dx >= 5 || dy >= 5 {
				C.nox_xxx_bookSpellDrop_45DCA0(C.int(C.dword_5d4594_1047528), 0, C.int(x), C.int(y), nil)
			} else {
				nox_xxx_bookMoveToPage_45B930(C.int(C.nox_xxx_aNox_cfg_0_587000_132136))
			}
		}
		if C.nox_xxx_wndGetCaptureMain_46AE00() != nil {
			C.nox_xxx_wndClearCaptureMain_46ADE0(win)
		}
		C.nox_xxx_bookSpellDnDclear_477660()
		return 1
	case 8:
		return 1
	case 11:
		nox_client_toggleSpellbook_45AC70()
		return 1
	default:
		return 0
	}
}

//export nox_xxx_bookMoveToPage_45B930
func nox_xxx_bookMoveToPage_45B930(index C.int) C.int {
	C.nox_xxx_aNox_cfg_0_587000_132132 = 0
	C.dword_5d4594_1046932 = C.uint32_t(index)
	C.dword_5d4594_1046936 = 99
	game2BookAnimChoose(C.dword_5d4594_1046924, C.dword_5d4594_1046872 != 0)
	C.dword_5d4594_1046868 = 2
	nox_book_anim_start(C.dword_5d4594_1046924)
	game2BookSound(788)
	game2BookHide(C.dword_5d4594_1046952, 0)
	game2BookHide(C.dword_5d4594_1046944, 0)
	return game2BookHide(C.dword_5d4594_1046948, 0)
}

//export nox_xxx_book_45BD30
func nox_xxx_book_45BD30(unused1 *C.nox_window, unused2 *C.nox_window_data) C.int { return 1 }

//export nox_xxx_bookInit_45B9D0
func nox_xxx_bookInit_45B9D0() C.int {
	C.dword_5d4594_1047516 = C.dword_8531A0_2576
	loadImage := func(name string) *C.nox_video_bag_image_t {
		p := C.CString(name)
		defer C.free(unsafe.Pointer(p))
		return C.nox_xxx_gLoadImg_42F970(p)
	}
	for i, name := range []string{"ArrowNW", "ArrowN", "ArrowNE", "ArrowW", "ArrowE", "ArrowSW", "ArrowS", "ArrowSE"} {
		image := loadImage(name)
		C.nox_book_arrow_images[i] = image
		if image == nil {
			return 0
		}
	}
	C.nox_book_background_image = loadImage("BookOfKnowledge")
	if C.nox_book_background_image == nil {
		return 0
	}
	C.nox_book_guide_tab_lit_image = loadImage("GuideTabLit")
	if C.nox_book_guide_tab_lit_image == nil {
		return 0
	}
	C.nox_book_spell_tab_lit_image = loadImage("SpellTabLit")
	if C.nox_book_spell_tab_lit_image == nil {
		return 0
	}
	loadAnim := func(name string) *C.nox_things_imageRef_t {
		p := C.CString(name)
		defer C.free(unsafe.Pointer(p))
		return C.nox_xxx_gLoadAnim_42FA20(p)
	}
	C.dword_5d4594_1046924 = loadAnim("BookPageForward")
	if C.dword_5d4594_1046924 == nil {
		return 0
	}
	nox_book_anim_set_end(C.dword_5d4594_1046924, unsafe.Pointer(C.nox_xxx_bookClickSpell_45B1F0))
	C.dword_5d4594_1046928 = loadAnim("BookPageBackward")
	if C.dword_5d4594_1046928 == nil {
		return 0
	}
	nox_book_anim_set_end(C.dword_5d4594_1046928, unsafe.Pointer(C.nox_xxx_bookClickSpell_45B1F0))
	C.nox_win_unk1 = C.nox_window_new(nil, 1196, 5, C.nox_win_height-323, 285, 168, nil)
	C.nox_window_set_all_funcs(C.nox_win_unk1, unsafe.Pointer(C.nox_xxx_bookListWndProc_45B5F0), unsafe.Pointer(C.nox_xxx_bookDrawList_45BD40), nil)
	game2BookHide(C.nox_win_unk1, 1)
	first := C.nox_window_new(C.nox_win_unk1, 8, 257, 15, 27, 40, nil)
	if first == nil {
		return 0
	}
	C.nox_window_set_all_funcs(first, unsafe.Pointer(C.nox_xxx_bookChildWndProcMB_45B360), unsafe.Pointer(C.nox_xxx_book_45BD30), unsafe.Pointer(C.nox_xxx_book_45CF00))
	first.id = 1320
	second := C.nox_window_new(C.nox_win_unk1, 8, 253, 61, 27, 40, nil)
	if second == nil {
		return 0
	}
	C.nox_window_set_all_funcs(second, unsafe.Pointer(C.nox_xxx_bookChildWndProcMB_45B360), unsafe.Pointer(C.nox_xxx_book_45BD30), unsafe.Pointer(C.nox_xxx_book_45CF00))
	second.id = 1310
	C.dword_5d4594_1046944 = C.nox_window_new(C.nox_win_unk1, 136, 24, 138, 20, 20, nil)
	C.nox_window_set_all_funcs(C.dword_5d4594_1046944, unsafe.Pointer(C.nox_xxx_book_45B210), nil, nil)
	C.nox_xxx_wndSetIcon_46AE60(C.dword_5d4594_1046944, loadImage("ArrowW"))
	C.dword_5d4594_1046948 = C.nox_window_new(C.nox_win_unk1, 136, 233, 138, 20, 20, nil)
	C.nox_window_set_all_funcs(C.dword_5d4594_1046948, unsafe.Pointer(C.nox_xxx_bookWndProc_45B070), nil, nil)
	C.nox_xxx_wndSetIcon_46AE60(C.dword_5d4594_1046948, loadImage("ArrowE"))
	C.dword_5d4594_1046952 = C.nox_window_new(C.nox_win_unk1, 8, 63, 19, 30, 30, nil)
	C.nox_window_set_all_funcs(C.dword_5d4594_1046952, unsafe.Pointer(C.nox_xxx_bookWndFn_45CC10), unsafe.Pointer(C.nox_xxx_bookDrawIconFn_45CB30), nil)
	game2BookHide(C.dword_5d4594_1046952, 1)
	C.dword_5d4594_1046956 = C.nox_window_new(nil, 40, 0, 0, 30, 30, nil)
	C.nox_window_set_all_funcs(C.dword_5d4594_1046956, nil, unsafe.Pointer(C.nox_xxx_bookDrawFn_45C7D0), nil)
	game2BookHide(C.dword_5d4594_1046956, 1)
	C.dword_5d4594_1046868 = 0
	C.dword_5d4594_1046872 = 0
	C.nox_xxx_aNox_cfg_0_587000_132132 = 1
	C.dword_5d4594_1046936 = 0
	return 1
}

//export nox_xxx_bookDrawIconFn_45CB30
func nox_xxx_bookDrawIconFn_45CB30(window *C.uint32_t) C.int {
	player := C.game2_book_player()
	var image unsafe.Pointer
	if C.dword_5d4594_1046868 != 0 {
		if C.dword_5d4594_1046868 != 1 || game2PlayerClass(player) != 2 || (player.spell_lvl[134] == 0 && (!game2BookGameFlag(0x2000) || game2BookGameFlag(4096))) {
			return 1
		}
		image = C.nox_xxx_spellIcon_424A90(C.int(game2BookID(uint32(C.dword_5d4594_1046932)) + 74))
	} else if game2PlayerClass(player) != 0 {
		image = C.nox_xxx_spellIcon_424A90(C.int(game2BookID(uint32(C.dword_5d4594_1046932))))
	} else {
		image = unsafe.Pointer(C.nox_xxx_spellGetAbilityIcon_425310(C.int(game2BookID(uint32(C.dword_5d4594_1046932))), 0))
	}
	if image != nil {
		var x, y C.uint
		C.nox_client_wndGetPosition_46AA60((*C.nox_window)(unsafe.Pointer(window)), &x, &y)
		C.nox_client_drawImageAt_47D2C0((*C.nox_video_bag_image_t)(image), C.int(x), C.int(y))
	}
	return 1
}

//export nox_xxx_bookWndFn_45CC10
func nox_xxx_bookWndFn_45CC10(window *C.uint32_t, event C.int, pos C.uint) C.int {
	player := C.game2_book_player()
	win := (*C.nox_window)(unsafe.Pointer(window))
	y := C.int(uint32(pos) >> 16)
	x := C.int(uint16(pos))
	if nox_client_localPlayerHasFlags30(2) != 0 {
		return 1
	}
	if C.dword_5d4594_1046868 == 1 {
		class := game2PlayerClass(player)
		if class == 1 || class == 0 || (!game2BookGameFlag(0x2000) && !game2BookGameFlag(4096) && player.spell_lvl[134] == 0) {
			return 0
		}
	}
	drag := func(id uint32) {
		C.dword_5d4594_1047540 = C.uint32_t(id)
		C.nox_xxx_bookSaveSpellForDragDrop_477640(C.int(id), 1)
		C.nox_xxx_wndSetCaptureMain_46ADC0(win)
		game2BookSound(793)
	}
	finish := func(clear bool) {
		if clear {
			C.dword_5d4594_1047540 = 0
			C.nox_xxx_wndClearCaptureMain_46ADE0(win)
		}
		C.nox_xxx_bookSpellDnDclear_477660()
	}
	switch event {
	case 5:
		if C.nox_xxx_wndGetCaptureMain_46AE00() != nil || C.dword_5d4594_1047540 != 0 {
			return 1
		}
		if game2PlayerClass(player) == 0 && C.dword_5d4594_1046868 == 0 {
			drag(game2BookID(uint32(C.dword_5d4594_1046932)))
			return 1
		}
		if game2PlayerClass(player) == 2 && C.dword_5d4594_1046868 == 1 {
			drag(game2BookID(uint32(C.dword_5d4594_1046932)) + 74)
			return 1
		}
		if C.dword_5d4594_1046868 != 0 {
			return 1
		}
		id := game2BookID(uint32(C.dword_5d4594_1046932))
		C.dword_5d4594_1047540 = C.uint32_t(id)
		if !game2BookSpellFlag(C.int(id), 0x15000) {
			drag(id)
		} else {
			C.dword_5d4594_1047540 = 0
		}
		return 1
	case 6, 7:
		if C.dword_5d4594_1047540 == 0 {
			return 1
		}
		id := C.int(C.dword_5d4594_1047540)
		if !bool(C.nox_xxx_wndPointInWnd_46AAB0(win, x, y)) {
			if C.nox_xxx_guiSpell_460650() != 0 {
				nox_xxx_guiSpellTargetClickSet_45D9D0(id)
			} else if C.sub_4611A0() != 0 {
				C.sub_4611B0()
			} else {
				C.nox_xxx_bookSpellDrop_45DCA0(id, 0, x, y, nil)
			}
			finish(true)
			return 1
		}
		if game2PlayerClass(player) != 0 {
			if C.nox_xxx_guiSpell_460650() != 0 {
				C.nox_xxx_guiSpellSetCursor_45DF60(0, 0)
				finish(true)
				return 1
			}
			if !game2BookSpellFlag(id, 0x600) {
				C.nox_xxx_guiSpellSetCursor_45DF60(id, 0)
				finish(false)
				return 1
			}
			if C.nox_xxx_guiSpellSetCursor_45DF60(id, 1) != 0 {
				finish(true)
				return 1
			}
		} else {
			if C.sub_4611A0() != 0 {
				C.sub_45DFC0(0)
				finish(true)
				return 1
			}
			if C.sub_45DFC0(id) != 0 {
				finish(true)
				return 1
			}
		}
		finish(false)
		return 1
	case 8:
		return 1
	default:
		return 0
	}
}
