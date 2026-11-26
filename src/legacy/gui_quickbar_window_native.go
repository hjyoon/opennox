package legacy

/*
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "client__gui__guispell.h"
#include "client__gui__window.h"

extern uint32_t dword_5d4594_1049484;
extern uint32_t dword_5d4594_1049536;
extern uint32_t dword_5d4594_1047548;
extern uint32_t dword_5d4594_1047552;
extern int nox_win_width;
extern int nox_win_height;
*/
import "C"

import (
	"fmt"
	"unsafe"

	playerlib "github.com/opennox/libs/player"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
)

func quickbarImage(name string) *nox_video_bag_image_t {
	return nox_xxx_gLoadImg_42F970(internCStr(name))
}

func quickbarIcons(win *nox_window, base, lit string) {
	nox_xxx_wndSetIcon_46AE60(win, quickbarImage(base))
	nox_xxx_wndSetIconLit_46AEA0(win, quickbarImage(lit))
}

func quickbarTooltip(win *nox_window, key string, line int32) {
	nox_xxx_wndWddSetTooltip_46B000(&win.draw_data, quickbarText(key, line))
}

type quickbarArrowSpec struct {
	x, y, width, height C.int
	iconX, iconY        C.int
	image, tooltip      string
	sourceLine          int32
	id                  C.uint
}

func newQuickbarArrow(parent *nox_window, spec quickbarArrowSpec) {
	win := nox_window_new(parent, 1032, spec.x, spec.y, spec.width, spec.height, nil)
	nox_xxx_wndSetOffsetMB_46AE40(win, spec.iconX, spec.iconY)
	nox_xxx_wndSetIcon_46AE60(win, nil)
	nox_xxx_wndSetIconLit_46AEA0(win, quickbarImage(spec.image))
	nox_window_set_all_funcs(win, unsafe.Pointer(C.nox_xxx_quickbarTrapUpDownProc_45F630),
		unsafe.Pointer(C.nox_xxx_quickbarTrapUpDownDraw_45F6F0), nil)
	win.field_92 = spec.id
	quickbarTooltip(win, spec.tooltip, spec.sourceLine)
}

func storeQuickbarControl(slot **nox_window, win *nox_window) {
	*slot = win
	asWindow(win).AddCleanup(func() {
		if *slot == win {
			*slot = nil
		}
	})
}

func createQuickbarRight(x, y C.int, player *C.nox_playerInfo, class playerlib.Class) {
	warrior := class == playerlib.Warrior
	style := C.int(1160)
	if warrior {
		style = 1672
	}
	root := nox_window_new(nil, style, x+260, y, 45, 66, nil)
	storeQuickbarControl(&quickbarControls.right, root)
	nox_xxx_wndSetOffsetMB_46AE40(root, -263, 0)
	C.dword_5d4594_1049536 = C.uint32_t(y)
	nox_xxx_wndSetWindowProc_46B300(root, unsafe.Pointer(C.nox_xxx_quickbar_45F8D0))
	if warrior {
		quickbarIcons(root, "QuickBarWarriorRight", "QuickBarWarriorRight")
		return
	}

	activate := nox_window_new(root, 1032, 9, 33, 32, 32, nil)
	storeQuickbarControl(&quickbarControls.trap_activate, activate)
	registerQuickbarDragCleanup(activate)
	nox_window_set_all_funcs(activate, unsafe.Pointer(C.nox_xxx_quickbarTrapButtonProc_45F7A0),
		unsafe.Pointer(C.nox_xxx_quickbarDrawFn_460000), nil)
	construct := nox_window_new(root, 1160, 0, 19, 12, 12, nil)
	storeQuickbarControl(&quickbarControls.trap_construct, construct)
	nox_xxx_wndSetWindowProc_46B300(construct, unsafe.Pointer(C.nox_xxx_quickbarTrapProc_45FB90))
	nox_xxx_wndSetOffsetMB_46AE40(construct, -265, -23)
	quickbarIcons(construct, "QuickBarTrapButton", "QuickBarTrapButtonLit")

	var icon, tooltip string
	var line int32
	switch class {
	case playerlib.Wizard:
		icon, tooltip, line = "QuickBarTrap", "ToolTipLayTrap", 1805
	case playerlib.Conjurer:
		icon, tooltip, line = "QuickBarBomber", "ToolTipSummonBomber", 1838
	default:
		return
	}
	if player.spell_lvl[34] == 0 && (!noxflags.HasGame(noxflags.GameOnline) ||
		noxflags.HasGame(noxflags.GameModeQuest) || nox_xxx_isQuest_4D6F50() != 0 || sub_4D6F70() != 0) {
		quickbarIcons(root, "QuickBarWarriorRight", "QuickBarWarriorRight")
		nox_xxx_wndClearFlag_46AD80(activate, 8)
		nox_xxx_wndClearFlag_46AD80(construct, 8)
		nox_xxx_wndSetDrawFn_46B340(construct, unsafe.Pointer(C.nox_xxx_quickbarButtonBookDraw_45EF30))
		return
	}
	quickbarIcons(root, icon, icon+"Hit")
	quickbarTooltip(activate, tooltip, line)
	nox_xxx_wnd_46AD60(activate, 8)
	quickbarTooltip(construct, "ToolTipTrapConstruct", line+3)
	nox_xxx_wnd_46AD60(construct, 8)
}

