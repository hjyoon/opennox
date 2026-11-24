package legacy

/*
#include "GAME3_1.h"
extern uint32_t dword_5d4594_1321196, dword_5d4594_1321208;
extern nox_window *dword_5d4594_1321032, *dword_5d4594_1321040;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

const (
	summonRecordsBase = 1321052
	summonRecordCount = 4
	summonGridBase    = 1321180
	summonGridSide    = 2
)

type summonRecord = C.nox_gui_summon_record

func summonRecordAt(index int) *summonRecord {
	if index < 0 || index >= summonRecordCount {
		return nil
	}
	return memmap.PtrT[summonRecord](0x5D4594, summonRecordsBase+uintptr(index)*unsafe.Sizeof(summonRecord{}))
}

func summonGridCell(x, y int32) **summonRecord {
	if x < 0 || x >= summonGridSide || y < 0 || y >= summonGridSide {
		return nil
	}
	slot := memmap.PtrPtr(0x5D4594, summonGridBase+4*uintptr(y+summonGridSide*x))
	return (**summonRecord)(unsafe.Pointer(slot))
}

//export nox_xxx_wndSummonGet_4C2410
func nox_xxx_wndSummonGet_4C2410(pos *C.nox_summon_const_position_t) *summonRecord {
	if pos == nil {
		return nil
	}
	if cell := summonGridCell(int32(pos.field_0), int32(pos.field_4)); cell != nil {
		return *cell
	}
	return nil
}

//export sub_4C2BF0
func sub_4C2BF0() *C.int {
	for x := int32(0); x < summonGridSide; x++ {
		for y := int32(0); y < summonGridSide; y++ {
			*summonGridCell(x, y) = nil
		}
	}
	return (*C.int)(memmap.PtrOff(0x5D4594, summonGridBase+20))
}

func firstActiveSummon(start int) *summonRecord {
	for i := start; i < summonRecordCount; i++ {
		if record := summonRecordAt(i); record.active != 0 {
			return record
		}
	}
	return nil
}

//export sub_4C2D60
func sub_4C2D60() *summonRecord {
	return firstActiveSummon(0)
}

//export sub_4C2D90
func sub_4C2D90(record *summonRecord) *summonRecord {
	if record == nil {
		return nil
	}
	return firstActiveSummon(int(int32(record.slot) + 1))
}

func summonCarnivorousPlant() uint32 {
	if C.dword_5d4594_1321208 == 0 {
		C.dword_5d4594_1321208 = C.uint32_t(nox_xxx_getNameId_4E3AA0(internCStr("CarnivorousPlant")))
	}
	return uint32(C.dword_5d4594_1321208)
}

//export sub_4C2DD0
func sub_4C2DD0(record *C.nox_summon_const_record_t) C.int {
	plant := summonCarnivorousPlant()
	return C.int(bool2int(uint32(record.thing_id) != plant))
}

//export sub_4C2E00
func sub_4C2E00() C.int {
	summonCarnivorousPlant()
	for record := sub_4C2D60(); record != nil; record = sub_4C2D90(record) {
		if sub_4C2DD0((*C.nox_summon_const_record_t)(record)) != 0 {
			return 1
		}
	}
	return 0
}

//export nox_xxx_cliSummonCreat_4C2E50
func nox_xxx_cliSummonCreat_4C2E50(code, thing, silent C.int) {
	for record := sub_4C2D60(); record != nil; record = sub_4C2D90(record) {
		if record.net_code == C.uint32_t(code) && record.thing_id == C.uint32_t(thing) {
			return
		}
	}
	record := sub_4C2F20()
	if record == nil {
		return
	}
	record.net_code, record.thing_id = C.uint32_t(code), C.uint32_t(thing)
	record.size = C.uint8_t(sub_4C2EF0(thing))
	sub_4C2F70()

	state := memmap.PtrUint8(0x5D4594, 1321200)
	if *state == 0 || *state == 3 {
		if silent == 0 {
			nox_xxx_clientPlaySoundSpecial_452D80(801, 100)
		}
		*state = 1
		g2ServerHidden(C.dword_5d4594_1321032, false)
		g2ServerHidden(C.dword_5d4594_1321040, false)
	}
	C.dword_5d4594_1321196++
}

//export sub_4C2EF0
func sub_4C2EF0(thing C.int) C.int {
	class := nox_get_thing(int32(thing)).sub_class
	if class&1 != 0 {
		return 1
	}
	if class&2 != 0 {
		return 2
	}
	return 4
}

//export sub_4C2F20
func sub_4C2F20() *summonRecord {
	for i := 0; i < summonRecordCount; i++ {
		if record := summonRecordAt(i); record.active == 0 {
			*record = summonRecord{active: 1, slot: C.uint32_t(i)}
			return record
		}
	}
	return nil
}

//export sub_4C2F70
func sub_4C2F70() *summonRecord {
	sub_4C2BF0()
	for _, size := range [...]C.uint8_t{1, 2, 4} {
		for record := sub_4C2D60(); record != nil; record = sub_4C2D90(record) {
			if record.size == size {
				sub_4C2FD0(record)
			}
		}
	}
	return nil
}

//export sub_4C2FD0
func sub_4C2FD0(record *summonRecord) C.int {
	size := int32(record.size)
	for slot := int32(0); slot < 4; slot += size {
		offset := 184456 + 8*uintptr(slot)
		x := *memmap.PtrInt32(0x587000, offset)
		y := *memmap.PtrInt32(0x587000, offset+4)
		record.grid_x, record.grid_y = C.int32_t(x), C.int32_t(y)
		if summonGridHasSpace(x, y, size) {
			return C.int(setSummonGridRect(x, y, size, record))
		}
	}
	return C.int(record.size)
}

func summonGridSpan(size int32) (width, height int32) {
	switch size {
	case 1, 2:
		return 1, size
	case 4:
		return 2, 2
	default:
		return size, size
	}
}

func setSummonGridRect(x, y, size int32, record *summonRecord) int32 {
	width, height := summonGridSpan(size)
	for row := y; row < y+height; row++ {
		for col := x; col < x+width; col++ {
			if cell := summonGridCell(col, row); cell != nil {
				*cell = record
			}
		}
	}
	return y + height
}

//export sub_4C3030
func sub_4C3030(pos *C.nox_summon_const_i32_t, size C.int, record *summonRecord) C.int {
	p := (*[2]int32)(unsafe.Pointer(pos))
	return C.int(setSummonGridRect(p[0], p[1], int32(size), record))
}

func summonGridHasSpace(x, y, size int32) bool {
	width, height := summonGridSpan(size)
	for row := y; row < y+height; row++ {
		for col := x; col < x+width; col++ {
			cell := summonGridCell(col, row)
			if cell == nil || *cell != nil {
				return false
			}
		}
	}
	return true
}

//export sub_4C30C0
func sub_4C30C0(pos *C.nox_summon_const_i32_t, size C.int) C.int {
	if size <= 0 {
		return 1
	}
	p := (*[2]int32)(unsafe.Pointer(pos))
	return C.int(bool2int(summonGridHasSpace(p[0], p[1], int32(size))))
}

//export nox_xxx_cliSummonOnDieOrBanish_4C3140
func nox_xxx_cliSummonOnDieOrBanish_4C3140(code C.int, silent unsafe.Pointer) {
	removeSummonCreature(code, silent != nil)
}

func removeSummonCreature(code C.int, silent bool) {
	record := sub_4C31D0(code)
	if record == nil {
		return
	}
	if record == C.dword_5d4594_1321204 {
		C.nox_xxx_guiHideSummonWindow_4C2470()
	}
	if sprite := nox_xxx_netSpriteByCodeDynamic_45A6F0(int32(code)); sprite != nil {
		sprite.flags30 &^= 0x40000000
	}
	setSummonGridRect(int32(record.grid_x), int32(record.grid_y), int32(record.size), nil)
	sub_4C3210(record)
	sub_4C2F70()
	C.dword_5d4594_1321196--
	if C.dword_5d4594_1321196 == 0 {
		if !silent {
			nox_xxx_clientPlaySoundSpecial_452D80(802, 100)
		}
		*memmap.PtrUint8(0x5D4594, 1321200) = 3
	}
}

//export sub_4C31D0
func sub_4C31D0(code C.int) *summonRecord {
	for i := 0; i < summonRecordCount; i++ {
		record := summonRecordAt(i)
		if record.active != 0 && record.net_code == C.uint32_t(code) {
			return record
		}
	}
	return nil
}

//export sub_4C3210
func sub_4C3210(record *summonRecord) *summonRecord {
	record.active = 0
	return record
}

//export nox_xxx_sprite_4C3220
func nox_xxx_sprite_4C3220(sprite *nox_drawable) C.int {
	return C.int(bool2int(sub_4C31D0(C.int(sprite.field_32)) != nil))
}

//export sub_4C3260
func sub_4C3260() C.int {
	return C.int(bool2int(sub_4C2D60() != nil))
}
