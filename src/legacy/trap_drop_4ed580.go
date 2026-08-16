package legacy

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func trapDropCall4ED580(
	owner, trap *server.Object,
	point *types.Pointf,
) int32 {
	outer := GetServer()
	return outer.S().TrapDrop4ED580(owner, trap, point, server.TrapDropRuntime4ED580{
		MapTileAllowTeleport: mapTileAllowTeleport411A90,
		DefaultDrop:          defaultDropCall4ED290,
	})
}