func createQuickbarTrapTray(x, y, fontHeight C.int) {
	bar := g2qTrap()
	nox_xxx_quickBarInitWindow_4601F0(bar, x+122, y-17, 3, 21,
		unsafe.Pointer(C.nox_xxx_quickBarWnd_45EF50), unsafe.Pointer(C.nox_xxx_quickBarWarriorDraw_45FDE0))
	root := nox_quickbar_root(bar)
	nox_xxx_wndSetIcon_46AE60(root, quickbarImage("QuickBarTrapTray"))
	nox_xxx_wndSetOffsetMB_46AE40(root, -40, -20)
	*memmap.PtrUint8(0x5D4594, 1048192) |= 1
	nox_window_set_hidden(root, 1)
	C.dword_5d4594_1049484 = 0

	label := nox_window_new(root, 1032, 20, -7, 110, fontHeight, nil)
	nox_window_set_all_funcs(label, nil, unsafe.Pointer(C.nox_xxx_quickbarDraw_45FAC0), nil)
	newQuickbarArrow(root, quickbarArrowSpec{
		x: 15, y: 12, width: 10, height: 14, iconX: -55, iconY: -32,
		image: "QuickBarTrapTrayUpLit", tooltip: "ToolTipPrevTrap", sourceLine: 1883, id: 3,
	})
	newQuickbarArrow(root, quickbarArrowSpec{
		x: 15, y: 32, width: 10, height: 14, iconX: -55, iconY: -52,
		image: "QuickBarTrapTrayDownLit", tooltip: "ToolTipNextTrap", sourceLine: 1892, id: 4,
	})
}

func createQuickbarLeft(x, y C.int, warrior bool) {
	root := nox_window_new(nil, 1032, x-1, y+26, 61, 48, nil)
	storeQuickbarControl(&quickbarControls.left, root)
	nox_xxx_wndSetOffsetMB_46AE40(root, 1, -26)
	icon := "QuickBarSpellSetBase"
	if warrior {
		icon = "QuickBarWarriorLeft"
	}
	quickbarIcons(root, icon, icon)
	if !warrior {
		root.field_92 = 5
	}
	nox_window_set_all_funcs(root, unsafe.Pointer(C.nox_xxx_quickbar_45F8D0),
		unsafe.Pointer(C.nox_xxx_quickbarTrapUpDownDraw_45F6F0), nil)
	bookBase := nox_window_new(root, 1160, 0, 9, 29, 30, nil)
	storeQuickbarControl(&quickbarControls.book_button, bookBase)
	quickbarIcons(bookBase, "SpellbookButton", "SpellbookButtonLit")
	book := nox_window_new(bookBase, 1064, 1, 2, 28, 28, nil)
	nox_window_set_all_funcs(book, unsafe.Pointer(C.nox_xxx_quickbarButtonBookWnd_45F450),
		unsafe.Pointer(C.nox_xxx_quickbarButtonBookDraw_45EF30), unsafe.Pointer(C.nox_xxx_quickbarButtonBook_45F3F0))
	quickbarTooltip(book, "OpenSpellBookTT", 1931)
	if warrior {
		return
	}
	for _, spec := range [...]quickbarArrowSpec{
		{x: 30, y: 0, width: 15, height: 19, iconX: -29, iconY: -26,
			image: "QuickBarSpellSetUpLit", tooltip: "ToolTipPrevSpellSet", sourceLine: 1943, id: 0},
		{x: 30, y: 29, width: 15, height: 19, iconX: -29, iconY: -55,
			image: "QuickBarSpellSetDownLit", tooltip: "ToolTipNextSpellSet", sourceLine: 1953, id: 1},
		{x: 48, y: 16, width: 13, height: 17, iconX: -47, iconY: -42,
			image: "QuickBarSpellSetMaxLit", tooltip: "ToolTipAllSpellSets", sourceLine: 1963, id: 2},
	} {
		newQuickbarArrow(root, spec)
	}
	drag := nox_window_new(nil, 1032, 0, 0, 1, 1, nil)
	storeQuickbarControl(&quickbarControls.drag_window, drag)
	nox_window_set_all_funcs(drag, unsafe.Pointer(C.sub_45EF40), unsafe.Pointer(C.sub_45F8F0), nil)
	title := nox_window_new(nil, 1152, x, y, 2, 2, nil)
	storeQuickbarControl(&quickbarControls.title, title)
	nox_xxx_wndSetIcon_46AE60(title, quickbarImage("QuickBarTitle"))
	label := nox_window_new(title, 8, 115, 6, 101, 14, nil)
	nox_window_set_all_funcs(label, nil, unsafe.Pointer(C.sub_45F9B0), nil)
}

