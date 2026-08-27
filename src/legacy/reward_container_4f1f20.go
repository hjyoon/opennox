package legacy

/*
#include "GAME3_3.h"
*/
import "C"

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func rewardContainerRuntime4F1F20(outer Server) server.RewardContainerRuntime4F1F20 {
	s := outer.S()
	return server.RewardContainerRuntime4F1F20{
		QuestStage: func() uint32 {
			return uint32(Nox_game_getQuestStage_4E3CC0())
		},
		PreprocessMarkers: func() {
			C.sub_4F2110()
		},
		PreprocessRewards: func() {
			C.sub_4F2210()
		},
		ActivateMarker: func(marker *server.Object, stage uint32) *server.Object {
			return rewardMarkerActivateCall4F0720(
				s,
				marker,
				stage,
				rewardMarkerActivateRuntime4F0720(s),
			)
		},
		CreateAt: func(object, owner *server.Object, point types.Pointf) {
			outer.CreateObjectAt(object, owner, point)
		},
		DelayedDelete:   outer.DelayedDelete,
		DetachInventory: inventoryDetach4ED0C0,
		InventoryPut: func(owner, item *server.Object, mode uint32) {
			inventoryPutImpl4F3070(owner, item, mode != 0)
		},
	}
}

func rewardContainerProcessCall4F1F20() {
	outer := GetServer()
	outer.S().RewardContainerProcess4F1F20(rewardContainerRuntime4F1F20(outer))
}
