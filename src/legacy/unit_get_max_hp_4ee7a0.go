package legacy

import "github.com/opennox/opennox/v1/server"

func unitGetMaxHPCall4EE7A0(unit *server.Object) uint16 {
	return server.UnitGetMaxHP4EE7A0(unit)
}

// Nox_xxx_unitGetMaxHP_4EE7A0 exposes the native Go path to Go-owned callers
// without a CGo round trip.
func Nox_xxx_unitGetMaxHP_4EE7A0(unit *server.Object) uint16 {
	return unitGetMaxHPCall4EE7A0(unit)
}
