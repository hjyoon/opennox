package legacy

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func ankhTradableDropCall4EE370(
	owner, item *server.Object,
	point *types.Pointf,
) int32 {
	outer := GetServer()
	return outer.S().AnkhTradableDrop4EE370(owner, item, point, server.AnkhTradableDropRuntime4EE370{
		DefaultDrop: defaultDropCall4ED290,
	})
}
