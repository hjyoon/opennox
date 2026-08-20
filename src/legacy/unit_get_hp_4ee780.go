package legacy

import "github.com/opennox/opennox/v1/server"

func unitGetHPCall4EE780(unit *server.Object) uint16 {
	return server.UnitGetHP4EE780(unit)
}

// Nox_xxx_unitGetHP_4EE780 exposes the native Go path to Go-owned callers
// without a CGo round trip.
func Nox_xxx_unitGetHP_4EE780(unit *server.Object) uint16 {
	return unitGetHPCall4EE780(unit)
}
