package legacy

import (
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

func networkInfoBookCall51BAD0(
	packet *[server.NetworkInfoBookPacketSize51BAD0]byte,
	unit *server.Object,
	update *server.PlayerUpdateData,
) int32 {
	srv := GetServer().S()
	return srv.NetworkInfoBook51BAD0(
		unit,
		update,
		packet,
		server.NetworkInfoBookRuntime51BAD0{
			NetDebug: func() bool {
				return noxflags.HasEngine(noxflags.EngineNetDebug)
			},
			TestHighBit: func(code uint16) {
				_ = code & 0x8000
			},
			Send: func(recipient uint8, response [server.NetworkInfoBookPacketSize51BAD0]byte) {
				srv.NetSendPacketXxx0(int(recipient), response[:], nil, 1)
			},
		},
	)
}
