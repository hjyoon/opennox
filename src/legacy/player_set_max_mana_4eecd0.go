package legacy

import "github.com/opennox/opennox/v1/server"

func playerSetMaxManaCall4EECD0(unit *server.Object, maximum uint16) uintptr {
	return server.PlayerSetMaxMana4EECD0(unit, maximum)
}

// Nox_xxx_playerSetMaxMana_4EECD0 exposes the native Go path to Go-owned
// callers. Its uintptr result preserves the original unit-or-UpdateData
// return register and must not be converted back into a Go pointer.
func Nox_xxx_playerSetMaxMana_4EECD0(unit *server.Object, maximum uint16) uintptr {
	return playerSetMaxManaCall4EECD0(unit, maximum)
}
