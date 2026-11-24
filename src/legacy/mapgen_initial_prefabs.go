package legacy

/*
#include "GAME4_2.h"

extern uint32_t dword_5d4594_2487656;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

func mapgenInitializePrefabExitTypes() {
	if C.dword_5d4594_2487656 != 0 {
		return
	}
	C.dword_5d4594_2487656 = C.uint32_t(nox_xxx_getNameId_4E3AA0(internCStr("ExitNorthMarker")))
	for i, name := range [...]string{"ExitSouthMarker", "ExitEastMarker", "ExitWestMarker"} {
		*memmap.PtrUint32(0x5D4594, uintptr(2487660+4*i)) = uint32(nox_xxx_getNameId_4E3AA0(internCStr(name)))
	}
}

func mapgenInitialPrefabs(theme *C.uint8_t, place func(*mapgenThemePrefabRecord) bool) bool {
	mapgenInitializePrefabExitTypes()
	sub_525DF0(theme)
	words := (*[mapgenThemeWordCount]C.uint32_t)(unsafe.Pointer(theme))
	const maxInitialPrefabs = 5
	placed := 0
	for _, required := range [...]bool{true, false} {
		for record := mapgenResolveThemePrefab(words[20]); record != nil; record = mapgenResolveThemePrefab(C.uint32_t(record.next)) {
			if (record.required != 0) != required {
				continue
			}
			if placed == maxInitialPrefabs {
				return !required
			}
			if !place(record) {
				return false
			}
			record.placed = 1
			placed++
		}
	}
	return true
}

//export nox_mapgenInitialPrefabsWithCallback_525D20
func nox_mapgenInitialPrefabsWithCallback_525D20(theme *C.uint8_t, callback C.nox_mapgen_prefab_room_cb_525d20) C.int {
	if theme == nil || callback == nil {
		return 0
	}
	return C.int(bool2int(mapgenInitialPrefabs(theme, func(prefab *mapgenThemePrefabRecord) bool {
		return ccall.CallIntPtr2(unsafe.Pointer(callback), unsafe.Pointer(theme), unsafe.Pointer(prefab)) != 0
	})))
}

//export nox_xxx_mapGen_InPrefab1_525D20
func nox_xxx_mapGen_InPrefab1_525D20(theme *C.uint8_t) C.int {
	if theme == nil {
		return 0
	}
	return C.int(bool2int(mapgenInitialPrefabs(theme, func(prefab *mapgenThemePrefabRecord) bool {
		return nox_xxx_mapGenPrefabMkRoom_526100(theme, (*C.uint8_t)(unsafe.Pointer(prefab))) != 0
	})))
}
