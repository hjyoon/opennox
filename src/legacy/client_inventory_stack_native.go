package legacy

/*
#include "GAME1_1.h"
#include "GAME2_1.h"
#include "client__gui__guiinv.h"

extern uint32_t dword_5d4594_1062484;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const (
	inventoryRows     = int(C.NOX_INVENTORY_ROW_COUNT)
	inventoryColumns  = int(C.NOX_INVENTORY_COL_COUNT)
	inventoryStackMax = 32
)

// Inventory callbacks retain cell addresses in the equipment and drag state.
var inventoryGrid, _ = alloc.New(C.nox_inventory_grid_t{})

//export nox_inventory_grid_storage
func nox_inventory_grid_storage() *C.nox_inventory_grid_t {
	return inventoryGrid
}

func inventoryGridCell(row, column int) *C.nox_inventory_cell_t {
	return &inventoryGrid[row+inventoryRows*column]
}

func inventoryCellCodes(cell *C.nox_inventory_cell_t) []C.uint32_t {
	return unsafe.Slice(&cell.field_4, inventoryStackMax)
}

//export nox_xxx_clientSetAltWeapon_461550
func nox_xxx_clientSetAltWeapon_461550(cell *C.nox_inventory_cell_t) C.int {
	previous := (*C.nox_inventory_cell_t)(globalStateNativePointer(2))
	C.dword_5d4594_1062484 = 0
	if previous != nil {
		C.dword_5d4594_1062484 = previous.field_4
	}
	if cell != nil && cell.field_0 == nil {
		cell = nil
	}
	globalStateNativeStorePointer(globalStateNativeSlot(2), unsafe.Pointer(cell))
	sub_4619F0()
	if cell == nil {
		return nox_xxx_clientReportSecondaryWeapon_4BF010(nil)
	}
	cell.field_0.field_32 = cell.field_4
	cell.field_136 = 1
	return nox_xxx_clientReportSecondaryWeapon_4BF010(cell.field_0)
}

//export sub_461970
func sub_461970(netCode, thingType C.int) *C.nox_inventory_cell_t {
	c, _, _ := clientInventoryStackPickup(netCode, thingType)
	return c
}

func clientInventoryStackPickup(netCode, thingType C.int) (*C.nox_inventory_cell_t, int, int) {
	thing := nox_get_thing(int32(thingType))
	if thing == nil || thing.pri_class&0x4000000 != 0 {
		return nil, 0, 0
	}
	for row := range inventoryRows - 1 {
		for column := range inventoryColumns {
			cell := inventoryGridCell(row, column)
			n := int(cell.field_140)
			if n != 0 && n < inventoryStackMax && cell.field_0 != nil && cell.field_0.field_27 == C.uint32_t(thingType) {
				inventoryCellCodes(cell)[n] = C.uint32_t(netCode)
				cell.field_140++
				return cell, row, column
			}
		}
	}
	return nil, 0, 0
}

//export sub_4619F0
func sub_4619F0() *C.char {
	for row := range inventoryRows {
		for column := range inventoryColumns {
			cell := inventoryGridCell(row, column)
			if cell.field_140 != 0 {
				cell.field_136 = 0
			}
		}
	}
	lastRow := inventoryGridCell(inventoryRows-1, 0)
	offset := uintptr(inventoryRows*inventoryColumns) * C.sizeof_nox_inventory_cell_t
	return (*C.char)(unsafe.Add(unsafe.Pointer(&lastRow.field_140), offset))
}

//export sub_461B50
func sub_461B50() {
	var ordered [(inventoryRows - 1) * inventoryColumns]*C.nox_inventory_cell_t
	for row := range inventoryRows - 1 {
		for column := range inventoryColumns {
			ordered[row*inventoryColumns+column] = inventoryGridCell(row, column)
		}
	}
	for i := 0; i < len(ordered); {
		target := ordered[i]
		if target.field_140 != 0 {
			i++
			continue
		}
		j := i + 1
		for j < len(ordered) && ordered[j].field_140 == 0 {
			j++
		}
		if j == len(ordered) {
			for _, empty := range ordered[i:] {
				empty.field_132, empty.field_136 = 0, 0
			}
			return
		}
		source := ordered[j]
		appearance := source.field_0
		if appearance != nil && appearance.flags28&0x4000000 == 0 && source.field_140 <= inventoryStackMax {
			codes := inventoryCellCodes(source)
			for _, dest := range ordered {
				if source.field_140 == 0 {
					break
				}
				if dest == source || dest.field_140 == 0 || dest.field_140 >= inventoryStackMax ||
					dest.field_0 == nil || dest.field_0.field_27 != appearance.field_27 {
					continue
				}
				destCodes := inventoryCellCodes(dest)
				for source.field_140 > 0 && dest.field_140 < inventoryStackMax {
					source.field_140--
					destCodes[dest.field_140] = codes[source.field_140]
					dest.field_140++
				}
			}
			if source.field_140 == 0 {
				nox_xxx_spriteDelete_45A4B0(source.field_0)
				source.field_0 = nil
				continue
			}
		}
		*target = *source
		if target.field_136 != 0 {
			globalStateNativeStorePointer(globalStateNativeSlot(2), unsafe.Pointer(target))
		}
		source.field_140, source.field_0, source.field_132 = 0, nil, 0
		i++
	}
}

//export sub_461E60
func sub_461E60(cell *C.nox_inventory_cell_t, stackIndex C.uint32_t) *C.nox_inventory_cell_t {
	if cell == nil || stackIndex >= C.uint32_t(cell.field_140) || cell.field_140 > inventoryStackMax {
		return cell
	}
	count := int(cell.field_140)
	codes := inventoryCellCodes(cell)
	copy(codes[stackIndex:count-1], codes[stackIndex+1:count])
	cell.field_140--
	if cell.field_140 == 0 {
		nox_xxx_spriteDelete_45A4B0(cell.field_0)
		cell.field_0 = nil
	}
	if cell.field_136 != 0 {
		nox_xxx_clientSetAltWeapon_461550(nil)
		cell.field_136 = 0
	}
	return cell
}

func inventoryItemCell(row, column int, thingType uint32) *C.nox_inventory_cell_t {
	cell := inventoryGridCell(row, column)
	if cell.field_140 == 0 || cell.field_0 == nil || uint32(cell.field_0.field_27) != thingType {
		return nil
	}
	return cell
}

func Nox_client_inventoryItemState(thingType uint32) (found bool, count uint32, currentHealth, maximumHealth uint16) {
	for row := range inventoryRows {
		for column := range inventoryColumns {
			cell := inventoryItemCell(row, column, thingType)
			if cell == nil {
				continue
			}
			count += uint32(cell.field_140)
			if !found {
				currentHealth = uint16(cell.field_0.field_73_1)
				maximumHealth = uint16(cell.field_0.field_73_2)
				found = true
			}
		}
	}
	return
}

func Nox_client_inventoryItemLocation(thingType uint32) (found bool, column, row int, netCode uint32) {
	for row := range inventoryRows {
		for column := range inventoryColumns {
			if cell := inventoryItemCell(row, column, thingType); cell != nil {
				return true, column, row, uint32(cell.field_4)
			}
		}
	}
	return false, -1, -1, 0
}
