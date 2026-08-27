package legacy

import (
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

func networkTryDequipCall51BAD0(
	packet *[server.NetworkTryDequipPacketSize51BAD0]byte,
	unit *server.Object,
	update *server.PlayerUpdateData,
) int32 {
	return GetServer().S().NetworkTryDequip51BAD0(
		unit,
		update,
		packet,
		server.NetworkTryDequipRuntime51BAD0{
			NetDebug: func() bool {
				return noxflags.HasEngine(noxflags.EngineNetDebug)
			},
			TestHighBit: func(code uint16) {
				_ = code & 0x8000
			},
			TryDequip: func(owner, item *server.Object) {
				Nox_xxx_playerTryDequip_4F2FB0(owner, item)
			},
		},
	)
}
