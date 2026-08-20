package legacy

import "github.com/opennox/opennox/v1/server"

func unitGetOldManaCall4EEC80(unit *server.Object) uint16 {
	return server.UnitGetOldMana4EEC80(unit)
}

// Nox_xxx_unitGetOldMana_4EEC80 exposes the native Go path to Go-owned
// callers without a CGo round trip. The historical name refers to the live
// current-mana word in the original Player update data.
func Nox_xxx_unitGetOldMana_4EEC80(unit *server.Object) uint16 {
	return unitGetOldManaCall4EEC80(unit)
}
