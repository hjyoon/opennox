package legacy

/*
#include "game2_book_native.h"
*/
import "C"

import (
	"github.com/opennox/opennox/v1/common/memmap"
	"unsafe"
)

func game2BookRewardMask(id C.int) C.int {
	if game2BookSpellFlag(id, 0x1000) {
		return 0x2000
	}
	if game2BookSpellFlag(id, 0x4000) {
		return 0x8000
	}
	if game2BookSpellFlag(id, 0x10000) {
		return 0x20000
	}
	return 0
}
func game2BookFind(id C.int, count int) int {
	for i := 0; i < count; i++ {
		if game2BookID(uint32(i)) == uint32(id) {
			return i
		}
	}
	return count
}
func game2BookNotify(id C.int, page int) {
	nox_xxx_bookHideMB_45ACA0(0)
	nox_xxx_bookMoveToPage_45B930(C.int(page))
	nox_xxx_book_45B010(0)
}

//export nox_xxx_netSpellRewardCli_45CFE0
func nox_xxx_netSpellRewardCli_45CFE0(id, level, notify, extra C.int) {
	player := C.game2_book_current_player()
	if player == nil {
		return
	}
	if C.nox_xxx_playerCheckSpellClass_57AEA0(C.int(game2PlayerClass(player)), id) == 9 {
		return
	}
	player.spell_lvl[int(id)] = C.uint32_t(level)
	if mask := game2BookRewardMask(id); mask != 0 {
		for spell := C.int(1); spell < 137; spell++ {
			if game2BookSpellFlag(spell, mask) && bool(C.nox_xxx_spellIsValid_424B50(spell)) {
				player.spell_lvl[int(spell)] = C.uint32_t(level)
			}
		}
	}
	C.dword_5d4594_1046868 = 0
	C.dword_5d4594_1046872 = 0
	nox_xxx_guiSpellSortList_45ADF0(C.int(game2PlayerClass(player)))
	if id == 34 {
		show := C.int(0)
		if player.spell_lvl[34] == 1 && notify != 0 {
			show = 1
		}
		C.nox_xxx_quickbarAddTrap_460EC0(show)
	} else if notify != 0 {
		if page := game2BookFind(id, 137); page != 137 {
			game2BookNotify(id, page)
			C.game2_book_reward_scalar(2, id, extra)
		}
	}
}

//export game2_book_forward_word
func game2_book_forward_word(kind C.uintptr_t, id C.int, position *C.int2) C.uintptr_t {
	C.nox_xxx_bookShowMB_45AD70(1)
	C.nox_window_setPos_46A9B0(C.nox_win_unk1, position.field_0, position.field_4)
	token := uintptr(kind)
	count := 0
	switch token {
	case 2, 4:
		count = 137
	case 3:
		count = 6
	default:
		return C.uintptr_t(token)
	}
	page := game2BookFind(id, count)
	if page != count {
		return C.uintptr_t(uintptr(int64(nox_xxx_bookMoveToPage_45B930(C.int(page)))))
	}
	return C.uintptr_t(uintptr(memmap.PtrOff(0x5D4594, 1046960+4*uintptr(count))))
}

//export game2_book_ability_reward
func game2_book_ability_reward(id C.int, notify C.uintptr_t, extra C.int) {
	player := C.nox_common_playerInfoGetByID_417040(C.int(C.nox_player_netCode_85319C))
	if player == nil {
		return
	}
	C.dword_5d4594_1046868 = 0
	C.dword_5d4594_1046872 = 0
	nox_xxx_guiSpellSortList_45ADF0(C.int(game2PlayerClass(player)))
	if notify != 0 {
		if page := game2BookFind(id, 6); page != 6 {
			game2BookNotify(id, page)
			if extra != 0 {
				C.game2_book_reward_scalar(3, id, extra)
			}
		}
	}
}

//export sub_45D320
func sub_45D320(id C.int) C.int {
	player := C.game2_book_current_player()
	result := nox_xxx_bookHideMB_45ACA0(1)
	if player != nil {
		player.spell_lvl[int(id)] = 0
		C.sub_461360(id)
		if mask := game2BookRewardMask(id); mask != 0 {
			for spell := C.int(1); spell < 137; spell++ {
				if game2BookSpellFlag(spell, mask) && bool(C.nox_xxx_spellIsValid_424B50(spell)) {
					player.spell_lvl[int(spell)] = 0
					C.sub_461360(spell)
				}
			}
		}
		return nox_xxx_guiSpellSortList_45ADF0(C.int(game2PlayerClass(player)))
	}
	return result
}

