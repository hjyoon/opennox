package legacy

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func networkInventoryFailCall51BAD0(
	packet *[server.NetworkInventoryFailPacketSize51BAD0]byte,
	unit *server.Object,
) int32 {
	srv := GetServer().S()
	return srv.NetworkInventoryFail51BAD0(
		unit,
		packet,
		server.NetworkInventoryFailRuntime51BAD0{
			Drop: func(owner, item *server.Object, point *types.Pointf) {
				objectDropDispatchCall4ED790(owner, item, point)
			},
			CarryingTooMuch: func(unit *server.Object) {
				srv.NetPriMsgToPlayer(unit, "pickup.c:CarryingTooMuch", 0)
			},
		},
	)
}
