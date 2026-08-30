package legacy

import (
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

func networkTryUseCall51BAD0(
	packet *[server.NetworkTryUsePacketSize51BAD0]byte,
	unit *server.Object,
	update *server.PlayerUpdateData,
) int32 {
	return GetServer().S().NetworkTryUse51BAD0(
		unit,
		update,
		packet,
		server.NetworkTryUseRuntime51BAD0{
			NetDebug: func() bool {
				return noxflags.HasEngine(noxflags.EngineNetDebug)
			},
			TestHighBit: func(code uint16) {
				_ = code & 0x8000
			},
			GameBlocked: Nox_xxx_gameGet_4DB1B0,
		},
	)
}
