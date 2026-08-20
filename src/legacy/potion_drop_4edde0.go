package legacy

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func potionDropCall4EDDE0(
	owner, item *server.Object,
	point *types.Pointf,
) int32 {
	outer := GetServer()
	return outer.S().PotionDrop4EDDE0(owner, item, point, server.PotionDropRuntime4EDDE0{
		DefaultDrop: defaultDropCall4ED290,
	})
}
