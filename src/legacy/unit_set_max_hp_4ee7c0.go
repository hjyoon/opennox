package legacy

import "github.com/opennox/opennox/v1/server"

func unitSetMaxHPCall4EE7C0(unit *server.Object, maximum uint16) *server.HealthData {
	return server.UnitSetMaxHP4EE7C0(unit, maximum)
}

// Nox_xxx_unitSetMaxHP_4EE7C0 exposes the native Go path to Go-owned callers
// without a CGo round trip and preserves the returned HealthData identity.
func Nox_xxx_unitSetMaxHP_4EE7C0(unit *server.Object, maximum uint16) *server.HealthData {
	return unitSetMaxHPCall4EE7C0(unit, maximum)
}
