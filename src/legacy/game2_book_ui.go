package legacy

/*
#include "game2_book_native.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// Share native window and image slots with the retained C book reference.
var bookUI, _ = alloc.New(C.nox_book_ui_slots{})

var bookDetails, _ = alloc.New(C.nox_book_details_t{})

var bookState = func() *C.nox_book_state_t {
	state, _ := alloc.New(C.nox_book_state_t{})
	state.show_detail_page = 1
	state.list_drag_index = ^C.uint32_t(0)
	return state
}()

//export nox_book_state_storage
func nox_book_state_storage() *C.nox_book_state_t { return bookState }

//export nox_book_details_storage
func nox_book_details_storage() *C.nox_book_details_t { return bookDetails }

//export nox_book_ui_storage
func nox_book_ui_storage() *C.nox_book_ui_slots {
	return bookUI
}

func game2BookInputBlocked() bool {
	return bookState.reward_active == 1 || (nox_client_localPlayerHasFlags30(2) != 0 && C.sub_478030() == 0 && C.sub_47A260() == 0)
}

//export nox_xxx_bookWndProc_45B070
func nox_xxx_bookWndProc_45B070(unused1 *C.nox_window, event C.int) C.int {
	if bookState.reward_active == 1 {
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
	game2BookHide(bookDetails.left_arrow, 0)
	if bookState.show_detail_page != 0 {
		// DWORD addition occurs before the unsigned C comparison.
		if uint32(bookState.list_page)+1 < uint32(*game2BookU(1046940)) {
			bookState.list_page++
		} else {
			bookState.entry_index = 0
			bookState.show_detail_page = 0
			game2BookHide(bookDetails.page_icon, 0)
		}
		game2BookAnimChoose(bookDetails.forward, bookState.guide_mode != 0)
		bookState.page_mode = 2
		nox_book_anim_start(bookDetails.forward)
		game2BookSound(788)
		return 1
	}
	if uint32(bookState.entry_index) < *game2BookU(1047508)-uint32(bookState.hidden_entry_count)-1 {
		bookState.entry_index++
		game2BookAnimChoose(bookDetails.forward, bookState.guide_mode != 0)
		bookState.page_mode = 2
		nox_book_anim_start(bookDetails.forward)
		game2BookSound(788)
	}
	return 1
}

//export nox_xxx_book_45B210
func nox_xxx_book_45B210(unused1 *C.nox_window, event C.int) C.int {
	if bookState.reward_active == 1 {
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
	game2BookHide(bookDetails.right_arrow, 0)
	if bookState.show_detail_page == 0 {
		if int32(bookState.entry_index) < int32(*game2BookU(1047508))-int32(bookState.hidden_entry_count) {
			if int32(bookState.entry_index) <= 0 {
				bookState.show_detail_page = 1
				bookState.list_page = C.uint32_t(*game2BookU(1046940) - 1)
				game2BookHide(bookDetails.page_icon, 1)
			} else {
				bookState.entry_index--
			}
		} else {
			bookState.show_detail_page = 1
			bookState.list_page = 0
			game2BookHide(bookDetails.page_icon, 1)
		}
	} else {
		if bookState.list_page == 0 {
			return 1
		}
		bookState.list_page--
	}
	game2BookAnimChoose(bookDetails.backward, bookState.guide_mode != 0)
	bookState.page_mode = 3
	nox_book_anim_start(bookDetails.backward)
	game2BookSound(788)
	return 1
}

//export nox_xxx_bookChildWndProcMB_45B360
func nox_xxx_bookChildWndProcMB_45B360(window *C.uint32_t, event C.uint) C.int {
	player := bookSelectedPlayer()
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
		if bookState.guide_mode != 0 {
			nox_xxx_bookClickSpell_45B1F0()
			if nox_xxx_guiSpellSortList_45ADF0(C.int(game2PlayerClass(player))) == 0 {
				nox_xxx_bookClickCreature_45B200()
				nox_xxx_guiSpellSortList_45ADF0(C.int(game2PlayerClass(player)))
				game2BookSound(925)
				return 1
			}
			changed = true
		}
		if bookState.show_detail_page == 0 || bookState.list_page != 0 || changed {
			bookState.show_detail_page = 1
			bookState.list_page = 0
			game2BookSound(788)
			bookState.page_mode = 3
			nox_book_anim_start(bookDetails.backward)
			nox_book_anim_set_end(bookDetails.backward, unsafe.Pointer(C.nox_xxx_bookClickSpell_45B1F0))
			game2BookHide(bookDetails.page_icon, 1)
			game2BookHide(bookDetails.right_arrow, 0)
			nox_xxx_wndClearCaptureMain_46ADE0(bookUI.list_window)
			C.nox_xxx_bookSpellDnDclear_477660()
		}
		return 1
	}
	if *window != 1320 {
		return 1
	}
	if bookState.guide_mode != 1 {
		nox_xxx_bookClickCreature_45B200()
		if nox_xxx_guiSpellSortList_45ADF0(C.int(game2PlayerClass(player))) == 0 {
			nox_xxx_bookClickSpell_45B1F0()
			nox_xxx_guiSpellSortList_45ADF0(C.int(game2PlayerClass(player)))
			game2BookSound(925)
			return 1
		}
		changed = true
	}
	if bookState.show_detail_page != 0 && bookState.list_page == 0 && !changed {
		return 1
	}
	bookState.show_detail_page = 1
	bookState.list_page = 0
	game2BookSound(788)
	ref := bookDetails.backward
	bookState.page_mode = 3
	if changed {
		ref = bookDetails.forward
		bookState.page_mode = 2
	}
	nox_book_anim_start(ref)
	nox_book_anim_set_end(ref, unsafe.Pointer(C.nox_xxx_bookClickCreature_45B200))
	game2BookHide(bookDetails.page_icon, 1)
	game2BookHide(bookDetails.right_arrow, 0)
	nox_xxx_wndClearCaptureMain_46ADE0(bookUI.list_window)
	C.nox_xxx_bookSpellDnDclear_477660()
	return 1
}

//export nox_xxx_bookListWndProc_45B5F0
func nox_xxx_bookListWndProc_45B5F0(win *C.nox_window, event, pos C.uint) C.int {
	player := bookSelectedPlayer()
	y := int32(uint32(pos) >> 16)
	x := int32(uint16(pos))
	if game2BookInputBlocked() {
		return 1
	}
	if bookState.show_detail_page == 0 {
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
	bookState.line_height = C.uint32_t(C.nox_xxx_guiFontHeightMB_43F320(nil) + 2)
	rows := int32(uint32(141)/uint32(bookState.line_height) - 1)
	var wx, wy C.uint
	nox_gui_getWindowOffs_46AA20(win, &wx, &wy)
	switch event {
	case 5:
		if nox_client_localPlayerHasFlags30(2) != 0 {
			return 1
		}
		bookState.list_drag_y = C.uint32_t(y)
		dy := int32(uint32(y) - uint32(wy) - 19)
		bookState.list_drag_x = C.uint32_t(x)
		if dy < 0 {
			bookState.list_drag_index = ^C.uint32_t(0)
			return 1
		}
		row := dy / int32(bookState.line_height)
		if row >= rows {
			bookState.list_drag_index = ^C.uint32_t(0)
			return 1
		}
		if x-int32(wx) > 145 {
			row += rows
		}
		bookState.list_drag_index = C.uint32_t(uint32(2*rows)*uint32(bookState.list_page) + uint32(row))
		nox_xxx_wndSetCaptureMain_46ADC0(win)
		index := uint32(bookState.list_drag_index)
		if index >= *game2BookU(1047508)-uint32(bookState.hidden_entry_count) {
			bookState.list_drag_index = ^C.uint32_t(0)
			return 1
		}
		class := game2PlayerClass(player)
		if class == 0 && bookState.page_mode == 0 {
			bookState.list_drag_id = C.uint32_t(game2BookID(index))
		} else if class != 2 || bookState.page_mode != 1 {
			if bookState.page_mode != 0 {
				return 1
			}
			bookState.list_drag_id = C.uint32_t(game2BookID(index))
		} else {
			if player.spell_lvl[134] == 0 && (!game2BookGameFlag(0x2000) || game2BookGameFlag(4096)) {
				nox_xxx_wndClearCaptureMain_46ADE0(win)
				nox_xxx_bookMoveToPage_45B930(C.int(index))
				bookState.list_drag_id = 0
				bookState.list_drag_index = ^C.uint32_t(0)
				return 1
			}
			bookState.list_drag_id = C.uint32_t(game2BookID(index) + 74)
		}
		C.nox_xxx_bookSaveSpellForDragDrop_477640(C.int(bookState.list_drag_id), 1)
		game2BookSound(793)
		return 1
	case 6, 7:
		if int32(bookState.list_drag_index) >= 0 {
			dx := int32(uint32(bookState.list_drag_x) - uint32(x))
			if dx < 0 {
				dx = int32(uint32(x) - uint32(bookState.list_drag_x))
			}
			dy := int32(0)
			if dx < 5 {
				dy = int32(uint32(bookState.list_drag_y) - uint32(y))
				if dy < 0 {
					dy = int32(uint32(y) - uint32(bookState.list_drag_y))
				}
			}
			if dx >= 5 || dy >= 5 {
				C.nox_xxx_bookSpellDrop_45DCA0(C.int(bookState.list_drag_id), 0, C.int(x), C.int(y), nil)
			} else {
				nox_xxx_bookMoveToPage_45B930(C.int(bookState.list_drag_index))
			}
		}
		if nox_xxx_wndGetCaptureMain_46AE00() != nil {
			nox_xxx_wndClearCaptureMain_46ADE0(win)
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
	bookState.show_detail_page = 0
	bookState.entry_index = C.uint32_t(index)
	bookState.list_page = 99
	game2BookAnimChoose(bookDetails.forward, bookState.guide_mode != 0)
	bookState.page_mode = 2
	nox_book_anim_start(bookDetails.forward)
	game2BookSound(788)
	game2BookHide(bookDetails.page_icon, 0)
	game2BookHide(bookDetails.left_arrow, 0)
	return game2BookHide(bookDetails.right_arrow, 0)
}

//export nox_xxx_book_45BD30
func nox_xxx_book_45BD30(unused1 *C.nox_window, unused2 *C.nox_window_data) C.int { return 1 }

//export nox_xxx_bookInit_45B9D0
func nox_xxx_bookInit_45B9D0() C.int {
	bookState.player_word = worldOwnerStorage.local_player_word
	loadImage := func(name string) *C.nox_video_bag_image_t {
		p := C.CString(name)
		defer C.free(unsafe.Pointer(p))
		return C.nox_xxx_gLoadImg_42F970(p)
	}
	for i, name := range []string{"ArrowNW", "ArrowN", "ArrowNE", "ArrowW", "ArrowE", "ArrowSW", "ArrowS", "ArrowSE"} {
		image := loadImage(name)
		bookUI.arrows[i] = image
		if image == nil {
			return 0
		}
	}
	bookUI.background = loadImage("BookOfKnowledge")
	if bookUI.background == nil {
		return 0
	}
	bookUI.guide_tab_lit = loadImage("GuideTabLit")
	if bookUI.guide_tab_lit == nil {
		return 0
	}
	bookUI.spell_tab_lit = loadImage("SpellTabLit")
	if bookUI.spell_tab_lit == nil {
		return 0
	}
	loadAnim := func(name string) *C.nox_things_imageRef_t {
		p := C.CString(name)
		defer C.free(unsafe.Pointer(p))
		return C.nox_xxx_gLoadAnim_42FA20(p)
	}
	bookDetails.forward = loadAnim("BookPageForward")
	if bookDetails.forward == nil {
		return 0
	}
	nox_book_anim_set_end(bookDetails.forward, unsafe.Pointer(C.nox_xxx_bookClickSpell_45B1F0))
	bookDetails.backward = loadAnim("BookPageBackward")
	if bookDetails.backward == nil {
		return 0
	}
	nox_book_anim_set_end(bookDetails.backward, unsafe.Pointer(C.nox_xxx_bookClickSpell_45B1F0))
	bookUI.list_window = nox_window_new(nil, 1196, 5, clientDisplayStorage.height-323, 285, 168, nil)
	root := bookUI.list_window
	asWindow(root).AddCleanup(func() {
		if bookUI.list_window == root {
			bookUI.list_window = nil
			bookDetails.left_arrow, bookDetails.right_arrow, bookDetails.page_icon = nil, nil, nil
		}
	})
	nox_window_set_all_funcs(bookUI.list_window, unsafe.Pointer(C.nox_xxx_bookListWndProc_45B5F0), unsafe.Pointer(C.nox_xxx_bookDrawList_45BD40), nil)
	game2BookHide(bookUI.list_window, 1)
	first := nox_window_new(bookUI.list_window, 8, 257, 15, 27, 40, nil)
	if first == nil {
		return 0
	}
	nox_window_set_all_funcs(first, unsafe.Pointer(C.nox_xxx_bookChildWndProcMB_45B360), unsafe.Pointer(C.nox_xxx_book_45BD30), unsafe.Pointer(C.nox_xxx_book_45CF00))
	first.id = 1320
	second := nox_window_new(bookUI.list_window, 8, 253, 61, 27, 40, nil)
	if second == nil {
		return 0
	}
	nox_window_set_all_funcs(second, unsafe.Pointer(C.nox_xxx_bookChildWndProcMB_45B360), unsafe.Pointer(C.nox_xxx_book_45BD30), unsafe.Pointer(C.nox_xxx_book_45CF00))
	second.id = 1310
	bookDetails.left_arrow = nox_window_new(bookUI.list_window, 136, 24, 138, 20, 20, nil)
	nox_window_set_all_funcs(bookDetails.left_arrow, unsafe.Pointer(C.nox_xxx_book_45B210), nil, nil)
	nox_xxx_wndSetIcon_46AE60(bookDetails.left_arrow, loadImage("ArrowW"))
	bookDetails.right_arrow = nox_window_new(bookUI.list_window, 136, 233, 138, 20, 20, nil)
	nox_window_set_all_funcs(bookDetails.right_arrow, unsafe.Pointer(C.nox_xxx_bookWndProc_45B070), nil, nil)
	nox_xxx_wndSetIcon_46AE60(bookDetails.right_arrow, loadImage("ArrowE"))
	bookDetails.page_icon = nox_window_new(bookUI.list_window, 8, 63, 19, 30, 30, nil)
	nox_window_set_all_funcs(bookDetails.page_icon, unsafe.Pointer(C.nox_xxx_bookWndFn_45CC10), unsafe.Pointer(C.nox_xxx_bookDrawIconFn_45CB30), nil)
	game2BookHide(bookDetails.page_icon, 1)
	bookDetails.reward = nox_window_new(nil, 40, 0, 0, 30, 30, nil)
	reward := bookDetails.reward
	asWindow(reward).AddCleanup(func() {
		if bookDetails.reward == reward {
			bookDetails.reward = nil
		}
	})
	nox_window_set_all_funcs(bookDetails.reward, nil, unsafe.Pointer(C.nox_xxx_bookDrawFn_45C7D0), nil)
	game2BookHide(bookDetails.reward, 1)
	bookState.page_mode = 0
	bookState.guide_mode = 0
	bookState.show_detail_page = 1
	bookState.list_page = 0
	return 1
}

//export nox_xxx_bookDrawIconFn_45CB30
func nox_xxx_bookDrawIconFn_45CB30(window *C.uint32_t) C.int {
	player := bookSelectedPlayer()
	var image unsafe.Pointer
	if bookState.page_mode != 0 {
		if bookState.page_mode != 1 || game2PlayerClass(player) != 2 || (player.spell_lvl[134] == 0 && (!game2BookGameFlag(0x2000) || game2BookGameFlag(4096))) {
			return 1
		}
		image = C.nox_xxx_spellIcon_424A90(C.int(game2BookID(uint32(bookState.entry_index)) + 74))
	} else if game2PlayerClass(player) != 0 {
		image = C.nox_xxx_spellIcon_424A90(C.int(game2BookID(uint32(bookState.entry_index))))
	} else {
		image = unsafe.Pointer(C.nox_xxx_spellGetAbilityIcon_425310(C.int(game2BookID(uint32(bookState.entry_index))), 0))
	}
	if image != nil {
		var x, y C.uint
		nox_client_wndGetPosition_46AA60((*C.nox_window)(unsafe.Pointer(window)), &x, &y)
		C.nox_client_drawImageAt_47D2C0((*C.nox_video_bag_image_t)(image), C.int(x), C.int(y))
	}
	return 1
}

//export nox_xxx_bookWndFn_45CC10
func nox_xxx_bookWndFn_45CC10(window *C.uint32_t, event C.int, pos C.uint) C.int {
	player := bookSelectedPlayer()
	win := (*C.nox_window)(unsafe.Pointer(window))
	y := C.int(uint32(pos) >> 16)
	x := C.int(uint16(pos))
	if nox_client_localPlayerHasFlags30(2) != 0 {
		return 1
	}
	if bookState.page_mode == 1 {
		class := game2PlayerClass(player)
		if class == 1 || class == 0 || (!game2BookGameFlag(0x2000) && !game2BookGameFlag(4096) && player.spell_lvl[134] == 0) {
			return 0
		}
	}
	drag := func(id uint32) {
		bookState.page_drag_id = C.uint32_t(id)
		C.nox_xxx_bookSaveSpellForDragDrop_477640(C.int(id), 1)
		nox_xxx_wndSetCaptureMain_46ADC0(win)
		game2BookSound(793)
	}
	finish := func(clear bool) {
		if clear {
			bookState.page_drag_id = 0
			nox_xxx_wndClearCaptureMain_46ADE0(win)
		}
		C.nox_xxx_bookSpellDnDclear_477660()
	}
	switch event {
	case 5:
		if nox_xxx_wndGetCaptureMain_46AE00() != nil || bookState.page_drag_id != 0 {
			return 1
		}
		if game2PlayerClass(player) == 0 && bookState.page_mode == 0 {
			drag(game2BookID(uint32(bookState.entry_index)))
			return 1
		}
		if game2PlayerClass(player) == 2 && bookState.page_mode == 1 {
			drag(game2BookID(uint32(bookState.entry_index)) + 74)
			return 1
		}
		if bookState.page_mode != 0 {
			return 1
		}
		id := game2BookID(uint32(bookState.entry_index))
		bookState.page_drag_id = C.uint32_t(id)
		if !game2BookSpellFlag(C.int(id), 0x15000) {
			drag(id)
		} else {
			bookState.page_drag_id = 0
		}
		return 1
	case 6, 7:
		if bookState.page_drag_id == 0 {
			return 1
		}
		id := C.int(bookState.page_drag_id)
		if !bool(nox_xxx_wndPointInWnd_46AAB0(win, x, y)) {
			if C.nox_xxx_guiSpell_460650() != 0 {
				nox_xxx_guiSpellTargetClickSet_45D9D0(id)
			} else if C.sub_4611A0() != 0 {
				sub_4611B0()
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
