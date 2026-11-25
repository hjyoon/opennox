package legacy

/*
#include "GAME4_2.h"
#include "cgo_const.h"
*/
import "C"

import "unsafe"

func mapgenCloneDecorFills(sourceToken C.uint32_t) (C.uint32_t, bool) {
	var head C.uint32_t
	link := &head
	for source := mapgenResolveDecorFill(sourceToken); source != nil; source = mapgenResolveDecorFill(source.next_token) {
		record, token := mapgenAllocateDecorRecord(C.sizeof_nox_mapgen_decor_fill_legacy)
		if record == nil {
			mapgenFreeDecorFillList(head)
			return 0, false
		}
		clone := (*C.nox_mapgen_decor_fill_legacy)(record)
		*clone = *source
		clone.next_token = 0
		*link = token
		link = &clone.next_token
	}
	return head, true
}

func mapgenFindNamedDecor(head *C.nox_mapgen_decor_definition_legacy, name string) *C.nox_mapgen_decor_definition_legacy {
	for def := head; def != nil; def = mapgenResolveDecorDefinition(def.next_token) {
		if GoString(&def.name[0]) == name {
			return def
		}
	}
	return nil
}

func mapgenReadDecorCopy(raw *C.uint32_t, destination *C.nox_mapgen_decor_definition_legacy, file *C.FILE) C.int {
	scratch := mapgenThemeTokenBuffer()
	if raw == nil || destination == nil || file == nil || nox_xxx_mapGenReadLine_51E540(file, scratch) == 0 {
		return 0
	}
	theme := (*[mapgenThemeWordCount]C.uint32_t)(unsafe.Pointer(raw))
	var first *C.nox_mapgen_decor_definition_legacy
	if destination.room_type < 3 {
		first = mapgenResolveDecorDefinition(theme[22+8*int(destination.room_type)])
	} else {
		first = mapgenResolveDecorDefinition(C.uint32_t(registerMapgenPointer(unsafe.Pointer(destination))))
	}
	name := GoStringP(unsafe.Pointer(scratch))
	source := mapgenFindNamedDecor(first, name)
	if source == nil {
		source = mapgenFindNamedDecor(mapgenResolveDecorDefinition(theme[38]), name)
	}
	if source == nil {
		return 0
	}

	var walls, sets, wallCount, setCount C.uint32_t
	published := false
	defer func() {
		if !published {
			mapgenFreeWallFloorList(walls)
			mapgenFreeDecorSetList(sets)
		}
	}()
	var wallTail *C.nox_mapgen_wall_floor_legacy
	for wall := mapgenResolveWallFloor(source.wall_floors_token); wall != nil; wall = mapgenResolveWallFloor(wall.next_token) {
		record, token := mapgenAllocateDecorRecord(C.sizeof_nox_mapgen_wall_floor_legacy)
		if record == nil {
			return 0
		}
		clone := (*C.nox_mapgen_wall_floor_legacy)(record)
		*clone = *wall
		fills, ok := mapgenCloneDecorFills(wall.fills_token)
		if !ok {
			mapgenReleaseDecorRecord(record)
			return 0
		}
		clone.fills_token = fills
		clone.next_token = walls
		walls = token
		if wallTail == nil {
			wallTail = clone
		}
		wallCount++
	}

	link := &sets
	for set := mapgenResolveDecorSet(source.decor_sets_token); set != nil; set = mapgenResolveDecorSet(set.next_token) {
		record, token := mapgenAllocateDecorRecord(C.sizeof_nox_mapgen_decor_set_legacy)
		if record == nil {
			return 0
		}
		clone := (*C.nox_mapgen_decor_set_legacy)(record)
		*clone = *set
		clone.shares_entries = 1
		clone.next_token = 0
		*link = token
		link = &clone.next_token
		setCount++
	}

	if wallTail != nil {
		wallTail.next_token = destination.wall_floors_token
		destination.wall_floors_token = walls
		destination.wall_floor_count += wallCount
	}
	if sets != 0 {
		*mapgenDecorSetTailLink(destination) = sets
		destination.decor_set_count += setCount
	}
	published = true
	return 1
}

//export nox_xxx_genDecorReadCopy_520660
func nox_xxx_genDecorReadCopy_520660(raw *C.uint32_t, definition *C.nox_cgo_const_char, file *C.FILE) C.int {
	record := resolveMapgenPointer(uintptr(unsafe.Pointer(definition)))
	return mapgenReadDecorCopy(raw, (*C.nox_mapgen_decor_definition_legacy)(record), file)
}
