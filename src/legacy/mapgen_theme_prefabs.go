package legacy

/*
#include "GAME4_2.h"
#include "mapgen_legacy_ptr.h"
*/
import "C"

import "unsafe"

type mapgenThemePrefabRecord struct {
	name     [60]byte
	_        [3]uint32
	required uint32
	_        [19]uint32
	foreach  uint32
	next     uint32
}

var _ [160]byte = [unsafe.Sizeof(mapgenThemePrefabRecord{})]byte{}

func mapgenResolveThemePrefab(token C.uint32_t) *mapgenThemePrefabRecord {
	return (*mapgenThemePrefabRecord)(C.nox_mapgenLegacyPtrResolve(C.uintptr_t(token)))
}

func mapgenReadThemePrefabEntry(theme *[mapgenThemeWordCount]C.uint32_t, file *C.FILE) *mapgenThemePrefabRecord {
	scratch := mapgenThemeTokenBuffer()
	if theme == nil || file == nil || nox_xxx_mapGenReadLine_51E540(file, scratch) == 0 {
		return nil
	}
	record, token := mapgenAllocateDecorRecord(C.size_t(unsafe.Sizeof(mapgenThemePrefabRecord{})))
	if record == nil {
		return nil
	}
	prefab := (*mapgenThemePrefabRecord)(record)
	if !copyMapgenName(prefab.name[:], GoStringP(unsafe.Pointer(scratch))) {
		mapgenReleaseDecorRecord(record)
		return nil
	}
	prefab.next = uint32(theme[20])
	theme[20] = token
	theme[21]++
	return prefab
}

//export sub_520CE0
func sub_520CE0(raw *C.uint8_t, file *C.FILE) *C.char {
	theme := (*[mapgenThemeWordCount]C.uint32_t)(unsafe.Pointer(raw))
	return (*C.char)(unsafe.Pointer(mapgenReadThemePrefabEntry(theme, file)))
}

//export nox_xxx_genReadPrefab_520BF0
func nox_xxx_genReadPrefab_520BF0(raw *C.uint8_t, file *C.FILE) C.int {
	if raw == nil || file == nil {
		return 0
	}
	theme := (*[mapgenThemeWordCount]C.uint32_t)(unsafe.Pointer(raw))
	scratch := mapgenThemeTokenBuffer()
	var current *mapgenThemePrefabRecord
	required := false
	for nox_xxx_mapGenReadLine_51E540(file, scratch) != 0 {
		switch mapgenUppercaseToken(GoStringP(unsafe.Pointer(scratch))) {
		case "END":
			return 1
		case "MUST_OCCUR":
			required = true
		case "AREAMAP":
			current = mapgenReadThemePrefabEntry(theme, file)
			if current == nil {
				return 0
			}
			current.required = uint32(bool2int(required))
			required = false
		case "FOREACH":
			if current == nil {
				return 0
			}
			entry, token := mapgenReadForeach(file)
			if entry == nil {
				return 0
			}
			entry.next = current.foreach
			current.foreach = uint32(token)
		}
	}
	return 0
}
