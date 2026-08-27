package legacy

import (
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

var Nox_server_tryPickup_51BAD0 func(*server.Object, *server.Object)

func networkTryGetCall51BAD0(
	packet *[server.NetworkTryGetPacketSize51BAD0]byte,
	unit *server.Object,
	update *server.PlayerUpdateData,
) int32 {
	srv := GetServer().S()
	return srv.NetworkTryGet51BAD0(
		unit,
		update,
		packet,
		server.NetworkTryGetRuntime51BAD0{
			NetDebug: func() bool {
				return noxflags.HasEngine(noxflags.EngineNetDebug)
			},
			TestHighBit: func(code uint16) {
				_ = code & 0x8000
			},
			GameBlocked: Nox_xxx_gameGet_4DB1B0,
			Pickup:      Nox_server_tryPickup_51BAD0,
			CarryingTooMuch: func(unit *server.Object) {
				srv.NetPriMsgToPlayer(unit, "pickup.c:CarryingTooMuch", 0)
			},
		},
	)
}
