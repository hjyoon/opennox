package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

//export sub_52BAF0
func sub_52BAF0(objC *nox_object_t) unsafe.Pointer {
	obj := asObjectS(objC)
	if obj == nil || obj.UpdateData == nil {
		return nil
	}
	resolveMonsterActionRefs52BAF0(obj, nil)
	return obj.UpdateData
}

// The original C entry reads IDs from the argument slots. The map loader
// keeps those IDs in refs until the objects and waypoints are available.
func resolveMonsterActionRefs52BAF0(obj *server.Object, refs *server.MonsterActionRefs) {
	update := obj.UpdateDataMonster()
	srv := obj.Server()
	for i := 0; i <= int(update.AIStackInd) && i < len(update.AIStack); i++ {
		item := &update.AIStack[i]
		if item.Action >= 72 {
			continue
		}
		descriptor := 16 * uintptr(item.Action)
		argCount := int32(memmap.Uint32(0x587000, 255604+descriptor))
		for j := 0; j < int(argCount) && j < len(item.Args)/2; j++ {
			slot := 2 * j
			switch memmap.Uint32(0x587000, 255608+descriptor+4*uintptr(j)) {
			case 1:
				id := int32(item.Args[slot])
				if refs != nil {
					id = refs.AIObjects[i][j]
				}
				item.Args[slot] = 0
				if id != 0 {
					item.Args[slot] = uintptr(unsafe.Pointer(srv.ObjectByScriptID4ECF10(id)))
				}
			case 2:
				index := uint32(item.Args[slot])
				if refs != nil {
					index = refs.AIWaypoints[i][j]
				}
				item.Args[slot] = 0
				if index != 0 {
					item.Args[slot] = uintptr(unsafe.Pointer(srv.WPs.ByInd(int(index))))
				}
			}
		}
	}
}
