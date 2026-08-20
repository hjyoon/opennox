package legacy

import "github.com/opennox/opennox/v1/server"

func playerGetMaxManaCall4EECB0(unit *server.Object) uint16 {
	return server.PlayerGetMaxMana4EECB0(unit)
}

// Nox_xxx_playerGetMaxMana_4EECB0 exposes the native Go path to Go-owned
// callers without a CGo round trip.
func Nox_xxx_playerGetMaxMana_4EECB0(unit *server.Object) uint16 {
	return playerGetMaxManaCall4EECB0(unit)
}
