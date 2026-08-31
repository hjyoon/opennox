package legacy

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

func networkTryCollideCall51BAD0(
	packet *[server.NetworkTryCollidePacketSize51BAD0]byte,
	unit *server.Object,
	update *server.PlayerUpdateData,
) int32 {
	return GetServer().S().NetworkTryCollide51BAD0(
		unit,
		update,
		packet,
		server.NetworkTryCollideRuntime51BAD0{
			NetDebug: func() bool {
				return noxflags.HasEngine(noxflags.EngineNetDebug)
			},
			TestHighBit: func(code uint16) {
				_ = code & 0x8000
			},
			CallCollide: func(callback unsafe.Pointer, target, unit *server.Object) {
				ccall.CallVoidUPtr3(
					callback,
					uintptr(target.CObj()),
					uintptr(unit.CObj()),
					0,
				)
			},
		},
	)
}
