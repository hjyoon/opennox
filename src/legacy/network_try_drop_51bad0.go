package legacy

import (
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

func networkTryDropCall51BAD0(
	packet *[server.NetworkTryDropPacketSize51BAD0]byte,
	unit *server.Object,
	update *server.PlayerUpdateData,
) int32 {
	return GetServer().S().NetworkTryDrop51BAD0(
		unit,
		update,
		packet,
		server.NetworkTryDropRuntime51BAD0{
			NetDebug: func() bool {
				return noxflags.HasEngine(noxflags.EngineNetDebug)
			},
			TestHighBit: func(code uint16) {
				_ = code & 0x8000
			},
			Drop: func(owner, item *server.Object, point *types.Pointf) {
				objectDropBoundedCall4ED810(owner, item, point)
			},
		},
	)
}
