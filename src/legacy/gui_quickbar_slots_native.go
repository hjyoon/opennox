package legacy

/*
#include "GAME2_1.h"
#include "client__gui__guispell.h"

extern nox_window *dword_5d4594_1049500, *dword_5d4594_1049504, *dword_5d4594_1049520;
extern nox_window *dword_5d4594_1049508, *dword_5d4594_1049512, *dword_5d4594_1049516, *dword_5d4594_1049532;
extern uint32_t dword_5d4594_1047932;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

type quickbarSpellSlot struct {
	id   uint32
	mode uint8
	_    [3]byte
}

func quickbarSpellRows() (*[5][5]quickbarSpellSlot, int) {
	data := g2qMain()
	if data == nil {
		return nil, 0
	}
	start := int(*(*uint8)(unsafe.Add(data, 200)))
	if start >= 5 {
		return nil, 0
	}
	return (*[5][5]quickbarSpellSlot)(data), start
}

//export sub_460D50
func sub_460D50() C.int {
	slots := [...]**nox_window{
		&C.dword_5d4594_1049500, &C.dword_5d4594_1049504,
		&C.dword_5d4594_1049520, &C.dword_5d4594_1049508,
		&C.dword_5d4594_1049512, &C.dword_5d4594_1049516,
	}
	for _, slot := range slots {
		nox_xxx_windowDestroyMB_46C4E0(*slot)
		*slot = nil
	}
	nox_quickbar_windows_destroy_all()
	for row := range 5 {
		clear(memmap.PtrT[[11]uint32](0x5D4594, 1048196+256*uintptr(row)+208)[:])
	}
	clear(memmap.PtrT[[4]uint32](0x5D4594, 1048148)[:])
	C.dword_5d4594_1049532 = nil
	*memmap.PtrUint32(0x5D4594, 1047928) = 0
	C.dword_5d4594_1047932 = 0
	return 0
}

//export nox_xxx_cliPrepareGameplay1_460E60
func nox_xxx_cliPrepareGameplay1_460E60() C.int {
	if sub_460D40() != 0 {
		sub_460D50()
	}
	if nox_xxx_quickBarCreate_45E190() == 0 {
		return 0
	}
	sub_460EA0(C.int(nox_client_getRenderGUI()))
	return 1
}

//export sub_460EA0
func sub_460EA0(show C.int) C.int {
	return sub_460B90(show)
}

//export nox_xxx_buttonFindFirstEmptySlot_461250
func nox_xxx_buttonFindFirstEmptySlot_461250() C.int {
	rows, start := quickbarSpellRows()
	if rows == nil {
		return -1
	}
	for i := range rows {
		row := (start + i) % len(rows)
		for slot, spell := range rows[row] {
			if spell.id == 0 {
				nox_xxx_clientUpdateButtonRow_45E110(C.int(row))
				return C.int(slot)
			}
		}
	}
	return -1
}

//export sub_4612A0
func sub_4612A0() C.int {
	rows, start := quickbarSpellRows()
	if rows == nil {
		return -1
	}
	for slot, spell := range rows[start] {
		if spell.id == 0 {
			return C.int(slot)
		}
	}
	return -1
}

//export nox_xxx_buttonHaveSpellInBarMB_4612D0
func nox_xxx_buttonHaveSpellInBarMB_4612D0(id C.int) C.int {
	rows, start := quickbarSpellRows()
	if rows == nil {
		return 0
	}
	for i := range rows {
		for _, spell := range rows[(start+i)%len(rows)] {
			if spell.id == uint32(id) {
				return 1
			}
		}
	}
	return 0
}

//export nox_xxx_buttonSetImgMB_461320
func nox_xxx_buttonSetImgMB_461320(slot C.int, dst *C.uint32_t) {
	if dst == nil {
		return
	}
	if win := nox_quickbar_button(g2qMain(), slot); win != nil {
		pos := asWindow(win).GlobalPos()
		coords := unsafe.Slice(dst, 2)
		coords[0], coords[1] = C.uint32_t(pos.X), C.uint32_t(pos.Y)
	}
}

//export sub_461360
func sub_461360(id C.int) C.int {
	rows, start := quickbarSpellRows()
	if rows == nil {
		return 0
	}
	for i := range rows {
		row := &rows[(start+i)%len(rows)]
		for slot := range row {
			if row[slot].id == uint32(id) {
				row[slot].id = 0
			}
		}
	}
	return C.int(40*((start+4)%5) + 32)
}

//export sub_461400
func sub_461400() C.int {
	rows := (*[5][5]quickbarSpellSlot)(g2qMain())
	if rows == nil {
		return 0
	}
	for slot := range rows[0] {
		for row := range rows {
			off := uintptr(1047564 + 40*row + 8*slot)
			rows[row][slot].id = memmap.Uint32(0x5D4594, off)
			rows[row][slot].mode = memmap.Uint8(0x5D4594, off+4)
		}
	}
	return 232
}

//export sub_461440
func sub_461440(value C.int) C.int {
	*memmap.PtrInt32(0x5D4594, 1049688) = int32(value)
	return value
}

//export sub_461450
func sub_461450() C.int {
	return C.int(memmap.Int32(0x5D4594, 1049688))
}
