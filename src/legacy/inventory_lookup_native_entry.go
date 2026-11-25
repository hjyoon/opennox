package legacy

/*
#include "GAME2_1.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func inventoryLookupNativeSearch(netCode uint32) (*C.nox_inventory_cell_t, uint32) {
	for row := range inventoryRows {
		for column := range inventoryColumns {
			cell := inventoryGridCell(row, column)
			codes := unsafe.Slice(&cell.field_4, int(cell.field_140))
			for index, code := range codes {
				if uint32(code) == netCode {
					return cell, uint32(index)
				}
			}
		}
	}
	return nil, 0
}

//export nox_inventory_find_cell_native_461EF0
func nox_inventory_find_cell_native_461EF0(netCode C.int, stackIndex *C.uint32_t) *C.nox_inventory_cell_t {
	return inventoryLookupFindCell(uint32(netCode), (*uint32)(stackIndex))
}

func inventoryLookupFindCell(netCode uint32, stackIndex *uint32) *C.nox_inventory_cell_t {
	cell, index := inventoryLookupNativeSearch(netCode)
	if cell != nil && stackIndex != nil {
		*stackIndex = index
	}
	return cell
}

//export sub_461EF0
func sub_461EF0(netCode C.int) *C.char {
	return (*C.char)(inventoryLookupLegacyResult(uint32(netCode)))
}

func inventoryLookupLegacyResult(netCode uint32) unsafe.Pointer {
	cell, index := inventoryLookupNativeSearch(netCode)
	if cell == nil {
		return nil
	}
	// Stock writes the index before publishing the cell pointer.
	*memmap.PtrUint32(0x5D4594, 1049792) = index
	slot := memmap.PtrPtr(0x5D4594, 1049788)
	if unsafe.Sizeof(uintptr(0)) == 4 {
		// The original packed C word need not contain an initialized pointer.
		*(*uintptr)(unsafe.Pointer(slot)) = uintptr(unsafe.Pointer(cell))
	} else {
		*slot = unsafe.Pointer(cell)
	}
	return memmap.PtrOff(0x5D4594, 1049788)
}
