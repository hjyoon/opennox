package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import "unsafe"

//export nox_xxx_genReadDecor_51F9F0
func nox_xxx_genReadDecor_51F9F0(raw *C.uint32_t, file *C.FILE) C.int {
	if raw == nil || file == nil {
		return 0
	}
	record, token := mapgenAllocateDecorRecord(C.sizeof_nox_mapgen_decor_definition_legacy)
	if record == nil {
		return 0
	}
	published := false
	defer func() {
		if !published {
			mapgenFreeDecorDefinitionList(token)
		}
	}()

	definition := (*C.nox_mapgen_decor_definition_legacy)(record)
	definition.frequency = 1000
	definition.max_room_size = 999999
	scratch := mapgenThemeTokenBuffer()
	if nox_xxx_mapGenReadLine_51E540(file, scratch) == 0 || mapgenReadFixedName(file, &definition.name) == 0 {
		return 0
	}
	switch mapgenUppercaseToken(GoStringP(unsafe.Pointer(scratch))) {
	case "ROOM":
		definition.room_type = 0
	case "HALL":
		definition.room_type = 1
	case "TEMPLATE":
		definition.room_type = 2
	case "BACKDROP":
		definition.room_type = 3
	default:
		return 0
	}

	var surface *C.nox_mapgen_wall_floor_legacy
	for nox_xxx_mapGenReadLine_51E540(file, scratch) != 0 {
		word := mapgenUppercaseToken(GoStringP(unsafe.Pointer(scratch)))
		result := C.int(1)
		switch word {
		case "END":
			theme := (*[mapgenThemeWordCount]C.uint32_t)(unsafe.Pointer(raw))
			index := 22 + 8*int(definition.room_type)
			definition.next_token = theme[index]
			theme[index] = token
			theme[index+1]++
			published = true
			return 1
		case "WALL_FLOOR":
			surface = nox_mapgenReadWallFloorNative_51FE00(definition, file)
			if surface == nil {
				return 0
			}
		case "OCCUR_CONSTRAINT":
			result = nox_xxx_genDecorReadOccurConstraint_520810(C.int(token), file)
		case "OCCUR_LIMIT":
			result = nox_xxx_genDecorReadOccurLimit_5208D0(C.int(token), file)
		case "MUST_OCCUR":
			definition.must_occur = 1
		case "FREQUENCY":
			result = nox_xxx_genDecorReadFrequency_520910(C.int(token), file)
		case "DOOR":
			result = mapgenReadFixedName(file, &definition.door_name)
		case "DOUBLE_DOOR":
			result = mapgenReadFixedName(file, &definition.double_door_name)
		case "ROOM_SIZE_CONSTRAINT":
			result = nox_xxx_genDecorReadRoomSizeCon_5209F0(C.int(token), file)
		case "COPY":
			result = C.nox_xxx_genDecorReadCopy_520660(raw, (*C.char)(record), file)
		case "DECOR_SET":
			result = C.nox_xxx_genDecorReadDecorSet_51FFA0(C.int(token), file)
		default:
			index, ok := mapgenDecorKeywordIndex(253200, word)
			if !ok || surface == nil {
				return 0
			}
			result = nox_mapgenReadDecorFillNative_51FEC0(surface, C.int(index), file)
		}
		if result == 0 {
			return 0
		}
	}
	return 0
}
