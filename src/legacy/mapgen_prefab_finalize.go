package legacy

/*
#include "GAME4_2.h"

extern uint32_t dword_5d4594_2487656;
*/
import "C"

import (
	"slices"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

func mapgenVisitPlacedPrefabs(theme *C.uint8_t, finish func(*mapgenThemePrefabRecord, *C.uint8_t, *mapgenRankedRoom)) {
	words := (*[mapgenThemeWordCount]C.uint32_t)(unsafe.Pointer(theme))
	for prefab := mapgenResolveThemePrefab(words[20]); prefab != nil; prefab = mapgenResolveThemePrefab(C.uint32_t(prefab.next)) {
		if prefab.placed == 0 {
			continue
		}
		entries := mapgenThemeByToken(C.int(prefab.foreach))
		room := mapgenRankedRoomByToken(prefab.room)
		finish(prefab, entries, room)
	}
}

func mapgenFinalizePrefab(theme *C.uint8_t, prefab *mapgenThemePrefabRecord, foreach *C.uint8_t, room *mapgenRankedRoom) {
	sub_502D70(C.int(prefab.areaIndex))
	markers := [...]uint32{
		uint32(C.dword_5d4594_2487656),
		memmap.Uint32(0x5D4594, 2487660),
		memmap.Uint32(0x5D4594, 2487664),
		memmap.Uint32(0x5D4594, 2487668),
	}
	for object := sub_504980(); object != nil; {
		next := sub_5049C0(object)
		if slices.Contains(markers[:], uint32(object.typ_ind)) {
			sub_504A10(object)
		}
		object = next
	}
	nox_mapgenApplyForeachNative_521C60(theme, C.uintptr_t(uintptr(unsafe.Pointer(foreach))))
	if room != nil {
		sub_503B30((*C.float2)(unsafe.Pointer(&room.position[0])))
	}
}

//export nox_mapgenPlacePrefabsWithCallback_526830
func nox_mapgenPlacePrefabsWithCallback_526830(theme *C.uint8_t, callback C.nox_mapgen_prefab_finalize_cb_526830) C.int {
	if theme == nil || callback == nil {
		return 0
	}
	mapgenVisitPlacedPrefabs(theme, func(prefab *mapgenThemePrefabRecord, foreach *C.uint8_t, room *mapgenRankedRoom) {
		ccall.CallVoidPtr4(unsafe.Pointer(callback), unsafe.Pointer(theme), unsafe.Pointer(prefab),
			unsafe.Pointer(foreach), unsafe.Pointer(room))
	})
	return 1
}

//export nox_xxx_mapGenPlacePrefabs_526830
func nox_xxx_mapGenPlacePrefabs_526830(theme *C.uint8_t) C.int {
	if theme == nil {
		return 0
	}
	mapgenVisitPlacedPrefabs(theme, func(prefab *mapgenThemePrefabRecord, foreach *C.uint8_t, room *mapgenRankedRoom) {
		mapgenFinalizePrefab(theme, prefab, foreach, room)
	})
	return 1
}
