package legacy

import "github.com/opennox/opennox/v1/server"

func networkGauntletCall51BAD0(
	packet *[server.NetworkGauntletPacketSize51BAD0]byte,
	unit *server.Object,
	update *server.PlayerUpdateData,
) int32 {
	return GetServer().S().NetworkGauntlet51BAD0(
		unit,
		update,
		packet,
		server.NetworkGauntletRuntime51BAD0{
			Respawn: func(unit *server.Object) {
				_ = playerRespawnCall4F7EF0(unit)
			},
			Exit: Sub_4DD0B0,
		},
	)
}

// NetworkGauntletCall51BAD0 exposes the native-width packet handler to the Go
// server dispatcher so valid packets never enter the PE32 C switch.
func NetworkGauntletCall51BAD0(
	packet *[server.NetworkGauntletPacketSize51BAD0]byte,
	unit *server.Object,
	update *server.PlayerUpdateData,
) int32 {
	return networkGauntletCall51BAD0(packet, unit, update)
}
