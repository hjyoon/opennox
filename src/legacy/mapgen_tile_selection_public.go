package legacy

/*
#include "GAME4_1.h"
#include "GAME4_3.h"
#include "cgo_const.h"
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const (
	mapgenTileStateBase  = 0x973F18
	mapgenTileIDOffset   = 35912
	mapgenTileModeOffset = 35916
	floorEdgeNameOffset  = 28644
	floorEdgeVariations  = 28688
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
			legacyWords.dword_5d4594_3835348 = C.uint32_t(variation)
			return 1
		}
	}
	legacyWords.dword_5d4594_3835348 = 0
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

//export sub_543FB0
func sub_543FB0(name *C.nox_cgo_const_char) C.int {
	count := int32(*thingEdgeCountCell411850())
	if name == nil || count <= 0 {
		return -1
	}
	text := alloc.GoString((*byte)(unsafe.Pointer(name)))
	for i := int32(0); i < count; i++ {
		record := thingEdgeBytes411850(uint32(i), floorEdgeNameOffset, 1)
		if alloc.GoString(&record[0]) == text {
			return C.int(i)
		}
	}
	return -1
}

//export sub_544020
func sub_544020(name *C.char) C.int {
	floorEditState.selected_edge = 0
	if C.nox_strcmpi(internCStr("NONE"), name) == 0 {
		floorEditState.edge_tile = 255
		return 1
	}
	return nox_xxx_tileCheckByte3_544070(sub_543FB0((*C.nox_cgo_const_char)(name)))
}

//export nox_xxx_tileCheckByte3_544070
func nox_xxx_tileCheckByte3_544070(index C.int) C.int {
	if index < 0 || index >= C.int(int32(*thingEdgeCountCell411850())) {
		return 0
	}
	floorEditState.edge_tile = C.uint32_t(index)
	floorEditState.selected_edge = 1
	return 1
}

//export nox_xxx_tileCheckByte4_5440A0
func nox_xxx_tileCheckByte4_5440A0(variation C.int) C.int {
	if floorEditState.selected_edge == 0 {
		return 1
	}
	if int32(floorEditState.edge_tile) == -1 || variation < 0 {
		return 0
	}
	// The original indexes the variation limit with the incoming value.
	count := binary.LittleEndian.Uint16(thingEdgeBytes411850(uint32(variation), floorEdgeVariations, 2))
	if variation >= C.int(count) {
		return 0
	}
	floorEditState.edge_variation = C.uint32_t(variation)
	return 1
}
