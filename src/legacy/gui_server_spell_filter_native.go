package legacy

/*
#include "game2_server_bridge.h"
#include "client__gui__servopts__spelllst.h"
#include "cgo_const.h"

*/
import "C"

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

const serverSpellFilterStrings = `C:\NoxPost\src\client\Gui\ServOpts\spelllst.c`

//export nox_xxx_guiSpelllistLoad_453850
func nox_xxx_guiSpelllistLoad_453850(parent *C.nox_window) *C.nox_window {
	root := nox_new_window_from_file(internCStr("spelllst.wnd"), unsafe.Pointer(C.sub_453C00))
	serverSharedWindows.spell_filter = root
	if root == nil {
		return nil
	}
	nox_xxx_wndSetDrawFn_46B340(root, unsafe.Pointer(C.sub_453B80))
	C.sub_46B120(root, parent)
	nox_xxx_wnd_46B280(root, parent)
	names, classes := g2ServerChild(root, 1110), g2ServerChild(root, 1112)
	serverSharedWindows.spell_names, serverSharedWindows.spell_classes = names, classes
	if names == nil || classes == nil {
		nox_xxx_windowDestroyMB_46C4E0(root)
		serverSharedWindows.spell_filter, serverSharedWindows.spell_names, serverSharedWindows.spell_classes = nil, nil, nil
		return nil
	}
	sub_453B00()
	g2ServerEvent(names, 16399, 0, 0)
	g2ServerEvent(classes, 16399, 0, 0)
	var text [64]wchar2_t
	for id := int32(1); id < int32(C.NOX_SPELLS_MAX); id++ {
		if !bool(nox_xxx_spellIsValid_424B50(id)) {
			continue
		}
		flags := uint32(nox_xxx_spellFlags_424A70(id))
		if flags&0x15000 != 0 || flags&0x7000000 == 0 {
			continue
		}
		var label *wchar2_t
		switch {
		case flags&0x1000000 != 0 || flags&0x6000000 == 0x6000000:
			label = g2ServerLoad("Common", serverSpellFilterStrings, 307)
		case flags&0x2000000 != 0:
			label = g2ServerLoad("SpellWizard", serverSpellFilterStrings, 314)
		default:
			label = g2ServerLoad("SpellConjurer", serverSpellFilterStrings, 318)
		}
		text[0] = 0
		appendWideString(&text[0], label)
		g2ServerText(classes, 16397, unsafe.Pointer(&text[0]), ^uintptr(0))
		nox_wcsncpy(&text[0], nox_xxx_spellTitle_424930(id), 63)
		text[63] = 0
		g2ServerText(names, 16397, unsafe.Pointer(&text[0]), ^uintptr(0))
	}
	up := unsafe.Pointer(g2ServerChild(root, 1113))
	g2ServerText(names, 16408, up, 0)
	g2ServerText(classes, 16408, up, 0)
	down := unsafe.Pointer(g2ServerChild(root, 1114))
	g2ServerText(names, 16409, down, 0)
	g2ServerText(classes, 16409, down, 0)
	sub_454040((*C.uint32_t)(g2ServerMem(1045488)))
	sub_454120()
	if !g2ServerGame(1) || g2ServerGame(49152) {
		sub_46AD20(root, 1115, 1133, 0)
	}
	return serverSharedWindows.spell_filter
}

func serverSpellFilterScroll(names, classes, control *C.nox_window) {
	g2ServerText(names, 0x4000, unsafe.Pointer(control), 0)
	g2ServerText(classes, 0x4000, unsafe.Pointer(control), 0)
	sub_454120()
}

func serverSpellUnavailableInFlagBall(settings *C.char, spell C.int) bool {
	return spell == 132 && (noxflags.HasGame(noxflags.GameModeFlagBall) ||
		settings != nil && *(*byte)(unsafe.Add(unsafe.Pointer(settings), 52))&0x40 != 0)
}

func setAllServerSpellFilters(names *C.nox_window, set C.int) {
	data := g2ServerList(names)
	settings := sub_4165B0()
	if items := g2ServerItems(data); items != nil {
		for i := 0; i < int(data.Field_11_0); i++ {
			title := (*wchar2_t)(unsafe.Pointer(&items[i].Text[0]))
			if *title == 0 {
				continue
			}
			spell := C.int(nox_xxx_spellByTitle_424960(title))
			if spell != 0 && (set == 0 || !serverSpellUnavailableInFlagBall(settings, spell)) {
				sub_453FA0(g2ServerMem(1045488), spell, set)
			}
		}
	}
	if Get_dword_5d4594_2650652() != 0 {
		var online [15]C.int
		sub_57A1E0(&online[0], nil, nil, 4, 6128)
		for i := 0; i < 5; i++ {
			*g2ServerWord(1045488 + 4*uintptr(i)) &= uint32(online[i+6])
		}
	}
	sub_454120()
	sub_459D50(1)
}

func serverSpellFilterNotice(root, control *C.nox_window, key string, line int) {
	control.draw_data.field_0 ^= 4
	message := g2ServerLoad(key, serverSpellFilterStrings, line)
	caption := g2ServerLoad("Notice", serverSpellFilterStrings, line-1)
	nox_xxx_dialogMsgBoxCreate_449A10(root, caption, message, 33, nil, nil)
	sub_44A360(1)
}

func setServerSpellFilter(root, names, control *C.nox_window, row int) {
	data := g2ServerList(names)
	index := g2ServerFirstVisible(data) + row
	items := g2ServerItems(data)
	if data == nil || items == nil || index < 0 || index >= int(data.Field_11_0) || items[index].Text[0] == 0 {
		sub_459D50(1)
		return
	}
	spell := C.int(nox_xxx_spellByTitle_424960((*wchar2_t)(unsafe.Pointer(&items[index].Text[0]))))
	if spell == 0 {
		sub_459D50(1)
		return
	}
	if Get_dword_5d4594_2650652() != 0 {
		var online [15]C.int
		sub_57A1E0(&online[0], nil, nil, 4, 6128)
		if sub_454000((*C.nox_cgo_const_void)(unsafe.Pointer(&online[6])), spell) == 0 {
			serverSpellFilterNotice(root, control, "NotInternet", 211)
			return
		}
	}
	if serverSpellUnavailableInFlagBall(sub_4165B0(), spell) {
		serverSpellFilterNotice(root, control, "plyrspel.c:Illegal", 226)
		return
	}
	sub_453FA0(g2ServerMem(1045488), spell, g2ServerBool(control.draw_data.field_0&4 == 0))
	sub_459D50(1)
}

//export sub_453C00
func sub_453C00(window *C.nox_window, event C.int, control *C.nox_window, arg C.uintptr_t) C.int {
	root, names, classes := serverSharedWindows.spell_filter, serverSharedWindows.spell_names, serverSharedWindows.spell_classes
	if root == nil || names == nil || classes == nil || control == nil {
		return 0
	}
	switch event {
	case 0x4000:
		if control == g2ServerChild(root, 1113) || control == g2ServerChild(root, 1114) {
			serverSpellFilterScroll(names, classes, control)
		}
	case 16391:
		id := int(control.id)
		switch id {
		case 1113, 1114:
			serverSpellFilterScroll(names, classes, control)
			return 0
		case 1115, 1116:
			setAllServerSpellFilters(names, g2ServerBool(id == 1115))
		default:
			if id >= 1120 && id <= 1133 {
				setServerSpellFilter(root, names, control, id-1120)
			}
		}
		nox_xxx_clientPlaySoundSpecial_452D80(766, 100)
	}
	return 0
}
