package legacy

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func audEventDropCall4EE2F0(
	owner, item *server.Object,
	point *types.Pointf,
) int32 {
	outer := GetServer()
	return outer.S().AudEventDrop4EE2F0(owner, item, point, server.AudEventDropRuntime4EE2F0{
		DefaultDrop: defaultDropCall4ED290,
	})
}
