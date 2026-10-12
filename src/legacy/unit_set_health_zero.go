package legacy

import "github.com/opennox/opennox/v1/server"

// UnitSetHealthZero4EE5E0 uses the same native effects as lethal damage, while
// retaining direct SetHealth assignment semantics rather than its GodMode gate.
func UnitSetHealthZero4EE5E0(unit *server.Object) {
	GetServer().S().UnitSetHealthZero4EE5E0(unit, unitDamageClearRuntime4EE5E0())
}
