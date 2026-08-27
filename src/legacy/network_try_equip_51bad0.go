package legacy

import (
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

func networkTryEquipCall51BAD0(
	packet *[server.NetworkTryEquipPacketSize51BAD0]byte,
	unit *server.Object,
	update *server.PlayerUpdateData,
) int32 {
	return GetServer().S().NetworkTryEquip51BAD0(
		unit,
		update,
		packet,
		server.NetworkTryEquipRuntime51BAD0{
			NetDebug: func() bool {
				return noxflags.HasEngine(noxflags.EngineNetDebug)
			},
			TestHighBit: func(code uint16) {
				_ = code & 0x8000
			},
			GameBlocked: Nox_xxx_gameGet_4DB1B0,
			TryEquip: func(owner, item *server.Object) {
				Nox_xxx_playerTryEquip_4F2F70(owner, item)
			},
		},
	)
}
