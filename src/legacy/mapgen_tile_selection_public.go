package legacy

/*
#include "GAME4_1.h"
extern uint32_t dword_5d4594_3835348;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const (
	mapgenTileStateBase  = 0x973F18
	mapgenTileIDOffset   = 35912
	mapgenTileModeOffset = 35916
)

func mapgenTileNamesEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		x, y := a[i], b[i]
		if x >= 'A' && x <= 'Z' {
			x += 'a' - 'A'
		}
		if y >= 'A' && y <= 'Z' {
			y += 'a' - 'A'
		}
		if x != y {
			return false
		}
	}
	return true
}

//export nox_xxx_tileGetDefByName_51D4D0
func nox_xxx_tileGetDefByName_51D4D0(name *C.char) C.int {
	selected := memmap.PtrUint32(mapgenTileStateBase, mapgenTileIDOffset)
	if name == nil {
		*selected = 0
		return 0
	}
	text := alloc.GoString((*byte)(unsafe.Pointer(name)))
	matched := false
	for i := range floorTileStorage.definitions {
		def := &floorTileStorage.definitions[i]
		buffer := unsafe.Slice((*byte)(unsafe.Pointer(&def.name[0])), len(def.name))
		if mapgenTileNamesEqual(alloc.GoStringS(buffer), text) {
			*selected = uint32(i)
			matched = true
		}
	}
	if mapgenTileNamesEqual(text, "NONE") {
		*selected = 255
		return 1
	}
	if matched {
		return 1
	}
	*selected = 0
	return 0
}

//export nox_xxx_tileCheckImage_51D540
func nox_xxx_tileCheckImage_51D540(index C.int) C.int {
	selected := memmap.PtrUint32(mapgenTileStateBase, mapgenTileIDOffset)
	if index < 0 || index >= C.int(len(floorTileStorage.definitions)) {
		*selected = 0
		return 0
	}
	*selected = uint32(index)
	return 1
}

//export nox_xxx_tileCheckImageVari_51D570
func nox_xxx_tileCheckImageVari_51D570(variation C.int) C.int {
	index := memmap.Uint32(mapgenTileStateBase, mapgenTileIDOffset)
	// The name selector can leave the NONE sentinel in this slot.
	if index < uint32(len(floorTileStorage.definitions)) {
		def := &floorTileStorage.definitions[index]
		if variation < C.int(def.field_52)*C.int(def.field_53) {
			C.dword_5d4594_3835348 = C.uint32_t(variation)
			return 1
		}
	}
	C.dword_5d4594_3835348 = 0
	return 0
}

//export nox_xxx_tile_51D5C0
func nox_xxx_tile_51D5C0(mode C.int) C.int {
	if mode != 0 && mode != 1 {
		return 0
	}
	*memmap.PtrUint32(mapgenTileStateBase, mapgenTileModeOffset) = uint32(mode)
	return 1
}
