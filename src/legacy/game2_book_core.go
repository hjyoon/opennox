package legacy

/*
#include "game2_book_native.h"
*/
import "C"

import (
	"encoding/binary"
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

// Cgo omits the packed info member; use the existing native Player accessor.
func game2PlayerClass(p *C.nox_playerInfo) C.uchar {
	return C.uchar(asPlayerS(p).PlayerClass())
}

func game2BookU(off uintptr) *uint32  { return memmap.PtrUint32(0x5D4594, off) }
func game2BookID(index uint32) uint32 { return *game2BookU(1046960 + 4*uintptr(index)) }
func game2BookClass() C.int           { return C.int(game2PlayerClass(C.game2_book_player())) }
func game2BookGameFlag(flag C.int) bool {
	return bool(C.nox_common_gameFlags_check_40A5C0(C.uint(flag)))
}
func game2BookSpellFlag(id, flag C.int) bool            { return bool(C.nox_xxx_spellHasFlags_424A50(id, flag)) }
func game2BookSound(id C.int)                           { C.nox_xxx_clientPlaySoundSpecial_452D80(id, 100) }
func game2BookHide(w *C.nox_window, hidden C.int) C.int { return C.nox_window_set_hidden(w, hidden) }
func game2BookPtr32(p unsafe.Pointer, i int) *uint32    { return (*uint32)(unsafe.Add(p, i*4)) }
func game2BookAnimChoose(ref *C.nox_things_imageRef_t, guide bool) {
	fn := unsafe.Pointer(C.nox_xxx_bookClickSpell_45B1F0)
	if guide {
		fn = unsafe.Pointer(C.nox_xxx_bookClickCreature_45B200)
	}
	nox_book_anim_set_end(ref, fn)
}

//export nox_book_anim_data
func nox_book_anim_data(ref *C.nox_things_imageRef_t) *C.nox_image_ref_anim_t {
	if ref == nil {
		return nil
	}
	return (*C.nox_image_ref_anim_t)(ref.field_24)
}

//export nox_book_anim_set_end
func nox_book_anim_set_end(ref *C.nox_things_imageRef_t, fn unsafe.Pointer) {
	if anim := nox_book_anim_data(ref); anim != nil {
		anim.on_end = fn
	}
}

//export nox_book_anim_start
func nox_book_anim_start(ref *C.nox_things_imageRef_t) {
	if anim := nox_book_anim_data(ref); anim != nil {
		anim.field_3 = C.uint32_t(C.nox_xxx_bookGet_430B40_get_mouse_prev_seq())
	}
}

//export nox_client_localPlayerHasFlags30
func nox_client_localPlayerHasFlags30(flags C.uint32_t) C.int {
	dr := (*C.nox_drawable)(*memmap.PtrPtr(0x852978, 8))
	if dr != nil && dr.flags30&flags == flags {
		return 1
	}
	return 0
}

//export game2_book_compare
func game2_book_compare(a, b unsafe.Pointer) C.int {
	id2 := C.int(*(*uint32)(b)) // The original reads argument 2 before argument 1.
	var s1, s2 *C.wchar2_t
	if C.dword_5d4594_1046868 == 1 {
		s1 = C.nox_xxx_guiCreatureGetName_427240(C.int(*(*uint32)(a)))
		s2 = C.nox_xxx_guiCreatureGetName_427240(id2)
	} else {
		id1 := C.int(*(*uint32)(a))
		if game2PlayerClass(C.game2_book_player()) != 0 {
			s1 = C.nox_xxx_spellTitle_424930(id1)
			s2 = C.nox_xxx_spellTitle_424930(id2)
		} else {
			s1 = C.nox_xxx_abilityGetName_0_425260(id1)
			s2 = C.nox_xxx_abilityGetName_0_425260(id2)
		}
	}
	if s1 != nil && s2 != nil {
		return C._nox_wcsicmp(s1, s2)
	}
	return 0
}

//export nox_xxx_bookSetColor_45AC40
func nox_xxx_bookSetColor_45AC40() C.int {
	*game2BookU(1046880) = uint32(C.nox_color_rgb_4344A0(15, 15, 15))
	value := C.nox_color_rgb_4344A0(115, 100, 100)
	*game2BookU(1046884) = uint32(value)
	return C.int(value)
}

//export nox_client_toggleSpellbook_45AC70
func nox_client_toggleSpellbook_45AC70() {
	if C.wndIsShown_nox_xxx_wndIsShown_46ACC0(C.nox_win_unk1) != 0 {
		C.nox_xxx_bookShowMB_45AD70(0)
	} else {
		nox_xxx_bookHideMB_45ACA0(0)
	}
}

//export nox_xxx_bookHideMB_45ACA0
func nox_xxx_bookHideMB_45ACA0(animate C.int) C.int {
	if C.wndIsShown_nox_xxx_wndIsShown_46ACC0(C.nox_win_unk1) != 0 || C.dword_5d4594_1047520 == 1 {
		return 0
	}
	C.dword_5d4594_1046864 = 0
	C.set_dword_5d4594_3799468(1)
	C.dword_5d4594_1046868 = 0
	if C.dword_5d4594_1046872 != 0 {
		C.dword_5d4594_1046868 = 1
	}
	if C.nox_xxx_wndGetCaptureMain_46AE00() == C.nox_win_unk1 {
		C.nox_xxx_wndClearCaptureMain_46ADE0(C.nox_win_unk1)
	}
	game2BookHide(C.nox_win_unk1, 1)
	game2BookSound(787)
	if animate != 0 {
		C.nox_xxx_aNox_cfg_0_587000_132132 = 1
		C.dword_5d4594_1046936 = 0
		game2BookHide(C.dword_5d4594_1046952, 1)
	}
	if C.nox_xxx_bookGetSpellDnDType_477670() == 1 {
		C.nox_xxx_bookSpellDnDclear_477660()
	}
	return 1
}

//export nox_xxx_guiSpellSortList_45ADF0
func nox_xxx_guiSpellSortList_45ADF0(class C.int) C.int {
	player := C.game2_book_player()
	found := C.int(0)
	C.dword_5d4594_1046656 = C.uint32_t(C.nox_xxx_guiFontHeightMB_43F320(nil) + 2)
	*game2BookU(1047508) = 0
	C.dword_5d4594_1047512 = 0
	perPage := uint32(2*(uint32(141)/uint32(C.dword_5d4594_1046656)) - 2)
	appendID := func(id C.int) {
		found = 1
		n := *game2BookU(1047508)
		*game2BookU(1046960 + 4*uintptr(n)) = uint32(id)
		*game2BookU(1047508) = n + 1
	}
	known := func(id C.int, guide bool) bool {
		if game2BookGameFlag(0x2000) && !game2BookGameFlag(4096) {
			return true
		}
		if guide {
			return player.beast_scroll_lvl[int(id)] != 0
		}
		return player.spell_lvl[int(id)] != 0
	}
	if C.dword_5d4594_1046868 == 1 {
		for id := C.nox_xxx_bookGetFirstCreMB_427300(); id != 0; id = C.nox_xxx_bookGetNextCre_427320(id) {
			if known(id, true) && C.nox_xxx_bookCreatureTest_4D70C0(id) != 0 {
				appendID(id)
			}
		}
	} else if class != 0 {
		for id := C.nox_xxx_spellFirstValid_424AD0(); id != 0; id = C.nox_xxx_spellNextValid_424AF0(id) {
			if id != 34 && C.nox_xxx_playerCheckSpellClass_57AEA0(class, id) == 0 && known(id, false) {
				if game2BookSpellFlag(id, 0x15000) {
					C.dword_5d4594_1047512++
				}
				if !game2BookSpellFlag(id, 0x2000) {
					appendID(id)
				}
			}
		}
	} else {
		for id := C.nox_xxx_bookFirstKnownAbil_425330(); id != 0; id = C.nox_xxx_bookNextKnownAbil_425350(id) {
			if known(id, false) {
				appendID(id)
			}
		}
	}
	n := *game2BookU(1047508) - uint32(C.dword_5d4594_1047512)
	*game2BookU(1046940) = n/perPage + 1
	// qsort is the same external host sorting service, with a native callback.
	C.qsort(memmap.PtrOff(0x5D4594, 1046960), C.size_t(n), 4, (*[0]byte)(unsafe.Pointer(C.nox_xxx_guiSpellSortFn_45ABC0)))
	return found
}

//export nox_xxx_book_45B010
func nox_xxx_book_45B010(force C.int) {
	if C.nox_gui_xxx_check_446360() != 0 || C.sub_4AE3D0() != 0 {
		return
	}
	if force != 0 || C.nox_xxx_get_57AF20() == 0 {
		C.dword_5d4594_1046864 = 1
		C.dword_5d4594_1046868 = 0
		if C.dword_5d4594_1046872 != 0 {
			C.dword_5d4594_1046868 = 1
		}
		C.nox_xxx_wndShowModalMB_46A8C0(C.nox_win_unk1)
		game2BookSound(786)
	}
}

//export nox_xxx_bookClickSpell_45B1F0
func nox_xxx_bookClickSpell_45B1F0() C.int {
	C.dword_5d4594_1046868 = 0
	C.dword_5d4594_1046872 = 0
	return 0
}

//export nox_xxx_bookClickCreature_45B200
func nox_xxx_bookClickCreature_45B200() C.int {
	C.dword_5d4594_1046868 = 1
	C.dword_5d4594_1046872 = 1
	return 1
}

//export sub_45CFC0
func sub_45CFC0() C.int {
	return C.int((^uint8(C.nox_xxx_wndGetFlags_46ADA0(C.nox_win_unk1)) >> 4) & 1)
}

//export sub_45D9B0
func sub_45D9B0() C.int { return C.int(C.dword_5d4594_1047520) }

//export nox_xxx_guiSpellTargetClickSet_45D9D0
func nox_xxx_guiSpellTargetClickSet_45D9D0(id C.int) C.int {
	if *game2BookU(1047916) != 0 {
		return 0
	}
	*game2BookU(1047916) = uint32(id)
	*memmap.PtrUint8(0x5D4594, 1047920) = 0
	C.nox_xxx_setKeybTimeout_4160D0(5)
	*game2BookU(1047928) = 0
	*game2BookU(1047924) = 0
	return 1
}

//export nox_xxx_guiSpell_45DA10
func nox_xxx_guiSpell_45DA10(id C.int) C.int {
	if *game2BookU(1047916) != 0 {
		return 0
	}
	*game2BookU(1047916) = uint32(id)
	*memmap.PtrUint8(0x5D4594, 1047920) = 0
	C.nox_xxx_setKeybTimeout_4160D0(5)
	*game2BookU(1047924) = 1
	return 1
}

//export nox_client_invokeSpellSlot_45DA50
func nox_client_invokeSpellSlot_45DA50(index C.int) {
	player := C.game2_book_current_player()
	if player == nil || C.nox_xxx_playerAnimCheck_4372B0() != 0 {
		return
	}
	slot := nox_quickbar_selected_slot(C.nox_xxx_aClosewoodengat_587000_133480, index)
	if slot == nil {
		return
	}
	if game2PlayerClass(player) != 0 {
		nox_xxx_clientSendSpell_45DB20((*C.char)(unsafe.Pointer(slot)), 1, C.char(*(*byte)(unsafe.Add(unsafe.Pointer(slot), 4))&1))
	} else {
		nox_xxx_clientSendAbil_45DAF0(C.int(*(*uint32)(unsafe.Pointer(slot))))
	}
	nox_xxx_clientStoreLastButton_45DAD0(index)
}

//export nox_xxx_clientStoreLastButton_45DAD0
func nox_xxx_clientStoreLastButton_45DAD0(index C.int) {
	*memmap.PtrUint32(0x587000, 133484) = uint32(index)
	*game2BookU(1049540) = uint32(C.gameFrame())
}

// The transport hooks are synchronous: packet storage is never retained by C.
var game2BookCursor = func() int32 { return int32(C.nox_xxx_guiCursor_477600()) }
var game2BookSend = func(packet []byte) int32 {
	result := int32(C.nox_netlist_addToMsgListCli_40EBC0(31, 0, (*C.uchar)(unsafe.Pointer(&packet[0])), C.int(len(packet))))
	runtime.KeepAlive(packet)
	return result
}

//export nox_xxx_clientSendAbil_45DAF0
func nox_xxx_clientSendAbil_45DAF0(id C.int) C.int {
	result := game2BookCursor()
	if result != 0 {
		return C.int(result)
	}
	if id == 0 {
		return 0
	}
	return C.int(game2BookSend([]byte{122, byte(id)}))
}

//export nox_xxx_clientSendSpell_45DB20
func nox_xxx_clientSendSpell_45DB20(ids *C.char, count C.int, flag C.char) C.int {
	result := game2BookCursor()
	if result != 0 {
		return C.int(result)
	}
	p := unsafe.Pointer(ids)
	if *(*uint32)(p) == 0 {
		return 0
	}
	packet := make([]byte, 22)
	packet[0] = 121
	for i := 0; i < 5; i++ {
		if i < int(count) {
			binary.LittleEndian.PutUint32(packet[1+4*i:], *game2BookPtr32(p, i))
		}
	}
	packet[21] = byte(flag)
	return C.int(game2BookSend(packet))
}

//export sub_45DB90
func sub_45DB90() C.int {
	clear(unsafe.Slice((*byte)(memmap.PtrOff(0x5D4594, 1049544)), 136))
	*memmap.PtrUint8(0x5D4594, 1049680) = 0
	return 0
}

//export nox_xxx_guiSpellTargetClickCheckSend_45DBB0
func nox_xxx_guiSpellTargetClickCheckSend_45DBB0() {
	if *game2BookU(1047916) != 0 {
		nox_xxx_clientSendSpell_45DB20((*C.char)(memmap.PtrOff(0x5D4594, 1047916)), 1, C.char(*memmap.PtrUint8(0x5D4594, 1047924)))
		*game2BookU(1047916) = 0
	}
}

//export nox_quickbar_selected_row
func nox_quickbar_selected_row(base unsafe.Pointer) *C.uchar {
	if base == nil {
		return nil
	}
	row := *(*byte)(unsafe.Add(base, 200))
	if row >= 5 {
		return nil
	}
	return (*C.uchar)(unsafe.Add(base, 40*uintptr(row)))
}

//export nox_quickbar_trap_selected_row
func nox_quickbar_trap_selected_row() *C.uchar {
	row := *memmap.PtrUint8(0x5D4594, 1048140)
	if row >= 3 {
		return nil
	}
	return (*C.uchar)(memmap.PtrOff(0x5D4594, 1047940+40*uintptr(row)))
}

//export nox_quickbar_selected_slot
func nox_quickbar_selected_slot(base unsafe.Pointer, slot C.int) *C.uchar {
	if slot < 0 || slot >= 5 {
		return nil
	}
	row := nox_quickbar_selected_row(base)
	if row == nil {
		return nil
	}
	return (*C.uchar)(unsafe.Add(unsafe.Pointer(row), 8*int(slot)))
}
