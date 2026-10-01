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

var questHecubahSpawnCall51A5A0 = func(position *types.Pointf) {
	outer := GetServer()
	outer.S().QuestSpawnHecubah51A5A0(position, server.QuestHecubahSpawnRuntime51A5A0{
		HealthScale: func() float64 { return float64(C.sub_4E40F0()) },
		SetHP:       Nox_xxx_unitSetHP_4E4560,
		CreateAt:    func(unit *server.Object, position types.Pointf) { outer.CreateObjectAt(unit, nil, position) },
		Stage:       func() uint32 { return uint32(Nox_game_getQuestStage_4E3CC0()) },
		InventoryPut: func(owner, item *server.Object, report int32) {
			Nox_xxx_inventoryPutImpl_4F3070(owner, item, int(report))
		},
	})
}

//export nox_xxx_spawnHecubahQuest_native_51A5A0
func nox_xxx_spawnHecubahQuest_native_51A5A0(position *C.float2) {
	questHecubahSpawnCall51A5A0((*types.Pointf)(unsafe.Pointer(position)))
}

func questHecubahSpawnCEntry51A5A0(position *types.Pointf) {
	C.nox_xxx_spawnHecubahQuest_51A5A0((*C.int)(unsafe.Pointer(position)))
}
