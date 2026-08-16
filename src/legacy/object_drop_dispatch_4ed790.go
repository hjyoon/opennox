package legacy

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func objectDropDispatchRuntime4ED790() server.ObjectDropRuntime4ED790 {
	return server.ObjectDropRuntime4ED790{
		DefaultDrop: defaultDropCall4ED290,
		RefreshUnit: Nox_xxx_unit_511810,
	}
}

func objectDropDispatchCall4ED790(
	owner, item *server.Object,
	point *types.Pointf,
) int32 {
	return GetServer().S().ObjectDrop4ED790(
		owner,
		item,
		point,
		objectDropDispatchRuntime4ED790(),
	)
}
