package legacy

/*
#include "GAME3_3.h"
#include "GAME4_1.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

var questNecroSpawnCall51A7A0 = func(position *types.Pointf) {
	outer := GetServer()
	outer.S().QuestSpawnNecro51A7A0(position, server.QuestNecroSpawnRuntime51A7A0{
		HealthScale: func() float64 { return float64(C.sub_4E40F0()) },
		SetHP:       Nox_xxx_unitSetHP_4E4560,
		CreateAt: func(unit *server.Object, position types.Pointf) {
			outer.CreateObjectAt(unit, nil, position)
		},
		Stage: func() uint32 { return uint32(Nox_game_getQuestStage_4E3CC0()) },
		InventoryPut: func(owner, item *server.Object, report int32) {
			Nox_xxx_inventoryPutImpl_4F3070(owner, item, int(report))
		},
	})
}

//export nox_xxx_spawnNecroQuest_native_51A7A0
func nox_xxx_spawnNecroQuest_native_51A7A0(position *C.float2) {
	// Preserve the late position dereference, including failed allocation with
	// a null position; conversion here only carries the native address.
	questNecroSpawnCall51A7A0((*types.Pointf)(unsafe.Pointer(position)))
}

func questNecroSpawnCEntry51A7A0(position *types.Pointf) {
	C.nox_xxx_spawnNecroQuest_51A7A0((*C.int)(unsafe.Pointer(position)))
}
