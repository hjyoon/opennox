package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import (
	"math/bits"
	"unsafe"
)

type mapgenDecorSettings [8]C.int

func mapgenDecorSettingsAt(raw *C.uint8_t, index int) *mapgenDecorSettings {
	theme := (*[mapgenThemeWordCount]C.int)(unsafe.Pointer(raw))
	return (*mapgenDecorSettings)(unsafe.Pointer(&theme[index]))
}

func mapgenDecorMatchesBucket(definition *C.nox_mapgen_decor_definition_legacy, room *mapgenRankedRoom) bool {
	return definition != nil && room != nil &&
		(definition.constraint_mask == 0 || room.bucket&uint32(definition.constraint_mask) != 0)
}

func mapgenDecorFitsRoom(definition *C.nox_mapgen_decor_definition_legacy, room *mapgenRankedRoom) bool {
	if definition == nil || room == nil {
		return false
	}
	size := max(room.cellDimensions[0], room.cellDimensions[1])
	return size >= definition.min_room_size && size <= definition.max_room_size
}

func mapgenConsumeDecorLimit(settings *mapgenDecorSettings, token C.uint32_t) {
	definition := mapgenResolveDecorDefinition(token)
	if settings == nil || definition == nil || definition.occur_limit == 0 {
		return
	}
	definition.occur_limit--
	if definition.occur_limit != 0 {
		return
	}
	for bucket := 1; bucket <= 4; bucket++ {
		if definition.constraint_mask&(1<<bucket) != 0 {
			settings[2+bucket] -= C.int(definition.frequency)
		}
	}

	head := (*C.uint32_t)(unsafe.Pointer(&settings[0]))
	link := head
	for *link != 0 && *link != token {
		node := mapgenResolveDecorDefinition(*link)
		if node == nil {
			return
		}
		link = &node.next_token
	}
	if *link == 0 {
		return
	}
	*link = definition.next_token
	definition.next_token = 0
	definition.exhausted = 1
	link = head
	for *link != 0 {
		node := mapgenResolveDecorDefinition(*link)
		if node == nil {
			return
		}
		link = &node.next_token
	}
	*link = token
}

func mapgenSelectRoomDecor(room *mapgenRankedRoom, settings *mapgenDecorSettings) C.uint32_t {
	if room == nil || settings == nil {
		return 0
	}
	for attempt := 0; attempt <= 100; attempt++ {
		bucket := room.bucket
		if bucket == 0 || bucket > 32 || bucket&(bucket-1) != 0 {
			break
		}
		total := settings[2+bits.TrailingZeros32(bucket)]
		if total <= 0 {
			break
		}
		roll := nox_xxx_mapGenRandFunc_526AC0(0, total)
		for token := C.uint32_t(settings[0]); token != 0; {
			definition := mapgenResolveDecorDefinition(token)
			if definition == nil {
				break
			}
			if mapgenDecorMatchesBucket(definition, room) {
				roll -= C.int(definition.frequency)
				if roll <= 0 {
					if mapgenDecorFitsRoom(definition, room) && definition.exhausted == 0 {
						mapgenConsumeDecorLimit(settings, token)
						return token
					}
					break
				}
			}
			token = definition.next_token
		}
	}
	mapgenRankedRoomToken(room)
	return 0
}

//export nox_mapgenSelectBackdropNative_524070
func nox_mapgenSelectBackdropNative_524070(raw, room *C.uint8_t) C.int {
	if raw == nil || room == nil {
		return 0
	}
	record := (*mapgenRankedRoom)(unsafe.Pointer(room))
	record.decor = mapgenSelectRoomDecor(record, mapgenDecorSettingsAt(raw, 46))
	return C.int(bool2int(record.decor != 0))
}

//export sub_524070
func sub_524070(theme, room C.int) C.int {
	return nox_mapgenSelectBackdropNative_524070(mapgenThemeByToken(theme),
		(*C.uint8_t)(unsafe.Pointer(mapgenRankedRoomByToken(uint32(room)))))
}

//export sub_524090
func sub_524090(room C.int, raw *C.int) C.int {
	settings := (*mapgenDecorSettings)(unsafe.Pointer(raw))
	return C.int(mapgenSelectRoomDecor(mapgenRankedRoomByToken(uint32(room)), settings))
}

//export nox_xxx_mapGenDecorChkConstaint_5241C0
func nox_xxx_mapGenDecorChkConstaint_5241C0(definition, room C.int) C.int {
	return C.int(bool2int(mapgenDecorMatchesBucket(mapgenResolveDecorDefinition(C.uint32_t(definition)),
		mapgenRankedRoomByToken(uint32(room)))))
}

//export nox_xxx_mapGenChkDecorFillsRoom_5241F0
func nox_xxx_mapGenChkDecorFillsRoom_5241F0(definition, room C.int) C.int {
	return C.int(bool2int(mapgenDecorFitsRoom(mapgenResolveDecorDefinition(C.uint32_t(definition)),
		mapgenRankedRoomByToken(uint32(room)))))
}

//export nox_xxx_mapGenDecorChkLimit_524220
func nox_xxx_mapGenDecorChkLimit_524220(raw *C.int, token C.int) C.char {
	mapgenConsumeDecorLimit((*mapgenDecorSettings)(unsafe.Pointer(raw)), C.uint32_t(token))
	return 0
}

func mapgenAssignRequiredDecor(settings *mapgenDecorSettings, hall bool) bool {
	for token := C.uint32_t(settings[0]); token != 0; {
		definition := mapgenResolveDecorDefinition(token)
		if definition == nil {
			return false
		}
		next := definition.next_token
		if definition.must_occur != 0 && definition.occurred == 0 {
			found := false
			for node := mapgenRooms.head; node != nil; node = mapgenNextListedRoom(node, mapgenResolveListedRoom) {
				room := (*mapgenRankedRoom)(unsafe.Pointer(node))
				if room.decor != 0 || mapgenIsHall(uint32(room.kind)) != hall || definition.exhausted != 0 ||
					!mapgenDecorMatchesBucket(definition, room) || !mapgenDecorFitsRoom(definition, room) {
					continue
				}
				mapgenConsumeDecorLimit(settings, token)
				room.decor = token
				definition.occurred = 1
				found = true
				break
			}
			if !found {
				return false
			}
		}
		token = next
	}
	return true
}

//export nox_mapgenMakeRoomsNative_524310
func nox_mapgenMakeRoomsNative_524310(theme *C.uint8_t) C.int {
	if theme == nil {
		return 0
	}
	rooms := mapgenDecorSettingsAt(theme, 22)
	halls := mapgenDecorSettingsAt(theme, 30)
	if !mapgenAssignRequiredDecor(rooms, false) || !mapgenAssignRequiredDecor(halls, true) {
		return 0
	}
	for node := mapgenRooms.head; node != nil; node = mapgenNextListedRoom(node, mapgenResolveListedRoom) {
		room := (*mapgenRankedRoom)(unsafe.Pointer(node))
		if room.decor != 0 {
			continue
		}
		settings := rooms
		if mapgenIsHall(uint32(room.kind)) {
			settings = halls
		}
		room.decor = mapgenSelectRoomDecor(room, settings)
		if room.decor == 0 {
			return 0
		}
	}
	return 1
}

//export nox_xxx_mapGenMakeRooms_524310
func nox_xxx_mapGenMakeRooms_524310(theme C.int) C.int {
	return nox_mapgenMakeRoomsNative_524310(mapgenThemeByToken(theme))
}