//export sub_45D400
func sub_45D400(id C.int) C.int {
	player := C.game2_book_current_player()
	result := nox_xxx_bookHideMB_45ACA0(1)
	if player != nil {
		player.beast_scroll_lvl[int(id)] = 0
		C.sub_461360(id + 74)
		for groupIndex := uintptr(0); groupIndex < 41; groupIndex++ {
			group := *memmap.PtrPtr(0x587000, 132124+4*groupIndex)
			if group == nil {
				break
			}
			if *(*uint32)(group) == uint32(id) {
				for member := unsafe.Add(group, 4); *(*uint32)(member) != 0; member = unsafe.Add(member, 4) {
					spell := *(*uint32)(member)
					player.beast_scroll_lvl[int(spell)] = 0
					C.sub_461360(C.int(spell) + 74)
				}
			}
		}
		result = nox_xxx_guiSpellSortList_45ADF0(C.int(game2PlayerClass(player)))
	}
	return result
}

//export game2_book_disable_ability_word
func game2_book_disable_ability_word(id C.int) C.uintptr_t {
	nox_xxx_bookHideMB_45ACA0(1)
	C.nox_xxx_netAbilityRewardCli_4611E0(id, 0, nil)
	if player := C.game2_book_player(); player != nil {
		player.spell_lvl[int(id)] = 0
	}
	C.sub_461360(id)
	player := C.nox_common_playerInfoGetByID_417040(C.int(C.nox_player_netCode_85319C))
	if player == nil {
		return 0
	}
	return C.uintptr_t(uintptr(int64(nox_xxx_guiSpellSortList_45ADF0(C.int(game2PlayerClass(player))))))
}

//export sub_45D500
func sub_45D500(show C.int) C.int {
	result := C.int(C.dword_5d4594_1046864)
	if show != 0 {
		if C.dword_5d4594_1046864 != 0 {
			result = game2BookHide(C.nox_win_unk1, 0)
		}
	} else if C.dword_5d4594_1046864 != 0 {
		result = C.sub_47A260()
		if result == 0 {
			result = game2BookHide(C.nox_win_unk1, 1)
		}
	}
	return result
}

//export game2_book_position_word
func game2_book_position_word(position *C.uint32_t) C.uintptr_t {
	if position == nil {
		return 0
	}
	result := C.nox_client_wndGetPosition_46AA60(C.dword_5d4594_1046952, (*C.uint)(unsafe.Pointer(position)), (*C.uint)(unsafe.Add(unsafe.Pointer(position), 4)))
	return C.uintptr_t(uintptr(int64(result)))
}

