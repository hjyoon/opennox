package legacy

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func rewardAnkhReplaceRuntime4F2110(outer Server) server.RewardAnkhReplaceRuntime4F2110 {
	return server.RewardAnkhReplaceRuntime4F2110{
		CreateAt: func(object, owner *server.Object, point types.Pointf) {
			outer.CreateObjectAt(object, owner, point)
		},
		DelayedDelete: outer.DelayedDelete,
	}
}
