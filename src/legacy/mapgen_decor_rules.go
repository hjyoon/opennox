package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import (
	"bytes"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const (
	mapgenDecorConstraintNames = 253144
	mapgenDecorFillNames       = 253200
	mapgenDecorEntryNames      = 253216
)

func mapgenDecorKeywordIndex(offset uintptr, uppercase string) (int, bool) {
	for index := 0; ; index++ {
		name := *memmap.PtrPtr(0x587000, offset+4*uintptr(index))
		if name == nil {
			return 0, false
		}
		if mapgenUppercaseToken(GoStringP(name)) == uppercase {
			return index, true
		}
	}
}

//export nox_xxx_genDecorReadOccurConstraint_520810
func nox_xxx_genDecorReadOccurConstraint_520810(token C.int, file *C.FILE) C.int {
	def := mapgenResolveDecorDefinition(C.uint32_t(token))
	if def == nil {
		return 0
	}
	def.constraint_mask = 0
	scratch := mapgenThemeTokenBuffer()
	if nox_xxx_mapGenReadLine_51E540(file, scratch) == 0 {
		return 0
	}
	start := (*byte)(unsafe.Pointer(scratch))
	text := unsafe.Slice(start, alloc.StrLen(start))
	if mapgenUppercaseToken(string(text)) == "NONE" {
		return 1
	}
	for len(text) != 0 {
		if text[0] == '+' {
			text = text[1:]
			continue
		}
		end := bytes.IndexByte(text, '+')
		part := text
		if end >= 0 {
			text[end] = 0
			part, text = text[:end], text[end+1:]
		} else {
			text = nil
		}
		index, ok := mapgenDecorKeywordIndex(mapgenDecorConstraintNames, mapgenUppercaseToken(string(part)))
		if !ok {
			return 0
		}
		def.constraint_mask |= C.uint8_t(1 << index)
	}
	return 1
}

//export nox_xxx_genDecorReadOccurLimit_5208D0
func nox_xxx_genDecorReadOccurLimit_5208D0(token C.int, file *C.FILE) C.int {
	def := mapgenResolveDecorDefinition(C.uint32_t(token))
	text := mapgenThemeTokenBuffer()
	if def == nil || nox_xxx_mapGenReadLine_51E540(file, text) == 0 {
		return 0
	}
	def.occur_limit = C.uint8_t(parseLegacyDecimal32((*C.char)(unsafe.Pointer(text))))
	return 1
}

//export nox_xxx_genDecorReadFrequency_520910
func nox_xxx_genDecorReadFrequency_520910(token C.int, file *C.FILE) C.int {
	def := mapgenResolveDecorDefinition(C.uint32_t(token))
	text := mapgenThemeTokenBuffer()
	if def == nil || nox_xxx_mapGenReadLine_51E540(file, text) == 0 {
		return 0
	}
	switch mapgenUppercaseToken(GoStringP(unsafe.Pointer(text))) {
	case "COMMON":
		def.frequency = 1000
	case "UNCOMMON":
		def.frequency = 500
	case "RARE":
		def.frequency = 100
	case "VERY_RARE":
		def.frequency = 10
	case "HARDLY_EVER":
		def.frequency = 1
	}
	return 1
}

//export nox_xxx_genDecorReadRoomSizeCon_5209F0
func nox_xxx_genDecorReadRoomSizeCon_5209F0(token C.int, file *C.FILE) C.int {
	def := mapgenResolveDecorDefinition(C.uint32_t(token))
	text := mapgenThemeTokenBuffer()
	if def == nil || nox_xxx_mapGenReadLine_51E540(file, text) == 0 {
		return 0
	}
	value := (*C.char)(unsafe.Pointer(text))
	def.min_room_size = C.int32_t(parseLegacyDecimal32(value))
	if nox_xxx_mapGenReadLine_51E540(file, text) == 0 {
		return 0
	}
	if GoString(value) == "*" {
		def.max_room_size = 999999
	} else {
		def.max_room_size = C.int32_t(parseLegacyDecimal32(value))
	}
	return 1
}

func mapgenReadFixedName(file *C.FILE, name *[60]C.char) C.int {
	length := 0
	parsed := mapgenReadPlainToken(file, func(index int, value byte) {
		length = index + 1
		if index < len(name) {
			name[index] = C.char(value)
		}
	})
	if length > len(name) {
		name[len(name)-1] = 0
		return 0
	}
	if parsed == 0 {
		return 0
	}
	return sub_51E670(file)
}

//export nox_xxx_genDecorReadDoor_520A90
func nox_xxx_genDecorReadDoor_520A90(token C.int, file *C.FILE) C.int {
	def := mapgenResolveDecorDefinition(C.uint32_t(token))
	if def == nil {
		return 0
	}
	return mapgenReadFixedName(file, &def.door_name)
}

//export nox_xxx_genDecorReadDoubleDoor_520AB0
func nox_xxx_genDecorReadDoubleDoor_520AB0(token C.int, file *C.FILE) C.int {
	def := mapgenResolveDecorDefinition(C.uint32_t(token))
	if def == nil {
		return 0
	}
	return mapgenReadFixedName(file, &def.double_door_name)
}