//export nox_xxx_bookFillAll_45D570
func nox_xxx_bookFillAll_45D570(kind, id C.int) {
	if C.dword_5d4594_1047520 == 1 {
		return
	}
	C.nox_xxx_quickBarClose_4606B0()
	*game2BookU(1046612) = uint32(C.nox_xxx_buttonsGetSelectedRow_45E180())
	if kind == 2 && game2BookSpellFlag(id, 0x15000) {
		return
	}
	if C.nox_xxx_buttonHaveSpellInBarMB_4612D0(id) == 1 {
		return
	}
	var slot C.int
	switch kind {
	case 2, 4:
		slot = C.nox_xxx_buttonFindFirstEmptySlot_461250()
	case 3:
		slot = C.sub_4612A0()
	default:
		return
	}
	C.dword_5d4594_1046852 = C.uint32_t(slot)
	if slot == -1 {
		return
	}
	C.dword_5d4594_1047520 = 1
	C.dword_5d4594_1046868 = 1
	C.dword_5d4594_1046872 = 1
	if kind == 2 || kind == 3 {
		C.dword_5d4594_1046868 = 0
		C.dword_5d4594_1046872 = 0
	}
	game2_book_position_word((*C.uint32_t)(memmap.PtrOff(0x5D4594, 1046844)))
	*(*float32)(unsafe.Pointer(&C.dword_5d4594_1046636)) = float32(int32(*game2BookU(1046844)))
	*(*float32)(unsafe.Pointer(&C.dword_5d4594_1046640)) = float32(int32(*game2BookU(1046848)))
	C.nox_window_setPos_46A9B0(C.dword_5d4594_1046956, C.int(*game2BookU(1046844)), C.int(*game2BookU(1046848)))
	C.nox_xxx_buttonSetImgMB_461320(slot, (*C.uint32_t)(memmap.PtrOff(0x5D4594, 1046668)))
	first := C.int2{field_0: 400, field_4: -500}
	second := C.int2{field_0: 350, field_4: 400}
	C.dword_5d4594_1047524 = C.uint32_t(id)
	C.dword_5d4594_1046652 = 0
	if kind == 3 {
		C.dword_5d4594_1046652 = 1
	}
	*game2BookU(1046676) = uint32(kind)
	C.nox_client_setRenderGUI(1)
	*game2BookU(1046680) = 0
	C.sub_4BEDE0((*C.int2)(memmap.PtrOff(0x5D4594, 1046844)), (*C.int2)(memmap.PtrOff(0x5D4594, 1046668)), &first, &second, 19, 0, C.uintptr_t(uintptr(unsafe.Pointer(C.sub_45D7D0))), nil)
	*game2BookU(1046628) = 0
	C.obj_5d4594_1046620.field_0 = C.float(*memmap.PtrFloat32(0x5D4594, 1046692) - *memmap.PtrFloat32(0x5D4594, 1046684))
	C.obj_5d4594_1046620.field_4 = C.float(*memmap.PtrFloat32(0x5D4594, 1046696) - *memmap.PtrFloat32(0x5D4594, 1046688))
	C.nox_xxx_utilNormalizeVector_509F20(&C.obj_5d4594_1046620)
	scale := float64(10)
	if C.nox_win_width < 1000 {
		scale = 8
		if C.nox_win_width < 750 {
			scale = 6
		}
	}
	C.obj_5d4594_1046620.field_0 = C.float(float64(C.obj_5d4594_1046620.field_0) * scale)
	C.obj_5d4594_1046620.field_4 = C.float(float64(C.obj_5d4594_1046620.field_4) * scale)
	game2BookHide(C.dword_5d4594_1046956, 0)
	C.nox_xxx_wndShowModalMB_46A8C0(C.dword_5d4594_1046956)
	if game2BookGameFlag(2048) {
		C.sub_57AF30(nil, kind)
	}
	C.dword_5d4594_1046648 = C.uint32_t(C.nox_xxx_bookGet_430B40_get_mouse_prev_seq())
	if !game2BookGameFlag(2048) || C.nox_gui_xxx_check_446360() == 1 || C.nox_xxx_gameGet_4DB1B0() == 1 {
		C.sub_45D870()
	}
}

//export sub_45D7D0
func sub_45D7D0(a, b *C.int) C.int {
	result := *game2BookU(1046680)
	if int32(result) < 20 {
		result++
		*memmap.PtrFloat32(0x5D4594, 1046676+8*uintptr(result)) = float32(*a)
		*memmap.PtrFloat32(0x5D4594, 1046680+8*uintptr(result)) = float32(*(*C.int)(unsafe.Add(unsafe.Pointer(a), 4)))
		*memmap.PtrFloat32(0x5D4594, 1046684+8*uintptr(result)) = float32(*b)
		last := float32(*(*C.int)(unsafe.Add(unsafe.Pointer(b), 4)))
		*game2BookU(1046680) = result
		*memmap.PtrFloat32(0x5D4594, 1046688+8*uintptr(result)) = last
	}
	return C.int(result)
}

//export sub_45D810
func sub_45D810() {
	if C.dword_5d4594_1047520 != 0 {
		C.dword_5d4594_1047520 = 0
		game2BookHide(C.dword_5d4594_1046956, 1)
		C.dword_5d4594_1046648 = 0
		C.nox_xxx_clientUpdateButtonRow_45E110(C.int(*game2BookU(1046612)))
		if game2BookGameFlag(2048) {
			C.sub_57B0A0()
			C.sub_413A00(0)
		}
	}
}
