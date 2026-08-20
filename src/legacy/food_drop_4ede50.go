package legacy

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func foodDropCall4EDE50(
	owner, food *server.Object,
	point *types.Pointf,
) int32 {
	outer := GetServer()
	return outer.S().FoodDrop4EDE50(owner, food, point, server.FoodDropRuntime4EDE50{
		DefaultDrop: defaultDropCall4ED290,
	})
}
