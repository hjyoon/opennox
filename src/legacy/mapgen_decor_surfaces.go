package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import "unsafe"

func mapgenAllocateDecorRecord(size C.size_t) (unsafe.Pointer, C.uint32_t) {
	record := nox_mapgen_alloc_memory(1, size)
	if record == nil {
		return nil, 0
	}
	token := registerMapgenPointer(record)
	if token == 0 {
		nox_mapgen_release_memory(record)
		return nil, 0
	}
	return record, C.uint32_t(token)
}

func mapgenReleaseDecorRecord(record unsafe.Pointer) {
	forgetMapgenPointer(record)
	nox_mapgen_release_memory(record)
}

func mapgenReadDecorName(file *C.FILE, destination *[60]C.char) bool {
	buf := mapgenThemeTokenBuffer()
	if nox_xxx_mapGenReadLine_51E540(file, buf) == 0 {
		return false
	}
	dst := (*[60]byte)(unsafe.Pointer(destination))
	return copyMapgenName(dst[:], GoStringP(unsafe.Pointer(buf)))
}

func mapgenResolveDecorDefinition(token C.uint32_t) *C.nox_mapgen_decor_definition_legacy {
	return (*C.nox_mapgen_decor_definition_legacy)(resolveMapgenPointer(uintptr(token)))
}

//export nox_mapgenReadWallFloorNative_51FE00
func nox_mapgenReadWallFloorNative_51FE00(definition *C.nox_mapgen_decor_definition_legacy, file *C.FILE) *C.nox_mapgen_wall_floor_legacy {
	if definition == nil || file == nil {
		return nil
	}
	record, token := mapgenAllocateDecorRecord(C.sizeof_nox_mapgen_wall_floor_legacy)
	if record == nil {
		return nil
	}
	surface := (*C.nox_mapgen_wall_floor_legacy)(record)
	if !mapgenReadDecorName(file, &surface.wall_name) || !mapgenReadDecorName(file, &surface.floor_name) {
		mapgenReleaseDecorRecord(record)
		return nil
	}
	surface.next_token = definition.wall_floors_token
	definition.wall_floors_token = token
	definition.wall_floor_count++
	return surface
}

//export nox_xxx_genDecorReadWallFloor_51FE00
func nox_xxx_genDecorReadWallFloor_51FE00(definitionToken C.int, file *C.FILE) *C.char {
	definition := mapgenResolveDecorDefinition(C.uint32_t(definitionToken))
	return (*C.char)(unsafe.Pointer(nox_mapgenReadWallFloorNative_51FE00(definition, file)))
}

func mapgenResolveDecorFill(token C.uint32_t) *C.nox_mapgen_decor_fill_legacy {
	return (*C.nox_mapgen_decor_fill_legacy)(resolveMapgenPointer(uintptr(token)))
}

//export nox_mapgenReadDecorFillNative_51FEC0
func nox_mapgenReadDecorFillNative_51FEC0(surface *C.nox_mapgen_wall_floor_legacy, kind C.int, file *C.FILE) C.int {
	if surface == nil || file == nil {
		return 0
	}
	record, token := mapgenAllocateDecorRecord(C.sizeof_nox_mapgen_decor_fill_legacy)
	if record == nil {
		return 0
	}
	fill := (*C.nox_mapgen_decor_fill_legacy)(record)
	fill._type = C.uint32_t(kind)
	if !mapgenReadDecorName(file, &fill.object_name) || !mapgenReadDecorName(file, &fill.tile_name) {
		mapgenReleaseDecorRecord(record)
		return 0
	}
	tail := mapgenResolveDecorFill(surface.fills_token)
	if tail == nil {
		surface.fills_token = token
		return 1
	}
	for tail.next_token != 0 {
		tail = mapgenResolveDecorFill(tail.next_token)
		if tail == nil {
			mapgenReleaseDecorRecord(record)
			return 0
		}
	}
	tail.next_token = token
	return 1
}

//export sub_51FEC0
func sub_51FEC0(surfaceToken, kind C.int, file *C.FILE) C.int {
	surface := mapgenResolveWallFloor(C.uint32_t(surfaceToken))
	return nox_mapgenReadDecorFillNative_51FEC0(surface, kind, file)
}

//export nox_xxx_mapgenCheckSettings_520AD0
func nox_xxx_mapgenCheckSettings_520AD0(settings *C.int) C.int {
	if settings == nil {
		return 0
	}
	fields := (*mapgenDecorSettings)(unsafe.Pointer(settings))
	totals := fields[2:]
	clear(totals)
	for def := mapgenResolveDecorDefinition(C.uint32_t(fields[0])); def != nil; def = mapgenResolveDecorDefinition(def.next_token) {
		for i := range totals {
			if def.constraint_mask == 0 || def.constraint_mask&(1<<i) != 0 {
				totals[i] += C.int(def.frequency)
			}
		}
	}
	for _, total := range totals {
		if total == 0 {
			return 0
		}
	}
	return 1
}