func createQuickbarNuggets(warrior bool) bool {
	prefix, suffix := "QuickBarNugget", memmap.String(0x587000, 134996)
	if warrior {
		prefix, suffix = "QuickBarWarriorNugget", ""
	}
	for rowIndex := range 5 {
		bar := g2qExpanded(rowIndex)
		root := nox_quickbar_root(bar)
		if root == nil {
			return false
		}
		x, y := C.int(root.off_x)+10, C.int(root.off_y)+5
		offset := C.int(0)
		for slot := range 5 {
			index := C.int(0)
			if rowIndex == 4 {
				index = C.int(slot)
			}
			win := nox_window_new(nil, 1160, x+offset, y, 30, 10, nil)
			image := fmt.Sprintf("%s%d", prefix, index+1)
			quickbarIcons(win, image, image+suffix)
			nox_xxx_wndSetOffsetMB_46AE40(win, -70-offset, -23)
			nox_window_set_all_funcs(win, unsafe.Pointer(C.nox_xxx_quickbar_45F8D0), nil, nil)
			nox_quickbar_set_nugget(bar, C.int(slot), win)
			nox_xxx_updateSpellIconDir_45DE10(index, bar)
			if !warrior {
				child := nox_window_new(win, 1032, 12, 0, 10, 10, nil)
				nox_window_set_all_funcs(child, unsafe.Pointer(C.sub_45F520),
					unsafe.Pointer(C.nox_xxx_quickbarButtonBookDraw_45EF30), unsafe.Pointer(C.sub_45F480))
				child.field_92 = C.uint(index) | C.uint(rowIndex)<<16
				nox_quickbar_set_nugget_child(bar, C.int(slot), child)
			}
			if rowIndex == 4 {
				key := nox_window_new(win, 1032, 13, 39, 10, 10, nil)
				nox_window_set_all_funcs(key, nil, unsafe.Pointer(C.sub_45F5D0), nil)
				key.field_92 = C.uint(index)
			} else {
				nox_window_set_hidden(win, 1)
			}
			offset += C.int(memmap.Uint32(0x587000, 133488+4*uintptr(slot)))
		}
	}
	return true
}

//export nox_xxx_quickBarCreate_45E190
func nox_xxx_quickBarCreate_45E190() C.int {
	*memmap.PtrUint32(0x5D4594, 1047916) = 0
	*memmap.PtrUint8(0x5D4594, 1047920) = 0
	sub_416170(5)
	*memmap.PtrUint32(0x5D4594, 1047924) = 0
	*memmap.PtrUint32(0x5D4594, 1047928) = 0
	*memmap.PtrUint32(0x5D4594, 1049480) = 0
	*memmap.PtrUint8(0x5D4594, 1049488) = 0
	fontHeight := C.int(nox_xxx_guiFontHeightMB_43F320(nil))
	x, y := (C.nox_win_width-320)/2, C.nox_win_height-74
	C.dword_5d4594_1047548, C.dword_5d4594_1047552 = C.uint32_t(x), C.uint32_t(y)
	C.nox_quickbar_font = nox_xxx_guiFontPtrByName_43F360(internCStr("small"))
	if unsafe.Sizeof(uintptr(0)) == 4 {
		*memmap.PtrUint32(0x5D4594, 1049684) = uint32(uintptr(C.nox_quickbar_font))
	}
	*memmap.PtrUint32(0x587000, 133656) = uint32(x)
	*memmap.PtrUint32(0x587000, 133660) = uint32(y - 17)
	*memmap.PtrUint32(0x587000, 133664) = uint32(x + 320)
	*memmap.PtrUint32(0x587000, 133668) = uint32(C.nox_win_height)
	player := g2qPlayer()
	if player == nil {
		return 0
	}
	class := asPlayerS(player).PlayerClass()
	warrior := class == playerlib.Warrior
	draw := unsafe.Pointer(C.nox_xxx_quickBarWarriorDraw_45FDE0)
	if warrior {
		draw = unsafe.Pointer(C.nox_xxx_quickBarDrawFn_45FBD0)
	}
	proc := unsafe.Pointer(C.nox_xxx_quickBarWnd_45EF50)
	nox_xxx_quickBarInitWindow_4601F0(g2qMain(), x+69, y+32, 5, 0, proc, draw)
	for row := 3; row >= 0; row-- {
		bar := g2qExpanded(row)
		nox_xxx_quickBarInitWindow_4601F0(bar, x+69, y+32-C.int(60*(4-row)), 5, 0,
			proc, unsafe.Pointer(C.nox_xxx_quickBarWarriorDraw_45FDE0))
		nox_window_set_hidden(nox_quickbar_root(bar), 1)
		*g2qB(bar, 200) = 0
		g2qMirror(bar)
	}
	createQuickbarRight(x, y, player, class)
	if !warrior {
		createQuickbarTrapTray(x, y, fontHeight)
	}
	createQuickbarLeft(x, y, warrior)
	if !createQuickbarNuggets(warrior) {
		return 0
	}
	if warrior {
		for i := uintptr(0); i < 5; i++ {
			dst := memmap.Slice(0x5D4594, 1047788+24*i)[:24]
			copy(dst[:20], memmap.Slice(0x587000, 133536+24*i)[:20])
			clear(dst[20:])
		}
	}
	nox_xxx_clientUpdateButtonRow_45E110(0)
	return 1
}
